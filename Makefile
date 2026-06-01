.PHONY: build run restart stop logs clean test integration lint load help

# Docker Compose command
COMPOSE ?= docker compose

# Tools container for testing and linting
TOOLS := $(COMPOSE) --profile tools run --rm tools

# ======================
# Main Targets
# ======================

# Build the goboxd Docker image
build:
	$(COMPOSE) build goboxd

# Run the service in detached mode
run:
	$(COMPOSE) up -d goboxd

# Restart the service
restart:
	$(COMPOSE) restart goboxd

# Stop all services
stop:
	$(COMPOSE) down

# View logs
logs:
	$(COMPOSE) logs -f goboxd

# Clean everything (containers, volumes, images)
clean:
	$(COMPOSE) down -v --rmi local
	docker system prune -f --filter "label=project=goboxd"

# ======================
# Testing & Quality
# ======================

# Unit tests (fast, no Docker dependencies required)
test:
	$(TOOLS) go test ./... -short -count=1

# Stage 1 Integration Tests (most important right now)
integration:
	@echo "=== Running Stage 1 Integration Tests ==="
	$(TOOLS) go clean -testcache
	$(TOOLS) go test -v ./tests/integration -run TestRunEndpoint_Stage1 -count=1

# Run all tests (unit + integration)
test-all: test integration

# Lint the codebase
lint:
	$(TOOLS) env GOFLAGS=-buildvcs=false golangci-lint run --timeout=5m ./...

# Basic load test (using hey - install via tools if needed)
load:
	@echo "=== Running basic load test (100 requests, 10 concurrent) ==="
	@$(TOOLS) hey -n 100 -c 10 -m POST \
		-H "Content-Type: application/json" \
		-d '{"language":"py3","source":"print(\"hello from load test\")","tests":[{"stdin":"","expected_stdout":"hello from load test"}]}' \
		http://goboxd:8080/run

# ======================
# Development Helpers
# ======================

# Full rebuild + restart (useful during development)
rebuild: build restart

# Run everything fresh (clean + build + run)
fresh: clean build run

# Show this help message
help:
	@echo "goboxd Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build Docker image"
	@echo "  run            - Start service in background"
	@echo "  restart        - Restart service"
	@echo "  stop           - Stop all services"
	@echo "  logs           - Follow logs"
	@echo "  clean          - Full cleanup"
	@echo "  test           - Run unit tests"
	@echo "  integration    - Run integration tests"
	@echo "  lint           - Run golangci-lint"
	@echo "  load           - Basic load test"
	@echo "  rebuild        - Rebuild + restart"
	@echo "  fresh          - Clean + build + run"
	@echo "  help           - Show this help"

# ======================
# Docker Compose Profile Setup (ensure tools exist)
# ======================

# Ensure tools container is ready (called automatically if needed)
tools-ready:
	@$(COMPOSE) build --quiet tools 2>/dev/null || true