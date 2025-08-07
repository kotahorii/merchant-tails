class_name ShopViewPanel
extends BasePanel

signal item_selected(item_data: Dictionary)
signal item_purchased(item_data: Dictionary)
signal item_sold(item_data: Dictionary)
signal price_changed(item_id: String, new_price: int)

@onready var shop_grid: GridContainer = $ScrollContainer/ShopGrid
@onready var item_details: Control = $ItemDetails
@onready var gold_display: Label = $TopBar/GoldLabel
@onready var capacity_display: Label = $TopBar/CapacityLabel
@onready var filter_buttons: HBoxContainer = $FilterBar/FilterButtons
@onready var search_bar: LineEdit = $FilterBar/SearchBar

var shop_inventory: Array[Dictionary] = []
var filtered_inventory: Array[Dictionary] = []
var selected_item: Dictionary = {}
var current_filter: String = "all"
var search_text: String = ""

# Shop configuration
var max_shop_capacity: int = 50
var current_capacity: int = 0
var player_gold: int = 1000

func _ready() -> void:
	super._ready()
	panel_name = "ShopView"
	_setup_ui()
	_load_shop_inventory()

func _setup_ui() -> void:
	if search_bar:
		search_bar.text_changed.connect(_on_search_text_changed)
	
	_create_filter_buttons()
	_update_displays()

func _create_filter_buttons() -> void:
	var categories = ["All", "Fruits", "Potions", "Weapons", "Accessories", "Books", "Gems"]
	
	for category in categories:
		var button = Button.new()
		button.text = category
		button.toggle_mode = true
		button.pressed.connect(_on_filter_button_pressed.bind(category.to_lower()))
		filter_buttons.add_child(button)
		
		if category == "All":
			button.button_pressed = true

func _load_shop_inventory() -> void:
	# Load inventory from game state
	# This would normally come from the Go backend
	shop_inventory = [
		{
			"id": "apple_001",
			"name": "Fresh Apple",
			"category": "fruits",
			"price": 10,
			"quantity": 5,
			"quality": "normal",
			"description": "A fresh, crisp apple"
		},
		{
			"id": "potion_001",
			"name": "Health Potion",
			"category": "potions",
			"price": 50,
			"quantity": 3,
			"quality": "good",
			"description": "Restores health"
		}
	]
	
	_refresh_display()

func _refresh_display() -> void:
	# Clear existing items
	for child in shop_grid.get_children():
		child.queue_free()
	
	# Apply filters
	filtered_inventory = shop_inventory.duplicate()
	
	if current_filter != "all":
		filtered_inventory = filtered_inventory.filter(
			func(item): return item.category == current_filter
		)
	
	if search_text != "":
		filtered_inventory = filtered_inventory.filter(
			func(item): return search_text.to_lower() in item.name.to_lower()
		)
	
	# Create item displays
	for item in filtered_inventory:
		var item_display = _create_item_display(item)
		shop_grid.add_child(item_display)
	
	_update_displays()

func _create_item_display(item_data: Dictionary) -> Control:
	var container = PanelContainer.new()
	container.custom_minimum_size = Vector2(150, 180)
	
	var vbox = VBoxContainer.new()
	container.add_child(vbox)
	
	# Item icon (placeholder)
	var icon = TextureRect.new()
	icon.custom_minimum_size = Vector2(100, 100)
	icon.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
	vbox.add_child(icon)
	
	# Item name
	var name_label = Label.new()
	name_label.text = item_data.name
	name_label.add_theme_font_size_override("font_size", 14)
	vbox.add_child(name_label)
	
	# Price
	var price_container = HBoxContainer.new()
	vbox.add_child(price_container)
	
	var price_label = Label.new()
	price_label.text = str(item_data.price) + " G"
	price_label.add_theme_color_override("font_color", Color.GOLD)
	price_container.add_child(price_label)
	
	# Quantity
	var quantity_label = Label.new()
	quantity_label.text = "x" + str(item_data.quantity)
	price_container.add_child(quantity_label)
	
	# Make clickable
	container.gui_input.connect(_on_item_clicked.bind(item_data))
	
	return container

func _on_item_clicked(event: InputEvent, item_data: Dictionary) -> void:
	if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		selected_item = item_data
		item_selected.emit(item_data)
		_show_item_details(item_data)

func _show_item_details(item_data: Dictionary) -> void:
	if not item_details:
		return
	
	item_details.visible = true
	# Update detail panel with item information
	# This would show more detailed info and action buttons

func _update_displays() -> void:
	if gold_display:
		gold_display.text = "Gold: " + str(player_gold) + " G"
	
	if capacity_display:
		current_capacity = shop_inventory.reduce(
			func(sum, item): return sum + item.quantity, 0
		)
		capacity_display.text = "Capacity: " + str(current_capacity) + "/" + str(max_shop_capacity)

func _on_filter_button_pressed(category: String) -> void:
	current_filter = category
	
	# Update button states
	for button in filter_buttons.get_children():
		button.button_pressed = button.text.to_lower() == category
	
	_refresh_display()

func _on_search_text_changed(new_text: String) -> void:
	search_text = new_text
	_refresh_display()

func set_player_gold(amount: int) -> void:
	player_gold = amount
	_update_displays()

func add_item_to_shop(item_data: Dictionary) -> bool:
	if current_capacity + item_data.quantity > max_shop_capacity:
		return false
	
	# Check if item already exists
	var existing = shop_inventory.filter(
		func(item): return item.id == item_data.id
	)
	
	if existing.size() > 0:
		existing[0].quantity += item_data.quantity
	else:
		shop_inventory.append(item_data)
	
	_refresh_display()
	return true

func remove_item_from_shop(item_id: String, quantity: int) -> bool:
	for item in shop_inventory:
		if item.id == item_id:
			if item.quantity >= quantity:
				item.quantity -= quantity
				if item.quantity == 0:
					shop_inventory.erase(item)
				_refresh_display()
				return true
			break
	return false

func update_item_price(item_id: String, new_price: int) -> void:
	for item in shop_inventory:
		if item.id == item_id:
			item.price = new_price
			price_changed.emit(item_id, new_price)
			_refresh_display()
			break