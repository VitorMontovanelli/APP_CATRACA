package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidPlanPrice    = errors.New("o preço do plano não pode ser negativo")
	ErrInvalidDueDay       = errors.New("o dia de vencimento deve estar entre 1 e 31")
	ErrInvoiceAlreadyPaid  = errors.New("fatura já se encontra paga")
	ErrStudentPlanInactive = errors.New("matrícula de plano do aluno está inativa")
)

type Plan struct {
	ID               uint
	Name             string
	Description      string
	DurationDays     int
	PriceCents       int
	PrecoCartaoCents *int
	GracePeriodDays  int
	Active           bool
}

type StudentPlan struct {
	ID              uint
	StudentID       uint
	PlanID          uint
	Status          string // active, overdue, suspended, cancelled, expired
	PaymentMethodID *uint
	DueDay          int
	StartDate       time.Time
	EndDate         *time.Time
}

type Invoice struct {
	ID              uint
	StudentPlanID   uint
	StudentID       uint
	PlanID          uint
	AmountCents     int
	Status          string // pending, paid, overdue, cancelled, refunded
	PaymentMethodID *uint
	DueDate         string // YYYY-MM-DD
	PaidAt          *time.Time
	PaidAmountCents *int
}

// IsOverdue verifica se uma fatura pendente ultrapassou a tolerância (grace period).
func IsOverdue(invoice *Invoice, gracePeriodDays int, now time.Time) bool {
	if invoice.Status != "pending" {
		return false
	}
	parsedDueDate, err := time.Parse("2006-01-02", invoice.DueDate)
	if err != nil {
		return false
	}
	limit := parsedDueDate.AddDate(0, 0, gracePeriodDays)
	return now.After(limit)
}

func (p *Plan) Validate() error {
	if p.PriceCents < 0 {
		return ErrInvalidPlanPrice
	}
	if p.DurationDays <= 0 {
		p.DurationDays = 30
	}
	return nil
}
