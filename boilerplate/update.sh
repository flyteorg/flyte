#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

OUT="$(mktemp -d)"
trap "rm -fr $OUT" EXIT

git clone git@github.com:flyteorg/boilerplate.git "${OUT}"

echo "Updating the update.sh script."
cp "${OUT}/boilerplate/update.sh" "${DIR}/update.sh"
echo ""


CONFIG_FILE="${DIR}/update.cfg"
README="https://github.com/flyteorg/boilerplate/blob/master/Readme.rst"

if [ ! -f "$CONFIG_FILE" ]; then
  echo "$CONFIG_FILE not found."
  echo "This file is required in order to select which features to include." 
  echo "See $README for more details."
  exit 1
fi

if [ -z "$REPOSITORY" ]; then
  echo '$REPOSITORY is required to run this script'
  echo "See $README for more details."
  exit 1
fi

while read directory junk; do
  # Skip comment lines (which can have leading whitespace)
  if [[ "$directory" == '#'* ]]; then
    continue
  fi
  # Skip blank or whitespace-only lines
  if [[ "$directory" == "" ]]; then
    continue
  fi
  # Lines like
  #    valid/path  other_junk
  # are not acceptable, unless `other_junk` is a comment
  if [[ "$junk" != "" ]] && [[ "$junk" != '#'* ]]; then
    echo "Invalid config! Only one directory is allowed per line. Found '$junk'"
    exit 1
  fi

  dir_path="${OUT}/boilerplate/${directory}"
  # Make sure the directory exists
  if ! [[ -d "$dir_path" ]]; then
    echo "Invalid boilerplate directory: '$directory'"
    exit 1
  fi

  echo "***********************************************************************************"
  echo "$directory is configured in update.cfg."
  echo "-----------------------------------------------------------------------------------"
  echo "syncing files from source."
  rm -rf "${DIR}/${directory}"
  mkdir -p $(dirname "${DIR}/${directory}")
  cp -r "$dir_path" "${DIR}/${directory}"
  if [ -f "${DIR}/${directory}/update.sh" ]; then
    echo "executing ${DIR}/${directory}/update.sh"
    "${DIR}/${directory}/update.sh"
  fi
  echo "***********************************************************************************"
  echo ""
done < "$CONFIG_FILE"