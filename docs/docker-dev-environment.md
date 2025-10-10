# Docker Development Environment

This document explains how to use the standardized Docker CI image for local development to ensure parity between your local machine and CI.

## Why Use Docker for Development?

The Flyte project uses a Docker image for both CI and local development to:
- **Eliminate environment discrepancies** between local and CI
- **Guarantee consistent tool versions** across the team
- **Simplify onboarding** - no need to install multiple tools
- **Reduce "works on my machine"** issues

## Quick Start

### Prerequisites

- Docker installed on your machine ([Install Docker](https://docs.docker.com/get-docker/))
- Git repository cloned locally

### Basic Usage

Run commands inside the Docker container:

```bash
# Pull the latest image
docker pull ghcr.io/flyteorg/flyte/ci:v2

# Run a command
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  make gen

# Run interactively
docker run --rm -it -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  bash
```

### Helper Script

Create a helper script for easier usage. Save this as `docker-dev.sh`:

```bash
#!/bin/bash
# Helper script to run commands in the CI Docker container

IMAGE="ghcr.io/flyteorg/flyte/ci:v2"
WORKSPACE="/workspace"

# Pull the latest image if not present
if [ "$1" == "pull" ]; then
    docker pull "$IMAGE"
    exit 0
fi

# Run interactively if no command provided
if [ $# -eq 0 ]; then
    docker run --rm -it \
        -v "$(pwd):$WORKSPACE" \
        -w "$WORKSPACE" \
        "$IMAGE" \
        bash
else
    # Run the provided command
    docker run --rm \
        -v "$(pwd):$WORKSPACE" \
        -w "$WORKSPACE" \
        "$IMAGE" \
        "$@"
fi
```

Make it executable:

```bash
chmod +x docker-dev.sh
```

Usage:

```bash
# Pull latest image
./docker-dev.sh pull

# Interactive shell
./docker-dev.sh

# Run specific commands
./docker-dev.sh make gen
./docker-dev.sh buf lint
./docker-dev.sh cargo build
```

## Common Development Tasks

### Generate Protocol Buffers

```bash
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  make gen
```

### Run Buf Commands

```bash
# Lint
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  buf lint

# Format
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  buf format -w
```

### Build Rust Crate

```bash
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  make build-crate
```

### Python Development

```bash
# Install Python dependencies
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  sh -c "cd gen/python && uv sync --all-groups"

# Run Python tests
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  sh -c "cd gen/python && uv run pytest"
```

### TypeScript/Node Development

```bash
# Install dependencies
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  sh -c "cd gen/ts && npm install"

# Build
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  sh -c "cd gen/ts && npm run build"
```

## Tool Versions

The Docker image includes these tools with pinned versions:

| Tool     | Version | Purpose                          |
|----------|---------|----------------------------------|
| Go       | 1.24.6  | Go code generation               |
| Python   | 3.12.9  | Python code generation           |
| uv       | 0.8.4   | Fast Python package manager      |
| Node.js  | 20.18.3 | TypeScript code generation       |
| Rust     | 1.84.0  | Rust code generation             |
| Buf      | 1.58.0  | Protocol buffer tooling          |
| mockery  | 2.53.5  | Go mock generation               |

## Advanced Usage

### Using Docker Compose

Create a `docker-compose.yml` for persistent development:

```yaml
version: '3.8'

services:
  dev:
    image: ghcr.io/flyteorg/flyte/ci:v2
    volumes:
      - .:/workspace
    working_dir: /workspace
    stdin_open: true
    tty: true
    command: bash
```

Usage:

```bash
# Start interactive shell
docker-compose run --rm dev

# Run specific command
docker-compose run --rm dev make gen
```

### Git Configuration Inside Container

If you need to commit from inside the container:

```bash
docker run --rm -it \
  -v $(pwd):/workspace \
  -v ~/.gitconfig:/root/.gitconfig:ro \
  -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  bash
```

### SSH Keys for Private Repositories

If you need SSH keys for private dependencies:

```bash
docker run --rm -it \
  -v $(pwd):/workspace \
  -v ~/.ssh:/root/.ssh:ro \
  -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 \
  bash
```

## Updating the Docker Image

### How Image Building Works

The Docker image is automatically rebuilt when:
- Changes are pushed to the `main` or `v2` branches
- Pull requests modify `Dockerfile.ci` or the build workflow

### PR-Based Workflow (Testing Image Updates)

When you modify `Dockerfile.ci` in a pull request:

1. **Automatic Build**: The workflow automatically builds a PR-specific image tagged as `pr-X` (where X is the PR number)
2. **Automatic Usage**: All CI workflows in that PR automatically use the new image
3. **Bot Comment**: A bot will comment on the PR with the image tag and instructions
4. **Local Testing**: You can pull and test the PR-specific image locally

Example workflow:

```bash
# 1. Create a branch and modify ci.Dockerfile
git checkout -b update-go-version
# Edit ci.Dockerfile to update Go version

# 2. Push to GitHub
git push origin update-go-version

# 3. Create a PR - the bot will comment with the image tag

# 4. Pull and test the PR-specific image locally
docker pull ghcr.io/flyteorg/flyte/ci:pr-123  # Replace 123 with your PR number
docker run --rm -it -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:pr-123 bash

# 5. Verify your changes work
./scripts/docker-dev.sh make gen
```

### Available Image Tags

```bash
# Use latest from main branch
docker pull ghcr.io/flyteorg/flyte/ci:latest

# Use latest from v2 branch (recommended for regular development)
docker pull ghcr.io/flyteorg/flyte/ci:v2

# Use PR-specific image (for testing Dockerfile changes)
docker pull ghcr.io/flyteorg/flyte/ci:pr-123

# Use specific commit
docker pull ghcr.io/flyteorg/flyte/ci:v2-sha-abc123
```

### Testing Dockerfile Changes

**Step 1**: Modify `Dockerfile.ci` in your PR

**Step 2**: Wait for the build (check Actions tab or PR comments)

**Step 3**: CI automatically uses your new image

**Step 4**: Test locally with the same image:
```bash
PR_NUMBER=123  # Your PR number
docker pull ghcr.io/flyteorg/flyte/ci:pr-$PR_NUMBER
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:pr-$PR_NUMBER \
  make gen
```

This ensures true parity: you can test the exact same image that CI uses!

## Troubleshooting

### Permission Issues

If you encounter permission issues with generated files:

```bash
# Run with your user ID
docker run --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  --user $(id -u):$(id -g) \
  ghcr.io/flyteorg/flyte/ci:v2 \
  make gen
```

### Image Not Found

If you get authentication errors pulling the image:

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Or use GitHub CLI
gh auth token | docker login ghcr.io -u USERNAME --password-stdin
```

### Out of Date Image

Always pull the latest image before reporting issues:

```bash
docker pull ghcr.io/flyteorg/flyte/ci:v2
```

## Contributing to the Docker Image

To modify the Docker image:

1. Edit `Dockerfile.ci`
2. Test locally:
   ```bash
   docker build -f ci.Dockerfile -t flyte-ci-test .
   docker run --rm -it flyte-ci-test bash
   ```
3. Submit a PR - the image will be built automatically for testing
4. After merge, the new image will be available as `ghcr.io/flyteorg/flyte/ci:v2`

## Alternatives

If Docker isn't suitable for your workflow, you can still install tools manually following the versions listed in `Dockerfile.ci`. However, this approach is more prone to environment discrepancies.
