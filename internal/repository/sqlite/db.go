// Xomoi-Core: Sovereign Edge Node
// Copyright (C) 2026 Adrish Bora (@code-grey) & Simanjit Hujuri (@code-zephyrus)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	// Standard driver. For cross-compilation later without CGO, we can swap this for modernc.org/sqlite
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// DB is the SQLite engine wrapper
type DB struct {
	*sql.DB
}

// NewDB initializes a new SQLite connection with WAL mode enabled.
func NewDB(dbPath string) (*DB, error) {
	// DSN parameters for WAL, busy_timeout, and foreign keys
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_fk=1&_synchronous=NORMAL", dbPath)
	
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Apply migrations
	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &DB{db}, nil
}

func applyMigrations(db *sql.DB) error {
	content, err := migrationFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("migration 001 failed: %w", err)
	}
	
	log.Println("SQLite engine online. Migrations applied successfully.")
	return nil
}
