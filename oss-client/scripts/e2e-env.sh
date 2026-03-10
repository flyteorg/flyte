#!/bin/sh
set -e

# Usage:
#   . ./scripts/e2e-env.sh
#   load_e2e_env   # sets HOST, TEST_ORG, SECRET_ID based on E2E_ENV

load_e2e_env() {
  # E2E_ENV determines which environment to test against
  # Set this single value in Buildkite to switch environments
  E2E_ENV="${E2E_ENV:-playground.canary}"

  echo "Configuring for environment: $E2E_ENV"

  case "$E2E_ENV" in
    playground.canary)
      HOST="playground.canary.unionai.cloud"
      TEST_ORG="playground"
      SECRET_ID="frontend/e2e/canary"
      ;;
    demo.hosted)
      HOST="demo.hosted.unionai.cloud"
      TEST_ORG="demo"
      SECRET_ID="frontend/e2e/demo"
      ;;
    *)
      echo "Unknown E2E_ENV: $E2E_ENV"
      echo "Supported environments: playground.canary, demo.hosted"
      exit 1
      ;;
  esac

  # make them available to the caller
  export E2E_ENV HOST TEST_ORG SECRET_ID
}
