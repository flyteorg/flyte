# Queue Service - Quick Start Guide

## Prerequisites

- Go 1.21+ installed
- PostgreSQL 12+ (or Docker)

## Option 1: Quick Test (Recommended)

Run the automated test script:

```bash
# If you don't have PostgreSQL running, start it with Docker:
docker run --name flyte-queue-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=flyte_queue \
  -p 5432:5432 \
  -d postgres:15

# Run the test script (builds, starts service, runs tests)
cd queue
./test.sh
```

The script will:
1. Build the service if needed
2. Check PostgreSQL connection
3. Start the Queue Service
4. Run health checks
5. Execute the test client
6. Show database contents
7. Clean up

## Option 2: Manual Testing

### Step 1: Start PostgreSQL

```bash
# Using Docker
docker run --name flyte-queue-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=flyte_queue \
  -p 5432:5432 \
  -d postgres:15

# Or create the database in your existing PostgreSQL
createdb -h localhost -U postgres flyte_queue
```

### Step 2: Build the Service

```bash
cd /path/to/flyte
go build -o queue/bin/queue-service ./queue/cmd/main.go
```

### Step 3: Start the Service

```bash
cd queue
./bin/queue-service --config config.yaml
```

You should see:
```
Starting Queue Service
Running database migrations
Queue Service listening on 0.0.0.0:8089
```

### Step 4: Test the Service

In a new terminal:

```bash
# Health check
curl http://localhost:8089/healthz
# Should return: OK

# Readiness check
curl http://localhost:8089/readyz
# Should return: OK

# Run the test client
cd queue
go run testclient/main.go
```

Expected output:
```
Test 1: Enqueuing an action...
✓ Action enqueued successfully

Test 2: Enqueuing another action in the same run...
✓ Second action enqueued successfully

Test 3: Aborting a specific action...
✓ Action aborted successfully

Test 4: Enqueuing action for a different run...
✓ Action for run-002 enqueued successfully

Test 5: Aborting entire run-001...
✓ Run aborted successfully

All tests completed successfully! 🎉
```

### Step 5: Inspect the Database

```bash
psql -h localhost -U postgres -d flyte_queue

# View all actions
SELECT id, org, project, domain, run_name, action_name, status, enqueued_at
FROM queued_actions
ORDER BY enqueued_at DESC;

# View aborted actions
SELECT id, org, project, domain, run_name, action_name, status, abort_reason
FROM queued_actions
WHERE status = 'aborted';
```

## Testing with curl

You can also test the service using curl (Connect protocol over HTTP):

```bash
# Enqueue an action
curl -X POST http://localhost:8089/flyteidl2.workflow.QueueService/EnqueueAction \
  -H "Content-Type: application/json" \
  -d '{
    "actionId": {
      "run": {
        "org": "test-org",
        "project": "test-project",
        "domain": "development",
        "name": "test-run-001"
      },
      "name": "test-action-001"
    },
    "inputUri": "s3://bucket/input",
    "runOutputBase": "s3://bucket/output",
    "task": {
      "spec": {
        "template": {
          "container": {
            "image": "alpine:latest",
            "args": ["echo", "hello"]
          }
        }
      }
    }
  }'

# Abort a run
curl -X POST http://localhost:8089/flyteidl2.workflow.QueueService/AbortQueuedRun \
  -H "Content-Type: application/json" \
  -d '{
    "runId": {
      "org": "test-org",
      "project": "test-project",
      "domain": "development",
      "name": "test-run-001"
    },
    "reason": "Testing abort"
  }'
```

## Configuration

Edit `queue/config.yaml` to customize:

```yaml
queue:
  server:
    host: "0.0.0.0"
    port: 8089  # Change port if needed
  maxQueueSize: 10000
  workerCount: 10

database:
  postgres:
    host: "localhost"
    port: 5432
    dbname: "flyte_queue"
    username: "postgres"
    password: "postgres"  # Or use passwordPath for secrets
    extraOptions: "sslmode=disable"
  maxIdleConnections: 10
  maxOpenConnections: 100
  connMaxLifeTime: 1h
```

## Troubleshooting

### Cannot connect to database

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check database exists
psql -h localhost -U postgres -l | grep flyte_queue

# If not, create it
createdb -h localhost -U postgres flyte_queue
```

### Port 8089 already in use

Edit `queue/config.yaml` and change the port:

```yaml
queue:
  server:
    port: 8090  # Use a different port
```

### Build errors

```bash
# Make sure you're in the root of the flyte repo
cd /path/to/flyte

# Tidy dependencies
go mod tidy

# Try building again
go build -o queue/bin/queue-service ./queue/cmd/main.go
```

## Clean Up

```bash
# Stop the service (Ctrl+C if running in foreground)

# Stop and remove PostgreSQL container
docker stop flyte-queue-postgres
docker rm flyte-queue-postgres

# Or just stop it to keep the data
docker stop flyte-queue-postgres
```

## Next Steps

- See [README.md](README.md) for detailed documentation
- Check [IMPLEMENTATION_SPEC.md](../IMPLEMENTATION_SPEC.md) for architecture details
- Implement worker logic to process queued actions
