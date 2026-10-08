package adapters

import (
	"context"
	"time"

	"catraca-app/internal/identity/domain"
	"catraca-app/internal/identity/ports"
	"gorm.io/gorm"
)

type UserModel struct {
	ID            uint      `gorm:"primaryKey"`
	Name          string    `gorm:"size:255;not null"`
	Email         string    `gorm:"size:255;uniqueIndex;not null"`
	Senha         string    `gorm:"size:255;not null"`
	Cargo         string    `gorm:"size:20;not null"`
	RegistradorID *uint
	Ativo         bool      `gorm:"default:true"`
	Foto          *string   `gorm:"size:500000"`
	Permissoes    string    `gorm:"size:1000"`
	CriadoEm      time.Time `gorm:"autoCreateTime"`
}

func (UserModel) TableName() string {
	return "users"
}

type AuditLogModel struct {
	ID        uint      `gorm:"primaryKey"`
	ActorID   uint      `gorm:"not null"`
	Action    string    `gorm:"size:100;not null"`
	TargetID  *uint
	Details   string    `gorm:"size:500"`
	Timestamp time.Time `gorm:"autoCreateTime"`
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}

type SQLiteUserRepository struct {
	db *gorm.DB
}

func NewSQLiteUserRepository(db *gorm.DB) ports.UserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) toDomain(m *UserModel) *domain.User {
	return &domain.User{
		ID:            m.ID,
		Name:          m.Name,
		Email:         m.Email,
		SenhaHash:     m.Senha,
		Cargo:         m.Cargo,
		RegistradorID: m.RegistradorID,
		Ativo:         m.Ativo,
		Foto:          m.Foto,
		Permissoes:    m.Permissoes,
		CriadoEm:      m.CriadoEm,
	}
}

func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		return nil, err
	}
	return r.toDomain(&m), nil
}

func (r *SQLiteUserRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return r.toDomain(&m), nil
}

func (r *SQLiteUserRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	users := make([]domain.User, len(models))
	for i, m := range models {
		users[i] = *r.toDomain(&m)
	}
	return users, nil
}

func (r *SQLiteUserRepository) Create(ctx context.Context, u *domain.User) error {
	m := UserModel{
		Name:          u.Name,
		Email:         u.Email,
		Senha:         u.SenhaHash,
		Cargo:         u.Cargo,
		RegistradorID: u.RegistradorID,
		Ativo:         u.Ativo,
		Permissoes:    u.Permissoes,
	}
	err := r.db.WithContext(ctx).Create(&m).Error
	if err == nil {
		u.ID = m.ID
		u.CriadoEm = m.CriadoEm
	}
	return err
}

func (r *SQLiteUserRepository) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"name":       u.Name,
		"email":      u.Email,
		"cargo":      u.Cargo,
		"permissoes": u.Permissoes,
	}).Error
}

func (r *SQLiteUserRepository) UpdatePassword(ctx context.Context, id uint, hash string) error {
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Update("senha", hash).Error
}

func (r *SQLiteUserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&UserModel{}, id).Error
}

func (r *SQLiteUserRepository) SaveAudit(ctx context.Context, log *domain.AuditLogRecord) error {
	m := AuditLogModel{
		ActorID:   log.ActorID,
		Action:    log.Action,
		TargetID:  log.TargetID,
		Details:   log.Details,
		Timestamp: log.Timestamp,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}
