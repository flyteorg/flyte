#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT_ID="${PROJECT_ID:-testproject}"

buf curl --schema . "$ENDPOINT/flyteidl2.project.ProjectService/GetProject" --data @- <<EOF
{
  "id": "$PROJECT_ID",
  "org": "$ORG"
}
EOF
