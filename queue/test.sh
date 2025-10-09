#!/bin/bash
set -e

echo "=== Queue Service Test Script ==="
echo

# Check if PostgreSQL is running
if ! command -v psql &> /dev/null; then
    echo "⚠️  psql not found. Make sure PostgreSQL is installed."
    echo "   You can use Docker: docker run --name flyte-queue-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=flyte_queue -p 5432:5432 -d postgres:15"
    exit 1
fi

# Check if the service binary exists
if [ ! -f "bin/queue-service" ]; then
    echo "📦 Building queue service..."
    cd ..
    go build -o queue/bin/queue-service ./queue/cmd/main.go
    cd queue
    echo "✓ Build complete"
    echo
fi

# Check if PostgreSQL is accessible
if ! psql -h localhost -U postgres -d flyte_queue -c "SELECT 1" &> /dev/null; then
    echo "⚠️  Cannot connect to PostgreSQL at localhost:5432"
    echo "   Make sure PostgreSQL is running with database 'flyte_queue'"
    echo "   Quick start: docker run --name flyte-queue-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=flyte_queue -p 5432:5432 -d postgres:15"
    exit 1
fi

echo "✓ PostgreSQL is running"
echo

# Start the service in the background
echo "🚀 Starting Queue Service..."
./bin/queue-service --config config.yaml &
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
if curl -sf http://localhost:8089/healthz > /dev/null; then
    echo "✓ Health check passed"
else
    echo "❌ Health check failed"
    kill $SERVICE_PID
    exit 1
fi
echo

# Run the test client
echo "🧪 Running test client..."
echo
cd .. && go run queue/client/main.go
TEST_RESULT=$?
cd queue

echo
if [ $TEST_RESULT -eq 0 ]; then
    echo "✅ All tests passed!"
else
    echo "❌ Tests failed"
fi

# Check database
echo
echo "📊 Database contents:"
psql -h localhost -U postgres -d flyte_queue -c "SELECT id, org, project, domain, run_name, action_name, status, enqueued_at FROM queued_actions ORDER BY id;"

# Cleanup
echo
echo "🧹 Stopping service..."
kill $SERVICE_PID
wait $SERVICE_PID 2>/dev/null || true

echo "✓ Done"
