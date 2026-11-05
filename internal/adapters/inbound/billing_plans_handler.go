package inbound

import (
	"context"
	"fmt"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type BillingPlansHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewBillingPlansHandler(db *gorm.DB, session *SessionManager) *BillingPlansHandler {
	return &BillingPlansHandler{db: db, session: session}
}

func (h *BillingPlansHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *BillingPlansHandler) ListarPlanos() ([]models.Plan, error) {
	list := []models.Plan{}
	err := h.db.Order("price_cents").Find(&list).Error
	return list, err
}

func (h *BillingPlansHandler) CriarPlano(name, description string, durationDays, priceCents, gracePeriodDays int, precoCartaoCents *int) (*models.Plan, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	plan := models.Plan{
		Name:             name,
		Description:      description,
		DurationDays:     durationDays,
		PriceCents:       priceCents,
		PrecoCartaoCents: precoCartaoCents,
		GracePeriodDays:  gracePeriodDays,
	}
	if err := h.db.Create(&plan).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAudit("criou_plano", plan.ID)
	return h.BuscarPlano(plan.ID)
}

func (h *BillingPlansHandler) BuscarPlano(id uint) (*models.Plan, error) {
	var p models.Plan
	if err := h.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (h *BillingPlansHandler) AtualizarPlano(id uint, name, description string, durationDays, priceCents, gracePeriodDays int, precoCartaoCents *int) (*models.Plan, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	plan, err := h.BuscarPlano(id)
	if err != nil {
		return nil, err
	}
	plan.Name = name
	plan.Description = description
	plan.DurationDays = durationDays
	plan.PriceCents = priceCents
	plan.PrecoCartaoCents = precoCartaoCents
	plan.GracePeriodDays = gracePeriodDays
	if err := h.db.Save(plan).Error; err != nil {
		return nil, err
	}
	// Faturas em aberto passam a refletir o preço atual do plano. Faturas
	// pagas, canceladas ou estornadas nunca são alteradas.
	res := h.db.Model(&models.Invoice{}).
		Where("plan_id = ? AND status IN ('pending', 'overdue')", id).
		Update("amount_cents", priceCents)
	h.session.RegistrarAuditComDetalhes("editou_plano", id,
		fmt.Sprintf("faturas_em_aberto_atualizadas=%d", res.RowsAffected))
	return h.BuscarPlano(id)
}

func (h *BillingPlansHandler) AtivarPlano(id uint, ativo bool) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	err := h.db.Model(&models.Plan{}).Where("id = ?", id).Update("active", ativo).Error
	if err == nil {
		action := "desativou_plano"
		if ativo {
			action = "ativou_plano"
		}
		h.session.RegistrarAudit(action, id)
	}
	return err
}

func (h *BillingPlansHandler) DeletarPlano(id uint) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	n := h.db.Delete(&models.Plan{}, id).RowsAffected
	if n == 0 {
		return fmt.Errorf("plano não encontrado")
	}
	h.session.RegistrarAudit("deletou_plano", id)
	return nil
}

func (h *BillingPlansHandler) ListarAlunosPorPlano(planID uint, page, pageSize int) ([]models.AlunoPorPlano, int, error) {
	var total int64
	h.db.Model(&models.StudentPlan{}).Where("plan_id = ?", planID).Count(&total)

	list := []models.AlunoPorPlano{}
	query := `SELECT s.id as student_id, s.nome, s.cpf, sp.status, sp.start_date
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		WHERE sp.plan_id = ?
		ORDER BY sp.start_date DESC
		LIMIT ? OFFSET ?`
	err := h.db.Raw(query, planID, pageSize, (page-1)*pageSize).Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, int(total), nil
}

func (h *BillingPlansHandler) ListarPaymentMethods() ([]models.PaymentMethod, error) {
	list := []models.PaymentMethod{}
	err := h.db.Order("name").Find(&list).Error
	return list, err
}

func (h *BillingPlansHandler) CriarPaymentMethod(name, mtype string, feePercent float64, feeFixedCents int) (*models.PaymentMethod, error) {
	if h.session.ActorCargo() != "super_admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	pm := models.PaymentMethod{Name: name, Type: mtype, FeePercent: feePercent, FeeFixedCents: feeFixedCents}
	if err := h.db.Create(&pm).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAudit("criou_metodo_pagamento", pm.ID)
	return &pm, nil
}

func (h *BillingPlansHandler) AtivarPaymentMethod(id uint, enabled bool) error {
	if h.session.ActorCargo() != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := h.db.Model(&models.PaymentMethod{}).Where("id = ?", id).Update("enabled", enabled).Error
	if err == nil {
		action := "desativou_metodo_pagamento"
		if enabled {
			action = "ativou_metodo_pagamento"
		}
		h.session.RegistrarAudit(action, id)
	}
	return err
}

func (h *BillingPlansHandler) ListarPaymentGateways() ([]models.PaymentGateway, error) {
	list := []models.PaymentGateway{}
	err := h.db.Order("name").Find(&list).Error
	return list, err
}

func (h *BillingPlansHandler) AtivarPaymentGateway(id uint, enabled bool) error {
	if h.session.ActorCargo() != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := h.db.Model(&models.PaymentGateway{}).Where("id = ?", id).Update("enabled", enabled).Error
	if err == nil {
		action := "desativou_gateway"
		if enabled {
			action = "ativou_gateway"
		}
		h.session.RegistrarAudit(action, id)
	}
	return err
}

func (h *BillingPlansHandler) AtualizarConfigGateway(id uint, config string) error {
	if h.session.ActorCargo() != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := h.db.Model(&models.PaymentGateway{}).Where("id = ?", id).Update("config", config).Error
	if err == nil {
		h.session.RegistrarAudit("configurou_gateway", id)
	}
	return err
}
