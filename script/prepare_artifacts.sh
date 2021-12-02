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
sed -i "s,repository:[^P]*# FLYTEADMIN_IMAGE,repository: cr.flyte.org/flyteorg/flyteadmin-release # FLYTEADMIN_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# FLYTEADMIN_IMAGE,repository: cr.flyte.org/flyteorg/flyteadmin-release # FLYTEADMIN_IMAGE," ./charts/flyte-core/values.yaml

sed -i "s,repository:[^P]*# FLYTESCHEDULER_IMAGE,repository: cr.flyte.org/flyteorg/flytescheduler-release # FLYTESCHEDULER_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# FLYTESCHEDULER_IMAGE,repository: cr.flyte.org/flyteorg/flytescheduler-release # FLYTESCHEDULER_IMAGE," ./charts/flyte-core/values.yaml

sed -i "s,repository:[^P]*# DATACATALOG_IMAGE,repository: cr.flyte.org/flyteorg/datacatalog-release # DATACATALOG_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# DATACATALOG_IMAGE,repository: cr.flyte.org/flyteorg/datacatalog-release # DATACATALOG_IMAGE," ./charts/flyte-core/values.yaml

sed -i "s,repository:[^P]*# FLYTECONSOLE_IMAGE,repository: cr.flyte.org/flyteorg/flyteconsole-release # FLYTECONSOLE_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# FLYTECONSOLE_IMAGE,repository: cr.flyte.org/flyteorg/flyteconsole-release # FLYTECONSOLE_IMAGE," ./charts/flyte-core/values.yaml

sed -i "s,repository:[^P]*# FLYTEPROPELLER_IMAGE,repository: cr.flyte.org/flyteorg/flytepropeller-release # FLYTEPROPELLER_IMAGE," ./charts/flyte/values.yaml
sed -i "s,repository:[^P]*# FLYTEPROPELLER_IMAGE,repository: cr.flyte.org/flyteorg/flytepropeller-release # FLYTEPROPELLER_IMAGE," ./charts/flyte-core/values.yaml
