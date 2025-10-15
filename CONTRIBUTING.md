# Contributing to Flyte 2

Thank you for your interest in contributing to Flyte 2! This guide will help you understand the contribution workflow, testing requirements, and release process.

## Table of Contents
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Making Changes](#making-changes)
- [Testing and Verification](#testing-and-verification)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)
- [Best Practices](#best-practices)

## Getting Started

### Prerequisites

Before contributing, ensure you have:
- [Buf CLI](https://buf.build/docs/installation) installed
- Go 1.24.6 or later
- Node.js and npm (for TypeScript)
- Python 3.9+ with `uv` package manager
- Rust toolchain (if working with Rust bindings)
- Git configured with your name and email

### Setting Up Your Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/flyte.git
   cd flyte/flyte
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/flyteorg/flyte.git
   ```

4. **Verify your setup**:
   ```bash
   make gen
   ```

## Development Workflow

### Creating a Feature Branch

Always create a new branch for your changes:

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/add-new-task-type` - for new features
- `fix/workflow-literal-bug` - for bug fixes
- `docs/improve-readme` - for documentation
- `refactor/simplify-interface` - for refactoring

### Keeping Your Branch Updated

Regularly sync with upstream:

```bash
git fetch upstream
git rebase upstream/v2
```

## Making Changes

### 1. Modifying Protocol Buffers

When editing `.proto` files in `flyteidl2/`:

**Naming Conventions:**
- Use `snake_case` for field names: `task_id`, `execution_time`
- Use `PascalCase` for message names: `TaskDefinition`, `WorkflowSpec`
- Use `SCREAMING_SNAKE_CASE` for enum values: `TASK_STATE_RUNNING`

**Backward Compatibility:**
- Never change field numbers
- Never change field types
- Never remove required fields
- Use `reserved` for removed fields:
  ```protobuf
  message Example {
    reserved 2, 15, 9 to 11;
    reserved "old_field_name";
  }
  ```

**Documentation:**
- Add clear comments for all messages, fields, and enums
- Include examples where helpful
- Document any constraints or validation rules

**Example:**
```protobuf
// TaskDefinition defines the structure and configuration of a task.
message TaskDefinition {
  // Unique identifier for the task
  string task_id = 1;

  // Human-readable name for the task
  string name = 2;

  // Optional description explaining the task's purpose
  string description = 3;
}
```

### 2. Generate and Format Code

After making changes:

```bash
# Format proto files
make buf-format

# Lint proto files
make buf-lint

# Generate all language bindings
make buf
```

### 3. Update Language-Specific Utilities (if needed)

If you need to customize generation for a specific language, update files in `flyteidl2/gen_utils/`:
- `flyteidl2/gen_utils/go/` - Go-specific utilities
- `flyteidl2/gen_utils/ts/` - TypeScript utilities
- `flyteidl2/gen_utils/python/` - Python package configuration
- `flyteidl2/gen_utils/rust/` - Rust crate configuration

## Testing and Verification

### 1. Lint and Format Checks

```bash
make buf-lint
make buf-format
```

All proto files must pass linting without errors.

### 2. Verify Generated Code Builds

**Go:**
```bash
make buf-go
make go-tidy
cd gen/go
go build ./...
go test ./...
```

**TypeScript:**
```bash
make buf-ts
cd gen/ts
npm install
npm run build  # if there's a build script
```

**Python:**
```bash
make buf-python
cd gen/python
uv lock
uv sync
```

**Rust:**
```bash
make buf-rust
make build-crate
```

### 3. Generate Mocks (Go)

If you've modified Go interfaces:

```bash
make mocks
```

### 4. Integration Testing

Test your changes in a downstream project that uses flyte:
1. Use `replace` directive in `go.mod` to point to your local changes
2. Generate code and verify it works as expected
3. Run integration tests in the consuming project

## Submitting Changes

### 1. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -s -m "feat: add new task execution state

- Add TASK_STATE_CACHED for cached task results
- Update state transitions to support caching
- Add documentation for caching behavior

Fixes #123"
```

**Commit message format:**
- Use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
- Include `-s` flag to sign your commits (required)
- Reference issue numbers with `Fixes #123` or `Closes #456`

### 2. Push Your Changes

```bash
git push origin feature/your-feature-name
```

### 3. Create a Pull Request

1. Go to the [Flyte repository](https://github.com/flyteorg/flyte)
2. Click "New Pull Request"
3. Select your fork and branch
4. Target the `v2` branch (not `main`)
5. Fill out the PR template with:
   - Description of changes
   - Motivation and context
   - Testing performed
   - Screenshots (if UI changes)
   - Breaking changes (if any)

### 4. Code Review Process

- Address reviewer feedback promptly
- Push new commits to the same branch
- Use `git commit --amend` for small fixes (then `git push --force`)
- Engage in constructive discussion
- Be patient - reviews may take time

## Release Process

Releases are created by maintainers following these steps:

### Version Management

Flyte 2 follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** version (v2.0.0 → v3.0.0): Breaking changes
- **MINOR** version (v2.1.0 → v2.2.0): New features, backward compatible
- **PATCH** version (v2.1.1 → v2.1.2): Bug fixes, backward compatible

### Creating a Release (Maintainers Only)

1. **Ensure all changes are merged into `v2` branch**

2. **Determine the version number** based on changes:
   ```bash
   # View changes since last release
   git log v2.X.Y..HEAD --oneline
   ```

3. **Create and push a tag**:
   ```bash
   git checkout v2
   git pull upstream v2
   git tag -a v2.X.Y -m "Release v2.X.Y"
   git push upstream v2.X.Y
   ```

4. **Create GitHub Release**:
   - Go to [Releases](https://github.com/flyteorg/flyte/releases)
   - Click "Draft a new release"
   - Select the tag you just created
   - Generate release notes
   - Add any additional context or highlights
   - Publish release

5. **Verify Artifacts**:
   The release triggers automated publishing to:
   - Go modules: `github.com/flyteorg/flyte/v2`
   - NPM: `@flyteorg/flyte`
   - PyPI: `flyteidl2`
   - Crates.io: `flyte`

### Post-Release

1. **Verify published packages**:
   ```bash
   # Go
   GOPROXY=https://proxy.golang.org go list -m github.com/flyteorg/flyte/v2@v2.X.Y

   # NPM
   npm view @flyteorg/flyte@2.X.Y

   # PyPI
   pip index versions flyteidl2

   # Crates.io
   cargo search flyte
   ```

2. **Update dependent projects** with the new version

3. **Announce the release** in community channels

## Best Practices

### Protocol Buffer Design

- **Keep messages small and focused** - Single responsibility principle
- **Use oneof for mutually exclusive fields**:
  ```protobuf
  message Task {
    oneof task_type {
      PythonTask python = 1;
      ContainerTask container = 2;
    }
  }
  ```
- **Use optional for truly optional fields** (proto3)
- **Use repeated for arrays/lists**
- **Provide sensible defaults** where appropriate

### Documentation

- Document all public APIs
- Include usage examples in comments
- Keep README.md and CONTRIBUTING.md up to date
- Add inline comments for complex logic

### Git Workflow

- Commit early and often
- Keep commits atomic and focused
- Write meaningful commit messages
- Squash small "fix" commits before PR
- Rebase instead of merge when updating your branch

### Communication

- Be respectful and professional
- Ask questions when unclear
- Provide context in discussions
- Update PR descriptions as scope changes
- Respond to reviews in a timely manner

## Getting Help

- **Documentation**: [Flyte Documentation](https://docs.flyte.org)
- **Community Slack**: [Flyte Slack](https://slack.flyte.org)
- **GitHub Issues**: [Report Issues](https://github.com/flyteorg/flyte/issues)
- **GitHub Discussions**: [Ask Questions](https://github.com/flyteorg/flyte/discussions)

## License

By contributing to Flyte 2, you agree that your contributions will be licensed under the Apache License 2.0.
