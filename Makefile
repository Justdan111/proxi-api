# ============================================================
#  Proxi Backend — Makefile
# ============================================================

APP_NAME   = proxi-api
BINARY     = bin/server
MAIN       = ./cmd/server
DOCKER_IMG = proxi-api:latest

.PHONY: help build run dev test clean docker-build docker-up docker-down lint tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ---- Local Development ----

build: ## Compile the binary
	@mkdir -p bin
	@echo "Building $(BINARY)..."
	go build -o $(BINARY) $(MAIN)
	@echo "✅ Done"

run: build ## Build and run locally
	./$(BINARY)

dev: ## Run with hot reload (requires: go install github.com/air-verse/air@latest)
	air

tidy: ## Tidy go modules
	go mod tidy

lint: ## Run linter (requires: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

test: ## Run all tests
	go test -v -race -cover ./...

test-cover: ## Run tests with HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html

# ---- Docker ----

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMG) .

docker-up: ## Start all services (API + MongoDB + Mongo Express)
	docker-compose up --build

docker-up-d: ## Start all services in background
	docker-compose up --build -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## Follow API logs
	docker-compose logs -f api

docker-shell: ## Open shell inside API container
	docker-compose exec api sh

# ---- Production ----

docker-prod-up: ## Start production compose
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# ---- Database ----

mongo-shell: ## Open MongoDB shell
	docker-compose exec mongo mongosh proxi

# ---- Helpers ----

health: ## Check API health
	curl -s http://localhost:8080/health | jq