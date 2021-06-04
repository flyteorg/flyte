#!/usr/bin/env bash

set -ex

FLYTEADMIN_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteadmin/releases/latest" | jq -r .tag_name)
DATACATALOG_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/datacatalog/releases/latest" | jq -r .tag_name)
FLYTECONSOLE_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteconsole/releases/latest" | jq -r .tag_name)
FLYTEPROPELLER_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flytepropeller/releases/latest" | jq -r .tag_name)

# bump latest release of flyte component in kustomize
grep -rlZ "newTag:[^P]*# FLYTEADMIN_TAG" ./kustomize/overlays | xargs -0 sed -i "s/newTag:[^P]*# FLYTEADMIN_TAG/newTag: ${FLYTEADMIN_TAG} # FLYTEADMIN_TAG/g"
grep -rlZ "newTag:[^P]*# DATACATALOG_TAG" ./kustomize/overlays | xargs -0 sed -i "s/newTag:[^P]*# DATACATALOG_TAG/newTag: ${DATACATALOG_TAG} # DATACATALOG_TAG/g"
grep -rlZ "newTag:[^P]*# FLYTECONSOLE_TAG" ./kustomize/overlays | xargs -0 sed -i "s/newTag:[^P]*# FLYTECONSOLE_TAG/newTag: ${FLYTECONSOLE_TAG} # FLYTECONSOLE_TAG/g"
grep -rlZ "newTag:[^P]*# FLYTEPROPELLER_TAG" ./kustomize/overlays | xargs -0 sed -i "s/newTag:[^P]*# FLYTEPROPELLER_TAG/newTag: ${FLYTEPROPELLER_TAG} # FLYTEPROPELLER_TAG/g"

# Backup template
cp ./helm/.values-template.yaml ./helm/.values-template.yaml.bak

# bump latest release of flyte component in helm
sed -i "s,FLYTEADMIN_TAG,${FLYTEADMIN_TAG}," ./helm/.values-template.yaml
sed -i "s,DATACATALOG_TAG,${DATACATALOG_TAG}," ./helm/.values-template.yaml
sed -i "s,FLYTECONSOLE_TAG,${FLYTECONSOLE_TAG}," ./helm/.values-template.yaml
sed -i "s,FLYTEPROPELLER_TAG,${FLYTEPROPELLER_TAG}," ./helm/.values-template.yaml

# Copy updated value from helm-template to values.yaml
cp ./helm/.values-template.yaml ./helm/values.yaml

# Restore template from backup and remove backup file
cp ./helm/.values-template.yaml.bak ./helm/.values-template.yaml
rm ./helm/.values-template.yaml.bak 

# Added comment on top 
sed -i '1s/^/# Auto generated file/' helm/values.yaml