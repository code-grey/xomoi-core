package state

import (
	"encoding/json"
	"sync"
	"time"
)

// DeviceState holds the latest telemetry payload for a device.
type DeviceState struct {
	DeviceID   string
	LastUpdate time.Time
	Payload    json.RawMessage // Unmarshaled from protobuf during ingestion
}

// HotState uses a concurrent map for O(1) reads/writes.
// This acts as the memory barrier protecting the SQLite disk from wear.
type HotState struct {
	sync.Map // key: deviceID (string), value: DeviceState
}

// NewHotState creates a new HotState manager.
func NewHotState() *HotState {
	return &HotState{}
}

// Update records the latest telemetry for a device.
func (h *HotState) Update(deviceID string, payload []byte) {
	h.Store(deviceID, DeviceState{
		DeviceID:   deviceID,
		LastUpdate: time.Now(),
		Payload:    payload,
	})
}

// Get retrieves the latest telemetry for a device.
func (h *HotState) Get(deviceID string) (DeviceState, bool) {
	val, ok := h.Load(deviceID)
	if !ok {
		return DeviceState{}, false
	}
	return val.(DeviceState), true
}

// GetAll returns a snapshot of all current device states.
func (h *HotState) GetAll() []DeviceState {
	var states []DeviceState
	h.Range(func(key, value any) bool {
		states = append(states, value.(DeviceState))
		return true
	})
	return states
}
