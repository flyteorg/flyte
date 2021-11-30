#!/usr/bin/env bash

set -e

# Create dir structure
mkdir -p release

# Copy all deployment manifest in release directory
for file in ./deployment/**/flyte_generated.yaml; do
    if [ -f "$file" ]; then
        result=${file/#"./deployment/"}
        result=${result/%"/flyte_generated.yaml"}
        cp $file "./release/flyte_${result}_manifest.yaml"
    fi
done

grep -rlZ "version:[^P]*# VERSION" ./charts/flyte/Chart.yaml | xargs -0 sed -i "s/version:[^P]*# VERSION/version: ${VERSION} # VERSION/g"
sed "s/v0.1.10/${VERSION}/g" ./charts/flyte/README.md  > temp.txt && mv temp.txt ./charts/flyte/README.md

grep -rlZ "version:[^P]*# VERSION" ./charts/flyte-core/Chart.yaml | xargs -0 sed -i "s/version:[^P]*# VERSION/version: ${VERSION} # VERSION/g"
sed "s/v0.1.10/${VERSION}/g" ./charts/flyte-core/README.md  > temp.txt && mv temp.txt ./charts/flyte-core/README.md

helm dep update ./charts/flyte

# bump latest release of flyte component in helm
sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${VERSION} # FLYTEADMIN_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${VERSION} # FLYTEADMIN_TAG," ./charts/flyte-core/values.yaml
sed -i "s,repository:[^P]*# FLYTEADMIN_IMAGE,repository: cr.flyte.org/flyteorg/flyteadmin-release:${VERSION} # FLYTEADMIN_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# FLYTEADMIN_IMAGE,repository: cr.flyte.org/flyteorg/flyteadmin-release:${VERSION} # FLYTEADMIN_IMAGE," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${VERSION} # FLYTESCHEDULER_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${VERSION} # FLYTESCHEDULER_TAG," ./charts/flyte-core/values.yaml
sed -i "s,repository:[^P]*# FLYTESCHEDULER_IMAGE,repository: cr.flyte.org/flyteorg/flytescheduler-release:${VERSION} # FLYTESCHEDULER_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# FLYTESCHEDULER_IMAGE,repository: cr.flyte.org/flyteorg/flytescheduler-release:${VERSION} # FLYTESCHEDULER_IMAGE," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${VERSION} # DATACATALOG_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${VERSION} # DATACATALOG_TAG," ./charts/flyte-core/values.yaml
sed -i "s,repository:[^P]*# DATACATALOG_IMAGE,repository: cr.flyte.org/flyteorg/datacatalog-release:${VERSION} # DATACATALOG_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# DATACATALOG_IMAGE,repository: cr.flyte.org/flyteorg/datacatalog-release:${VERSION} # DATACATALOG_IMAGE," ./charts/flyte-core/values.yaml

sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${VERSION} # FLYTECONSOLE_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${VERSION} # FLYTECONSOLE_TAG," ./charts/flyte-core/values.yaml
sed -i "s,repository:[^P]*# FLYTECONSOLE_IMAGE,repository: cr.flyte.org/flyteorg/flyteconsole-release:${VERSION} # FLYTECONSOLE_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# FLYTECONSOLE_IMAGE,repository: cr.flyte.org/flyteorg/flyteconsole-release:${VERSION} # FLYTECONSOLE_IMAGE," ./charts/flyte-core/values.yaml


sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${VERSION} # FLYTEPROPELLER_TAG," ./charts/flyte/values.yaml
sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${VERSION} # FLYTEPROPELLER_TAG," ./charts/flyte-core/values.yaml
sed -i "s,repository:[^P]*# FLYTEPROPELLER_IMAGE,repository: cr.flyte.org/flyteorg/flytepropeller-release:${VERSION} # FLYTEPROPELLER_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# FLYTEPROPELLER_IMAGE,repository: cr.flyte.org/flyteorg/flytepropeller-release:${VERSION} # FLYTEPROPELLER_IMAGE," ./charts/flyte-core/values.yaml
