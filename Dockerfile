ARG FLYTE_VERSION="latest"

FROM ghcr.io/flyteorg/flyteconsole-release:$FLYTE_VERSION AS flyteconsole


FROM golang:1.19.1-bullseye AS flytebuilder

ARG FLYTE_VERSION="master"

WORKDIR /flyteorg/build
RUN git clone -b add-secrets --depth=1 https://github.com/flyteorg/flyte.git ./flyte
WORKDIR /flyteorg/build/flyte
RUN go get "github.com/flyteorg/flyteadmin@move-secrets"
RUN go mod download
COPY --from=flyteconsole /app/dist cmd/single/dist
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    go build -tags console -v -o dist/flyte cmd/main.go


FROM gcr.io/distroless/base-debian11

COPY --from=flytebuilder /flyteorg/build/flyte/dist/flyte /flyte
ENTRYPOINT [ "/flyte" ]
