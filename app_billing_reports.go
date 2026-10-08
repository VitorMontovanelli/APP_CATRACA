package main

import (
	"fmt"
	"time"
)

func (a *App) ListarCobranca() ([]CobrancaAluno, error) {
	type alunoRaw struct {
		StudentID   uint
		Nome        string
		CPF         string
		PlanoNome   string
		PlanoPreco  int
		StatusPlano string
		DueDay      int
	}
	var alunos []alunoRaw
	err := a.db.Raw(`
		SELECT s.id AS student_id, s.nome, s.cpf, p.name AS plano_nome,
			p.price_cents AS plano_preco, sp.status AS status_plano, sp.due_day AS due_day
		FROM students s
		JOIN student_plans sp ON sp.student_id = s.id AND sp.status IN ('active','overdue','expired')
		JOIN plans p ON p.id = sp.plan_id
		ORDER BY s.nome
	`).Scan(&alunos).Error
	if err != nil {
		return nil, err
	}

	result := make([]CobrancaAluno, len(alunos))
	for i, al := range alunos {
		var faturas []CobrancaFatura
		a.db.Raw(`
			SELECT id, amount_cents, status, due_date,
				paid_at, comprovante,
				CAST(julianday('now') - julianday(due_date) AS INTEGER) AS dias_vencido
			FROM invoices
			WHERE student_plan_id = (SELECT id FROM student_plans WHERE student_id = ? ORDER BY id DESC LIMIT 1)
			ORDER BY due_date DESC
		`, al.StudentID).Scan(&faturas)

		if faturas == nil {
			faturas = []CobrancaFatura{}
		}
		result[i] = CobrancaAluno{
			StudentID:   al.StudentID,
			Nome:        al.Nome,
			CPF:         al.CPF,
			PlanoNome:   al.PlanoNome,
			PlanoPreco:  al.PlanoPreco,
			StatusPlano: al.StatusPlano,
			DueDay:      al.DueDay,
			Faturas:     faturas,
		}
	}
	return result, nil
}

func (a *App) ListarInadimplentes() ([]InadimplenteReport, error) {
	type rawResult struct {
		StudentID       uint
		Nome            string
		CPF             string
		Telefone        *string
		Email           *string
		PlanoNome       string
		PlanoPreco      int
		StatusPlano     string
		InvoiceID       uint
		ValorDevido     int
		DataVencimento  string
		DiasVencido     int
		UltimoPagamento *string
		TotalEmAberto   int
	}

	var raw []rawResult
	err := a.db.Raw(`
		SELECT
			s.id AS student_id,
			s.nome,
			s.cpf,
			s.telefone,
			s.email,
			p.name AS plano_nome,
			p.price_cents AS plano_preco,
			sp.status AS status_plano,
			i.id AS invoice_id,
			i.amount_cents AS valor_devido,
			i.due_date AS data_vencimento,
			CAST(julianday('now') - julianday(i.due_date) AS INTEGER) AS dias_vencido,
			ult.ultimo_pagamento,
			aberto.total AS total_em_aberto
		FROM students s
		JOIN student_plans sp ON sp.student_id = s.id AND sp.status IN ('active','overdue','expired')
		JOIN plans p ON p.id = sp.plan_id
		JOIN invoices i ON i.student_plan_id = sp.id AND i.status IN ('pending','overdue') AND i.due_date < date('now')
		LEFT JOIN (
			SELECT student_id, MAX(paid_at) as ultimo_pagamento
			FROM invoices WHERE status = 'paid' GROUP BY student_id
		) ult ON ult.student_id = s.id
		JOIN (
			SELECT sp2.id as spid, SUM(i2.amount_cents) as total
			FROM student_plans sp2
			JOIN invoices i2 ON i2.student_plan_id = sp2.id AND i2.status IN ('pending','overdue')
			GROUP BY sp2.id
		) aberto ON aberto.spid = sp.id
		ORDER BY i.due_date ASC
	`).Scan(&raw).Error
	if err != nil {
		return nil, err
	}

	list := make([]InadimplenteReport, len(raw))
	for i, r := range raw {
		list[i] = InadimplenteReport{
			StudentID:       r.StudentID,
			Nome:            r.Nome,
			CPF:             r.CPF,
			Telefone:        r.Telefone,
			Email:           r.Email,
			PlanoNome:       r.PlanoNome,
			PlanoPreco:      r.PlanoPreco,
			StatusPlano:     r.StatusPlano,
			InvoiceID:       r.InvoiceID,
			ValorDevido:     r.ValorDevido,
			DataVencimento:  r.DataVencimento,
			DiasVencido:     r.DiasVencido,
			UltimoPagamento: r.UltimoPagamento,
			TotalEmAberto:   r.TotalEmAberto,
		}
	}
	return list, nil
}

func (a *App) CriarInvoiceTeste() error {
	plano := Plan{
		Name:            "Plano Eduardo",
		Description:     "Plano anual do Eduardo",
		DurationDays:    365,
		PriceCents:      120000,
		GracePeriodDays: 10,
		Active:          true,
	}
	if err := a.db.Create(&plano).Error; err != nil {
		return fmt.Errorf("erro ao criar plano: %w", err)
	}

	aluno := Student{
		Nome:  "Eduardo Teste",
		CPF:   fmt.Sprintf("888%09d", plano.ID),
		Ativo: true,
	}
	if err := a.db.Create(&aluno).Error; err != nil {
		return fmt.Errorf("erro ao criar aluno: %w", err)
	}

	sp := StudentPlan{
		StudentID: aluno.ID,
		PlanID:    plano.ID,
		DueDay:    15,
		Status:    "active",
		StartDate: time.Now().AddDate(0, -1, 0),
	}
	if err := a.db.Create(&sp).Error; err != nil {
		return fmt.Errorf("erro ao criar vínculo: %w", err)
	}

	inv := Invoice{
		StudentPlanID: sp.ID,
		StudentID:     aluno.ID,
		PlanID:        plano.ID,
		AmountCents:   plano.PriceCents,
		Status:        "overdue",
		DueDate:       time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
	}
	if err := a.db.Create(&inv).Error; err != nil {
		return fmt.Errorf("erro ao criar fatura: %w", err)
	}

	a.registrarAudit("criou_teste", inv.ID)
	return nil
}

func (a *App) DeletarDadosTeste() error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}

	a.db.Exec("DELETE FROM invoices WHERE student_id IN (SELECT id FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%')")
	a.db.Exec("DELETE FROM student_plans WHERE student_id IN (SELECT id FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%')")
	a.db.Exec("DELETE FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%'")
	a.db.Exec("DELETE FROM plans WHERE name LIKE '%Teste%' OR name LIKE '%Eduardo%'")

	a.registrarAudit("deletou_dados_teste", 0)
	return nil
}

// =============== Utilitários ===============
