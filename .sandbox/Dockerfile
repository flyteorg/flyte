ARG DOCKER_VERSION
FROM docker:${DOCKER_VERSION}-dind

# Install dependencies
RUN apk add --no-cache bash git make tini

# Install k3s
ARG K3S_VERSION
ADD https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s /usr/local/bin/k3s
RUN chmod +x /usr/local/bin/k3s
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# Copy entrypoints
COPY k3s-entrypoint.sh /usr/local/bin

ENTRYPOINT ["tini", "k3s-entrypoint.sh"]
