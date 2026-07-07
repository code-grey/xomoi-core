package broker

import (
	"bytes"

	"github.com/code-grey/xomoi-core/internal/state"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Processor struct {
	hotState *state.HotState
}

func NewProcessor(hs *state.HotState) *Processor {
	return &Processor{hotState: hs}
}

// Process receives a pointer to the recycled Job struct.
func (p *Processor) Process(job *Job) error {
	// The HotState update is O(1) and copies the bytes into the sync.Map safely.
	p.hotState.Update(job.DeviceID, job.Payload)
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
