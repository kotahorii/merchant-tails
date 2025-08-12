package profiling

import (
	"fmt"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// MemoryProfiler provides memory profiling capabilities
type MemoryProfiler struct {
	mu           sync.RWMutex
	enabled      bool
	interval     time.Duration
	snapshots    []*MemorySnapshot
	maxSnapshots int
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// MemorySnapshot represents a point-in-time memory state
type MemorySnapshot struct {
	Timestamp    time.Time
	Allocated    uint64 // bytes allocated and still in use
	TotalAlloc   uint64 // bytes allocated (even if freed)
	Sys          uint64 // bytes obtained from system
	NumGC        uint32 // number of completed GC cycles
	NumGoroutine int    // number of goroutines
	HeapAlloc    uint64 // heap bytes allocated
	HeapSys      uint64 // heap bytes obtained from system
	HeapInuse    uint64 // heap bytes in use
	HeapReleased uint64 // heap bytes released to OS
	StackInuse   uint64 // stack bytes in use
	StackSys     uint64 // stack bytes obtained from system
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{
		interval:     30 * time.Second,
		maxSnapshots: 100,
		snapshots:    make([]*MemorySnapshot, 0),
		stopChan:     make(chan struct{}),
	}
}

// Start begins memory profiling
func (mp *MemoryProfiler) Start() {
	mp.mu.Lock()
	if mp.enabled {
		mp.mu.Unlock()
		return
	}
	mp.enabled = true
	mp.mu.Unlock()

	mp.wg.Add(1)
	go mp.profileLoop()
}

// Stop stops memory profiling
func (mp *MemoryProfiler) Stop() {
	mp.mu.Lock()
	if !mp.enabled {
		mp.mu.Unlock()
		return
	}
	mp.enabled = false
	mp.mu.Unlock()

	close(mp.stopChan)
	mp.wg.Wait()
}

// profileLoop continuously profiles memory
func (mp *MemoryProfiler) profileLoop() {
	defer mp.wg.Done()
	ticker := time.NewTicker(mp.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mp.takeSnapshot()
		case <-mp.stopChan:
			return
		}
	}
}

// takeSnapshot captures current memory state
func (mp *MemoryProfiler) takeSnapshot() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := &MemorySnapshot{
		Timestamp:    time.Now(),
		Allocated:    m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
	}

	mp.mu.Lock()
	mp.snapshots = append(mp.snapshots, snapshot)
	if len(mp.snapshots) > mp.maxSnapshots {
		mp.snapshots = mp.snapshots[1:]
	}
	mp.mu.Unlock()
}

// GetCurrentMemory returns current memory usage
func (mp *MemoryProfiler) GetCurrentMemory() *MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemorySnapshot{
		Timestamp:    time.Now(),
		Allocated:    m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
	}
}

// GetSnapshots returns recent memory snapshots
func (mp *MemoryProfiler) GetSnapshots() []*MemorySnapshot {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	result := make([]*MemorySnapshot, len(mp.snapshots))
	copy(result, mp.snapshots)
	return result
}

// AnalyzeMemoryGrowth analyzes memory growth over time
func (mp *MemoryProfiler) AnalyzeMemoryGrowth() *MemoryAnalysis {
	mp.mu.RLock()
	snapshots := mp.snapshots
	mp.mu.RUnlock()

	if len(snapshots) < 2 {
		return nil
	}

	first := snapshots[0]
	last := snapshots[len(snapshots)-1]
	duration := last.Timestamp.Sub(first.Timestamp)

	// Calculate growth rates
	allocGrowth := float64(last.Allocated-first.Allocated) / float64(first.Allocated) * 100
	sysGrowth := float64(last.Sys-first.Sys) / float64(first.Sys) * 100

	// Find peak memory usage
	var peakAlloc uint64
	var peakTimestamp time.Time
	for _, s := range snapshots {
		if s.Allocated > peakAlloc {
			peakAlloc = s.Allocated
			peakTimestamp = s.Timestamp
		}
	}

	// Calculate average allocation rate
	totalAllocDiff := last.TotalAlloc - first.TotalAlloc
	allocRate := float64(totalAllocDiff) / duration.Seconds()

	return &MemoryAnalysis{
		Duration:        duration,
		AllocGrowth:     allocGrowth,
		SysGrowth:       sysGrowth,
		PeakAlloc:       peakAlloc,
		PeakTimestamp:   peakTimestamp,
		AllocRate:       allocRate,
		NumSnapshots:    len(snapshots),
		GCCount:         last.NumGC - first.NumGC,
		GoroutineGrowth: last.NumGoroutine - first.NumGoroutine,
	}
}

// MemoryAnalysis contains memory usage analysis
type MemoryAnalysis struct {
	Duration        time.Duration
	AllocGrowth     float64   // percentage growth in allocated memory
	SysGrowth       float64   // percentage growth in system memory
	PeakAlloc       uint64    // peak allocated memory
	PeakTimestamp   time.Time // when peak occurred
	AllocRate       float64   // bytes allocated per second
	NumSnapshots    int       // number of snapshots analyzed
	GCCount         uint32    // number of GC cycles
	GoroutineGrowth int       // change in goroutine count
}

// String returns a formatted analysis report
func (ma *MemoryAnalysis) String() string {
	return fmt.Sprintf(
		"Memory Analysis Report:\n"+
			"Duration: %v\n"+
			"Allocated Memory Growth: %.2f%%\n"+
			"System Memory Growth: %.2f%%\n"+
			"Peak Allocation: %s at %s\n"+
			"Allocation Rate: %s/sec\n"+
			"GC Cycles: %d\n"+
			"Goroutine Change: %+d\n",
		ma.Duration,
		ma.AllocGrowth,
		ma.SysGrowth,
		formatBytes(ma.PeakAlloc),
		ma.PeakTimestamp.Format("15:04:05"),
		formatBytes(uint64(ma.AllocRate)),
		ma.GCCount,
		ma.GoroutineGrowth,
	)
}

// DetectMemoryLeaks attempts to identify potential memory leaks
func (mp *MemoryProfiler) DetectMemoryLeaks() []MemoryLeak {
	mp.mu.RLock()
	snapshots := mp.snapshots
	mp.mu.RUnlock()

	if len(snapshots) < 10 {
		return nil
	}

	var leaks []MemoryLeak

	// Check for consistent memory growth
	const checkWindow = 5
	for i := checkWindow; i < len(snapshots); i++ {
		window := snapshots[i-checkWindow : i]
		if isConsistentGrowth(window) {
			leaks = append(leaks, MemoryLeak{
				Type:        "ConsistentGrowth",
				Description: "Memory shows consistent growth pattern",
				StartTime:   window[0].Timestamp,
				EndTime:     window[len(window)-1].Timestamp,
				Growth:      window[len(window)-1].Allocated - window[0].Allocated,
			})
		}
	}

	// Check for goroutine leaks
	if len(snapshots) > 1 {
		first := snapshots[0]
		last := snapshots[len(snapshots)-1]
		goroutineGrowth := last.NumGoroutine - first.NumGoroutine

		if goroutineGrowth > 100 {
			leaks = append(leaks, MemoryLeak{
				Type:        "GoroutineLeak",
				Description: fmt.Sprintf("Goroutine count increased by %d", goroutineGrowth),
				StartTime:   first.Timestamp,
				EndTime:     last.Timestamp,
				Growth:      uint64(goroutineGrowth),
			})
		}
	}

	return leaks
}

// MemoryLeak represents a potential memory leak
type MemoryLeak struct {
	Type        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Growth      uint64
}

// isConsistentGrowth checks if memory shows consistent growth
func isConsistentGrowth(snapshots []*MemorySnapshot) bool {
	if len(snapshots) < 2 {
		return false
	}

	growthCount := 0
	for i := 1; i < len(snapshots); i++ {
		if snapshots[i].Allocated > snapshots[i-1].Allocated {
			growthCount++
		}
	}

	// If memory grew in 80% or more of the intervals
	return float64(growthCount)/float64(len(snapshots)-1) >= 0.8
}

// formatBytes formats bytes in human-readable format
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// WriteHeapProfile writes heap profile to writer
func (mp *MemoryProfiler) WriteHeapProfile(profile *pprof.Profile) error {
	return pprof.WriteHeapProfile(profile)
}

// ForceGC forces garbage collection
func (mp *MemoryProfiler) ForceGC() {
	runtime.GC()
	runtime.Gosched()
}

// SetInterval sets the profiling interval
func (mp *MemoryProfiler) SetInterval(interval time.Duration) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.interval = interval
}
