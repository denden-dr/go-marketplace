# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database connection string
DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_PATH=internal/database/migrations


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
	go build -o go-marketplace main.go

run: ## Run the application
	@echo "Running application..."
	go run main.go

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


	@echo "Setting up social test users in emulator..."
	@go run scratch/setup_social_users.go

# Example to generate a token:
	@go run scratch/gen_token.go -provider=$(provider)
