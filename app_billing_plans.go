package main

import (
	"fmt"
)

// =============== Planos (CRUD) ===============

func (a *App) ListarPlanos() ([]Plan, error) {
	list := []Plan{}
	err := a.db.Order("price_cents").Find(&list).Error
	return list, err
}

func (a *App) CriarPlano(name, description string, durationDays, priceCents, gracePeriodDays int, precoCartaoCents *int) (*Plan, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	plan := Plan{
		Name:             name,
		Description:      description,
		DurationDays:     durationDays,
		PriceCents:       priceCents,
		PrecoCartaoCents: precoCartaoCents,
		GracePeriodDays:  gracePeriodDays,
	}
	if err := a.db.Create(&plan).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_plano", plan.ID)
	return a.BuscarPlano(plan.ID)
}

func (a *App) BuscarPlano(id uint) (*Plan, error) {
	var p Plan
	err := a.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (a *App) AtualizarPlano(id uint, name, description string, durationDays, priceCents, gracePeriodDays int, precoCartaoCents *int) (*Plan, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	plan, err := a.BuscarPlano(id)
	if err != nil {
		return nil, err
	}
	plan.Name = name
	plan.Description = description
	plan.DurationDays = durationDays
	plan.PriceCents = priceCents
	plan.PrecoCartaoCents = precoCartaoCents
	plan.GracePeriodDays = gracePeriodDays
	if err := a.db.Save(plan).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("editou_plano", id)
	return a.BuscarPlano(id)
}

func (a *App) AtivarPlano(id uint, ativo bool) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&Plan{}).Where("id = ?", id).Update("active", ativo).Error
	if err == nil {
		action := "desativou_plano"
		if ativo {
			action = "ativou_plano"
		}
		a.registrarAudit(action, id)
	}
	return err
}

func (a *App) DeletarPlano(id uint) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	n := a.db.Delete(&Plan{}, id).RowsAffected
	if n == 0 {
		return fmt.Errorf("plano não encontrado")
	}
	a.registrarAudit("deletou_plano", id)
	return nil
}

type AlunoPorPlano struct {
	StudentID uint   `json:"student_id"`
	Nome      string `json:"nome"`
	CPF       string `json:"cpf"`
	Status    string `json:"status"`
	StartDate string `json:"start_date"`
}

func (a *App) ListarAlunosPorPlano(planID uint, page, pageSize int) ([]AlunoPorPlano, int, error) {
	var total int64
	a.db.Model(&StudentPlan{}).Where("plan_id = ?", planID).Count(&total)

	list := []AlunoPorPlano{}
	err := a.db.Raw(`
		SELECT s.id, s.nome, s.cpf, sp.status, sp.start_date
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		WHERE sp.plan_id = ?
		ORDER BY sp.start_date DESC
		LIMIT ? OFFSET ?
	`, planID, pageSize, (page-1)*pageSize).Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, int(total), nil
}

// =============== Métodos de Pagamento ===============

func (a *App) ListarPaymentMethods() ([]PaymentMethod, error) {
	list := []PaymentMethod{}
	err := a.db.Order("name").Find(&list).Error
	return list, err
}

func (a *App) CriarPaymentMethod(name, mtype string, feePercent float64, feeFixedCents int) (*PaymentMethod, error) {
	if a.actorCargo != "super_admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	pm := PaymentMethod{
		Name:          name,
		Type:          mtype,
		FeePercent:    feePercent,
		FeeFixedCents: feeFixedCents,
	}
	if err := a.db.Create(&pm).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_metodo_pagamento", pm.ID)
	return &pm, nil
}

func (a *App) AtivarPaymentMethod(id uint, enabled bool) error {
	if a.actorCargo != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&PaymentMethod{}).Where("id = ?", id).Update("enabled", enabled).Error
	if err == nil {
		action := "desativou_metodo_pagamento"
		if enabled {
			action = "ativou_metodo_pagamento"
		}
		a.registrarAudit(action, id)
	}
	return err
}

// =============== Gateways ===============

func (a *App) ListarPaymentGateways() ([]PaymentGateway, error) {
	list := []PaymentGateway{}
	err := a.db.Order("name").Find(&list).Error
	return list, err
}

func (a *App) AtivarPaymentGateway(id uint, enabled bool) error {
	if a.actorCargo != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&PaymentGateway{}).Where("id = ?", id).Update("enabled", enabled).Error
	if err == nil {
		action := "desativou_gateway"
		if enabled {
			action = "ativou_gateway"
		}
		a.registrarAudit(action, id)
	}
	return err
}

func (a *App) AtualizarConfigGateway(id uint, config string) error {
	if a.actorCargo != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&PaymentGateway{}).Where("id = ?", id).Update("config", config).Error
	if err == nil {
		a.registrarAudit("configurou_gateway", id)
	}
	return err
}
