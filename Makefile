.PHONY: help run test test-integration test-all clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the server locally
	go run cmd/api/main.go

test: ## Run unit tests
	go test ./internal/... -v

test-integration: ## Run integration tests
	go test ./tests/integration/... -v

test-all: ## Run all tests (unit + integration)
	go test ./... -v

clean: ## Clean build artifacts and coverage files
	rm -f main

.DEFAULT_GOAL := help
