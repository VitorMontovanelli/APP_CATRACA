package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) gerarArquivoBackup() (string, error) {
	tmpDir, err := os.MkdirTemp("", "catraca_backup_*")
	if err != nil {
		return "", fmt.Errorf("erro ao criar diretório temporário: %w", err)
	}

	dbPath := filepath.Join(tmpDir, "catraca.db")
	escaped := strings.ReplaceAll(dbPath, "'", "''")
	if err := a.db.Exec("VACUUM INTO '" + escaped + "'").Error; err != nil {
		return "", fmt.Errorf("erro ao gerar snapshot do banco: %w", err)
	}

	zipPath := filepath.Join(tmpDir, "catraca_backup_"+time.Now().Format("20060102_150405")+".zip")
	if err := criarZip(zipPath, dbPath); err != nil {
		return "", err
	}
	return zipPath, nil
}

func criarZip(zipPath, filePath string) error {
	zf, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo zip: %w", err)
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir snapshot: %w", err)
	}
	defer src.Close()

	dst, err := zw.Create(filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("erro ao criar entrada no zip: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("erro ao compactar snapshot: %w", err)
	}
	return nil
}

func enviarDocumentoTelegram(token, chatID, filePath, caption string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir backup: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return err
	}
	if err := writer.WriteField("caption", caption); err != nil {
		return err
	}
	part, err := writer.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+token+"/sendDocument", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar backup para o Telegram: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("resposta inválida do Telegram: %w", err)
	}
	if !result.Ok {
		return fmt.Errorf("falha ao enviar backup: %s", result.Description)
	}
	return nil
}
