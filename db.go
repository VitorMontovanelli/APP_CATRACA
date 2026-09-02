package main

import (
	"log"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("catraca.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Falha ao abrir banco de dados: %v", err)
	}

	// Configura pragmas
	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA journal_mode = WAL")

	// AutoMigrate: cria/atualiza todas as tabelas
	err = db.AutoMigrate(
		&User{},
		&Aluno{},
		&RegistroAcesso{},
		&Student{},
		&Plan{},
		&PaymentMethod{},
		&PaymentGateway{},
		&StudentPlan{},
		&Invoice{},
		&AccessLog{},
		&AuditLog{},
		&Setting{},
		&Agendamento{},
		&CapacidadeDia{},
		&AgendamentoIndividual{},
	)
	if err != nil {
		log.Fatalf("Falha ao migrar banco de dados: %v", err)
	}

	// Migration manual: cria tabelas que podem não ter sido criadas pelo AutoMigrate
	db.Exec("CREATE TABLE IF NOT EXISTS capacidade_dias (data TEXT PRIMARY KEY, capacidade INTEGER NOT NULL)")
	db.Exec("CREATE TABLE IF NOT EXISTS agendamento_individuals (id INTEGER PRIMARY KEY AUTOINCREMENT, data TEXT, hora TEXT, student_id INTEGER, nome TEXT, pratica TEXT, observacao TEXT, created_at DATETIME)")
	db.Exec("CREATE TABLE IF NOT EXISTS settings (`key` TEXT PRIMARY KEY, value TEXT)")

	// Migração manual: adiciona colunas que podem não ter sido criadas pelo AutoMigrate
	if !db.Migrator().HasColumn(&User{}, "Foto") {
		db.Migrator().AddColumn(&User{}, "Foto")
	}
	if !db.Migrator().HasColumn(&Invoice{}, "Comprovante") {
		db.Migrator().AddColumn(&Invoice{}, "Comprovante")
	}
	if !db.Migrator().HasColumn(&Invoice{}, "PaidAt") {
		db.Migrator().AddColumn(&Invoice{}, "PaidAt")
	}
	if !db.Migrator().HasColumn(&Invoice{}, "PaidAmountCents") {
		db.Migrator().AddColumn(&Invoice{}, "PaidAmountCents")
	}
	if !db.Migrator().HasColumn(&Plan{}, "PrecoCartaoCents") {
		db.Migrator().AddColumn(&Plan{}, "PrecoCartaoCents")
	}
	if !db.Migrator().HasColumn(&Student{}, "TelefoneUrgencia") {
		db.Migrator().AddColumn(&Student{}, "TelefoneUrgencia")
	}
	if !db.Migrator().HasColumn(&Student{}, "LaudoMedico") {
		db.Migrator().AddColumn(&Student{}, "LaudoMedico")
	}
	if !db.Migrator().HasColumn(&Student{}, "Foto") {
		db.Migrator().AddColumn(&Student{}, "Foto")
	}

	seedData(db)
	migrateDefaultPlans(db)

	return db
}

func seedData(db *gorm.DB) {
	seedSuperAdmin(db)
	seedPaymentData(db)
}

func intPtr(v int) *int {
	return &v
}

func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Falha ao hashear senha de seed: %v", err)
	}
	return string(hashed)
}

func seedSuperAdmin(db *gorm.DB) {
	var count int64
	db.Model(&User{}).Where("cargo = ?", "super_admin").Count(&count)
	if count > 0 {
		return
	}
	db.Create(&User{
		Name:  "Admin Supremo",
		Email: "admin@catraca.com",
		Senha: hashPassword("LLI81o9wtVn$cD%CsmuLZvhxL"),
		Cargo: "super_admin",
	})
}

func seedPaymentData(db *gorm.DB) {
	var count int64
	db.Model(&Plan{}).Count(&count)
	if count > 0 {
		return
	}

	plans := []Plan{
		{
			Name:             "Avulso (Diária)",
			Description:      "Aula avulsa de pilates, acesso por um único dia.",
			DurationDays:     1,
			PriceCents:       3500,
			PrecoCartaoCents: intPtr(3500),
			GracePeriodDays:  0,
		},
		{
			Name:             "Plano 2x por semana",
			Description:      "Duas aulas semanais de pilates. Valor no pix/dinheiro: R$ 220,00 · Cartão: R$ 230,00.",
			DurationDays:     30,
			PriceCents:       22000,
			PrecoCartaoCents: intPtr(23000),
			GracePeriodDays:  5,
		},
		{
			Name:             "Plano 3x por semana",
			Description:      "Três aulas semanais de pilates. Valor no pix/dinheiro: R$ 280,00 · Cartão: R$ 290,00.",
			DurationDays:     30,
			PriceCents:       28000,
			PrecoCartaoCents: intPtr(29000),
			GracePeriodDays:  5,
		},
		{
			Name:             "Plano 4x por semana",
			Description:      "Quatro aulas semanais de pilates. Valor no pix/dinheiro: R$ 330,00 · Cartão: R$ 340,00.",
			DurationDays:     30,
			PriceCents:       33000,
			PrecoCartaoCents: intPtr(34000),
			GracePeriodDays:  5,
		},
		{
			Name:             "Plano 5x por semana",
			Description:      "Cinco aulas semanais de pilates. Valor no pix/dinheiro: R$ 430,00 · Cartão: R$ 440,00.",
			DurationDays:     30,
			PriceCents:       43000,
			PrecoCartaoCents: intPtr(44000),
			GracePeriodDays:  5,
		},
	}
	for _, p := range plans {
		db.Create(&p)
	}

	methods := []PaymentMethod{
		{Name: "Pix", Type: "pix", FeePercent: 0, FeeFixedCents: 0},
		{Name: "Cartão de Crédito", Type: "credit_card", FeePercent: 3.5, FeeFixedCents: 0},
		{Name: "Boleto Bancário", Type: "boleto", FeePercent: 0, FeeFixedCents: 349},
		{Name: "Dinheiro", Type: "cash", FeePercent: 0, FeeFixedCents: 0},
		{Name: "PicPay", Type: "wallet", FeePercent: 1.5, FeeFixedCents: 0},
	}
	for _, m := range methods {
		db.Create(&m)
	}

	gateways := []PaymentGateway{
		{Name: "Mercado Pago", Type: "mercadopago", Enabled: false},
		{Name: "PicPay", Type: "picpay", Enabled: false},
	}
	for _, g := range gateways {
		db.Create(&g)
	}
}

// migrateDefaultPlans converte os planos padrão antigos (baseados em duração:
// Mensal/Trimestral/Semestral/Anual) para o novo conjunto de planos de pilates
// (Avulso diária e planos por frequência semanal), preservando os IDs para não
// quebrar vínculos existentes com StudentPlan/Invoice.
func migrateDefaultPlans(db *gorm.DB) {
	type target struct {
		oldName       string
		newName       string
		description   string
		durationDays  int
		priceCents    int
		cardCents     int
		gracePeriod   int
	}

	targets := []target{
		{oldName: "Mensal", newName: "Plano 2x por semana", description: "Duas aulas semanais de pilates. Valor no pix/dinheiro: R$ 220,00 · Cartão: R$ 230,00.", durationDays: 30, priceCents: 22000, cardCents: 23000, gracePeriod: 5},
		{oldName: "Trimestral", newName: "Plano 3x por semana", description: "Três aulas semanais de pilates. Valor no pix/dinheiro: R$ 280,00 · Cartão: R$ 290,00.", durationDays: 30, priceCents: 28000, cardCents: 29000, gracePeriod: 5},
		{oldName: "Semestral", newName: "Plano 4x por semana", description: "Quatro aulas semanais de pilates. Valor no pix/dinheiro: R$ 330,00 · Cartão: R$ 340,00.", durationDays: 30, priceCents: 33000, cardCents: 34000, gracePeriod: 5},
		{oldName: "Anual", newName: "Plano 5x por semana", description: "Cinco aulas semanais de pilates. Valor no pix/dinheiro: R$ 430,00 · Cartão: R$ 440,00.", durationDays: 30, priceCents: 43000, cardCents: 44000, gracePeriod: 5},
	}

	for _, t := range targets {
		var count int64
		db.Model(&Plan{}).Where("name = ?", t.oldName).Count(&count)
		if count == 0 {
			continue
		}
		db.Model(&Plan{}).
			Where("name = ?", t.oldName).
			Updates(map[string]interface{}{
				"name":                t.newName,
				"description":         t.description,
				"duration_days":       t.durationDays,
				"price_cents":         t.priceCents,
				"preco_cartao_cents":  t.cardCents,
				"grace_period_days":   t.gracePeriod,
			})
	}

	// Garante que o plano Avulso (Diária) exista.
	var avulsoCount int64
	db.Model(&Plan{}).Where("name = ?", "Avulso (Diária)").Count(&avulsoCount)
	if avulsoCount == 0 {
		db.Create(&Plan{
			Name:             "Avulso (Diária)",
			Description:      "Aula avulsa de pilates, acesso por um único dia.",
			DurationDays:     1,
			PriceCents:       3500,
			PrecoCartaoCents: intPtr(3500),
			GracePeriodDays:  0,
			Active:           true,
		})
	}
}
