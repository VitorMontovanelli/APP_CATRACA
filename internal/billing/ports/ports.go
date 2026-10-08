package ports

import (
	"context"

	"catraca-app/internal/billing/domain"
)

type BillingRepository interface {
	FindPlanByID(ctx context.Context, id uint) (*domain.Plan, error)
	ListPlans(ctx context.Context) ([]domain.Plan, error)
	CreatePlan(ctx context.Context, p *domain.Plan) error
	FindStudentPlan(ctx context.Context, id uint) (*domain.StudentPlan, error)
	CreateInvoice(ctx context.Context, inv *domain.Invoice) error
	FindInvoiceByID(ctx context.Context, id uint) (*domain.Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, id uint, status string, paidAt *string, paidAmount *int) error
}

type BillingServicePort interface {
	CriarPlano(ctx context.Context, p *domain.Plan) error
	GerarFatura(ctx context.Context, studentPlanID uint) (*domain.Invoice, error)
	ConfirmarPagamento(ctx context.Context, invoiceID, paymentMethodID uint, paidAmountCents int) error
}
