extends BasePanel

## Inventory Placement Panel - handles inventory placement and transfer UI
class_name InventoryPlacementPanel

# Signals
signal transfer_completed(result: Dictionary)
signal bulk_transfer_completed(results: Array)
signal optimization_applied(results: Array)

# UI References
@onready var shop_list: ItemList = $VBoxContainer/ContentArea/HSplitContainer/ShopPanel/ShopList
@onready var warehouse_list: ItemList = $VBoxContainer/ContentArea/HSplitContainer/WarehousePanel/WarehouseList
@onready var shop_capacity_bar: ProgressBar = $VBoxContainer/ContentArea/HSplitContainer/ShopPanel/CapacityBar
@onready var warehouse_capacity_bar: ProgressBar = $VBoxContainer/ContentArea/HSplitContainer/WarehousePanel/CapacityBar
@onready var shop_label: Label = $VBoxContainer/ContentArea/HSplitContainer/ShopPanel/ShopLabel
@onready var warehouse_label: Label = $VBoxContainer/ContentArea/HSplitContainer/WarehousePanel/WarehouseLabel

# Transfer controls
@onready var transfer_slider: HSlider = $VBoxContainer/ContentArea/TransferPanel/TransferSlider
@onready var transfer_quantity_label: Label = $VBoxContainer/ContentArea/TransferPanel/QuantityLabel
@onready var transfer_to_shop_btn: Button = $VBoxContainer/ContentArea/TransferPanel/ToShopButton
@onready var transfer_to_warehouse_btn: Button = $VBoxContainer/ContentArea/TransferPanel/ToWarehouseButton

# Filter controls
@onready var category_filter: OptionButton = $VBoxContainer/HeaderArea/FilterBar/CategoryFilter
@onready var location_filter: OptionButton = $VBoxContainer/HeaderArea/FilterBar/LocationFilter
@onready var perishable_check: CheckBox = $VBoxContainer/HeaderArea/FilterBar/PerishableCheck
@onready var sort_option: OptionButton = $VBoxContainer/HeaderArea/FilterBar/SortOption

# Stats display
@onready var stats_panel: Panel = $VBoxContainer/FooterArea/StatsPanel
@onready var total_value_label: Label = $VBoxContainer/FooterArea/StatsPanel/TotalValueLabel
@onready var total_items_label: Label = $VBoxContainer/FooterArea/StatsPanel/TotalItemsLabel
@onready var expiring_warning: Label = $VBoxContainer/FooterArea/StatsPanel/ExpiringWarning

# Optimization panel
@onready var suggestions_list: ItemList = $VBoxContainer/ContentArea/OptimizationPanel/SuggestionsList
@onready var auto_optimize_btn: Button = $VBoxContainer/ContentArea/OptimizationPanel/AutoOptimizeButton
@onready var apply_suggestion_btn: Button = $VBoxContainer/ContentArea/OptimizationPanel/ApplySuggestionButton

# Data
var game_bridge: GameBridge
var inventory_items: Array = []
var selected_item: Dictionary = {}
var selected_location: String = ""
var suggestions: Array = []
var stats: Dictionary = {}

# Visual settings
const SHOP_COLOR = Color(0.5, 0.8, 0.5)
const WAREHOUSE_COLOR = Color(0.5, 0.6, 0.8)
const PERISHABLE_COLOR = Color(1.0, 0.8, 0.5)
const EXPIRING_COLOR = Color(1.0, 0.5, 0.5)

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
	category_filter.add_item("Magic Book")
	category_filter.add_item("Gem")
	
	# Setup location filter
	location_filter.clear()
	location_filter.add_item("All Locations")
	location_filter.add_item("Shop")
	location_filter.add_item("Warehouse")
	
	# Setup sort options
	sort_option.clear()
	sort_option.add_item("Name")
	sort_option.add_item("Quantity")
	sort_option.add_item("Value")
	sort_option.add_item("Sales Velocity")
	sort_option.add_item("Days in Stock")
	
	# Setup transfer slider
	transfer_slider.min_value = 0
	transfer_slider.max_value = 100
	transfer_slider.value = 1
	transfer_slider.step = 1

func _connect_signals():
	# Connect filter signals
	category_filter.item_selected.connect(_on_filter_changed)
	location_filter.item_selected.connect(_on_filter_changed)
	perishable_check.toggled.connect(_on_filter_changed)
	sort_option.item_selected.connect(_on_filter_changed)
	
	# Connect list signals
	shop_list.item_selected.connect(_on_shop_item_selected)
	warehouse_list.item_selected.connect(_on_warehouse_item_selected)
	
	# Connect transfer signals
	transfer_slider.value_changed.connect(_on_transfer_quantity_changed)
	transfer_to_shop_btn.pressed.connect(_on_transfer_to_shop)
	transfer_to_warehouse_btn.pressed.connect(_on_transfer_to_warehouse)
	
	# Connect optimization signals
	suggestions_list.item_selected.connect(_on_suggestion_selected)
	auto_optimize_btn.pressed.connect(_on_auto_optimize)
	apply_suggestion_btn.pressed.connect(_on_apply_suggestion)

func refresh_data():
	if not game_bridge:
		return
	
	# Get filter settings
	var filter = _get_current_filter()
	
	# Get inventory items
	inventory_items = game_bridge.get_inventory_items(filter)
	
	# Get inventory stats
	stats = game_bridge.get_inventory_stats()
	
	# Get optimization suggestions
	suggestions = game_bridge.get_optimization_suggestions()
	
	# Update UI
	_update_inventory_lists()
	_update_stats_display()
	_update_suggestions_list()
	_update_capacity_bars()

func _update_inventory_lists():
	shop_list.clear()
	warehouse_list.clear()
	
	for item in inventory_items:
		var item_text = "%s x%d (%.2f g/day)" % [
			item.name,
			item.quantity,
			item.sales_velocity
		]
		
		var list_to_add = shop_list if item.location == "shop" else warehouse_list
		list_to_add.add_item(item_text)
		
		var idx = list_to_add.get_item_count() - 1
		list_to_add.set_item_metadata(idx, item)
		
		# Add icon if available
		if item.has("icon") and item.icon != "":
			var icon = load(item.icon)
			if icon:
				list_to_add.set_item_icon(idx, icon)
		
		# Color code based on status
		if item.durability > 0 and item.durability - item.days_in_stock <= 2:
			list_to_add.set_item_custom_bg_color(idx, EXPIRING_COLOR)
		elif item.durability > 0:
			list_to_add.set_item_custom_bg_color(idx, PERISHABLE_COLOR)
		
		# Add tooltip
		var tooltip = _generate_item_tooltip(item)
		list_to_add.set_item_tooltip(idx, tooltip)

func _update_stats_display():
	if stats.is_empty():
		return
	
	# Update labels
	shop_label.text = "Shop (%d/%d)" % [stats.shop_used, stats.shop_capacity]
	warehouse_label.text = "Warehouse (%d/%d)" % [stats.warehouse_used, stats.warehouse_capacity]
	
	total_value_label.text = "Total Value: %.2f gold" % stats.total_value
	total_items_label.text = "Total Items: %d" % stats.total_items
	
	# Update expiring warning
	if stats.expiring_items > 0:
		expiring_warning.text = "⚠ %d items expiring soon!" % stats.expiring_items
		expiring_warning.modulate = Color.RED
		expiring_warning.visible = true
	else:
		expiring_warning.visible = false

func _update_capacity_bars():
	if stats.is_empty():
		return
	
	# Update shop capacity bar
	shop_capacity_bar.max_value = stats.shop_capacity
	shop_capacity_bar.value = stats.shop_used
	
	# Color based on utilization
	if stats.shop_utilization > 0.9:
		shop_capacity_bar.modulate = Color.RED
	elif stats.shop_utilization > 0.7:
		shop_capacity_bar.modulate = Color.YELLOW
	else:
		shop_capacity_bar.modulate = Color.GREEN
	
	# Update warehouse capacity bar
	warehouse_capacity_bar.max_value = stats.warehouse_capacity
	warehouse_capacity_bar.value = stats.warehouse_used
	
	if stats.warehouse_utilization > 0.9:
		warehouse_capacity_bar.modulate = Color.RED
	elif stats.warehouse_utilization > 0.7:
		warehouse_capacity_bar.modulate = Color.YELLOW
	else:
		warehouse_capacity_bar.modulate = Color.GREEN

func _update_suggestions_list():
	suggestions_list.clear()
	
	for suggestion in suggestions:
		var icon_text = _get_suggestion_icon(suggestion.type)
		var text = "%s %s: %s x%d - %s" % [
			icon_text,
			suggestion.item_name,
			suggestion.type,
			suggestion.quantity,
			suggestion.reason
		]
		
		suggestions_list.add_item(text)
		var idx = suggestions_list.get_item_count() - 1
		suggestions_list.set_item_metadata(idx, suggestion)
		
		# Color by priority
		match suggestion.priority:
			5:
				suggestions_list.set_item_custom_fg_color(idx, Color.RED)
			4:
				suggestions_list.set_item_custom_fg_color(idx, Color.ORANGE)
			3:
				suggestions_list.set_item_custom_fg_color(idx, Color.YELLOW)

func _on_shop_item_selected(index: int):
	selected_item = shop_list.get_item_metadata(index)
	selected_location = "shop"
	_update_transfer_controls()

func _on_warehouse_item_selected(index: int):
	selected_item = warehouse_list.get_item_metadata(index)
	selected_location = "warehouse"
	_update_transfer_controls()

func _update_transfer_controls():
	if selected_item.is_empty():
		transfer_to_shop_btn.disabled = true
		transfer_to_warehouse_btn.disabled = true
		return
	
	# Update slider max value
	transfer_slider.max_value = selected_item.quantity
	transfer_slider.value = min(1, selected_item.quantity)
	
	# Enable/disable buttons based on location
	transfer_to_shop_btn.disabled = (selected_location == "shop")
	transfer_to_warehouse_btn.disabled = (selected_location == "warehouse")
	
	# Update quantity label
	_on_transfer_quantity_changed(transfer_slider.value)

func _on_transfer_quantity_changed(value: float):
	var quantity = int(value)
	transfer_quantity_label.text = "Transfer: %d items" % quantity
	
	# Show space requirement
	if not selected_item.is_empty():
		var space_text = " (Requires %d space)" % quantity
		transfer_quantity_label.text += space_text

func _on_transfer_to_shop():
	if selected_item.is_empty() or selected_location != "warehouse":
		return
	
	var request = {
		"item_id": selected_item.item_id,
		"quantity": int(transfer_slider.value),
		"from_location": "warehouse",
		"to_location": "shop"
	}
	
	var result = game_bridge.transfer_inventory_item(request)
	_handle_transfer_result(result)

func _on_transfer_to_warehouse():
	if selected_item.is_empty() or selected_location != "shop":
		return
	
	var request = {
		"item_id": selected_item.item_id,
		"quantity": int(transfer_slider.value),
		"from_location": "shop",
		"to_location": "warehouse"
	}
	
	var result = game_bridge.transfer_inventory_item(request)
	_handle_transfer_result(result)

func _handle_transfer_result(result: Dictionary):
	if result.success:
		_show_success_message("Transferred %d %s from %s to %s" % [
			result.quantity,
			result.item_id,
			result.from_location,
			result.to_location
		])
		refresh_data()
		transfer_completed.emit(result)
	else:
		_show_error_message("Transfer failed: %s" % result.message)

func _on_suggestion_selected(index: int):
	if index < 0:
		return
	
	var suggestion = suggestions_list.get_item_metadata(index)
	apply_suggestion_btn.disabled = false
	apply_suggestion_btn.tooltip_text = "Apply: " + suggestion.reason

func _on_apply_suggestion():
	var selected_idx = suggestions_list.get_selected_items()
	if selected_idx.is_empty():
		return
	
	var suggestion = suggestions_list.get_item_metadata(selected_idx[0])
	
	# Convert suggestion to transfer request
	match suggestion.type:
		"move_to_shop":
			var request = {
				"item_id": suggestion.item_id,
				"quantity": suggestion.quantity,
				"from_location": "warehouse",
				"to_location": "shop"
			}
			var result = game_bridge.transfer_inventory_item(request)
			_handle_transfer_result(result)
		
		"move_to_warehouse":
			var request = {
				"item_id": suggestion.item_id,
				"quantity": suggestion.quantity,
				"from_location": "shop",
				"to_location": "warehouse"
			}
			var result = game_bridge.transfer_inventory_item(request)
			_handle_transfer_result(result)
		
		"sell_soon":
			_show_warning_message("Marked %s for urgent sale" % suggestion.item_name)
		
		"restock":
			_show_info_message("Restock recommendation noted for %s" % suggestion.item_name)

func _on_auto_optimize():
	var result = game_bridge.optimize_inventory_layout()
	if result:
		_show_success_message("Inventory optimized automatically")
		refresh_data()
		optimization_applied.emit(result)
	else:
		_show_error_message("Optimization failed")

func _on_filter_changed(_index = null):
	refresh_data()

func _get_current_filter() -> Dictionary:
	var filter = {}
	
	# Category filter
	var cat_idx = category_filter.selected
	if cat_idx > 0:
		filter["category"] = category_filter.get_item_text(cat_idx).to_lower().replace(" ", "_")
	
	# Location filter
	var loc_idx = location_filter.selected
	if loc_idx > 0:
		filter["location"] = location_filter.get_item_text(loc_idx).to_lower()
	
	# Perishable filter
	if perishable_check.button_pressed:
		filter["perishable"] = true
	
	# Sort settings
	var sort_idx = sort_option.selected
	match sort_idx:
		0: filter["sort_by"] = "name"
		1: filter["sort_by"] = "quantity"
		2: filter["sort_by"] = "value"
		3: filter["sort_by"] = "velocity"
		4: filter["sort_by"] = "age"
	
	filter["sort_order"] = "desc" if Input.is_key_pressed(KEY_SHIFT) else "asc"
	
	return filter

func _generate_item_tooltip(item: Dictionary) -> String:
	var tooltip = "%s\n" % item.name
	tooltip += "Quantity: %d\n" % item.quantity
	tooltip += "Location: %s\n" % item.location.capitalize()
	tooltip += "Purchase Price: %.2f gold\n" % item.purchase_price
	tooltip += "Current Price: %.2f gold\n" % item.current_price
	tooltip += "Profit Margin: %.1f%%\n" % item.profit_margin
	tooltip += "Days in Stock: %d\n" % item.days_in_stock
	
	if item.durability > 0:
		var days_left = item.durability - item.days_in_stock
		tooltip += "⚠ Expires in %d days\n" % days_left
	
	tooltip += "Sales Velocity: %.2f/day" % item.sales_velocity
	
	return tooltip

func _get_suggestion_icon(type: String) -> String:
	match type:
		"move_to_shop": return "📥"
		"move_to_warehouse": return "📤"
		"sell_soon": return "⚠️"
		"restock": return "🔄"
		_: return "💡"

func _show_success_message(message: String):
	print("[SUCCESS] " + message)
	# TODO: Show in UI notification system

func _show_error_message(message: String):
	print("[ERROR] " + message)
	# TODO: Show in UI notification system

func _show_warning_message(message: String):
	print("[WARNING] " + message)
	# TODO: Show in UI notification system

func _show_info_message(message: String):
	print("[INFO] " + message)
	# TODO: Show in UI notification system

# Drag and drop support for quick transfers
func _can_drop_data(position: Vector2, data) -> bool:
	if not data.has("item_id"):
		return false
	
	# Check if dropping on valid target
	var local_pos = get_local_mouse_position()
	return shop_list.get_rect().has_point(local_pos) or warehouse_list.get_rect().has_point(local_pos)

func _drop_data(position: Vector2, data):
	if not data.has("item_id"):
		return
	
	var local_pos = get_local_mouse_position()
	var target_location = ""
	
	if shop_list.get_rect().has_point(local_pos):
		target_location = "shop"
	elif warehouse_list.get_rect().has_point(local_pos):
		target_location = "warehouse"
	else:
		return
	
	# Execute transfer
	var request = {
		"item_id": data.item_id,
		"quantity": data.quantity,
		"from_location": data.location,
		"to_location": target_location
	}
	
	if request.from_location != request.to_location:
		var result = game_bridge.transfer_inventory_item(request)
		_handle_transfer_result(result)