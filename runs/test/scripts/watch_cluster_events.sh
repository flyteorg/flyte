#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
RUN_NAME="${RUN_NAME:?Error: RUN_NAME environment variable is required}"
ACTION="${ACTION:-$RUN_NAME}"
ATTEMPT="${ATTEMPT:-1}"

buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/WatchClusterEvents" --data @- <<EOF
{
    "id": {
        "run": {
            "org": "$ORG",
            "project": "$PROJECT",
            "domain": "$DOMAIN",
            "name": "$RUN_NAME"
        },
        "name": "$ACTION"
    },
    "attempt": $ATTEMPT
}
EOF
