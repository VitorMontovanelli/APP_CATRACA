package inbound

import (
	"context"
	"testing"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

func setupPaymentMethodsTest(t *testing.T) (*BillingPlansHandler, *BillingInvoicesHandler, *SessionManager, *gorm.DB) {
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
	actor := models.User{Name: "Admin Comum", Email: "admin@teste", Senha: "x", Cargo: "admin", Ativo: true}
	if err := database.Create(&actor).Error; err != nil {
		t.Fatalf("criar ator: %v", err)
	}
	session.SetActor(actor.ID, "admin")

	recurring := NewBillingRecurringHandler(database, session)
	invoices := NewBillingInvoicesHandler(database, session, recurring)
	invoices.SetContext(context.Background())
	plans := NewBillingPlansHandler(database, session)
	plans.SetContext(context.Background())
	return plans, invoices, session, database
}

func TestAdminComumGerenciaMetodosPagamento(t *testing.T) {
	plans, _, _, _ := setupPaymentMethodsTest(t)

	metodo, err := plans.CriarPaymentMethod("Pix", "pix", 1.5, 0)
	if err != nil {
		t.Fatalf("admin comum deveria criar método: %v", err)
	}

	if _, err := plans.AtualizarPaymentMethod(metodo.ID, "Pix novo", "transfer", 0, 250); err != nil {
		t.Fatalf("admin comum deveria editar método: %v", err)
	}
	if err := plans.AtivarPaymentMethod(metodo.ID, false); err != nil {
		t.Fatalf("admin comum deveria desativar método: %v", err)
	}
	if err := plans.DeletarPaymentMethod(metodo.ID); err != nil {
		t.Fatalf("método sem uso deveria ser excluído: %v", err)
	}

	if _, err := plans.CriarPaymentMethod("Inválido", "tipo_inexistente", 0, 0); err == nil {
		t.Fatal("tipo inválido deveria ser recusado")
	}
	if _, err := plans.CriarPaymentMethod("", "pix", 0, 0); err == nil {
		t.Fatal("nome vazio deveria ser recusado")
	}
	if _, err := plans.CriarPaymentMethod("Taxa dupla", "pix", 1, 100); err == nil {
		t.Fatal("percentual e valor fixo juntos deveriam ser recusados")
	}
	if _, err := plans.CriarPaymentMethod("Taxa negativa", "pix", -1, 0); err == nil {
		t.Fatal("taxa negativa deveria ser recusada")
	}
}

func TestMetodoUsadoEmFaturaPagaGuardaHistoricoENaoPodeSerExcluido(t *testing.T) {
	plans, invoices, _, database := setupPaymentMethodsTest(t)

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
	metodo, err := plans.CriarPaymentMethod("Pix antigo", "pix", 1.5, 0)
	if err != nil {
		t.Fatalf("criar método: %v", err)
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

	if err := invoices.ConfirmarPagamento(inv.ID, metodo.ID, 3500, "2026-11-05"); err != nil {
		t.Fatalf("ConfirmarPagamento: %v", err)
	}

	var paga models.Invoice
	database.First(&paga, inv.ID)
	if paga.PaidMethodName != "Pix antigo" || paga.PaidFeePercent != 1.5 || paga.PaidFeeFixedCents != 0 {
		t.Fatalf("snapshot incorreto: nome=%q percentual=%v fixo=%d", paga.PaidMethodName, paga.PaidFeePercent, paga.PaidFeeFixedCents)
	}

	// Renomear e alterar a taxa do método não pode mudar o histórico da fatura paga.
	if _, err := plans.AtualizarPaymentMethod(metodo.ID, "Pix renomeado", "pix", 9.9, 0); err != nil {
		t.Fatalf("editar método: %v", err)
	}
	historico, err := invoices.ListarHistoricoPagamentos()
	if err != nil {
		t.Fatalf("ListarHistoricoPagamentos: %v", err)
	}
	var encontrada *models.Invoice
	for i := range historico {
		if historico[i].ID == inv.ID {
			encontrada = &historico[i]
			break
		}
	}
	if encontrada == nil {
		t.Fatal("fatura paga não apareceu no histórico")
	}
	if encontrada.PaymentMethodName != "Pix antigo" || encontrada.PaidFeePercent != 1.5 {
		t.Fatalf("histórico mudou após renomear método: nome=%q percentual=%v", encontrada.PaymentMethodName, encontrada.PaidFeePercent)
	}

	if err := plans.DeletarPaymentMethod(metodo.ID); err == nil {
		t.Fatal("método usado em fatura paga não deveria ser excluído")
	}
	if err := plans.AtivarPaymentMethod(metodo.ID, false); err != nil {
		t.Fatalf("método em uso deveria poder ser desativado: %v", err)
	}
}

func TestMetodoUsadoEmPlanoDeAlunoNaoPodeSerExcluido(t *testing.T) {
	plans, _, _, database := setupPaymentMethodsTest(t)

	metodo, err := plans.CriarPaymentMethod("Dinheiro", "cash", 0, 0)
	if err != nil {
		t.Fatalf("criar método: %v", err)
	}
	plan := models.Plan{Name: "Plano Teste", DurationDays: 30, PriceCents: 3500, Active: true}
	if err := database.Create(&plan).Error; err != nil {
		t.Fatalf("criar plano: %v", err)
	}
	student := models.Student{Nome: "Aluno Teste", Ativo: true}
	if err := database.Create(&student).Error; err != nil {
		t.Fatalf("criar aluno: %v", err)
	}
	sp := models.StudentPlan{StudentID: student.ID, PlanID: plan.ID, DueDay: 5, Status: "active", StartDate: time.Now(), PaymentMethodID: &metodo.ID}
	if err := database.Create(&sp).Error; err != nil {
		t.Fatalf("criar plano do aluno: %v", err)
	}

	if err := plans.DeletarPaymentMethod(metodo.ID); err == nil {
		t.Fatal("método usado em plano de aluno não deveria ser excluído")
	}
}

func TestConfirmarPagamentoComMetodoDesativadoFalha(t *testing.T) {
	plans, invoices, _, database := setupPaymentMethodsTest(t)

	metodo, err := plans.CriarPaymentMethod("Boleto", "boleto", 0, 349)
	if err != nil {
		t.Fatalf("criar método: %v", err)
	}
	if err := plans.AtivarPaymentMethod(metodo.ID, false); err != nil {
		t.Fatalf("desativar método: %v", err)
	}
	plan := models.Plan{Name: "Plano Teste", DurationDays: 30, PriceCents: 3500, Active: true}
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
	inv := models.Invoice{StudentPlanID: sp.ID, StudentID: student.ID, PlanID: plan.ID, AmountCents: 3500, Status: "pending", DueDate: "2026-12-05"}
	if err := database.Create(&inv).Error; err != nil {
		t.Fatalf("criar fatura: %v", err)
	}

	if err := invoices.ConfirmarPagamento(inv.ID, metodo.ID, 3500, "2026-11-05"); err == nil {
		t.Fatal("não deveria confirmar pagamento com método desativado")
	}
}
