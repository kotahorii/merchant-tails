package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector manages Prometheus metrics collection
type MetricsCollector struct {
	// Game metrics
	totalTransactions prometheus.Counter
	successfulTrades  prometheus.Counter
	failedTrades      prometheus.Counter
	revenueTotal      prometheus.Counter
	totalGoldSpent    prometheus.Counter
	profitTotal       prometheus.Gauge
	playerGold        prometheus.Gauge
	playerLevel       prometheus.Gauge
	playerReputation  prometheus.Gauge
	inventorySize     prometheus.Gauge
	inventoryValue    prometheus.Gauge

	// Market metrics
	marketPrices        *prometheus.GaugeVec
	priceVolatility     *prometheus.GaugeVec
	marketDemand        *prometheus.GaugeVec
	marketSupply        *prometheus.GaugeVec
	priceUpdateDuration prometheus.Histogram

	// Performance metrics
	frameTime          prometheus.Histogram
	updateLoopDuration prometheus.Histogram
	renderTime         prometheus.Histogram
	memoryUsage        prometheus.Gauge
	goroutineCount     prometheus.Gauge
	gcPauseTime        prometheus.Histogram

	// System metrics
	saveOperations prometheus.Counter
	loadOperations prometheus.Counter
	saveErrors     prometheus.Counter
	loadErrors     prometheus.Counter
	saveDuration   prometheus.Histogram
	loadDuration   prometheus.Histogram

	// Event metrics
	eventsProcessed     *prometheus.CounterVec
	eventProcessingTime *prometheus.HistogramVec
	eventQueueSize      prometheus.Gauge

	// Quest metrics
	questsCompleted     prometheus.Counter
	questsAbandoned     prometheus.Counter
	questRewardsClaimed prometheus.Counter
	activeQuests        prometheus.Gauge

	// Achievement metrics
	achievementsUnlocked prometheus.Counter
	achievementProgress  *prometheus.GaugeVec

	// Investment metrics
	totalInvestments  prometheus.Counter
	investmentReturns prometheus.Counter
	roi               prometheus.Gauge
	shopLevel         prometheus.Gauge
	bankBalance       prometheus.Gauge
	interestEarned    prometheus.Counter

	// Weather metrics
	weatherChanges prometheus.Counter
	weatherImpact  *prometheus.GaugeVec

	// Concurrent processing metrics
	workerPoolUtilization prometheus.Gauge
	jobsQueued            prometheus.Gauge
	jobsCompleted         prometheus.Counter
	jobsFailed            prometheus.Counter
	jobProcessingTime     prometheus.Histogram

	// Object pool metrics
	poolHitRate       prometheus.Gauge
	poolSize          prometheus.Gauge
	poolAllocations   prometheus.Counter
	poolDeallocations prometheus.Counter

	server *http.Server
	mu     sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		// Game metrics
		totalTransactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_transactions_total",
			Help: "Total number of transactions",
		}),
		successfulTrades: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_successful_trades_total",
			Help: "Total number of successful trades",
		}),
		failedTrades: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_failed_trades_total",
			Help: "Total number of failed trades",
		}),
		revenueTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_revenue_total",
			Help: "Total revenue generated",
		}),
		totalGoldSpent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_gold_spent_total",
			Help: "Total gold spent",
		}),
		profitTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_profit_current",
			Help: "Current total profit",
		}),
		playerGold: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_player_gold",
			Help: "Current player gold amount",
		}),
		playerLevel: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_player_level",
			Help: "Current player level",
		}),
		playerReputation: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_player_reputation",
			Help: "Current player reputation",
		}),
		inventorySize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_inventory_size",
			Help: "Current inventory size",
		}),
		inventoryValue: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_inventory_value",
			Help: "Current inventory value",
		}),

		// Market metrics
		marketPrices: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "merchant_game_market_prices",
			Help: "Current market prices by item",
		}, []string{"item"}),
		priceVolatility: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "merchant_game_price_volatility",
			Help: "Price volatility by item",
		}, []string{"item"}),
		marketDemand: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "merchant_game_market_demand",
			Help: "Market demand by item",
		}, []string{"item"}),
		marketSupply: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "merchant_game_market_supply",
			Help: "Market supply by item",
		}, []string{"item"}),
		priceUpdateDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "merchant_game_price_update_duration_seconds",
			Help:    "Duration of price update operations",
			Buckets: prometheus.DefBuckets,
		}),

		// Performance metrics
		frameTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "merchant_game_frame_time_seconds",
			Help:    "Frame rendering time",
			Buckets: []float64{0.001, 0.005, 0.01, 0.016, 0.02, 0.05, 0.1},
		}),
		updateLoopDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "merchant_game_update_loop_duration_seconds",
			Help:    "Game update loop duration",
			Buckets: prometheus.DefBuckets,
		}),
		renderTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "merchant_game_render_time_seconds",
			Help:    "Rendering time",
			Buckets: prometheus.DefBuckets,
		}),
		memoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		}),
		goroutineCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_goroutines",
			Help: "Current number of goroutines",
		}),
		gcPauseTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "merchant_game_gc_pause_seconds",
			Help:    "GC pause duration",
			Buckets: []float64{0.00001, 0.00005, 0.0001, 0.0005, 0.001, 0.005, 0.01},
		}),

		// System metrics
		saveOperations: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_save_operations_total",
			Help: "Total number of save operations",
		}),
		loadOperations: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_load_operations_total",
			Help: "Total number of load operations",
		}),
		saveErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_save_errors_total",
			Help: "Total number of save errors",
		}),
		loadErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_load_errors_total",
			Help: "Total number of load errors",
		}),
		saveDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "merchant_game_save_duration_seconds",
			Help:    "Duration of save operations",
			Buckets: prometheus.DefBuckets,
		}),
		loadDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "merchant_game_load_duration_seconds",
			Help:    "Duration of load operations",
			Buckets: prometheus.DefBuckets,
		}),

		// Event metrics
		eventsProcessed: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "merchant_game_events_processed_total",
			Help: "Total events processed by type",
		}, []string{"event_type"}),
		eventProcessingTime: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "merchant_game_event_processing_seconds",
			Help:    "Event processing time by type",
			Buckets: prometheus.DefBuckets,
		}, []string{"event_type"}),
		eventQueueSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_event_queue_size",
			Help: "Current event queue size",
		}),

		// Quest metrics
		questsCompleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_quests_completed_total",
			Help: "Total quests completed",
		}),
		questsAbandoned: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_quests_abandoned_total",
			Help: "Total quests abandoned",
		}),
		questRewardsClaimed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_quest_rewards_claimed_total",
			Help: "Total quest rewards claimed",
		}),
		activeQuests: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_active_quests",
			Help: "Current number of active quests",
		}),

		// Achievement metrics
		achievementsUnlocked: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_achievements_unlocked_total",
			Help: "Total achievements unlocked",
		}),
		achievementProgress: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "merchant_game_achievement_progress",
			Help: "Achievement progress by ID",
		}, []string{"achievement_id"}),

		// Investment metrics
		totalInvestments: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_investments_total",
			Help: "Total investments made",
		}),
		investmentReturns: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_investment_returns_total",
			Help: "Total investment returns",
		}),
		roi: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_roi_percentage",
			Help: "Current ROI percentage",
		}),
		shopLevel: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_shop_level",
			Help: "Current shop level",
		}),
		bankBalance: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_bank_balance",
			Help: "Current bank balance",
		}),
		interestEarned: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_interest_earned_total",
			Help: "Total interest earned",
		}),

		// Weather metrics
		weatherChanges: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_weather_changes_total",
			Help: "Total weather changes",
		}),
		weatherImpact: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "merchant_game_weather_impact",
			Help: "Weather impact on market by type",
		}, []string{"weather_type"}),

		// Concurrent processing metrics
		workerPoolUtilization: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_worker_pool_utilization",
			Help: "Worker pool utilization percentage",
		}),
		jobsQueued: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_jobs_queued",
			Help: "Current number of queued jobs",
		}),
		jobsCompleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_jobs_completed_total",
			Help: "Total jobs completed",
		}),
		jobsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_jobs_failed_total",
			Help: "Total jobs failed",
		}),
		jobProcessingTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "merchant_game_job_processing_seconds",
			Help:    "Job processing time",
			Buckets: prometheus.DefBuckets,
		}),

		// Object pool metrics
		poolHitRate: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_pool_hit_rate",
			Help: "Object pool hit rate",
		}),
		poolSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "merchant_game_pool_size",
			Help: "Current pool size",
		}),
		poolAllocations: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_pool_allocations_total",
			Help: "Total pool allocations",
		}),
		poolDeallocations: promauto.NewCounter(prometheus.CounterOpts{
			Name: "merchant_game_pool_deallocations_total",
			Help: "Total pool deallocations",
		}),
	}

	return mc
}

// StartServer starts the Prometheus metrics HTTP server
func (mc *MetricsCollector) StartServer(port int) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.server != nil {
		return fmt.Errorf("metrics server already running")
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mc.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := mc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	return nil
}

// StopServer stops the metrics server
func (mc *MetricsCollector) StopServer() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mc.server.Shutdown(ctx)
	mc.server = nil
	return err
}

// RecordTransaction records a transaction metric
func (mc *MetricsCollector) RecordTransaction(success bool, revenue float64) {
	mc.totalTransactions.Inc()
	if success {
		mc.successfulTrades.Inc()
		mc.revenueTotal.Add(revenue)
	} else {
		mc.failedTrades.Inc()
	}
}

// UpdatePlayerMetrics updates player-related metrics
func (mc *MetricsCollector) UpdatePlayerMetrics(gold int, level int, reputation float64) {
	mc.playerGold.Set(float64(gold))
	mc.playerLevel.Set(float64(level))
	mc.playerReputation.Set(reputation)
}

// UpdateInventoryMetrics updates inventory metrics
func (mc *MetricsCollector) UpdateInventoryMetrics(size int, value float64) {
	mc.inventorySize.Set(float64(size))
	mc.inventoryValue.Set(value)
}

// UpdateMarketPrice updates market price for an item
func (mc *MetricsCollector) UpdateMarketPrice(itemID string, price float64) {
	mc.marketPrices.WithLabelValues(itemID).Set(price)
}

// UpdateMarketVolatility updates price volatility for an item
func (mc *MetricsCollector) UpdateMarketVolatility(itemID string, volatility float64) {
	mc.priceVolatility.WithLabelValues(itemID).Set(volatility)
}

// UpdateMarketDemand updates market demand for an item
func (mc *MetricsCollector) UpdateMarketDemand(itemID string, demand float64) {
	mc.marketDemand.WithLabelValues(itemID).Set(demand)
}

// UpdateMarketSupply updates market supply for an item
func (mc *MetricsCollector) UpdateMarketSupply(itemID string, supply float64) {
	mc.marketSupply.WithLabelValues(itemID).Set(supply)
}

// RecordPriceUpdate records the duration of a price update
func (mc *MetricsCollector) RecordPriceUpdate(duration time.Duration) {
	mc.priceUpdateDuration.Observe(duration.Seconds())
}

// RecordFrameTime records frame rendering time
func (mc *MetricsCollector) RecordFrameTime(duration time.Duration) {
	mc.frameTime.Observe(duration.Seconds())
}

// RecordUpdateLoop records game update loop duration
func (mc *MetricsCollector) RecordUpdateLoop(duration time.Duration) {
	mc.updateLoopDuration.Observe(duration.Seconds())
}

// RecordUpgrade records an upgrade purchase
func (mc *MetricsCollector) RecordUpgrade(upgradeType string, cost int) {
	mc.totalTransactions.Inc()
	mc.totalGoldSpent.Add(float64(cost))
}

// RecordRenderTime records rendering duration
func (mc *MetricsCollector) RecordRenderTime(duration time.Duration) {
	mc.renderTime.Observe(duration.Seconds())
}

// UpdateMemoryUsage updates memory usage metric
func (mc *MetricsCollector) UpdateMemoryUsage(bytes uint64) {
	mc.memoryUsage.Set(float64(bytes))
}

// UpdateGoroutineCount updates goroutine count
func (mc *MetricsCollector) UpdateGoroutineCount(count int) {
	mc.goroutineCount.Set(float64(count))
}

// RecordGCPause records GC pause duration
func (mc *MetricsCollector) RecordGCPause(duration time.Duration) {
	mc.gcPauseTime.Observe(duration.Seconds())
}

// RecordSaveOperation records a save operation
func (mc *MetricsCollector) RecordSaveOperation(success bool, duration time.Duration) {
	mc.saveOperations.Inc()
	mc.saveDuration.Observe(duration.Seconds())
	if !success {
		mc.saveErrors.Inc()
	}
}

// RecordLoadOperation records a load operation
func (mc *MetricsCollector) RecordLoadOperation(success bool, duration time.Duration) {
	mc.loadOperations.Inc()
	mc.loadDuration.Observe(duration.Seconds())
	if !success {
		mc.loadErrors.Inc()
	}
}

// RecordEvent records an event processing metric
func (mc *MetricsCollector) RecordEvent(eventType string, duration time.Duration) {
	mc.eventsProcessed.WithLabelValues(eventType).Inc()
	mc.eventProcessingTime.WithLabelValues(eventType).Observe(duration.Seconds())
}

// UpdateEventQueueSize updates event queue size
func (mc *MetricsCollector) UpdateEventQueueSize(size int) {
	mc.eventQueueSize.Set(float64(size))
}

// RecordQuestCompletion records quest completion
func (mc *MetricsCollector) RecordQuestCompletion() {
	mc.questsCompleted.Inc()
}

// RecordQuestAbandonment records quest abandonment
func (mc *MetricsCollector) RecordQuestAbandonment() {
	mc.questsAbandoned.Inc()
}

// RecordQuestRewardClaim records quest reward claim
func (mc *MetricsCollector) RecordQuestRewardClaim() {
	mc.questRewardsClaimed.Inc()
}

// UpdateActiveQuests updates active quest count
func (mc *MetricsCollector) UpdateActiveQuests(count int) {
	mc.activeQuests.Set(float64(count))
}

// RecordAchievementUnlock records achievement unlock
func (mc *MetricsCollector) RecordAchievementUnlock() {
	mc.achievementsUnlocked.Inc()
}

// UpdateAchievementProgress updates achievement progress
func (mc *MetricsCollector) UpdateAchievementProgress(achievementID string, progress float64) {
	mc.achievementProgress.WithLabelValues(achievementID).Set(progress)
}

// RecordInvestment records an investment
func (mc *MetricsCollector) RecordInvestment(amount float64) {
	mc.totalInvestments.Inc()
}

// RecordInvestmentReturn records investment return
func (mc *MetricsCollector) RecordInvestmentReturn(amount float64) {
	mc.investmentReturns.Add(amount)
}

// UpdateROI updates ROI metric
func (mc *MetricsCollector) UpdateROI(roi float64) {
	mc.roi.Set(roi)
}

// UpdateShopLevel updates shop level
func (mc *MetricsCollector) UpdateShopLevel(level int) {
	mc.shopLevel.Set(float64(level))
}

// UpdateBankBalance updates bank balance
func (mc *MetricsCollector) UpdateBankBalance(balance float64) {
	mc.bankBalance.Set(balance)
}

// RecordInterestEarned records interest earned
func (mc *MetricsCollector) RecordInterestEarned(amount float64) {
	mc.interestEarned.Add(amount)
}

// RecordWeatherChange records weather change
func (mc *MetricsCollector) RecordWeatherChange() {
	mc.weatherChanges.Inc()
}

// UpdateWeatherImpact updates weather impact
func (mc *MetricsCollector) UpdateWeatherImpact(weatherType string, impact float64) {
	mc.weatherImpact.WithLabelValues(weatherType).Set(impact)
}

// UpdateWorkerPoolUtilization updates worker pool utilization
func (mc *MetricsCollector) UpdateWorkerPoolUtilization(utilization float64) {
	mc.workerPoolUtilization.Set(utilization)
}

// UpdateJobsQueued updates queued job count
func (mc *MetricsCollector) UpdateJobsQueued(count int) {
	mc.jobsQueued.Set(float64(count))
}

// RecordJobCompletion records job completion
func (mc *MetricsCollector) RecordJobCompletion(success bool, duration time.Duration) {
	if success {
		mc.jobsCompleted.Inc()
	} else {
		mc.jobsFailed.Inc()
	}
	mc.jobProcessingTime.Observe(duration.Seconds())
}

// UpdatePoolMetrics updates object pool metrics
func (mc *MetricsCollector) UpdatePoolMetrics(hitRate float64, size int) {
	mc.poolHitRate.Set(hitRate)
	mc.poolSize.Set(float64(size))
}

// RecordPoolAllocation records pool allocation
func (mc *MetricsCollector) RecordPoolAllocation() {
	mc.poolAllocations.Inc()
}

// RecordPoolDeallocation records pool deallocation
func (mc *MetricsCollector) RecordPoolDeallocation() {
	mc.poolDeallocations.Inc()
}
