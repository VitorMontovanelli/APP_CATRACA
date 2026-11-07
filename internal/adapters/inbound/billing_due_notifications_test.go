package inbound

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

var cpfTesteSeq int

func proximoCPFTeste() string {
	cpfTesteSeq++
	return fmt.Sprintf("%011d", cpfTesteSeq)
}

func setupDueNotificationsTest(t *testing.T) (*BillingDueNotificationsHandler, *BillingInvoicesHandler, *gorm.DB) {
	t.Helper()

	database := setupTestDB(t)
	if err := database.AutoMigrate(
		&models.Student{},
		&models.StudentPlan{},
		&models.Invoice{},
		&models.DueNotification{},
		&models.Setting{},
	); err != nil {
		t.Fatalf("failed to migrate notification models: %v", err)
	}

	session := NewSessionManager(database)
	actor := models.User{Name: "Admin Teste", Email: "admin-notif@teste", Senha: "x", Cargo: "super_admin", Ativo: true}
	if err := database.Create(&actor).Error; err != nil {
		t.Fatalf("criar ator: %v", err)
	}
	session.SetActor(actor.ID, "super_admin")

	recurring := NewBillingRecurringHandler(database, session)
	invoices := NewBillingInvoicesHandler(database, session, recurring)
	notifications := NewBillingDueNotificationsHandler(database, session)
	return notifications, invoices, database
}

func criarAlunoPlanoTeste(t *testing.T, database *gorm.DB, ativo bool) (*models.Student, *models.StudentPlan) {
	t.Helper()

	aluno := models.Student{Nome: "Vitor Aluno", CPF: proximoCPFTeste(), Ativo: ativo}
	if err := database.Create(&aluno).Error; err != nil {
		t.Fatalf("criar aluno: %v", err)
	}
	if !ativo {
		// O campo Ativo tem default:true no GORM; força a desativação para o teste.
		if err := database.Model(&models.Student{}).Where("id = ?", aluno.ID).Update("ativo", false).Error; err != nil {
			t.Fatalf("desativar aluno: %v", err)
		}
	}
	plano := models.Plan{Name: "Plano Teste " + proximoCPFTeste(), DurationDays: 30, PriceCents: 12000, Active: true}
	if err := database.Create(&plano).Error; err != nil {
		t.Fatalf("criar plano: %v", err)
	}
	sp := models.StudentPlan{StudentID: aluno.ID, PlanID: plano.ID, DueDay: 10, Status: "active"}
	if err := database.Create(&sp).Error; err != nil {
		t.Fatalf("criar plano do aluno: %v", err)
	}
	return &aluno, &sp
}

func criarFaturaTeste(t *testing.T, database *gorm.DB, dueDate, status string, ativo bool) uint {
	t.Helper()

	aluno, sp := criarAlunoPlanoTeste(t, database, ativo)
	inv := models.Invoice{
		StudentPlanID: sp.ID,
		StudentID:     aluno.ID,
		PlanID:        sp.PlanID,
		AmountCents:   12000,
		Status:        status,
		DueDate:       dueDate,
	}
	if err := database.Create(&inv).Error; err != nil {
		t.Fatalf("criar fatura: %v", err)
	}
	return inv.ID
}

func TestNextDueDateRespeitaUltimoDiaDoMes(t *testing.T) {
	loc := time.Local
	casos := []struct {
		from time.Time
		day  int
		want string
	}{
		{time.Date(2026, 1, 15, 0, 0, 0, 0, loc), 31, "2026-01-31"},
		{time.Date(2026, 1, 31, 0, 0, 0, 0, loc), 31, "2026-02-28"},
		{time.Date(2026, 2, 1, 0, 0, 0, 0, loc), 31, "2026-02-28"},
		{time.Date(2026, 4, 1, 0, 0, 0, 0, loc), 31, "2026-04-30"},
		{time.Date(2026, 1, 15, 0, 0, 0, 0, loc), 5, "2026-02-05"},
		{time.Date(2024, 2, 1, 0, 0, 0, 0, loc), 31, "2024-02-29"},
	}
	for _, c := range casos {
		if got := nextDueDate(c.from, c.day); got != c.want {
			t.Errorf("nextDueDate(%s, %d) = %s; want %s", c.from.Format("2006-01-02"), c.day, got, c.want)
		}
	}
}

func TestAjustarParaDiaUtil(t *testing.T) {
	sexta := time.Date(2026, 11, 6, 0, 0, 0, 0, time.Local)
	sabado := time.Date(2026, 11, 7, 0, 0, 0, 0, time.Local)
	domingo := time.Date(2026, 11, 8, 0, 0, 0, 0, time.Local)
	segunda := time.Date(2026, 11, 9, 0, 0, 0, 0, time.Local)

	casos := []struct {
		entrada time.Time
		want    time.Time
	}{
		{sexta, sexta},
		{sabado, sexta},
		{domingo, sexta},
		{segunda, segunda},
	}
	for _, c := range casos {
		if got := ajustarParaDiaUtil(c.entrada); !got.Equal(c.want) {
			t.Errorf("ajustarParaDiaUtil(%s) = %s; want %s",
				c.entrada.Format("2006-01-02"), got.Format("2006-01-02"), c.want.Format("2006-01-02"))
		}
	}
}

func TestFormatarReais(t *testing.T) {
	casos := map[int]string{
		99:      "R$ 0,99",
		100:     "R$ 1,00",
		12000:   "R$ 120,00",
		123456:  "R$ 1.234,56",
		1000000: "R$ 10.000,00",
	}
	for cents, want := range casos {
		if got := formatarReais(cents); got != want {
			t.Errorf("formatarReais(%d) = %q; want %q", cents, got, want)
		}
	}
}

func TestVerificarAvisosCriaUmaVezEMarcaComoLido(t *testing.T) {
	h, _, database := setupDueNotificationsTest(t)
	due := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	criarFaturaTeste(t, database, due, "pending", true)

	lista, err := h.VerificarAvisosVencimento()
	if err != nil {
		t.Fatalf("verificar avisos: %v", err)
	}
	if len(lista) != 1 {
		t.Fatalf("esperava 1 aviso, obteve %d", len(lista))
	}
	aviso := lista[0]
	if aviso.Kind != "due" {
		t.Errorf("kind = %q; want due", aviso.Kind)
	}
	if !strings.Contains(aviso.Message, "vence amanhã") {
		t.Errorf("mensagem sem 'vence amanhã': %q", aviso.Message)
	}
	if !strings.Contains(aviso.Message, "R$ 120,00") {
		t.Errorf("mensagem sem valor formatado: %q", aviso.Message)
	}

	if err := h.MarcarAvisoComoLido(aviso.ID); err != nil {
		t.Fatalf("marcar como lido: %v", err)
	}
	lista, err = h.VerificarAvisosVencimento()
	if err != nil {
		t.Fatalf("verificar avisos novamente: %v", err)
	}
	if len(lista) != 1 {
		t.Fatalf("não deveria duplicar aviso; obteve %d", len(lista))
	}
	if lista[0].ReadAt == nil {
		t.Error("aviso deveria permanecer marcado como lido")
	}
}

func TestFaturasPagasCanceladasEAlunosInativosNaoGeramAviso(t *testing.T) {
	h, _, database := setupDueNotificationsTest(t)
	due := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	criarFaturaTeste(t, database, due, "pending", true)
	criarFaturaTeste(t, database, due, "paid", true)
	criarFaturaTeste(t, database, due, "cancelled", true)
	criarFaturaTeste(t, database, due, "pending", false)

	lista, err := h.VerificarAvisosVencimento()
	if err != nil {
		t.Fatalf("verificar avisos: %v", err)
	}
	if len(lista) != 1 {
		t.Fatalf("esperava somente o aviso da fatura pendente de aluno ativo; obteve %d", len(lista))
	}
}

func TestAvisoDeAtrasoNaoRepeteESubstituiAvisoDeVencimento(t *testing.T) {
	h, _, database := setupDueNotificationsTest(t)
	amanha := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	ontem := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	invoiceID := criarFaturaTeste(t, database, amanha, "pending", true)

	if _, err := h.VerificarAvisosVencimento(); err != nil {
		t.Fatalf("verificar aviso de vencimento: %v", err)
	}

	if err := database.Model(&models.Invoice{}).Where("id = ?", invoiceID).Update("due_date", ontem).Error; err != nil {
		t.Fatalf("atualizar vencimento: %v", err)
	}

	lista, err := h.VerificarAvisosVencimento()
	if err != nil {
		t.Fatalf("verificar aviso de atraso: %v", err)
	}
	if len(lista) != 1 {
		t.Fatalf("esperava apenas o aviso de atraso, obteve %d", len(lista))
	}
	if lista[0].Kind != "overdue" {
		t.Errorf("kind = %q; want overdue", lista[0].Kind)
	}
	if !strings.Contains(lista[0].Message, "venceu há 1 dia") {
		t.Errorf("mensagem de atraso inesperada: %q", lista[0].Message)
	}

	lista, err = h.VerificarAvisosVencimento()
	if err != nil {
		t.Fatalf("verificar aviso de atraso novamente: %v", err)
	}
	if len(lista) != 1 {
		t.Fatalf("aviso de atraso não deveria repetir; obteve %d", len(lista))
	}
}

func TestGerarInvoiceUsaDiaDeVencimentoENaoDuplicaPeriodo(t *testing.T) {
	_, invoices, database := setupDueNotificationsTest(t)
	_, sp := criarAlunoPlanoTeste(t, database, true)
	if err := database.Model(&models.StudentPlan{}).Where("id = ?", sp.ID).Update("due_day", 10).Error; err != nil {
		t.Fatalf("ajustar dia de vencimento: %v", err)
	}

	inv, err := invoices.GerarInvoice(sp.ID)
	if err != nil {
		t.Fatalf("gerar fatura: %v", err)
	}
	want := nextDueDate(time.Now(), 10)
	if inv.DueDate != want {
		t.Errorf("due_date = %s; want %s", inv.DueDate, want)
	}

	if _, err := invoices.GerarInvoice(sp.ID); err == nil {
		t.Fatal("esperava erro ao gerar outra fatura em aberto no mesmo período")
	} else if !strings.Contains(err.Error(), "já existe uma fatura em aberto") {
		t.Errorf("erro inesperado: %v", err)
	}
}

func TestConfigAvisoVencimentoPadraoEPersistencia(t *testing.T) {
	h, _, database := setupDueNotificationsTest(t)

	cfg, err := h.ObterConfigAvisoVencimento()
	if err != nil {
		t.Fatalf("obter config: %v", err)
	}
	if cfg.DiasAntecedencia != 1 {
		t.Errorf("padrão = %d; want 1", cfg.DiasAntecedencia)
	}

	if err := h.SalvarConfigAvisoVencimento(5); err != nil {
		t.Fatalf("salvar config: %v", err)
	}
	cfg, err = h.ObterConfigAvisoVencimento()
	if err != nil {
		t.Fatalf("obter config salva: %v", err)
	}
	if cfg.DiasAntecedencia != 5 {
		t.Errorf("config salva = %d; want 5", cfg.DiasAntecedencia)
	}

	var valor string
	if err := database.Model(&models.Setting{}).Where("key = ?", chaveAvisoVencimentoDias).Pluck("value", &valor).Error; err != nil {
		t.Fatalf("ler setting: %v", err)
	}
	if valor != "5" {
		t.Errorf("setting = %q; want 5", valor)
	}

	if err := h.SalvarConfigAvisoVencimento(31); err == nil {
		t.Fatal("esperava erro para antecedência acima de 30")
	}
}
