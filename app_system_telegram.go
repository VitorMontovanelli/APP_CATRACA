package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// =============== Backup Telegram ===============

const (
	keyTelegramToken = "telegram_token"
	keyTelegramChat  = "telegram_chat_id"
	keyTelegramAuto  = "telegram_auto"
	keyTelegramLast  = "telegram_last_backup"
)

func (a *App) settingValue(key string) string {
	var v string
	a.db.Model(&Setting{}).Where("key = ?", key).Pluck("value", &v)
	return v
}

func (a *App) saveSetting(key, value string) {
	a.db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
}

func (a *App) ObterConfigTelegram() (*BackupConfig, error) {
	if a.actorCargo != "super_admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	cfg := &BackupConfig{
		Token:      a.settingValue(keyTelegramToken),
		ChatID:     a.settingValue(keyTelegramChat),
		LastBackup: a.settingValue(keyTelegramLast),
	}
	cfg.AutoBackup = a.settingValue(keyTelegramAuto) == "1"
	return cfg, nil
}

func (a *App) SalvarConfigTelegram(token, chatID string, autoBackup bool) error {
	if a.actorCargo != "super_admin" {
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
	a.saveSetting(keyTelegramToken, token)
	a.saveSetting(keyTelegramChat, chatID)
	a.saveSetting(keyTelegramAuto, auto)
	a.registrarAudit("configurou_backup_telegram", 0)
	return nil
}

func (a *App) TestarTelegram() error {
	if a.actorCargo != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	token := a.settingValue(keyTelegramToken)
	chatID := a.settingValue(keyTelegramChat)
	if token == "" {
		return fmt.Errorf("token do bot não configurado")
	}
	if chatID == "" {
		return fmt.Errorf("chat ID não configurado")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}

	resp, err := client.Get("https://api.telegram.org/bot" + token + "/getMe")
	if err != nil {
		return fmt.Errorf("erro ao conectar com o Telegram: %w", err)
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("resposta inválida do Telegram: %w", err)
	}
	if !result.Ok {
		return fmt.Errorf("token inválido: %s", result.Description)
	}

	resp, err = client.Get("https://api.telegram.org/bot" + token + "/getChat?chat_id=" + url.QueryEscape(chatID))
	if err != nil {
		return fmt.Errorf("erro ao validar o chat: %w", err)
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("resposta inválida do Telegram: %w", err)
	}
	if !result.Ok {
		return fmt.Errorf("chat ID inválido: %s", result.Description)
	}
	return nil
}

func (a *App) EnviarBackupTelegram() (string, error) {
	if a.actorCargo != "super_admin" {
		return "", fmt.Errorf("permissão negada")
	}
	return a.enviarBackupTelegram()
}

func (a *App) enviarBackupTelegram() (string, error) {
	token := a.settingValue(keyTelegramToken)
	chatID := a.settingValue(keyTelegramChat)
	if token == "" || chatID == "" {
		return "", fmt.Errorf("configure o token do bot e o chat ID antes de enviar o backup")
	}

	zipPath, err := a.gerarArquivoBackup()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(filepath.Dir(zipPath))

	caption := fmt.Sprintf("Backup CatracaVMD - %s", time.Now().Format("02/01/2006 15:04"))
	if err := enviarDocumentoTelegram(token, chatID, zipPath, caption); err != nil {
		return "", err
	}

	now := time.Now().Format(time.RFC3339)
	a.saveSetting(keyTelegramLast, now)
	a.registrarAudit("enviou_backup_telegram", 0)
	return now, nil
}
