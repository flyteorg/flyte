#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
LIMIT="${LIMIT:-10}"
TOKEN="${TOKEN:-}"

# List runs by project with basic pagination support
buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/ListRuns" --data @- <<EOF
{
    "request": {
        "limit": $LIMIT,
        "token": "$TOKEN"
    },
    "project_id": {
        "organization": "$ORG",
        "name": "$PROJECT",
        "domain": "$DOMAIN"
    }
}
EOF
