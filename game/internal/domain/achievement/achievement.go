package achievement

import (
	"fmt"
	"sync"
	"time"
)

// AchievementID represents unique achievement identifier
type AchievementID string

// Achievement IDs
const (
	// Trading Achievements
	AchievementFirstProfit   AchievementID = "first_profit"
	AchievementFirstLoss     AchievementID = "first_loss"
	AchievementProfitStreak5 AchievementID = "profit_streak_5"
	AchievementMasterTrader  AchievementID = "master_trader"

	// Investment Achievements
	AchievementROI50       AchievementID = "roi_50"
	AchievementROI100      AchievementID = "roi_100"
	AchievementDiversified AchievementID = "diversified_portfolio"
	AchievementRiskManager AchievementID = "risk_manager"

	// Wealth Achievements
	AchievementGold1000    AchievementID = "gold_1000"
	AchievementGold10000   AchievementID = "gold_10000"
	AchievementGold100000  AchievementID = "gold_100000"
	AchievementMillionaire AchievementID = "millionaire"

	// Shop Achievements
	AchievementShopLevel3    AchievementID = "shop_level_3"
	AchievementShopLevel5    AchievementID = "shop_level_5"
	AchievementFullyEquipped AchievementID = "fully_equipped"
	AchievementEfficiencyMax AchievementID = "efficiency_max"

	// Market Achievements
	AchievementBuyLowSellHigh   AchievementID = "buy_low_sell_high"
	AchievementMarketTiming     AchievementID = "market_timing"
	AchievementVolatilityMaster AchievementID = "volatility_master"
	AchievementSeasonalExpert   AchievementID = "seasonal_expert"

	// Progress Achievements
	AchievementDay30            AchievementID = "survived_30_days"
	AchievementDay100           AchievementID = "survived_100_days"
	AchievementYear1            AchievementID = "survived_1_year"
	AchievementCompleteTutorial AchievementID = "complete_tutorial"
)

// AchievementTier represents achievement rarity
type AchievementTier int

const (
	TierBronze AchievementTier = iota
	TierSilver
	TierGold
	TierPlatinum
)

// Achievement represents a single achievement
type Achievement struct {
	ID          AchievementID
	Name        string
	Description string
	Tier        AchievementTier
	Points      int
	Hidden      bool   // Hidden until unlocked
	Icon        string // Icon path for UI
	UnlockedAt  *time.Time
	Progress    float64 // 0.0 to 1.0 for progressive achievements
	MaxProgress float64 // Maximum value for progress
}

// AchievementManager manages all achievements
type AchievementManager struct {
	achievements map[AchievementID]*Achievement
	unlocked     map[AchievementID]bool
	statistics   *PlayerStatistics
	callbacks    []AchievementCallback
	mu           sync.RWMutex
}

// PlayerStatistics tracks stats for achievement progress
type PlayerStatistics struct {
	TotalProfit        int
	TotalLoss          int
	ConsecutiveProfits int
	ConsecutiveLosses  int
	HighestGold        int
	TotalTrades        int
	SuccessfulTrades   int
	ItemTypesSold      map[string]int
	BestROI            float64
	CurrentDay         int
	TutorialCompleted  bool
	ShopLevel          int
	EquipmentCount     int
}

// AchievementCallback is called when achievement is unlocked
type AchievementCallback func(achievement *Achievement)

// NewAchievementManager creates a new achievement manager
func NewAchievementManager() *AchievementManager {
	am := &AchievementManager{
		achievements: make(map[AchievementID]*Achievement),
		unlocked:     make(map[AchievementID]bool),
		statistics: &PlayerStatistics{
			ItemTypesSold: make(map[string]int),
		},
		callbacks: make([]AchievementCallback, 0),
	}

	am.initializeAchievements()
	return am
}

// initializeAchievements sets up all achievements
func (am *AchievementManager) initializeAchievements() {
	// Trading Achievements
	am.registerAchievement(&Achievement{
		ID:          AchievementFirstProfit,
		Name:        "First Profit",
		Description: "Make your first profitable trade",
		Tier:        TierBronze,
		Points:      10,
		Icon:        "icons/profit.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementProfitStreak5,
		Name:        "Profit Streak",
		Description: "Make 5 profitable trades in a row",
		Tier:        TierSilver,
		Points:      25,
		Icon:        "icons/streak.png",
		MaxProgress: 5,
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementMasterTrader,
		Name:        "Master Trader",
		Description: "Complete 100 successful trades",
		Tier:        TierGold,
		Points:      50,
		Icon:        "icons/master.png",
		MaxProgress: 100,
	})

	// Investment Achievements
	am.registerAchievement(&Achievement{
		ID:          AchievementROI50,
		Name:        "Smart Investor",
		Description: "Achieve 50% return on investment",
		Tier:        TierBronze,
		Points:      15,
		Icon:        "icons/roi.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementROI100,
		Name:        "Investment Genius",
		Description: "Double your investment (100% ROI)",
		Tier:        TierSilver,
		Points:      30,
		Icon:        "icons/roi_gold.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementDiversified,
		Name:        "Diversified Portfolio",
		Description: "Trade in 5 different item categories",
		Tier:        TierSilver,
		Points:      20,
		Icon:        "icons/diverse.png",
		MaxProgress: 5,
	})

	// Wealth Achievements
	am.registerAchievement(&Achievement{
		ID:          AchievementGold1000,
		Name:        "Thousand Gold",
		Description: "Accumulate 1,000 gold",
		Tier:        TierBronze,
		Points:      10,
		Icon:        "icons/gold_1k.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementGold10000,
		Name:        "Ten Thousand Gold",
		Description: "Accumulate 10,000 gold",
		Tier:        TierSilver,
		Points:      25,
		Icon:        "icons/gold_10k.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementMillionaire,
		Name:        "Millionaire",
		Description: "Accumulate 1,000,000 gold",
		Tier:        TierPlatinum,
		Points:      100,
		Icon:        "icons/millionaire.png",
		Hidden:      true,
	})

	// Shop Achievements
	am.registerAchievement(&Achievement{
		ID:          AchievementShopLevel3,
		Name:        "Growing Business",
		Description: "Upgrade your shop to level 3",
		Tier:        TierBronze,
		Points:      15,
		Icon:        "icons/shop.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementFullyEquipped,
		Name:        "Fully Equipped",
		Description: "Purchase all available equipment",
		Tier:        TierGold,
		Points:      40,
		Icon:        "icons/equipment.png",
	})

	// Market Achievements
	am.registerAchievement(&Achievement{
		ID:          AchievementBuyLowSellHigh,
		Name:        "Buy Low, Sell High",
		Description: "Sell an item for 50% more than purchase price",
		Tier:        TierBronze,
		Points:      15,
		Icon:        "icons/market.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementMarketTiming,
		Name:        "Perfect Timing",
		Description: "Sell during a high demand event",
		Tier:        TierSilver,
		Points:      20,
		Icon:        "icons/timing.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementVolatilityMaster,
		Name:        "Volatility Master",
		Description: "Profit from highly volatile items 10 times",
		Tier:        TierGold,
		Points:      35,
		Icon:        "icons/volatile.png",
		MaxProgress: 10,
	})

	// Progress Achievements
	am.registerAchievement(&Achievement{
		ID:          AchievementDay30,
		Name:        "Month Survivor",
		Description: "Survive 30 days",
		Tier:        TierBronze,
		Points:      10,
		Icon:        "icons/calendar.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementYear1,
		Name:        "Yearly Veteran",
		Description: "Survive 365 days",
		Tier:        TierPlatinum,
		Points:      75,
		Icon:        "icons/year.png",
	})

	am.registerAchievement(&Achievement{
		ID:          AchievementCompleteTutorial,
		Name:        "Ready to Trade",
		Description: "Complete the tutorial",
		Tier:        TierBronze,
		Points:      5,
		Icon:        "icons/tutorial.png",
	})
}

// registerAchievement adds an achievement to the manager
func (am *AchievementManager) registerAchievement(achievement *Achievement) {
	am.achievements[achievement.ID] = achievement
}

// UpdateTradeStats updates statistics after a trade
func (am *AchievementManager) UpdateTradeStats(profit int, roi float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.statistics.TotalTrades++

	if profit > 0 {
		am.statistics.TotalProfit += profit
		am.statistics.ConsecutiveProfits++
		am.statistics.ConsecutiveLosses = 0
		am.statistics.SuccessfulTrades++

		// Check first profit
		if am.statistics.SuccessfulTrades == 1 {
			am.unlockAchievement(AchievementFirstProfit)
		}

		// Check profit streaks
		if am.statistics.ConsecutiveProfits >= 5 {
			am.updateProgress(AchievementProfitStreak5, float64(am.statistics.ConsecutiveProfits))
		}
	} else if profit < 0 {
		am.statistics.TotalLoss += -profit
		am.statistics.ConsecutiveLosses++
		am.statistics.ConsecutiveProfits = 0

		// Check first loss
		if am.statistics.TotalLoss > 0 && !am.unlocked[AchievementFirstLoss] {
			am.unlockAchievement(AchievementFirstLoss)
		}
	}

	// Update ROI stats
	if roi > am.statistics.BestROI {
		am.statistics.BestROI = roi

		if roi >= 0.5 {
			am.unlockAchievement(AchievementROI50)
		}
		if roi >= 1.0 {
			am.unlockAchievement(AchievementROI100)
		}
	}

	// Check master trader progress
	am.updateProgress(AchievementMasterTrader, float64(am.statistics.SuccessfulTrades))
}

// UpdateGoldStats updates gold-related statistics
func (am *AchievementManager) UpdateGoldStats(currentGold int) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if currentGold > am.statistics.HighestGold {
		am.statistics.HighestGold = currentGold

		if currentGold >= 1000 {
			am.unlockAchievement(AchievementGold1000)
		}
		if currentGold >= 10000 {
			am.unlockAchievement(AchievementGold10000)
		}
		if currentGold >= 100000 {
			am.unlockAchievement(AchievementGold100000)
		}
		if currentGold >= 1000000 {
			am.unlockAchievement(AchievementMillionaire)
		}
	}
}

// UpdateDayStats updates day-related statistics
func (am *AchievementManager) UpdateDayStats(currentDay int) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.statistics.CurrentDay = currentDay

	if currentDay >= 30 {
		am.unlockAchievement(AchievementDay30)
	}
	if currentDay >= 100 {
		am.unlockAchievement(AchievementDay100)
	}
	if currentDay >= 365 {
		am.unlockAchievement(AchievementYear1)
	}
}

// UpdateShopStats updates shop-related statistics
func (am *AchievementManager) UpdateShopStats(shopLevel int, equipmentCount int) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.statistics.ShopLevel = shopLevel
	am.statistics.EquipmentCount = equipmentCount

	if shopLevel >= 3 {
		am.unlockAchievement(AchievementShopLevel3)
	}
	if shopLevel >= 5 {
		am.unlockAchievement(AchievementShopLevel5)
	}
	if equipmentCount >= 5 { // Assuming 5 total equipment available
		am.unlockAchievement(AchievementFullyEquipped)
	}
}

// UpdateItemCategoryStats updates item diversity statistics
func (am *AchievementManager) UpdateItemCategoryStats(category string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.statistics.ItemTypesSold == nil {
		am.statistics.ItemTypesSold = make(map[string]int)
	}

	am.statistics.ItemTypesSold[category]++

	// Check diversification
	if len(am.statistics.ItemTypesSold) >= 5 {
		am.updateProgress(AchievementDiversified, float64(len(am.statistics.ItemTypesSold)))
	}
}

// CompleteTutorial marks tutorial as completed
func (am *AchievementManager) CompleteTutorial() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.statistics.TutorialCompleted = true
	am.unlockAchievement(AchievementCompleteTutorial)
}

// unlockAchievement unlocks an achievement
func (am *AchievementManager) unlockAchievement(id AchievementID) {
	if am.unlocked[id] {
		return
	}

	achievement, exists := am.achievements[id]
	if !exists {
		return
	}

	now := time.Now()
	achievement.UnlockedAt = &now
	achievement.Progress = achievement.MaxProgress
	if achievement.Progress == 0 {
		achievement.Progress = 1.0
	}
	am.unlocked[id] = true

	// Notify callbacks
	for _, callback := range am.callbacks {
		callback(achievement)
	}
}

// updateProgress updates progress for progressive achievements
func (am *AchievementManager) updateProgress(id AchievementID, progress float64) {
	achievement, exists := am.achievements[id]
	if !exists || am.unlocked[id] {
		return
	}

	achievement.Progress = progress

	if achievement.MaxProgress > 0 && progress >= achievement.MaxProgress {
		am.unlockAchievement(id)
	}
}

// GetAchievement returns a specific achievement
func (am *AchievementManager) GetAchievement(id AchievementID) (*Achievement, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	achievement, exists := am.achievements[id]
	return achievement, exists
}

// GetAllAchievements returns all achievements
func (am *AchievementManager) GetAllAchievements() []*Achievement {
	am.mu.RLock()
	defer am.mu.RUnlock()

	achievements := make([]*Achievement, 0, len(am.achievements))
	for _, achievement := range am.achievements {
		// Don't return hidden achievements unless unlocked
		if !achievement.Hidden || am.unlocked[achievement.ID] {
			achievements = append(achievements, achievement)
		}
	}
	return achievements
}

// GetUnlockedAchievements returns only unlocked achievements
func (am *AchievementManager) GetUnlockedAchievements() []*Achievement {
	am.mu.RLock()
	defer am.mu.RUnlock()

	achievements := make([]*Achievement, 0)
	for id, achievement := range am.achievements {
		if am.unlocked[id] {
			achievements = append(achievements, achievement)
		}
	}
	return achievements
}

// GetProgress returns overall achievement progress
func (am *AchievementManager) GetProgress() (unlocked int, total int, points int) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	total = len(am.achievements)
	for id, achievement := range am.achievements {
		if am.unlocked[id] {
			unlocked++
			points += achievement.Points
		}
	}
	return
}

// RegisterCallback registers a callback for achievement unlocks
func (am *AchievementManager) RegisterCallback(callback AchievementCallback) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.callbacks = append(am.callbacks, callback)
}

// GetStatistics returns current player statistics
func (am *AchievementManager) GetStatistics() *PlayerStatistics {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.statistics
}

// Reset resets all achievements and statistics
func (am *AchievementManager) Reset() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.unlocked = make(map[AchievementID]bool)
	am.statistics = &PlayerStatistics{
		ItemTypesSold: make(map[string]int),
	}

	// Reset all achievement progress
	for _, achievement := range am.achievements {
		achievement.UnlockedAt = nil
		achievement.Progress = 0
	}
}

// ExportForSteam exports achievement data for Steam integration
func (am *AchievementManager) ExportForSteam() map[string]bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	steamAchievements := make(map[string]bool)
	for id := range am.unlocked {
		// Convert to Steam achievement ID format
		steamID := fmt.Sprintf("ACH_%s", string(id))
		steamAchievements[steamID] = true
	}
	return steamAchievements
}

// ImportFromSave imports achievement data from save file
func (am *AchievementManager) ImportFromSave(unlockedIDs []string, stats *PlayerStatistics) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Reset first
	am.unlocked = make(map[AchievementID]bool)

	// Import unlocked achievements
	for _, idStr := range unlockedIDs {
		id := AchievementID(idStr)
		if achievement, exists := am.achievements[id]; exists {
			am.unlocked[id] = true
			now := time.Now()
			achievement.UnlockedAt = &now
		}
	}

	// Import statistics
	if stats != nil {
		am.statistics = stats
		if am.statistics.ItemTypesSold == nil {
			am.statistics.ItemTypesSold = make(map[string]int)
		}
	}
}
