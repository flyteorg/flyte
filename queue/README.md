# Queue Service

The Queue Service implements the `QueueService` gRPC interface using buf connect. It manages the execution queue for actions.

## Features

- Enqueue actions for execution
- Abort queued runs
- Abort individual queued actions
- PostgreSQL-backed persistence
- Health and readiness checks

## Running the Service

### Prerequisites

1. PostgreSQL running (default: localhost:5432)
2. Go 1.21 or later

### Setup PostgreSQL

```bash
# Using Docker
docker run --name flyte-queue-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=flyte_queue \
  -p 5432:5432 \
  -d postgres:15

# Or use your existing PostgreSQL instance
createdb flyte_queue
```

### Run the service

```bash
# From the queue directory
go run cmd/main.go serve --config config.yaml

# Or build and run
go build -o queue-service cmd/main.go
./queue-service serve --config config.yaml
```

The service will:
1. Connect to PostgreSQL
2. Run database migrations
3. Start HTTP/2 server on port 8089 (configurable)

### Configuration

Edit `config.yaml` to customize:

```yaml
queue:
  server:
    host: "0.0.0.0"
    port: 8089
  maxQueueSize: 10000
  workerCount: 10

database:
  postgres:
    host: "localhost"
    port: 5432
    dbname: "flyte_queue"
    username: "postgres"
    password: "postgres"
    extraOptions: "sslmode=disable"
```

## Testing

### Run unit tests

```bash
go test ./...
```

### Run integration test with client

```bash
# Terminal 1: Start the service
go run cmd/main.go serve --config config.yaml

# Terminal 2: Run the test client
go run client/main.go
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

### Check service health

```bash
# Health check
curl http://localhost:8089/healthz

# Readiness check
curl http://localhost:8089/readyz
```

### Inspect database

```bash
# Connect to PostgreSQL
psql -h localhost -U postgres -d flyte_queue

# View queued actions
SELECT id, org, project, domain, run_name, action_name, status, enqueued_at
FROM queued_actions
ORDER BY enqueued_at DESC;

# View aborted actions
SELECT id, org, project, domain, run_name, action_name, status, abort_reason
FROM queued_actions
WHERE status = 'aborted';
```

## API Endpoints

The service exposes the following buf connect endpoints:

- `POST /flyteidl2.workflow.QueueService/EnqueueAction` - Enqueue a new action
- `POST /flyteidl2.workflow.QueueService/AbortQueuedRun` - Abort all actions in a run
- `POST /flyteidl2.workflow.QueueService/AbortQueuedAction` - Abort a specific action

Plus health endpoints:
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

## Database Schema

### queued_actions table

```sql
CREATE TABLE queued_actions (
    id BIGSERIAL PRIMARY KEY,
    org VARCHAR(255) NOT NULL,
    project VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    run_name VARCHAR(255) NOT NULL,
    action_name VARCHAR(255) NOT NULL,
    parent_action_name VARCHAR(255),
    action_group VARCHAR(255),
    subject VARCHAR(255),
    action_spec JSONB NOT NULL,
    input_uri TEXT NOT NULL,
    run_output_base TEXT NOT NULL,
    enqueued_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'queued',
    abort_reason TEXT,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(org, project, domain, run_name, action_name)
);
```

## Project Structure

```
queue/
├── cmd/
│   ├── main.go          # Entry point
│   └── serve.go         # Serve command implementation
├── config/
│   └── config.go        # Configuration structs
├── repository/
│   ├── interfaces.go    # Repository interface
│   ├── models.go        # Database models
│   └── postgres.go      # PostgreSQL implementation
├── service/
│   ├── queue_service.go      # Service implementation
│   └── queue_service_test.go # Unit tests
├── migrations/
│   └── migrations.go    # Database migrations
├── client/
│   └── main.go          # Test client
├── config.yaml          # Sample configuration
└── README.md           # This file
```
