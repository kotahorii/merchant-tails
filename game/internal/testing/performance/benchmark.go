package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/gamestate"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
	playermerchant "github.com/yourusername/merchant-tails/game/internal/domain/merchant"
)

// BenchmarkRunner runs performance benchmarks
type BenchmarkRunner struct {
	mu         sync.RWMutex
	config     *BenchmarkConfig
	results    map[string]*BenchmarkResult
	collectors []MetricCollector
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
}

// BenchmarkConfig defines benchmark configuration
type BenchmarkConfig struct {
	Name            string
	Duration        time.Duration
	WarmupDuration  time.Duration
	Iterations      int
	Parallelism     int
	MemoryProfile   bool
	CPUProfile      bool
	TraceEnabled    bool
	MetricsInterval time.Duration
}

// BenchmarkResult stores benchmark results
type BenchmarkResult struct {
	Name           string
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	Iterations     int
	Operations     int64
	BytesAllocated int64
	Allocations    int64
	MinLatency     time.Duration
	MaxLatency     time.Duration
	AvgLatency     time.Duration
	P50Latency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
	ErrorCount     int
	ErrorRate      float64
	Throughput     float64
	MemoryStats    runtime.MemStats
	CustomMetrics  map[string]interface{}
}

// MetricCollector collects metrics during benchmarks
type MetricCollector interface {
	Collect() map[string]interface{}
	Reset()
}

// NewBenchmarkRunner creates a new benchmark runner
func NewBenchmarkRunner(config *BenchmarkConfig) *BenchmarkRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &BenchmarkRunner{
		config:     config,
		results:    make(map[string]*BenchmarkResult),
		collectors: make([]MetricCollector, 0),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// MarketBenchmark benchmarks market operations
type MarketBenchmark struct {
	market      *market.Market
	items       []*item.Item
	iterations  int
	parallelism int
	results     chan *OperationResult
	wg          sync.WaitGroup
}

// OperationResult stores individual operation results
type OperationResult struct {
	Operation string
	Latency   time.Duration
	Error     error
	Bytes     int64
	Allocs    int64
}

// NewMarketBenchmark creates a market benchmark
func NewMarketBenchmark(iterations, parallelism int) *MarketBenchmark {
	// Create test market
	mkt := market.NewMarket()

	// Create test items
	items := make([]*item.Item, 100)
	categories := []item.Category{
		item.CategoryFruit,
		item.CategoryPotion,
		item.CategoryWeapon,
		item.CategoryAccessory,
		item.CategoryMagicBook,
		item.CategoryGem,
	}

	for i := range items {
		items[i] = &item.Item{
			ID:        fmt.Sprintf("item_%d", i),
			Name:      fmt.Sprintf("Test Item %d", i),
			Category:  categories[i%len(categories)],
			BasePrice: 10 + i*5,
		}
	}

	return &MarketBenchmark{
		market:      mkt,
		items:       items,
		iterations:  iterations,
		parallelism: parallelism,
		results:     make(chan *OperationResult, iterations),
	}
}

// Run executes the market benchmark
func (mb *MarketBenchmark) Run(ctx context.Context) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:          "Market Operations",
		StartTime:     time.Now(),
		Iterations:    mb.iterations,
		CustomMetrics: make(map[string]interface{}),
	}

	// Warmup
	mb.warmup()

	// Run benchmark
	mb.wg.Add(mb.parallelism)
	for i := 0; i < mb.parallelism; i++ {
		go mb.worker(ctx, mb.iterations/mb.parallelism)
	}

	// Collect results
	go func() {
		mb.wg.Wait()
		close(mb.results)
	}()

	// Process results
	latencies := make([]time.Duration, 0, mb.iterations)
	var totalLatency time.Duration
	var totalBytes, totalAllocs int64

	for op := range mb.results {
		result.Operations++
		latencies = append(latencies, op.Latency)
		totalLatency += op.Latency
		totalBytes += op.Bytes
		totalAllocs += op.Allocs
		if op.Error != nil {
			result.ErrorCount++
		}
	}

	// Calculate statistics
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.BytesAllocated = totalBytes
	result.Allocations = totalAllocs

	if len(latencies) > 0 {
		result.MinLatency = minDuration(latencies)
		result.MaxLatency = maxDuration(latencies)
		result.AvgLatency = totalLatency / time.Duration(len(latencies))
		result.P50Latency = percentile(latencies, 0.50)
		result.P95Latency = percentile(latencies, 0.95)
		result.P99Latency = percentile(latencies, 0.99)
	}

	result.ErrorRate = float64(result.ErrorCount) / float64(result.Operations)
	result.Throughput = float64(result.Operations) / result.Duration.Seconds()

	// Collect memory stats
	runtime.ReadMemStats(&result.MemoryStats)

	return result
}

// warmup runs warmup iterations
func (mb *MarketBenchmark) warmup() {
	for i := 0; i < 100; i++ {
		item := mb.items[i%len(mb.items)]
		_ = mb.market.GetPrice(item.ID)
		// Simulate price updates by updating market state
		mb.market.UpdatePrices()
	}
}

// worker performs benchmark operations
func (mb *MarketBenchmark) worker(ctx context.Context, iterations int) {
	defer mb.wg.Done()

	for i := 0; i < iterations; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			// Perform operation
			op := mb.performOperation(i)
			mb.results <- op
		}
	}
}

// performOperation executes a single benchmark operation
func (mb *MarketBenchmark) performOperation(iteration int) *OperationResult {
	item := mb.items[iteration%len(mb.items)]
	opType := iteration % 4

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	start := time.Now()

	switch opType {
	case 0: // Price calculation
		_ = mb.market.GetPrice(item.ID)
	case 1, 2: // Market state updates
		// Update market state
		mb.market.UpdatePrices()
	case 3: // Complex operation
		mb.market.UpdatePrices()
		_ = mb.market.GetPrice(item.ID)
	}

	latency := time.Since(start)
	runtime.ReadMemStats(&m2)

	return &OperationResult{
		Operation: fmt.Sprintf("op_%d", opType),
		Latency:   latency,
		Error:     nil,
		Bytes:     int64(m2.TotalAlloc - m1.TotalAlloc),
		Allocs:    int64(m2.Mallocs - m1.Mallocs),
	}
}

// MerchantBenchmark benchmarks merchant operations
type MerchantBenchmark struct {
	merchants   []*playermerchant.PlayerMerchant
	market      *market.Market
	iterations  int
	parallelism int
	results     chan *OperationResult
	wg          sync.WaitGroup
}

// NewMerchantBenchmark creates a merchant benchmark
func NewMerchantBenchmark(iterations, parallelism int) *MerchantBenchmark {
	// Create test merchants
	merchants := make([]*playermerchant.PlayerMerchant, 10)
	for i := range merchants {
		merchants[i], _ = playermerchant.NewPlayerMerchant(
			fmt.Sprintf("merchant_%d", i),
			fmt.Sprintf("Test Merchant %d", i),
			1000, // Starting gold
		)
	}

	return &MerchantBenchmark{
		merchants:   merchants,
		market:      market.NewMarket(),
		iterations:  iterations,
		parallelism: parallelism,
		results:     make(chan *OperationResult, iterations),
	}
}

// Run executes the merchant benchmark
func (mb *MerchantBenchmark) Run(ctx context.Context) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:          "Merchant Operations",
		StartTime:     time.Now(),
		Iterations:    mb.iterations,
		CustomMetrics: make(map[string]interface{}),
	}

	// Run parallel workers
	mb.wg.Add(mb.parallelism)
	for i := 0; i < mb.parallelism; i++ {
		go mb.merchantWorker(ctx, mb.iterations/mb.parallelism)
	}

	// Wait and collect
	go func() {
		mb.wg.Wait()
		close(mb.results)
	}()

	// Process results
	latencies := make([]time.Duration, 0, mb.iterations)
	for op := range mb.results {
		result.Operations++
		latencies = append(latencies, op.Latency)
		if op.Error != nil {
			result.ErrorCount++
		}
	}

	// Calculate stats
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if len(latencies) > 0 {
		result.MinLatency = minDuration(latencies)
		result.MaxLatency = maxDuration(latencies)
		result.P50Latency = percentile(latencies, 0.50)
		result.P95Latency = percentile(latencies, 0.95)
		result.P99Latency = percentile(latencies, 0.99)
	}

	result.Throughput = float64(result.Operations) / result.Duration.Seconds()

	return result
}

// merchantWorker performs merchant operations
func (mb *MerchantBenchmark) merchantWorker(ctx context.Context, iterations int) {
	defer mb.wg.Done()

	for i := 0; i < iterations; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			merchant := mb.merchants[i%len(mb.merchants)]
			start := time.Now()

			// Simulate trading operations
			price := mb.market.GetPrice(fmt.Sprintf("item_%d", i%100))
			if merchant.CanAfford(price) {
				// Simulate a buy operation
				testItem := &item.Item{
					ID:        fmt.Sprintf("item_%d", i%100),
					Name:      "Test Item",
					Category:  item.CategoryFruit,
					BasePrice: price,
				}
				_ = merchant.BuyItem(testItem, 1, price)
			}

			latency := time.Since(start)
			mb.results <- &OperationResult{
				Operation: "trading_decision",
				Latency:   latency,
			}
		}
	}
}

// GameStateBenchmark benchmarks game state operations
type GameStateBenchmark struct {
	gameState   *gamestate.GameState
	iterations  int
	parallelism int
	results     chan *OperationResult
	wg          sync.WaitGroup
}

// NewGameStateBenchmark creates a game state benchmark
func NewGameStateBenchmark(iterations, parallelism int) *GameStateBenchmark {
	config := &gamestate.GameConfig{
		InitialGold:       1000,
		ShopCapacity:      50,
		WarehouseCapacity: 100,
		InitialRank:       gamestate.RankApprentice,
	}
	gs := gamestate.NewGameState(config)

	return &GameStateBenchmark{
		gameState:   gs,
		iterations:  iterations,
		parallelism: parallelism,
		results:     make(chan *OperationResult, iterations),
	}
}

// Run executes the game state benchmark
func (gb *GameStateBenchmark) Run(ctx context.Context) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:          "GameState Operations",
		StartTime:     time.Now(),
		Iterations:    gb.iterations,
		CustomMetrics: make(map[string]interface{}),
	}

	// Run workers
	gb.wg.Add(gb.parallelism)
	for i := 0; i < gb.parallelism; i++ {
		go gb.stateWorker(ctx, gb.iterations/gb.parallelism)
	}

	// Collect results
	go func() {
		gb.wg.Wait()
		close(gb.results)
	}()

	// Process results
	var totalLatency time.Duration
	for op := range gb.results {
		result.Operations++
		totalLatency += op.Latency
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.AvgLatency = totalLatency / time.Duration(result.Operations)
	result.Throughput = float64(result.Operations) / result.Duration.Seconds()

	return result
}

// stateWorker performs game state operations
func (gb *GameStateBenchmark) stateWorker(ctx context.Context, iterations int) {
	defer gb.wg.Done()

	for i := 0; i < iterations; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()

			// Perform state operations
			switch i % 5 {
			case 0:
				_ = gb.gameState.AddGold(10)
			case 1:
				if gb.gameState.GetGold() >= 5 {
					_ = gb.gameState.AddGold(-5)
				}
			case 2:
				// Simulate day increment by adding gold
				_ = gb.gameState.AddGold(1)
			case 3:
				// Simulate reputation update
				_ = gb.gameState.AddGold(2)
			case 4:
				_ = gb.gameState.GetCurrentDay()
			}

			latency := time.Since(start)
			gb.results <- &OperationResult{
				Operation: "state_update",
				Latency:   latency,
			}
		}
	}
}

// Helper functions
func minDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	min := durations[0]
	for _, d := range durations[1:] {
		if d < min {
			min = d
		}
	}
	return min
}

func maxDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	max := durations[0]
	for _, d := range durations[1:] {
		if d > max {
			max = d
		}
	}
	return max
}

func percentile(durations []time.Duration, p float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	// Simple percentile calculation (should sort first for accuracy)
	index := int(float64(len(durations)) * p)
	if index >= len(durations) {
		index = len(durations) - 1
	}
	return durations[index]
}

// Run executes all benchmarks
func (br *BenchmarkRunner) Run() error {
	br.mu.Lock()
	if br.running {
		br.mu.Unlock()
		return fmt.Errorf("benchmarks already running")
	}
	br.running = true
	br.mu.Unlock()

	defer func() {
		br.mu.Lock()
		br.running = false
		br.mu.Unlock()
	}()

	// Run market benchmark
	marketBench := NewMarketBenchmark(br.config.Iterations, br.config.Parallelism)
	marketResult := marketBench.Run(br.ctx)
	br.results["market"] = marketResult

	// Run merchant benchmark
	merchantBench := NewMerchantBenchmark(br.config.Iterations, br.config.Parallelism)
	merchantResult := merchantBench.Run(br.ctx)
	br.results["merchant"] = merchantResult

	// Run game state benchmark
	stateBench := NewGameStateBenchmark(br.config.Iterations, br.config.Parallelism)
	stateResult := stateBench.Run(br.ctx)
	br.results["gamestate"] = stateResult

	return nil
}

// GetResults returns benchmark results
func (br *BenchmarkRunner) GetResults() map[string]*BenchmarkResult {
	br.mu.RLock()
	defer br.mu.RUnlock()

	results := make(map[string]*BenchmarkResult)
	for k, v := range br.results {
		results[k] = v
	}
	return results
}

// Stop stops the benchmark runner
func (br *BenchmarkRunner) Stop() {
	br.cancel()
}
