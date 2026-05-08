# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Container engine auto-detection
DOCKER_BIN := $(shell which podman 2>/dev/null || which docker 2>/dev/null)
DOCKER_COMPOSE := $(shell if $(DOCKER_BIN) compose version >/dev/null 2>&1; then echo "$(DOCKER_BIN) compose"; else echo "docker-compose"; fi)

# Database connection string
DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_PATH=internal/database/migrations


docker-up: ## Start containers in detached mode
	@echo "Starting containers..."
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop and remove containers
	@echo "Stopping containers..."
	$(DOCKER_COMPOSE) down

docker-logs: ## Follow container logs
	@echo "Following logs..."
	$(DOCKER_COMPOSE) logs -f

docker-shell: ## Interactive shell in the API container
	@echo "Opening shell..."
	$(DOCKER_BIN) exec -it marketplace-api sh

test-docker: ## Run tests in Docker containers
	@echo "Running tests in Docker..."
	$(DOCKER_COMPOSE) -f docker-compose.test.yaml up --build --abort-on-container-exit --exit-code-from test-runner
	$(DOCKER_COMPOSE) -f docker-compose.test.yaml down -v




swagger: ## Generate swagger documentation
	@echo "Generating swagger..."
	swag init -g cmd/api/main.go --parseDependency --parseInternal

test: ## Run all tests
	@echo "Running tests..."
	go test ./... -v -count=1

help: ## Show this help message
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	go build -o bin/go-marketplace cmd/api/main.go

run: ## Run the application
	@echo "Running application..."
	go run cmd/api/main.go

clean: ## Remove the compiled binary
	@echo "Cleaning up..."
	rm -f go-marketplace

fmt: ## Format the code
	@echo "Formatting code..."
	go fmt ./...

tidy: ## Tidy up go modules
	@echo "Tidying modules..."
	go mod tidy

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down

migrate-create: ## Create a new migration file (usage: make migrate-create name=name_of_migration)
	@echo "Creating migration..."
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

mock: ## Generate mocks using mockery
	@echo "Generating mocks..."
	mockery
