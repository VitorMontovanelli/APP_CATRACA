package main

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// =============== Catraca Virtual ===============

// CadastrarAluno cria um novo aluno para a catraca com vencimento calculado automaticamente.
func (a *App) CadastrarAluno(nome, cpf string, diasValidade int, idBiometria int) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	vencimento := time.Now().AddDate(0, 0, diasValidade)
	aluno := Aluno{
		Nome:            nome,
		CPF:             cpf,
		Status:          true,
		VencimentoPlano: vencimento,
		IDBiometriaMock: idBiometria,
	}
	if err := a.db.Create(&aluno).Error; err != nil {
		return err
	}
	a.registrarAudit("cadastrou_aluno_catraca", aluno.ID)
	return nil
}

// VerificarAcessoAluno simula a leitura da digital na catraca e retorna o resultado.
func (a *App) VerificarAcessoAluno(idBiometria int) (ResultadoAcesso, error) {
	var aluno Aluno
	err := a.db.Where("id_biometria_mock = ?", idBiometria).First(&aluno).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			registro := RegistroAcesso{
				AlunoID:  0,
				DataHora: time.Now(),
				Liberado: false,
				Motivo:   "Aluno não encontrado",
			}
			a.db.Create(&registro)
			return ResultadoAcesso{
				Nome:     "",
				Liberado: false,
				Mensagem: "Aluno não encontrado",
			}, nil
		}
		return ResultadoAcesso{}, err
	}

	if !aluno.Status {
		registro := RegistroAcesso{
			AlunoID:  aluno.ID,
			Liberado: false,
			Motivo:   "Aluno inativo",
		}
		a.db.Create(&registro)
		return ResultadoAcesso{
			Nome:     aluno.Nome,
			Liberado: false,
			Mensagem: "Aluno inativo",
		}, nil
	}

	if time.Now().After(aluno.VencimentoPlano) {
		registro := RegistroAcesso{
			AlunoID:  aluno.ID,
			Liberado: false,
			Motivo:   "Plano Vencido",
		}
		a.db.Create(&registro)
		return ResultadoAcesso{
			Nome:     aluno.Nome,
			Liberado: false,
			Mensagem: "Plano Vencido",
		}, nil
	}

	registro := RegistroAcesso{
		AlunoID:  aluno.ID,
		Liberado: true,
		Motivo:   "Acesso Liberado",
	}
	a.db.Create(&registro)
	return ResultadoAcesso{
		Nome:     aluno.Nome,
		Liberado: true,
		Mensagem: "Acesso Liberado",
	}, nil
}

func (a *App) ListarAccessLogs() ([]AccessLog, error) {
	logs := []AccessLog{}
	err := a.db.Order("timestamp DESC").Limit(100).Find(&logs).Error
	return logs, err
}

func (a *App) ObterEstatisticasAcesso(filtro string) ([]AcessoEstatistica, error) {
	var list []AcessoEstatistica
	var query string
	var since time.Time

	now := time.Now()

	switch filtro {
	case "dia":
		// Últimos 7 dias, agrupados por dia
		since = now.AddDate(0, 0, -6) // 7 dias incluindo hoje
		query = `
			SELECT strftime('%d/%m', data_hora) as label,
			       SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			       SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos
			WHERE data_hora >= ?
			GROUP BY strftime('%Y-%m-%d', data_hora)
			ORDER BY data_hora ASC
		`
	case "semana":
		// Últimas 4 semanas, agrupadas por semana
		since = now.AddDate(0, 0, -27) // 28 dias
		query = `
			SELECT 'Sem ' || strftime('%W', data_hora) as label,
			       SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			       SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos
			WHERE data_hora >= ?
			GROUP BY strftime('%Y-%W', data_hora)
			ORDER BY data_hora ASC
		`
	case "mes":
		// Últimos 6 meses, agrupados por mês
		since = now.AddDate(0, -5, 0) // 6 meses
		query = `
			SELECT strftime('%m/%Y', data_hora) as label,
			       SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			       SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos
			WHERE data_hora >= ?
			GROUP BY strftime('%Y-%m', data_hora)
			ORDER BY data_hora ASC
		`
	default:
		return nil, fmt.Errorf("filtro inválido: %s", filtro)
	}

	err := a.db.Raw(query, since).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []AcessoEstatistica{}
	}
	return list, nil
}
