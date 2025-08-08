package quest

import (
	"fmt"
	"sync"
	"time"
)

// QuestID represents unique quest identifier
type QuestID string

// Quest IDs for investment learning quests
const (
	// Tutorial Quests
	QuestFirstTrade      QuestID = "first_trade"
	QuestFirstProfit     QuestID = "first_profit"
	QuestBasicInvestment QuestID = "basic_investment"

	// Investment Learning Quests
	QuestROI25          QuestID = "roi_25"
	QuestROI50          QuestID = "roi_50"
	QuestROI100         QuestID = "roi_100"
	QuestCompoundReturn QuestID = "compound_return"
	QuestRiskManagement QuestID = "risk_management"

	// Portfolio Quests
	QuestDiversifyPortfolio QuestID = "diversify_portfolio"
	QuestBalancedInvestment QuestID = "balanced_investment"
	QuestSeasonalStrategy   QuestID = "seasonal_strategy"

	// Advanced Investment Quests
	QuestMarketTiming       QuestID = "market_timing"
	QuestVolatilityProfit   QuestID = "volatility_profit"
	QuestLongTermInvestment QuestID = "long_term_investment"
	QuestValueInvesting     QuestID = "value_investing"

	// Shop Investment Quests
	QuestShopUpgrade      QuestID = "shop_upgrade"
	QuestEquipmentROI     QuestID = "equipment_roi"
	QuestPassiveIncome    QuestID = "passive_income"
	QuestInvestmentMaster QuestID = "investment_master"
)

// QuestType represents the type of quest
type QuestType int

const (
	QuestTypeMain QuestType = iota
	QuestTypeSide
	QuestTypeDaily
	QuestTypeTutorial
	QuestTypeChallenge
)

// QuestStatus represents the current status of a quest
type QuestStatus int

const (
	QuestStatusLocked QuestStatus = iota
	QuestStatusAvailable
	QuestStatusActive
	QuestStatusCompleted
	QuestStatusFailed
)

// QuestObjective represents a single objective within a quest
type QuestObjective struct {
	ID          string
	Description string
	Current     int
	Target      int
	Completed   bool
}

// QuestReward represents rewards for completing a quest
type QuestReward struct {
	Gold       int
	Experience int
	Items      map[string]int // ItemID -> Quantity
	Reputation float64
	Unlocks    []QuestID // Quests to unlock
}

// Quest represents a single quest
type Quest struct {
	ID          QuestID
	Name        string
	Description string
	Type        QuestType
	Status      QuestStatus
	Level       int // Required player level
	Objectives  []*QuestObjective
	Rewards     *QuestReward
	TimeLimit   *time.Duration // Optional time limit
	StartedAt   *time.Time
	CompletedAt *time.Time
	FailedAt    *time.Time
	Chain       []QuestID // Quest chain sequence
	ChainIndex  int       // Current position in chain
}

// QuestManager manages all quests
type QuestManager struct {
	quests          map[QuestID]*Quest
	activeQuests    map[QuestID]*Quest
	completedQuests map[QuestID]bool
	questChains     map[string][]QuestID // Chain name -> Quest IDs
	statistics      *QuestStatistics
	callbacks       []QuestCallback
	mu              sync.RWMutex
}

// QuestStatistics tracks quest completion stats
type QuestStatistics struct {
	TotalCompleted   int
	MainCompleted    int
	SideCompleted    int
	DailyCompleted   int
	TotalFailed      int
	TotalRewards     int
	BestROIAchieved  float64
	InvestmentProfit int
	QuestStreak      int
	LastCompletedAt  *time.Time
}

// QuestCallback is called when quest status changes
type QuestCallback func(quest *Quest, oldStatus QuestStatus)

// NewQuestManager creates a new quest manager
func NewQuestManager() *QuestManager {
	qm := &QuestManager{
		quests:          make(map[QuestID]*Quest),
		activeQuests:    make(map[QuestID]*Quest),
		completedQuests: make(map[QuestID]bool),
		questChains:     make(map[string][]QuestID),
		statistics:      &QuestStatistics{},
		callbacks:       make([]QuestCallback, 0),
	}

	qm.initializeQuests()
	qm.initializeQuestChains()
	return qm
}

// initializeQuests sets up all quests
func (qm *QuestManager) initializeQuests() {
	// Tutorial Quests
	qm.registerQuest(&Quest{
		ID:          QuestFirstTrade,
		Name:        "Your First Trade",
		Description: "Learn the basics of trading by making your first purchase and sale",
		Type:        QuestTypeTutorial,
		Status:      QuestStatusAvailable,
		Level:       1,
		Objectives: []*QuestObjective{
			{ID: "buy_item", Description: "Buy any item from the market", Target: 1},
			{ID: "sell_item", Description: "Sell any item for profit", Target: 1},
		},
		Rewards: &QuestReward{
			Gold:       100,
			Experience: 50,
			Reputation: 5,
			Unlocks:    []QuestID{QuestFirstProfit},
		},
	})

	qm.registerQuest(&Quest{
		ID:          QuestFirstProfit,
		Name:        "Profitable Trader",
		Description: "Make your first 100 gold profit from trading",
		Type:        QuestTypeTutorial,
		Status:      QuestStatusLocked,
		Level:       1,
		Objectives: []*QuestObjective{
			{ID: "earn_profit", Description: "Earn 100 gold profit", Target: 100},
		},
		Rewards: &QuestReward{
			Gold:       200,
			Experience: 100,
			Reputation: 10,
			Unlocks:    []QuestID{QuestROI25},
		},
	})

	// Investment Learning Quests
	qm.registerQuest(&Quest{
		ID:          QuestROI25,
		Name:        "Smart Investment",
		Description: "Achieve 25% return on investment in a single trade",
		Type:        QuestTypeMain,
		Status:      QuestStatusLocked,
		Level:       2,
		Objectives: []*QuestObjective{
			{ID: "achieve_roi", Description: "Achieve 25% ROI", Target: 25},
		},
		Rewards: &QuestReward{
			Gold:       500,
			Experience: 200,
			Reputation: 15,
			Unlocks:    []QuestID{QuestROI50},
		},
	})

	qm.registerQuest(&Quest{
		ID:          QuestROI50,
		Name:        "Double Your Money",
		Description: "Achieve 50% return on investment",
		Type:        QuestTypeMain,
		Status:      QuestStatusLocked,
		Level:       3,
		Objectives: []*QuestObjective{
			{ID: "achieve_roi", Description: "Achieve 50% ROI", Target: 50},
		},
		Rewards: &QuestReward{
			Gold:       1000,
			Experience: 400,
			Reputation: 25,
			Unlocks:    []QuestID{QuestROI100},
		},
	})

	qm.registerQuest(&Quest{
		ID:          QuestROI100,
		Name:        "Investment Master",
		Description: "Double your investment with 100% ROI",
		Type:        QuestTypeChallenge,
		Status:      QuestStatusLocked,
		Level:       5,
		Objectives: []*QuestObjective{
			{ID: "achieve_roi", Description: "Achieve 100% ROI", Target: 100},
		},
		Rewards: &QuestReward{
			Gold:       2500,
			Experience: 800,
			Reputation: 50,
			Unlocks:    []QuestID{QuestInvestmentMaster},
		},
	})

	// Portfolio Diversification Quests
	qm.registerQuest(&Quest{
		ID:          QuestDiversifyPortfolio,
		Name:        "Diversified Investor",
		Description: "Learn risk management by diversifying your portfolio across 5 different item categories",
		Type:        QuestTypeMain,
		Status:      QuestStatusLocked,
		Level:       3,
		Objectives: []*QuestObjective{
			{ID: "diversify", Description: "Trade in 5 different categories", Target: 5},
			{ID: "profit_each", Description: "Make profit in each category", Target: 5},
		},
		Rewards: &QuestReward{
			Gold:       1500,
			Experience: 500,
			Reputation: 30,
			Items: map[string]int{
				"investment_guide": 1,
			},
		},
	})

	// Shop Investment Quests
	qm.registerQuest(&Quest{
		ID:          QuestShopUpgrade,
		Name:        "Business Expansion",
		Description: "Invest in your shop infrastructure to increase capacity and efficiency",
		Type:        QuestTypeMain,
		Status:      QuestStatusLocked,
		Level:       4,
		Objectives: []*QuestObjective{
			{ID: "upgrade_shop", Description: "Upgrade shop to level 3", Target: 3},
			{ID: "purchase_equipment", Description: "Purchase 3 equipment items", Target: 3},
		},
		Rewards: &QuestReward{
			Gold:       2000,
			Experience: 600,
			Reputation: 35,
			Unlocks:    []QuestID{QuestEquipmentROI},
		},
	})

	qm.registerQuest(&Quest{
		ID:          QuestEquipmentROI,
		Name:        "Equipment Returns",
		Description: "Achieve positive ROI from equipment investments within 30 days",
		Type:        QuestTypeChallenge,
		Status:      QuestStatusLocked,
		Level:       5,
		Objectives: []*QuestObjective{
			{ID: "equipment_roi", Description: "Earn back equipment cost", Target: 100},
			{ID: "time_limit", Description: "Within 30 days", Target: 30},
		},
		Rewards: &QuestReward{
			Gold:       3000,
			Experience: 1000,
			Reputation: 50,
			Items: map[string]int{
				"premium_equipment": 1,
			},
		},
	})

	// Market Timing Quests
	qm.registerQuest(&Quest{
		ID:          QuestMarketTiming,
		Name:        "Perfect Timing",
		Description: "Master market timing by buying low and selling high during events",
		Type:        QuestTypeSide,
		Status:      QuestStatusLocked,
		Level:       4,
		Objectives: []*QuestObjective{
			{ID: "buy_low", Description: "Buy during price dip", Target: 10},
			{ID: "sell_high", Description: "Sell during price surge", Target: 10},
			{ID: "profit_margin", Description: "Achieve 40% profit margin", Target: 40},
		},
		Rewards: &QuestReward{
			Gold:       1800,
			Experience: 450,
			Reputation: 25,
		},
	})

	// Long-term Investment Quest
	timeLimit := 7 * 24 * time.Hour // 7 days
	qm.registerQuest(&Quest{
		ID:          QuestLongTermInvestment,
		Name:        "Patient Investor",
		Description: "Hold items for 7 days and sell for significant profit",
		Type:        QuestTypeChallenge,
		Status:      QuestStatusLocked,
		Level:       6,
		TimeLimit:   &timeLimit,
		Objectives: []*QuestObjective{
			{ID: "hold_items", Description: "Hold items for 7 days", Target: 7},
			{ID: "final_profit", Description: "Achieve 75% total ROI", Target: 75},
		},
		Rewards: &QuestReward{
			Gold:       5000,
			Experience: 1500,
			Reputation: 75,
			Items: map[string]int{
				"investor_badge": 1,
			},
		},
	})
}

// initializeQuestChains sets up quest chains
func (qm *QuestManager) initializeQuestChains() {
	// Investment Learning Chain
	qm.questChains["investment_basics"] = []QuestID{
		QuestFirstTrade,
		QuestFirstProfit,
		QuestROI25,
		QuestROI50,
		QuestROI100,
	}

	// Shop Development Chain
	qm.questChains["shop_development"] = []QuestID{
		QuestShopUpgrade,
		QuestEquipmentROI,
		QuestPassiveIncome,
	}

	// Market Mastery Chain
	qm.questChains["market_mastery"] = []QuestID{
		QuestMarketTiming,
		QuestVolatilityProfit,
		QuestLongTermInvestment,
	}
}

// registerQuest adds a quest to the manager
func (qm *QuestManager) registerQuest(quest *Quest) {
	qm.quests[quest.ID] = quest
}

// StartQuest starts a quest if available
func (qm *QuestManager) StartQuest(questID QuestID, playerLevel int) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quest, exists := qm.quests[questID]
	if !exists {
		return fmt.Errorf("quest not found: %s", questID)
	}

	if quest.Status != QuestStatusAvailable {
		return fmt.Errorf("quest not available: %s", questID)
	}

	if playerLevel < quest.Level {
		return fmt.Errorf("player level too low: required %d, current %d", quest.Level, playerLevel)
	}

	// Check if already active
	if _, active := qm.activeQuests[questID]; active {
		return fmt.Errorf("quest already active: %s", questID)
	}

	// Start the quest
	now := time.Now()
	quest.Status = QuestStatusActive
	quest.StartedAt = &now
	qm.activeQuests[questID] = quest

	// Notify callbacks
	qm.notifyCallbacks(quest, QuestStatusAvailable)

	return nil
}

// UpdateObjective updates progress on a quest objective
func (qm *QuestManager) UpdateObjective(questID QuestID, objectiveID string, progress int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quest, exists := qm.activeQuests[questID]
	if !exists || quest.Status != QuestStatusActive {
		return
	}

	for _, objective := range quest.Objectives {
		if objective.ID == objectiveID {
			objective.Current = min(progress, objective.Target)
			if objective.Current >= objective.Target {
				objective.Completed = true
			}
			qm.checkQuestCompletion(quest)
			break
		}
	}
}

// checkQuestCompletion checks if all objectives are complete
func (qm *QuestManager) checkQuestCompletion(quest *Quest) {
	allCompleted := true
	for _, objective := range quest.Objectives {
		if !objective.Completed {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		qm.completeQuest(quest)
	}
}

// completeQuest marks a quest as completed
func (qm *QuestManager) completeQuest(quest *Quest) {
	now := time.Now()
	quest.Status = QuestStatusCompleted
	quest.CompletedAt = &now

	delete(qm.activeQuests, quest.ID)
	qm.completedQuests[quest.ID] = true

	// Update statistics
	qm.statistics.TotalCompleted++
	qm.statistics.LastCompletedAt = &now

	switch quest.Type {
	case QuestTypeMain:
		qm.statistics.MainCompleted++
	case QuestTypeSide:
		qm.statistics.SideCompleted++
	case QuestTypeDaily:
		qm.statistics.DailyCompleted++
	}

	// Update quest streak
	if qm.statistics.LastCompletedAt != nil {
		diff := now.Sub(*qm.statistics.LastCompletedAt)
		if diff < 24*time.Hour {
			qm.statistics.QuestStreak++
		} else {
			qm.statistics.QuestStreak = 1
		}
	}

	// Unlock next quests
	if quest.Rewards != nil {
		for _, unlockID := range quest.Rewards.Unlocks {
			if nextQuest, exists := qm.quests[unlockID]; exists {
				if nextQuest.Status == QuestStatusLocked {
					nextQuest.Status = QuestStatusAvailable
				}
			}
		}
	}

	// Handle quest chain progression
	if len(quest.Chain) > 0 && quest.ChainIndex < len(quest.Chain)-1 {
		nextQuestID := quest.Chain[quest.ChainIndex+1]
		if nextQuest, exists := qm.quests[nextQuestID]; exists {
			nextQuest.Status = QuestStatusAvailable
			nextQuest.ChainIndex = quest.ChainIndex + 1
		}
	}

	// Notify callbacks
	qm.notifyCallbacks(quest, QuestStatusActive)
}

// FailQuest marks a quest as failed
func (qm *QuestManager) FailQuest(questID QuestID) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quest, exists := qm.activeQuests[questID]
	if !exists || quest.Status != QuestStatusActive {
		return
	}

	now := time.Now()
	quest.Status = QuestStatusFailed
	quest.FailedAt = &now

	delete(qm.activeQuests, questID)
	qm.statistics.TotalFailed++

	// Notify callbacks
	qm.notifyCallbacks(quest, QuestStatusActive)
}

// GetQuest returns a specific quest
func (qm *QuestManager) GetQuest(questID QuestID) (*Quest, bool) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quest, exists := qm.quests[questID]
	return quest, exists
}

// GetActiveQuests returns all active quests
func (qm *QuestManager) GetActiveQuests() []*Quest {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quests := make([]*Quest, 0, len(qm.activeQuests))
	for _, quest := range qm.activeQuests {
		quests = append(quests, quest)
	}
	return quests
}

// GetAvailableQuests returns quests available for the player level
func (qm *QuestManager) GetAvailableQuests(playerLevel int) []*Quest {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quests := make([]*Quest, 0)
	for _, quest := range qm.quests {
		if quest.Status == QuestStatusAvailable && quest.Level <= playerLevel {
			quests = append(quests, quest)
		}
	}
	return quests
}

// GetCompletedQuests returns all completed quests
func (qm *QuestManager) GetCompletedQuests() []*Quest {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quests := make([]*Quest, 0)
	for questID := range qm.completedQuests {
		if quest, exists := qm.quests[questID]; exists {
			quests = append(quests, quest)
		}
	}
	return quests
}

// GetQuestChain returns quests in a chain
func (qm *QuestManager) GetQuestChain(chainName string) []*Quest {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	questIDs, exists := qm.questChains[chainName]
	if !exists {
		return nil
	}

	quests := make([]*Quest, 0, len(questIDs))
	for _, questID := range questIDs {
		if quest, exists := qm.quests[questID]; exists {
			quests = append(quests, quest)
		}
	}
	return quests
}

// GetStatistics returns quest statistics
func (qm *QuestManager) GetStatistics() *QuestStatistics {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	return qm.statistics
}

// RegisterCallback registers a callback for quest status changes
func (qm *QuestManager) RegisterCallback(callback QuestCallback) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.callbacks = append(qm.callbacks, callback)
}

// notifyCallbacks notifies all registered callbacks
func (qm *QuestManager) notifyCallbacks(quest *Quest, oldStatus QuestStatus) {
	for _, callback := range qm.callbacks {
		callback(quest, oldStatus)
	}
}

// CheckTimeouts checks for quest timeouts
func (qm *QuestManager) CheckTimeouts() {
	qm.mu.Lock()

	toFail := make([]QuestID, 0)
	now := time.Now()
	for questID, quest := range qm.activeQuests {
		if quest.TimeLimit != nil && quest.StartedAt != nil {
			elapsed := now.Sub(*quest.StartedAt)
			if elapsed > *quest.TimeLimit {
				toFail = append(toFail, questID)
			}
		}
	}
	qm.mu.Unlock()

	// Fail quests outside of the lock
	for _, questID := range toFail {
		qm.FailQuest(questID)
	}
}

// ClaimReward claims rewards for a completed quest
func (qm *QuestManager) ClaimReward(questID QuestID) (*QuestReward, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quest, exists := qm.quests[questID]
	if !exists {
		return nil, fmt.Errorf("quest not found: %s", questID)
	}

	if quest.Status != QuestStatusCompleted {
		return nil, fmt.Errorf("quest not completed: %s", questID)
	}

	if quest.Rewards == nil {
		return nil, fmt.Errorf("quest has no rewards: %s", questID)
	}

	// Track reward statistics
	qm.statistics.TotalRewards += quest.Rewards.Gold

	return quest.Rewards, nil
}

// Reset resets all quest progress
func (qm *QuestManager) Reset() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// Reset all quests
	for _, quest := range qm.quests {
		quest.Status = QuestStatusLocked
		quest.StartedAt = nil
		quest.CompletedAt = nil
		quest.FailedAt = nil

		// Reset objectives
		for _, objective := range quest.Objectives {
			objective.Current = 0
			objective.Completed = false
		}
	}

	// Unlock initial quests
	if quest, exists := qm.quests[QuestFirstTrade]; exists {
		quest.Status = QuestStatusAvailable
	}

	// Clear tracking
	qm.activeQuests = make(map[QuestID]*Quest)
	qm.completedQuests = make(map[QuestID]bool)
	qm.statistics = &QuestStatistics{}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
