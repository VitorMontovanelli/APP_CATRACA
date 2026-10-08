package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) SalvarFotoAluno(studentID uint, base64Data string) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	var student Student
	if err := a.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("aluno não encontrado")
	}

	dir := filepath.Join("cliente", "fotos")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %w", err)
	}

	dest := filepath.Join(dir, fmt.Sprintf("foto_aluno_%d.jpg", studentID))

	_, raw, found := strings.Cut(base64Data, ",")
	if !found {
		raw = base64Data
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("erro ao decodificar foto: %w", err)
	}

	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar foto: %w", err)
	}

	if err := a.db.Model(&Student{}).Where("id = ?", studentID).Update("foto", dest).Error; err != nil {
		return fmt.Errorf("erro ao atualizar aluno: %w", err)
	}

	a.registrarAuditComDetalhes("salvou_foto_aluno", studentID, fmt.Sprintf("foto=%s", dest))
	return nil
}

// RemoverFotoAluno remove a foto do aluno: apaga o arquivo (se existir) e
// limpa o campo foto no banco.
func (a *App) RemoverFotoAluno(studentID uint) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	var student Student
	if err := a.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("aluno não encontrado")
	}

	if student.Foto != nil && *student.Foto != "" {
		if _, err := os.Stat(*student.Foto); err == nil {
			_ = os.Remove(*student.Foto)
		}
	}

	if err := a.db.Model(&Student{}).Where("id = ?", studentID).Update("foto", nil).Error; err != nil {
		return fmt.Errorf("erro ao atualizar aluno: %w", err)
	}

	a.registrarAuditComDetalhes("removeu_foto_aluno", studentID, "")
	return nil
}

// ObterFotoAluno retorna a foto do aluno como data URL (base64) para exibição
// inline no webview, ou uma string vazia quando não há foto.
func (a *App) ObterFotoAluno(studentID uint) (string, error) {
	var student Student
	if err := a.db.First(&student, studentID).Error; err != nil {
		return "", fmt.Errorf("aluno não encontrado")
	}
	if student.Foto == nil || *student.Foto == "" {
		return "", nil
	}
	if _, err := os.Stat(*student.Foto); err != nil {
		return "", nil
	}
	data, err := os.ReadFile(*student.Foto)
	if err != nil {
		return "", fmt.Errorf("erro ao ler foto: %w", err)
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data), nil
}
