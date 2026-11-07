package inbound

import (
	"context"
	"fmt"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type BillingRecurringHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewBillingRecurringHandler(db *gorm.DB, session *SessionManager) *BillingRecurringHandler {
	return &BillingRecurringHandler{db: db, session: session}
}

func (h *BillingRecurringHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func normalizarDueDay(day int) int {
	if day < 1 {
		return 5
	}
	if day > 31 {
		return 31
	}
	return day
}

// ultimoDiaDoMes devolve o último dia do mês informado.
func ultimoDiaDoMes(ano int, mes time.Month, loc *time.Location) int {
	return time.Date(ano, mes+1, 0, 0, 0, 0, 0, loc).Day()
}

// dataComDia monta uma data no mês informado usando o dia desejado. Quando o
// mês não possui esse dia (ex.: 31 em fevereiro), usa o último dia do mês.
func dataComDia(ano int, mes time.Month, dia int, loc *time.Location) time.Time {
	if dia < 1 {
		dia = 1
	}
	ultimo := ultimoDiaDoMes(ano, mes, loc)
	if dia > ultimo {
		dia = ultimo
	}
	return time.Date(ano, mes, dia, 0, 0, 0, 0, loc)
}

func nextDueDate(from time.Time, day int) string {
	day = normalizarDueDay(day)
	next := dataComDia(from.Year(), from.Month(), day, from.Location())
	if !next.After(from) {
		next = dataComDia(from.Year(), from.Month()+1, day, from.Location())
	}
	return next.Format("2006-01-02")
}

func (h *BillingRecurringHandler) GerarProximaFatura(invoiceID uint) {
	type invData struct {
		StudentPlanID, StudentID, PlanID uint
		AmountCents                      int
	}
	var cur invData
	err := h.db.Raw("SELECT student_plan_id, student_id, plan_id, amount_cents FROM invoices WHERE id = ?", invoiceID).Scan(&cur).Error
	if err != nil || cur.StudentPlanID == 0 {
		return
	}

	var dueDay int
	h.db.Raw("SELECT due_day FROM student_plans WHERE id = ?", cur.StudentPlanID).Scan(&dueDay)
	dueDay = normalizarDueDay(dueDay)

	// Planos avulsos (diária) não geram fatura recorrente.
	var durationDays int
	h.db.Raw("SELECT p.duration_days FROM plans p JOIN student_plans sp ON sp.plan_id = p.id WHERE sp.id = ?", cur.StudentPlanID).Scan(&durationDays)
	if durationDays <= 1 {
		return
	}

	nextDue := nextDueDate(time.Now(), dueDay)
	periodo := nextDue[:7]
	var count int64
	h.db.Model(&models.Invoice{}).
		Where("student_plan_id = ? AND status IN ('pending', 'overdue') AND substr(due_date, 1, 7) = ?", cur.StudentPlanID, periodo).
		Count(&count)
	if count > 0 {
		return
	}

	inv := models.Invoice{
		StudentPlanID: cur.StudentPlanID,
		StudentID:     cur.StudentID,
		PlanID:        cur.PlanID,
		AmountCents:   cur.AmountCents,
		Status:        "pending",
		DueDate:       nextDue,
	}
	if err := h.db.Create(&inv).Error; err == nil {
		h.session.RegistrarAudit("gerou_proxima_fatura", inv.ID)
	}
}

func (h *BillingRecurringHandler) ListarStudentPlans() ([]models.StudentPlan, error) {
	type spRaw struct {
		models.StudentPlan
		StudentName    string
		PlanName       string
		PlanPriceCents int
	}
	var raw []spRaw
	query := `SELECT sp.id, sp.student_id, sp.plan_id, sp.status, sp.payment_method_id,
		sp.due_day, sp.start_date, sp.end_date, sp.cancelled_at, sp.notes, sp.created_at, sp.updated_at,
		s.nome AS student_name, p.name AS plan_name, p.price_cents AS plan_price_cents
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		JOIN plans p ON p.id = sp.plan_id
		ORDER BY sp.start_date DESC`
	if err := h.db.Raw(query).Scan(&raw).Error; err != nil {
		return nil, err
	}
	list := make([]models.StudentPlan, len(raw))
	for i, r := range raw {
		list[i] = r.StudentPlan
		list[i].StudentName = r.StudentName
		list[i].PlanName = r.PlanName
		list[i].PlanPriceCents = r.PlanPriceCents
	}
	return list, nil
}

func (h *BillingRecurringHandler) CriarStudentPlan(studentID, planID uint, dueDay int) (*models.StudentPlan, error) {
	sp := models.StudentPlan{
		StudentID: studentID,
		PlanID:    planID,
		DueDay:    normalizarDueDay(dueDay),
	}
	if err := h.db.Create(&sp).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAudit("criou_plano_aluno", sp.ID)
	return h.BuscarStudentPlan(sp.ID)
}

func (h *BillingRecurringHandler) BuscarStudentPlan(id uint) (*models.StudentPlan, error) {
	type spRaw struct {
		models.StudentPlan
		StudentName    string
		PlanName       string
		PlanPriceCents int
	}
	var raw spRaw
	query := `SELECT sp.id, sp.student_id, sp.plan_id, sp.status, sp.payment_method_id,
		sp.due_day, sp.start_date, sp.end_date, sp.cancelled_at, sp.notes, sp.created_at, sp.updated_at,
		s.nome AS student_name, p.name AS plan_name, p.price_cents AS plan_price_cents
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		JOIN plans p ON p.id = sp.plan_id
		WHERE sp.id = ?`
	if err := h.db.Raw(query, id).Scan(&raw).Error; err != nil {
		return nil, err
	}
	raw.StudentPlan.StudentName = raw.StudentName
	raw.StudentPlan.PlanName = raw.PlanName
	raw.StudentPlan.PlanPriceCents = raw.PlanPriceCents
	return &raw.StudentPlan, nil
}

func (h *BillingRecurringHandler) CancelarStudentPlan(id uint) error {
	err := h.db.Model(&models.StudentPlan{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "cancelled",
		"cancelled_at": time.Now(),
	}).Error
	if err == nil {
		h.session.RegistrarAudit("cancelou_plano_aluno", id)
	}
	return err
}

func (h *BillingRecurringHandler) AtualizarPlanoAluno(studentID, planID, formaPagamentoID uint, dueDay int) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	dueDay = normalizarDueDay(dueDay)

	var current models.StudentPlan
	hasCurrent := h.db.Where("student_id = ? AND status = 'active'", studentID).First(&current).Error == nil

	if planID == 0 {
		if hasCurrent {
			return h.db.Model(&current).Updates(map[string]interface{}{"status": "cancelled", "cancelled_at": time.Now()}).Error
		}
		return nil
	}

	if hasCurrent && current.PlanID == planID {
		return h.db.Model(&current).Updates(map[string]interface{}{"payment_method_id": uintPtr(formaPagamentoID), "due_day": dueDay}).Error
	}

	if hasCurrent {
		h.db.Model(&current).Updates(map[string]interface{}{"status": "cancelled", "cancelled_at": time.Now()})
	}

	sp := models.StudentPlan{StudentID: studentID, PlanID: planID, DueDay: dueDay, PaymentMethodID: uintPtr(formaPagamentoID)}
	if err := h.db.Create(&sp).Error; err != nil {
		return err
	}
	h.session.RegistrarAuditComDetalhes("vinculou_plano_aluno", sp.ID, fmt.Sprintf("aluno=%d, plano=%d, forma_pagamento=%d, due_day=%d", studentID, planID, formaPagamentoID, dueDay))
	return nil
}
