package adapters

import (
	"context"
	"time"

	"catraca-app/internal/scheduling/domain"
	"catraca-app/internal/scheduling/ports"
	"gorm.io/gorm"
)

type AgendamentoModel struct {
	ID         uint      `gorm:"primaryKey"`
	Data       string    `gorm:"size:10;index;not null"`
	StudentID  *uint     `gorm:"index"`
	Nome       string    `gorm:"size:255;not null"`
	Turno      string    `gorm:"size:10;not null"`
	Telefone   *string   `gorm:"size:20"`
	Observacao *string   `gorm:"size:500"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (AgendamentoModel) TableName() string {
	return "agendamentos"
}

type AgendamentoIndividualModel struct {
	ID         uint      `gorm:"primaryKey"`
	Data       string    `gorm:"size:10;index;not null"`
	Hora       string    `gorm:"size:5;not null"`
	StudentID  *uint     `gorm:"index"`
	Nome       string    `gorm:"size:255;not null"`
	Pratica    string    `gorm:"size:30;not null"`
	Observacao *string   `gorm:"size:500"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (AgendamentoIndividualModel) TableName() string {
	return "agendamento_individuals"
}

type CapacidadeDiaModel struct {
	Data       string `gorm:"primaryKey;size:10"`
	Capacidade int    `gorm:"not null"`
}

func (CapacidadeDiaModel) TableName() string {
	return "capacidade_dias"
}

type SQLiteScheduleRepository struct {
	db *gorm.DB
}

func NewSQLiteScheduleRepository(db *gorm.DB) ports.ScheduleRepository {
	return &SQLiteScheduleRepository{db: db}
}

func (r *SQLiteScheduleRepository) ObterCapacidadeDia(ctx context.Context, data string) (int, error) {
	var cd CapacidadeDiaModel
	if err := r.db.WithContext(ctx).Where("data = ?", data).First(&cd).Error; err != nil {
		return domain.CapacidadeDiaPadrao, nil
	}
	return cd.Capacidade, nil
}

func (r *SQLiteScheduleRepository) DefinirCapacidadeDia(ctx context.Context, data string, cap int) error {
	return r.db.WithContext(ctx).Save(&CapacidadeDiaModel{Data: data, Capacidade: cap}).Error
}

func (r *SQLiteScheduleRepository) ContarAgendamentosDia(ctx context.Context, data string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AgendamentoModel{}).Where("data = ?", data).Count(&count).Error
	return int(count), err
}

func (r *SQLiteScheduleRepository) CriarAgendamento(ctx context.Context, ag *domain.Agendamento) error {
	m := AgendamentoModel{
		Data:       ag.Data,
		StudentID:  ag.StudentID,
		Nome:       ag.Nome,
		Turno:      ag.Turno,
		Telefone:   ag.Telefone,
		Observacao: ag.Observacao,
	}
	err := r.db.WithContext(ctx).Create(&m).Error
	if err == nil {
		ag.ID = m.ID
		ag.CreatedAt = m.CreatedAt
	}
	return err
}

func (r *SQLiteScheduleRepository) RemoverAgendamento(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&AgendamentoModel{}, id).Error
}

func (r *SQLiteScheduleRepository) ObterLimiteIndividual(ctx context.Context) (int, error) {
	return 10, nil
}

func (r *SQLiteScheduleRepository) ContarIndividuaisDia(ctx context.Context, data string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AgendamentoIndividualModel{}).Where("data = ?", data).Count(&count).Error
	return int(count), err
}

func (r *SQLiteScheduleRepository) ExisteHorarioIndividual(ctx context.Context, data, hora string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AgendamentoIndividualModel{}).Where("data = ? AND hora = ?", data, hora).Count(&count).Error
	return count > 0, err
}

func (r *SQLiteScheduleRepository) CriarAgendamentoIndividual(ctx context.Context, ag *domain.AgendamentoIndividual) error {
	m := AgendamentoIndividualModel{
		Data:       ag.Data,
		Hora:       ag.Hora,
		StudentID:  ag.StudentID,
		Nome:       ag.Nome,
		Pratica:    ag.Pratica,
		Observacao: ag.Observacao,
	}
	err := r.db.WithContext(ctx).Create(&m).Error
	if err == nil {
		ag.ID = m.ID
		ag.CreatedAt = m.CreatedAt
	}
	return err
}

func (r *SQLiteScheduleRepository) RemoverAgendamentoIndividual(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&AgendamentoIndividualModel{}, id).Error
}
