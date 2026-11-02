package inbound

import (
	"context"
	"fmt"
	"strings"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

const CapacidadeDiaPadrao = 10

type ScheduleHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
	*ScheduleIndividualHandler
	*ScheduleIndividualCrudHandler
}

func NewScheduleHandler(db *gorm.DB, session *SessionManager) *ScheduleHandler {
	indHandler := NewScheduleIndividualHandler(db, session)
	crudHandler := NewScheduleIndividualCrudHandler(db, session)
	return &ScheduleHandler{
		db:                            db,
		session:                       session,
		ScheduleIndividualHandler:     indHandler,
		ScheduleIndividualCrudHandler: crudHandler,
	}
}

func (h *ScheduleHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
	h.ScheduleIndividualHandler.SetContext(ctx)
	h.ScheduleIndividualCrudHandler.SetContext(ctx)
}

func turnoValido(turno string) bool {
	return turno == "manha" || turno == "tarde" || turno == "noite"
}

func (h *ScheduleHandler) capacidadeDoDia(data string) int {
	var cd models.CapacidadeDia
	if err := h.db.Where("data = ?", data).First(&cd).Error; err != nil {
		return CapacidadeDiaPadrao
	}
	return cd.Capacidade
}

func (h *ScheduleHandler) ListarAgendamentosMes(ano, mes int) ([]models.DiaAgenda, error) {
	prefix := fmt.Sprintf("%04d-%02d", ano, mes)
	list := []models.DiaAgenda{}
	err := h.db.Raw(`
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
		list = []models.DiaAgenda{}
	}
	return list, nil
}

func (h *ScheduleHandler) DefinirCapacidadeDia(data string, capacidade int) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return fmt.Errorf("data inválida: use o formato DD/MM/AAAA")
	}
	if capacidade < 1 {
		return fmt.Errorf("a capacidade deve ser de no mínimo 1 vaga")
	}

	var count int64
	h.db.Model(&models.Agendamento{}).Where("data = ?", data).Count(&count)
	if count > int64(capacidade) {
		return fmt.Errorf("não foi possível aplicar: %d agendamento(s) já existem, acima da nova capacidade de %d", count, capacidade)
	}

	var cd models.CapacidadeDia
	err := h.db.Where("data = ?", data).First(&cd).Error
	if err == gorm.ErrRecordNotFound {
		return h.db.Create(&models.CapacidadeDia{Data: data, Capacidade: capacidade}).Error
	} else if err != nil {
		return err
	}
	cd.Capacidade = capacidade
	return h.db.Save(&cd).Error
}

func (h *ScheduleHandler) ListarCapacidadesMes(ano, mes int) ([]models.CapacidadeDia, error) {
	prefix := fmt.Sprintf("%04d-%02d", ano, mes)
	list := []models.CapacidadeDia{}
	if err := h.db.Where("data LIKE ?", prefix+"%").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (h *ScheduleHandler) ListarAgendamentosDia(data string) ([]models.Agendamento, error) {
	list := []models.Agendamento{}
	err := h.db.Where("data = ?", data).Order("turno, id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.Agendamento{}
	}
	return list, nil
}

func (h *ScheduleHandler) AdicionarAgendamento(data string, alunoID uint, nome, turno string) (*models.Agendamento, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return nil, fmt.Errorf("data inválida: use o formato DD/MM/AAAA")
	}
	if !turnoValido(turno) {
		return nil, fmt.Errorf("turno inválido: escolha Manhã, Tarde ou Noite")
	}

	nome, err := h.resolverNomeAluno(alunoID, nome)
	if err != nil {
		return nil, err
	}

	if err := h.validarCapacidadeEDUplicidade(data, alunoID); err != nil {
		return nil, err
	}

	ag := models.Agendamento{
		Data:      data,
		StudentID: uintPtr(alunoID),
		Nome:      nome,
		Turno:     turno,
	}
	if err := h.db.Create(&ag).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAuditComDetalhes("agendou_aluno", ag.ID, fmt.Sprintf("data=%s, aluno=%s, student_id=%d, turno=%s", data, nome, alunoID, turno))
	return &ag, nil
}

func (h *ScheduleHandler) resolverNomeAluno(alunoID uint, nome string) (string, error) {
	nome = strings.TrimSpace(nome)
	if alunoID > 0 {
		var s models.Student
		if err := h.db.First(&s, alunoID).Error; err != nil {
			return "", fmt.Errorf("aluno não encontrado")
		}
		return s.Nome, nil
	}
	if nome == "" {
		return "", fmt.Errorf("informe o nome do novo aluno")
	}
	return nome, nil
}

func (h *ScheduleHandler) validarCapacidadeEDUplicidade(data string, alunoID uint) error {
	var count int64
	h.db.Model(&models.Agendamento{}).Where("data = ?", data).Count(&count)
	capacidade := h.capacidadeDoDia(data)
	if count >= int64(capacidade) {
		return fmt.Errorf("capacidade máxima de %d alunos por dia atingida", capacidade)
	}

	if alunoID > 0 {
		var duplicado int64
		h.db.Model(&models.Agendamento{}).Where("data = ? AND student_id = ?", data, alunoID).Count(&duplicado)
		if duplicado > 0 {
			return fmt.Errorf("este aluno já está agendado neste dia")
		}
	}
	return nil
}

func (h *ScheduleHandler) RemoverAgendamento(id uint) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	var ag models.Agendamento
	if err := h.db.First(&ag, id).Error; err != nil {
		return fmt.Errorf("agendamento não encontrado")
	}
	if err := h.db.Delete(&ag).Error; err != nil {
		return err
	}
	h.session.RegistrarAuditComDetalhes("removeu_agendamento", id, fmt.Sprintf("data=%s, aluno=%s", ag.Data, ag.Nome))
	return nil
}
