#!/usr/bin/env bash

set -ex

mkdir -p release
cp deployment/eks/flyte_generated.yaml ./release/flyte_eks_menifest.yaml
cp deployment/gcp/flyte_generated.yaml ./release/flyte_gcp_menifest.yaml
cp deployment/sandbox/flyte_generated.yaml ./release/flyte_sandbox_menifest.yaml
cp deployment/test/flyte_generated.yaml ./release/flyte_test_menifest.yaml