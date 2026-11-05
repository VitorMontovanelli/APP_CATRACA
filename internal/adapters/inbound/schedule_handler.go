package inbound

import (
	"context"
	"fmt"
	"strings"
	"time"

	"catraca-app/internal/models"
	"catraca-app/internal/scheduling/domain"
	"gorm.io/gorm"
)

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
	return capacidadeDoCalendario(h.db, domain.CalendarioPilates, data)
}

func (h *ScheduleHandler) ObterConfiguracaoAgenda() (*models.ConfigAgenda, error) {
	return h.ObterConfiguracaoAgendaCalendario(domain.CalendarioPilates)
}

func (h *ScheduleHandler) ListarAgendamentosMes(ano, mes int) ([]models.DiaAgenda, error) {
	prefix := fmt.Sprintf("%04d-%02d", ano, mes)
	list := []models.DiaAgenda{}
	err := h.db.Raw(`
		SELECT a.data, COUNT(a.id) AS total, COALESCE(c.capacidade, ?) AS capacidade
		FROM agendamentos a
		LEFT JOIN capacidade_dias c ON c.data = a.data AND c.calendario = ?
		WHERE a.data LIKE ?
		GROUP BY a.data, c.capacidade
	`, domain.CapacidadeDiaPadrao, domain.CalendarioPilates, prefix+"%").Scan(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.DiaAgenda{}
	}
	return list, nil
}

func (h *ScheduleHandler) DefinirCapacidadeDia(data string, capacidade int) error {
	return h.DefinirCapacidadeDiaCalendario(domain.CalendarioPilates, data, capacidade)
}

func (h *ScheduleHandler) ListarCapacidadesMes(ano, mes int) ([]models.CapacidadeDia, error) {
	return h.ListarCapacidadesCalendarioMes(domain.CalendarioPilates, ano, mes)
}

func (h *ScheduleHandler) ListarAgendamentosDia(data string) ([]models.Agendamento, error) {
	list := []models.Agendamento{}
	err := h.db.Where("data = ?", data).Order("CASE WHEN hora = '' THEN 1 ELSE 0 END, hora, id").Find(&list).Error
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

	if err := h.validarCapacidadeEDuplicidade(data, alunoID); err != nil {
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

func (h *ScheduleHandler) AdicionarAgendamentoComHorario(data string, alunoID uint, nome, hora string) (*models.Agendamento, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return nil, fmt.Errorf("data inválida: use o formato DD/MM/AAAA")
	}
	if !domain.HorarioAulaValido(hora) {
		return nil, fmt.Errorf("horário inválido: escolha um dos horários disponíveis")
	}

	nome, err := h.resolverNomeAluno(alunoID, nome)
	if err != nil {
		return nil, err
	}

	if err := h.validarCapacidadeEDuplicidade(data, alunoID); err != nil {
		return nil, err
	}
	if err := h.garantirDiaComCapacidadePadrao(data); err != nil {
		return nil, err
	}

	ag := models.Agendamento{
		Data:      data,
		StudentID: uintPtr(alunoID),
		Nome:      nome,
		Turno:     domain.TurnoDoHorario(hora),
		Hora:      hora,
	}
	if err := h.db.Create(&ag).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAuditComDetalhes("agendou_aluno", ag.ID, fmt.Sprintf("data=%s, aluno=%s, student_id=%d, hora=%s", data, nome, alunoID, hora))
	return &ag, nil
}

func (h *ScheduleHandler) AtualizarHorarioAgendamento(id uint, hora string) (*models.Agendamento, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	if !domain.HorarioAulaValido(hora) {
		return nil, fmt.Errorf("horário inválido: escolha um dos horários disponíveis")
	}
	var ag models.Agendamento
	if err := h.db.First(&ag, id).Error; err != nil {
		return nil, fmt.Errorf("agendamento não encontrado")
	}
	if domain.DiaIndisponivel(ag.Data) {
		return nil, fmt.Errorf("sábados e domingos estão indisponíveis")
	}
	ag.Hora = hora
	ag.Turno = domain.TurnoDoHorario(hora)
	if err := h.db.Save(&ag).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAuditComDetalhes("editou_horario_agendamento", ag.ID, fmt.Sprintf("data=%s, aluno=%s, hora=%s", ag.Data, ag.Nome, hora))
	return &ag, nil
}

func (h *ScheduleHandler) garantirDiaComCapacidadePadrao(data string) error {
	return garantirCapacidadeDiaCalendario(h.db, domain.CalendarioPilates, data, domain.CapacidadeDiaPadrao)
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

func (h *ScheduleHandler) validarCapacidadeEDuplicidade(data string, alunoID uint) error {
	if domain.DiaIndisponivel(data) {
		return fmt.Errorf("sábados e domingos estão indisponíveis para agendamento")
	}

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
