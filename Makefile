.PHONY: help build run test clean install-deps install-proto-tools proto proto-clean setup-db docker-build docker-run

# Variables
BINARY_NAME=terminal-velocity
GO=go
GOFLAGS=-v
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Protobuf variables
PROTO_DIR=api/proto
PROTO_OUT_DIR=api/gen/go/v1
PROTO_FILES=$(shell find $(PROTO_DIR) -name '*.proto')

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

install-deps: ## Install Go dependencies
	$(GO) mod download
	$(GO) mod tidy

install-proto-tools: ## Install protobuf tools
	@echo "Installing protobuf tools..."
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@command -v protoc >/dev/null 2>&1 || { echo "protoc not found. Please install Protocol Buffers compiler."; exit 1; }
	@echo "Protobuf tools installed successfully"

proto: ## Generate Go code from protobuf schemas
	@echo "Generating protobuf code..."
	@mkdir -p $(PROTO_OUT_DIR)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
	@echo "Protobuf code generated in $(PROTO_OUT_DIR)"

proto-clean: ## Remove generated protobuf code
	rm -rf $(PROTO_OUT_DIR)

build: proto ## Build the server binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) cmd/server/main.go

build-tools: ## Build utility tools
	$(GO) build $(GOFLAGS) -o genmap cmd/genmap/main.go
	$(GO) build $(GOFLAGS) -o accounts cmd/accounts/main.go

genmap: build-tools ## Generate and preview a universe
	./genmap -systems 100 -stats

# Docker targets
docker-build: ## Build Docker image
	docker build -t terminal-velocity:latest .

docker-run: ## Run server in Docker
	docker run -p 2222:2222 terminal-velocity:latest

docker compose-up: ## Start full stack with docker compose
	docker compose up -d

docker compose-down: ## Stop docker compose stack
	docker compose down

docker compose-logs: ## View docker compose logs
	docker compose logs -f

docker compose-restart: ## Restart docker compose stack
	docker compose restart

docker-clean: ## Remove all Docker artifacts
	docker compose down -v
	docker system prune -f

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
	rm -rf $(PROTO_OUT_DIR)
	rm -f configs/ssh_host_key* data/ssh_host_key*

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

docker compose-up: ## Start with docker compose
	docker compose up -d

docker compose-down: ## Stop docker compose
	docker compose down

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
