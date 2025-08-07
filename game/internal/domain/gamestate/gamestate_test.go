package gamestate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGameState(t *testing.T) {
	config := &GameConfig{
		InitialGold:       1000,
		ShopCapacity:      20,
		WarehouseCapacity: 100,
		InitialRank:       RankApprentice,
	}

	gs := NewGameState(config)
	require.NotNil(t, gs)

	// Check initial values
	assert.Equal(t, config.InitialGold, gs.GetGold())
	assert.Equal(t, config.InitialRank, gs.GetPlayerRank())
	assert.Equal(t, config.ShopCapacity, gs.GetShopCapacity())
	assert.Equal(t, config.WarehouseCapacity, gs.GetWarehouseCapacity())
	assert.Equal(t, StateInitializing, gs.GetCurrentState())
	assert.False(t, gs.IsGameOver())
	assert.Equal(t, 0, gs.GetTotalTransactions())
	assert.Equal(t, 0, gs.GetTotalProfit())
	assert.Equal(t, 0.0, gs.GetReputation())
}

func TestGameStateTransitions(t *testing.T) {
	gs := NewGameState(nil)

	// Test state transitions
	tests := []struct {
		name        string
		fromState   State
		toState     State
		shouldError bool
	}{
		{
			name:        "Initialize to Menu",
			fromState:   StateInitializing,
			toState:     StateMenu,
			shouldError: false,
		},
		{
			name:        "Menu to Playing",
			fromState:   StateMenu,
			toState:     StatePlaying,
			shouldError: false,
		},
		{
			name:        "Playing to Paused",
			fromState:   StatePlaying,
			toState:     StatePaused,
			shouldError: false,
		},
		{
			name:        "Paused to Playing",
			fromState:   StatePaused,
			toState:     StatePlaying,
			shouldError: false,
		},
		{
			name:        "Playing to GameOver",
			fromState:   StatePlaying,
			toState:     StateGameOver,
			shouldError: false,
		},
		{
			name:        "Invalid: GameOver to Playing",
			fromState:   StateGameOver,
			toState:     StatePlaying,
			shouldError: true,
		},
		{
			name:        "Playing to Saving",
			fromState:   StatePlaying,
			toState:     StateSaving,
			shouldError: false,
		},
		{
			name:        "Saving to Playing",
			fromState:   StateSaving,
			toState:     StatePlaying,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs.currentState = tt.fromState
			err := gs.TransitionTo(tt.toState)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.toState, gs.GetCurrentState())
			}
		})
	}
}

func TestGameStateGoldOperations(t *testing.T) {
	gs := NewGameState(&GameConfig{InitialGold: 1000})

	// Test adding gold
	err := gs.AddGold(500)
	assert.NoError(t, err)
	assert.Equal(t, 1500, gs.GetGold())

	// Test spending gold
	err = gs.SpendGold(300)
	assert.NoError(t, err)
	assert.Equal(t, 1200, gs.GetGold())

	// Test spending more than available
	err = gs.SpendGold(1500)
	assert.Error(t, err)
	assert.Equal(t, 1200, gs.GetGold()) // Should remain unchanged

	// Test negative amounts
	err = gs.AddGold(-100)
	assert.Error(t, err)

	err = gs.SpendGold(-100)
	assert.Error(t, err)
}

func TestGameStateRankProgression(t *testing.T) {
	gs := NewGameState(&GameConfig{InitialRank: RankApprentice})

	// Test rank progression
	tests := []struct {
		fromRank   PlayerRank
		toRank     PlayerRank
		canPromote bool
	}{
		{RankApprentice, RankJourneyman, true},
		{RankJourneyman, RankExpert, true},
		{RankExpert, RankMaster, true},
		{RankMaster, RankMaster, false}, // Can't promote beyond Master
	}

	for _, tt := range tests {
		gs.playerRank = tt.fromRank
		if tt.canPromote {
			err := gs.PromoteRank()
			assert.NoError(t, err)
			assert.Equal(t, tt.toRank, gs.GetPlayerRank())
		} else {
			err := gs.PromoteRank()
			assert.Error(t, err)
			assert.Equal(t, tt.fromRank, gs.GetPlayerRank())
		}
	}
}

func TestGameStateCapacityUpgrades(t *testing.T) {
	gs := NewGameState(&GameConfig{
		ShopCapacity:      20,
		WarehouseCapacity: 100,
	})

	// Upgrade shop capacity
	err := gs.UpgradeShopCapacity(10)
	assert.NoError(t, err)
	assert.Equal(t, 30, gs.GetShopCapacity())

	// Upgrade warehouse capacity
	err = gs.UpgradeWarehouseCapacity(50)
	assert.NoError(t, err)
	assert.Equal(t, 150, gs.GetWarehouseCapacity())

	// Test negative upgrades
	err = gs.UpgradeShopCapacity(-10)
	assert.Error(t, err)

	// Test max capacity limits
	gs.shopCapacity = 990
	err = gs.UpgradeShopCapacity(20) // Would exceed 1000
	assert.Error(t, err)
}

func TestGameStateReputation(t *testing.T) {
	gs := NewGameState(nil)

	// Add reputation
	gs.ModifyReputation(10.0)
	assert.Equal(t, 10.0, gs.GetReputation())

	// Reputation should be clamped between -100 and 100
	gs.ModifyReputation(150.0)
	assert.Equal(t, 100.0, gs.GetReputation())

	gs.ModifyReputation(-250.0)
	assert.Equal(t, -100.0, gs.GetReputation())

	// Test reputation effects (formula: 1.0 + reputation/500)
	gs.SetReputation(50.0)
	assert.Equal(t, 1.1, gs.GetReputationMultiplier()) // Good reputation: 1.0 + 50/500 = 1.1

	gs.SetReputation(-50.0)
	assert.Equal(t, 0.9, gs.GetReputationMultiplier()) // Bad reputation: 1.0 + (-50)/500 = 0.9

	gs.SetReputation(0.0)
	assert.Equal(t, 1.0, gs.GetReputationMultiplier()) // Neutral
}

func TestGameStateTransactionTracking(t *testing.T) {
	gs := NewGameState(nil)

	// Record purchases
	gs.RecordPurchase(100)
	gs.RecordPurchase(200)
	assert.Equal(t, 2, gs.GetTotalTransactions())
	assert.Equal(t, -300, gs.GetTotalProfit()) // Negative because we spent money

	// Record sales
	gs.RecordSale(400)
	gs.RecordSale(250)
	assert.Equal(t, 4, gs.GetTotalTransactions())
	assert.Equal(t, 350, gs.GetTotalProfit()) // -300 + 650 = 350

	// Get profit margin
	margin := gs.GetProfitMargin()
	assert.Greater(t, margin, 0.0)
}

func TestGameStateCallbacks(t *testing.T) {
	gs := NewGameState(nil)

	// Test state change callback
	stateChangeCalled := false
	var oldStateReceived, newStateReceived State
	gs.RegisterStateChangeCallback(func(oldState, newState State) {
		stateChangeCalled = true
		oldStateReceived = oldState
		newStateReceived = newState
	})

	gs.TransitionTo(StateMenu)
	assert.True(t, stateChangeCalled)
	assert.Equal(t, StateInitializing, oldStateReceived)
	assert.Equal(t, StateMenu, newStateReceived)

	// Test rank change callback
	rankChangeCalled := false
	var oldRankReceived, newRankReceived PlayerRank
	gs.RegisterRankChangeCallback(func(oldRank, newRank PlayerRank) {
		rankChangeCalled = true
		oldRankReceived = oldRank
		newRankReceived = newRank
	})

	gs.PromoteRank()
	assert.True(t, rankChangeCalled)
	assert.Equal(t, RankApprentice, oldRankReceived)
	assert.Equal(t, RankJourneyman, newRankReceived)

	// Test gold change callback
	goldChangeCalled := false
	var goldChangeAmount int
	gs.RegisterGoldChangeCallback(func(amount int) {
		goldChangeCalled = true
		goldChangeAmount = amount
	})

	gs.AddGold(500)
	assert.True(t, goldChangeCalled)
	assert.Equal(t, 500, goldChangeAmount)
}

func TestGameStateStatistics(t *testing.T) {
	gs := NewGameState(&GameConfig{InitialGold: 1000})

	// Simulate some game activity
	gs.RecordPurchase(300)
	gs.RecordSale(500)
	gs.RecordPurchase(200)
	gs.RecordSale(400)

	stats := gs.GetStatistics()
	assert.NotNil(t, stats)
	assert.Equal(t, 4, stats.TotalTransactions)
	assert.Equal(t, 400, stats.TotalProfit) // -500 + 900 = 400
	assert.Equal(t, 1000, stats.CurrentGold)
	assert.Equal(t, RankApprentice, stats.CurrentRank)
	assert.GreaterOrEqual(t, stats.SessionDuration, int64(0)) // May be 0 in fast tests
}

func TestGameStateSaveLoad(t *testing.T) {
	gs := NewGameState(&GameConfig{
		InitialGold: 1000,
		InitialRank: RankJourneyman,
	})

	// Modify state
	gs.AddGold(500)
	gs.RecordSale(300)
	gs.ModifyReputation(25.0)

	// Create save data
	saveData := gs.CreateSaveData()
	assert.NotNil(t, saveData)
	assert.Equal(t, 1500, saveData.Gold)
	assert.Equal(t, RankJourneyman, saveData.PlayerRank)
	assert.Equal(t, 25.0, saveData.Reputation)
	assert.Equal(t, 1, saveData.TotalTransactions)

	// Load save data into new state
	newGS := NewGameState(nil)
	err := newGS.LoadSaveData(saveData)
	assert.NoError(t, err)
	assert.Equal(t, 1500, newGS.GetGold())
	assert.Equal(t, RankJourneyman, newGS.GetPlayerRank())
	assert.Equal(t, 25.0, newGS.GetReputation())
	assert.Equal(t, 1, newGS.GetTotalTransactions())
}

func TestGameStateVictoryConditions(t *testing.T) {
	gs := NewGameState(&GameConfig{
		InitialGold: 1000,
		InitialRank: RankMaster,
	})

	// Not victory initially
	assert.False(t, gs.CheckVictoryCondition())

	// Set victory conditions
	gs.AddGold(49000) // Total: 50000
	gs.SetReputation(80.0)

	// Should be victory now
	assert.True(t, gs.CheckVictoryCondition())

	// Test defeat conditions
	gs.SpendGold(50000)     // Go to 0 gold
	gs.SetReputation(-90.0) // Very bad reputation

	assert.True(t, gs.CheckDefeatCondition())
}

func TestGetStateName(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateInitializing, "Initializing"},
		{StateMenu, "Menu"},
		{StatePlaying, "Playing"},
		{StatePaused, "Paused"},
		{StateSaving, "Saving"},
		{StateLoading, "Loading"},
		{StateGameOver, "GameOver"},
		{State(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetStateName(tt.state))
		})
	}
}

func TestGetRankName(t *testing.T) {
	tests := []struct {
		rank     PlayerRank
		expected string
	}{
		{RankApprentice, "Apprentice"},
		{RankJourneyman, "Journeyman"},
		{RankExpert, "Expert"},
		{RankMaster, "Master"},
		{PlayerRank(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetRankName(tt.rank))
		})
	}
}

func TestGameStateIntegration(t *testing.T) {
	gs := NewGameState(&GameConfig{
		InitialGold:       1000,
		ShopCapacity:      20,
		WarehouseCapacity: 100,
		InitialRank:       RankApprentice,
	})

	// Start game
	err := gs.TransitionTo(StateMenu)
	require.NoError(t, err)
	err = gs.TransitionTo(StatePlaying)
	require.NoError(t, err)

	// Simulate trading
	gs.RecordPurchase(300) // Buy items
	gs.SpendGold(300)
	assert.Equal(t, 700, gs.GetGold())

	gs.RecordSale(500) // Sell items
	gs.AddGold(500)
	assert.Equal(t, 1200, gs.GetGold())

	// Check profit
	assert.Equal(t, 200, gs.GetTotalProfit())

	// Earn reputation from good trade
	gs.ModifyReputation(5.0)

	// Upgrade shop
	if gs.GetGold() >= 500 {
		gs.SpendGold(500)
		gs.UpgradeShopCapacity(10)
		assert.Equal(t, 30, gs.GetShopCapacity())
	}

	// Check for rank promotion eligibility
	if gs.GetGold() >= 1000 && gs.GetReputation() >= 20 {
		err = gs.PromoteRank()
		assert.NoError(t, err)
	}

	// Save game
	err = gs.TransitionTo(StateSaving)
	assert.NoError(t, err)
	saveData := gs.CreateSaveData()
	assert.NotNil(t, saveData)

	// Return to playing
	err = gs.TransitionTo(StatePlaying)
	assert.NoError(t, err)
}
