# syntax=docker/dockerfile:1.4-labs
FROM --platform=${BUILDPLATFORM} mgoltzsche/podman:minimal AS builder

ARG TARGETARCH
ENV TARGETARCH "${TARGETARCH}"

WORKDIR /build

COPY images/manifest.txt images/preload ./
RUN --security=insecure ./preload manifest.txt


FROM --platform=${BUILDPLATFORM} golang:1.22-bullseye AS bootstrap

ARG TARGETARCH
ENV CGO_ENABLED 0
ENV GOARCH "${TARGETARCH}"
ENV GOOS linux

WORKDIR /flyteorg/build
COPY bootstrap/go.mod bootstrap/go.sum ./
RUN go mod download
COPY bootstrap/ ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    go build -o dist/flyte-sandbox-bootstrap cmd/bootstrap/main.go


FROM rancher/k3s:v1.29.0-k3s1

ARG TARGETARCH

ARG FLYTE_SANDBOX_VERSION
ENV FLYTE_SANDBOX_VERSION "${FLYTE_SANDBOX_VERSION}"

COPY --from=builder /build/images/ /var/lib/rancher/k3s/agent/images/
COPY images/tar/${TARGETARCH}/ /var/lib/rancher/k3s/agent/images/
COPY manifests/ /var/lib/rancher/k3s/server/manifests-staging/
COPY bin/ /bin/

# Install bootstrap
COPY --from=bootstrap /flyteorg/build/dist/flyte-sandbox-bootstrap /bin/

VOLUME /var/lib/flyte/storage

# Set environment variable for picking up additional CA certificates
ENV SSL_CERT_DIR /var/lib/flyte/config/ca-certificates

ENTRYPOINT [ "/bin/k3d-entrypoint.sh" ]
CMD [ "server", "--disable=traefik", "--disable=servicelb" ]
