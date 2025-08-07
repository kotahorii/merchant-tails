package progression

import (
	"sync"
)

// Feature represents an unlockable game feature
type Feature struct {
	ID                   string
	Name                 string
	Description          string
	RequiredRank         Rank
	RequiredAchievements []string
	RequiredFeatures     []string
	RequiredGold         int
	Unlocked             bool
}

// FeatureUnlockSystem manages feature unlocking
type FeatureUnlockSystem struct {
	features         map[string]*Feature
	unlockedFeatures map[string]bool
	mu               sync.RWMutex
}

// NewFeatureUnlockSystem creates a new feature unlock system
func NewFeatureUnlockSystem() *FeatureUnlockSystem {
	return &FeatureUnlockSystem{
		features:         make(map[string]*Feature),
		unlockedFeatures: make(map[string]bool),
	}
}

// RegisterFeature adds a new feature to the system
func (fs *FeatureUnlockSystem) RegisterFeature(feature *Feature) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.features[feature.ID] = feature
	// Features start as locked
	fs.unlockedFeatures[feature.ID] = false
}

// IsAvailable checks if a feature is available based on rank
func (fs *FeatureUnlockSystem) IsAvailable(featureID string, currentRank Rank) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	feature, exists := fs.features[featureID]
	if !exists {
		return false
	}

	// Check rank requirement
	return currentRank >= feature.RequiredRank
}

// IsAvailableWithAchievements checks if a feature is available with achievements
func (fs *FeatureUnlockSystem) IsAvailableWithAchievements(
	featureID string,
	currentRank Rank,
	unlockedAchievements []string,
) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	feature, exists := fs.features[featureID]
	if !exists {
		return false
	}

	// Check rank requirement
	if currentRank < feature.RequiredRank {
		return false
	}

	// Check achievement requirements
	if len(feature.RequiredAchievements) > 0 {
		achievementMap := make(map[string]bool)
		for _, achievement := range unlockedAchievements {
			achievementMap[achievement] = true
		}

		for _, required := range feature.RequiredAchievements {
			if !achievementMap[required] {
				return false
			}
		}
	}

	return true
}

// IsAvailableWithFeatures checks if a feature is available with other features
func (fs *FeatureUnlockSystem) IsAvailableWithFeatures(
	featureID string,
	currentRank Rank,
	unlockedFeatures []string,
) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	feature, exists := fs.features[featureID]
	if !exists {
		return false
	}

	// Check rank requirement
	if currentRank < feature.RequiredRank {
		return false
	}

	// Check feature requirements
	if len(feature.RequiredFeatures) > 0 {
		featureMap := make(map[string]bool)
		for _, f := range unlockedFeatures {
			featureMap[f] = true
		}

		for _, required := range feature.RequiredFeatures {
			if !featureMap[required] {
				return false
			}
		}
	}

	return true
}

// UnlockFeature marks a feature as unlocked
func (fs *FeatureUnlockSystem) UnlockFeature(featureID string) bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	feature, exists := fs.features[featureID]
	if !exists {
		return false
	}

	// Check if already unlocked
	if fs.unlockedFeatures[featureID] {
		return false
	}

	// Unlock the feature
	fs.unlockedFeatures[featureID] = true
	feature.Unlocked = true

	return true
}

// IsUnlocked checks if a feature is unlocked
func (fs *FeatureUnlockSystem) IsUnlocked(featureID string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.unlockedFeatures[featureID]
}

// GetAllFeatures returns all registered features
func (fs *FeatureUnlockSystem) GetAllFeatures() []*Feature {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	features := make([]*Feature, 0, len(fs.features))
	for _, feature := range fs.features {
		features = append(features, feature)
	}
	return features
}

// GetUnlockedFeatures returns all unlocked features
func (fs *FeatureUnlockSystem) GetUnlockedFeatures() []*Feature {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	features := make([]*Feature, 0)
	for id, unlocked := range fs.unlockedFeatures {
		if unlocked {
			if feature, exists := fs.features[id]; exists {
				features = append(features, feature)
			}
		}
	}
	return features
}

// GetAvailableFeatures returns features available to unlock
func (fs *FeatureUnlockSystem) GetAvailableFeatures(
	currentRank Rank,
	unlockedAchievements []string,
	currentGold int,
) []*Feature {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	availableFeatures := make([]*Feature, 0)
	unlockedList := fs.getUnlockedFeaturesList()

	for id, feature := range fs.features {
		// Skip if already unlocked
		if fs.unlockedFeatures[id] {
			continue
		}

		// Check rank
		if currentRank < feature.RequiredRank {
			continue
		}

		// Check gold
		if currentGold < feature.RequiredGold {
			continue
		}

		// Check achievements
		if !fs.hasRequiredAchievements(feature, unlockedAchievements) {
			continue
		}

		// Check prerequisite features
		if !fs.hasRequiredFeatures(feature, unlockedList) {
			continue
		}

		availableFeatures = append(availableFeatures, feature)
	}

	return availableFeatures
}

// getUnlockedFeaturesList returns a list of unlocked feature IDs
func (fs *FeatureUnlockSystem) getUnlockedFeaturesList() []string {
	unlockedList := make([]string, 0)
	for id, unlocked := range fs.unlockedFeatures {
		if unlocked {
			unlockedList = append(unlockedList, id)
		}
	}
	return unlockedList
}

// hasRequiredAchievements checks if all required achievements are unlocked
func (fs *FeatureUnlockSystem) hasRequiredAchievements(feature *Feature, unlockedAchievements []string) bool {
	if len(feature.RequiredAchievements) == 0 {
		return true
	}

	achievementMap := make(map[string]bool)
	for _, achievement := range unlockedAchievements {
		achievementMap[achievement] = true
	}

	for _, required := range feature.RequiredAchievements {
		if !achievementMap[required] {
			return false
		}
	}

	return true
}

// hasRequiredFeatures checks if all required features are unlocked
func (fs *FeatureUnlockSystem) hasRequiredFeatures(feature *Feature, unlockedFeatures []string) bool {
	if len(feature.RequiredFeatures) == 0 {
		return true
	}

	featureMap := make(map[string]bool)
	for _, f := range unlockedFeatures {
		featureMap[f] = true
	}

	for _, required := range feature.RequiredFeatures {
		if !featureMap[required] {
			return false
		}
	}

	return true
}

// ResetFeatures resets all feature unlocks (for new game)
func (fs *FeatureUnlockSystem) ResetFeatures() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for id := range fs.unlockedFeatures {
		fs.unlockedFeatures[id] = false
		if feature, exists := fs.features[id]; exists {
			feature.Unlocked = false
		}
	}
}
