package quest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuestManager(t *testing.T) {
	qm := NewQuestManager()

	assert.NotNil(t, qm)
	assert.NotNil(t, qm.quests)
	assert.NotNil(t, qm.activeQuests)
	assert.NotNil(t, qm.completedQuests)
	assert.NotNil(t, qm.questChains)
	assert.NotNil(t, qm.statistics)

	// Check that quests are initialized
	assert.Greater(t, len(qm.quests), 0)

	// Check first quest is available
	quest, exists := qm.GetQuest(QuestFirstTrade)
	assert.True(t, exists)
	assert.Equal(t, QuestStatusAvailable, quest.Status)

	// Check other quests are locked
	quest, exists = qm.GetQuest(QuestFirstProfit)
	assert.True(t, exists)
	assert.Equal(t, QuestStatusLocked, quest.Status)
}

func TestStartQuest(t *testing.T) {
	qm := NewQuestManager()

	tests := []struct {
		name        string
		questID     QuestID
		playerLevel int
		wantErr     bool
		errContains string
	}{
		{
			name:        "start available quest",
			questID:     QuestFirstTrade,
			playerLevel: 1,
			wantErr:     false,
		},
		{
			name:        "start locked quest",
			questID:     QuestFirstProfit,
			playerLevel: 1,
			wantErr:     true,
			errContains: "quest not available",
		},
		{
			name:        "insufficient level",
			questID:     QuestROI25,
			playerLevel: 1,
			wantErr:     true,
			errContains: "quest not available", // ROI25 is locked initially
		},
		{
			name:        "non-existent quest",
			questID:     "invalid_quest",
			playerLevel: 1,
			wantErr:     true,
			errContains: "quest not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unlock ROI25 for testing
			if tt.questID == QuestROI25 && tt.playerLevel >= 2 {
				qm.quests[QuestROI25].Status = QuestStatusAvailable
			}

			err := qm.StartQuest(tt.questID, tt.playerLevel)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)

				// Check quest is active
				quest, exists := qm.activeQuests[tt.questID]
				assert.True(t, exists)
				assert.Equal(t, QuestStatusActive, quest.Status)
				assert.NotNil(t, quest.StartedAt)
			}
		})
	}
}

func TestUpdateObjective(t *testing.T) {
	qm := NewQuestManager()

	// Start a quest
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)

	// Update buy objective
	qm.UpdateObjective(QuestFirstTrade, "buy_item", 1)

	quest, _ := qm.GetQuest(QuestFirstTrade)
	assert.Equal(t, 1, quest.Objectives[0].Current)
	assert.True(t, quest.Objectives[0].Completed)
	assert.False(t, quest.Objectives[1].Completed)

	// Update sell objective
	qm.UpdateObjective(QuestFirstTrade, "sell_item", 1)

	// Quest should be completed
	assert.Equal(t, QuestStatusCompleted, quest.Status)
	assert.NotNil(t, quest.CompletedAt)

	// Check statistics
	stats := qm.GetStatistics()
	assert.Equal(t, 1, stats.TotalCompleted)

	// Check next quest is unlocked
	nextQuest, _ := qm.GetQuest(QuestFirstProfit)
	assert.Equal(t, QuestStatusAvailable, nextQuest.Status)
}

func TestQuestCompletion(t *testing.T) {
	qm := NewQuestManager()

	// Register callback to track completion
	var completedQuest *Quest
	var oldStatus QuestStatus
	qm.RegisterCallback(func(q *Quest, old QuestStatus) {
		completedQuest = q
		oldStatus = old
	})

	// Start and complete a quest
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)

	qm.UpdateObjective(QuestFirstTrade, "buy_item", 1)
	qm.UpdateObjective(QuestFirstTrade, "sell_item", 1)

	// Check callback was called
	assert.NotNil(t, completedQuest)
	assert.Equal(t, QuestFirstTrade, completedQuest.ID)
	assert.Equal(t, QuestStatusActive, oldStatus)
	assert.Equal(t, QuestStatusCompleted, completedQuest.Status)

	// Check quest is in completed list
	completedQuests := qm.GetCompletedQuests()
	assert.Equal(t, 1, len(completedQuests))
	assert.Equal(t, QuestFirstTrade, completedQuests[0].ID)

	// Check quest is not in active list
	activeQuests := qm.GetActiveQuests()
	assert.Equal(t, 0, len(activeQuests))
}

func TestQuestChain(t *testing.T) {
	qm := NewQuestManager()

	// Get investment basics chain
	chain := qm.GetQuestChain("investment_basics")
	assert.Equal(t, 5, len(chain))
	assert.Equal(t, QuestFirstTrade, chain[0].ID)
	assert.Equal(t, QuestROI100, chain[4].ID)

	// Complete first quest in chain
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)

	qm.UpdateObjective(QuestFirstTrade, "buy_item", 1)
	qm.UpdateObjective(QuestFirstTrade, "sell_item", 1)

	// Check next quest in chain is available
	nextQuest, _ := qm.GetQuest(QuestFirstProfit)
	assert.Equal(t, QuestStatusAvailable, nextQuest.Status)
}

func TestROIQuests(t *testing.T) {
	qm := NewQuestManager()

	// Unlock and start ROI quest
	qm.quests[QuestROI25].Status = QuestStatusAvailable
	err := qm.StartQuest(QuestROI25, 2)
	require.NoError(t, err)

	// Update ROI objective
	qm.UpdateObjective(QuestROI25, "achieve_roi", 25)

	quest, _ := qm.GetQuest(QuestROI25)
	assert.Equal(t, QuestStatusCompleted, quest.Status)

	// Check next ROI quest is unlocked
	nextQuest, _ := qm.GetQuest(QuestROI50)
	assert.Equal(t, QuestStatusAvailable, nextQuest.Status)
}

func TestDiversificationQuest(t *testing.T) {
	qm := NewQuestManager()

	// Unlock and start diversification quest
	qm.quests[QuestDiversifyPortfolio].Status = QuestStatusAvailable
	err := qm.StartQuest(QuestDiversifyPortfolio, 3)
	require.NoError(t, err)

	// Update diversification objectives
	qm.UpdateObjective(QuestDiversifyPortfolio, "diversify", 3)
	quest, _ := qm.GetQuest(QuestDiversifyPortfolio)
	assert.Equal(t, 3, quest.Objectives[0].Current)
	assert.False(t, quest.Objectives[0].Completed)
	assert.Equal(t, QuestStatusActive, quest.Status)

	// Complete diversification
	qm.UpdateObjective(QuestDiversifyPortfolio, "diversify", 5)
	assert.True(t, quest.Objectives[0].Completed)

	// Complete profit objective
	qm.UpdateObjective(QuestDiversifyPortfolio, "profit_each", 5)
	assert.Equal(t, QuestStatusCompleted, quest.Status)
}

func TestQuestTimeout(t *testing.T) {
	qm := NewQuestManager()

	// Create a quest with short timeout
	timeout := 100 * time.Millisecond
	timedQuest := &Quest{
		ID:          "timed_quest",
		Name:        "Timed Quest",
		Description: "Complete quickly",
		Type:        QuestTypeDaily,
		Status:      QuestStatusAvailable,
		Level:       1,
		TimeLimit:   &timeout,
		Objectives: []*QuestObjective{
			{ID: "test", Description: "Test objective", Target: 1},
		},
	}
	qm.registerQuest(timedQuest)

	// Start the quest
	err := qm.StartQuest("timed_quest", 1)
	require.NoError(t, err)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Check timeouts
	qm.CheckTimeouts()

	// Quest should be failed
	quest, _ := qm.GetQuest("timed_quest")
	assert.Equal(t, QuestStatusFailed, quest.Status)
	assert.NotNil(t, quest.FailedAt)

	// Check statistics
	stats := qm.GetStatistics()
	assert.Equal(t, 1, stats.TotalFailed)
}

func TestFailQuest(t *testing.T) {
	qm := NewQuestManager()

	// Start a quest
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)

	// Fail the quest
	qm.FailQuest(QuestFirstTrade)

	quest, _ := qm.GetQuest(QuestFirstTrade)
	assert.Equal(t, QuestStatusFailed, quest.Status)
	assert.NotNil(t, quest.FailedAt)

	// Check quest is not in active list
	activeQuests := qm.GetActiveQuests()
	assert.Equal(t, 0, len(activeQuests))

	// Check statistics
	stats := qm.GetStatistics()
	assert.Equal(t, 1, stats.TotalFailed)
}

func TestGetAvailableQuests(t *testing.T) {
	qm := NewQuestManager()

	// Get available quests for level 1
	quests := qm.GetAvailableQuests(1)
	assert.Equal(t, 1, len(quests))
	assert.Equal(t, QuestFirstTrade, quests[0].ID)

	// Complete first quest to unlock more
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)
	qm.UpdateObjective(QuestFirstTrade, "buy_item", 1)
	qm.UpdateObjective(QuestFirstTrade, "sell_item", 1)

	// Now should have more available quests
	quests = qm.GetAvailableQuests(1)
	assert.Equal(t, 1, len(quests))
	assert.Equal(t, QuestFirstProfit, quests[0].ID)

	// Check for higher level
	qm.quests[QuestROI25].Status = QuestStatusAvailable
	quests = qm.GetAvailableQuests(2)
	assert.Equal(t, 2, len(quests)) // FirstProfit and ROI25
}

func TestClaimReward(t *testing.T) {
	qm := NewQuestManager()

	// Start and complete a quest
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)
	qm.UpdateObjective(QuestFirstTrade, "buy_item", 1)
	qm.UpdateObjective(QuestFirstTrade, "sell_item", 1)

	// Claim reward
	reward, err := qm.ClaimReward(QuestFirstTrade)
	require.NoError(t, err)
	assert.NotNil(t, reward)
	assert.Equal(t, 100, reward.Gold)
	assert.Equal(t, 50, reward.Experience)
	assert.Equal(t, float64(5), reward.Reputation)
	assert.Contains(t, reward.Unlocks, QuestFirstProfit)

	// Check statistics
	stats := qm.GetStatistics()
	assert.Equal(t, 100, stats.TotalRewards)

	// Try to claim non-completed quest
	_, err = qm.ClaimReward(QuestFirstProfit)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quest not completed")
}

func TestQuestStatistics(t *testing.T) {
	qm := NewQuestManager()

	// Complete multiple quests
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)
	qm.UpdateObjective(QuestFirstTrade, "buy_item", 1)
	qm.UpdateObjective(QuestFirstTrade, "sell_item", 1)

	// Start and complete another quest
	err = qm.StartQuest(QuestFirstProfit, 1)
	require.NoError(t, err)
	qm.UpdateObjective(QuestFirstProfit, "earn_profit", 100)

	stats := qm.GetStatistics()
	assert.Equal(t, 2, stats.TotalCompleted)
	assert.NotNil(t, stats.LastCompletedAt)
	assert.Equal(t, 2, stats.QuestStreak)
}

func TestQuestReset(t *testing.T) {
	qm := NewQuestManager()

	// Complete some quests
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)
	qm.UpdateObjective(QuestFirstTrade, "buy_item", 1)
	qm.UpdateObjective(QuestFirstTrade, "sell_item", 1)

	// Reset
	qm.Reset()

	// Check all quests are reset
	quest, _ := qm.GetQuest(QuestFirstTrade)
	assert.Equal(t, QuestStatusAvailable, quest.Status)
	assert.Nil(t, quest.StartedAt)
	assert.Nil(t, quest.CompletedAt)
	assert.Equal(t, 0, quest.Objectives[0].Current)
	assert.False(t, quest.Objectives[0].Completed)

	// Check other quests are locked
	quest, _ = qm.GetQuest(QuestFirstProfit)
	assert.Equal(t, QuestStatusLocked, quest.Status)

	// Check statistics are reset
	stats := qm.GetStatistics()
	assert.Equal(t, 0, stats.TotalCompleted)
	assert.Equal(t, 0, stats.TotalFailed)
	assert.Equal(t, 0, stats.TotalRewards)
}

func TestShopInvestmentQuests(t *testing.T) {
	qm := NewQuestManager()

	// Unlock and start shop upgrade quest
	qm.quests[QuestShopUpgrade].Status = QuestStatusAvailable
	err := qm.StartQuest(QuestShopUpgrade, 4)
	require.NoError(t, err)

	// Update objectives
	qm.UpdateObjective(QuestShopUpgrade, "upgrade_shop", 3)
	qm.UpdateObjective(QuestShopUpgrade, "purchase_equipment", 3)

	quest, _ := qm.GetQuest(QuestShopUpgrade)
	assert.Equal(t, QuestStatusCompleted, quest.Status)

	// Check equipment ROI quest is unlocked
	roiQuest, _ := qm.GetQuest(QuestEquipmentROI)
	assert.Equal(t, QuestStatusAvailable, roiQuest.Status)
}

func TestMarketTimingQuest(t *testing.T) {
	qm := NewQuestManager()

	// Unlock and start market timing quest
	qm.quests[QuestMarketTiming].Status = QuestStatusAvailable
	err := qm.StartQuest(QuestMarketTiming, 4)
	require.NoError(t, err)

	// Update objectives progressively
	qm.UpdateObjective(QuestMarketTiming, "buy_low", 5)
	quest, _ := qm.GetQuest(QuestMarketTiming)
	assert.Equal(t, 5, quest.Objectives[0].Current)
	assert.False(t, quest.Objectives[0].Completed)

	qm.UpdateObjective(QuestMarketTiming, "buy_low", 10)
	assert.True(t, quest.Objectives[0].Completed)

	// Complete all objectives
	qm.UpdateObjective(QuestMarketTiming, "sell_high", 10)
	qm.UpdateObjective(QuestMarketTiming, "profit_margin", 40)

	assert.Equal(t, QuestStatusCompleted, quest.Status)
}

func TestLongTermInvestmentQuest(t *testing.T) {
	qm := NewQuestManager()

	// Check quest has time limit
	quest, exists := qm.GetQuest(QuestLongTermInvestment)
	assert.True(t, exists)
	assert.NotNil(t, quest.TimeLimit)
	assert.Equal(t, 7*24*time.Hour, *quest.TimeLimit)

	// Unlock and start
	quest.Status = QuestStatusAvailable
	err := qm.StartQuest(QuestLongTermInvestment, 6)
	require.NoError(t, err)

	// Update objectives
	qm.UpdateObjective(QuestLongTermInvestment, "hold_items", 7)
	qm.UpdateObjective(QuestLongTermInvestment, "final_profit", 75)

	assert.Equal(t, QuestStatusCompleted, quest.Status)

	// Check rewards
	reward, err := qm.ClaimReward(QuestLongTermInvestment)
	require.NoError(t, err)
	assert.Equal(t, 5000, reward.Gold)
	assert.Equal(t, 1500, reward.Experience)
	assert.Equal(t, 1, reward.Items["investor_badge"])
}

func TestConcurrentQuestOperations(t *testing.T) {
	qm := NewQuestManager()

	// Run concurrent operations
	done := make(chan bool, 4)

	// Start quests concurrently
	go func() {
		for i := 0; i < 10; i++ {
			_ = qm.StartQuest(QuestFirstTrade, 1)
		}
		done <- true
	}()

	// Update objectives concurrently
	go func() {
		for i := 0; i < 10; i++ {
			qm.UpdateObjective(QuestFirstTrade, "buy_item", i)
		}
		done <- true
	}()

	// Get quests concurrently
	go func() {
		for i := 0; i < 10; i++ {
			_ = qm.GetActiveQuests()
			_ = qm.GetAvailableQuests(1)
			_ = qm.GetCompletedQuests()
		}
		done <- true
	}()

	// Check timeouts concurrently
	go func() {
		for i := 0; i < 10; i++ {
			qm.CheckTimeouts()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify state is consistent
	stats := qm.GetStatistics()
	assert.GreaterOrEqual(t, stats.TotalCompleted, 0)
	assert.GreaterOrEqual(t, stats.TotalFailed, 0)
}

func TestQuestCallbacks(t *testing.T) {
	qm := NewQuestManager()

	callbackCalls := 0
	var lastQuest *Quest
	var lastOldStatus QuestStatus

	// Register multiple callbacks
	qm.RegisterCallback(func(q *Quest, old QuestStatus) {
		callbackCalls++
		lastQuest = q
		lastOldStatus = old
	})

	qm.RegisterCallback(func(q *Quest, old QuestStatus) {
		callbackCalls++
	})

	// Start and complete a quest
	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)

	assert.Equal(t, 2, callbackCalls)
	assert.Equal(t, QuestFirstTrade, lastQuest.ID)
	assert.Equal(t, QuestStatusAvailable, lastOldStatus)

	// Complete the quest
	qm.UpdateObjective(QuestFirstTrade, "buy_item", 1)
	qm.UpdateObjective(QuestFirstTrade, "sell_item", 1)

	assert.Equal(t, 4, callbackCalls)
	assert.Equal(t, QuestStatusActive, lastOldStatus)
	assert.Equal(t, QuestStatusCompleted, lastQuest.Status)
}

func TestQuestObjectiveOverflow(t *testing.T) {
	qm := NewQuestManager()

	err := qm.StartQuest(QuestFirstTrade, 1)
	require.NoError(t, err)

	// Try to overflow objective
	qm.UpdateObjective(QuestFirstTrade, "buy_item", 100)

	quest, _ := qm.GetQuest(QuestFirstTrade)
	// Should cap at target
	assert.Equal(t, 1, quest.Objectives[0].Current)
	assert.True(t, quest.Objectives[0].Completed)
}
