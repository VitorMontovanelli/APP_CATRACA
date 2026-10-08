package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidCPF       = errors.New("cpf inválido: formato esperado 000.000.000-00 ou 11 dígitos numéricos")
	ErrEmptyStudentName = errors.New("nome do aluno é obrigatório")
	ErrPhotoTooLarge    = errors.New("a foto não pode exceder 5MB")
)

var cpfRegex = regexp.MustCompile(`^\d{3}\.?\d{3}\.?\d{3}-?\d{2}$`)

type Student struct {
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
	DataEntrada      time.Time
	Observacao       *string
	Ativo            bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ValidateCPF valida o formato básico do CPF antes de persistir.
func ValidateCPF(cpf string) error {
	trimmed := strings.TrimSpace(cpf)
	if !cpfRegex.MatchString(trimmed) {
		return ErrInvalidCPF
	}
	return nil
}

// ValidateStudent valida invariantes essenciais de negócio para cadastro de aluno.
func (s *Student) Validate() error {
	if strings.TrimSpace(s.Nome) == "" {
		return ErrEmptyStudentName
	}
	if err := ValidateCPF(s.CPF); err != nil {
		return err
	}
	return nil
}
