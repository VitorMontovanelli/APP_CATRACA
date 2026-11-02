package inbound

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

const (
	keyTelegramToken = "telegram_token"
	keyTelegramChat  = "telegram_chat_id"
	keyTelegramAuto  = "telegram_auto"
	keyTelegramLast  = "telegram_last_backup"
)

type SystemTelegramHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
	backup  *SystemBackupHandler
}

func NewSystemTelegramHandler(db *gorm.DB, session *SessionManager, backup *SystemBackupHandler) *SystemTelegramHandler {
	return &SystemTelegramHandler{db: db, session: session, backup: backup}
}

func (h *SystemTelegramHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *SystemTelegramHandler) settingValue(key string) string {
	var v string
	h.db.Model(&models.Setting{}).Where("key = ?", key).Pluck("value", &v)
	return v
}

func (h *SystemTelegramHandler) saveSetting(key, value string) {
	h.db.Exec("INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value)
}

func (h *SystemTelegramHandler) ObterConfigTelegram() (*models.BackupConfig, error) {
	if h.session.ActorCargo() != "super_admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	cfg := &models.BackupConfig{
		Token:      h.settingValue(keyTelegramToken),
		ChatID:     h.settingValue(keyTelegramChat),
		LastBackup: h.settingValue(keyTelegramLast),
	}
	cfg.AutoBackup = h.settingValue(keyTelegramAuto) == "1"
	return cfg, nil
}

func (h *SystemTelegramHandler) SalvarConfigTelegram(token, chatID string, autoBackup bool) error {
	if h.session.ActorCargo() != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	token = strings.TrimSpace(token)
	chatID = strings.TrimSpace(chatID)
	if token == "" || chatID == "" {
		return fmt.Errorf("preencha o token do bot e o chat ID")
	}
	auto := "0"
	if autoBackup {
		auto = "1"
	}
	h.saveSetting(keyTelegramToken, token)
	h.saveSetting(keyTelegramChat, chatID)
	h.saveSetting(keyTelegramAuto, auto)
	h.session.RegistrarAudit("configurou_backup_telegram", 0)
	return nil
}

func (h *SystemTelegramHandler) TestarTelegram() error {
	if h.session.ActorCargo() != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	token := h.settingValue(keyTelegramToken)
	chatID := h.settingValue(keyTelegramChat)
	if token == "" || chatID == "" {
		return fmt.Errorf("token ou chat ID não configurado")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	if err := validarTokenBot(client, token); err != nil {
		return err
	}
	return validarChatID(client, token, chatID)
}

func validarTokenBot(client *http.Client, token string) error {
	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	resp, err := client.Get("https://api.telegram.org/bot" + token + "/getMe")
	if err != nil {
		return fmt.Errorf("erro ao conectar com o Telegram: %w", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || !result.Ok {
		return fmt.Errorf("token inválido: %s", result.Description)
	}
	return nil
}

func validarChatID(client *http.Client, token, chatID string) error {
	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	resp, err := client.Get("https://api.telegram.org/bot" + token + "/getChat?chat_id=" + url.QueryEscape(chatID))
	if err != nil {
		return fmt.Errorf("erro ao validar o chat: %w", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || !result.Ok {
		return fmt.Errorf("chat ID inválido: %s", result.Description)
	}
	return nil
}

func (h *SystemTelegramHandler) EnviarBackupTelegram() (string, error) {
	if h.session.ActorCargo() != "super_admin" {
		return "", fmt.Errorf("permissão negada")
	}
	return h.enviarBackupTelegram()
}

func (h *SystemTelegramHandler) enviarBackupTelegram() (string, error) {
	token := h.settingValue(keyTelegramToken)
	chatID := h.settingValue(keyTelegramChat)
	if token == "" || chatID == "" {
		return "", fmt.Errorf("configure o token do bot e o chat ID antes de enviar o backup")
	}

	zipPath, err := h.backup.gerarArquivoBackup()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(filepath.Dir(zipPath))

	caption := fmt.Sprintf("Backup CatracaVMD - %s", time.Now().Format("02/01/2006 15:04"))
	if err := enviarDocumentoTelegram(token, chatID, zipPath, caption); err != nil {
		return "", err
	}

	now := time.Now().Format(time.RFC3339)
	h.saveSetting(keyTelegramLast, now)
	h.session.RegistrarAudit("enviou_backup_telegram", 0)
	return now, nil
}

func (h *SystemTelegramHandler) OnShutdown() {
	if h.settingValue(keyTelegramAuto) == "1" {
		_, _ = h.enviarBackupTelegram()
	}
}
