BINARY_NAME=api
BUILD_DIR=bin
CMD_DIR=cmd/api

.PHONY: all build run test clean tidy fmt

all: build

build:
	@echo "Building binary..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

run:
	@echo "Running application..."
	go run ./$(CMD_DIR)

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

tidy:
	@echo "Tidying go modules..."
	go mod tidy

fmt:
	@echo "Formatting code..."
	go fmt ./...
