#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
RUN_NAME="${RUN_NAME:?Error: RUN_NAME environment variable is required}"
LIMIT="${LIMIT:-10}"
TOKEN="${TOKEN:-}"

buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/ListActions" --data @- <<EOF
{
    "request": {
        "limit": $LIMIT,
        "token": "$TOKEN"
    },
    "run_id": {
        "org": "$ORG",
        "project": "$PROJECT",
        "domain": "$DOMAIN",
        "name": "$RUN_NAME"
    }
}
EOF
