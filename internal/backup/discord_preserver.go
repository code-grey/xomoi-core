package backup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// DiscordPreserver implements Provider using a Discord Webhook for 100% free snapshot storage.
type DiscordPreserver struct {
	WebhookURL string
}

func NewDiscordPreserver(webhookURL string) *DiscordPreserver {
	return &DiscordPreserver{WebhookURL: webhookURL}
}

func (p *DiscordPreserver) Save(ctx context.Context, localDBPath string) error {
	file, err := os.Open(localDBPath)
	if err != nil {
		return fmt.Errorf("failed to open database file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Attach the SQLite file to the payload
	part, err := writer.CreateFormFile("file", filepath.Base(localDBPath))
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", p.WebhookURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("discord network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord webhook rejected payload with status: %d", resp.StatusCode)
	}

	return nil
}

func (p *DiscordPreserver) Restore(ctx context.Context, localDBPath string) error {
	// Restoring from Discord requires parsing the channel history via the Bot API, 
	// locating the last message with a valid `.db` attachment, and downloading it.
	return fmt.Errorf("not implemented: discord restore requires Bot API integration to fetch channel history")
}
