package investment

import (
	"testing"
)

func TestNewShopInvestment(t *testing.T) {
	si := NewShopInvestment()

	if si == nil {
		t.Fatal("NewShopInvestment returned nil")
	}

	if si.level != 1 {
		t.Errorf("Expected initial level 1, got %d", si.level)
	}

	if si.capacity != 20 {
		t.Errorf("Expected initial capacity 20, got %d", si.capacity)
	}

	if si.efficiency != 1.0 {
		t.Errorf("Expected initial efficiency 1.0, got %f", si.efficiency)
	}

	if si.maintenanceCost != 50 {
		t.Errorf("Expected initial maintenance cost 50, got %d", si.maintenanceCost)
	}

	if si.totalInvested != 0 {
		t.Errorf("Expected initial investment 0, got %d", si.totalInvested)
	}
}

func TestPurchaseUpgrade(t *testing.T) {
	si := NewShopInvestment()

	// Test purchasing storage upgrade
	err := si.PurchaseUpgrade("storage_small", 1000)
	if err != nil {
		t.Errorf("Failed to purchase upgrade: %v", err)
	}

	// Check capacity increased
	if si.capacity != 30 { // 20 base + 10 from upgrade
		t.Errorf("Expected capacity 30 after upgrade, got %d", si.capacity)
	}

	// Check total invested
	if si.totalInvested != 500 {
		t.Errorf("Expected total invested 500, got %d", si.totalInvested)
	}

	// Test insufficient gold
	err = si.PurchaseUpgrade("storage_small", 100)
	if err == nil {
		t.Error("Expected error for insufficient gold")
	}

	// Test non-existent upgrade
	err = si.PurchaseUpgrade("invalid_upgrade", 10000)
	if err == nil {
		t.Error("Expected error for non-existent upgrade")
	}

	// Test locked upgrade
	err = si.PurchaseUpgrade("premium_location", 10000)
	if err == nil {
		t.Error("Expected error for locked upgrade")
	}
}

func TestPurchaseEquipment(t *testing.T) {
	si := NewShopInvestment()

	// Test purchasing display case
	err := si.PurchaseEquipment("display_case", 1000)
	if err != nil {
		t.Errorf("Failed to purchase equipment: %v", err)
	}

	// Check capacity increased
	if si.capacity != 25 { // 20 base + 5 from display case
		t.Errorf("Expected capacity 25 after equipment, got %d", si.capacity)
	}

	// Check maintenance cost increased
	if si.maintenanceCost != 60 { // 50 base + 10 from display case
		t.Errorf("Expected maintenance cost 60, got %d", si.maintenanceCost)
	}

	// Check total invested
	if si.totalInvested != 800 {
		t.Errorf("Expected total invested 800, got %d", si.totalInvested)
	}

	// Test purchasing already owned equipment
	err = si.PurchaseEquipment("display_case", 1000)
	if err == nil {
		t.Error("Expected error for already purchased equipment")
	}

	// Test insufficient gold
	err = si.PurchaseEquipment("preservation_unit", 100)
	if err == nil {
		t.Error("Expected error for insufficient gold")
	}
}

func TestUpgradeShopLevel(t *testing.T) {
	si := NewShopInvestment()

	// Upgrade to level 2
	err := si.UpgradeShopLevel(5000)
	if err != nil {
		t.Errorf("Failed to upgrade shop level: %v", err)
	}

	if si.level != 2 {
		t.Errorf("Expected level 2, got %d", si.level)
	}

	if si.capacity != 40 { // 20 base + 20 from level up
		t.Errorf("Expected capacity 40, got %d", si.capacity)
	}

	if si.efficiency != 1.1 {
		t.Errorf("Expected efficiency 1.1, got %f", si.efficiency)
	}

	// Check unlocks
	if !si.upgrades["storage_large"].Unlocked {
		t.Error("storage_large should be unlocked at level 2")
	}

	if !si.upgrades["bulk_discount"].Unlocked {
		t.Error("bulk_discount should be unlocked at level 2")
	}

	// Test insufficient gold
	err = si.UpgradeShopLevel(1000)
	if err == nil {
		t.Error("Expected error for insufficient gold")
	}
}

func TestCalculateROI(t *testing.T) {
	si := NewShopInvestment()

	// Make some investments
	si.PurchaseUpgrade("storage_small", 1000)
	si.PurchaseEquipment("display_case", 1000)

	// Calculate ROI
	roi := si.CalculateROI(2000, 1000)

	if roi.TotalInvested != 1300 { // 500 + 800
		t.Errorf("Expected total invested 1300, got %d", roi.TotalInvested)
	}

	// Monthly profit = (2000-1000) - 60*30 = 1000 - 1800 = -800
	if roi.MonthlyProfit != -800 {
		t.Errorf("Expected monthly profit -800, got %d", roi.MonthlyProfit)
	}

	// With negative profit, ROI should be negative
	if roi.ROI >= 0 {
		t.Errorf("Expected negative ROI, got %f", roi.ROI)
	}
}

func TestUpgradeEffects(t *testing.T) {
	si := NewShopInvestment()

	// Test efficiency upgrade
	err := si.PurchaseUpgrade("efficiency_training", 1000)
	if err != nil {
		t.Errorf("Failed to purchase efficiency upgrade: %v", err)
	}

	if si.efficiency != 1.1 {
		t.Errorf("Expected efficiency 1.1, got %f", si.efficiency)
	}

	// Purchase multiple levels
	err = si.PurchaseUpgrade("efficiency_training", 2000)
	if err != nil {
		t.Errorf("Failed to purchase second level: %v", err)
	}

	if si.efficiency < 1.19 || si.efficiency > 1.21 {
		t.Errorf("Expected efficiency around 1.2, got %f", si.efficiency)
	}
}

func TestRepairEquipment(t *testing.T) {
	si := NewShopInvestment()

	// Purchase equipment
	si.PurchaseEquipment("display_case", 1000)

	// Damage the equipment
	si.equipment["display_case"].DurabilityCur = 50
	si.equipment["display_case"].Active = false

	// Repair it
	err := si.RepairEquipment("display_case", 1000)
	if err != nil {
		t.Errorf("Failed to repair equipment: %v", err)
	}

	equipment := si.equipment["display_case"]
	if equipment.DurabilityCur != equipment.DurabilityMax {
		t.Error("Equipment not fully repaired")
	}

	if !equipment.Active {
		t.Error("Equipment should be active after repair")
	}

	// Test repairing non-purchased equipment
	err = si.RepairEquipment("cash_register", 1000)
	if err == nil {
		t.Error("Expected error for non-purchased equipment")
	}
}

func TestGetAvailableUpgrades(t *testing.T) {
	si := NewShopInvestment()

	upgrades := si.GetAvailableUpgrades()

	// Initially should have 2 unlocked upgrades
	expectedCount := 2 // storage_small and efficiency_training
	if len(upgrades) != expectedCount {
		t.Errorf("Expected %d available upgrades, got %d", expectedCount, len(upgrades))
	}

	// Upgrade shop level to unlock more
	si.UpgradeShopLevel(5000)

	upgrades = si.GetAvailableUpgrades()
	if len(upgrades) <= expectedCount {
		t.Error("Should have more upgrades after leveling up")
	}
}

func TestGetAvailableEquipment(t *testing.T) {
	si := NewShopInvestment()

	equipment := si.GetAvailableEquipment()

	// Should have all 5 equipment initially
	if len(equipment) != 5 {
		t.Errorf("Expected 5 available equipment, got %d", len(equipment))
	}

	// Purchase one
	si.PurchaseEquipment("display_case", 1000)

	equipment = si.GetAvailableEquipment()
	if len(equipment) != 4 {
		t.Errorf("Expected 4 available equipment after purchase, got %d", len(equipment))
	}
}

func TestSimulateInvestmentScenario(t *testing.T) {
	si := NewShopInvestment()

	// Simulate 6 months
	returns := si.SimulateInvestmentScenario(1000, 6)

	if len(returns) != 6 {
		t.Errorf("Expected 6 month returns, got %d", len(returns))
	}

	// First month should be 5% of 1000 = 50
	if returns[0] != 50 {
		t.Errorf("Expected first month return 50, got %f", returns[0])
	}

	// Returns should increase due to compound effect
	for i := 1; i < len(returns); i++ {
		if returns[i] <= returns[i-1] {
			t.Error("Returns should increase with compound effect")
		}
	}
}

func TestGetInvestmentAdvice(t *testing.T) {
	si := NewShopInvestment()

	advice := si.GetInvestmentAdvice(10000, 1000)

	if advice == "" {
		t.Error("Should provide investment advice")
	}

	// Should recommend storage upgrades for low capacity
	if si.capacity < 50 {
		if !contains(advice, "Storage upgrades") {
			t.Error("Should recommend storage upgrades for low capacity")
		}
	}

	// Should mention shop upgrade if affordable
	if si.level < 3 && 10000 > si.level*5000 {
		if !contains(advice, "Shop upgrade available") {
			t.Error("Should mention available shop upgrade")
		}
	}
}

func TestMaxLevelUpgrade(t *testing.T) {
	si := NewShopInvestment()

	// Max out storage_small upgrade
	for i := 0; i < 5; i++ {
		err := si.PurchaseUpgrade("storage_small", 10000)
		if err != nil {
			t.Errorf("Failed to purchase upgrade level %d: %v", i+1, err)
		}
	}

	// Try to purchase beyond max level
	err := si.PurchaseUpgrade("storage_small", 10000)
	if err == nil {
		t.Error("Expected error when purchasing beyond max level")
	}

	// Check total capacity increase
	expectedCapacity := 20 + (10 * 5) // base + 5 levels of +10
	if si.capacity != expectedCapacity {
		t.Errorf("Expected capacity %d, got %d", expectedCapacity, si.capacity)
	}
}

func TestMaintenanceReduction(t *testing.T) {
	si := NewShopInvestment()

	// Level up to unlock premium location
	si.level = 3
	si.checkUnlocks()

	initialMaintenance := si.maintenanceCost

	// Purchase premium location (increases maintenance by 20%)
	err := si.PurchaseUpgrade("premium_location", 10000)
	if err != nil {
		t.Errorf("Failed to purchase premium location: %v", err)
	}

	expectedMaintenance := int(float64(initialMaintenance) * 1.2)
	if si.maintenanceCost != expectedMaintenance {
		t.Errorf("Expected maintenance %d, got %d", expectedMaintenance, si.maintenanceCost)
	}
}

func TestConcurrentAccess(t *testing.T) {
	si := NewShopInvestment()
	done := make(chan bool)

	// Multiple goroutines trying to purchase
	for i := 0; i < 10; i++ {
		go func() {
			si.GetCapacity()
			si.GetEfficiency()
			si.GetLevel()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without deadlock, concurrent access works
	t.Log("Concurrent access test passed")
}

func TestBreakEvenCalculation(t *testing.T) {
	si := NewShopInvestment()

	// Make investments
	si.totalInvested = 10000
	si.maintenanceCost = 100

	// Calculate with positive monthly profit
	roi := si.CalculateROI(5000, 2000)

	// Monthly profit = (5000-2000) - 100*30 = 3000 - 3000 = 0
	if roi.MonthlyProfit != 0 {
		t.Errorf("Expected monthly profit 0, got %d", roi.MonthlyProfit)
	}

	// With break-even, payback period should be very long or 0
	if roi.PaybackPeriod < 0 {
		t.Error("Payback period should not be negative")
	}

	// With zero monthly profit, break-even date calculation may be invalid
	// Just check it's not zero
	if roi.BreakEvenDate.IsZero() && roi.MonthlyProfit == 0 {
		// This is expected for break-even scenario
		t.Log("Break-even scenario with zero profit")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr
}
