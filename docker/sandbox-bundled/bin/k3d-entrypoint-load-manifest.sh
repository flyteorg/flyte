#!/bin/sh

set -o errexit
set -o nounset

REPLACEMENTS=$(mktemp)
trap 'rm -f ${REPLACEMENTS}' EXIT

cat << EOF > ${REPLACEMENTS}
s/%{HOST_GATEWAY_IP}%/$(ip route | awk '/default/ {print $3}')/g
EOF

TEMPLATE=/var/lib/rancher/k3s/server/manifests-staging/complete.yaml
if [ "${FLYTE_DEV:-}" = "True" ]; then
  TEMPLATE=/var/lib/rancher/k3s/server/manifests-staging/dev.yaml
fi

mkdir -p /var/lib/rancher/k3s/server/manifests
sed -f ${REPLACEMENTS} ${TEMPLATE} > /var/lib/rancher/k3s/server/manifests/sandbox.yaml
