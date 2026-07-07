package repository

import (
	"context"

	"github.com/code-grey/xomoi-core/internal/core"
)

// UserRepository handles persistence for User entities.
type UserRepository interface {
	Create(ctx context.Context, user *core.User) error
	GetByID(ctx context.Context, id string) (*core.User, error)
	GetByUsername(ctx context.Context, username string) (*core.User, error)
	Update(ctx context.Context, user *core.User) error
	Delete(ctx context.Context, id string) error
}

// SessionRepository handles API and UI session tokens.
type SessionRepository interface {
	Create(ctx context.Context, session *core.Session) error
	GetByID(ctx context.Context, id string) (*core.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

// DeviceRepository manages IoT edge devices.
type DeviceRepository interface {
	Create(ctx context.Context, device *core.Device) error
	GetByID(ctx context.Context, id string) (*core.Device, error)
	GetByMAC(ctx context.Context, macAddress string) (*core.Device, error)
	UpdateLastSeen(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

// SensorTagRepository manages the mapping between Field IDs and tag names.
type SensorTagRepository interface {
	Upsert(ctx context.Context, tag *core.SensorTag) error
	GetByDevice(ctx context.Context, deviceID string) ([]*core.SensorTag, error)
	GetByFieldID(ctx context.Context, deviceID string, fieldID uint8) (*core.SensorTag, error)
}

// TelemetryRepository is responsible for reading bulk telemetry data.
// Note: Write operations are typically handled by the Hot State engine and bulk-flushed.
type TelemetryRepository interface {
	// GetDeviceHistory retrieves historical telemetry points for a specific device and time range.
	GetDeviceHistory(ctx context.Context, deviceID string, start, end int64) ([]byte, error)
}

// AlertRuleRepository manages threshold rules for devices.
type AlertRuleRepository interface {
	Create(ctx context.Context, rule *core.AlertRule) error
	GetByDevice(ctx context.Context, deviceID string) ([]*core.AlertRule, error)
	Update(ctx context.Context, rule *core.AlertRule) error
	Delete(ctx context.Context, id string) error
}
