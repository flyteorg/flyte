# Manager Service Makefile
# This Makefile includes common Go targets from ../go.Makefile

# Service configuration
SERVICE_NAME := flyte-manager
BIN_NAME := flyte-manager
CMD_PATH := cmd/main.go
RUN_ARGS ?= --config config.yaml

# Include common Go targets
include ../go.Makefile

##@ Manager Specific

.PHONY: run-testclient
run-testclient: ## Run the test client
	@echo "Running test client..."
	@go run testclient/main.go

.PHONY: clean-db
clean-db: ## Remove local database file
	@echo "Removing flyte.db..."
	@rm -f flyte.db

# Override clean to also remove database
clean: clean-db ## Remove build artifacts and database
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
