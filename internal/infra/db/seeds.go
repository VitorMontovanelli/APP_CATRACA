package db

import (
	"log"

	"catraca-app/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) {
	seedSuperAdmin(db)
	seedPaymentData(db)
}

func intPtr(v int) *int {
	return &v
}

func HashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Falha ao hashear senha de seed: %v", err)
	}
	return string(hashed)
}

func seedSuperAdmin(db *gorm.DB) {
	var admin models.User
	err := db.Where("email = ? OR email = ?", "admin", "admin@catraca.com").First(&admin).Error
	if err != nil {
		db.Create(&models.User{
			Name:       "Administrador",
			Email:      "admin",
			Senha:      HashPassword("admin"),
			Cargo:      "super_admin",
			Ativo:      true,
			Permissoes: "home,alunos,planos,financeiro,cobranca,metodos_pagamento,usuarios,logs",
		})
		return
	}

	db.Model(&admin).Updates(map[string]interface{}{
		"name":       "Administrador",
		"email":      "admin",
		"cargo":      "super_admin",
		"ativo":      true,
		"senha":      HashPassword("admin"),
		"permissoes": "home,alunos,planos,financeiro,cobranca,metodos_pagamento,usuarios,logs",
	})
}

func seedPaymentData(db *gorm.DB) {
	seedPlans(db)
	seedMethods(db)
	seedGateways(db)
}

func seedPlans(db *gorm.DB) {
	var count int64
	db.Model(&models.Plan{}).Count(&count)
	if count > 0 {
		return
	}

	plans := []models.Plan{
		{Name: "Avulso (Diária)", Description: "Aula avulsa de pilates, acesso por um único dia.", DurationDays: 1, PriceCents: 3500, PrecoCartaoCents: intPtr(3500), GracePeriodDays: 0, Active: true},
		{Name: "Plano 2x por semana", Description: "Duas aulas semanais de pilates. Valor no pix/dinheiro: R$ 220,00 · Cartão: R$ 230,00.", DurationDays: 30, PriceCents: 22000, PrecoCartaoCents: intPtr(23000), GracePeriodDays: 5, Active: true},
		{Name: "Plano 3x por semana", Description: "Três aulas semanais de pilates. Valor no pix/dinheiro: R$ 280,00 · Cartão: R$ 290,00.", DurationDays: 30, PriceCents: 28000, PrecoCartaoCents: intPtr(29000), GracePeriodDays: 5, Active: true},
		{Name: "Plano 4x por semana", Description: "Quatro aulas semanais de pilates. Valor no pix/dinheiro: R$ 330,00 · Cartão: R$ 340,00.", DurationDays: 30, PriceCents: 33000, PrecoCartaoCents: intPtr(34000), GracePeriodDays: 5, Active: true},
		{Name: "Plano 5x por semana", Description: "Cinco aulas semanais de pilates. Valor no pix/dinheiro: R$ 430,00 · Cartão: R$ 440,00.", DurationDays: 30, PriceCents: 43000, PrecoCartaoCents: intPtr(44000), GracePeriodDays: 5, Active: true},
	}
	for _, p := range plans {
		db.Create(&p)
	}
}

func seedMethods(db *gorm.DB) {
	var count int64
	db.Model(&models.PaymentMethod{}).Count(&count)
	if count > 0 {
		return
	}
	methods := []models.PaymentMethod{
		{Name: "Pix", Type: "pix", FeePercent: 0, FeeFixedCents: 0},
		{Name: "Cartão de Crédito", Type: "credit_card", FeePercent: 3.5, FeeFixedCents: 0},
		{Name: "Boleto Bancário", Type: "boleto", FeePercent: 0, FeeFixedCents: 349},
		{Name: "Dinheiro", Type: "cash", FeePercent: 0, FeeFixedCents: 0},
		{Name: "PicPay", Type: "wallet", FeePercent: 1.5, FeeFixedCents: 0},
	}
	for _, m := range methods {
		db.Create(&m)
	}
}

func seedGateways(db *gorm.DB) {
	var count int64
	db.Model(&models.PaymentGateway{}).Count(&count)
	if count > 0 {
		return
	}
	gateways := []models.PaymentGateway{
		{Name: "Mercado Pago", Type: "mercadopago", Enabled: false},
		{Name: "PicPay", Type: "picpay", Enabled: false},
	}
	for _, g := range gateways {
		db.Create(&g)
	}
}
