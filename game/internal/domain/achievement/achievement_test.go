package achievement

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAchievementManager(t *testing.T) {
	am := NewAchievementManager()

	assert.NotNil(t, am)
	assert.NotNil(t, am.achievements)
	assert.NotNil(t, am.unlocked)
	assert.NotNil(t, am.statistics)
	assert.NotNil(t, am.statistics.ItemTypesSold)

	// Check that achievements are initialized
	assert.Greater(t, len(am.achievements), 0)

	// Check specific achievements exist
	_, exists := am.GetAchievement(AchievementFirstProfit)
	assert.True(t, exists)

	_, exists = am.GetAchievement(AchievementMillionaire)
	assert.True(t, exists)
}

func TestUpdateTradeStats(t *testing.T) {
	am := NewAchievementManager()

	// Test first profit achievement
	am.UpdateTradeStats(100, 0.1)

	stats := am.GetStatistics()
	assert.Equal(t, 1, stats.TotalTrades)
	assert.Equal(t, 1, stats.SuccessfulTrades)
	assert.Equal(t, 100, stats.TotalProfit)
	assert.Equal(t, 1, stats.ConsecutiveProfits)

	// Check first profit achievement unlocked
	assert.True(t, am.unlocked[AchievementFirstProfit])

	// Test profit streak
	for i := 0; i < 4; i++ {
		am.UpdateTradeStats(50, 0.1)
	}

	assert.Equal(t, 5, stats.ConsecutiveProfits)
	// Profit streak 5 should be unlocked
	assert.True(t, am.unlocked[AchievementProfitStreak5])

	// Test loss
	am.UpdateTradeStats(-50, -0.1)
	assert.Equal(t, 50, stats.TotalLoss)
	assert.Equal(t, 0, stats.ConsecutiveProfits)
	assert.Equal(t, 1, stats.ConsecutiveLosses)

	// Test ROI achievements
	am.UpdateTradeStats(1000, 0.5)
	assert.True(t, am.unlocked[AchievementROI50])

	am.UpdateTradeStats(2000, 1.0)
	assert.True(t, am.unlocked[AchievementROI100])
}

func TestUpdateGoldStats(t *testing.T) {
	am := NewAchievementManager()

	// Test gold achievements
	t.Logf("Initial HighestGold: %d", am.statistics.HighestGold)

	am.UpdateGoldStats(500)
	t.Logf("After 500: HighestGold=%d", am.statistics.HighestGold)
	assert.False(t, am.unlocked[AchievementGold1000])

	am.UpdateGoldStats(1000)
	t.Logf("After 1000: HighestGold=%d, unlocked[Gold1000]=%v", am.statistics.HighestGold, am.unlocked[AchievementGold1000])
	assert.True(t, am.unlocked[AchievementGold1000])
	assert.False(t, am.unlocked[AchievementGold10000])

	am.UpdateGoldStats(10000)
	t.Logf("After 10000: HighestGold=%d, unlocked[Gold10000]=%v", am.statistics.HighestGold, am.unlocked[AchievementGold10000])
	assert.True(t, am.unlocked[AchievementGold10000])
	assert.False(t, am.unlocked[AchievementGold100000])

	am.UpdateGoldStats(100000)
	t.Logf("After 100000: HighestGold=%d, unlocked[Gold100000]=%v", am.statistics.HighestGold, am.unlocked[AchievementGold100000])
	assert.True(t, am.unlocked[AchievementGold100000])
	assert.False(t, am.unlocked[AchievementMillionaire])

	am.UpdateGoldStats(1000000)
	assert.True(t, am.unlocked[AchievementMillionaire])

	// Check highest gold stat
	stats := am.GetStatistics()
	assert.Equal(t, 1000000, stats.HighestGold)
}

func TestUpdateDayStats(t *testing.T) {
	am := NewAchievementManager()

	// Test day achievements
	am.UpdateDayStats(15)
	assert.False(t, am.unlocked[AchievementDay30])

	am.UpdateDayStats(30)
	assert.True(t, am.unlocked[AchievementDay30])
	assert.False(t, am.unlocked[AchievementDay100])

	am.UpdateDayStats(100)
	assert.True(t, am.unlocked[AchievementDay100])
	assert.False(t, am.unlocked[AchievementYear1])

	am.UpdateDayStats(365)
	assert.True(t, am.unlocked[AchievementYear1])

	stats := am.GetStatistics()
	assert.Equal(t, 365, stats.CurrentDay)
}

func TestUpdateShopStats(t *testing.T) {
	am := NewAchievementManager()

	// Test shop level achievements
	am.UpdateShopStats(2, 2)
	assert.False(t, am.unlocked[AchievementShopLevel3])

	am.UpdateShopStats(3, 3)
	assert.True(t, am.unlocked[AchievementShopLevel3])
	assert.False(t, am.unlocked[AchievementShopLevel5])

	am.UpdateShopStats(5, 4)
	assert.True(t, am.unlocked[AchievementShopLevel5])
	assert.False(t, am.unlocked[AchievementFullyEquipped])

	// Test equipment achievement
	am.UpdateShopStats(5, 5)
	assert.True(t, am.unlocked[AchievementFullyEquipped])

	stats := am.GetStatistics()
	assert.Equal(t, 5, stats.ShopLevel)
	assert.Equal(t, 5, stats.EquipmentCount)
}

func TestUpdateItemCategoryStats(t *testing.T) {
	am := NewAchievementManager()

	// Test diversification achievement
	categories := []string{"FRUIT", "POTION", "WEAPON", "GEM"}

	for _, category := range categories {
		am.UpdateItemCategoryStats(category)
	}

	// Not enough categories yet
	assert.False(t, am.unlocked[AchievementDiversified])

	// Add 5th category
	am.UpdateItemCategoryStats("ACCESSORY")
	assert.True(t, am.unlocked[AchievementDiversified])

	stats := am.GetStatistics()
	assert.Equal(t, 5, len(stats.ItemTypesSold))
	assert.Equal(t, 1, stats.ItemTypesSold["FRUIT"])
}

func TestCompleteTutorial(t *testing.T) {
	am := NewAchievementManager()

	assert.False(t, am.unlocked[AchievementCompleteTutorial])

	am.CompleteTutorial()

	assert.True(t, am.unlocked[AchievementCompleteTutorial])
	stats := am.GetStatistics()
	assert.True(t, stats.TutorialCompleted)
}

func TestProgressiveAchievements(t *testing.T) {
	am := NewAchievementManager()

	// Test progressive achievement
	achievement, exists := am.GetAchievement(AchievementMasterTrader)
	require.True(t, exists)

	assert.Equal(t, float64(0), achievement.Progress)

	// Update progress gradually
	for i := 1; i <= 100; i++ {
		am.UpdateTradeStats(100, 0.1)

		if i < 100 {
			assert.False(t, am.unlocked[AchievementMasterTrader])
		}
	}

	// Should be unlocked after 100 successful trades
	assert.True(t, am.unlocked[AchievementMasterTrader])
	assert.Equal(t, float64(100), achievement.Progress)
}

func TestGetAllAchievements(t *testing.T) {
	am := NewAchievementManager()

	achievements := am.GetAllAchievements()
	assert.Greater(t, len(achievements), 0)

	// Hidden achievements should not be returned unless unlocked
	hiddenFound := false
	for _, ach := range achievements {
		if ach.ID == AchievementMillionaire {
			hiddenFound = true
			break
		}
	}
	assert.False(t, hiddenFound, "Hidden achievement should not be visible")

	// Unlock the hidden achievement
	am.UpdateGoldStats(1000000)

	achievements = am.GetAllAchievements()
	hiddenFound = false
	for _, ach := range achievements {
		if ach.ID == AchievementMillionaire {
			hiddenFound = true
			break
		}
	}
	assert.True(t, hiddenFound, "Unlocked hidden achievement should be visible")
}

func TestGetUnlockedAchievements(t *testing.T) {
	am := NewAchievementManager()

	// Initially no achievements unlocked
	unlocked := am.GetUnlockedAchievements()
	assert.Equal(t, 0, len(unlocked))

	// Unlock some achievements
	am.UpdateTradeStats(100, 0.1)
	am.UpdateGoldStats(1000)
	am.CompleteTutorial()

	unlocked = am.GetUnlockedAchievements()
	assert.Equal(t, 3, len(unlocked))

	// Check specific achievements are unlocked
	unlockedIDs := make(map[AchievementID]bool)
	for _, ach := range unlocked {
		unlockedIDs[ach.ID] = true
	}

	assert.True(t, unlockedIDs[AchievementFirstProfit])
	assert.True(t, unlockedIDs[AchievementGold1000])
	assert.True(t, unlockedIDs[AchievementCompleteTutorial])
}

func TestGetProgress(t *testing.T) {
	am := NewAchievementManager()

	// Get initial progress
	unlocked, total, points := am.GetProgress()
	assert.Equal(t, 0, unlocked)
	assert.Greater(t, total, 0)
	assert.Equal(t, 0, points)

	// Unlock some achievements
	am.UpdateTradeStats(100, 0.1) // First Profit: 10 points
	am.UpdateGoldStats(1000)      // Gold 1000: 10 points
	am.CompleteTutorial()         // Tutorial: 5 points

	unlocked, _, points = am.GetProgress()
	assert.Equal(t, 3, unlocked)
	assert.Equal(t, 25, points)
}

func TestAchievementCallback(t *testing.T) {
	am := NewAchievementManager()

	var calledAchievements []*Achievement
	callback := func(achievement *Achievement) {
		calledAchievements = append(calledAchievements, achievement)
	}

	am.RegisterCallback(callback)

	// Trigger some achievements
	am.UpdateTradeStats(100, 0.1)
	am.CompleteTutorial()

	assert.Equal(t, 2, len(calledAchievements))
	assert.Equal(t, AchievementFirstProfit, calledAchievements[0].ID)
	assert.Equal(t, AchievementCompleteTutorial, calledAchievements[1].ID)
}

func TestReset(t *testing.T) {
	am := NewAchievementManager()

	// Unlock some achievements
	am.UpdateTradeStats(100, 0.5)
	am.UpdateGoldStats(10000)
	am.UpdateDayStats(100)

	// Verify achievements are unlocked
	assert.True(t, am.unlocked[AchievementFirstProfit])
	assert.True(t, am.unlocked[AchievementROI50])
	assert.True(t, am.unlocked[AchievementGold10000])
	assert.True(t, am.unlocked[AchievementDay100])

	stats := am.GetStatistics()
	assert.Greater(t, stats.TotalProfit, 0)
	assert.Greater(t, stats.HighestGold, 0)

	// Reset
	am.Reset()

	// Verify all achievements are locked
	assert.False(t, am.unlocked[AchievementFirstProfit])
	assert.False(t, am.unlocked[AchievementROI50])
	assert.False(t, am.unlocked[AchievementGold10000])
	assert.False(t, am.unlocked[AchievementDay100])

	// Verify statistics are reset
	stats = am.GetStatistics()
	assert.Equal(t, 0, stats.TotalProfit)
	assert.Equal(t, 0, stats.HighestGold)
	assert.Equal(t, 0, stats.CurrentDay)

	// Verify progress is reset
	achievement, _ := am.GetAchievement(AchievementFirstProfit)
	assert.Nil(t, achievement.UnlockedAt)
	assert.Equal(t, float64(0), achievement.Progress)
}

func TestExportForSteam(t *testing.T) {
	am := NewAchievementManager()

	// Unlock some achievements
	am.UpdateTradeStats(100, 0.1)
	am.UpdateGoldStats(1000)
	am.CompleteTutorial()

	steamData := am.ExportForSteam()

	// Check Steam format
	assert.Equal(t, 3, len(steamData))
	assert.True(t, steamData["ACH_first_profit"])
	assert.True(t, steamData["ACH_gold_1000"])
	assert.True(t, steamData["ACH_complete_tutorial"])

	// Non-unlocked achievements should not be exported
	assert.False(t, steamData["ACH_gold_10000"])
}

func TestImportFromSave(t *testing.T) {
	am := NewAchievementManager()

	// Create save data
	unlockedIDs := []string{
		string(AchievementFirstProfit),
		string(AchievementGold1000),
		string(AchievementDay30),
	}

	savedStats := &PlayerStatistics{
		TotalProfit:      5000,
		HighestGold:      2000,
		CurrentDay:       35,
		SuccessfulTrades: 25,
		ItemTypesSold: map[string]int{
			"FRUIT":  10,
			"WEAPON": 5,
		},
	}

	// Import
	am.ImportFromSave(unlockedIDs, savedStats)

	// Verify achievements are unlocked
	assert.True(t, am.unlocked[AchievementFirstProfit])
	assert.True(t, am.unlocked[AchievementGold1000])
	assert.True(t, am.unlocked[AchievementDay30])
	assert.False(t, am.unlocked[AchievementDay100])

	// Verify statistics are imported
	stats := am.GetStatistics()
	assert.Equal(t, 5000, stats.TotalProfit)
	assert.Equal(t, 2000, stats.HighestGold)
	assert.Equal(t, 35, stats.CurrentDay)
	assert.Equal(t, 25, stats.SuccessfulTrades)
	assert.Equal(t, 10, stats.ItemTypesSold["FRUIT"])
	assert.Equal(t, 5, stats.ItemTypesSold["WEAPON"])
}

func TestAchievementTiers(t *testing.T) {
	am := NewAchievementManager()

	// Check tier assignments
	firstProfit, _ := am.GetAchievement(AchievementFirstProfit)
	assert.Equal(t, TierBronze, firstProfit.Tier)

	roi100, _ := am.GetAchievement(AchievementROI100)
	assert.Equal(t, TierSilver, roi100.Tier)

	masterTrader, _ := am.GetAchievement(AchievementMasterTrader)
	assert.Equal(t, TierGold, masterTrader.Tier)

	millionaire, _ := am.GetAchievement(AchievementMillionaire)
	assert.Equal(t, TierPlatinum, millionaire.Tier)
}

func TestDuplicateUnlock(t *testing.T) {
	am := NewAchievementManager()

	var callCount int
	callback := func(achievement *Achievement) {
		if achievement.ID == AchievementFirstProfit {
			callCount++
		}
	}

	am.RegisterCallback(callback)

	// Unlock achievement first time
	am.UpdateTradeStats(100, 0.1)
	assert.Equal(t, 1, callCount)

	// Try to unlock again
	am.UpdateTradeStats(100, 0.1)
	assert.Equal(t, 1, callCount, "Achievement should not be unlocked twice")
}

func TestConcurrentAccess(t *testing.T) {
	am := NewAchievementManager()

	// Run concurrent operations
	done := make(chan bool, 5)

	// Goroutine 1: Update trade stats
	go func() {
		for i := 0; i < 100; i++ {
			am.UpdateTradeStats(50, 0.1)
		}
		done <- true
	}()

	// Goroutine 2: Update gold stats
	go func() {
		for i := 0; i < 100; i++ {
			am.UpdateGoldStats(i * 100)
		}
		done <- true
	}()

	// Goroutine 3: Update day stats
	go func() {
		for i := 0; i < 100; i++ {
			am.UpdateDayStats(i)
		}
		done <- true
	}()

	// Goroutine 4: Get achievements
	go func() {
		for i := 0; i < 100; i++ {
			_ = am.GetAllAchievements()
			_ = am.GetUnlockedAchievements()
		}
		done <- true
	}()

	// Goroutine 5: Get progress
	go func() {
		for i := 0; i < 100; i++ {
			_, _, _ = am.GetProgress()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify state is consistent
	stats := am.GetStatistics()
	assert.GreaterOrEqual(t, stats.TotalTrades, 0)
	assert.GreaterOrEqual(t, stats.HighestGold, 0)
	assert.GreaterOrEqual(t, stats.CurrentDay, 0)
}

func TestUnlockTiming(t *testing.T) {
	am := NewAchievementManager()

	// Unlock an achievement
	am.UpdateTradeStats(100, 0.1)

	achievement, _ := am.GetAchievement(AchievementFirstProfit)
	assert.NotNil(t, achievement.UnlockedAt)

	// Check unlock time is recent
	now := time.Now()
	diff := now.Sub(*achievement.UnlockedAt)
	assert.Less(t, diff.Seconds(), 1.0, "Unlock time should be recent")
}
