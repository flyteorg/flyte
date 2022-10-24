FROM rancher/k3s:v1.24.4-k3s1

ARG ARCH="amd64"

COPY bootstrap/output.yaml /var/lib/rancher/k3s/server/manifests/
RUN echo $ARCH
COPY images/tar/$ARCH/* /var/lib/rancher/k3s/agent/images/
COPY images/dockerfiles/k3s/bin /bin/

COPY flyte.yaml /opt/flyte/defaults.flyte.yaml

ENTRYPOINT [ "/bin/k3d-entrypoint.sh" ]
CMD [ "server", "--disable=traefik", "--disable=servicelb" ]
