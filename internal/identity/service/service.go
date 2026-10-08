package service

import (
	"context"
	"time"

	"catraca-app/internal/identity/domain"
	"catraca-app/internal/identity/ports"
)

type AuthService struct {
	repo ports.UserRepository
}

func NewAuthService(repo ports.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.Ativo {
		return nil, domain.ErrUserInactive
	}

	if !domain.CheckPasswordHash(password, user.SenhaHash) {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}

func (s *AuthService) AlterarSenha(ctx context.Context, userID uint, senhaAtual, novaSenha string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	if !domain.CheckPasswordHash(senhaAtual, user.SenhaHash) {
		return domain.ErrInvalidCredentials
	}

	newHash, err := domain.HashPassword(novaSenha)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, newHash)
}

func (s *AuthService) RegistrarAuditoria(ctx context.Context, actorID uint, action string, targetID *uint, details string) error {
	log := &domain.AuditLogRecord{
		ActorID:   actorID,
		Action:    action,
		TargetID:  targetID,
		Details:   details,
		Timestamp: time.Now(),
	}
	return s.repo.SaveAudit(ctx, log)
}
