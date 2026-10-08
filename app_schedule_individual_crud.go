package main

import (
	"fmt"
	"strings"
	"time"
)

func (a *App) AdicionarAgendamentoIndividual(data string, alunoID uint, nome, hora, pratica, observacao string) (*AgendamentoIndividual, error) {
	if !ehAdmin(a.actorCargo) {
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

	nome = strings.TrimSpace(nome)
	if alunoID > 0 {
		var s Student
		if err := a.db.First(&s, alunoID).Error; err != nil {
			return nil, fmt.Errorf("aluno não encontrado")
		}
		nome = s.Nome
	} else if nome == "" {
		return nil, fmt.Errorf("selecione um aluno cadastrado ou informe o nome")
	}

	var count int64
	a.db.Model(&AgendamentoIndividual{}).Where("data = ?", data).Count(&count)
	limite := a.limiteIndividual()
	if count >= int64(limite) {
		return nil, fmt.Errorf("limite diário de %d atendimentos atingido para este dia", limite)
	}

	var ocupado int64
	a.db.Model(&AgendamentoIndividual{}).Where("data = ? AND hora = ?", data, hora).Count(&ocupado)
	if ocupado > 0 {
		return nil, fmt.Errorf("o horário %s já está ocupado neste dia", hora)
	}

	ag := AgendamentoIndividual{
		Data:      data,
		Hora:      hora,
		StudentID: uintPtr(alunoID),
		Nome:      nome,
		Pratica:   pratica,
	}
	observacao = strings.TrimSpace(observacao)
	if observacao != "" {
		ag.Observacao = &observacao
	}
	if err := a.db.Create(&ag).Error; err != nil {
		return nil, err
	}
	a.registrarAuditComDetalhes("agendou_atendimento_individual", ag.ID,
		fmt.Sprintf("data=%s, hora=%s, aluno=%s, pratica=%s", data, hora, nome, pratica))
	return &ag, nil
}

// AtualizarAgendamentoIndividual altera horário, prática ou observação de um atendimento.
func (a *App) AtualizarAgendamentoIndividual(id uint, hora, pratica, observacao string) (*AgendamentoIndividual, error) {
	if !ehAdmin(a.actorCargo) {
		return nil, fmt.Errorf("permissão negada")
	}
	var ag AgendamentoIndividual
	if err := a.db.First(&ag, id).Error; err != nil {
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
		a.db.Model(&AgendamentoIndividual{}).Where("data = ? AND hora = ? AND id <> ?", ag.Data, hora, id).Count(&ocupado)
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
	if err := a.db.Save(&ag).Error; err != nil {
		return nil, err
	}
	a.registrarAuditComDetalhes("editou_atendimento_individual", ag.ID,
		fmt.Sprintf("data=%s, hora=%s, aluno=%s, pratica=%s", ag.Data, ag.Hora, ag.Nome, ag.Pratica))
	return &ag, nil
}

// RemoverAgendamentoIndividual libera o horário removendo o atendimento individual.
func (a *App) RemoverAgendamentoIndividual(id uint) error {
	if !ehAdmin(a.actorCargo) {
		return fmt.Errorf("permissão negada")
	}
	var ag AgendamentoIndividual
	if err := a.db.First(&ag, id).Error; err != nil {
		return fmt.Errorf("atendimento não encontrado")
	}
	if err := a.db.Delete(&ag).Error; err != nil {
		return err
	}
	a.registrarAuditComDetalhes("removeu_atendimento_individual", id,
		fmt.Sprintf("data=%s, hora=%s, aluno=%s, pratica=%s", ag.Data, ag.Hora, ag.Nome, ag.Pratica))
	return nil
}
