package main

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// =============== Agenda Individual (Professora) ===============

const LimiteIndividualPadrao = 10
const chaveLimiteIndividual = "agenda_individual_limite"

var horariosIndividuais = []string{
	"06:00", "07:00", "08:00", "09:00", "10:00", // Manhã 06:00–11:00
	"15:00", "16:00", "17:00", "18:00", // Tarde 15:00–19:00
	"19:00", // Noite 19:00–20:00
}

var praticasIndividuais = []string{
	PraticaVentosaterapia,
	PraticaLiberacaoMiofascial,
	PraticaPersonalTrainer,
	PraticaKinesioTape,
}

func horarioIndividualValido(hora string) bool {
	for _, h := range horariosIndividuais {
		if h == hora {
			return true
		}
	}
	return false
}

func praticaValida(pratica string) bool {
	for _, p := range praticasIndividuais {
		if p == pratica {
			return true
		}
	}
	return false
}

// limiteIndividual retorna o limite diário configurado (padrão 10).
func (a *App) limiteIndividual() int {
	var s Setting
	if err := a.db.Where("`key` = ?", chaveLimiteIndividual).First(&s).Error; err != nil {
		return LimiteIndividualPadrao
	}
	n, err := strconv.Atoi(s.Value)
	if err != nil || n < 1 || n > LimiteIndividualPadrao {
		return LimiteIndividualPadrao
	}
	return n
}

// ObterConfigAgendaIndividual retorna a configuração atual da agenda individual.
func (a *App) ObterConfigAgendaIndividual() (*ConfigAgendaIndividual, error) {
	if !ehAdmin(a.actorCargo) {
		return nil, fmt.Errorf("permissão negada")
	}
	return &ConfigAgendaIndividual{
		LimiteDiario: a.limiteIndividual(),
		LimiteMaximo: LimiteIndividualPadrao,
	}, nil
}

// DefinirLimiteDiarioIndividual ajusta o limite diário de atendimentos individuais (1 a 10).
func (a *App) DefinirLimiteDiarioIndividual(limite int) error {
	if !ehAdmin(a.actorCargo) {
		return fmt.Errorf("permissão negada")
	}
	if limite < 1 || limite > LimiteIndividualPadrao {
		return fmt.Errorf("o limite diário deve estar entre 1 e %d atendimentos", LimiteIndividualPadrao)
	}
	var s Setting
	err := a.db.Where("`key` = ?", chaveLimiteIndividual).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		s = Setting{Key: chaveLimiteIndividual, Value: strconv.Itoa(limite)}
		if err := a.db.Create(&s).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		s.Value = strconv.Itoa(limite)
		if err := a.db.Save(&s).Error; err != nil {
			return err
		}
	}
	a.registrarAuditComDetalhes("ajustou_limite_agenda_individual", 0,
		fmt.Sprintf("limite_diario=%d", limite))
	return nil
}

// ListarAgendamentosIndividuaisMes retorna a ocupação individual de cada dia do mês.
func (a *App) ListarAgendamentosIndividuaisMes(ano, mes int) ([]DiaIndividualAgenda, error) {
	if !ehAdmin(a.actorCargo) {
		return nil, fmt.Errorf("permissão negada")
	}
	prefix := fmt.Sprintf("%04d-%02d", ano, mes)
	list := []DiaIndividualAgenda{}
	limite := a.limiteIndividual()
	err := a.db.Raw(`
		SELECT COALESCE(i.data, c.data) AS data,
		       COUNT(i.id) AS total,
		       COALESCE(c.capacidade, ?) AS capacidade,
		       COALESCE(GROUP_CONCAT(i.pratica), '') AS praticas
		FROM agendamento_individuals i
		LEFT JOIN capacidade_dias c ON c.data = i.data
		WHERE COALESCE(i.data, c.data) LIKE ?
		GROUP BY COALESCE(i.data, c.data)

		UNION

		SELECT c.data AS data,
		       0 AS total,
		       c.capacidade AS capacidade,
		       '' AS praticas
		FROM capacidade_dias c
		WHERE c.data LIKE ?
		  AND c.data NOT IN (SELECT DISTINCT data FROM agendamento_individuals WHERE data LIKE ?)
	`, limite, prefix+"%", prefix+"%", prefix+"%").Scan(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []DiaIndividualAgenda{}
	}
	return list, nil
}

// ListarAgendamentosIndividuaisDia retorna os atendimentos individuais de um dia.
func (a *App) ListarAgendamentosIndividuaisDia(data string) ([]AgendamentoIndividual, error) {
	if !ehAdmin(a.actorCargo) {
		return nil, fmt.Errorf("permissão negada")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return nil, fmt.Errorf("data inválida")
	}
	list := []AgendamentoIndividual{}
	err := a.db.Where("data = ?", data).Order("hora, id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []AgendamentoIndividual{}
	}
	return list, nil
}

// AdicionarAgendamentoIndividual marca um atendimento individual no dia e horário escolhidos.
// alunoID > 0 vincula um aluno cadastrado; alunoID == 0 usa o nome digitado.
