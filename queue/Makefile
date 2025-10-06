.PHONY: build test run clean docker-postgres

# Build the queue service
build:
	go build -o bin/queue-service cmd/main.go

# Run tests
test:
	go test -v ./...

# Run the service (requires PostgreSQL)
run: build
	./bin/queue-service serve --config config.yaml

# Run the test client
run-testclient:
	go run testclient/main.go

# Start PostgreSQL in Docker
docker-postgres:
	docker run --name flyte-queue-postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=flyte_queue \
		-p 5432:5432 \
		-d postgres:15

# Stop and remove PostgreSQL container
docker-postgres-stop:
	docker stop flyte-queue-postgres || true
	docker rm flyte-queue-postgres || true

# Clean build artifacts
clean:
	rm -rf bin/

# Full integration test
integration-test: docker-postgres-stop docker-postgres
	@echo "Waiting for PostgreSQL to start..."
	@sleep 3
	@echo "Running service..."
	@./bin/queue-service serve --config config.yaml &
	@SERVER_PID=$$!; \
	sleep 2; \
	echo "Running client tests..."; \
	go run client/main.go; \
	kill $$SERVER_PID; \
	docker stop flyte-queue-postgres; \
	docker rm flyte-queue-postgres
