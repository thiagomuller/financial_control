package database

import "database/sql"

func SeedSystemTags(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO tags (name, color, is_system, user_id)
		VALUES
			('Income',  '#16a34a', true, NULL),
			('Expense', '#dc2626', true, NULL)
		ON CONFLICT DO NOTHING
	`)
	return err
}
