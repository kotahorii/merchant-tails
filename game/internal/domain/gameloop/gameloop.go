// Package gameloop provides the core game loop system for the merchant game
package gameloop

import (
	"context"
	"time"
)

// Phase represents the different phases of a game day
type Phase int

const (
	PhaseMorning Phase = iota
	PhaseNoon
	PhaseEvening
	PhaseNight
)

// GameLoop defines the interface for the main game loop
type GameLoop interface {
	// Start begins the game loop
	Start(ctx context.Context) error

	// Stop gracefully stops the game loop
	Stop() error

	// Update is called each frame to update game state
	Update(deltaTime time.Duration) error

	// GetCurrentPhase returns the current phase of the day
	GetCurrentPhase() Phase

	// GetElapsedTime returns the total elapsed game time
	GetElapsedTime() time.Duration

	// SetSpeed sets the game speed multiplier (1.0 = normal, 2.0 = double speed)
	SetSpeed(multiplier float64)

	// Pause pauses the game loop
	Pause()

	// Resume resumes the game loop
	Resume()

	// IsPaused returns whether the game is paused
	IsPaused() bool

	// RegisterUpdateCallback registers a callback to be called on each update
	RegisterUpdateCallback(callback UpdateCallback)

	// RegisterPhaseChangeCallback registers a callback for phase changes
	RegisterPhaseChangeCallback(callback PhaseChangeCallback)
}

// UpdateCallback is called on each game update
type UpdateCallback func(deltaTime time.Duration) error

// PhaseChangeCallback is called when the game phase changes
type PhaseChangeCallback func(oldPhase, newPhase Phase)

// Config contains configuration for the game loop
type Config struct {
	// TargetFPS is the target frames per second
	TargetFPS int

	// DayDuration is the real-time duration of one game day
	DayDuration time.Duration

	// PhaseDistribution defines the proportion of day for each phase
	PhaseDistribution map[Phase]float64

	// AutoStart determines if the loop starts automatically
	AutoStart bool
}

// DefaultConfig returns a default game loop configuration
func DefaultConfig() *Config {
	return &Config{
		TargetFPS:   60,
		DayDuration: 10 * time.Minute, // 10 minutes per game day
		PhaseDistribution: map[Phase]float64{
			PhaseMorning: 0.25, // 25% of day
			PhaseNoon:    0.25, // 25% of day
			PhaseEvening: 0.25, // 25% of day
			PhaseNight:   0.25, // 25% of day
		},
		AutoStart: false,
	}
}

// GetPhaseName returns the string name of a phase
func GetPhaseName(phase Phase) string {
	switch phase {
	case PhaseMorning:
		return "Morning"
	case PhaseNoon:
		return "Noon"
	case PhaseEvening:
		return "Evening"
	case PhaseNight:
		return "Night"
	default:
		return "Unknown"
	}
}

// GetPhaseFromTime calculates the phase based on elapsed time
func GetPhaseFromTime(elapsed time.Duration, dayDuration time.Duration, distribution map[Phase]float64) Phase {
	dayProgress := float64(elapsed%dayDuration) / float64(dayDuration)

	threshold := 0.0
	for phase := PhaseMorning; phase <= PhaseNight; phase++ {
		threshold += distribution[phase]
		if dayProgress < threshold {
			return phase
		}
	}

	return PhaseNight
}
