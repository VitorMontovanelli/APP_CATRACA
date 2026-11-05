package models

import (
	"time"
)

type Agendamento struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Data       string    `gorm:"size:10;index;not null" json:"data"`
	StudentID  *uint     `gorm:"index" json:"student_id"`
	Nome       string    `gorm:"size:255;not null" json:"nome"`
	Turno      string    `gorm:"size:10;not null;check:turno IN ('manha','tarde','noite')" json:"turno"`
	Hora       string    `gorm:"size:5;not null;default:''" json:"hora"`
	Telefone   *string   `gorm:"size:20" json:"telefone"`
	Observacao *string   `gorm:"size:500" json:"observacao"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type DiaAgenda struct {
	Data       string `json:"data"`
	Total      int    `json:"total"`
	Capacidade int    `json:"capacidade"`
}

type CapacidadeDia struct {
	Data       string `gorm:"primaryKey;size:10" json:"data"`
	Calendario string `gorm:"primaryKey;size:30;not null;default:pilates" json:"calendario"`
	Capacidade int    `gorm:"not null" json:"capacidade"`
}

// TableName fixa a tabela usada pelo restante do sistema (SQL manual e
// calendários). Sem isso o GORM grava em "capacidade_dia", enquanto as
// consultas leem "capacidade_dias".
func (CapacidadeDia) TableName() string {
	return "capacidade_dias"
}

type ConfigAgenda struct {
	Calendario       string   `json:"calendario"`
	CapacidadePadrao int      `json:"capacidade_padrao"`
	CapacidadeMaxima int      `json:"capacidade_maxima"`
	Horarios         []string `json:"horarios"`
}

const (
	PraticaVentosaterapia      = "ventosaterapia"
	PraticaLiberacaoMiofascial = "liberacao_miofascial"
	PraticaPersonalTrainer     = "personal_trainer"
	PraticaKinesioTape         = "kinesio_tape"
)

type AgendamentoIndividual struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Data       string    `gorm:"size:10;index;not null" json:"data"`
	Hora       string    `gorm:"size:5;not null" json:"hora"`
	StudentID  *uint     `gorm:"index" json:"student_id"`
	Nome       string    `gorm:"size:255;not null" json:"nome"`
	Pratica    string    `gorm:"size:30;not null;check:pratica IN ('ventosaterapia','liberacao_miofascial','personal_trainer','kinesio_tape')" json:"pratica"`
	Observacao *string   `gorm:"size:500" json:"observacao"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type DiaIndividualAgenda struct {
	Data       string `json:"data"`
	Total      int    `json:"total"`
	Capacidade int    `json:"capacidade"`
	Praticas   string `json:"praticas"`
	Contagens  string `json:"contagens"`
}

type ConfigAgendaIndividual struct {
	LimiteDiario int `json:"limite_diario"`
	LimiteMaximo int `json:"limite_maximo"`
}
