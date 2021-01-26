#!/bin/bash

#Usage
#  make_all_subdirs.sh target
#
# This script is intended to be run from the base cookbook directory. It will go through all sub-folders in the
# recipes folder and run the 'target' supplied if a Makefile exists in that directory.


DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

shopt -s dotglob
find recipes/* -type d | while IFS= read -r d; do
    if [ -f "$d/Makefile" ]; then
        echo "Running make in $d..."
        cd $d;
        make $1;
        cd -;
    fi
done
