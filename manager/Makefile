.PHONY: build run clean test

.PHONY: help
help: ## Show this help message
	@echo '🆘  Showing help message'
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the unified manager binary
build:
	go build -o bin/flyte-manager cmd/main.go

run: # Run the manager
	go run cmd/main.go --config config.yaml

run-testclient: # Run the manager
	go run testclient/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f flyte.db

# Test the manager
test:
	go test ./...

# Build and install
install: build
	cp bin/flyte-manager /usr/local/bin/

.DEFAULT_GOAL := build
