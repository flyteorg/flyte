#!/bin/sh

flyte-sandbox-bootstrap

KUBECONFIG_PATH="${K3S_KUBECONFIG_OUTPUT:-/etc/rancher/k3s/k3s.yaml}"
(
  while ! [ -s "$KUBECONFIG_PATH" ]; do sleep 1; done
  sed -i 's/: default/: flytev2-sandbox/g' "$KUBECONFIG_PATH"
) &
