FROM rancher/k3s:v1.24.4-k3s1

ARG ARCH="linux/amd64"

COPY bootstrap/output.yaml /var/lib/rancher/k3s/server/manifests/
COPY images/tar/${TARGETARCH}${TARGETVARIANT}/* /var/lib/rancher/k3s/agent/imagestmp/
RUN mv /var/lib/rancher/k3s/agent/imagestmp/${TARGETARCH}${TARGETVARIANT} /var/lib/rancher/k3s/agent/images
COPY images/dockerfiles/k3s/bin /bin/

COPY flyte.yaml /opt/flyte/defaults.flyte.yaml

ENTRYPOINT [ "/bin/k3d-entrypoint.sh" ]
CMD [ "server", "--disable=traefik", "--disable=servicelb" ]
