# Todo(alex): We should add UI into the image when UI is done

FROM --platform=${BUILDPLATFORM} golang:1.25-bookworm AS flytebuilder

ARG TARGETARCH
ENV GOARCH="${TARGETARCH}"
ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /flyteorg/build

COPY dataproxy dataproxy
COPY executor executor
COPY flytecopilot flytecopilot
COPY flyteidl2 flyteidl2
COPY flyteplugins flyteplugins
COPY flytestdlib flytestdlib
COPY gen/go gen/go
COPY actions actions
COPY app app
COPY events events
COPY runs runs
COPY cache_service cache_service
COPY secret secret

COPY go.mod go.sum ./
RUN go mod download
COPY manager manager
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
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
