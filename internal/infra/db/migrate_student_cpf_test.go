package db

import (
	"os"
	"path/filepath"
	"testing"

	"catraca-app/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrateEmptyStudentCPFToNull(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(old)

	database, err := gorm.Open(sqlite.Open(filepath.Join(dir, "catraca.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.AutoMigrate(&models.Student{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	defer sqlDB.Close()

	// Simula um banco legado, anterior ao índice único, com vários CPFs vazios.
	if err := database.Exec("DROP INDEX IF EXISTS idx_students_cpf").Error; err != nil {
		t.Fatalf("drop index: %v", err)
	}
	// Insere direto via SQL para simular dados antigos com CPF vazio.
	if err := database.Exec("INSERT INTO students (nome, cpf) VALUES (?, '')", "Ana").Error; err != nil {
		t.Fatalf("insert ana: %v", err)
	}
	if err := database.Exec("INSERT INTO students (nome, cpf) VALUES (?, '')", "Bia").Error; err != nil {
		t.Fatalf("insert bia: %v", err)
	}

	MigrateEmptyStudentCPFToNull(database)

	var nulos int64
	if err := database.Raw("SELECT COUNT(*) FROM students WHERE cpf IS NULL").Scan(&nulos).Error; err != nil {
		t.Fatalf("count null: %v", err)
	}
	if nulos != 2 {
		t.Fatalf("esperava 2 CPFs NULL, encontrou %d", nulos)
	}

	backups, err := filepath.Glob(filepath.Join(dir, "backup_cpf_vazio_*", "catraca.db"))
	if err != nil {
		t.Fatalf("glob backup: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("esperava 1 backup do banco, encontrou %d", len(backups))
	}
}

func TestMigrateEmptyStudentCPFToNullSemDadosNaoCriaBackup(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(old)

	database, err := gorm.Open(sqlite.Open(filepath.Join(dir, "catraca.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.AutoMigrate(&models.Student{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	defer sqlDB.Close()
	if err := database.Exec("INSERT INTO students (nome, cpf) VALUES (?, ?)", "Ana", "147.089.317-78").Error; err != nil {
		t.Fatalf("insert: %v", err)
	}

	MigrateEmptyStudentCPFToNull(database)

	backups, err := filepath.Glob(filepath.Join(dir, "backup_cpf_vazio_*"))
	if err != nil {
		t.Fatalf("glob backup: %v", err)
	}
	if len(backups) != 0 {
		t.Fatalf("não deveria criar backup sem CPF vazio, encontrou %d", len(backups))
	}
}
