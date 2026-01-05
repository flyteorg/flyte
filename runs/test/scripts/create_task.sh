#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
TASK_NAME="${TASK_NAME:-my_task}"
VERSION="${VERSION:-1}"

buf curl --schema . $ENDPOINT/flyteidl2.task.TaskService/DeployTask --data @- <<EOF
{
    "task_id": {
        "org": "$ORG",
        "project": "$PROJECT",
        "domain": "$DOMAIN",
        "name": "$TASK_NAME",
        "version": "$VERSION"
    },
    "spec": {
        "task_template": {
            "type": "python"
        }
    }
}
EOF
