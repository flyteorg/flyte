# Docker Development - Quick Start

This guide gets you started with Docker-based development in under 5 minutes.

## Why Docker?

Using Docker ensures your local environment matches CI exactly, eliminating "works on my machine" issues.

## Quick Start

### 1. Pull the Image

```bash
make docker-pull
```

or

```bash
docker pull ghcr.io/flyteorg/flyte/ci:v2
```

### 2. Run Common Commands

#### Generate Protocol Buffers
```bash
make gen
```

#### Build Rust Crate
```bash
make build-crate
```

#### Interactive Shell
```bash
make docker-shell
```

### 3. Manual Docker Commands

If you prefer not to use Make:

```bash
# Generate files
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 make gen

# Interactive shell
docker run --rm -it -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:v2 bash
```

## Available Make Targets

Run `make help` to see all available targets including Docker-based ones:

```bash
make help
```

Docker-specific targets:
- `make docker-pull` - Pull the latest CI image
- `make docker-shell` - Start interactive shell
- `make gen` - Run code generation
- `make build-crate` - Build Rust crate

## Troubleshooting

### Permission Issues

If generated files have wrong ownership:

```bash
docker run --rm -v $(pwd):/workspace -w /workspace \
  --user $(id -u):$(id -g) \
  ghcr.io/flyteorg/flyte/ci:v2 make gen
```

### Authentication Issues

Login to GitHub Container Registry:

```bash
gh auth token | docker login ghcr.io -u YOUR_USERNAME --password-stdin
```

## Updating the Docker Image

### Fast Local Iteration (Recommended)

If you're modifying `ci.Dockerfile`, build and test locally first:

```bash
# One command to build and test
make docker-dev

# Or step-by-step
make docker-build          # Build image
make docker-shell-local    # Test interactively
make docker-gen-local      # Run generation
```

This is **much faster** than waiting for PR builds!

### PR Testing (After Local Testing)

Once your local changes work:

1. **Create a PR** with your changes
2. **Wait for build** - A bot will comment with the PR-specific image tag
3. **Test with PR image** to verify CI works:
   ```bash
   docker pull ghcr.io/flyteorg/flyte/ci:pr-123  # Use your PR number
   docker run --rm -it -v $(pwd):/workspace -w /workspace \
     ghcr.io/flyteorg/flyte/ci:pr-123 bash
   ```
4. **CI automatically uses** your new image in the PR

### Workflow Comparison

**Local iteration** (seconds to minutes):
```bash
vim ci.Dockerfile
make docker-dev           # Fast!
# Repeat until it works
```

**PR iteration** (5-10 minutes per build):
```bash
git push
# Wait for build...
docker pull ghcr.io/flyteorg/flyte/ci:pr-123
# Test
```

Use local iteration first, then validate with PR!

## More Information

See [docs/docker-dev-environment.md](docs/docker-dev-environment.md) for comprehensive documentation.
