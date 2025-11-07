.PHONY: help build run test clean install-deps setup-db docker-build docker-run

# Variables
BINARY_NAME=terminal-velocity
GO=go
GOFLAGS=-v
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

install-deps: ## Install Go dependencies
	$(GO) mod download
	$(GO) mod tidy

build: ## Build the server binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) cmd/server/main.go

build-tools: ## Build utility tools
	$(GO) build $(GOFLAGS) -o genmap cmd/genmap/main.go

genmap: build-tools ## Generate and preview a universe
	./genmap -systems 100 -stats

run: ## Run the server (development)
	$(GO) run $(GOFLAGS) cmd/server/main.go

test: ## Run tests
	$(GO) test -v -race -coverprofile=coverage.out ./...

coverage: test ## Show test coverage
	$(GO) tool cover -html=coverage.out

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -rf build/
	rm -f configs/ssh_host_key*

setup-db: ## Set up PostgreSQL database
	@echo "Setting up database..."
	@command -v psql >/dev/null 2>&1 || { echo "PostgreSQL client not found. Please install it."; exit 1; }
	psql -U postgres -f scripts/schema.sql

lint: ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Install from https://golangci-lint.run/"; exit 1; }
	golangci-lint run

fmt: ## Format code
	$(GO) fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	$(GO) vet ./...

# Docker targets
docker-build: ## Build Docker image
	docker build -t terminal-velocity:$(VERSION) .

docker-run: ## Run in Docker
	docker run -p 2222:2222 terminal-velocity:$(VERSION)

docker-compose-up: ## Start with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop docker-compose
	docker-compose down

# Development helpers
dev-setup: install-deps ## Complete development setup
	@echo "Setting up development environment..."
	@mkdir -p logs
	@mkdir -p configs
	@cp configs/config.example.yaml configs/config.yaml 2>/dev/null || true
	@echo "Development setup complete!"
	@echo "1. Set up PostgreSQL database: make setup-db"
	@echo "2. Edit configs/config.yaml with your settings"
	@echo "3. Run server: make run"

watch: ## Watch for changes and rebuild (requires entr)
	@command -v entr >/dev/null 2>&1 || { echo "entr not found. Install with: apt-get install entr"; exit 1; }
	find . -name '*.go' | entr -r make run

# Release build
release: ## Build release binary
	mkdir -p build
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o build/$(BINARY_NAME)-linux-amd64 cmd/server/main.go
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o build/$(BINARY_NAME)-linux-arm64 cmd/server/main.go
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o build/$(BINARY_NAME)-darwin-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o build/$(BINARY_NAME)-darwin-arm64 cmd/server/main.go
	@echo "Release binaries built in build/"
