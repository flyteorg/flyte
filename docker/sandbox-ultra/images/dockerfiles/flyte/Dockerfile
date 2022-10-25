FROM ghcr.io/flyteorg/flyteconsole-release:latest AS flyteconsole


FROM golang:1.19.1-bullseye AS flytebuilder

WORKDIR /flyteorg/build
RUN git clone --depth=1 https://github.com/flyteorg/flyte.git .
RUN go mod download
COPY --from=flyteconsole /app/dist cmd/single/dist
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -tags console -v -o dist/flyte cmd/main.go


FROM gcr.io/distroless/base-debian11

COPY --from=flytebuilder /flyteorg/build/dist/flyte /flyte
ENTRYPOINT [ "/flyte" ]
