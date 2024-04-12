#!/usr/bin/env bash

set -ex

FLYTEKIT_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flytekit/releases/latest" | jq -r .tag_name | sed 's/^v//')
FLYTECONSOLE_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteconsole/releases/latest" | jq -r .tag_name)

# bump latest release of flyte component in kustomize
grep -rlZ "newTag:[^P]*# FLYTEADMIN_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# FLYTEADMIN_TAG/newTag: ${VERSION} # FLYTEADMIN_TAG/g" {}
grep -rlZ "newTag:[^P]*# CACHESERVICE_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# CACHESERVICE_TAG/newTag: ${VERSION} # CACHESERVICE_TAG/g" {}
grep -rlZ "newTag:[^P]*# DATACATALOG_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# DATACATALOG_TAG/newTag: ${VERSION} # DATACATALOG_TAG/g" {}
grep -rlZ "newTag:[^P]*# FLYTECONSOLE_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# FLYTECONSOLE_TAG/newTag: ${FLYTECONSOLE_TAG} # FLYTECONSOLE_TAG/g" {}
grep -rlZ "newTag:[^P]*# FLYTEPROPELLER_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# FLYTEPROPELLER_TAG/newTag: ${VERSION} # FLYTEPROPELLER_TAG/g" {}

# bump latest release of flyte component in helm
sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${VERSION} # FLYTEADMIN_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${VERSION} # FLYTEADMIN_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${VERSION} # FLYTESCHEDULER_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${VERSION} # FLYTESCHEDULER_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# CACHESERVICE_TAG,tag: ${VERSION} # CACHESERVICE_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# CACHESERVICE_TAG,tag: ${VERSION} # CACHESERVICE_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${VERSION} # DATACATALOG_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${VERSION} # DATACATALOG_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${FLYTECONSOLE_TAG} # FLYTECONSOLE_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${FLYTECONSOLE_TAG} # FLYTECONSOLE_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${VERSION} # FLYTEPROPELLER_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${VERSION} # FLYTEPROPELLER_TAG," ./charts/flyte-core/values.yaml

sed -i "s,image:[^P]*# FLYTECOPILOT_IMAGE,image: cr.flyte.org/flyteorg/flytecopilot:${VERSION} # FLYTECOPILOT_IMAGE," ./charts/flyte/values.yaml
sed -i "s,image:[^P]*# FLYTECOPILOT_IMAGE,image: cr.flyte.org/flyteorg/flytecopilot:${VERSION} # FLYTECOPILOT_IMAGE," ./charts/flyte-core/values.yaml
sed -i "s,tag:[^P]*# FLYTECOPILOT_TAG,tag: ${VERSION} # FLYTECOPILOT_TAG," ./charts/flyte-binary/values.yaml

sed -i "s,tag:[^P]*# FLYTEAGENT_TAG,tag: ${FLYTEKIT_TAG} # FLYTEAGENT_TAG," ./charts/flyteagent/values.yaml
