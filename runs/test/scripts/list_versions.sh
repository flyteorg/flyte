#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
TASK_NAME="${TASK_NAME:-my_task}"

# List versions of a specific task
buf curl --schema . $ENDPOINT/flyteidl2.task.TaskService/ListVersions --data @- <<EOF
{
    "task_name": {
				"org": "$ORG",
				"project": "$PROJECT",
        "domain": "$DOMAIN",
        "name": "$TASK_NAME"
    }
}
EOF
