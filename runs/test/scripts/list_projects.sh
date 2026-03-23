#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
LIMIT="${LIMIT:-10}"
TOKEN="${TOKEN:-}"
FILTERS="${FILTERS:-}"

buf curl --schema . "$ENDPOINT/flyteidl2.project.ProjectService/ListProjects" --data @- <<EOF
{
  "limit": $LIMIT,
  "token": "$TOKEN",
  "filters": "$FILTERS",
  "org": "$ORG"
}
EOF
