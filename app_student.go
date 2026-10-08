package main

import (
	"fmt"
	"time"
)

// =============== Alunos (Students CRUD) ===============

func (a *App) ListarStudents() ([]Student, error) {
	list := []Student{}
	err := a.db.Order("nome").Find(&list).Error
	return list, err
}

func (a *App) CriarStudent(nome, cpf, dataNascimento, telefone, email string, formaPagamentoID uint) (*Student, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	student := Student{
		Nome:             nome,
		CPF:              cpf,
		DataNascimento:   strPtr(dataNascimento),
		Telefone:         strPtr(telefone),
		Email:            strPtr(email),
		FormaPagamentoID: uintPtr(formaPagamentoID),
	}
	if err := a.db.Create(&student).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_aluno", student.ID)
	return a.BuscarStudent(student.ID)
}

func (a *App) BuscarStudent(id uint) (*Student, error) {
	var s Student
	err := a.db.First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (a *App) AtualizarStudent(id uint, nome, cpf, dataNascimento, telefone, telefoneUrgencia, email string, formaPagamentoID uint) (*Student, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	student, err := a.BuscarStudent(id)
	if err != nil {
		return nil, err
	}
	student.Nome = nome
	student.CPF = cpf
	student.DataNascimento = strPtr(dataNascimento)
	student.Telefone = strPtr(telefone)
	student.TelefoneUrgencia = strPtr(telefoneUrgencia)
	student.Email = strPtr(email)
	student.FormaPagamentoID = uintPtr(formaPagamentoID)

	if err := a.db.Save(student).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("editou_aluno", id)
	return a.BuscarStudent(id)
}

func (a *App) AtivarStudent(id uint, ativo bool) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&Student{}).Where("id = ?", id).Update("ativo", ativo).Error
	if err == nil {
		action := "desativou_aluno"
		if ativo {
			action = "ativou_aluno"
		}
		a.registrarAudit(action, id)
	}
	return err
}

func (a *App) ListarStudentsComPlanos() ([]StudentComPlano, error) {
	type rawResult struct {
		ID               uint
		Nome             string
		CPF              string
		DataNascimento   *string
		Telefone         *string
		TelefoneUrgencia *string
		Email            *string
		LaudoMedico      *string
		Foto             *string
		FormaPagamentoID *uint
		DataEntrada      string
		Observacao       *string
		Ativo            bool
		CreatedAt        string
		UpdatedAt        *string
		PlanoNome        *string
		PlanoPreco       *int
		PlanoStatus      *string
		StudentPlanID    *uint
		PaymentMethod    *string
		DueDay           *int
		UltimaFatura     *string
		FaturaValor      *int
		FaturaVencimento *string
	}

	var raw []rawResult
	err := a.db.Raw(`
		SELECT
			s.id, s.nome, s.cpf, s.data_nascimento, s.telefone, s.telefone_urgencia, s.email, s.laudo_medico, s.foto,
			s.forma_pagamento_id, s.data_entrada, s.observacao, s.ativo,
			s.created_at, s.updated_at,
			p.name, p.price_cents, sp.status,
			sp.id, pm.name, sp.due_day,
			i.status, i.amount_cents, i.due_date
		FROM students s
		LEFT JOIN student_plans sp ON sp.student_id = s.id AND sp.status = 'active'
		LEFT JOIN plans p ON p.id = sp.plan_id
		LEFT JOIN payment_methods pm ON pm.id = COALESCE(sp.payment_method_id, s.forma_pagamento_id)
		LEFT JOIN invoices i ON i.id = (
			SELECT id FROM invoices WHERE student_id = s.id ORDER BY due_date DESC LIMIT 1
		)
		ORDER BY s.nome
	`).Scan(&raw).Error
	if err != nil {
		return nil, err
	}

	list := make([]StudentComPlano, len(raw))
	for i, r := range raw {
		de := time.Time{}
		ca := time.Time{}
		var ua time.Time
		if t, err := time.Parse(time.RFC3339, r.DataEntrada); err == nil {
			de = t
		}
		if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
			ca = t
		}
		if r.UpdatedAt != nil {
			if t, err := time.Parse(time.RFC3339, *r.UpdatedAt); err == nil {
				ua = t
			}
		}

		list[i] = StudentComPlano{
			Student: Student{
				ID:               r.ID,
				Nome:             r.Nome,
				CPF:              r.CPF,
				DataNascimento:   r.DataNascimento,
				Telefone:         r.Telefone,
				TelefoneUrgencia: r.TelefoneUrgencia,
				Email:            r.Email,
				LaudoMedico:      r.LaudoMedico,
				Foto:             r.Foto,
				FormaPagamentoID: r.FormaPagamentoID,
				DataEntrada:      de,
				Observacao:       r.Observacao,
				Ativo:            r.Ativo,
				CreatedAt:        ca,
				UpdatedAt:        ua,
			},
			PlanoNome:        r.PlanoNome,
			PlanoPreco:       r.PlanoPreco,
			PlanoStatus:      r.PlanoStatus,
			StudentPlanID:    r.StudentPlanID,
			PaymentMethod:    r.PaymentMethod,
			DueDay:           r.DueDay,
			UltimaFatura:     r.UltimaFatura,
			FaturaValor:      r.FaturaValor,
			FaturaVencimento: r.FaturaVencimento,
		}
	}
	return list, nil
}
