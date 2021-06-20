#!/usr/bin/env bash

set -ex

# Download helm chart releaser
wget -q -O /tmp/chart-releaser.tar.gz https://github.com/helm/chart-releaser/releases/download/v1.2.1/chart-releaser_1.2.1_darwin_amd64.tar.gz 
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