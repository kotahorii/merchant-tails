package pool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Poolable interface that objects must implement to be pooled
type Poolable interface {
	Reset()
	IsValid() bool
}

// ObjectPool manages a pool of reusable objects
type ObjectPool[T Poolable] struct {
	pool        chan T
	factory     func() T
	resetFunc   func(T)
	maxSize     int
	created     int32
	inUse       int32
	hits        int64
	misses      int64
	mu          sync.RWMutex
	metrics     *PoolMetrics
	lastCleanup time.Time
}

// PoolMetrics tracks pool performance
type PoolMetrics struct {
	Created     int32
	InUse       int32
	Available   int
	TotalHits   int64
	TotalMisses int64
	HitRate     float64
	LastCleanup time.Time
	MemoryUsage int64
}

// NewObjectPool creates a new object pool
func NewObjectPool[T Poolable](factory func() T, resetFunc func(T), maxSize int) *ObjectPool[T] {
	if maxSize <= 0 {
		maxSize = 100
	}

	pool := &ObjectPool[T]{
		pool:        make(chan T, maxSize),
		factory:     factory,
		resetFunc:   resetFunc,
		maxSize:     maxSize,
		metrics:     &PoolMetrics{},
		lastCleanup: time.Now(),
	}

	// Pre-allocate some objects
	preAllocate := maxSize / 4
	for i := 0; i < preAllocate; i++ {
		obj := factory()
		pool.pool <- obj
		atomic.AddInt32(&pool.created, 1)
	}

	// Start cleanup goroutine
	go pool.cleanupRoutine()

	return pool
}

// Get retrieves an object from the pool
func (p *ObjectPool[T]) Get() T {
	select {
	case obj := <-p.pool:
		// Object retrieved from pool
		atomic.AddInt32(&p.inUse, 1)
		atomic.AddInt64(&p.hits, 1)

		// Validate object
		if obj.IsValid() {
			return obj
		}
		// Invalid object, create new one
		atomic.AddInt32(&p.created, 1)
		return p.factory()

	default:
		// Pool is empty, create new object
		atomic.AddInt32(&p.created, 1)
		atomic.AddInt32(&p.inUse, 1)
		atomic.AddInt64(&p.misses, 1)
		return p.factory()
	}
}

// Put returns an object to the pool
func (p *ObjectPool[T]) Put(obj T) {
	// Check if object is valid
	// For interface types, we need to check if it's valid rather than nil
	if !obj.IsValid() {
		atomic.AddInt32(&p.inUse, -1)
		return
	}

	// Reset the object
	if p.resetFunc != nil {
		p.resetFunc(obj)
	} else {
		obj.Reset()
	}

	select {
	case p.pool <- obj:
		// Object returned to pool
		atomic.AddInt32(&p.inUse, -1)
	default:
		// Pool is full, discard object
		atomic.AddInt32(&p.inUse, -1)
		atomic.AddInt32(&p.created, -1)
	}
}

// GetMetrics returns current pool metrics
func (p *ObjectPool[T]) GetMetrics() PoolMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	created := atomic.LoadInt32(&p.created)
	inUse := atomic.LoadInt32(&p.inUse)
	hits := atomic.LoadInt64(&p.hits)
	misses := atomic.LoadInt64(&p.misses)

	hitRate := 0.0
	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return PoolMetrics{
		Created:     created,
		InUse:       inUse,
		Available:   len(p.pool),
		TotalHits:   hits,
		TotalMisses: misses,
		HitRate:     hitRate,
		LastCleanup: p.lastCleanup,
	}
}

// Clear empties the pool
func (p *ObjectPool[T]) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Drain the pool
	for {
		select {
		case <-p.pool:
			atomic.AddInt32(&p.created, -1)
		default:
			return
		}
	}
}

// cleanupRoutine periodically cleans up the pool
func (p *ObjectPool[T]) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanup()
	}
}

// cleanup removes excess objects from the pool
func (p *ObjectPool[T]) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	available := len(p.pool)
	if available > p.maxSize/2 {
		// Remove half of the excess objects
		toRemove := (available - p.maxSize/2) / 2
		for i := 0; i < toRemove; i++ {
			select {
			case <-p.pool:
				atomic.AddInt32(&p.created, -1)
			default:
				break
			}
		}
	}

	p.lastCleanup = time.Now()
}

// SyncPool wraps sync.Pool for better type safety
type SyncPool[T any] struct {
	pool    *sync.Pool
	factory func() T
	reset   func(*T)
}

// NewSyncPool creates a new sync pool wrapper
func NewSyncPool[T any](factory func() T, reset func(*T)) *SyncPool[T] {
	return &SyncPool[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return factory()
			},
		},
		factory: factory,
		reset:   reset,
	}
}

// Get retrieves an object from the sync pool
func (sp *SyncPool[T]) Get() T {
	return sp.pool.Get().(T)
}

// Put returns an object to the sync pool
func (sp *SyncPool[T]) Put(obj T) {
	if sp.reset != nil {
		sp.reset(&obj)
	}
	sp.pool.Put(obj)
}

// BufferPool manages byte buffer pools
type BufferPool struct {
	small  *sync.Pool
	medium *sync.Pool
	large  *sync.Pool
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		small: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024) // 1KB
			},
		},
		medium: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 8192) // 8KB
			},
		},
		large: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 65536) // 64KB
			},
		},
	}
}

// Get retrieves a buffer of the specified size
func (bp *BufferPool) Get(size int) []byte {
	switch {
	case size <= 1024:
		buf := bp.small.Get().([]byte)
		return buf[:size]
	case size <= 8192:
		buf := bp.medium.Get().([]byte)
		return buf[:size]
	case size <= 65536:
		buf := bp.large.Get().([]byte)
		return buf[:size]
	default:
		return make([]byte, size)
	}
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf []byte) {
	// Clear the buffer
	for i := range buf {
		buf[i] = 0
	}

	switch cap(buf) {
	case 1024:
		bp.small.Put(buf)
	case 8192:
		bp.medium.Put(buf)
	case 65536:
		bp.large.Put(buf)
	default:
		// Don't pool non-standard sizes
	}
}

// ResourcePool manages pooled resources with lifecycle
type ResourcePool[T any] struct {
	available   chan *PooledResource[T]
	factory     func() (T, error)
	destroyer   func(T) error
	validator   func(T) bool
	maxSize     int
	maxAge      time.Duration
	idleTimeout time.Duration
	mu          sync.RWMutex
	closed      bool
	stats       *ResourcePoolStats
}

// PooledResource wraps a pooled resource with metadata
type PooledResource[T any] struct {
	Resource   T
	CreatedAt  time.Time
	LastUsedAt time.Time
	UsageCount int
}

// ResourcePoolStats tracks resource pool statistics
type ResourcePoolStats struct {
	Created   int64
	Destroyed int64
	InUse     int32
	Available int32
	Timeouts  int64
	Errors    int64
}

// NewResourcePool creates a new resource pool
func NewResourcePool[T any](
	factory func() (T, error),
	destroyer func(T) error,
	validator func(T) bool,
	maxSize int,
	maxAge time.Duration,
	idleTimeout time.Duration,
) *ResourcePool[T] {
	pool := &ResourcePool[T]{
		available:   make(chan *PooledResource[T], maxSize),
		factory:     factory,
		destroyer:   destroyer,
		validator:   validator,
		maxSize:     maxSize,
		maxAge:      maxAge,
		idleTimeout: idleTimeout,
		stats:       &ResourcePoolStats{},
	}

	// Start maintenance routine
	go pool.maintain()

	return pool
}

// Acquire gets a resource from the pool
func (rp *ResourcePool[T]) Acquire() (T, error) {
	rp.mu.RLock()
	if rp.closed {
		rp.mu.RUnlock()
		var zero T
		return zero, fmt.Errorf("pool is closed")
	}
	rp.mu.RUnlock()

	select {
	case res := <-rp.available:
		// Validate resource
		if rp.validator != nil && !rp.validator(res.Resource) {
			// Invalid resource, destroy and create new
			if rp.destroyer != nil {
				_ = rp.destroyer(res.Resource)
			}
			atomic.AddInt64(&rp.stats.Destroyed, 1)
			return rp.createNew()
		}

		// Check age
		if time.Since(res.CreatedAt) > rp.maxAge {
			// Resource too old, destroy and create new
			if rp.destroyer != nil {
				_ = rp.destroyer(res.Resource)
			}
			atomic.AddInt64(&rp.stats.Destroyed, 1)
			return rp.createNew()
		}

		// Update usage
		res.LastUsedAt = time.Now()
		res.UsageCount++
		atomic.AddInt32(&rp.stats.InUse, 1)
		atomic.AddInt32(&rp.stats.Available, -1)

		return res.Resource, nil

	default:
		// No available resources, create new
		return rp.createNew()
	}
}

// Release returns a resource to the pool
func (rp *ResourcePool[T]) Release(resource T) {
	rp.mu.RLock()
	if rp.closed {
		rp.mu.RUnlock()
		if rp.destroyer != nil {
			_ = rp.destroyer(resource)
		}
		return
	}
	rp.mu.RUnlock()

	// Validate resource before returning to pool
	if rp.validator != nil && !rp.validator(resource) {
		if rp.destroyer != nil {
			_ = rp.destroyer(resource)
		}
		atomic.AddInt64(&rp.stats.Destroyed, 1)
		atomic.AddInt32(&rp.stats.InUse, -1)
		return
	}

	res := &PooledResource[T]{
		Resource:   resource,
		LastUsedAt: time.Now(),
	}

	select {
	case rp.available <- res:
		atomic.AddInt32(&rp.stats.InUse, -1)
		atomic.AddInt32(&rp.stats.Available, 1)
	default:
		// Pool is full, destroy resource
		if rp.destroyer != nil {
			_ = rp.destroyer(resource)
		}
		atomic.AddInt64(&rp.stats.Destroyed, 1)
		atomic.AddInt32(&rp.stats.InUse, -1)
	}
}

// Close closes the resource pool
func (rp *ResourcePool[T]) Close() error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.closed {
		return nil
	}

	rp.closed = true
	close(rp.available)

	// Destroy all resources
	for res := range rp.available {
		if rp.destroyer != nil {
			_ = rp.destroyer(res.Resource)
		}
		atomic.AddInt64(&rp.stats.Destroyed, 1)
	}

	return nil
}

// createNew creates a new resource
func (rp *ResourcePool[T]) createNew() (T, error) {
	resource, err := rp.factory()
	if err != nil {
		atomic.AddInt64(&rp.stats.Errors, 1)
		return resource, err
	}

	atomic.AddInt64(&rp.stats.Created, 1)
	atomic.AddInt32(&rp.stats.InUse, 1)

	return resource, nil
}

// maintain performs periodic maintenance on the pool
func (rp *ResourcePool[T]) maintain() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rp.mu.RLock()
		if rp.closed {
			rp.mu.RUnlock()
			return
		}
		rp.mu.RUnlock()

		// Remove idle or old resources
		now := time.Now()
		var toRemove []*PooledResource[T]

		// Check all available resources
		n := len(rp.available)
		for i := 0; i < n; i++ {
			select {
			case res := <-rp.available:
				if now.Sub(res.LastUsedAt) > rp.idleTimeout ||
					now.Sub(res.CreatedAt) > rp.maxAge {
					toRemove = append(toRemove, res)
				} else {
					// Put back
					select {
					case rp.available <- res:
					default:
						toRemove = append(toRemove, res)
					}
				}
			default:
				break
			}
		}

		// Destroy removed resources
		for _, res := range toRemove {
			if rp.destroyer != nil {
				_ = rp.destroyer(res.Resource)
			}
			atomic.AddInt64(&rp.stats.Destroyed, 1)
			atomic.AddInt32(&rp.stats.Available, -1)
		}
	}
}

// GetStats returns pool statistics
func (rp *ResourcePool[T]) GetStats() ResourcePoolStats {
	return ResourcePoolStats{
		Created:   atomic.LoadInt64(&rp.stats.Created),
		Destroyed: atomic.LoadInt64(&rp.stats.Destroyed),
		InUse:     atomic.LoadInt32(&rp.stats.InUse),
		Available: atomic.LoadInt32(&rp.stats.Available),
		Timeouts:  atomic.LoadInt64(&rp.stats.Timeouts),
		Errors:    atomic.LoadInt64(&rp.stats.Errors),
	}
}
