package progression

import (
	"sync"
)

// Rank represents a player's rank level
type Rank int

const (
	RankApprentice Rank = iota
	RankJourneyman
	RankExpert
	RankMaster
)

// RankBenefits represents the benefits of each rank
type RankBenefits struct {
	ShopCapacity      int
	WarehouseCapacity int
	PriceModifier     float64
	MaxAICompetitors  int
	UnlockFeatures    []string
}

// RankSystem manages player rank progression
type RankSystem struct {
	currentRank Rank
	experience  int
	mu          sync.RWMutex
}

// NewRankSystem creates a new rank system
func NewRankSystem() *RankSystem {
	return &RankSystem{
		currentRank: RankApprentice,
		experience:  0,
	}
}

// GetCurrentRank returns the current player rank
func (rs *RankSystem) GetCurrentRank() Rank {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.currentRank
}

// SetRank sets the player's rank (for testing or loading save)
func (rs *RankSystem) SetRank(rank Rank) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.currentRank = rank
	rs.experience = 0
}

// GetExperience returns current experience points
func (rs *RankSystem) GetExperience() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.experience
}

// AddExperience adds experience and checks for rank up
func (rs *RankSystem) AddExperience(amount int) bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.experience += amount

	// Check if we should rank up
	required := rs.getExperienceRequiredForRank(rs.currentRank)
	if required > 0 && rs.experience >= required {
		return rs.rankUp()
	}

	return false
}

// rankUp promotes the player to the next rank
func (rs *RankSystem) rankUp() bool {
	switch rs.currentRank {
	case RankApprentice:
		rs.currentRank = RankJourneyman
		rs.experience = 0
		return true
	case RankJourneyman:
		rs.currentRank = RankExpert
		rs.experience = 0
		return true
	case RankExpert:
		rs.currentRank = RankMaster
		rs.experience = 0
		return true
	case RankMaster:
		// Already at max rank
		return false
	}
	return false
}

// GetExperienceRequired returns experience needed for next rank
func (rs *RankSystem) GetExperienceRequired() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.getExperienceRequiredForRank(rs.currentRank)
}

// getExperienceRequiredForRank returns experience needed for each rank
func (rs *RankSystem) getExperienceRequiredForRank(rank Rank) int {
	switch rank {
	case RankApprentice:
		return 1000
	case RankJourneyman:
		return 2500
	case RankExpert:
		return 5000
	case RankMaster:
		return 0 // No further promotion
	default:
		return 0
	}
}

// GetExperienceToNextRank returns remaining experience needed
func (rs *RankSystem) GetExperienceToNextRank() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	required := rs.getExperienceRequiredForRank(rs.currentRank)
	if required == 0 {
		return 0 // Already at max rank
	}

	return required - rs.experience
}

// GetRankBenefits returns the benefits for the current rank
func (rs *RankSystem) GetRankBenefits() RankBenefits {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	switch rs.currentRank {
	case RankApprentice:
		return RankBenefits{
			ShopCapacity:      50,
			WarehouseCapacity: 100,
			PriceModifier:     1.0,
			MaxAICompetitors:  2,
			UnlockFeatures:    []string{"basic_trading"},
		}
	case RankJourneyman:
		return RankBenefits{
			ShopCapacity:      75,
			WarehouseCapacity: 150,
			PriceModifier:     0.95,
			MaxAICompetitors:  3,
			UnlockFeatures:    []string{"basic_trading", "market_analysis"},
		}
	case RankExpert:
		return RankBenefits{
			ShopCapacity:      100,
			WarehouseCapacity: 250,
			PriceModifier:     0.90,
			MaxAICompetitors:  4,
			UnlockFeatures:    []string{"basic_trading", "market_analysis", "advanced_trading"},
		}
	case RankMaster:
		return RankBenefits{
			ShopCapacity:      150,
			WarehouseCapacity: 500,
			PriceModifier:     0.85,
			MaxAICompetitors:  6,
			UnlockFeatures:    []string{"basic_trading", "market_analysis", "advanced_trading", "master_trader"},
		}
	default:
		return RankBenefits{
			ShopCapacity:      50,
			WarehouseCapacity: 100,
			PriceModifier:     1.0,
			MaxAICompetitors:  2,
			UnlockFeatures:    []string{"basic_trading"},
		}
	}
}

// GetRankName returns the string name of a rank
func GetRankName(rank Rank) string {
	switch rank {
	case RankApprentice:
		return "Apprentice"
	case RankJourneyman:
		return "Journeyman"
	case RankExpert:
		return "Expert"
	case RankMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

// GetRankDescription returns a description of the rank
func GetRankDescription(rank Rank) string {
	switch rank {
	case RankApprentice:
		return "A novice merchant learning the basics of trade"
	case RankJourneyman:
		return "An experienced trader with proven skills"
	case RankExpert:
		return "A master of markets with advanced techniques"
	case RankMaster:
		return "A legendary merchant whose name echoes through history"
	default:
		return "Unknown rank"
	}
}
