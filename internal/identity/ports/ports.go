package ports

import (
	"context"

	"catraca-app/internal/identity/domain"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uint) (*domain.User, error)
	ListAll(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	UpdatePassword(ctx context.Context, id uint, hash string) error
	Delete(ctx context.Context, id uint) error
	SaveAudit(ctx context.Context, log *domain.AuditLogRecord) error
}

type AuthServicePort interface {
	Login(ctx context.Context, email, password string) (*domain.User, error)
	AlterarSenha(ctx context.Context, userID uint, senhaAtual, novaSenha string) error
	RegistrarAuditoria(ctx context.Context, actorID uint, action string, targetID *uint, details string) error
}
