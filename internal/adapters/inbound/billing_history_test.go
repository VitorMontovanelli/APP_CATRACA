package inbound

import (
	"context"
	"testing"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

func setupBillingHistoryTest(t *testing.T) (*BillingInvoicesHandler, *BillingPlansHandler, *gorm.DB) {
	t.Helper()

	database := setupTestDB(t)
	if err := database.AutoMigrate(
		&models.Student{},
		&models.StudentPlan{},
		&models.Invoice{},
	); err != nil {
		t.Fatalf("failed to migrate billing models: %v", err)
	}

	session := NewSessionManager(database)
	actor := models.User{Name: "Maria Admin", Email: "maria@teste", Senha: "x", Cargo: "super_admin", Ativo: true}
	if err := database.Create(&actor).Error; err != nil {
		t.Fatalf("criar ator: %v", err)
	}
	session.SetActor(actor.ID, "super_admin")

	recurring := NewBillingRecurringHandler(database, session)
	invoices := NewBillingInvoicesHandler(database, session, recurring)
	invoices.SetContext(context.Background())
	plans := NewBillingPlansHandler(database, session)
	plans.SetContext(context.Background())
	return invoices, plans, database
}

func contemFatura(list []models.Invoice, id uint) bool {
	for _, inv := range list {
		if inv.ID == id {
			return true
		}
	}
	return false
}

func TestConfirmarPagamentoMoveParaHistoricoEEstornoVoltaParaPendente(t *testing.T) {
	invoices, _, database := setupBillingHistoryTest(t)

	plan := models.Plan{Name: "Plano Teste", DurationDays: 30, PriceCents: 3500, GracePeriodDays: 5, Active: true}
	if err := database.Create(&plan).Error; err != nil {
		t.Fatalf("criar plano: %v", err)
	}
	student := models.Student{Nome: "Aluno Teste", Ativo: true}
	if err := database.Create(&student).Error; err != nil {
		t.Fatalf("criar aluno: %v", err)
	}
	sp := models.StudentPlan{StudentID: student.ID, PlanID: plan.ID, DueDay: 5, Status: "active", StartDate: time.Now()}
	if err := database.Create(&sp).Error; err != nil {
		t.Fatalf("criar plano do aluno: %v", err)
	}
	metodo := models.PaymentMethod{Name: "Pix", Type: "pix", Enabled: true}
	if err := database.Create(&metodo).Error; err != nil {
		t.Fatalf("criar forma de pagamento: %v", err)
	}

	inv := models.Invoice{
		StudentPlanID: sp.ID,
		StudentID:     student.ID,
		PlanID:        plan.ID,
		AmountCents:   3500,
		Status:        "pending",
		DueDate:       time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}
	if err := database.Create(&inv).Error; err != nil {
		t.Fatalf("criar fatura: %v", err)
	}

	// Pagamento com desconto: a fatura guarda 3500 e o valor pago 3000.
	if err := invoices.ConfirmarPagamento(inv.ID, metodo.ID, 3000, "2026-11-05"); err != nil {
		t.Fatalf("ConfirmarPagamento: %v", err)
	}

	var paga models.Invoice
	if err := database.First(&paga, inv.ID).Error; err != nil {
		t.Fatalf("buscar fatura paga: %v", err)
	}
	if paga.Status != "paid" || paga.AmountCents != 3500 || paga.PaidAmountCents == nil || *paga.PaidAmountCents != 3000 {
		t.Fatalf("fatura paga inesperada: status=%s amount=%d paid=%v", paga.Status, paga.AmountCents, paga.PaidAmountCents)
	}
	if paga.ConfirmedBy != "Maria Admin" {
		t.Fatalf("confirmado por = %q, esperado Maria Admin", paga.ConfirmedBy)
	}

	historico, err := invoices.ListarHistoricoPagamentos()
	if err != nil {
		t.Fatalf("ListarHistoricoPagamentos: %v", err)
	}
	if !contemFatura(historico, inv.ID) {
		t.Fatal("fatura paga não apareceu no histórico")
	}
	pendentes, err := invoices.ListarInvoicesPendentes()
	if err != nil {
		t.Fatalf("ListarInvoicesPendentes: %v", err)
	}
	if contemFatura(pendentes, inv.ID) {
		t.Fatal("fatura paga continuou em Faturas Emitidas")
	}

	if err := invoices.DeletarInvoice(inv.ID); err == nil {
		t.Fatal("fatura paga deveria ser protegida contra exclusão")
	}

	if err := invoices.EstornarPagamento(inv.ID); err != nil {
		t.Fatalf("EstornarPagamento: %v", err)
	}
	var estornada models.Invoice
	if err := database.First(&estornada, inv.ID).Error; err != nil {
		t.Fatalf("buscar fatura estornada: %v", err)
	}
	if estornada.Status != "pending" || estornada.PaidAmountCents != nil || estornada.PaidAt != nil {
		t.Fatalf("fatura estornada inesperada: status=%s paid=%v paid_at=%v", estornada.Status, estornada.PaidAmountCents, estornada.PaidAt)
	}
	if estornada.RefundedBy != "Maria Admin" || estornada.RefundedAt == nil {
		t.Fatalf("estorno não registrou quem/quando: by=%q at=%v", estornada.RefundedBy, estornada.RefundedAt)
	}

	pendentes, err = invoices.ListarInvoicesPendentes()
	if err != nil {
		t.Fatalf("ListarInvoicesPendentes após estorno: %v", err)
	}
	if !contemFatura(pendentes, inv.ID) {
		t.Fatal("fatura estornada não voltou para Faturas Emitidas")
	}
	historico, err = invoices.ListarHistoricoPagamentos()
	if err != nil {
		t.Fatalf("ListarHistoricoPagamentos após estorno: %v", err)
	}
	if contemFatura(historico, inv.ID) {
		t.Fatal("fatura estornada continuou no histórico")
	}
}

func TestAtualizarPlanoReprecificaSomenteFaturasEmAberto(t *testing.T) {
	_, plans, database := setupBillingHistoryTest(t)

	plan := models.Plan{Name: "Plano Teste", DurationDays: 30, PriceCents: 3500, GracePeriodDays: 5, Active: true}
	if err := database.Create(&plan).Error; err != nil {
		t.Fatalf("criar plano: %v", err)
	}
	student := models.Student{Nome: "Aluno Teste", Ativo: true}
	if err := database.Create(&student).Error; err != nil {
		t.Fatalf("criar aluno: %v", err)
	}
	sp := models.StudentPlan{StudentID: student.ID, PlanID: plan.ID, DueDay: 5, Status: "active", StartDate: time.Now()}
	if err := database.Create(&sp).Error; err != nil {
		t.Fatalf("criar plano do aluno: %v", err)
	}

	pendente := models.Invoice{StudentPlanID: sp.ID, StudentID: student.ID, PlanID: plan.ID, AmountCents: 0, Status: "pending", DueDate: "2026-12-05"}
	paga := models.Invoice{StudentPlanID: sp.ID, StudentID: student.ID, PlanID: plan.ID, AmountCents: 0, Status: "paid", DueDate: "2026-11-05"}
	if err := database.Create(&pendente).Error; err != nil {
		t.Fatalf("criar pendente: %v", err)
	}
	if err := database.Create(&paga).Error; err != nil {
		t.Fatalf("criar paga: %v", err)
	}

	if _, err := plans.AtualizarPlano(plan.ID, plan.Name, plan.Description, plan.DurationDays, 4000, plan.GracePeriodDays, nil); err != nil {
		t.Fatalf("AtualizarPlano: %v", err)
	}

	var pendenteAtualizada, pagaAtualizada models.Invoice
	database.First(&pendenteAtualizada, pendente.ID)
	database.First(&pagaAtualizada, paga.ID)
	if pendenteAtualizada.AmountCents != 4000 {
		t.Fatalf("fatura pendente = %d, esperado 4000", pendenteAtualizada.AmountCents)
	}
	if pagaAtualizada.AmountCents != 0 {
		t.Fatalf("fatura paga foi alterada para %d", pagaAtualizada.AmountCents)
	}
}
