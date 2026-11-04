package domain

import (
	"errors"
	"time"
)

const CapacidadeDiaPadrao = 48

// HorariosAula é a lista única de horários do calendário de Pilates.
var HorariosAula = []string{
	"06:00", "07:00", "08:00", "09:00", "10:00",
	"15:00", "16:00", "17:00", "18:00", "19:00",
}

func HorarioAulaValido(hora string) bool {
	for _, h := range HorariosAula {
		if h == hora {
			return true
		}
	}
	return false
}

func TurnoDoHorario(hora string) string {
	switch hora {
	case "06:00", "07:00", "08:00", "09:00", "10:00":
		return "manha"
	case "15:00", "16:00", "17:00", "18:00":
		return "tarde"
	default:
		return "noite"
	}
}

// DiaIndisponivel informa se a data cai em sábado ou domingo.
func DiaIndisponivel(data string) bool {
	t, err := time.Parse("2006-01-02", data)
	if err != nil {
		return false
	}
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

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
