#!/usr/bin/env bash
set -e

df -H
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# stop any existing test container that might be running
docker kill dockernetes || true

# initialize the k8s container, mounting the flyte repo to /flyte.

# The container must start with systemd (/sbin/init) as PID 1

docker run \
  --detach \
  --rm \
  --privileged \
  --volume /var/lib/docker \
  --volume /lib/modules:/lib/modules \
  --volume $DIR/..:/flyte \
  --name dockernetes \
  --env "DOCKERNETES_DEBUG=${DOCKERNETES_DEBUG}" \
  lyft/dockernetes:a65a0237bfaf9055b7503473515c8dd5c99df07a /sbin/init

# wait for the system to initalize, then run execute.sh
if [ -n "$DOCKERNETES_DEBUG" ]; then
  docker exec -it \
    dockernetes /flyte/end2end/dockernetes_run.sh /flyte/end2end/execute.sh
else
  docker exec \
    dockernetes /flyte/end2end/dockernetes_run.sh /flyte/end2end/execute.sh
fi
