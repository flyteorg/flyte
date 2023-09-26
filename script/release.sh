#!/usr/bin/env bash

set -ex

# Get latest release tag (minus the component prefix)
# TODO(monorepo): This only works if we have at least one component tag per release.
#                 In other words, if we have two consecutive releases the latest tag in the second release is going to point to an invalid
#                 tag (because there will not be images tagged with the previous release tag).
LATEST_TAG=$(git tag | sed 's#[^/]*/##' | sort | tail -n 1)
FLYTEKIT_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flytekit/releases/latest" | jq -r .tag_name)
FLYTECONSOLE_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteconsole/releases/latest" | jq -r .tag_name)

# bump latest release of flyte component in kustomize
grep -rlZ "newTag:[^P]*# FLYTEADMIN_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# FLYTEADMIN_TAG/newTag: ${LATEST_TAG} # FLYTEADMIN_TAG/g" {}
grep -rlZ "newTag:[^P]*# DATACATALOG_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# DATACATALOG_TAG/newTag: ${LATEST_TAG} # DATACATALOG_TAG/g" {}
grep -rlZ "newTag:[^P]*# FLYTECONSOLE_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# FLYTECONSOLE_TAG/newTag: ${FLYTECONSOLE_TAG} # FLYTECONSOLE_TAG/g" {}
grep -rlZ "newTag:[^P]*# FLYTEPROPELLER_TAG" ./kustomize/overlays | xargs -I {} sed -i "s/newTag:[^P]*# FLYTEPROPELLER_TAG/newTag: ${LATEST_TAG} # FLYTEPROPELLER_TAG/g" {}

# bump latest release of flyte component in helm
sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${LATEST_TAG} # FLYTEADMIN_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${LATEST_TAG} # FLYTEADMIN_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${LATEST_TAG} # FLYTESCHEDULER_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${LATEST_TAG} # FLYTESCHEDULER_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${LATEST_TAG} # DATACATALOG_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${LATEST_TAG} # DATACATALOG_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${FLYTECONSOLE_TAG} # FLYTECONSOLE_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${FLYTECONSOLE_TAG} # FLYTECONSOLE_TAG," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${LATEST_TAG} # FLYTEPROPELLER_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${LATEST_TAG} # FLYTEPROPELLER_TAG," ./charts/flyte-core/values.yaml

sed -i "s,image:[^P]*# FLYTECOPILOT_IMAGE,image: cr.flyte.org/flyteorg/flytecopilot:${LATEST_TAG} # FLYTECOPILOT_IMAGE," ./charts/flyte/values.yaml
sed -i "s,image:[^P]*# FLYTECOPILOT_IMAGE,image: cr.flyte.org/flyteorg/flytecopilot:${LATEST_TAG} # FLYTECOPILOT_IMAGE," ./charts/flyte-core/values.yaml
sed -i "s,tag:[^P]*# FLYTECOPILOT_TAG,tag: ${LATEST_TAG} # FLYTECOPILOT_TAG," ./charts/flyte-binary/values.yaml

sed -i "s,tag:[^P]*# FLYTEAGENT_TAG,tag: ${FLYTEKIT_TAG} # FLYTEAGENT_TAG," ./charts/flyteagent/values.yaml

go get github.com/flyteorg/flyte/flyteadmin@${LATEST_TAG}
go get github.com/flyteorg/flyte/flytepropeller@${LATEST_TAG}
go get github.com/flyteorg/flyte/datacatalog@${LATEST_TAG}
go get github.com/flyteorg/flyte/flytestdlib@${LATEST_TAG}
go mod tidy
