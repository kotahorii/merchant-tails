extends Control

# Settings data
var settings = {
	"difficulty": "Normal",
	"language": "en",
	"auto_save": true,
	"resolution": "1920x1080",
	"fullscreen": false,
	"vsync": true,
	"master_volume": 80,
	"music_volume": 70,
	"sfx_volume": 90
}

# UI nodes
@onready var difficulty_option = $Container/TabContainer/Gameplay/VBox/DifficultyOption
@onready var language_option = $Container/TabContainer/Gameplay/VBox/LanguageOption
@onready var auto_save_check = $Container/TabContainer/Gameplay/VBox/AutoSaveCheck
@onready var resolution_option = $Container/TabContainer/Graphics/VBox/ResolutionOption
@onready var fullscreen_check = $Container/TabContainer/Graphics/VBox/FullscreenCheck
@onready var vsync_check = $Container/TabContainer/Graphics/VBox/VSyncCheck
@onready var master_slider = $Container/TabContainer/Audio/VBox/MasterSlider
@onready var master_value = $Container/TabContainer/Audio/VBox/MasterValue
@onready var music_slider = $Container/TabContainer/Audio/VBox/MusicSlider
@onready var music_value = $Container/TabContainer/Audio/VBox/MusicValue
@onready var sfx_slider = $Container/TabContainer/Audio/VBox/SFXSlider
@onready var sfx_value = $Container/TabContainer/Audio/VBox/SFXValue

# Available options
var difficulties = ["Easy", "Normal", "Hard"]
var languages = {"English": "en", "日本語": "ja"}
var resolutions = ["1280x720", "1920x1080", "2560x1440", "3840x2160"]

func _ready():
	_load_settings()
	_setup_ui()
	_apply_settings()
	_update_localized_text()

func _setup_ui():
	# Setup difficulty options
	difficulty_option.clear()
	for diff in difficulties:
		difficulty_option.add_item(tr("DIFFICULTY_" + diff.to_upper()))
	
	# Setup language options
	language_option.clear()
	for lang_name in languages.keys():
		language_option.add_item(lang_name)
	
	# Setup resolution options
	resolution_option.clear()
	for res in resolutions:
		resolution_option.add_item(res)
	
	# Set current values
	_update_ui_from_settings()

func _update_ui_from_settings():
	# Difficulty
	var diff_index = difficulties.find(settings.difficulty)
	if diff_index != -1:
		difficulty_option.selected = diff_index
	
	# Language
	var lang_index = 0
	var i = 0
	for lang_code in languages.values():
		if lang_code == settings.language:
			lang_index = i
			break
		i += 1
	language_option.selected = lang_index
	
	# Auto save
	auto_save_check.button_pressed = settings.auto_save
	
	# Resolution
	var res_index = resolutions.find(settings.resolution)
	if res_index != -1:
		resolution_option.selected = res_index
	
	# Fullscreen
	fullscreen_check.button_pressed = settings.fullscreen
	
	# VSync
	vsync_check.button_pressed = settings.vsync
	
	# Volume sliders
	master_slider.value = settings.master_volume
	music_slider.value = settings.music_volume
	sfx_slider.value = settings.sfx_volume
	
	# Update volume labels
	master_value.text = str(settings.master_volume) + "%"
	music_value.text = str(settings.music_volume) + "%"
	sfx_value.text = str(settings.sfx_volume) + "%"

func _update_localized_text():
	# Update all localized strings
	$Container/Title.text = tr("SETTINGS_TITLE")
	$Container/TabContainer.set_tab_title(0, tr("SETTINGS_GAMEPLAY"))
	$Container/TabContainer.set_tab_title(1, tr("SETTINGS_GRAPHICS"))
	$Container/TabContainer.set_tab_title(2, tr("SETTINGS_AUDIO"))
	
	# Update labels
	$Container/TabContainer/Gameplay/VBox/DifficultyLabel.text = tr("SETTINGS_DIFFICULTY")
	$Container/TabContainer/Gameplay/VBox/LanguageLabel.text = tr("SETTINGS_LANGUAGE")
	$Container/TabContainer/Graphics/VBox/ResolutionLabel.text = tr("SETTINGS_RESOLUTION")
	$Container/TabContainer/Graphics/VBox/FullscreenCheck.text = tr("SETTINGS_FULLSCREEN")
	$Container/TabContainer/Audio/VBox/MasterLabel.text = tr("SETTINGS_MASTER_VOLUME")
	$Container/TabContainer/Audio/VBox/MusicLabel.text = tr("SETTINGS_MUSIC_VOLUME")
	$Container/TabContainer/Audio/VBox/SFXLabel.text = tr("SETTINGS_SFX_VOLUME")
	
	# Update buttons
	$Container/ButtonContainer/ApplyButton.text = tr("SETTINGS_APPLY")
	$Container/ButtonContainer/BackButton.text = tr("SETTINGS_BACK")

func _load_settings():
	var config = ConfigFile.new()
	var error = config.load("user://settings.cfg")
	
	if error == OK:
		# Load each setting
		settings.difficulty = config.get_value("gameplay", "difficulty", "Normal")
		settings.language = config.get_value("gameplay", "language", "en")
		settings.auto_save = config.get_value("gameplay", "auto_save", true)
		settings.resolution = config.get_value("graphics", "resolution", "1920x1080")
		settings.fullscreen = config.get_value("graphics", "fullscreen", false)
		settings.vsync = config.get_value("graphics", "vsync", true)
		settings.master_volume = config.get_value("audio", "master_volume", 80)
		settings.music_volume = config.get_value("audio", "music_volume", 70)
		settings.sfx_volume = config.get_value("audio", "sfx_volume", 90)

func _save_settings():
	var config = ConfigFile.new()
	
	# Save each setting
	config.set_value("gameplay", "difficulty", settings.difficulty)
	config.set_value("gameplay", "language", settings.language)
	config.set_value("gameplay", "auto_save", settings.auto_save)
	config.set_value("graphics", "resolution", settings.resolution)
	config.set_value("graphics", "fullscreen", settings.fullscreen)
	config.set_value("graphics", "vsync", settings.vsync)
	config.set_value("audio", "master_volume", settings.master_volume)
	config.set_value("audio", "music_volume", settings.music_volume)
	config.set_value("audio", "sfx_volume", settings.sfx_volume)
	
	# Save to file
	config.save("user://settings.cfg")

func _apply_settings():
	# Apply language
	TranslationServer.set_locale(settings.language)
	
	# Apply fullscreen
	if settings.fullscreen:
		DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)
	else:
		DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_WINDOWED)
	
	# Apply resolution (only in windowed mode)
	if not settings.fullscreen:
		var res_parts = settings.resolution.split("x")
		if res_parts.size() == 2:
			var width = int(res_parts[0])
			var height = int(res_parts[1])
			DisplayServer.window_set_size(Vector2i(width, height))
			# Center window
			var screen_size = DisplayServer.screen_get_size()
			var window_pos = (screen_size - Vector2i(width, height)) / 2
			DisplayServer.window_set_position(window_pos)
	
	# Apply VSync
	if settings.vsync:
		DisplayServer.window_set_vsync_mode(DisplayServer.VSYNC_ENABLED)
	else:
		DisplayServer.window_set_vsync_mode(DisplayServer.VSYNC_DISABLED)
	
	# Apply audio volumes
	AudioServer.set_bus_volume_db(AudioServer.get_bus_index("Master"), linear_to_db(settings.master_volume / 100.0))
	AudioServer.set_bus_volume_db(AudioServer.get_bus_index("Music"), linear_to_db(settings.music_volume / 100.0))
	AudioServer.set_bus_volume_db(AudioServer.get_bus_index("SFX"), linear_to_db(settings.sfx_volume / 100.0))

# UI callbacks
func _on_difficulty_selected(index: int):
	settings.difficulty = difficulties[index]

func _on_language_selected(index: int):
	var lang_keys = languages.keys()
	settings.language = languages[lang_keys[index]]
	TranslationServer.set_locale(settings.language)
	_update_localized_text()

func _on_auto_save_toggled(pressed: bool):
	settings.auto_save = pressed

func _on_resolution_selected(index: int):
	settings.resolution = resolutions[index]

func _on_fullscreen_toggled(pressed: bool):
	settings.fullscreen = pressed

func _on_vsync_toggled(pressed: bool):
	settings.vsync = pressed

func _on_master_volume_changed(value: float):
	settings.master_volume = int(value)
	master_value.text = str(int(value)) + "%"
	AudioServer.set_bus_volume_db(AudioServer.get_bus_index("Master"), linear_to_db(value / 100.0))

func _on_music_volume_changed(value: float):
	settings.music_volume = int(value)
	music_value.text = str(int(value)) + "%"
	AudioServer.set_bus_volume_db(AudioServer.get_bus_index("Music"), linear_to_db(value / 100.0))

func _on_sfx_volume_changed(value: float):
	settings.sfx_volume = int(value)
	sfx_value.text = str(int(value)) + "%"
	AudioServer.set_bus_volume_db(AudioServer.get_bus_index("SFX"), linear_to_db(value / 100.0))

func _on_apply_pressed():
	_save_settings()
	_apply_settings()
	_show_notification(tr("Settings saved!"))

func _on_back_pressed():
	# Return to main menu
	get_tree().change_scene_to_file("res://scenes/MainMenu.tscn")

func _show_notification(message: String):
	# Create a simple notification
	var label = Label.new()
	label.text = message
	label.add_theme_font_size_override("font_size", 20)
	label.position = Vector2(get_viewport().size.x / 2 - 100, 50)
	add_child(label)
	
	# Fade out and remove after 2 seconds
	var tween = create_tween()
	tween.tween_property(label, "modulate:a", 0.0, 2.0)
	tween.tween_callback(label.queue_free)