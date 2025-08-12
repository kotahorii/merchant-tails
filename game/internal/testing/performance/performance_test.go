package performance

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// TestMarketBenchmark tests market operation performance
func TestMarketBenchmark(t *testing.T) {
	config := &BenchmarkConfig{
		Name:            "Market Operations Test",
		Duration:        5 * time.Second,
		WarmupDuration:  1 * time.Second,
		Iterations:      10000,
		Parallelism:     4,
		MemoryProfile:   true,
		MetricsInterval: 100 * time.Millisecond,
	}

	_ = NewBenchmarkRunner(config)

	// Run specific benchmark
	bench := NewMarketBenchmark(config.Iterations, config.Parallelism)
	result := bench.Run(context.Background())

	// Validate results
	if result.Operations == 0 {
		t.Error("No operations were performed")
	}

	if result.Throughput < 1000 {
		t.Errorf("Throughput too low: %.2f ops/sec", result.Throughput)
	}

	if result.ErrorRate > 0.01 {
		t.Errorf("Error rate too high: %.2f%%", result.ErrorRate*100)
	}

	// Check latency
	if result.P95Latency > 10*time.Millisecond {
		t.Errorf("P95 latency too high: %v", result.P95Latency)
	}

	t.Logf("Market Benchmark Results:")
	t.Logf("  Operations: %d", result.Operations)
	t.Logf("  Throughput: %.2f ops/sec", result.Throughput)
	t.Logf("  Avg Latency: %v", result.AvgLatency)
	t.Logf("  P95 Latency: %v", result.P95Latency)
	t.Logf("  P99 Latency: %v", result.P99Latency)
	t.Logf("  Memory Allocated: %d bytes", result.BytesAllocated)
}

// TestMerchantBenchmark tests merchant operation performance
func TestMerchantBenchmark(t *testing.T) {
	bench := NewMerchantBenchmark(5000, 2)
	result := bench.Run(context.Background())

	if result.Operations == 0 {
		t.Error("No merchant operations were performed")
	}

	if result.Throughput < 500 {
		t.Errorf("Merchant throughput too low: %.2f ops/sec", result.Throughput)
	}

	t.Logf("Merchant Benchmark Results:")
	t.Logf("  Operations: %d", result.Operations)
	t.Logf("  Throughput: %.2f ops/sec", result.Throughput)
	t.Logf("  Min Latency: %v", result.MinLatency)
	t.Logf("  Max Latency: %v", result.MaxLatency)
}

// TestGameStateBenchmark tests game state operation performance
func TestGameStateBenchmark(t *testing.T) {
	bench := NewGameStateBenchmark(10000, 4)
	result := bench.Run(context.Background())

	if result.Operations == 0 {
		t.Error("No game state operations were performed")
	}

	if result.AvgLatency > time.Microsecond*100 {
		t.Errorf("Game state operations too slow: %v", result.AvgLatency)
	}

	t.Logf("GameState Benchmark Results:")
	t.Logf("  Operations: %d", result.Operations)
	t.Logf("  Throughput: %.2f ops/sec", result.Throughput)
	t.Logf("  Avg Latency: %v", result.AvgLatency)
}

// TestFullBenchmarkSuite runs all benchmarks
func TestFullBenchmarkSuite(t *testing.T) {
	config := &BenchmarkConfig{
		Name:            "Full Benchmark Suite",
		Duration:        10 * time.Second,
		WarmupDuration:  2 * time.Second,
		Iterations:      50000,
		Parallelism:     8,
		MemoryProfile:   true,
		CPUProfile:      true,
		MetricsInterval: 500 * time.Millisecond,
	}

	runner := NewBenchmarkRunner(config)

	// Run all benchmarks
	err := runner.Run()
	if err != nil {
		t.Fatalf("Failed to run benchmarks: %v", err)
	}

	// Get results
	results := runner.GetResults()

	// Validate each benchmark
	for name, result := range results {
		t.Logf("\n%s Results:", name)
		t.Logf("  Duration: %v", result.Duration)
		t.Logf("  Operations: %d", result.Operations)
		t.Logf("  Throughput: %.2f ops/sec", result.Throughput)
		t.Logf("  Error Rate: %.2f%%", result.ErrorRate*100)
		t.Logf("  Avg Latency: %v", result.AvgLatency)
		t.Logf("  P50 Latency: %v", result.P50Latency)
		t.Logf("  P95 Latency: %v", result.P95Latency)
		t.Logf("  P99 Latency: %v", result.P99Latency)

		// Performance assertions
		if result.Throughput < 100 {
			t.Errorf("%s: Throughput below minimum (100 ops/sec): %.2f", name, result.Throughput)
		}

		if result.ErrorRate > 0.05 {
			t.Errorf("%s: Error rate above threshold (5%%): %.2f%%", name, result.ErrorRate*100)
		}
	}

	// Stop runner
	runner.Stop()
}

// TestNormalLoadScenario tests normal load conditions
func TestNormalLoadScenario(t *testing.T) {
	config := &LoadTestConfig{
		Name:             "Normal Load Test",
		Duration:         30 * time.Second,
		RampUpDuration:   5 * time.Second,
		MaxUsers:         50,
		StartUsers:       10,
		UserSpawnRate:    2,
		ThinkTime:        time.Second,
		CollectInterval:  time.Second,
		ScenarioType:     ScenarioNormal,
		TargetTPS:        100,
		ErrorThreshold:   0.05,
		LatencyThreshold: 100 * time.Millisecond,
	}

	runner := NewLoadTestRunner(config)

	// Run load test
	go func() {
		err := runner.Run()
		if err != nil {
			t.Errorf("Load test failed: %v", err)
		}
	}()

	// Let it run for a bit
	time.Sleep(10 * time.Second)

	// Check intermediate metrics
	metrics := runner.GetMetrics()
	if metrics.CurrentUsers == 0 {
		t.Error("No users are active")
	}

	// Stop test
	runner.Stop()

	// Wait for completion
	time.Sleep(2 * time.Second)

	// Get final metrics
	finalMetrics := runner.GetMetrics()

	// Generate report
	report := runner.GenerateReport()
	t.Log(report)

	// Validate results
	if finalMetrics.TotalTransactions == 0 {
		t.Error("No transactions were completed")
	}

	if finalMetrics.ErrorRate > config.ErrorThreshold {
		t.Errorf("Error rate %.2f%% exceeds threshold %.2f%%",
			finalMetrics.ErrorRate*100, config.ErrorThreshold*100)
	}

	if finalMetrics.P95Latency > config.LatencyThreshold {
		t.Errorf("P95 latency %v exceeds threshold %v",
			finalMetrics.P95Latency, config.LatencyThreshold)
	}
}

// TestPeakLoadScenario tests peak load conditions
func TestPeakLoadScenario(t *testing.T) {
	config := &LoadTestConfig{
		Name:             "Peak Load Test",
		Duration:         20 * time.Second,
		RampUpDuration:   3 * time.Second,
		MaxUsers:         100,
		StartUsers:       20,
		UserSpawnRate:    5,
		ThinkTime:        500 * time.Millisecond,
		CollectInterval:  500 * time.Millisecond,
		ScenarioType:     ScenarioPeakTrading,
		TargetTPS:        500,
		ErrorThreshold:   0.1,
		LatencyThreshold: 200 * time.Millisecond,
	}

	runner := NewLoadTestRunner(config)

	// Run test in background
	done := make(chan bool)
	go func() {
		err := runner.Run()
		if err != nil {
			t.Errorf("Peak load test failed: %v", err)
		}
		done <- true
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// Test completed
	case <-time.After(25 * time.Second):
		runner.Stop()
		t.Error("Peak load test timed out")
	}

	// Get metrics
	metrics := runner.GetMetrics()

	t.Logf("\nPeak Load Test Results:")
	t.Logf("  Peak Users: %d", metrics.PeakUsers)
	t.Logf("  Total Transactions: %d", metrics.TotalTransactions)
	t.Logf("  TPS: %.2f", metrics.TPS)
	t.Logf("  Error Rate: %.2f%%", metrics.ErrorRate*100)
	t.Logf("  P95 Latency: %v", metrics.P95Latency)
	t.Logf("  P99 Latency: %v", metrics.P99Latency)
}

// TestStressTest tests system under stress
func TestStressTest(t *testing.T) {
	t.Skip("Skipping stress test in normal test runs")

	config := &LoadTestConfig{
		Name:            "Stress Test",
		Duration:        60 * time.Second,
		RampUpDuration:  10 * time.Second,
		MaxUsers:        500,
		StartUsers:      50,
		UserSpawnRate:   10,
		ThinkTime:       100 * time.Millisecond,
		CollectInterval: time.Second,
		ScenarioType:    ScenarioNormal,
		// No targets - we want to find breaking point
	}

	runner := NewLoadTestRunner(config)

	// Run stress test
	go runner.Run()

	// Monitor for issues
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(70 * time.Second)

	for {
		select {
		case <-ticker.C:
			metrics := runner.GetMetrics()
			t.Logf("Stress Test Progress: Users=%d, TPS=%.2f, Errors=%.2f%%",
				metrics.CurrentUsers, metrics.TPS, metrics.ErrorRate*100)

			// Check for system degradation
			if metrics.ErrorRate > 0.5 {
				t.Log("System showing signs of stress (50% error rate)")
				runner.Stop()
				return
			}

		case <-timeout:
			runner.Stop()

			finalMetrics := runner.GetMetrics()
			t.Logf("\nStress Test Completed:")
			t.Logf("  Max Users Handled: %d", finalMetrics.PeakUsers)
			t.Logf("  Max TPS Achieved: %.2f", finalMetrics.TPS)
			t.Logf("  Final Error Rate: %.2f%%", finalMetrics.ErrorRate*100)
			return
		}
	}
}

// BenchmarkMarketOperations standard Go benchmark
func BenchmarkMarketOperations(b *testing.B) {
	bench := NewMarketBenchmark(b.N, 1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = bench.performOperation(i)
	}

	b.StopTimer()
}

// BenchmarkMerchantDecisions standard Go benchmark
func BenchmarkMerchantDecisions(b *testing.B) {
	bench := NewMerchantBenchmark(b.N, 1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		merchant := bench.merchants[i%len(bench.merchants)]
		// Simulate trading operations
		price := bench.market.GetPrice(fmt.Sprintf("item_%d", i%100))
		if merchant.CanAfford(price) {
			// Simple operation to avoid extra dependencies
			_ = price
		}
	}

	b.StopTimer()
}

// BenchmarkGameStateUpdates standard Go benchmark
func BenchmarkGameStateUpdates(b *testing.B) {
	bench := NewGameStateBenchmark(b.N, 1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 5 {
		case 0:
			bench.gameState.AddGold(10)
		case 1:
			if bench.gameState.GetGold() >= 5 {
				bench.gameState.AddGold(-5)
			}
		case 2:
			// Simulate day increment
			bench.gameState.AddGold(1)
		case 3:
			// Simulate reputation update
			bench.gameState.AddGold(2)
		case 4:
			// Get current day
			_ = bench.gameState.GetCurrentDay()
		}
	}

	b.StopTimer()
}

// TestMemoryLeaks checks for memory leaks
func TestMemoryLeaks(t *testing.T) {
	// Run a short load test and check for memory growth
	config := &LoadTestConfig{
		Name:            "Memory Leak Test",
		Duration:        10 * time.Second,
		MaxUsers:        10,
		StartUsers:      10,
		ThinkTime:       100 * time.Millisecond,
		CollectInterval: time.Second,
		ScenarioType:    ScenarioNormal,
	}

	runner := NewLoadTestRunner(config)

	// Capture initial memory
	var initialMem, finalMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)

	// Run test
	err := runner.Run()
	if err != nil {
		t.Fatalf("Memory leak test failed: %v", err)
	}

	// Force GC and capture final memory
	runtime.GC()
	runtime.ReadMemStats(&finalMem)

	// Check for excessive memory growth
	growth := finalMem.HeapAlloc - initialMem.HeapAlloc
	growthMB := float64(growth) / (1024 * 1024)

	t.Logf("Memory Growth: %.2f MB", growthMB)
	t.Logf("Goroutines: %d", runtime.NumGoroutine())

	if growthMB > 100 {
		t.Errorf("Excessive memory growth detected: %.2f MB", growthMB)
	}

	if runtime.NumGoroutine() > 100 {
		t.Errorf("Possible goroutine leak: %d goroutines", runtime.NumGoroutine())
	}
}
