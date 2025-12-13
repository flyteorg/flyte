# Queue Service Makefile
# This Makefile includes common Go targets from ../go.Makefile

# Service configuration
SERVICE_NAME := queue-service
BIN_NAME := queue-service
CMD_PATH := cmd/main.go
RUN_ARGS ?= serve --config config.yaml

# Include common Go targets
include ../go.Makefile

##@ Queue Specific

.PHONY: run-testclient
run-testclient: ## Run the test client
	@echo "Running test client..."
	@go run testclient/main.go

##@ Docker - PostgreSQL

.PHONY: docker-postgres
docker-postgres: ## Start PostgreSQL in Docker
	@echo "Starting PostgreSQL container..."
	@docker run --name flyte-queue-postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=flyte_queue \
		-p 5432:5432 \
		-d postgres:15

.PHONY: docker-postgres-stop
docker-postgres-stop: ## Stop and remove PostgreSQL container
	@echo "Stopping PostgreSQL container..."
	@docker stop flyte-queue-postgres || true
	@docker rm flyte-queue-postgres || true

##@ Integration Tests

.PHONY: integration-test
integration-test: docker-postgres-stop docker-postgres ## Run integration test with PostgreSQL
	@echo "Waiting for PostgreSQL to start..."
	@sleep 3
	@echo "Building service..."
	@$(MAKE) build-fast
	@echo "Running service..."
	@./bin/queue-service serve --config config.yaml & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Running client tests..."; \
	go run testclient/main.go; \
	kill $$SERVER_PID; \
	docker stop flyte-queue-postgres; \
	docker rm flyte-queue-postgres
