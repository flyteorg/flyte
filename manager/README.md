# Flyte Manager - Unified Binary

The Flyte Manager is a unified binary that runs all Flyte services in a single process:

- **Runs Service** (port 8090) - Manages workflow runs and action state
- **Queue Service** (port 8089) - Creates and manages TaskAction CRs in Kubernetes
- **Executor/Operator** (port 8081 health) - Reconciles TaskAction CRs and transitions them through states

## Features

✅ **Single Binary** - One process to deploy and manage
✅ **Single SQLite Database** - All data in one file
✅ **Auto Kubernetes Detection** - Uses current kubeconfig
✅ **Unified Configuration** - One config file for all services
✅ **HTTP/2 Support** - Buf Connect compatible

## Quick Start

### Prerequisites

1. **Kubernetes cluster** (k3d, kind, minikube, or any cluster)
2. **Go 1.21 or later**
3. **TaskAction CRD** installed in the cluster
4. **Kubeconfig** configured (or running in-cluster)

### Install TaskAction CRD

```bash
kubectl apply -f ../executor/config/crd/bases/flyte.org_taskactions.yaml
```

### Build and Run

```bash
# Build the binary
make build

# Run the manager
make run

# Or run directly
./bin/flyte-manager --config config.yaml
```

The manager will:
1. Initialize a SQLite database (`flyte.db`)
2. Run database migrations
3. Start all three services in parallel goroutines
4. Connect to your Kubernetes cluster
5. Begin reconciling TaskAction CRs

## Configuration

Edit `config.yaml`:

```yaml
manager:
  runsService:
    host: "0.0.0.0"
    port: 8090

  queueService:
    host: "0.0.0.0"
    port: 8089

  executor:
    healthProbePort: 8081

  kubernetes:
    namespace: "flyte"
    # Optional: specify custom kubeconfig path
    # kubeconfig: "/path/to/kubeconfig"

database:
  type: "sqlite"
  sqlite:
    file: "flyte.db"

logger:
  level: 4  # Info level
  show-source: true
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Flyte Manager Process                   │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐ │
│  │ Runs Service │  │ Queue Service │  │   Executor   │ │
│  │   :8090      │  │    :8089      │  │              │ │
│  │              │  │               │  │  Reconciles  │ │
│  │ - RunService │  │ Creates K8s   │  │  TaskActions │ │
│  │ - StateServ. │  │ TaskAction CRs│  │              │ │
│  └──────┬───────┘  └───────┬───────┘  └──────┬───────┘ │
│         │                  │                  │         │
│         └──────────────────┴──────────────────┘         │
│                            │                             │
│                   ┌────────┴────────┐                   │
│                   │  SQLite DB      │                   │
│                   │  flyte.db       │                   │
│                   └─────────────────┘                   │
└─────────────────────────────────────────────────────────┘
                            │
                            ↓
                   Kubernetes Cluster
                   (TaskAction CRs)
```

## API Endpoints

### Runs Service (port 8090)

- `POST /flyteidl2.workflow.RunService/CreateRun` - Create a new run
- `POST /flyteidl2.workflow.RunService/GetRun` - Get run details
- `POST /flyteidl2.workflow.RunService/ListRuns` - List runs
- `POST /flyteidl2.workflow.RunService/AbortRun` - Abort a run
- `POST /flyteidl2.workflow.StateService/Put` - Update action state
- `POST /flyteidl2.workflow.StateService/Get` - Get action state
- `POST /flyteidl2.workflow.StateService/Watch` - Watch state updates
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

### Queue Service (port 8089)

- `POST /flyteidl2.workflow.QueueService/EnqueueAction` - Create TaskAction CR
- `POST /flyteidl2.workflow.QueueService/AbortQueuedRun` - Delete root TaskAction
- `POST /flyteidl2.workflow.QueueService/AbortQueuedAction` - Delete specific TaskAction
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

### Executor (port 8081)

- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

## How It Works

1. **CreateRun** → Runs Service persists run to SQLite DB
2. **CreateRun** → Runs Service calls Queue Service to enqueue root action
3. **EnqueueAction** → Queue Service creates TaskAction CR in Kubernetes
4. **Executor** → Watches TaskAction CRs and reconciles them
5. **Executor** → Transitions: Queued → Initializing → Running → Succeeded
6. **Executor** → Calls State Service Put() on each transition
7. **State Service** → Persists state updates to SQLite DB
8. **State Service** → Notifies watchers of state changes

## Testing

### Check Services

```bash
# Runs Service health
curl http://localhost:8090/healthz

# Queue Service health
curl http://localhost:8089/healthz

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
# Open SQLite database
sqlite3 flyte.db

# List tables
.tables

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

### Docker

```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN cd manager && make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/manager/bin/flyte-manager .
COPY --from=builder /app/manager/config.yaml .
CMD ["./flyte-manager", "--config", "config.yaml"]
```

### Kubernetes

Deploy as a single pod with access to the Kubernetes API:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flyte-manager
  namespace: flyte
spec:
  replicas: 1
  selector:
    matchLabels:
      app: flyte-manager
  template:
    metadata:
      labels:
        app: flyte-manager
    spec:
      serviceAccountName: flyte-manager
      containers:
      - name: manager
        image: flyte-manager:latest
        ports:
        - containerPort: 8090
          name: runs
        - containerPort: 8089
          name: queue
        - containerPort: 8081
          name: health
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

**Error:** "failed to run migrations"

```bash
# Remove and recreate database
rm flyte.db
./bin/flyte-manager --config config.yaml
```

### Port Conflicts

**Error:** "address already in use"

```bash
# Check what's using the ports
lsof -i :8090
lsof -i :8089
lsof -i :8081

# Change ports in config.yaml
manager:
  runsService:
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
