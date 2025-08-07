package event

import "time"

// Event names as constants
const (
	EventNameItemRegistered      = "item.registered"
	EventNamePriceUpdated        = "price.updated"
	EventNameMarketEventOccurred = "market.event.occurred"
	EventNameInventoryChanged    = "inventory.changed"
	EventNameTransactionComplete = "transaction.complete"
	EventNameMerchantAction      = "merchant.action"
	EventNameSeasonChanged       = "season.changed"
	EventNameDayEnded            = "day.ended"
)

// BaseEvent provides common fields for all events
type BaseEvent struct {
	Name      string
	Timestamp int64
}

// EventName returns the event name
func (e *BaseEvent) EventName() string {
	return e.Name
}

// OccurredAt returns when the event occurred
func (e *BaseEvent) OccurredAt() int64 {
	return e.Timestamp
}

// NewBaseEvent creates a new base event
func NewBaseEvent(name string) *BaseEvent {
	return &BaseEvent{
		Name:      name,
		Timestamp: time.Now().Unix(),
	}
}

// ItemRegisteredEvent is fired when a new item is registered in the market
type ItemRegisteredEvent struct {
	*BaseEvent
	ItemID   string
	ItemName string
	Category string
	Price    int
}

// NewItemRegisteredEvent creates a new item registered event
func NewItemRegisteredEvent(itemID, itemName, category string, price int) *ItemRegisteredEvent {
	return &ItemRegisteredEvent{
		BaseEvent: NewBaseEvent(EventNameItemRegistered),
		ItemID:    itemID,
		ItemName:  itemName,
		Category:  category,
		Price:     price,
	}
}

// PriceUpdatedEvent is fired when an item's price changes
type PriceUpdatedEvent struct {
	*BaseEvent
	ItemID   string
	OldPrice int
	NewPrice int
	Reason   string
}

// NewPriceUpdatedEvent creates a new price updated event
func NewPriceUpdatedEvent(itemID string, oldPrice, newPrice int, reason string) *PriceUpdatedEvent {
	return &PriceUpdatedEvent{
		BaseEvent: NewBaseEvent(EventNamePriceUpdated),
		ItemID:    itemID,
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
		Reason:    reason,
	}
}

// MarketEventOccurredEvent is fired when a special market event occurs
type MarketEventOccurredEvent struct {
	*BaseEvent
	EventType   string
	Description string
	Impact      map[string]interface{}
}

// NewMarketEventOccurredEvent creates a new market event occurred event
func NewMarketEventOccurredEvent(eventType, description string, impact map[string]interface{}) *MarketEventOccurredEvent {
	return &MarketEventOccurredEvent{
		BaseEvent:   NewBaseEvent(EventNameMarketEventOccurred),
		EventType:   eventType,
		Description: description,
		Impact:      impact,
	}
}

// InventoryChangedEvent is fired when inventory changes
type InventoryChangedEvent struct {
	*BaseEvent
	ItemID       string
	OldQuantity  int
	NewQuantity  int
	ChangeReason string
}

// NewInventoryChangedEvent creates a new inventory changed event
func NewInventoryChangedEvent(itemID string, oldQty, newQty int, reason string) *InventoryChangedEvent {
	return &InventoryChangedEvent{
		BaseEvent:    NewBaseEvent(EventNameInventoryChanged),
		ItemID:       itemID,
		OldQuantity:  oldQty,
		NewQuantity:  newQty,
		ChangeReason: reason,
	}
}

// TransactionCompleteEvent is fired when a transaction is completed
type TransactionCompleteEvent struct {
	*BaseEvent
	TransactionID string
	Type          string // "buy" or "sell"
	ItemID        string
	Quantity      int
	TotalPrice    int
	PartyID       string // ID of the merchant or customer
}

// NewTransactionCompleteEvent creates a new transaction complete event
func NewTransactionCompleteEvent(txID, txType, itemID string, qty, price int, partyID string) *TransactionCompleteEvent {
	return &TransactionCompleteEvent{
		BaseEvent:     NewBaseEvent(EventNameTransactionComplete),
		TransactionID: txID,
		Type:          txType,
		ItemID:        itemID,
		Quantity:      qty,
		TotalPrice:    price,
		PartyID:       partyID,
	}
}

// MerchantActionEvent is fired when an AI merchant takes an action
type MerchantActionEvent struct {
	*BaseEvent
	MerchantID   string
	ActionType   string
	TargetItemID string
	Details      map[string]interface{}
}

// NewMerchantActionEvent creates a new merchant action event
func NewMerchantActionEvent(merchantID, actionType, targetItemID string, details map[string]interface{}) *MerchantActionEvent {
	return &MerchantActionEvent{
		BaseEvent:    NewBaseEvent(EventNameMerchantAction),
		MerchantID:   merchantID,
		ActionType:   actionType,
		TargetItemID: targetItemID,
		Details:      details,
	}
}

// SeasonChangedEvent is fired when the season changes
type SeasonChangedEvent struct {
	*BaseEvent
	OldSeason string
	NewSeason string
	Effects   map[string]float64
}

// NewSeasonChangedEvent creates a new season changed event
func NewSeasonChangedEvent(oldSeason, newSeason string, effects map[string]float64) *SeasonChangedEvent {
	return &SeasonChangedEvent{
		BaseEvent: NewBaseEvent(EventNameSeasonChanged),
		OldSeason: oldSeason,
		NewSeason: newSeason,
		Effects:   effects,
	}
}

// DayEndedEvent is fired at the end of each game day
type DayEndedEvent struct {
	*BaseEvent
	DayNumber      int
	TotalSales     int
	TotalPurchases int
	NetProfit      int
}

// NewDayEndedEvent creates a new day ended event
func NewDayEndedEvent(dayNum, sales, purchases, profit int) *DayEndedEvent {
	return &DayEndedEvent{
		BaseEvent:      NewBaseEvent(EventNameDayEnded),
		DayNumber:      dayNum,
		TotalSales:     sales,
		TotalPurchases: purchases,
		NetProfit:      profit,
	}
}
