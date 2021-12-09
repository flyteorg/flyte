#! /usr/bin/env sh

a=$(cat $1/a)
b=$(cat $1/b)

echo "4*a(1) * $a * $b" | bc -l | tee $2/area

echo "[from shell rawcontainer]" | tee $2/metadata
