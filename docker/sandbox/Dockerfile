FROM golang:1.16.5-alpine3.13 AS go_builder_

# Install dependencies
RUN apk add --no-cache build-base git make

# Create directory to store built artifacts
ARG INSTALL_DIR="/install"
RUN mkdir -p ${INSTALL_DIR}

ARG BUILDKIT_CLI_FOR_KUBECTL_VERSION="v0.1.2"
RUN git clone -b ${BUILDKIT_CLI_FOR_KUBECTL_VERSION} --single-branch --depth 1 https://github.com/vmware-tanzu/buildkit-cli-for-kubectl.git ${GOPATH}/src/github.com/vmware-tanzu/buildkit-cli-for-kubectl \
    && make -C ${GOPATH}/src/github.com/vmware-tanzu/buildkit-cli-for-kubectl ${INSTALL_DIR}/linux/kubectl-build BIN_DIR=${INSTALL_DIR} VERSION=${BUILDKIT_CLI_FOR_KUBECTL_VERSION}


FROM alpine:3.13.5 AS base_

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
COPY --from=go_builder_ /install/linux/ /flyteorg/bin/

# Copy flyte chart
COPY charts/flyte/ /flyteorg/share/flyte
COPY charts/flyte-core/ /flyteorg/share/flyte-core

# Copy scripts
COPY docker/sandbox/kubectl docker/sandbox/cgroup-v2-hack.sh docker/sandbox/wait-for-flyte.sh /flyteorg/bin/


FROM base_ AS default

# Install dependencies
RUN apk add --no-cache bash git make tini curl jq

# Copy entrypoints
COPY docker/sandbox/flyte-entrypoint-default.sh /flyteorg/bin/flyte-entrypoint.sh
COPY docker/sandbox/bashrc /root/.bashrc

ARG FLYTE_VERSION="latest"
ENV FLYTE_VERSION "${FLYTE_VERSION}"

ARG FLYTE_TEST="release"
ENV FLYTE_TEST "${FLYTE_TEST}"

# Update PATH variable
ENV PATH "/flyteorg/bin:${PATH}"

# Declare volumes for k3s
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# Expose Flyte ports
EXPOSE 30081 30082 30084 30087

ENTRYPOINT ["tini", "flyte-entrypoint.sh"]


FROM docker:20.10.11-dind AS dind

# Install dependencies
RUN apk add --no-cache bash git make tini curl jq

# Copy artifacts from base
COPY --from=base_ /flyteorg/ /flyteorg/

# Copy entrypoints
COPY docker/sandbox/flyte-entrypoint-dind.sh /flyteorg/bin/flyte-entrypoint.sh

ARG FLYTE_VERSION="latest"
ENV FLYTE_VERSION "${FLYTE_VERSION}"

ARG FLYTE_TEST="release"
ENV FLYTE_TEST "${FLYTE_TEST}"

# Update PATH variable
ENV PATH "/flyteorg/bin:${PATH}"

# Declare volumes for k3s
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# Expose Flyte ports
EXPOSE 30081 30082 30084 30087

ENTRYPOINT ["tini", "flyte-entrypoint.sh"]
