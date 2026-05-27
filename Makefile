.PHONY: build run restart stop logs clean test integration lint

COMPOSE ?= docker compose
TOOLS   := $(COMPOSE) --profile tools run --rm tools

build:
	$(COMPOSE) build goboxd

run:
	$(COMPOSE) up -d goboxd

restart:
	$(COMPOSE) restart goboxd

stop:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f goboxd

clean:
	$(COMPOSE) down -v --rmi local

test:
	$(TOOLS) go test ./...

integration:
	$(TOOLS) go test -tags=integration ./tests/...

lint:
	$(TOOLS) golangci-lint run ./..