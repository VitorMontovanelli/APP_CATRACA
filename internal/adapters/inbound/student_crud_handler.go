package inbound

import (
	"context"
	"fmt"
	"strings"
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

func nilOrString(v string) interface{} {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return strings.TrimSpace(v)
}

func normalizarCPF(cpf string) string {
	var b strings.Builder
	for _, r := range cpf {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func validarDadosAluno(nome, cpf, telefone string) error {
	if strings.TrimSpace(nome) == "" {
		return fmt.Errorf("informe o nome do aluno")
	}
	if normalizarCPF(cpf) == "" {
		return fmt.Errorf("informe o CPF do aluno")
	}
	if len(normalizarCPF(cpf)) != 11 {
		return fmt.Errorf("CPF inválido: informe os 11 dígitos")
	}
	if strings.TrimSpace(telefone) == "" {
		return fmt.Errorf("informe o telefone do aluno")
	}
	return nil
}

// verificarCPFDuplicado compara apenas os dígitos, então CPFs com ou sem
// pontuação são tratados como o mesmo documento. CPF vazio nunca conflita.
func verificarCPFDuplicado(db *gorm.DB, cpf string, ignorarID uint) error {
	digitos := normalizarCPF(cpf)
	if digitos == "" {
		return nil
	}

	var conflito struct {
		ID   uint
		Nome string
	}
	query := `SELECT id, nome FROM students
		WHERE id <> ? AND REPLACE(REPLACE(REPLACE(REPLACE(COALESCE(cpf,''),'.',''),'-',''),' ',''),'/','') = ?
		LIMIT 1`
	if err := db.Raw(query, ignorarID, digitos).Scan(&conflito).Error; err != nil {
		return err
	}
	if conflito.ID != 0 {
		return fmt.Errorf("Este CPF já está cadastrado para %s", conflito.Nome)
	}
	return nil
}

func tratarErroCPF(db *gorm.DB, err error, cpf string, ignorarID uint) error {
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: students.cpf") {
		if dupErr := verificarCPFDuplicado(db, cpf, ignorarID); dupErr != nil {
			return dupErr
		}
		return fmt.Errorf("Este CPF já está cadastrado para outro aluno")
	}
	return err
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
	nome = strings.TrimSpace(nome)
	cpf = strings.TrimSpace(cpf)
	if err := validarDadosAluno(nome, cpf, telefone); err != nil {
		return nil, err
	}
	if err := verificarCPFDuplicado(h.db, cpf, 0); err != nil {
		return nil, err
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
		return nil, tratarErroCPF(h.db, err, cpf, 0)
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
	if _, err := h.BuscarStudent(id); err != nil {
		return nil, err
	}
	nome = strings.TrimSpace(nome)
	cpf = strings.TrimSpace(cpf)
	if err := validarDadosAluno(nome, cpf, telefone); err != nil {
		return nil, err
	}
	if err := verificarCPFDuplicado(h.db, cpf, id); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"nome":               nome,
		"cpf":                nilOrString(cpf),
		"data_nascimento":    strPtr(dataNascimento),
		"telefone":           strPtr(telefone),
		"telefone_urgencia":  strPtr(telefoneUrgencia),
		"email":              strPtr(email),
		"forma_pagamento_id": uintPtr(formaPagamentoID),
	}
	if err := h.db.Model(&models.Student{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, tratarErroCPF(h.db, err, cpf, id)
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
		p.name AS plano_nome, p.price_cents AS plano_preco, sp.status AS plano_status, sp.id AS student_plan_id,
		pm.name AS payment_method, sp.due_day AS due_day, i.status AS ultima_fatura, i.amount_cents AS fatura_valor,
		i.due_date AS fatura_vencimento
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
