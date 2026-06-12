.PHONY: build build-local up run restart stop logs clean test integration test-all ci lint load load-internal rebuild fresh help run-local

# Docker Compose command
COMPOSE ?= docker compose -f deployments/docker/docker-compose.yml

# Tools container for testing and linting
TOOLS := $(COMPOSE) --profile tools run --rm tools

# ======================
# Main Targets
# ======================

build:
	$(COMPOSE) build goboxd

# Build natively (faster for local development)
build-local:
	go build -o goboxd ./cmd/goboxd

# Start service using Docker
up run:
	$(COMPOSE) up -d goboxd

# Run natively (recommended for development)
run-local: build-local
	@echo "🚀 Starting goboxd locally on :8080"
	./goboxd

# Restart Docker service
restart:
	$(COMPOSE) restart goboxd

stop:
	$(COMPOSE) down --remove-orphans

logs:
	$(COMPOSE) logs -f goboxd

# ======================
# Cleanup
# ======================

clean:
	@echo "🧹 Cleaning up..."
	$(COMPOSE) down -v --rmi local --remove-orphans
	docker rm -f goboxd 2>/dev/null || true
	docker network rm goboxd_network 2>/dev/null || true
	docker system prune -f --filter "label=project=goboxd"
	rm -f goboxd
	go clean -cache -testcache

# ======================
# Testing & Quality
# ======================

test: tools-ready
	@echo "=== Running Unit Tests ==="
	$(TOOLS) go test ./tests/... -short -count=1 -race

integration: tools-ready
	@echo "=== Running Integration Tests ==="
	$(TOOLS) go clean -testcache
	$(TOOLS) go test -v ./tests/integration -count=1

test-all: test integration

ci:
	./scripts/test.sh

# Lint - Try Docker first, fallback to local
lint:
	@echo "=== Running linter ==="
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m ./...; \
	else \
		$(TOOLS) env GOFLAGS=-buildvcs=false golangci-lint run --timeout=5m ./...; \
	fi

# ======================
# Load Testing
# ======================

load:
	@echo "=== Running Load Test (Host) ==="
	@which hey || (echo "Install hey: go install github.com/rakyll/hey@latest" && exit 1)
	hey -n 1000 -c 50 -m POST -H "Content-Type: application/json" \
		-d '{"language":"py3","source":"print(\"load test\")","tests":[{"stdin":"","expected_stdout":"load test"}]}' \
		http://localhost:8080/run

load-internal: tools-ready
	@echo "=== Running Load Test (Internal) ==="
	$(TOOLS) hey -n 100 -c 10 -m POST \
		-H "Content-Type: application/json" \
		-d '{"language":"py3","source":"print(\"hello from load test\")","tests":[{"stdin":"","expected_stdout":"hello from load test"}]}' \
		http://goboxd:8080/run

# ======================
# Development Workflow
# ======================

rebuild: build restart

fresh: clean build run

# ======================
# Helpers
# ======================

tools-ready:
	@$(COMPOSE) build --quiet tools

help:
	@echo "goboxd Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build           - Build Docker image"
	@echo "  build-local     - Build native binary"
	@echo "  run             - Start service (Docker)"
	@echo "  run-local       - Start service natively (recommended for dev)"
	@echo "  stop            - Stop services"
	@echo "  logs            - Follow logs"
	@echo "  clean           - Full cleanup"
	@echo "  test            - Run unit tests"
	@echo "  integration     - Run integration tests"
	@echo "  lint            - Run golangci-lint"
	@echo "  load            - Load test from host"
	@echo "  fresh           - Clean + build + run"
	@echo "  help            - Show this help"