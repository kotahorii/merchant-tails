// Package gamestate manages the overall game state and progression
package gamestate

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents the current state of the game
type State int

const (
	StateInitializing State = iota
	StateMenu
	StatePlaying
	StatePaused
	StateSaving
	StateLoading
	StateGameOver
)

// PlayerRank represents the player's progression level
type PlayerRank int

const (
	RankApprentice PlayerRank = iota
	RankJourneyman
	RankExpert
	RankMaster
)

// GameConfig holds the configuration for initializing a game state
type GameConfig struct {
	InitialGold       int
	ShopCapacity      int
	WarehouseCapacity int
	InitialRank       PlayerRank
}

// SaveData represents the data structure for saving/loading game state
type SaveData struct {
	Gold              int
	PlayerRank        PlayerRank
	ShopCapacity      int
	WarehouseCapacity int
	Reputation        float64
	TotalTransactions int
	TotalProfit       int
	SaveTime          time.Time
}

// Statistics holds game statistics
type Statistics struct {
	TotalTransactions int
	TotalProfit       int
	CurrentGold       int
	CurrentRank       PlayerRank
	Reputation        float64
	SessionDuration   int64 // in seconds
	ShopCapacity      int
	WarehouseCapacity int
}

// StateChangeCallback is called when the game state changes
type StateChangeCallback func(oldState, newState State)

// RankChangeCallback is called when the player's rank changes
type RankChangeCallback func(oldRank, newRank PlayerRank)

// GoldChangeCallback is called when gold amount changes
type GoldChangeCallback func(amount int)

// GameState manages the overall state of the game
type GameState struct {
	// Core state
	currentState State
	playerRank   PlayerRank
	gold         int
	reputation   float64

	// Capacity
	shopCapacity      int
	warehouseCapacity int

	// Statistics
	totalTransactions int
	totalProfit       int
	totalExpenses     int
	totalRevenue      int
	sessionStartTime  time.Time

	// Callbacks
	stateChangeCallbacks []StateChangeCallback
	rankChangeCallbacks  []RankChangeCallback
	goldChangeCallbacks  []GoldChangeCallback

	// Thread safety
	mu sync.RWMutex
}

// Default configuration values
var defaultConfig = &GameConfig{
	InitialGold:       1000,
	ShopCapacity:      20,
	WarehouseCapacity: 100,
	InitialRank:       RankApprentice,
}

// Constants for game mechanics
const (
	MaxShopCapacity      = 1000
	MaxWarehouseCapacity = 5000
	MaxReputation        = 100.0
	MinReputation        = -100.0
	VictoryGoldThreshold = 50000
	VictoryRepThreshold  = 75.0
	DefeatGoldThreshold  = 0
	DefeatRepThreshold   = -75.0
)

// NewGameState creates a new game state
func NewGameState(config *GameConfig) *GameState {
	if config == nil {
		config = defaultConfig
	}

	gs := &GameState{
		currentState:         StateInitializing,
		playerRank:           config.InitialRank,
		gold:                 config.InitialGold,
		reputation:           0.0,
		shopCapacity:         config.ShopCapacity,
		warehouseCapacity:    config.WarehouseCapacity,
		totalTransactions:    0,
		totalProfit:          0,
		sessionStartTime:     time.Now(),
		stateChangeCallbacks: make([]StateChangeCallback, 0),
		rankChangeCallbacks:  make([]RankChangeCallback, 0),
		goldChangeCallbacks:  make([]GoldChangeCallback, 0),
	}

	return gs
}

// GetCurrentState returns the current game state
func (gs *GameState) GetCurrentState() State {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.currentState
}

// TransitionTo attempts to transition to a new state
func (gs *GameState) TransitionTo(newState State) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Validate transition
	if !gs.isValidTransition(gs.currentState, newState) {
		return fmt.Errorf("invalid state transition from %s to %s",
			GetStateName(gs.currentState), GetStateName(newState))
	}

	oldState := gs.currentState
	gs.currentState = newState

	// Notify callbacks
	for _, callback := range gs.stateChangeCallbacks {
		callback(oldState, newState)
	}

	return nil
}

// isValidTransition checks if a state transition is valid
func (gs *GameState) isValidTransition(from, to State) bool {
	// Define valid transitions
	validTransitions := map[State][]State{
		StateInitializing: {StateMenu, StateLoading},
		StateMenu:         {StatePlaying, StateLoading, StateGameOver},
		StatePlaying:      {StatePaused, StateSaving, StateGameOver, StateMenu},
		StatePaused:       {StatePlaying, StateSaving, StateMenu, StateGameOver},
		StateSaving:       {StatePlaying, StatePaused, StateMenu},
		StateLoading:      {StateMenu, StatePlaying},
		StateGameOver:     {StateMenu}, // Can only go back to menu from game over
	}

	validStates, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, validState := range validStates {
		if validState == to {
			return true
		}
	}

	return false
}

// GetGold returns the current gold amount
func (gs *GameState) GetGold() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.gold
}

// AddGold adds gold to the player's inventory
func (gs *GameState) AddGold(amount int) error {
	if amount < 0 {
		return errors.New("cannot add negative gold amount")
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.gold += amount

	// Notify callbacks
	for _, callback := range gs.goldChangeCallbacks {
		callback(amount)
	}

	return nil
}

// SpendGold removes gold from the player's inventory
func (gs *GameState) SpendGold(amount int) error {
	if amount < 0 {
		return errors.New("cannot spend negative gold amount")
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.gold < amount {
		return fmt.Errorf("insufficient gold: have %d, need %d", gs.gold, amount)
	}

	gs.gold -= amount

	// Notify callbacks
	for _, callback := range gs.goldChangeCallbacks {
		callback(-amount)
	}

	return nil
}

// GetPlayerRank returns the current player rank
func (gs *GameState) GetPlayerRank() PlayerRank {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.playerRank
}

// PromoteRank promotes the player to the next rank
func (gs *GameState) PromoteRank() error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.playerRank >= RankMaster {
		return errors.New("already at maximum rank")
	}

	oldRank := gs.playerRank
	gs.playerRank++

	// Notify callbacks
	for _, callback := range gs.rankChangeCallbacks {
		callback(oldRank, gs.playerRank)
	}

	return nil
}

// GetShopCapacity returns the current shop capacity
func (gs *GameState) GetShopCapacity() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.shopCapacity
}

// GetWarehouseCapacity returns the current warehouse capacity
func (gs *GameState) GetWarehouseCapacity() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.warehouseCapacity
}

// UpgradeShopCapacity increases shop capacity
func (gs *GameState) UpgradeShopCapacity(amount int) error {
	if amount <= 0 {
		return errors.New("upgrade amount must be positive")
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.shopCapacity+amount > MaxShopCapacity {
		return fmt.Errorf("would exceed maximum shop capacity of %d", MaxShopCapacity)
	}

	gs.shopCapacity += amount
	return nil
}

// UpgradeWarehouseCapacity increases warehouse capacity
func (gs *GameState) UpgradeWarehouseCapacity(amount int) error {
	if amount <= 0 {
		return errors.New("upgrade amount must be positive")
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.warehouseCapacity+amount > MaxWarehouseCapacity {
		return fmt.Errorf("would exceed maximum warehouse capacity of %d", MaxWarehouseCapacity)
	}

	gs.warehouseCapacity += amount
	return nil
}

// GetReputation returns the current reputation
func (gs *GameState) GetReputation() float64 {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.reputation
}

// ModifyReputation changes the reputation by the given amount
func (gs *GameState) ModifyReputation(delta float64) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.reputation += delta
	gs.reputation = clampFloat64(gs.reputation, MinReputation, MaxReputation)
}

// SetReputation sets the reputation to a specific value
func (gs *GameState) SetReputation(value float64) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.reputation = clampFloat64(value, MinReputation, MaxReputation)
}

// GetReputationMultiplier returns a multiplier based on reputation
func (gs *GameState) GetReputationMultiplier() float64 {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// Linear scaling: -100 = 0.9x, 0 = 1.0x, 100 = 1.1x
	return 1.0 + (gs.reputation / 500.0)
}

// RecordPurchase records a purchase transaction
func (gs *GameState) RecordPurchase(amount int) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.totalTransactions++
	gs.totalExpenses += amount
	gs.totalProfit -= amount
}

// RecordSale records a sale transaction
func (gs *GameState) RecordSale(amount int) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.totalTransactions++
	gs.totalRevenue += amount
	gs.totalProfit += amount
}

// GetTotalTransactions returns the total number of transactions
func (gs *GameState) GetTotalTransactions() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.totalTransactions
}

// GetTotalProfit returns the total profit
func (gs *GameState) GetTotalProfit() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.totalProfit
}

// GetProfitMargin returns the profit margin as a percentage
func (gs *GameState) GetProfitMargin() float64 {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.totalRevenue == 0 {
		return 0.0
	}

	return float64(gs.totalProfit) / float64(gs.totalRevenue) * 100.0
}

// IsGameOver returns true if the game is over
func (gs *GameState) IsGameOver() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.currentState == StateGameOver
}

// CheckVictoryCondition checks if the player has won
func (gs *GameState) CheckVictoryCondition() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return gs.gold >= VictoryGoldThreshold &&
		gs.reputation >= VictoryRepThreshold &&
		gs.playerRank == RankMaster
}

// CheckDefeatCondition checks if the player has lost
func (gs *GameState) CheckDefeatCondition() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return gs.gold <= DefeatGoldThreshold ||
		gs.reputation <= DefeatRepThreshold
}

// RegisterStateChangeCallback registers a callback for state changes
func (gs *GameState) RegisterStateChangeCallback(callback StateChangeCallback) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.stateChangeCallbacks = append(gs.stateChangeCallbacks, callback)
}

// RegisterRankChangeCallback registers a callback for rank changes
func (gs *GameState) RegisterRankChangeCallback(callback RankChangeCallback) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.rankChangeCallbacks = append(gs.rankChangeCallbacks, callback)
}

// RegisterGoldChangeCallback registers a callback for gold changes
func (gs *GameState) RegisterGoldChangeCallback(callback GoldChangeCallback) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.goldChangeCallbacks = append(gs.goldChangeCallbacks, callback)
}

// GetStatistics returns current game statistics
func (gs *GameState) GetStatistics() *Statistics {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return &Statistics{
		TotalTransactions: gs.totalTransactions,
		TotalProfit:       gs.totalProfit,
		CurrentGold:       gs.gold,
		CurrentRank:       gs.playerRank,
		Reputation:        gs.reputation,
		SessionDuration:   int64(time.Since(gs.sessionStartTime).Seconds()),
		ShopCapacity:      gs.shopCapacity,
		WarehouseCapacity: gs.warehouseCapacity,
	}
}

// CreateSaveData creates a save data snapshot
func (gs *GameState) CreateSaveData() *SaveData {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return &SaveData{
		Gold:              gs.gold,
		PlayerRank:        gs.playerRank,
		ShopCapacity:      gs.shopCapacity,
		WarehouseCapacity: gs.warehouseCapacity,
		Reputation:        gs.reputation,
		TotalTransactions: gs.totalTransactions,
		TotalProfit:       gs.totalProfit,
		SaveTime:          time.Now(),
	}
}

// LoadSaveData loads game state from save data
func (gs *GameState) LoadSaveData(data *SaveData) error {
	if data == nil {
		return errors.New("save data is nil")
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.gold = data.Gold
	gs.playerRank = data.PlayerRank
	gs.shopCapacity = data.ShopCapacity
	gs.warehouseCapacity = data.WarehouseCapacity
	gs.reputation = data.Reputation
	gs.totalTransactions = data.TotalTransactions
	gs.totalProfit = data.TotalProfit

	return nil
}

// GetStateName returns the string name of a state
func GetStateName(state State) string {
	switch state {
	case StateInitializing:
		return "Initializing"
	case StateMenu:
		return "Menu"
	case StatePlaying:
		return "Playing"
	case StatePaused:
		return "Paused"
	case StateSaving:
		return "Saving"
	case StateLoading:
		return "Loading"
	case StateGameOver:
		return "GameOver"
	default:
		return "Unknown"
	}
}

// GetRankName returns the string name of a rank
func GetRankName(rank PlayerRank) string {
	switch rank {
	case RankApprentice:
		return "Apprentice"
	case RankJourneyman:
		return "Journeyman"
	case RankExpert:
		return "Expert"
	case RankMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

// clampFloat64 clamps a float64 value between min and max
func clampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
