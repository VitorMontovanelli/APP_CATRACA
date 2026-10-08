package main

func (a *App) ObterMetricasHome() (*MetricasHome, error) {
	var stats MetricasHome

	// Total usuários
	a.db.Model(&User{}).Count(&stats.TotalUsuarios)

	// Admins
	a.db.Model(&User{}).Where("cargo = ? OR cargo = ?", "super_admin", "admin").Count(&stats.TotalAdmins)

	// Alunos matriculados
	a.db.Model(&Student{}).Count(&stats.TotalAlunos)

	// Total Faturado
	a.db.Model(&Invoice{}).Where("status = ?", "paid").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalFaturadoCents)

	// Total Atraso
	a.db.Model(&Invoice{}).Where("status = ?", "overdue").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalAtrasoCents)

	// Total Pendente
	a.db.Model(&Invoice{}).Where("status = ?", "pending").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalPendenteCents)

	// Distribuição de planos
	type result struct {
		Name string
		Qtd  int64
	}
	var dist []result
	err := a.db.Raw(`
		SELECT p.name, COUNT(sp.id) as qtd
		FROM plans p
		LEFT JOIN student_plans sp ON sp.plan_id = p.id AND sp.status = 'active'
		GROUP BY p.id
	`).Scan(&dist).Error
	if err == nil {
		stats.PlanosDistribuicao = make([]PlanoContador, len(dist))
		for i, r := range dist {
			stats.PlanosDistribuicao[i] = PlanoContador{
				Nome: r.Name,
				Qtd:  r.Qtd,
			}
		}
	}

	return &stats, nil
}

type AcessoEstatistica struct {
	Label    string `json:"label"`
	Liberado int    `json:"liberado"`
	Negado   int    `json:"negado"`
}
