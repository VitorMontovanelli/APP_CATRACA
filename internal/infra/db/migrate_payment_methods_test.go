package db

import (
	"strings"
	"testing"

	"catraca-app/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMigratePaymentMethodTypesPreservaDadosERemoveRestricao(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("abrir banco: %v", err)
	}

	if err := database.Exec(`CREATE TABLE payment_methods (
		id integer PRIMARY KEY AUTOINCREMENT,
		name text NOT NULL,
		type text NOT NULL,
		enabled numeric DEFAULT true,
		fee_percent real DEFAULT 0,
		fee_fixed_cents integer DEFAULT 0,
		created_at datetime,
		CONSTRAINT chk_payment_methods_type CHECK (type IN ('pix','credit_card','boleto','debit_card','wallet','cash'))
	)`).Error; err != nil {
		t.Fatalf("criar tabela antiga: %v", err)
	}
	if err := database.Exec(`INSERT INTO payment_methods (id, name, type, enabled, fee_percent, fee_fixed_cents) VALUES
		(1, 'Pix', 'pix', 1, 0, 0),
		(6, 'Transferência', 'pix', 1, 0, 0)`).Error; err != nil {
		t.Fatalf("inserir dados antigos: %v", err)
	}

	MigratePaymentMethodTypes(database)

	var schema string
	database.Raw("SELECT COALESCE(sql, '') FROM sqlite_master WHERE type = 'table' AND name = 'payment_methods'").Scan(&schema)
	if strings.Contains(schema, "chk_payment_methods_type") {
		t.Fatalf("restrição antiga ainda existe: %s", schema)
	}

	var methods []models.PaymentMethod
	if err := database.Order("id").Find(&methods).Error; err != nil {
		t.Fatalf("listar métodos: %v", err)
	}
	if len(methods) != 2 {
		t.Fatalf("quantidade de métodos = %d, esperado 2", len(methods))
	}
	if methods[0].ID != 1 || methods[0].Name != "Pix" {
		t.Fatalf("dados do Pix não preservados: %+v", methods[0])
	}
	if methods[1].ID != 6 || methods[1].Name != "Transferência" || methods[1].Type != "transfer" {
		t.Fatalf("Transferência não corrigida: %+v", methods[1])
	}

	// Tipos novos passam a ser aceitos após a reconstrução da tabela.
	if err := database.Create(&models.PaymentMethod{Name: "Outro", Type: "other"}).Error; err != nil {
		t.Fatalf("tipo other deveria ser aceito: %v", err)
	}
	if err := database.Create(&models.PaymentMethod{Name: "Débito", Type: "debit_card"}).Error; err != nil {
		t.Fatalf("tipo debit_card deveria ser aceito: %v", err)
	}
}
