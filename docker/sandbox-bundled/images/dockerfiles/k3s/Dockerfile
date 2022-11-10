# syntax=docker/dockerfile:1.4-labs

FROM --platform=linux/amd64 mgoltzsche/podman:minimal AS builder

ARG TARGETARCH
RUN mkdir /tmp/images
RUN --security=insecure podman pull --arch ${TARGETARCH} kubernetesui/dashboard:v2.6.1 && podman save -o /tmp/images/kubernetesui-dashboard.tar kubernetesui/dashboard:v2.6.1
RUN --security=insecure podman pull --arch ${TARGETARCH} postgres:14-alpine && podman save -o /tmp/images/postgres.tar postgres:14-alpine
RUN --security=insecure podman pull --arch ${TARGETARCH} envoyproxy/envoy:v1.23-latest && podman save -o /tmp/images/envoy.tar envoyproxy/envoy:v1.23-latest
RUN --security=insecure podman pull --arch ${TARGETARCH} rancher/pause:3.1 && podman save -o /tmp/images/pause.tar rancher/pause:3.1
RUN --security=insecure podman pull --arch ${TARGETARCH} kubernetesui/metrics-scraper:v1.0.8 && podman save -o /tmp/images/metrics-scraper.tar kubernetesui/metrics-scraper:v1.0.8
RUN --security=insecure podman pull --arch ${TARGETARCH} registry:2 && podman save -o /tmp/images/registry.tar registry:2
RUN --security=insecure podman pull --arch ${TARGETARCH} bitnami/minio:2022-debian-11 && podman save -o /tmp/images/minio.tar bitnami/minio:2022-debian-11

FROM rancher/k3s:v1.24.4-k3s1

ARG TARGETARCH
COPY bootstrap/output.yaml /var/lib/rancher/k3s/server/manifests/
COPY images/dockerfiles/tar/${TARGETARCH}/* /var/lib/rancher/k3s/agent/images/
COPY --from=builder /tmp/images/  /var/lib/rancher/k3s/agent/images/
COPY images/dockerfiles/k3s/bin /bin/

COPY flyte.yaml /opt/flyte/defaults.flyte.yaml

ENTRYPOINT [ "/bin/k3d-entrypoint.sh" ]
CMD [ "server", "--disable=traefik", "--disable=servicelb" ]

