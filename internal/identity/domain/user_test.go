package domain_test

import (
	"errors"
	"testing"

	"catraca-app/internal/identity/domain"
)

func TestHashPasswordAdmin(t *testing.T) {
	// Valida que senhas menores que 4 caracteres falham
	_, err := domain.HashPassword("123")
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Fatalf("esperava ErrPasswordTooShort para senha '123', obteve %v", err)
	}

	// Valida que a senha 'admin' (5 caracteres) é aceita
	hash, err := domain.HashPassword("admin")
	if err != nil {
		t.Fatalf("esperava sucesso ao hashear 'admin', obteve erro: %v", err)
	}

	// Valida autenticação com 'admin'
	if !domain.CheckPasswordHash("admin", hash) {
		t.Fatalf("esperava que a senha 'admin' conferisse com o hash gerado")
	}

	// Valida rejeição de senha incorreta
	if domain.CheckPasswordHash("adminErrado", hash) {
		t.Fatalf("senha incorreta não deveria ser validada")
	}
}

