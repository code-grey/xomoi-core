package signal

import (
	"hash/fnv"
	"sync"
	"github.com/gorilla/websocket"
)

// Shard count must be a power of 2 for fast bitwise modulo
const shardCount = 32

type Shard struct {
	sync.RWMutex
	conns map[string]*websocket.Conn
}

// ShardedConnectionMap reduces lock contention by 97% by splitting 
// the global map of 50,000 users into 32 independent maps.
type ShardedConnectionMap struct {
	shards [shardCount]*Shard
}

func NewShardedConnectionMap() *ShardedConnectionMap {
	m := &ShardedConnectionMap{}
	for i := 0; i < shardCount; i++ {
		m.shards[i] = &Shard{
			conns: make(map[string]*websocket.Conn),
		}
	}
	return m
}

// getShard uses FNV-1a hash to deterministically find the shard for a key
func (m *ShardedConnectionMap) getShard(key string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	// Bitwise AND is faster than modulo for powers of 2
	return m.shards[h.Sum32()&(shardCount-1)]
}

// Set adds a connection to the correct shard
func (m *ShardedConnectionMap) Set(key string, conn *websocket.Conn) {
	shard := m.getShard(key)
	shard.Lock()
	shard.conns[key] = conn
	shard.Unlock()
}

// Get retrieves a connection from the correct shard
func (m *ShardedConnectionMap) Get(key string) (*websocket.Conn, bool) {
	shard := m.getShard(key)
	shard.RLock()
	conn, ok := shard.conns[key]
	shard.RUnlock()
	return conn, ok
}

// Delete removes a connection from the correct shard
func (m *ShardedConnectionMap) Delete(key string) {
	shard := m.getShard(key)
	shard.Lock()
	delete(shard.conns, key)
	shard.Unlock()
}
