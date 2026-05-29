#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
RUN_NAME="${RUN_NAME:?Error: RUN_NAME environment variable is required}"
REASON="${REASON:-User requested abort}"

buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/AbortRun" --data @- <<EOF
{
    "run_id": {
        "org": "$ORG",
        "project": "$PROJECT",
        "domain": "$DOMAIN",
        "name": "$RUN_NAME"
    },
    "reason": "$REASON"
}
EOF
