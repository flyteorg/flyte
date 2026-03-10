#!/bin/sh
set -e

if ! command -v pnpm; then 
    npm install -g pnpm
fi

pnpm install
pnpm test:unit
