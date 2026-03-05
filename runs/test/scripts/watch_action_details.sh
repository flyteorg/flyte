#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
TASK_NAME="${TASK_NAME:-my_task}"
RUN_NAME="${RUN_NAME:?Error: RUN_NAME environment variable is required}"
ACTION="${ACTION:-a0}"

buf curl --schema . $ENDPOINT/flyteidl2.workflow.RunService/WatchActionDetails --data @- <<EOF
{
    "action_id": {
        "run": {
            "org": "$ORG",
            "project": "$PROJECT",
            "domain": "$DOMAIN",
            "name": "$RUN_NAME"
        },
        "name": "$ACTION"
    }
}
EOF
