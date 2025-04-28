# Copyright (C) 2023-present

TARGET := $(shell uname -s)
GO     := GO111MODULE=on go

ifeq ("$(TARGET)", "Darwin")
	ARCH := darwin
endif

ifeq ("$(TARGET)", "Linux")
	ARCH := linux
endif

ifeq ("$(os)", "darwin")
	TARGET := Darwin
	ARCH   := darwin
endif

ifeq ("$(os)", "linux")
	TARGET := Linux
	ARCH   := linux
endif

ifeq ("$(os)", "windows")
	TARGET := Windows
	ARCH   := windows
	EXT    := .exe
endif

GOBUILD = CGO_ENABLED=0 GOOS=$(ARCH) GOARCH=amd64 $(GO) build -ldflags "-s -w"
BIN     = ./bin
CMD     = ./cmd
.PHONY: info all clean go-mcp-k8s windows-build

default: all

all: info clean go-mcp-k8s

info:
	@echo ---Building go-mcp-k8s for $(TARGET)...

windows-build:
	go build -ldflags "-s -w" -o go-mcp-k8s.exe $(CMD)/server/main.go
	@echo "Build $@ successfully!"

go-mcp-k8s: info
	@$(GOBUILD) -o $(BIN)/$@$(EXT) $(CMD)/server/main.go
	@echo "Build $@ successfully!"

clean:
	@rm -f $(BIN)/*
	@echo "Clean successfully!"
