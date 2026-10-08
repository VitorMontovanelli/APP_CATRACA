package domain

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("e-mail ou senha inválidos")
	ErrUserInactive       = errors.New("usuário inativo no sistema")
	ErrPasswordTooShort   = errors.New("a senha deve ter no mínimo 6 caracteres")
	ErrInvalidRole        = errors.New("cargo inválido: permitido super_admin ou admin")
	ErrPermissionDenied   = errors.New("permissão negada para este recurso")
)

type User struct {
	ID            uint
	Name          string
	Email         string
	SenhaHash     string
	Cargo         string // super_admin, admin
	RegistradorID *uint
	Ativo         bool
	Foto          *string
	Permissoes    string
	CriadoEm      time.Time
}

type AuditLogRecord struct {
	ID        uint
	ActorID   uint
	Action    string
	TargetID  *uint
	Details   string
	Timestamp time.Time
}

func HashPassword(plainPassword string) (string, error) {
	if len(plainPassword) < 6 {
		return "", ErrPasswordTooShort
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 12)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func HasPermission(userPermissoes, required string) bool {
	if userPermissoes == "" {
		return false
	}
	parts := strings.Split(userPermissoes, ",")
	for _, p := range parts {
		if strings.TrimSpace(p) == required {
			return true
		}
	}
	return false
}
