package main

import "C"
import (
	"fmt"
)

//export godot_gdextension_init
func godot_gdextension_init() {
	fmt.Println("Merchant Tails GDExtension initialized")
}

//export godot_gdextension_terminate
func godot_gdextension_terminate() {
	fmt.Println("Merchant Tails GDExtension terminated")
}

func main() {
	// This is required for building a C shared library
	// The actual entry points are the exported functions above
}
