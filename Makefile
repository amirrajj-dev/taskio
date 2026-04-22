.PHONY: help run build dev prod test clean migrate swagger docker-up-dev docker-up-prod docker-down

# Variables
BINARY_NAME=taskio
ENV_FILE?=.env.dev

help:
	@echo "Available commands:"
	@echo "  make run            - Run application (dev mode)"
	@echo "  make build          - Build binary"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo "  make prod           - Run production binary"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make swagger        - Generate swagger docs"
	@echo "  make docker-up-dev  - Start dev dependencies (postgres,redis,rabbitmq,mailhog)"
	@echo "  make docker-up-prod - Start full production stack"
	@echo "  make docker-down    - Stop all docker containers"

run:
	ENV_FILE=$(ENV_FILE) go run ./cmd/api

build:
	go build -o $(BINARY_NAME) ./cmd/api

dev:
	air

prod: build
	ENV_FILE=.env.prod ./$(BINARY_NAME)

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf tmp/

swagger:
	swag init --parseDependency --parseInternal --output ./docs --dir ./cmd/api,./internal/handlers

docker-up-dev:
	docker compose -f docker-compose.dev.yml --env-file .env.dev up -d

docker-up-prod:
	docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

docker-down:
	docker compose -f docker-compose.dev.yml -f docker-compose.prod.yml down