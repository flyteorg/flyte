#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
TASK_NAME="${TASK_NAME:-my_task}"
VERSION="${VERSION:-1}"

# Test 1: Auto-generated run name
echo "Test 1: Auto-generated run name (project_id + task_spec)"
buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/CreateRun" --data @- <<EOF
{
  "project_id": {
    "organization": "$ORG",
    "name": "$PROJECT",
    "domain": "$DOMAIN"
  },
  "task_spec": {
    "short_name": "test"
  },
  "source": "RUN_SOURCE_CLI"
}
EOF

echo ""
echo "---"
echo ""

# Test 2: Custom run name
echo "Test 2: Custom run name (run_id + task_id)"
buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/CreateRun" --data @- <<EOF
{
  "run_id": {
    "org": "$ORG",
    "project": "$PROJECT",
    "domain": "$DOMAIN",
    "name": "my-custom-run-123"
  },
  "task_id": {
    "org": "$ORG",
    "project": "$PROJECT",
    "domain": "$DOMAIN",
    "name": "$TASK_NAME",
    "version": "$VERSION"
  },
  "source": "RUN_SOURCE_WEB"
}
EOF