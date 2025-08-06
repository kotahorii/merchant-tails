extends Control

func _ready():
	print("Merchant Tails - Main Menu Loaded")
	
func _on_start_button_pressed():
	print("Starting game...")
	# TODO: Load game scene
	
func _on_settings_button_pressed():
	print("Opening settings...")
	# TODO: Open settings menu
	
func _on_quit_button_pressed():
	get_tree().quit()