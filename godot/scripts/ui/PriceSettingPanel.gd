extends BasePanel

## Price Setting Panel - handles item pricing UI
class_name PriceSettingPanel

# Signals
signal price_updated(result: Dictionary)
signal bulk_price_updated(results: Array)
signal strategy_applied(results: Array)
signal rules_applied(results: Array)

# UI References
@onready var item_list: ItemList = $VBoxContainer/ContentArea/HSplitContainer/ItemsPanel/ItemList
@onready var price_slider: HSlider = $VBoxContainer/ContentArea/HSplitContainer/PricingPanel/PriceSlider
@onready var price_input: SpinBox = $VBoxContainer/ContentArea/HSplitContainer/PricingPanel/PriceInput
@onready var current_price_label: Label = $VBoxContainer/ContentArea/HSplitContainer/PricingPanel/CurrentPriceLabel
@onready var market_price_label: Label = $VBoxContainer/ContentArea/HSplitContainer/PricingPanel/MarketPriceLabel
@onready var profit_margin_label: Label = $VBoxContainer/ContentArea/HSplitContainer/PricingPanel/ProfitMarginLabel
@onready var expected_sales_label: Label = $VBoxContainer/ContentArea/HSplitContainer/PricingPanel/ExpectedSalesLabel

# Strategy controls
@onready var strategy_list: OptionButton = $VBoxContainer/ContentArea/StrategyPanel/StrategyList
@onready var apply_strategy_btn: Button = $VBoxContainer/ContentArea/StrategyPanel/ApplyStrategyButton
@onready var strategy_description: Label = $VBoxContainer/ContentArea/StrategyPanel/StrategyDescription

# Analytics display
@onready var analytics_chart: Control = $VBoxContainer/ContentArea/AnalyticsPanel/Chart
@onready var optimal_price_label: Label = $VBoxContainer/ContentArea/AnalyticsPanel/OptimalPriceLabel
@onready var elasticity_label: Label = $VBoxContainer/ContentArea/AnalyticsPanel/ElasticityLabel
@onready var revenue_label: Label = $VBoxContainer/ContentArea/AnalyticsPanel/RevenueLabel

# Rules panel
@onready var rules_list: ItemList = $VBoxContainer/ContentArea/RulesPanel/RulesList
@onready var apply_rules_btn: Button = $VBoxContainer/ContentArea/RulesPanel/ApplyRulesButton
@onready var toggle_rule_btn: Button = $VBoxContainer/ContentArea/RulesPanel/ToggleRuleButton

# Filter controls
@onready var category_filter: OptionButton = $VBoxContainer/HeaderArea/FilterBar/CategoryFilter
@onready var demand_filter: OptionButton = $VBoxContainer/HeaderArea/FilterBar/DemandFilter
@onready var sort_option: OptionButton = $VBoxContainer/HeaderArea/FilterBar/SortOption

# Bulk actions
@onready var select_all_check: CheckBox = $VBoxContainer/FooterArea/BulkActions/SelectAllCheck
@onready var bulk_strategy: OptionButton = $VBoxContainer/FooterArea/BulkActions/BulkStrategyOption
@onready var bulk_apply_btn: Button = $VBoxContainer/FooterArea/BulkActions/BulkApplyButton

# Data
var game_bridge: GameBridge
var price_items: Array = []
var selected_items: Array = []
var strategies: Array = []
var rules: Array = []
var analytics_data: Dictionary = {}

# Visual settings
const PROFIT_COLOR = Color(0.3, 0.8, 0.3)
const LOSS_COLOR = Color(0.8, 0.3, 0.3)
const MARKET_COLOR = Color(0.5, 0.5, 0.8)
const RECOMMENDED_COLOR = Color(0.8, 0.8, 0.3)

func _ready():
	super._ready()
	_setup_ui()
	_connect_signals()
	
	# Get game bridge reference
	game_bridge = get_node("/root/GameBridge")
	if game_bridge:
		refresh_data()

func _setup_ui():
	# Setup category filter
	category_filter.clear()
	category_filter.add_item("All Categories")
	category_filter.add_item("Fruit")
	category_filter.add_item("Potion")
	category_filter.add_item("Weapon")
	category_filter.add_item("Accessory")
	category_filter.add_item("Magic Book")
	category_filter.add_item("Gem")
	
	# Setup demand filter
	demand_filter.clear()
	demand_filter.add_item("All Demand")
	demand_filter.add_item("Very High")
	demand_filter.add_item("High")
	demand_filter.add_item("Normal")
	demand_filter.add_item("Low")
	demand_filter.add_item("Very Low")
	
	# Setup sort options
	sort_option.clear()
	sort_option.add_item("Name")
	sort_option.add_item("Price")
	sort_option.add_item("Profit Margin")
	sort_option.add_item("Demand")
	sort_option.add_item("Quantity")
	
	# Setup price controls
	price_slider.min_value = 0
	price_slider.max_value = 1000
	price_slider.step = 0.01
	
	price_input.min_value = 0
	price_input.max_value = 10000
	price_input.step = 0.01

func _connect_signals():
	# Connect filter signals
	category_filter.item_selected.connect(_on_filter_changed)
	demand_filter.item_selected.connect(_on_filter_changed)
	sort_option.item_selected.connect(_on_filter_changed)
	
	# Connect item selection
	item_list.item_selected.connect(_on_item_selected)
	item_list.multi_selected.connect(_on_item_multi_selected)
	
	# Connect price controls
	price_slider.value_changed.connect(_on_price_slider_changed)
	price_input.value_changed.connect(_on_price_input_changed)
	
	# Connect strategy controls
	strategy_list.item_selected.connect(_on_strategy_selected)
	apply_strategy_btn.pressed.connect(_on_apply_strategy)
	
	# Connect rules controls
	rules_list.item_selected.connect(_on_rule_selected)
	apply_rules_btn.pressed.connect(_on_apply_rules)
	toggle_rule_btn.pressed.connect(_on_toggle_rule)
	
	# Connect bulk actions
	select_all_check.toggled.connect(_on_select_all_toggled)
	bulk_apply_btn.pressed.connect(_on_bulk_apply)

func refresh_data():
	if not game_bridge:
		return
	
	# Get filter settings
	var filter = _get_current_filter()
	
	# Get price setting items
	price_items = game_bridge.get_price_setting_items(filter)
	
	# Get strategies
	strategies = game_bridge.get_pricing_strategies()
	
	# Get rules
	rules = game_bridge.get_pricing_rules()
	
	# Update UI
	_update_item_list()
	_update_strategy_list()
	_update_rules_list()

func _update_item_list():
	item_list.clear()
	
	for item in price_items:
		var text = "%s - %.2f gold (x%d)" % [
			item.name,
			item.current_price,
			item.quantity
		]
		
		item_list.add_item(text)
		var idx = item_list.get_item_count() - 1
		item_list.set_item_metadata(idx, item)
		
		# Add icon if available
		if item.has("icon") and item.icon != "":
			var icon = load(item.icon)
			if icon:
				item_list.set_item_icon(idx, icon)
		
		# Color code by profit margin
		if item.profit_margin > 20:
			item_list.set_item_custom_fg_color(idx, PROFIT_COLOR)
		elif item.profit_margin < 5:
			item_list.set_item_custom_fg_color(idx, LOSS_COLOR)
		
		# Add demand indicator
		var demand_icon = _get_demand_icon(item.demand_level)
		item_list.set_item_text(idx, text + " " + demand_icon)
		
		# Add tooltip
		var tooltip = _generate_item_tooltip(item)
		item_list.set_item_tooltip(idx, tooltip)

func _update_strategy_list():
	strategy_list.clear()
	strategy_list.add_item("Manual Pricing")
	
	for strategy in strategies:
		strategy_list.add_item(strategy.name)
		var idx = strategy_list.get_item_count() - 1
		strategy_list.set_item_metadata(idx, strategy)

func _update_rules_list():
	rules_list.clear()
	
	for rule in rules:
		var status = "✓" if rule.enabled else "✗"
		var text = "%s %s - %s" % [status, rule.name, rule.condition]
		
		rules_list.add_item(text)
		var idx = rules_list.get_item_count() - 1
		rules_list.set_item_metadata(idx, rule)
		
		# Color by enabled status
		if rule.enabled:
			rules_list.set_item_custom_fg_color(idx, Color.GREEN)
		else:
			rules_list.set_item_custom_fg_color(idx, Color.GRAY)

func _on_item_selected(index: int):
	var item = item_list.get_item_metadata(index)
	selected_items = [item]
	_update_price_controls(item)
	_load_analytics(item.item_id)

func _on_item_multi_selected(index: int, selected: bool):
	var item = item_list.get_item_metadata(index)
	
	if selected:
		if not item in selected_items:
			selected_items.append(item)
	else:
		selected_items.erase(item)
	
	# Update UI based on selection
	if selected_items.size() == 1:
		_update_price_controls(selected_items[0])
	else:
		_update_bulk_controls()

func _update_price_controls(item: Dictionary):
	# Update price slider bounds
	price_slider.min_value = item.min_price
	price_slider.max_value = item.max_price
	price_slider.value = item.current_price
	
	# Update price input
	price_input.value = item.current_price
	
	# Update labels
	current_price_label.text = "Current: %.2f gold" % item.current_price
	market_price_label.text = "Market: %.2f gold" % item.market_price
	profit_margin_label.text = "Margin: %.1f%%" % item.profit_margin
	expected_sales_label.text = "Expected Sales: %d/day" % item.expected_sales
	
	# Show recommended price
	var recommended_text = "Recommended: %.2f gold" % item.recommended_price
	var price_comparison = ""
	
	if item.current_price < item.recommended_price * 0.9:
		price_comparison = " (↑ Increase)"
		profit_margin_label.modulate = Color.YELLOW
	elif item.current_price > item.recommended_price * 1.1:
		price_comparison = " (↓ Decrease)"
		profit_margin_label.modulate = Color.ORANGE
	else:
		price_comparison = " (✓ Optimal)"
		profit_margin_label.modulate = Color.GREEN
	
	profit_margin_label.text += price_comparison
	
	# Update competitor comparison
	if item.competitor_price > 0:
		var comp_diff = ((item.current_price - item.competitor_price) / item.competitor_price) * 100
		var comp_text = "vs Competition: "
		if comp_diff > 5:
			comp_text += "%.1f%% higher" % comp_diff
		elif comp_diff < -5:
			comp_text += "%.1f%% lower" % abs(comp_diff)
		else:
			comp_text += "Competitive"
		market_price_label.text += "\n" + comp_text

func _update_bulk_controls():
	# Enable bulk actions
	bulk_strategy.disabled = false
	bulk_apply_btn.disabled = false
	bulk_apply_btn.text = "Apply to %d items" % selected_items.size()

func _on_price_slider_changed(value: float):
	price_input.set_value_no_signal(value)
	_preview_price_change(value)

func _on_price_input_changed(value: float):
	price_slider.set_value_no_signal(value)
	_preview_price_change(value)

func _preview_price_change(new_price: float):
	if selected_items.is_empty():
		return
	
	var item = selected_items[0]
	
	# Calculate new profit margin
	var new_margin = ((new_price - item.purchase_price) / item.purchase_price) * 100
	profit_margin_label.text = "Margin: %.1f%% → %.1f%%" % [item.profit_margin, new_margin]
	
	# Estimate new sales based on elasticity
	var price_change = (new_price - item.current_price) / item.current_price
	var quantity_change = -price_change * item.elasticity
	var new_sales = max(0, int(item.expected_sales * (1 + quantity_change)))
	expected_sales_label.text = "Expected Sales: %d → %d/day" % [item.expected_sales, new_sales]
	
	# Calculate expected revenue
	var old_revenue = item.current_price * item.expected_sales
	var new_revenue = new_price * new_sales
	var revenue_change = new_revenue - old_revenue
	
	var revenue_text = "Revenue: %.2f" % new_revenue
	if revenue_change > 0:
		revenue_text += " (+%.2f)" % revenue_change
		revenue_label.modulate = Color.GREEN
	elif revenue_change < 0:
		revenue_text += " (%.2f)" % revenue_change
		revenue_label.modulate = Color.RED
	else:
		revenue_label.modulate = Color.WHITE
	
	revenue_label.text = revenue_text

func _on_strategy_selected(index: int):
	if index == 0:
		strategy_description.text = "Set prices manually for each item"
		return
	
	var strategy = strategy_list.get_item_metadata(index)
	if strategy:
		strategy_description.text = strategy.description
		strategy_description.text += "\nTarget Profit: %.0f%%" % strategy.target_profit
		strategy_description.text += "\nRisk Level: %s" % strategy.risk_level.capitalize()

func _on_apply_strategy():
	if selected_items.is_empty():
		_show_error_message("No items selected")
		return
	
	var strategy_idx = strategy_list.selected
	if strategy_idx <= 0:
		_show_error_message("Please select a strategy")
		return
	
	var strategy = strategy_list.get_item_metadata(strategy_idx)
	var item_ids = []
	for item in selected_items:
		item_ids.append(item.item_id)
	
	var results = game_bridge.apply_pricing_strategy(strategy.id, item_ids)
	_handle_strategy_results(results)

func _handle_strategy_results(results: Array):
	var successful = 0
	var failed = 0
	
	for result in results:
		if result.success:
			successful += 1
		else:
			failed += 1
	
	if successful > 0:
		_show_success_message("Updated %d prices successfully" % successful)
		refresh_data()
		strategy_applied.emit(results)
	
	if failed > 0:
		_show_error_message("%d price updates failed" % failed)

func _on_rule_selected(index: int):
	var rule = rules_list.get_item_metadata(index)
	toggle_rule_btn.disabled = false
	toggle_rule_btn.text = "Disable" if rule.enabled else "Enable"

func _on_toggle_rule():
	var selected = rules_list.get_selected_items()
	if selected.is_empty():
		return
	
	var rule = rules_list.get_item_metadata(selected[0])
	var enabled = not rule.enabled
	
	if game_bridge.toggle_pricing_rule(rule.id, enabled):
		_show_success_message("Rule %s" % ("enabled" if enabled else "disabled"))
		refresh_data()

func _on_apply_rules():
	var results = game_bridge.apply_pricing_rules()
	
	if results and results.size() > 0:
		_show_success_message("Applied rules to %d items" % results.size())
		refresh_data()
		rules_applied.emit(results)
	else:
		_show_info_message("No rules were applied")

func _on_select_all_toggled(pressed: bool):
	selected_items.clear()
	
	if pressed:
		for i in range(item_list.get_item_count()):
			item_list.select(i, false)
			var item = item_list.get_item_metadata(i)
			selected_items.append(item)
	else:
		for i in range(item_list.get_item_count()):
			item_list.deselect(i)
	
	_update_bulk_controls()

func _on_bulk_apply():
	if selected_items.is_empty():
		_show_error_message("No items selected")
		return
	
	var strategy_idx = bulk_strategy.selected
	if strategy_idx < 0:
		_show_error_message("Please select a bulk strategy")
		return
	
	var updates = []
	for item in selected_items:
		var update = {
			"item_id": item.item_id,
			"new_price": price_input.value,
			"strategy": bulk_strategy.get_item_text(strategy_idx).to_lower()
		}
		updates.append(update)
	
	var request = {
		"updates": updates,
		"strategy": bulk_strategy.get_item_text(strategy_idx).to_lower()
	}
	
	var results = game_bridge.bulk_update_prices(request)
	_handle_bulk_results(results)

func _handle_bulk_results(results: Array):
	var successful = 0
	var total_revenue = 0.0
	
	for result in results:
		if result.success:
			successful += 1
			total_revenue += result.expected_revenue
	
	if successful > 0:
		_show_success_message("Updated %d prices\nExpected revenue: %.2f gold/day" % [successful, total_revenue])
		refresh_data()
		bulk_price_updated.emit(results)

func _load_analytics(item_id: String):
	analytics_data = game_bridge.get_price_analytics(item_id)
	
	if analytics_data:
		optimal_price_label.text = "Optimal Price: %.2f gold" % analytics_data.optimal_price
		elasticity_label.text = "Price Elasticity: %.2f" % analytics_data.price_elasticity
		
		# Draw chart
		_draw_analytics_chart()

func _draw_analytics_chart():
	# Simple revenue/profit chart visualization
	if not analytics_data or not analytics_data.has("revenue_history"):
		return
	
	# This would draw a chart showing historical revenue and profit
	# For now, just show the latest values
	if analytics_data.revenue_history.size() > 0:
		var latest_revenue = analytics_data.revenue_history[-1]
		revenue_label.text = "Latest Revenue: %.2f gold" % latest_revenue

func _on_filter_changed(_index = null):
	refresh_data()

func _get_current_filter() -> String:
	var cat_idx = category_filter.selected
	if cat_idx > 0:
		return category_filter.get_item_text(cat_idx).to_lower().replace(" ", "_")
	return "all"

func _get_demand_icon(demand_level: String) -> String:
	match demand_level:
		"very_high": return "⬆⬆"
		"high": return "⬆"
		"normal": return "➡"
		"low": return "⬇"
		"very_low": return "⬇⬇"
		_: return ""

func _generate_item_tooltip(item: Dictionary) -> String:
	var tooltip = "%s\n" % item.name
	tooltip += "Current Price: %.2f gold\n" % item.current_price
	tooltip += "Purchase Price: %.2f gold\n" % item.purchase_price
	tooltip += "Market Price: %.2f gold\n" % item.market_price
	
	if item.competitor_price > 0:
		tooltip += "Competitor Price: %.2f gold\n" % item.competitor_price
	
	tooltip += "Recommended: %.2f gold\n" % item.recommended_price
	tooltip += "Profit Margin: %.1f%%\n" % item.profit_margin
	tooltip += "Demand: %s\n" % item.demand_level.capitalize()
	tooltip += "Price Elasticity: %.2f\n" % item.elasticity
	tooltip += "Expected Sales: %d/day\n" % item.expected_sales
	tooltip += "Risk Level: %s" % item.risk_level.capitalize()
	
	return tooltip

func _show_success_message(message: String):
	print("[SUCCESS] " + message)
	# TODO: Show in UI notification system

func _show_error_message(message: String):
	print("[ERROR] " + message)
	# TODO: Show in UI notification system

func _show_info_message(message: String):
	print("[INFO] " + message)
	# TODO: Show in UI notification system

# Support for keyboard shortcuts
func _unhandled_key_input(event: InputEvent):
	if not visible:
		return
	
	if event is InputEventKey and event.pressed:
		match event.keycode:
			KEY_ENTER:
				_apply_price_change()
			KEY_ESCAPE:
				_cancel_price_change()
			KEY_A:
				if event.ctrl_pressed:
					select_all_check.button_pressed = true
					_on_select_all_toggled(true)

func _apply_price_change():
	if selected_items.is_empty():
		return
	
	var item = selected_items[0]
	var request = {
		"item_id": item.item_id,
		"new_price": price_input.value,
		"strategy": "manual"
	}
	
	var result = game_bridge.update_item_price(request)
	if result.success:
		_show_success_message("Price updated to %.2f gold" % result.new_price)
		refresh_data()
		price_updated.emit(result)

func _cancel_price_change():
	if not selected_items.is_empty():
		var item = selected_items[0]
		price_slider.value = item.current_price
		price_input.value = item.current_price