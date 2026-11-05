package db

import (
	"log"
	"strings"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

// MigratePaymentMethodTypes remove a restrição antiga de tipos da tabela
// payment_methods, que não permitia "transfer" nem "other". A tabela é
// reconstruída preservando IDs, dados e a sequência de autoincremento.
func MigratePaymentMethodTypes(db *gorm.DB) {
	if !db.Migrator().HasTable("payment_methods") {
		return
	}

	var schema string
	if err := db.Raw("SELECT COALESCE(sql, '') FROM sqlite_master WHERE type = 'table' AND name = 'payment_methods'").Scan(&schema).Error; err != nil {
		return
	}
	if strings.Contains(schema, "chk_payment_methods_type") {
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE payment_methods RENAME TO payment_methods_old").Error; err != nil {
				return err
			}
			if err := tx.Exec(`CREATE TABLE payment_methods (
				id integer PRIMARY KEY AUTOINCREMENT,
				name text NOT NULL,
				type text NOT NULL,
				enabled numeric DEFAULT true,
				fee_percent real DEFAULT 0,
				fee_fixed_cents integer DEFAULT 0,
				created_at datetime
			)`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`INSERT INTO payment_methods (id, name, type, enabled, fee_percent, fee_fixed_cents, created_at)
				SELECT id, name, type, enabled, fee_percent, fee_fixed_cents, created_at FROM payment_methods_old`).Error; err != nil {
				return err
			}
			return tx.Exec("DROP TABLE payment_methods_old").Error
		})
		if err != nil {
			log.Printf("falha ao atualizar restrição de tipos dos métodos de pagamento: %v", err)
		}
	}

	// A forma "Transferência" era gravada como "pix"; passa a ter tipo próprio
	// para aparecer corretamente na tela e nos relatórios.
	db.Model(&models.PaymentMethod{}).
		Where("name = ? AND type <> ?", "Transferência", "transfer").
		Update("type", "transfer")
}
