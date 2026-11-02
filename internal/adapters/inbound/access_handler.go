package inbound

import (
	"context"
	"fmt"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type AccessHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewAccessHandler(db *gorm.DB, session *SessionManager) *AccessHandler {
	return &AccessHandler{db: db, session: session}
}

func (h *AccessHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *AccessHandler) CadastrarAluno(nome, cpf string, diasValidade int, idBiometria int) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	vencimento := time.Now().AddDate(0, 0, diasValidade)
	aluno := models.Aluno{
		Nome:            nome,
		CPF:             cpf,
		Status:          true,
		VencimentoPlano: vencimento,
		IDBiometriaMock: idBiometria,
	}
	if err := h.db.Create(&aluno).Error; err != nil {
		return err
	}
	h.session.RegistrarAudit("cadastrou_aluno_catraca", aluno.ID)
	return nil
}

func (h *AccessHandler) VerificarAcessoAluno(idBiometria int) (models.ResultadoAcesso, error) {
	var aluno models.Aluno
	err := h.db.Where("id_biometria_mock = ?", idBiometria).First(&aluno).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			registro := models.RegistroAcesso{
				AlunoID:  0,
				DataHora: time.Now(),
				Liberado: false,
				Motivo:   "Aluno não encontrado",
			}
			h.db.Create(&registro)
			return models.ResultadoAcesso{Nome: "", Liberado: false, Mensagem: "Aluno não encontrado"}, nil
		}
		return models.ResultadoAcesso{}, err
	}

	return h.processarAcesso(&aluno)
}

func (h *AccessHandler) processarAcesso(aluno *models.Aluno) (models.ResultadoAcesso, error) {
	if !aluno.Status {
		h.db.Create(&models.RegistroAcesso{AlunoID: aluno.ID, Liberado: false, Motivo: "Aluno inativo"})
		return models.ResultadoAcesso{Nome: aluno.Nome, Liberado: false, Mensagem: "Aluno inativo"}, nil
	}

	if time.Now().After(aluno.VencimentoPlano) {
		h.db.Create(&models.RegistroAcesso{AlunoID: aluno.ID, Liberado: false, Motivo: "Plano Vencido"})
		return models.ResultadoAcesso{Nome: aluno.Nome, Liberado: false, Mensagem: "Plano Vencido"}, nil
	}

	h.db.Create(&models.RegistroAcesso{AlunoID: aluno.ID, Liberado: true, Motivo: "Acesso Liberado"})
	return models.ResultadoAcesso{Nome: aluno.Nome, Liberado: true, Mensagem: "Acesso Liberado"}, nil
}

func (h *AccessHandler) ListarAccessLogs() ([]models.AccessLog, error) {
	logs := []models.AccessLog{}
	err := h.db.Order("timestamp DESC").Limit(100).Find(&logs).Error
	return logs, err
}

func (h *AccessHandler) ObterEstatisticasAcesso(filtro string) ([]models.AcessoEstatistica, error) {
	query, since, err := buildEstatisticasQuery(filtro)
	if err != nil {
		return nil, err
	}

	var list []models.AcessoEstatistica
	err = h.db.Raw(query, since).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.AcessoEstatistica{}
	}
	return list, nil
}

func buildEstatisticasQuery(filtro string) (string, time.Time, error) {
	now := time.Now()
	switch filtro {
	case "dia":
		return `SELECT strftime('%d/%m', data_hora) as label,
			SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos WHERE data_hora >= ?
			GROUP BY strftime('%Y-%m-%d', data_hora) ORDER BY data_hora ASC`, now.AddDate(0, 0, -6), nil
	case "semana":
		return `SELECT 'Sem ' || strftime('%W', data_hora) as label,
			SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos WHERE data_hora >= ?
			GROUP BY strftime('%Y-%W', data_hora) ORDER BY data_hora ASC`, now.AddDate(0, 0, -27), nil
	case "mes":
		return `SELECT strftime('%m/%Y', data_hora) as label,
			SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos WHERE data_hora >= ?
			GROUP BY strftime('%Y-%m', data_hora) ORDER BY data_hora ASC`, now.AddDate(0, -5, 0), nil
	default:
		return "", now, fmt.Errorf("filtro inválido: %s", filtro)
	}
}
