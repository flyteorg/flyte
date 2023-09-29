FROM --platform=${BUILDPLATFORM} golang:1.19-alpine3.16 as builder

ARG TARGETARCH
ENV GOARCH "${TARGETARCH}"
ENV GOOS linux

RUN apk add git openssh-client make curl

# Create the artifacts directory
RUN mkdir /artifacts


# COPY only the go mod files for efficient caching
COPY go.mod go.sum /go/src/github.com/flyteorg/flyte/flyteartifacts/
WORKDIR /go/src/github.com/flyteorg/flyte/flyteartifacts/


# Pull dependencies
RUN go mod download

# COPY the rest of the source code
COPY . /go/src/github.com/flyteorg/flyte/flyteartifacts/

# This 'linux_compile' target should compile binaries to the /artifacts directory
# The main entrypoint should be compiled to /artifacts/flyteadmin
RUN make linux_compile

# update the PATH to include the /artifacts directory
ENV PATH="/artifacts:${PATH}"

# This will eventually move to centurylink/ca-certs:latest for minimum possible image size
FROM alpine:3.16
LABEL org.opencontainers.image.source https://github.com/flyteorg/flyte/

COPY --from=builder /artifacts /bin

# Ensure the latest CA certs are present to authenticate SSL connections.
RUN apk --update add ca-certificates

RUN addgroup -S flyte && adduser -S flyte -G flyte
USER flyte

CMD ["artifacts"]
