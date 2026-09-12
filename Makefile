.DEFAULT_GOAL := help

APP_NAME      ?= petstore
BINARY_DIR    ?= bin
BINARY_NAME   ?= $(BINARY_DIR)/$(APP_NAME)
MAIN_FILE     ?= main.go
DOCKER_IMAGE  ?= go-server-petstore:latest

# Tooling
GO            ?= go
GOLANGCI_LINT ?= golangci-lint

# Colors for terminal output
CYAN   := \033[36m
GREEN  := \033[32m
YELLOW := \033[33m
RESET  := \033[0m

## help: Display this help message with descriptions of each target
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## [a-zA-Z_-]+:.*?$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ": "}; {printf "  $(CYAN)%-18s$(RESET) %s\n", substr($$1, 4), $$2}'

## build: Build binary for current platform into bin/
.PHONY: build
build:
	@echo "$(CYAN)Building binary $(BINARY_NAME)...$(RESET)"
	@mkdir -p $(BINARY_DIR)
	$(GO) build -v -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)Build successful: $(BINARY_NAME)$(RESET)"

## build-linux: Cross-compile static Linux binary (for minimal containers)
.PHONY: build-linux
build-linux:
	@echo "$(CYAN)Building static Linux binary...$(RESET)"
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -v -ldflags="-w -s" -o $(BINARY_NAME)-linux $(MAIN_FILE)
	@echo "$(GREEN)Static Linux build successful: $(BINARY_NAME)-linux$(RESET)"

## run: Run the service locally using go run
.PHONY: run
run:
	@echo "$(CYAN)Starting server locally...$(RESET)"
	$(GO) run $(MAIN_FILE)

## test: Run unit and security regression tests
.PHONY: test
test:
	@echo "$(CYAN)Running tests...$(RESET)"
	$(GO) test -v ./...

## test-race: Run tests with race condition detector
.PHONY: test-race
test-race:
	@echo "$(CYAN)Running tests with race detector...$(RESET)"
	$(GO) test -v -race ./...

## test-coverage: Run tests and generate HTML coverage report
.PHONY: test-coverage
test-coverage:
	@echo "$(CYAN)Running tests with coverage analysis...$(RESET)"
	@mkdir -p $(BINARY_DIR)
	$(GO) test -v -coverprofile=$(BINARY_DIR)/coverage.out ./...
	$(GO) tool cover -func=$(BINARY_DIR)/coverage.out
	@echo "$(YELLOW)Generating HTML coverage report at $(BINARY_DIR)/coverage.html...$(RESET)"
	$(GO) tool cover -html=$(BINARY_DIR)/coverage.out -o $(BINARY_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(BINARY_DIR)/coverage.html$(RESET)"

## lint: Run golangci-lint static analysis across all packages
.PHONY: lint
lint:
	@echo "$(CYAN)Running golangci-lint...$(RESET)"
	$(GOLANGCI_LINT) run ./...

## lint-fix: Run golangci-lint and automatically apply safe fixes
.PHONY: lint-fix
lint-fix:
	@echo "$(CYAN)Running golangci-lint with auto-fix...$(RESET)"
	$(GOLANGCI_LINT) run --fix ./...

## lint-config: Verify .golangci.yml schema
.PHONY: lint-config
lint-config:
	@echo "$(CYAN)Verifying linter configuration...$(RESET)"
	$(GOLANGCI_LINT) config verify

## fmt: Format Go code with gofmt
.PHONY: fmt
fmt:
	@echo "$(CYAN)Formatting Go code...$(RESET)"
	gofmt -s -w .

## tidy: Tidy and verify go.mod and go.sum dependencies
.PHONY: tidy
tidy:
	@echo "$(CYAN)Tidying Go modules...$(RESET)"
	$(GO) mod tidy
	$(GO) mod verify

## check: Run full validation suite (fmt, tidy, lint, test-race, build)
.PHONY: check
check: fmt tidy lint test-race build
	@echo "$(GREEN)All CI checks passed successfully!$(RESET)"

## clean: Remove build artifacts and temporary files
.PHONY: clean
clean:
	@echo "$(CYAN)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BINARY_DIR) coverage.out
	@echo "$(GREEN)Clean complete.$(RESET)"

## docker-build: Build Docker container image
.PHONY: docker-build
docker-build:
	@echo "$(CYAN)Building Docker image $(DOCKER_IMAGE)...$(RESET)"
	docker build -t $(DOCKER_IMAGE) .

## docker-run: Run Docker container locally
.PHONY: docker-run
docker-run:
	@echo "$(CYAN)Running Docker container $(DOCKER_IMAGE)...$(RESET)"
	docker run --rm -p 8080:8080 -e SERVER_ADDR=0.0.0.0:8080 $(DOCKER_IMAGE)

## compose-up: Start full stack (MongoDB + Petstore API) via Docker Compose
.PHONY: compose-up
compose-up:
	@echo "$(CYAN)Starting services with Docker Compose...$(RESET)"
	docker compose up -d --build
	@echo "$(GREEN)Services started. API listening on http://localhost:8080$(RESET)"

## compose-down: Stop and clean up Docker Compose services
.PHONY: compose-down
compose-down:
	@echo "$(CYAN)Stopping Docker Compose services...$(RESET)"
	docker compose down
	@echo "$(GREEN)Services stopped.$(RESET)"
