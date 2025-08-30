.PHONY: build run test clean deps lint fmt docker-build docker-run

APP_NAME=polyagent-server
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)

build: deps
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

run: build
	@echo "Running $(APP_NAME)..."
	@./$(BUILD_DIR)/$(APP_NAME)

dev:
	@echo "Running in development mode..."
	@air -c .air.toml

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

lint:
	@echo "Running linter..."
	@golangci-lint run

fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@go mod tidy

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

docker-build:
	@echo "Building Docker image..."
	@docker build -t polyagent:latest .

docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env polyagent:latest

install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

generate:
	@echo "Running go generate..."
	@go generate ./...

mod-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

security-check:
	@echo "Running security checks..."
	@gosec ./...