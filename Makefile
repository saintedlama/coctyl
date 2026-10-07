.DEFAULT_GOAL := help

BINARY_NAME := coctyl
ifeq ($(OS),Windows_NT)
	BINARY_EXT := .exe
else
	BINARY_EXT :=
endif
BINARY := bin/$(BINARY_NAME)$(BINARY_EXT)

.PHONY: help build run vet test test-coverage fmt clean

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the coctyl CLI binary into bin/
	go build -o $(BINARY) ./cmd/coctyl

run: ## Run coctyl CLI (use ARGS="hash file.go")
	go run ./cmd/coctyl $(ARGS)

vet: ## Run go vet across all packages
	go vet ./...

test: ## Run unit and integration tests
	go test -v ./...

test-coverage: ## Run tests and print statement coverage breakdown
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt: ## Format source files using go fmt
	go fmt ./...

clean: ## Remove build artifacts and temporary files
	go clean
	@rm -rf bin coverage.out
