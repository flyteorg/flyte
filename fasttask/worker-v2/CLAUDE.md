# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This project is comprised of two parts - the bridge, and the executor - the bridge subprocesses the exector and communicates over a TCP socket.

## Architecture

This is a Rust/Python hybrid project that implements Union.ai Actors v2. The system consists of:

- **Bridge** (`bridge/`): Rust binary (`unionai-actor-bridge`) that manages communication and subprocess execution via TCP sockets
- **Executor** (`executor/`): Rust library compiled as a Python extension (`_lib`) that handles task execution
- **Python Package** (`python_src/unionai_actor_executor/`): Python wrapper and entry points

The workspace uses Cargo for Rust dependency management and setuptools-rust for Python packaging integration.

## Development Setup

1. **Environment Setup**:
   ```bash
   uv venv --python 3.13
   source .venv/bin/activate
   ```

2. **Development Dependencies**:
   ```bash
   uv sync --reinstall
   uv pip install --pre flyte
   ```

## Common Commands

### Rust Development
- **Check Rust code**: `cargo check`
- **Build Rust components**: `cargo build --release`

### Python Development  
- **Build a local wheel**: `python -m build --wheel --outdir dist`

## Key Files

- `pyproject.toml`: Main Python package configuration with setuptools-rust integration
- `Cargo.toml`: Workspace configuration for Rust components
- `Makefile`: Wheel building automation with Docker cross-compilation
- `bridge/Cargo.toml`: Bridge binary configuration
- `executor/Cargo.toml`: Executor library configuration (PyO3 extension)

## Project Structure

- Entry points: `unionai-actor-executor` (Python), `unionai-actor-tester` (Python), `unionai-actor-bridge` (Rust binary)
- The executor creates a `_lib.cpython-*.so` file that gets installed as part of the Python package
- Version management uses setuptools-scm with git integration (root: `../..`)
- Cross-platform wheel building via Docker with maturin

## Testing Integration

Tests require an internal sandbox environment that may or may not be active.

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.
