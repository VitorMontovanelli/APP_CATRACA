package main

import (
	"gorm.io/gorm"
)

// migrateDefaultPlans converte os planos padrão antigos (baseados em duração:
// Mensal/Trimestral/Semestral/Anual) para o novo conjunto de planos de pilates
// (Avulso diária e planos por frequência semanal), preservando os IDs para não
// quebrar vínculos existentes com StudentPlan/Invoice.
func migrateDefaultPlans(db *gorm.DB) {
	type target struct {
		oldName      string
		newName      string
		description  string
		durationDays int
		priceCents   int
		cardCents    int
		gracePeriod  int
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
				"name":               t.newName,
				"description":        t.description,
				"duration_days":      t.durationDays,
				"price_cents":        t.priceCents,
				"preco_cartao_cents": t.cardCents,
				"grace_period_days":  t.gracePeriod,
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
