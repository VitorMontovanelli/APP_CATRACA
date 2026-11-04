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

// MigrateEmptyStudentCPFToNull converte CPFs vazios ("") em NULL. O índice
// único de students.cpf aceita vários NULL, mas apenas um texto vazio.
// Antes de alterar qualquer registro, gera um snapshot do banco em disco.
func MigrateEmptyStudentCPFToNull(db *gorm.DB) {
	const somenteTextoVazio = "cpf IS NOT NULL AND TRIM(cpf) = ''"

	var total int64
	if err := db.Raw("SELECT COUNT(*) FROM students WHERE " + somenteTextoVazio).Scan(&total).Error; err != nil {
		log.Printf("Aviso: não foi possível verificar CPFs vazios: %v", err)
		return
	}
	if total == 0 {
		return
	}

	backup, err := gerarBackupMigracaoCPF(db)
	if err != nil {
		log.Printf("Aviso: não foi possível gerar backup antes de corrigir CPFs vazios: %v", err)
		return
	}
	log.Printf("Backup do banco gerado em %s antes de converter %d CPF(s) vazio(s) em NULL", backup, total)

	if err := db.Exec("UPDATE students SET cpf = NULL WHERE " + somenteTextoVazio).Error; err != nil {
		log.Printf("Aviso: falha ao converter CPFs vazios em NULL: %v", err)
		return
	}
	log.Printf("%d CPF(s) vazio(s) convertido(s) em NULL", total)
}

func gerarBackupMigracaoCPF(db *gorm.DB) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	dir := filepath.Join(cwd, "backup_cpf_vazio_"+time.Now().Format("2006-01-02_150405"))
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
