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
	query := `SELECT sp.id, sp.student_id, sp.plan_id, p.price_cents AS plan_price_cents, sp.status
		FROM student_plans sp JOIN plans p ON p.id = sp.plan_id WHERE sp.id = ?`
	if err := h.db.Raw(query, studentPlanID).Scan(&sp).Error; err != nil {
		return nil, err
	}
	// Se o plano estiver sem preço cadastrado, usa o preço atual do plano
	// diretamente como reforço (o alias acima já evita o valor zerado).
	if sp.PlanPriceCents == 0 {
		var preco int
		if err := h.db.Raw("SELECT price_cents FROM plans WHERE id = ?", sp.PlanID).Scan(&preco).Error; err == nil && preco > 0 {
			sp.PlanPriceCents = preco
		}
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
		COALESCE(i.confirmed_by, '') AS confirmed_by, i.refunded_by, i.refunded_at,
		i.paid_method_name, i.paid_fee_percent, i.paid_fee_fixed_cents,
		s.nome AS student_name, p.name AS plan_name,
		COALESCE(NULLIF(i.paid_method_name, ''), pm.name, '') AS payment_method_name
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

func (h *BillingInvoicesHandler) listarInvoices(where string, args ...interface{}) ([]models.Invoice, error) {
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
		COALESCE(i.confirmed_by, '') AS confirmed_by,
		i.refunded_by, i.refunded_at,
		i.paid_method_name, i.paid_fee_percent, i.paid_fee_fixed_cents,
		s.nome AS student_name, p.name AS plan_name,
		COALESCE(NULLIF(i.paid_method_name, ''), pm.name, '') AS payment_method_name
		FROM invoices i JOIN students s ON s.id = i.student_id JOIN plans p ON p.id = i.plan_id
		LEFT JOIN payment_methods pm ON pm.id = i.payment_method_id WHERE ` + where
	if err := h.db.Raw(query, args...).Scan(&raw).Error; err != nil {
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

func (h *BillingInvoicesHandler) ListarInvoices() ([]models.Invoice, error) {
	return h.listarInvoices("1 = 1 ORDER BY i.due_date DESC, i.id DESC")
}

// ListarInvoicesPendentes devolve apenas as faturas em aberto (pendentes e
// vencidas). Faturas pagas ficam exclusivamente no histórico.
func (h *BillingInvoicesHandler) ListarInvoicesPendentes() ([]models.Invoice, error) {
	return h.listarInvoices("i.status IN ('pending', 'overdue') ORDER BY i.due_date ASC, i.id ASC")
}

// ListarHistoricoPagamentos devolve as faturas pagas, da mais recente para a
// mais antiga, com os dados de quem confirmou o pagamento.
func (h *BillingInvoicesHandler) ListarHistoricoPagamentos() ([]models.Invoice, error) {
	return h.listarInvoices("i.status = 'paid' ORDER BY i.paid_at DESC, i.id DESC")
}

func (h *BillingInvoicesHandler) actorName() string {
	actorID := h.session.ActorID()
	if actorID > 0 {
		var name string
		h.db.Model(&models.User{}).Where("id = ?", actorID).Pluck("name", &name)
		if strings.TrimSpace(name) != "" {
			return name
		}
	}
	if cargo := h.session.ActorCargo(); cargo != "" {
		return cargo
	}
	return "Sistema"
}

func (h *BillingInvoicesHandler) ConfirmarPagamento(invoiceID uint, paymentMethodID uint, paidAmountCents int, paidDate string) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	var inv models.Invoice
	if err := h.db.First(&inv, invoiceID).Error; err != nil {
		return fmt.Errorf("fatura não encontrada")
	}
	if inv.Status == "paid" {
		return fmt.Errorf("esta fatura já está paga")
	}
	if inv.Status == "cancelled" {
		return fmt.Errorf("fatura cancelada não pode ser confirmada como paga")
	}
	if paymentMethodID == 0 {
		return fmt.Errorf("selecione a forma de pagamento")
	}
	if paidAmountCents < 0 {
		return fmt.Errorf("o valor pago não pode ser negativo")
	}

	var metodo models.PaymentMethod
	if err := h.db.First(&metodo, paymentMethodID).Error; err != nil {
		return fmt.Errorf("forma de pagamento não encontrada")
	}
	if !metodo.Enabled {
		return fmt.Errorf("esta forma de pagamento está desativada")
	}

	paidAt := time.Now()
	if strings.TrimSpace(paidDate) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(paidDate), time.Local)
		if err != nil {
			return fmt.Errorf("data do pagamento inválida")
		}
		paidAt = parsed
	}

	confirmedBy := h.actorName()
	err := h.db.Model(&models.Invoice{}).Where("id = ?", invoiceID).Updates(map[string]interface{}{
		"status":               "paid",
		"payment_method_id":    paymentMethodID,
		"paid_method_name":     metodo.Name,
		"paid_fee_percent":     metodo.FeePercent,
		"paid_fee_fixed_cents": metodo.FeeFixedCents,
		"paid_at":              paidAt,
		"paid_amount_cents":    paidAmountCents,
		"confirmed_by":         confirmedBy,
		"refunded_by":          "",
		"refunded_at":          nil,
	}).Error
	if err != nil {
		return err
	}
	h.session.RegistrarAuditComDetalhes("confirmou_pagamento", invoiceID,
		fmt.Sprintf("valor=%d, metodo=%d (%s), data=%s, por=%s", paidAmountCents, paymentMethodID, metodo.Name, paidAt.Format("2006-01-02"), confirmedBy))
	h.recurring.GerarProximaFatura(invoiceID)
	return nil
}

// EstornarPagamento devolve uma fatura paga para pendente, registrando quem
// estornou e quando. A fatura sai do histórico e volta para "Faturas Emitidas".
func (h *BillingInvoicesHandler) EstornarPagamento(invoiceID uint) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	var inv models.Invoice
	if err := h.db.First(&inv, invoiceID).Error; err != nil {
		return fmt.Errorf("fatura não encontrada")
	}
	if inv.Status != "paid" {
		return fmt.Errorf("somente faturas pagas podem ser estornadas")
	}

	now := time.Now()
	refundedBy := h.actorName()
	nota := fmt.Sprintf("Estornado por %s em %s", refundedBy, now.Format("02/01/2006 15:04"))
	err := h.db.Model(&models.Invoice{}).Where("id = ?", invoiceID).Updates(map[string]interface{}{
		"status":               "pending",
		"payment_method_id":    nil,
		"paid_method_name":     "",
		"paid_fee_percent":     0,
		"paid_fee_fixed_cents": 0,
		"paid_at":              nil,
		"paid_amount_cents":    nil,
		"confirmed_by":         "",
		"refunded_by":          refundedBy,
		"refunded_at":          now,
		"notes":                gorm.Expr("COALESCE(notes || ' | ', '') || ?", nota),
	}).Error
	if err != nil {
		return err
	}
	h.session.RegistrarAuditComDetalhes("estornou_pagamento", invoiceID,
		fmt.Sprintf("por=%s, em=%s", refundedBy, now.Format(time.RFC3339)))
	return nil
}

func (h *BillingInvoicesHandler) CancelarInvoice(invoiceID uint) error {
	var inv models.Invoice
	if err := h.db.First(&inv, invoiceID).Error; err != nil {
		return fmt.Errorf("fatura não encontrada")
	}
	if inv.Status == "paid" {
		return fmt.Errorf("fatura paga não pode ser cancelada; use o estorno")
	}
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
	var inv models.Invoice
	if err := h.db.First(&inv, invoiceID).Error; err != nil {
		return fmt.Errorf("fatura não encontrada")
	}
	if inv.Status == "paid" {
		return fmt.Errorf("fatura paga não pode ser excluída; use o estorno")
	}
	if err := h.db.Delete(&models.Invoice{}, invoiceID).Error; err != nil {
		return err
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
