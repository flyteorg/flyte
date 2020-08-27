#!/usr/bin/env bash

set -ex

df -H
kubectl get all --all-namespaces
kubectl describe nodes

function get_flyte_pods() {
    echo $(kubectl -n flyte get pods | awk '{print $1}' | grep -v NAME)
}

pods=$(get_flyte_pods)
echo $pods | while read -a podarray; do
  for i in "${podarray[@]}"; do
      echo "Logs for ${i}"
      echo "------------------------------------------"

      kubectl -n flyte describe pod $i || true
      if [[ $i == *"flyteadmin"* ]]; then
        kubectl -n flyte logs $i -c flyteadmin || true
      else
        kubectl -n flyte logs $i || true
      fi
  done
done

