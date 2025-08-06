.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Variables
GO_VERSION := 1.24
GODOT_VERSION := 4.4.1
GAME_DIR := game
GODOT_DIR := godot
PROTO_DIR := proto
BUILD_DIR := build

# Go commands
GO := go
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOMOD := $(GO) mod
GOFMT := gofmt
GOVET := $(GO) vet

# Godot commands
GODOT := godot

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# Development
.PHONY: dev
dev: ## Run in development mode with hot reload
	@echo "$(GREEN)Starting development server...$(NC)"
	cd $(GAME_DIR) && air

.PHONY: run
run: ## Run the game
	@echo "$(GREEN)Running the game...$(NC)"
	$(GODOT) --path $(GODOT_DIR)

# Testing
.PHONY: test
test: test-go ## Run all tests

.PHONY: test-go
test-go: ## Run Go tests
	@echo "$(GREEN)Running Go tests...$(NC)"
	cd $(GAME_DIR) && $(GOTEST) -v -race -coverprofile=coverage.out ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(NC)"
	cd $(GAME_DIR) && $(GOTEST) -v -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	cd tests/integration && $(GOTEST) -v -tags=integration ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	cd $(GAME_DIR) && $(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	cd $(GAME_DIR) && $(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: game/coverage.html$(NC)"

.PHONY: test-watch
test-watch: ## Run tests in watch mode for TDD
	@echo "$(GREEN)Starting test watcher for TDD...$(NC)"
	@which watchexec > /dev/null || (echo "$(RED)Please install watchexec: brew install watchexec$(NC)" && exit 1)
	watchexec -w $(GAME_DIR) -e go -- make test-unit

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	cd $(GAME_DIR) && $(GOTEST) -bench=. -benchmem ./...

# Building
.PHONY: build
build: build-go build-godot ## Build everything

.PHONY: build-go
build-go: ## Build Go GDExtension
	@echo "$(GREEN)Building Go GDExtension...$(NC)"
	@mkdir -p $(GODOT_DIR)/bin
	cd $(GAME_DIR) && CGO_ENABLED=1 $(GOBUILD) -buildmode=c-shared \
		-o ../$(GODOT_DIR)/bin/merchant_game.so \
		./cmd/gdextension/main.go

.PHONY: build-godot
build-godot: ## Build Godot project
	@echo "$(GREEN)Building Godot project...$(NC)"
	@mkdir -p $(BUILD_DIR)/windows $(BUILD_DIR)/linux $(BUILD_DIR)/mac
	@if [ -f "$(GODOT_DIR)/project.godot" ]; then \
		$(GODOT) --headless --export-release "Windows Desktop" $(BUILD_DIR)/windows/merchant_tails.exe; \
		$(GODOT) --headless --export-release "Linux/X11" $(BUILD_DIR)/linux/merchant_tails; \
		$(GODOT) --headless --export-release "macOS" $(BUILD_DIR)/mac/merchant_tails.app; \
	else \
		echo "$(YELLOW)Godot project not found, skipping Godot build$(NC)"; \
	fi

.PHONY: build-all
build-all: build ## Build for all platforms

.PHONY: build-windows
build-windows: ## Build for Windows
	@echo "$(GREEN)Building for Windows...$(NC)"
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) -buildmode=c-shared \
		-o $(GODOT_DIR)/bin/merchant_game.dll \
		$(GAME_DIR)/cmd/gdextension/main.go

.PHONY: build-mac
build-mac: ## Build for macOS
	@echo "$(GREEN)Building for macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)/mac
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) -buildmode=c-shared \
		-o $(GODOT_DIR)/bin/merchant_game.dylib \
		$(GAME_DIR)/cmd/gdextension/main.go

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "$(GREEN)Building for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) -buildmode=c-shared \
		-o $(GODOT_DIR)/bin/merchant_game.so \
		$(GAME_DIR)/cmd/gdextension/main.go

# Code Quality
.PHONY: fmt
fmt: ## Format Go code
	@echo "$(GREEN)Formatting Go code...$(NC)"
	cd $(GAME_DIR) && $(GOFMT) -w -s .
	@echo "$(GREEN)Code formatted$(NC)"

.PHONY: lint
lint: ## Run linters
	@echo "$(GREEN)Running linters...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(RED)Please install golangci-lint$(NC)" && exit 1)
	cd $(GAME_DIR) && golangci-lint run --timeout 5m

.PHONY: vet
vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	cd $(GAME_DIR) && $(GOVET) ./...

.PHONY: check
check: fmt vet lint test ## Run all checks

# Protocol Buffers
.PHONY: proto
proto: ## Generate protobuf files
	@echo "$(GREEN)Generating protobuf files...$(NC)"
	@which protoc > /dev/null || (echo "$(RED)Please install protoc$(NC)" && exit 1)
	protoc --go_out=$(GAME_DIR) --go_opt=paths=source_relative \
		$(PROTO_DIR)/**/*.proto

# Dependencies
.PHONY: deps
deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	cd $(GAME_DIR) && $(GOMOD) download
	cd $(GAME_DIR) && $(GOMOD) tidy

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	cd $(GAME_DIR) && $(GOMOD) get -u ./...
	cd $(GAME_DIR) && $(GOMOD) tidy

.PHONY: tools
tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install golang.org/x/tools/cmd/godoc@latest
	@echo "$(GREEN)Tools installed$(NC)"

# Documentation
.PHONY: docs
docs: ## Generate documentation
	@echo "$(GREEN)Generating documentation...$(NC)"
	cd $(GAME_DIR) && godoc -http=:6060 &
	@echo "$(GREEN)Documentation server started at http://localhost:6060$(NC)"

.PHONY: docs-api
docs-api: ## Generate API documentation
	@echo "$(GREEN)Generating API documentation...$(NC)"
	cd $(GAME_DIR) && $(GO) doc -all ./... > ../docs/api/api.md

# Cleanup
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -rf $(GODOT_DIR)/bin
	rm -f $(GAME_DIR)/coverage.out $(GAME_DIR)/coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

.PHONY: clean-cache
clean-cache: ## Clean Go module cache
	@echo "$(GREEN)Cleaning Go module cache...$(NC)"
	$(GO) clean -modcache

# Git hooks
.PHONY: install-hooks
install-hooks: ## Install git hooks
	@echo "$(GREEN)Installing git hooks...$(NC)"
	@echo '#!/bin/sh\nmake check' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)Git hooks installed$(NC)"

# Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t merchant-tails:latest .

.PHONY: docker-run
docker-run: ## Run in Docker container
	@echo "$(GREEN)Running in Docker container...$(NC)"
	docker run -it --rm -p 8080:8080 merchant-tails:latest

# Info
.PHONY: info
info: ## Show project information
	@echo "$(GREEN)Project Information:$(NC)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Godot Version: $(GODOT_VERSION)"
	@echo "  Game Directory: $(GAME_DIR)"
	@echo "  Godot Directory: $(GODOT_DIR)"
	@echo ""
	@echo "$(GREEN)Go Module Information:$(NC)"
	@cd $(GAME_DIR) && $(GO) list -m all | head -5
	@echo ""
	@echo "$(GREEN)Test Coverage:$(NC)"
	@if [ -f "$(GAME_DIR)/coverage.out" ]; then \
		cd $(GAME_DIR) && $(GO) tool cover -func=coverage.out | tail -1; \
	else \
		echo "  No coverage data available. Run 'make test-coverage' first."; \
	fi

# Default target
.DEFAULT_GOAL := help
