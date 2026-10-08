package main

import (
	"time"

	"gorm.io/gorm"
)

type Aluno struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Nome            string    `gorm:"size:255;not null" json:"nome"`
	CPF             string    `gorm:"size:14;uniqueIndex" json:"cpf"`
	Status          bool      `gorm:"default:true" json:"status"`
	VencimentoPlano time.Time `gorm:"not null" json:"vencimento_plano"`
	IDBiometriaMock int       `gorm:"uniqueIndex;default:0" json:"id_biometria_mock"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RegistroAcesso struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	AlunoID  uint      `gorm:"index;not null" json:"aluno_id"`
	Aluno    Aluno     `gorm:"foreignKey:AlunoID" json:"-"`
	DataHora time.Time `gorm:"autoCreateTime" json:"data_hora"`
	Liberado bool      `gorm:"not null" json:"liberado"`
	Motivo   string    `gorm:"size:255" json:"motivo"`
}

type ResultadoAcesso struct {
	Nome     string `json:"nome"`
	Liberado bool   `json:"liberado"`
	Mensagem string `json:"mensagem"`
}

type AccessLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `json:"user_id"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Resultado string    `gorm:"size:20;not null;check:resultado IN ('liberado','negado')" json:"resultado"`
	Motivo    string    `gorm:"size:255" json:"motivo"`
}

// BeforeCreate hook para alunos da catraca: gera vencimento padrão (30 dias) se não definido
func (a *Aluno) BeforeCreate(tx *gorm.DB) error {
	if a.VencimentoPlano.IsZero() {
		a.VencimentoPlano = time.Now().AddDate(0, 1, 0)
	}
	return nil
}
