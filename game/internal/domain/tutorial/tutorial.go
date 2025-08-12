package tutorial

import (
	"sync"
	"time"
)

// StepStatus represents the status of a tutorial step
type StepStatus int

const (
	StatusPending StepStatus = iota
	StatusActive
	StatusCompleted
	StatusSkipped
)

// TutorialStep represents a single step in the tutorial
type TutorialStep struct {
	ID          string
	Title       string
	Description string
	Objective   string
	Hint        string
	Status      StepStatus
	Required    bool // Can't skip if true
	Order       int
	Condition   func() bool // Condition to check if step is completed
	OnComplete  func()      // Callback when step is completed
}

// TutorialManager manages the tutorial flow
type TutorialManager struct {
	steps       map[string]*TutorialStep
	currentStep *TutorialStep
	completed   map[string]bool
	isActive    bool
	callbacks   []func(step *TutorialStep)
	mu          sync.RWMutex
}

// NewTutorialManager creates a new tutorial manager
func NewTutorialManager() *TutorialManager {
	tm := &TutorialManager{
		steps:     make(map[string]*TutorialStep),
		completed: make(map[string]bool),
		callbacks: make([]func(step *TutorialStep), 0),
	}
	tm.initializeTutorialSteps()
	return tm
}

// initializeTutorialSteps sets up the tutorial flow
func (tm *TutorialManager) initializeTutorialSteps() {
	// Step 1: Welcome
	tm.AddStep(&TutorialStep{
		ID:          "welcome",
		Title:       "Welcome to Merchant Tails!",
		Description: "Welcome to the capital city of Elm! You're about to begin your journey as a merchant.",
		Objective:   "Read the introduction",
		Hint:        "Click 'Next' to continue",
		Required:    true,
		Order:       1,
	})

	// Step 2: Shop Introduction
	tm.AddStep(&TutorialStep{
		ID:          "shop_intro",
		Title:       "Your Shop",
		Description: "This is your shop where customers will come to buy items. Keep it well-stocked!",
		Objective:   "Explore your shop interface",
		Hint:        "Click on different areas to learn more",
		Required:    true,
		Order:       2,
	})

	// Step 3: First Purchase
	tm.AddStep(&TutorialStep{
		ID:          "first_purchase",
		Title:       "Buying Items",
		Description: "Let's buy your first items from suppliers to stock your shop.",
		Objective:   "Buy 5 apples from a supplier",
		Hint:        "Go to the Market and find the fruit supplier",
		Required:    true,
		Order:       3,
	})

	// Step 4: Price Setting
	tm.AddStep(&TutorialStep{
		ID:          "price_setting",
		Title:       "Setting Prices",
		Description: "Set competitive prices for your items. Too high and customers won't buy, too low and you'll lose money!",
		Objective:   "Set a price for apples",
		Hint:        "Try setting a price 20% higher than what you paid",
		Required:    true,
		Order:       4,
	})

	// Step 5: First Sale
	tm.AddStep(&TutorialStep{
		ID:          "first_sale",
		Title:       "Making Sales",
		Description: "Customers will visit your shop. Serve them well to build reputation!",
		Objective:   "Make your first sale",
		Hint:        "Wait for a customer and click on them to serve",
		Required:    true,
		Order:       5,
	})

	// Step 6: Market Dynamics
	tm.AddStep(&TutorialStep{
		ID:          "market_dynamics",
		Title:       "Understanding the Market",
		Description: "Prices change based on supply and demand. Watch the market trends!",
		Objective:   "Check the market prices",
		Hint:        "Open the Market View to see price trends",
		Required:    false,
		Order:       6,
	})

	// Step 7: Inventory Management
	tm.AddStep(&TutorialStep{
		ID:          "inventory_management",
		Title:       "Managing Inventory",
		Description: "Balance your shop and warehouse inventory. Some items spoil over time!",
		Objective:   "Move items between shop and warehouse",
		Hint:        "Open Inventory View and transfer items",
		Required:    false,
		Order:       7,
	})

	// Step 8: Daily Cycle
	tm.AddStep(&TutorialStep{
		ID:          "daily_cycle",
		Title:       "Day and Night",
		Description: "Each day has different phases. Plan your activities accordingly!",
		Objective:   "Complete your first day",
		Hint:        "The day ends automatically, review your profits",
		Required:    true,
		Order:       8,
	})

	// Step 9: Progression
	tm.AddStep(&TutorialStep{
		ID:          "progression",
		Title:       "Growing Your Business",
		Description: "Earn experience and gold to rank up from Apprentice to Master Merchant!",
		Objective:   "View your progress",
		Hint:        "Check the progression panel",
		Required:    false,
		Order:       9,
	})

	// Step 10: Advanced Tips
	tm.AddStep(&TutorialStep{
		ID:          "advanced_tips",
		Title:       "Pro Tips",
		Description: "Watch for seasonal events, manage relationships with other merchants, and diversify your inventory!",
		Objective:   "Complete the tutorial",
		Hint:        "You're ready to begin your merchant journey!",
		Required:    false,
		Order:       10,
	})
}

// AddStep adds a tutorial step
func (tm *TutorialManager) AddStep(step *TutorialStep) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.steps[step.ID] = step
}

// StartTutorial begins the tutorial
func (tm *TutorialManager) StartTutorial() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.isActive = true
	tm.currentStep = tm.getFirstStep()
	if tm.currentStep != nil {
		tm.currentStep.Status = StatusActive
		tm.notifyCallbacks(tm.currentStep)
	}
}

// SkipTutorial skips the entire tutorial
func (tm *TutorialManager) SkipTutorial() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, step := range tm.steps {
		if !step.Required {
			step.Status = StatusSkipped
			tm.completed[step.ID] = true
		}
	}
	tm.isActive = false
}

// CompleteCurrentStep marks the current step as completed
func (tm *TutorialManager) CompleteCurrentStep() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.currentStep == nil || !tm.isActive {
		return false
	}

	// Check condition if exists
	if tm.currentStep.Condition != nil && !tm.currentStep.Condition() {
		return false
	}

	// Mark as completed
	tm.currentStep.Status = StatusCompleted
	tm.completed[tm.currentStep.ID] = true

	// Execute completion callback
	if tm.currentStep.OnComplete != nil {
		tm.currentStep.OnComplete()
	}

	// Move to next step
	nextStep := tm.getNextStep()
	if nextStep != nil {
		tm.currentStep = nextStep
		tm.currentStep.Status = StatusActive
		tm.notifyCallbacks(tm.currentStep)
	} else {
		// Tutorial completed
		tm.isActive = false
		tm.currentStep = nil
	}

	return true
}

// SkipCurrentStep skips the current step if allowed
func (tm *TutorialManager) SkipCurrentStep() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.currentStep == nil || tm.currentStep.Required {
		return false
	}

	tm.currentStep.Status = StatusSkipped
	tm.completed[tm.currentStep.ID] = true

	// Move to next step
	nextStep := tm.getNextStep()
	if nextStep != nil {
		tm.currentStep = nextStep
		tm.currentStep.Status = StatusActive
		tm.notifyCallbacks(tm.currentStep)
	} else {
		tm.isActive = false
		tm.currentStep = nil
	}

	return true
}

// GetCurrentStep returns the current tutorial step
func (tm *TutorialManager) GetCurrentStep() *TutorialStep {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentStep
}

// IsActive returns whether tutorial is active
func (tm *TutorialManager) IsActive() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.isActive
}

// GetProgress returns tutorial completion percentage
func (tm *TutorialManager) GetProgress() float64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if len(tm.steps) == 0 {
		return 0
	}

	completedCount := 0
	for _, completed := range tm.completed {
		if completed {
			completedCount++
		}
	}

	return float64(completedCount) / float64(len(tm.steps)) * 100
}

// RegisterCallback registers a callback for step changes
func (tm *TutorialManager) RegisterCallback(callback func(step *TutorialStep)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.callbacks = append(tm.callbacks, callback)
}

// Helper methods

func (tm *TutorialManager) getFirstStep() *TutorialStep {
	var firstStep *TutorialStep
	minOrder := int(^uint(0) >> 1) // Max int

	for _, step := range tm.steps {
		if step.Order < minOrder {
			minOrder = step.Order
			firstStep = step
		}
	}
	return firstStep
}

func (tm *TutorialManager) getNextStep() *TutorialStep {
	if tm.currentStep == nil {
		return nil
	}

	var nextStep *TutorialStep
	minOrder := int(^uint(0) >> 1) // Max int
	currentOrder := tm.currentStep.Order

	for _, step := range tm.steps {
		if step.Order > currentOrder && step.Order < minOrder && !tm.completed[step.ID] {
			minOrder = step.Order
			nextStep = step
		}
	}
	return nextStep
}

func (tm *TutorialManager) notifyCallbacks(step *TutorialStep) {
	for _, callback := range tm.callbacks {
		go callback(step)
	}
}

// TutorialHint provides contextual hints
type TutorialHint struct {
	ID        string
	Context   string // Where/when to show this hint
	Message   string
	Priority  int
	Shown     bool
	ShowCount int
	MaxShows  int // Maximum times to show this hint
}

// HintSystem manages contextual hints
type HintSystem struct {
	hints   map[string]*TutorialHint
	active  []*TutorialHint
	enabled bool
	mu      sync.RWMutex
}

// NewHintSystem creates a new hint system
func NewHintSystem() *HintSystem {
	hs := &HintSystem{
		hints:   make(map[string]*TutorialHint),
		active:  make([]*TutorialHint, 0),
		enabled: true,
	}
	hs.initializeHints()
	return hs
}

// initializeHints sets up default hints
func (hs *HintSystem) initializeHints() {
	// Market hints
	hs.AddHint(&TutorialHint{
		ID:       "buy_low",
		Context:  "market_view",
		Message:  "Buy items when prices are below average (green indicators)",
		Priority: 1,
		MaxShows: 3,
	})

	hs.AddHint(&TutorialHint{
		ID:       "sell_high",
		Context:  "market_view",
		Message:  "Sell items when prices are above average (red indicators)",
		Priority: 1,
		MaxShows: 3,
	})

	// Inventory hints
	hs.AddHint(&TutorialHint{
		ID:       "fruit_spoilage",
		Context:  "inventory_fruit",
		Message:  "Fruits spoil quickly! Sell them within 3 days",
		Priority: 2,
		MaxShows: 2,
	})

	hs.AddHint(&TutorialHint{
		ID:       "warehouse_usage",
		Context:  "inventory_full",
		Message:  "Use your warehouse to store items you're not selling immediately",
		Priority: 1,
		MaxShows: 2,
	})

	// Trading hints
	hs.AddHint(&TutorialHint{
		ID:       "reputation_matters",
		Context:  "customer_interaction",
		Message:  "Good reputation attracts more customers and better prices",
		Priority: 1,
		MaxShows: 2,
	})

	hs.AddHint(&TutorialHint{
		ID:       "bulk_discount",
		Context:  "bulk_purchase",
		Message:  "Buying in bulk often gets you better prices",
		Priority: 2,
		MaxShows: 1,
	})

	// Seasonal hints
	hs.AddHint(&TutorialHint{
		ID:       "seasonal_items",
		Context:  "season_change",
		Message:  "Some items sell better in certain seasons",
		Priority: 1,
		MaxShows: 4,
	})

	// Event hints
	hs.AddHint(&TutorialHint{
		ID:       "dragon_event",
		Context:  "event_dragon",
		Message:  "Dragon attacks reduce supply - prices will rise!",
		Priority: 3,
		MaxShows: 1,
	})

	hs.AddHint(&TutorialHint{
		ID:       "festival_event",
		Context:  "event_festival",
		Message:  "Festivals increase demand - stock up beforehand!",
		Priority: 3,
		MaxShows: 1,
	})
}

// AddHint adds a new hint
func (hs *HintSystem) AddHint(hint *TutorialHint) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.hints[hint.ID] = hint
}

// GetHint returns a hint for the given context
func (hs *HintSystem) GetHint(context string) *TutorialHint {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if !hs.enabled {
		return nil
	}

	var bestHint *TutorialHint
	highestPriority := -1

	// Sort hints by ID for consistent ordering
	var sortedHints []*TutorialHint
	for _, hint := range hs.hints {
		sortedHints = append(sortedHints, hint)
	}
	// Sort by ID to ensure consistent order
	for i := 0; i < len(sortedHints)-1; i++ {
		for j := i + 1; j < len(sortedHints); j++ {
			if sortedHints[i].ID > sortedHints[j].ID {
				sortedHints[i], sortedHints[j] = sortedHints[j], sortedHints[i]
			}
		}
	}

	for _, hint := range sortedHints {
		if hint.Context == context &&
			!hint.Shown &&
			hint.ShowCount < hint.MaxShows &&
			hint.Priority > highestPriority {
			bestHint = hint
			highestPriority = hint.Priority
		}
	}

	if bestHint != nil {
		bestHint.Shown = true
		bestHint.ShowCount++
		hs.active = append(hs.active, bestHint)

		// Reset shown flag after some time
		go func() {
			time.Sleep(30 * time.Second)
			bestHint.Shown = false
		}()
	}

	return bestHint
}

// GetActiveHints returns all currently active hints
func (hs *HintSystem) GetActiveHints() []*TutorialHint {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.active
}

// ClearActiveHints clears all active hints
func (hs *HintSystem) ClearActiveHints() {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.active = make([]*TutorialHint, 0)
}

// SetEnabled enables or disables the hint system
func (hs *HintSystem) SetEnabled(enabled bool) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.enabled = enabled
}

// ResetHints resets all hint show counts
func (hs *HintSystem) ResetHints() {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	for _, hint := range hs.hints {
		hint.Shown = false
		hint.ShowCount = 0
	}
	hs.active = make([]*TutorialHint, 0)
}
