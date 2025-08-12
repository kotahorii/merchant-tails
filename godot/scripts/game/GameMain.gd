extends Control

# Game API reference
var game_api

# UI nodes
@onready var player_name_label = $MainContainer/LeftPanel/PlayerInfo/VBox/PlayerName
@onready var gold_label = $MainContainer/LeftPanel/PlayerInfo/VBox/GoldLabel
@onready var rank_label = $MainContainer/LeftPanel/PlayerInfo/VBox/RankLabel
@onready var reputation_label = $MainContainer/LeftPanel/PlayerInfo/VBox/ReputationLabel
@onready var day_label = $MainContainer/LeftPanel/PlayerInfo/VBox/DayLabel
@onready var season_label = $MainContainer/LeftPanel/PlayerInfo/VBox/SeasonLabel

@onready var weather_label = $MainContainer/LeftPanel/WeatherInfo/VBox/WeatherLabel
@onready var weather_effect_label = $MainContainer/LeftPanel/WeatherInfo/VBox/WeatherEffect

@onready var center_tabs = $MainContainer/CenterPanel
@onready var shop_item_grid = $MainContainer/CenterPanel/Shop/ShopView/VBox/ItemGrid
@onready var market_item_list = $MainContainer/CenterPanel/Market/MarketView/VBox/ItemList
@onready var market_quantity_spin = $MainContainer/CenterPanel/Market/MarketView/VBox/BuyPanel/QuantitySpinBox
@onready var shop_inventory_list = $MainContainer/CenterPanel/Inventory/InventoryView/VBox/HSplitContainer/ShopInventory/ShopList
@onready var warehouse_inventory_list = $MainContainer/CenterPanel/Inventory/InventoryView/VBox/HSplitContainer/WarehouseInventory/WarehouseList
@onready var bank_balance_label = $MainContainer/CenterPanel/Bank/BankView/VBox/BalanceLabel
@onready var bank_amount_spin = $MainContainer/CenterPanel/Bank/BankView/VBox/HBoxContainer/AmountSpinBox

@onready var notification_panel = $NotificationPanel
@onready var notification_text = $NotificationPanel/NotificationText

# Game state
var player_info = {}
var market_prices = {}
var inventory = {}
var bank_account = {}
var current_weather = {}

func _ready():
	# Initialize game API
	_initialize_game_api()
	
	# Load game state
	_load_game_state()
	
	# Update all UI elements
	_update_player_info()
	_update_market_prices()
	_update_inventory()
	_update_bank_info()
	_update_weather()
	
	# Connect to game events
	_connect_game_events()

func _initialize_game_api():
	# Load the GDExtension
	game_api = load("res://bin/merchant_game.gdextension")
	
	if not game_api:
		push_error("Failed to load game API")
		return
	
	# Initialize the game
	if game_api.initialize_game():
		print("Game initialized successfully")
	else:
		push_error("Failed to initialize game")

func _load_game_state():
	# Check if we're continuing or starting new
	var save_exists = FileAccess.file_exists("user://savegame.dat")
	
	if save_exists:
		# Load saved game
		var result = game_api.load_game(0)
		if result.success:
			print("Game loaded successfully")
		else:
			push_error("Failed to load game: " + result.message)
			_start_new_game()
	else:
		_start_new_game()

func _start_new_game():
	# Start a new game with default player name
	var result = game_api.start_new_game("Player")
	if result.success:
		print("New game started")
	else:
		push_error("Failed to start new game: " + result.message)

func _update_player_info():
	player_info = game_api.get_player_info()
	
	player_name_label.text = player_info.name
	gold_label.text = tr("UI_GOLD") + ": " + str(player_info.gold)
	rank_label.text = tr("UI_RANK") + ": " + tr(player_info.rank)
	reputation_label.text = tr("UI_REPUTATION") + ": " + str(player_info.reputation)
	day_label.text = tr("UI_DAY") + ": " + str(player_info.day)
	season_label.text = tr("UI_SEASON") + ": " + tr(player_info.season)

func _update_market_prices():
	market_prices = game_api.get_market_prices()
	
	# Clear and populate market list
	market_item_list.clear()
	for item_id in market_prices:
		var price = market_prices[item_id]
		var item_text = tr(item_id.to_upper()) + " - " + str(price) + " " + tr("UI_GOLD")
		market_item_list.add_item(item_text)

func _update_inventory():
	inventory = game_api.get_inventory()
	
	# Update shop inventory
	shop_inventory_list.clear()
	if "shop" in inventory:
		for item_id in inventory.shop:
			var quantity = inventory.shop[item_id]
			var item_text = tr(item_id.to_upper()) + " x" + str(quantity)
			shop_inventory_list.add_item(item_text)
	
	# Update warehouse inventory
	warehouse_inventory_list.clear()
	if "warehouse" in inventory:
		for item_id in inventory.warehouse:
			var quantity = inventory.warehouse[item_id]
			var item_text = tr(item_id.to_upper()) + " x" + str(quantity)
			warehouse_inventory_list.add_item(item_text)

func _update_bank_info():
	bank_account = game_api.get_bank_account()
	
	var balance_text = tr("UI_BALANCE") + ": " + str(bank_account.balance) + " " + tr("UI_GOLD")
	bank_balance_label.text = balance_text

func _update_weather():
	current_weather = game_api.get_current_weather()
	
	weather_label.text = tr("UI_WEATHER") + ": " + tr(current_weather.type)
	
	var effect_text = ""
	if current_weather.price_effect != 0:
		effect_text += tr("UI_PRICE") + ": " + ("+" if current_weather.price_effect > 0 else "") + str(current_weather.price_effect * 100) + "%"
	if current_weather.demand_effect != 0:
		if effect_text != "":
			effect_text += ", "
		effect_text += tr("UI_DEMAND") + ": " + ("+" if current_weather.demand_effect > 0 else "") + str(current_weather.demand_effect * 100) + "%"
	
	weather_effect_label.text = effect_text

func _connect_game_events():
	# Connect to game events if API supports signals
	if game_api.has_signal("price_updated"):
		game_api.price_updated.connect(_on_price_updated)
	if game_api.has_signal("transaction_complete"):
		game_api.transaction_complete.connect(_on_transaction_complete)
	if game_api.has_signal("day_changed"):
		game_api.day_changed.connect(_on_day_changed)
	if game_api.has_signal("weather_changed"):
		game_api.weather_changed.connect(_on_weather_changed)

func _on_price_updated(item_id: String, new_price: int):
	market_prices[item_id] = new_price
	_update_market_prices()

func _on_transaction_complete(transaction: Dictionary):
	_show_notification(tr("MSG_TRANSACTION_COMPLETE") + ": " + transaction.type + " " + str(transaction.quantity) + " " + tr(transaction.item_id.to_upper()))
	_update_player_info()
	_update_inventory()

func _on_day_changed(new_day: int, season: String):
	_update_player_info()
	_update_weather()
	_show_notification(tr("MSG_DAY_CHANGED") + ": " + str(new_day) + " (" + tr(season) + ")")

func _on_weather_changed(weather: String):
	_update_weather()
	_show_notification(tr("MSG_WEATHER_CHANGED") + ": " + tr(weather))

func _show_notification(message: String, duration: float = 3.0):
	notification_text.text = message
	notification_panel.visible = true
	
	# Hide after duration
	await get_tree().create_timer(duration).timeout
	notification_panel.visible = false

# Button handlers
func _on_market_button_pressed():
	center_tabs.current_tab = 1  # Market tab

func _on_inventory_button_pressed():
	center_tabs.current_tab = 2  # Inventory tab

func _on_bank_button_pressed():
	center_tabs.current_tab = 3  # Bank tab

func _on_save_button_pressed():
	var result = game_api.save_game(0)
	if result.success:
		_show_notification(tr("MSG_GAME_SAVED"))
	else:
		_show_notification(tr("MSG_SAVE_FAILED") + ": " + result.message, 5.0)

func _on_buy_button_pressed():
	var selected_items = market_item_list.get_selected_items()
	if selected_items.is_empty():
		_show_notification(tr("MSG_NO_ITEM_SELECTED"))
		return
	
	var selected_index = selected_items[0]
	var item_keys = market_prices.keys()
	if selected_index >= item_keys.size():
		return
	
	var item_id = item_keys[selected_index]
	var quantity = int(market_quantity_spin.value)
	
	var result = game_api.buy_item(item_id, quantity)
	if result.success:
		_show_notification(tr("MSG_PURCHASE_SUCCESS") + ": " + str(quantity) + " " + tr(item_id.to_upper()))
		_update_player_info()
		_update_inventory()
	else:
		_show_notification(tr("MSG_PURCHASE_FAILED") + ": " + result.message, 5.0)

func _on_deposit_button_pressed():
	var amount = bank_amount_spin.value
	
	var result = game_api.deposit_money(amount)
	if result.success:
		_show_notification(tr("MSG_DEPOSIT_SUCCESS") + ": " + str(amount) + " " + tr("UI_GOLD"))
		_update_player_info()
		_update_bank_info()
	else:
		_show_notification(tr("MSG_DEPOSIT_FAILED") + ": " + result.message, 5.0)

func _on_withdraw_button_pressed():
	var amount = bank_amount_spin.value
	
	var result = game_api.withdraw_money(amount)
	if result.success:
		_show_notification(tr("MSG_WITHDRAW_SUCCESS") + ": " + str(amount) + " " + tr("UI_GOLD"))
		_update_player_info()
		_update_bank_info()
	else:
		_show_notification(tr("MSG_WITHDRAW_FAILED") + ": " + result.message, 5.0)

func _notification(what):
	if what == NOTIFICATION_WM_CLOSE_REQUEST:
		# Auto-save on close
		game_api.save_game(0)