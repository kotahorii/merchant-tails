// Package time provides time management for the game
package time

import (
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/gameloop"
)

// GameTime represents the in-game time
type GameTime struct {
	Year   int
	Season Season
	Day    int // Day within the season
	Phase  gameloop.Phase
}

// Season represents the four seasons
type Season int

const (
	Spring Season = iota
	Summer
	Autumn
	Winter
)

const (
	DaysPerSeason = 30
	DaysPerYear   = DaysPerSeason * 4
)

// TimeManager manages game time and calendar
type TimeManager interface {
	// GetCurrentTime returns the current game time
	GetCurrentTime() GameTime

	// GetRealTime returns the real elapsed time
	GetRealTime() time.Duration

	// GetGameSpeed returns the current game speed multiplier
	GetGameSpeed() float64

	// SetGameSpeed sets the game speed multiplier
	SetGameSpeed(multiplier float64)

	// AdvanceDay advances the game by one day
	AdvanceDay()

	// GetDayOfYear returns the current day of the year (1-120)
	GetDayOfYear() int

	// GetSeasonName returns the name of the current season
	GetSeasonName() string

	// IsNewSeason returns true if it's the first day of a new season
	IsNewSeason() bool

	// IsNewYear returns true if it's the first day of a new year
	IsNewYear() bool

	// RegisterTimeChangeCallback registers a callback for time changes
	RegisterTimeChangeCallback(callback TimeChangeCallback)

	// RegisterSeasonChangeCallback registers a callback for season changes
	RegisterSeasonChangeCallback(callback SeasonChangeCallback)

	// Start starts the time manager
	Start()

	// Stop stops the time manager
	Stop()

	// Update updates the time manager
	Update(deltaTime time.Duration)
}

// TimeChangeCallback is called when game time changes
type TimeChangeCallback func(newTime GameTime)

// SeasonChangeCallback is called when the season changes
type SeasonChangeCallback func(oldSeason, newSeason Season)

// StandardTimeManager implements TimeManager
type StandardTimeManager struct {
	currentTime           GameTime
	realTime              time.Duration
	gameSpeed             float64
	timePerDay            time.Duration
	timeChangeCallbacks   []TimeChangeCallback
	seasonChangeCallbacks []SeasonChangeCallback
	gameLoop              gameloop.GameLoop
	lastDay               int
	running               bool
	mu                    sync.RWMutex
}

// NewStandardTimeManager creates a new time manager
func NewStandardTimeManager(gameLoop gameloop.GameLoop, timePerDay time.Duration) *StandardTimeManager {
	tm := &StandardTimeManager{
		currentTime: GameTime{
			Year:   1,
			Season: Spring,
			Day:    1,
			Phase:  gameloop.PhaseMorning,
		},
		gameSpeed:             1.0,
		timePerDay:            timePerDay,
		timeChangeCallbacks:   make([]TimeChangeCallback, 0),
		seasonChangeCallbacks: make([]SeasonChangeCallback, 0),
		gameLoop:              gameLoop,
		lastDay:               1,
		running:               false,
	}

	// Register with game loop
	if gameLoop != nil {
		gameLoop.RegisterUpdateCallback(tm.onGameLoopUpdate)
		gameLoop.RegisterPhaseChangeCallback(tm.onPhaseChange)
	}

	return tm
}

// GetCurrentTime returns the current game time
func (tm *StandardTimeManager) GetCurrentTime() GameTime {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentTime
}

// GetRealTime returns the real elapsed time
func (tm *StandardTimeManager) GetRealTime() time.Duration {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.realTime
}

// GetGameSpeed returns the current game speed multiplier
func (tm *StandardTimeManager) GetGameSpeed() float64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.gameSpeed
}

// SetGameSpeed sets the game speed multiplier
func (tm *StandardTimeManager) SetGameSpeed(multiplier float64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.gameSpeed = multiplier
	if tm.gameLoop != nil {
		tm.gameLoop.SetSpeed(multiplier)
	}
}

// AdvanceDay advances the game by one day
func (tm *StandardTimeManager) AdvanceDay() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	oldSeason := tm.currentTime.Season
	tm.currentTime.Day++

	// Check for season change
	if tm.currentTime.Day > DaysPerSeason {
		tm.currentTime.Day = 1
		tm.currentTime.Season++

		// Check for year change
		if tm.currentTime.Season > Winter {
			tm.currentTime.Season = Spring
			tm.currentTime.Year++
		}

		// Notify season change
		if oldSeason != tm.currentTime.Season {
			tm.notifySeasonChange(oldSeason, tm.currentTime.Season)
		}
	}

	tm.notifyTimeChange()
}

// GetDayOfYear returns the current day of the year (1-120)
func (tm *StandardTimeManager) GetDayOfYear() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return int(tm.currentTime.Season)*DaysPerSeason + tm.currentTime.Day
}

// GetSeasonName returns the name of the current season
func (tm *StandardTimeManager) GetSeasonName() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return GetSeasonName(tm.currentTime.Season)
}

// IsNewSeason returns true if it's the first day of a new season
func (tm *StandardTimeManager) IsNewSeason() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentTime.Day == 1
}

// IsNewYear returns true if it's the first day of a new year
func (tm *StandardTimeManager) IsNewYear() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentTime.Season == Spring && tm.currentTime.Day == 1
}

// RegisterTimeChangeCallback registers a callback for time changes
func (tm *StandardTimeManager) RegisterTimeChangeCallback(callback TimeChangeCallback) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.timeChangeCallbacks = append(tm.timeChangeCallbacks, callback)
}

// RegisterSeasonChangeCallback registers a callback for season changes
func (tm *StandardTimeManager) RegisterSeasonChangeCallback(callback SeasonChangeCallback) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.seasonChangeCallbacks = append(tm.seasonChangeCallbacks, callback)
}

// Start starts the time manager
func (tm *StandardTimeManager) Start() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.running = true
}

// Stop stops the time manager
func (tm *StandardTimeManager) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.running = false
}

// Update updates the time manager
func (tm *StandardTimeManager) Update(deltaTime time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.running {
		return
	}

	// Update real time
	tm.realTime += deltaTime

	// Calculate current day based on elapsed time
	currentDay := int(tm.realTime/tm.timePerDay) + 1

	// Check if day has changed
	if currentDay != tm.lastDay {
		daysToAdvance := currentDay - tm.lastDay
		for i := 0; i < daysToAdvance; i++ {
			tm.advanceDayInternal()
		}
		tm.lastDay = currentDay
	}
}

// onGameLoopUpdate is called by the game loop
func (tm *StandardTimeManager) onGameLoopUpdate(deltaTime time.Duration) error {
	tm.Update(deltaTime)
	return nil
}

// onPhaseChange is called when the game phase changes
func (tm *StandardTimeManager) onPhaseChange(oldPhase, newPhase gameloop.Phase) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.currentTime.Phase = newPhase
	tm.notifyTimeChange()
}

// advanceDayInternal advances the day without locking (must be called with lock held)
func (tm *StandardTimeManager) advanceDayInternal() {
	oldSeason := tm.currentTime.Season
	tm.currentTime.Day++

	// Check for season change
	if tm.currentTime.Day > DaysPerSeason {
		tm.currentTime.Day = 1
		tm.currentTime.Season++

		// Check for year change
		if tm.currentTime.Season > Winter {
			tm.currentTime.Season = Spring
			tm.currentTime.Year++
		}

		// Notify season change
		if oldSeason != tm.currentTime.Season {
			tm.notifySeasonChange(oldSeason, tm.currentTime.Season)
		}
	}

	tm.notifyTimeChange()
}

// notifyTimeChange notifies all time change callbacks
func (tm *StandardTimeManager) notifyTimeChange() {
	for _, callback := range tm.timeChangeCallbacks {
		callback(tm.currentTime)
	}
}

// notifySeasonChange notifies all season change callbacks
func (tm *StandardTimeManager) notifySeasonChange(oldSeason, newSeason Season) {
	for _, callback := range tm.seasonChangeCallbacks {
		callback(oldSeason, newSeason)
	}
}

// GetSeasonName returns the string name of a season
func GetSeasonName(season Season) string {
	switch season {
	case Spring:
		return "Spring"
	case Summer:
		return "Summer"
	case Autumn:
		return "Autumn"
	case Winter:
		return "Winter"
	default:
		return "Unknown"
	}
}

// FormatGameTime formats a GameTime into a readable string
func FormatGameTime(gt GameTime) string {
	return fmt.Sprintf("Year %d, %s Day %d, %s",
		gt.Year,
		GetSeasonName(gt.Season),
		gt.Day,
		gameloop.GetPhaseName(gt.Phase))
}
