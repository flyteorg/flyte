#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
RUN_NAME="${RUN_NAME:?Error: RUN_NAME environment variable is required}"
ACTION="${ACTION:-$RUN_NAME}"

buf curl --schema . -H "Cookie: ${FLYTE_COOKIE}" "$ENDPOINT/flyteidl2.workflow.RunService/GetActionData" --data @- <<EOF
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
