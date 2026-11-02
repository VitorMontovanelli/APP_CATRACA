package inbound

import (
	"context"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type SystemMetricsHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewSystemMetricsHandler(db *gorm.DB, session *SessionManager) *SystemMetricsHandler {
	return &SystemMetricsHandler{db: db, session: session}
}

func (h *SystemMetricsHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *SystemMetricsHandler) ObterMetricasHome() (*models.MetricasHome, error) {
	var stats models.MetricasHome

	h.db.Model(&models.User{}).Count(&stats.TotalUsuarios)
	h.db.Model(&models.User{}).Where("cargo = ? OR cargo = ?", "super_admin", "admin").Count(&stats.TotalAdmins)
	h.db.Model(&models.Student{}).Count(&stats.TotalAlunos)
	h.db.Model(&models.Invoice{}).Where("status = ?", "paid").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalFaturadoCents)
	h.db.Model(&models.Invoice{}).Where("status = ?", "overdue").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalAtrasoCents)
	h.db.Model(&models.Invoice{}).Where("status = ?", "pending").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalPendenteCents)

	type result struct {
		Name string
		Qtd  int64
	}
	var dist []result
	query := `SELECT p.name, COUNT(sp.id) as qtd FROM plans p
		LEFT JOIN student_plans sp ON sp.plan_id = p.id AND sp.status = 'active'
		GROUP BY p.id`
	if err := h.db.Raw(query).Scan(&dist).Error; err == nil {
		stats.PlanosDistribuicao = make([]models.PlanoContador, len(dist))
		for i, r := range dist {
			stats.PlanosDistribuicao[i] = models.PlanoContador{Nome: r.Name, Qtd: r.Qtd}
		}
	}

	return &stats, nil
}
