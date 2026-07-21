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
