class_name InventoryViewPanel
extends BasePanel

signal item_transferred(item_id: String, quantity: int, to_shop: bool)
signal auto_arrange_requested()

@onready var warehouse_items: VBoxContainer = $WarehousePanel/ItemList/Items
@onready var shop_items: VBoxContainer = $ShopPanel/ItemList/Items
@onready var warehouse_capacity_bar: ProgressBar = $WarehousePanel/CapacityBar
@onready var warehouse_capacity_label: Label = $WarehousePanel/CapacityLabel
@onready var shop_capacity_bar: ProgressBar = $ShopPanel/CapacityBar
@onready var shop_capacity_label: Label = $ShopPanel/CapacityLabel

@onready var to_shop_button: Button = $TransferArea/ToShopButton
@onready var to_warehouse_button: Button = $TransferArea/ToWarehouseButton
@onready var sort_button: Button = $ActionButtons/SortButton
@onready var optimize_button: Button = $ActionButtons/OptimizeButton
@onready var back_button: Button = $ActionButtons/BackButton

var warehouse_inventory: Array = []
var shop_inventory: Array = []
var selected_warehouse_item: Dictionary = {}
var selected_shop_item: Dictionary = {}

var warehouse_capacity: int = 200
var shop_capacity: int = 50

func _ready() -> void:
	super._ready()
	_setup_signals()
	_update_transfer_buttons()

func _setup_signals() -> void:
	if to_shop_button:
		to_shop_button.pressed.connect(_on_transfer_to_shop)
	if to_warehouse_button:
		to_warehouse_button.pressed.connect(_on_transfer_to_warehouse)
	if sort_button:
		sort_button.pressed.connect(_on_sort_pressed)
	if optimize_button:
		optimize_button.pressed.connect(_on_optimize_pressed)
	if back_button:
		back_button.pressed.connect(_on_back_pressed)

func update_inventories(warehouse: Array, shop: Array) -> void:
	warehouse_inventory = warehouse
	shop_inventory = shop
	_refresh_displays()
	_update_capacity_displays()

func _refresh_displays() -> void:
	_clear_item_displays()
	_populate_warehouse_items()
	_populate_shop_items()

func _clear_item_displays() -> void:
	for child in warehouse_items.get_children():
		child.queue_free()
	for child in shop_items.get_children():
		child.queue_free()

func _populate_warehouse_items() -> void:
	for item in warehouse_inventory:
		var item_entry = _create_item_entry(item, false)
		warehouse_items.add_child(item_entry)

func _populate_shop_items() -> void:
	for item in shop_inventory:
		var item_entry = _create_item_entry(item, true)
		shop_items.add_child(item_entry)

func _create_item_entry(item: Dictionary, is_shop: bool) -> Control:
	var hbox = HBoxContainer.new()
	hbox.custom_minimum_size.y = 40
	
	# Selection toggle
	var select_btn = CheckBox.new()
	select_btn.toggled.connect(func(pressed): _on_item_selected(item, is_shop, pressed))
	hbox.add_child(select_btn)
	
	# Item icon placeholder
	var icon = TextureRect.new()
	icon.custom_minimum_size = Vector2(32, 32)
	icon.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
	hbox.add_child(icon)
	
	# Item name
	var name_label = Label.new()
	name_label.text = item.get("name", item.get("id", "Unknown"))
	name_label.custom_minimum_size.x = 150
	hbox.add_child(name_label)
	
	# Quantity
	var qty_label = Label.new()
	qty_label.text = "x" + str(item.get("quantity", 0))
	qty_label.custom_minimum_size.x = 50
	hbox.add_child(qty_label)
	
	# Price (if in shop)
	if is_shop:
		var price_label = Label.new()
		price_label.text = str(item.get("price", 0)) + " G"
		price_label.custom_minimum_size.x = 60
		price_label.modulate = Color.YELLOW
		hbox.add_child(price_label)
	
	# Durability indicator (for perishables)
	var durability = item.get("durability", -1)
	if durability >= 0:
		var dur_label = Label.new()
		if durability <= 1:
			dur_label.text = "⚠"
			dur_label.modulate = Color.RED
		elif durability <= 3:
			dur_label.text = "!"
			dur_label.modulate = Color.YELLOW
		else:
			dur_label.text = str(durability) + "d"
		dur_label.custom_minimum_size.x = 30
		hbox.add_child(dur_label)
	
	return hbox

func _on_item_selected(item: Dictionary, is_shop: bool, selected: bool) -> void:
	if is_shop:
		if selected:
			selected_shop_item = item
		else:
			selected_shop_item = {}
	else:
		if selected:
			selected_warehouse_item = item
		else:
			selected_warehouse_item = {}
	
	_update_transfer_buttons()

func _update_transfer_buttons() -> void:
	if to_shop_button:
		to_shop_button.disabled = selected_warehouse_item.is_empty()
	if to_warehouse_button:
		to_warehouse_button.disabled = selected_shop_item.is_empty()

func _update_capacity_displays() -> void:
	# Calculate warehouse usage
	var warehouse_used = 0
	for item in warehouse_inventory:
		warehouse_used += item.get("quantity", 0)
	
	if warehouse_capacity_bar:
		warehouse_capacity_bar.max_value = warehouse_capacity
		warehouse_capacity_bar.value = warehouse_used
		
		# Color based on capacity
		if warehouse_used >= warehouse_capacity * 0.9:
			warehouse_capacity_bar.modulate = Color.RED
		elif warehouse_used >= warehouse_capacity * 0.7:
			warehouse_capacity_bar.modulate = Color.YELLOW
		else:
			warehouse_capacity_bar.modulate = Color.GREEN
	
	if warehouse_capacity_label:
		warehouse_capacity_label.text = "%d/%d" % [warehouse_used, warehouse_capacity]
	
	# Calculate shop usage
	var shop_used = 0
	for item in shop_inventory:
		shop_used += item.get("quantity", 0)
	
	if shop_capacity_bar:
		shop_capacity_bar.max_value = shop_capacity
		shop_capacity_bar.value = shop_used
		
		# Color based on capacity
		if shop_used >= shop_capacity * 0.9:
			shop_capacity_bar.modulate = Color.RED
		elif shop_used >= shop_capacity * 0.7:
			shop_capacity_bar.modulate = Color.YELLOW
		else:
			shop_capacity_bar.modulate = Color.GREEN
	
	if shop_capacity_label:
		shop_capacity_label.text = "%d/%d" % [shop_used, shop_capacity]

func _on_transfer_to_shop() -> void:
	if selected_warehouse_item.is_empty():
		return
	
	# Check shop capacity
	var shop_used = _calculate_total_quantity(shop_inventory)
	var transfer_qty = selected_warehouse_item.get("quantity", 0)
	
	if shop_used + transfer_qty > shop_capacity:
		_show_capacity_warning("Shop is full!")
		return
	
	item_transferred.emit(
		selected_warehouse_item.get("id", ""),
		transfer_qty,
		true
	)
	
	# Move item
	_transfer_item(selected_warehouse_item, warehouse_inventory, shop_inventory)
	selected_warehouse_item = {}
	_refresh_displays()
	_update_capacity_displays()

func _on_transfer_to_warehouse() -> void:
	if selected_shop_item.is_empty():
		return
	
	# Check warehouse capacity
	var warehouse_used = _calculate_total_quantity(warehouse_inventory)
	var transfer_qty = selected_shop_item.get("quantity", 0)
	
	if warehouse_used + transfer_qty > warehouse_capacity:
		_show_capacity_warning("Warehouse is full!")
		return
	
	item_transferred.emit(
		selected_shop_item.get("id", ""),
		transfer_qty,
		false
	)
	
	# Move item
	_transfer_item(selected_shop_item, shop_inventory, warehouse_inventory)
	selected_shop_item = {}
	_refresh_displays()
	_update_capacity_displays()

func _transfer_item(item: Dictionary, from_inventory: Array, to_inventory: Array) -> void:
	# Remove from source
	from_inventory.erase(item)
	
	# Add to destination
	var existing = _find_item_in_inventory(item.get("id", ""), to_inventory)
	if existing:
		existing["quantity"] += item.get("quantity", 0)
	else:
		to_inventory.append(item.duplicate())

func _find_item_in_inventory(item_id: String, inventory: Array) -> Dictionary:
	for item in inventory:
		if item.get("id", "") == item_id:
			return item
	return {}

func _calculate_total_quantity(inventory: Array) -> int:
	var total = 0
	for item in inventory:
		total += item.get("quantity", 0)
	return total

func _show_capacity_warning(message: String) -> void:
	# TODO: Show warning dialog
	print(message)

func _on_sort_pressed() -> void:
	# Sort both inventories by category then name
	warehouse_inventory.sort_custom(_sort_items)
	shop_inventory.sort_custom(_sort_items)
	_refresh_displays()

func _sort_items(a: Dictionary, b: Dictionary) -> bool:
	var cat_a = a.get("category", "")
	var cat_b = b.get("category", "")
	
	if cat_a != cat_b:
		return cat_a < cat_b
	
	return a.get("name", "") < b.get("name", "")

func _on_optimize_pressed() -> void:
	auto_arrange_requested.emit()

func _on_back_pressed() -> void:
	close_panel()

func _on_panel_open(data: Dictionary) -> void:
	# Load initial inventory data if provided
	if data.has("warehouse"):
		warehouse_inventory = data["warehouse"]
	if data.has("shop"):
		shop_inventory = data["shop"]
	
	_refresh_displays()
	_update_capacity_displays()