#!/bin/bash

# Fetches and install Dolt. To be invoked by the Dockerfile

# echo commands to the terminal output
set -eox pipefail

# Install Dolt

apt-get update -y \
    && apt-get install curl \
    && sudo bash -c 'curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | sudo bash'
