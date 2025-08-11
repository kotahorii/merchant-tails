# Merchant Tails Monitoring System

This directory contains the monitoring configuration for Merchant Tails using Prometheus and Grafana.

## Overview

The monitoring system provides real-time metrics and alerting for:
- Game performance (frame time, memory usage, GC pauses)
- Business metrics (transactions, revenue, inventory)
- System health (job queues, worker pools, save operations)
- Player progression (level, achievements, quests)

## Quick Start

### 1. Start the monitoring stack

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

### 2. Start the game with metrics enabled

The game automatically starts a metrics server on port 8080 when initialized.

### 3. Access the dashboards

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (default login: admin/admin)

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│    Game     │────▶│  Prometheus  │────▶│   Grafana   │
│  (port 8080)│     │  (port 9090) │     │ (port 3000) │
└─────────────┘     └──────────────┘     └─────────────┘
       │                    │
       │                    ▼
       │            ┌──────────────┐
       └───────────▶│    Alerts    │
                    └──────────────┘
```

## Metrics Exported

### Game Metrics
- `merchant_game_player_gold` - Current player gold
- `merchant_game_player_level` - Player level
- `merchant_game_player_reputation` - Player reputation
- `merchant_game_transactions_total` - Total transactions
- `merchant_game_successful_trades_total` - Successful trades
- `merchant_game_failed_trades_total` - Failed trades
- `merchant_game_revenue_total` - Total revenue

### Market Metrics
- `merchant_game_market_prices{item=""}` - Item prices
- `merchant_game_price_volatility{item=""}` - Price volatility
- `merchant_game_market_demand{item=""}` - Market demand
- `merchant_game_market_supply{item=""}` - Market supply

### Performance Metrics
- `merchant_game_frame_time_seconds` - Frame rendering time
- `merchant_game_update_loop_duration_seconds` - Game update loop duration
- `merchant_game_memory_usage_bytes` - Memory usage
- `merchant_game_goroutines` - Active goroutines
- `merchant_game_gc_pause_seconds` - GC pause duration

### System Metrics
- `merchant_game_save_operations_total` - Save operations
- `merchant_game_save_errors_total` - Save errors
- `merchant_game_load_operations_total` - Load operations
- `merchant_game_jobs_queued` - Queued jobs
- `merchant_game_jobs_completed_total` - Completed jobs
- `merchant_game_worker_pool_utilization` - Worker pool usage

## Alerts

Configured alerts include:
- **HighFrameTime**: Frame time exceeds 33ms (30 FPS threshold)
- **ExcessiveMemoryUsage**: Memory usage exceeds 2GB
- **HighGCPause**: GC pause exceeds 10ms
- **HighTransactionFailureRate**: Failure rate exceeds 10%
- **GameServerDown**: Metrics endpoint not responding
- **HighSaveOperationFailures**: Save failures detected

## Custom Dashboards

The included Grafana dashboard (`grafana-dashboard.json`) provides:
- Player progression overview
- Market price trends
- Performance metrics
- System health indicators
- Transaction success rates
- Memory and resource usage

## Integration

### Using in Code

```go
// Initialize metrics collector
collector := monitoring.NewMetricsCollector()

// Start metrics server
err := collector.StartServer(8080)
if err != nil {
    log.Fatal(err)
}

// Create middleware for automatic collection
middleware := monitoring.NewMetricsMiddleware(
    collector,
    gameState,
    market,
    inventory,
)

// Start automatic metrics collection (every 5 seconds)
middleware.Start(5 * time.Second)

// Record custom metrics
middleware.RecordTransaction(true, 150.0)
middleware.RecordFrameTime(16 * time.Millisecond)
```

## Development

### Adding New Metrics

1. Define the metric in `metrics.go`
2. Add collection method in `MetricsCollector`
3. Update middleware if automatic collection needed
4. Add to Prometheus scrape config if new endpoint
5. Update Grafana dashboard

### Testing Alerts

You can test alerts by simulating conditions:

```bash
# Simulate high memory usage
curl -X POST http://localhost:8080/debug/simulate/high-memory

# Simulate transaction failures
curl -X POST http://localhost:8080/debug/simulate/failures
```

## Troubleshooting

### Metrics not appearing
- Check game is running and metrics server started
- Verify Prometheus can reach the game (check targets page)
- Check firewall/network settings

### High memory usage in monitoring
- Reduce retention period in Prometheus
- Adjust scrape intervals
- Limit number of metrics collected

### Grafana dashboard not loading
- Ensure Prometheus datasource is configured
- Check dashboard JSON is valid
- Verify metric names match those exported

## Performance Impact

The monitoring system has minimal performance impact:
- Metrics collection: ~1-2% CPU overhead
- Memory overhead: ~10-20MB
- Network: ~100KB/s at default scrape interval

## Security

For production deployments:
1. Change default Grafana password
2. Enable authentication on Prometheus
3. Use TLS for metrics endpoints
4. Restrict network access to monitoring ports
5. Implement RBAC for dashboard access