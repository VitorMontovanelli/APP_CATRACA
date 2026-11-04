package inbound

import (
	"context"
	"fmt"
	"strings"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type StudentPlansHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewStudentPlansHandler(db *gorm.DB, session *SessionManager) *StudentPlansHandler {
	return &StudentPlansHandler{db: db, session: session}
}

func (h *StudentPlansHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *StudentPlansHandler) CriarStudentComPlano(nome, cpf, dataNascimento, telefone, telefoneUrgencia, email string, formaPagamentoID, planID, dueDay uint) (*models.Student, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}

	nome = strings.TrimSpace(nome)
	cpf = strings.TrimSpace(cpf)
	if err := validarDadosAluno(nome, cpf, telefone); err != nil {
		return nil, err
	}
	if err := verificarCPFDuplicado(h.db, cpf, 0); err != nil {
		return nil, err
	}

	tx := h.db.Begin()

	student := models.Student{
		Nome:             nome,
		CPF:              cpf,
		DataNascimento:   strPtr(dataNascimento),
		Telefone:         strPtr(telefone),
		TelefoneUrgencia: strPtr(telefoneUrgencia),
		Email:            strPtr(email),
		FormaPagamentoID: uintPtr(formaPagamentoID),
	}
	if err := tx.Create(&student).Error; err != nil {
		tx.Rollback()
		return nil, tratarErroCPF(h.db, err, cpf, 0)
	}

	if planID > 0 {
		sp := models.StudentPlan{
			StudentID:       student.ID,
			PlanID:          planID,
			DueDay:          int(dueDay),
			PaymentMethodID: uintPtr(formaPagamentoID),
		}
		if err := tx.Create(&sp).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	h.session.RegistrarAudit("criou_aluno", student.ID)
	var created models.Student
	if err := h.db.First(&created, student.ID).Error; err != nil {
		return nil, err
	}
	return &created, nil
}
