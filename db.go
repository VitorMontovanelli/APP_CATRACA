package main

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("catraca.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Falha ao abrir banco de dados: %v", err)
	}

	// Configura pragmas
	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA journal_mode = WAL")

	// AutoMigrate: cria/atualiza todas as tabelas
	err = db.AutoMigrate(
		&User{},
		&Aluno{},
		&RegistroAcesso{},
		&Student{},
		&Plan{},
		&PaymentMethod{},
		&PaymentGateway{},
		&StudentPlan{},
		&Invoice{},
		&AccessLog{},
		&AuditLog{},
		&Setting{},
		&Agendamento{},
		&CapacidadeDia{},
		&AgendamentoIndividual{},
	)
	if err != nil {
		log.Fatalf("Falha ao migrar banco de dados: %v", err)
	}

	// Migration manual: cria tabelas que podem não ter sido criadas pelo AutoMigrate
	db.Exec("CREATE TABLE IF NOT EXISTS capacidade_dias (data TEXT PRIMARY KEY, capacidade INTEGER NOT NULL)")
	db.Exec("CREATE TABLE IF NOT EXISTS agendamento_individuals (id INTEGER PRIMARY KEY AUTOINCREMENT, data TEXT, hora TEXT, student_id INTEGER, nome TEXT, pratica TEXT, observacao TEXT, created_at DATETIME)")
	db.Exec("CREATE TABLE IF NOT EXISTS settings (`key` TEXT PRIMARY KEY, value TEXT)")

	// Migração manual: adiciona colunas que podem não ter sido criadas pelo AutoMigrate
	if !db.Migrator().HasColumn(&User{}, "Foto") {
		db.Migrator().AddColumn(&User{}, "Foto")
	}
	if !db.Migrator().HasColumn(&Invoice{}, "Comprovante") {
		db.Migrator().AddColumn(&Invoice{}, "Comprovante")
	}
	if !db.Migrator().HasColumn(&Invoice{}, "PaidAt") {
		db.Migrator().AddColumn(&Invoice{}, "PaidAt")
	}
	if !db.Migrator().HasColumn(&Invoice{}, "PaidAmountCents") {
		db.Migrator().AddColumn(&Invoice{}, "PaidAmountCents")
	}
	if !db.Migrator().HasColumn(&Plan{}, "PrecoCartaoCents") {
		db.Migrator().AddColumn(&Plan{}, "PrecoCartaoCents")
	}
	if !db.Migrator().HasColumn(&Student{}, "TelefoneUrgencia") {
		db.Migrator().AddColumn(&Student{}, "TelefoneUrgencia")
	}
	if !db.Migrator().HasColumn(&Student{}, "LaudoMedico") {
		db.Migrator().AddColumn(&Student{}, "LaudoMedico")
	}
	if !db.Migrator().HasColumn(&Student{}, "Foto") {
		db.Migrator().AddColumn(&Student{}, "Foto")
	}

	seedData(db)
	migrateDefaultPlans(db)

	return db
}
