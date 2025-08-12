# Merchant Tails 🏪

[![Godot Engine](https://img.shields.io/badge/Godot-4.4.1-blue.svg)](https://godotengine.org/)
[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![CI/CD](https://github.com/yourusername/merchant-tails/workflows/CI/badge.svg)](https://github.com/yourusername/merchant-tails/actions)
[![Documentation](https://img.shields.io/badge/docs-complete-brightgreen.svg)](docs/)

An educational business simulation game teaching investment fundamentals through fantasy merchant trading.

## 🎮 Overview

**Merchant Tails** is a beginner-friendly business simulation game where players learn investment principles by trading goods in a fantasy marketplace. Built with simplicity in mind, it focuses on core trading mechanics without overwhelming complexity.

### ✨ Features

- 🏰 **Fantasy Setting** - Trade in the merchant city of Elm
- 📈 **Dynamic Markets** - Learn supply, demand, and price patterns
- 🏦 **Simple Banking** - Save gold and earn interest
- 🌤️ **Weather Effects** - Weather impacts market conditions
- 🌍 **Localization** - Full English and Japanese support
- 💾 **Save System** - Auto-save and manual save options

### 🎯 Educational Goals

Learn investment fundamentals through gameplay:
- Buy low, sell high strategies
- Risk management and diversification
- Market timing and trend analysis
- Compound interest through banking
- Seasonal market patterns

## 🏗️ Architecture

### Technology Stack
- **Frontend**: Godot 4.4.1 (GDScript)
- **Backend**: Go 1.24 (Game Logic)
- **Integration**: GDExtension (C ABI)
- **Data**: Protocol Buffers
- **Build**: Make, Docker
- **CI/CD**: GitHub Actions

### Design Principles
- **Clean Architecture** - Separated layers with clear boundaries
- **Domain-Driven Design** - Business logic in domain layer
- **Test-Driven Development** - Tests written before code
- **Simplicity First** - Avoid over-engineering

## 🚀 Quick Start

### Prerequisites
- Godot 4.4.1
- Go 1.24 (specifically, not 1.23)
- Make
- Protocol Buffers compiler (optional)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/merchant-tails.git
cd merchant-tails
```

2. **Build the project**
```bash
# Build everything
make build-all

# Or build separately
make build-go      # Build Go library
make build-godot   # Build Godot project
```

3. **Run the game**
```bash
make run
```

### Development

```bash
# Run tests
cd game && go test ./...

# Run with hot reload
make dev

# Format and lint
make fmt
make lint

# Run all checks
make check
```

## 📁 Project Structure

```
merchant-tails/
├── game/                  # Go game logic
│   ├── cmd/              # Entry points
│   ├── internal/         # Core implementation
│   │   ├── domain/      # Business logic
│   │   ├── application/ # Use cases
│   │   └── presentation/# API layer
│   └── tests/           # Test files
├── godot/                # Godot project
│   ├── scenes/          # Game scenes
│   ├── scripts/         # GDScript
│   └── localization/   # Translations
├── docs/                 # Documentation
│   ├── api-specification.md
│   ├── developer-guide.md
│   └── user-manual.md
└── proto/               # Protocol Buffers
```

## 🎮 Gameplay

### Core Loop
1. **Buy** items when prices are low
2. **Store** in shop or warehouse
3. **Sell** when prices rise
4. **Bank** profits for interest
5. **Advance** through merchant ranks

### Item Categories
- 🍎 **Fruits** - Fast turnover, low margins
- 🧪 **Potions** - Medium risk/reward
- ⚔️ **Weapons** - High value items
- 💍 **Accessories** - Luxury goods
- 📚 **Spellbooks** - Specialized market
- 💎 **Gems** - High risk investments

### Progression
- **Apprentice** → **Journeyman** → **Expert** → **Master**

## 📚 Documentation

- [API Specification](docs/api-specification.md) - GDExtension API reference
- [Developer Guide](docs/developer-guide.md) - Setup and development workflow
- [User Manual](docs/user-manual.md) - How to play the game
- [CLAUDE.md](CLAUDE.md) - AI assistant instructions
- [Release Notes](RELEASE_NOTES.md) - Version history

## 🧪 Testing

```bash
# Unit tests
cd game && go test ./...

# Specific package tests
cd game && go test ./internal/domain/market -v

# With coverage
cd game && go test -cover ./...

# Benchmarks
cd game && go test -bench=. ./...
```

## 🐛 Known Issues

- GDExtension may require manual library path configuration
- Save files are not compatible between alpha versions
- Some tutorial steps may not trigger correctly

## 🤝 Contributing

### Before Pushing
Always run these commands:
```bash
cd game && go test ./...        # Run tests
cd game && golangci-lint run    # Run linter
make build-go                    # Build project
make check                       # Run all checks
```

### Guidelines
1. Follow TDD - write tests first
2. Keep it simple - avoid over-engineering
3. Document your code
4. Run all checks before committing

## 📝 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Godot Engine community
- Go community
- Beta testers and contributors
- Investment education advisors

## 📞 Support

- 📖 Check the [documentation](docs/)
- 🐛 Report [issues on GitHub](https://github.com/yourusername/merchant-tails/issues)
- 💬 Join our Discord (coming soon)

## 🚦 Project Status

**Current Version**: 0.1.0 (Alpha)

This project is in active development. Core gameplay is functional but expect:
- Balance adjustments
- Bug fixes
- Performance improvements
- Additional content

---

Made with ❤️ for learning investment through gaming