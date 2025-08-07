package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/merchant-tails/game/internal/domain/gameloop"
)

func TestNewStandardTimeManager(t *testing.T) {
	tm := NewStandardTimeManager(nil, 1*time.Hour)
	assert.NotNil(t, tm)

	// Check initial state
	gameTime := tm.GetCurrentTime()
	assert.Equal(t, 1, gameTime.Year)
	assert.Equal(t, Spring, gameTime.Season)
	assert.Equal(t, 1, gameTime.Day)
	assert.Equal(t, gameloop.PhaseMorning, gameTime.Phase)
	assert.Equal(t, 1.0, tm.GetGameSpeed())
}

func TestTimeManagerAdvanceDay(t *testing.T) {
	tm := NewStandardTimeManager(nil, 1*time.Hour)

	// Advance within season
	tm.AdvanceDay()
	gameTime := tm.GetCurrentTime()
	assert.Equal(t, 2, gameTime.Day)
	assert.Equal(t, Spring, gameTime.Season)
	assert.Equal(t, 1, gameTime.Year)

	// Advance to next season
	for i := 0; i < 29; i++ {
		tm.AdvanceDay()
	}
	gameTime = tm.GetCurrentTime()
	assert.Equal(t, 1, gameTime.Day)
	assert.Equal(t, Summer, gameTime.Season)
	assert.Equal(t, 1, gameTime.Year)

	// Advance to next year
	for i := 0; i < 90; i++ { // 3 more seasons
		tm.AdvanceDay()
	}
	gameTime = tm.GetCurrentTime()
	assert.Equal(t, 1, gameTime.Day)
	assert.Equal(t, Spring, gameTime.Season)
	assert.Equal(t, 2, gameTime.Year)
}

func TestTimeManagerGetDayOfYear(t *testing.T) {
	tm := NewStandardTimeManager(nil, 1*time.Hour)

	// Spring Day 1
	assert.Equal(t, 1, tm.GetDayOfYear())

	// Spring Day 30
	for i := 0; i < 29; i++ {
		tm.AdvanceDay()
	}
	assert.Equal(t, 30, tm.GetDayOfYear())

	// Summer Day 1
	tm.AdvanceDay()
	assert.Equal(t, 31, tm.GetDayOfYear())

	// Winter Day 30 (last day of year)
	for i := 0; i < 89; i++ {
		tm.AdvanceDay()
	}
	assert.Equal(t, 120, tm.GetDayOfYear())
}

func TestTimeManagerSeasonChecks(t *testing.T) {
	tm := NewStandardTimeManager(nil, 1*time.Hour)

	// New year and new season
	assert.True(t, tm.IsNewYear())
	assert.True(t, tm.IsNewSeason())

	// Not new year or season
	tm.AdvanceDay()
	assert.False(t, tm.IsNewYear())
	assert.False(t, tm.IsNewSeason())

	// New season but not new year
	for i := 0; i < 29; i++ {
		tm.AdvanceDay()
	}
	assert.True(t, tm.IsNewSeason())
	assert.False(t, tm.IsNewYear())
}

func TestTimeManagerCallbacks(t *testing.T) {
	tm := NewStandardTimeManager(nil, 1*time.Hour)

	// Time change callback
	timeChangeCalled := false
	var receivedTime GameTime
	tm.RegisterTimeChangeCallback(func(newTime GameTime) {
		timeChangeCalled = true
		receivedTime = newTime
	})

	// Season change callback
	seasonChangeCalled := false
	var oldSeasonReceived, newSeasonReceived Season
	tm.RegisterSeasonChangeCallback(func(oldSeason, newSeason Season) {
		seasonChangeCalled = true
		oldSeasonReceived = oldSeason
		newSeasonReceived = newSeason
	})

	// Advance day should trigger time change
	tm.AdvanceDay()
	assert.True(t, timeChangeCalled)
	assert.Equal(t, 2, receivedTime.Day)

	// Advance to next season
	for i := 0; i < 29; i++ {
		tm.AdvanceDay()
	}
	assert.True(t, seasonChangeCalled)
	assert.Equal(t, Spring, oldSeasonReceived)
	assert.Equal(t, Summer, newSeasonReceived)
}

func TestTimeManagerUpdate(t *testing.T) {
	tm := NewStandardTimeManager(nil, 100*time.Millisecond) // 100ms per day
	tm.Start()

	// Update with half a day
	tm.Update(50 * time.Millisecond)
	assert.Equal(t, 1, tm.GetCurrentTime().Day)
	assert.Equal(t, 50*time.Millisecond, tm.GetRealTime())

	// Update to complete a day
	tm.Update(50 * time.Millisecond)
	assert.Equal(t, 2, tm.GetCurrentTime().Day)
	assert.Equal(t, 100*time.Millisecond, tm.GetRealTime())

	// Update multiple days at once
	tm.Update(300 * time.Millisecond)
	assert.Equal(t, 5, tm.GetCurrentTime().Day)
	assert.Equal(t, 400*time.Millisecond, tm.GetRealTime())
}

func TestTimeManagerStartStop(t *testing.T) {
	tm := NewStandardTimeManager(nil, 1*time.Hour)

	// Initially not running
	tm.Update(1 * time.Hour)
	assert.Equal(t, 1, tm.GetCurrentTime().Day) // Should not advance

	// Start and update
	tm.Start()
	tm.Update(1 * time.Hour)
	assert.Equal(t, 2, tm.GetCurrentTime().Day) // Should advance

	// Stop and update
	tm.Stop()
	tm.Update(1 * time.Hour)
	assert.Equal(t, 2, tm.GetCurrentTime().Day) // Should not advance
}

func TestTimeManagerGameSpeed(t *testing.T) {
	tm := NewStandardTimeManager(nil, 1*time.Hour)

	// Default speed
	assert.Equal(t, 1.0, tm.GetGameSpeed())

	// Set double speed
	tm.SetGameSpeed(2.0)
	assert.Equal(t, 2.0, tm.GetGameSpeed())
}

func TestGetSeasonName(t *testing.T) {
	tests := []struct {
		season   Season
		expected string
	}{
		{Spring, "Spring"},
		{Summer, "Summer"},
		{Autumn, "Autumn"},
		{Winter, "Winter"},
		{Season(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetSeasonName(tt.season))
		})
	}
}

func TestFormatGameTime(t *testing.T) {
	gt := GameTime{
		Year:   2,
		Season: Summer,
		Day:    15,
		Phase:  gameloop.PhaseNoon,
	}

	formatted := FormatGameTime(gt)
	assert.Equal(t, "Year 2, Summer Day 15, Noon", formatted)
}

func TestTimeManagerWithGameLoop(t *testing.T) {
	// Create a game loop
	config := &gameloop.Config{
		TargetFPS:   60,
		DayDuration: 100 * time.Millisecond,
		PhaseDistribution: map[gameloop.Phase]float64{
			gameloop.PhaseMorning: 0.25,
			gameloop.PhaseNoon:    0.25,
			gameloop.PhaseEvening: 0.25,
			gameloop.PhaseNight:   0.25,
		},
		AutoStart: false,
	}
	gl := gameloop.NewStandardGameLoop(config)

	// Create time manager with game loop
	tm := NewStandardTimeManager(gl, 100*time.Millisecond)
	tm.Start()

	// Phase change should update time manager
	phaseChangeCalled := false
	tm.RegisterTimeChangeCallback(func(newTime GameTime) {
		if newTime.Phase == gameloop.PhaseNoon {
			phaseChangeCalled = true
		}
	})

	// Simulate phase change
	gl.Update(30 * time.Millisecond) // Should trigger phase change to Noon
	assert.True(t, phaseChangeCalled)
	assert.Equal(t, gameloop.PhaseNoon, tm.GetCurrentTime().Phase)

	// Speed change should affect game loop
	tm.SetGameSpeed(2.0)
	// This would be reflected in the game loop's speed multiplier
}

func TestTimeManagerIntegration(t *testing.T) {
	tm := NewStandardTimeManager(nil, 10*time.Millisecond) // Fast time for testing
	tm.Start()

	seasonChanges := 0
	tm.RegisterSeasonChangeCallback(func(oldSeason, newSeason Season) {
		seasonChanges++
		t.Logf("Season changed from %s to %s", GetSeasonName(oldSeason), GetSeasonName(newSeason))
	})

	dayChanges := 0
	lastDay := 1
	tm.RegisterTimeChangeCallback(func(newTime GameTime) {
		if newTime.Day != lastDay {
			dayChanges++
			lastDay = newTime.Day
			t.Logf("Day changed to %d", newTime.Day)
		}
	})

	// Simulate a long period
	tm.Update(350 * time.Millisecond) // 35 days

	// Should have gone through at least one season
	assert.Greater(t, seasonChanges, 0)
	assert.Greater(t, dayChanges, 30) // Should have passed at least 30 days

	currentTime := tm.GetCurrentTime()
	assert.Equal(t, Summer, currentTime.Season)
	assert.Greater(t, currentTime.Day, 0)
}
