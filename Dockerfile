FROM --platform=${BUILDPLATFORM} node:22-bookworm-slim AS console-builder

WORKDIR /oss-client
COPY oss-client/package.json oss-client/pnpm-lock.yaml ./
RUN npm install -g pnpm@10.20.0 && pnpm install --frozen-lockfile
COPY oss-client .
# Remove server-side route handlers (legacy URL redirects) incompatible with static export
RUN find src/app -name "route.ts" -delete
RUN BUILD_STATIC=1 pnpm build

FROM --platform=${BUILDPLATFORM} golang:1.24-bookworm AS flytebuilder

ARG TARGETARCH
ENV GOARCH="${TARGETARCH}"
ENV GOOS=linux

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
COPY runs runs

COPY go.mod go.sum ./
RUN go mod download
COPY manager manager
COPY --from=console-builder /oss-client/out console/dist
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    go build -tags console -v -o dist/flyte ./manager/cmd/


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
