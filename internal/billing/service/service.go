package service

import (
	"context"
	"fmt"
	"time"

	"catraca-app/internal/billing/domain"
	"catraca-app/internal/billing/ports"
)

type BillingService struct {
	repo ports.BillingRepository
}

func NewBillingService(repo ports.BillingRepository) *BillingService {
	return &BillingService{repo: repo}
}

func (s *BillingService) CriarPlano(ctx context.Context, p *domain.Plan) error {
	if err := p.Validate(); err != nil {
		return err
	}
	p.Active = true
	return s.repo.CreatePlan(ctx, p)
}

func (s *BillingService) GerarFatura(ctx context.Context, studentPlanID uint) (*domain.Invoice, error) {
	sp, err := s.repo.FindStudentPlan(ctx, studentPlanID)
	if err != nil {
		return nil, err
	}
	if sp.Status == "cancelled" || sp.Status == "expired" {
		return nil, domain.ErrStudentPlanInactive
	}

	plan, err := s.repo.FindPlanByID(ctx, sp.PlanID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	dueYear := now.Year()
	dueMonth := now.Month()
	if sp.DueDay < now.Day() {
		// Vencimento para o próximo mês
		dueMonth++
		if dueMonth > 12 {
			dueMonth = 1
			dueYear++
		}
	}
	dueDate := fmt.Sprintf("%04d-%02d-%02d", dueYear, dueMonth, sp.DueDay)

	inv := &domain.Invoice{
		StudentPlanID: sp.ID,
		StudentID:     sp.StudentID,
		PlanID:        sp.PlanID,
		AmountCents:   plan.PriceCents,
		Status:        "pending",
		DueDate:       dueDate,
	}

	if err := s.repo.CreateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *BillingService) ConfirmarPagamento(ctx context.Context, invoiceID, paymentMethodID uint, paidAmountCents int) error {
	inv, err := s.repo.FindInvoiceByID(ctx, invoiceID)
	if err != nil {
		return err
	}
	if inv.Status == "paid" {
		return domain.ErrInvoiceAlreadyPaid
	}

	nowStr := time.Now().Format(time.RFC3339)
	return s.repo.UpdateInvoiceStatus(ctx, invoiceID, "paid", &nowStr, &paidAmountCents)
}
