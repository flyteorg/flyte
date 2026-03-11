FROM --platform=${BUILDPLATFORM} golang:1.24-bookworm AS flytebuilder

ARG TARGETARCH
ENV GOARCH="${TARGETARCH}"
ENV GOOS=linux
ENV CGO_ENABLED=1

# Install cross-compilation toolchains for CGO (required by go-sqlite3)
RUN apt-get update && apt-get install --no-install-recommends --yes \
        gcc-aarch64-linux-gnu \
        gcc-x86-64-linux-gnu \
        libc6-dev-amd64-cross \
        libc6-dev-arm64-cross \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /flyteorg/build

COPY app app
COPY dataproxy dataproxy
COPY executor executor
COPY flytecopilot flytecopilot
COPY flyteidl2 flyteidl2
COPY flyteplugins flyteplugins
COPY flytestdlib flytestdlib
COPY gen gen
COPY actions actions
COPY events events
COPY runs runs

COPY go.mod go.sum ./
RUN go mod download
COPY manager manager
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    CC=$(case "${TARGETARCH}" in amd64) echo x86_64-linux-gnu-gcc;; arm64) echo aarch64-linux-gnu-gcc;; esac) \
    go build -v -o dist/flyte ./manager/cmd/


FROM debian:bookworm-slim

ARG FLYTE_VERSION
ENV FLYTE_VERSION="${FLYTE_VERSION}"

ENV DEBCONF_NONINTERACTIVE_SEEN=true
ENV DEBIAN_FRONTEND=noninteractive

# Install core packages
RUN apt-get update && apt-get install --no-install-recommends --yes \
        ca-certificates \
        tini \
    && rm -rf /var/lib/apt/lists/*

# Copy compiled executable into image
COPY --from=flytebuilder /flyteorg/build/dist/flyte /usr/local/bin/

# Set entrypoint
ENTRYPOINT [ "/usr/bin/tini", "-g", "--", "/usr/local/bin/flyte" ]