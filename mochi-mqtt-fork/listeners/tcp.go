// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 mochi-mqtt, mochi-co
// SPDX-FileContributor: mochi-co

package listeners

import (
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"log/slog"
)

const TypeTCP = "tcp"

// TCP is a listener for establishing client connections on basic TCP protocol.
type TCP struct { // [MQTT-4.2.0-1]
	sync.RWMutex
	id      string       // the internal id of the listener
	address string       // the network address to bind to
	listen  net.Listener // a net.Listener which will listen for new clients
	config  Config       // configuration values for the listener
	log     *slog.Logger // server logger
	end     uint32       // ensure the close methods are only called once
}

// NewTCP initializes and returns a new TCP listener, listening on an address.
func NewTCP(config Config) *TCP {
	return &TCP{
		id:      config.ID,
		address: config.Address,
		config:  config,
	}
}

// ID returns the id of the listener.
func (l *TCP) ID() string {
	return l.id
}

// Address returns the address of the listener.
func (l *TCP) Address() string {
	if l.listen != nil {
		return l.listen.Addr().String()
	}
	return l.address
}

// Protocol returns the address of the listener.
func (l *TCP) Protocol() string {
	return "tcp"
}

// Init initializes the listener.
func (l *TCP) Init(log *slog.Logger) error {
	l.log = log

	var err error
	if l.config.TLSConfig != nil {
		l.listen, err = tls.Listen("tcp", l.address, l.config.TLSConfig)
	} else {
		l.listen, err = net.Listen("tcp", l.address)
	}

	return err
}

// Serve starts waiting for new TCP connections, and calls the establish
// connection callback for any received.
func (l *TCP) Serve(establish EstablishFn) {
	// Create a buffered channel to act as the User-Space TCP Backlog (bypassing OS SOMAXCONN limits)
	// We use 8192 as a massive buffer to absorb thundering herds.
	connChan := make(chan net.Conn, 8192)

	// Start a fixed pool of 100 "Authenticator" workers to process connections.
	// This prevents Goroutine Sprawl when 5,000 devices connect simultaneously.
	workerCount := 100
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for conn := range connChan {
				// We enforce an 8-second deadline for the *initial* CONNECT packet to protect against Slowloris.
				// After the handshake succeeds, Mochi-MQTT automatically resets this to the Keepalive timer.
				conn.SetReadDeadline(time.Now().Add(8 * time.Second))

				err := establish(l.id, conn)
				if err != nil {
					l.log.Warn("connection establish failed", "error", err)
				}
			}
		}()
	}

	for {
		if atomic.LoadUint32(&l.end) == 1 {
			break
		}

		conn, err := l.listen.Accept()
		if err != nil {
			break
		}

		if atomic.LoadUint32(&l.end) == 0 {
			// Fast-path: push to the buffer instead of spawning a goroutine.
			select {
			case connChan <- conn:
				// Successfully buffered in user-space!
			default:
				// If the 8192 buffer is FULL, we are under a severe DoS attack.
				// Drop the connection to protect the broker.
				l.log.Warn("user-space backlog full, dropping connection to prevent DoS")
				conn.Close()
			}
		}
	}

	// Graceful shutdown
	close(connChan)
	wg.Wait()
}

// Close closes the listener and any client connections.
func (l *TCP) Close(closeClients CloseFn) {
	l.Lock()
	defer l.Unlock()

	if atomic.CompareAndSwapUint32(&l.end, 0, 1) {
		closeClients(l.id)
	}

	if l.listen != nil {
		err := l.listen.Close()
		if err != nil {
			return
		}
	}
}
