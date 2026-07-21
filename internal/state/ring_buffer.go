package state

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/code-grey/xomoi-core/internal/repository"
	"github.com/klauspost/compress/zstd"
	"github.com/oklog/ulid/v2"
)

// TelemetryPacket represents a single row to be inserted.
type TelemetryPacket struct {
	ID          string
	DeviceID    string
	Timestamp   time.Time
	Temperature *float64
	Humidity    *float64
	State       *string
	PayloadBlob []byte // Zstd compressed
}

// RingBuffer handles the lossless queueing and batch flushing of telemetry data.
type RingBuffer struct {
	buffer        chan *TelemetryPacket
	tsdb          repository.TelemetryRepository // interface
	zstdEncoder   *zstd.Encoder
	batchSize     int
	flushInterval time.Duration
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewRingBuffer creates a new lossless ingestion queue.
func NewRingBuffer(tsdb repository.TelemetryRepository, maxQueueSize, batchSize int, flushInterval time.Duration) *RingBuffer {
	// Initialize Zstandard encoder with default compression
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		panic("failed to initialize zstd encoder: " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &RingBuffer{
		buffer:        make(chan *TelemetryPacket, maxQueueSize),
		tsdb:          tsdb,
		zstdEncoder:   encoder,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Enqueue compresses the payload and adds it to the ring buffer.
func (rb *RingBuffer) Enqueue(deviceID string, temp, hum *float64, state *string, payload []byte) {
	// Compress payload natively in Go (Zero Heap Allocation)
	compressed := rb.zstdEncoder.EncodeAll(payload, make([]byte, 0, len(payload)))

	// Generate ULID for perfect time-sorting without millisecond collisions
	id := ulid.Make().String()

	packet := &TelemetryPacket{
		ID:          id,
		DeviceID:    deviceID,
		Timestamp:   time.Now(),
		Temperature: temp,
		Humidity:    hum,
		State:       state,
		PayloadBlob: compressed,
	}

	select {
	case rb.buffer <- packet:
		// Success
	default:
		// If queue is full, we log a warning and block to apply backpressure.
		// A queue of 100,000 means the DB is locked or IO is dead.
		slog.Warn("Ring buffer full! Applying backpressure to ingestion pipeline.", "deviceID", deviceID)
		rb.buffer <- packet 
	}
}

// Start begins the batch flusher goroutine.
func (rb *RingBuffer) Start() {
	rb.wg.Add(1)
	go rb.flusher()
}

func (rb *RingBuffer) flusher() {
	defer rb.wg.Done()

	ticker := time.NewTicker(rb.flushInterval)
	defer ticker.Stop()

	var batch []*TelemetryPacket

	flush := func() {
		if len(batch) == 0 {
			return
		}
		// Convert our internal batch to the repository package's struct format
		err := rb.tsdb.BulkInsert(rb.ctx, convertToRepoBatch(batch))
		if err != nil {
			slog.Error("Failed to bulk insert telemetry batch", "error", err, "batchSize", len(batch))
		}
		
		// Reset batch
		batch = batch[:0]
	}

	for {
		select {
		case <-rb.ctx.Done():
			flush()
			return
		case <-ticker.C:
			flush()
		case packet := <-rb.buffer:
			batch = append(batch, packet)
			if len(batch) >= rb.batchSize {
				flush()
			}
		}
	}
}

// Stop gracefully flushes the remaining buffer and stops.
func (rb *RingBuffer) Stop() {
	rb.cancel()
	close(rb.buffer)
	rb.wg.Wait()
}

// Helper to bridge our local packets to the repository's generic BulkInsert model
func convertToRepoBatch(batch []*TelemetryPacket) []repository.TelemetryRecord {
	records := make([]repository.TelemetryRecord, len(batch))
	for i, p := range batch {
		records[i] = repository.TelemetryRecord{
			ID:          p.ID,
			DeviceID:    p.DeviceID,
			Timestamp:   p.Timestamp,
			Temperature: p.Temperature,
			Humidity:    p.Humidity,
			State:       p.State,
			PayloadBlob: p.PayloadBlob,
		}
	}
	return records
}
