#!/bin/bash
# Helper script to run commands in the CI Docker container
# Usage:
#   ./scripts/docker-dev.sh          # Start interactive shell
#   ./scripts/docker-dev.sh pull     # Pull latest image
#   ./scripts/docker-dev.sh make gen # Run make gen

set -e

IMAGE="ghcr.io/flyteorg/flyte/ci:v2"
WORKSPACE="/workspace"

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the repository root (parent of scripts directory)
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Function to print colored messages
info() {
    echo -e "${BLUE}🐳 $1${NC}"
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Handle special commands
case "$1" in
    pull)
        info "Pulling latest CI image..."
        docker pull "$IMAGE"
        success "Image pulled successfully"
        exit 0
        ;;
    help|--help|-h)
        echo "Docker Development Helper"
        echo ""
        echo "Usage:"
        echo "  $0                  Start interactive shell in CI container"
        echo "  $0 pull             Pull the latest CI image"
        echo "  $0 <command>        Run a command in the CI container"
        echo ""
        echo "Examples:"
        echo "  $0                  # Interactive shell"
        echo "  $0 make gen         # Generate protocol buffers"
        echo "  $0 buf lint         # Lint protocol buffers"
        echo "  $0 cargo build      # Build Rust crate"
        exit 0
        ;;
esac

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker and try again."
    exit 1
fi

# Run interactively if no command provided
if [ $# -eq 0 ]; then
    info "Starting interactive shell..."
    docker run --rm -it \
        -v "$REPO_ROOT:$WORKSPACE" \
        -w "$WORKSPACE" \
        "$IMAGE" \
        bash
else
    # Run the provided command
    info "Running: $*"
    docker run --rm \
        -v "$REPO_ROOT:$WORKSPACE" \
        -w "$WORKSPACE" \
        "$IMAGE" \
        "$@"
fi
