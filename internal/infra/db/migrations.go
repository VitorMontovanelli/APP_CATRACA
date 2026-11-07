package db

import (
	"strings"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type planMigrationTarget struct {
	oldName      string
	newName      string
	description  string
	durationDays int
	priceCents   int
	cardCents    int
	gracePeriod  int
}

func MigrateDefaultPlans(db *gorm.DB) {
	targets := []planMigrationTarget{
		{"Mensal", "Plano 2x por semana", "Duas aulas semanais de pilates. Valor no pix/dinheiro: R$ 220,00 · Cartão: R$ 230,00.", 30, 22000, 23000, 5},
		{"Trimestral", "Plano 3x por semana", "Três aulas semanais de pilates. Valor no pix/dinheiro: R$ 280,00 · Cartão: R$ 290,00.", 30, 28000, 29000, 5},
		{"Semestral", "Plano 4x por semana", "Quatro aulas semanais de pilates. Valor no pix/dinheiro: R$ 330,00 · Cartão: R$ 340,00.", 30, 33000, 34000, 5},
		{"Anual", "Plano 5x por semana", "Cinco aulas semanais de pilates. Valor no pix/dinheiro: R$ 430,00 · Cartão: R$ 440,00.", 30, 43000, 44000, 5},
	}

	for _, t := range targets {
		updateLegacyPlan(db, t)
	}

	ensureAvulsoPlan(db)
}

func updateLegacyPlan(db *gorm.DB, t planMigrationTarget) {
	var count int64
	db.Model(&models.Plan{}).Where("name = ?", t.oldName).Count(&count)
	if count == 0 {
		return
	}
	db.Model(&models.Plan{}).
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

func MigrateAdminPermissions(db *gorm.DB) {
	var admins []models.User
	if err := db.Where("cargo = ?", "admin").Find(&admins).Error; err != nil {
		return
	}
	for _, u := range admins {
		atualizadas := semPermissaoLogs(comPermissaoAgenda(u.Permissoes))
		if atualizadas == u.Permissoes {
			continue
		}
		db.Model(&models.User{}).Where("id = ?", u.ID).Update("permissoes", atualizadas)
	}
}

// semPermissaoLogs remove o acesso aos logs de acesso do administrador
// comum, que passou a ser exclusivo do super_admin.
func semPermissaoLogs(permissoes string) string {
	itens := strings.Split(permissoes, ",")
	filtrados := make([]string, 0, len(itens))
	for _, p := range itens {
		p = strings.TrimSpace(p)
		if p == "" || p == "logs" {
			continue
		}
		filtrados = append(filtrados, p)
	}
	return strings.Join(filtrados, ",")
}

func comPermissaoAgenda(permissoes string) string {
	permissoes = strings.TrimSpace(permissoes)
	if permissoes == "" {
		return models.DefaultAdminPermissoes
	}
	for _, p := range strings.Split(permissoes, ",") {
		if strings.TrimSpace(p) == "agenda" {
			return permissoes
		}
	}
	return permissoes + ",agenda"
}

func ensureAvulsoPlan(db *gorm.DB) {
	var avulsoCount int64
	db.Model(&models.Plan{}).Where("name = ?", "Avulso (Diária)").Count(&avulsoCount)
	if avulsoCount == 0 {
		db.Create(&models.Plan{
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
