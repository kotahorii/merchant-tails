package inventory

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// CapacityManager manages inventory capacity and optimization
type CapacityManager struct {
	baseShopCapacity           int
	baseWarehouseCapacity      int
	shopCapacityModifiers      map[string]float64
	warehouseCapacityModifiers map[string]float64
	utilizationHistory         []UtilizationRecord
	capacityAlerts             []CapacityAlert
	autoExpandEnabled          bool
	maxShopCapacity            int
	maxWarehouseCapacity       int
	mu                         sync.RWMutex
}

// UtilizationRecord tracks capacity usage over time
type UtilizationRecord struct {
	Timestamp            time.Time
	ShopUtilization      float64
	WarehouseUtilization float64
	ShopItems            int
	WarehouseItems       int
	ShopCapacity         int
	WarehouseCapacity    int
}

// CapacityAlert represents a capacity-related alert
type CapacityAlert struct {
	Type        CapacityAlertType
	Location    InventoryLocation
	Severity    AlertSeverity
	Message     string
	Timestamp   time.Time
	Utilization float64
}

// CapacityAlertType represents the type of capacity alert
type CapacityAlertType int

const (
	AlertTypeHighUtilization CapacityAlertType = iota
	AlertTypeFull
	AlertTypeInefficient
	AlertTypeRebalanceNeeded
)

// AlertSeverity represents the severity of an alert
type AlertSeverity int

const (
	SeverityInfo AlertSeverity = iota
	SeverityWarning
	SeverityCritical
)

// CapacityConfig holds capacity configuration
type CapacityConfig struct {
	BaseShopCapacity      int
	BaseWarehouseCapacity int
	MaxShopCapacity       int
	MaxWarehouseCapacity  int
	AutoExpandEnabled     bool
}

// CapacityStats provides capacity statistics
type CapacityStats struct {
	CurrentShopCapacity          int
	CurrentWarehouseCapacity     int
	ShopUtilization              float64
	WarehouseUtilization         float64
	TotalUtilization             float64
	AvgShopUtilization           float64
	AvgWarehouseUtilization      float64
	PeakShopUtilization          float64
	PeakWarehouseUtilization     float64
	RecommendedShopCapacity      int
	RecommendedWarehouseCapacity int
}

// NewCapacityManager creates a new capacity manager
func NewCapacityManager(config *CapacityConfig) *CapacityManager {
	if config == nil {
		config = &CapacityConfig{
			BaseShopCapacity:      100,
			BaseWarehouseCapacity: 500,
			MaxShopCapacity:       1000,
			MaxWarehouseCapacity:  5000,
			AutoExpandEnabled:     false,
		}
	}

	return &CapacityManager{
		baseShopCapacity:           config.BaseShopCapacity,
		baseWarehouseCapacity:      config.BaseWarehouseCapacity,
		maxShopCapacity:            config.MaxShopCapacity,
		maxWarehouseCapacity:       config.MaxWarehouseCapacity,
		autoExpandEnabled:          config.AutoExpandEnabled,
		shopCapacityModifiers:      make(map[string]float64),
		warehouseCapacityModifiers: make(map[string]float64),
		utilizationHistory:         make([]UtilizationRecord, 0),
		capacityAlerts:             make([]CapacityAlert, 0),
	}
}

// GetShopCapacity returns the current shop capacity with all modifiers
func (cm *CapacityManager) GetShopCapacity() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	capacity := float64(cm.baseShopCapacity)
	for _, modifier := range cm.shopCapacityModifiers {
		capacity *= modifier
	}

	finalCapacity := int(math.Ceil(capacity))
	if finalCapacity > cm.maxShopCapacity {
		return cm.maxShopCapacity
	}
	return finalCapacity
}

// GetWarehouseCapacity returns the current warehouse capacity with all modifiers
func (cm *CapacityManager) GetWarehouseCapacity() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	capacity := float64(cm.baseWarehouseCapacity)
	for _, modifier := range cm.warehouseCapacityModifiers {
		capacity *= modifier
	}

	finalCapacity := int(math.Ceil(capacity))
	if finalCapacity > cm.maxWarehouseCapacity {
		return cm.maxWarehouseCapacity
	}
	return finalCapacity
}

// AddCapacityModifier adds a capacity modifier
func (cm *CapacityManager) AddCapacityModifier(location InventoryLocation, name string, modifier float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if location == LocationShop {
		cm.shopCapacityModifiers[name] = modifier
	} else {
		cm.warehouseCapacityModifiers[name] = modifier
	}
}

// RemoveCapacityModifier removes a capacity modifier
func (cm *CapacityManager) RemoveCapacityModifier(location InventoryLocation, name string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if location == LocationShop {
		delete(cm.shopCapacityModifiers, name)
	} else {
		delete(cm.warehouseCapacityModifiers, name)
	}
}

// RecordUtilization records current capacity utilization
func (cm *CapacityManager) RecordUtilization(shopItems, warehouseItems int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	shopCapacity := cm.GetShopCapacity()
	warehouseCapacity := cm.GetWarehouseCapacity()

	record := UtilizationRecord{
		Timestamp:            time.Now(),
		ShopItems:            shopItems,
		WarehouseItems:       warehouseItems,
		ShopCapacity:         shopCapacity,
		WarehouseCapacity:    warehouseCapacity,
		ShopUtilization:      float64(shopItems) / float64(shopCapacity),
		WarehouseUtilization: float64(warehouseItems) / float64(warehouseCapacity),
	}

	cm.utilizationHistory = append(cm.utilizationHistory, record)

	// Keep only last 100 records
	if len(cm.utilizationHistory) > 100 {
		cm.utilizationHistory = cm.utilizationHistory[len(cm.utilizationHistory)-100:]
	}

	// Check for alerts
	cm.checkCapacityAlerts(record)
}

// checkCapacityAlerts checks for capacity-related alerts
func (cm *CapacityManager) checkCapacityAlerts(record UtilizationRecord) {
	// Clear old alerts
	cm.capacityAlerts = []CapacityAlert{}

	// Check shop utilization
	if record.ShopUtilization > 0.9 {
		cm.capacityAlerts = append(cm.capacityAlerts, CapacityAlert{
			Type:        AlertTypeFull,
			Location:    LocationShop,
			Severity:    SeverityCritical,
			Message:     fmt.Sprintf("Shop capacity critical: %.1f%% utilized", record.ShopUtilization*100),
			Timestamp:   time.Now(),
			Utilization: record.ShopUtilization,
		})
	} else if record.ShopUtilization > 0.75 {
		cm.capacityAlerts = append(cm.capacityAlerts, CapacityAlert{
			Type:        AlertTypeHighUtilization,
			Location:    LocationShop,
			Severity:    SeverityWarning,
			Message:     fmt.Sprintf("Shop capacity high: %.1f%% utilized", record.ShopUtilization*100),
			Timestamp:   time.Now(),
			Utilization: record.ShopUtilization,
		})
	}

	// Check warehouse utilization
	if record.WarehouseUtilization > 0.9 {
		cm.capacityAlerts = append(cm.capacityAlerts, CapacityAlert{
			Type:        AlertTypeFull,
			Location:    LocationWarehouse,
			Severity:    SeverityCritical,
			Message:     fmt.Sprintf("Warehouse capacity critical: %.1f%% utilized", record.WarehouseUtilization*100),
			Timestamp:   time.Now(),
			Utilization: record.WarehouseUtilization,
		})
	} else if record.WarehouseUtilization > 0.75 {
		cm.capacityAlerts = append(cm.capacityAlerts, CapacityAlert{
			Type:        AlertTypeHighUtilization,
			Location:    LocationWarehouse,
			Severity:    SeverityWarning,
			Message:     fmt.Sprintf("Warehouse capacity high: %.1f%% utilized", record.WarehouseUtilization*100),
			Timestamp:   time.Now(),
			Utilization: record.WarehouseUtilization,
		})
	}

	// Check for inefficient usage
	if record.ShopUtilization < 0.2 && record.WarehouseUtilization > 0.5 {
		cm.capacityAlerts = append(cm.capacityAlerts, CapacityAlert{
			Type:        AlertTypeRebalanceNeeded,
			Location:    LocationShop,
			Severity:    SeverityInfo,
			Message:     "Consider moving items from warehouse to shop",
			Timestamp:   time.Now(),
			Utilization: record.ShopUtilization,
		})
	}
}

// GetCapacityStats returns capacity statistics
func (cm *CapacityManager) GetCapacityStats(currentShopItems, currentWarehouseItems int) *CapacityStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	shopCapacity := cm.GetShopCapacity()
	warehouseCapacity := cm.GetWarehouseCapacity()

	stats := &CapacityStats{
		CurrentShopCapacity:      shopCapacity,
		CurrentWarehouseCapacity: warehouseCapacity,
		ShopUtilization:          float64(currentShopItems) / float64(shopCapacity),
		WarehouseUtilization:     float64(currentWarehouseItems) / float64(warehouseCapacity),
	}

	totalCapacity := shopCapacity + warehouseCapacity
	totalItems := currentShopItems + currentWarehouseItems
	stats.TotalUtilization = float64(totalItems) / float64(totalCapacity)

	// Calculate averages and peaks from history
	if len(cm.utilizationHistory) > 0 {
		var sumShop, sumWarehouse float64
		peakShop := 0.0
		peakWarehouse := 0.0

		for _, record := range cm.utilizationHistory {
			sumShop += record.ShopUtilization
			sumWarehouse += record.WarehouseUtilization

			if record.ShopUtilization > peakShop {
				peakShop = record.ShopUtilization
			}
			if record.WarehouseUtilization > peakWarehouse {
				peakWarehouse = record.WarehouseUtilization
			}
		}

		stats.AvgShopUtilization = sumShop / float64(len(cm.utilizationHistory))
		stats.AvgWarehouseUtilization = sumWarehouse / float64(len(cm.utilizationHistory))
		stats.PeakShopUtilization = peakShop
		stats.PeakWarehouseUtilization = peakWarehouse

		// Calculate recommended capacities based on peak usage
		stats.RecommendedShopCapacity = int(float64(shopCapacity) * (peakShop / 0.75))
		stats.RecommendedWarehouseCapacity = int(float64(warehouseCapacity) * (peakWarehouse / 0.75))

		// Ensure recommendations don't exceed maximums
		if stats.RecommendedShopCapacity > cm.maxShopCapacity {
			stats.RecommendedShopCapacity = cm.maxShopCapacity
		}
		if stats.RecommendedWarehouseCapacity > cm.maxWarehouseCapacity {
			stats.RecommendedWarehouseCapacity = cm.maxWarehouseCapacity
		}
	}

	return stats
}

// GetCapacityAlerts returns current capacity alerts
func (cm *CapacityManager) GetCapacityAlerts() []CapacityAlert {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	alerts := make([]CapacityAlert, len(cm.capacityAlerts))
	copy(alerts, cm.capacityAlerts)
	return alerts
}

// UpgradeCapacity upgrades base capacity
func (cm *CapacityManager) UpgradeCapacity(location InventoryLocation, amount int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if amount <= 0 {
		return errors.New("upgrade amount must be positive")
	}

	if location == LocationShop {
		newCapacity := cm.baseShopCapacity + amount
		if newCapacity > cm.maxShopCapacity {
			return fmt.Errorf("would exceed maximum shop capacity of %d", cm.maxShopCapacity)
		}
		cm.baseShopCapacity = newCapacity
	} else {
		newCapacity := cm.baseWarehouseCapacity + amount
		if newCapacity > cm.maxWarehouseCapacity {
			return fmt.Errorf("would exceed maximum warehouse capacity of %d", cm.maxWarehouseCapacity)
		}
		cm.baseWarehouseCapacity = newCapacity
	}

	return nil
}

// ShouldAutoExpand determines if capacity should be auto-expanded
func (cm *CapacityManager) ShouldAutoExpand(location InventoryLocation) bool {
	if !cm.autoExpandEnabled {
		return false
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Check if we have enough history
	if len(cm.utilizationHistory) < 10 {
		return false
	}

	// Calculate recent average utilization
	recentRecords := cm.utilizationHistory[len(cm.utilizationHistory)-10:]
	var avgUtil float64

	for _, record := range recentRecords {
		if location == LocationShop {
			avgUtil += record.ShopUtilization
		} else {
			avgUtil += record.WarehouseUtilization
		}
	}
	avgUtil /= float64(len(recentRecords))

	// Auto-expand if average utilization is above 80%
	return avgUtil > 0.8
}

// CalculateOptimalTransfer calculates optimal item transfer between locations
func (cm *CapacityManager) CalculateOptimalTransfer(shopItems, warehouseItems int) (toShop, toWarehouse int) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	shopCapacity := cm.GetShopCapacity()
	warehouseCapacity := cm.GetWarehouseCapacity()

	shopUtil := float64(shopItems) / float64(shopCapacity)
	warehouseUtil := float64(warehouseItems) / float64(warehouseCapacity)

	// Balance utilization between locations
	targetUtil := (shopUtil + warehouseUtil) / 2

	targetShopItems := int(targetUtil * float64(shopCapacity))
	targetWarehouseItems := int(targetUtil * float64(warehouseCapacity))

	// Calculate transfers
	if shopItems < targetShopItems && warehouseItems > targetWarehouseItems {
		// Transfer from warehouse to shop
		toShop = targetShopItems - shopItems
		if toShop > warehouseItems-targetWarehouseItems {
			toShop = warehouseItems - targetWarehouseItems
		}
	} else if warehouseItems < targetWarehouseItems && shopItems > targetShopItems {
		// Transfer from shop to warehouse
		toWarehouse = targetWarehouseItems - warehouseItems
		if toWarehouse > shopItems-targetShopItems {
			toWarehouse = shopItems - targetShopItems
		}
	}

	return toShop, toWarehouse
}

// GetUtilizationHistory returns utilization history
func (cm *CapacityManager) GetUtilizationHistory() []UtilizationRecord {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	history := make([]UtilizationRecord, len(cm.utilizationHistory))
	copy(history, cm.utilizationHistory)
	return history
}

// SetAutoExpand enables or disables auto-expansion
func (cm *CapacityManager) SetAutoExpand(enabled bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.autoExpandEnabled = enabled
}

// Reset resets the capacity manager to initial state
func (cm *CapacityManager) Reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.shopCapacityModifiers = make(map[string]float64)
	cm.warehouseCapacityModifiers = make(map[string]float64)
	cm.utilizationHistory = []UtilizationRecord{}
	cm.capacityAlerts = []CapacityAlert{}
}
