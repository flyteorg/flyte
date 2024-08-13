ARG FLYTECONSOLE_VERSION=latest
FROM ghcr.io/flyteorg/flyteconsole:${FLYTECONSOLE_VERSION} AS flyteconsole


FROM --platform=${BUILDPLATFORM} golang:1.22-bookworm AS flytebuilder

ARG TARGETARCH
ENV GOARCH "${TARGETARCH}"
ENV GOOS linux

WORKDIR /flyteorg/build

COPY datacatalog datacatalog
COPY flyteadmin flyteadmin
COPY flytecopilot flytecopilot
COPY flyteidl flyteidl
COPY flyteplugins flyteplugins
COPY flytepropeller flytepropeller
COPY flytestdlib flytestdlib

COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
COPY --from=flyteconsole /app/ cmd/single/dist
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    go build -tags console -v -o dist/flyte cmd/main.go


FROM debian:bookworm-slim

ARG FLYTE_VERSION
ENV FLYTE_VERSION "${FLYTE_VERSION}"

ENV DEBCONF_NONINTERACTIVE_SEEN true
ENV DEBIAN_FRONTEND noninteractive

# Install core packages
RUN apt-get update && apt-get install --no-install-recommends --yes \
        ca-certificates \
        tini \
    && rm -rf /var/lib/apt/lists/*

# Copy compiled executable into image
COPY --from=flytebuilder /flyteorg/build/dist/flyte /usr/local/bin/

# Set entrypoint
ENTRYPOINT [ "/usr/bin/tini", "-g", "--", "/usr/local/bin/flyte" ]
