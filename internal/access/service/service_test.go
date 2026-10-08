package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"catraca-app/internal/access/domain"
	"catraca-app/internal/access/service"
)

type mockAccessRepo struct {
	alunos map[int]*domain.CatracaAluno
	logs   []*domain.AccessLogRecord
}

func (m *mockAccessRepo) FindAlunoByBiometria(ctx context.Context, id int) (*domain.CatracaAluno, error) {
	if a, ok := m.alunos[id]; ok {
		return a, nil
	}
	return nil, errors.New("record not found")
}

func (m *mockAccessRepo) SaveAccessLog(ctx context.Context, log *domain.AccessLogRecord) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockAccessRepo) CreateAluno(ctx context.Context, aluno *domain.CatracaAluno) error {
	m.alunos[aluno.IDBiometriaMock] = aluno
	return nil
}

func TestVerificarAcesso(t *testing.T) {
	now := time.Now()
	repo := &mockAccessRepo{
		alunos: map[int]*domain.CatracaAluno{
			1: {
				ID:              1,
				Nome:            "Carlos Ativo",
				Status:          true,
				VencimentoPlano: now.Add(48 * time.Hour),
				IDBiometriaMock: 1,
			},
			2: {
				ID:              2,
				Nome:            "Ana Vencida",
				Status:          true,
				VencimentoPlano: now.Add(-24 * time.Hour),
				IDBiometriaMock: 2,
			},
			3: {
				ID:              3,
				Nome:            "Bruno Inativo",
				Status:          false,
				VencimentoPlano: now.Add(48 * time.Hour),
				IDBiometriaMock: 3,
			},
		},
	}

	svc := service.NewAccessService(repo)
	ctx := context.Background()

	tests := []struct {
		name         string
		biometriaID  int
		wantLiberado bool
		wantMensagem string
	}{
		{"Aluno Ativo e em Dia", 1, true, "Acesso Liberado"},
		{"Aluno com Plano Vencido", 2, false, "Plano Vencido"},
		{"Aluno Inativo", 3, false, "Aluno inativo"},
		{"Aluno Inexistente", 99, false, "Aluno não encontrado"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := svc.VerificarAcesso(ctx, tt.biometriaID)
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if res.Liberado != tt.wantLiberado {
				t.Errorf("Liberado = %v, esperado %v", res.Liberado, tt.wantLiberado)
			}
			if res.Mensagem != tt.wantMensagem {
				t.Errorf("Mensagem = %q, esperado %q", res.Mensagem, tt.wantMensagem)
			}
		})
	}
}
