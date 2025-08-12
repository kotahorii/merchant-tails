package update

import (
	"sync"
	"time"
)

// UpdateType represents the type of update
type UpdateType string

const (
	UpdateTypeGold        UpdateType = "gold"
	UpdateTypeInventory   UpdateType = "inventory"
	UpdateTypeMarket      UpdateType = "market"
	UpdateTypePlayer      UpdateType = "player"
	UpdateTypeStats       UpdateType = "stats"
	UpdateTypeQuest       UpdateType = "quest"
	UpdateTypeAchievement UpdateType = "achievement"
	UpdateTypeWeather     UpdateType = "weather"
	UpdateTypeTime        UpdateType = "time"
	UpdateTypeFull        UpdateType = "full"
)

// UpdatePriority represents the priority of an update
type UpdatePriority int

const (
	PriorityLow      UpdatePriority = 0
	PriorityNormal   UpdatePriority = 1
	PriorityHigh     UpdatePriority = 2
	PriorityCritical UpdatePriority = 3
)

// PartialUpdate represents a partial update to the game state
type PartialUpdate struct {
	ID        string
	Type      UpdateType
	Priority  UpdatePriority
	Timestamp time.Time
	Data      interface{}
	Version   int64
	Metadata  map[string]interface{}
}

// UpdateBatch represents a batch of updates
type UpdateBatch struct {
	Updates   []*PartialUpdate
	BatchID   string
	CreatedAt time.Time
	Applied   bool
}

// UpdateManager manages partial updates
type UpdateManager struct {
	updates      map[string]*PartialUpdate
	pendingQueue []*PartialUpdate
	history      []*UpdateBatch
	version      int64
	subscribers  map[UpdateType][]UpdateSubscriber
	mu           sync.RWMutex
	maxHistory   int
	batchSize    int
}

// UpdateSubscriber is a callback for update notifications
type UpdateSubscriber func(update *PartialUpdate)

// NewUpdateManager creates a new update manager
func NewUpdateManager() *UpdateManager {
	return &UpdateManager{
		updates:      make(map[string]*PartialUpdate),
		pendingQueue: make([]*PartialUpdate, 0),
		history:      make([]*UpdateBatch, 0),
		subscribers:  make(map[UpdateType][]UpdateSubscriber),
		version:      0,
		maxHistory:   100,
		batchSize:    10,
	}
}

// QueueUpdate queues a partial update
func (um *UpdateManager) QueueUpdate(updateType UpdateType, data interface{}, priority UpdatePriority) string {
	um.mu.Lock()
	defer um.mu.Unlock()

	um.version++
	update := &PartialUpdate{
		ID:        generateUpdateID(),
		Type:      updateType,
		Priority:  priority,
		Timestamp: time.Now(),
		Data:      data,
		Version:   um.version,
		Metadata:  make(map[string]interface{}),
	}

	um.pendingQueue = append(um.pendingQueue, update)
	um.sortQueueByPriority()

	// Auto-flush if batch size reached
	if len(um.pendingQueue) >= um.batchSize {
		um.flushUpdates()
	}

	return update.ID
}

// FlushUpdates processes all pending updates
func (um *UpdateManager) FlushUpdates() *UpdateBatch {
	um.mu.Lock()
	defer um.mu.Unlock()

	return um.flushUpdates()
}

// flushUpdates internal implementation (must be called with lock held)
func (um *UpdateManager) flushUpdates() *UpdateBatch {
	if len(um.pendingQueue) == 0 {
		return nil
	}

	batch := &UpdateBatch{
		Updates:   make([]*PartialUpdate, len(um.pendingQueue)),
		BatchID:   generateBatchID(),
		CreatedAt: time.Now(),
		Applied:   false,
	}

	copy(batch.Updates, um.pendingQueue)
	um.pendingQueue = make([]*PartialUpdate, 0)

	// Apply updates
	for _, update := range batch.Updates {
		um.updates[update.ID] = update
		um.notifySubscribers(update)
	}

	batch.Applied = true
	um.addToHistory(batch)

	return batch
}

// Subscribe adds a subscriber for specific update types
func (um *UpdateManager) Subscribe(updateType UpdateType, subscriber UpdateSubscriber) {
	um.mu.Lock()
	defer um.mu.Unlock()

	if um.subscribers[updateType] == nil {
		um.subscribers[updateType] = make([]UpdateSubscriber, 0)
	}
	um.subscribers[updateType] = append(um.subscribers[updateType], subscriber)
}

// GetUpdate retrieves a specific update by ID
func (um *UpdateManager) GetUpdate(id string) (*PartialUpdate, bool) {
	um.mu.RLock()
	defer um.mu.RUnlock()

	update, exists := um.updates[id]
	return update, exists
}

// GetUpdatesSince returns all updates since a specific version
func (um *UpdateManager) GetUpdatesSince(version int64) []*PartialUpdate {
	um.mu.RLock()
	defer um.mu.RUnlock()

	var updates []*PartialUpdate
	for _, update := range um.updates {
		if update.Version > version {
			updates = append(updates, update)
		}
	}

	// Sort by version
	for i := 0; i < len(updates)-1; i++ {
		for j := i + 1; j < len(updates); j++ {
			if updates[i].Version > updates[j].Version {
				updates[i], updates[j] = updates[j], updates[i]
			}
		}
	}

	return updates
}

// GetPendingCount returns the number of pending updates
func (um *UpdateManager) GetPendingCount() int {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return len(um.pendingQueue)
}

// GetVersion returns the current version
func (um *UpdateManager) GetVersion() int64 {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.version
}

// ClearHistory clears old update history
func (um *UpdateManager) ClearHistory() {
	um.mu.Lock()
	defer um.mu.Unlock()

	if len(um.history) > um.maxHistory {
		um.history = um.history[len(um.history)-um.maxHistory:]
	}
}

// GetHistory returns update history
func (um *UpdateManager) GetHistory(limit int) []*UpdateBatch {
	um.mu.RLock()
	defer um.mu.RUnlock()

	if limit <= 0 || limit > len(um.history) {
		limit = len(um.history)
	}

	result := make([]*UpdateBatch, limit)
	start := len(um.history) - limit
	copy(result, um.history[start:])
	return result
}

// MergeUpdates merges multiple updates of the same type
func (um *UpdateManager) MergeUpdates(updates []*PartialUpdate) *PartialUpdate {
	if len(updates) == 0 {
		return nil
	}

	if len(updates) == 1 {
		return updates[0]
	}

	// Find highest priority
	maxPriority := PriorityLow
	for _, update := range updates {
		if update.Priority > maxPriority {
			maxPriority = update.Priority
		}
	}

	// Create merged update
	merged := &PartialUpdate{
		ID:        generateUpdateID(),
		Type:      updates[0].Type,
		Priority:  maxPriority,
		Timestamp: time.Now(),
		Version:   um.version,
		Data:      mergeData(updates),
		Metadata: map[string]interface{}{
			"merged_count": len(updates),
			"merged_ids":   getUpdateIDs(updates),
		},
	}

	return merged
}

// SetBatchSize sets the batch size for auto-flushing
func (um *UpdateManager) SetBatchSize(size int) {
	um.mu.Lock()
	defer um.mu.Unlock()
	if size > 0 {
		um.batchSize = size
	}
}

// SetMaxHistory sets the maximum history size
func (um *UpdateManager) SetMaxHistory(max int) {
	um.mu.Lock()
	defer um.mu.Unlock()
	if max > 0 {
		um.maxHistory = max
	}
}

// Helper functions

func (um *UpdateManager) sortQueueByPriority() {
	// Sort pending queue by priority (higher first) and timestamp (older first)
	for i := 0; i < len(um.pendingQueue)-1; i++ {
		for j := i + 1; j < len(um.pendingQueue); j++ {
			if um.pendingQueue[i].Priority < um.pendingQueue[j].Priority ||
				(um.pendingQueue[i].Priority == um.pendingQueue[j].Priority &&
					um.pendingQueue[i].Timestamp.After(um.pendingQueue[j].Timestamp)) {
				um.pendingQueue[i], um.pendingQueue[j] = um.pendingQueue[j], um.pendingQueue[i]
			}
		}
	}
}

func (um *UpdateManager) notifySubscribers(update *PartialUpdate) {
	// Notify specific type subscribers
	if subscribers, exists := um.subscribers[update.Type]; exists {
		for _, subscriber := range subscribers {
			go subscriber(update)
		}
	}

	// Notify "all" subscribers
	if subscribers, exists := um.subscribers[UpdateTypeFull]; exists {
		for _, subscriber := range subscribers {
			go subscriber(update)
		}
	}
}

func (um *UpdateManager) addToHistory(batch *UpdateBatch) {
	um.history = append(um.history, batch)
	if len(um.history) > um.maxHistory*2 {
		// Keep only recent history
		um.history = um.history[len(um.history)-um.maxHistory:]
	}
}

func generateUpdateID() string {
	return "update_" + time.Now().Format("20060102150405") + "_" + generateRandomString(6)
}

func generateBatchID() string {
	return "batch_" + time.Now().Format("20060102150405") + "_" + generateRandomString(6)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

func mergeData(updates []*PartialUpdate) interface{} {
	// This is a simplified merge - in real implementation,
	// this would be type-specific merging logic
	merged := make(map[string]interface{})
	for _, update := range updates {
		if data, ok := update.Data.(map[string]interface{}); ok {
			for k, v := range data {
				merged[k] = v
			}
		}
	}
	return merged
}

func getUpdateIDs(updates []*PartialUpdate) []string {
	ids := make([]string, len(updates))
	for i, update := range updates {
		ids[i] = update.ID
	}
	return ids
}

// DeltaTracker tracks changes between states
type DeltaTracker struct {
	oldState map[string]interface{}
	newState map[string]interface{}
	changes  map[string]*Delta
	mu       sync.RWMutex
}

// Delta represents a change in value
type Delta struct {
	Field    string
	OldValue interface{}
	NewValue interface{}
	Type     UpdateType
}

// NewDeltaTracker creates a new delta tracker
func NewDeltaTracker() *DeltaTracker {
	return &DeltaTracker{
		oldState: make(map[string]interface{}),
		newState: make(map[string]interface{}),
		changes:  make(map[string]*Delta),
	}
}

// TrackField tracks a field for changes
func (dt *DeltaTracker) TrackField(field string, value interface{}, updateType UpdateType) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if oldValue, exists := dt.oldState[field]; exists {
		if !isEqual(oldValue, value) {
			dt.changes[field] = &Delta{
				Field:    field,
				OldValue: oldValue,
				NewValue: value,
				Type:     updateType,
			}
			dt.newState[field] = value
		}
	} else {
		dt.oldState[field] = value
		dt.newState[field] = value
	}
}

// GetChanges returns all tracked changes
func (dt *DeltaTracker) GetChanges() map[string]*Delta {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	result := make(map[string]*Delta)
	for k, v := range dt.changes {
		result[k] = v
	}
	return result
}

// HasChanges returns true if there are any changes
func (dt *DeltaTracker) HasChanges() bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return len(dt.changes) > 0
}

// Reset resets the tracker
func (dt *DeltaTracker) Reset() {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	dt.oldState = dt.newState
	dt.newState = make(map[string]interface{})
	dt.changes = make(map[string]*Delta)
}

func isEqual(a, b interface{}) bool {
	// Simple equality check - could be enhanced
	return a == b
}
