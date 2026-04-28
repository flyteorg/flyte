# Flyte Manager - Unified Binary

The Flyte Manager is a unified binary that runs all Flyte services in a single process:

- **Runs Service** - Manages workflow runs and action state
- **Executor/Operator** - Reconciles and transitions TaskAction CRs through their lifecycle
- **Actions Service** - Serves action metadata and lifecycle APIs, including enqueueing TaskAction CRs
- **DataProxy Service** - Proxies signed-URL and blob access for task I/O
- **Events Service** - Ingests and fans out task/run events
- **Cache Service** - Backs task output caching and lookups
- **App Service** (+ internal proxy) - Hosts the Flyte UI/app and routes to internal services
- **Secret Service** - Manages secret references used by tasks

## Features

✅ **Single Binary** - One process to deploy and manage
✅ **PostgreSQL Backend** - Shared database for all services
✅ **Auto Kubernetes Detection** - Uses current kubeconfig
✅ **Unified Configuration** - One config file for all services
✅ **HTTP/2 Support** - Buf Connect compatible

## API Endpoints

### Manager (port 8090)

All Connect/gRPC services are mounted on a single port. Notable handlers:

- `flyteidl2.workflow.RunService` - Create / Get / List / Abort runs
- `flyteidl2.workflow.InternalRunService` - Internal run-control APIs used by the executor
- `flyteidl2.workflow.TranslatorService` - Translates user task definitions
- `flyteidl2.workflow.RunLogsService` - Stream logs for a run
- `flyteidl2.actions.ActionsService` - Action lifecycle and metadata
- `flyteidl2.task.TaskService` - Task registration and lookup
- `flyteidl2.trigger.TriggerService` - Schedules and triggers
- `flyteidl2.project.ProjectService` - Project management
- `flyteidl2.auth.IdentityService` / `AuthMetadataService` - Identity and auth metadata
- DataProxy, Events, Cache, Secret, and App services (see their respective packages)
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

### Executor (port 8081)

- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

## How It Works

1. **CreateRun** → Runs Service persists the run to PostgreSQL and calls `ActionsService.Enqueue(...)` to enqueue the root action
2. **Actions Service / Executor** → That enqueue flow results in the root TaskAction CR being created in Kubernetes, which the Executor then watches and reconciles
3. **Executor** → Transitions: Queued → Initializing → Running → Succeeded
4. **Actions Service** → Watches TaskAction CRs via a shared informer and forwards status updates (phase, output URI, error state) to subscribers; sdk controller consumes these updates through `WatchForUpdates` to drive the run forward
5. **Runs Service** → Persists state changes to PostgreSQL and notifies its own watchers

## Testing

### Check Services

```bash
# Manager (Connect services) health
curl http://localhost:8090/healthz

# Executor health
curl http://localhost:8081/healthz
```

### Watch TaskActions

```bash
# List all TaskActions
kubectl get taskactions -n flyte

# Watch TaskActions in real-time
kubectl get taskactions -n flyte -w

# Get details of a specific TaskAction
kubectl describe taskaction <name> -n flyte
```

### Check Database

```bash
# Connect to the PostgreSQL backend (devbox defaults)
psql -h localhost -p 30001 -U postgres -d runs

# List tables
\dt

# Query runs
SELECT * FROM runs;

# Query actions
SELECT name, phase, state FROM actions;
```

## Deployment

### Local Development

```bash
make run
```

## Troubleshooting

### Connection Issues

**Error:** "failed to get Kubernetes config"

```bash
# Verify kubeconfig
kubectl cluster-info

# Or set explicit kubeconfig in config.yaml
manager:
  kubernetes:
    kubeconfig: "/path/to/kubeconfig"
```

### Database Issues

**Error:** relation/table not found (runs schema missing)

```bash
# Restart manager so startup migrations run
./bin/flyte-manager --config config.yaml
```

### Port Conflicts

**Error:** "address already in use"

```bash
# Check what's using the ports
lsof -i :8090
lsof -i :8081

# Change ports in config.yaml
manager:
  server:
    port: 9090  # Changed from 8090
```

## Development

### Build

```bash
make build
```

### Clean

```bash
make clean
```

### Run Tests

```bash
make test
```
