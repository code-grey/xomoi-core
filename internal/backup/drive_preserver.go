package backup

import (
	"context"
	"fmt"
)

// DrivePreserver implements Provider using Google Drive API and a Service Account.
type DrivePreserver struct {
	ServiceAccountJSON []byte
	FolderID           string
}

func NewDrivePreserver(saJSON []byte, folderID string) *DrivePreserver {
	return &DrivePreserver{
		ServiceAccountJSON: saJSON,
		FolderID:           folderID,
	}
}

func (p *DrivePreserver) Save(ctx context.Context, localDBPath string) error {
	// Strategy:
	// 1. Authenticate using google.golang.org/api/drive/v3 with ServiceAccountJSON.
	// 2. Upload the binary stream of localDBPath to the specified FolderID.
	return fmt.Errorf("not implemented: Google Drive API upload logic")
}

func (p *DrivePreserver) Restore(ctx context.Context, localDBPath string) error {
	// Strategy:
	// 1. Authenticate.
	// 2. Query FolderID for the latest backup file (ordered by createdTime).
	// 3. Download the stream and overwrite localDBPath locally.
	return fmt.Errorf("not implemented: Google Drive API download logic")
}
