class_name MarketViewPanel
extends BasePanel

signal item_selected(item_data: Dictionary)
signal buy_requested(item_data: Dictionary, quantity: int)
signal refresh_requested()

@onready var price_grid: GridContainer = $TabContainer/CurrentPrices/ItemList/PriceGrid
@onready var chart_area: Panel = $TabContainer/PriceTrends/ChartArea
@onready var item_selector: OptionButton = $TabContainer/PriceTrends/ItemSelector
@onready var time_range: OptionButton = $TabContainer/PriceTrends/TimeRange
@onready var trend_info: RichTextLabel = $TabContainer/PriceTrends/TrendInfo
@onready var events_container: VBoxContainer = $TabContainer/MarketEvents/EventList/Events

@onready var refresh_button: Button = $ActionButtons/RefreshButton
@onready var buy_button: Button = $ActionButtons/BuyButton
@onready var back_button: Button = $ActionButtons/BackButton

var market_data: Dictionary = {}
var selected_item: Dictionary = {}
var price_history: Dictionary = {}

func _ready() -> void:
	super._ready()
	_setup_ui()
	_setup_signals()
	_populate_time_ranges()

func _setup_ui() -> void:
	# Setup grid headers
	_add_grid_header("Item", 150)
	_add_grid_header("Category", 100)
	_add_grid_header("Price", 80)
	_add_grid_header("Trend", 60)
	_add_grid_header("Supply", 80)
	_add_grid_header("Demand", 80)

func _setup_signals() -> void:
	if refresh_button:
		refresh_button.pressed.connect(_on_refresh_pressed)
	if buy_button:
		buy_button.pressed.connect(_on_buy_pressed)
	if back_button:
		back_button.pressed.connect(_on_back_pressed)
	if item_selector:
		item_selector.item_selected.connect(_on_item_selected)
	if time_range:
		time_range.item_selected.connect(_on_time_range_changed)

func _add_grid_header(text: String, min_width: int = 100) -> void:
	var label = Label.new()
	label.text = text
	label.custom_minimum_size.x = min_width
	label.add_theme_style_override("normal", preload("res://themes/header_style.tres"))
	price_grid.add_child(label)

func _populate_time_ranges() -> void:
	if not time_range:
		return
	
	time_range.clear()
	time_range.add_item("1 Day")
	time_range.add_item("7 Days")
	time_range.add_item("30 Days")
	time_range.add_item("All Time")
	time_range.selected = 1  # Default to 7 days

func update_market_data(data: Dictionary) -> void:
	market_data = data
	_refresh_price_display()
	_refresh_item_selector()
	_refresh_events()

func _refresh_price_display() -> void:
	# Clear existing items (except headers)
	for i in range(price_grid.get_child_count() - 1, 5, -1):
		price_grid.get_child(i).queue_free()
	
	# Add market items
	var items = market_data.get("items", [])
	for item in items:
		_add_price_row(item)

func _add_price_row(item: Dictionary) -> void:
	# Item name
	var name_label = Label.new()
	name_label.text = item.get("name", "Unknown")
	name_label.custom_minimum_size.x = 150
	price_grid.add_child(name_label)
	
	# Category
	var category_label = Label.new()
	category_label.text = item.get("category", "")
	category_label.custom_minimum_size.x = 100
	price_grid.add_child(category_label)
	
	# Current price
	var price_label = Label.new()
	var current_price = item.get("currentPrice", 0)
	var base_price = item.get("basePrice", 0)
	price_label.text = str(current_price) + " G"
	price_label.custom_minimum_size.x = 80
	
	# Color code based on price vs base
	if current_price > base_price * 1.1:
		price_label.modulate = Color.GREEN
	elif current_price < base_price * 0.9:
		price_label.modulate = Color.RED
	
	price_grid.add_child(price_label)
	
	# Trend indicator
	var trend_label = Label.new()
	var trend = _calculate_trend(item)
	trend_label.text = trend
	trend_label.custom_minimum_size.x = 60
	if "↑" in trend:
		trend_label.modulate = Color.GREEN
	elif "↓" in trend:
		trend_label.modulate = Color.RED
	price_grid.add_child(trend_label)
	
	# Supply
	var supply_label = Label.new()
	var supply = item.get("supply", 0)
	supply_label.text = _get_supply_text(supply)
	supply_label.custom_minimum_size.x = 80
	price_grid.add_child(supply_label)
	
	# Demand
	var demand_label = Label.new()
	var demand = item.get("demand", 0)
	demand_label.text = _get_demand_text(demand)
	demand_label.custom_minimum_size.x = 80
	price_grid.add_child(demand_label)
	
	# Make row clickable
	name_label.gui_input.connect(func(event): _on_item_clicked(event, item))

func _calculate_trend(item: Dictionary) -> String:
	var current = item.get("currentPrice", 0)
	var base = item.get("basePrice", 0)
	
	if current > base * 1.2:
		return "↑↑"
	elif current > base * 1.05:
		return "↑"
	elif current < base * 0.8:
		return "↓↓"
	elif current < base * 0.95:
		return "↓"
	else:
		return "→"

func _get_supply_text(supply: int) -> String:
	if supply > 150:
		return "Very High"
	elif supply > 100:
		return "High"
	elif supply > 50:
		return "Normal"
	elif supply > 20:
		return "Low"
	else:
		return "Very Low"

func _get_demand_text(demand: int) -> String:
	if demand > 150:
		return "Very High"
	elif demand > 100:
		return "High"
	elif demand > 50:
		return "Normal"
	elif demand > 20:
		return "Low"
	else:
		return "Very Low"

func _refresh_item_selector() -> void:
	if not item_selector:
		return
	
	item_selector.clear()
	item_selector.add_item("Select Item")
	
	var items = market_data.get("items", [])
	for item in items:
		item_selector.add_item(item.get("name", "Unknown"))

func _refresh_events() -> void:
	# Clear existing events
	for child in events_container.get_children():
		child.queue_free()
	
	# Add active events
	var events = market_data.get("events", [])
	if events.is_empty():
		var no_events = Label.new()
		no_events.text = "No active market events"
		no_events.modulate = Color(0.7, 0.7, 0.7)
		events_container.add_child(no_events)
	else:
		for event in events:
			_add_event_entry(event)

func _add_event_entry(event: Dictionary) -> void:
	var event_panel = PanelContainer.new()
	event_panel.custom_minimum_size = Vector2(0, 60)
	
	var vbox = VBoxContainer.new()
	event_panel.add_child(vbox)
	
	var title = Label.new()
	title.text = event.get("name", "Unknown Event")
	title.add_theme_font_size_override("font_size", 18)
	vbox.add_child(title)
	
	var description = Label.new()
	description.text = event.get("description", "")
	description.add_theme_color_override("font_color", Color(0.8, 0.8, 0.8))
	vbox.add_child(description)
	
	var effect = Label.new()
	effect.text = "Effect: " + event.get("effect", "Unknown")
	effect.add_theme_color_override("font_color", Color(1, 0.8, 0.3))
	vbox.add_child(effect)
	
	events_container.add_child(event_panel)

func _on_item_clicked(event: InputEvent, item: Dictionary) -> void:
	if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		selected_item = item
		item_selected.emit(item)
		_highlight_selected_item(item)

func _highlight_selected_item(item: Dictionary) -> void:
	# Update buy button
	if buy_button:
		buy_button.text = "Buy " + item.get("name", "Item")
		buy_button.disabled = false

func _on_item_selected(index: int) -> void:
	if index == 0:  # "Select Item"
		return
	
	# Draw price chart for selected item
	_draw_price_chart(index - 1)

func _on_time_range_changed(index: int) -> void:
	# Refresh chart with new time range
	if item_selector.selected > 0:
		_draw_price_chart(item_selector.selected - 1)

func _draw_price_chart(item_index: int) -> void:
	# This would draw an actual price chart
	# For now, just show trend info
	var items = market_data.get("items", [])
	if item_index < items.size():
		var item = items[item_index]
		_update_trend_info(item)

func _update_trend_info(item: Dictionary) -> void:
	if not trend_info:
		return
	
	var info_text = "[b]%s Price Analysis[/b]\n\n" % item.get("name", "Unknown")
	info_text += "Current Price: [color=yellow]%d G[/color]\n" % item.get("currentPrice", 0)
	info_text += "Base Price: %d G\n" % item.get("basePrice", 0)
	info_text += "Volatility: %s\n" % _get_volatility_text(item.get("volatility", 0))
	info_text += "\n[b]Recommendation:[/b] "
	
	var current = item.get("currentPrice", 0)
	var base = item.get("basePrice", 0)
	
	if current < base * 0.8:
		info_text += "[color=green]Strong Buy[/color] - Price is significantly below average"
	elif current < base * 0.95:
		info_text += "[color=lightgreen]Buy[/color] - Good opportunity"
	elif current > base * 1.2:
		info_text += "[color=red]Sell[/color] - Price is high, good time to sell"
	else:
		info_text += "[color=gray]Hold[/color] - Wait for better opportunity"
	
	trend_info.text = info_text

func _get_volatility_text(volatility: float) -> String:
	if volatility > 0.5:
		return "Very High Risk"
	elif volatility > 0.3:
		return "High Risk"
	elif volatility > 0.15:
		return "Moderate Risk"
	else:
		return "Low Risk"

func _on_refresh_pressed() -> void:
	refresh_requested.emit()

func _on_buy_pressed() -> void:
	if selected_item.is_empty():
		return
	
	# Open buy dialog
	var quantity = 1  # TODO: Show quantity dialog
	buy_requested.emit(selected_item, quantity)

func _on_back_pressed() -> void:
	close_panel()

func _on_panel_open(data: Dictionary) -> void:
	# Refresh market data when panel opens
	refresh_requested.emit()