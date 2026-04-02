# Mockery v2 → v3 Migration Plan

## Overview

Migrate mockery from v2 (currently v2.52.1/v2.53.5) to v3.7.0 (latest). This involves updating install references, converting the `.mockery.yaml` config to v3 format, consolidating all `go:generate mockery` directives into the central config, and removing the `go:generate` lines.

Reference: https://vektra.github.io/mockery/v3.0/v3/

---

## Step 1: Update mockery version in all install locations

### 1a. `gen.Dockerfile` (line 16)

```diff
- ARG MOCKERY_VERSION=2.53.5
+ ARG MOCKERY_VERSION=3.7.0
```

```diff
- go install "github.com/vektra/mockery/v2@v${MOCKERY_VERSION}"
+ go install "github.com/vektra/mockery/v3@v${MOCKERY_VERSION}"
```

### 1b. `boilerplate/flyte/golang_support_tools/go.mod` (line 10)

```diff
- github.com/vektra/mockery/v2 v2.52.1
+ github.com/vektra/mockery/v3 v3.7.0
```

Then run `go mod tidy` in that directory.

### 1c. `boilerplate/flyte/golang_support_tools/tools.go` (line 7)

```diff
- _ "github.com/vektra/mockery/v2/cmd"
+ _ "github.com/vektra/mockery/v3/cmd"
```

### 1d. `boilerplate/flyte/golang_test_targets/download_tooling.sh` (line 19)

```diff
- "github.com/vektra/mockery/v2@v2.52.1"
+ "github.com/vektra/mockery/v3@v3.7.0"
```

---

## Step 2: Convert `.mockery.yaml` to v3 format

**Use `mockery migrate`** to auto-convert, then verify. Do NOT manually rewrite the config — the migrator handles field mapping accurately.

```bash
mockery migrate --config .mockery.yaml
# Produces .mockery_v3.yml — review, then replace:
mv .mockery_v3.yml .mockery.yaml
```

The migrator produces this v3 config (already verified on this repo):

```yaml
log-level: warn
structname: '{{.InterfaceName}}'
pkgname: mocks
template: testify
template-data:
  unroll-variadic: true
packages:
  # ... (all existing packages carried over with correct field names)
```

### Key changes the migrator applies:

| v2 field                 | v3 equivalent                       | Notes                                        |
|--------------------------|-------------------------------------|----------------------------------------------|
| `outpkg`                 | `pkgname`                           | Renamed                                      |
| `inpackage`              | *(removed)*                         | v3 auto-detects same-package placement       |
| `with-expecter`          | *(removed)*                         | Always enabled in v3                         |
| `resolve-type-alias`     | *(removed)*                         | Permanently `False` in v3                    |
| `disable-version-string` | *(removed)*                         | Permanently disabled in v3                   |
| `issue-845-fix`          | *(removed)*                         | No longer applicable in v3                   |
| `mockname`               | *(removed by migrator)*             | Default behavior matches `{{.InterfaceName}}`|
| `filename`               | *(removed by migrator)*             | Default snakecase matches existing behavior  |
| *(new)*                  | `template: testify`                 | Required — specifies template                |
| *(new)*                  | `template-data.unroll-variadic`     | Preserves v2 default variadic behavior       |

### Important: `structname` stays `{{.InterfaceName}}`

The v2 config had `structname: ""` (empty). The migrator correctly maps this to `structname: '{{.InterfaceName}}'`, which preserves mock struct names matching the interface name. Do NOT change this to `{{.InterfaceNameCamel}}` — that would rename mocks and break all test code.

---

## Step 3: Consolidate `go:generate mockery` directives into `.mockery.yaml`

The following 5 packages have mocks generated via `go:generate` directives or other means and are NOT in the migrated `.mockery.yaml`. They all need to be added.

### 3a. `flytestdlib/autorefreshcache/` (1 directive)

**File:** `flytestdlib/autorefreshcache/auto_refresh.go` (line 28)
**Directive:** `//go:generate mockery -all`
**Existing mocks:** `AutoRefresh.go`, `AutoRefreshWithUpdate.go`, `Item.go`, `ItemWrapper.go`

**Add to `.mockery.yaml`:**
```yaml
  github.com/flyteorg/flyte/v2/flytestdlib/autorefreshcache:
    config:
      all: true
      dir: 'flytestdlib/autorefreshcache/mocks'
```

### 3b. `flyteplugins/go/tasks/pluginmachinery/webapi/` (1 directive)

**File:** `flyteplugins/go/tasks/pluginmachinery/webapi/plugin.go` (line 19)
**Directive:** `//go:generate mockery -all -case=underscore`
**Existing mocks:** 9 files (async_plugin.go, delete_context.go, etc.)

**Add to `.mockery.yaml`:**
```yaml
  github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/webapi:
    config:
      all: true
      dir: 'flyteplugins/go/tasks/pluginmachinery/webapi/mocks'
```

> Note: `-case=underscore` is handled by the default v3 snakecase filename behavior.

### 3c. `flyteplugins/go/tasks/pluginmachinery/internal/webapi/` (1 directive)

**File:** `flyteplugins/go/tasks/pluginmachinery/internal/webapi/cache.go` (line 19)
**Directive:** `//go:generate mockery -all -case=underscore`
**Existing mocks:** `client.go`

**Add to `.mockery.yaml`:**
```yaml
  github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/internal/webapi:
    config:
      all: true
      dir: 'flyteplugins/go/tasks/pluginmachinery/internal/webapi/mocks'
```

### 3d. `flyteplugins/go/tasks/pluginmachinery/secret/` (6 directives across 4 files)

**Files and directives:**
- `global_secrets.go:15` — `--name=GlobalSecretProvider`
- `secrets_injector.go:20` — `--name=SecretsInjector`
- `secret_manager_client.go:12` — `--name=AWSSecretManagerClient`
- `secret_manager_client.go:21` — `--name=GCPSecretManagerClient`
- `secret_manager_client.go:28` — `--name=AzureKeyVaultClient`
- `embedded_secret_manager.go:500` — `--name=MockableControllerRuntimeClient`

**Existing mocks:** 6 files in `./mocks/`

**Add to `.mockery.yaml`:**
```yaml
  github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret:
    config:
      all: false
      dir: 'flyteplugins/go/tasks/pluginmachinery/secret/mocks'
    interfaces:
      GlobalSecretProvider:
      SecretsInjector:
      AWSSecretManagerClient:
      GCPSecretManagerClient:
      AzureKeyVaultClient:
      MockableControllerRuntimeClient:
```

### 3e. `actions/service/` (no `go:generate`, but has generated mock not in config)

**File:** `actions/service/mocks/actions_client_interface.go` (generated by mockery, used in tests)
**Interface:** `ActionsClientInterface` defined in `actions/service/interfaces.go:16`

**Add to `.mockery.yaml`:**
```yaml
  github.com/flyteorg/flyte/v2/actions/service:
    config:
      all: false
      dir: 'actions/service/mocks'
    interfaces:
      ActionsClientInterface:
```

---

## Step 4: Replace hand-written fakes

### 4a. Delete unused hand-written fakes

The following 2 hand-written mock files are **not imported or referenced anywhere** in the codebase. Delete them:

```bash
rm flyteplugins/go/tasks/pluginmachinery/core/mocks/fake_k8s_client.go
rm flyteplugins/go/tasks/pluginmachinery/core/mocks/fake_k8s_cache.go
```

### 4b. Replace `mock_cache.go` with mockery-generated mock

`flytestdlib/cache/mocks/mock_cache.go` is a hand-written generic stub for `github.com/eko/gocache/lib/v4/cache.CacheInterface[T]`.
It is used in 9 places in `flyteplugins/go/tasks/pluginmachinery/secret/embedded_secret_manager_test.go`.

**Add to `.mockery.yaml`:**
```yaml
  github.com/eko/gocache/lib/v4/cache:
    config:
      all: false
      dir: 'flytestdlib/cache/mocks'
    interfaces:
      CacheInterface:
```

**Delete the hand-written file:**
```bash
rm flytestdlib/cache/mocks/mock_cache.go
```

**Update test file** `flyteplugins/go/tasks/pluginmachinery/secret/embedded_secret_manager_test.go`:

The import stays the same (`cacheMocks "github.com/flyteorg/flyte/v2/flytestdlib/cache/mocks"`), but each call site needs updating.

**Before** (9 occurrences):
```go
secretCache := cacheMocks.NewMockCache[SecretValue](true)
```

**After:**
```go
secretCache := cacheMocks.NewCacheInterface[SecretValue](t)
secretCache.EXPECT().Get(mock.Anything, mock.Anything).Return(SecretValue{}, fmt.Errorf("cache miss")).Maybe()
secretCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
secretCache.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil).Maybe()
secretCache.EXPECT().Invalidate(mock.Anything).Return(nil).Maybe()
secretCache.EXPECT().Clear(mock.Anything).Return(nil).Maybe()
secretCache.EXPECT().GetType().Return("mock").Maybe()
```

> **Tip:** To reduce repetition, extract a helper function in the test file:
> ```go
> func newAlwaysMissCache(t *testing.T) *cacheMocks.CacheInterface[SecretValue] {
> 	t.Helper()
> 	m := cacheMocks.NewCacheInterface[SecretValue](t)
> 	m.EXPECT().Get(mock.Anything, mock.Anything).Return(SecretValue{}, fmt.Errorf("cache miss")).Maybe()
> 	m.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
> 	m.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil).Maybe()
> 	m.EXPECT().Invalidate(mock.Anything).Return(nil).Maybe()
> 	m.EXPECT().Clear(mock.Anything).Return(nil).Maybe()
> 	m.EXPECT().GetType().Return("mock").Maybe()
> 	return m
> }
> ```
> Then each call site becomes: `secretCache := newAlwaysMissCache(t)`

---

## Step 5: Remove unused mock packages from `.mockery.yaml`

The following 11 mock packages are generated but **never imported by any test or source file**. Remove them from `.mockery.yaml` and delete their mocks directories:

### flytestdlib (3 packages):

| Package in `.mockery.yaml` | Directory |
|---|---|
| `github.com/flyteorg/flyte/v2/flytestdlib/fastcheck` | `flytestdlib/fastcheck/mocks/` |
| `github.com/flyteorg/flyte/v2/flytestdlib/random` | `flytestdlib/random/mocks/` |
| `github.com/flyteorg/flyte/v2/flytestdlib/utils` | `flytestdlib/utils/mocks/` |

### gen/go/flyteidl2 (8 packages):

| Package in `.mockery.yaml` | Directory |
|---|---|
| `github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions` | `gen/go/flyteidl2/actions/mocks/` |
| `github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect` | `gen/go/flyteidl2/actions/actionsconnect/mocks/` |
| `github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app` | `gen/go/flyteidl2/app/mocks/` |
| `github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common` | `gen/go/flyteidl2/common/mocks/` |
| `github.com/flyteorg/flyte/v2/gen/go/flyteidl2/imagebuilder` | `gen/go/flyteidl2/imagebuilder/mocks/` |
| `github.com/flyteorg/flyte/v2/gen/go/flyteidl2/secret` | `gen/go/flyteidl2/secret/mocks/` |
| `github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task` | `gen/go/flyteidl2/task/mocks/` |
| `github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow` | `gen/go/flyteidl2/workflow/mocks/` |

```bash
# Delete unused mock directories
rm -rf flytestdlib/fastcheck/mocks/ \
       flytestdlib/random/mocks/ \
       flytestdlib/utils/mocks/ \
       gen/go/flyteidl2/actions/mocks/ \
       gen/go/flyteidl2/actions/actionsconnect/mocks/ \
       gen/go/flyteidl2/app/mocks/ \
       gen/go/flyteidl2/common/mocks/ \
       gen/go/flyteidl2/imagebuilder/mocks/ \
       gen/go/flyteidl2/secret/mocks/ \
       gen/go/flyteidl2/task/mocks/ \
       gen/go/flyteidl2/workflow/mocks/
```

> **Note:** These packages may have been generated proactively for future use. Confirm with the team before removing — they are easy to re-add later if needed.

---

## Step 6: Remove all `go:generate mockery` directives

Remove the following lines from source files:

| File | Line | Directive |
|------|------|-----------|
| `flytestdlib/autorefreshcache/auto_refresh.go` | 28 | `//go:generate mockery -all` |
| `flyteplugins/go/tasks/pluginmachinery/webapi/plugin.go` | 19 | `//go:generate mockery -all -case=underscore` |
| `flyteplugins/go/tasks/pluginmachinery/internal/webapi/cache.go` | 19 | `//go:generate mockery -all -case=underscore` |
| `flyteplugins/go/tasks/pluginmachinery/secret/global_secrets.go` | 15 | `//go:generate mockery --output=./mocks --case=underscore --name=GlobalSecretProvider` |
| `flyteplugins/go/tasks/pluginmachinery/secret/secrets_injector.go` | 20 | `//go:generate mockery --output=./mocks --case=underscore --name=SecretsInjector` |
| `flyteplugins/go/tasks/pluginmachinery/secret/secret_manager_client.go` | 12 | `//go:generate mockery --output=./mocks --case=underscore -name=AWSSecretManagerClient` |
| `flyteplugins/go/tasks/pluginmachinery/secret/secret_manager_client.go` | 21 | `//go:generate mockery --output=./mocks --case=underscore -name=GCPSecretManagerClient` |
| `flyteplugins/go/tasks/pluginmachinery/secret/secret_manager_client.go` | 28 | `//go:generate mockery --output=./mocks --case=underscore -name=AzureKeyVaultClient` |
| `flyteplugins/go/tasks/pluginmachinery/secret/embedded_secret_manager.go` | 500 | `//go:generate mockery -name=MockableControllerRuntimeClient -output=./mocks -case=underscore` |

---

## Step 7: Regenerate all mocks and verify

```bash
# Step 7a: Delete all remaining generated mock files (recursive from repo root)
find . -path '*/mocks/*.go' -exec grep -l "Code generated by mockery" {} + | xargs rm -f

# Step 7b: Regenerate using new v3 config
mockery

# Step 7c: Verify compilation
go build ./...

# Step 7d: Run tests
go test ./...
```

---

## Step 8: Update Makefile (if needed)

The `Makefile` (line 149) already runs `mockery` without arguments (uses `.mockery.yaml`), so no changes should be needed there. Verify the `mocks` target still works:

```bash
make mocks
```

---

## Step 9: Clean up migrator artifacts

```bash
# Remove the migrator output if it was created separately
rm -f .mockery_v3.yml
```

---

## Step 10: Run full test suite

Final validation that the entire migration is correct — build and test everything:

```bash
# Ensure everything compiles
go build ./...

# Run all tests
go test ./...
```

Fix any failures before committing. Common things to watch for:
- **Mock constructor signature changes** — v3 constructors require `*testing.T` as argument
- **Missing expectations** — v3 mocks with expecter enabled will fail on unexpected calls unless `.Maybe()` is used
- **Function type mocks removed** — v3 skips function types; if any test relied on a function mock, it will get a compile error

---

## Summary of files to modify

| File | Action |
|------|--------|
| `.mockery.yaml` | Replace with `mockery migrate` output, add 6 packages (5 from Step 3 + gocache from Step 4b), remove 11 unused packages |
| `gen.Dockerfile` | Update version `2.53.5` → `3.7.0`, change module path `v2` → `v3` |
| `boilerplate/flyte/golang_support_tools/go.mod` | `mockery/v2` → `mockery/v3`, run `go mod tidy` |
| `boilerplate/flyte/golang_support_tools/tools.go` | `mockery/v2/cmd` → `mockery/v3/cmd` |
| `boilerplate/flyte/golang_test_targets/download_tooling.sh` | `mockery/v2@v2.52.1` → `mockery/v3@v3.7.0` |
| `flytestdlib/autorefreshcache/auto_refresh.go` | Remove `go:generate` line |
| `flyteplugins/go/tasks/pluginmachinery/webapi/plugin.go` | Remove `go:generate` line |
| `flyteplugins/go/tasks/pluginmachinery/internal/webapi/cache.go` | Remove `go:generate` line |
| `flyteplugins/go/tasks/pluginmachinery/secret/global_secrets.go` | Remove `go:generate` line |
| `flyteplugins/go/tasks/pluginmachinery/secret/secrets_injector.go` | Remove `go:generate` line |
| `flyteplugins/go/tasks/pluginmachinery/secret/secret_manager_client.go` | Remove 3 `go:generate` lines |
| `flyteplugins/go/tasks/pluginmachinery/secret/embedded_secret_manager.go` | Remove `go:generate` line |
| `flyteplugins/go/tasks/pluginmachinery/secret/embedded_secret_manager_test.go` | Update 9 `NewMockCache` calls to use mockery-generated `NewCacheInterface` |

### Files/directories to delete

| Path | Reason |
|------|--------|
| `flyteplugins/go/tasks/pluginmachinery/core/mocks/fake_k8s_client.go` | Unused hand-written fake |
| `flyteplugins/go/tasks/pluginmachinery/core/mocks/fake_k8s_cache.go` | Unused hand-written fake |
| `flytestdlib/cache/mocks/mock_cache.go` | Replace with mockery-generated `CacheInterface` mock |
| `flytestdlib/fastcheck/mocks/` | Unused generated mocks (no imports) |
| `flytestdlib/random/mocks/` | Unused generated mocks (no imports) |
| `flytestdlib/utils/mocks/` | Unused generated mocks (no imports) |
| `gen/go/flyteidl2/actions/mocks/` | Unused generated mocks (no imports) |
| `gen/go/flyteidl2/actions/actionsconnect/mocks/` | Unused generated mocks (no imports) |
| `gen/go/flyteidl2/app/mocks/` | Unused generated mocks (no imports) |
| `gen/go/flyteidl2/common/mocks/` | Unused generated mocks (no imports) |
| `gen/go/flyteidl2/imagebuilder/mocks/` | Unused generated mocks (no imports) |
| `gen/go/flyteidl2/secret/mocks/` | Unused generated mocks (no imports) |
| `gen/go/flyteidl2/task/mocks/` | Unused generated mocks (no imports) |
| `gen/go/flyteidl2/workflow/mocks/` | Unused generated mocks (no imports) |

---

## Risks and considerations

1. **Generated mock API changes**: v3 may generate slightly different mock code (e.g., always includes expecter methods). Run full test suite to catch any incompatibilities.
2. **Function type mocks**: v3 no longer generates mocks for function types. The `webapi` package has `PluginLoader` which is a function type — verify this doesn't break anything.
3. **`include-auto-generated`**: Verify this flag still exists in v3 for the protobuf-generated interface packages.
4. **`actions/service/mocks`**: This package has a generated mock with no `go:generate` directive and was missing from `.mockery.yaml`. It must be added (Step 3e) or regeneration will leave it missing and break `actions/service/actions_service_test.go`.
5. **Unused mock removal**: The 11 unused mock packages (Step 5) may have been generated proactively for future use. Confirm with the team before removing — they are easy to re-add to `.mockery.yaml` later if needed.
