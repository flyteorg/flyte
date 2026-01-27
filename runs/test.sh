#!/bin/bash
set -e

echo "=== Runs Service Test Script ==="
echo

# Determine which database to use
DB_TYPE=${1:-sqlite}

if [ "$DB_TYPE" = "postgres" ]; then
    CONFIG_FILE="config-postgres.yaml"
    echo "📊 Using PostgreSQL"

    # Check if PostgreSQL is running
    if ! command -v psql &> /dev/null; then
        echo "⚠️  psql not found. Make sure PostgreSQL is installed."
        echo "   You can use Docker: docker run --name flyte-runs-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=flyte_runs -p 5433:5432 -d postgres:15"
        exit 1
    fi

    # Check if PostgreSQL is accessible
    if ! psql -h localhost -p 5433 -U postgres -d flyte_runs -c "SELECT 1" &> /dev/null; then
        echo "⚠️  Cannot connect to PostgreSQL at localhost:5433"
        echo "   Make sure PostgreSQL is running with database 'flyte_runs'"
        echo "   Quick start: ./start-postgres.sh"
        exit 1
    fi

    echo "✓ PostgreSQL is running"
else
    CONFIG_FILE="config.yaml"
    echo "📊 Using SQLite"
    echo "✓ SQLite ready (no setup required)"
fi

echo

# Check if the service binary exists
if [ ! -f "bin/runs-service" ]; then
    echo "📦 Building runs service..."
    go build -o bin/runs-service ./cmd/main.go
    echo "✓ Build complete"
    echo
fi

# Check if the client binary exists
if [ ! -f "bin/runs-client" ]; then
    echo "📦 Building test client..."
    go build -o bin/runs-testclient ./testclient/main.go
    echo "✓ Build complete"
    echo
fi

# Start the service in the background
echo "🚀 Starting Runs Service..."
./bin/runs-service --config $CONFIG_FILE &
SERVICE_PID=$!

# Wait for service to start
sleep 3

# Check if service is running
if ! kill -0 $SERVICE_PID 2>/dev/null; then
    echo "❌ Failed to start service"
    exit 1
fi

echo "✓ Service started (PID: $SERVICE_PID)"
echo

# Test health endpoint
echo "🔍 Testing health endpoint..."
if curl -sf http://localhost:8090/healthz > /dev/null; then
    echo "✓ Health check passed"
else
    echo "❌ Health check failed"
    kill $SERVICE_PID
    exit 1
fi
echo

# Test readiness endpoint
echo "🔍 Testing readiness endpoint..."
if curl -sf http://localhost:8090/readyz > /dev/null; then
    echo "✓ Readiness check passed"
else
    echo "❌ Readiness check failed"
    kill $SERVICE_PID
    exit 1
fi
echo

# Run the test client
echo "🧪 Running test client..."
echo
./bin/runs-testclient
TEST_RESULT=$?

echo
if [ $TEST_RESULT -eq 0 ]; then
    echo "✅ All tests passed!"
else
    echo "❌ Tests failed"
fi

# Check database contents
echo
echo "📊 Database contents:"
if [ "$DB_TYPE" = "postgres" ]; then
    psql -h localhost -p 5433 -U postgres -d flyte_runs -c "SELECT id, org, project, domain, name, root_action_name, created_at FROM runs ORDER BY id LIMIT 10;"
    echo
    psql -h localhost -p 5433 -U postgres -d flyte_runs -c "SELECT id, org, project, domain, run_name, name, phase FROM actions ORDER BY id LIMIT 10;"
else
    echo "Runs:"
    sqlite3 runs.db "SELECT id, org, project, domain, name, root_action_name FROM runs ORDER BY id LIMIT 10;"
    echo
    echo "Actions:"
    sqlite3 runs.db "SELECT id, org, project, domain, run_name, name, phase FROM actions ORDER BY id LIMIT 10;" || echo "No actions found"
fi

# Cleanup
echo
echo "🧹 Stopping service..."
kill $SERVICE_PID
wait $SERVICE_PID 2>/dev/null || true

echo "✓ Done"

exit $TEST_RESULT
