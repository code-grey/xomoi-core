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
	"strings"
	"sync"
)

// Job represents a raw telemetry payload to be processed.
type Job struct {
	DeviceID string
	Payload  []byte
}

// WorkerPool manages a fixed number of goroutines to process telemetry without sprawling.
type WorkerPool struct {
	jobs         chan *Job
	criticalJobs chan *Job // High priority queue
	wg           sync.WaitGroup
	numWorkers   int
	processor    *Processor
	jobPool      sync.Pool
	strategy     string
}

// NewWorkerPool creates a new constrained worker pool with a Zero-Allocation sync.Pool.
func NewWorkerPool(numWorkers int, bufferSize int, strategy string, proc *Processor) *WorkerPool {
	return &WorkerPool{
		jobs:         make(chan *Job, bufferSize),
		criticalJobs: make(chan *Job, bufferSize),
		numWorkers:   numWorkers,
		processor:    proc,
		strategy:     strategy,
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

// worker listens on the jobs channels, prioritizing criticalJobs.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for {
		// Priority Check: Always drain critical jobs first
		select {
		case job, ok := <-wp.criticalJobs:
			if !ok {
				return
			}
			wp.processJob(id, job)
			continue // Skip to next iteration to re-check critical
		default:
		}

		// Normal Check: Wait for either
		select {
		case job, ok := <-wp.criticalJobs:
			if !ok {
				return
			}
			wp.processJob(id, job)
		case job, ok := <-wp.jobs:
			if !ok {
				return
			}
			wp.processJob(id, job)
		}
	}
}

func (wp *WorkerPool) processJob(id int, job *Job) {
	if err := wp.processor.Process(job); err != nil {
		log.Printf("[Worker %d] Failed to process job for device %s: %v", id, job.DeviceID, err)
	}
	wp.jobPool.Put(job)
}

// Submit grabs a recycled job from the pool, fills it, and enqueues it.
func (wp *WorkerPool) Submit(deviceID string, topic string, payload []byte) {
	job := wp.jobPool.Get().(*Job)
	
	job.DeviceID = deviceID
	job.Payload = payload 

	isCritical := strings.HasPrefix(topic, "/xomoi/critical")
	targetChan := wp.jobs
	if isCritical {
		targetChan = wp.criticalJobs
	}

	if wp.strategy == "drop" && !isCritical {
		select {
		case targetChan <- job:
		default:
			if wp.processor != nil && wp.processor.hotState != nil {
				wp.processor.hotState.Update(deviceID, payload)
			}
			wp.jobPool.Put(job)
		}
	} else {
		targetChan <- job
	}
}

// Stop gracefully shuts down the pool.
func (wp *WorkerPool) Stop() {
	close(wp.jobs)
	close(wp.criticalJobs)
	wp.wg.Wait()
	log.Println("Ingestion Worker Pool stopped cleanly.")
}
