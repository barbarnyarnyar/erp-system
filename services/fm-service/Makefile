# Variables
APP_NAME=fm-service
DOCKER_IMAGE=fm-service
DOCKER_TAG=latest
GO_VERSION=1.21

# Build commands
.PHONY: build
build:
	go build -o bin/$(APP_NAME) cmd/server/main.go

.PHONY: run
run:
	go run cmd/server/main.go

.PHONY: test
test:
	go test ./...

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: lint
lint:
	golangci-lint run

# Docker commands
.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run:
	docker run -p 8001:8001 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Development commands
.PHONY: dev
dev:
	air

.PHONY: deps
deps:
	go mod tidy
	go mod download

# Database commands
.PHONY: migrate-up
migrate-up:
	migrate -path internal/data/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

.PHONY: migrate-down
migrate-down:
	migrate -path internal/data/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir internal/data/migrations $(name)

# Clean commands
.PHONY: clean
clean:
	rm -rf bin/
	go clean -cache

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run linter"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  dev           - Run in development mode with air"
	@echo "  deps          - Download dependencies"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  migrate-create - Create new migration file"
	@echo "  clean         - Clean build files"