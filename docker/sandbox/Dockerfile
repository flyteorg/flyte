# syntax=docker/dockerfile:1.4-labs

# Stage 1: Base setup with Alpine
FROM alpine:3.20.3 AS base

# Install necessary dependencies
RUN apk add --no-cache openssl bash git make tini curl jq

# Set up directories for Flyte artifacts
RUN mkdir -p /flyteorg/bin /flyteorg/share

# Install k3s with version and architecture specification
ARG K3S_VERSION="v1.21.1+k3s1"
ARG TARGETARCH

RUN case $TARGETARCH in \
    amd64) export SUFFIX=;; \
    arm64 | aarch64) export SUFFIX=-arm64;; \
    *) echo "Unsupported architecture"; exit 1;; \
    esac; \
    wget -q -O /flyteorg/bin/k3s https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s${SUFFIX} \
    && chmod +x /flyteorg/bin/k3s

# Install Helm 
ARG HELM_VERSION="v3.6.3"
RUN wget -q -O /flyteorg/bin/get_helm.sh https://raw.githubusercontent.com/helm/helm/${HELM_VERSION}/scripts/get-helm-3 && \
    chmod 700 /flyteorg/bin/get_helm.sh && \
    sh /flyteorg/bin/get_helm.sh --version ${HELM_VERSION} && \
    mv /usr/local/bin/helm /flyteorg/bin/helm && \
    rm /flyteorg/bin/get_helm.sh

# Install flytectl
RUN wget -q -O - https://raw.githubusercontent.com/flyteorg/flytectl/master/install.sh | BINDIR=/flyteorg/bin sh -s

# Copy Flyte Helm charts and scripts
COPY charts/flyte/ /flyteorg/share/flyte
COPY charts/flyte-core/ /flyteorg/share/flyte-core
COPY charts/flyte-deps/ /flyteorg/share/flyte-deps
COPY docker/sandbox/kubectl docker/sandbox/cgroup-v2-hack.sh docker/sandbox/wait-for-flyte.sh /flyteorg/bin/


# Stage 2: Docker-in-Docker setup with dependencies
FROM docker:20.10.25-dind-alpine3.20 AS dind

# Copy artifacts from base stage
COPY --from=base /flyteorg/ /flyteorg/

# Set environment variables for Flyte version and test type
ARG FLYTE_VERSION="latest"
ENV FLYTE_VERSION="${FLYTE_VERSION}"
ARG FLYTE_TEST="release"
ENV FLYTE_TEST="${FLYTE_TEST}"

# Update PATH to include Flyte binaries
ENV PATH="/flyteorg/bin:${PATH}"

# Declare volumes for k3s to persist data across containers
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# Expose Flyte service ports
EXPOSE 30081 30082 30084 30088

# Copy entrypoint script and set it as entrypoint
COPY docker/sandbox/flyte-entrypoint-default.sh /flyteorg/bin/flyte-entrypoint.sh
ENTRYPOINT ["tini", "--", "/flyteorg/bin/flyte-entrypoint.sh"]

