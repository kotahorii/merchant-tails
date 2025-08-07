# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Merchant Tails is a business simulation game teaching investment fundamentals through fantasy merchant trading. Built with Godot 4.4.1 for rendering/UI and Go 1.24 for core game logic via GDExtension.

## Build and Development Commands

### Essential Commands
```bash
# Run tests (TDD workflow)
cd game && go test ./...                    # Run all Go tests
cd game && go test ./internal/domain/market -v  # Run specific domain tests
make test-watch                             # Watch mode for TDD development

# Build the project
make build-go                               # Build Go GDExtension library
make build-godot                            # Build Godot project
make build-all                              # Build everything

# Code quality
make fmt                                    # Format Go code
make lint                                   # Run golangci-lint
make check                                  # Run fmt, vet, lint, and tests

# Development
make dev                                    # Run with hot reload (requires air)
make run                                    # Run the game in Godot
```

### Single Test Execution
```bash
# Run a specific test function
cd game && go test -run TestNewMarket ./internal/domain/market

# Run tests for a specific package with verbose output
cd game && go test -v ./internal/domain/merchant

# Run with race detection
cd game && go test -race ./...

# Generate coverage for specific package
cd game && go test -coverprofile=coverage.out ./internal/domain/item
```

## Architecture Overview

### Core Architecture Principles
1. **Clean Architecture** - Strict layer separation with dependency inversion
2. **Domain-Driven Design (DDD)** - Business logic in domain layer
3. **Test-Driven Development (TDD)** - Tests written before implementation
4. **Event-Driven Architecture** - Loose coupling via event bus

### Layer Structure
```
Presentation (Godot/GDScript) → Application (Go) → Domain (Go) → Infrastructure (Go)
```

- **Presentation**: Godot scenes and minimal GDScript for UI only
- **Application**: Use cases and DTOs in `game/internal/application/`
- **Domain**: Core business logic in `game/internal/domain/` (merchant, market, item, inventory)
- **Infrastructure**: External interfaces in `game/internal/infrastructure/`

### Critical Integration Points

#### Go-Godot Bridge (GDExtension)
- Entry point: `game/cmd/gdextension/main.go`
- API layer: `game/internal/presentation/api/`
- Godot calls Go functions via GDExtension bindings
- Data serialization uses Protocol Buffers

#### Domain Models
Key domain packages with complex interactions:
- `game/internal/domain/market/` - Dynamic pricing engine with supply/demand
- `game/internal/domain/merchant/` - AI behavior and trading strategies
- `game/internal/domain/item/` - Item categories with durability/volatility
- These domains interact through interfaces to maintain loose coupling

#### Market System Architecture
The market system uses:
- `PricingEngine` with modular price modifiers
- `MarketState` tracking demand/supply/season
- Event system for market events (dragon attacks, festivals)
- Price history with trend analysis

#### Merchant System Architecture
Merchants have:
- Personality types (Aggressive, Conservative, Balanced, Opportunistic)
- Trading strategies implementing `TradingStrategy` interface
- Network effects for information propagation
- Risk tolerance calculations based on personality and gold

## Testing Strategy

### TDD Workflow
1. Write failing test in `*_test.go`
2. Run test to verify failure
3. Implement minimal code to pass
4. Refactor while keeping tests green

### Test Organization
- Unit tests: Same directory as code with `_test.go` suffix
- Integration tests: `tests/integration/` with build tag
- Benchmarks: Use `Benchmark*` prefix in test files

## Data Flow

### Save System
- Protocol Buffers definitions in `proto/`
- Save data: `proto/save/gamestate.proto`
- Config data: `proto/config/game_config.proto`
- Generated Go code via `make proto`

### Event System
- Central `EventBus` for decoupled communication
- Events defined in `game/internal/domain/event/`
- Systems subscribe to relevant events

## Key Development Patterns

### Adding New Domain Feature
1. Define interfaces in domain layer
2. Write comprehensive tests first (TDD)
3. Implement domain logic
4. Create application use case
5. Expose via GDExtension API
6. Minimal GDScript for UI binding

### Modifying Market Mechanics
- Price calculations in `market.PricingEngine.CalculatePrice()`
- Add new modifiers by implementing `PriceModifier` interface
- Market events affect state through `Market.ApplyEvent()`

### Adding Merchant Behavior
- Implement `TradingStrategy` interface
- Add personality type in `merchant.PersonalityType`
- Modify `Merchant.MakeTradingDecision()` for decision logic

## Performance Considerations

### Critical Paths
- Market price updates (target: < 16ms)
- Merchant AI decisions (batch processing)
- Inventory operations (O(1) lookups via maps)

### Optimization Points
- Use object pooling for frequently created objects
- Batch market updates in `Market.UpdatePrices()`
- Cache price calculations when market state unchanged

## Common Pitfalls

1. **Don't put game logic in GDScript** - Only UI handling
2. **Always write tests first** - TDD is mandatory
3. **Use interfaces for dependencies** - Maintain loose coupling
4. **Event names are constants** - Define in domain layer
5. **Protocol Buffer changes** - Run `make proto` after modifications

## CI/CD Pipeline

GitHub Actions workflow (`.github/workflows/ci.yml`):
- Go tests with coverage
- Godot project build
- Benchmarks with performance tracking
- Security scanning with Trivy
- Runs on push to main/develop and PRs

## Environment Requirements

- Go 1.24 (not 1.23 - explicitly required)
- Godot 4.4.1
- Protocol Buffers compiler
- golangci-lint for code quality
- air for hot reload development

## Critical Pre-Push Checklist

**IMPORTANT: ALWAYS run these commands before pushing to GitHub:**

```bash
# 1. Run all tests
cd game && go test ./...

# 2. Run linter
cd game && golangci-lint run

# 3. Build the project
make build-go

# 4. Run all checks at once
make check
```

**Rules:**
- NEVER push if tests are failing
- NEVER push if build errors exist
- ALWAYS fix all issues before committing
- If CI fails after push, fix immediately
