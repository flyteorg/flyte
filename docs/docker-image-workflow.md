# Docker Image Build Workflow

This document explains how the Docker CI image is built and used across different scenarios.

## Workflow Diagrams

### Scenario 1: Regular PR (No Dockerfile Changes)

```
┌─────────────────────────────────────────────────────────────┐
│ Developer creates PR (no gen.Dockerfile changes)             │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│ check-generate workflow triggers                            │
│ • Checks if gen.Dockerfile modified: NO                      │
│ • Uses image: ghcr.io/flyteorg/flyte/ci:v2                 │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│ Pulls existing v2 image and runs checks                     │
│ ✓ Fast: No image build needed                              │
└─────────────────────────────────────────────────────────────┘
```

### Scenario 2: PR with Dockerfile Changes

```
┌─────────────────────────────────────────────────────────────┐
│ Developer creates PR (modifies gen.Dockerfile)               │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ├──────────────────────┬─────────────────────┐
                  ▼                      ▼                     ▼
         ┌────────────────┐    ┌─────────────────┐   ┌──────────────┐
         │ build-ci-image │    │ check-generate  │   │ Other checks │
         │   workflow     │    │    workflow     │   │              │
         └────────┬───────┘    └────────┬────────┘   └──────┬───────┘
                  │                     │                    │
                  ▼                     ▼                    ▼
         ┌────────────────┐    ┌─────────────────┐   ┌──────────────┐
         │ Builds image   │    │ Detects Docker- │   │ Uses PR      │
         │ with tag:      │    │ file modified   │   │ image        │
         │ pr-123         │    │                 │   │              │
         └────────┬───────┘    └────────┬────────┘   └──────────────┘
                  │                     │
                  │                     ▼
                  │            ┌─────────────────┐
                  │            │ Waits for image │
                  │            │ to be available │
                  │            │ (polls for ~10m)│
                  │            └────────┬────────┘
                  │                     │
                  ▼                     ▼
         ┌────────────────────────────────────┐
         │ Pushes to ghcr.io                  │
         │ ghcr.io/flyteorg/flyte/ci:pr-123   │
         └────────┬───────────────────────────┘
                  │
                  ▼
         ┌────────────────┐
         │ Comments on PR │
         │ with image tag │
         └────────────────┘
                  │
                  ▼
         ┌────────────────────────────────────┐
         │ All workflows use pr-123 image     │
         │ Developer can test same image      │
         └────────────────────────────────────┘
```

### Scenario 3: Merged to v2 Branch

```
┌─────────────────────────────────────────────────────────────┐
│ PR merged to v2 branch                                      │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│ build-ci-image workflow triggers on push                    │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│ Builds and pushes with tags:                                │
│ • ghcr.io/flyteorg/flyte/ci:v2                             │
│ • ghcr.io/flyteorg/flyte/ci:v2-sha-abc123                  │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│ Future PRs use updated v2 image                             │
└─────────────────────────────────────────────────────────────┘
```

## Image Tag Strategy

| Context            | Image Tag                              | When Created                    |
|--------------------|----------------------------------------|---------------------------------|
| Regular PR         | `v2` or `latest`                      | Uses existing branch image      |
| PR with Dockerfile | `pr-123`                              | Built when PR is created/updated|
| v2 branch push     | `v2`                                  | Built on every push to v2       |
| main branch push   | `latest`                              | Built on every push to master   |
| Any branch push    | `{branch}-sha-{commit}`               | Built on push (with commit SHA) |

## Workflow Files

### Primary Workflows

1. **`.github/workflows/build-ci-image.yml`**
   - Builds Docker image
   - Runs on: PR with Dockerfile changes, push to main/v2
   - Publishes to GHCR with appropriate tags
   - Comments on PR with image information

2. **`.github/workflows/check-generate.yml`**
   - Validates generated files
   - Automatically detects if Dockerfile was modified
   - Uses PR-specific image if available, otherwise uses v2

3. **`.github/workflows/regenerate-on-comment.yml`**
   - Regenerates files via `/regen` comment
   - Uses PR-specific image if available

## Developer Experience

### For Regular Development (No Docker Changes)

```bash
# Just use the standard v2 image
make docker-pull
make gen
```

### For Docker Image Updates

#### Option 1: Local Development (Fastest - Recommended)

```bash
# Modify gen.Dockerfile
vim gen.Dockerfile

# Build and test locally
make docker-dev

# Iterate quickly
vim gen.Dockerfile
make docker-build  # Uses cache, faster rebuilds
make gen

# When it works, push to PR
git commit -am "Update Python to 3.13"
git push
```

**Benefits:**
- ⚡ Fast iteration (seconds to minutes)
- 🔄 No network dependency
- 🧪 Test before pushing
- 💰 No CI minutes wasted

#### Option 2: PR-Based Testing

```bash
# 1. Modify gen.Dockerfile
vim gen.Dockerfile

# 2. Create PR
git checkout -b update-python
git commit -am "Update Python to 3.13"
git push origin update-python
# Create PR on GitHub

# 3. Wait for build (~5-10 minutes)
# Bot will comment: "🐳 Docker CI Image Built"
# Image: ghcr.io/flyteorg/flyte/ci:pr-456

# 4. Test locally with YOUR image
make docker-pull gen DOCKER_CI_IMAGE=ghcr.io/flyteorg/flyte/ci:pr-456

# 5. CI automatically uses the same image!
# No need to merge before testing
```

## Benefits of This Approach

1. **Test Before Merge**: You can fully test Docker image changes before merging
2. **True Parity**: Local testing uses the exact same image as CI
3. **Fast Iteration**: No need to merge to test, iterate in the PR
4. **Automatic Detection**: Workflows automatically detect which image to use
5. **Clean Fallback**: If Docker isn't modified, uses the stable v2 image
6. **Transparent**: Bot comments show exactly what image is being used

## Implementation Details

### Image Detection Logic

Workflows check if `gen.Dockerfile` or `.github/workflows/build-ci-image.yml` were modified:

```bash
git diff --name-only origin/$BASE_BRANCH...HEAD | \
  grep -E '^(Dockerfile\.ci|\.github/workflows/build-ci-image\.yml)$'
```

If modified:
- Use `ghcr.io/flyteorg/flyte/ci:pr-{NUMBER}`
- Wait for image to be available (with timeout)

If not modified:
- Use `ghcr.io/flyteorg/flyte/ci:v2`
- Proceed immediately

### Wait Strategy

For PRs with Dockerfile changes, workflows intelligently wait for the build:

1. **Check for build workflow**: Queries GitHub API for `build-ci-image.yml` runs
2. **Find matching run**: Looks for run with same commit SHA
3. **Wait for completion**: Polls every 20 seconds for up to 20 minutes
4. **Verify success**: Ensures build succeeded before proceeding
5. **Pull fresh image**: Downloads the newly built image

**Benefits:**
- Always waits for the actual build to complete (not just image existence)
- Works correctly on subsequent pushes to the same PR
- Provides clear feedback on build status
- Fails fast if build fails

This ensures tests always run with the freshly built image, even when pushing multiple commits to a PR.

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
