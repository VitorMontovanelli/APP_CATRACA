package inbound

import (
	"context"
	"fmt"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type StudentCrudHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewStudentCrudHandler(db *gorm.DB, session *SessionManager) *StudentCrudHandler {
	return &StudentCrudHandler{db: db, session: session}
}

func (h *StudentCrudHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func strPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func (h *StudentCrudHandler) ListarStudents() ([]models.Student, error) {
	list := []models.Student{}
	err := h.db.Order("nome").Find(&list).Error
	return list, err
}

func (h *StudentCrudHandler) CriarStudent(nome, cpf, dataNascimento, telefone, email string, formaPagamentoID uint) (*models.Student, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	student := models.Student{
		Nome:             nome,
		CPF:              cpf,
		DataNascimento:   strPtr(dataNascimento),
		Telefone:         strPtr(telefone),
		Email:            strPtr(email),
		FormaPagamentoID: uintPtr(formaPagamentoID),
	}
	if err := h.db.Create(&student).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAudit("criou_aluno", student.ID)
	return h.BuscarStudent(student.ID)
}

func (h *StudentCrudHandler) BuscarStudent(id uint) (*models.Student, error) {
	var s models.Student
	if err := h.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (h *StudentCrudHandler) AtualizarStudent(id uint, nome, cpf, dataNascimento, telefone, telefoneUrgencia, email string, formaPagamentoID uint) (*models.Student, error) {
	if !h.session.IsAdmin() {
		return nil, fmt.Errorf("permissão negada")
	}
	student, err := h.BuscarStudent(id)
	if err != nil {
		return nil, err
	}
	student.Nome = nome
	student.CPF = cpf
	student.DataNascimento = strPtr(dataNascimento)
	student.Telefone = strPtr(telefone)
	student.TelefoneUrgencia = strPtr(telefoneUrgencia)
	student.Email = strPtr(email)
	student.FormaPagamentoID = uintPtr(formaPagamentoID)

	if err := h.db.Save(student).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAudit("editou_aluno", id)
	return h.BuscarStudent(id)
}

func (h *StudentCrudHandler) AtivarStudent(id uint, ativo bool) error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	err := h.db.Model(&models.Student{}).Where("id = ?", id).Update("ativo", ativo).Error
	if err == nil {
		action := "desativou_aluno"
		if ativo {
			action = "ativou_aluno"
		}
		h.session.RegistrarAudit(action, id)
	}
	return err
}

type studentRawRow struct {
	ID               uint
	Nome             string
	CPF              string
	DataNascimento   *string
	Telefone         *string
	TelefoneUrgencia *string
	Email            *string
	LaudoMedico      *string
	Foto             *string
	FormaPagamentoID *uint
	DataEntrada      string
	Observacao       *string
	Ativo            bool
	CreatedAt        string
	UpdatedAt        *string
	PlanoNome        *string
	PlanoPreco       *int
	PlanoStatus      *string
	StudentPlanID    *uint
	PaymentMethod    *string
	DueDay           *int
	UltimaFatura     *string
	FaturaValor      *int
	FaturaVencimento *string
}

func (h *StudentCrudHandler) ListarStudentsComPlanos() ([]models.StudentComPlano, error) {
	var raw []studentRawRow
	query := `SELECT s.id, s.nome, s.cpf, s.data_nascimento, s.telefone, s.telefone_urgencia, s.email, s.laudo_medico, s.foto,
		s.forma_pagamento_id, s.data_entrada, s.observacao, s.ativo, s.created_at, s.updated_at,
		p.name, p.price_cents, sp.status, sp.id, pm.name, sp.due_day, i.status, i.amount_cents, i.due_date
		FROM students s
		LEFT JOIN student_plans sp ON sp.student_id = s.id AND sp.status = 'active'
		LEFT JOIN plans p ON p.id = sp.plan_id
		LEFT JOIN payment_methods pm ON pm.id = COALESCE(sp.payment_method_id, s.forma_pagamento_id)
		LEFT JOIN invoices i ON i.id = (SELECT id FROM invoices WHERE student_id = s.id ORDER BY due_date DESC LIMIT 1)
		ORDER BY s.nome`
	if err := h.db.Raw(query).Scan(&raw).Error; err != nil {
		return nil, err
	}

	list := make([]models.StudentComPlano, len(raw))
	for i, r := range raw {
		list[i] = convertRawToStudentComPlano(&r)
	}
	return list, nil
}

func convertRawToStudentComPlano(r *studentRawRow) models.StudentComPlano {
	de, ca, ua := parseStudentDates(r.DataEntrada, r.CreatedAt, r.UpdatedAt)
	return models.StudentComPlano{
		Student: models.Student{
			ID:               r.ID,
			Nome:             r.Nome,
			CPF:              r.CPF,
			DataNascimento:   r.DataNascimento,
			Telefone:         r.Telefone,
			TelefoneUrgencia: r.TelefoneUrgencia,
			Email:            r.Email,
			LaudoMedico:      r.LaudoMedico,
			Foto:             r.Foto,
			FormaPagamentoID: r.FormaPagamentoID,
			DataEntrada:      de,
			Observacao:       r.Observacao,
			Ativo:            r.Ativo,
			CreatedAt:        ca,
			UpdatedAt:        ua,
		},
		PlanoNome:        r.PlanoNome,
		PlanoPreco:       r.PlanoPreco,
		PlanoStatus:      r.PlanoStatus,
		StudentPlanID:    r.StudentPlanID,
		PaymentMethod:    r.PaymentMethod,
		DueDay:           r.DueDay,
		UltimaFatura:     r.UltimaFatura,
		FaturaValor:      r.FaturaValor,
		FaturaVencimento: r.FaturaVencimento,
	}
}

func parseStudentDates(dataEntrada, createdAt string, updatedAt *string) (time.Time, time.Time, time.Time) {
	var de, ca, ua time.Time
	if t, err := time.Parse(time.RFC3339, dataEntrada); err == nil {
		de = t
	}
	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		ca = t
	}
	if updatedAt != nil {
		if t, err := time.Parse(time.RFC3339, *updatedAt); err == nil {
			ua = t
		}
	}
	return de, ca, ua
}
