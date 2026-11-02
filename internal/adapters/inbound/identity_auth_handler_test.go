package inbound

import (
	"context"
	"testing"

	"catraca-app/internal/infra/db"
	"catraca-app/internal/models"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to memory db: %v", err)
	}
	err = database.AutoMigrate(
		&models.User{},
		&models.AuditLog{},
		&models.Plan{},
		&models.PaymentMethod{},
		&models.PaymentGateway{},
	)
	if err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}
	return database
}

func TestAdminLoginSuccess(t *testing.T) {
	database := setupTestDB(t)
	db.SeedData(database)
	session := NewSessionManager(database)
	handler := NewIdentityAuthHandler(database, session)
	handler.SetContext(context.Background())

	user, err := handler.Login("admin", "admin")
	if err != nil {
		t.Fatalf("expected successful login, got: %v", err)
	}
	if user.Cargo != "super_admin" {
		t.Errorf("expected cargo super_admin, got: %s", user.Cargo)
	}
	if !user.Ativo {
		t.Errorf("expected user to be ativo")
	}
}

func TestAdminLoginSelfHealing(t *testing.T) {
	database := setupTestDB(t)
	// Insere um super admin com senha alterada/desconhecida
	wrongHash, _ := bcrypt.GenerateFromPassword([]byte("outrasenha"), bcrypt.DefaultCost)
	database.Create(&models.User{
		Name:       "Administrador",
		Email:      "admin",
		Senha:      string(wrongHash),
		Cargo:      "super_admin",
		Ativo:      false, // Desativado propositalmente
		Permissoes: "all",
	})

	session := NewSessionManager(database)
	handler := NewIdentityAuthHandler(database, session)
	handler.SetContext(context.Background())

	// Tenta logar com admin/admin -> deve auto-recuperar, reativar e atualizar hash
	user, err := handler.Login("admin", "admin")
	if err != nil {
		t.Fatalf("expected self-healing login to succeed, got: %v", err)
	}
	if !user.Ativo {
		t.Errorf("expected user to be reactivated")
	}

	// Verifica se a senha no banco foi atualizada para "admin"
	var updated models.User
	database.First(&updated, user.ID)
	if err := bcrypt.CompareHashAndPassword([]byte(updated.Senha), []byte("admin")); err != nil {
		t.Errorf("expected password in DB to be updated to bcrypt of admin: %v", err)
	}
}
