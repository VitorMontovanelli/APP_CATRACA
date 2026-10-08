package service_test

import (
	"context"
	"errors"
	"testing"

	"catraca-app/internal/identity/domain"
	"catraca-app/internal/identity/service"
)

type mockUserRepo struct {
	users map[uint]*domain.User
	audits []*domain.AuditLogRecord
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[uint]*domain.User),
	}
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	var list []domain.User
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, nil
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uint, hash string) error {
	if u, ok := m.users[id]; ok {
		u.SenhaHash = hash
		return nil
	}
	return errors.New("not found")
}

func (m *mockUserRepo) Delete(ctx context.Context, id uint) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) SaveAudit(ctx context.Context, log *domain.AuditLogRecord) error {
	m.audits = append(m.audits, log)
	return nil
}

func TestAuthService(t *testing.T) {
	repo := newMockUserRepo()
	hash, _ := domain.HashPassword("senhaSegura123")
	repo.users[1] = &domain.User{
		ID:         1,
		Name:       "Admin",
		Email:      "admin@catraca.com",
		SenhaHash:  hash,
		Cargo:      "super_admin",
		Ativo:      true,
		Permissoes: "home,alunos,usuarios",
	}
	repo.users[2] = &domain.User{
		ID:        2,
		Name:      "Inativo",
		Email:     "inativo@catraca.com",
		SenhaHash: hash,
		Cargo:     "admin",
		Ativo:     false,
	}

	svc := service.NewAuthService(repo)
	ctx := context.Background()

	// 1. Login com sucesso
	u, err := svc.Login(ctx, "admin@catraca.com", "senhaSegura123")
	if err != nil || u == nil {
		t.Fatalf("esperado login com sucesso, recebido err=%v", err)
	}

	// 2. Senha errada
	_, err = svc.Login(ctx, "admin@catraca.com", "senhaErrada")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("esperado ErrInvalidCredentials, recebido %v", err)
	}

	// 3. Usuário inativo
	_, err = svc.Login(ctx, "inativo@catraca.com", "senhaSegura123")
	if !errors.Is(err, domain.ErrUserInactive) {
		t.Errorf("esperado ErrUserInactive, recebido %v", err)
	}

	// 4. Auditoria
	err = svc.RegistrarAuditoria(ctx, 1, "login_realizado", nil, "IP 127.0.0.1")
	if err != nil || len(repo.audits) != 1 {
		t.Errorf("esperado 1 auditoria gravada, erro: %v", err)
	}
}
