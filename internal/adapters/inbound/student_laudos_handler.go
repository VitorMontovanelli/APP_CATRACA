package inbound

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type StudentLaudosHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewStudentLaudosHandler(db *gorm.DB, session *SessionManager) *StudentLaudosHandler {
	return &StudentLaudosHandler{db: db, session: session}
}

func (h *StudentLaudosHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *StudentLaudosHandler) SalvarLaudoAluno(studentID uint, base64Data, fileName string) error {
	var student models.Student
	if err := h.db.First(&student, studentID).Error; err != nil {
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

	if err := h.db.Model(&models.Student{}).Where("id = ?", studentID).Update("laudo_medico", dest).Error; err != nil {
		return fmt.Errorf("erro ao atualizar aluno: %w", err)
	}

	h.session.RegistrarAuditComDetalhes("salvou_laudo_aluno", studentID, fmt.Sprintf("laudo=%s", dest))
	return nil
}

func (h *StudentLaudosHandler) RemoverLaudoAluno(studentID uint) error {
	var student models.Student
	if err := h.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("aluno não encontrado")
	}

	if student.LaudoMedico != nil && *student.LaudoMedico != "" {
		if _, err := os.Stat(*student.LaudoMedico); err == nil {
			_ = os.Remove(*student.LaudoMedico)
		}
	}

	if err := h.db.Model(&models.Student{}).Where("id = ?", studentID).Update("laudo_medico", nil).Error; err != nil {
		return fmt.Errorf("erro ao atualizar aluno: %w", err)
	}

	h.session.RegistrarAuditComDetalhes("removeu_laudo_aluno", studentID, "")
	return nil
}

func (h *StudentLaudosHandler) BaixarLaudo(studentID uint) error {
	var student models.Student
	if err := h.db.First(&student, studentID).Error; err != nil {
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
