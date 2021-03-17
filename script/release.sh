#!/usr/bin/env bash

set -ex

FLYTEADMIN_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteadmin/releases/latest" | jq -r .tag_name)
DATACATALOG_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/datacatalog/releases/latest" | jq -r .tag_name)
FLYTECONSOLE_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteconsole/releases/latest" | jq -r .tag_name)
FLYTEPROPELLER_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flytepropeller/releases/latest" | jq -r .tag_name)

grep -rlZ FLYTEADMIN_TAG ./kustomize/overlays | xargs -0 sed -i "s/FLYTEADMIN_TAG/${FLYTEADMIN_TAG}/g"
grep -rlZ DATACATALOG_TAG ./kustomize/overlays | xargs -0 sed -i "s/DATACATALOG_TAG/${DATACATALOG_TAG}/g"
grep -rlZ FLYTECONSOLE_TAG ./kustomize/overlays | xargs -0 sed -i "s/FLYTECONSOLE_TAG/${FLYTECONSOLE_TAG}/g"
grep -rlZ FLYTEPROPELLER_TAG ./kustomize/overlays | xargs -0 sed -i "s/FLYTEPROPELLER_TAG/${FLYTEPROPELLER_TAG}/g"