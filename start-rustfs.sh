#!/bin/bash
# Starts RustFS (S3-compatible object store) for local development.
# RustFS is a lightweight, fast alternative to MinIO.

CONTAINER_NAME="flyte-rustfs"
BUCKET_NAME="flyte-data"

# Stop any existing container
docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm "$CONTAINER_NAME" 2>/dev/null || true

echo "Starting RustFS..."
docker run --name "$CONTAINER_NAME" \
  -p 9000:9000 -p 9001:9001 \
  -e RUSTFS_ACCESS_KEY=minioadmin \
  -e RUSTFS_SECRET_KEY=minioadmin \
  -d rustfs/rustfs:latest

echo "Waiting for RustFS to be ready..."
for i in $(seq 1 15); do
  if curl -sf http://localhost:9000/health >/dev/null 2>&1; then
    echo "RustFS is ready!"
    # Create the default bucket via S3 PUT request
    curl -sf -X PUT "http://localhost:9000/$BUCKET_NAME" \
      -u minioadmin:minioadmin >/dev/null 2>&1 || true
    echo "  API:         http://localhost:9000"
    echo "  Console:     http://localhost:9001"
    echo "  Credentials: minioadmin / minioadmin"
    echo "  Bucket:      $BUCKET_NAME"
    echo ""
    echo "From k3d/orbstack pods: http://host.docker.internal:9000"
    exit 0
  fi
  sleep 1
done

echo "RustFS failed to start"
exit 1
