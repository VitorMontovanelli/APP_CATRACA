package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MigrateLegacyCapacityTable move os valores que o GORM gravava na tabela
// singular "capacidade_dia" para a tabela canônica "capacidade_dias", usada
// pelo calendário e pelas consultas SQL. Depois de copiar, remove a tabela
// antiga para que não existam duas fontes de capacidade.
func MigrateLegacyCapacityTable(db *gorm.DB) {
	var existe int64
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'capacidade_dia'").Scan(&existe).Error; err != nil || existe == 0 {
		return
	}

	var total int64
	if err := db.Raw("SELECT COUNT(*) FROM capacidade_dia").Scan(&total).Error; err != nil {
		log.Printf("Aviso: não foi possível verificar a tabela capacidade_dia: %v", err)
		return
	}
	if total == 0 {
		if err := db.Exec("DROP TABLE IF EXISTS capacidade_dia").Error; err != nil {
			log.Printf("Aviso: não foi possível remover a tabela capacidade_dia vazia: %v", err)
		}
		return
	}

	backup, err := gerarBackupMigracaoCapacidade(db)
	if err != nil {
		log.Printf("Aviso: não foi possível gerar backup antes de migrar capacidade_dia: %v", err)
		return
	}
	log.Printf("Backup do banco gerado em %s antes de migrar %d capacidade(s) de capacidade_dia", backup, total)

	if err := db.Exec(`INSERT OR IGNORE INTO capacidade_dias (data, calendario, capacidade)
		SELECT data, 'pilates', capacidade FROM capacidade_dia WHERE data IS NOT NULL`).Error; err != nil {
		log.Printf("Aviso: falha ao migrar capacidade_dia para capacidade_dias: %v", err)
		return
	}
	if err := db.Exec("DROP TABLE IF EXISTS capacidade_dia").Error; err != nil {
		log.Printf("Aviso: capacidade_dia foi migrada, mas a tabela antiga não pôde ser removida: %v", err)
		return
	}
	log.Printf("%d capacidade(s) migrada(s) de capacidade_dia para capacidade_dias", total)
}

func gerarBackupMigracaoCapacidade(db *gorm.DB) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	dir := filepath.Join(cwd, "backup_capacidade_legada_"+time.Now().Format("2006-01-02_150405"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "catraca.db")
	escaped := strings.ReplaceAll(path, "'", "''")
	if err := db.Exec("VACUUM INTO '" + escaped + "'").Error; err != nil {
		return "", fmt.Errorf("erro ao gerar snapshot: %w", err)
	}
	return path, nil
}
