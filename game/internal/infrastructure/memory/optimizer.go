package memory

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// Optimizer manages memory optimization strategies
type Optimizer struct {
	mu           sync.RWMutex
	gcPercent    int
	memLimit     int64
	autoTune     bool
	lastGC       time.Time
	gcInterval   time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	trimMemory   bool
	aggressiveGC bool
}

// NewOptimizer creates a new memory optimizer
func NewOptimizer() *Optimizer {
	return &Optimizer{
		gcPercent:    100, // Default Go GC percentage
		memLimit:     0,   // No limit by default
		gcInterval:   30 * time.Second,
		stopChan:     make(chan struct{}),
		trimMemory:   true,
		aggressiveGC: false,
	}
}

// Start begins memory optimization
func (o *Optimizer) Start() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.autoTune {
		o.wg.Add(1)
		go o.optimizationLoop()
	}
}

// Stop stops memory optimization
func (o *Optimizer) Stop() {
	close(o.stopChan)
	o.wg.Wait()
}

// optimizationLoop runs periodic optimization
func (o *Optimizer) optimizationLoop() {
	defer o.wg.Done()
	ticker := time.NewTicker(o.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.optimize()
		case <-o.stopChan:
			return
		}
	}
}

// optimize performs memory optimization
func (o *Optimizer) optimize() {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Run garbage collection if needed
	if time.Since(o.lastGC) > o.gcInterval {
		runtime.GC()
		o.lastGC = time.Now()
	}

	// Trim memory if enabled
	if o.trimMemory {
		o.trimMemoryUsage()
	}

	// Apply aggressive GC if memory usage is high
	if o.aggressiveGC {
		o.applyAggressiveGC()
	}
}

// trimMemoryUsage returns memory to OS
func (o *Optimizer) trimMemoryUsage() {
	debug.FreeOSMemory()
}

// applyAggressiveGC applies aggressive garbage collection
func (o *Optimizer) applyAggressiveGC() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// If heap is more than 80% of system memory, force GC
	if m.HeapAlloc > uint64(float64(m.HeapSys)*0.8) {
		runtime.GC()
		debug.FreeOSMemory()
	}
}

// SetGCPercent sets the garbage collection percentage
func (o *Optimizer) SetGCPercent(percent int) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.gcPercent = percent
	debug.SetGCPercent(percent)
}

// SetMemoryLimit sets memory limit
func (o *Optimizer) SetMemoryLimit(bytes int64) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.memLimit = bytes
	debug.SetMemoryLimit(bytes)
}

// EnableAutoTune enables automatic tuning
func (o *Optimizer) EnableAutoTune(enable bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.autoTune = enable
}

// SetAggressiveGC enables aggressive garbage collection
func (o *Optimizer) SetAggressiveGC(enable bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.aggressiveGC = enable
}

// GetMemoryStats returns current memory statistics
func (o *Optimizer) GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Allocated:    m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		System:       m.Sys,
		NumGC:        m.NumGC,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
		NextGC:       m.NextGC,
		LastGC:       time.Unix(0, int64(m.LastGC)), //nolint:gosec // LastGC is always valid nanoseconds
		PauseTotal:   m.PauseTotalNs,
		NumGoroutine: runtime.NumGoroutine(),
	}
}

// MemoryStats contains memory statistics
type MemoryStats struct {
	Allocated    uint64
	TotalAlloc   uint64
	System       uint64
	NumGC        uint32
	HeapAlloc    uint64
	HeapSys      uint64
	HeapInuse    uint64
	HeapReleased uint64
	StackInuse   uint64
	StackSys     uint64
	NextGC       uint64
	LastGC       time.Time
	PauseTotal   uint64
	NumGoroutine int
}

// OptimizationStrategy represents a memory optimization strategy
type OptimizationStrategy interface {
	Apply(*Optimizer)
	Name() string
}

// LowMemoryStrategy for low memory environments
type LowMemoryStrategy struct{}

func (s LowMemoryStrategy) Apply(o *Optimizer) {
	o.SetGCPercent(50)              // More frequent GC
	o.SetAggressiveGC(true)         // Enable aggressive GC
	o.trimMemory = true             // Always trim memory
	o.gcInterval = 10 * time.Second // More frequent optimization
}

func (s LowMemoryStrategy) Name() string {
	return "LowMemory"
}

// HighPerformanceStrategy for performance-critical scenarios
type HighPerformanceStrategy struct{}

func (s HighPerformanceStrategy) Apply(o *Optimizer) {
	o.SetGCPercent(200)             // Less frequent GC
	o.SetAggressiveGC(false)        // Disable aggressive GC
	o.trimMemory = false            // Don't trim memory
	o.gcInterval = 60 * time.Second // Less frequent optimization
}

func (s HighPerformanceStrategy) Name() string {
	return "HighPerformance"
}

// BalancedStrategy for balanced memory/performance
type BalancedStrategy struct{}

func (s BalancedStrategy) Apply(o *Optimizer) {
	o.SetGCPercent(100)             // Default GC
	o.SetAggressiveGC(false)        // Disable aggressive GC
	o.trimMemory = true             // Trim memory periodically
	o.gcInterval = 30 * time.Second // Default interval
}

func (s BalancedStrategy) Name() string {
	return "Balanced"
}

// ApplyStrategy applies an optimization strategy
func (o *Optimizer) ApplyStrategy(strategy OptimizationStrategy) {
	strategy.Apply(o)
}
