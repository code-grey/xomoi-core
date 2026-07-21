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
