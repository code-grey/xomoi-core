package broker

import (
	"bytes"
	"log/slog"

	"github.com/buger/jsonparser"
	"github.com/code-grey/xomoi-core/internal/repository"
	"github.com/code-grey/xomoi-core/internal/state"
	"github.com/code-grey/xomoi-core/internal/worker"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Processor struct {
	hotState *state.HotState
	tsdb     repository.TelemetryRepository
	rules    *worker.RulesEngine
}

func NewProcessor(hs *state.HotState, tsdb repository.TelemetryRepository, rules *worker.RulesEngine) *Processor {
	return &Processor{hotState: hs, tsdb: tsdb, rules: rules}
}

// Process receives a pointer to the recycled Job struct.
func (p *Processor) Process(job *Job) (err error) {
	// 1. Zero-Panic Guarantee: Catch any unexpected panics during processing
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic in ingestion processor", "panic", r, "device", job.DeviceID)
		}
	}()

	// 2. The HotState update is O(1) and copies the bytes into the sync.Map safely.
	p.hotState.Update(job.DeviceID, job.Payload)

	// 3. Zero-Allocation JSON Parsing using buger/jsonparser
	// This entirely bypasses the Go heap and avoids struct reflection overhead
	temp, err := jsonparser.GetFloat(job.Payload, "temp")
	if err != nil {
		slog.Warn("Dropped malformed telemetry payload: missing temp", "error", err, "device", job.DeviceID)
		return nil
	}

	hum, err := jsonparser.GetFloat(job.Payload, "hum")
	if err != nil {
		slog.Warn("Dropped malformed telemetry payload: missing hum", "error", err, "device", job.DeviceID)
		return nil
	}

	stateVal, err := jsonparser.GetString(job.Payload, "state")
	if err != nil && err != jsonparser.KeyPathNotFoundError {
		slog.Warn("Error parsing state", "error", err, "device", job.DeviceID)
	}

	// 4. TSDB Insertion is DEFERRED to the HotState SnapshotWorker to protect the SD Card.
	// The HotState update above is all we need for immediate persistence.

	// 5. Zero-Allocation Rules Engine Evaluation
	if p.rules != nil {
		var pTemp, pHum *float64
		if temp != 0 || err == nil { pTemp = &temp }
		if hum != 0 || err == nil { pHum = &hum }
		
		p.rules.Evaluate(job.DeviceID, pTemp, pHum, stateVal)
	}

	return nil
}

type PublishHook struct {
	mqtt.HookBase
	pool *WorkerPool
}

func NewPublishHook(pool *WorkerPool) *PublishHook {
	return &PublishHook{pool: pool}
}

func (h *PublishHook) ID() string {
	return "xomoi-ingestion-hook"
}

func (h *PublishHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *PublishHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	deviceID := string(cl.Properties.Username)
	if deviceID == "" {
		deviceID = cl.ID
	}

	// Hand off to the WorkerPool which uses sync.Pool under the hood.
	// This delegates all allocations and copies entirely away from the MQTT thread.
	h.pool.Submit(deviceID, pk.Payload)

	return pk, nil
}
