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
	"log"
	"sync"
)

// Job represents a raw telemetry payload to be processed.
type Job struct {
	DeviceID string
	Payload  []byte
}

// WorkerPool manages a fixed number of goroutines to process telemetry without sprawling.
type WorkerPool struct {
	jobs       chan *Job
	wg         sync.WaitGroup
	numWorkers int
	processor  *Processor
	jobPool    sync.Pool // The zero-allocation object pool
}

// NewWorkerPool creates a new constrained worker pool with a Zero-Allocation sync.Pool.
func NewWorkerPool(numWorkers int, bufferSize int, proc *Processor) *WorkerPool {
	return &WorkerPool{
		jobs:       make(chan *Job, bufferSize), // Buffered to handle backpressure
		numWorkers: numWorkers,
		processor:  proc,
		jobPool: sync.Pool{
			New: func() any {
				// Pre-allocate the struct. It will be reused infinitely.
				// This prevents the struct from ever escaping to the heap and triggering the GC.
				return &Job{}
			},
		},
	}
}

// Start spins up the fixed goroutines.
func (wp *WorkerPool) Start() {
	log.Printf("Starting GC-Optimized Ingestion Pool with %d workers", wp.numWorkers)
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker listens on the jobs channel.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for job := range wp.jobs {
		// Pass to the processor for update
		if err := wp.processor.Process(job); err != nil {
			log.Printf("[Worker %d] Failed to process job for device %s: %v", id, job.DeviceID, err)
		}
		
		// CRITICAL SAFETY RULES FOR SYNC.POOL:
		// 1. You MUST NOT hold any references to 'job' after this Put().
		// 2. You MUST NOT call Put() twice on the same object (fatal memory corruption).
		// By doing it exactly once at the bottom of the range loop, we guarantee safety.
		wp.jobPool.Put(job)
	}
}

// Submit grabs a recycled job from the pool, fills it, and enqueues it.
func (wp *WorkerPool) Submit(deviceID string, payload []byte) {
	// Borrow a job struct from the pool (Zero Heap Allocation)
	job := wp.jobPool.Get().(*Job)
	
	// DEFENSIVE ZEROING: 
	// We MUST overwrite every single field of the struct here. If we miss a field,
	// data from the previous request (which used this struct) will leak into the new request,
	// causing silent data corruption.
	job.DeviceID = deviceID
	job.Payload = payload 

	// NON-BLOCKING HANDOFF:
	// If 5,000 devices hit at once, the jobs channel will fill up. 
	// If we block here, Mochi-MQTT's read loops freeze, choking the OS TCP window.
	select {
	case wp.jobs <- job:
		// Queue accepted, worker will handle RingBuffer and HotState.
	default:
		// QUEUE FULL: Thundering Herd Detected!
		// We bypass the worker pool completely to protect the broker.
		// We update the HotState synchronously (O(1) fast sync.Map write) so the 
		// dashboard doesn't lose data, but we drop the RingBuffer insertion.
		if wp.processor != nil && wp.processor.hotState != nil {
			wp.processor.hotState.Update(deviceID, payload)
		}
		// Return the struct to the pool to prevent memory leaks
		wp.jobPool.Put(job)
	}
}

// Stop gracefully shuts down the pool.
func (wp *WorkerPool) Stop() {
	close(wp.jobs)
	wp.wg.Wait()
	log.Println("Ingestion Worker Pool stopped cleanly.")
}
