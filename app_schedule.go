package main

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const CapacidadeDiaPadrao = 10

func turnoValido(turno string) bool {
	return turno == "manha" || turno == "tarde" || turno == "noite"
}

func (a *App) capacidadeDoDia(data string) int {
	var cd CapacidadeDia
	if err := a.db.Where("data = ?", data).First(&cd).Error; err != nil {
		return CapacidadeDiaPadrao
	}
	return cd.Capacidade
}

// ListarAgendamentosMes retorna a ocupação (total de vagas preenchidas) e a capacidade de cada dia do mês.
func (a *App) ListarAgendamentosMes(ano, mes int) ([]DiaAgenda, error) {
	prefix := fmt.Sprintf("%04d-%02d", ano, mes)
	list := []DiaAgenda{}
	err := a.db.Raw(`
		SELECT a.data, COUNT(a.id) AS total, COALESCE(c.capacidade, ?) AS capacidade
		FROM agendamentos a
		LEFT JOIN capacidade_dias c ON c.data = a.data
		WHERE a.data LIKE ?
		GROUP BY a.data, c.capacidade
	`, CapacidadeDiaPadrao, prefix+"%").Scan(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []DiaAgenda{}
	}
	return list, nil
}

// DefinirCapacidadeDia define a capacidade de vagas de um dia específico.
func (a *App) DefinirCapacidadeDia(data string, capacidade int) error {
	if !ehAdmin(a.actorCargo) {
		return fmt.Errorf("permissão negada")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return fmt.Errorf("data inválida: use o formato DD/MM/AAAA")
	}
	if capacidade < 1 {
		return fmt.Errorf("a capacidade deve ser de no mínimo 1 vaga")
	}

	var count int64
	a.db.Model(&Agendamento{}).Where("data = ?", data).Count(&count)
	if count > int64(capacidade) {
		return fmt.Errorf("não foi possível aplicar: %d agendamento(s) já existem, acima da nova capacidade de %d", count, capacidade)
	}

	var cd CapacidadeDia
	err := a.db.Where("data = ?", data).First(&cd).Error
	if err == gorm.ErrRecordNotFound {
		cd = CapacidadeDia{Data: data, Capacidade: capacidade}
		if err := a.db.Create(&cd).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		cd.Capacidade = capacidade
		if err := a.db.Save(&cd).Error; err != nil {
			return err
		}
	}

	return nil
}

// ListarCapacidadesMes retorna as capacidades customizadas do mês.
func (a *App) ListarCapacidadesMes(ano, mes int) ([]CapacidadeDia, error) {
	prefix := fmt.Sprintf("%04d-%02d", ano, mes)
	list := []CapacidadeDia{}
	if err := a.db.Where("data LIKE ?", prefix+"%").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListarAgendamentosDia retorna os agendamentos de um dia específico (formato YYYY-MM-DD).
func (a *App) ListarAgendamentosDia(data string) ([]Agendamento, error) {
	list := []Agendamento{}
	err := a.db.Where("data = ?", data).Order("turno, id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []Agendamento{}
	}
	return list, nil
}

// AdicionarAgendamento reserva uma vaga para o dia. alunoID > 0 vincula um aluno cadastrado;
// alunoID == 0 cadastra um aluno novo apenas na agenda (nome digitado).
func (a *App) AdicionarAgendamento(data string, alunoID uint, nome, turno string) (*Agendamento, error) {
	if !ehAdmin(a.actorCargo) {
		return nil, fmt.Errorf("permissão negada")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return nil, fmt.Errorf("data inválida: use o formato DD/MM/AAAA")
	}
	if !turnoValido(turno) {
		return nil, fmt.Errorf("turno inválido: escolha Manhã, Tarde ou Noite")
	}

	nome = strings.TrimSpace(nome)
	if alunoID > 0 {
		var s Student
		if err := a.db.First(&s, alunoID).Error; err != nil {
			return nil, fmt.Errorf("aluno não encontrado")
		}
		nome = s.Nome
	} else {
		if nome == "" {
			return nil, fmt.Errorf("informe o nome do novo aluno")
		}
	}

	var count int64
	a.db.Model(&Agendamento{}).Where("data = ?", data).Count(&count)
	capacidade := a.capacidadeDoDia(data)
	if count >= int64(capacidade) {
		return nil, fmt.Errorf("capacidade máxima de %d alunos por dia atingida", capacidade)
	}

	if alunoID > 0 {
		var duplicado int64
		a.db.Model(&Agendamento{}).Where("data = ? AND student_id = ?", data, alunoID).Count(&duplicado)
		if duplicado > 0 {
			return nil, fmt.Errorf("este aluno já está agendado neste dia")
		}
	}

	ag := Agendamento{
		Data:      data,
		StudentID: uintPtr(alunoID),
		Nome:      nome,
		Turno:     turno,
	}
	if err := a.db.Create(&ag).Error; err != nil {
		return nil, err
	}
	a.registrarAuditComDetalhes("agendou_aluno", ag.ID,
		fmt.Sprintf("data=%s, aluno=%s, student_id=%d, turno=%s", data, nome, alunoID, turno))
	return &ag, nil
}

// RemoverAgendamento libera a vaga removendo o agendamento.
func (a *App) RemoverAgendamento(id uint) error {
	if !ehAdmin(a.actorCargo) {
		return fmt.Errorf("permissão negada")
	}
	var ag Agendamento
	if err := a.db.First(&ag, id).Error; err != nil {
		return fmt.Errorf("agendamento não encontrado")
	}
	if err := a.db.Delete(&ag).Error; err != nil {
		return err
	}
	a.registrarAuditComDetalhes("removeu_agendamento", id,
		fmt.Sprintf("data=%s, aluno=%s", ag.Data, ag.Nome))
	return nil
}
