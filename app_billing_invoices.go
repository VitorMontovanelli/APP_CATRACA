package main

import (
	"fmt"
	"time"
)

func (a *App) GerarInvoice(studentPlanID uint) (*Invoice, error) {
	type spMini struct {
		ID             uint
		StudentID      uint
		PlanID         uint
		PlanPriceCents int
		Status         string
	}
	var sp spMini
	err := a.db.Raw(`
		SELECT sp.id, sp.student_id, sp.plan_id, p.price_cents, sp.status
		FROM student_plans sp
		JOIN plans p ON p.id = sp.plan_id
		WHERE sp.id = ?
	`, studentPlanID).Scan(&sp).Error
	if err != nil {
		return nil, err
	}
	if sp.Status != "active" {
		return nil, fmt.Errorf("plano do aluno não está ativo")
	}

	inv := Invoice{
		StudentPlanID: studentPlanID,
		StudentID:     sp.StudentID,
		PlanID:        sp.PlanID,
		AmountCents:   sp.PlanPriceCents,
		DueDate:       time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}
	if err := a.db.Create(&inv).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("gerou_fatura", inv.ID)
	return a.BuscarInvoice(inv.ID)
}

func (a *App) BuscarInvoice(id uint) (*Invoice, error) {
	type invRaw struct {
		Invoice
		StudentName       string
		PlanName          string
		PaymentMethodName *string
	}
	var raw invRaw
	err := a.db.Raw(`
		SELECT i.id, i.student_plan_id, i.student_id, i.plan_id, i.amount_cents,
			i.status, i.payment_method_id, i.due_date, i.paid_at, i.paid_amount_cents,
			i.gateway_transaction_id, i.pix_qr_code, i.pix_br_code, i.notes, i.created_at,
			s.nome AS student_name, p.name AS plan_name, pm.name AS payment_method_name
		FROM invoices i
		JOIN students s ON s.id = i.student_id
		JOIN plans p ON p.id = i.plan_id
		LEFT JOIN payment_methods pm ON pm.id = i.payment_method_id
		WHERE i.id = ?
	`, id).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	raw.Invoice.StudentName = raw.StudentName
	raw.Invoice.PlanName = raw.PlanName
	if raw.PaymentMethodName != nil {
		raw.Invoice.PaymentMethodName = *raw.PaymentMethodName
	}
	return &raw.Invoice, nil
}

func (a *App) ListarInvoices() ([]Invoice, error) {
	type invRaw struct {
		Invoice
		StudentName       string
		PlanName          string
		PaymentMethodName string
	}
	var raw []invRaw
	err := a.db.Raw(`
		SELECT i.id, i.student_plan_id, i.student_id, i.plan_id, i.amount_cents,
			i.status, i.payment_method_id, i.due_date, i.paid_at, i.paid_amount_cents,
			i.gateway_transaction_id, i.pix_qr_code, i.pix_br_code, i.notes, i.created_at,
			s.nome AS student_name, p.name AS plan_name, COALESCE(pm.name, '') AS payment_method_name
		FROM invoices i
		JOIN students s ON s.id = i.student_id
		JOIN plans p ON p.id = i.plan_id
		LEFT JOIN payment_methods pm ON pm.id = i.payment_method_id
		ORDER BY i.due_date DESC
	`).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	list := make([]Invoice, len(raw))
	for i, r := range raw {
		list[i] = r.Invoice
		list[i].StudentName = r.StudentName
		list[i].PlanName = r.PlanName
		list[i].PaymentMethodName = r.PaymentMethodName
	}
	return list, nil
}

func (a *App) ConfirmarPagamento(invoiceID uint, paymentMethodID uint, paidAmountCents int) error {
	userName := a.actorCargo
	err := a.db.Exec(
		"UPDATE invoices SET status='paid', payment_method_id=?, paid_at=datetime('now'), paid_amount_cents=?, notes=COALESCE(notes||' | ','')||'Confirmado por '||? WHERE id=?",
		paymentMethodID, paidAmountCents, userName, invoiceID,
	).Error
	if err == nil {
		a.registrarAuditComDetalhes("confirmou_pagamento", invoiceID,
			fmt.Sprintf("valor=%d, metodo=%d", paidAmountCents, paymentMethodID))
		a.gerarProximaFatura(invoiceID)
	}
	return err
}

func (a *App) CancelarInvoice(invoiceID uint) error {
	err := a.db.Model(&Invoice{}).Where("id = ?", invoiceID).Update("status", "cancelled").Error
	if err == nil {
		a.registrarAudit("cancelou_fatura", invoiceID)
	}
	return err
}

func (a *App) DeletarInvoice(invoiceID uint) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	n := a.db.Delete(&Invoice{}, invoiceID).RowsAffected
	if n == 0 {
		return fmt.Errorf("fatura não encontrada")
	}
	a.registrarAudit("deletou_fatura", invoiceID)
	return nil
}

// =============== AccessLogs ===============

type CobrancaFatura struct {
	ID          uint    `json:"id"`
	AmountCents int     `json:"amount_cents"`
	Status      string  `json:"status"`
	DueDate     string  `json:"due_date"`
	PaidAt      *string `json:"paid_at"`
	Comprovante *string `json:"comprovante"`
	DiasVencido int     `json:"dias_vencido"`
}

type CobrancaAluno struct {
	StudentID   uint             `json:"student_id"`
	Nome        string           `json:"nome"`
	CPF         string           `json:"cpf"`
	PlanoNome   string           `json:"plano_nome"`
	PlanoPreco  int              `json:"plano_preco"`
	StatusPlano string           `json:"status_plano"`
	DueDay      int              `json:"due_day"`
	Faturas     []CobrancaFatura `json:"faturas"`
}
