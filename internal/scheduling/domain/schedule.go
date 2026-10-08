package domain

import (
	"errors"
	"time"
)

const CapacidadeDiaPadrao = 10

var (
	ErrCapacidadeEsgotada   = errors.New("capacidade de vagas esgotada para o dia selecionado")
	ErrTurnoInvalido        = errors.New("turno inválido: esperado manha, tarde ou noite")
	ErrPraticaInvalida      = errors.New("prática inválida para agendamento individual")
	ErrHorarioOcupado       = errors.New("o horário selecionado já está ocupado")
	ErrLimiteIndividualAtingido = errors.New("limite diário de atendimentos individuais atingido")
)

type Agendamento struct {
	ID         uint
	Data       string // YYYY-MM-DD
	StudentID  *uint
	Nome       string
	Turno      string // manha, tarde, noite
	Telefone   *string
	Observacao *string
	CreatedAt  time.Time
}

type AgendamentoIndividual struct {
	ID         uint
	Data       string // YYYY-MM-DD
	Hora       string // HH:MM
	StudentID  *uint
	Nome       string
	Pratica    string
	Observacao *string
	CreatedAt  time.Time
}

func ValidarTurno(turno string) bool {
	return turno == "manha" || turno == "tarde" || turno == "noite"
}

func ValidarPratica(pratica string) bool {
	switch pratica {
	case "ventosaterapia", "liberacao_miofascial", "personal_trainer", "kinesio_tape":
		return true
	default:
		return false
	}
}
