# PolyAgent Go Services Makefile

.PHONY: build clean run test deps docker help

# 变量定义
APP_NAME = polyagent
VERSION = $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION = $(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS = -ldflags "\
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GoVersion=${GO_VERSION}"

# Go 相关变量
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# 目录定义
BUILD_DIR = build
DIST_DIR = dist

# 默认目标
all: clean deps test build

# 安装依赖
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 构建
build: deps
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/gateway ./gateway
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/scheduler ./scheduler
	@echo "Build completed!"

# 构建所有平台
build-all: deps
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/gateway-linux-amd64 ./gateway
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/scheduler-linux-amd64 ./scheduler
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/gateway-linux-arm64 ./gateway
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/scheduler-linux-arm64 ./scheduler
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/gateway-darwin-amd64 ./gateway
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/scheduler-darwin-amd64 ./scheduler
	
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/gateway-darwin-arm64 ./gateway
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/scheduler-darwin-arm64 ./scheduler
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/gateway-windows-amd64.exe ./gateway
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/scheduler-windows-amd64.exe ./scheduler
	
	@echo "Cross-platform build completed!"

# 运行网关服务
run-gateway:
	@echo "Starting gateway service..."
	$(GOBUILD) -o $(BUILD_DIR)/gateway ./gateway && $(BUILD_DIR)/gateway

# 运行调度服务
run-scheduler:
	@echo "Starting scheduler service..."
	$(GOBUILD) -o $(BUILD_DIR)/scheduler ./scheduler && $(BUILD_DIR)/scheduler

# 运行所有服务
run-all: build
	@echo "Starting all services..."
	$(BUILD_DIR)/gateway &
	$(BUILD_DIR)/scheduler &
	wait

# 测试
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests completed!"

# 测试覆盖率
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 基准测试
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# 代码检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
		golangci-lint run; \
	fi

# 代码格式化
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# 生成模拟文件
generate:
	@echo "Generating mock files..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=internal/storage/postgres.go -destination=internal/storage/mocks/postgres_mock.go; \
		mockgen -source=internal/storage/redis.go -destination=internal/storage/mocks/redis_mock.go; \
		mockgen -source=internal/ai/client.go -destination=internal/ai/mocks/client_mock.go; \
	else \
		echo "mockgen not installed. Installing..."; \
		$(GOGET) github.com/golang/mock/mockgen@latest; \
		make generate; \
	fi

# Docker 构建
docker-build:
	@echo "Building Docker images..."
	docker build -f Dockerfile.gateway -t $(APP_NAME)-gateway:$(VERSION) .
	docker build -f Dockerfile.scheduler -t $(APP_NAME)-scheduler:$(VERSION) .
	@echo "Docker images built!"

# Docker 运行
docker-run:
	@echo "Running with Docker Compose..."
	docker-compose up --build

# Docker 停止
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# 清理构建文件
clean:
	@echo "Cleaning build files..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean completed!"

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) golang.org/x/tools/cmd/goimports@latest
	$(GOGET) github.com/golang/mock/mockgen@latest
	$(GOGET) github.com/swaggo/swag/cmd/swag@latest
	@echo "Tools installed!"

# 生成 API 文档
docs:
	@echo "Generating API documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g gateway/main.go; \
	else \
		echo "swag not installed. Installing..."; \
		$(GOGET) github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g gateway/main.go; \
	fi

# 数据库迁移
migrate-up:
	@echo "Running database migrations..."
	# 这里可以添加数据库迁移命令

migrate-down:
	@echo "Rolling back database migrations..."
	# 这里可以添加回滚命令

# 生产环境构建
build-prod: clean
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo $(LDFLAGS) -o $(BUILD_DIR)/gateway ./gateway
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo $(LDFLAGS) -o $(BUILD_DIR)/scheduler ./scheduler
	@echo "Production build completed!"

# 性能分析
profile:
	@echo "Running performance profiling..."
	$(GOTEST) -cpuprofile cpu.prof -memprofile mem.prof -bench . ./...
	@echo "Profile files generated: cpu.prof, mem.prof"

# 安全检查
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Installing..."; \
		$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# 依赖检查
deps-check:
	@echo "Checking for outdated dependencies..."
	$(GOCMD) list -u -m all

# 漏洞检查
vuln-check:
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Installing..."; \
		$(GOGET) golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

# 代码统计
stats:
	@echo "Code statistics:"
	@find . -name "*.go" | xargs wc -l | tail -1
	@find . -name "*.go" | wc -l | awk '{print $$1 " Go files"}'

# 帮助信息
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  build-all     - Build for all platforms"
	@echo "  build-prod    - Build for production"
	@echo "  clean         - Clean build files"
	@echo "  deps          - Install dependencies"
	@echo "  deps-check    - Check for outdated dependencies"
	@echo "  docs          - Generate API documentation"
	@echo "  fmt           - Format code"
	@echo "  generate      - Generate mock files"
	@echo "  help          - Show this help"
	@echo "  install-tools - Install development tools"
	@echo "  lint          - Run linter"
	@echo "  profile       - Run performance profiling"
	@echo "  run-gateway   - Run gateway service"
	@echo "  run-scheduler - Run scheduler service"
	@echo "  run-all       - Run all services"
	@echo "  security      - Run security checks"
	@echo "  stats         - Show code statistics"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Generate test coverage report"
	@echo "  vuln-check    - Check for vulnerabilities"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker containers"