class_name GameHUD
extends Control

signal menu_requested
signal shop_requested
signal inventory_requested
signal market_requested
signal stats_requested

# HUD Elements
@onready var gold_label: Label = $TopBar/HBoxContainer/GoldContainer/GoldLabel
@onready var rank_label: Label = $TopBar/HBoxContainer/RankContainer/RankLabel
@onready var day_label: Label = $TopBar/HBoxContainer/TimeContainer/DayLabel
@onready var season_label: Label = $TopBar/HBoxContainer/TimeContainer/SeasonLabel
@onready var time_label: Label = $TopBar/HBoxContainer/TimeContainer/TimeLabel

# Quick Access Buttons
@onready var shop_button: Button = $QuickAccess/ShopButton
@onready var inventory_button: Button = $QuickAccess/InventoryButton
@onready var market_button: Button = $QuickAccess/MarketButton
@onready var stats_button: Button = $QuickAccess/StatsButton
@onready var menu_button: Button = $TopBar/MenuButton

# Notification Area
@onready var notification_container: VBoxContainer = $NotificationArea/VBoxContainer

var current_gold: int = 0
var current_rank: String = "Apprentice"
var current_day: int = 1
var current_season: String = "Spring"
var current_time: String = "Morning"

func _ready() -> void:
	_setup_buttons()
	_update_display()

func _setup_buttons() -> void:
	if shop_button:
		shop_button.pressed.connect(_on_shop_button_pressed)
	if inventory_button:
		inventory_button.pressed.connect(_on_inventory_button_pressed)
	if market_button:
		market_button.pressed.connect(_on_market_button_pressed)
	if stats_button:
		stats_button.pressed.connect(_on_stats_button_pressed)
	if menu_button:
		menu_button.pressed.connect(_on_menu_button_pressed)

func _update_display() -> void:
	if gold_label:
		gold_label.text = str(current_gold) + " G"
	if rank_label:
		rank_label.text = current_rank
	if day_label:
		day_label.text = "Day " + str(current_day)
	if season_label:
		season_label.text = current_season
	if time_label:
		time_label.text = current_time

func update_gold(amount: int) -> void:
	current_gold = amount
	if gold_label:
		gold_label.text = str(current_gold) + " G"
		_animate_gold_change()

func update_rank(rank: String) -> void:
	current_rank = rank
	if rank_label:
		rank_label.text = current_rank
		_animate_rank_change()

func update_time(day: int, season: String, time_of_day: String) -> void:
	current_day = day
	current_season = season
	current_time = time_of_day
	_update_display()

func show_notification(message: String, type: String = "info", duration: float = 3.0) -> void:
	var notification = preload("res://scenes/ui/Notification.tscn").instantiate()
	notification_container.add_child(notification)
	notification.setup(message, type, duration)

func _animate_gold_change() -> void:
	if not gold_label:
		return

	var tween = create_tween()
	tween.tween_property(gold_label, "scale", Vector2(1.2, 1.2), 0.1)
	tween.tween_property(gold_label, "scale", Vector2.ONE, 0.1)

func _animate_rank_change() -> void:
	if not rank_label:
		return

	var tween = create_tween()
	tween.tween_property(rank_label, "modulate", Color.YELLOW, 0.2)
	tween.tween_property(rank_label, "modulate", Color.WHITE, 0.3)

func _on_shop_button_pressed() -> void:
	shop_requested.emit()

func _on_inventory_button_pressed() -> void:
	inventory_requested.emit()

func _on_market_button_pressed() -> void:
	market_requested.emit()

func _on_stats_button_pressed() -> void:
	stats_requested.emit()

func _on_menu_button_pressed() -> void:
	menu_requested.emit()

func set_button_enabled(button_name: String, enabled: bool) -> void:
	match button_name:
		"shop":
			if shop_button:
				shop_button.disabled = not enabled
		"inventory":
			if inventory_button:
				inventory_button.disabled = not enabled
		"market":
			if market_button:
				market_button.disabled = not enabled
		"stats":
			if stats_button:
				stats_button.disabled = not enabled
