package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"catraca-app/internal/billing/domain"
	"catraca-app/internal/billing/service"
)

type mockBillingRepo struct {
	plans        map[uint]*domain.Plan
	studentPlans map[uint]*domain.StudentPlan
	invoices     map[uint]*domain.Invoice
	nextInvID    uint
}

func newMockBillingRepo() *mockBillingRepo {
	return &mockBillingRepo{
		plans:        make(map[uint]*domain.Plan),
		studentPlans: make(map[uint]*domain.StudentPlan),
		invoices:     make(map[uint]*domain.Invoice),
		nextInvID:    1,
	}
}

func (m *mockBillingRepo) FindPlanByID(ctx context.Context, id uint) (*domain.Plan, error) {
	if p, ok := m.plans[id]; ok {
		return p, nil
	}
	return nil, errors.New("not found")
}

func (m *mockBillingRepo) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	var list []domain.Plan
	for _, p := range m.plans {
		list = append(list, *p)
	}
	return list, nil
}

func (m *mockBillingRepo) CreatePlan(ctx context.Context, p *domain.Plan) error {
	p.ID = uint(len(m.plans) + 1)
	m.plans[p.ID] = p
	return nil
}

func (m *mockBillingRepo) FindStudentPlan(ctx context.Context, id uint) (*domain.StudentPlan, error) {
	if sp, ok := m.studentPlans[id]; ok {
		return sp, nil
	}
	return nil, errors.New("not found")
}

func (m *mockBillingRepo) CreateInvoice(ctx context.Context, inv *domain.Invoice) error {
	inv.ID = m.nextInvID
	m.nextInvID++
	m.invoices[inv.ID] = inv
	return nil
}

func (m *mockBillingRepo) FindInvoiceByID(ctx context.Context, id uint) (*domain.Invoice, error) {
	if inv, ok := m.invoices[id]; ok {
		return inv, nil
	}
	return nil, errors.New("not found")
}

func (m *mockBillingRepo) UpdateInvoiceStatus(ctx context.Context, id uint, status string, paidAt *string, paidAmount *int) error {
	if inv, ok := m.invoices[id]; ok {
		inv.Status = status
		inv.PaidAmountCents = paidAmount
		return nil
	}
	return errors.New("not found")
}

func TestBillingService(t *testing.T) {
	repo := newMockBillingRepo()
	repo.plans[1] = &domain.Plan{
		ID:              1,
		Name:            "Mensal Standard",
		PriceCents:      15000,
		DurationDays:    30,
		GracePeriodDays: 5,
	}
	repo.studentPlans[1] = &domain.StudentPlan{
		ID:        1,
		StudentID: 10,
		PlanID:    1,
		Status:    "active",
		DueDay:    15,
	}
	repo.studentPlans[2] = &domain.StudentPlan{
		ID:        2,
		StudentID: 20,
		PlanID:    1,
		Status:    "cancelled",
		DueDay:    15,
	}

	svc := service.NewBillingService(repo)
	ctx := context.Background()

	// 1. Gerar fatura para matrícula ativa
	inv, err := svc.GerarFatura(ctx, 1)
	if err != nil {
		t.Fatalf("erro ao gerar fatura: %v", err)
	}
	if inv.AmountCents != 15000 {
		t.Errorf("esperado 15000 centavos, recebido %d", inv.AmountCents)
	}
	if inv.Status != "pending" {
		t.Errorf("status esperado 'pending', recebido %s", inv.Status)
	}

	// 2. Tentar gerar fatura para matrícula cancelada
	_, err = svc.GerarFatura(ctx, 2)
	if !errors.Is(err, domain.ErrStudentPlanInactive) {
		t.Errorf("esperado ErrStudentPlanInactive, recebido %v", err)
	}

	// 3. Confirmar pagamento
	err = svc.ConfirmarPagamento(ctx, inv.ID, 1, 15000)
	if err != nil {
		t.Fatalf("erro ao confirmar pagamento: %v", err)
	}
	if repo.invoices[inv.ID].Status != "paid" {
		t.Errorf("status da fatura esperado 'paid', recebido %s", repo.invoices[inv.ID].Status)
	}

	// 4. Teste de Inadimplência (Domain Rule)
	now := time.Now()
	overdueInv := &domain.Invoice{
		Status:  "pending",
		DueDate: now.AddDate(0, 0, -10).Format("2006-01-02"),
	}
	if !domain.IsOverdue(overdueInv, 5, now) {
		t.Errorf("fatura deveria ser considerada inadimplente")
	}
}
