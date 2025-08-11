extends BasePanel

## Purchase Panel - handles the purchase UI
class_name PurchasePanel

# Signals
signal purchase_completed(result: Dictionary)
signal bulk_purchase_completed(results: Array)
signal preset_executed(preset_id: String, results: Array)

# UI References
@onready var item_list: ItemList = $VBoxContainer/ContentArea/HSplitContainer/ItemListPanel/ItemList
@onready var item_details: RichTextLabel = $VBoxContainer/ContentArea/HSplitContainer/ItemDetailsPanel/ScrollContainer/ItemDetails
@onready var category_filter: OptionButton = $VBoxContainer/HeaderArea/FilterBar/CategoryFilter
@onready var sort_option: OptionButton = $VBoxContainer/HeaderArea/FilterBar/SortOption
@onready var search_bar: LineEdit = $VBoxContainer/HeaderArea/FilterBar/SearchBar

# Purchase controls
@onready var quantity_slider: HSlider = $VBoxContainer/ContentArea/HSplitContainer/ItemDetailsPanel/PurchaseControls/QuantitySlider
@onready var quantity_label: Label = $VBoxContainer/ContentArea/HSplitContainer/ItemDetailsPanel/PurchaseControls/QuantityLabel
@onready var price_label: Label = $VBoxContainer/ContentArea/HSplitContainer/ItemDetailsPanel/PurchaseControls/PriceLabel
@onready var total_cost_label: Label = $VBoxContainer/ContentArea/HSplitContainer/ItemDetailsPanel/PurchaseControls/TotalCostLabel
@onready var negotiate_check: CheckBox = $VBoxContainer/ContentArea/HSplitContainer/ItemDetailsPanel/PurchaseControls/NegotiateCheck
@onready var buy_button: Button = $VBoxContainer/ContentArea/HSplitContainer/ItemDetailsPanel/PurchaseControls/BuyButton

# Quick buy presets
@onready var presets_container: HBoxContainer = $VBoxContainer/FooterArea/PresetsContainer

# Bulk purchase
@onready var bulk_list: ItemList = $VBoxContainer/ContentArea/BulkPurchasePanel/BulkList
@onready var bulk_total_label: Label = $VBoxContainer/ContentArea/BulkPurchasePanel/BulkTotalLabel
@onready var bulk_buy_button: Button = $VBoxContainer/ContentArea/BulkPurchasePanel/BulkBuyButton

# Data
var game_bridge: GameBridge
var purchase_options: Array = []
var selected_item: Dictionary = {}
var bulk_items: Array = []
var presets: Array = []
var player_gold: float = 0.0
var inventory_space: int = 0

func _ready():
	super._ready()
	_setup_ui()
	_connect_signals()
	
	# Get game bridge reference
	game_bridge = get_node("/root/GameBridge")
	if game_bridge:
		refresh_data()

func _setup_ui():
	# Setup category filter
	category_filter.clear()
	category_filter.add_item("All Categories")
	category_filter.add_item("Fruit")
	category_filter.add_item("Potion")
	category_filter.add_item("Weapon")
	category_filter.add_item("Accessory")
	category_filter.add_item("Spellbook")
	category_filter.add_item("Gem")
	
	# Setup sort options
	sort_option.clear()
	sort_option.add_item("Name")
	sort_option.add_item("Price")
	sort_option.add_item("Profit Potential")
	sort_option.add_item("Risk Level")
	sort_option.add_item("Market Trend")
	
	# Setup quantity slider
	quantity_slider.min_value = 1
	quantity_slider.max_value = 100
	quantity_slider.value = 1
	quantity_slider.step = 1

func _connect_signals():
	# Connect UI signals
	category_filter.item_selected.connect(_on_category_filter_changed)
	sort_option.item_selected.connect(_on_sort_option_changed)
	search_bar.text_changed.connect(_on_search_text_changed)
	item_list.item_selected.connect(_on_item_selected)
	quantity_slider.value_changed.connect(_on_quantity_changed)
	negotiate_check.toggled.connect(_on_negotiate_toggled)
	buy_button.pressed.connect(_on_buy_button_pressed)
	bulk_buy_button.pressed.connect(_on_bulk_buy_pressed)

func refresh_data():
	if not game_bridge:
		return
	
	# Get purchase options from backend
	var category = _get_selected_category()
	var sort_by = _get_selected_sort()
	purchase_options = game_bridge.get_purchase_options(category, sort_by)
	
	# Get player info
	player_gold = game_bridge.get_player_gold()
	inventory_space = game_bridge.get_inventory_space()
	
	# Get presets
	presets = game_bridge.get_quick_buy_presets()
	
	# Update UI
	_update_item_list()
	_update_presets()
	_update_player_info()

func _update_item_list():
	item_list.clear()
	
	for option in purchase_options:
		# Apply search filter
		if search_bar.text != "" and not option.name.to_lower().contains(search_bar.text.to_lower()):
			continue
		
		# Create item entry
		var item_text = "%s - %s gold" % [option.name, option.current_price]
		item_list.add_item(item_text)
		
		# Add metadata
		var idx = item_list.get_item_count() - 1
		item_list.set_item_metadata(idx, option)
		
		# Add icon if available
		if option.has("icon") and option.icon != "":
			var icon = load(option.icon)
			if icon:
				item_list.set_item_icon(idx, icon)
		
		# Color code by trend
		match option.market_trend:
			"up":
				item_list.set_item_custom_bg_color(idx, Color(1.0, 0.9, 0.9))
			"down":
				item_list.set_item_custom_bg_color(idx, Color(0.9, 1.0, 0.9))
			_:
				pass

func _on_item_selected(index: int):
	if index < 0:
		return
	
	selected_item = item_list.get_item_metadata(index)
	_update_item_details()
	_update_purchase_controls()

func _update_item_details():
	if selected_item.is_empty():
		item_details.text = "Select an item to view details"
		return
	
	var details_text = "[b]%s[/b]\n\n" % selected_item.name
	details_text += "%s\n\n" % selected_item.description
	
	# Market information
	details_text += "[b]Market Information:[/b]\n"
	details_text += "Current Price: [color=yellow]%s gold[/color]\n" % selected_item.current_price
	details_text += "Market Trend: %s %s\n" % [_get_trend_icon(selected_item.market_trend), selected_item.market_trend]
	details_text += "Price Change: %s%%\n" % _format_percentage(selected_item.price_change)
	details_text += "Supply Level: %s\n" % selected_item.supply_level
	details_text += "\n"
	
	# Investment analysis
	details_text += "[b]Investment Analysis:[/b]\n"
	details_text += "Profit Potential: [color=green]%s%%[/color]\n" % selected_item.profit_potential
	details_text += "Risk Level: %s\n" % _get_risk_color(selected_item.risk_level)
	details_text += "Recommended Quantity: %d\n" % selected_item.recommended_qty
	details_text += "\n"
	
	# Warnings
	if selected_item.has("warnings") and selected_item.warnings.size() > 0:
		details_text += "[b]Warnings:[/b]\n"
		for warning in selected_item.warnings:
			details_text += "⚠ %s\n" % warning
	
	item_details.bbcode_text = details_text

func _update_purchase_controls():
	if selected_item.is_empty():
		buy_button.disabled = true
		return
	
	# Update quantity slider
	quantity_slider.max_value = min(selected_item.max_quantity, _calculate_max_affordable())
	quantity_slider.value = min(selected_item.recommended_qty, quantity_slider.max_value)
	
	# Update labels
	_on_quantity_changed(quantity_slider.value)
	
	# Enable buy button
	buy_button.disabled = false

func _on_quantity_changed(value: float):
	var quantity = int(value)
	quantity_label.text = "Quantity: %d" % quantity
	
	if not selected_item.is_empty():
		var unit_price = selected_item.current_price
		if negotiate_check.button_pressed:
			unit_price *= 0.9  # 10% discount when negotiating
		
		var total_cost = unit_price * quantity
		price_label.text = "Price: %s gold" % unit_price
		total_cost_label.text = "Total: %s gold" % total_cost
		
		# Color code based on affordability
		if total_cost > player_gold:
			total_cost_label.modulate = Color.RED
			buy_button.disabled = true
		else:
			total_cost_label.modulate = Color.WHITE
			buy_button.disabled = false

func _on_negotiate_toggled(pressed: bool):
	_on_quantity_changed(quantity_slider.value)

func _on_buy_button_pressed():
	if selected_item.is_empty():
		return
	
	var purchase_request = {
		"item_id": selected_item.item_id,
		"quantity": int(quantity_slider.value),
		"max_price": selected_item.current_price * 1.1,  # Allow 10% price increase
		"negotiate_price": negotiate_check.button_pressed
	}
	
	var result = game_bridge.execute_purchase(purchase_request)
	_handle_purchase_result(result)

func _handle_purchase_result(result: Dictionary):
	if result.success:
		# Show success message
		_show_success_message("Purchase successful! Bought %d %s for %s gold" % [
			result.quantity,
			result.item_id,
			result.total_cost
		])
		
		# Update player info
		player_gold = result.gold_remaining
		inventory_space = result.inventory_space
		_update_player_info()
		
		# Refresh item list to update supply/demand
		refresh_data()
		
		# Emit signal
		purchase_completed.emit(result)
	else:
		# Show error message
		_show_error_message("Purchase failed: %s" % result.message)
	
	# Show warnings if any
	if result.has("warnings") and result.warnings.size() > 0:
		for warning in result.warnings:
			_show_warning_message(warning)

func _update_presets():
	# Clear existing preset buttons
	for child in presets_container.get_children():
		child.queue_free()
	
	# Create preset buttons
	for preset in presets:
		var button = Button.new()
		button.text = preset.name
		button.tooltip_text = "%s\nTotal Cost: %s gold" % [preset.description, preset.total_cost]
		button.pressed.connect(_on_preset_button_pressed.bind(preset.id))
		
		# Color code based on affordability
		if preset.total_cost > player_gold:
			button.modulate = Color(0.7, 0.7, 0.7)
			button.disabled = true
		
		presets_container.add_child(button)

func _on_preset_button_pressed(preset_id: String):
	var results = game_bridge.execute_quick_buy(preset_id)
	_handle_bulk_purchase_results(results)
	preset_executed.emit(preset_id, results)

func _handle_bulk_purchase_results(results: Array):
	var success_count = 0
	var total_spent = 0.0
	
	for result in results:
		if result.success:
			success_count += 1
			total_spent += result.total_cost
	
	if success_count > 0:
		_show_success_message("Bulk purchase: %d/%d successful, spent %s gold" % [
			success_count,
			results.size(),
			total_spent
		])
		refresh_data()
	else:
		_show_error_message("Bulk purchase failed")
	
	bulk_purchase_completed.emit(results)

func _update_player_info():
	# Update header with player info
	var info_label = $VBoxContainer/HeaderArea/PlayerInfo
	if info_label:
		info_label.text = "Gold: %s | Inventory: %d spaces" % [player_gold, inventory_space]

# Helper functions
func _get_selected_category() -> String:
	var idx = category_filter.selected
	if idx <= 0:
		return "all"
	return category_filter.get_item_text(idx).to_lower()

func _get_selected_sort() -> String:
	var idx = sort_option.selected
	match idx:
		0: return "name"
		1: return "price"
		2: return "profit"
		3: return "risk"
		4: return "trend"
		_: return "name"

func _calculate_max_affordable() -> int:
	if selected_item.is_empty():
		return 0
	
	var unit_price = selected_item.current_price
	if negotiate_check.button_pressed:
		unit_price *= 0.9
	
	if unit_price <= 0:
		return 0
	
	return int(player_gold / unit_price)

func _get_trend_icon(trend: String) -> String:
	match trend:
		"up": return "📈"
		"down": return "📉"
		_: return "➡"

func _format_percentage(value: float) -> String:
	if value > 0:
		return "[color=green]+%.1f[/color]" % value
	elif value < 0:
		return "[color=red]%.1f[/color]" % value
	else:
		return "0.0"

func _get_risk_color(risk: String) -> String:
	match risk:
		"low": return "[color=green]Low[/color]"
		"medium": return "[color=yellow]Medium[/color]"
		"high": return "[color=red]High[/color]"
		_: return risk

func _show_success_message(message: String):
	print("[SUCCESS] " + message)
	# TODO: Show in UI notification system

func _show_error_message(message: String):
	print("[ERROR] " + message)
	# TODO: Show in UI notification system

func _show_warning_message(message: String):
	print("[WARNING] " + message)
	# TODO: Show in UI notification system

func _on_category_filter_changed(index: int):
	refresh_data()

func _on_sort_option_changed(index: int):
	refresh_data()

func _on_search_text_changed(text: String):
	_update_item_list()

func _on_bulk_buy_pressed():
	if bulk_items.is_empty():
		return
	
	var bulk_request = {
		"purchases": bulk_items,
		"total_budget": player_gold,
		"optimize_for_profit": true
	}
	
	var results = game_bridge.execute_bulk_purchase(bulk_request)
	_handle_bulk_purchase_results(results)