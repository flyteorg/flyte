#!/bin/bash

# This script is used to generate SSL certificates for the development environment.
# It uses mkcert to generate the certificates, it is a simple tool for making locally-trusted development certificates.
# This script will try to install mkcert with homebrew if it's not already installed.
#
# more info:
#   - https://github.com/FiloSottile/mkcert
#   - https://nextjs.org/docs/pages/api-reference/next-cli#https-for-local-development

# Check if the operating system is macOS
if [[ "$(uname)" != "Darwin" ]]; then
    # Check if Homebrew is installed
    if ! command -v brew &> /dev/null; then
        # Install mkcert
        brew install mkcert
    fi
fi


# Check if mkcert is installed successfully
if ! command -v mkcert &> /dev/null; then
    echo "Failed to locate command mkcert. Please check your mkcert installation. Instructions can be found at https://github.com/FiloSottile/mkcert"
    exit 1
fi

echo "mkcert is installed successfully. Generating local SSL certificates..."

# Verify if ADMIN_DOMAIN environment variable is present
if [ -z "$NEXT_PUBLIC_ADMIN_DOMAIN" ]; then
    echo "ADMIN_DOMAIN environment variable is not set. Please set the variable and try again."
    exit 1
fi

# make directory ./scripts/certificate if it doesnt exist
mkdir -p ./scripts/certificate

mkcert -cert-file ./scripts/certificate/server.crt -key-file ./scripts/certificate/server.key ${NEXT_PUBLIC_ADMIN_DOMAIN} https://${NEXT_PUBLIC_ADMIN_DOMAIN}:8080 "*.${NEXT_PUBLIC_ADMIN_DOMAIN}" ${ADMIN_DOMAIN}

mkcert -install
