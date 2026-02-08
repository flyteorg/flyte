# Local Development Setup

This guide walks you through setting up the full Flyte v2 system locally using the unified manager binary (single process).

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.24.0+ | [golang.org/dl](https://go.dev/dl/) |
| Docker | 17.03+ | [docker.com](https://docs.docker.com/get-docker/) |
| kind | latest | `go install sigs.k8s.io/kind@latest` |
| kubectl | 1.11+ | [kubernetes.io](https://kubernetes.io/docs/tasks/tools/) |

> If your system Go version is older than 1.24, you can install it side-by-side:
> ```sh
> go install golang.org/dl/go1.24.0@latest
> go1.24.0 download
> export GOROOT=$(go1.24.0 env GOROOT)
> export PATH="$GOROOT/bin:$PATH"
> ```

## Quick Start

Run these commands from the repository root to get the whole system up and running.

### 1. Create a kind cluster

```sh
kind create cluster --image=kindest/node:v1.26.0 --name flytev2
```

### 2. Install the TaskAction CRD and create the `flyte` namespace

```sh
kubectl apply -f executor/config/crd/bases/flyte.org_taskactions.yaml
kubectl create namespace flyte
```

### 3. Deploy MinIO (object storage backend)

```sh
kubectl apply -f dataproxy/deployment/minio.yaml
kubectl wait --for=condition=ready pod -l app=minio -n flyte-dataproxy --timeout=120s
```

### 4. Port-forward MinIO

Open a **separate terminal** and keep it running:

```sh
kubectl port-forward -n flyte-dataproxy svc/minio 9000:9000 9001:9001
```

### 5. Create the `flyte-data` storage bucket

Open [http://localhost:9001](http://localhost:9001) in your browser (username: `minioadmin`, password: `minioadmin`) and create a bucket named `flyte-data`.

Alternatively, use the MinIO client CLI:

```sh
# Install mc (MinIO Client) if you don't have it
# macOS: brew install minio/stable/mc
# Linux: https://min.io/docs/minio/linux/reference/minio-mc.html

mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/flyte-data
```

### 6. Build and run the manager

```sh
cd manager
make run
```

This starts all services in a single process:

| Service | Port | Description |
|---------|------|-------------|
| Runs Service | 8090 | Manages workflow runs and action state |
| Queue Service | 8089 | Creates and manages TaskAction CRs in Kubernetes |
| Executor | 8081 | Reconciles TaskAction CRs (health probe) |

### 7. Verify everything is working

```sh
# Check Runs Service
curl http://localhost:8090/healthz

# Check Queue Service
curl http://localhost:8089/healthz

# Check Executor
curl http://localhost:8081/healthz

# List TaskActions in the cluster
kubectl get taskactions -n flyte
```

## Teardown

```sh
# Stop the manager (Ctrl+C in the terminal running it)

# Delete the kind cluster (removes everything)
kind delete cluster --name flytev2
```

To clean up the local database without removing the cluster:

```sh
cd manager
make clean-db
```

## Running Individual Services

If you're only working on a specific service, you can run it standalone instead of using the unified manager.

### Runs Service

```sh
cd runs
make run        # Uses SQLite by default (zero setup)
```

For PostgreSQL:

```sh
cd runs
make docker-postgres    # Start PostgreSQL in Docker
make run-postgres       # Run with PostgreSQL config
```

### Queue Service

```sh
cd queue
make run
```

### Executor (standalone)

```sh
cd executor
make install   # Install CRD into cluster
make build
./bin/manager
```

### Data Proxy

```sh
cd dataproxy
go run cmd/main.go --config config/config.example.yaml
```

Requires MinIO to be running and port-forwarded (steps 3-5 above).

## Configuration

The manager uses `manager/config.yaml` by default. Key settings:

```yaml
manager:
  server:
    host: "0.0.0.0"
    port: 8090
  executor:
    healthProbePort: 8081
  kubernetes:
    namespace: "flyte"

database:
  type: "sqlite"
  sqlite:
    file: "flyte.db"

storage:
  type: minio
  container: "flyte-data"
  stow:
    kind: "s3"
    config:
      access_key_id: "minioadmin"
      secret_key: "minioadmin"
      endpoint: "http://localhost:9000"
```

See `manager/config.yaml` for the full configuration reference.

## Testing with flyte-sdk

Once the system is running, you can submit workflows using the [flyte-sdk](https://github.com/flyteorg/flyte-sdk):

```yaml
# ~/.flyte/config.yaml
admin:
  endpoint: dns:///localhost:8090
  insecure: true
image:
  builder: local
task:
  domain: development
  project: testproject
  org: testorg
```

```sh
python example/basics/hello.py
```

## Useful Commands

```sh
# Watch TaskActions in real-time
kubectl get taskactions -n flyte -w

# Inspect a specific TaskAction
kubectl describe taskaction <name> -n flyte

# Query the local database
sqlite3 manager/flyte.db ".tables"
sqlite3 manager/flyte.db "SELECT * FROM runs;"

# Run unit tests for a specific service
cd manager && make test
cd runs && make test
cd executor && make test
```

## Troubleshooting

### "failed to get Kubernetes config"

Make sure your kubeconfig points to the kind cluster:

```sh
kubectl cluster-info --context kind-flytev2
```

### Port already in use

Check what's using the port and either stop it or change the port in `config.yaml`:

```sh
lsof -i :8090
lsof -i :8089
lsof -i :8081
```

### MinIO connection refused

Verify the port-forward is still running and MinIO pod is healthy:

```sh
kubectl get pods -n flyte-dataproxy
kubectl port-forward -n flyte-dataproxy svc/minio 9000:9000 9001:9001
```

### Database migration errors

Remove and recreate the database:

```sh
cd manager
make clean-db
make run
```

## Architecture

```
                    Flyte Manager (single binary)
  ┌──────────────────────────────────────────────────────────┐
  │                                                          │
  │  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  │
  │  │ Runs Service │  │ Queue Service │  │   Executor   │  │
  │  │   :8090      │  │    :8089      │  │   :8081      │  │
  │  └──────┬───────┘  └───────┬───────┘  └──────┬───────┘  │
  │         └──────────────────┼──────────────────┘          │
  │                    ┌───────┴───────┐                     │
  │                    │  SQLite DB    │                     │
  │                    └───────────────┘                     │
  └──────────────────────────┬───────────────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   kind cluster  │
                    │  (TaskAction    │
                    │    CRs + MinIO) │
                    └─────────────────┘
```

For the full architecture specification, see [IMPLEMENTATION_SPEC.md](IMPLEMENTATION_SPEC.md).
