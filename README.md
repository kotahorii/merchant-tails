# Merchant Tails 🏪

[![Godot Engine](https://img.shields.io/badge/Godot-4.4.1-blue.svg)](https://godotengine.org/)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![CI/CD](https://github.com/yourusername/merchant-tails/workflows/Build%20and%20Test/badge.svg)](https://github.com/yourusername/merchant-tails/actions)
[![Documentation](https://img.shields.io/badge/docs-available-brightgreen.svg)](docs/)

A business simulation game where players learn investment fundamentals through fantasy merchant trading, built with Godot Engine and Go.

## 🎮 Game Overview

**Merchant Tails** is an educational business simulation game set in a fantasy world. Players take on the role of a merchant, learning investment principles through buying and selling goods without explicitly using financial terminology.

### Key Features
- 🏰 Fantasy merchant simulation in the capital city of Elm
- 📈 Dynamic market system teaching investment fundamentals
- 🎨 Warm, approachable art style
- 🌍 Localization support (Japanese/English)
- 💾 Cross-platform save system with Steam Cloud support

## 🏗️ Architecture

Built with **Test-Driven Development (TDD)** and **Documentation-Driven Development (DDD)** principles.

### Tech Stack
- **Game Engine**: Godot 4.4.1
- **Core Logic**: Go 1.23+ (via GDExtension)
- **Data Format**: Protocol Buffers
- **Architecture**: Clean Architecture + Domain-Driven Design
- **Testing**: Go testing + Godot unit tests

## 📋 Prerequisites

- Godot Engine 4.4.1
- Go 1.23 or higher
- Protocol Buffers compiler
- Make (for build automation)
- Git LFS (for asset management)

## 🚀 Getting Started

### 1. Clone the Repository
```bash
git clone https://github.com/yourusername/merchant-tails.git
cd merchant-tails
git lfs pull
```

### 2. Install Dependencies
```bash
# Install Go dependencies
cd game
go mod download

# Install protobuf compiler
# macOS
brew install protobuf

# Linux
apt-get install protobuf-compiler

# Windows
choco install protobuf
```

### 3. Build the Project
```bash
# Build Go extension
make build-go

# Build Godot project
make build-godot

# Or build everything
make build-all
```

### 4. Run Tests
```bash
# Run Go tests
make test-go

# Run Godot tests
make test-godot

# Run all tests with coverage
make test-coverage
```

## 🧪 Test-Driven Development

We follow strict TDD principles. Every feature must have tests written first.

### Running Tests
```bash
# Unit tests
go test ./game/...

# Integration tests
go test ./tests/integration -tags=integration

# Benchmark tests
go test -bench=. ./game/...

# Watch mode for TDD
make test-watch
```

### Test Structure
```
tests/
├── unit/           # Unit tests for individual components
├── integration/    # Integration tests for system interactions
├── e2e/           # End-to-end game flow tests
└── performance/   # Performance and benchmark tests
```

## 📚 Documentation-Driven Development

All features must be documented before implementation.

### Documentation Structure
```
docs/
├── prd.md              # Product Requirements Document
├── design-doc.md       # Technical Design Document
├── development-todo.md # Development task tracking
├── api/               # API documentation
├── architecture/      # Architecture decisions
└── guides/           # Development guides
```

### Generate Documentation
```bash
# Generate API documentation
make docs-api

# Generate architecture diagrams
make docs-architecture

# Serve documentation locally
make docs-serve
```

## 🏃‍♂️ Development Workflow

### 1. Documentation First
```bash
# 1. Update design documentation
vim docs/design-doc.md

# 2. Create/update API specs
vim docs/api/feature-name.md
```

### 2. Write Tests
```bash
# 3. Write failing tests
vim game/internal/domain/feature_test.go

# 4. Run tests (should fail)
go test ./game/internal/domain
```

### 3. Implement Feature
```bash
# 5. Implement minimum code to pass tests
vim game/internal/domain/feature.go

# 6. Run tests (should pass)
go test ./game/internal/domain
```

### 4. Refactor
```bash
# 7. Refactor while keeping tests green
make test-watch
```

## 🎯 Development Phases

### Phase 0: Creative Assets (Weeks 1-8)
- Character design and sprites
- Background art and environments
- Music and sound effects
- UI/UX design

### Phase 1: Foundation (Weeks 1-4)
- Environment setup
- Core architecture implementation
- CI/CD pipeline

### Phase 2: Core Systems (Weeks 5-12)
- Game loop implementation
- ECS system
- Market mechanics
- Inventory management

### Phase 3: Game Logic (Weeks 13-20)
- Trading systems
- Event systems
- AI merchants
- Progression systems

[See full roadmap](docs/development-todo.md)

## 🔧 Build Commands

```bash
# Development
make dev           # Run in development mode
make hot-reload    # Run with hot reload

# Testing
make test          # Run all tests
make test-unit     # Run unit tests only
make test-integration # Run integration tests
make bench         # Run benchmarks

# Building
make build         # Build for current platform
make build-all     # Build for all platforms
make build-windows # Build for Windows
make build-mac     # Build for macOS
make build-linux   # Build for Linux

# Documentation
make docs          # Generate all documentation
make docs-serve    # Serve documentation locally

# Utilities
make clean         # Clean build artifacts
make fmt           # Format code
make lint          # Run linters
make proto         # Generate protobuf files
```

## 📦 Project Structure

```
merchant-tails/
├── godot/                 # Godot project files
│   ├── scenes/           # Game scenes
│   ├── scripts/          # GDScript files (UI only)
│   └── resources/        # Assets and resources
├── game/                  # Go game logic
│   ├── cmd/              # Entry points
│   ├── internal/         # Internal packages
│   │   ├── domain/       # Domain models
│   │   ├── application/  # Use cases
│   │   └── infrastructure/ # External interfaces
│   └── pkg/              # Public packages
├── proto/                 # Protocol buffer definitions
├── tests/                 # Test suites
├── docs/                  # Documentation
└── scripts/              # Build and utility scripts
```

## 🧪 Testing Strategy

### Test Coverage Requirements
- Unit Tests: 80% minimum coverage
- Integration Tests: All critical paths
- E2E Tests: Main game flows
- Performance Tests: All systems under load

### Continuous Integration
```yaml
# Tests run on every push
- Unit tests
- Integration tests
- Code coverage report
- Performance regression tests
- Documentation generation
```

## 📈 Performance Targets

- **Frame Rate**: Stable 60 FPS
- **Memory Usage**: < 500MB RAM
- **Load Time**: < 3 seconds
- **Save/Load**: < 1 second
- **Market Update**: < 16ms per frame

## 🤝 Contributing

We welcome contributions! Please follow our TDD/DDD approach:

1. **Fork** the repository
2. **Document** your feature in `/docs`
3. **Write tests** for your feature
4. **Implement** the feature
5. **Ensure** all tests pass
6. **Submit** a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Design Document](docs/design-doc.md)
- [Product Requirements](docs/prd.md)
- [Development Tasks](docs/development-todo.md)
- [API Documentation](docs/api/)
- [Wiki](https://github.com/yourusername/merchant-tails/wiki)

## 💬 Support

- [Discord Server](https://discord.gg/merchanttails)
- [Issue Tracker](https://github.com/yourusername/merchant-tails/issues)
- [Discussions](https://github.com/yourusername/merchant-tails/discussions)

## 🎮 Minimum Requirements

### Development
- **OS**: Windows 10/11, macOS 12+, Ubuntu 20.04+
- **RAM**: 8GB minimum, 16GB recommended
- **Storage**: 10GB free space
- **GPU**: OpenGL 4.6 compatible

### Runtime
- **OS**: Windows 10+, macOS 10.15+, Ubuntu 18.04+
- **RAM**: 4GB
- **Storage**: 2GB free space
- **GPU**: OpenGL 3.3 compatible

---

<p align="center">
  Made with ❤️ using Godot Engine and Go
</p>
