#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This file was derived from https://github.com/kubernetes/sample-controller/blob/4d46ec53ca337e118754c5fc50f02634b6a83380/hack/update-codegen.sh

set -o errexit
set -o nounset
set -o pipefail
set +u

: "${RESOURCE_NAME:?should be set for CRD}"
: "${OPERATOR_PKG:?should be set for operator}"

echo "Generating CRD: ${RESOURCE_NAME}, in package ${OPERATOR_PKG}..."

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

if [[ -d ${CODEGEN_PKG} ]]; then
    # generate the code with:
    # --output-base    because this script should also be able to run inside the vendor dir of
    #                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
    #                  instead of the $GOPATH directly. For normal projects this can be dropped.
    bash ${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
      ${OPERATOR_PKG}/pkg/client \
      ${OPERATOR_PKG}/pkg/apis \
      ${RESOURCE_NAME}:v1alpha1 \
      --output-base "$(dirname ${BASH_SOURCE})/../../../.." \
      --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt
fi

# To use your own boilerplate text use:
#   --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt

# This section is used by GitHub workflow to ensure that the generation step was run
if [[ "$DELTA_CHECK" == true ]]; then
  DIRTY=$(git status --porcelain)
  if [ -n "$DIRTY" ]; then
    echo "FAILED: CRD code updated without committing generated code."
    echo "Ensure make op_code_generate has run and all changes are committed."
    DIFF=$(git diff)
    echo "diff detected: $DIFF"
    DIFF=$(git diff --name-only)
    echo "files different: $DIFF"
    exit 1
  else
    echo "SUCCESS: Generated CRD code is up to date."
  fi
fi
