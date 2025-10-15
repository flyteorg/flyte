# Docker Image Build Workflow

This document explains how the Docker CI image is built and used across different scenarios.

## Workflow Diagrams

### Scenario 1: Regular PR (No Dockerfile Changes)

```
┌─────────────────────────────────────────────────────────────┐
│ Developer creates PR (no Dockerfile.ci changes)             │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│ check-generate workflow triggers                            │
│ • Checks if Dockerfile.ci modified: NO                      │
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
│ Developer creates PR (modifies Dockerfile.ci)               │
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
# Modify ci.Dockerfile
vim ci.Dockerfile

# Build and test locally
make docker-dev

# Iterate quickly
vim ci.Dockerfile
make docker-build-fast  # Uses cache, faster rebuilds
make docker-gen-local

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
# 1. Modify ci.Dockerfile
vim ci.Dockerfile

# 2. Create PR
git checkout -b update-python
git commit -am "Update Python to 3.13"
git push origin update-python
# Create PR on GitHub

# 3. Wait for build (~5-10 minutes)
# Bot will comment: "🐳 Docker CI Image Built"
# Image: ghcr.io/flyteorg/flyte/ci:pr-456

# 4. Test locally with YOUR image
docker pull ghcr.io/flyteorg/flyte/ci:pr-456
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/flyteorg/flyte/ci:pr-456 make gen

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

Workflows check if `Dockerfile.ci` or `.github/workflows/build-ci-image.yml` were modified:

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
