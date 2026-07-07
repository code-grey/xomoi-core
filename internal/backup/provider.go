package backup

import "context"

// Provider defines the Hexagonal Interface for disaster recovery and cloud syncing.
// Xomoi-Core can hot-swap these implementations via environment variables without 
// altering any business logic.
type Provider interface {
	// Save executes the backup strategy (e.g., snapshotting SQLite and uploading, or syncing rows).
	Save(ctx context.Context, localDBPath string) error
	
	// Restore pulls the latest backup from the target and rehydrates the local edge node.
	Restore(ctx context.Context, localDBPath string) error
}
