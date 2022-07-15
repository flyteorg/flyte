# syntax=docker/dockerfile:1.3


ARG FLYTE_VERSION="latest"
FROM ghcr.io/flyteorg/flyteconsole-release:${FLYTE_VERSION} AS flyteconsole

FROM golang:1.18.0-alpine3.15 AS go_builder

# Install dependencies
RUN apk add --no-cache build-base

COPY go.mod go.sum /app/flyte/
WORKDIR /app/flyte
RUN go mod download

COPY --from=flyteconsole /app/dist /app/flyte/cmd/single/dist

COPY cmd/ /app/flyte/cmd/
RUN --mount=type=cache,target=/root/.cache/go-build go build -tags console -v -o /flyte cmd/main.go

FROM alpine:3.15 AS base

# Install dependencies
RUN apk add --no-cache openssl

# Make directory to store artifacts
RUN mkdir -p /flyteorg/bin /flyteorg/share

# Install k3s
ARG K3S_VERSION="v1.21.1%2Bk3s1"
ARG TARGETARCH

RUN case $TARGETARCH in \
    amd64) export SUFFIX=;; \
    arm64) export SUFFIX=-arm64;; \
    aarch64)  export SUFFIX=-arm64;; \
    # TODO: Check if we need to add case fail
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

# Install buildkit-cli-for-kubectl
COPY --from=go_builder /flyte /flyteorg/bin/

# Copy flyte chart
COPY charts/flyte-deps/ /flyteorg/share/flyte-deps

# Copy scripts
COPY docker/sandbox/kubectl docker/sandbox/cgroup-v2-hack.sh /flyteorg/bin/

# Copy Flyte config
COPY flyte.yaml /flyteorg/share/flyte.yaml

FROM docker:20.10.14-dind-alpine3.15 AS dind

# Install dependencies
RUN apk add --no-cache bash git make tini curl jq

# Copy artifacts from base
COPY --from=base /flyteorg/ /flyteorg/

# Copy entrypoints
COPY docker/sandbox-lite/flyte-entrypoint-dind.sh /flyteorg/bin/flyte-entrypoint.sh

# Copy cluster resource templates
COPY docker/sandbox-lite/templates/ /etc/flyte/clusterresource/templates/

ENV FLYTE_VERSION "${FLYTE_VERSION}"

ARG FLYTE_TEST="release"
ENV FLYTE_TEST "${FLYTE_TEST}"

RUN addgroup -S docker

# Update PATH variable
ENV PATH "/flyteorg/bin:${PATH}"
ENV POD_NAMESPACE "flyte"

# Declare volumes for k3s
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# Expose Flyte ports
# 30080 for console, 30081 for gRPC, 30082 for k8s dashboard, 30084 for minio api, 30088 for minio console
EXPOSE 30080 30081 30082 30084 30088 30089

ENTRYPOINT ["tini", "flyte-entrypoint.sh"]
