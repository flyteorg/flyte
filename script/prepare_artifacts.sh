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

# Download helm chart releaser
wget -q -O /tmp/chart-releaser.tar.gz https://github.com/helm/chart-releaser/releases/download/v1.2.1/chart-releaser_1.2.1_linux_amd64.tar.gz 
mkdir -p bin
tar -xf /tmp/chart-releaser.tar.gz -C bin
chmod +x bin/cr
rm /tmp/chart-releaser.tar.gz \

# Clean Git History
git stash

# Package helm chart
bin/cr package helm

# Commit Chart registry to github gh-pages
bin/cr index --owner flyteorg --git-repo flyte --charts-repo="https://flyteorg.github.io/flyte" --push --token=${FLYTE_BOT_PAT} --release-name-template="{{ .Version }}"