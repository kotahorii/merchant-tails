package concurrent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConcurrentManager manages all concurrent operations in the game
type ConcurrentManager struct {
	mainPool       *WorkerPool
	backgroundPool *WorkerPool
	scheduler      *TaskScheduler
	batchProcessor *BatchProcessor
	throttler      *Throttler
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex

	// Performance metrics
	metrics          *ManagerMetrics
	metricsTimer     *time.Timer
	metricsCollector interface{} // Will hold *monitoring.MetricsCollector if available
}

// ManagerMetrics tracks overall concurrent system metrics
type ManagerMetrics struct {
	TotalJobsSubmitted  int64
	TotalJobsCompleted  int64
	TotalJobsFailed     int64
	AverageResponseTime time.Duration
	CurrentQueueDepth   int
	WorkerUtilization   float64
	LastMetricsUpdate   time.Time
}

// ConcurrentConfig configures the concurrent manager
type ConcurrentConfig struct {
	MainWorkers       int
	BackgroundWorkers int
	QueueSize         int
	BatchSize         int
	ThrottleRate      int
	ThrottleInterval  time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() *ConcurrentConfig {
	return &ConcurrentConfig{
		MainWorkers:       4,
		BackgroundWorkers: 2,
		QueueSize:         100,
		BatchSize:         10,
		ThrottleRate:      100,
		ThrottleInterval:  time.Second,
	}
}

// NewConcurrentManager creates a new concurrent manager
func NewConcurrentManager(config *ConcurrentConfig) *ConcurrentManager {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create worker pools
	mainPool := NewWorkerPool(config.MainWorkers, config.QueueSize)
	backgroundPool := NewWorkerPool(config.BackgroundWorkers, config.QueueSize/2)

	// Create scheduler for background pool
	scheduler := NewTaskScheduler(backgroundPool, 100*time.Millisecond)

	// Create batch processor
	batchProcessor := NewBatchProcessor(mainPool, config.BatchSize, 5*time.Second)

	// Create throttler
	throttler := NewThrottler(config.ThrottleRate, config.ThrottleInterval)

	manager := &ConcurrentManager{
		mainPool:       mainPool,
		backgroundPool: backgroundPool,
		scheduler:      scheduler,
		batchProcessor: batchProcessor,
		throttler:      throttler,
		ctx:            ctx,
		cancel:         cancel,
		metrics:        &ManagerMetrics{LastMetricsUpdate: time.Now()},
	}

	// Start metrics collection
	manager.startMetricsCollection()

	return manager
}

// SubmitJob submits a job to the main pool
func (cm *ConcurrentManager) SubmitJob(job Job) error {
	// Apply throttling
	if !cm.throttler.Allow() {
		return fmt.Errorf("rate limit exceeded")
	}

	jobStart := time.Now()
	err := cm.mainPool.Submit(job)
	if err == nil {
		cm.updateSubmittedCount()
		// Record job completion metrics if monitoring is available
		if metricsCollector := cm.getMetricsCollector(); metricsCollector != nil {
			// Type assert to check if it has RecordJobCompletion method
			type jobRecorder interface {
				RecordJobCompletion(success bool, duration time.Duration)
			}
			if metrics, ok := metricsCollector.(jobRecorder); ok {
				metrics.RecordJobCompletion(err == nil, time.Since(jobStart))
			}
		}
	}
	return err
}

// SubmitBackgroundJob submits a low-priority job to background pool
func (cm *ConcurrentManager) SubmitBackgroundJob(job Job) error {
	err := cm.backgroundPool.Submit(job)
	if err == nil {
		cm.updateSubmittedCount()
	}
	return err
}

// ScheduleJob schedules a job for later execution
func (cm *ConcurrentManager) ScheduleJob(job Job) {
	cm.scheduler.Schedule(job)
	cm.updateSubmittedCount()
}

// ProcessBatch processes a batch of jobs
func (cm *ConcurrentManager) ProcessBatch(jobs []Job) ([]Result, error) {
	// Update metrics
	for range jobs {
		cm.updateSubmittedCount()
	}

	return cm.batchProcessor.ProcessBatch(jobs)
}

// ProcessMarketUpdate processes market updates concurrently
func (cm *ConcurrentManager) ProcessMarketUpdate(itemGroups [][]string) error {
	jobs := make([]Job, 0, len(itemGroups))

	for i, group := range itemGroups {
		job := NewMarketUpdateJob(
			fmt.Sprintf("market-update-%d", i),
			nil, // Market would be passed here
			group,
		)
		jobs = append(jobs, job)
	}

	results, err := cm.ProcessBatch(jobs)
	if err != nil {
		return err
	}

	// Process results
	for _, result := range results {
		if result.Error != nil {
			cm.updateFailedCount()
		} else {
			cm.updateCompletedCount()
		}
	}

	return nil
}

// ProcessPriceAnalysis performs concurrent price analysis
func (cm *ConcurrentManager) ProcessPriceAnalysis(items map[string][]float64) (map[string]PriceAnalysisResult, error) {
	results := make(map[string]PriceAnalysisResult)
	resultChannels := make(map[string]chan PriceAnalysisResult)

	// Submit analysis jobs
	for itemID, history := range items {
		job := NewPriceAnalysisJob(
			fmt.Sprintf("price-analysis-%s", itemID),
			itemID,
			history,
		)

		resultChannels[itemID] = job.Result

		if err := cm.SubmitJob(job); err != nil {
			return nil, err
		}
	}

	// Collect results
	timeout := time.After(5 * time.Second)
	for itemID, ch := range resultChannels {
		select {
		case result := <-ch:
			results[itemID] = result
			cm.updateCompletedCount()
		case <-timeout:
			return results, fmt.Errorf("timeout waiting for price analysis")
		}
	}

	return results, nil
}

// ProcessEvents processes game events concurrently
func (cm *ConcurrentManager) ProcessEvents(events []interface{}, handler func(interface{}) error) error {
	var wg sync.WaitGroup
	errors := make(chan error, len(events))

	for i, event := range events {
		wg.Add(1)

		job := NewEventProcessingJob(
			fmt.Sprintf("event-%d", i),
			fmt.Sprintf("event-%d", i),
			"game_event",
			event,
			handler,
		)

		go func() {
			defer wg.Done()

			if err := cm.SubmitJob(job); err != nil {
				errors <- err
				return
			}

			// Wait for result
			result, err := cm.mainPool.GetResultWithTimeout(2 * time.Second)
			if err != nil {
				errors <- err
				cm.updateFailedCount()
			} else if result.Error != nil {
				errors <- result.Error
				cm.updateFailedCount()
			} else {
				cm.updateCompletedCount()
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

// ProcessDataInParallel processes data chunks in parallel
func (cm *ConcurrentManager) ProcessDataInParallel(data []interface{}, chunkSize int, processFn func(interface{}) error) error {
	// Split data into chunks
	chunks := make([][]interface{}, 0)
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}

	// Create jobs for each chunk
	jobs := make([]Job, 0, len(chunks))
	for i, chunk := range chunks {
		job := NewBulkDataProcessingJob(
			fmt.Sprintf("data-chunk-%d", i),
			chunk,
			processFn,
		)
		jobs = append(jobs, job)
	}

	// Process batch
	results, err := cm.ProcessBatch(jobs)
	if err != nil {
		return err
	}

	// Check results
	for _, result := range results {
		if result.Error != nil {
			cm.updateFailedCount()
			return result.Error
		}
		cm.updateCompletedCount()
	}

	return nil
}

// GetMetrics returns current performance metrics
func (cm *ConcurrentManager) GetMetrics() ManagerMetrics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	metrics := *cm.metrics

	// Get pool metrics
	mainMetrics := cm.mainPool.GetMetrics()
	bgMetrics := cm.backgroundPool.GetMetrics()

	metrics.CurrentQueueDepth = mainMetrics.QueueSize + bgMetrics.QueueSize

	totalWorkers := float64(cm.mainPool.workers + cm.backgroundPool.workers)
	activeWorkers := float64(mainMetrics.ActiveWorkers + bgMetrics.ActiveWorkers)
	if totalWorkers > 0 {
		metrics.WorkerUtilization = activeWorkers / totalWorkers
	}

	return metrics
}

// Shutdown gracefully shuts down the concurrent manager
func (cm *ConcurrentManager) Shutdown() {
	cm.cancel()

	// Stop scheduler
	cm.scheduler.Stop()

	// Stop throttler
	cm.throttler.Stop()

	// Shutdown pools
	cm.mainPool.Shutdown()
	cm.backgroundPool.Shutdown()
}

// startMetricsCollection starts periodic metrics collection
func (cm *ConcurrentManager) startMetricsCollection() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cm.collectMetrics()
			case <-cm.ctx.Done():
				return
			}
		}
	}()
}

// collectMetrics collects and updates metrics
func (cm *ConcurrentManager) collectMetrics() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	mainMetrics := cm.mainPool.GetMetrics()
	bgMetrics := cm.backgroundPool.GetMetrics()

	cm.metrics.TotalJobsCompleted = mainMetrics.CompletedJobs + bgMetrics.CompletedJobs
	cm.metrics.TotalJobsFailed = mainMetrics.FailedJobs + bgMetrics.FailedJobs

	// Calculate average response time
	totalTime := mainMetrics.AverageExecTime + bgMetrics.AverageExecTime
	if totalTime > 0 {
		cm.metrics.AverageResponseTime = totalTime / 2
	}

	cm.metrics.LastMetricsUpdate = time.Now()
}

// updateSubmittedCount increments submitted job count
func (cm *ConcurrentManager) updateSubmittedCount() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.metrics.TotalJobsSubmitted++
}

// updateCompletedCount increments completed job count
func (cm *ConcurrentManager) updateCompletedCount() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.metrics.TotalJobsCompleted++
}

// updateFailedCount increments failed job count
func (cm *ConcurrentManager) updateFailedCount() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.metrics.TotalJobsFailed++
}

// RunWithTimeout runs a function with timeout
func RunWithTimeout(fn func() error, timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// ParallelMap applies a function to each element in parallel
func ParallelMap[T any, R any](items []T, fn func(T) (R, error), workers int) ([]R, error) {
	if workers <= 0 {
		workers = 4
	}

	type result struct {
		index int
		value R
		err   error
	}

	jobs := make(chan int, len(items))
	results := make(chan result, len(items))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				val, err := fn(items[i])
				results <- result{index: i, value: val, err: err}
			}
		}()
	}

	// Send jobs
	for i := range items {
		jobs <- i
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	close(results)

	// Collect results
	output := make([]R, len(items))
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		output[res.index] = res.value
	}

	return output, nil
}

// SetMetricsCollector sets the metrics collector
func (cm *ConcurrentManager) SetMetricsCollector(collector interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.metricsCollector = collector
}

// getMetricsCollector returns the metrics collector (internal helper)
func (cm *ConcurrentManager) getMetricsCollector() interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.metricsCollector
}
