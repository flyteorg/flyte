set -ex

setup_opta () {
  sudo apt-get update
  sudo apt-get install -y locales locales-all
  /bin/bash -c "$(curl -fsSL https://docs.opta.dev/install.sh)"
  sudo ln -fs ~/.opta/opta /usr/local/bin/opta
  export PATH=$PATH:/home/runner/.opta
}

setup_jq () {
  sudo apt-get install -y jq
}

commit_changelog () {
  sed -i "s,changelog:[^*]*# OPTA_CHANGELOG,changelog: $(date +%s) # OPTA_CHANGELOG," ./opta/aws/flyte.yaml
}

prepare_flyte_core () {
 helm repo add flyteorg https://flyteorg.github.io/flyte
 helm repo update
 helm fetch --untar --untardir dev/opta flyteorg/flyte-core --version $1
 echo "changelog: $(date +%s)" >> dev/opta/flyte-core/values-eks.yaml
}

prepare_helm_chart () {
    commit_changelog
    FLYTE_VERSION=$(curl --silent "https://api.github.com/repos/flyteorg/flyte/releases/latest" | jq -r .tag_name)
    prepare_flyte_core $FLYTE_VERSION
    sed -i "s,chart_version:[^P]*# FLYTE_VERSION,chart_version: $FLYTE_VERSION # FLYTE_VERSION," ./opta/aws/flyte.yaml
}

opta_action () {
  cd $2
  if [[ $1 == "force-unlock" ]]
  then
    y | OPTA_DEBUG=true HELM_DEBUG=true opta $1 -c ./opta/aws/flyte.yaml --env $ENV
    exit 0
  fi
  y | OPTA_DEBUG=true HELM_DEBUG=true opta $1 -c ./opta/aws/flyte.yaml --env $ENV --auto-approve --detailed-plan
}


prepare_flyte_release_build () {
  REPOSITORY="flyte"
  [ "${OPTA_ENVIRONMENT}" == "nightly" ] && REPOSITORY="flyteadmin" || REPOSITORY="flyte"
  if [[ -z "$FLYTEADMIN_TAG" ]]; then
    FLYTEADMIN_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/$REPOSITORY/releases/latest" | jq -r .tag_name)
  else
    [ "${FLYTEADMIN_TAG}" == "latest" ] && FLYTEADMIN_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/$REPOSITORY/releases/latest" | jq -r .tag_name) || FLYTEADMIN_TAG=$FLYTEADMIN_TAG
  fi

  [ "${OPTA_ENVIRONMENT}" == "nightly" ] && REPOSITORY="datacatalog" || REPOSITORY="flyte"
  if [[ -z "$DATACATALOG_TAG" ]]; then
    DATACATALOG_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/$REPOSITORY/releases/latest" | jq -r .tag_name)
  else
    [ "${DATACATALOG_TAG}" == "latest" ] && DATACATALOG_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/$REPOSITORY/releases/latest" | jq -r .tag_name) || DATACATALOG_TAG=$DATACATALOG_TAG
  fi

  [ "${OPTA_ENVIRONMENT}" == "nightly" ] && REPOSITORY="flyteconsole" || REPOSITORY="flyte"
  if [[ -z "$FLYTECONSOLE_TAG" ]]; then
    FLYTECONSOLE_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/$REPOSITORY/releases/latest" | jq -r .tag_name)
  else
    [ "${FLYTECONSOLE_TAG}" == "latest" ] && FLYTECONSOLE_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/$REPOSITORY/releases/latest" | jq -r .tag_name) || FLYTECONSOLE_TAG=$FLYTECONSOLE_TAG
  fi

  [ "${OPTA_ENVIRONMENT}" == "nightly" ] && REPOSITORY="flytepropeller" || REPOSITORY="flyte"
  if [[ -z "$FLYTEPROPELLER_TAG" ]]; then
    FLYTEPROPELLER_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/$REPOSITORY/releases/latest" | jq -r .tag_name)
  else
    [ "${FLYTEPROPELLER_TAG}" == "latest" ] && FLYTEPROPELLER_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/$REPOSITORY/releases/latest" | jq -r .tag_name) || FLYTEPROPELLER_TAG=$FLYTEPROPELLER_TAG
  fi
  sed -i "s,tag:[^P]*# FLYTEADMIN_TAG,tag: ${FLYTEADMIN_TAG} # FLYTEADMIN_TAG," ./opta/aws/flyte.yaml
  sed -i "s,tag:[^P]*# FLYTESCHEDULER_TAG,tag: ${FLYTEADMIN_TAG} # FLYTESCHEDULER_TAG," ./opta/aws/flyte.yaml
  sed -i "s,tag:[^P]*# DATACATALOG_TAG,tag: ${DATACATALOG_TAG} # DATACATALOG_TAG," ./opta/aws/flyte.yaml
  sed -i "s,tag:[^P]*# FLYTECONSOLE_TAG,tag: ${FLYTECONSOLE_TAG} # FLYTECONSOLE_TAG," ./opta/aws/flyte.yaml
  sed -i "s,tag:[^P]*# FLYTEPROPELLER_TAG,tag: ${FLYTEPROPELLER_TAG} # FLYTEPROPELLER_TAG," ./opta/aws/flyte.yaml
}

install_flytekit () {
  python -m pip install --upgrade pip
  pip install flytekit
  pip freeze
}

setup_flytekit () {
  install_flytekit
  echo $CLIENT_SECRET >> /home/runner/secret_location
  flyte-cli setup-config --host=$DNS
}

$1 $2 $3