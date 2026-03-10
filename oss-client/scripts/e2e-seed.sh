#!/bin/sh
set -e

if ! command -v uv; then
 pip install uv
fi

# Load shared env mapping util
. "$(dirname "$0")/e2e-env.sh"

load_e2e_env

BASE_URL="https://$HOST"
TEST_PROJECT="${TEST_PROJECT:-e2e}"
TEST_DOMAIN="${TEST_DOMAIN:-development}"

echo "  HOST: $HOST"
echo "  BASE_URL: $BASE_URL"
echo "  TEST_ORG: $TEST_ORG"
echo "  SECRET_ID: $SECRET_ID"

echo "Pulling down secrets..."
AWS_REGION=us-east-2 aws secretsmanager get-secret-value --region us-east-2 --secret-id "${SECRET_ID}" --query SecretString --output text > secrets.json

# Parse the JSON and export each key as an environment variable (secrets only)
if command -v jq >/dev/null 2>&1; then
  export $(jq -r 'to_entries|map("\(.key)=\(.value|tostring)")|.[]' secrets.json)
else
  echo "jq is required to parse secrets.json. Please install jq."
  exit 1
fi

# Export our derived config values (override any from secrets for consistency)
export HOST
export BASE_URL
export TEST_ORG
export TEST_PROJECT
export TEST_DOMAIN

# if secrets exported include SKIP_E2E, skip tests. this is intended to be used rarely
if [ "$(printf '%s' "$SKIP_E2E" | tr '[:upper:]' '[:lower:]')" = "true" ]; then
  echo "SKIP_E2E=true. Skipping E2E tests."
  exit 0
fi


echo "Installing Python 3.13 and creating venv at /opt/venv"
export UV_VENV_CLEAR=1
uv python install 3.13
uv python pin 3.13 --global
uv venv /opt/venv

# Activate the venv explicitly
. /opt/venv/bin/activate

# Ensure pip exists inside the venv
/opt/venv/bin/python -m ensurepip --upgrade
# Install Flyte 2 packages from the official Union v2 index
echo "Installing Flyte 2..."
uv pip install --no-cache --prerelease=allow --upgrade flyte
echo "Installing FastAPI and dependencies..."
uv pip install fastapi uvicorn

# Run seed scripts using the venv Python
echo "Running seed scripts..."
/opt/venv/bin/python e2e/scripts/seed_run_with_groups.py
/opt/venv/bin/python e2e/scripts/seed_basic_task.py
/opt/venv/bin/python e2e/scripts/seed_basic_app.py
/opt/venv/bin/python e2e/scripts/seed_trigger.py
# /opt/venv/bin/python e2e/scripts/seed_fanout_with_failures.py

deactivate 2>/dev/null || true
echo "✅ Done seeding"
exit 0
