package domain

import (
	"time"
)

// AccessResult representa a decisão final do motor de regras de acesso.
type AccessResult struct {
	Nome     string `json:"nome"`
	Liberado bool   `json:"liberado"`
	Mensagem string `json:"mensagem"`
}

// CatracaAluno representa os dados essenciais de um aluno para validação de acesso.
type CatracaAluno struct {
	ID              uint
	Nome            string
	CPF             string
	Status          bool
	VencimentoPlano time.Time
	IDBiometriaMock int
}

// AccessLogRecord representa o registro de auditoria de uma tentativa de passagem.
type AccessLogRecord struct {
	AlunoID  uint
	DataHora time.Time
	Liberado bool
	Motivo   string
}

// EvaluateAccess é uma regra de domínio pura que determina se a catraca deve liberar o acesso.
func EvaluateAccess(aluno *CatracaAluno, now time.Time) (AccessResult, string) {
	if aluno == nil {
		return AccessResult{
			Nome:     "",
			Liberado: false,
			Mensagem: "Aluno não encontrado",
		}, "Aluno não encontrado"
	}

	if !aluno.Status {
		return AccessResult{
			Nome:     aluno.Nome,
			Liberado: false,
			Mensagem: "Aluno inativo",
		}, "Aluno inativo"
	}

	if now.After(aluno.VencimentoPlano) {
		return AccessResult{
			Nome:     aluno.Nome,
			Liberado: false,
			Mensagem: "Plano Vencido",
		}, "Plano Vencido"
	}

	return AccessResult{
		Nome:     aluno.Nome,
		Liberado: true,
		Mensagem: "Acesso Liberado",
	}, "Acesso Liberado"
}
