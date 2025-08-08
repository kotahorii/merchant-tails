package analytics

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

func TestNewPriceChart(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeDaily, ChartTypeLine)

	assert.NotNil(t, chart)
	assert.Equal(t, "apple", chart.itemID)
	assert.Equal(t, TimeframeDaily, chart.timeframe)
	assert.Equal(t, ChartTypeLine, chart.chartType)
	assert.Equal(t, 100, chart.maxDataPoints)
	assert.Equal(t, 24*time.Hour, chart.updateInterval)
}

func TestPriceChart_AddPricePoint(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Add price points
	chart.AddPricePoint(100.0, 10)
	chart.AddPricePoint(105.0, 15)
	chart.AddPricePoint(102.0, 12)

	points := chart.GetDataPoints()
	assert.Len(t, points, 3)
	assert.Equal(t, 100.0, points[0].Price)
	assert.Equal(t, 10, points[0].Volume)
	assert.Equal(t, 105.0, points[1].Price)
	assert.Equal(t, 102.0, points[2].Price)
}

func TestPriceChart_MaxDataPoints(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)
	chart.maxDataPoints = 5

	// Add more than max points
	for i := 0; i < 10; i++ {
		chart.AddPricePoint(float64(100+i), i)
	}

	points := chart.GetDataPoints()
	assert.Len(t, points, 5)
	assert.Equal(t, 105.0, points[0].Price) // Should keep latest 5 points
	assert.Equal(t, 109.0, points[4].Price)
}

func TestPriceChart_CandleData(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeCandle)

	// Add price points within same hour
	baseTime := time.Now()
	chart.AddPricePoint(100.0, 10)
	chart.AddPricePoint(105.0, 15)
	chart.AddPricePoint(95.0, 12)
	chart.AddPricePoint(102.0, 8)

	candles := chart.GetCandles()
	assert.Len(t, candles, 1)

	candle := candles[0]
	assert.Equal(t, 100.0, candle.Open)
	assert.Equal(t, 105.0, candle.High)
	assert.Equal(t, 95.0, candle.Low)
	assert.Equal(t, 102.0, candle.Close)
	assert.Equal(t, 45, candle.Volume) // 10+15+12+8

	_ = baseTime // Suppress unused variable warning
}

func TestPriceChart_MovingAverage(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Add 25 price points for 20-period MA
	for i := 0; i < 25; i++ {
		price := 100.0 + float64(i%5) // Creates some variation
		chart.AddPricePoint(price, 10)
	}

	ma, _, _ := chart.GetIndicators()
	assert.Greater(t, len(ma), 0)

	// MA should be around 102 (average of 100-104 pattern)
	lastMA := ma[len(ma)-1]
	assert.InDelta(t, 102.0, lastMA, 1.0)
}

func TestPriceChart_RSI(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Add increasing prices (should show high RSI)
	for i := 0; i < 20; i++ {
		price := 100.0 + float64(i*2)
		chart.AddPricePoint(price, 10)
	}

	_, rsi, _ := chart.GetIndicators()
	assert.Greater(t, rsi, 70.0) // Should be overbought

	// Add decreasing prices (should show low RSI)
	chart = NewPriceChart("apple", TimeframeHourly, ChartTypeLine)
	for i := 0; i < 20; i++ {
		price := 200.0 - float64(i*2)
		chart.AddPricePoint(price, 10)
	}

	_, rsi, _ = chart.GetIndicators()
	assert.Less(t, rsi, 30.0) // Should be oversold
}

func TestPriceChart_Volatility(t *testing.T) {
	// Test low volatility
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)
	for i := 0; i < 20; i++ {
		price := 100.0 + float64(i%2)*0.5 // Small variation
		chart.AddPricePoint(price, 10)
	}

	_, _, volatility := chart.GetIndicators()
	assert.Less(t, volatility, 5.0) // Should have low volatility

	// Test high volatility
	chart = NewPriceChart("apple", TimeframeHourly, ChartTypeLine)
	for i := 0; i < 20; i++ {
		price := 100.0 + float64(i%2)*20 // Large variation
		chart.AddPricePoint(price, 10)
	}

	_, _, volatility = chart.GetIndicators()
	assert.Greater(t, volatility, 10.0) // Should have high volatility
}

func TestPriceChart_TrendLines(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Create an upward trend
	for i := 0; i < 30; i++ {
		// Add some noise but overall upward trend
		price := 100.0 + float64(i)*2 + math.Sin(float64(i))*5
		chart.AddPricePoint(price, 10)
	}

	trends := chart.GetTrendLines()
	assert.Greater(t, len(trends), 0)

	// Should detect at least one trend line
	foundUpTrend := false
	for _, trend := range trends {
		if trend.Slope > 0 {
			foundUpTrend = true
			break
		}
	}
	assert.True(t, foundUpTrend, "Should detect upward trend")
}

func TestPriceChart_GetPricePrediction(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Not enough data
	prediction := chart.GetPricePrediction()
	assert.Nil(t, prediction)

	// Add upward trending data
	for i := 0; i < 20; i++ {
		price := 100.0 + float64(i)*2
		chart.AddPricePoint(price, 10)
	}

	prediction = chart.GetPricePrediction()
	assert.NotNil(t, prediction)
	assert.Equal(t, market.TrendUp, prediction.Direction)
	// Should predict higher price for upward trend
	assert.Greater(t, prediction.PredictedPrice, prediction.CurrentPrice*0.99) // Allow small margin
	assert.Greater(t, prediction.Confidence, 0.0)
	assert.LessOrEqual(t, prediction.Confidence, 1.0)
}

func TestPriceChart_Analyze(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Not enough data
	analysis := chart.Analyze()
	assert.Nil(t, analysis)

	// Add mixed data with patterns
	for i := 0; i < 30; i++ {
		// Create support at 90, resistance at 110
		price := 100.0 + math.Sin(float64(i)*0.5)*10
		chart.AddPricePoint(price, 10+i%5)
	}

	analysis = chart.Analyze()
	assert.NotNil(t, analysis)
	assert.Equal(t, "apple", analysis.ItemID)
	assert.Greater(t, analysis.VolumeAverage, 0)

	// Should have a recommendation
	assert.Contains(t, []market.TradeAction{
		market.ActionBuy,
		market.ActionSell,
		market.ActionHold,
	}, analysis.Recommendation)

	assert.Greater(t, analysis.ConfidenceLevel, 0.0)
	assert.LessOrEqual(t, analysis.ConfidenceLevel, 0.95)
}

func TestPriceChart_Recommendation(t *testing.T) {
	tests := []struct {
		name           string
		setupChart     func() *PriceChart
		expectedAction market.TradeAction
	}{
		{
			name: "Oversold condition should recommend buy",
			setupChart: func() *PriceChart {
				chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)
				// Create oversold condition (falling prices)
				for i := 0; i < 20; i++ {
					price := 150.0 - float64(i)*3
					chart.AddPricePoint(price, 10)
				}
				return chart
			},
			expectedAction: market.ActionBuy,
		},
		{
			name: "Overbought condition should recommend sell",
			setupChart: func() *PriceChart {
				chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)
				// Create overbought condition (rising prices)
				for i := 0; i < 20; i++ {
					price := 50.0 + float64(i)*3
					chart.AddPricePoint(price, 10)
				}
				return chart
			},
			expectedAction: market.ActionSell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart := tt.setupChart()
			analysis := chart.Analyze()

			assert.NotNil(t, analysis)
			if analysis.Recommendation != tt.expectedAction {
				t.Logf("RSI: %v, MA: %v, Volatility: %v", analysis.RSI, analysis.MovingAverage, analysis.Volatility)
			}
			assert.Equal(t, tt.expectedAction, analysis.Recommendation)
		})
	}
}

func TestPriceChart_TimeframeSeparation(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeDaily, ChartTypeCandle)

	// Add points on different days
	now := time.Now()

	// Day 1
	chart.dataPoints = append(chart.dataPoints, PricePoint{
		Timestamp: now,
		Price:     100,
		Volume:    10,
	})
	chart.updateCandleData(chart.dataPoints[0])

	// Day 2
	tomorrow := now.Add(25 * time.Hour)
	point2 := PricePoint{
		Timestamp: tomorrow,
		Price:     110,
		Volume:    15,
	}
	chart.dataPoints = append(chart.dataPoints, point2)
	chart.updateCandleData(point2)

	candles := chart.GetCandles()
	assert.Len(t, candles, 2) // Should have 2 separate candles for 2 days
}

func TestPriceChart_LocalExtrema(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Create clear peaks and valleys
	prices := []float64{
		100, 105, 110, 105, 100, // Peak at 110
		95, 90, 95, 100, 105, // Valley at 90
		110, 115, 110, 105, 100, // Peak at 115
		95, 90, 85, 90, 95, // Valley at 85
	}

	for _, price := range prices {
		chart.AddPricePoint(price, 10)
	}

	highs := chart.findLocalExtrema(true)
	lows := chart.findLocalExtrema(false)

	// Should find some peaks and valleys
	assert.Greater(t, len(highs), 0, "Should find high points")
	assert.Greater(t, len(lows), 0, "Should find low points")
}

func TestPriceChart_LinearRegression(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Create points with clear linear trend
	points := []PricePoint{}
	baseTime := time.Now()

	for i := 0; i < 10; i++ {
		points = append(points, PricePoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour),
			Price:     100.0 + float64(i)*5, // y = 100 + 5x
			Volume:    10,
		})
	}

	trend := chart.fitTrendLine(points, TrendTypeSupport)
	assert.NotNil(t, trend)

	// Slope should be approximately 5/3600 (5 per hour in seconds)
	expectedSlope := 5.0 / 3600.0
	assert.InDelta(t, expectedSlope, trend.Slope, 0.001)
}

func TestPriceChart_ConcurrentAccess(t *testing.T) {
	chart := NewPriceChart("apple", TimeframeHourly, ChartTypeLine)

	// Concurrent writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			chart.AddPricePoint(float64(100+i), i)
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Reader goroutines
	for i := 0; i < 3; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				_ = chart.GetDataPoints()
				_ = chart.GetCandles()
				_ = chart.GetTrendLines()
				_, _, _ = chart.GetIndicators()
				_ = chart.Analyze()
				time.Sleep(time.Microsecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Should have processed without panic
	points := chart.GetDataPoints()
	assert.Greater(t, len(points), 0)
}
