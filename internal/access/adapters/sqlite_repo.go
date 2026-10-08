package adapters

import (
	"context"
	"time"

	"catraca-app/internal/access/domain"
	"catraca-app/internal/access/ports"
	"gorm.io/gorm"
)

// AlunoEntity representa a tabela no banco SQLite.
type AlunoEntity struct {
	ID              uint      `gorm:"primaryKey"`
	Nome            string    `gorm:"size:255;not null"`
	CPF             string    `gorm:"size:14;uniqueIndex"`
	Status          bool      `gorm:"default:true"`
	VencimentoPlano time.Time `gorm:"not null"`
	IDBiometriaMock int       `gorm:"uniqueIndex;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (AlunoEntity) TableName() string {
	return "alunos"
}

// RegistroAcessoEntity representa a tabela de logs de passagem.
type RegistroAcessoEntity struct {
	ID       uint      `gorm:"primaryKey"`
	AlunoID  uint      `gorm:"index;not null"`
	DataHora time.Time `gorm:"autoCreateTime"`
	Liberado bool      `gorm:"not null"`
	Motivo   string    `gorm:"size:255"`
}

func (RegistroAcessoEntity) TableName() string {
	return "registro_acessos"
}

type SQLiteAccessRepository struct {
	db *gorm.DB
}

func NewSQLiteAccessRepository(db *gorm.DB) ports.AccessRepository {
	return &SQLiteAccessRepository{db: db}
}

func (r *SQLiteAccessRepository) FindAlunoByBiometria(ctx context.Context, id int) (*domain.CatracaAluno, error) {
	var entity AlunoEntity
	err := r.db.WithContext(ctx).Where("id_biometria_mock = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &domain.CatracaAluno{
		ID:              entity.ID,
		Nome:            entity.Nome,
		CPF:             entity.CPF,
		Status:          entity.Status,
		VencimentoPlano: entity.VencimentoPlano,
		IDBiometriaMock: entity.IDBiometriaMock,
	}, nil
}

func (r *SQLiteAccessRepository) SaveAccessLog(ctx context.Context, log *domain.AccessLogRecord) error {
	entity := RegistroAcessoEntity{
		AlunoID:  log.AlunoID,
		DataHora: log.DataHora,
		Liberado: log.Liberado,
		Motivo:   log.Motivo,
	}
	return r.db.WithContext(ctx).Create(&entity).Error
}

func (r *SQLiteAccessRepository) CreateAluno(ctx context.Context, aluno *domain.CatracaAluno) error {
	entity := AlunoEntity{
		Nome:            aluno.Nome,
		CPF:             aluno.CPF,
		Status:          aluno.Status,
		VencimentoPlano: aluno.VencimentoPlano,
		IDBiometriaMock: aluno.IDBiometriaMock,
	}
	err := r.db.WithContext(ctx).Create(&entity).Error
	if err == nil {
		aluno.ID = entity.ID
	}
	return err
}
