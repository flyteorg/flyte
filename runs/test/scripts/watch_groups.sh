#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
LIMIT="${LIMIT:-20}"
START_DATE="${START_DATE:-2026-01-01T00:00:00Z}"
END_DATE="${END_DATE:-}"

if [ -n "$END_DATE" ]; then
buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/WatchGroups" --data @- <<EOF
{
    "project_id": {
        "organization": "$ORG",
        "name": "$PROJECT",
        "domain": "$DOMAIN"
    },
    "start_date": "$START_DATE",
    "end_date": "$END_DATE",
    "request": {
        "limit": $LIMIT
    }
}
EOF
else
buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/WatchGroups" --data @- <<EOF
{
    "project_id": {
        "organization": "$ORG",
        "name": "$PROJECT",
        "domain": "$DOMAIN"
    },
    "start_date": "$START_DATE",
    "request": {
        "limit": $LIMIT
    }
}
EOF
fi
