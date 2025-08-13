# Merchant Tails - Build System
# =============================

# Variables
GO := go
GODOT := godot
GOFLAGS := -v
LDFLAGS := -s -w
BUILD_DIR := build
GAME_DIR := game
GODOT_DIR := godot

# Platform detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	PLATFORM := linux
	LIB_EXT := so
	GODOT_PLATFORM := x11
endif
ifeq ($(UNAME_S),Darwin)
	PLATFORM := macos
	LIB_EXT := dylib
	GODOT_PLATFORM := macos
endif
ifeq ($(OS),Windows_NT)
	PLATFORM := windows
	LIB_EXT := dll
	GODOT_PLATFORM := windows
endif

# Build targets
LIB_NAME := libmerchant_game.$(LIB_EXT)
TARGET := $(BUILD_DIR)/$(LIB_NAME)

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

.PHONY: all build test clean run help

# Default target
all: build

# Help command
help:
	@echo "$(GREEN)Merchant Tails Build System$(NC)"
	@echo "============================"
	@echo ""
	@echo "Available targets:"
	@echo "  $(YELLOW)make build$(NC)      - Build the Go GDExtension library"
	@echo "  $(YELLOW)make build-go$(NC)   - Build only the Go library"
	@echo "  $(YELLOW)make build-godot$(NC) - Build the Godot project"
	@echo "  $(YELLOW)make build-all$(NC)  - Build everything (Go + Godot)"
	@echo "  $(YELLOW)make test$(NC)       - Run all Go tests"
	@echo "  $(YELLOW)make test-watch$(NC) - Run tests in watch mode (requires entr)"
	@echo "  $(YELLOW)make test-cover$(NC) - Run tests with coverage"
	@echo "  $(YELLOW)make bench$(NC)      - Run benchmarks"
	@echo "  $(YELLOW)make fmt$(NC)        - Format Go code"
	@echo "  $(YELLOW)make lint$(NC)       - Run golangci-lint"
	@echo "  $(YELLOW)make check$(NC)      - Run fmt, vet, lint, and tests"
	@echo "  $(YELLOW)make clean$(NC)      - Clean build artifacts"
	@echo "  $(YELLOW)make run$(NC)        - Run the game in Godot"
	@echo "  $(YELLOW)make dev$(NC)        - Run with hot reload (requires air)"
	@echo "  $(YELLOW)make proto$(NC)      - Generate protobuf files"
	@echo "  $(YELLOW)make deps$(NC)       - Install dependencies"
	@echo ""
	@echo "Platform: $(PLATFORM)"
	@echo "Library Extension: $(LIB_EXT)"

# Build the Go GDExtension library
build: build-go

build-go:
	@echo "$(GREEN)Building Go GDExtension library...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(GODOT_DIR)/lib
	cd $(GAME_DIR) && CGO_ENABLED=1 $(GO) build -buildmode=c-shared \
		-ldflags "$(LDFLAGS)" \
		-o ../$(BUILD_DIR)/$(LIB_NAME) \
		./cmd/gdextension
	@cp $(BUILD_DIR)/$(LIB_NAME) $(GODOT_DIR)/lib/$(LIB_NAME)
	@echo "$(GREEN)✓ GDExtension library built: $(GODOT_DIR)/lib/$(LIB_NAME)$(NC)"

# Build the Godot project
build-godot:
	@echo "$(GREEN)Building Godot project...$(NC)"
	cd $(GODOT_DIR) && $(GODOT) --export-release "$(GODOT_PLATFORM)" ../$(BUILD_DIR)/merchant_tails
	@echo "$(GREEN)✓ Godot build complete$(NC)"

# Build everything
build-all: build-go build-godot
	@echo "$(GREEN)✓ Full build complete$(NC)"

# Run unit tests
test:
	@echo "$(GREEN)Running unit tests...$(NC)"
	cd $(GAME_DIR) && $(GO) test $(GOFLAGS) ./...

# Run integration tests
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	cd $(GAME_DIR) && $(GO) test -tags=integration $(GOFLAGS) ./tests/integration

# Run all tests
test-all: test test-integration
	@echo "$(GREEN)✓ All tests complete$(NC)"

# Run tests in watch mode
test-watch:
	@echo "$(GREEN)Running tests in watch mode...$(NC)"
	@which entr > /dev/null || (echo "$(RED)Error: entr not installed. Install with: brew install entr$(NC)" && exit 1)
	find $(GAME_DIR) -name "*.go" | entr -c make test

# Run tests with coverage
test-cover:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	cd $(GAME_DIR) && $(GO) test -coverprofile=coverage.out ./...
	cd $(GAME_DIR) && $(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: $(GAME_DIR)/coverage.html$(NC)"

# Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	cd $(GAME_DIR) && $(GO) test -bench=. -benchmem ./...

# Format Go code
fmt:
	@echo "$(GREEN)Formatting Go code...$(NC)"
	cd $(GAME_DIR) && $(GO) fmt ./...
	cd $(GAME_DIR) && gofumpt -l -w .
	@echo "$(GREEN)✓ Code formatted$(NC)"

# Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	cd $(GAME_DIR) && golangci-lint run
	@echo "$(GREEN)✓ Linting complete$(NC)"

# Run all checks
check: fmt
	@echo "$(GREEN)Running checks...$(NC)"
	cd $(GAME_DIR) && $(GO) vet ./...
	@make lint
	@make test
	@echo "$(GREEN)✓ All checks passed$(NC)"

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f $(GODOT_DIR)/bin/*.$(LIB_EXT)
	rm -f $(GODOT_DIR)/bin/*.h
	rm -f $(GAME_DIR)/coverage.out
	rm -f $(GAME_DIR)/coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

# Run the game in Godot
run:
	@echo "$(GREEN)Starting Merchant Tails...$(NC)"
	cd $(GODOT_DIR) && $(GODOT) --debug

# Development mode with hot reload
dev:
	@echo "$(GREEN)Starting development mode...$(NC)"
	@which air > /dev/null || (echo "$(RED)Error: air not installed. Install with: go install github.com/cosmtrek/air@latest$(NC)" && exit 1)
	cd $(GAME_DIR) && air

# Generate protobuf files
proto:
	@echo "$(GREEN)Generating protobuf files...$(NC)"
	protoc --go_out=. --go_opt=paths=source_relative \
		proto/save/*.proto proto/config/*.proto
	@echo "$(GREEN)✓ Protobuf generation complete$(NC)"

# Install dependencies
deps:
	@echo "$(GREEN)Installing dependencies...$(NC)"
	cd $(GAME_DIR) && $(GO) mod download
	cd $(GAME_DIR) && $(GO) mod tidy
	@echo "$(GREEN)Installing development tools...$(NC)"
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install github.com/cosmtrek/air@latest
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

# Docker support
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t merchant-tails:latest .

docker-run:
	@echo "$(GREEN)Running in Docker...$(NC)"
	docker run -it --rm merchant-tails:latest

# CI/CD targets
ci-test:
	@echo "$(GREEN)Running CI tests...$(NC)"
	cd $(GAME_DIR) && $(GO) test -race -coverprofile=coverage.out ./...

ci-lint:
	@echo "$(GREEN)Running CI linting...$(NC)"
	cd $(GAME_DIR) && golangci-lint run --timeout 5m

# Version management
VERSION ?= $(shell git describe --tags --always --dirty)

version:
	@echo "$(GREEN)Version: $(VERSION)$(NC)"

release:
	@echo "$(GREEN)Building release version $(VERSION)...$(NC)"
	$(MAKE) clean
	$(MAKE) build-all LDFLAGS="-s -w -X main.Version=$(VERSION)"
	@echo "$(GREEN)✓ Release $(VERSION) built$(NC)"

.DEFAULT_GOAL := help
