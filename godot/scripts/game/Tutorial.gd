extends Control

# Tutorial steps data
var tutorial_steps = [
	{
		"title": "Welcome",
		"description": "Welcome to Merchant Tails! In this tutorial, you'll learn the basics of trading and investment through our fantasy merchant simulation.\n\nYour journey begins as an apprentice merchant in the capital city of Elm. You'll start with 1000 gold and work your way up to becoming a Master Merchant.\n\nThe key to success is simple: Buy low, sell high. But timing and strategy are everything!",
		"hint": "Watch market prices carefully - they change based on supply and demand!"
	},
	{
		"title": "Buying Items",
		"description": "To start trading, you need to buy items from the market.\n\n1. Go to the Market tab\n2. Check current prices for different items\n3. Look for items that seem underpriced\n4. Select an item and choose quantity\n5. Click 'Buy' to purchase\n\nRemember: Different items have different characteristics. Fruits spoil quickly but sell fast. Weapons are expensive but have high profit margins.",
		"hint": "Start with fruits - they're cheap and help you learn the basics!"
	},
	{
		"title": "Setting Prices",
		"description": "Once you have items, they'll appear in your shop inventory.\n\nCustomers will automatically buy items from your shop if:\n- The item is available\n- Your price is reasonable\n- There's demand for it\n\nYour reputation affects how many customers visit your shop. Keep customers happy with fair prices and good availability!",
		"hint": "Price too high = no sales. Price too low = lost profit. Find the sweet spot!"
	},
	{
		"title": "Managing Inventory",
		"description": "You have two storage locations:\n\n• Shop: Items here are available for customers to buy\n• Warehouse: Extra storage for items you're holding\n\nManage your inventory wisely:\n- Keep popular items in the shop\n- Store seasonal items in warehouse\n- Don't let fruits spoil!\n- Balance variety with quantity",
		"hint": "Move items between shop and warehouse based on demand patterns!"
	},
	{
		"title": "Understanding Markets",
		"description": "Market prices fluctuate based on:\n\n• Supply and Demand: Basic economic forces\n• Seasons: Some items are seasonal\n• Weather: Affects certain goods\n• Events: Special occasions change demand\n\nLearn to recognize patterns:\n- Rising prices = increasing demand or decreasing supply\n- Falling prices = decreasing demand or increasing supply",
		"hint": "Buy during market dips, sell during peaks!"
	},
	{
		"title": "Making Profit",
		"description": "Your goal is to grow your wealth through smart trading:\n\n1. Buy when prices are low\n2. Hold items until prices rise\n3. Sell for profit\n4. Reinvest earnings\n5. Diversify your inventory\n\nTrack your performance:\n- Total profit made\n- Best trades\n- Current net worth",
		"hint": "Don't put all your gold in one basket - diversify!"
	},
	{
		"title": "Using the Bank",
		"description": "The bank offers a safe way to grow your wealth:\n\n• Deposit excess gold to earn interest\n• 2% annual interest rate\n• Compounds daily\n• Withdraw anytime for opportunities\n\nBanking strategy:\n- Keep some gold for trading\n- Deposit profits for steady growth\n- Build an emergency fund",
		"hint": "Let compound interest work for you while you sleep!"
	},
	{
		"title": "Complete!",
		"description": "Congratulations! You've learned the basics of Merchant Tails!\n\nKey takeaways:\n✓ Buy low, sell high\n✓ Watch market trends\n✓ Manage inventory wisely\n✓ Diversify investments\n✓ Use the bank for passive income\n\nNow you're ready to start your journey as a merchant. Good luck, and may your trades be profitable!",
		"hint": "Remember: Patience and strategy beat luck every time!"
	}
]

# Current step
var current_step = 0

# UI nodes
@onready var title_label = $Container/Title
@onready var step_title = $Container/ContentPanel/VBox/StepTitle
@onready var description = $Container/ContentPanel/VBox/Description
@onready var hint_label = $Container/ContentPanel/VBox/Hint
@onready var skip_button = $Container/ButtonContainer/SkipButton
@onready var previous_button = $Container/ButtonContainer/PreviousButton
@onready var next_button = $Container/ButtonContainer/NextButton
@onready var complete_button = $Container/ButtonContainer/CompleteButton

# Step indicators
@onready var step_indicators = [
	$Container/StepIndicator/Step1,
	$Container/StepIndicator/Step2,
	$Container/StepIndicator/Step3,
	$Container/StepIndicator/Step4,
	$Container/StepIndicator/Step5,
	$Container/StepIndicator/Step6,
	$Container/StepIndicator/Step7,
	$Container/StepIndicator/Step8
]

func _ready():
	_update_display()
	_update_localized_text()

func _update_display():
	# Update content
	var step = tutorial_steps[current_step]
	step_title.text = "Step " + str(current_step + 1) + ": " + tr("TUTORIAL_" + step.title.to_upper().replace(" ", "_"))
	description.text = tr("TUTORIAL_" + step.title.to_upper().replace(" ", "_") + "_DESC")
	hint_label.text = "💡 " + tr("TUTORIAL_" + step.title.to_upper().replace(" ", "_") + "_HINT")
	
	# Fallback to hardcoded text if translation not found
	if description.text.begins_with("TUTORIAL_"):
		description.text = step.description
	if hint_label.text.begins_with("💡 TUTORIAL_"):
		hint_label.text = "💡 " + step.hint
	
	# Update step indicators
	for i in range(step_indicators.size()):
		if i < current_step:
			step_indicators[i].color = Color(0.2, 0.8, 0.2, 1)  # Completed (green)
		elif i == current_step:
			step_indicators[i].color = Color(0.2, 0.6, 1, 1)    # Current (blue)
		else:
			step_indicators[i].color = Color(0.3, 0.3, 0.3, 1)  # Upcoming (gray)
	
	# Update buttons
	previous_button.disabled = (current_step == 0)
	next_button.visible = (current_step < tutorial_steps.size() - 1)
	complete_button.visible = (current_step == tutorial_steps.size() - 1)
	
	# Update title for last step
	if current_step == tutorial_steps.size() - 1:
		title_label.text = tr("TUTORIAL_COMPLETE")

func _update_localized_text():
	# Update button text
	skip_button.text = tr("TUTORIAL_SKIP")
	previous_button.text = tr("TUTORIAL_PREVIOUS")
	next_button.text = tr("TUTORIAL_NEXT")
	complete_button.text = tr("TUTORIAL_COMPLETE")

func _on_skip_pressed():
	_show_confirm_dialog(
		tr("Are you sure you want to skip the tutorial?"),
		_skip_tutorial
	)

func _skip_tutorial():
	# Mark tutorial as completed
	_save_tutorial_progress(true)
	# Go to main game
	get_tree().change_scene_to_file("res://scenes/game/GameMain.tscn")

func _on_previous_pressed():
	if current_step > 0:
		current_step -= 1
		_update_display()

func _on_next_pressed():
	if current_step < tutorial_steps.size() - 1:
		current_step += 1
		_update_display()
		_save_tutorial_progress(false)

func _on_complete_pressed():
	# Mark tutorial as completed
	_save_tutorial_progress(true)
	# Go to main game
	get_tree().change_scene_to_file("res://scenes/game/GameMain.tscn")

func _save_tutorial_progress(completed: bool):
	var config = ConfigFile.new()
	config.load("user://settings.cfg")
	config.set_value("tutorial", "completed", completed)
	config.set_value("tutorial", "last_step", current_step)
	config.save("user://settings.cfg")

func _show_confirm_dialog(message: String, callback: Callable):
	# Create confirmation dialog
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

func _input(event):
	# Keyboard shortcuts
	if event.is_action_pressed("ui_left"):
		_on_previous_pressed()
	elif event.is_action_pressed("ui_right"):
		if next_button.visible:
			_on_next_pressed()
	elif event.is_action_pressed("ui_accept"):
		if complete_button.visible:
			_on_complete_pressed()
		elif next_button.visible:
			_on_next_pressed()
	elif event.is_action_pressed("ui_cancel"):
		_on_skip_pressed()