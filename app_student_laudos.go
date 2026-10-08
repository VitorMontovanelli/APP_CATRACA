package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) SalvarComprovante(invoiceID uint, base64Data, fileName string) error {
	dir := "comprovantes"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".pdf" {
		return fmt.Errorf("formato de arquivo inválido: apenas comprovantes em formato PDF são permitidos")
	}
	dest := filepath.Join(dir, fmt.Sprintf("invoice_%d%s", invoiceID, ext))

	_, raw, found := strings.Cut(base64Data, ",")
	if !found {
		raw = base64Data
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("erro ao decodificar arquivo: %w", err)
	}

	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}

	now := time.Now()
	err = a.db.Exec(
		"UPDATE invoices SET comprovante = ?, status = 'paid', paid_at = ? WHERE id = ?",
		dest, now, invoiceID,
	).Error
	if err != nil {
		return fmt.Errorf("erro ao atualizar fatura: %w", err)
	}

	a.registrarAuditComDetalhes("confirmou_pagamento", invoiceID,
		fmt.Sprintf("comprovante=%s, invoice=%d", dest, invoiceID))
	a.gerarProximaFatura(invoiceID)
	return nil
}

// SalvarLaudoAluno salva o laudo médico (PDF) do aluno na pasta cliente/laudo/
// e grava o caminho no campo laudo_medico do aluno. Reutiliza a mesma estratégia
// de armazenamento local usada em SalvarComprovante.
func (a *App) SalvarLaudoAluno(studentID uint, base64Data, fileName string) error {
	var student Student
	if err := a.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("aluno não encontrado")
	}

	dir := filepath.Join("cliente", "laudo")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".pdf" {
		return fmt.Errorf("formato de arquivo inválido: apenas o laudo em formato PDF é permitido")
	}

	dest := filepath.Join(dir, fmt.Sprintf("laudo_aluno_%d%s", studentID, ext))

	_, raw, found := strings.Cut(base64Data, ",")
	if !found {
		raw = base64Data
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("erro ao decodificar arquivo: %w", err)
	}

	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}

	if err := a.db.Model(&Student{}).Where("id = ?", studentID).Update("laudo_medico", dest).Error; err != nil {
		return fmt.Errorf("erro ao atualizar aluno: %w", err)
	}

	a.registrarAuditComDetalhes("salvou_laudo_aluno", studentID, fmt.Sprintf("laudo=%s", dest))
	return nil
}

// RemoverLaudoAluno remove o laudo médico do aluno: apaga o arquivo (se existir)
// e limpa o campo laudo_medico no banco.
func (a *App) RemoverLaudoAluno(studentID uint) error {
	var student Student
	if err := a.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("aluno não encontrado")
	}

	if student.LaudoMedico != nil && *student.LaudoMedico != "" {
		if _, err := os.Stat(*student.LaudoMedico); err == nil {
			_ = os.Remove(*student.LaudoMedico)
		}
	}

	if err := a.db.Model(&Student{}).Where("id = ?", studentID).Update("laudo_medico", nil).Error; err != nil {
		return fmt.Errorf("erro ao atualizar aluno: %w", err)
	}

	a.registrarAuditComDetalhes("removeu_laudo_aluno", studentID, "")
	return nil
}

// BaixarLaudo abre o laudo médico em PDF do aluno com o visualizador padrão do SO.
func (a *App) BaixarLaudo(studentID uint) error {
	var student Student
	if err := a.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("aluno não encontrado")
	}
	if student.LaudoMedico == nil || *student.LaudoMedico == "" {
		return fmt.Errorf("aluno não possui laudo anexado")
	}
	if _, err := os.Stat(*student.LaudoMedico); err != nil {
		return fmt.Errorf("arquivo do laudo não encontrado em disco")
	}

	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", *student.LaudoMedico)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erro ao abrir laudo: %w", err)
	}
	return nil
}

// SalvarFotoAluno salva a foto do aluno (JPEG) em cliente/fotos/ e grava o
// caminho relativo no campo foto. Segue a mesma estratégia do laudo em PDF:
// base64 -> arquivo em disco -> caminho no banco.
