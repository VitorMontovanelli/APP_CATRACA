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
	seedSuperAdmin(db)
	seedPaymentData(db)
	seedAnualPlan(db)
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
