package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMigrateLegacyCapacityTableMoveValoresERemoveTabelaAntiga(t *testing.T) {
	dir := t.TempDir()
	anterior, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(anterior) }()

	database, err := gorm.Open(sqlite.Open(filepath.Join(dir, "teste.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("abrir banco: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("obter conexão: %v", err)
	}
	defer sqlDB.Close()

	if err := database.Exec("CREATE TABLE capacidade_dia (data TEXT PRIMARY KEY, capacidade INTEGER NOT NULL)").Error; err != nil {
		t.Fatalf("criar capacidade_dia: %v", err)
	}
	if err := database.Exec("CREATE TABLE capacidade_dias (data TEXT NOT NULL, calendario TEXT NOT NULL DEFAULT 'pilates', capacidade INTEGER NOT NULL, PRIMARY KEY (data, calendario))").Error; err != nil {
		t.Fatalf("criar capacidade_dias: %v", err)
	}
	if err := database.Exec("INSERT INTO capacidade_dia (data, capacidade) VALUES ('2026-11-01', 10)").Error; err != nil {
		t.Fatalf("inserir capacidade legada: %v", err)
	}

	MigrateLegacyCapacityTable(database)

	var capacidade int
	if err := database.Raw("SELECT capacidade FROM capacidade_dias WHERE data = '2026-11-01'").Scan(&capacidade).Error; err != nil {
		t.Fatalf("ler capacidade migrada: %v", err)
	}
	if capacidade != 10 {
		t.Fatalf("capacidade migrada = %d, esperado 10", capacidade)
	}

	var tabelaAntiga int64
	if err := database.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'capacidade_dia'").Scan(&tabelaAntiga).Error; err != nil {
		t.Fatalf("verificar tabela antiga: %v", err)
	}
	if tabelaAntiga != 0 {
		t.Fatal("tabela capacidade_dia ainda existe após a migração")
	}
}

func TestMigrateLegacyCapacityTableSemTabelaAntigaNaoFalha(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("abrir banco: %v", err)
	}

	MigrateLegacyCapacityTable(database)

	var total int64
	if err := database.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'capacidade_dia'").Scan(&total).Error; err != nil {
		t.Fatalf("verificar tabela: %v", err)
	}
	if total != 0 {
		t.Fatal("tabela capacidade_dia não deveria existir")
	}
}
