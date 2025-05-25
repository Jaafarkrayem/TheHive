# Hexagonal Chain Development Makefile

.PHONY: help build test clean docker-build docker-up docker-down deps fmt lint vet geth-clone geth-analyze

# Default target
help: ## Show this help message
	@echo "Hexagonal Chain Development Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the hexnode binary
	@echo "Building hexnode..."
	@go build -o bin/hexnode ./cmd/hexnode

build-all: ## Build all binaries
	@echo "Building all binaries..."
	@go build -o bin/hexnode ./cmd/hexnode
	@go build -o bin/hex-indexer ./pkg/hexindexer

# Development
dev: ## Run development environment
	@echo "Starting development environment..."
	@docker-compose up hexnode-dev postgres redis

# Testing
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Code quality
fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t hexagonal-chain:latest .

docker-up: ## Start all services with Docker Compose
	@echo "Starting all services..."
	@docker-compose up -d

docker-down: ## Stop all Docker services
	@echo "Stopping all services..."
	@docker-compose down

docker-logs: ## Show Docker logs
	@docker-compose logs -f

# Geth research (Phase 1.2)
geth-clone: ## Clone Geth repository for analysis
	@echo "Cloning Geth repository..."
	@if [ ! -d "research/go-ethereum" ]; then \
		mkdir -p research; \
		git clone https://github.com/ethereum/go-ethereum.git research/go-ethereum; \
	else \
		echo "Geth repository already exists"; \
	fi

geth-analyze: geth-clone ## Analyze Geth codebase structure
	@echo "Analyzing Geth codebase..."
	@find research/go-ethereum -name "*.go" -path "*/core/*" | head -20
	@echo "\nKey Geth directories:"
	@ls -la research/go-ethereum/core/
	@echo "\nBlock-related files:"
	@find research/go-ethereum -name "*block*.go" | head -10

# Clean up
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -rf coverage.out coverage.html
	@go clean

clean-all: clean docker-down ## Clean everything including Docker
	@echo "Cleaning all..."
	@docker system prune -f
	@rm -rf data/
	@rm -rf logs/

# Genesis and configuration
genesis: ## Create genesis block configuration
	@echo "Creating genesis configuration..."
	@mkdir -p internal/config
	@go run scripts/create-genesis.go

# Node operations
node-init: ## Initialize a new node
	@echo "Initializing node..."
	@./bin/hexnode init --datadir ./data/node1

node-run: build ## Run a single node
	@echo "Running hexnode..."
	@./bin/hexnode --datadir ./data/node1 --networkid 1337

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@godoc -http=:6060
	@echo "Documentation available at http://localhost:6060"

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Project setup
setup: deps install-tools geth-clone ## Complete project setup
	@echo "Project setup complete!"
	@echo "Run 'make help' to see available commands" 