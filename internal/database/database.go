package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewConnection(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Migrate() error {
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		member_id TEXT UNIQUE NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		phone TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		membership_date DATETIME NOT NULL,
		membership_level TEXT NOT NULL CHECK (membership_level IN ('Gold', 'Silver', 'Bronze')),
		points INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	createTransfersTable := `
	CREATE TABLE IF NOT EXISTS transfers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_user_id INTEGER NOT NULL,
		to_user_id INTEGER NOT NULL,
		amount REAL NOT NULL CHECK (amount > 0 AND amount <= 2.0 AND ROUND(amount, 2) = amount),
		status TEXT NOT NULL CHECK (status IN ('pending','processing','completed','failed','cancelled','reversed')),
		note TEXT,
		idempotency_key TEXT NOT NULL UNIQUE,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		completed_at DATETIME,
		fail_reason TEXT,
		FOREIGN KEY (from_user_id) REFERENCES users(id),
		FOREIGN KEY (to_user_id) REFERENCES users(id)
	);`

	createPointLedgerTable := `
	CREATE TABLE IF NOT EXISTS point_ledger (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		change REAL NOT NULL,
		balance_after REAL NOT NULL,
		event_type TEXT NOT NULL CHECK (event_type IN ('transfer_out','transfer_in','adjust','earn','redeem')),
		transfer_id INTEGER,
		reference TEXT,
		metadata TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (transfer_id) REFERENCES transfers(id)
	);`

	// Create indexes
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_transfers_from ON transfers(from_user_id);",
		"CREATE INDEX IF NOT EXISTS idx_transfers_to ON transfers(to_user_id);",
		"CREATE INDEX IF NOT EXISTS idx_transfers_created ON transfers(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_ledger_user ON point_ledger(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_ledger_transfer ON point_ledger(transfer_id);",
		"CREATE INDEX IF NOT EXISTS idx_ledger_created ON point_ledger(created_at);",
	}

	// Execute migrations
	tables := []string{createUsersTable, createTransfersTable, createPointLedgerTable}
	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			log.Printf("Error creating table: %v", err)
			return err
		}
	}

	// Create indexes
	for _, index := range createIndexes {
		if _, err := db.Exec(index); err != nil {
			log.Printf("Error creating index: %v", err)
			return err
		}
	}

	log.Println("Database migration completed successfully")
	return nil
}
