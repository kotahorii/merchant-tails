package market

import (
	"testing"

	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

func BenchmarkPricingEngine_CalculatePrice(b *testing.B) {
	pe := NewPricingEngine()
	testItem, _ := item.NewItem("test_001", "Test Item", item.CategoryFruit, 100)
	state := &MarketState{
		CurrentDemand: DemandNormal,
		CurrentSupply: SupplyNormal,
		CurrentSeason: item.SeasonSpring,
		CurrentDay:    1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pe.CalculatePrice(testItem, state)
	}
}

func BenchmarkMarket_UpdatePrices(b *testing.B) {
	market := NewMarket()

	// Register some items
	for i := 0; i < 10; i++ {
		testItem, _ := item.NewItem(
			"item_"+string(rune(i)),
			"Test Item",
			item.CategoryFruit,
			100,
		)
		market.RegisterItem(testItem)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		market.UpdatePrices()
	}
}
