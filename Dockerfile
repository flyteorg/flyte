# Todo(alex): We should add UI into the image when UI is done

FROM --platform=${BUILDPLATFORM} golang:1.26.3-bookworm AS gobase

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


# Builds the flyte single binary (the default image).
FROM gobase AS flytebuilder
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    go build -v -o dist/flyte ./manager/cmd/

# Builds the flytecopilot data init/sidecar binary.
FROM gobase AS flytecopilotbuilder
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    go build -v -o dist/flyte-copilot ./flytecopilot


# flytecopilot runtime image. Build with `--target flytecopilot`.
# flyteplugins invokes copilot as `/bin/flyte-copilot` (flytek8s.CopilotCommandArgs),
# so the binary must live at that exact path.
FROM debian:bookworm-slim AS flytecopilot

ARG FLYTE_VERSION
ENV FLYTE_VERSION="${FLYTE_VERSION}"

ENV DEBCONF_NONINTERACTIVE_SEEN=true
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install --no-install-recommends --yes \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=flytecopilotbuilder /flyteorg/build/dist/flyte-copilot /bin/flyte-copilot

CMD ["flyte-copilot"]


# flyte single binary runtime image. This is the default build target and must
# remain the last stage so `docker build .` (no --target) keeps building it.
FROM debian:bookworm-slim AS flyte

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
