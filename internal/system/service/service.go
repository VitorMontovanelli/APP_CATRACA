package service

import (
	"context"

	"catraca-app/internal/system/domain"
	"catraca-app/internal/system/ports"
)

type SystemService struct {
	repo     ports.SettingsRepository
	telegram ports.TelegramPort
}

func NewSystemService(repo ports.SettingsRepository, telegram ports.TelegramPort) *SystemService {
	return &SystemService{
		repo:     repo,
		telegram: telegram,
	}
}

func (s *SystemService) ObterConfigBackup(ctx context.Context) (*domain.BackupConfig, error) {
	return s.repo.GetBackupConfig(ctx)
}

func (s *SystemService) SalvarConfigBackup(ctx context.Context, token, chatID string, autoBackup bool) error {
	config := &domain.BackupConfig{
		Token:      token,
		ChatID:     chatID,
		AutoBackup: autoBackup,
	}
	if err := config.Validate(); err != nil {
		return err
	}
	return s.repo.SaveBackupConfig(ctx, config)
}

func (s *SystemService) TestarTelegram(ctx context.Context, token, chatID string) error {
	if s.telegram == nil {
		return nil
	}
	return s.telegram.SendMessage(ctx, token, chatID, "🤖 Teste de Notificação - Catraca App OK!")
}
