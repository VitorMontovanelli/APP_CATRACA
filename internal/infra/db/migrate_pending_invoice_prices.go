package db

import (
	"log"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

// MigratePendingInvoicePrices corrige faturas em aberto que ficaram com valor
// zerado por causa da leitura incorreta do preço do plano. Faturas pagas,
// canceladas ou estornadas não são tocadas.
func MigratePendingInvoicePrices(db *gorm.DB) {
	type pendente struct {
		ID         uint
		PlanID     uint
		PriceCents int
	}
	var pendentes []pendente
	err := db.Raw(`SELECT i.id, i.plan_id, p.price_cents
		FROM invoices i JOIN plans p ON p.id = i.plan_id
		WHERE i.status IN ('pending', 'overdue') AND i.amount_cents <= 0`).Scan(&pendentes).Error
	if err != nil {
		return
	}
	corrigidas := 0
	for _, p := range pendentes {
		if p.PriceCents <= 0 {
			continue
		}
		res := db.Model(&models.Invoice{}).
			Where("id = ? AND status IN ('pending', 'overdue') AND amount_cents <= 0", p.ID).
			Update("amount_cents", p.PriceCents)
		if res.Error == nil && res.RowsAffected > 0 {
			corrigidas++
		}
	}
	if len(pendentes) > 0 {
		log.Printf("faturas em aberto com valor zerado encontradas: %d, corrigidas com o preço do plano: %d", len(pendentes), corrigidas)
	}
}
