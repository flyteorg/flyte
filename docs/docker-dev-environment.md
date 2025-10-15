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
| Python   | 3.12.9  | Python code generation (installed via uv) |
| uv       | 0.8.4   | Python version & package manager |
| Node.js  | 20.18.3 | TypeScript code generation       |
| Rust     | 1.84.0  | Rust code generation             |
| Buf      | 1.58.0  | Protocol buffer tooling          |
| mockery  | 2.53.5  | Go mock generation               |

**Supported Architectures:**
- `linux/amd64` (Intel/AMD x86_64)
- `linux/arm64` (Apple Silicon, ARM64)

The image is built as a multi-architecture image, so it will automatically pull the correct version for your platform.

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

## Developing the Docker Image

### Fast Local Iteration

When modifying `ci.Dockerfile`, you can build and test locally for faster iteration:

#### Quick Start
```bash
# Build local image and run generation
make docker-dev

# Or step-by-step:
make docker-build        # Build local image
make docker-gen-local    # Run generation with local image
make docker-shell-local  # Interactive shell with local image
```

#### Available Local Targets

| Command | Description |
|---------|-------------|
| `make docker-build` | Build Docker image locally as `flyte-ci:local` |
| `make docker-build-fast` | Build with cache (faster rebuilds) |
| `make docker-gen-local` | Run `make gen` in local container |
| `make docker-shell-local` | Interactive shell in local container |
| `make docker-dev` | Build + generate in one command (recommended) |

#### Full Development Workflow

```bash
# 1. Modify the Dockerfile
vim ci.Dockerfile  # e.g., update Go version

# 2. Build and test locally (fast!)
make docker-dev

# 3. If it works, commit and push
git add ci.Dockerfile
git commit -m "Update Go to 1.25.0"
git push origin update-go

# 4. PR will build image as pr-XXX
# 5. CI automatically uses your PR image
# 6. Merge when ready
```

### Testing Different Approaches

**Local testing** (fastest, no push needed):
```bash
make docker-build
make docker-shell-local
# Test inside container
```

**PR testing** (tests the exact CI image):
```bash
git push  # Triggers PR image build
# Wait for build
docker pull ghcr.io/flyteorg/flyte/ci:pr-123
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:pr-123 bash
```

### Contributing to the Docker Image

To modify the Docker image:

1. Edit `ci.Dockerfile`
2. Test locally:
   ```bash
   make docker-build
   make docker-shell-local
   # Or use docker commands directly:
   # docker build -f ci.Dockerfile -t flyte-ci:local .
   # docker run --rm -it flyte-ci:local bash
   ```
3. Submit a PR - the image will be built automatically as `pr-XXX` for testing
4. After merge, the new image will be available as `ghcr.io/flyteorg/flyte/ci:v2`

## Build Performance

The Docker image is optimized for fast builds using several techniques:

### Multi-Stage Build Strategy

The Dockerfile uses parallel multi-stage builds to download tools simultaneously:

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Go Stage    │  │ Node Stage  │  │Python Stage │  │ Buf Stage   │
│ (parallel)  │  │ (parallel)  │  │ (parallel)  │  │ (parallel)  │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │                │
       └────────────────┴────────────────┴────────────────┘
                              │
                       ┌──────▼──────┐
                       │Final Image  │
                       │ (assembly)  │
                       └─────────────┘
```

**Benefits:**
- 4x parallelization of downloads
- Leverages official Docker images (pre-built, cached)
- Faster than sequential downloads

### Caching Strategy

The build uses a multi-layer caching approach:

1. **Registry cache** (primary): Stored in GHCR, fastest to pull
2. **GitHub Actions cache** (secondary): Fallback for layers
3. **BuildKit inline cache**: Metadata in image layers
4. **Cache mounts**: For package managers (apt, go mod, cargo)

**Cache hierarchy:**
```
1. Try buildcache tag (dedicated cache image)
2. Try current PR tag (if exists)
3. Try v2 tag (stable baseline)
4. Try GHA cache
5. Build from scratch
```

### Performance Improvements

| Optimization | Time Saved |
|--------------|------------|
| Multi-stage parallel builds | ~5-8 min |
| Official image copying vs downloads | ~2-3 min |
| Registry cache (vs no cache) | ~10-12 min |
| Cache mounts for packages | ~1-2 min |
| **Total potential savings** | **15-20 min** |

**Build times:**
- Cold build (no cache): ~15 min
- Warm build (full cache): ~2-3 min
- Incremental build (partial cache): ~5-8 min

### Cache Mounts

The Dockerfile uses BuildKit cache mounts for package managers:

```dockerfile
# APT packages cached
RUN --mount=type=cache,target=/var/cache/apt

# Go modules cached
RUN --mount=type=cache,target=/root/go/pkg/mod

# Cargo packages cached
RUN --mount=type=cache,target=/root/.cargo/registry
```

These persist across builds, dramatically speeding up package installation.

## Alternatives

If Docker isn't suitable for your workflow, you can still install tools manually following the versions listed in `ci.Dockerfile`. However, this approach is more prone to environment discrepancies.
