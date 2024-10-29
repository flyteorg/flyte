# syntax=docker/dockerfile:1.4-labs

# Stage 1: Builder stage for preloading images
FROM --platform=${BUILDPLATFORM} mgoltzsche/podman:minimal AS builder

ARG TARGETARCH
ENV TARGETARCH="${TARGETARCH}"

WORKDIR /build


COPY images/manifest.txt images/preload ./
RUN --security=insecure ./preload manifest.txt


# Stage 2: Bootstrap stage to compile Go binary
FROM --platform=${BUILDPLATFORM} golang:1.22-bullseye AS bootstrap

ARG TARGETARCH
ENV CGO_ENABLED=0
ENV GOARCH="${TARGETARCH}"
ENV GOOS=linux

WORKDIR /flyteorg/build

# Separate dependency caching step for Go modules
COPY bootstrap/go.mod bootstrap/go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Build the bootstrap binary
COPY bootstrap/ ./
RUN go build -o dist/flyte-sandbox-bootstrap cmd/bootstrap/main.go


FROM rancher/k3s:v1.29.0-k3s1

ARG TARGETARCH
ENV TARGETARCH="${TARGETARCH}"
ARG FLYTE_SANDBOX_VERSION
ENV FLYTE_SANDBOX_VERSION="${FLYTE_SANDBOX_VERSION}"

# Copy preloaded images and manifests for k3s
COPY --from=builder /build/images/ /var/lib/rancher/k3s/agent/images/
COPY images/tar/${TARGETARCH}/ /var/lib/rancher/k3s/agent/images/
COPY manifests/ /var/lib/rancher/k3s/server/manifests-staging/
COPY bin/ /bin/


COPY --from=bootstrap /flyteorg/build/dist/flyte-sandbox-bootstrap /bin/

# Volume for Flyte storage
VOLUME /var/lib/flyte/storage


ENV SSL_CERT_DIR=/var/lib/flyte/config/ca-certificates

ENTRYPOINT [ "/bin/k3d-entrypoint.sh" ]
CMD [ "server", "--disable=traefik", "--disable=servicelb" ]
