# Data Proxy Development Guide

This guide provides steps on how to develop and iterate changes for the Data Proxy service.

## Prerequisites

- go version v1.24.0+
- docker version 17.03+
- kubectl version v1.11.3+

## Use go v1.24

Currently, flyte-v2 uses go v1.24 for development.

```sh
go install golang.org/dl/go1.24.0@latest
go1.24.0 download
export GOROOT=$(go1.24.0 env GOROOT)
export PATH="$GOROOT/bin:$PATH"
```

## Local Development

Following steps require you to switch your working directory to the `dataproxy/`.

```sh
cd dataproxy
```

### Run Data Proxy Service

1. Create a kind cluster

```sh
kind create cluster --image=kindest/node:v1.26.0 --name flytev2
```

2. Deploy MinIO storage backend

```sh
kubectl apply -f deployment/minio.yaml
```

3. Wait for MinIO to be ready

```sh
kubectl wait --for=condition=ready pod -l app=minio -n flyte-dataproxy --timeout=60s
```

4. Port-forward MinIO for local access

```sh
# Port-forward both MinIO API (9000) and Console (9001)
kubectl port-forward -n flyte-dataproxy svc/minio 9000:9000 9001:9001
```

Keep this running in a separate terminal.

You can now access MinIO Console in your browser:
- URL: http://localhost:9001
- Username: `minioadmin`
- Password: `minioadmin`

5. Run the Data Proxy service

```sh
go run cmd/main.go --config config/config.example.yaml
```

The service will start on port 8088.

6. Test the service

```sh
# Check health
curl http://localhost:8088/healthz

# Check readiness
curl http://localhost:8088/readyz
```

## Scripts

Convenient scripts are provided in `dataproxy/test/scripts/` to interact with the service using `buf curl`.
Ensure the service is running before executing these scripts.

- `./dataproxy/test/scripts/create_upload_location.sh` - Create a signed upload URL with all options

### Usage

```sh
# Use default values
./dataproxy/test/scripts/create_upload_location.sh

# Customize with environment variables
PROJECT=myproject DOMAIN=staging ./dataproxy/test/scripts/create_upload_location.sh

# Custom filename
FILENAME=mydata.csv FILENAME_ROOT=my-upload ./dataproxy/test/scripts/create_upload_location.sh

# Custom endpoint
ENDPOINT=http://localhost:9000 ./dataproxy/test/scripts/create_upload_location.sh
```

### Clean Up

```sh
kind delete cluster --name flytev2
```

## Tests

### Run unit tests

```sh
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test package
go test -v ./service
```
