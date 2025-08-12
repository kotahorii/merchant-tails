# Merchant Tails API Specification

## Overview
This document describes the GDExtension API interface between the Godot frontend and Go backend for Merchant Tails.

## Version
- API Version: 1.0.0
- Last Updated: 2025-08-12

## Base Architecture
The game uses GDExtension to bind Go functions to Godot, allowing the game logic to run in Go while the UI is handled by Godot.

## Core APIs

### 1. Game Management API

#### InitializeGame
Initializes the game systems.
```gdscript
func initialize_game() -> bool
```
**Returns:** `true` if initialization successful

#### StartNewGame
Starts a new game with the given player name.
```gdscript
func start_new_game(player_name: String) -> Dictionary
```
**Parameters:**
- `player_name`: The player's chosen name

**Returns:** Dictionary containing:
- `success`: bool
- `message`: String
- `game_id`: String

#### SaveGame
Saves the current game state.
```gdscript
func save_game(slot: int) -> Dictionary
```
**Parameters:**
- `slot`: Save slot number (0-9)

**Returns:** Dictionary containing:
- `success`: bool
- `message`: String
- `timestamp`: String

#### LoadGame
Loads a saved game.
```gdscript
func load_game(slot: int) -> Dictionary
```
**Parameters:**
- `slot`: Save slot number (0-9)

**Returns:** Dictionary containing:
- `success`: bool
- `message`: String
- `game_state`: Dictionary

### 2. Player State API

#### GetPlayerInfo
Gets current player information.
```gdscript
func get_player_info() -> Dictionary
```
**Returns:** Dictionary containing:
- `name`: String
- `gold`: int
- `rank`: String
- `reputation`: float
- `day`: int
- `season`: String

#### SetPlayerGold
Updates player's gold amount.
```gdscript
func set_player_gold(amount: int) -> void
```
**Parameters:**
- `amount`: New gold amount

### 3. Market API

#### GetMarketPrices
Gets current market prices for all items.
```gdscript
func get_market_prices() -> Dictionary
```
**Returns:** Dictionary with item IDs as keys and prices as values

#### GetItemPrice
Gets the current price for a specific item.
```gdscript
func get_item_price(item_id: String) -> int
```
**Parameters:**
- `item_id`: The item identifier

**Returns:** Current price in gold

#### UpdateMarketPrices
Updates market prices based on time and events.
```gdscript
func update_market_prices() -> void
```

### 4. Trading API

#### BuyItem
Purchases an item from the market.
```gdscript
func buy_item(item_id: String, quantity: int) -> Dictionary
```
**Parameters:**
- `item_id`: The item to purchase
- `quantity`: Amount to buy

**Returns:** Dictionary containing:
- `success`: bool
- `message`: String
- `total_cost`: int
- `gold_remaining`: int

#### SellItem
Sells an item from inventory.
```gdscript
func sell_item(item_id: String, quantity: int) -> Dictionary
```
**Parameters:**
- `item_id`: The item to sell
- `quantity`: Amount to sell

**Returns:** Dictionary containing:
- `success`: bool
- `message`: String
- `total_revenue`: int
- `gold_gained`: int

### 5. Inventory API

#### GetInventory
Gets the current inventory status.
```gdscript
func get_inventory() -> Dictionary
```
**Returns:** Dictionary containing:
- `shop`: Dictionary of items in shop
- `warehouse`: Dictionary of items in warehouse
- `shop_capacity`: int
- `warehouse_capacity`: int

#### MoveItem
Moves items between shop and warehouse.
```gdscript
func move_item(item_id: String, quantity: int, from: String, to: String) -> Dictionary
```
**Parameters:**
- `item_id`: The item to move
- `quantity`: Amount to move
- `from`: Source location ("shop" or "warehouse")
- `to`: Destination location ("shop" or "warehouse")

**Returns:** Dictionary containing:
- `success`: bool
- `message`: String

### 6. Bank API

#### GetBankAccount
Gets bank account information.
```gdscript
func get_bank_account() -> Dictionary
```
**Returns:** Dictionary containing:
- `balance`: float
- `interest_rate`: float
- `last_interest_date`: String

#### DepositMoney
Deposits money into the bank.
```gdscript
func deposit_money(amount: float) -> Dictionary
```
**Parameters:**
- `amount`: Amount to deposit

**Returns:** Dictionary containing:
- `success`: bool
- `message`: String
- `new_balance`: float

#### WithdrawMoney
Withdraws money from the bank.
```gdscript
func withdraw_money(amount: float) -> Dictionary
```
**Parameters:**
- `amount`: Amount to withdraw

**Returns:** Dictionary containing:
- `success`: bool
- `message`: String
- `new_balance`: float

### 7. Tutorial API

#### GetTutorialStep
Gets the current tutorial step.
```gdscript
func get_tutorial_step() -> Dictionary
```
**Returns:** Dictionary containing:
- `id`: String
- `title`: String
- `description`: String
- `hint`: String
- `completed`: bool

#### NextTutorialStep
Advances to the next tutorial step.
```gdscript
func next_tutorial_step() -> bool
```
**Returns:** `true` if there are more steps

#### SkipTutorial
Skips the tutorial.
```gdscript
func skip_tutorial() -> void
```

### 8. Weather API

#### GetCurrentWeather
Gets the current weather.
```gdscript
func get_current_weather() -> Dictionary
```
**Returns:** Dictionary containing:
- `type`: String ("Sunny", "Rainy", "Cloudy")
- `description`: String
- `price_effect`: float
- `demand_effect`: float

#### UpdateWeather
Updates the weather (called each day).
```gdscript
func update_weather() -> void
```

### 9. Difficulty API

#### GetDifficulty
Gets the current difficulty setting.
```gdscript
func get_difficulty() -> String
```
**Returns:** "Easy", "Normal", or "Hard"

#### SetDifficulty
Sets the game difficulty.
```gdscript
func set_difficulty(level: String) -> void
```
**Parameters:**
- `level`: "Easy", "Normal", or "Hard"

## Error Codes

| Code | Description |
|------|-------------|
| 0    | Success |
| 1001 | Insufficient gold |
| 1002 | Insufficient inventory space |
| 1003 | Item not found |
| 1004 | Invalid quantity |
| 1005 | Bank account not found |
| 1006 | Invalid save slot |
| 1007 | Save file corrupted |
| 1008 | Network error (reserved for future use) |

## Data Structures

### Item Structure
```gdscript
{
    "id": "apple",
    "name": "Apple",
    "category": "fruit",
    "base_price": 10,
    "current_price": 12,
    "quantity": 5,
    "quality": 1
}
```

### Transaction Structure
```gdscript
{
    "id": "tx_12345",
    "type": "buy",
    "item_id": "apple",
    "quantity": 10,
    "unit_price": 12,
    "total": 120,
    "timestamp": "2025-08-12T21:00:00Z"
}
```

## Events

The backend can emit events that the frontend listens to:

### PriceUpdated
Emitted when market prices change.
```gdscript
signal price_updated(item_id: String, new_price: int)
```

### TransactionComplete
Emitted when a transaction is completed.
```gdscript
signal transaction_complete(transaction: Dictionary)
```

### DayChanged
Emitted when the game day advances.
```gdscript
signal day_changed(new_day: int, season: String)
```

### WeatherChanged
Emitted when weather changes.
```gdscript
signal weather_changed(weather: String)
```

## Usage Example

```gdscript
extends Node

var game_api

func _ready():
    # Initialize the game
    game_api = load("res://bin/merchant_game.gdextension")
    if game_api.initialize_game():
        print("Game initialized successfully")
    
    # Start a new game
    var result = game_api.start_new_game("Player1")
    if result.success:
        print("New game started")
    
    # Get market prices
    var prices = game_api.get_market_prices()
    for item_id in prices:
        print("%s: %d gold" % [item_id, prices[item_id]])
    
    # Buy an item
    var buy_result = game_api.buy_item("apple", 10)
    if buy_result.success:
        print("Bought 10 apples for %d gold" % buy_result.total_cost)
```

## Notes

- All monetary values are in gold (integer)
- All dates/times are in ISO 8601 format
- The API is synchronous; all calls block until completion
- The backend maintains game state; frontend is for display only