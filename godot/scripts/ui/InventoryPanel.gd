class_name InventoryPanel
extends BasePanel

signal item_moved_to_shop(item_data: Dictionary, quantity: int)
signal item_moved_to_warehouse(item_data: Dictionary, quantity: int)
signal item_discarded(item_data: Dictionary, quantity: int)

@onready var shop_inventory: ItemList = $HSplitContainer/ShopSection/ShopInventory
@onready var warehouse_inventory: ItemList = $HSplitContainer/WarehouseSection/WarehouseInventory
@onready var item_info: RichTextLabel = $HSplitContainer/ItemInfo/ScrollContainer/ItemDetails
@onready var shop_capacity_bar: ProgressBar = $HSplitContainer/ShopSection/CapacityBar
@onready var warehouse_capacity_bar: ProgressBar = $HSplitContainer/WarehouseSection/CapacityBar
@onready var move_to_shop_btn: Button = $HSplitContainer/ItemInfo/Actions/MoveToShop
@onready var move_to_warehouse_btn: Button = $HSplitContainer/ItemInfo/Actions/MoveToWarehouse
@onready var discard_btn: Button = $HSplitContainer/ItemInfo/Actions/Discard
@onready var quantity_slider: HSlider = $HSplitContainer/ItemInfo/Actions/QuantitySlider
@onready var quantity_label: Label = $HSplitContainer/ItemInfo/Actions/QuantityLabel

var shop_items: Array[Dictionary] = []
var warehouse_items: Array[Dictionary] = []
var selected_item: Dictionary = {}
var selected_location: String = ""
var max_shop_capacity: int = 50
var max_warehouse_capacity: int = 100

func _ready() -> void:
	super._ready()
	panel_name = "Inventory"
	_setup_ui()
	_load_inventory_data()

func _setup_ui() -> void:
	# Connect signals
	if shop_inventory:
		shop_inventory.item_selected.connect(_on_shop_item_selected)
	if warehouse_inventory:
		warehouse_inventory.item_selected.connect(_on_warehouse_item_selected)
	if move_to_shop_btn:
		move_to_shop_btn.pressed.connect(_on_move_to_shop)
	if move_to_warehouse_btn:
		move_to_warehouse_btn.pressed.connect(_on_move_to_warehouse)
	if discard_btn:
		discard_btn.pressed.connect(_on_discard_item)
	if quantity_slider:
		quantity_slider.value_changed.connect(_on_quantity_changed)
	
	# Setup capacity bars
	if shop_capacity_bar:
		shop_capacity_bar.max_value = max_shop_capacity
	if warehouse_capacity_bar:
		warehouse_capacity_bar.max_value = max_warehouse_capacity

func _load_inventory_data() -> void:
	# Load from game state (placeholder data)
	shop_items = [
		{
			"id": "apple_001",
			"name": "Fresh Apple",
			"category": "fruits",
			"quantity": 10,
			"quality": "normal",
			"weight": 0.5,
			"value": 10
		},
		{
			"id": "sword_001",
			"name": "Iron Sword",
			"category": "weapons",
			"quantity": 2,
			"quality": "good",
			"weight": 3.0,
			"value": 150
		}
	]
	
	warehouse_items = [
		{
			"id": "potion_001",
			"name": "Health Potion",
			"category": "potions",
			"quantity": 20,
			"quality": "normal",
			"weight": 0.3,
			"value": 50
		},
		{
			"id": "gem_001",
			"name": "Ruby",
			"category": "gems",
			"quantity": 5,
			"quality": "excellent",
			"weight": 0.1,
			"value": 500
		}
	]
	
	_refresh_displays()

func _refresh_displays() -> void:
	_update_shop_display()
	_update_warehouse_display()
	_update_capacity_bars()

func _update_shop_display() -> void:
	if not shop_inventory:
		return
	
	shop_inventory.clear()
	for item in shop_items:
		var text = "%s (x%d)" % [item.name, item.quantity]
		shop_inventory.add_item(text)
		
		# Set item metadata
		var idx = shop_inventory.get_item_count() - 1
		shop_inventory.set_item_metadata(idx, item)
		
		# Color code by quality
		match item.quality:
			"excellent":
				shop_inventory.set_item_custom_fg_color(idx, Color.GOLD)
			"good":
				shop_inventory.set_item_custom_fg_color(idx, Color.LIGHT_BLUE)
			"poor":
				shop_inventory.set_item_custom_fg_color(idx, Color.GRAY)

func _update_warehouse_display() -> void:
	if not warehouse_inventory:
		return
	
	warehouse_inventory.clear()
	for item in warehouse_items:
		var text = "%s (x%d)" % [item.name, item.quantity]
		warehouse_inventory.add_item(text)
		
		var idx = warehouse_inventory.get_item_count() - 1
		warehouse_inventory.set_item_metadata(idx, item)
		
		match item.quality:
			"excellent":
				warehouse_inventory.set_item_custom_fg_color(idx, Color.GOLD)
			"good":
				warehouse_inventory.set_item_custom_fg_color(idx, Color.LIGHT_BLUE)
			"poor":
				warehouse_inventory.set_item_custom_fg_color(idx, Color.GRAY)

func _update_capacity_bars() -> void:
	var shop_weight = _calculate_total_weight(shop_items)
	var warehouse_weight = _calculate_total_weight(warehouse_items)
	
	if shop_capacity_bar:
		shop_capacity_bar.value = shop_weight
		shop_capacity_bar.modulate = Color.GREEN if shop_weight < max_shop_capacity * 0.8 else Color.YELLOW
		if shop_weight >= max_shop_capacity:
			shop_capacity_bar.modulate = Color.RED
	
	if warehouse_capacity_bar:
		warehouse_capacity_bar.value = warehouse_weight
		warehouse_capacity_bar.modulate = Color.GREEN if warehouse_weight < max_warehouse_capacity * 0.8 else Color.YELLOW
		if warehouse_weight >= max_warehouse_capacity:
			warehouse_capacity_bar.modulate = Color.RED

func _calculate_total_weight(items: Array) -> float:
	var total = 0.0
	for item in items:
		total += item.weight * item.quantity
	return total

func _on_shop_item_selected(index: int) -> void:
	if index < 0 or index >= shop_inventory.get_item_count():
		return
	
	selected_item = shop_inventory.get_item_metadata(index)
	selected_location = "shop"
	_display_item_info()
	_update_action_buttons()

func _on_warehouse_item_selected(index: int) -> void:
	if index < 0 or index >= warehouse_inventory.get_item_count():
		return
	
	selected_item = warehouse_inventory.get_item_metadata(index)
	selected_location = "warehouse"
	_display_item_info()
	_update_action_buttons()

func _display_item_info() -> void:
	if not item_info or selected_item.is_empty():
		return
	
	var info_text = "[b]%s[/b]\n" % selected_item.name
	info_text += "Category: %s\n" % selected_item.category.capitalize()
	info_text += "Quality: [color=%s]%s[/color]\n" % [_get_quality_color(selected_item.quality), selected_item.quality.capitalize()]
	info_text += "Quantity: %d\n" % selected_item.quantity
	info_text += "Weight: %.1f per unit\n" % selected_item.weight
	info_text += "Value: %d G per unit\n" % selected_item.value
	info_text += "Total Value: %d G\n" % (selected_item.value * selected_item.quantity)
	
	item_info.bbcode_text = info_text
	
	# Update quantity slider
	if quantity_slider:
		quantity_slider.max_value = selected_item.quantity
		quantity_slider.value = 1
	if quantity_label:
		quantity_label.text = "1"

func _get_quality_color(quality: String) -> String:
	match quality:
		"excellent":
			return "gold"
		"good":
			return "aqua"
		"normal":
			return "white"
		"poor":
			return "gray"
		_:
			return "white"

func _update_action_buttons() -> void:
	if not selected_item.is_empty():
		if move_to_shop_btn:
			move_to_shop_btn.disabled = (selected_location == "shop")
		if move_to_warehouse_btn:
			move_to_warehouse_btn.disabled = (selected_location == "warehouse")
		if discard_btn:
			discard_btn.disabled = false
	else:
		if move_to_shop_btn:
			move_to_shop_btn.disabled = true
		if move_to_warehouse_btn:
			move_to_warehouse_btn.disabled = true
		if discard_btn:
			discard_btn.disabled = true

func _on_quantity_changed(value: float) -> void:
	if quantity_label:
		quantity_label.text = str(int(value))

func _on_move_to_shop() -> void:
	if selected_item.is_empty() or selected_location != "warehouse":
		return
	
	var quantity = int(quantity_slider.value) if quantity_slider else 1
	
	# Check capacity
	var new_weight = _calculate_total_weight(shop_items) + (selected_item.weight * quantity)
	if new_weight > max_shop_capacity:
		# Show error notification
		return
	
	# Move item
	_move_item(warehouse_items, shop_items, selected_item.id, quantity)
	item_moved_to_shop.emit(selected_item, quantity)
	_refresh_displays()

func _on_move_to_warehouse() -> void:
	if selected_item.is_empty() or selected_location != "shop":
		return
	
	var quantity = int(quantity_slider.value) if quantity_slider else 1
	
	# Check capacity
	var new_weight = _calculate_total_weight(warehouse_items) + (selected_item.weight * quantity)
	if new_weight > max_warehouse_capacity:
		# Show error notification
		return
	
	# Move item
	_move_item(shop_items, warehouse_items, selected_item.id, quantity)
	item_moved_to_warehouse.emit(selected_item, quantity)
	_refresh_displays()

func _on_discard_item() -> void:
	if selected_item.is_empty():
		return
	
	# Show confirmation dialog
	var quantity = int(quantity_slider.value) if quantity_slider else 1
	
	# Remove item
	var source = shop_items if selected_location == "shop" else warehouse_items
	_remove_item_quantity(source, selected_item.id, quantity)
	
	item_discarded.emit(selected_item, quantity)
	selected_item = {}
	_refresh_displays()
	_display_item_info()

func _move_item(source: Array, destination: Array, item_id: String, quantity: int) -> void:
	# Find and remove from source
	for item in source:
		if item.id == item_id:
			if item.quantity <= quantity:
				source.erase(item)
				_add_item_to_array(destination, item)
			else:
				item.quantity -= quantity
				var moved_item = item.duplicate()
				moved_item.quantity = quantity
				_add_item_to_array(destination, moved_item)
			break

func _add_item_to_array(array: Array, item_data: Dictionary) -> void:
	# Check if item already exists
	for existing in array:
		if existing.id == item_data.id:
			existing.quantity += item_data.quantity
			return
	
	# Add as new item
	array.append(item_data)

func _remove_item_quantity(array: Array, item_id: String, quantity: int) -> void:
	for item in array:
		if item.id == item_id:
			if item.quantity <= quantity:
				array.erase(item)
			else:
				item.quantity -= quantity
			break