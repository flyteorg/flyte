#! /usr/bin/env sh

echo "4*a(1) * $1 * $2" | bc -l | tee $3/area

echo "[from shell rawcontainer]" | tee $3/metadata
