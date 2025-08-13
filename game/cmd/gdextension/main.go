//go:build cgo
// +build cgo

package main

// #include <stdlib.h>
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/yourusername/merchant-tails/game/internal/presentation/api"
)

// Global game manager instance
var gameManager *api.GameManager

// Game state cache
var (
	currentGold float64 = 1000.0
	currentDay  int32   = 1
	marketData  map[string]float64
	inventory   map[string]int
)

//export godot_gdextension_init
func godot_gdextension_init() C.int {
	fmt.Println("Initializing Merchant Tails GDExtension")
	gameManager = api.NewGameManager()

	// Initialize market data
	marketData = make(map[string]float64)
	inventory = make(map[string]int)

	// Set initial market prices
	marketData["apple"] = 10.0
	marketData["bread"] = 15.0
	marketData["sword"] = 100.0
	marketData["potion"] = 50.0
	marketData["armor"] = 200.0
	marketData["herb"] = 5.0

	return 1 // Success
}

//export godot_gdextension_terminate
func godot_gdextension_terminate() {
	fmt.Println("Terminating Merchant Tails GDExtension")
	if gameManager != nil {
		_ = gameManager.SaveGame(0) // Auto-save on exit
		gameManager.Cleanup()
	}
}

//export start_new_game
func start_new_game(playerName *C.char) C.int {
	name := C.GoString(playerName)
	err := gameManager.StartNewGame(name)
	if err != nil {
		fmt.Printf("Failed to start new game: %v\n", err)
		return 0
	}

	// Reset game state
	currentGold = 1000.0
	currentDay = 1
	inventory = make(map[string]int)

	fmt.Printf("New game started for player: %s\n", name)
	return 1
}

//export get_player_gold
func get_player_gold() C.double {
	return C.double(currentGold)
}

//export set_player_gold
func set_player_gold(gold C.double) {
	currentGold = float64(gold)
}

//export get_current_day
func get_current_day() C.int {
	return C.int(currentDay)
}

//export advance_day
func advance_day() {
	currentDay++

	// Update market prices (simple simulation)
	for item := range marketData {
		// Random price fluctuation between -20% and +20%
		factor := 0.8 + (float64(currentDay%5) * 0.1)
		marketData[item] *= factor
	}

	fmt.Printf("Advanced to day %d\n", currentDay)
}

//export save_game
func save_game() C.int {
	err := gameManager.SaveGame(0)
	if err != nil {
		fmt.Printf("Failed to save game: %v\n", err)
		return 0
	}
	fmt.Println("Game saved successfully")
	return 1
}

//export load_game
func load_game() C.int {
	err := gameManager.LoadGame(0)
	if err != nil {
		fmt.Printf("Failed to load game: %v\n", err)
		return 0
	}

	// Load game state from manager
	if stateJSON, err := gameManager.GetGameState(); err == nil {
		var state map[string]interface{}
		if err := json.Unmarshal([]byte(stateJSON), &state); err == nil {
			if gold, ok := state["gold"].(float64); ok {
				currentGold = gold
			}
			if day, ok := state["day"].(float64); ok {
				currentDay = int32(day)
			}
		}
	}

	fmt.Println("Game loaded successfully")
	return 1
}

//export buy_item
func buy_item(itemID *C.char, quantity C.int) C.int {
	item := C.GoString(itemID)
	qty := int(quantity)

	// Check if item exists in market
	price, exists := marketData[item]
	if !exists {
		fmt.Printf("Item %s not found in market\n", item)
		return 0
	}

	totalCost := price * float64(qty)

	// Check if player has enough gold
	if currentGold < totalCost {
		fmt.Printf("Not enough gold. Need %.2f, have %.2f\n", totalCost, currentGold)
		return 0
	}

	// Process purchase
	currentGold -= totalCost
	inventory[item] += qty

	fmt.Printf("Bought %d %s for %.2f gold\n", qty, item, totalCost)
	return 1
}

//export sell_item
func sell_item(itemID *C.char, quantity C.int) C.int {
	item := C.GoString(itemID)
	qty := int(quantity)

	// Check if player has the item
	currentQty, hasItem := inventory[item]
	if !hasItem || currentQty < qty {
		fmt.Printf("Not enough %s in inventory\n", item)
		return 0
	}

	// Get sell price (80% of market price)
	marketPrice, exists := marketData[item]
	if !exists {
		marketPrice = 10.0 // Default price
	}
	sellPrice := marketPrice * 0.8
	totalRevenue := sellPrice * float64(qty)

	// Process sale
	inventory[item] -= qty
	if inventory[item] == 0 {
		delete(inventory, item)
	}
	currentGold += totalRevenue

	fmt.Printf("Sold %d %s for %.2f gold\n", qty, item, totalRevenue)
	return 1
}

//export get_market_price
func get_market_price(itemID *C.char) C.double {
	item := C.GoString(itemID)
	if price, exists := marketData[item]; exists {
		return C.double(price)
	}
	return C.double(0.0)
}

//export get_inventory_quantity
func get_inventory_quantity(itemID *C.char) C.int {
	item := C.GoString(itemID)
	if qty, exists := inventory[item]; exists {
		return C.int(qty)
	}
	return C.int(0)
}

//export get_market_items_json
func get_market_items_json() *C.char {
	jsonData, _ := json.Marshal(marketData)
	return C.CString(string(jsonData))
}

//export get_inventory_json
func get_inventory_json() *C.char {
	jsonData, _ := json.Marshal(inventory)
	return C.CString(string(jsonData))
}

//export free_string
func free_string(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func main() {
	// Required for building as shared library
}
