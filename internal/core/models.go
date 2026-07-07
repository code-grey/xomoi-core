package core

import "time"

// User represents an owner/admin of the edge node.
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never serialize to JSON
	Role         string    `json:"role" db:"role"`       // e.g., admin, viewer
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Session represents an active user session (API or UI).
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"` // Hashed token for security
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Device represents an IoT node authenticated via HMAC-Lite.
type Device struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	MACAddress string    `json:"mac_address" db:"mac_address"`
	SecretKey  string    `json:"-" db:"secret_key"` // Used for HMAC-Lite signature verification
	LastSeen   time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// SensorTag maps a 1-byte field ID to a human-readable string and type.
// This is critical for protocol efficiency (sending ID '1' instead of string 'temperature').
type SensorTag struct {
	DeviceID string `json:"device_id" db:"device_id"`
	FieldID  uint8  `json:"field_id" db:"field_id"`
	TagName  string `json:"tag_name" db:"tag_name"`   // e.g., "temperature_c"
	DataType string `json:"data_type" db:"data_type"` // e.g., "float", "int", "bool"
}

// AlertRule defines a threshold condition for a specific sensor.
type AlertRule struct {
	ID        string    `json:"id" db:"id"`
	DeviceID  string    `json:"device_id" db:"device_id"`
	TagName   string    `json:"tag_name" db:"tag_name"`
	Condition string    `json:"condition" db:"condition"` // e.g., ">", "<", "=="
	Threshold float64   `json:"threshold" db:"threshold"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
