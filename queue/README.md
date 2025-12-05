# Queue Service

The Queue Service implements the `QueueService` gRPC interface using buf connect. It manages action execution by creating and managing Kubernetes TaskAction Custom Resources.

## Architecture

**Kubernetes-Native Design:**
- No database required
- Creates TaskAction CRs in Kubernetes for each action
- Uses OwnerReferences for automatic cascading deletion
- Leverages Kubernetes garbage collection for cleanup

## Features

- **Enqueue actions** - Creates TaskAction CRs in Kubernetes
- **Abort queued runs** - Deletes root TaskAction (children cascade automatically)
- **Abort individual actions** - Deletes specific TaskAction (children cascade automatically)
- **OwnerReference hierarchy** - Parent-child relationships via Kubernetes OwnerReferences
- **Automatic cleanup** - Kubernetes handles cascading deletion of dependent resources
- **Health and readiness checks**

## Running the Service

### Prerequisites

1. **Kubernetes cluster** (k3d, kind, minikube, or any K8s cluster)
2. **Go 1.24 or later**
3. **TaskAction CRD** installed in the cluster
4. **Kubeconfig** configured (or running in-cluster)

### Quick Start with k3d

```bash
# Create a k3d cluster
k3d cluster create flyte-dev

# Verify cluster is running
kubectl cluster-info

# The service will automatically use your ~/.kube/config
```

### Run the service

```bash
# From the queue directory
go build -o bin/queue-service ./cmd
./bin/queue-service --config config.yaml
```

The service will:
1. **Auto-detect Kubernetes config:**
   - Try in-cluster config (when running in K8s)
   - Fall back to `~/.kube/config` (for local development)
   - Or use explicit kubeconfig path from config
2. Initialize Kubernetes client
3. Start HTTP/2 server on port 8089 (configurable)

### Configuration

Edit `config.yaml`:

```yaml
queue:
  server:
    host: "0.0.0.0"
    port: 8089
  kubernetes:
    namespace: "flyte"
    # Optional: specify custom kubeconfig path
    # kubeconfig: "/path/to/kubeconfig"
```

**Kubeconfig Resolution:**
1. If `kubeconfig` is set → uses that file
2. Else tries in-cluster config
3. Falls back to default kubeconfig (`~/.kube/config`, `$KUBECONFIG`)

## How It Works

### Enqueue Action

When `EnqueueAction` is called:

1. Creates a TaskAction CR in Kubernetes
2. **Root action:** No OwnerReference, labeled `flyte.org/is-root: "true"`
3. **Child action:** Sets OwnerReference to parent TaskAction
4. Executor watches TaskAction CRs and executes them

**Example TaskAction CR:**
```yaml
apiVersion: flyte.org/v1
kind: TaskAction
metadata:
  name: my-org-my-project-dev-run-001-task-001
  namespace: flyte
  labels:
    flyte.org/org: my-org
    flyte.org/project: my-project
    flyte.org/domain: dev
    flyte.org/run: run-001
    flyte.org/action: task-001
    flyte.org/is-root: "true"
spec:
  taskActionBytes: <protobuf-encoded ActionSpec>
```

**Child action with OwnerReference:**
```yaml
apiVersion: flyte.org/v1
kind: TaskAction
metadata:
  name: my-org-my-project-dev-run-001-task-002
  namespace: flyte
  labels:
    flyte.org/is-root: "false"
  ownerReferences:
  - apiVersion: flyte.org/v1
    kind: TaskAction
    name: my-org-my-project-dev-run-001-task-001
    uid: <parent-uid>
    blockOwnerDeletion: true
spec:
  taskActionBytes: <protobuf-encoded ActionSpec>
```

### Abort Queued Run

When `AbortQueuedRun` is called:

1. Finds root TaskAction (labeled `flyte.org/is-root: "true"`)
2. Deletes the root TaskAction
3. **Kubernetes automatically cascades the deletion** to all child TaskActions

**Hierarchy example:**
```
root-action (deleted manually)
├─ child-1 (deleted by K8s)
│  ├─ grandchild-1 (deleted by K8s)
│  └─ grandchild-2 (deleted by K8s)
└─ child-2 (deleted by K8s)
```

### Abort Queued Action

When `AbortQueuedAction` is called:

1. Deletes the specific TaskAction
2. **Kubernetes automatically cascades the deletion** to any child TaskActions
3. Parent and sibling actions remain unaffected

## Testing

### Check service health

```bash
# Health check
curl http://localhost:8089/healthz

# Readiness check
curl http://localhost:8089/readyz
```

### View TaskActions in Kubernetes

```bash
# List all TaskActions
kubectl get taskactions -n flyte

# Watch TaskActions in real-time
kubectl get taskactions -n flyte -w

# Get details of a specific TaskAction
kubectl describe taskaction <name> -n flyte

# View TaskAction hierarchy (via OwnerReferences)
kubectl get taskaction <name> -n flyte -o yaml | grep -A 5 ownerReferences
```

### Test with queue client

```bash
# Run the test client
go run testclient/main.go
```

Expected flow:
1. Client calls `EnqueueAction` → TaskAction CR created in K8s
2. Client calls `AbortQueuedRun` → Root TaskAction deleted, children cascade deleted
3. Verify in K8s: `kubectl get taskactions -n flyte` (should show deletions)

## API Endpoints

The service exposes the following buf connect endpoints:

- `POST /flyteidl2.workflow.QueueService/EnqueueAction` - Create TaskAction CR
- `POST /flyteidl2.workflow.QueueService/AbortQueuedRun` - Delete root TaskAction (cascades)
- `POST /flyteidl2.workflow.QueueService/AbortQueuedAction` - Delete specific TaskAction (cascades)

Plus health endpoints:
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

## Kubernetes Resources

### TaskAction CR Structure

The TaskAction Custom Resource stores:
- **Spec:** Protobuf-encoded `ActionSpec` (includes task definition, inputs, etc.)
- **Labels:** For organization, filtering, and identifying root actions
- **OwnerReferences:** For automatic cascading deletion

### Required RBAC

The Queue Service needs permissions to:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: queue-service
  namespace: flyte
rules:
- apiGroups: ["flyte.org"]
  resources: ["taskactions"]
  verbs: ["get", "list", "watch", "create", "delete"]
```

## Project Structure

```
queue/
├── cmd/
│   └── main.go              # Entry point with K8s client setup
├── config/
│   └── config.go            # Configuration structs
├── k8s/
│   └── client.go            # Kubernetes operations (create/delete TaskActions)
├── service/
│   └── queue_service.go     # gRPC service handlers
├── client/
│   └── main.go              # Test client
├── config.yaml              # Configuration file
├── Makefile                 # Build and test commands
└── README.md                # This file
```

## Development

### Build

```bash
# Build the service
make build

# Or manually
go build -o bin/queue-service ./cmd
```

### Run locally

```bash
# Ensure you have a kubeconfig configured
export KUBECONFIG=~/.kube/config

# Run the service
./bin/queue-service --config config.yaml
```

### Run in Kubernetes

```bash
# Build Docker image
docker build -t queue-service:latest .

# Deploy to K8s (requires deployment manifests)
kubectl apply -f k8s-manifests/
```

## Advantages of Kubernetes-Native Design

✅ **No Database** - Kubernetes is the single source of truth
✅ **Built-in Durability** - K8s etcd provides persistence
✅ **Automatic Cleanup** - Cascading deletion via OwnerReferences
✅ **Native Watching** - Controllers can watch CR changes
✅ **Scalability** - Kubernetes handles distribution and scheduling
✅ **Simpler Architecture** - One less component to manage
✅ **Idiomatic K8s** - Leverages native Kubernetes patterns

## Troubleshooting

### Connection Issues

**Error:** "failed to get Kubernetes config"
```bash
# Verify kubeconfig is valid
kubectl cluster-info

# Or set explicit kubeconfig in config.yaml
queue:
  kubernetes:
    kubeconfig: "/path/to/kubeconfig"
```

### Permission Issues

**Error:** "failed to create TaskAction CR: forbidden"

Ensure the service has proper RBAC permissions:
```bash
kubectl create role queue-service \
  --verb=get,list,watch,create,delete \
  --resource=taskactions \
  -n flyte

kubectl create rolebinding queue-service \
  --role=queue-service \
  --serviceaccount=flyte:queue-service \
  -n flyte
```

### TaskAction CRD Not Found

**Error:** "no matches for kind TaskAction"

Install the TaskAction CRD:
```bash
kubectl apply -f executor/config/crd/
```
