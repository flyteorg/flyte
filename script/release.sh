#!/usr/bin/env bash

set -ex

FLYTEKIT_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flytekit/releases/latest" | jq -r .tag_name | sed 's/^v//')
FLYTECONSOLE_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteconsole/releases/latest" | jq -r .tag_name)

# bump latest release of flyte component in helm
sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${VERSION}  # FLYTEADMIN_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${VERSION}  # FLYTEADMIN_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${VERSION}  # FLYTESCHEDULER_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${VERSION}  # FLYTESCHEDULER_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${VERSION}  # DATACATALOG_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${VERSION}  # DATACATALOG_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${FLYTECONSOLE_TAG}  # FLYTECONSOLE_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${FLYTECONSOLE_TAG}  # FLYTECONSOLE_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${VERSION}  # FLYTEPROPELLER_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${VERSION}  # FLYTEPROPELLER_TAG," ./charts/flyte-core/values.yaml

sed -i "s,image:[^P]*# FLYTECOPILOT_IMAGE,image: cr.flyte.org/flyteorg/flytecopilot:${VERSION}  # FLYTECOPILOT_IMAGE," ./charts/flyte/values.yaml
sed -i "s,image:[^P]*# FLYTECOPILOT_IMAGE,image: cr.flyte.org/flyteorg/flytecopilot:${VERSION}  # FLYTECOPILOT_IMAGE," ./charts/flyte-core/values.yaml
sed -i "s,tag:[^P]*# FLYTECOPILOT_TAG,tag: ${VERSION}  # FLYTECOPILOT_TAG," ./charts/flyte-binary/values.yaml

sed -i "s,tag:[^P]*# FLYTEAGENT_TAG,tag: ${FLYTEKIT_TAG}  # FLYTEAGENT_TAG," ./charts/flyteagent/values.yaml
