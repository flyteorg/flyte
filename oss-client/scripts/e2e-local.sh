# e2e-local.sh
#!/bin/bash
set -euo pipefail

# add environment variables from .env
export $(grep -v '^#' .env | xargs)

echo "🔧 Setting up local Flyte 2 environment with uv..."

# Ensure uv is available
if ! command -v uv &> /dev/null; then
  echo "Installing uv..."
  pip install uv
fi

export UV_VENV_CLEAR=1
uv python install 3.13
uv venv .venv
source .venv/bin/activate

# Ensure pip inside
python -m ensurepip --upgrade

echo "📦 Installing Flyte 2 (Union BYOC) SDK..."
uv pip install --no-cache --prerelease=allow --upgrade flyte

echo "📦 Installing FastAPI and dependencies..."
uv pip install fastapi uvicorn

echo "🐍 Running seed scripts..."
python e2e/scripts/seed_run_with_groups.py
python e2e/scripts/seed_basic_task.py
python e2e/scripts/seed_basic_app.py
python e2e/scripts/seed_trigger.py

echo "✅ Seeds complete. Run your Playwright tests with:"
echo "   pnpm exec playwright test --project=chromium"
