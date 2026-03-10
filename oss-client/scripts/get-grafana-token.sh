#!/bin/sh
set -e

SECRET_ID="frontend/build/grafana"

SECRET_JSON=$(aws secretsmanager get-secret-value \
  --region us-east-2 \
  --secret-id "$SECRET_ID" \
  --query SecretString \
  --output text)

echo "$SECRET_JSON" | jq -r '.GRAFANA_SOURCEMAP_TOKEN'
