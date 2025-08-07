class_name UIManager
extends Node

signal ui_state_changed(new_state: String)

@export var initial_panel: String = "MainMenu"

var panels: Dictionary = {}
var current_panel: BasePanel = null
var panel_stack: Array[BasePanel] = []

func _ready() -> void:
	_register_panels()
	
	if initial_panel != "":
		open_panel(initial_panel)

func _register_panels() -> void:
	# Find all panels in the scene
	var all_panels = get_tree().get_nodes_in_group("ui_panels")
	for panel in all_panels:
		if panel is BasePanel:
			panels[panel.panel_name] = panel
			panel.panel_transition_requested.connect(_on_panel_transition_requested)
			panel.visible = false

func open_panel(panel_name: String, data: Dictionary = {}) -> void:
	if not panels.has(panel_name):
		push_error("Panel not found: " + panel_name)
		return
	
	var panel = panels[panel_name]
	
	# Close current panel if not modal
	if current_panel and not panel.is_modal:
		if current_panel != panel:
			current_panel.close_panel()
	
	# Add to stack if modal
	if panel.is_modal and current_panel:
		panel_stack.append(current_panel)
	
	current_panel = panel
	panel.open_panel(data)
	ui_state_changed.emit(panel_name)

func close_current_panel() -> void:
	if not current_panel:
		return
	
	current_panel.close_panel()
	
	# Pop from stack if there are modal panels
	if panel_stack.size() > 0:
		current_panel = panel_stack.pop_back()
		if current_panel:
			current_panel.visible = true
	else:
		current_panel = null

func _on_panel_transition_requested(target_panel: String) -> void:
	open_panel(target_panel)

func get_current_panel() -> BasePanel:
	return current_panel

func is_panel_open(panel_name: String) -> bool:
	if panels.has(panel_name):
		return panels[panel_name].is_open
	return false

func close_all_panels() -> void:
	panel_stack.clear()
	for panel in panels.values():
		if panel.is_open:
			panel.close_panel()
	current_panel = null