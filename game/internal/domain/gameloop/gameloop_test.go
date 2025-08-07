package gameloop

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStandardGameLoop(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		gl := NewStandardGameLoop(nil)
		assert.NotNil(t, gl)
		assert.Equal(t, PhaseMorning, gl.currentPhase)
		assert.Equal(t, 1.0, gl.speedMultiplier)
		assert.False(t, gl.paused)
		assert.False(t, gl.running)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			TargetFPS:   30,
			DayDuration: 5 * time.Minute,
			PhaseDistribution: map[Phase]float64{
				PhaseMorning: 0.3,
				PhaseNoon:    0.2,
				PhaseEvening: 0.3,
				PhaseNight:   0.2,
			},
			AutoStart: false,
		}
		gl := NewStandardGameLoop(config)
		assert.NotNil(t, gl)
		assert.Equal(t, config, gl.config)
	})
}

func TestGameLoopStartStop(t *testing.T) {
	gl := NewStandardGameLoop(nil)
	ctx := context.Background()

	// Test start
	err := gl.Start(ctx)
	require.NoError(t, err)
	assert.True(t, gl.running)

	// Test double start
	err = gl.Start(ctx)
	assert.Error(t, err)

	// Test stop
	err = gl.Stop()
	require.NoError(t, err)
	assert.False(t, gl.running)

	// Test double stop
	err = gl.Stop()
	assert.Error(t, err)
}

func TestGameLoopPauseResume(t *testing.T) {
	gl := NewStandardGameLoop(nil)

	assert.False(t, gl.IsPaused())

	gl.Pause()
	assert.True(t, gl.IsPaused())

	gl.Resume()
	assert.False(t, gl.IsPaused())
}

func TestGameLoopSetSpeed(t *testing.T) {
	gl := NewStandardGameLoop(nil)

	// Normal speed
	gl.SetSpeed(1.0)
	assert.Equal(t, 1.0, gl.speedMultiplier)

	// Double speed
	gl.SetSpeed(2.0)
	assert.Equal(t, 2.0, gl.speedMultiplier)

	// Test clamping
	gl.SetSpeed(0.05)
	assert.Equal(t, 0.1, gl.speedMultiplier)

	gl.SetSpeed(15.0)
	assert.Equal(t, 10.0, gl.speedMultiplier)
}

func TestGetPhaseName(t *testing.T) {
	tests := []struct {
		phase    Phase
		expected string
	}{
		{PhaseMorning, "Morning"},
		{PhaseNoon, "Noon"},
		{PhaseEvening, "Evening"},
		{PhaseNight, "Night"},
		{Phase(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetPhaseName(tt.phase))
		})
	}
}

func TestGetPhaseFromTime(t *testing.T) {
	dayDuration := 4 * time.Hour
	distribution := map[Phase]float64{
		PhaseMorning: 0.25,
		PhaseNoon:    0.25,
		PhaseEvening: 0.25,
		PhaseNight:   0.25,
	}

	tests := []struct {
		elapsed  time.Duration
		expected Phase
	}{
		{0 * time.Hour, PhaseMorning},
		{1 * time.Hour, PhaseNoon},
		{2 * time.Hour, PhaseEvening},
		{3 * time.Hour, PhaseNight},
		{4 * time.Hour, PhaseMorning}, // New day
		{5 * time.Hour, PhaseNoon},
	}

	for _, tt := range tests {
		t.Run(tt.elapsed.String(), func(t *testing.T) {
			phase := GetPhaseFromTime(tt.elapsed, dayDuration, distribution)
			assert.Equal(t, tt.expected, phase)
		})
	}
}

func TestGameLoopUpdate(t *testing.T) {
	config := &Config{
		TargetFPS:   60,
		DayDuration: 100 * time.Millisecond,
		PhaseDistribution: map[Phase]float64{
			PhaseMorning: 0.25,
			PhaseNoon:    0.25,
			PhaseEvening: 0.25,
			PhaseNight:   0.25,
		},
		AutoStart: false,
	}
	gl := NewStandardGameLoop(config)

	// Test update callback
	updateCalled := false
	gl.RegisterUpdateCallback(func(deltaTime time.Duration) error {
		updateCalled = true
		return nil
	})

	// Test phase change callback
	phaseChangeCalled := false
	var oldPhaseReceived, newPhaseReceived Phase
	gl.RegisterPhaseChangeCallback(func(oldPhase, newPhase Phase) {
		phaseChangeCalled = true
		oldPhaseReceived = oldPhase
		newPhaseReceived = newPhase
	})

	// Initial state
	assert.Equal(t, PhaseMorning, gl.GetCurrentPhase())
	assert.Equal(t, time.Duration(0), gl.GetElapsedTime())

	// Update to advance time
	err := gl.Update(10 * time.Millisecond)
	require.NoError(t, err)
	assert.True(t, updateCalled)
	assert.Equal(t, 10*time.Millisecond, gl.GetElapsedTime())

	// Update to trigger phase change
	err = gl.Update(20 * time.Millisecond) // Total: 30ms, should be in PhaseNoon
	require.NoError(t, err)
	assert.True(t, phaseChangeCalled)
	assert.Equal(t, PhaseMorning, oldPhaseReceived)
	assert.Equal(t, PhaseNoon, newPhaseReceived)
	assert.Equal(t, PhaseNoon, gl.GetCurrentPhase())
}

func TestGameLoopGetCurrentDay(t *testing.T) {
	config := &Config{
		TargetFPS:   60,
		DayDuration: 1 * time.Hour,
		PhaseDistribution: map[Phase]float64{
			PhaseMorning: 0.25,
			PhaseNoon:    0.25,
			PhaseEvening: 0.25,
			PhaseNight:   0.25,
		},
		AutoStart: false,
	}
	gl := NewStandardGameLoop(config)

	// Day 1
	assert.Equal(t, 1, gl.GetCurrentDay())

	// Advance to day 2
	gl.elapsedTime = 1 * time.Hour
	assert.Equal(t, 2, gl.GetCurrentDay())

	// Advance to day 3
	gl.elapsedTime = 2 * time.Hour
	assert.Equal(t, 3, gl.GetCurrentDay())
}

func TestGameLoopIntegration(t *testing.T) {
	config := &Config{
		TargetFPS:   60,
		DayDuration: 100 * time.Millisecond,
		PhaseDistribution: map[Phase]float64{
			PhaseMorning: 0.25,
			PhaseNoon:    0.25,
			PhaseEvening: 0.25,
			PhaseNight:   0.25,
		},
		AutoStart: false,
	}
	gl := NewStandardGameLoop(config)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	phaseChanges := 0
	gl.RegisterPhaseChangeCallback(func(oldPhase, newPhase Phase) {
		phaseChanges++
		t.Logf("Phase changed from %s to %s", GetPhaseName(oldPhase), GetPhaseName(newPhase))
	})

	err := gl.Start(ctx)
	require.NoError(t, err)

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	err = gl.Stop()
	require.NoError(t, err)

	// Should have gone through at least some phase changes
	assert.Greater(t, phaseChanges, 0)
	assert.Greater(t, gl.GetElapsedTime(), time.Duration(0))
}
