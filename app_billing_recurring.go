package main

import "time"

func (a *App) gerarProximaFatura(invoiceID uint) {
	type invData struct {
		StudentPlanID uint
		StudentID     uint
		PlanID        uint
		AmountCents   int
	}
	var cur invData
	err := a.db.Raw("SELECT student_plan_id, student_id, plan_id, amount_cents FROM invoices WHERE id = ?", invoiceID).Scan(&cur).Error
	if err != nil || cur.StudentPlanID == 0 {
		return
	}

	var dueDay int
	a.db.Raw("SELECT due_day FROM student_plans WHERE id = ?", cur.StudentPlanID).Scan(&dueDay)
	if dueDay == 0 {
		dueDay = 5
	}

	nextDue := nextDueDate(time.Now(), dueDay)

	var count int64
	a.db.Model(&Invoice{}).Where("student_plan_id = ? AND due_date = ?", cur.StudentPlanID, nextDue).Count(&count)
	if count > 0 {
		return
	}

	inv := Invoice{
		StudentPlanID: cur.StudentPlanID,
		StudentID:     cur.StudentID,
		PlanID:        cur.PlanID,
		AmountCents:   cur.AmountCents,
		Status:        "pending",
		DueDate:       nextDue,
	}
	if err := a.db.Create(&inv).Error; err == nil {
		a.registrarAudit("gerou_proxima_fatura", inv.ID)
	}
}

func nextDueDate(from time.Time, day int) string {
	next := time.Date(from.Year(), from.Month(), day, 0, 0, 0, 0, from.Location())
	if next.Before(from) || next.Equal(from) {
		next = next.AddDate(0, 1, 0)
	}
	return next.Format("2006-01-02")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func uintPtr(n uint) *uint {
	if n == 0 {
		return nil
	}
	return &n
}

type PlanoContador struct {
	Nome string `json:"nome"`
	Qtd  int64  `json:"qtd"`
}

type MetricasHome struct {
	TotalUsuarios      int64           `json:"total_usuarios"`
	TotalAdmins        int64           `json:"total_admins"`
	TotalAlunos        int64           `json:"total_alunos"`
	TotalFaturadoCents int64           `json:"total_faturado_cents"`
	TotalAtrasoCents   int64           `json:"total_atraso_cents"`
	TotalPendenteCents int64           `json:"total_pendente_cents"`
	PlanosDistribuicao []PlanoContador `json:"planos_distribuicao"`
}
