#!/bin/sh

set -o errexit
set -o nounset

mkdir -p /var/lib/rancher/k3s/server/manifests
if [ "${FLYTE_DEV:-}" = "True" ]; then
  cp /var/lib/rancher/k3s/server/manifests-staging/dev.yaml /var/lib/rancher/k3s/server/manifests/
else
  cp /var/lib/rancher/k3s/server/manifests-staging/complete.yaml /var/lib/rancher/k3s/server/manifests/
fi
