package ports

import (
	"context"

	"catraca-app/internal/system/domain"
)

type SettingsRepository interface {
	GetSetting(ctx context.Context, key string) (string, error)
	SaveSetting(ctx context.Context, key, value string) error
	GetBackupConfig(ctx context.Context) (*domain.BackupConfig, error)
	SaveBackupConfig(ctx context.Context, config *domain.BackupConfig) error
}

type TelegramPort interface {
	SendMessage(ctx context.Context, token, chatID, text string) error
	SendDocument(ctx context.Context, token, chatID, filePath, caption string) error
}

type SystemServicePort interface {
	ObterConfigBackup(ctx context.Context) (*domain.BackupConfig, error)
	SalvarConfigBackup(ctx context.Context, token, chatID string, autoBackup bool) error
}
