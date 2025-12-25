#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
NAME="${NAME:-my_task}"

# List tasks by project, match case-insensitive to name, sort by name ascending
buf curl --schema . $ENDPOINT/flyteidl2.task.TaskService/ListTasks --data @- <<EOF
{
    "request": {
        "limit": 10,
        "token": "",
        "filters": [
            {
                "field": "name",
                "function": 14,
                "values": ["$NAME"]
            }
        ],
        "sort_by_fields": [
            {
                "key": "name",
                "direction": 1
            }
        ]
    },
    "project_id": {
        "organization": "$ORG",
        "domain": "$DOMAIN",
        "name": "$PROJECT"
    }
}
EOF
