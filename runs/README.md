# Runs Service

The Runs Service implements the `RunService` gRPC interface using buf connect. It manages workflow runs, actions, attempts, and cluster events.

## Features

- Create and manage runs
- Track actions and their execution state
- Record action attempts and retries
- Store cluster events and phase transitions
- List and filter runs and actions
- Abort runs and individual actions
- Dual database support: **SQLite** (default) and **PostgreSQL**
- Health and readiness checks
- Streaming RPCs for real-time updates

## Running the Service

### Prerequisites

1. Go 1.24 or later
2. (Optional) PostgreSQL running for production use

### Quick Start with SQLite

The service uses SQLite by default, no database setup required!

```bash
# Build and run
go build -o runs-service cmd/main.go
./runs-service --config config.yaml
```

Or using make:

```bash
make build
make run
```

The service will:
1. Create/connect to SQLite database (`runs.db`)
2. Run database migrations
3. Start HTTP/2 server on port 8090 (configurable)

### Setup with PostgreSQL

For production deployments, use PostgreSQL:

```bash
# Using Docker
docker run --name flyte-runs-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=flyte_runs \
  -p 5433:5432 \
  -d postgres:15

# Or use your existing PostgreSQL instance
createdb flyte_runs
```

Run with PostgreSQL config:

```bash
./runs-service --config config-postgres.yaml
```

### Configuration

**SQLite (default) - `config.yaml`:**

```yaml
runs:
  server:
    host: "0.0.0.0"
    port: 8090
  watchBufferSize: 100

database:
  sqlite:
    file: "./runs.db"

logger:
  level: 4  # Info level
  show-source: true
```

**PostgreSQL - `config-postgres.yaml`:**

```yaml
runs:
  server:
    host: "0.0.0.0"
    port: 8090
  watchBufferSize: 100

database:
  postgres:
    host: "localhost"
    port: 5433
    dbname: "flyte_runs"
    username: "postgres"
    password: "postgres"
    extraOptions: "sslmode=disable"
  maxIdleConnections: 10
  maxOpenConnections: 100
  connMaxLifeTime: 1h

logger:
  level: 4
  show-source: true
```

## Testing

### Run unit tests

```bash
go test ./...
# or
make test
```

### Run API tests

Pleaes ensure the service is started by following [Quick Start with SQLite](#quick-start-with-sqlite) section
before running API tests.

```sh
make api-test
```

### Run integration test with client

**Using SQLite:**

```bash
# Terminal 1: Start the service
go run cmd/main.go --config config.yaml

# Terminal 2: Run the test client
go run testclient/main.go
```

**Or use make:**

```bash
make integration-test
```

**Using PostgreSQL:**

```bash
make integration-test-postgres
```

Expected output:
```
Test 1: Creating a run with task spec...
✓ Run created successfully: test-run-001

Test 2: Creating a run with auto-generated name...
✓ Run with auto-generated name created: run-1759535665

Test 3: Getting run details...
✓ Retrieved run details: details:{}

Test 4: Listing runs...
✓ Found 2 runs

Test 5: Listing actions for the run...
✓ Found 0 actions

Test 6: Aborting an action...
✓ Action aborted successfully

Test 7: Aborting a run...
✓ Run aborted successfully

All tests completed successfully! 🎉
```

## Scripts

Convenient scripts are provided in `runs/tests/scripts/` to interact with the service using `buf curl`.
Ensure the service is running before executing these scripts.

- `./runs/tests/scripts/create_task.sh` - create a new task
- `./runs/tests/scripts/list_tasks.sh` - list tasks with name filtering

### Check service health

```bash
# Health check
curl http://localhost:8090/healthz

# Readiness check
curl http://localhost:8090/readyz
```

### Inspect database

**SQLite:**

```bash
sqlite3 runs.db

# View runs
SELECT id, org, project, domain, name, root_action_name, created_at
FROM runs
ORDER BY created_at DESC;

# View actions
SELECT id, org, project, domain, run_name, name, phase, action_type
FROM actions
ORDER BY created_at DESC;
```

**PostgreSQL:**

```bash
# Connect to PostgreSQL
psql -h localhost -p 5433 -U postgres -d flyte_runs

# View runs
SELECT id, org, project, domain, name, root_action_name, created_at
FROM runs
ORDER BY created_at DESC;

# View actions
SELECT id, org, project, domain, run_name, name, phase, action_type
FROM actions
ORDER BY created_at DESC;
```

## API Endpoints

The service exposes the following buf connect endpoints:

### Run Management
- `POST /flyteidl2.workflow.RunService/CreateRun` - Create a new run
- `POST /flyteidl2.workflow.RunService/GetRunDetails` - Get run details
- `POST /flyteidl2.workflow.RunService/ListRuns` - List runs with filtering
- `POST /flyteidl2.workflow.RunService/AbortRun` - Abort a run

### Action Management
- `POST /flyteidl2.workflow.RunService/GetActionDetails` - Get action details
- `POST /flyteidl2.workflow.RunService/GetActionData` - Get action input/output data
- `POST /flyteidl2.workflow.RunService/ListActions` - List actions for a run
- `POST /flyteidl2.workflow.RunService/AbortAction` - Abort a specific action

### Task Management
- `POST /flyteidl2.workflow.TaskService/CreateTask` - Create a new task
- `POST /flyteidl2.workflow.TaskService/GetTask` - Get task details
- `POST /flyteidl2.workflow.TaskService/ListTasks` - List tasks with filtering and sorting
- `POST /flyteidl2.workflow.TaskService/UpdateTask` - Update an existing task

### Streaming (Watch) RPCs
- `POST /flyteidl2.workflow.RunService/WatchRunDetails` - Stream run detail updates
- `POST /flyteidl2.workflow.RunService/WatchActionDetails` - Stream action detail updates
- `POST /flyteidl2.workflow.RunService/WatchRuns` - Stream run updates
- `POST /flyteidl2.workflow.RunService/WatchActions` - Stream action updates
- `POST /flyteidl2.workflow.RunService/WatchClusterEvents` - Stream cluster events

### Health Endpoints
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

## Database Schema

### runs table

```sql
CREATE TABLE runs (
    id BIGSERIAL PRIMARY KEY,
    org VARCHAR(255) NOT NULL,
    project VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    root_action_name VARCHAR(255) NOT NULL,
    trigger_org VARCHAR(255),
    trigger_project VARCHAR(255),
    trigger_domain VARCHAR(255),
    trigger_name VARCHAR(255),
    run_spec JSONB NOT NULL,
    created_by_principal VARCHAR(255),
    created_by_k8s_service_account VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(org, project, domain, name)
);
```

### actions table

```sql
CREATE TABLE actions (
    id BIGSERIAL PRIMARY KEY,
    org VARCHAR(255) NOT NULL,
    project VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    run_name VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    run_id BIGINT NOT NULL REFERENCES runs(id),
    parent_action_name VARCHAR(255),
    action_type VARCHAR(50) NOT NULL,
    phase VARCHAR(50) NOT NULL DEFAULT 'PHASE_QUEUED',
    task_spec JSONB,
    trace_spec JSONB,
    condition_spec JSONB,
    input_uri TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(org, project, domain, run_name, name)
);
```

### action_attempts table

```sql
CREATE TABLE action_attempts (
    id BIGSERIAL PRIMARY KEY,
    action_id BIGINT NOT NULL REFERENCES actions(id),
    attempt_number INT NOT NULL,
    phase VARCHAR(50) NOT NULL DEFAULT 'PHASE_QUEUED',
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    outputs JSONB,
    log_info JSONB,
    log_context JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(action_id, attempt_number)
);
```

### cluster_events table

```sql
CREATE TABLE cluster_events (
    id BIGSERIAL PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES action_attempts(id),
    occurred_at TIMESTAMP NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### phase_transitions table

```sql
CREATE TABLE phase_transitions (
    id BIGSERIAL PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES action_attempts(id),
    phase VARCHAR(50) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Project Structure

```
runs/
├── cmd/
│   └── main.go              # Entry point
├── config/
│   └── config.go            # Configuration structs
├── repository/
│   ├── interfaces.go        # Repository interface
│   ├── models.go            # Database models
│   └── postgres.go          # PostgreSQL/SQLite implementation
├── service/
│   └── run_service.go       # Service implementation
├── migrations/
│   └── migrations.go        # Database migrations
├── client/
│   └── main.go              # Test client
├── config.yaml              # SQLite configuration (default)
├── config-postgres.yaml     # PostgreSQL configuration
├── Makefile                 # Build and test commands
└── README.md                # This file
```

## Development

### Build commands

```bash
# Build service
make build

# Build client
make build-testclient

# Clean artifacts
make clean
```

### Database switching

The service automatically selects the database based on configuration:
- If `database.sqlite.file` is set → uses SQLite
- If `database.postgres` is set → uses PostgreSQL

No code changes needed, just update `config.yaml`!

### Database migration

We are using gorm for auto migration, please ensure addding `gorm` tags in `/runs/repository/models/` when you
add any more columns/models.

### Adding new features

1. Update proto definitions in `/flyteidl2/workflow/`
2. Regenerate code: `buf generate`
3. Update repository interface in `repository/interfaces/`
4. Update DB models in `repository/models/`
5. Implement in `repository/impl/`
6. Add service handler in `service/xxx_service.go`
7. Add tests and update client

## Troubleshooting

### Port already in use

If port 8090 is already in use, update in `config.yaml`:

```yaml
runs:
  server:
    port: 8091  # Use a different port
```

### Database connection issues

**SQLite:**
- Check file permissions on `./runs.db`
- Ensure directory is writable

**PostgreSQL:**
- Verify PostgreSQL is running: `pg_isready -h localhost -p 5433`
- Check credentials in `config-postgres.yaml`
- Ensure database exists: `psql -h localhost -p 5433 -U postgres -l`

### Service won't start

Check logs for detailed error messages. Common issues:
- Config file not found (use `--config` flag)
- Database connection failed
- Port already in use
