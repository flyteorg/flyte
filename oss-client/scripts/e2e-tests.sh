#!/bin/sh
set -e
echo "Starting test script"

if ! command -v pnpm >/dev/null 2>&1; then
    npm install -g pnpm  >/dev/null
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

echo "pulling down secrets "
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

# Pull the Slack API token from its dedicated secret.
# Soft-fails so local runs and environments without this secret still work.
SLACK_APP_TOKEN=$(AWS_REGION=us-east-2 aws secretsmanager get-secret-value --region us-east-2 \
  --secret-id "buildkite/slack-api-token" --query SecretString --output text 2>/dev/null \
  | jq -r '."buildkite/slack-api-token"' 2>/dev/null || true)
export SLACK_APP_TOKEN

# Install project dependencies first so postMessageToSlack.mjs is available for all notifications below.
echo "installing dependencies"
pnpm install > /dev/null

BUILD_INFO="${BUILDKITE_BUILD_URL:-(local run)}"

# if secrets exported included SKIP_E2E, skip tests. this is intended to be used rarely
if [ "$(printf '%s' "$SKIP_E2E" | tr '[:upper:]' '[:lower:]')" = "true" ]; then
  echo "SKIP_E2E=true. Skipping E2E tests."
  node ./scripts/postMessageToSlack.mjs -m ":no_entry_sign: Playwright e2e tests skipped (SKIP_E2E=true) for ${E2E_ENV}: ${BUILD_INFO}"
  exit 0
fi

export DEV_HOST="localhost.${HOST}"
export DEV_PORT=8080
HEALTH_URL="https://${DEV_HOST}:${DEV_PORT}/v2"

mkdir -p ./scripts/certificate

openssl req -x509 -out ./scripts/certificate/server.crt -keyout ./scripts/certificate/server.key \
    -newkey rsa:2048 -nodes -sha256 -days 365 \
    -subj "/CN=${DEV_HOST}" \
    -addext "subjectAltName=DNS:${DEV_HOST},DNS:localhost,DNS:127.0.0.1"

npx playwright install-deps > /dev/null
pnpm exec playwright install > /dev/null
echo "127.0.0.1 ${DEV_HOST}" >> /etc/hosts

export PLAYWRIGHT_HTML_OPEN='never'

# Create a temporary log file
LOG_FILE=$(mktemp)

export NODE_ENV=test
export NEXT_PUBLIC_ADMIN_DOMAIN="$BASE_URL"
# Set API URL - handle BASE_URL with or without https:// prefix
case "$BASE_URL" in
  https://*) export NEXT_PUBLIC_ADMIN_API_URL="$BASE_URL" ;;
  *) export NEXT_PUBLIC_ADMIN_API_URL="https://$BASE_URL" ;;
esac
export PORT="$DEV_PORT"

# Start the dev server and capture output
pnpm dev:ci > "$LOG_FILE" 2>&1 &
DEV_PID=$!

echo "Starting dev server (PID: $DEV_PID)..."
echo "Log file: $LOG_FILE"

# Function to cleanup on exit - more aggressive cleanup
cleanup() {
    echo "Stopping dev server..."

    # Kill the main process
    if kill -0 $DEV_PID 2>/dev/null; then
        kill $DEV_PID 2>/dev/null || true
        sleep 2

        # Force kill if still running
        if kill -0 $DEV_PID 2>/dev/null; then
            kill -9 $DEV_PID 2>/dev/null || true
        fi
    fi

    # Kill any remaining pnpm/node processes on the dev port
    pkill -f "pnpm dev" 2>/dev/null || true
    pkill -f "next-server" 2>/dev/null || true

    # Find and kill any process using the dev port
    PORT_PID=$(lsof -ti:"${DEV_PORT}" 2>/dev/null || true)
    if [ -n "$PORT_PID" ]; then
        echo "Killing process on port ${DEV_PORT}: $PORT_PID"
        kill -9 $PORT_PID 2>/dev/null || true
    fi

    rm -f "$LOG_FILE"
    exit
}

trap cleanup EXIT INT TERM

# Wait for "Ready" message in logs
echo "Waiting for dev server to be ready..."
timeout=120  # Increased timeout
counter=0

while [ $counter -lt $timeout ]; do
    if ! kill -0 $DEV_PID 2>/dev/null; then
        echo "Dev server process died unexpectedly"
        echo "Last few lines of log:"
        tail -n 50 "$LOG_FILE"
        exit 1
    fi

    # Check for Next.js ready messages
    if grep -q -E "(Ready in|ready|Local:|Network:)" "$LOG_FILE"; then
        echo "Dev server is ready!"
        echo "Server output:"
        cat "$LOG_FILE"
        break
    fi

    sleep 1  # Check more frequently
    counter=$((counter + 1))
done

if [ $counter -eq $timeout ]; then
    echo "Timeout waiting for dev server to start"
    echo "Full log output:"
    cat "$LOG_FILE"
    exit 1
fi

# Additional wait to ensure server is fully stable
echo "Waiting additional 5 seconds for server to fully stabilize..."
sleep 5

echo "Verifying server is responding..."
HEALTH_CHECK_ATTEMPTS=0
MAX_HEALTH_ATTEMPTS=15

while [ "$HEALTH_CHECK_ATTEMPTS" -lt "$MAX_HEALTH_ATTEMPTS" ]; do
  # If dev server process died, bail out and show logs
  if ! kill -0 "$DEV_PID" 2>/dev/null; then
    echo "Dev server process ($DEV_PID) died during health checks."
    echo "===== Dev server log ====="
    cat "$LOG_FILE"
    exit 1
  fi

  HTTP_CODE=$(curl -k -sS -o /dev/null -w '%{http_code}' "$HEALTH_URL" || echo "curl_error")
  echo "Health attempt $((HEALTH_CHECK_ATTEMPTS+1)) - $HEALTH_URL -> $HTTP_CODE"

  if [ "$HTTP_CODE" != "curl_error" ] && [ "$HTTP_CODE" -ge 200 ] && [ "$HTTP_CODE" -lt 400 ]; then
    echo "Server responded successfully."
    break
  fi

  echo "Server not responding yet, waiting... (attempt $((HEALTH_CHECK_ATTEMPTS+1))/$MAX_HEALTH_ATTEMPTS)"
  HEALTH_CHECK_ATTEMPTS=$((HEALTH_CHECK_ATTEMPTS+1))
  sleep 1
done

if [ "$HEALTH_CHECK_ATTEMPTS" -eq "$MAX_HEALTH_ATTEMPTS" ]; then
  echo "Server failed health checks."
  echo "===== Dev server log ====="
  cat "$LOG_FILE"
  exit 1
fi

# Additional buffer time before starting tests
echo "Waiting additional 5 seconds before starting tests..."
sleep 5

echo "Checking dev server process:"
ps -p "$DEV_PID" -o pid,ppid,cmd || echo "Dev server process not found"


# Run the tests, capturing exit code so we can post a Slack notification regardless of outcome.
echo "Starting e2e tests..."
set +e
#  specify chrome only
pnpm test:e2e --project=chromium
TEST_EXIT_CODE=$?
set -e

echo "Tests completed with exit code: $TEST_EXIT_CODE"

# Detect skipped tests from the JSON report written by playwright.config.ts on CI.
SKIPPED_COUNT=0
if [ -f "playwright-report/results.json" ]; then
  SKIPPED_COUNT=$(jq '.stats.skipped // 0' playwright-report/results.json 2>/dev/null || echo "0")
fi

if [ "$TEST_EXIT_CODE" -ne 0 ]; then
  node ./scripts/postMessageToSlack.mjs -m ":rotating_light: Playwright e2e tests *failed* for ${E2E_ENV}: ${BUILD_INFO}"
elif [ "$SKIPPED_COUNT" -gt 0 ]; then
  node ./scripts/postMessageToSlack.mjs -m ":large_yellow_circle: Playwright e2e tests passed with ${SKIPPED_COUNT} skipped test(s) for ${E2E_ENV}: ${BUILD_INFO}"
else
  node ./scripts/postMessageToSlack.mjs -m ":large_green_circle: Playwright e2e tests *passed* for ${E2E_ENV}: ${BUILD_INFO}"
fi

exit $TEST_EXIT_CODE

