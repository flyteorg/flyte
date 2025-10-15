# Flyte 2

Flyte 2 is a next-generation IDL (Interface Definition Language) repository that defines the protocol buffer schemas for Flyte's APIs. This repository generates client libraries and type definitions for Go, TypeScript, Python, and Rust.

## Repository Structure

```
flyte/
├── flyteidl2/           # Protocol buffer definitions
│   ├── common/          # Common types and utilities
│   ├── core/            # Core Flyte types (tasks, workflows, literals)
│   ├── imagebuilder/    # Image builder service definitions
│   ├── logs/            # Logging types
│   ├── secret/          # Secret management types
│   ├── task/            # Task execution types
│   ├── trigger/         # Trigger service definitions
│   ├── workflow/        # Workflow types
│   └── gen_utils/       # Language-specific generation utilities
├── gen/                 # Generated code (not checked into version control)
│   ├── go/              # Generated Go code
│   ├── ts/              # Generated TypeScript code
│   ├── python/          # Generated Python code
│   └── rust/            # Generated Rust code
├── buf.yaml             # Buf configuration
├── buf.gen.*.yaml       # Language-specific generation configs
└── Makefile             # Build automation
```

## Prerequisites

- [Buf CLI](https://buf.build/docs/installation) - Protocol buffer tooling
- Go 1.24.6 or later
- Node.js/npm (for TypeScript generation)
- Python 3.9+ with `uv` package manager (for Python generation)
- Rust toolchain (for Rust generation)

## Quick Start

### Generate All Code

To generate code for all supported languages:

```bash
make gen
```

This will:
1. Update buf dependencies
2. Format and lint proto files
3. Generate code for Go, TypeScript, Python, and Rust
4. Generate mocks for Go
5. Run `go mod tidy`

### Generate for Specific Languages Locally

```bash
make buf-go      # Generate Go code only
make buf-ts      # Generate TypeScript code only
make buf-python  # Generate Python code only
make buf-rust    # Generate Rust code only
```

## Making Changes

### 1. Modify Protocol Buffers

Edit `.proto` files in the `flyteidl2/` directory following these guidelines:
- Follow the existing naming conventions
- Use proper protobuf style (snake_case for fields, PascalCase for messages)
- Add appropriate comments and documentation
- Ensure backward compatibility when modifying existing messages

### 2. Generate Code

After modifying proto files:

```bash
make gen
```

### 3. Verify Your Changes

Run the following to ensure everything builds correctly:

```bash
# For Go
make go-tidy
go build ./...

# For Rust
make build-crate

# For Python
cd gen/python && uv lock

# For TypeScript
cd gen/ts && npm install
```

### 4. Generate Mocks (Go only)

If you've added or modified Go interfaces:

```bash
make gen
```

## Development Workflow

1. **Format proto files**: `make buf-format`
2. **Lint proto files**: `make buf-lint`
3. **Generate code**: `make buf` or `make gen`
4. **Verify builds**: Build generated code in your target language
5. **Commit changes**: Commit both proto files and generated code

## Common Tasks

### Update Buf Dependencies

```bash
make gen
```

### View Available Commands

```bash
make help
```

## Versioning and Releases

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed release instructions.

## Generated Code

The `gen/` directory contains auto-generated code and should not be manually edited. Changes to generated code should be made by:
1. Modifying the source `.proto` files in `flyteidl2/`
2. Updating generation utilities in `flyteidl2/gen_utils/` if needed
3. Running `make gen` to regenerate all code

## Troubleshooting

### Buf Errors
- Ensure you have the latest version of Buf: `buf --version`
- Update dependencies: `make buf-dep`
- Check `buf.lock` for dependency conflicts

### Go Module Issues
- Run `make go-tidy` to clean up dependencies
- Ensure you're using Go 1.24.6 or later

### Python Generation Issues
- Ensure `uv` is installed: `pip install uv`
- Set the environment variable: `export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0`

### Rust Build Issues
- Update Rust toolchain: `rustup update`
- Navigate to `gen/rust` and run `cargo update`

## Contributing

We welcome contributions to Flyte 2! Please follow the guide [here](CONTRIBUTING.md).