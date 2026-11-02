package inbound

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

const LimiteIndividualPadrao = 10
const chaveLimiteIndividual = "agenda_individual_limite"

var horariosIndividuais = []string{
	"06:00", "07:00", "08:00", "09:00", "10:00",
	"15:00", "16:00", "17:00", "18:00",
	"19:00",
}

var praticasIndividuais = []string{
	models.PraticaVentosaterapia,
	models.PraticaLiberacaoMiofascial,
	models.PraticaPersonalTrainer,
	models.PraticaKinesioTape,
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

type ScheduleIndividualHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewScheduleIndividualHandler(db *gorm.DB, session *SessionManager) *ScheduleIndividualHandler {
	return &ScheduleIndividualHandler{db: db, session: session}
}

func (h *ScheduleIndividualHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *ScheduleIndividualHandler) limiteIndividual() int {
	var s models.Setting
	if err := h.db.Where("`key` = ?", chaveLimiteIndividual).First(&s).Error; err != nil {
		return LimiteIndividualPadrao
	}
	n, err := strconv.Atoi(s.Value)
	if err != nil || n < 1 || n > LimiteIndividualPadrao {
		return LimiteIndividualPadrao
	}
	return n
}

func (h *ScheduleIndividualHandler) ObterConfigAgendaIndividual() (*models.ConfigAgendaIndividual, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	return &models.ConfigAgendaIndividual{
		LimiteDiario: h.limiteIndividual(),
		LimiteMaximo: LimiteIndividualPadrao,
	}, nil
}

func (h *ScheduleIndividualHandler) DefinirLimiteDiarioIndividual(limite int) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	if limite < 1 || limite > LimiteIndividualPadrao {
		return fmt.Errorf("o limite diário deve estar entre 1 e %d atendimentos", LimiteIndividualPadrao)
	}
	var s models.Setting
	err := h.db.Where("`key` = ?", chaveLimiteIndividual).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		s = models.Setting{Key: chaveLimiteIndividual, Value: strconv.Itoa(limite)}
		if err := h.db.Create(&s).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		s.Value = strconv.Itoa(limite)
		if err := h.db.Save(&s).Error; err != nil {
			return err
		}
	}
	h.session.RegistrarAuditComDetalhes("ajustou_limite_agenda_individual", 0, fmt.Sprintf("limite_diario=%d", limite))
	return nil
}

func (h *ScheduleIndividualHandler) ListarAgendamentosIndividuaisMes(ano, mes int) ([]models.DiaIndividualAgenda, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	prefix := fmt.Sprintf("%04d-%02d", ano, mes)
	list := []models.DiaIndividualAgenda{}
	limite := h.limiteIndividual()
	err := h.db.Raw(`
		SELECT COALESCE(i.data, c.data) AS data,
		       COUNT(i.id) AS total,
		       COALESCE(c.capacidade, ?) AS capacidade,
		       COALESCE(GROUP_CONCAT(i.pratica), '') AS praticas
		FROM agendamento_individuals i
		LEFT JOIN capacidade_dias c ON c.data = i.data
		WHERE COALESCE(i.data, c.data) LIKE ?
		GROUP BY COALESCE(i.data, c.data)
		UNION
		SELECT c.data AS data, 0 AS total, c.capacidade AS capacidade, '' AS praticas
		FROM capacidade_dias c
		WHERE c.data LIKE ? AND c.data NOT IN (SELECT DISTINCT data FROM agendamento_individuals WHERE data LIKE ?)
	`, limite, prefix+"%", prefix+"%", prefix+"%").Scan(&list).Error
	if err != nil {
		return nil, err
	}
	h.preencherContagensPratica(prefix, list)
	return list, nil
}

func (h *ScheduleIndividualHandler) preencherContagensPratica(prefix string, list []models.DiaIndividualAgenda) {
	type contagemPratica struct {
		Data    string
		Pratica string
		Total   int
	}
	var contagens []contagemPratica
	h.db.Raw(`SELECT data, pratica, COUNT(*) AS total FROM agendamento_individuals WHERE data LIKE ? GROUP BY data, pratica`, prefix+"%").Scan(&contagens)
	porDia := map[string][]string{}
	for _, c := range contagens {
		porDia[c.Data] = append(porDia[c.Data], fmt.Sprintf("%s:%d", c.Pratica, c.Total))
	}
	for i := range list {
		list[i].Contagens = strings.Join(porDia[list[i].Data], ",")
	}
}

func (h *ScheduleIndividualHandler) ListarAgendamentosIndividuaisDia(data string) ([]models.AgendamentoIndividual, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return nil, fmt.Errorf("data inválida")
	}
	list := []models.AgendamentoIndividual{}
	err := h.db.Where("data = ?", data).Order("hora, id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.AgendamentoIndividual{}
	}
	return list, nil
}
