extends Node

# ゲーム状態
var player_gold: float = 1000.0
var current_day: int = 1
var player_name: String = "Merchant"
var inventory = {}
var market_prices = {}

# GDExtension library
var gdextension = null

# UI References
@onready var main_container = $MainContainer
@onready var player_info_label = $MainContainer/TopBar/PlayerInfo
@onready var gold_label = $MainContainer/TopBar/GoldLabel
@onready var day_label = $MainContainer/TopBar/DayLabel
@onready var market_list = $MainContainer/MarketPanel/ItemList
@onready var inventory_list = $MainContainer/InventoryPanel/ItemList
@onready var notification_label = $MainContainer/NotificationPanel/Label

func _ready():
	print("Game Main ready")
	_initialize_gdextension()
	_initialize_game()

func _initialize_gdextension():
	# Load the GDExtension library
	if FileAccess.file_exists("res://lib/libmerchant_game.dylib"):
		gdextension = load("res://merchant_game.gdextension")
		if gdextension:
			print("GDExtension loaded successfully")
			# Initialize the extension
			var init_result = OS.call_deferred("godot_gdextension_init")
			if init_result:
				print("GDExtension initialized")
		else:
			print("Failed to load GDExtension")
	else:
		print("GDExtension library not found, running in test mode")

func _initialize_game():
	if gdextension:
		_load_game_data()
	else:
		_setup_test_data()

func _load_game_data():
	# Get data from Go GDExtension
	player_gold = OS.call_deferred("get_player_gold")
	current_day = OS.call_deferred("get_current_day")
	
	# Get market data
	var market_json = OS.call_deferred("get_market_items_json")
	if market_json:
		var json = JSON.new()
		var parse_result = json.parse(market_json)
		if parse_result == OK:
			market_prices = json.data
		OS.call_deferred("free_string", market_json)
	
	# Get inventory data
	var inventory_json = OS.call_deferred("get_inventory_json")
	if inventory_json:
		var json = JSON.new()
		var parse_result = json.parse(inventory_json)
		if parse_result == OK:
			inventory = json.data
		OS.call_deferred("free_string", inventory_json)
	
	_update_ui()

func _setup_test_data():
	# Test mode data
	market_prices = {
		"apple": 10,
		"bread": 15,
		"sword": 100,
		"potion": 50,
		"armor": 200,
		"herb": 5
	}
	inventory = {
		"apple": 5,
		"bread": 3
	}
	_update_ui()

func start_new_game(name: String):
	player_name = name
	if gdextension:
		var success = OS.call_deferred("start_new_game", name)
		if success:
			_load_game_data()
			_show_notification("New game started!")
		else:
			_show_notification("Failed to start new game", true)
	else:
		# Test mode
		player_gold = 1000.0
		current_day = 1
		inventory.clear()
		_update_ui()
		_show_notification("New game started (test mode)")

func save_game():
	if gdextension:
		var success = OS.call_deferred("save_game")
		if success:
			_show_notification("Game saved!")
		else:
			_show_notification("Failed to save game", true)
	else:
		_show_notification("Save not available in test mode", true)

func load_game():
	if gdextension:
		var success = OS.call_deferred("load_game")
		if success:
			_load_game_data()
			_show_notification("Game loaded!")
		else:
			_show_notification("No saved game found", true)
	else:
		_show_notification("Load not available in test mode", true)

func advance_day():
	if gdextension:
		OS.call_deferred("advance_day")
		current_day = OS.call_deferred("get_current_day")
		_load_game_data()  # Reload to get updated prices
	else:
		current_day += 1
		# Simulate price changes in test mode
		for item_id in market_prices:
			var factor = 0.8 + randf() * 0.4  # -20% to +20%
			market_prices[item_id] = int(market_prices[item_id] * factor)
	
	_update_ui()
	_show_notification("Advanced to day " + str(current_day))

func buy_item(item_id: String, quantity: int):
	if gdextension:
		var success = OS.call_deferred("buy_item", item_id, quantity)
		if success:
			player_gold = OS.call_deferred("get_player_gold")
			_load_game_data()
			_show_notification("Bought %d %s" % [quantity, item_id])
		else:
			_show_notification("Purchase failed", true)
	else:
		# Test mode
		var price = market_prices.get(item_id, 10)
		var total_cost = price * quantity
		if player_gold >= total_cost:
			player_gold -= total_cost
			if not inventory.has(item_id):
				inventory[item_id] = 0
			inventory[item_id] += quantity
			_update_ui()
			_show_notification("Bought %d %s for %d gold" % [quantity, item_id, total_cost])
		else:
			_show_notification("Not enough gold!", true)

func sell_item(item_id: String, quantity: int):
	if gdextension:
		var success = OS.call_deferred("sell_item", item_id, quantity)
		if success:
			player_gold = OS.call_deferred("get_player_gold")
			_load_game_data()
			_show_notification("Sold %d %s" % [quantity, item_id])
		else:
			_show_notification("Sale failed", true)
	else:
		# Test mode
		if inventory.has(item_id) and inventory[item_id] >= quantity:
			inventory[item_id] -= quantity
			if inventory[item_id] == 0:
				inventory.erase(item_id)
			var price = market_prices.get(item_id, 10) * 0.8
			var total_revenue = price * quantity
			player_gold += total_revenue
			_update_ui()
			_show_notification("Sold %d %s for %d gold" % [quantity, item_id, int(total_revenue)])
		else:
			_show_notification("Not enough items!", true)

func _update_ui():
	if player_info_label:
		player_info_label.text = player_name
	if gold_label:
		gold_label.text = "Gold: " + str(int(player_gold))
	if day_label:
		day_label.text = "Day: " + str(current_day)
	
	_update_market_display()
	_update_inventory_display()

func _update_market_display():
	if not market_list:
		return
	
	market_list.clear()
	for item_id in market_prices:
		var price = market_prices[item_id]
		if price is float:
			price = int(price)
		market_list.add_item("%s - %d gold" % [item_id.capitalize(), price])

func _update_inventory_display():
	if not inventory_list:
		return
		
	inventory_list.clear()
	for item_id in inventory:
		var quantity = inventory[item_id]
		inventory_list.add_item("%s x%d" % [item_id.capitalize(), quantity])

var notification_timer: Timer

func _show_notification(message: String, is_error: bool = false):
	if notification_label:
		notification_label.text = message
		if is_error:
			notification_label.modulate = Color.RED
		else:
			notification_label.modulate = Color.WHITE
		
		notification_label.get_parent().visible = true
		
		if not notification_timer:
			notification_timer = Timer.new()
			add_child(notification_timer)
			notification_timer.timeout.connect(func(): 
				if notification_label:
					notification_label.get_parent().visible = false
			)
		
		notification_timer.wait_time = 3.0
		notification_timer.one_shot = true
		notification_timer.start()
	else:
		print(message)

# Button handlers
func _on_new_game_button_pressed():
	start_new_game("Player")

func _on_save_button_pressed():
	save_game()

func _on_load_button_pressed():
	load_game()

func _on_next_day_button_pressed():
	advance_day()

func _on_buy_button_pressed():
	# Buy selected item from market
	if market_list and market_list.is_anything_selected():
		var selected = market_list.get_selected_items()[0]
		var items = market_prices.keys()
		if selected < items.size():
			var item_id = items[selected]
			buy_item(item_id, 1)

func _on_sell_button_pressed():
	# Sell selected item from inventory
	if inventory_list and inventory_list.is_anything_selected():
		var selected = inventory_list.get_selected_items()[0]
		var items = inventory.keys()
		if selected < items.size():
			var item_id = items[selected]
			sell_item(item_id, 1)