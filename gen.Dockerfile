# Flyte CI/Development Docker Image
# This image provides a consistent environment for both CI and local development
# Multi-stage build for parallel downloads and optimized caching

# Get target architecture from buildx
ARG TARGETARCH
ARG TARGETOS

# Set versions for all tools
ARG GO_VERSION=1.24.6
ARG PYTHON_VERSION=3.12.9
ARG NODE_VERSION=20.18.3
ARG RUST_VERSION=1.84.0
ARG UV_VERSION=0.8.4
ARG BUF_VERSION=1.58.0
ARG MOCKERY_VERSION=2.53.5

# Stage 1: Download Go (parallel)
FROM golang:${GO_VERSION}-alpine AS go-source
# Just copy from official image

# Stage 2: Download Node.js (parallel)
FROM node:${NODE_VERSION}-bookworm-slim AS node-source
# Using Debian-based image for glibc compatibility with Ubuntu
RUN npm install -g pnpm

# Stage 3: Download uv and Python (parallel)
FROM ubuntu:24.04 AS python-installer
ARG TARGETARCH
ARG UV_VERSION
ARG PYTHON_VERSION
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y curl ca-certificates && rm -rf /var/lib/apt/lists/*
RUN curl -LsSf "https://astral.sh/uv/${UV_VERSION}/install.sh" | sh
ENV PATH="/root/.local/bin:${PATH}"
RUN uv python install ${PYTHON_VERSION}

# Stage 4: Download Buf (parallel)
FROM alpine:latest AS buf-downloader
ARG TARGETARCH
ARG BUF_VERSION
RUN apk add --no-cache curl tar
RUN BUFARCH=$(case ${TARGETARCH} in amd64) echo "x86_64" ;; arm64) echo "aarch64" ;; *) echo "x86_64" ;; esac) && \
    curl -fsSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-${BUFARCH}.tar.gz" | \
    tar -xzC /tmp

# Stage 5: Final image
FROM ubuntu:24.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Set versions again for reference
ARG GO_VERSION
ARG PYTHON_VERSION
ARG NODE_VERSION
ARG RUST_VERSION
ARG UV_VERSION
ARG MOCKERY_VERSION

# Install system dependencies with cache mount
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y \
    # Basic tools
    curl \
    wget \
    git \
    make \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    unzip \
    xz-utils \
    jq \
    # Minimal Python build deps (most come from official Python)
    libssl-dev \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy Go from official image
COPY --from=go-source /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"
ENV GOPATH="/root/go"

# Copy uv and Python from installer stage
COPY --from=python-installer /root/.local /root/.local
COPY --from=python-installer /root/.local/share/uv /root/.local/share/uv
ENV PATH="/root/.local/bin:${PATH}"
RUN UV_PYTHON=$(uv python find ${PYTHON_VERSION}) && \
    ln -sf ${UV_PYTHON} /usr/local/bin/python3 && \
    ln -sf ${UV_PYTHON} /usr/local/bin/python

# Copy Node.js from official image (using Debian-based for glibc compatibility)
# Copy the entire /usr/local structure to preserve npm/pnpm directory layout
COPY --from=node-source /usr/local /usr/local

# Install Rust (still need to run installer for architecture detection)
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain ${RUST_VERSION} --profile minimal
ENV PATH="/root/.cargo/bin:${PATH}"

# Copy Buf from downloader stage
COPY --from=buf-downloader /tmp/buf/bin/buf /usr/local/bin/buf
COPY --from=buf-downloader /tmp/buf/bin/protoc-gen-buf-breaking /usr/local/bin/protoc-gen-buf-breaking
COPY --from=buf-downloader /tmp/buf/bin/protoc-gen-buf-lint /usr/local/bin/protoc-gen-buf-lint

# Install Go tools with cache mount
RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install "github.com/vektra/mockery/v2@v${MOCKERY_VERSION}"

# Set working directory
WORKDIR /workspace

# Verify installations & print versions
RUN echo "=== Tool Versions ===" && \
    echo "Go: $(go version)" && \
    echo "Python: $(python --version)" && \
    echo "uv: $(uv --version)" && \
    echo "Node: $(node --version)" && \
    echo "npm: $(npm --version)" && \
    echo "pnpm: $(pnpm --version)" && \
    echo "Rust: $(rustc --version)" && \
    echo "Cargo: $(cargo --version)" && \
    echo "Buf: $(buf --version)" && \
    echo "Mockery: $(mockery --version)"

# Set default command
CMD ["/bin/bash"]
