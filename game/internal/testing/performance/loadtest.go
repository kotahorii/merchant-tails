package performance

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/event"
	"github.com/yourusername/merchant-tails/game/internal/domain/gameloop"
	"github.com/yourusername/merchant-tails/game/internal/domain/gamestate"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
	playermerchant "github.com/yourusername/merchant-tails/game/internal/domain/merchant"
)

// LoadTestRunner runs load tests
type LoadTestRunner struct {
	mu           sync.RWMutex
	config       *LoadTestConfig
	scenario     LoadScenario
	metrics      *LoadTestMetrics
	gameLoop     gameloop.GameLoop
	market       *market.Market
	gameState    *gamestate.GameState
	eventBus     *event.EventBus
	virtualUsers []*VirtualUser
	running      atomic.Bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// LoadTestConfig defines load test configuration
type LoadTestConfig struct {
	Name             string
	Duration         time.Duration
	RampUpDuration   time.Duration
	MaxUsers         int
	StartUsers       int
	UserSpawnRate    int // Users per second
	ThinkTime        time.Duration
	MaxTransactions  int
	CollectInterval  time.Duration
	ScenarioType     ScenarioType
	TargetTPS        float64 // Target transactions per second
	ErrorThreshold   float64 // Maximum error rate
	LatencyThreshold time.Duration
}

// ScenarioType defines different load test scenarios
type ScenarioType int

const (
	ScenarioNormal ScenarioType = iota
	ScenarioPeakTrading
	ScenarioMarketCrash
	ScenarioHighVolume
	ScenarioEndurance
	ScenarioSpike
	ScenarioStress
)

// LoadScenario defines behavior for load test scenarios
type LoadScenario interface {
	Execute(ctx context.Context, user *VirtualUser) error
	GetName() string
	GetWeight() float64
}

// VirtualUser simulates a player
type VirtualUser struct {
	ID           string
	merchant     *playermerchant.PlayerMerchant
	market       *market.Market
	gameState    *gamestate.GameState
	eventBus     *event.EventBus
	transactions int64
	errors       int64
	totalLatency int64
	lastAction   time.Time
	thinkTime    time.Duration
	rng          *rand.Rand
}

// LoadTestMetrics tracks load test metrics
type LoadTestMetrics struct {
	mu                     sync.RWMutex
	StartTime              time.Time
	EndTime                time.Time
	TotalTransactions      int64
	SuccessfulTransactions int64
	FailedTransactions     int64
	TotalLatency           int64
	MinLatency             time.Duration
	MaxLatency             time.Duration
	AvgLatency             time.Duration
	P50Latency             time.Duration
	P95Latency             time.Duration
	P99Latency             time.Duration
	CurrentUsers           int32
	PeakUsers              int32
	TPS                    float64 // Transactions per second
	ErrorRate              float64
	ResponseTimes          []time.Duration
	UserMetrics            map[string]*UserMetrics
	SystemMetrics          *SystemMetrics
}

// UserMetrics tracks per-user metrics
type UserMetrics struct {
	Transactions int64
	Errors       int64
	AvgLatency   time.Duration
	LastActive   time.Time
}

// SystemMetrics tracks system-level metrics
type SystemMetrics struct {
	CPUUsage       float64
	MemoryUsage    uint64
	GoroutineCount int
	HeapAlloc      uint64
	HeapObjects    uint64
	GCPauseTime    time.Duration
	NetworkLatency time.Duration
}

// NewLoadTestRunner creates a new load test runner
func NewLoadTestRunner(config *LoadTestConfig) *LoadTestRunner {
	ctx, cancel := context.WithCancel(context.Background())

	return &LoadTestRunner{
		config:  config,
		metrics: NewLoadTestMetrics(),
		gameLoop: gameloop.NewStandardGameLoop(&gameloop.Config{
			TargetFPS: 60,
		}),
		market: market.NewMarket(),
		gameState: gamestate.NewGameState(&gamestate.GameConfig{
			InitialGold:       1000,
			ShopCapacity:      50,
			WarehouseCapacity: 100,
			InitialRank:       gamestate.RankApprentice,
		}),
		eventBus: event.GetGlobalEventBus(),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// NewLoadTestMetrics creates new metrics
func NewLoadTestMetrics() *LoadTestMetrics {
	return &LoadTestMetrics{
		UserMetrics:   make(map[string]*UserMetrics),
		ResponseTimes: make([]time.Duration, 0, 10000),
		SystemMetrics: &SystemMetrics{},
	}
}

// NormalScenario simulates normal gameplay
type NormalScenario struct {
	market    *market.Market
	gameState *gamestate.GameState
	eventBus  *event.EventBus
}

// Execute runs the normal scenario
func (ns *NormalScenario) Execute(ctx context.Context, user *VirtualUser) error {
	actions := []func() error{
		func() error { return ns.buyItem(user) },
		func() error { return ns.sellItem(user) },
		func() error { return ns.checkPrices(user) },
		func() error { return ns.manageInventory(user) },
	}

	// Random action selection
	action := actions[user.rng.Intn(len(actions))]
	return action()
}

// GetName returns scenario name
func (ns *NormalScenario) GetName() string {
	return "Normal Trading"
}

// GetWeight returns scenario weight
func (ns *NormalScenario) GetWeight() float64 {
	return 1.0
}

// buyItem simulates buying an item
func (ns *NormalScenario) buyItem(user *VirtualUser) error {
	// Create test item
	testItem := &item.Item{
		ID:        fmt.Sprintf("item_%d", user.rng.Intn(100)),
		Name:      "Test Item",
		Category:  item.CategoryFruit,
		BasePrice: 10 + user.rng.Intn(90),
	}

	// Get price
	price := ns.market.GetPrice(testItem.ID)

	// Check if user has enough gold
	if !user.merchant.CanAfford(price) {
		return fmt.Errorf("insufficient gold")
	}

	// Execute purchase
	_ = user.merchant.BuyItem(testItem, 1, price)

	// Simulate event publishing (simplified for testing)

	return nil
}

// sellItem simulates selling an item
func (ns *NormalScenario) sellItem(user *VirtualUser) error {
	// Create test item
	testItem := &item.Item{
		ID:        fmt.Sprintf("item_%d", user.rng.Intn(100)),
		Name:      "Test Item",
		Category:  item.CategoryPotion,
		BasePrice: 15 + user.rng.Intn(85),
	}

	// Get price
	price := ns.market.GetPrice(testItem.ID)

	// Execute sale (simulate by selling an item)
	_ = user.merchant.SellItem(testItem.ID, 1, price)

	// Simulate event publishing (simplified for testing)

	return nil
}

// checkPrices simulates checking market prices
func (ns *NormalScenario) checkPrices(user *VirtualUser) error {
	// Check multiple item prices
	for i := 0; i < 5; i++ {
		itemID := fmt.Sprintf("item_%d", user.rng.Intn(100))
		_ = ns.market.GetPrice(itemID)
	}
	return nil
}

// manageInventory simulates inventory management
func (ns *NormalScenario) manageInventory(user *VirtualUser) error {
	// Simulate inventory operations
	time.Sleep(time.Millisecond * time.Duration(10+user.rng.Intn(40)))
	return nil
}

// PeakTradingScenario simulates peak trading hours
type PeakTradingScenario struct {
	NormalScenario
	intensity float64
}

// Execute runs the peak trading scenario
func (pts *PeakTradingScenario) Execute(ctx context.Context, user *VirtualUser) error {
	// Increase trading frequency
	for i := 0; i < int(pts.intensity); i++ {
		if err := pts.NormalScenario.Execute(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

// GetName returns scenario name
func (pts *PeakTradingScenario) GetName() string {
	return "Peak Trading"
}

// GetWeight returns scenario weight
func (pts *PeakTradingScenario) GetWeight() float64 {
	return pts.intensity
}

// NewVirtualUser creates a new virtual user
func NewVirtualUser(id string, market *market.Market, gameState *gamestate.GameState, eventBus *event.EventBus) *VirtualUser {
	merchant, _ := playermerchant.NewPlayerMerchant(
		fmt.Sprintf("user_%s", id),
		fmt.Sprintf("Virtual User %s", id),
		1000, // Starting gold
	)

	return &VirtualUser{
		ID:        id,
		merchant:  merchant,
		market:    market,
		gameState: gameState,
		eventBus:  eventBus,
		thinkTime: time.Second,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run starts the load test
func (ltr *LoadTestRunner) Run() error {
	if ltr.running.Load() {
		return fmt.Errorf("load test already running")
	}
	ltr.running.Store(true)
	defer ltr.running.Store(false)

	// Initialize scenario
	ltr.scenario = ltr.createScenario()

	// Start metrics collection
	go ltr.collectMetrics()

	// Start game loop
	_ = ltr.gameLoop.Start(ltr.ctx)
	defer func() { _ = ltr.gameLoop.Stop() }()

	// Record start time
	ltr.metrics.StartTime = time.Now()

	// Ramp up users
	if err := ltr.rampUpUsers(); err != nil {
		return err
	}

	// Run main test
	testCtx, cancel := context.WithTimeout(ltr.ctx, ltr.config.Duration)
	defer cancel()

	// Run users
	var wg sync.WaitGroup
	for _, user := range ltr.virtualUsers {
		wg.Add(1)
		go ltr.runUser(testCtx, user, &wg)
	}

	// Wait for completion
	wg.Wait()

	// Record end time
	ltr.metrics.EndTime = time.Now()

	// Calculate final metrics
	ltr.calculateFinalMetrics()

	return nil
}

// createScenario creates the appropriate scenario
func (ltr *LoadTestRunner) createScenario() LoadScenario {
	switch ltr.config.ScenarioType {
	case ScenarioPeakTrading:
		return &PeakTradingScenario{
			NormalScenario: NormalScenario{
				market:    ltr.market,
				gameState: ltr.gameState,
				eventBus:  ltr.eventBus,
			},
			intensity: 3.0,
		}
	default:
		return &NormalScenario{
			market:    ltr.market,
			gameState: ltr.gameState,
			eventBus:  ltr.eventBus,
		}
	}
}

// rampUpUsers gradually adds users
func (ltr *LoadTestRunner) rampUpUsers() error {
	ltr.virtualUsers = make([]*VirtualUser, 0, ltr.config.MaxUsers)

	// Start with initial users
	for i := 0; i < ltr.config.StartUsers; i++ {
		user := NewVirtualUser(
			fmt.Sprintf("%d", i),
			ltr.market,
			ltr.gameState,
			ltr.eventBus,
		)
		ltr.virtualUsers = append(ltr.virtualUsers, user)
		atomic.AddInt32(&ltr.metrics.CurrentUsers, 1)
	}

	// Ramp up additional users
	if ltr.config.RampUpDuration > 0 {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		rampUpEnd := time.Now().Add(ltr.config.RampUpDuration)
		userCount := len(ltr.virtualUsers)

		for time.Now().Before(rampUpEnd) && userCount < ltr.config.MaxUsers {
			select {
			case <-ticker.C:
				// Add users based on spawn rate
				for i := 0; i < ltr.config.UserSpawnRate && userCount < ltr.config.MaxUsers; i++ {
					user := NewVirtualUser(
						fmt.Sprintf("%d", userCount),
						ltr.market,
						ltr.gameState,
						ltr.eventBus,
					)
					ltr.virtualUsers = append(ltr.virtualUsers, user)
					userCount++
					atomic.AddInt32(&ltr.metrics.CurrentUsers, 1)
				}
			case <-ltr.ctx.Done():
				return ltr.ctx.Err()
			}
		}
	}

	// Update peak users
	current := atomic.LoadInt32(&ltr.metrics.CurrentUsers)
	if current > atomic.LoadInt32(&ltr.metrics.PeakUsers) {
		atomic.StoreInt32(&ltr.metrics.PeakUsers, current)
	}

	return nil
}

// runUser runs a virtual user
func (ltr *LoadTestRunner) runUser(ctx context.Context, user *VirtualUser, wg *sync.WaitGroup) {
	defer wg.Done()
	defer atomic.AddInt32(&ltr.metrics.CurrentUsers, -1)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Think time
			time.Sleep(user.thinkTime)

			// Execute scenario
			start := time.Now()
			err := ltr.scenario.Execute(ctx, user)
			latency := time.Since(start)

			// Record metrics
			atomic.AddInt64(&user.transactions, 1)
			atomic.AddInt64(&ltr.metrics.TotalTransactions, 1)

			if err != nil {
				atomic.AddInt64(&user.errors, 1)
				atomic.AddInt64(&ltr.metrics.FailedTransactions, 1)
			} else {
				atomic.AddInt64(&ltr.metrics.SuccessfulTransactions, 1)
			}

			// Record latency
			atomic.AddInt64(&user.totalLatency, int64(latency))
			atomic.AddInt64(&ltr.metrics.TotalLatency, int64(latency))
			ltr.recordLatency(latency)

			user.lastAction = time.Now()

			// Check transaction limit
			if ltr.config.MaxTransactions > 0 &&
				atomic.LoadInt64(&user.transactions) >= int64(ltr.config.MaxTransactions) {
				return
			}
		}
	}
}

// recordLatency records response time
func (ltr *LoadTestRunner) recordLatency(latency time.Duration) {
	ltr.metrics.mu.Lock()
	defer ltr.metrics.mu.Unlock()

	ltr.metrics.ResponseTimes = append(ltr.metrics.ResponseTimes, latency)

	if ltr.metrics.MinLatency == 0 || latency < ltr.metrics.MinLatency {
		ltr.metrics.MinLatency = latency
	}
	if latency > ltr.metrics.MaxLatency {
		ltr.metrics.MaxLatency = latency
	}
}

// collectMetrics collects metrics periodically
func (ltr *LoadTestRunner) collectMetrics() {
	ticker := time.NewTicker(ltr.config.CollectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ltr.updateMetrics()
		case <-ltr.ctx.Done():
			return
		}
	}
}

// updateMetrics updates current metrics
func (ltr *LoadTestRunner) updateMetrics() {
	ltr.metrics.mu.Lock()
	defer ltr.metrics.mu.Unlock()

	// Calculate TPS
	duration := time.Since(ltr.metrics.StartTime).Seconds()
	if duration > 0 {
		ltr.metrics.TPS = float64(atomic.LoadInt64(&ltr.metrics.TotalTransactions)) / duration
	}

	// Calculate error rate
	total := atomic.LoadInt64(&ltr.metrics.TotalTransactions)
	if total > 0 {
		failed := atomic.LoadInt64(&ltr.metrics.FailedTransactions)
		ltr.metrics.ErrorRate = float64(failed) / float64(total)
	}

	// Update user metrics
	for _, user := range ltr.virtualUsers {
		userMetric := &UserMetrics{
			Transactions: atomic.LoadInt64(&user.transactions),
			Errors:       atomic.LoadInt64(&user.errors),
			LastActive:   user.lastAction,
		}

		if userMetric.Transactions > 0 {
			totalLatency := atomic.LoadInt64(&user.totalLatency)
			userMetric.AvgLatency = time.Duration(totalLatency / userMetric.Transactions)
		}

		ltr.metrics.UserMetrics[user.ID] = userMetric
	}
}

// calculateFinalMetrics calculates final test metrics
func (ltr *LoadTestRunner) calculateFinalMetrics() {
	ltr.metrics.mu.Lock()
	defer ltr.metrics.mu.Unlock()

	// Calculate average latency
	if len(ltr.metrics.ResponseTimes) > 0 {
		var total time.Duration
		for _, latency := range ltr.metrics.ResponseTimes {
			total += latency
		}
		ltr.metrics.AvgLatency = total / time.Duration(len(ltr.metrics.ResponseTimes))

		// Calculate percentiles
		ltr.metrics.P50Latency = percentile(ltr.metrics.ResponseTimes, 0.50)
		ltr.metrics.P95Latency = percentile(ltr.metrics.ResponseTimes, 0.95)
		ltr.metrics.P99Latency = percentile(ltr.metrics.ResponseTimes, 0.99)
	}

	// Final TPS calculation
	duration := ltr.metrics.EndTime.Sub(ltr.metrics.StartTime).Seconds()
	if duration > 0 {
		ltr.metrics.TPS = float64(ltr.metrics.TotalTransactions) / duration
	}
}

// GetMetrics returns current metrics
func (ltr *LoadTestRunner) GetMetrics() *LoadTestMetrics {
	ltr.metrics.mu.RLock()
	defer ltr.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metricsCopy := *ltr.metrics
	return &metricsCopy
}

// Stop stops the load test
func (ltr *LoadTestRunner) Stop() {
	ltr.cancel()
	ltr.running.Store(false)
}

// IsRunning checks if test is running
func (ltr *LoadTestRunner) IsRunning() bool {
	return ltr.running.Load()
}

// GenerateReport generates a test report
func (ltr *LoadTestRunner) GenerateReport() string {
	metrics := ltr.GetMetrics()

	report := fmt.Sprintf(`
Load Test Report
================
Test Name: %s
Duration: %v
Users: Start=%d, Peak=%d, Max=%d

Performance Metrics:
- Total Transactions: %d
- Successful: %d
- Failed: %d
- Error Rate: %.2f%%
- Throughput: %.2f TPS

Latency Statistics:
- Min: %v
- Max: %v
- Avg: %v
- P50: %v
- P95: %v
- P99: %v

Test Status: %s
`,
		ltr.config.Name,
		metrics.EndTime.Sub(metrics.StartTime),
		ltr.config.StartUsers,
		metrics.PeakUsers,
		ltr.config.MaxUsers,
		metrics.TotalTransactions,
		metrics.SuccessfulTransactions,
		metrics.FailedTransactions,
		metrics.ErrorRate*100,
		metrics.TPS,
		metrics.MinLatency,
		metrics.MaxLatency,
		metrics.AvgLatency,
		metrics.P50Latency,
		metrics.P95Latency,
		metrics.P99Latency,
		ltr.getTestStatus(metrics),
	)

	return report
}

// getTestStatus determines if test passed or failed
func (ltr *LoadTestRunner) getTestStatus(metrics *LoadTestMetrics) string {
	// Check error threshold
	if metrics.ErrorRate > ltr.config.ErrorThreshold {
		return fmt.Sprintf("FAILED - Error rate %.2f%% exceeds threshold %.2f%%",
			metrics.ErrorRate*100, ltr.config.ErrorThreshold*100)
	}

	// Check latency threshold
	if ltr.config.LatencyThreshold > 0 && metrics.P95Latency > ltr.config.LatencyThreshold {
		return fmt.Sprintf("FAILED - P95 latency %v exceeds threshold %v",
			metrics.P95Latency, ltr.config.LatencyThreshold)
	}

	// Check TPS target
	if ltr.config.TargetTPS > 0 && metrics.TPS < ltr.config.TargetTPS {
		return fmt.Sprintf("FAILED - TPS %.2f below target %.2f",
			metrics.TPS, ltr.config.TargetTPS)
	}

	return "PASSED"
}
