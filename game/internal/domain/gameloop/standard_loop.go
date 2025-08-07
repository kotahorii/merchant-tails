package gameloop

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StandardGameLoop implements the GameLoop interface
type StandardGameLoop struct {
	config               *Config
	currentPhase         Phase
	elapsedTime          time.Duration
	lastUpdateTime       time.Time
	speedMultiplier      float64
	paused               bool
	running              bool
	updateCallbacks      []UpdateCallback
	phaseChangeCallbacks []PhaseChangeCallback
	mu                   sync.RWMutex
	stopChan             chan struct{}
	ctx                  context.Context
	cancel               context.CancelFunc
}

// NewStandardGameLoop creates a new standard game loop
func NewStandardGameLoop(config *Config) *StandardGameLoop {
	if config == nil {
		config = DefaultConfig()
	}

	return &StandardGameLoop{
		config:               config,
		currentPhase:         PhaseMorning,
		speedMultiplier:      1.0,
		paused:               false,
		running:              false,
		updateCallbacks:      make([]UpdateCallback, 0),
		phaseChangeCallbacks: make([]PhaseChangeCallback, 0),
		stopChan:             make(chan struct{}),
	}
}

// Start begins the game loop
func (gl *StandardGameLoop) Start(ctx context.Context) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if gl.running {
		return fmt.Errorf("game loop is already running")
	}

	gl.ctx, gl.cancel = context.WithCancel(ctx)
	gl.running = true
	gl.lastUpdateTime = time.Now()

	go gl.run()

	return nil
}

// Stop gracefully stops the game loop
func (gl *StandardGameLoop) Stop() error {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if !gl.running {
		return fmt.Errorf("game loop is not running")
	}

	if gl.cancel != nil {
		gl.cancel()
	}

	close(gl.stopChan)
	gl.running = false

	return nil
}

// run is the main loop goroutine
func (gl *StandardGameLoop) run() {
	ticker := time.NewTicker(time.Second / time.Duration(gl.config.TargetFPS))
	defer ticker.Stop()

	for {
		select {
		case <-gl.ctx.Done():
			return
		case <-gl.stopChan:
			return
		case <-ticker.C:
			if !gl.paused {
				now := time.Now()
				deltaTime := now.Sub(gl.lastUpdateTime)
				gl.lastUpdateTime = now

				if err := gl.Update(deltaTime); err != nil {
					// Log error but continue running
					fmt.Printf("Game loop update error: %v\n", err)
				}
			}
		}
	}
}

// Update is called each frame to update game state
func (gl *StandardGameLoop) Update(deltaTime time.Duration) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if gl.paused {
		return nil
	}

	// Apply speed multiplier
	adjustedDelta := time.Duration(float64(deltaTime) * gl.speedMultiplier)
	oldPhase := gl.currentPhase

	// Update elapsed time
	gl.elapsedTime += adjustedDelta

	// Calculate current phase
	newPhase := GetPhaseFromTime(gl.elapsedTime, gl.config.DayDuration, gl.config.PhaseDistribution)

	// Check for phase change
	if newPhase != oldPhase {
		gl.currentPhase = newPhase
		gl.notifyPhaseChange(oldPhase, newPhase)
	}

	// Call update callbacks
	for _, callback := range gl.updateCallbacks {
		if err := callback(adjustedDelta); err != nil {
			return fmt.Errorf("update callback error: %w", err)
		}
	}

	return nil
}

// GetCurrentPhase returns the current phase of the day
func (gl *StandardGameLoop) GetCurrentPhase() Phase {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.currentPhase
}

// GetElapsedTime returns the total elapsed game time
func (gl *StandardGameLoop) GetElapsedTime() time.Duration {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.elapsedTime
}

// SetSpeed sets the game speed multiplier
func (gl *StandardGameLoop) SetSpeed(multiplier float64) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if multiplier < 0.1 {
		multiplier = 0.1
	} else if multiplier > 10.0 {
		multiplier = 10.0
	}

	gl.speedMultiplier = multiplier
}

// Pause pauses the game loop
func (gl *StandardGameLoop) Pause() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.paused = true
}

// Resume resumes the game loop
func (gl *StandardGameLoop) Resume() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.paused = false
	gl.lastUpdateTime = time.Now() // Reset to avoid large delta
}

// IsPaused returns whether the game is paused
func (gl *StandardGameLoop) IsPaused() bool {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.paused
}

// RegisterUpdateCallback registers a callback to be called on each update
func (gl *StandardGameLoop) RegisterUpdateCallback(callback UpdateCallback) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.updateCallbacks = append(gl.updateCallbacks, callback)
}

// RegisterPhaseChangeCallback registers a callback for phase changes
func (gl *StandardGameLoop) RegisterPhaseChangeCallback(callback PhaseChangeCallback) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.phaseChangeCallbacks = append(gl.phaseChangeCallbacks, callback)
}

// notifyPhaseChange notifies all phase change callbacks
func (gl *StandardGameLoop) notifyPhaseChange(oldPhase, newPhase Phase) {
	for _, callback := range gl.phaseChangeCallbacks {
		callback(oldPhase, newPhase)
	}
}

// GetCurrentDay returns the current game day number
func (gl *StandardGameLoop) GetCurrentDay() int {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return int(gl.elapsedTime/gl.config.DayDuration) + 1
}

// GetTimeInCurrentPhase returns how long we've been in the current phase
func (gl *StandardGameLoop) GetTimeInCurrentPhase() time.Duration {
	gl.mu.RLock()
	defer gl.mu.RUnlock()

	dayProgress := gl.elapsedTime % gl.config.DayDuration
	phaseStart := 0.0

	for phase := PhaseMorning; phase < gl.currentPhase; phase++ {
		phaseStart += gl.config.PhaseDistribution[phase]
	}

	phaseStartTime := time.Duration(phaseStart * float64(gl.config.DayDuration))
	return dayProgress - phaseStartTime
}
