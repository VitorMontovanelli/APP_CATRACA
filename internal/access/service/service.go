package service

import (
	"context"
	"time"

	"catraca-app/internal/access/domain"
	"catraca-app/internal/access/ports"
)

type AccessService struct {
	repo ports.AccessRepository
}

func NewAccessService(repo ports.AccessRepository) *AccessService {
	return &AccessService{repo: repo}
}

func (s *AccessService) VerificarAcesso(ctx context.Context, idBiometria int) (domain.AccessResult, error) {
	now := time.Now()
	aluno, err := s.repo.FindAlunoByBiometria(ctx, idBiometria)
	if err != nil {
		// Loga a falha por não encontrar o aluno
		result, motivo := domain.EvaluateAccess(nil, now)
		_ = s.repo.SaveAccessLog(ctx, &domain.AccessLogRecord{
			AlunoID:  0,
			DataHora: now,
			Liberado: false,
			Motivo:   motivo,
		})
		return result, nil
	}

	result, motivo := domain.EvaluateAccess(aluno, now)
	_ = s.repo.SaveAccessLog(ctx, &domain.AccessLogRecord{
		AlunoID:  aluno.ID,
		DataHora: now,
		Liberado: result.Liberado,
		Motivo:   motivo,
	})

	return result, nil
}

func (s *AccessService) CadastrarAluno(ctx context.Context, nome, cpf string, diasValidade int, idBiometria int) error {
	vencimento := time.Now().AddDate(0, 0, diasValidade)
	aluno := &domain.CatracaAluno{
		Nome:            nome,
		CPF:             cpf,
		Status:          true,
		VencimentoPlano: vencimento,
		IDBiometriaMock: idBiometria,
	}
	return s.repo.CreateAluno(ctx, aluno)
}
