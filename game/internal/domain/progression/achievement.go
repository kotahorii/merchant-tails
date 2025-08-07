package progression

import (
	"sync"
	"time"
)

// AchievementCategory represents the category of an achievement
type AchievementCategory int

const (
	AchievementCategoryTrading AchievementCategory = iota
	AchievementCategoryWealth
	AchievementCategoryReputation
	AchievementCategoryExploration
	AchievementCategoryMastery
)

// Achievement represents a single achievement
type Achievement struct {
	ID            string
	Name          string
	Description   string
	Points        int
	Category      AchievementCategory
	MaxProgress   int
	IsProgressive bool
	Hidden        bool
	UnlockedAt    *time.Time
}

// AchievementProgress tracks progress for progressive achievements
type AchievementProgress struct {
	AchievementID string
	CurrentValue  int
	MaxValue      int
	Completed     bool
}

// AchievementSystem manages player achievements
type AchievementSystem struct {
	achievements map[string]*Achievement
	unlocked     map[string]bool
	progress     map[string]*AchievementProgress
	totalPoints  int
	mu           sync.RWMutex
}

// NewAchievementSystem creates a new achievement system
func NewAchievementSystem() *AchievementSystem {
	return &AchievementSystem{
		achievements: make(map[string]*Achievement),
		unlocked:     make(map[string]bool),
		progress:     make(map[string]*AchievementProgress),
		totalPoints:  0,
	}
}

// RegisterAchievement adds a new achievement to the system
func (as *AchievementSystem) RegisterAchievement(achievement *Achievement) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.achievements[achievement.ID] = achievement

	// Initialize progress for progressive achievements
	if achievement.IsProgressive {
		as.progress[achievement.ID] = &AchievementProgress{
			AchievementID: achievement.ID,
			CurrentValue:  0,
			MaxValue:      achievement.MaxProgress,
			Completed:     false,
		}
	}
}

// UnlockAchievement unlocks an achievement
func (as *AchievementSystem) UnlockAchievement(achievementID string) bool {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Check if achievement exists
	achievement, exists := as.achievements[achievementID]
	if !exists {
		return false
	}

	// Check if already unlocked
	if as.unlocked[achievementID] {
		return false
	}

	// Unlock the achievement
	as.unlocked[achievementID] = true
	now := time.Now()
	achievement.UnlockedAt = &now

	// Add points
	as.totalPoints += achievement.Points

	// Mark progress as completed if progressive
	if progress, exists := as.progress[achievementID]; exists {
		progress.Completed = true
		progress.CurrentValue = progress.MaxValue
	}

	return true
}

// IsUnlocked checks if an achievement is unlocked
func (as *AchievementSystem) IsUnlocked(achievementID string) bool {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.unlocked[achievementID]
}

// UpdateProgress updates progress for a progressive achievement
func (as *AchievementSystem) UpdateProgress(achievementID string, amount int) {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Check if achievement exists and is progressive
	achievement, exists := as.achievements[achievementID]
	if !exists || !achievement.IsProgressive {
		return
	}

	// Get or create progress
	progress, exists := as.progress[achievementID]
	if !exists {
		progress = &AchievementProgress{
			AchievementID: achievementID,
			CurrentValue:  0,
			MaxValue:      achievement.MaxProgress,
			Completed:     false,
		}
		as.progress[achievementID] = progress
	}

	// Update progress
	progress.CurrentValue += amount

	// Check if completed
	if progress.CurrentValue >= progress.MaxValue && !progress.Completed {
		// Unlock the achievement
		as.unlocked[achievementID] = true
		now := time.Now()
		achievement.UnlockedAt = &now
		as.totalPoints += achievement.Points
		progress.Completed = true
	}
}

// GetProgress returns the current progress for an achievement
func (as *AchievementSystem) GetProgress(achievementID string) int {
	as.mu.RLock()
	defer as.mu.RUnlock()

	if progress, exists := as.progress[achievementID]; exists {
		return progress.CurrentValue
	}
	return 0
}

// GetAllAchievements returns all registered achievements
func (as *AchievementSystem) GetAllAchievements() []*Achievement {
	as.mu.RLock()
	defer as.mu.RUnlock()

	achievements := make([]*Achievement, 0, len(as.achievements))
	for _, achievement := range as.achievements {
		achievements = append(achievements, achievement)
	}
	return achievements
}

// GetUnlockedAchievements returns all unlocked achievements
func (as *AchievementSystem) GetUnlockedAchievements() []*Achievement {
	as.mu.RLock()
	defer as.mu.RUnlock()

	achievements := make([]*Achievement, 0)
	for id, unlocked := range as.unlocked {
		if unlocked {
			if achievement, exists := as.achievements[id]; exists {
				achievements = append(achievements, achievement)
			}
		}
	}
	return achievements
}

// GetAchievementsByCategory returns achievements by category
func (as *AchievementSystem) GetAchievementsByCategory(category AchievementCategory) []*Achievement {
	as.mu.RLock()
	defer as.mu.RUnlock()

	achievements := make([]*Achievement, 0)
	for _, achievement := range as.achievements {
		if achievement.Category == category {
			achievements = append(achievements, achievement)
		}
	}
	return achievements
}

// GetTotalPoints returns total achievement points earned
func (as *AchievementSystem) GetTotalPoints() int {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.totalPoints
}

// GetCompletionPercentage returns the percentage of achievements completed
func (as *AchievementSystem) GetCompletionPercentage() float64 {
	as.mu.RLock()
	defer as.mu.RUnlock()

	if len(as.achievements) == 0 {
		return 0.0
	}

	unlockedCount := 0
	for _, unlocked := range as.unlocked {
		if unlocked {
			unlockedCount++
		}
	}

	return float64(unlockedCount) / float64(len(as.achievements)) * 100.0
}

// ResetAchievements resets all achievements (for new game)
func (as *AchievementSystem) ResetAchievements() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.unlocked = make(map[string]bool)
	as.totalPoints = 0

	// Reset progress for progressive achievements
	for id, achievement := range as.achievements {
		if achievement.IsProgressive {
			as.progress[id] = &AchievementProgress{
				AchievementID: id,
				CurrentValue:  0,
				MaxValue:      achievement.MaxProgress,
				Completed:     false,
			}
		}
		achievement.UnlockedAt = nil
	}
}

// GetAchievementCategoryName returns the name of an achievement category
func GetAchievementCategoryName(category AchievementCategory) string {
	switch category {
	case AchievementCategoryTrading:
		return "Trading"
	case AchievementCategoryWealth:
		return "Wealth"
	case AchievementCategoryReputation:
		return "Reputation"
	case AchievementCategoryExploration:
		return "Exploration"
	case AchievementCategoryMastery:
		return "Mastery"
	default:
		return "Unknown"
	}
}
