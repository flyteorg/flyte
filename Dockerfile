# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

# Using go1.10.4
FROM golang:1.13.3-alpine3.10 as builder
RUN apk add git openssh-client make curl dep

# COPY only the dep files for efficient caching
COPY Gopkg.* /go/src/github.com/lyft/datacatalog/
WORKDIR /go/src/github.com/lyft/datacatalog

# Pull dependencies
RUN dep ensure -vendor-only

# COPY the rest of the source code
COPY . /go/src/github.com/lyft/datacatalog/

# This 'linux_compile' target should compile binaries to the /artifacts directory
# The main entrypoint should be compiled to /artifacts/datacatalog
RUN make linux_compile

# update the PATH to include the /artifacts directory
ENV PATH="/artifacts:${PATH}"

# This will eventually move to centurylink/ca-certs:latest for minimum possible image size
FROM alpine:3.10
COPY --from=builder /artifacts /bin

RUN apk --update add ca-certificates

CMD ["datacatalog"]
