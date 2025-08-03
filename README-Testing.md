# Merchant Tails - Testing Guide

## Overview

This guide covers the testing framework and procedures for Merchant Tails. The game includes comprehensive unit tests, integration tests, performance tests, and save/load integrity tests.

## Test Structure

```
Assets/Tests/
├── Runtime/
│   ├── TestBase.cs              # Base class for all tests
│   ├── MarketSystemTests.cs     # Market system unit tests
│   ├── InventorySystemTests.cs  # Inventory system unit tests
│   ├── TimeManagerTests.cs      # Time manager unit tests
│   ├── IntegrationTests.cs      # System integration tests
│   ├── PerformanceTests.cs      # Performance benchmarks
│   ├── SaveLoadTests.cs         # Save/load integrity tests
│   └── MerchantTails.Tests.Runtime.asmdef
└── Editor/
    └── (Editor-only tests)
```

## Running Tests

### In Unity Editor

1. Open Unity Test Runner: `Window > General > Test Runner`
2. Switch to "PlayMode" tab for runtime tests
3. Click "Run All" or select specific tests

### From Command Line

```bash
# Run all tests
Unity -runTests -projectPath . -testResults results.xml -testPlatform PlayMode

# Run specific test category
Unity -runTests -projectPath . -testResults results.xml -testPlatform PlayMode -testFilter "MarketSystemTests"
```

## Test Categories

### 1. Unit Tests

Test individual components in isolation:

- **MarketSystemTests**: Price calculations, trends, seasonal modifiers
- **InventorySystemTests**: Item management, transfers, quality degradation
- **TimeManagerTests**: Time progression, day/season cycles, phase transitions

### 2. Integration Tests

Test interactions between multiple systems:

- **BuyAndSellFlow**: Complete trading cycle
- **DayProgressionFlow**: Time advancement affecting all systems
- **RankProgression**: Feature unlocking with rank advancement
- **EventTriggering**: Event system affecting market prices
- **CompleteGameLoop**: 24-hour gameplay simulation

### 3. Performance Tests

Measure system performance and resource usage:

- **MarketPriceUpdate**: Price calculation speed
- **InventoryOperations**: Add/remove item performance
- **MemoryAllocation**: Memory usage during gameplay
- **FrameRate**: FPS stability under load
- **SaveSystem**: Save/load operation speed

### 4. Save/Load Tests

Ensure data persistence integrity:

- **PlayerData**: Money, rank, statistics
- **TimeData**: Day, season, phase, progress
- **ComplexInventory**: Multiple items with various states
- **MarketHistory**: Price history preservation
- **Investments**: Bank deposits and shop investments

## Debug Features

### Debug Manager (F1)

Access debug menu for testing:

- **Money Manipulation**: Add/subtract money instantly
- **Time Control**: Advance hours, days, or seasons
- **Item Spawning**: Add items to inventory
- **Rank Setting**: Change merchant rank
- **Event Triggering**: Force specific events

### Console Commands (~)

Use console for advanced debugging:

```
money <amount>        # Add/subtract money
day <count>          # Advance days
season [name]        # Change/advance season
rank <rank>          # Set merchant rank
item <type> <qty>    # Add items
event [id]           # Trigger event
save                 # Quick save
load                 # Quick load
```

### Keyboard Shortcuts

- **Shift+T**: Advance 1 hour
- **Shift+D**: Advance 1 day
- **Shift+S**: Advance to next season
- **Ctrl+M**: Add 1000G
- **Ctrl+N**: Remove 1000G
- **Alt+S**: Quick save
- **Alt+L**: Quick load

## Writing New Tests

### Test Base Class

All tests should inherit from `TestBase`:

```csharp
public class MySystemTests : TestBase
{
    [Test]
    public void MyTest()
    {
        // Arrange
        var item = CreateTestItem(ItemType.Fruit, 10);
        
        // Act
        inventorySystem.AddItem(item);
        
        // Assert
        Assert.AreEqual(10, inventorySystem.GetItemCount(ItemType.Fruit));
    }
}
```

### Common Helpers

- `CreateTestItem()`: Create test items with specific properties
- `AdvanceTime()`: Progress time by hours
- `AdvanceDays()`: Progress time by days
- `AssertFloatEquals()`: Compare floats with tolerance
- `WaitForCondition()`: Wait for async conditions

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Use `Teardown()` to reset state
3. **Naming**: Use descriptive test names: `MethodName_Scenario_ExpectedResult`
4. **Assertions**: Include meaningful assertion messages
5. **Performance**: Keep tests fast (< 1 second each)

## Continuous Integration

Tests are automatically run on:
- Pull request creation
- Commits to main branch
- Commits to develop branch

CI Pipeline includes:
- Unity PlayMode tests execution
- Code coverage reporting (target: 75%+)
- Test results published to GitHub
- Automatic PR comments with coverage data

Failed tests will block merges and deployments.

### Setting up CI

See `.github/workflows/README.md` for detailed setup instructions including Unity license configuration.

## Troubleshooting

### Common Issues

1. **Static Instance Conflicts**: TestBase clears static instances between tests
2. **File System Access**: Use Unity's Application paths for file operations
3. **Timing Issues**: Use coroutines and WaitForCondition for async operations
4. **Memory Leaks**: Check GameObject destruction in Teardown

### Debug Output

Enable verbose logging:

```csharp
ErrorHandler.SetDebugMode(true);
```

View test logs in:
- Unity Console
- Test Runner output
- `Logs/test_results.log`

## Performance Benchmarks

Target performance metrics:

- Market price update: < 16.67ms (60 FPS)
- Inventory operations: < 1ms per item
- Save/Load: < 100ms for full game state
- Memory growth: < 10MB per minute
- Minimum FPS under load: 30

## Coverage Goals

- Unit test coverage: > 80%
- Integration test coverage: > 60%
- Critical path coverage: 100%
- Edge case coverage: > 70%