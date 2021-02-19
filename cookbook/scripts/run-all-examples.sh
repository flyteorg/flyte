#!/bin/bash

#set -e

for script in `ls ./*/*.py | grep -v __init__`; do
  echo "\033[33mExecuting ${script} ...\033[0m"
  echo "Executing ${script}..." >> /tmp/logs
  python ${script} 1> /tmp/logs 2>&1
  out=$?
  if [[ "$out" == "0" ]]; then
    echo "\033[32mSuccess"
  else
    echo "\033[31mFailed"
  fi
done
