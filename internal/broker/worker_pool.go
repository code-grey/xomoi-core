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
		
		// CRITICAL OPTIMIZATION: Return the object to the pool so the GC never has to clean it up.
		wp.jobPool.Put(job)
	}
}

// Submit grabs a recycled job from the pool, fills it, and enqueues it.
func (wp *WorkerPool) Submit(deviceID string, payload []byte) {
	// Borrow a job struct from the pool (Zero Heap Allocation)
	job := wp.jobPool.Get().(*Job)
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
