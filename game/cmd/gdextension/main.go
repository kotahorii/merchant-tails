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
	"fmt"
	"unsafe"

	"github.com/yourusername/merchant-tails/game/internal/domain/event"
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

// initializeCore sets up core systems
func initializeCore() {
	fmt.Println("Initializing core systems...")

	// Initialize the global event bus
	eventBus := event.GetGlobalEventBus()
	if eventBus != nil {
		fmt.Println("Event bus initialized")
	}

	// Register base classes with Godot
	registry := gdextension.GetClassRegistry()

	// Register MerchantGame class
	_, err := registry.RegisterClass("MerchantGame", "Node",
		func(userData unsafe.Pointer) unsafe.Pointer {
			// Create instance
			return userData // Placeholder
		},
		func(userData unsafe.Pointer, instance unsafe.Pointer) {
			// Free instance
		},
	)
	if err != nil {
		fmt.Printf("Failed to register MerchantGame class: %v\n", err)
	}
}

// initializeServers sets up server systems
func initializeServers() {
	fmt.Println("Initializing server systems...")
	// Initialize game servers (physics, rendering, etc.)
}

// initializeScene sets up scene systems
func initializeScene() {
	fmt.Println("Initializing scene systems...")

	registry := gdextension.GetClassRegistry()

	// Register MarketManager class
	_, err := registry.RegisterClass("MarketManager", "Resource",
		func(userData unsafe.Pointer) unsafe.Pointer {
			return userData
		},
		func(userData unsafe.Pointer, instance unsafe.Pointer) {
		},
	)
	if err != nil {
		fmt.Printf("Failed to register MarketManager class: %v\n", err)
	}

	// Register methods for MarketManager
	_, err = registry.RegisterMethod("MarketManager", "get_current_price",
		func(methodData unsafe.Pointer, instance unsafe.Pointer, args []unsafe.Pointer, ret unsafe.Pointer) error {
			// Method implementation
			return nil
		},
	)
	if err != nil {
		fmt.Printf("Failed to register method: %v\n", err)
	}
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

func main() {
	// This is required for building a C shared library
	// The actual entry points are the exported functions above
}
