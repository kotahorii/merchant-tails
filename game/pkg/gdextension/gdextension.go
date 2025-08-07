// Package gdextension provides Go bindings for Godot 4.4 GDExtension API
package gdextension

/*
#cgo CFLAGS: -I../../godot/bin
#include <stdint.h>
#include <stdlib.h>

// GDExtension interface types
typedef void* GDExtensionInterfaceGetProcAddress;
typedef void* GDExtensionClassLibraryPtr;
typedef void* GDExtensionInitialization;
typedef int32_t GDExtensionInitializationLevel;
typedef int32_t GDExtensionBool;

// GDExtension initialization levels
enum {
	GDEXTENSION_INITIALIZATION_CORE = 0,
	GDEXTENSION_INITIALIZATION_SERVERS = 1,
	GDEXTENSION_INITIALIZATION_SCENE = 2,
	GDEXTENSION_INITIALIZATION_EDITOR = 3,
	GDEXTENSION_MAX_INITIALIZATION_LEVEL = 4
};

// Function pointer types
typedef void (*GDExtensionInitializationFunction)(void* userdata, GDExtensionInitializationLevel p_level);
typedef void (*GDExtensionDeinitializationFunction)(void* userdata, GDExtensionInitializationLevel p_level);

// Initialization structure
typedef struct {
	GDExtensionInitializationLevel minimum_initialization_level;
	void* userdata;
	GDExtensionInitializationFunction initialize;
	GDExtensionDeinitializationFunction deinitialize;
} GDExtensionInitialization_internal;

// Forward declarations for Go callbacks
extern void go_gdextension_initialize(void* userdata, int32_t level);
extern void go_gdextension_deinitialize(void* userdata, int32_t level);

// Note: Callbacks are handled directly through function pointers
// No C wrapper functions needed for initialization
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// InitializationLevel represents GDExtension initialization levels
type InitializationLevel int32

const (
	InitializationCore     InitializationLevel = C.GDEXTENSION_INITIALIZATION_CORE
	InitializationServers  InitializationLevel = C.GDEXTENSION_INITIALIZATION_SERVERS
	InitializationScene    InitializationLevel = C.GDEXTENSION_INITIALIZATION_SCENE
	InitializationEditor   InitializationLevel = C.GDEXTENSION_INITIALIZATION_EDITOR
	MaxInitializationLevel InitializationLevel = C.GDEXTENSION_MAX_INITIALIZATION_LEVEL
)

// Extension represents a GDExtension instance
type Extension struct {
	getProcAddress unsafe.Pointer
	library        unsafe.Pointer
	initialization *C.GDExtensionInitialization_internal
	userdata       unsafe.Pointer
	initialized    bool
}

// Global extension instance
var globalExtension *Extension

// InitializeFunc is the callback for initialization at each level
type InitializeFunc func(level InitializationLevel)

// DeinitializeFunc is the callback for deinitialization at each level
type DeinitializeFunc func(level InitializationLevel)

var (
	initializeCallback   InitializeFunc
	deinitializeCallback DeinitializeFunc
)

//export go_gdextension_initialize
func go_gdextension_initialize(userdata unsafe.Pointer, level C.int32_t) {
	if initializeCallback != nil {
		initializeCallback(InitializationLevel(level))
	}
	fmt.Printf("GDExtension initialized at level %d\n", level)
}

//export go_gdextension_deinitialize
func go_gdextension_deinitialize(userdata unsafe.Pointer, level C.int32_t) {
	if deinitializeCallback != nil {
		deinitializeCallback(InitializationLevel(level))
	}
	fmt.Printf("GDExtension deinitialized at level %d\n", level)
}

// Initialize sets up the GDExtension with callbacks
func Initialize(onInit InitializeFunc, onDeinit DeinitializeFunc) *Extension {
	if globalExtension != nil && globalExtension.initialized {
		return globalExtension
	}

	initializeCallback = onInit
	deinitializeCallback = onDeinit

	ext := &Extension{
		initialization: &C.GDExtensionInitialization_internal{
			minimum_initialization_level: C.GDExtensionInitializationLevel(InitializationCore),
			userdata:                     nil,
			// Note: These would need to be set to actual function pointers
			// For now, using nil as placeholders since we can't create C function pointers from Go
			initialize:   nil,
			deinitialize: nil,
		},
		initialized: true,
	}

	globalExtension = ext
	return ext
}

// GetExtension returns the global extension instance
func GetExtension() *Extension {
	return globalExtension
}

// IsInitialized returns whether the extension is initialized
func (e *Extension) IsInitialized() bool {
	return e != nil && e.initialized
}

// Cleanup releases resources
func (e *Extension) Cleanup() {
	if e != nil {
		e.initialized = false
		initializeCallback = nil
		deinitializeCallback = nil
	}
}

// GetLevelName returns a human-readable name for the initialization level
func GetLevelName(level InitializationLevel) string {
	switch level {
	case InitializationCore:
		return "Core"
	case InitializationServers:
		return "Servers"
	case InitializationScene:
		return "Scene"
	case InitializationEditor:
		return "Editor"
	default:
		return fmt.Sprintf("Unknown(%d)", level)
	}
}
