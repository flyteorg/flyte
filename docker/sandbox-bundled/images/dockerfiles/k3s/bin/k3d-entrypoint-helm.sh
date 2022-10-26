#!/bin/sh

set -o errexit
set -o nounset

# Flyte will run from /etc/flyte/flyte.yaml as configured in Helm deployment.
# /etc/flyte/ will be mounted in from host path /srv/flyte/
# Image should be prebuilt with a default flyte.yaml file stored in /opt/flyte/defaults.flyte.yaml
# If /srv/flyte/flyte.yaml is not there (as in, not mounted in from the user's state dir) then create it from the default.
mkdir -p /srv/flyte

if [[ ! -e /srv/flyte/flyte.yaml ]]; then
  echo "Creating flyte.yaml file from default file"
  cp /opt/flyte/defaults.flyte.yaml /srv/flyte/flyte.yaml
fi
