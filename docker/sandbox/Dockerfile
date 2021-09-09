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

# Make directory to store artifacts
RUN mkdir -p /flyteorg/bin /flyteorg/share

# Install k3s
ARG K3S_VERSION="v1.21.1%2Bk3s1"
RUN wget -q -O /flyteorg/bin/k3s https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s \
    && chmod +x /flyteorg/bin/k3s

# Install Helm
ENV HELM_URL="https://get.helm.sh"
ARG HELM_VERSION="v3.6.3"
RUN wget ${HELM_URL}/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xz && \
    mv linux-amd64/helm /flyteorg/bin/helm && \
    chmod +x /flyteorg/bin/helm && \
    rm -rf linux-amd64

# Install flytectl
RUN wget -q -O - https://raw.githubusercontent.com/flyteorg/flytectl/master/install.sh | BINDIR=/flyteorg/bin sh -s

# Install buildkit-cli-for-kubectl
COPY --from=go_builder_ /install/linux/ /flyteorg/bin/

# Copy flyte chart
COPY charts/flyte/ /flyteorg/share/flyte

# Copy scripts
COPY docker/sandbox/kubectl docker/sandbox/cgroup-v2-hack.sh docker/sandbox/wait-for-flyte.sh /flyteorg/bin/


FROM base_ AS default

# Install dependencies
RUN apk add --no-cache bash git make tini

# Copy entrypoints
COPY docker/sandbox/flyte-entrypoint-default.sh /flyteorg/bin/flyte-entrypoint.sh
COPY docker/sandbox/bashrc /root/.bashrc

# Update PATH variable
ENV PATH "/flyteorg/bin:${PATH}"

# Declare volumes for k3s
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# Expose Flyte ports
EXPOSE 30081 30082 30084

ENTRYPOINT ["tini", "flyte-entrypoint.sh"]


FROM docker:20.10.7-dind AS dind

# Install dependencies
RUN apk add --no-cache bash git make tini

# Copy artifacts from base
COPY --from=base_ /flyteorg/ /flyteorg/

# Copy entrypoints
COPY docker/sandbox/flyte-entrypoint-dind.sh /flyteorg/bin/flyte-entrypoint.sh

# Update PATH variable
ENV PATH "/flyteorg/bin:${PATH}"

# Declare volumes for k3s
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# Expose Flyte ports
EXPOSE 30081 30082 30084

ENTRYPOINT ["tini", "flyte-entrypoint.sh"]
