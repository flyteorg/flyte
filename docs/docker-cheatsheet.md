# Docker Development Cheatsheet

Quick reference for Docker-based development in Flyte.

## Common Commands

### Regular Development (Using Remote Image)

```bash
# Pull latest image
make docker-pull

# Run generation
make docker-gen

# Build Rust crate
make docker-build-crate

# Interactive shell
make docker-shell
```

### Local Development (Building Image Locally)

```bash
# Build and test in one command (recommended)
make docker-dev

# Or step by step:
make docker-build           # Build local image
make docker-gen-local       # Run generation
make docker-build-crate-local  # Build Rust
make docker-shell-local     # Interactive shell

# Fast rebuilds (uses cache)
make docker-build-fast
```

## Image Tags

| Tag | Usage | When to Use |
|-----|-------|-------------|
| `flyte-ci:local` | Local builds | Fast iteration on Dockerfile changes |
| `ghcr.io/flyteorg/flyte/ci:v2` | Stable v2 branch | Regular development |
| `ghcr.io/flyteorg/flyte/ci:pr-123` | PR-specific | Testing Dockerfile changes in CI |
| `ghcr.io/flyteorg/flyte/ci:latest` | Main branch | Latest stable |

## Workflow Decision Tree

```
Are you modifying ci.Dockerfile?
│
├─ YES → Use local builds
│   └─ make docker-dev (fastest!)
│
└─ NO → Use remote image
    └─ make docker-gen
```

## Quick Workflows

### Update Tool Version in Dockerfile

```bash
# 1. Edit
vim ci.Dockerfile

# 2. Build and test locally
make docker-dev

# 3. Iterate if needed
vim ci.Dockerfile
make docker-build-fast && make docker-gen-local

# 4. Push when ready
git commit -am "Update Go to 1.25"
git push
```

### Run Generation Only

```bash
# With remote image
make docker-gen

# With local image
make docker-gen-local
```

### Debug Inside Container

```bash
# Remote image
make docker-shell

# Local image
make docker-shell-local
```

## Manual Docker Commands

If you prefer raw Docker commands:

```bash
# Build
docker build -f ci.Dockerfile -t flyte-ci:local .

# Run generation
docker run --rm -v $(pwd):/workspace -w /workspace \
  flyte-ci:local make gen

# Interactive
docker run --rm -it -v $(pwd):/workspace -w /workspace \
  flyte-ci:local bash

# Pull PR image
docker pull ghcr.io/flyteorg/flyte/ci:pr-123
```

## Troubleshooting

### Permission Issues
```bash
# Run with your user ID
docker run --rm -v $(pwd):/workspace -w /workspace \
  --user $(id -u):$(id -g) \
  flyte-ci:local make gen
```

### Cache Issues
```bash
# Clear Docker build cache
docker builder prune

# Rebuild without cache
docker build --no-cache -f ci.Dockerfile -t flyte-ci:local .
```

### See All Available Targets
```bash
make help
```

## Architecture Support

- **linux/amd64** (Intel/AMD x86_64)
- **linux/arm64** (Apple Silicon M1/M2/M3/M4)

Docker automatically pulls the correct architecture.

## Performance Tips

1. **Local first**: Always test with `make docker-dev` before pushing
2. **Use cache**: Use `make docker-build-fast` for subsequent builds
3. **Layer optimization**: Put frequently changing commands at the end of Dockerfile
4. **Multi-stage builds**: Consider multi-stage for large dependencies

## See Also

- [DOCKER_QUICK_START.md](../DOCKER_QUICK_START.md) - Quick start guide
- [docker-dev-environment.md](docker-dev-environment.md) - Comprehensive documentation
- [docker-image-workflow.md](docker-image-workflow.md) - Workflow diagrams
