# Merchant Tails Developer Guide

## Table of Contents
1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Development Setup](#development-setup)
4. [Building the Project](#building-the-project)
5. [Testing](#testing)
6. [Project Structure](#project-structure)
7. [Key Systems](#key-systems)
8. [Development Workflow](#development-workflow)
9. [Debugging](#debugging)
10. [Contributing](#contributing)

## Project Overview

Merchant Tails is an educational business simulation game that teaches investment fundamentals through fantasy merchant trading. The game is built with a unique architecture combining Godot 4.4.1 for rendering and UI with Go 1.24 for core game logic via GDExtension.

### Technology Stack
- **Frontend**: Godot 4.4.1 (GDScript)
- **Backend**: Go 1.24
- **Integration**: GDExtension (C ABI)
- **Data**: Protocol Buffers
- **Build**: Make, Docker
- **CI/CD**: GitHub Actions

## Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────┐
│     Presentation (Godot/GDScript)   │
├─────────────────────────────────────┤
│      Application (Go Use Cases)      │
├─────────────────────────────────────┤
│       Domain (Go Business Logic)     │
├─────────────────────────────────────┤
│    Infrastructure (Go External)      │
└─────────────────────────────────────┘
```

### Key Principles
1. **Domain-Driven Design (DDD)** - Business logic isolated in domain layer
2. **Test-Driven Development (TDD)** - Tests written before implementation
3. **Event-Driven Architecture** - Loose coupling via event bus
4. **Simplicity First** - Avoid over-engineering

## Development Setup

### Prerequisites
- Go 1.24 (not 1.23)
- Godot 4.4.1
- Protocol Buffers compiler
- Make
- Docker (optional)

### Initial Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/merchant-tails.git
cd merchant-tails
```

2. Install dependencies:
```bash
# Install Go dependencies
cd game
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2
go install github.com/cosmtrek/air@latest
```

3. Generate Protocol Buffers:
```bash
make proto
```

## Building the Project

### Quick Build
```bash
# Build everything
make build-all

# Build Go library only
make build-go

# Build Godot project only
make build-godot
```

### Platform-Specific Builds
```bash
# Windows
GOOS=windows GOARCH=amd64 make build-go

# macOS
GOOS=darwin GOARCH=amd64 make build-go

# Linux
GOOS=linux GOARCH=amd64 make build-go
```

### Docker Build
```bash
docker-compose up --build
```

## Testing

### Running Tests

```bash
# Run all tests
cd game && go test ./...

# Run specific domain tests
cd game && go test ./internal/domain/market -v

# Run with coverage
cd game && go test -cover ./...

# Run with race detection
cd game && go test -race ./...

# Watch mode for TDD
make test-watch
```

### Test Organization
- Unit tests: Same directory as code with `_test.go` suffix
- Integration tests: `tests/integration/`
- Benchmarks: Use `Benchmark*` prefix

### Writing Tests

Example test structure:
```go
func TestMarket_CalculatePrice(t *testing.T) {
    // Arrange
    market := NewMarket()
    item := NewItem("apple", 10.0)
    
    // Act
    price := market.CalculatePrice(item)
    
    // Assert
    assert.Equal(t, 12.0, price)
}
```

## Project Structure

```
merchant-tails/
├── game/                      # Go game logic
│   ├── cmd/
│   │   └── gdextension/      # GDExtension entry point
│   ├── internal/
│   │   ├── domain/           # Core business logic
│   │   │   ├── market/      # Market system
│   │   │   ├── merchant/    # Merchant logic
│   │   │   ├── item/        # Item definitions
│   │   │   └── inventory/   # Inventory management
│   │   ├── application/      # Use cases
│   │   ├── infrastructure/  # External interfaces
│   │   └── presentation/    # API layer
│   └── tests/               # Test files
├── godot/                    # Godot project
│   ├── scenes/              # Game scenes
│   ├── scripts/             # GDScript files
│   ├── resources/           # Game resources
│   └── localization/       # Translation files
├── proto/                   # Protocol Buffers
├── docs/                    # Documentation
└── Makefile                # Build commands
```

## Key Systems

### Market System
The market system handles dynamic pricing based on supply and demand:

```go
// Domain model
type Market struct {
    pricingEngine *PricingEngine
    state        *MarketState
    eventBus     *EventBus
}

// Price calculation with modifiers
func (m *Market) CalculatePrice(item *Item) float64 {
    basePrice := item.BasePrice
    modifiers := m.pricingEngine.GetModifiers()
    return m.pricingEngine.ApplyModifiers(basePrice, modifiers)
}
```

### Merchant System
Simplified merchant system for single-player gameplay:

```go
type Merchant struct {
    ID         string
    Gold       float64
    Inventory  *Inventory
    Reputation float64
}
```

### Event System
Central event bus for decoupled communication:

```go
// Publishing events
eventBus.Publish("market.price_changed", PriceChangedEvent{
    ItemID:   "apple",
    NewPrice: 15.0,
})

// Subscribing to events
eventBus.Subscribe("market.price_changed", func(e Event) {
    // Handle price change
})
```

## Development Workflow

### TDD Workflow
1. Write failing test
2. Run test to verify failure
3. Implement minimal code to pass
4. Refactor while keeping tests green

### Adding New Features
1. Define interfaces in domain layer
2. Write comprehensive tests first
3. Implement domain logic
4. Create application use case
5. Expose via GDExtension API
6. Add minimal GDScript for UI

### Code Quality
```bash
# Format code
make fmt

# Run linter
make lint

# Run all checks
make check
```

## Debugging

### Logging
The game uses a simple logging system:

```go
log.Info("Player purchased item", "item", itemID, "price", price)
log.Warn("Low inventory", "item", itemID)
log.Error("Transaction failed", "error", err)
```

### Godot Debugging
1. Enable debug mode in project settings
2. Use `print()` statements in GDScript
3. Use Godot's built-in debugger

### Go Debugging
1. Use `dlv` debugger:
```bash
dlv test ./internal/domain/market
```

2. Add debug prints:
```go
fmt.Printf("DEBUG: price=%v\n", price)
```

## Contributing

### Pre-Push Checklist
**IMPORTANT**: Always run these before pushing:

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

### Code Style Guidelines
- Follow Go idioms and best practices
- Keep functions small and focused
- Write descriptive variable names
- Add comments for complex logic
- Maintain test coverage above 80%

### Commit Messages
Use conventional commits:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Tests
- `refactor:` Code refactoring
- `chore:` Maintenance

Example:
```
feat: add bank deposit functionality

- Implement deposit and withdrawal methods
- Add interest calculation
- Update UI to show bank balance
```

### Pull Request Process
1. Create feature branch from `develop`
2. Write tests first (TDD)
3. Implement feature
4. Ensure all tests pass
5. Run linter and fix issues
6. Create PR with description
7. Wait for CI to pass
8. Request review

## Common Issues

### Build Errors
- Ensure Go 1.24 is installed (not 1.23)
- Run `make clean` before rebuilding
- Check that Protocol Buffers are generated

### Test Failures
- Run tests in isolation first
- Check for race conditions with `-race`
- Ensure test data is properly initialized

### GDExtension Issues
- Verify the library path in Godot project
- Check that the C library is properly built
- Ensure function signatures match

## Performance Optimization

### Critical Paths
- Market price updates: Target < 16ms
- Merchant decisions: Use batch processing
- Inventory operations: O(1) lookups with maps

### Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

## Resources

- [Godot Documentation](https://docs.godotengine.org/)
- [Go Documentation](https://go.dev/doc/)
- [GDExtension Guide](https://docs.godotengine.org/en/stable/tutorials/scripting/gdextension/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)

## Support

For questions or issues:
1. Check existing GitHub issues
2. Read the FAQ in docs/
3. Create a new issue with details
4. Join our Discord community (if applicable)