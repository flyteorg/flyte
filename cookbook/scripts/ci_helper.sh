#!/bin/bash

set -e

for row in $(cat flyte_tests_manifest.json | jq -c '.[]'); do
  if [ -d "./$(echo ${row} | jq -r '.path')/_pb_output/" ]; then
    tar -cvzf "./release-snacks/flytesnacks-$(echo ${row} | jq -r '.name').tgz"  "./$(echo ${row} | jq -r '.path')/_pb_output/"
  fi
done
