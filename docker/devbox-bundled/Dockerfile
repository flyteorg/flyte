# syntax=docker/dockerfile:1.4-labs
FROM --platform=${BUILDPLATFORM} mgoltzsche/podman:minimal AS builder

ARG TARGETARCH
ENV TARGETARCH "${TARGETARCH}"

WORKDIR /build

COPY images/manifest.txt images/preload ./
RUN --security=insecure ./preload manifest.txt


FROM --platform=${BUILDPLATFORM} golang:1.24-bullseye AS bootstrap

ARG TARGETARCH
ENV CGO_ENABLED 0
ENV GOARCH "${TARGETARCH}"
ENV GOOS linux

WORKDIR /flyteorg/build
COPY bootstrap/go.mod bootstrap/go.sum ./
RUN go mod download
COPY bootstrap/ ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/go/pkg/mod \
    go build -o dist/flyte-devbox-bootstrap cmd/bootstrap/main.go && \
    go build -o dist/embedded-postgres cmd/embedded-postgres/main.go


# Pre-download PostgreSQL binaries from Maven for the embedded-postgres library.
# The library expects a cached .txz file; we also extract the dynamic linker since
# the K3s base image has no libc.
FROM debian:bookworm-slim AS pg-cache

ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends curl unzip xz-utils ca-certificates && rm -rf /var/lib/apt/lists/*

RUN set -ex; \
    ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "arm64v8" || echo "amd64"); \
    PG_VERSION="16.9.0"; \
    CACHE_FILE="embedded-postgres-binaries-linux-${ARCH}-${PG_VERSION}.txz"; \
    mkdir -p /cache /glibc-libs; \
    # Download JAR (actually a ZIP) from Maven Central
    curl -fL "https://repo1.maven.org/maven2/io/zonky/test/postgres/embedded-postgres-binaries-linux-${ARCH}/${PG_VERSION}/embedded-postgres-binaries-linux-${ARCH}-${PG_VERSION}.jar" -o /tmp/pg.jar; \
    # Extract the .txz from inside the JAR
    unzip -p /tmp/pg.jar "*.txz" > "/cache/${CACHE_FILE}"; \
    rm -f /tmp/pg.jar; \
    # Collect glibc libraries needed by PostgreSQL binaries (not included in zonkyio bundle)
    mkdir -p /tmp/pg-tmp && tar xJf "/cache/${CACHE_FILE}" -C /tmp/pg-tmp/; \
    for bin in /tmp/pg-tmp/bin/*; do \
        ldd "$bin" 2>/dev/null | grep "=>" | awk '{print $3}' | while read lib; do \
            [ -f "$lib" ] && cp -n "$lib" /glibc-libs/ 2>/dev/null || true; \
        done; \
    done; \
    # Copy dynamic linker
    cp /lib/ld-linux-aarch64.so.1 /glibc-libs/ 2>/dev/null || true; \
    cp /lib/aarch64-linux-gnu/ld-linux-aarch64.so.1 /glibc-libs/ 2>/dev/null || true; \
    cp /lib64/ld-linux-x86-64.so.2 /glibc-libs/ 2>/dev/null || true; \
    cp /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 /glibc-libs/ 2>/dev/null || true; \
    rm -rf /tmp/pg-tmp


FROM debian:bookworm-slim AS knative-manifests

RUN apt-get update && apt-get install -y --no-install-recommends curl python3 python3-yaml ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /knative

RUN set -ex; \
    curl -fsSL "https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-crds.yaml" \
        -o 00-serving-crds.yaml; \
    curl -fsSL "https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-core.yaml" \
        -o 01-serving-core.yaml; \
    curl -fsSL "https://github.com/knative-extensions/net-kourier/releases/download/knative-v1.21.0/kourier.yaml" \
        -o /tmp/kourier-orig.yaml

# Patch kourier Service: change type to NodePort and assign nodePort 30081 on the http2 port.
# We load and re-dump the full manifest so the selector and other fields are preserved exactly.
# Outer <<'DOCKEREOF' is the Dockerfile-level heredoc (becomes the shell script).
# Inner << 'PYEOF' is a shell-level heredoc piping Python code to python3's stdin.
RUN <<'DOCKEREOF'
python3 - << 'PYEOF'
import yaml

with open('/tmp/kourier-orig.yaml') as f:
    docs = [d for d in yaml.safe_load_all(f) if d is not None]

for doc in docs:
    meta = doc.get('metadata', {})
    if (doc.get('kind') == 'Service'
            and meta.get('name') == 'kourier'
            and meta.get('namespace') == 'kourier-system'):
        doc['spec']['type'] = 'NodePort'
        for port in doc['spec'].get('ports', []):
            if port.get('name') == 'http2':
                port['nodePort'] = 30081

with open('02-kourier.yaml', 'w') as f:
    yaml.dump_all(docs, f, default_flow_style=False)
PYEOF
DOCKEREOF

# ConfigMap overrides applied after the knative manifests above.
# Alphabetical ordering (03/04) ensures these win over the defaults in 01/02.
RUN <<'DOCKEREOF'
cat > 03-config-network.yaml << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-network
  namespace: knative-serving
data:
  ingress-class: kourier.ingress.networking.knative.dev
  domain-template: "{{.Name}}-{{.Namespace}}.{{.Domain}}"
EOF
DOCKEREOF

RUN <<'DOCKEREOF'
cat > 04-config-domain.yaml << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-domain
  namespace: knative-serving
data:
  "localhost": ""
EOF
DOCKEREOF


FROM rancher/k3s:v1.34.6-k3s1

ARG TARGETARCH

ARG FLYTE_DEVBOX_VERSION
ENV FLYTE_DEVBOX_VERSION "${FLYTE_DEVBOX_VERSION}"

COPY --from=builder /build/images/ /var/lib/rancher/k3s/agent/images/
COPY images/tar/${TARGETARCH}/ /var/lib/rancher/k3s/agent/images/
COPY manifests/ /var/lib/rancher/k3s/server/manifests-staging/
COPY --from=knative-manifests /knative/ /var/lib/rancher/k3s/server/manifests/knative/
COPY bin/ /bin/

# Install bootstrap and embedded postgres
COPY --from=bootstrap /flyteorg/build/dist/flyte-devbox-bootstrap /bin/
COPY --from=bootstrap /flyteorg/build/dist/embedded-postgres /bin/

# Install pre-cached PostgreSQL binaries and glibc libraries
COPY --from=pg-cache /cache/ /var/cache/embedded-postgres/
COPY --from=pg-cache /glibc-libs/ /usr/lib/pg-glibc/

# Create dynamic linker symlinks and add postgres user/group
RUN for f in /usr/lib/pg-glibc/ld-linux-aarch64*; do \
      [ -f "$f" ] && ln -sf "$f" /lib/$(basename "$f"); \
    done 2>/dev/null; \
    for f in /usr/lib/pg-glibc/ld-linux-x86-64*; do \
      [ -f "$f" ] && mkdir -p /lib64 && ln -sf "$f" /lib64/$(basename "$f"); \
    done 2>/dev/null; \
    echo "postgres:x:999:999:PostgreSQL:/tmp:/bin/sh" >> /etc/passwd && \
    echo "postgres:x:999:" >> /etc/group

ENV LD_LIBRARY_PATH="/usr/lib/pg-glibc"

VOLUME /var/lib/flyte/storage

# Set environment variable for picking up additional CA certificates
ENV SSL_CERT_DIR /var/lib/flyte/config/ca-certificates

ENTRYPOINT [ "/bin/k3d-entrypoint.sh" ]
CMD [ "server", "--disable=servicelb", "--disable=metrics-server" ]
