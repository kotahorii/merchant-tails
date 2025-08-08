package tutorial

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTutorialManager(t *testing.T) {
	tm := NewTutorialManager()

	assert.NotNil(t, tm)
	assert.NotNil(t, tm.steps)
	assert.NotNil(t, tm.completed)
	assert.False(t, tm.isActive)
	assert.Len(t, tm.steps, 10) // Should have 10 default steps
}

func TestTutorialManager_StartTutorial(t *testing.T) {
	tm := NewTutorialManager()

	// Register callback to track step changes
	var calledStep *TutorialStep
	tm.RegisterCallback(func(step *TutorialStep) {
		calledStep = step
	})

	tm.StartTutorial()

	assert.True(t, tm.IsActive())
	assert.NotNil(t, tm.GetCurrentStep())
	assert.Equal(t, "welcome", tm.GetCurrentStep().ID)
	assert.Equal(t, StatusActive, tm.GetCurrentStep().Status)

	// Wait for callback
	time.Sleep(10 * time.Millisecond)
	assert.NotNil(t, calledStep)
	assert.Equal(t, "welcome", calledStep.ID)
}

func TestTutorialManager_CompleteCurrentStep(t *testing.T) {
	tm := NewTutorialManager()
	tm.StartTutorial()

	// Complete first step
	success := tm.CompleteCurrentStep()
	assert.True(t, success)

	// Should move to second step
	currentStep := tm.GetCurrentStep()
	assert.NotNil(t, currentStep)
	assert.Equal(t, "shop_intro", currentStep.ID)
	assert.Equal(t, StatusActive, currentStep.Status)

	// First step should be marked complete
	assert.True(t, tm.completed["welcome"])
}

func TestTutorialManager_SkipCurrentStep(t *testing.T) {
	tm := NewTutorialManager()
	tm.StartTutorial()

	// First step is required, cannot skip
	success := tm.SkipCurrentStep()
	assert.False(t, success)
	assert.Equal(t, "welcome", tm.GetCurrentStep().ID)

	// Complete required steps until we reach an optional one
	for range 5 {
		tm.CompleteCurrentStep()
	}

	// Now on "market_dynamics" which is optional
	currentStep := tm.GetCurrentStep()
	assert.Equal(t, "market_dynamics", currentStep.ID)
	assert.False(t, currentStep.Required)

	// Should be able to skip
	success = tm.SkipCurrentStep()
	assert.True(t, success)
	assert.Equal(t, "inventory_management", tm.GetCurrentStep().ID)
}

func TestTutorialManager_GetProgress(t *testing.T) {
	tm := NewTutorialManager()

	// Initially 0%
	assert.Equal(t, float64(0), tm.GetProgress())

	tm.StartTutorial()
	tm.CompleteCurrentStep() // Complete welcome

	// Should be 10% (1 of 10 steps)
	assert.Equal(t, float64(10), tm.GetProgress())

	// Complete more steps
	for range 4 {
		tm.CompleteCurrentStep()
	}

	// Should be 50% (5 of 10 steps)
	assert.Equal(t, float64(50), tm.GetProgress())
}

func TestTutorialManager_SkipTutorial(t *testing.T) {
	tm := NewTutorialManager()
	tm.StartTutorial()

	tm.SkipTutorial()

	assert.False(t, tm.IsActive())

	// Required steps should not be marked as completed
	// Optional steps should be marked as skipped
	for _, step := range tm.steps {
		if !step.Required {
			assert.Equal(t, StatusSkipped, step.Status)
			assert.True(t, tm.completed[step.ID])
		}
	}
}

func TestTutorialManager_StepWithCondition(t *testing.T) {
	tm := NewTutorialManager()

	conditionMet := false
	onCompleteCalled := false

	// Add a step with condition
	step := &TutorialStep{
		ID:          "conditional_step",
		Title:       "Conditional Step",
		Description: "This step has a condition",
		Order:       0, // Make it first
		Condition: func() bool {
			return conditionMet
		},
		OnComplete: func() {
			onCompleteCalled = true
		},
	}

	tm.AddStep(step)
	tm.currentStep = step
	tm.isActive = true

	// Try to complete when condition not met
	success := tm.CompleteCurrentStep()
	assert.False(t, success)
	assert.False(t, onCompleteCalled)

	// Set condition to true and try again
	conditionMet = true
	success = tm.CompleteCurrentStep()
	assert.True(t, success)
	assert.True(t, onCompleteCalled)
}

func TestTutorialManager_ConcurrentAccess(t *testing.T) {
	tm := NewTutorialManager()
	tm.StartTutorial()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	wg.Add(iterations)
	for range iterations {
		go func() {
			defer wg.Done()
			_ = tm.GetCurrentStep()
			_ = tm.IsActive()
			_ = tm.GetProgress()
		}()
	}

	// Concurrent writes
	wg.Add(iterations)
	for range iterations {
		go func() {
			defer wg.Done()
			tm.CompleteCurrentStep()
		}()
	}

	wg.Wait()

	// Should not panic and should be in a valid state
	assert.NotNil(t, tm)
}

func TestNewHintSystem(t *testing.T) {
	hs := NewHintSystem()

	assert.NotNil(t, hs)
	assert.NotNil(t, hs.hints)
	assert.NotNil(t, hs.active)
	assert.True(t, hs.enabled)
	assert.Greater(t, len(hs.hints), 0) // Should have default hints
}

func TestHintSystem_GetHint(t *testing.T) {
	hs := NewHintSystem()

	// Get a market hint
	hint := hs.GetHint("market_view")
	assert.NotNil(t, hint)
	assert.Equal(t, "buy_low", hint.ID)
	assert.True(t, hint.Shown)
	assert.Equal(t, 1, hint.ShowCount)

	// Should be in active hints
	activeHints := hs.GetActiveHints()
	assert.Len(t, activeHints, 1)
	assert.Equal(t, hint.ID, activeHints[0].ID)

	// Getting same context again should get different hint (buy_low is shown)
	hint2 := hs.GetHint("market_view")
	assert.NotNil(t, hint2)
	assert.NotEqual(t, hint.ID, hint2.ID)
	assert.Equal(t, "sell_high", hint2.ID)
}

func TestHintSystem_MaxShows(t *testing.T) {
	hs := NewHintSystem()

	// Add a hint with MaxShows = 1
	testHint := &TutorialHint{
		ID:       "test_hint",
		Context:  "test_context",
		Message:  "Test message",
		Priority: 10,
		MaxShows: 1,
	}
	hs.AddHint(testHint)

	// First get should work
	hint := hs.GetHint("test_context")
	assert.NotNil(t, hint)
	assert.Equal(t, "test_hint", hint.ID)

	// Wait for shown flag to reset
	time.Sleep(35 * time.Second)

	// Second get should not return the hint (exceeded MaxShows)
	hint = hs.GetHint("test_context")
	assert.Nil(t, hint)
}

func TestHintSystem_Priority(t *testing.T) {
	hs := NewHintSystem()

	// Add hints with different priorities
	hs.AddHint(&TutorialHint{
		ID:       "low_priority",
		Context:  "test",
		Priority: 1,
		MaxShows: 10,
	})

	hs.AddHint(&TutorialHint{
		ID:       "high_priority",
		Context:  "test",
		Priority: 5,
		MaxShows: 10,
	})

	hs.AddHint(&TutorialHint{
		ID:       "medium_priority",
		Context:  "test",
		Priority: 3,
		MaxShows: 10,
	})

	// Should get highest priority first
	hint := hs.GetHint("test")
	assert.NotNil(t, hint)
	assert.Equal(t, "high_priority", hint.ID)
}

func TestHintSystem_SetEnabled(t *testing.T) {
	hs := NewHintSystem()

	// Disable hint system
	hs.SetEnabled(false)

	hint := hs.GetHint("market_view")
	assert.Nil(t, hint)

	// Re-enable
	hs.SetEnabled(true)

	hint = hs.GetHint("market_view")
	assert.NotNil(t, hint)
}

func TestHintSystem_ClearActiveHints(t *testing.T) {
	hs := NewHintSystem()

	// Get some hints
	hs.GetHint("market_view")
	hs.GetHint("inventory_fruit")

	activeHints := hs.GetActiveHints()
	assert.Len(t, activeHints, 2)

	// Clear active hints
	hs.ClearActiveHints()

	activeHints = hs.GetActiveHints()
	assert.Len(t, activeHints, 0)
}

func TestHintSystem_ResetHints(t *testing.T) {
	hs := NewHintSystem()

	// Get and use up a hint
	hint := hs.GetHint("bulk_purchase")
	assert.NotNil(t, hint)
	assert.Equal(t, 1, hint.ShowCount)

	// Reset all hints
	hs.ResetHints()

	// Check that hint is reset
	for _, h := range hs.hints {
		assert.False(t, h.Shown)
		assert.Equal(t, 0, h.ShowCount)
	}

	// Should be able to get the hint again
	hint = hs.GetHint("bulk_purchase")
	assert.NotNil(t, hint)
}

func TestHintSystem_ConcurrentAccess(t *testing.T) {
	hs := NewHintSystem()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent hint requests
	wg.Add(iterations)
	for i := range iterations {
		go func(index int) {
			defer wg.Done()
			contexts := []string{"market_view", "inventory_fruit", "bulk_purchase"}
			context := contexts[index%len(contexts)]
			_ = hs.GetHint(context)
		}(i)
	}

	// Concurrent operations
	wg.Add(3)
	go func() {
		defer wg.Done()
		for range 10 {
			hs.GetActiveHints()
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for range 5 {
			hs.ClearActiveHints()
			time.Sleep(2 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for i := range 3 {
			hs.SetEnabled(i%2 == 0)
			time.Sleep(3 * time.Millisecond)
		}
	}()

	wg.Wait()

	// Should not panic and should be in a valid state
	assert.NotNil(t, hs)
}
