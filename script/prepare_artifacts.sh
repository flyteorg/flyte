#!/usr/bin/env bash

set -e

# Create dir structure
mkdir -p release .cr-index .cr-release-packages

# Copy all deployment manifest in release directory
for file in ./deployment/**/flyte_generated.yaml; do 
    if [ -f "$file" ]; then
        result=${file/#"./deployment/"}
        result=${result/%"/flyte_generated.yaml"}
        cp $file "./release/flyte_${result}_manifest.yaml"
    fi
done

grep -rlZ "version:[^P]*# VERSION" ./charts/flyte/Chart.yaml | xargs -0 sed -i "s/version:[^P]*# VERSION/version: ${VERSION} # VERSION/g"
grep -rlZ "version:[^P]*# VERSION" ./charts/flyte-core/Chart.yaml | xargs -0 sed -i "s/version:[^P]*# VERSION/version: ${VERSION} # VERSION/g"

# Download helm chart releaser
wget -q -O /tmp/chart-releaser.tar.gz https://github.com/helm/chart-releaser/releases/download/v1.2.1/chart-releaser_1.2.1_linux_amd64.tar.gz 
mkdir -p bin
tar -xf /tmp/chart-releaser.tar.gz -C bin
chmod +x bin/cr
rm /tmp/chart-releaser.tar.gz 

# Package helm chart
bin/cr package charts/flyte
bin/cr package charts/flyte-core

# Clean git history
git stash