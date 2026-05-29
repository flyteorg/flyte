#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
SCOPE="${SCOPE:-project}" # project|org

if [ "$SCOPE" = "org" ]; then
buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/WatchRuns" --data @- <<EOF
{
    "org": "$ORG"
}
EOF
else
buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/WatchRuns" --data @- <<EOF
{
    "project_id": {
        "organization": "$ORG",
        "name": "$PROJECT",
        "domain": "$DOMAIN"
    }
}
EOF
fi
