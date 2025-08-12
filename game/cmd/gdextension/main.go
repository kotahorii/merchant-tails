package main

/*
#cgo CFLAGS: -I../../godot/bin
#include <stdint.h>

// Entry point type definition
typedef int32_t (*GDExtensionEntryPoint)(void* get_proc_address, void* library, void* initialization);

// Export the library entry point
extern int32_t merchant_game_library_init(void* get_proc_address, void* library, void* initialization);
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/yourusername/merchant-tails/game/internal/domain/event"
	"github.com/yourusername/merchant-tails/game/internal/presentation/api"
	"github.com/yourusername/merchant-tails/game/pkg/gdextension"
)

//export merchant_game_library_init
func merchant_game_library_init(getProcAddress unsafe.Pointer, library unsafe.Pointer, initialization unsafe.Pointer) C.int32_t {
	fmt.Println("Merchant Tails GDExtension library initializing...")

	// Initialize the extension
	ext := gdextension.Initialize(
		// Initialize callback
		func(level gdextension.InitializationLevel) {
			fmt.Printf("Initializing at level: %s\n", gdextension.GetLevelName(level))

			switch level {
			case gdextension.InitializationCore:
				initializeCore()
			case gdextension.InitializationServers:
				initializeServers()
			case gdextension.InitializationScene:
				initializeScene()
			case gdextension.InitializationEditor:
				initializeEditor()
			}
		},
		// Deinitialize callback
		func(level gdextension.InitializationLevel) {
			fmt.Printf("Deinitializing at level: %s\n", gdextension.GetLevelName(level))

			switch level {
			case gdextension.InitializationEditor:
				deinitializeEditor()
			case gdextension.InitializationScene:
				deinitializeScene()
			case gdextension.InitializationServers:
				deinitializeServers()
			case gdextension.InitializationCore:
				deinitializeCore()
			}
		},
	)

	if ext.IsInitialized() {
		fmt.Println("Merchant Tails GDExtension successfully initialized")
		return 1 // Success
	}

	return 0 // Failure
}

// Global game manager instance
var gameManager *api.GameManager

// initializeCore sets up core systems
func initializeCore() {
	fmt.Println("Initializing core systems...")

	// Initialize the global event bus
	eventBus := event.GetGlobalEventBus()
	if eventBus != nil {
		fmt.Println("Event bus initialized")
	}

	// Create game manager
	gameManager = api.NewGameManager()

	// Register base classes with Godot
	registry := gdextension.GetClassRegistry()

	// Register MerchantGame class
	_, err := registry.RegisterClass("MerchantGame", "Node",
		func(userData unsafe.Pointer) unsafe.Pointer {
			// Return the game manager as the instance
			return unsafe.Pointer(gameManager)
		},
		func(userData unsafe.Pointer, instance unsafe.Pointer) {
			// Cleanup on free
			if gameManager != nil {
				gameManager.Cleanup()
			}
		},
	)
	if err != nil {
		fmt.Printf("Failed to register MerchantGame class: %v\n", err)
	}

	// Register game manager methods
	registerGameManagerMethods(registry)
}

// initializeServers sets up server systems
func initializeServers() {
	fmt.Println("Initializing server systems...")
	// Initialize game servers (physics, rendering, etc.)
}

// initializeScene sets up scene systems
func initializeScene() {
	fmt.Println("Initializing scene systems...")
	// Scene-specific initialization
}

// initializeEditor sets up editor systems
func initializeEditor() {
	fmt.Println("Initializing editor systems...")
	// Initialize editor-specific functionality
}

// deinitializeEditor cleans up editor systems
func deinitializeEditor() {
	fmt.Println("Deinitializing editor systems...")
}

// deinitializeScene cleans up scene systems
func deinitializeScene() {
	fmt.Println("Deinitializing scene systems...")
}

// deinitializeServers cleans up server systems
func deinitializeServers() {
	fmt.Println("Deinitializing server systems...")
}

// deinitializeCore cleans up core systems
func deinitializeCore() {
	fmt.Println("Deinitializing core systems...")

	// Clear the class registry
	registry := gdextension.GetClassRegistry()
	registry.Clear()

	// Reset the event bus
	event.ResetGlobalEventBus()

	// Clean up the extension
	ext := gdextension.GetExtension()
	if ext != nil {
		ext.Cleanup()
	}
}

//export godot_gdextension_init
func godot_gdextension_init() {
	// Legacy entry point for compatibility
	fmt.Println("Legacy init called - redirecting to new entry point")
}

//export godot_gdextension_terminate
func godot_gdextension_terminate() {
	// Legacy termination point for compatibility
	fmt.Println("Legacy terminate called")
}

// registerGameManagerMethods registers all GameManager methods with Godot
func registerGameManagerMethods(registry *gdextension.ClassRegistry) {
	// Start new game
	_, _ = registry.RegisterMethod("MerchantGame", "start_new_game",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			// Get player name from args
			playerName := gdextension.GetStringFromPointer(args[0])
			err := gameManager.StartNewGame(playerName)

			// Return success
			gdextension.SetBoolToPointer(ret, err == nil)
			return nil
		},
	)

	// Get game state
	_, _ = registry.RegisterMethod("MerchantGame", "get_game_state",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			stateJSON, err := gameManager.GetGameState()
			if err != nil {
				gdextension.SetStringToPointer(ret, "")
				return nil
			}

			gdextension.SetStringToPointer(ret, stateJSON)
			return nil
		},
	)

	// Get market data
	_, _ = registry.RegisterMethod("MerchantGame", "get_market_data",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			// Get market data from game manager
			marketData := gameManager.GetMarketData()
			jsonData, err := json.Marshal(marketData)
			if err != nil {
				gdextension.SetStringToPointer(ret, "")
				return nil
			}

			gdextension.SetStringToPointer(ret, string(jsonData))
			return nil
		},
	)

	// Get inventory data
	_, _ = registry.RegisterMethod("MerchantGame", "get_inventory_data",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			// Get inventory data from game manager
			inventoryData := gameManager.GetInventoryData()
			jsonData, err := json.Marshal(inventoryData)
			if err != nil {
				gdextension.SetStringToPointer(ret, "")
				return nil
			}

			gdextension.SetStringToPointer(ret, string(jsonData))
			return nil
		},
	)

	// Pause game
	_, _ = registry.RegisterMethod("MerchantGame", "pause_game",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			gameManager.PauseGame()
			return nil
		},
	)

	// Resume game
	_, _ = registry.RegisterMethod("MerchantGame", "resume_game",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			gameManager.ResumeGame()
			return nil
		},
	)

	// Buy item
	_, _ = registry.RegisterMethod("MerchantGame", "buy_item",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			itemID := gdextension.GetStringFromPointer(args[0])
			quantity := gdextension.GetIntFromPointer(args[1])
			price := gdextension.GetFloatFromPointer(args[2])

			result := gameManager.BuyItem(itemID, quantity, price)
			jsonData, err := json.Marshal(result)
			if err != nil {
				gdextension.SetStringToPointer(ret, "")
				return nil
			}

			gdextension.SetStringToPointer(ret, string(jsonData))
			return nil
		},
	)

	// Sell item
	_, _ = registry.RegisterMethod("MerchantGame", "sell_item",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			itemID := gdextension.GetStringFromPointer(args[0])
			quantity := gdextension.GetIntFromPointer(args[1])
			price := gdextension.GetFloatFromPointer(args[2])

			result := gameManager.SellItem(itemID, quantity, price)
			jsonData, err := json.Marshal(result)
			if err != nil {
				gdextension.SetStringToPointer(ret, "")
				return nil
			}

			gdextension.SetStringToPointer(ret, string(jsonData))
			return nil
		},
	)

	// Save game
	_, _ = registry.RegisterMethod("MerchantGame", "save_game",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			slot := gdextension.GetIntFromPointer(args[0])
			err := gameManager.SaveGame(slot)
			gdextension.SetBoolToPointer(ret, err == nil)
			return nil
		},
	)

	// Load game
	_, _ = registry.RegisterMethod("MerchantGame", "load_game",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				return fmt.Errorf("game manager not initialized")
			}

			slot := gdextension.GetIntFromPointer(args[0])
			err := gameManager.LoadGame(slot)
			gdextension.SetBoolToPointer(ret, err == nil)
			return nil
		},
	)

	// Get save slots
	registry.RegisterMethod("MerchantGame", "get_save_slots",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				gdextension.SetStringToPointer(ret, "[]")
				return nil
			}

			slotsJSON, err := gameManager.GetSaveSlots()
			if err != nil {
				gdextension.SetStringToPointer(ret, "[]")
				return nil
			}

			gdextension.SetStringToPointer(ret, slotsJSON)
			return nil
		},
	)

	// Get queued events
	registry.RegisterMethod("MerchantGame", "get_queued_events",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			if gameManager == nil {
				gdextension.SetStringToPointer(ret, "[]")
				return nil
			}

			eventsJSON, err := gameManager.GetQueuedEvents()
			if err != nil {
				gdextension.SetStringToPointer(ret, "[]")
				return nil
			}

			gdextension.SetStringToPointer(ret, eventsJSON)
			return nil
		},
	)

	// Register signals
	_ = registry.RegisterSignal("MerchantGame", "state_changed", []string{"state_json"})
	_ = registry.RegisterSignal("MerchantGame", "event_published", []string{"event_name", "event_data"})
}

func main() {
	// This is required for building a C shared library
	// The actual entry points are the exported functions above
}
