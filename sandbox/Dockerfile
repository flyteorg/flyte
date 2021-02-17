ARG DOCKER_VERSION

FROM docker:20.10.3-dind
LABEL org.opencontainers.image.source https://github.com/flyteorg/flyte

# Install dependencies
RUN apk add --no-cache bash git make rsync tini

# Install k3s
ARG K3S_VERSION
ADD https://github.com/k3s-io/k3s/releases/download/v1.20.2%2Bk3s1/k3s /usr/local/bin/k3s
RUN chmod +x /usr/local/bin/k3s
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# Copy entrypoints
COPY k3s-entrypoint.sh with-src-snapshot.sh /usr/local/bin

ENTRYPOINT ["tini"]
