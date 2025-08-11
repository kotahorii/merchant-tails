package concurrent

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Job represents a unit of work to be executed
type Job interface {
	Execute() error
	GetID() string
	GetPriority() int
}

// Result represents the result of a job execution
type Result struct {
	JobID    string
	Error    error
	Duration time.Duration
	Data     interface{}
}

// WorkerPool manages a pool of workers for concurrent job execution
type WorkerPool struct {
	workers       int
	jobQueue      chan Job
	resultQueue   chan Result
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	activeJobs    int32
	completedJobs int64
	failedJobs    int64
	metrics       *PoolMetrics
	mu            sync.RWMutex
}

// PoolMetrics tracks performance metrics
type PoolMetrics struct {
	TotalJobs       int64
	CompletedJobs   int64
	FailedJobs      int64
	AverageExecTime time.Duration
	MaxExecTime     time.Duration
	MinExecTime     time.Duration
	QueueSize       int
	ActiveWorkers   int32
	TotalExecTime   time.Duration
	LastUpdateTime  time.Time
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:     workers,
		jobQueue:    make(chan Job, queueSize),
		resultQueue: make(chan Result, queueSize),
		ctx:         ctx,
		cancel:      cancel,
		metrics:     &PoolMetrics{LastUpdateTime: time.Now()},
	}

	// Start workers
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	// Start metrics collector
	go pool.metricsCollector()

	return pool
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}

			atomic.AddInt32(&wp.activeJobs, 1)
			start := time.Now()

			// Execute the job
			err := job.Execute()
			duration := time.Since(start)

			// Send result
			result := Result{
				JobID:    job.GetID(),
				Error:    err,
				Duration: duration,
			}

			select {
			case wp.resultQueue <- result:
			case <-wp.ctx.Done():
				atomic.AddInt32(&wp.activeJobs, -1)
				return
			}

			// Update metrics
			if err != nil {
				atomic.AddInt64(&wp.failedJobs, 1)
			} else {
				atomic.AddInt64(&wp.completedJobs, 1)
			}

			atomic.AddInt32(&wp.activeJobs, -1)
		}
	}
}

// Submit adds a job to the queue
func (wp *WorkerPool) Submit(job Job) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// SubmitWithTimeout submits a job with a timeout
func (wp *WorkerPool) SubmitWithTimeout(job Job, timeout time.Duration) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout submitting job")
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	}
}

// GetResult retrieves a result from the result queue
func (wp *WorkerPool) GetResult() (Result, bool) {
	select {
	case result, ok := <-wp.resultQueue:
		return result, ok
	default:
		return Result{}, false
	}
}

// GetResultWithTimeout retrieves a result with a timeout
func (wp *WorkerPool) GetResultWithTimeout(timeout time.Duration) (Result, error) {
	select {
	case result := <-wp.resultQueue:
		return result, nil
	case <-time.After(timeout):
		return Result{}, fmt.Errorf("timeout waiting for result")
	case <-wp.ctx.Done():
		return Result{}, fmt.Errorf("worker pool is shutting down")
	}
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown() {
	wp.cancel()
	close(wp.jobQueue)
	wp.wg.Wait()
	close(wp.resultQueue)
}

// GetMetrics returns current pool metrics
func (wp *WorkerPool) GetMetrics() PoolMetrics {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	metrics := *wp.metrics
	metrics.ActiveWorkers = atomic.LoadInt32(&wp.activeJobs)
	metrics.CompletedJobs = atomic.LoadInt64(&wp.completedJobs)
	metrics.FailedJobs = atomic.LoadInt64(&wp.failedJobs)
	metrics.QueueSize = len(wp.jobQueue)

	return metrics
}

// metricsCollector periodically updates metrics
func (wp *WorkerPool) metricsCollector() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wp.updateMetrics()
		case <-wp.ctx.Done():
			return
		}
	}
}

// updateMetrics updates the pool metrics
func (wp *WorkerPool) updateMetrics() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	completed := atomic.LoadInt64(&wp.completedJobs)
	failed := atomic.LoadInt64(&wp.failedJobs)
	total := completed + failed

	wp.metrics.TotalJobs = total
	wp.metrics.CompletedJobs = completed
	wp.metrics.FailedJobs = failed
	wp.metrics.LastUpdateTime = time.Now()

	if total > 0 && wp.metrics.TotalExecTime > 0 {
		wp.metrics.AverageExecTime = time.Duration(int64(wp.metrics.TotalExecTime) / total)
	}
}

// PriorityQueue implements a priority queue for jobs
type PriorityQueue struct {
	items []*PriorityJob
	mu    sync.RWMutex
}

// PriorityJob wraps a job with priority
type PriorityJob struct {
	Job      Job
	Priority int
	AddedAt  time.Time
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*PriorityJob, 0),
	}
}

// Push adds a job to the priority queue
func (pq *PriorityQueue) Push(job Job) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pJob := &PriorityJob{
		Job:      job,
		Priority: job.GetPriority(),
		AddedAt:  time.Now(),
	}

	// Insert in priority order
	inserted := false
	for i, item := range pq.items {
		if pJob.Priority > item.Priority {
			pq.items = append(pq.items[:i], append([]*PriorityJob{pJob}, pq.items[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		pq.items = append(pq.items, pJob)
	}
}

// Pop removes and returns the highest priority job
func (pq *PriorityQueue) Pop() (Job, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return nil, false
	}

	job := pq.items[0].Job
	pq.items = pq.items[1:]

	return job, true
}

// Size returns the number of jobs in the queue
func (pq *PriorityQueue) Size() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}

// BatchProcessor processes jobs in batches
type BatchProcessor struct {
	pool      *WorkerPool
	batchSize int
	timeout   time.Duration
	mu        sync.Mutex
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(pool *WorkerPool, batchSize int, timeout time.Duration) *BatchProcessor {
	return &BatchProcessor{
		pool:      pool,
		batchSize: batchSize,
		timeout:   timeout,
	}
}

// ProcessBatch processes a batch of jobs
func (bp *BatchProcessor) ProcessBatch(jobs []Job) ([]Result, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	results := make([]Result, 0, len(jobs))
	resultMap := make(map[string]Result)

	// Submit all jobs
	for _, job := range jobs {
		if err := bp.pool.Submit(job); err != nil {
			return nil, fmt.Errorf("failed to submit job %s: %w", job.GetID(), err)
		}
	}

	// Collect results
	deadline := time.Now().Add(bp.timeout)
	for len(resultMap) < len(jobs) {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("batch processing timeout")
		}

		result, err := bp.pool.GetResultWithTimeout(100 * time.Millisecond)
		if err == nil {
			resultMap[result.JobID] = result
		}
	}

	// Order results according to input jobs
	for _, job := range jobs {
		if result, ok := resultMap[job.GetID()]; ok {
			results = append(results, result)
		}
	}

	return results, nil
}

// TaskScheduler schedules tasks for execution
type TaskScheduler struct {
	pool          *WorkerPool
	priorityQueue *PriorityQueue
	ticker        *time.Ticker
	stopCh        chan struct{}
	mu            sync.Mutex
}

// NewTaskScheduler creates a new task scheduler
func NewTaskScheduler(pool *WorkerPool, interval time.Duration) *TaskScheduler {
	scheduler := &TaskScheduler{
		pool:          pool,
		priorityQueue: NewPriorityQueue(),
		ticker:        time.NewTicker(interval),
		stopCh:        make(chan struct{}),
	}

	go scheduler.run()
	return scheduler
}

// Schedule adds a job to be scheduled
func (ts *TaskScheduler) Schedule(job Job) {
	ts.priorityQueue.Push(job)
}

// run processes scheduled jobs
func (ts *TaskScheduler) run() {
	for {
		select {
		case <-ts.ticker.C:
			ts.processJobs()
		case <-ts.stopCh:
			return
		}
	}
}

// processJobs processes pending jobs
func (ts *TaskScheduler) processJobs() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Process up to batch size
	batchSize := 10
	for i := 0; i < batchSize; i++ {
		job, ok := ts.priorityQueue.Pop()
		if !ok {
			break
		}

		_ = ts.pool.Submit(job)
	}
}

// Stop stops the scheduler
func (ts *TaskScheduler) Stop() {
	ts.ticker.Stop()
	close(ts.stopCh)
}

// Throttler limits the rate of job execution
type Throttler struct {
	rate     int
	interval time.Duration
	tokens   chan struct{}
	stopCh   chan struct{}
}

// NewThrottler creates a new throttler
func NewThrottler(rate int, interval time.Duration) *Throttler {
	throttler := &Throttler{
		rate:     rate,
		interval: interval,
		tokens:   make(chan struct{}, rate),
		stopCh:   make(chan struct{}),
	}

	// Fill initial tokens
	for i := 0; i < rate; i++ {
		throttler.tokens <- struct{}{}
	}

	// Start token refill
	go throttler.refill()

	return throttler
}

// refill periodically refills tokens
func (t *Throttler) refill() {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Try to add tokens up to rate
			for i := 0; i < t.rate; i++ {
				select {
				case t.tokens <- struct{}{}:
				default:
					// Channel is full
				}
			}
		case <-t.stopCh:
			return
		}
	}
}

// Allow checks if an operation is allowed
func (t *Throttler) Allow() bool {
	select {
	case <-t.tokens:
		return true
	default:
		return false
	}
}

// Wait waits for permission to proceed
func (t *Throttler) Wait() {
	<-t.tokens
}

// Stop stops the throttler
func (t *Throttler) Stop() {
	close(t.stopCh)
}
