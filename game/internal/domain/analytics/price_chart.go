package analytics

import (
	"math"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

// ChartTimeframe represents the time period for the chart
type ChartTimeframe int

const (
	TimeframeHourly ChartTimeframe = iota
	TimeframeDaily
	TimeframeWeekly
	TimeframeMonthly
)

// ChartType represents the type of chart
type ChartType int

const (
	ChartTypeLine ChartType = iota
	ChartTypeCandle
	ChartTypeArea
	ChartTypeBar
)

// PricePoint represents a single price data point
type PricePoint struct {
	Timestamp time.Time
	Price     float64
	Volume    int // Number of transactions
}

// CandleData represents OHLC (Open, High, Low, Close) data
type CandleData struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int
}

// TrendLine represents a trend line on the chart
type TrendLine struct {
	StartPoint PricePoint
	EndPoint   PricePoint
	Slope      float64
	Intercept  float64
	Type       TrendType
}

// TrendType represents the type of trend
type TrendType int

const (
	TrendTypeSupport TrendType = iota
	TrendTypeResistance
	TrendTypeMovingAverage
)

// PriceChart manages price chart data and analysis
type PriceChart struct {
	itemID     string
	timeframe  ChartTimeframe
	chartType  ChartType
	dataPoints []PricePoint
	candles    []CandleData
	trendLines []TrendLine

	// Technical indicators
	movingAverage    []float64
	relativeStrength float64
	volatility       float64

	// Chart settings
	maxDataPoints  int
	updateInterval time.Duration

	mu sync.RWMutex
}

// NewPriceChart creates a new price chart
func NewPriceChart(itemID string, timeframe ChartTimeframe, chartType ChartType) *PriceChart {
	pc := &PriceChart{
		itemID:         itemID,
		timeframe:      timeframe,
		chartType:      chartType,
		dataPoints:     make([]PricePoint, 0),
		candles:        make([]CandleData, 0),
		trendLines:     make([]TrendLine, 0),
		maxDataPoints:  100,
		updateInterval: getUpdateInterval(timeframe),
	}

	return pc
}

// AddPricePoint adds a new price point to the chart
func (pc *PriceChart) AddPricePoint(price float64, volume int) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	point := PricePoint{
		Timestamp: time.Now(),
		Price:     price,
		Volume:    volume,
	}

	pc.dataPoints = append(pc.dataPoints, point)

	// Maintain max data points
	if len(pc.dataPoints) > pc.maxDataPoints {
		pc.dataPoints = pc.dataPoints[len(pc.dataPoints)-pc.maxDataPoints:]
	}

	// Update candle data if using candle chart
	if pc.chartType == ChartTypeCandle {
		pc.updateCandleData(point)
	}

	// Recalculate indicators
	pc.calculateIndicators()
}

// updateCandleData updates OHLC data
func (pc *PriceChart) updateCandleData(point PricePoint) {
	if len(pc.candles) == 0 {
		// Create first candle
		pc.candles = append(pc.candles, CandleData{
			Timestamp: point.Timestamp,
			Open:      point.Price,
			High:      point.Price,
			Low:       point.Price,
			Close:     point.Price,
			Volume:    point.Volume,
		})
		return
	}

	lastCandle := &pc.candles[len(pc.candles)-1]

	// Check if we need a new candle based on timeframe
	if pc.shouldCreateNewCandle(lastCandle.Timestamp, point.Timestamp) {
		// Create new candle
		pc.candles = append(pc.candles, CandleData{
			Timestamp: point.Timestamp,
			Open:      point.Price,
			High:      point.Price,
			Low:       point.Price,
			Close:     point.Price,
			Volume:    point.Volume,
		})
	} else {
		// Update existing candle
		lastCandle.High = math.Max(lastCandle.High, point.Price)
		lastCandle.Low = math.Min(lastCandle.Low, point.Price)
		lastCandle.Close = point.Price
		lastCandle.Volume += point.Volume
	}

	// Maintain max candles
	if len(pc.candles) > pc.maxDataPoints {
		pc.candles = pc.candles[len(pc.candles)-pc.maxDataPoints:]
	}
}

// shouldCreateNewCandle determines if a new candle should be created
func (pc *PriceChart) shouldCreateNewCandle(lastTime, currentTime time.Time) bool {
	switch pc.timeframe {
	case TimeframeHourly:
		return currentTime.Hour() != lastTime.Hour()
	case TimeframeDaily:
		return currentTime.Day() != lastTime.Day()
	case TimeframeWeekly:
		_, lastWeek := lastTime.ISOWeek()
		_, currentWeek := currentTime.ISOWeek()
		return currentWeek != lastWeek
	case TimeframeMonthly:
		return currentTime.Month() != lastTime.Month()
	default:
		return false
	}
}

// calculateIndicators calculates technical indicators
func (pc *PriceChart) calculateIndicators() {
	if len(pc.dataPoints) < 2 {
		return
	}

	// Calculate moving average
	pc.calculateMovingAverage(20) // 20-period MA

	// Calculate RSI
	pc.calculateRSI(14) // 14-period RSI

	// Calculate volatility
	pc.calculateVolatility()

	// Detect trend lines
	pc.detectTrendLines()
}

// calculateMovingAverage calculates the moving average
func (pc *PriceChart) calculateMovingAverage(period int) {
	if len(pc.dataPoints) < period {
		return
	}

	pc.movingAverage = make([]float64, 0)

	for i := period - 1; i < len(pc.dataPoints); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += pc.dataPoints[j].Price
		}
		pc.movingAverage = append(pc.movingAverage, sum/float64(period))
	}
}

// calculateRSI calculates the Relative Strength Index
func (pc *PriceChart) calculateRSI(period int) {
	if len(pc.dataPoints) < period+1 {
		return
	}

	gains := 0.0
	losses := 0.0

	// Calculate initial average gain/loss
	for i := 1; i <= period; i++ {
		change := pc.dataPoints[i].Price - pc.dataPoints[i-1].Price
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Calculate RSI
	if avgLoss == 0 {
		pc.relativeStrength = 100
	} else {
		rs := avgGain / avgLoss
		pc.relativeStrength = 100 - (100 / (1 + rs))
	}
}

// calculateVolatility calculates price volatility
func (pc *PriceChart) calculateVolatility() {
	if len(pc.dataPoints) < 2 {
		return
	}

	// Calculate standard deviation of price changes
	changes := make([]float64, len(pc.dataPoints)-1)
	sum := 0.0

	for i := 1; i < len(pc.dataPoints); i++ {
		change := (pc.dataPoints[i].Price - pc.dataPoints[i-1].Price) / pc.dataPoints[i-1].Price
		changes[i-1] = change
		sum += change
	}

	mean := sum / float64(len(changes))

	variance := 0.0
	for _, change := range changes {
		variance += math.Pow(change-mean, 2)
	}
	variance /= float64(len(changes))

	pc.volatility = math.Sqrt(variance) * 100 // As percentage
}

// detectTrendLines detects support and resistance levels
func (pc *PriceChart) detectTrendLines() {
	if len(pc.dataPoints) < 10 {
		return
	}

	pc.trendLines = make([]TrendLine, 0)

	// Find local maxima and minima
	highs := pc.findLocalExtrema(true)
	lows := pc.findLocalExtrema(false)

	// Create resistance lines from highs
	if len(highs) >= 2 {
		trend := pc.fitTrendLine(highs, TrendTypeResistance)
		if trend != nil {
			pc.trendLines = append(pc.trendLines, *trend)
		}
	}

	// Create support lines from lows
	if len(lows) >= 2 {
		trend := pc.fitTrendLine(lows, TrendTypeSupport)
		if trend != nil {
			pc.trendLines = append(pc.trendLines, *trend)
		}
	}

	// If no extrema found, fit trend line to all points
	if len(pc.trendLines) == 0 && len(pc.dataPoints) >= 10 {
		trend := pc.fitTrendLine(pc.dataPoints, TrendTypeSupport)
		if trend != nil {
			pc.trendLines = append(pc.trendLines, *trend)
		}
	}

	// Add moving average trend line
	if len(pc.movingAverage) > 0 {
		maTrend := pc.createMovingAverageTrend()
		if maTrend != nil {
			pc.trendLines = append(pc.trendLines, *maTrend)
		}
	}
}

// findLocalExtrema finds local highs or lows
func (pc *PriceChart) findLocalExtrema(findHighs bool) []PricePoint {
	extrema := make([]PricePoint, 0)
	lookback := 5 // Number of points to look back/forward

	for i := lookback; i < len(pc.dataPoints)-lookback; i++ {
		isExtremum := true
		currentPrice := pc.dataPoints[i].Price

		for j := i - lookback; j <= i+lookback; j++ {
			if j == i {
				continue
			}

			if findHighs {
				if pc.dataPoints[j].Price > currentPrice {
					isExtremum = false
					break
				}
			} else {
				if pc.dataPoints[j].Price < currentPrice {
					isExtremum = false
					break
				}
			}
		}

		if isExtremum {
			extrema = append(extrema, pc.dataPoints[i])
		}
	}

	return extrema
}

// fitTrendLine fits a trend line to points using linear regression
func (pc *PriceChart) fitTrendLine(points []PricePoint, trendType TrendType) *TrendLine {
	if len(points) < 2 {
		return nil
	}

	// Convert timestamps to numeric values (seconds since first point)
	x := make([]float64, len(points))
	y := make([]float64, len(points))

	startTime := points[0].Timestamp
	for i, point := range points {
		x[i] = point.Timestamp.Sub(startTime).Seconds()
		y[i] = point.Price
	}

	// Calculate linear regression
	n := float64(len(points))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i := 0; i < len(points); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	return &TrendLine{
		StartPoint: points[0],
		EndPoint:   points[len(points)-1],
		Slope:      slope,
		Intercept:  intercept,
		Type:       trendType,
	}
}

// createMovingAverageTrend creates a trend line from moving average
func (pc *PriceChart) createMovingAverageTrend() *TrendLine {
	if len(pc.movingAverage) < 2 {
		return nil
	}

	// Use first and last MA points
	startIdx := len(pc.dataPoints) - len(pc.movingAverage)

	return &TrendLine{
		StartPoint: PricePoint{
			Timestamp: pc.dataPoints[startIdx].Timestamp,
			Price:     pc.movingAverage[0],
		},
		EndPoint: PricePoint{
			Timestamp: pc.dataPoints[len(pc.dataPoints)-1].Timestamp,
			Price:     pc.movingAverage[len(pc.movingAverage)-1],
		},
		Type: TrendTypeMovingAverage,
	}
}

// GetDataPoints returns the chart data points
func (pc *PriceChart) GetDataPoints() []PricePoint {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	result := make([]PricePoint, len(pc.dataPoints))
	copy(result, pc.dataPoints)
	return result
}

// GetCandles returns the candle data
func (pc *PriceChart) GetCandles() []CandleData {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	result := make([]CandleData, len(pc.candles))
	copy(result, pc.candles)
	return result
}

// GetTrendLines returns the trend lines
func (pc *PriceChart) GetTrendLines() []TrendLine {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	result := make([]TrendLine, len(pc.trendLines))
	copy(result, pc.trendLines)
	return result
}

// GetIndicators returns the technical indicators
func (pc *PriceChart) GetIndicators() (movingAverage []float64, rsi float64, volatility float64) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	ma := make([]float64, len(pc.movingAverage))
	copy(ma, pc.movingAverage)

	return ma, pc.relativeStrength, pc.volatility
}

// GetPricePrediction returns a simple price prediction
func (pc *PriceChart) GetPricePrediction() *PricePrediction {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if len(pc.dataPoints) < 10 {
		return nil
	}

	// Simple prediction based on trend and momentum
	lastPrice := pc.dataPoints[len(pc.dataPoints)-1].Price
	var predictedPrice float64
	var confidence float64
	var direction market.PriceTrend

	// Calculate simple trend from recent price movements
	if len(pc.dataPoints) >= 2 {
		recentTrend := 0.0
		// Calculate average trend from last 5 points (or available)
		lookback := 5
		if len(pc.dataPoints) < lookback {
			lookback = len(pc.dataPoints)
		}

		for i := len(pc.dataPoints) - lookback + 1; i < len(pc.dataPoints); i++ {
			change := pc.dataPoints[i].Price - pc.dataPoints[i-1].Price
			recentTrend += change
		}
		recentTrend /= float64(lookback - 1)

		// Determine direction based on recent trend
		switch {
		case recentTrend > 0.5:
			direction = market.TrendUp
			predictedPrice = lastPrice * 1.02
		case recentTrend < -0.5:
			direction = market.TrendDown
			predictedPrice = lastPrice * 0.98
		default:
			direction = market.TrendStable
			predictedPrice = lastPrice
		}

		// Base confidence
		confidence = 0.5
	}

	// Check trend direction from trend lines if available
	if len(pc.trendLines) > 0 {
		mainTrend := pc.trendLines[0]
		switch {
		case mainTrend.Slope > 0:
			direction = market.TrendUp
			predictedPrice = lastPrice * (1 + mainTrend.Slope*0.01)
		case mainTrend.Slope < 0:
			direction = market.TrendDown
			predictedPrice = lastPrice * (1 + mainTrend.Slope*0.01)
		default:
			direction = market.TrendStable
			predictedPrice = lastPrice
		}

		// Calculate confidence based on volatility
		switch {
		case pc.volatility < 5:
			confidence = 0.8
		case pc.volatility < 10:
			confidence = 0.6
		case pc.volatility < 20:
			confidence = 0.4
		default:
			confidence = 0.2
		}
	}

	// Adjust based on RSI
	if pc.relativeStrength > 70 {
		// Overbought, likely to go down
		predictedPrice *= 0.98
		confidence *= 0.9
	} else if pc.relativeStrength < 30 {
		// Oversold, likely to go up
		predictedPrice *= 1.02
		confidence *= 0.9
	}

	return &PricePrediction{
		CurrentPrice:   lastPrice,
		PredictedPrice: predictedPrice,
		Direction:      direction,
		Confidence:     confidence,
		TimeHorizon:    pc.updateInterval,
	}
}

// PricePrediction represents a price prediction
type PricePrediction struct {
	CurrentPrice   float64
	PredictedPrice float64
	Direction      market.PriceTrend
	Confidence     float64 // 0-1 scale
	TimeHorizon    time.Duration
}

// ChartAnalysis provides analysis of the chart
type ChartAnalysis struct {
	ItemID          string
	TrendDirection  market.PriceTrend
	SupportLevel    float64
	ResistanceLevel float64
	MovingAverage   float64
	RSI             float64
	Volatility      float64
	VolumeAverage   int
	Recommendation  market.TradeAction
	ConfidenceLevel float64
}

// Analyze provides comprehensive chart analysis
func (pc *PriceChart) Analyze() *ChartAnalysis {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if len(pc.dataPoints) < 10 {
		return nil
	}

	analysis := &ChartAnalysis{
		ItemID:     pc.itemID,
		RSI:        pc.relativeStrength,
		Volatility: pc.volatility,
	}

	// Determine trend direction
	if len(pc.trendLines) > 0 {
		mainTrend := pc.trendLines[0]
		switch {
		case mainTrend.Slope > 0.01:
			analysis.TrendDirection = market.TrendUp
		case mainTrend.Slope < -0.01:
			analysis.TrendDirection = market.TrendDown
		default:
			analysis.TrendDirection = market.TrendStable
		}
	}

	// Find support and resistance
	for _, trend := range pc.trendLines {
		if trend.Type == TrendTypeSupport {
			analysis.SupportLevel = trend.EndPoint.Price
		} else if trend.Type == TrendTypeResistance {
			analysis.ResistanceLevel = trend.EndPoint.Price
		}
	}

	// Current moving average
	if len(pc.movingAverage) > 0 {
		analysis.MovingAverage = pc.movingAverage[len(pc.movingAverage)-1]
	}

	// Calculate average volume
	totalVolume := 0
	for _, point := range pc.dataPoints {
		totalVolume += point.Volume
	}
	analysis.VolumeAverage = totalVolume / len(pc.dataPoints)

	// Generate recommendation
	analysis.Recommendation, analysis.ConfidenceLevel = pc.generateRecommendation(analysis)

	return analysis
}

// generateRecommendation generates trading recommendation
func (pc *PriceChart) generateRecommendation(analysis *ChartAnalysis) (market.TradeAction, float64) {
	confidence := 0.5 // Base confidence
	action := market.ActionHold

	// Score-based system to avoid conflicting signals
	buyScore := 0.0
	sellScore := 0.0

	lastPrice := pc.dataPoints[len(pc.dataPoints)-1].Price

	// Check RSI (highest priority)
	if pc.relativeStrength > 70 {
		// Overbought
		sellScore += 0.4
	} else if pc.relativeStrength < 30 {
		// Oversold
		buyScore += 0.4
	}

	// Check price vs moving average
	if analysis.MovingAverage > 0 {
		if lastPrice < analysis.MovingAverage*0.95 {
			buyScore += 0.2
		} else if lastPrice > analysis.MovingAverage*1.05 {
			sellScore += 0.2
		}
	}

	// Check support/resistance
	if analysis.SupportLevel > 0 && lastPrice <= analysis.SupportLevel*1.02 {
		buyScore += 0.15
	} else if analysis.ResistanceLevel > 0 && lastPrice >= analysis.ResistanceLevel*0.98 {
		sellScore += 0.15
	}

	// Determine action based on scores
	if sellScore > buyScore && sellScore > 0.2 {
		action = market.ActionSell
		confidence += sellScore
	} else if buyScore > sellScore && buyScore > 0.2 {
		action = market.ActionBuy
		confidence += buyScore
	}

	// Adjust confidence based on volatility
	if pc.volatility > 20 {
		confidence *= 0.7 // Less confident in volatile markets
	} else if pc.volatility < 5 {
		confidence *= 1.2 // More confident in stable markets
	}

	// Cap confidence at 0.95
	if confidence > 0.95 {
		confidence = 0.95
	}

	return action, confidence
}

// Helper function to get update interval based on timeframe
func getUpdateInterval(timeframe ChartTimeframe) time.Duration {
	switch timeframe {
	case TimeframeHourly:
		return time.Hour
	case TimeframeDaily:
		return 24 * time.Hour
	case TimeframeWeekly:
		return 7 * 24 * time.Hour
	case TimeframeMonthly:
		return 30 * 24 * time.Hour
	default:
		return time.Hour
	}
}
