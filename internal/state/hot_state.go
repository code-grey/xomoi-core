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

// Shard count for HotState (Must be power of 2)
const shardCount = 256

type StateShard struct {
	sync.RWMutex
	devices map[string]DeviceState
}

// HotState uses a sharded map for ultra-fast concurrent writes.
// This acts as the memory barrier protecting the SQLite disk from wear.
type HotState struct {
	shards [shardCount]*StateShard
}

// NewHotState creates a new 16-shard HotState manager.
func NewHotState() *HotState {
	h := &HotState{}
	for i := 0; i < shardCount; i++ {
		h.shards[i] = &StateShard{
			devices: make(map[string]DeviceState),
		}
	}
	return h
}

// getShard uses FNV-1a to find the correct shard deterministically
func (h *HotState) getShard(deviceID string) *StateShard {
	var hash uint32 = 2166136261
	for i := 0; i < len(deviceID); i++ {
		hash ^= uint32(deviceID[i])
		hash *= 16777619
	}
	return h.shards[hash&(shardCount-1)]
}

// Update records the latest telemetry for a device (Locks only 1/16th of the map).
func (h *HotState) Update(deviceID string, payload []byte) {
	shard := h.getShard(deviceID)
	
	newState := DeviceState{
		DeviceID:   deviceID,
		LastUpdate: time.Now(),
		Payload:    payload,
	}

	shard.Lock()
	shard.devices[deviceID] = newState
	shard.Unlock()
}

// Get retrieves the latest telemetry for a device.
func (h *HotState) Get(deviceID string) (DeviceState, bool) {
	shard := h.getShard(deviceID)
	
	shard.RLock()
	val, ok := shard.devices[deviceID]
	shard.RUnlock()
	
	return val, ok
}

// GetAll returns a snapshot of all current device states.
func (h *HotState) GetAll() []DeviceState {
	var states []DeviceState
	
	// We must lock each shard momentarily to safely copy the data
	for i := 0; i < shardCount; i++ {
		shard := h.shards[i]
		shard.RLock()
		for _, state := range shard.devices {
			states = append(states, state)
		}
		shard.RUnlock()
	}
	
	return states
}
