#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT_ID="${PROJECT_ID:-testproject}"
PROJECT_NAME="${PROJECT_NAME:-Test Project}"
DESCRIPTION="${DESCRIPTION:-Project created from script}"
STATE="${STATE:-PROJECT_STATE_ACTIVE}"

buf curl --schema . "$ENDPOINT/flyteidl2.project.ProjectService/CreateProject" --data @- <<EOF
{
  "project": {
    "id": "$PROJECT_ID",
    "name": "$PROJECT_NAME",
    "description": "$DESCRIPTION",
    "labels": {
      "values": {
        "team": "runs"
      }
    },
    "state": "$STATE",
    "org": "$ORG"
  }
}
EOF
