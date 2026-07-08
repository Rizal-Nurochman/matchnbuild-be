package migrations

import (
	"github.com/Rizal-Nurochman/matchnbuild/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20240613000000_add_client_message_id_to_messages", Up20240613000000AddClientMessageIDToMessages, Down20240613000000AddClientMessageIDToMessages)
}

func Up20240613000000AddClientMessageIDToMessages(db *gorm.DB) error {
	// Idempotent: AutoMigrate (database/migration.go) sudah membuat kolom & index
	// ini dari entity Message pada DB fresh, jadi pakai IF NOT EXISTS agar migrasi
	// tidak gagal di DB baru maupun lama.
	if err := db.Exec("ALTER TABLE messages ADD COLUMN IF NOT EXISTS client_message_id VARCHAR(36)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_sender_client_msg ON messages(sender_id, client_message_id)").Error; err != nil {
		return err
	}

	return nil
}

func Down20240613000000AddClientMessageIDToMessages(db *gorm.DB) error {
	if err := db.Exec("DROP INDEX IF EXISTS idx_sender_client_msg").Error; err != nil {
		return err
	}

	if err := db.Exec("ALTER TABLE messages DROP COLUMN IF EXISTS client_message_id").Error; err != nil {
		return err
	}

	return nil
}
