package adapters

import (
	"context"
	"time"

	"catraca-app/internal/student/domain"
	"catraca-app/internal/student/ports"
	"gorm.io/gorm"
)

type StudentModel struct {
	ID               uint      `gorm:"primaryKey"`
	Nome             string    `gorm:"size:255;not null"`
	CPF              string    `gorm:"size:14;uniqueIndex"`
	DataNascimento   *string   `gorm:"size:10"`
	Telefone         *string   `gorm:"size:20"`
	TelefoneUrgencia *string   `gorm:"size:20"`
	Email            *string   `gorm:"size:255"`
	LaudoMedico      *string   `gorm:"size:500"`
	Foto             *string   `gorm:"size:500000"`
	FormaPagamentoID *uint
	DataEntrada      time.Time `gorm:"autoCreateTime"`
	Observacao       *string   `gorm:"size:500"`
	Ativo            bool      `gorm:"default:true"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (StudentModel) TableName() string {
	return "students"
}

type SQLiteStudentRepository struct {
	db *gorm.DB
}

func NewSQLiteStudentRepository(db *gorm.DB) ports.StudentRepository {
	return &SQLiteStudentRepository{db: db}
}

func (r *SQLiteStudentRepository) toDomain(m *StudentModel) *domain.Student {
	return &domain.Student{
		ID:               m.ID,
		Nome:             m.Nome,
		CPF:              m.CPF,
		DataNascimento:   m.DataNascimento,
		Telefone:         m.Telefone,
		TelefoneUrgencia: m.TelefoneUrgencia,
		Email:            m.Email,
		LaudoMedico:      m.LaudoMedico,
		Foto:             m.Foto,
		FormaPagamentoID: m.FormaPagamentoID,
		DataEntrada:      m.DataEntrada,
		Observacao:       m.Observacao,
		Ativo:            m.Ativo,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func (r *SQLiteStudentRepository) FindByID(ctx context.Context, id uint) (*domain.Student, error) {
	var m StudentModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return r.toDomain(&m), nil
}

func (r *SQLiteStudentRepository) ListAll(ctx context.Context) ([]domain.Student, error) {
	var models []StudentModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	students := make([]domain.Student, len(models))
	for i, m := range models {
		students[i] = *r.toDomain(&m)
	}
	return students, nil
}

func (r *SQLiteStudentRepository) Create(ctx context.Context, s *domain.Student) error {
	m := StudentModel{
		Nome:             s.Nome,
		CPF:              s.CPF,
		DataNascimento:   s.DataNascimento,
		Telefone:         s.Telefone,
		TelefoneUrgencia: s.TelefoneUrgencia,
		Email:            s.Email,
		FormaPagamentoID: s.FormaPagamentoID,
		Observacao:       s.Observacao,
		Ativo:            s.Ativo,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	s.ID = m.ID
	s.DataEntrada = m.DataEntrada
	return nil
}

func (r *SQLiteStudentRepository) Update(ctx context.Context, s *domain.Student) error {
	return r.db.WithContext(ctx).Model(&StudentModel{}).Where("id = ?", s.ID).Updates(map[string]interface{}{
		"nome":               s.Nome,
		"cpf":                s.CPF,
		"data_nascimento":    s.DataNascimento,
		"telefone":           s.Telefone,
		"telefone_urgencia":  s.TelefoneUrgencia,
		"email":              s.Email,
		"forma_pagamento_id": s.FormaPagamentoID,
		"observacao":         s.Observacao,
	}).Error
}

func (r *SQLiteStudentRepository) SetActive(ctx context.Context, id uint, active bool) error {
	return r.db.WithContext(ctx).Model(&StudentModel{}).Where("id = ?", id).Update("ativo", active).Error
}

func (r *SQLiteStudentRepository) SavePhoto(ctx context.Context, id uint, base64Photo string) error {
	return r.db.WithContext(ctx).Model(&StudentModel{}).Where("id = ?", id).Update("foto", base64Photo).Error
}

func (r *SQLiteStudentRepository) GetPhoto(ctx context.Context, id uint) (string, error) {
	var m StudentModel
	if err := r.db.WithContext(ctx).Select("foto").First(&m, id).Error; err != nil {
		return "", err
	}
	if m.Foto == nil {
		return "", nil
	}
	return *m.Foto, nil
}

func (r *SQLiteStudentRepository) RemovePhoto(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&StudentModel{}).Where("id = ?", id).Update("foto", nil).Error
}
