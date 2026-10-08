package ports

import (
	"context"

	"catraca-app/internal/access/domain"
)

// AccessRepository define a porta de saída para persistência de dados de acesso.
type AccessRepository interface {
	FindAlunoByBiometria(ctx context.Context, idBiometria int) (*domain.CatracaAluno, error)
	SaveAccessLog(ctx context.Context, log *domain.AccessLogRecord) error
	CreateAluno(ctx context.Context, aluno *domain.CatracaAluno) error
}

// AccessServicePort define a porta de entrada para casos de uso de catraca.
type AccessServicePort interface {
	VerificarAcesso(ctx context.Context, idBiometria int) (domain.AccessResult, error)
	CadastrarAluno(ctx context.Context, nome, cpf string, diasValidade int, idBiometria int) error
}
