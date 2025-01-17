#!/bin/bash

#Usage
#  run_all_plugins.sh $@
#
# This script is intended to be run from the plugins/ top-level directory. It will go through all the sub-folders
# of plugins and run command supplied if a setup.py exists in that directory (implying a plugin).


shopt -s dotglob
find ./* -prune -type d | while IFS= read -r d; do
    if [ -f "$d/setup.py" ]; then
        echo "Running your command in $d..."
        cd "$d" || exit;
        "$@"
        cd - || exit;
    fi
done
