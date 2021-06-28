#!/bin/bash

set -e

release_example() {
  for row in $(cat flyte_tests_manifest.json | jq -c '.[]'); do
    if [ -d "./$(echo ${row} | jq -r '.path')/_pb_output/" ]; then
      tar -cvzf "./release-snacks/flytesnacks-$(echo ${row} | jq -r '.name').tgz"  "./$(echo ${row} | jq -r '.path')/_pb_output/"
    fi
  done
}

if [ "$1" == "RELEASE" ]; then
  release_example
fi