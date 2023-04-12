# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

FROM --platform=${BUILDPLATFORM} golang:1.18-alpine3.15 as builder

ARG TARGETARCH
ENV GOARCH "${TARGETARCH}"
ENV GOOS linux

RUN apk add git openssh-client make curl

# COPY only the go mod files for efficient caching
COPY go.mod go.sum /go/src/github.com/flyteorg/flyteadmin/
WORKDIR /go/src/github.com/flyteorg/flyteadmin

# Pull dependencies
RUN go mod download

# COPY the rest of the source code
COPY . /go/src/github.com/flyteorg/flyteadmin/

# This 'linux_compile_scheduler' target should compile binaries to the /artifacts directory
# The main entrypoint should be compiled to /artifacts/flytescheduler
RUN make linux_compile_scheduler

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

CMD ["flytescheduler"]

