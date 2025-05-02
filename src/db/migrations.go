package db

import (
	"database/sql"
	"fmt"
)

func migrateDatabase(db *sql.DB) error {
	// Check if status column exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name = 'status'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for status column: %v", err)
	}

	// Add status column if it doesn't exist
	if count == 0 {
		_, err := db.Exec("ALTER TABLE tasks ADD COLUMN status INTEGER NOT NULL DEFAULT 0")
		if err != nil {
			return fmt.Errorf("failed to add status column: %v", err)
		}
	}

	return nil
}
