extends Control

# Nodes
@onready var new_game_button = $VBoxContainer/NewGameButton
@onready var continue_button = $VBoxContainer/ContinueButton
@onready var tutorial_button = $VBoxContainer/TutorialButton
@onready var settings_button = $VBoxContainer/SettingsButton
@onready var quit_button = $VBoxContainer/QuitButton
@onready var title_label = $VBoxContainer/Title
@onready var language_selector = $LanguageSelector

# Available languages
var languages = {
	"English": "en",
	"日本語": "ja"
}

func _ready():
	# Setup language selector
	_setup_language_selector()
	
	# Update UI with current language
	_update_localized_text()
	
	# Check if there's a saved game
	_check_saved_game()
	
	# Connect to localization changed signal
	if not TranslationServer.get_singleton().property_list_changed.is_connected(_on_localization_changed):
		TranslationServer.get_singleton().property_list_changed.connect(_on_localization_changed)

func _setup_language_selector():
	language_selector.clear()
	var current_locale = TranslationServer.get_locale()
	var selected_index = 0
	var index = 0
	
	for lang_name in languages.keys():
		language_selector.add_item(lang_name)
		if languages[lang_name] == current_locale.substr(0, 2):
			selected_index = index
		index += 1
	
	language_selector.selected = selected_index

func _update_localized_text():
	# Update all text with translations
	title_label.text = tr("TITLE_GAME")
	new_game_button.text = tr("MENU_NEW_GAME")
	continue_button.text = tr("MENU_CONTINUE")
	tutorial_button.text = tr("MENU_TUTORIAL")
	settings_button.text = tr("MENU_SETTINGS")
	quit_button.text = tr("MENU_QUIT")

func _check_saved_game():
	# Check if save file exists
	var save_path = "user://savegame.dat"
	if FileAccess.file_exists(save_path):
		continue_button.disabled = false
	else:
		continue_button.disabled = true

func _on_new_game_pressed():
	# Check if there's an existing save
	if not continue_button.disabled:
		_show_confirm_dialog(
			tr("MSG_CONFIRM_NEW_GAME"),
			_start_new_game
		)
	else:
		_start_new_game()

func _start_new_game():
	print("Starting new game...")
	# Load the game scene
	get_tree().change_scene_to_file("res://scenes/game/GameMain.tscn")

func _on_continue_pressed():
	if not continue_button.disabled:
		print("Loading saved game...")
		# Load the game scene with saved data
		get_tree().change_scene_to_file("res://scenes/game/GameMain.tscn")

func _on_tutorial_pressed():
	print("Starting tutorial...")
	# Load tutorial scene
	get_tree().change_scene_to_file("res://scenes/game/Tutorial.tscn")

func _on_settings_pressed():
	print("Opening settings...")
	# Load settings scene
	get_tree().change_scene_to_file("res://scenes/ui/Settings.tscn")

func _on_quit_pressed():
	_show_confirm_dialog(
		tr("MSG_CONFIRM_QUIT"),
		_quit_game
	)

func _quit_game():
	get_tree().quit()

func _on_language_selected(index: int):
	var selected_language = language_selector.get_item_text(index)
	if selected_language in languages:
		var locale = languages[selected_language]
		TranslationServer.set_locale(locale)
		_update_localized_text()
		
		# Save language preference
		_save_language_preference(locale)

func _save_language_preference(locale: String):
	var config = ConfigFile.new()
	config.set_value("settings", "language", locale)
	config.save("user://settings.cfg")

func _load_language_preference():
	var config = ConfigFile.new()
	var error = config.load("user://settings.cfg")
	if error == OK:
		var locale = config.get_value("settings", "language", "en")
		TranslationServer.set_locale(locale)

func _on_localization_changed():
	_update_localized_text()

func _show_confirm_dialog(message: String, callback: Callable):
	# Create a simple confirmation dialog
	var dialog = AcceptDialog.new()
	dialog.dialog_text = message
	dialog.add_button(tr("MSG_NO"), true, "cancel")
	dialog.get_ok_button().text = tr("MSG_YES")
	
	add_child(dialog)
	dialog.popup_centered(Vector2(400, 150))
	
	var result = await dialog.confirmed
	dialog.queue_free()
	
	if result:
		callback.call()

func _notification(what):
	if what == NOTIFICATION_WM_CLOSE_REQUEST:
		_on_quit_pressed()