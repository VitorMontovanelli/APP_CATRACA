package inbound

import (
	"context"
	"encoding/base64"
	"errors"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("aluno não encontrado")
		}
		return fmt.Errorf("não foi possível localizar o aluno: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".pdf" {
		return fmt.Errorf("apenas arquivos em formato PDF são permitidos")
	}

	dir := filepath.Join("cliente", "laudo")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("não foi possível criar a pasta de laudos: %w", err)
	}

	dest := filepath.Join(dir, fmt.Sprintf("laudo_aluno_%d%s", studentID, ext))

	_, raw, found := strings.Cut(base64Data, ",")
	if !found {
		raw = base64Data
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("o arquivo PDF enviado é inválido")
	}
	if len(data) == 0 {
		return fmt.Errorf("o arquivo PDF está vazio")
	}

	// Guarda se já existia um laudo no mesmo caminho: assim, se a atualização do
	// banco falhar, só removemos o arquivo novo quando ele não substituía um
	// laudo anterior (evita deixar arquivo órfão e evita apagar o laudo antigo).
	_, statErr := os.Stat(dest)
	existiaAntes := statErr == nil

	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("não foi possível gravar o arquivo do laudo: %w", err)
	}

	if err := h.db.Model(&models.Student{}).Where("id = ?", studentID).Update("laudo_medico", dest).Error; err != nil {
		if !existiaAntes {
			_ = os.Remove(dest)
		}
		return fmt.Errorf("não foi possível vincular o laudo ao cadastro do aluno: %w", err)
	}

	h.session.RegistrarAuditComDetalhes("salvou_laudo_aluno", studentID, fmt.Sprintf("laudo=%s", dest))
	return nil
}

func (h *StudentLaudosHandler) RemoverLaudoAluno(studentID uint) error {
	var student models.Student
	if err := h.db.First(&student, studentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("aluno não encontrado")
		}
		return fmt.Errorf("não foi possível localizar o aluno: %w", err)
	}

	if student.LaudoMedico != nil && *student.LaudoMedico != "" {
		if _, err := os.Stat(*student.LaudoMedico); err == nil {
			_ = os.Remove(*student.LaudoMedico)
		}
	}

	if err := h.db.Model(&models.Student{}).Where("id = ?", studentID).Update("laudo_medico", nil).Error; err != nil {
		return fmt.Errorf("não foi possível remover o laudo do cadastro do aluno: %w", err)
	}

	h.session.RegistrarAuditComDetalhes("removeu_laudo_aluno", studentID, "")
	return nil
}

func (h *StudentLaudosHandler) BaixarLaudo(studentID uint) error {
	var student models.Student
	if err := h.db.First(&student, studentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("aluno não encontrado")
		}
		return fmt.Errorf("não foi possível localizar o aluno: %w", err)
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
