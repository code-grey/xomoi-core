package backup

import (
	"context"
	"database/sql"
	"fmt"
)

// PostgresPreserver implements Provider by syncing telemetry rows to a cloud PostgreSQL instance (Neon/Supabase).
type PostgresPreserver struct {
	PGConnString string
}

func NewPostgresPreserver(connString string) *PostgresPreserver {
	return &PostgresPreserver{PGConnString: connString}
}

func (p *PostgresPreserver) Save(ctx context.Context, localDBPath string) error {
	// Unlike the snapshot preservers, this executes the "Store-and-Forward" pattern.
	// It reads rows from local SQLite that haven't been synced yet, and executes
	// batch INSERTs into the cloud Postgres database, keeping bandwidth usage minimal.
	return fmt.Errorf("not implemented: Postgres incremental sync logic")
}

func (p *PostgresPreserver) Restore(ctx context.Context, localDBPath string) error {
	// Connect to Postgres cloud, pull historical data, and run bulk INSERTs into local SQLite
	// to re-hydrate the edge node if the ephemeral disk was wiped.
	pgDB, err := sql.Open("postgres", p.PGConnString)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pgDB.Close()

	if err := pgDB.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}

	// Execution of hydration queries goes here.
	return nil
}
