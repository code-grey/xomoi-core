package core

import "time"

// User represents an owner/admin of the edge node.
// PERFORMANCE: Fields are bit-packed (ordered largest to smallest by memory footprint)
// to prevent the compiler from injecting empty memory padding.
type User struct {
	CreatedAt    time.Time `json:"created_at" db:"created_at"` // 24 bytes
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"` // 24 bytes
	ID           string    `json:"id" db:"id"`                 // 16 bytes
	Username     string    `json:"username" db:"username"`     // 16 bytes
	PasswordHash string    `json:"-" db:"password_hash"`       // 16 bytes
	Role         string    `json:"role" db:"role"`             // 16 bytes
}

// Session represents an active user session (API or UI).
type Session struct {
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"` // 24 bytes
	CreatedAt time.Time `json:"created_at" db:"created_at"` // 24 bytes
	ID        string    `json:"id" db:"id"`                 // 16 bytes
	UserID    string    `json:"user_id" db:"user_id"`       // 16 bytes
	TokenHash string    `json:"-" db:"token_hash"`          // 16 bytes
}

// Device represents an IoT node authenticated via HMAC-Lite.
type Device struct {
	LastSeen   time.Time `json:"last_seen" db:"last_seen"`   // 24 bytes
	CreatedAt  time.Time `json:"created_at" db:"created_at"` // 24 bytes
	ID         string    `json:"id" db:"id"`                 // 16 bytes
	Name       string    `json:"name" db:"name"`             // 16 bytes
	MACAddress string    `json:"mac_address" db:"mac_address"`// 16 bytes
	SecretKey  string    `json:"-" db:"secret_key"`          // 16 bytes
}

// TelemetryPoint represents a single TSDB historical point
type TelemetryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Temp      *float64  `json:"temp,omitempty"`
	Hum       *float64  `json:"hum,omitempty"`
	State     string    `json:"state,omitempty"`
}

// SensorTag maps a 1-byte field ID to a human-readable string and type.
// This is critical for protocol efficiency (sending ID '1' instead of string 'temperature').
type SensorTag struct {
	DeviceID string `json:"device_id" db:"device_id"` // 16 bytes
	TagName  string `json:"tag_name" db:"tag_name"`   // 16 bytes
	DataType string `json:"data_type" db:"data_type"` // 16 bytes
	FieldID  uint8  `json:"field_id" db:"field_id"`   // 1 byte (Packed at end)
}

// AlertRule defines a threshold condition for a specific sensor.
type AlertRule struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"` // 24 bytes
	ID        string    `json:"id" db:"id"`                 // 16 bytes
	DeviceID  string    `json:"device_id" db:"device_id"`   // 16 bytes
	TagName   string    `json:"tag_name" db:"tag_name"`     // 16 bytes
	Condition string    `json:"condition" db:"condition"`   // 16 bytes
	Threshold float64   `json:"threshold" db:"threshold"`   // 8 bytes
	IsActive  bool      `json:"is_active" db:"is_active"`   // 1 byte (Packed at end)
}
