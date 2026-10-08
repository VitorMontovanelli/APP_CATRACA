package domain

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidTelegramConfig = errors.New("token e chat_id são obrigatórios para habilitar backup")
	ErrPathTraversal         = errors.New("caminho de arquivo inválido ou tentativa de path traversal")
)

type BackupConfig struct {
	Token      string `json:"token"`
	ChatID     string `json:"chat_id"`
	AutoBackup bool   `json:"auto_backup"`
	LastBackup string `json:"last_backup"`
	UpdatedAt  string `json:"updated_at"`
}

type Setting struct {
	Key   string
	Value string
}

// SanitizeFileName remove caracteres perigosos e previne path traversal em nomes de arquivos gerados.
func SanitizeFileName(baseDir, rawName string) (string, error) {
	cleanName := filepath.Base(rawName)
	if cleanName == "." || cleanName == "/" || strings.Contains(cleanName, "..") {
		return "", ErrPathTraversal
	}
	return filepath.Join(baseDir, cleanName), nil
}

func (c *BackupConfig) Validate() error {
	if c.AutoBackup && (strings.TrimSpace(c.Token) == "" || strings.TrimSpace(c.ChatID) == "") {
		return ErrInvalidTelegramConfig
	}
	c.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}
