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
