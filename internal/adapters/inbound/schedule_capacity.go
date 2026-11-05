package inbound

import (
	"fmt"
	"time"

	"catraca-app/internal/models"
	"catraca-app/internal/scheduling/domain"
	"gorm.io/gorm"
)

// capacidadeDoCalendario retorna a capacidade de um dia em um calendário
// específico, usando o padrão do calendário quando não há registro.
func capacidadeDoCalendario(db *gorm.DB, calendario, data string) int {
	var cd models.CapacidadeDia
	if err := db.Where("data = ? AND calendario = ?", data, calendario).First(&cd).Error; err != nil {
		return domain.CapacidadePadraoDoCalendario(calendario)
	}
	return cd.Capacidade
}

// garantirCapacidadeDiaCalendario cria o registro do dia com a capacidade
// informada quando ele ainda não existe para o calendário.
func garantirCapacidadeDiaCalendario(db *gorm.DB, calendario, data string, capacidade int) error {
	var count int64
	if err := db.Model(&models.CapacidadeDia{}).
		Where("data = ? AND calendario = ?", data, calendario).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&models.CapacidadeDia{
		Data:       data,
		Calendario: calendario,
		Capacidade: capacidade,
	}).Error
}

// contarAgendamentosCalendario conta os agendamentos do dia no calendário.
func contarAgendamentosCalendario(db *gorm.DB, calendario, data string) (int64, error) {
	var count int64
	if calendario == domain.CalendarioPilates {
		err := db.Model(&models.Agendamento{}).Where("data = ?", data).Count(&count).Error
		return count, err
	}
	err := db.Model(&models.AgendamentoIndividual{}).
		Where("data = ? AND pratica = ?", data, calendario).
		Count(&count).Error
	return count, err
}

// ObterConfiguracaoAgendaCalendario devolve a configuração de um calendário:
// capacidade padrão, limite máximo e grade de horários.
func (h *ScheduleHandler) ObterConfiguracaoAgendaCalendario(calendario string) (*models.ConfigAgenda, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	if !domain.CalendarioValido(calendario) {
		return nil, fmt.Errorf("calendário inválido")
	}
	horarios := make([]string, len(domain.HorariosAula))
	copy(horarios, domain.HorariosAula)
	return &models.ConfigAgenda{
		Calendario:       calendario,
		CapacidadePadrao: domain.CapacidadePadraoDoCalendario(calendario),
		CapacidadeMaxima: domain.CapacidadeMaximaDoCalendario(calendario),
		Horarios:         horarios,
	}, nil
}

// ListarCapacidadesCalendarioMes lista as capacidades explícitas do mês para
// um calendário. Dias sem registro usam a capacidade padrão no frontend.
func (h *ScheduleHandler) ListarCapacidadesCalendarioMes(calendario string, ano, mes int) ([]models.CapacidadeDia, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	if !domain.CalendarioValido(calendario) {
		return nil, fmt.Errorf("calendário inválido")
	}
	prefix := fmt.Sprintf("%04d-%02d", ano, mes)
	list := []models.CapacidadeDia{}
	if err := h.db.Where("calendario = ? AND data LIKE ?", calendario, prefix+"%").
		Order("data").Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.CapacidadeDia{}
	}
	return list, nil
}

// DefinirCapacidadeDiaCalendario define a capacidade de um dia apenas para o
// calendário informado, sem afetar os demais.
func (h *ScheduleHandler) DefinirCapacidadeDiaCalendario(calendario, data string, capacidade int) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	if !domain.CalendarioValido(calendario) {
		return fmt.Errorf("calendário inválido")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return fmt.Errorf("data inválida: use o formato DD/MM/AAAA")
	}
	if domain.DiaIndisponivel(data) {
		return fmt.Errorf("sábados e domingos estão indisponíveis")
	}
	if capacidade < 1 {
		return fmt.Errorf("a capacidade deve ser de no mínimo 1 vaga")
	}
	if maximo := domain.CapacidadeMaximaDoCalendario(calendario); maximo > 0 && capacidade > maximo {
		return fmt.Errorf("a capacidade deve ser de no máximo %d vagas para este calendário", maximo)
	}

	count, err := contarAgendamentosCalendario(h.db, calendario, data)
	if err != nil {
		return err
	}
	if count > int64(capacidade) {
		return fmt.Errorf("não foi possível aplicar: %d agendamento(s) já existem, acima da nova capacidade de %d", count, capacidade)
	}

	var cd models.CapacidadeDia
	err = h.db.Where("data = ? AND calendario = ?", data, calendario).First(&cd).Error
	if err == gorm.ErrRecordNotFound {
		err = h.db.Create(&models.CapacidadeDia{
			Data:       data,
			Calendario: calendario,
			Capacidade: capacidade,
		}).Error
	} else if err == nil {
		cd.Capacidade = capacidade
		err = h.db.Save(&cd).Error
	}
	if err != nil {
		return err
	}
	h.session.RegistrarAuditComDetalhes("ajustou_capacidade_dia", 0,
		fmt.Sprintf("calendario=%s, data=%s, capacidade=%d", calendario, data, capacidade))
	return nil
}
