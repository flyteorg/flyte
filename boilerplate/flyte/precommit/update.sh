#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'FLYTEORG/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/flyteorg/boilerplate/blob/master/Readme.rst

set -e

# Helper script for Automatically add DCO signoff with commit hooks 
# Taken from Envoy https://gitlab.cncf.ci/envoyproxy/envoy
if [ ! "$PWD" == "$(git rev-parse --show-toplevel)" ]; then
    cat >&2 <<__EOF__
ERROR: this script must be run at the root of the envoy source tree
__EOF__
    exit 1
fi

# Helper functions that calculate `abspath` and `relpath`. Taken from Mesos
# commit 82b040a60561cf94dec3197ea88ae15e57bcaa97, which also carries the Apache
# V2 license, and has deployed this code successfully for some time.
abspath() {
    cd "$(dirname "${1}")"
    echo "${PWD}"/"$(basename "${1}")"
    cd "${OLDPWD}"
}
relpath() {
  local FROM TO UP
  FROM="$(abspath "${1%/}")" TO="$(abspath "${2%/}"/)"
  while test "${TO}"  = "${TO#"${FROM}"/}" \
          -a "${TO}" != "${FROM}"; do
    FROM="${FROM%/*}" UP="../${UP}"
  done
  TO="${UP%/}${TO#${FROM}}"
  echo "${TO:-.}"
}

# Try to find the `.git` directory, even if it's not in Flyte project root (as
# it wouldn't be if, say, this were in a submodule). The "blessed" but fairly
# new way to do this is to use `--git-common-dir`.
DOT_GIT_DIR=$(git rev-parse --git-common-dir)
if test ! -d "${DOT_GIT_DIR}"; then
  # If `--git-common-dir` is not available, fall back to older way of doing it.
  DOT_GIT_DIR=$(git rev-parse --git-dir)
fi

mkdir -p ${DOT_GIT_DIR}/hooks

HOOKS_DIR="${DOT_GIT_DIR}/hooks"
HOOKS_DIR_RELPATH=$(relpath "${HOOKS_DIR}" "${PWD}")

if [ ! -e "${HOOKS_DIR}/prepare-commit-msg" ]; then
  echo "Installing hook 'prepare-commit-msg'"
  ln -s "${HOOKS_DIR_RELPATH}/boilerplate/flyte/precommit/hooks/prepare-commit-msg" "${HOOKS_DIR}/prepare-commit-msg"
fi

if [ ! -e "${HOOKS_DIR}/pre-push" ]; then
  echo "Installing hook 'pre-push'"
  ln -s "${HOOKS_DIR_RELPATH}/boilerplate/flyte/precommit/hooks/pre-push" "${HOOKS_DIR}/pre-push"
fi
