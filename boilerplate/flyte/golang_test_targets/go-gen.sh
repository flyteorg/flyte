#!/usr/bin/env bash

set -ex

echo "Running go generate"
go generate ./...
# Only run go mod tidy when explicitly requested (avoids incidental go.sum/go.mod churn from make generate)
if [ -n "${RUN_GO_TIDY:-}" ]; then
  go mod tidy
fi
# This section is used by GitHub workflow to ensure that the generation step was run
if [ -n "$DELTA_CHECK" ]; then
  DIRTY=$(git status --porcelain)
  if [ -n "$DIRTY" ]; then
    echo "FAILED: Go code updated without committing generated code."
    echo "Ensure make generate has run and all changes are committed."
    DIFF=$(git diff)
    echo "diff detected: $DIFF"
    DIFF=$(git diff --name-only)
    echo "files different: $DIFF"
    exit 1
  else
    echo "SUCCESS: Generated code is up to date."
  fi
fi
