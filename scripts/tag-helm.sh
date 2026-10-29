#!/usr/bin/env bash

set -e

grep -rlZ "version:[^P]*# VERSION" ./charts/flyte-binary/Chart.yaml | xargs -0 sed -i "s/version:[^P]*# VERSION/version: ${VERSION} # VERSION/g"
grep -rlZ "version:[^P]*# VERSION" ./charts/flyte-core/Chart.yaml | xargs -0 sed -i "s/version:[^P]*# VERSION/version: ${VERSION} # VERSION/g"
sed "s/v0.2.0/${VERSION}/g" ./charts/flyte-binary/README.md > temp.txt && mv temp.txt ./charts/flyte-binary/README.md

# bump latest release of flyte component in helm
sed -i "s,tag:[^P]*# FLYTE_TAG,tag: ${VERSION} # FLYTE_TAG," ./charts/flyte-binary/values.yaml
sed -i "s,tag:[^P]*# FLYTE_TAG,tag: ${VERSION} # FLYTE_TAG," ./charts/flyte-core/values.yaml
