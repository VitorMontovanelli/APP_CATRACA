package db

import (
	"log"

	"gorm.io/gorm"
)

// MigrateCapacityCalendars adiciona a dimensão de calendário à tabela
// capacidade_dias. Antes dela, a capacidade era compartilhada apenas pela data,
// então editar o Pilates afetava os calendários individuais (e vice-versa).
//
// Os valores existentes passam a pertencer ao calendário de Pilates. Os
// calendários individuais começam com a capacidade padrão (10) e podem ser
// ajustados individualmente depois.
func MigrateCapacityCalendars(db *gorm.DB) {
	var tabela int64
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'capacidade_dias'").Scan(&tabela).Error; err != nil || tabela == 0 {
		return
	}

	var temColuna int64
	if err := db.Raw("SELECT COUNT(*) FROM pragma_table_info('capacidade_dias') WHERE name = 'calendario'").Scan(&temColuna).Error; err != nil || temColuna > 0 {
		return
	}

	backup, err := gerarBackupMigracaoCapacidade(db)
	if err != nil {
		log.Printf("Aviso: não foi possível gerar backup antes de separar a capacidade por calendário: %v", err)
		return
	}
	log.Printf("Backup do banco gerado em %s antes de separar a capacidade por calendário", backup)

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("ALTER TABLE capacidade_dias RENAME TO capacidade_dias_legado").Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE TABLE capacidade_dias (
			data TEXT NOT NULL,
			calendario TEXT NOT NULL DEFAULT 'pilates',
			capacidade INTEGER NOT NULL,
			PRIMARY KEY (data, calendario)
		)`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT OR IGNORE INTO capacidade_dias (data, calendario, capacidade)
			SELECT data, 'pilates', capacidade FROM capacidade_dias_legado WHERE data IS NOT NULL`).Error; err != nil {
			return err
		}
		return tx.Exec("DROP TABLE IF EXISTS capacidade_dias_legado").Error
	})
	if err != nil {
		log.Printf("Aviso: falha ao separar a capacidade por calendário: %v", err)
		return
	}
	log.Printf("capacidade_dias migrada: registros antigos agora pertencem ao calendário pilates")
}
