package tutorial

import (
	"sync"
)

// SimpleTutorial provides basic tutorial functionality
type SimpleTutorial struct {
	currentStep int
	completed   bool
	steps       []TutorialStep
	mu          sync.RWMutex
}

// TutorialStep represents a simple tutorial step
type TutorialStep struct {
	ID          string
	Title       string
	Description string
	Hint        string
	Completed   bool
}

// NewSimpleTutorial creates a basic tutorial system
func NewSimpleTutorial() *SimpleTutorial {
	return &SimpleTutorial{
		currentStep: 0,
		completed:   false,
		steps:       createBasicTutorialSteps(),
	}
}

// createBasicTutorialSteps creates the essential tutorial steps
func createBasicTutorialSteps() []TutorialStep {
	return []TutorialStep{
		{
			ID:          "welcome",
			Title:       "Welcome to Merchant Tails",
			Description: "Welcome! You're about to start your journey as a merchant. Let's learn the basics of trading.",
			Hint:        "Click Next to continue",
		},
		{
			ID:          "buy_items",
			Title:       "How to Buy Items",
			Description: "Visit the market to buy items from suppliers. Look for items with low prices that you can sell for profit.",
			Hint:        "Open the market and select a supplier",
		},
		{
			ID:          "set_prices",
			Title:       "Setting Prices",
			Description: "Set your selling prices wisely. Too high and customers won't buy. Too low and you won't make profit.",
			Hint:        "Try a 20-30% markup on your purchase price",
		},
		{
			ID:          "manage_inventory",
			Title:       "Managing Inventory",
			Description: "Keep track of your inventory. Some items like fruits can spoil over time.",
			Hint:        "Check your inventory regularly",
		},
		{
			ID:          "watch_market",
			Title:       "Market Trends",
			Description: "Prices change based on supply and demand. Buy low, sell high!",
			Hint:        "Watch for price changes in the market",
		},
		{
			ID:          "make_profit",
			Title:       "Making Profit",
			Description: "Your goal is to make profit. Calculate your ROI (Return on Investment) for each trade.",
			Hint:        "ROI = (Selling Price - Purchase Price) / Purchase Price × 100%",
		},
		{
			ID:          "save_money",
			Title:       "Using the Bank",
			Description: "Save your profits in the bank to earn interest. This is your first investment!",
			Hint:        "Deposit gold to earn 2% annual interest",
		},
		{
			ID:          "complete",
			Title:       "Tutorial Complete",
			Description: "Congratulations! You've learned the basics. Now go build your merchant empire!",
			Hint:        "Good luck!",
		},
	}
}

// GetCurrentStep returns the current tutorial step
func (st *SimpleTutorial) GetCurrentStep() *TutorialStep {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if st.currentStep >= len(st.steps) {
		return nil
	}
	return &st.steps[st.currentStep]
}

// NextStep advances to the next tutorial step
func (st *SimpleTutorial) NextStep() bool {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.currentStep < len(st.steps)-1 {
		st.steps[st.currentStep].Completed = true
		st.currentStep++
		return true
	}

	// Mark tutorial as completed
	if st.currentStep == len(st.steps)-1 {
		st.steps[st.currentStep].Completed = true
		st.completed = true
	}

	return false
}

// Skip skips the tutorial
func (st *SimpleTutorial) Skip() {
	st.mu.Lock()
	defer st.mu.Unlock()

	for i := range st.steps {
		st.steps[i].Completed = true
	}
	st.currentStep = len(st.steps) - 1
	st.completed = true
}

// IsCompleted returns whether the tutorial is completed
func (st *SimpleTutorial) IsCompleted() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.completed
}

// GetProgress returns the current progress
func (st *SimpleTutorial) GetProgress() (current int, total int) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	completed := 0
	for _, step := range st.steps {
		if step.Completed {
			completed++
		}
	}

	return completed, len(st.steps)
}

// Reset resets the tutorial to the beginning
func (st *SimpleTutorial) Reset() {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.currentStep = 0
	st.completed = false
	for i := range st.steps {
		st.steps[i].Completed = false
	}
}

// GetAllSteps returns all tutorial steps for display
func (st *SimpleTutorial) GetAllSteps() []TutorialStep {
	st.mu.RLock()
	defer st.mu.RUnlock()

	// Return a copy to prevent external modification
	steps := make([]TutorialStep, len(st.steps))
	copy(steps, st.steps)
	return steps
}

// GetHint returns a hint for the current step
func (st *SimpleTutorial) GetHint() string {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if st.currentStep >= len(st.steps) {
		return ""
	}
	return st.steps[st.currentStep].Hint
}
