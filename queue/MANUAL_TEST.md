# Manual Testing Guide for Queue Service

## Step 1: Prepare PostgreSQL Database

You have PostgreSQL running on localhost:5432. Create the database manually:

```bash
# Option A: If you know your PostgreSQL password
psql -h localhost -U postgres
# Then run:
CREATE DATABASE flyte_queue;
\q

# Option B: Use your existing database
# Just update config.yaml with your database name
```

## Step 2: Update Configuration

Edit `queue/config.yaml` and set your PostgreSQL credentials:

```yaml
database:
  postgres:
    host: "localhost"
    port: 5432
    dbname: "flyte_queue"  # or your existing database name
    username: "postgres"    # or your PostgreSQL username
    password: "YOUR_PASSWORD_HERE"
    extraOptions: "sslmode=disable"
```

## Step 3: Start the Queue Service

```bash
cd /Users/haytham/src/github.com/flyteorg/flyte/queue
./bin/queue-service --config config.yaml
```

Expected output:
```
Starting Queue Service
Running database migrations
Queue Service listening on 0.0.0.0:8089
```

## Step 4: Test Health Endpoints (New Terminal)

```bash
# Health check
curl http://localhost:8089/healthz
# Should return: OK

# Readiness check
curl http://localhost:8089/readyz
# Should return: OK
```

## Step 5: Run the Test Client

```bash
cd /Users/haytham/src/github.com/flyteorg/flyte
go run queue/client/main.go
```

Expected output:
```
Test 1: Enqueuing an action...
✓ Action enqueued successfully: {}

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

## Step 6: Verify Database Contents

```bash
# Connect to your database
psql -h localhost -U postgres -d flyte_queue

# View all queued actions
SELECT id, org, project, domain, run_name, action_name, status, enqueued_at
FROM queued_actions
ORDER BY enqueued_at DESC;

# View aborted actions
SELECT id, run_name, action_name, status, abort_reason
FROM queued_actions
WHERE status = 'aborted';

# Exit
\q
```

## Alternative: Use SQLite for Quick Testing

If PostgreSQL setup is complex, you can modify the code to use SQLite instead:

1. Update `queue/config.yaml`:
```yaml
database:
  sqlite:
    file: "./queue.db"
```

2. Update `queue/cmd/main.go` to use SQLite (if needed)

## Troubleshooting

### PostgreSQL password issues

If you don't know your PostgreSQL password:

```bash
# Check pg_hba.conf location
psql -U postgres -c "SHOW hba_file;"

# Or try without password (if configured for trust auth)
psql -h localhost -U postgres -c "CREATE DATABASE flyte_queue;"
```

### Port already in use

If port 8089 is busy, edit `config.yaml`:
```yaml
queue:
  server:
    port: 8090  # Use different port
```

### Service won't start

Check the logs for specific errors. Common issues:
- Database connection failed → Check credentials in config.yaml
- Port in use → Change port in config.yaml
- Migration failed → Check database permissions

---

## Quick curl Test (Alternative to Go client)

```bash
# Enqueue an action
curl -X POST http://localhost:8089/flyteidl2.workflow.QueueService/EnqueueAction \
  -H "Content-Type: application/json" \
  -d '{
    "actionId": {
      "run": {
        "org": "test-org",
        "project": "test-proj",
        "domain": "dev",
        "name": "run-001"
      },
      "name": "action-001"
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

# Should return: {}
```
