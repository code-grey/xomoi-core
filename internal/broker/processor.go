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

package broker

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"

	"github.com/buger/jsonparser"
	"github.com/code-grey/xomoi-core/internal/state"
	"github.com/code-grey/xomoi-core/internal/worker"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"golang.org/x/time/rate"
)

type Processor struct {
	hotState *state.HotState
	ringBuf  *state.RingBuffer
	rules    *worker.RulesEngine
}

func NewProcessor(hs *state.HotState, rb *state.RingBuffer, rules *worker.RulesEngine) *Processor {
	return &Processor{hotState: hs, ringBuf: rb, rules: rules}
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

	// 3. Opportunistic Zero-Allocation JSON Parsing
	// If the payload is JSON, we extract specific fields for the Rules Engine.
	// If it's NOT JSON (e.g. Protobuf or raw bytes), we silently ignore the error
	// and pass the raw byte stream directly to the Ring Buffer (Phase 1.1 Support).
	var pTemp, pHum *float64
	var pState *string

	temp, errTemp := jsonparser.GetFloat(job.Payload, "temp")
	if errTemp == nil {
		pTemp = &temp
	}

	hum, errHum := jsonparser.GetFloat(job.Payload, "hum")
	if errHum == nil {
		pHum = &hum
	}

	stateVal, errState := jsonparser.GetString(job.Payload, "state")
	if errState == nil {
		pState = &stateVal
	}

	// 5. Lossless Ring Buffer TSDB Insertion
	// We push the raw payload and extracted fields to the ring buffer.
	// It will natively compress it with Zstd and bulk insert it into SQLite asynchronously.
	if p.ringBuf != nil {
		p.ringBuf.Enqueue(job.DeviceID, pTemp, pHum, pState, job.Payload)
	}

	// 6. Zero-Allocation Rules Engine Evaluation
	if p.rules != nil {
		p.rules.Evaluate(job.DeviceID, pTemp, pHum, stateVal)
	}

	return nil
}

type PublishHook struct {
	mqtt.HookBase
	pool   *WorkerPool
	limits sync.Map // map[string]*rate.Limiter
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

	isCritical := strings.HasPrefix(pk.TopicName, "/xomoi/critical")

	// FAIRNESS: Token Bucket Rate Limiting (100 msgs/sec limit per device)
	// We bypass the limiter for critical topics.
	if !isCritical {
		var limiter *rate.Limiter
		if val, ok := h.limits.Load(deviceID); ok {
			limiter = val.(*rate.Limiter)
		} else {
			limiter = rate.NewLimiter(rate.Limit(100), 50)
			h.limits.Store(deviceID, limiter)
		}

		if !limiter.Allow() {
			// Noisy Neighbor detected! Silently drop the packet to protect the broker.
			return pk, nil
		}
	}

	// Hand off to the WorkerPool which uses sync.Pool under the hood.
	h.pool.Submit(deviceID, pk.TopicName, pk.Payload)

	return pk, nil
}
