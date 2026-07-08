package signal

import (
	"sync"
)

// MessageBuffer represents a reusable byte slice to prevent heap allocations
type MessageBuffer struct {
	Data []byte
}

// Global sync.Pool for WebSocket payloads
// By recycling these buffers, we prevent the Garbage Collector from 
// trashing the heap every time a WebRTC SDP offer passes through the server.
var BufferPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate 2KB (enough for standard SDP offers and ICE candidates)
		return &MessageBuffer{
			Data: make([]byte, 2048), 
		}
	},
}

// GetBuffer retrieves a clean buffer from the pool
func GetBuffer() *MessageBuffer {
	buf := BufferPool.Get().(*MessageBuffer)
	// We do NOT clear the data here for performance. 
	// The reader will slice it up to the bytes actually read: buf.Data[:n]
	return buf
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *MessageBuffer) {
	// Re-slice to max capacity before returning
	buf.Data = buf.Data[:cap(buf.Data)]
	BufferPool.Put(buf)
}
