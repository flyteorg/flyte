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
make docker-gen
```

#### Build Rust Crate
```bash
make docker-build-crate
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
- `make docker-gen` - Run code generation
- `make docker-build-crate` - Build Rust crate

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

If you're modifying `Dockerfile.ci`:

1. **Create a PR** with your changes
2. **Wait for build** - A bot will comment with the PR-specific image tag
3. **Test locally** with the same image:
   ```bash
   docker pull ghcr.io/flyteorg/flyte/ci:pr-123  # Use your PR number
   docker run --rm -it -v $(pwd):/workspace -w /workspace \
     ghcr.io/flyteorg/flyte/ci:pr-123 bash
   ```
4. **CI automatically uses** your new image in the PR

This ensures you can test and iterate on Docker image changes in the same PR!

## More Information

See [docs/docker-dev-environment.md](docs/docker-dev-environment.md) for comprehensive documentation.
