# Max Parallelism Workflows

This directory contains Flyte workflow definitions designed to test and reproduce max-parallelism behavior, specifically for issue BDP-47029.

## Overview

These workflows test scenarios where a reference launch plan completes execution but the parent's node shows as running for an extended duration. The workflows use various combinations of fast and slow tasks with constrained max-parallelism settings to expose evaluation timing issues.

### Workflows Included

1. **starve_lp_parent_max**: Tests max-parallelism with multiple slow tasks and a fast reference launch plan
2. **starve_lp_parent_fixed**: Refined version with specific task fanout patterns
3. **starve_lp_parent_nested**: Nested workflow calling starve_lp_parent_fixed

## Prerequisites

1. **Local Flyte Cluster**: A running Flyte sandbox or local cluster
2. **Python Environment**: Python 3.8+
3. **Flytekit**: Flyte Python SDK installed

## Setup Instructions

### 1. Start Local Flyte Cluster

This project uses a custom k3d cluster with local development scripts. Follow these steps:

```bash
# Navigate to the local-dev scripts directory
cd /Users/sshardoo/source/flyte-wt/fork-multifix/local-dev/scripts

# Start the k3d cluster and all dependencies (PostgreSQL, MinIO)
./setup-k3d.sh

# Compile and run the Flyte binary with local configuration
./compile-and-run.sh

# The script will output service endpoints:
# - FlyteAdmin HTTP:  http://localhost:8088
# - FlyteAdmin gRPC:  localhost:8089
# - DataCatalog gRPC: localhost:8081
```

**Alternative: Using flytectl demo (Standard Method)**

```bash
# Start the Flyte sandbox
flytectl demo start

# Verify cluster is running
flytectl get project
```

Access the Flyte Console at:
- **Local k3d setup**: http://localhost:8089/console or http://localhost:8088
- **flytectl demo**: http://localhost:30080/console

### 2. Set Up Python Environment

This project uses flytekit installed from source:

```bash
# Create a Python 3.9 virtual environment
python3.9 -m venv ~/.virtualenvs/flytekit-dev
source ~/.virtualenvs/flytekit-dev/bin/activate

# Install flyteidl from the local repository
cd /Users/sshardoo/source/flyte-wt/fork-multifix/flyteidl
pip install .

# Install flytekit from source
cd /Users/sshardoo/source/flyte-flytekit
pip install -e .

# Verify installation
pyflyte --version
```

**Note**: Python 3.9 is required due to dependency compatibility. Using Python 3.13 or newer may cause version constraint issues.

### 3. Configure Flyte Context and Create Project

```bash
# Set up flytectl config for local cluster (if using flytectl demo)
flytectl config init

# Create the project using the Admin API (for local k3d setup)
curl -X POST http://localhost:8088/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "project": {
      "id": "shardool-project",
      "name": "Shardool Project",
      "description": "Test project for max parallelism workflows"
    }
  }'

# Create required Kubernetes namespaces
kubectl --kubeconfig ~/.config/k3d/kubeconfig-flyte-local.yaml \
  create namespace shardool-project-development

kubectl --kubeconfig ~/.config/k3d/kubeconfig-flyte-local.yaml \
  create namespace flytesnacks-development

# Verify project creation
curl -s http://localhost:8088/api/v1/projects | jq
```

### 4. Configure Global MinIO Credentials

For fast registration to work properly, configure global AWS/MinIO environment variables in the Flyte configuration:

Edit `/Users/sshardoo/source/flyte-wt/fork-multifix/local-dev/config/flyte-local-runtime.yaml`:

```yaml
# Add this at the end of the file
plugins:
  k8s:
    default-env-vars:
      FLYTE_AWS_ENDPOINT: http://minio.default.svc.cluster.local:9000
      FLYTE_AWS_ACCESS_KEY_ID: minio
      FLYTE_AWS_SECRET_ACCESS_KEY: miniostorage
```

**Important**: These environment variables are injected into all task pods by the K8s plugin, allowing tasks to download fast-registered code packages from MinIO without requiring credentials in each task definition.

After updating the configuration, restart Flyte:
```bash
# Stop the current Flyte process
pkill -f "flyte start"

# Restart with the updated configuration
cd /Users/sshardoo/source/flyte-wt/fork-multifix/local-dev/scripts
./compile-and-run.sh
```

## Running the Workflows

### Option 1: Local Execution (Testing)

Run workflows locally without registering to Flyte:

```bash
# Navigate to the workflows directory
cd workflows/max-parallel

# Run a workflow locally
pyflyte run max_parallelism_workflows.py starve_lp_parent_max

# Run with specific inputs (if needed)
pyflyte run max_parallelism_workflows.py starve_lp_parent_fixed
```

**Note**: Local execution won't use reference launch plans or max-parallelism constraints. Use this only for basic testing.

### Option 2: Register and Execute on Flyte Cluster

#### Step 1: Create the Referenced Launch Plan

Before registering the main workflows, you need to create the referenced launch plan `another_sleepy_lp`.

The reference workflow is located at `workflows/max-parallel/sleepy.py`.

```bash
# Activate the flytekit virtual environment
source ~/.virtualenvs/flytekit-dev/bin/activate

# Navigate to repository root
cd /Users/sshardoo/source/flyte-wt/fork-multifix

# Register the reference workflow with environment variables for local Flyte cluster
FLYTE_PLATFORM_URL=localhost:8089 FLYTE_PLATFORM_INSECURE=True \
  pyflyte register \
  --project shardool-project \
  --domain development \
  --version level0-v.08 \
  workflows/max-parallel/sleepy.py
```

# Regisger the repro workflow with ref-lps
FLYTE_PLATFORM_URL=localhost:8089 FLYTE_PLATFORM_INSECURE=True   pyflyte register   --project shardool-project   --domain development   --version level0-v.10   workflows/max-parallel/max_parallelism_workflows.py

**Note**: The environment variables are required for connecting to the local Flyte cluster:
- `FLYTE_PLATFORM_URL=localhost:8089`: Points to the local Flyte gRPC endpoint
- `FLYTE_PLATFORM_INSECURE=True`: Disables TLS for local development

#### Step 2: Register the Max Parallelism Workflows

```bash
# Ensure you're in the repository root and the virtual environment is activated
cd /Users/sshardoo/source/flyte-wt/fork-multifix
source ~/.virtualenvs/flytekit-dev/bin/activate

# Register all workflows in the file
FLYTE_PLATFORM_URL=localhost:8089 FLYTE_PLATFORM_INSECURE=True \
  pyflyte register \
  --project shardool-project \
  --domain development \
  --version level0-v.08 \
  workflows/max-parallel/max_parallelism_workflows.py
```

#### Step 3: Execute Workflows

**Via Command Line:**

```bash
# Ensure you're in the repository root and the virtual environment is activated
cd /Users/sshardoo/source/flyte-wt/fork-multifix
source ~/.virtualenvs/flytekit-dev/bin/activate

# Execute the sleepy workflow with pyflyte
FLYTE_PLATFORM_URL=localhost:8089 FLYTE_PLATFORM_INSECURE=True \
  pyflyte run --remote \
  --project shardool-project \
  --domain development \
  workflows/max-parallel/sleepy.py another_sleepy_wf \
  --sleep_seconds 180

# Execute a max-parallelism workflow
FLYTE_PLATFORM_URL=localhost:8089 FLYTE_PLATFORM_INSECURE=True \
  pyflyte run --remote \
  --project shardool-project \
  --domain development \
  workflows/max-parallel/max_parallelism_workflows.py starve_lp_parent_max

# Execute a specific launch plan using flytectl
flytectl create execution \
  --project shardool-project \
  --domain development \
  --launchplan starve_lp_parent_lp_max \
  --version level0-v.08
```

**Via Flyte Console:**

1. Navigate to http://localhost:8089/console (or http://localhost:8088 for the alternate HTTP endpoint)
2. Go to the project `shardool-project`
3. Select domain `development`
4. Find the launch plan (e.g., `starve_lp_parent_lp_max`)
5. Click "Launch" to create a new execution

**Note**: The console URL will be printed when you execute workflows remotely. Look for output like:
```
Go to http://localhost:8089/console/projects/shardool-project/domains/development/executions/<execution-id> to see execution in the console.
```

### Option 3: Package and Deploy with Docker

For a more production-like setup:

#### Step 1: Create a Dockerfile

```bash
cat > Dockerfile << 'EOF'
FROM python:3.9-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy workflow files
COPY . .

# Set the entrypoint for Flyte
ENTRYPOINT ["pyflyte-execute"]
EOF
```

#### Step 2: Build and Push Image

```bash
# Build the Docker image
docker build -t localhost:30000/max-parallel-workflows:v1.0 .

# Push to local registry (if using k3d/kind)
docker push localhost:30000/max-parallel-workflows:v1.0
```

#### Step 3: Register with Docker Image

```bash
pyflyte register max_parallelism_workflows.py \
  --project shardool-project \
  --domain development \
  --version v1.0 \
  --image localhost:30000/max-parallel-workflows:v1.0
```

## Monitoring Workflows

### View Execution Status

```bash
# List all executions for the project
flytectl get execution \
  --project shardool-project \
  --domain development

# Get details of a specific execution
flytectl get execution <execution-name> \
  --project shardool-project \
  --domain development
```

### View Logs

```bash
# Get logs for a specific execution
flytectl get logs <execution-name> \
  --project shardool-project \
  --domain development
```

### Monitor in Console

Navigate to the Flyte Console and view:
- Execution graph showing task dependencies
- Real-time task status
- Max parallelism constraints in action
- Node transition timings

## Troubleshooting

### Issue: Reference Launch Plan Not Found

If you see errors about `another_sleepy_lp` not found:

1. Ensure the reference workflow is registered first (see Step 1 above)
2. Verify the project, domain, and version match exactly
3. Check registered launch plans:

```bash
flytectl get launchplan \
  --project shardool-project \
  --domain development
```

### Issue: Tasks Failing with Resource Constraints

If tasks fail due to resource issues:

1. Check your cluster resources: `kubectl get nodes`
2. Adjust resource requests in the workflow file
3. Reduce max-parallelism settings

### Issue: Proxy Identity Not Working

If `proxy_as` causes issues:

1. Remove the `proxy_as` parameter from task decorators
2. Or configure the identity in your Flyte cluster

### Issue: Project Not Found

Create the project if it doesn't exist:

```bash
flytectl create project \
  --id shardool-project \
  --name "Shardool Project" \
  --description "Testing max parallelism workflows"
```

## Testing the Issue

To reproduce BDP-47029:

1. Register and execute one of the `starve_*` workflows
2. Monitor the execution in the Flyte Console
3. Observe the behavior of reference launch plan nodes
4. Look for scenarios where:
   - The reference launch plan completes execution
   - But the parent node continues showing as "running"
   - Until another node in the workflow completes

The max-parallelism constraint of 5 combined with the task layout should trigger the evaluation delays that expose the issue.

## Configuration Notes

- **Max Parallelism**: Set to 5 for all launch plans
- **Concurrency**: Limited to 5 concurrent executions with SKIP behavior
- **Schedule**: Configured for every 15 minutes (won't auto-trigger in local mode)
- **Resource Requests**: 1 CPU, 500Mi memory per task
- **Task Durations**: Mix of fast (60s-180s) and slow (45-60min) tasks

## Cleanup

```bash
# Stop the Flyte sandbox
flytectl demo teardown

# Deactivate Python environment
deactivate
```

## Additional Resources

- [Flyte Documentation](https://docs.flyte.org/)
- [Flytekit Python SDK](https://docs.flyte.org/projects/flytekit/en/latest/)
- [Max Parallelism Guide](https://docs.flyte.org/en/latest/user_guide/advanced_composition/max_parallelism.html)
