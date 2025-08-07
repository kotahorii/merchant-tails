package progression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRankSystem(t *testing.T) {
	t.Run("create new rank system", func(t *testing.T) {
		rankSystem := NewRankSystem()
		assert.NotNil(t, rankSystem)
		assert.Equal(t, RankApprentice, rankSystem.GetCurrentRank())
		assert.Equal(t, 0, rankSystem.GetExperience())
	})

	t.Run("add experience", func(t *testing.T) {
		rankSystem := NewRankSystem()

		// Add experience
		rankSystem.AddExperience(50)
		assert.Equal(t, 50, rankSystem.GetExperience())

		// Add more experience
		rankSystem.AddExperience(100)
		assert.Equal(t, 150, rankSystem.GetExperience())
	})

	t.Run("rank up from apprentice to journeyman", func(t *testing.T) {
		rankSystem := NewRankSystem()

		// Add enough experience to rank up
		promoted := rankSystem.AddExperience(1000)

		assert.True(t, promoted)
		assert.Equal(t, RankJourneyman, rankSystem.GetCurrentRank())
		assert.Equal(t, 0, rankSystem.GetExperience()) // Experience resets after promotion
	})

	t.Run("rank progression", func(t *testing.T) {
		rankSystem := NewRankSystem()

		// Apprentice -> Journeyman
		rankSystem.AddExperience(1000)
		assert.Equal(t, RankJourneyman, rankSystem.GetCurrentRank())

		// Journeyman -> Expert
		rankSystem.AddExperience(2500)
		assert.Equal(t, RankExpert, rankSystem.GetCurrentRank())

		// Expert -> Master
		rankSystem.AddExperience(5000)
		assert.Equal(t, RankMaster, rankSystem.GetCurrentRank())

		// Can't go beyond Master
		promoted := rankSystem.AddExperience(10000)
		assert.False(t, promoted)
		assert.Equal(t, RankMaster, rankSystem.GetCurrentRank())
	})

	t.Run("get rank benefits", func(t *testing.T) {
		rankSystem := NewRankSystem()

		// Apprentice benefits
		benefits := rankSystem.GetRankBenefits()
		assert.Equal(t, 50, benefits.ShopCapacity)
		assert.Equal(t, 100, benefits.WarehouseCapacity)
		assert.Equal(t, 1.0, benefits.PriceModifier)

		// Rank up to Journeyman
		rankSystem.AddExperience(1000)
		benefits = rankSystem.GetRankBenefits()
		assert.Equal(t, 75, benefits.ShopCapacity)
		assert.Equal(t, 150, benefits.WarehouseCapacity)
		assert.Equal(t, 0.95, benefits.PriceModifier) // 5% discount
	})

	t.Run("experience required for next rank", func(t *testing.T) {
		rankSystem := NewRankSystem()

		// Apprentice needs 1000 exp
		assert.Equal(t, 1000, rankSystem.GetExperienceRequired())

		// Add some experience
		rankSystem.AddExperience(300)
		assert.Equal(t, 700, rankSystem.GetExperienceToNextRank())

		// Rank up to Journeyman
		rankSystem.AddExperience(700)
		assert.Equal(t, 2500, rankSystem.GetExperienceRequired())
	})
}

func TestAchievementSystem(t *testing.T) {
	t.Run("create achievement system", func(t *testing.T) {
		achievementSystem := NewAchievementSystem()
		assert.NotNil(t, achievementSystem)

		// Register achievements
		achievementSystem.RegisterAchievement(&Achievement{
			ID:          "first_sale",
			Name:        "First Sale",
			Description: "Make your first sale",
			Points:      10,
			Category:    AchievementCategoryTrading,
		})

		achievements := achievementSystem.GetAllAchievements()
		assert.Equal(t, 1, len(achievements))
	})

	t.Run("unlock achievement", func(t *testing.T) {
		achievementSystem := NewAchievementSystem()

		// Register achievement
		achievementSystem.RegisterAchievement(&Achievement{
			ID:          "gold_hoarder",
			Name:        "Gold Hoarder",
			Description: "Accumulate 10,000 gold",
			Points:      50,
			Category:    AchievementCategoryWealth,
		})

		// Try to unlock
		unlocked := achievementSystem.UnlockAchievement("gold_hoarder")
		assert.True(t, unlocked)

		// Check if unlocked
		assert.True(t, achievementSystem.IsUnlocked("gold_hoarder"))

		// Can't unlock twice
		unlocked = achievementSystem.UnlockAchievement("gold_hoarder")
		assert.False(t, unlocked)
	})

	t.Run("achievement progress tracking", func(t *testing.T) {
		achievementSystem := NewAchievementSystem()

		// Register progressive achievement
		achievementSystem.RegisterAchievement(&Achievement{
			ID:            "trade_master",
			Name:          "Trade Master",
			Description:   "Complete 100 trades",
			Points:        100,
			Category:      AchievementCategoryTrading,
			MaxProgress:   100,
			IsProgressive: true,
		})

		// Update progress
		achievementSystem.UpdateProgress("trade_master", 25)
		progress := achievementSystem.GetProgress("trade_master")
		assert.Equal(t, 25, progress)

		// Update more progress
		achievementSystem.UpdateProgress("trade_master", 30)
		progress = achievementSystem.GetProgress("trade_master")
		assert.Equal(t, 55, progress)

		// Complete the achievement
		achievementSystem.UpdateProgress("trade_master", 45)
		assert.True(t, achievementSystem.IsUnlocked("trade_master"))
	})

	t.Run("achievement categories", func(t *testing.T) {
		achievementSystem := NewAchievementSystem()

		// Register achievements in different categories
		achievementSystem.RegisterAchievement(&Achievement{
			ID:       "trader_1",
			Category: AchievementCategoryTrading,
		})
		achievementSystem.RegisterAchievement(&Achievement{
			ID:       "trader_2",
			Category: AchievementCategoryTrading,
		})
		achievementSystem.RegisterAchievement(&Achievement{
			ID:       "wealth_1",
			Category: AchievementCategoryWealth,
		})

		// Get by category
		tradingAchievements := achievementSystem.GetAchievementsByCategory(AchievementCategoryTrading)
		assert.Equal(t, 2, len(tradingAchievements))

		wealthAchievements := achievementSystem.GetAchievementsByCategory(AchievementCategoryWealth)
		assert.Equal(t, 1, len(wealthAchievements))
	})

	t.Run("achievement points", func(t *testing.T) {
		achievementSystem := NewAchievementSystem()

		// Register achievements
		achievementSystem.RegisterAchievement(&Achievement{
			ID:     "achievement_1",
			Points: 10,
		})
		achievementSystem.RegisterAchievement(&Achievement{
			ID:     "achievement_2",
			Points: 25,
		})
		achievementSystem.RegisterAchievement(&Achievement{
			ID:     "achievement_3",
			Points: 50,
		})

		// Unlock some achievements
		achievementSystem.UnlockAchievement("achievement_1")
		achievementSystem.UnlockAchievement("achievement_3")

		// Check total points
		totalPoints := achievementSystem.GetTotalPoints()
		assert.Equal(t, 60, totalPoints) // 10 + 50
	})
}

func TestFeatureUnlockSystem(t *testing.T) {
	t.Run("create feature unlock system", func(t *testing.T) {
		unlockSystem := NewFeatureUnlockSystem()
		assert.NotNil(t, unlockSystem)

		// Register features
		unlockSystem.RegisterFeature(&Feature{
			ID:           "market_trading",
			Name:         "Market Trading",
			Description:  "Trade directly on the market",
			RequiredRank: RankJourneyman,
		})

		features := unlockSystem.GetAllFeatures()
		assert.Equal(t, 1, len(features))
	})

	t.Run("check feature availability", func(t *testing.T) {
		unlockSystem := NewFeatureUnlockSystem()

		// Register features with different requirements
		unlockSystem.RegisterFeature(&Feature{
			ID:           "basic_trading",
			Name:         "Basic Trading",
			RequiredRank: RankApprentice,
		})

		unlockSystem.RegisterFeature(&Feature{
			ID:           "advanced_trading",
			Name:         "Advanced Trading",
			RequiredRank: RankExpert,
		})

		// Check with Apprentice rank
		available := unlockSystem.IsAvailable("basic_trading", RankApprentice)
		assert.True(t, available)

		available = unlockSystem.IsAvailable("advanced_trading", RankApprentice)
		assert.False(t, available)

		// Check with Expert rank
		available = unlockSystem.IsAvailable("advanced_trading", RankExpert)
		assert.True(t, available)
	})

	t.Run("unlock features with achievements", func(t *testing.T) {
		unlockSystem := NewFeatureUnlockSystem()

		// Register feature that requires achievements
		unlockSystem.RegisterFeature(&Feature{
			ID:                   "special_vendor",
			Name:                 "Special Vendor Access",
			RequiredRank:         RankJourneyman,
			RequiredAchievements: []string{"gold_hoarder", "trade_master"},
		})

		// Check without achievements
		unlockedAchievements := []string{"gold_hoarder"}
		available := unlockSystem.IsAvailableWithAchievements(
			"special_vendor",
			RankJourneyman,
			unlockedAchievements,
		)
		assert.False(t, available)

		// Check with all achievements
		unlockedAchievements = []string{"gold_hoarder", "trade_master", "other_achievement"}
		available = unlockSystem.IsAvailableWithAchievements(
			"special_vendor",
			RankJourneyman,
			unlockedAchievements,
		)
		assert.True(t, available)
	})

	t.Run("progressive unlocks", func(t *testing.T) {
		unlockSystem := NewFeatureUnlockSystem()

		// Register tiered features
		unlockSystem.RegisterFeature(&Feature{
			ID:           "shop_upgrade_1",
			Name:         "Shop Upgrade Tier 1",
			RequiredRank: RankApprentice,
		})

		unlockSystem.RegisterFeature(&Feature{
			ID:               "shop_upgrade_2",
			Name:             "Shop Upgrade Tier 2",
			RequiredRank:     RankJourneyman,
			RequiredFeatures: []string{"shop_upgrade_1"},
		})

		unlockSystem.RegisterFeature(&Feature{
			ID:               "shop_upgrade_3",
			Name:             "Shop Upgrade Tier 3",
			RequiredRank:     RankExpert,
			RequiredFeatures: []string{"shop_upgrade_2"},
		})

		// Check chain availability
		unlockedFeatures := []string{"shop_upgrade_1"}
		available := unlockSystem.IsAvailableWithFeatures(
			"shop_upgrade_2",
			RankJourneyman,
			unlockedFeatures,
		)
		assert.True(t, available)

		// Can't skip tiers
		available = unlockSystem.IsAvailableWithFeatures(
			"shop_upgrade_3",
			RankExpert,
			unlockedFeatures, // Only tier 1 unlocked
		)
		assert.False(t, available)
	})
}

func TestProgressionManager(t *testing.T) {
	t.Run("create progression manager", func(t *testing.T) {
		manager := NewProgressionManager()
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.GetRankSystem())
		assert.NotNil(t, manager.GetAchievementSystem())
		assert.NotNil(t, manager.GetFeatureUnlockSystem())
	})

	t.Run("handle trade completion", func(t *testing.T) {
		manager := NewProgressionManager()

		// Register trade achievements
		manager.GetAchievementSystem().RegisterAchievement(&Achievement{
			ID:            "first_trade",
			Name:          "First Trade",
			Points:        10,
			Category:      AchievementCategoryTrading,
			IsProgressive: false,
		})

		manager.GetAchievementSystem().RegisterAchievement(&Achievement{
			ID:            "trade_10",
			Name:          "10 Trades",
			Points:        25,
			Category:      AchievementCategoryTrading,
			MaxProgress:   10,
			IsProgressive: true,
		})

		// Complete a trade
		result := manager.HandleTradeCompletion(100, 150) // Bought for 100, sold for 150

		assert.Equal(t, 10, result.ExperienceGained) // Base experience for trade
		assert.Equal(t, 5, result.BonusExperience)   // Bonus for profit

		// Check if first trade achievement was unlocked
		assert.True(t, manager.GetAchievementSystem().IsUnlocked("first_trade"))
	})

	t.Run("check feature availability with full context", func(t *testing.T) {
		manager := NewProgressionManager()

		// Register a complex feature
		manager.GetFeatureUnlockSystem().RegisterFeature(&Feature{
			ID:                   "elite_trading",
			Name:                 "Elite Trading",
			RequiredRank:         RankExpert,
			RequiredAchievements: []string{"trade_master"},
			RequiredFeatures:     []string{"advanced_trading"},
		})

		// Register prerequisite feature
		manager.GetFeatureUnlockSystem().RegisterFeature(&Feature{
			ID:           "advanced_trading",
			Name:         "Advanced Trading",
			RequiredRank: RankJourneyman,
		})

		// Register and unlock achievement
		manager.GetAchievementSystem().RegisterAchievement(&Achievement{
			ID: "trade_master",
		})
		manager.GetAchievementSystem().UnlockAchievement("trade_master")

		// Unlock prerequisite feature (simulate)
		manager.UnlockFeature("advanced_trading")

		// Set rank to Expert
		manager.GetRankSystem().SetRank(RankExpert)

		// Check availability
		available := manager.IsFeatureAvailable("elite_trading")
		assert.True(t, available)
	})

	t.Run("milestone rewards", func(t *testing.T) {
		manager := NewProgressionManager()

		// Register milestone achievements
		manager.RegisterMilestone(&Milestone{
			ID:        "gold_1000",
			Name:      "1000 Gold Milestone",
			Threshold: 1000,
			Type:      MilestoneTypeGold,
			Reward:    50, // 50 experience
		})

		// Check milestone
		rewards := manager.CheckMilestone(MilestoneTypeGold, 1000)
		assert.NotNil(t, rewards)
		assert.Equal(t, 50, rewards.Experience)

		// Milestone should only trigger once
		rewards = manager.CheckMilestone(MilestoneTypeGold, 1000)
		assert.Nil(t, rewards)
	})
}

func TestPlayerStats(t *testing.T) {
	t.Run("create player stats", func(t *testing.T) {
		stats := NewPlayerStats()
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats.GetTotalTrades())
		assert.Equal(t, 0, stats.GetTotalProfit())
	})

	t.Run("track trades", func(t *testing.T) {
		stats := NewPlayerStats()

		// Record trades
		stats.RecordTrade(100, 150) // 50 profit
		stats.RecordTrade(200, 180) // -20 loss
		stats.RecordTrade(50, 75)   // 25 profit

		assert.Equal(t, 3, stats.GetTotalTrades())
		assert.Equal(t, 55, stats.GetTotalProfit()) // 50 - 20 + 25
		assert.Equal(t, 2, stats.GetProfitableTrades())
		assert.InDelta(t, 0.667, stats.GetSuccessRate(), 0.001)
	})

	t.Run("track best trade", func(t *testing.T) {
		stats := NewPlayerStats()

		stats.RecordTrade(100, 200) // 100 profit
		stats.RecordTrade(50, 150)  // 100 profit
		stats.RecordTrade(200, 400) // 200 profit - best
		stats.RecordTrade(100, 150) // 50 profit

		assert.Equal(t, 200, stats.GetBestTrade())
		assert.Equal(t, 0, stats.GetWorstTrade()) // No losses

		stats.RecordTrade(100, 50) // -50 loss
		assert.Equal(t, -50, stats.GetWorstTrade())
	})

	t.Run("track item statistics", func(t *testing.T) {
		stats := NewPlayerStats()

		// Record item trades
		stats.RecordItemTrade("sword", 100, 150)
		stats.RecordItemTrade("sword", 100, 140)
		stats.RecordItemTrade("potion", 20, 30)
		stats.RecordItemTrade("potion", 20, 25)
		stats.RecordItemTrade("potion", 20, 15) // Loss

		// Get item stats
		swordStats := stats.GetItemStats("sword")
		assert.Equal(t, 2, swordStats.TradeCount)
		assert.Equal(t, 90, swordStats.TotalProfit) // 50 + 40

		potionStats := stats.GetItemStats("potion")
		assert.Equal(t, 3, potionStats.TradeCount)
		assert.Equal(t, 15, potionStats.TotalProfit) // 10 + 5 = 15 (loss is tracked separately)
	})
}

func TestProgressionIntegration(t *testing.T) {
	t.Run("full progression flow", func(t *testing.T) {
		manager := NewProgressionManager()

		// Setup achievements and features
		setupTestProgressionData(manager)

		// Simulate player progression

		// Start as Apprentice
		assert.Equal(t, RankApprentice, manager.GetRankSystem().GetCurrentRank())

		// Complete trades to gain experience
		// Each trade gives 10 base + 5 bonus (minimum) = 15 XP
		// Need 1000 XP to rank up from Apprentice
		// So we need at least 67 trades
		for i := 0; i < 70; i++ {
			result := manager.HandleTradeCompletion(100, 120)
			assert.NotNil(t, result)
		}

		// Should have gained enough experience to rank up
		currentRank := manager.GetRankSystem().GetCurrentRank()
		assert.True(t, currentRank > RankApprentice)

		// Check unlocked features
		available := manager.IsFeatureAvailable("basic_trading")
		assert.True(t, available)

		// Check achievements
		achievements := manager.GetAchievementSystem().GetUnlockedAchievements()
		assert.Greater(t, len(achievements), 0)

		// Check total progression score
		score := manager.GetProgressionScore()
		assert.Greater(t, score, 0)
	})
}

// Helper function to setup test data
func setupTestProgressionData(manager *ProgressionManager) {
	// Register achievements
	manager.GetAchievementSystem().RegisterAchievement(&Achievement{
		ID:       "first_trade",
		Name:     "First Trade",
		Points:   10,
		Category: AchievementCategoryTrading,
	})

	manager.GetAchievementSystem().RegisterAchievement(&Achievement{
		ID:            "trades_10",
		Name:          "10 Trades",
		Points:        25,
		Category:      AchievementCategoryTrading,
		MaxProgress:   10,
		IsProgressive: true,
	})

	// Register features
	manager.GetFeatureUnlockSystem().RegisterFeature(&Feature{
		ID:           "basic_trading",
		Name:         "Basic Trading",
		RequiredRank: RankApprentice,
	})

	manager.GetFeatureUnlockSystem().RegisterFeature(&Feature{
		ID:           "market_access",
		Name:         "Market Access",
		RequiredRank: RankJourneyman,
	})

	// Register milestones
	manager.RegisterMilestone(&Milestone{
		ID:        "profit_100",
		Name:      "100 Gold Profit",
		Threshold: 100,
		Type:      MilestoneTypeProfit,
		Reward:    25,
	})
}
