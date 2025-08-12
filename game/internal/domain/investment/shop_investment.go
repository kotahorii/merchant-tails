package investment

import (
	"fmt"
	"sync"
	"time"
)

// ShopInvestment represents the shop investment system
type ShopInvestment struct {
	level           int
	capacity        int
	efficiency      float64
	maintenanceCost int
	upgrades        map[string]*Upgrade
	equipment       map[string]*Equipment
	totalInvested   int
	monthlyRevenue  int
	mu              sync.RWMutex
}

// Upgrade represents a shop upgrade
type Upgrade struct {
	ID          string
	Name        string
	Description string
	Cost        int
	Level       int
	MaxLevel    int
	Effects     UpgradeEffects
	Unlocked    bool
	Purchased   bool
}

// UpgradeEffects represents the effects of an upgrade
type UpgradeEffects struct {
	CapacityBonus    int     // Additional storage capacity
	EfficiencyBonus  float64 // Sales efficiency multiplier
	MaintenanceReduc float64 // Maintenance cost reduction
	RevenueBonus     float64 // Revenue multiplier
	DiscountRate     float64 // Purchase discount rate
}

// Equipment represents shop equipment
type Equipment struct {
	ID              string
	Name            string
	Description     string
	Cost            int
	MaintenanceCost int
	DurabilityMax   int
	DurabilityCur   int
	Effects         EquipmentEffects
	Purchased       bool
	Active          bool
}

// EquipmentEffects represents the effects of equipment
type EquipmentEffects struct {
	StorageBonus     int     // Extra storage slots
	PreservationRate float64 // Reduces item decay rate
	DisplayBonus     float64 // Increases customer attraction
	ProcessingSpeed  float64 // Faster transaction processing
	SecurityLevel    int     // Reduces theft risk
}

// InvestmentReturn represents ROI calculation
type InvestmentReturn struct {
	TotalInvested int
	TotalReturns  int
	ROI           float64
	PaybackPeriod int // In days
	MonthlyProfit int
	BreakEvenDate time.Time
}

// NewShopInvestment creates a new shop investment system
func NewShopInvestment() *ShopInvestment {
	return &ShopInvestment{
		level:           1,
		capacity:        20, // Base capacity
		efficiency:      1.0,
		maintenanceCost: 50, // Base maintenance per day
		upgrades:        initializeUpgrades(),
		equipment:       initializeEquipment(),
		totalInvested:   0,
		monthlyRevenue:  0,
		mu:              sync.RWMutex{},
	}
}

// initializeUpgrades creates the upgrade tree
func initializeUpgrades() map[string]*Upgrade {
	return map[string]*Upgrade{
		"storage_small": {
			ID:          "storage_small",
			Name:        "Small Storage Expansion",
			Description: "Adds 10 storage slots",
			Cost:        500,
			Level:       0,
			MaxLevel:    5,
			Effects: UpgradeEffects{
				CapacityBonus: 10,
			},
			Unlocked: true,
		},
		"storage_large": {
			ID:          "storage_large",
			Name:        "Large Storage Expansion",
			Description: "Adds 25 storage slots",
			Cost:        2000,
			Level:       0,
			MaxLevel:    3,
			Effects: UpgradeEffects{
				CapacityBonus: 25,
			},
			Unlocked: false, // Unlocks at shop level 2
		},
		"efficiency_training": {
			ID:          "efficiency_training",
			Name:        "Staff Training",
			Description: "Improves sales efficiency by 10%",
			Cost:        1000,
			Level:       0,
			MaxLevel:    10,
			Effects: UpgradeEffects{
				EfficiencyBonus: 0.1,
			},
			Unlocked: true,
		},
		"premium_location": {
			ID:          "premium_location",
			Name:        "Premium Location",
			Description: "Move to a better location for increased traffic",
			Cost:        5000,
			Level:       0,
			MaxLevel:    1,
			Effects: UpgradeEffects{
				RevenueBonus:     0.25,
				MaintenanceReduc: -0.2, // Actually increases maintenance
			},
			Unlocked: false, // Unlocks at shop level 3
		},
		"bulk_discount": {
			ID:          "bulk_discount",
			Name:        "Supplier Partnership",
			Description: "Get 5% discount on all purchases",
			Cost:        3000,
			Level:       0,
			MaxLevel:    5,
			Effects: UpgradeEffects{
				DiscountRate: 0.05,
			},
			Unlocked: false, // Unlocks at shop level 2
		},
	}
}

// initializeEquipment creates available equipment
func initializeEquipment() map[string]*Equipment {
	return map[string]*Equipment{
		"display_case": {
			ID:              "display_case",
			Name:            "Display Case",
			Description:     "Attractive display for premium items",
			Cost:            800,
			MaintenanceCost: 10,
			DurabilityMax:   100,
			DurabilityCur:   100,
			Effects: EquipmentEffects{
				DisplayBonus: 0.15,
				StorageBonus: 5,
			},
		},
		"preservation_unit": {
			ID:              "preservation_unit",
			Name:            "Preservation Unit",
			Description:     "Keeps perishable items fresh longer",
			Cost:            1500,
			MaintenanceCost: 25,
			DurabilityMax:   150,
			DurabilityCur:   150,
			Effects: EquipmentEffects{
				PreservationRate: 0.5, // 50% slower decay
			},
		},
		"cash_register": {
			ID:              "cash_register",
			Name:            "Advanced Cash Register",
			Description:     "Speeds up transactions",
			Cost:            600,
			MaintenanceCost: 5,
			DurabilityMax:   200,
			DurabilityCur:   200,
			Effects: EquipmentEffects{
				ProcessingSpeed: 1.3,
			},
		},
		"security_system": {
			ID:              "security_system",
			Name:            "Security System",
			Description:     "Reduces theft and losses",
			Cost:            2000,
			MaintenanceCost: 30,
			DurabilityMax:   300,
			DurabilityCur:   300,
			Effects: EquipmentEffects{
				SecurityLevel: 3,
			},
		},
		"warehouse_extension": {
			ID:              "warehouse_extension",
			Name:            "Warehouse Extension",
			Description:     "Large storage expansion",
			Cost:            5000,
			MaintenanceCost: 50,
			DurabilityMax:   500,
			DurabilityCur:   500,
			Effects: EquipmentEffects{
				StorageBonus: 50,
			},
		},
	}
}

// PurchaseUpgrade purchases an upgrade
func (si *ShopInvestment) PurchaseUpgrade(upgradeID string, gold int) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	upgrade, exists := si.upgrades[upgradeID]
	if !exists {
		return fmt.Errorf("upgrade %s does not exist", upgradeID)
	}

	if !upgrade.Unlocked {
		return fmt.Errorf("upgrade %s is not unlocked yet", upgradeID)
	}

	if upgrade.Level >= upgrade.MaxLevel {
		return fmt.Errorf("upgrade %s is already at max level", upgradeID)
	}

	// Calculate cost for current level
	cost := upgrade.Cost * (upgrade.Level + 1)
	if gold < cost {
		return fmt.Errorf("insufficient gold: need %d, have %d", cost, gold)
	}

	// Apply upgrade
	upgrade.Level++
	upgrade.Purchased = true
	si.applyUpgradeEffects(upgrade)
	si.totalInvested += cost

	// Check for unlock conditions
	si.checkUnlocks()

	return nil
}

// PurchaseEquipment purchases equipment
func (si *ShopInvestment) PurchaseEquipment(equipmentID string, gold int) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	equipment, exists := si.equipment[equipmentID]
	if !exists {
		return fmt.Errorf("equipment %s does not exist", equipmentID)
	}

	if equipment.Purchased {
		return fmt.Errorf("equipment %s already purchased", equipmentID)
	}

	if gold < equipment.Cost {
		return fmt.Errorf("insufficient gold: need %d, have %d", equipment.Cost, gold)
	}

	// Purchase equipment
	equipment.Purchased = true
	equipment.Active = true
	si.applyEquipmentEffects(equipment)
	si.totalInvested += equipment.Cost

	// Add to maintenance cost
	si.maintenanceCost += equipment.MaintenanceCost

	return nil
}

// applyUpgradeEffects applies the effects of an upgrade
func (si *ShopInvestment) applyUpgradeEffects(upgrade *Upgrade) {
	effects := upgrade.Effects

	si.capacity += effects.CapacityBonus
	si.efficiency += effects.EfficiencyBonus

	if effects.MaintenanceReduc != 0 {
		si.maintenanceCost = int(float64(si.maintenanceCost) * (1 - effects.MaintenanceReduc))
	}
}

// applyEquipmentEffects applies the effects of equipment
func (si *ShopInvestment) applyEquipmentEffects(equipment *Equipment) {
	effects := equipment.Effects

	si.capacity += effects.StorageBonus

	// Other effects would be applied to relevant systems
	// For example, preservation rate would affect item decay
	// Display bonus would affect customer attraction
}

// checkUnlocks checks and unlocks new upgrades based on conditions
func (si *ShopInvestment) checkUnlocks() {
	// Unlock based on shop level
	if si.level >= 2 {
		si.upgrades["storage_large"].Unlocked = true
		si.upgrades["bulk_discount"].Unlocked = true
	}

	if si.level >= 3 {
		si.upgrades["premium_location"].Unlocked = true
	}

	// Unlock based on total investment
	// TODO: Unlock special upgrades when totalInvested >= 10000
}

// UpgradeShopLevel upgrades the shop to the next level
func (si *ShopInvestment) UpgradeShopLevel(gold int) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	// Calculate upgrade cost
	upgradeCost := si.level * 5000

	if gold < upgradeCost {
		return fmt.Errorf("insufficient gold: need %d, have %d", upgradeCost, gold)
	}

	si.level++
	si.capacity += 20    // Bonus capacity per level
	si.efficiency += 0.1 // 10% efficiency boost per level
	si.totalInvested += upgradeCost

	// Check for new unlocks
	si.checkUnlocks()

	return nil
}

// CalculateROI calculates return on investment
func (si *ShopInvestment) CalculateROI(currentRevenue, baseRevenue int) *InvestmentReturn {
	si.mu.RLock()
	defer si.mu.RUnlock()

	// Calculate total returns
	additionalRevenue := currentRevenue - baseRevenue
	monthlyProfit := additionalRevenue - si.maintenanceCost*30

	// Calculate ROI percentage
	roi := 0.0
	if si.totalInvested > 0 {
		roi = float64(monthlyProfit*12) / float64(si.totalInvested) * 100
	}

	// Calculate payback period
	paybackDays := 0
	if monthlyProfit > 0 {
		paybackDays = si.totalInvested / (monthlyProfit / 30)
	}

	return &InvestmentReturn{
		TotalInvested: si.totalInvested,
		TotalReturns:  additionalRevenue * 12, // Annual
		ROI:           roi,
		PaybackPeriod: paybackDays,
		MonthlyProfit: monthlyProfit,
		BreakEvenDate: time.Now().AddDate(0, 0, paybackDays),
	}
}

// GetMaintenanceCost returns daily maintenance cost
func (si *ShopInvestment) GetMaintenanceCost() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.maintenanceCost
}

// GetCapacity returns current shop capacity
func (si *ShopInvestment) GetCapacity() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.capacity
}

// GetEfficiency returns current efficiency multiplier
func (si *ShopInvestment) GetEfficiency() float64 {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.efficiency
}

// GetLevel returns current shop level
func (si *ShopInvestment) GetLevel() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.level
}

// GetTotalInvested returns total amount invested
func (si *ShopInvestment) GetTotalInvested() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.totalInvested
}

// GetAvailableUpgrades returns list of available upgrades
func (si *ShopInvestment) GetAvailableUpgrades() []*Upgrade {
	si.mu.RLock()
	defer si.mu.RUnlock()

	var available []*Upgrade
	for _, upgrade := range si.upgrades {
		if upgrade.Unlocked && upgrade.Level < upgrade.MaxLevel {
			available = append(available, upgrade)
		}
	}
	return available
}

// GetAvailableEquipment returns list of available equipment
func (si *ShopInvestment) GetAvailableEquipment() []*Equipment {
	si.mu.RLock()
	defer si.mu.RUnlock()

	var available []*Equipment
	for _, equipment := range si.equipment {
		if !equipment.Purchased {
			available = append(available, equipment)
		}
	}
	return available
}

// RepairEquipment repairs damaged equipment
func (si *ShopInvestment) RepairEquipment(equipmentID string, gold int) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	equipment, exists := si.equipment[equipmentID]
	if !exists {
		return fmt.Errorf("equipment %s does not exist", equipmentID)
	}

	if !equipment.Purchased {
		return fmt.Errorf("equipment %s not purchased", equipmentID)
	}

	// Calculate repair cost based on damage
	damage := equipment.DurabilityMax - equipment.DurabilityCur
	repairCost := int(float64(equipment.Cost) * float64(damage) / float64(equipment.DurabilityMax) * 0.3)

	if gold < repairCost {
		return fmt.Errorf("insufficient gold: need %d, have %d", repairCost, gold)
	}

	equipment.DurabilityCur = equipment.DurabilityMax
	equipment.Active = true

	return nil
}

// SimulateInvestmentScenario simulates investment returns over time
func (si *ShopInvestment) SimulateInvestmentScenario(investmentAmount int, months int) []float64 {
	si.mu.RLock()
	defer si.mu.RUnlock()

	returns := make([]float64, months)

	// Simple compound growth simulation
	baseReturn := 0.05 // 5% monthly return
	efficiency := si.efficiency

	for i := 0; i < months; i++ {
		monthlyReturn := float64(investmentAmount) * baseReturn * efficiency
		returns[i] = monthlyReturn

		// Compound effect
		investmentAmount += int(monthlyReturn)
	}

	return returns
}

// GetInvestmentAdvice provides investment recommendations
func (si *ShopInvestment) GetInvestmentAdvice(currentGold int, monthlyIncome int) string {
	si.mu.RLock()
	defer si.mu.RUnlock()

	advice := "Investment Recommendations:\n"

	// Check if maintenance is too high
	if si.maintenanceCost > monthlyIncome/10 {
		advice += "⚠️ Maintenance costs are high. Consider efficiency upgrades.\n"
	}

	// Check capacity utilization (would need actual inventory data)
	if si.capacity < 50 {
		advice += "📦 Low capacity. Storage upgrades recommended for growth.\n"
	}

	// ROI-based advice
	potentialROI := (si.efficiency - 1.0) * 100
	if potentialROI < 10 {
		advice += "📈 Focus on efficiency upgrades for better returns.\n"
	}

	// Level-based advice
	if si.level < 3 && currentGold > si.level*5000 {
		advice += fmt.Sprintf("⬆️ Shop upgrade available! Cost: %d gold\n", si.level*5000)
	}

	// Equipment recommendations
	if currentGold > 2000 {
		advice += "🛠️ Consider equipment purchases for passive bonuses.\n"
	}

	return advice
}
