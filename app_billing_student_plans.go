package main

import (
	"fmt"
	"time"
)

// =============== Planos dos Alunos (StudentPlans) ===============

func (a *App) ListarStudentPlans() ([]StudentPlan, error) {
	type spRaw struct {
		StudentPlan
		StudentName    string
		PlanName       string
		PlanPriceCents int
	}
	var raw []spRaw
	err := a.db.Raw(`
		SELECT sp.id, sp.student_id, sp.plan_id, sp.status, sp.payment_method_id,
			sp.due_day, sp.start_date, sp.end_date, sp.cancelled_at, sp.notes,
			sp.created_at, sp.updated_at,
			s.nome AS student_name, p.name AS plan_name, p.price_cents AS plan_price_cents
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		JOIN plans p ON p.id = sp.plan_id
		ORDER BY sp.start_date DESC
	`).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	list := make([]StudentPlan, len(raw))
	for i, r := range raw {
		list[i] = r.StudentPlan
		list[i].StudentName = r.StudentName
		list[i].PlanName = r.PlanName
		list[i].PlanPriceCents = r.PlanPriceCents
	}
	return list, nil
}

func (a *App) CriarStudentPlan(studentID, planID uint, dueDay int) (*StudentPlan, error) {
	sp := StudentPlan{
		StudentID: studentID,
		PlanID:    planID,
		DueDay:    dueDay,
	}
	if err := a.db.Create(&sp).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_plano_aluno", sp.ID)
	return a.BuscarStudentPlan(sp.ID)
}

func (a *App) BuscarStudentPlan(id uint) (*StudentPlan, error) {
	type spRaw struct {
		StudentPlan
		StudentName    string
		PlanName       string
		PlanPriceCents int
	}
	var raw spRaw
	err := a.db.Raw(`
		SELECT sp.id, sp.student_id, sp.plan_id, sp.status, sp.payment_method_id,
			sp.due_day, sp.start_date, sp.end_date, sp.cancelled_at, sp.notes,
			sp.created_at, sp.updated_at,
			s.nome AS student_name, p.name AS plan_name, p.price_cents AS plan_price_cents
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		JOIN plans p ON p.id = sp.plan_id
		WHERE sp.id = ?
	`, id).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	raw.StudentPlan.StudentName = raw.StudentName
	raw.StudentPlan.PlanName = raw.PlanName
	raw.StudentPlan.PlanPriceCents = raw.PlanPriceCents
	return &raw.StudentPlan, nil
}

func (a *App) CancelarStudentPlan(id uint) error {
	err := a.db.Model(&StudentPlan{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "cancelled",
		"cancelled_at": time.Now(),
	}).Error
	if err == nil {
		a.registrarAudit("cancelou_plano_aluno", id)
	}
	return err
}

func (a *App) AtualizarPlanoAluno(studentID, planID, formaPagamentoID uint, dueDay int) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}

	var current StudentPlan
	hasCurrent := a.db.Where("student_id = ? AND status = 'active'", studentID).First(&current).Error == nil

	if planID == 0 {
		if hasCurrent {
			return a.db.Model(&current).Updates(map[string]interface{}{
				"status":       "cancelled",
				"cancelled_at": time.Now(),
			}).Error
		}
		return nil
	}

	if hasCurrent && current.PlanID == planID {
		return a.db.Model(&current).Updates(map[string]interface{}{
			"payment_method_id": uintPtr(formaPagamentoID),
			"due_day":           dueDay,
		}).Error
	}

	if hasCurrent {
		a.db.Model(&current).Updates(map[string]interface{}{
			"status":       "cancelled",
			"cancelled_at": time.Now(),
		})
	}

	sp := StudentPlan{
		StudentID:       studentID,
		PlanID:          planID,
		DueDay:          dueDay,
		PaymentMethodID: uintPtr(formaPagamentoID),
	}
	if err := a.db.Create(&sp).Error; err != nil {
		return err
	}
	a.registrarAuditComDetalhes("vinculou_plano_aluno", sp.ID,
		fmt.Sprintf("aluno=%d, plano=%d, forma_pagamento=%d, due_day=%d", studentID, planID, formaPagamentoID, dueDay))
	return nil
}

// =============== Faturas (Invoices) ===============
