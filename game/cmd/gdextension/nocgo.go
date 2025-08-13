//go:build !cgo
// +build !cgo

package main

// This file provides stub implementations when CGO is disabled
// to prevent editor errors

type C struct{}

func (C) int(v int) int32          { return int32(v) }
func (C) double(v float64) float64 { return v }
func (C) CString(s string) *byte   { return nil }
func (C) GoString(s *byte) string  { return "" }
func (C) free(ptr *byte)           {}

// Stub functions for non-CGO builds
func godot_gdextension_init() int                { return 0 }
func godot_gdextension_terminate()               {}
func start_new_game(playerName *byte) int        { return 0 }
func get_player_gold() float64                   { return 0 }
func get_current_day() int32                     { return 0 }
func advance_day()                               {}
func save_game() int                             { return 0 }
func load_game() int                             { return 0 }
func buy_item(itemID *byte, quantity int32) int  { return 0 }
func sell_item(itemID *byte, quantity int32) int { return 0 }
func get_market_price(itemID *byte) float64      { return 0 }
func get_inventory_quantity(itemID *byte) int32  { return 0 }
func get_market_items_json() *byte               { return nil }
func get_inventory_json() *byte                  { return nil }
func free_string(str *byte)                      {}
