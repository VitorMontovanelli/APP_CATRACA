package inbound

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"catraca-app/internal/models"
	adapters "catraca-app/internal/student/adapters"
	"gorm.io/gorm"
)

type StudentPhotosHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewStudentPhotosHandler(db *gorm.DB, session *SessionManager) *StudentPhotosHandler {
	return &StudentPhotosHandler{db: db, session: session}
}

func (h *StudentPhotosHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *StudentPhotosHandler) SalvarFotoAluno(studentID uint, base64Data string) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	var student models.Student
	if err := h.db.First(&student, studentID).Error; err != nil {
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

	if err := h.db.Model(&models.Student{}).Where("id = ?", studentID).Update("foto", dest).Error; err != nil {
		return fmt.Errorf("erro ao atualizar aluno: %w", err)
	}

	h.session.RegistrarAuditComDetalhes("salvou_foto_aluno", studentID, fmt.Sprintf("foto=%s", dest))
	return nil
}

func (h *StudentPhotosHandler) RemoverFotoAluno(studentID uint) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	var student models.Student
	if err := h.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("aluno não encontrado")
	}

	if student.Foto != nil && *student.Foto != "" {
		if _, err := os.Stat(*student.Foto); err == nil {
			_ = os.Remove(*student.Foto)
		}
	}

	if err := h.db.Model(&models.Student{}).Where("id = ?", studentID).Update("foto", nil).Error; err != nil {
		return fmt.Errorf("erro ao atualizar aluno: %w", err)
	}

	h.session.RegistrarAuditComDetalhes("removeu_foto_aluno", studentID, "")
	return nil
}

func (h *StudentPhotosHandler) ObterFotoAluno(studentID uint) (string, error) {
	var student models.Student
	if err := h.db.First(&student, studentID).Error; err != nil {
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

func (h *StudentPhotosHandler) CapturarFotoWebcam() (string, error) {
	adapter := adapters.NewHardwareCameraAdapter()
	return adapter.CaptureFrame(context.Background())
}

func (h *StudentPhotosHandler) IniciarStreamWebcam() string {
	return adapters.GetCameraStreamer().GetStreamURL()
}

func (h *StudentPhotosHandler) PararStreamWebcam() {
	adapters.GetCameraStreamer().Stop()
}

func (h *StudentPhotosHandler) CapturarSnapshotStream() (string, error) {
	return adapters.GetCameraStreamer().CaptureSnapshot()
}
