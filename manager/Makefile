.PHONY: build run clean test

# Build the unified manager binary
build:
	go build -o bin/flyte-manager cmd/main.go

run: build # Run the manager
	./bin/flyte-manager --config config.yaml

run-client: # Run the manager
	go run client/main.go

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
