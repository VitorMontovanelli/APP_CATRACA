package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMigrateCapacityCalendarsSeparaColuna(t *testing.T) {
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

	if err := database.Exec("CREATE TABLE capacidade_dias (data TEXT PRIMARY KEY, capacidade INTEGER NOT NULL)").Error; err != nil {
		t.Fatalf("criar tabela antiga: %v", err)
	}
	if err := database.Exec("INSERT INTO capacidade_dias (data, capacidade) VALUES ('2026-11-02', 22)").Error; err != nil {
		t.Fatalf("inserir capacidade antiga: %v", err)
	}

	MigrateCapacityCalendars(database)

	var capacidade int
	if err := database.Raw("SELECT capacidade FROM capacidade_dias WHERE data = '2026-11-02' AND calendario = 'pilates'").Scan(&capacidade).Error; err != nil {
		t.Fatalf("ler capacidade migrada: %v", err)
	}
	if capacidade != 22 {
		t.Fatalf("capacidade migrada = %d, esperado 22", capacidade)
	}

	var colunas int64
	if err := database.Raw("SELECT COUNT(*) FROM pragma_table_info('capacidade_dias') WHERE name = 'calendario'").Scan(&colunas).Error; err != nil {
		t.Fatalf("verificar coluna: %v", err)
	}
	if colunas != 1 {
		t.Fatal("coluna calendario não foi criada")
	}

	// Rodar de novo não deve duplicar nem apagar os dados.
	MigrateCapacityCalendars(database)
	var total int64
	if err := database.Raw("SELECT COUNT(*) FROM capacidade_dias").Scan(&total).Error; err != nil {
		t.Fatalf("contar registros: %v", err)
	}
	if total != 1 {
		t.Fatalf("registros após segunda migração = %d, esperado 1", total)
	}
}
