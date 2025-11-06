# Runs Service Makefile
# This Makefile includes common Go targets from ../go.Makefile

# Service configuration
SERVICE_NAME := runs-service
BIN_NAME := runs-service
CMD_PATH := cmd/main.go
RUN_ARGS ?= --config config.yaml

# Include common Go targets
include ../go.Makefile

##@ Runs Specific

.PHONY: build-client
build-client: ## Build the test client
	@echo "Building test client..."
	@mkdir -p bin
	@go build -o bin/runs-client testclient/main.go

.PHONY: run-postgres
run-postgres: build-fast ## Run the service with PostgreSQL config
	@echo "Running $(SERVICE_NAME) with PostgreSQL..."
	@./$(BIN_DIR)/$(BIN_NAME) --config config-postgres.yaml

.PHONY: run-testclient
run-testclient: build-client ## Run the test client
	@echo "Running test client..."
	@./bin/runs-client

.PHONY: clean-db
clean-db: ## Remove local database file
	@echo "Removing runs.db..."
	@rm -f runs.db

##@ Docker - PostgreSQL

.PHONY: docker-postgres
docker-postgres: ## Start PostgreSQL in Docker
	@echo "Starting PostgreSQL container..."
	@docker run --name flyte-runs-postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=flyte_runs \
		-p 5433:5432 \
		-d postgres:15

.PHONY: docker-postgres-stop
docker-postgres-stop: ## Stop and remove PostgreSQL container
	@echo "Stopping PostgreSQL container..."
	@docker stop flyte-runs-postgres || true
	@docker rm flyte-runs-postgres || true

##@ Integration Tests

.PHONY: integration-test
integration-test: build-fast build-client ## Run integration test with SQLite
	@echo "Running service with SQLite..."
	@./bin/runs-service --config config.yaml & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Running client tests..."; \
	./bin/runs-client; \
	kill $$SERVER_PID || true

.PHONY: integration-test-postgres
integration-test-postgres: docker-postgres-stop docker-postgres ## Run integration test with PostgreSQL
	@echo "Waiting for PostgreSQL to start..."
	@sleep 3
	@echo "Running service with PostgreSQL..."
	@$(MAKE) build-fast build-client
	@./bin/runs-service --config config-postgres.yaml & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Running client tests..."; \
	./bin/runs-client; \
	kill $$SERVER_PID; \
	docker stop flyte-runs-postgres; \
	docker rm flyte-runs-postgres

##@ Docker - Service

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t flyte-runs-service:latest -f Dockerfile ..

.PHONY: docker-run
docker-run: docker-build ## Build and run with docker-compose
	@docker-compose up -d

.PHONY: docker-stop
docker-stop: ## Stop docker-compose
	@docker-compose down

.PHONY: docker-logs
docker-logs: ## Show docker-compose logs
	@docker-compose logs -f runs-service

.PHONY: docker-up
docker-up: ## Run everything with docker-compose
	@docker-compose up --build

.PHONY: docker-down
docker-down: ## Stop docker-compose and remove volumes
	@docker-compose down -v

# Override clean to also remove database
clean: clean-db ## Remove build artifacts and database
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
