class_name MainMenuPanel
extends BasePanel

@onready var new_game_button: Button = $VBoxContainer/NewGameButton
@onready var continue_button: Button = $VBoxContainer/ContinueButton
@onready var load_button: Button = $VBoxContainer/LoadButton
@onready var settings_button: Button = $VBoxContainer/SettingsButton
@onready var quit_button: Button = $VBoxContainer/QuitButton
@onready var version_label: Label = $VersionLabel

var has_save_file: bool = false

func _ready() -> void:
	super._ready()
	panel_name = "MainMenu"
	_setup_buttons()
	_check_save_files()
	_display_version()

func _setup_buttons() -> void:
	if new_game_button:
		new_game_button.pressed.connect(_on_new_game_pressed)
	if continue_button:
		continue_button.pressed.connect(_on_continue_pressed)
	if load_button:
		load_button.pressed.connect(_on_load_pressed)
	if settings_button:
		settings_button.pressed.connect(_on_settings_pressed)
	if quit_button:
		quit_button.pressed.connect(_on_quit_pressed)

func _check_save_files() -> void:
	# Check if save files exist
	var save_dir = DirAccess.open("user://saves/")
	if save_dir:
		has_save_file = save_dir.get_files().size() > 0
	
	# Enable/disable continue button based on save files
	if continue_button:
		continue_button.disabled = not has_save_file
		if not has_save_file:
			continue_button.modulate.a = 0.5

func _display_version() -> void:
	if version_label:
		var version = ProjectSettings.get_setting("application/config/version", "0.0.0")
		version_label.text = "v" + version

func _on_new_game_pressed() -> void:
	# Transition to character creation or directly to game
	request_transition("CharacterCreation")
	# Or start new game directly
	# GameManager.start_new_game()

func _on_continue_pressed() -> void:
	if has_save_file:
		# Load most recent save
		# GameManager.load_most_recent_save()
		get_tree().change_scene_to_file("res://scenes/game/GameScene.tscn")

func _on_load_pressed() -> void:
	request_transition("LoadGame")

func _on_settings_pressed() -> void:
	request_transition("Settings")

func _on_quit_pressed() -> void:
	# Show confirmation dialog
	var confirm_dialog = AcceptDialog.new()
	confirm_dialog.dialog_text = "Are you sure you want to quit?"
	confirm_dialog.add_cancel_button("Cancel")
	add_child(confirm_dialog)
	confirm_dialog.popup_centered()
	await confirm_dialog.confirmed
	get_tree().quit()

func _on_panel_open(data: Dictionary) -> void:
	_check_save_files()
	# Play menu music if needed
	# AudioManager.play_menu_music()

func _input(event: InputEvent) -> void:
	if event.is_action_pressed("ui_cancel") and is_open:
		_on_quit_pressed()