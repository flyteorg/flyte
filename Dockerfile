# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

FROM golang:1.18-alpine3.15 as builder
RUN apk add git openssh-client make curl

# Create the artifacts directory
RUN mkdir /artifacts

# Pull GRPC health probe binary for liveness and readiness checks
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.11 && \
    wget -qO/artifacts/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /artifacts/grpc_health_probe

# COPY only the go mod files for efficient caching
COPY go.mod go.sum /go/src/github.com/flyteorg/flyteadmin/
WORKDIR /go/src/github.com/flyteorg/flyteadmin

# Pull dependencies
RUN go mod download


# COPY the rest of the source code
COPY . /go/src/github.com/flyteorg/flyteadmin/

# This 'linux_compile' target should compile binaries to the /artifacts directory
# The main entrypoint should be compiled to /artifacts/flyteadmin
RUN make linux_compile

# update the PATH to include the /artifacts directory
ENV PATH="/artifacts:${PATH}"

# This will eventually move to centurylink/ca-certs:latest for minimum possible image size
FROM alpine:3.15
LABEL org.opencontainers.image.source https://github.com/flyteorg/flyteadmin

COPY --from=builder /artifacts /bin

# Ensure the latest CA certs are present to authenticate SSL connections.
RUN apk --update add ca-certificates

RUN addgroup -S flyte && adduser -S flyte -G flyte
USER flyte

CMD ["flyteadmin"]
