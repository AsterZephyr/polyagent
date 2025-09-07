.PHONY: build build-rec run run-rec test clean deps lint fmt docker-build docker-run

APP_NAME=polyagent-server
REC_APP_NAME=recommendation-agent-server
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)

# 构建通用Agent服务器 (包含对话、工作流等功能)
build: deps
	@echo "Building $(APP_NAME) (General Agent Server)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

# 构建推荐业务专用服务器 (推荐)
build-rec: deps
	@echo "Building $(REC_APP_NAME) (Recommendation Agent Server)..."
	@go build -o $(BUILD_DIR)/$(REC_APP_NAME) cmd/server/recommendation_server.go

run: build
	@echo "Running $(APP_NAME) (General Agent Server)..."
	@./$(BUILD_DIR)/$(APP_NAME)

# 运行推荐业务专用服务器 (推荐)
run-rec: build-rec
	@echo "Running $(REC_APP_NAME) (Recommendation Agent Server)..."
	@./$(BUILD_DIR)/$(REC_APP_NAME)

dev:
	@echo "Running in development mode..."
	@air -c .air.toml

test:
	@echo "Running tests..."
	@go test -v ./...

# 运行推荐业务测试 (专用)
test-rec:
	@echo "Running recommendation system tests..."
	@go test -v ./internal/recommendation

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