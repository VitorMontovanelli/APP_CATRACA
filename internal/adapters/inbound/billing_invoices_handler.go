package inbound

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type BillingInvoicesHandler struct {
	ctx       context.Context
	db        *gorm.DB
	session   *SessionManager
	recurring *BillingRecurringHandler
}

func NewBillingInvoicesHandler(db *gorm.DB, session *SessionManager, recurring *BillingRecurringHandler) *BillingInvoicesHandler {
	return &BillingInvoicesHandler{db: db, session: session, recurring: recurring}
}

func (h *BillingInvoicesHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *BillingInvoicesHandler) GerarInvoice(studentPlanID uint) (*models.Invoice, error) {
	type spMini struct {
		ID, StudentID, PlanID uint
		PlanPriceCents        int
		Status                string
	}
	var sp spMini
	query := `SELECT sp.id, sp.student_id, sp.plan_id, p.price_cents, sp.status
		FROM student_plans sp JOIN plans p ON p.id = sp.plan_id WHERE sp.id = ?`
	if err := h.db.Raw(query, studentPlanID).Scan(&sp).Error; err != nil {
		return nil, err
	}
	if sp.Status != "active" {
		return nil, fmt.Errorf("plano do aluno não está ativo")
	}

	inv := models.Invoice{
		StudentPlanID: studentPlanID,
		StudentID:     sp.StudentID,
		PlanID:        sp.PlanID,
		AmountCents:   sp.PlanPriceCents,
		DueDate:       time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}
	if err := h.db.Create(&inv).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAudit("gerou_fatura", inv.ID)
	return h.BuscarInvoice(inv.ID)
}

func (h *BillingInvoicesHandler) BuscarInvoice(id uint) (*models.Invoice, error) {
	type invRaw struct {
		models.Invoice
		StudentName       string
		PlanName          string
		PaymentMethodName *string
	}
	var raw invRaw
	query := `SELECT i.id, i.student_plan_id, i.student_id, i.plan_id, i.amount_cents,
		i.status, i.payment_method_id, i.due_date, i.paid_at, i.paid_amount_cents,
		i.gateway_transaction_id, i.pix_qr_code, i.pix_br_code, i.notes, i.created_at,
		s.nome AS student_name, p.name AS plan_name, pm.name AS payment_method_name
		FROM invoices i JOIN students s ON s.id = i.student_id JOIN plans p ON p.id = i.plan_id
		LEFT JOIN payment_methods pm ON pm.id = i.payment_method_id WHERE i.id = ?`
	if err := h.db.Raw(query, id).Scan(&raw).Error; err != nil {
		return nil, err
	}
	raw.Invoice.StudentName = raw.StudentName
	raw.Invoice.PlanName = raw.PlanName
	if raw.PaymentMethodName != nil {
		raw.Invoice.PaymentMethodName = *raw.PaymentMethodName
	}
	return &raw.Invoice, nil
}

func (h *BillingInvoicesHandler) ListarInvoices() ([]models.Invoice, error) {
	type invRaw struct {
		models.Invoice
		StudentName       string
		PlanName          string
		PaymentMethodName string
	}
	var raw []invRaw
	query := `SELECT i.id, i.student_plan_id, i.student_id, i.plan_id, i.amount_cents,
		i.status, i.payment_method_id, i.due_date, i.paid_at, i.paid_amount_cents,
		i.gateway_transaction_id, i.pix_qr_code, i.pix_br_code, i.notes, i.created_at,
		s.nome AS student_name, p.name AS plan_name, COALESCE(pm.name, '') AS payment_method_name
		FROM invoices i JOIN students s ON s.id = i.student_id JOIN plans p ON p.id = i.plan_id
		LEFT JOIN payment_methods pm ON pm.id = i.payment_method_id ORDER BY i.due_date DESC`
	if err := h.db.Raw(query).Scan(&raw).Error; err != nil {
		return nil, err
	}
	list := make([]models.Invoice, len(raw))
	for i, r := range raw {
		list[i] = r.Invoice
		list[i].StudentName = r.StudentName
		list[i].PlanName = r.PlanName
		list[i].PaymentMethodName = r.PaymentMethodName
	}
	return list, nil
}

func (h *BillingInvoicesHandler) ConfirmarPagamento(invoiceID uint, paymentMethodID uint, paidAmountCents int) error {
	userName := h.session.ActorCargo()
	err := h.db.Exec(
		"UPDATE invoices SET status='paid', payment_method_id=?, paid_at=datetime('now'), paid_amount_cents=?, notes=COALESCE(notes||' | ','')||'Confirmado por '||? WHERE id=?",
		paymentMethodID, paidAmountCents, userName, invoiceID,
	).Error
	if err == nil {
		h.session.RegistrarAuditComDetalhes("confirmou_pagamento", invoiceID, fmt.Sprintf("valor=%d, metodo=%d", paidAmountCents, paymentMethodID))
		h.recurring.GerarProximaFatura(invoiceID)
	}
	return err
}

func (h *BillingInvoicesHandler) CancelarInvoice(invoiceID uint) error {
	err := h.db.Model(&models.Invoice{}).Where("id = ?", invoiceID).Update("status", "cancelled").Error
	if err == nil {
		h.session.RegistrarAudit("cancelou_fatura", invoiceID)
	}
	return err
}

func (h *BillingInvoicesHandler) DeletarInvoice(invoiceID uint) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	n := h.db.Delete(&models.Invoice{}, invoiceID).RowsAffected
	if n == 0 {
		return fmt.Errorf("fatura não encontrada")
	}
	h.session.RegistrarAudit("deletou_fatura", invoiceID)
	return nil
}

func (h *BillingInvoicesHandler) SalvarComprovante(invoiceID uint, base64Data, fileName string) error {
	dir := "comprovantes"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".pdf" {
		return fmt.Errorf("formato de arquivo inválido: apenas comprovantes em formato PDF são permitidos")
	}
	dest := filepath.Join(dir, fmt.Sprintf("invoice_%d%s", invoiceID, ext))

	_, raw, found := strings.Cut(base64Data, ",")
	if !found {
		raw = base64Data
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("erro ao decodificar arquivo: %w", err)
	}
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}

	err = h.db.Exec("UPDATE invoices SET comprovante = ?, status = 'paid', paid_at = ? WHERE id = ?", dest, time.Now(), invoiceID).Error
	if err != nil {
		return fmt.Errorf("erro ao atualizar fatura: %w", err)
	}
	h.session.RegistrarAuditComDetalhes("confirmou_pagamento", invoiceID, fmt.Sprintf("comprovante=%s, invoice=%d", dest, invoiceID))
	h.recurring.GerarProximaFatura(invoiceID)
	return nil
}
