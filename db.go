package main

import (
	"log"
	"time"

	"github.com/glebarez/sqlite"
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
	)
	if err != nil {
		log.Fatalf("Falha ao migrar banco de dados: %v", err)
	}

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

	seedData(db)

	return db
}

func seedData(db *gorm.DB) {
	cleanTestData(db)
	seedSuperAdmin(db)
	seedTestUsers(db)
	seedPaymentData(db)
	seedAlunosTeste(db)
	seedInadimplentesTeste(db)
	seedAnualPlan(db)
}

func cleanTestData(db *gorm.DB) {
	db.Exec("DELETE FROM invoices WHERE student_id IN (SELECT id FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%')")
	db.Exec("DELETE FROM student_plans WHERE student_id IN (SELECT id FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%')")
	db.Exec("DELETE FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%'")
	db.Exec("DELETE FROM plans WHERE name LIKE '%Teste%' OR name LIKE '%Eduardo%'")
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
		Senha: "admin123",
		Cargo: "super_admin",
	})
}

func seedTestUsers(db *gorm.DB) {
	var count int64
	db.Model(&User{}).Where("email = ?", "admin_teste@catraca.com").Count(&count)
	if count > 0 {
		return
	}
	var superAdmin User
	db.Where("cargo = ?", "super_admin").First(&superAdmin)

	db.Create(&User{
		Name:          "Admin Teste",
		Email:         "admin_teste@catraca.com",
		Senha:         "admin123",
		Cargo:         "admin",
		RegistradorID: &superAdmin.ID,
	})
}

func seedAlunosTeste(db *gorm.DB) {
	var count int64
	db.Model(&Aluno{}).Count(&count)
	if count > 0 {
		return
	}

	db.Create(&Aluno{
		Nome:            "Carlos Aluno",
		CPF:             "111.222.333-44",
		Status:          true,
		VencimentoPlano: time.Now().AddDate(0, 1, 0),
		IDBiometriaMock: 1001,
	})
	db.Create(&Aluno{
		Nome:            "Maria Atleta",
		CPF:             "555.666.777-88",
		Status:          true,
		VencimentoPlano: time.Now().AddDate(0, 2, 0),
		IDBiometriaMock: 1002,
	})
	db.Create(&Aluno{
		Nome:            "João Vencido",
		CPF:             "999.888.777-66",
		Status:          true,
		VencimentoPlano: time.Now().AddDate(0, -1, 0),
		IDBiometriaMock: 1003,
	})

	// Students legado
	var sCount int64
	db.Model(&Student{}).Count(&sCount)
	if sCount == 0 {
		db.Create(&Student{Nome: "Aluno Teste", CPF: "111.222.333-45"})
		db.Create(&Student{Nome: "Maria Silva", CPF: "555.666.777-89"})
	}
}

func seedPaymentData(db *gorm.DB) {
	var count int64
	db.Model(&Plan{}).Count(&count)
	if count > 0 {
		return
	}

	plans := []Plan{
		{Name: "Mensal", Description: "Pagamento mensal com acesso ilimitado", DurationDays: 30, PriceCents: 12900},
		{Name: "Trimestral", Description: "3 meses com 10% de desconto", DurationDays: 90, PriceCents: 34900},
		{Name: "Semestral", Description: "6 meses com 15% de desconto", DurationDays: 180, PriceCents: 64900},
		{Name: "Anual", Description: "12 meses com 20% de desconto", DurationDays: 365, PriceCents: 119900},
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

func seedInadimplentesTeste(db *gorm.DB) {
	var count int64
	db.Model(&StudentPlan{}).Count(&count)
	if count > 0 {
		return
	}

	var students []Student
	db.Find(&students)
	var plans []Plan
	db.Find(&plans)
	if len(students) < 2 || len(plans) == 0 {
		return
	}

	now := time.Now()

	// Aluno 1 — 2 dias vencido (amarelo)
	sp1 := StudentPlan{
		StudentID: students[0].ID,
		PlanID:    plans[0].ID,
		DueDay:    5,
		Status:    "active",
		StartDate: now.AddDate(0, -2, 0),
	}
	db.Create(&sp1)
	db.Create(&Invoice{
		StudentPlanID: sp1.ID,
		StudentID:     students[0].ID,
		PlanID:        plans[0].ID,
		AmountCents:   plans[0].PriceCents,
		Status:        "pending",
		DueDate:       now.AddDate(0, 0, -2).Format("2006-01-02"),
	})

	// Aluno 1 — fatura paga anterior (histórico de pagamento)
	db.Create(&Invoice{
		StudentPlanID: sp1.ID,
		StudentID:     students[0].ID,
		PlanID:        plans[0].ID,
		AmountCents:   plans[0].PriceCents,
		Status:        "paid",
		DueDate:       now.AddDate(0, -1, -2).Format("2006-01-02"),
		PaidAt:        timePtr(now.AddDate(0, -1, 0)),
		PaidAmountCents: intPtr(plans[0].PriceCents),
	})

	// Aluno 2 — 10 dias vencido (vermelho), plano trimestral
	var plan2 Plan
	if len(plans) > 1 {
		plan2 = plans[1]
	} else {
		plan2 = plans[0]
	}
	sp2 := StudentPlan{
		StudentID: students[1].ID,
		PlanID:    plan2.ID,
		DueDay:    10,
		Status:    "active",
		StartDate: now.AddDate(0, -3, 0),
	}
	db.Create(&sp2)
	db.Create(&Invoice{
		StudentPlanID: sp2.ID,
		StudentID:     students[1].ID,
		PlanID:        plan2.ID,
		AmountCents:   plan2.PriceCents,
		Status:        "pending",
		DueDate:       now.AddDate(0, 0, -10).Format("2006-01-02"),
	})

	// Aluno 2 — segunda fatura pendente (total em aberto maior)
	db.Create(&Invoice{
		StudentPlanID: sp2.ID,
		StudentID:     students[1].ID,
		PlanID:        plan2.ID,
		AmountCents:   plan2.PriceCents,
		Status:        "overdue",
		DueDate:       now.AddDate(0, -1, -10).Format("2006-01-02"),
	})

	// Aluno 2 — pagamento antigo
	db.Create(&Invoice{
		StudentPlanID: sp2.ID,
		StudentID:     students[1].ID,
		PlanID:        plan2.ID,
		AmountCents:   plan2.PriceCents,
		Status:        "paid",
		DueDate:       now.AddDate(0, -2, -10).Format("2006-01-02"),
		PaidAt:        timePtr(now.AddDate(0, -2, -5)),
		PaidAmountCents: intPtr(plan2.PriceCents),
	})
}

func seedAnualPlan(db *gorm.DB) {
	var count int64
	db.Model(&Plan{}).Where("name = ?", "Anual").Count(&count)
	if count > 0 {
		return
	}
	db.Create(&Plan{
		Name: "Anual", Description: "12 meses com 20% de desconto",
		DurationDays: 365, PriceCents: 119900, GracePeriodDays: 5, Active: true,
	})
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func intPtr(n int) *int {
	return &n
}
