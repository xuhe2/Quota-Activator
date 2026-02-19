# Quota-Activator Makefile

BINARY_NAME=quota-activator
BUILD_DIR=build
GO=go
GOFLAGS=-v

.PHONY: all build clean run test fmt vet lint help

# 默认目标
all: build

# 编译项目
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 安装到 $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) .
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

# 清理构建文件
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "Clean complete"

# 运行程序
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# 快速运行（不编译，直接 go run）
run-dev:
	@echo "Running with 'go run'..."
	$(GO) run .

# 运行测试
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# 测试覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# 格式化代码
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# 代码检查
vet:
	@echo "Vetting code..."
	$(GO) vet ./...

# 使用 golangci-lint 进行检查（需要先安装: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest）
lint:
	@echo "Linting code..."
	@golangci-lint run ./... || echo "golangci-lint not installed, skipping..."

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# 更新依赖
deps-update:
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  all           - 构建项目（默认）"
	@echo "  build         - 编译项目到 $(BUILD_DIR)/"
	@echo "  install       - 安装到 \$GOPATH/bin"
	@echo "  clean         - 清理构建文件"
	@echo "  run           - 编译并运行"
	@echo "  run-dev       - 使用 'go run' 直接运行"
	@echo "  test          - 运行测试"
	@echo "  test-coverage - 运行测试并生成覆盖率报告"
	@echo "  fmt           - 格式化代码"
	@echo "  vet           - 代码检查"
	@echo "  lint          - 使用 golangci-lint 检查"
	@echo "  deps          - 下载依赖"
	@echo "  deps-update   - 更新依赖"
	@echo "  help          - 显示此帮助信息"
