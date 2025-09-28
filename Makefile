# Teable Go Backend Makefile

# 变量定义
BINARY_NAME=teable-backend
DOCKER_IMAGE=teable-go-backend
VERSION=1.0.0
BUILD_TIME=$(shell date +%Y%m%d-%H%M%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
BUILD_FLAGS=-v $(LDFLAGS)

# 默认目标
.PHONY: all
all: test build

# 构建
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(BUILD_FLAGS) -o bin/$(BINARY_NAME) ./cmd/server

# 构建(跨平台)
.PHONY: build-linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux ./cmd/server

.PHONY: build-windows
build-windows:
	@echo "Building $(BINARY_NAME) for Windows..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o bin/$(BINARY_NAME).exe ./cmd/server

.PHONY: build-darwin
build-darwin:
	@echo "Building $(BINARY_NAME) for macOS..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin ./cmd/server

# 运行
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run ./cmd/server

# 开发模式运行
.PHONY: dev
dev:
	@echo "Running in development mode..."
	air -c .air.toml

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...

# 基准测试
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# 代码质量
.PHONY: lint
lint:
	@echo "Running linters..."
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# 依赖管理
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) tidy

.PHONY: deps-verify
deps-verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html

# Docker
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 3000:3000 $(DOCKER_IMAGE):latest

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

# Docker Compose
.PHONY: up
up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

.PHONY: down
down:
	@echo "Stopping services..."
	docker-compose down

.PHONY: logs
logs:
	@echo "Showing logs..."
	docker-compose logs -f

# 数据库
.PHONY: db-up
db-up:
	@echo "Starting database services..."
	docker-compose up -d postgres redis

.PHONY: db-down
db-down:
	@echo "Stopping database services..."
	docker-compose stop postgres redis

.PHONY: db-reset
db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres redis

# 生成
.PHONY: generate
generate:
	@echo "Running go generate..."
	$(GOCMD) generate ./...

# API文档
.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go -o api/swagger

# 工具安装
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) -u github.com/cosmtrek/air@latest
	$(GOGET) -u github.com/swaggo/swag/cmd/swag@latest
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-linux    - Build for Linux"
	@echo "  build-windows  - Build for Windows"
	@echo "  build-darwin   - Build for macOS"
	@echo "  run            - Run the application"
	@echo "  dev            - Run in development mode with hot reload"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  test-race      - Run tests with race detection"
	@echo "  benchmark      - Run benchmarks"
	@echo "  lint           - Run linters"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  deps           - Download dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  deps-verify    - Verify dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-push    - Push Docker image"
	@echo "  up             - Start services with docker-compose"
	@echo "  down           - Stop services"
	@echo "  logs           - Show logs"
	@echo "  db-up          - Start database services"
	@echo "  db-down        - Stop database services"
	@echo "  db-reset       - Reset database"
	@echo "  generate       - Run go generate"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  install-tools  - Install development tools"
	@echo "  help           - Show this help message"