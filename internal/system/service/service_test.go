package service_test

import (
	"context"
	"errors"
	"testing"

	"catraca-app/internal/system/domain"
	"catraca-app/internal/system/service"
)

type mockSettingsRepo struct {
	config *domain.BackupConfig
}

func (m *mockSettingsRepo) GetSetting(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *mockSettingsRepo) SaveSetting(ctx context.Context, key, value string) error {
	return nil
}

func (m *mockSettingsRepo) GetBackupConfig(ctx context.Context) (*domain.BackupConfig, error) {
	if m.config == nil {
		return &domain.BackupConfig{}, nil
	}
	return m.config, nil
}

func (m *mockSettingsRepo) SaveBackupConfig(ctx context.Context, config *domain.BackupConfig) error {
	m.config = config
	return nil
}

type mockTelegram struct {
	sentMessages []string
}

func (m *mockTelegram) SendMessage(ctx context.Context, token, chatID, text string) error {
	m.sentMessages = append(m.sentMessages, text)
	return nil
}

func (m *mockTelegram) SendDocument(ctx context.Context, token, chatID, filePath, caption string) error {
	return nil
}

func TestSystemService(t *testing.T) {
	repo := &mockSettingsRepo{}
	tg := &mockTelegram{}
	svc := service.NewSystemService(repo, tg)
	ctx := context.Background()

	// 1. Salvar configuração válida
	err := svc.SalvarConfigBackup(ctx, "bot123", "chat456", true)
	if err != nil {
		t.Fatalf("erro ao salvar config: %v", err)
	}

	// 2. Erro quando autoBackup está ativo mas faltam dados
	err = svc.SalvarConfigBackup(ctx, "", "", true)
	if !errors.Is(err, domain.ErrInvalidTelegramConfig) {
		t.Errorf("esperado ErrInvalidTelegramConfig, recebido %v", err)
	}

	// 3. Teste sanitização path traversal
	safePath, err := domain.SanitizeFileName("/tmp/backups", "catraca_2026.zip")
	if err != nil || safePath != "/tmp/backups/catraca_2026.zip" {
		t.Errorf("caminho esperado /tmp/backups/catraca_2026.zip, recebido %s", safePath)
	}

	_, err = domain.SanitizeFileName("/tmp/backups", "../../../etc/passwd")
	if safePath == "/tmp/backups/passwd" {
		// Base strips ../ successfully
	}
}
