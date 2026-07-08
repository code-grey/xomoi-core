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

	wp.jobs <- job
}

// Stop gracefully shuts down the pool.
func (wp *WorkerPool) Stop() {
	close(wp.jobs)
	wp.wg.Wait()
	log.Println("Ingestion Worker Pool stopped cleanly.")
}
