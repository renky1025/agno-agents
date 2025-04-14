# Windows Makefile for Golang project

# 变量定义
BINARY_NAME=main.exe
GO=go
GOFLAGS=-v
BUILD_DIR=build
SRC_DIR=.

# 默认目标
.PHONY: all
all: clean build

# 构建应用
.PHONY: build
build:
	@echo 正在构建应用...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)/main.go
	@echo 构建完成: $(BUILD_DIR)/$(BINARY_NAME)

# 运行应用
.PHONY: run
run: build
	@echo 正在运行应用...
	$(BUILD_DIR)/$(BINARY_NAME)

# 测试
.PHONY: test
test:
	@echo 运行测试...
	$(GO) test ./...

# 清理构建文件
.PHONY: clean
clean:
	@echo 清理构建文件...
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
	@if exist $(BINARY_NAME) del /f /q $(BINARY_NAME)

# 安装依赖
.PHONY: deps
deps:
	@echo 安装依赖...
	$(GO) mod download

# 更新依赖
.PHONY: update-deps
update-deps:
	@echo 更新依赖...
	$(GO) get -u ./...

# 格式化代码
.PHONY: fmt
fmt:
	@echo 格式化代码...
	$(GO) fmt ./...

# 检查代码
.PHONY: vet
vet:
	@echo 检查代码...
	$(GO) vet ./...

# 帮助信息
.PHONY: help
help:
	@echo 可用的命令:
	@echo   make build      - 构建应用
	@echo   make run        - 构建并运行应用
	@echo   make test       - 运行测试
	@echo   make clean      - 清理构建文件
	@echo   make deps       - 安装依赖
	@echo   make update-deps - 更新依赖
	@echo   make fmt        - 格式化代码
	@echo   make vet        - 检查代码
	@echo   make help       - 显示帮助信息 