.PHONY: build test run clean docker-postgres

# Build the runs service
build:
	go build -o bin/runs-service cmd/main.go

# Build the client
build-client:
	go build -o bin/runs-client client/main.go

# Run tests
test:
	go test -v ./...

# Run the service with SQLite (default)
run: build
	./bin/runs-service --config config.yaml

# Run the service with PostgreSQL
run-postgres: build
	./bin/runs-service --config config-postgres.yaml

# Run the test client
run-client: build-client
	./bin/runs-client

# Start PostgreSQL in Docker
docker-postgres:
	docker run --name flyte-runs-postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=flyte_runs \
		-p 5433:5432 \
		-d postgres:15

# Stop and remove PostgreSQL container
docker-postgres-stop:
	docker stop flyte-runs-postgres || true
	docker rm flyte-runs-postgres || true

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f runs.db

# Full integration test with SQLite
integration-test: build build-client
	@echo "Running service with SQLite..."
	@./bin/runs-service --config config.yaml & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Running client tests..."; \
	./bin/runs-client; \
	kill $$SERVER_PID || true

# Full integration test with PostgreSQL
integration-test-postgres: docker-postgres-stop docker-postgres
	@echo "Waiting for PostgreSQL to start..."
	@sleep 3
	@echo "Running service with PostgreSQL..."
	@$(MAKE) build build-client
	@./bin/runs-service --config config-postgres.yaml & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Running client tests..."; \
	./bin/runs-client; \
	kill $$SERVER_PID; \
	docker stop flyte-runs-postgres; \
	docker rm flyte-runs-postgres

# Docker commands
docker-build:
	docker build -t flyte-runs-service:latest -f Dockerfile ..

docker-run: docker-build
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f runs-service

# Run everything with docker-compose
docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v
