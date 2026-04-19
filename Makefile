# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database connection string
DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_PATH=internal/database/migrations

.PHONY: help build run clean fmt tidy migrate-up migrate-down migrate-create test mock swagger firebase-emulator firebase-setup

swagger: ## Generate swagger documentation
	@echo "Generating swagger..."
	swag init -g main.go --parseDependency --parseInternal

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
	go build -o go-shop-yourself main.go

run: ## Run the application
	@echo "Running application..."
	go run main.go

clean: ## Remove the compiled binary
	@echo "Cleaning up..."
	rm -f go-shop-yourself

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

firebase-emulator: ## Run Firebase Auth Emulator
	@echo "Starting Firebase Auth Emulator..."
	firebase emulators:start --only auth --project $(FIREBASE_PROJECT_ID)

firebase-setup: ## Setup initial test users (Google, FB, Apple, Twitter) in Firebase Emulator
	@echo "Setting up social test users in emulator..."
	@go run scratch/setup_social_users.go

# Example to generate a token:
# make gen-token provider=facebook.com
gen-token: ## Generate a mock Firebase ID token (usage: make gen-token provider=google.com)
	@go run scratch/gen_token.go -provider=$(provider)
