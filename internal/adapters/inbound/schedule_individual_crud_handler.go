package inbound

import (
	"context"
	"fmt"
	"strings"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type ScheduleIndividualCrudHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewScheduleIndividualCrudHandler(db *gorm.DB, session *SessionManager) *ScheduleIndividualCrudHandler {
	return &ScheduleIndividualCrudHandler{db: db, session: session}
}

func (h *ScheduleIndividualCrudHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *ScheduleIndividualCrudHandler) AdicionarAgendamentoIndividual(data string, alunoID uint, nome, hora, pratica, observacao string) (*models.AgendamentoIndividual, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return nil, fmt.Errorf("data inválida")
	}
	if !horarioIndividualValido(hora) {
		return nil, fmt.Errorf("horário inválido: use um dos horários disponíveis da grade")
	}
	if !praticaValida(pratica) {
		return nil, fmt.Errorf("prática inválida")
	}

	nome, err := h.resolverNomeIndividual(alunoID, nome)
	if err != nil {
		return nil, err
	}

	if err := h.validarVagaIndividual(data, hora); err != nil {
		return nil, err
	}

	ag := models.AgendamentoIndividual{Data: data, Hora: hora, StudentID: uintPtr(alunoID), Nome: nome, Pratica: pratica}
	observacao = strings.TrimSpace(observacao)
	if observacao != "" {
		ag.Observacao = &observacao
	}
	if err := h.db.Create(&ag).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAuditComDetalhes("agendou_atendimento_individual", ag.ID, fmt.Sprintf("data=%s, hora=%s, aluno=%s, pratica=%s", data, hora, nome, pratica))
	return &ag, nil
}

func (h *ScheduleIndividualCrudHandler) resolverNomeIndividual(alunoID uint, nome string) (string, error) {
	nome = strings.TrimSpace(nome)
	if alunoID > 0 {
		var s models.Student
		if err := h.db.First(&s, alunoID).Error; err != nil {
			return "", fmt.Errorf("aluno não encontrado")
		}
		return s.Nome, nil
	}
	if nome == "" {
		return "", fmt.Errorf("selecione um aluno cadastrado ou informe o nome")
	}
	return nome, nil
}

func (h *ScheduleIndividualCrudHandler) validarVagaIndividual(data, hora string) error {
	var count int64
	h.db.Model(&models.AgendamentoIndividual{}).Where("data = ?", data).Count(&count)
	limite := LimiteIndividualPadrao
	if count >= int64(limite) {
		return fmt.Errorf("limite diário de %d atendimentos atingido para este dia", limite)
	}

	var ocupado int64
	h.db.Model(&models.AgendamentoIndividual{}).Where("data = ? AND hora = ?", data, hora).Count(&ocupado)
	if ocupado > 0 {
		return fmt.Errorf("o horário %s já está ocupado neste dia", hora)
	}
	return nil
}

func (h *ScheduleIndividualCrudHandler) AtualizarAgendamentoIndividual(id uint, hora, pratica, observacao string) (*models.AgendamentoIndividual, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	var ag models.AgendamentoIndividual
	if err := h.db.First(&ag, id).Error; err != nil {
		return nil, fmt.Errorf("atendimento não encontrado")
	}
	if !horarioIndividualValido(hora) {
		return nil, fmt.Errorf("horário inválido: use um dos horários disponíveis da grade")
	}
	if !praticaValida(pratica) {
		return nil, fmt.Errorf("prática inválida")
	}
	if hora != ag.Hora {
		var ocupado int64
		h.db.Model(&models.AgendamentoIndividual{}).Where("data = ? AND hora = ? AND id <> ?", ag.Data, hora, id).Count(&ocupado)
		if ocupado > 0 {
			return nil, fmt.Errorf("o horário %s já está ocupado neste dia", hora)
		}
	}
	ag.Hora = hora
	ag.Pratica = pratica
	observacao = strings.TrimSpace(observacao)
	if observacao != "" {
		ag.Observacao = &observacao
	} else {
		ag.Observacao = nil
	}
	if err := h.db.Save(&ag).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAuditComDetalhes("editou_atendimento_individual", ag.ID, fmt.Sprintf("data=%s, hora=%s, aluno=%s, pratica=%s", ag.Data, ag.Hora, ag.Nome, ag.Pratica))
	return &ag, nil
}

func (h *ScheduleIndividualCrudHandler) RemoverAgendamentoIndividual(id uint) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	var ag models.AgendamentoIndividual
	if err := h.db.First(&ag, id).Error; err != nil {
		return fmt.Errorf("atendimento não encontrado")
	}
	if err := h.db.Delete(&ag).Error; err != nil {
		return err
	}
	h.session.RegistrarAuditComDetalhes("removeu_atendimento_individual", id, fmt.Sprintf("data=%s, hora=%s, aluno=%s, pratica=%s", ag.Data, ag.Hora, ag.Nome, ag.Pratica))
	return nil
}
