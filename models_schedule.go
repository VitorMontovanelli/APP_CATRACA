package main

import (
	"time"
)

type Agendamento struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Data       string    `gorm:"size:10;index;not null" json:"data"`
	StudentID  *uint     `gorm:"index" json:"student_id"`
	Nome       string    `gorm:"size:255;not null" json:"nome"`
	Turno      string    `gorm:"size:10;not null;check:turno IN ('manha','tarde','noite')" json:"turno"`
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
	Capacidade int    `gorm:"not null" json:"capacidade"`
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
}

type ConfigAgendaIndividual struct {
	LimiteDiario int `json:"limite_diario"`
	LimiteMaximo int `json:"limite_maximo"`
}
