#!/usr/bin/env bash
set -e

echo ""
echo "waiting up to 5 minutes for kubernetes to start..."

K8S_TIMEOUT="300"

SECONDS=0
while ! systemctl is-active --quiet multi-user.target; do
  sleep 2
  if [ "$SECONDS" -gt "$K8S_TIMEOUT" ]; then
    echo "ERROR: timed out waiting for kubernetes to start."
    kubectl get all --all-namespaces
    kubectl describe nodes
    exit 1
  fi
done

echo "kubernetes started in $SECONDS seconds."
echo ""

exec $1
