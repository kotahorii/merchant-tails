class_name GameBridge
extends Node

# Singleton for bridging Godot and Go
signal game_state_updated(state: Dictionary)
signal market_data_updated(data: Dictionary)
signal inventory_updated(inventory: Dictionary)
signal trade_completed(result: Dictionary)
signal event_received(event_name: String, data: Dictionary)

var game_manager: Object = null
var is_connected: bool = false

# Game state cache
var current_state: Dictionary = {}
var market_data: Dictionary = {}
var inventory_data: Dictionary = {}

func _ready() -> void:
	# Try to connect to Go backend
	_connect_to_backend()
	
	# Start update timer
	var timer = Timer.new()
	timer.wait_time = 0.1  # Update every 100ms
	timer.timeout.connect(_update_from_backend)
	add_child(timer)
	timer.start()

func _connect_to_backend() -> void:
	# Check if GDExtension is loaded
	if ClassDB.class_exists("MerchantGame"):
		game_manager = ClassDB.instantiate("MerchantGame")
		if game_manager:
			is_connected = true
			print("Connected to Go backend via GDExtension")
			_initialize_game_manager()
		else:
			push_error("Failed to instantiate MerchantGame class")
	else:
		push_warning("MerchantGame class not found - running in mock mode")
		_setup_mock_data()

func _initialize_game_manager() -> void:
	if not game_manager:
		return
	
	# Connect signals if available
	if game_manager.has_signal("state_changed"):
		game_manager.state_changed.connect(_on_backend_state_changed)
	
	if game_manager.has_signal("event_published"):
		game_manager.event_published.connect(_on_backend_event)

func _update_from_backend() -> void:
	if not is_connected or not game_manager:
		return
	
	# Get current game state
	if game_manager.has_method("get_game_state"):
		var state_json = game_manager.get_game_state()
		if state_json != "":
			var json = JSON.new()
			var result = json.parse(state_json)
			if result == OK:
				current_state = json.data
				game_state_updated.emit(current_state)
	
	# Get market data
	if game_manager.has_method("get_market_data"):
		var market_json = game_manager.get_market_data()
		if market_json != "":
			var json = JSON.new()
			var result = json.parse(market_json)
			if result == OK:
				market_data = json.data
				market_data_updated.emit(market_data)
	
	# Get inventory data
	if game_manager.has_method("get_inventory_data"):
		var inv_json = game_manager.get_inventory_data()
		if inv_json != "":
			var json = JSON.new()
			var result = json.parse(inv_json)
			if result == OK:
				inventory_data = json.data
				inventory_updated.emit(inventory_data)
	
	# Poll for events
	if game_manager.has_method("get_queued_events"):
		var events_json = game_manager.get_queued_events()
		if events_json != "" and events_json != "[]":
			var json = JSON.new()
			var result = json.parse(events_json)
			if result == OK:
				var events = json.data
				for event in events:
					_process_backend_event(event)

func _process_backend_event(event: Dictionary) -> void:
	var event_name = event.get("Name", "")
	var event_data_str = event.get("Data", "{}")
	
	var json = JSON.new()
	var result = json.parse(event_data_str)
	var event_data = {}
	if result == OK:
		event_data = json.data
	
	# Emit the event to listeners
	event_received.emit(event_name, event_data)

func _on_backend_state_changed(state_json: String) -> void:
	var json = JSON.new()
	var result = json.parse(state_json)
	if result == OK:
		current_state = json.data
		game_state_updated.emit(current_state)

func _on_backend_event(event_name: String, event_data: String) -> void:
	var json = JSON.new()
	var result = json.parse(event_data)
	if result == OK:
		event_received.emit(event_name, json.data)

# Public API for Godot scripts

func start_new_game(player_name: String) -> bool:
	if not is_connected or not game_manager:
		# Mock mode
		current_state = _get_mock_initial_state()
		game_state_updated.emit(current_state)
		return true
	
	if game_manager.has_method("start_new_game"):
		var success = game_manager.start_new_game(player_name)
		return success
	
	return false

func pause_game() -> void:
	if is_connected and game_manager and game_manager.has_method("pause_game"):
		game_manager.pause_game()
	else:
		current_state["isPaused"] = true
		game_state_updated.emit(current_state)

func resume_game() -> void:
	if is_connected and game_manager and game_manager.has_method("resume_game"):
		game_manager.resume_game()
	else:
		current_state["isPaused"] = false
		game_state_updated.emit(current_state)

func buy_item(item_id: String, quantity: int, price: float) -> Dictionary:
	if is_connected and game_manager and game_manager.has_method("buy_item"):
		var result_json = game_manager.buy_item(item_id, quantity, price)
		var json = JSON.new()
		var parse_result = json.parse(result_json)
		if parse_result == OK:
			return json.data
	
	# Mock response
	return {
		"success": true,
		"message": "Item purchased",
		"gold_remaining": current_state.get("gold", 1000) - int(price * quantity)
	}

func sell_item(item_id: String, quantity: int, price: float) -> Dictionary:
	if is_connected and game_manager and game_manager.has_method("sell_item"):
		var result_json = game_manager.sell_item(item_id, quantity, price)
		var json = JSON.new()
		var parse_result = json.parse(result_json)
		if parse_result == OK:
			return json.data
	
	# Mock response
	return {
		"success": true,
		"message": "Item sold",
		"gold_gained": int(price * quantity)
	}

func move_item_to_shop(item_id: String, quantity: int) -> bool:
	if is_connected and game_manager and game_manager.has_method("move_item_to_shop"):
		return game_manager.move_item_to_shop(item_id, quantity)
	
	return true  # Mock success

func move_item_to_warehouse(item_id: String, quantity: int) -> bool:
	if is_connected and game_manager and game_manager.has_method("move_item_to_warehouse"):
		return game_manager.move_item_to_warehouse(item_id, quantity)
	
	return true  # Mock success

func update_item_price(item_id: String, new_price: float) -> void:
	if is_connected and game_manager and game_manager.has_method("update_item_price"):
		game_manager.update_item_price(item_id, new_price)

func save_game(slot: int) -> bool:
	if is_connected and game_manager and game_manager.has_method("save_game"):
		return game_manager.save_game(slot)
	
	# Mock save
	print("Mock: Saving game to slot ", slot)
	return true

func load_game(slot: int) -> bool:
	if is_connected and game_manager and game_manager.has_method("load_game"):
		return game_manager.load_game(slot)
	
	# Mock load
	print("Mock: Loading game from slot ", slot)
	return true

func get_save_slots() -> Array:
	if is_connected and game_manager and game_manager.has_method("get_save_slots"):
		var slots_json = game_manager.get_save_slots()
		var json = JSON.new()
		var result = json.parse(slots_json)
		if result == OK:
			return json.data
	
	# Mock save slots
	return [
		{"slot": 0, "exists": false},
		{"slot": 1, "exists": true, "metadata": {"playerName": "Test Player", "gold": 5000, "rank": "Journeyman", "day": 15}},
		{"slot": 2, "exists": false}
	]

# Mock data for testing without backend

func _setup_mock_data() -> void:
	current_state = _get_mock_initial_state()
	market_data = _get_mock_market_data()
	inventory_data = _get_mock_inventory_data()

func _get_mock_initial_state() -> Dictionary:
	return {
		"isRunning": false,
		"isPaused": false,
		"gold": 1000,
		"rank": "Apprentice",
		"reputation": 0,
		"day": 1,
		"season": "Spring",
		"timeOfDay": "Morning"
	}

func _get_mock_market_data() -> Dictionary:
	return {
		"items": [
			{
				"id": "apple",
				"name": "Fresh Apple",
				"category": "fruits",
				"basePrice": 10,
				"currentPrice": 12,
				"demand": 80,
				"supply": 100,
				"volatility": 0.2
			},
			{
				"id": "sword",
				"name": "Iron Sword",
				"category": "weapons",
				"basePrice": 150,
				"currentPrice": 145,
				"demand": 30,
				"supply": 20,
				"volatility": 0.1
			},
			{
				"id": "potion",
				"name": "Health Potion",
				"category": "potions",
				"basePrice": 50,
				"currentPrice": 55,
				"demand": 120,
				"supply": 80,
				"volatility": 0.3
			}
		]
	}

func _get_mock_inventory_data() -> Dictionary:
	return {
		"shop": [
			{"id": "apple", "quantity": 10, "price": 15},
			{"id": "potion", "quantity": 5, "price": 60}
		],
		"warehouse": [
			{"id": "sword", "quantity": 2, "price": 160},
			{"id": "apple", "quantity": 20, "price": 12}
		],
		"shopCapacity": 50,
		"warehouseCapacity": 100,
		"shopUsed": 15,
		"warehouseUsed": 22
	}

# Getters for current data

func get_current_gold() -> int:
	return current_state.get("gold", 0)

func get_current_rank() -> String:
	return current_state.get("rank", "Apprentice")

func get_current_day() -> int:
	return current_state.get("day", 1)

func get_current_season() -> String:
	return current_state.get("season", "Spring")

func get_market_items() -> Array:
	return market_data.get("items", [])

func get_shop_inventory() -> Array:
	return inventory_data.get("shop", [])

func get_warehouse_inventory() -> Array:
	return inventory_data.get("warehouse", [])