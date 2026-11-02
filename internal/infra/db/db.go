package db

import (
	"log"

	"catraca-app/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("catraca.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Falha ao abrir banco de dados: %v", err)
	}

	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA journal_mode = WAL")

	autoMigrateModels(db)
	executeRawTableMigrations(db)
	migrateColumns(db)

	SeedData(db)
	MigrateDefaultPlans(db)

	return db
}

func autoMigrateModels(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.Aluno{},
		&models.RegistroAcesso{},
		&models.Student{},
		&models.Plan{},
		&models.PaymentMethod{},
		&models.PaymentGateway{},
		&models.StudentPlan{},
		&models.Invoice{},
		&models.AccessLog{},
		&models.AuditLog{},
		&models.Setting{},
		&models.Agendamento{},
		&models.CapacidadeDia{},
		&models.AgendamentoIndividual{},
	)
	if err != nil {
		log.Fatalf("Falha ao migrar banco de dados: %v", err)
	}
}

func executeRawTableMigrations(db *gorm.DB) {
	db.Exec("CREATE TABLE IF NOT EXISTS capacidade_dias (data TEXT PRIMARY KEY, capacidade INTEGER NOT NULL)")
	db.Exec("CREATE TABLE IF NOT EXISTS agendamento_individuals (id INTEGER PRIMARY KEY AUTOINCREMENT, data TEXT, hora TEXT, student_id INTEGER, nome TEXT, pratica TEXT, observacao TEXT, created_at DATETIME)")
	db.Exec("CREATE TABLE IF NOT EXISTS settings (`key` TEXT PRIMARY KEY, value TEXT)")
}

func migrateColumns(db *gorm.DB) {
	cols := []struct {
		dst    interface{}
		column string
	}{
		{&models.User{}, "Foto"},
		{&models.Invoice{}, "Comprovante"},
		{&models.Invoice{}, "PaidAt"},
		{&models.Invoice{}, "PaidAmountCents"},
		{&models.Plan{}, "PrecoCartaoCents"},
		{&models.Student{}, "TelefoneUrgencia"},
		{&models.Student{}, "LaudoMedico"},
		{&models.Student{}, "Foto"},
	}
	for _, c := range cols {
		if !db.Migrator().HasColumn(c.dst, c.column) {
			db.Migrator().AddColumn(c.dst, c.column)
		}
	}
}
