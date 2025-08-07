class_name BasePanel
extends Control

signal panel_opened
signal panel_closed
signal panel_transition_requested(panel_name: String)

@export var panel_name: String = ""
@export var is_modal: bool = false
@export var animate_on_open: bool = true
@export var animate_on_close: bool = true

var is_open: bool = false
var animation_duration: float = 0.3

func _ready() -> void:
	visible = false
	if panel_name == "":
		panel_name = name

func open_panel(data: Dictionary = {}) -> void:
	if is_open:
		return

	is_open = true
	visible = true

	if animate_on_open:
		_animate_open()

	_on_panel_open(data)
	panel_opened.emit()

func close_panel() -> void:
	if not is_open:
		return

	is_open = false

	if animate_on_close:
		await _animate_close()

	visible = false
	_on_panel_close()
	panel_closed.emit()

func _animate_open() -> void:
	modulate.a = 0.0
	scale = Vector2(0.9, 0.9)

	var tween = create_tween()
	tween.set_parallel(true)
	tween.tween_property(self, "modulate:a", 1.0, animation_duration)
	tween.tween_property(self, "scale", Vector2.ONE, animation_duration).set_trans(Tween.TRANS_ELASTIC).set_ease(Tween.EASE_OUT)

func _animate_close() -> void:
	var tween = create_tween()
	tween.set_parallel(true)
	tween.tween_property(self, "modulate:a", 0.0, animation_duration * 0.5)
	tween.tween_property(self, "scale", Vector2(0.9, 0.9), animation_duration * 0.5)
	await tween.finished

func request_transition(target_panel: String, data: Dictionary = {}) -> void:
	panel_transition_requested.emit(target_panel)
	close_panel()

# Virtual methods to be overridden by child classes
func _on_panel_open(data: Dictionary) -> void:
	pass

func _on_panel_close() -> void:
	pass

func refresh_content() -> void:
	pass

func handle_input(event: InputEvent) -> bool:
	return false
