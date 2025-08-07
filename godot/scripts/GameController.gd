class_name GameController
extends Node

# Main game controller that coordinates between UI and backend

@onready var game_bridge: GameBridge = $GameBridge
@onready var ui_manager: UIManager = $UIManager
@onready var game_hud: GameHUD = $CanvasLayer/GameHUD

var is_game_active: bool = false
var player_name: String = "Merchant"

func _ready() -> void:
	# Create game bridge if not exists
	if not game_bridge:
		game_bridge = GameBridge.new()
		game_bridge.name = "GameBridge"
		add_child(game_bridge)

	# Connect bridge signals
	game_bridge.game_state_updated.connect(_on_game_state_updated)
	game_bridge.market_data_updated.connect(_on_market_data_updated)
	game_bridge.inventory_updated.connect(_on_inventory_updated)
	game_bridge.event_received.connect(_on_backend_event)

	# Connect HUD signals if in game
	if game_hud:
		game_hud.menu_requested.connect(_on_menu_requested)
		game_hud.shop_requested.connect(_on_shop_requested)
		game_hud.inventory_requested.connect(_on_inventory_requested)
		game_hud.market_requested.connect(_on_market_requested)
		game_hud.stats_requested.connect(_on_stats_requested)

	# Start with main menu
	if ui_manager:
		ui_manager.open_panel("MainMenu")

func start_new_game(name: String) -> void:
	player_name = name

	# Start game in backend
	var success = game_bridge.start_new_game(player_name)

	if success:
		is_game_active = true

		# Hide main menu
		if ui_manager:
			ui_manager.close_all_panels()

		# Show game HUD
		if game_hud:
			game_hud.visible = true
			game_hud.update_gold(1000)  # Starting gold
			game_hud.update_rank("Apprentice")
			game_hud.update_time(1, "Spring", "Morning")

		# Show shop view by default
		if ui_manager:
			ui_manager.open_panel("ShopView")

		# Show welcome notification
		if game_hud:
			game_hud.show_notification("Welcome to Merchant Tails, " + player_name + "!", "success", 5.0)

func pause_game() -> void:
	if not is_game_active:
		return

	game_bridge.pause_game()

	# Show pause menu
	if ui_manager:
		ui_manager.open_panel("PauseMenu")

func resume_game() -> void:
	if not is_game_active:
		return

	game_bridge.resume_game()

	# Close pause menu
	if ui_manager and ui_manager.is_panel_open("PauseMenu"):
		ui_manager.close_current_panel()

func save_game(slot: int) -> void:
	var success = game_bridge.save_game(slot)

	if success and game_hud:
		game_hud.show_notification("Game saved to slot " + str(slot), "success")
	elif game_hud:
		game_hud.show_notification("Failed to save game", "error")

func load_game(slot: int) -> void:
	var success = game_bridge.load_game(slot)

	if success:
		is_game_active = true

		# Close menu panels
		if ui_manager:
			ui_manager.close_all_panels()

		# Show game HUD
		if game_hud:
			game_hud.visible = true
			game_hud.show_notification("Game loaded from slot " + str(slot), "success")

		# Show shop view
		if ui_manager:
			ui_manager.open_panel("ShopView")

func quit_to_menu() -> void:
	is_game_active = false

	# Hide game HUD
	if game_hud:
		game_hud.visible = false

	# Close all panels and show main menu
	if ui_manager:
		ui_manager.close_all_panels()
		ui_manager.open_panel("MainMenu")

# Signal handlers

func _on_game_state_updated(state: Dictionary) -> void:
	if not is_game_active or not game_hud:
		return

	# Update HUD with new state
	game_hud.update_gold(state.get("gold", 0))
	game_hud.update_rank(state.get("rank", "Apprentice"))
	game_hud.update_time(
		state.get("day", 1),
		state.get("season", "Spring"),
		state.get("timeOfDay", "Morning")
	)

func _on_market_data_updated(data: Dictionary) -> void:
	# Update market panel if open
	if ui_manager and ui_manager.is_panel_open("MarketView"):
		var panel = ui_manager.get_current_panel()
		if panel and panel.has_method("update_market_data"):
			panel.update_market_data(data)

func _on_inventory_updated(inventory: Dictionary) -> void:
	# Update shop and inventory panels if open
	if ui_manager:
		if ui_manager.is_panel_open("ShopView"):
			var panel = ui_manager.get_current_panel()
			if panel and panel.has_method("update_inventory"):
				panel.update_inventory(inventory.get("shop", []))

		if ui_manager.is_panel_open("Inventory"):
			var panel = ui_manager.get_current_panel()
			if panel and panel.has_method("update_inventories"):
				panel.update_inventories(
					inventory.get("shop", []),
					inventory.get("warehouse", [])
				)

func _on_backend_event(event_name: String, data: Dictionary) -> void:
	match event_name:
		"TradeCompleted":
			_handle_trade_completed(data)
		"RankUp":
			_handle_rank_up(data)
		"MarketEvent":
			_handle_market_event(data)
		"GameVictory":
			_handle_victory(data)
		"GameDefeat":
			_handle_defeat(data)

func _handle_trade_completed(data: Dictionary) -> void:
	if game_hud:
		var profit = data.get("profit", 0)
		if profit > 0:
			game_hud.show_notification("Trade profit: " + str(profit) + " G", "success")
		else:
			game_hud.show_notification("Trade loss: " + str(abs(profit)) + " G", "warning")

func _handle_rank_up(data: Dictionary) -> void:
	var new_rank = data.get("rank", "")
	if game_hud and new_rank != "":
		game_hud.show_notification("Congratulations! You've been promoted to " + new_rank + "!", "success", 5.0)
		game_hud.update_rank(new_rank)

func _handle_market_event(data: Dictionary) -> void:
	var event_type = data.get("type", "")
	var description = data.get("description", "A market event has occurred")

	if game_hud:
		game_hud.show_notification(description, "info", 4.0)

func _handle_victory(data: Dictionary) -> void:
	var victory_type = data.get("type", "default")

	# Show victory screen
	if ui_manager:
		ui_manager.open_panel("VictoryScreen", {"type": victory_type})

	is_game_active = false

func _handle_defeat(data: Dictionary) -> void:
	var defeat_type = data.get("type", "default")

	# Show defeat screen
	if ui_manager:
		ui_manager.open_panel("DefeatScreen", {"type": defeat_type})

	is_game_active = false

# HUD request handlers

func _on_menu_requested() -> void:
	pause_game()

func _on_shop_requested() -> void:
	if ui_manager:
		ui_manager.open_panel("ShopView")

func _on_inventory_requested() -> void:
	if ui_manager:
		ui_manager.open_panel("Inventory")

func _on_market_requested() -> void:
	if ui_manager:
		ui_manager.open_panel("MarketView")

func _on_stats_requested() -> void:
	if ui_manager:
		ui_manager.open_panel("StatsView")

# Input handling

func _input(event: InputEvent) -> void:
	if not is_game_active:
		return

	# Pause with ESC
	if event.is_action_pressed("ui_cancel"):
		if ui_manager and ui_manager.get_current_panel():
			ui_manager.close_current_panel()
		else:
			pause_game()

	# Quick save
	if event.is_action_pressed("quick_save"):
		save_game(0)  # Quick save to slot 0

	# Quick load
	if event.is_action_pressed("quick_load"):
		load_game(0)  # Quick load from slot 0
