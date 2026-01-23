# Executor Development Guide

This guide provides steps on how to develop and iterates changes.

## Prerequisites
- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.

### Setup on kind

We recommend using kind to create a Kubernetes cluster for local development.

### Use go v1.24

Currently, flyte-v2 use go v1.24 for development.

```sh
go install golang.org/dl/go1.24.0@latest
go1.24.0 download
export GOROOT=$(go1.24.0 env GOROOT)
export PATH="$GOROOT/bin:$PATH"
```

### Clean up local binaries

To keep consistent code generation and testing result, it is recommended to remove outdated binaries installed by the Makefile.

```sh
make clean
```

## Local development on kind cluster

Following steps requires you to switch your working directory to the `executor/`.

```sh
cd executor
```

### Run executor inside the cluster

1. Create a kind cluster

```sh
kind create cluster --image=kindest/node:v1.26.0 --name flytev2
```

2. Build the image

```sh
IMG=flyteorg/executor:nightly make docker-build
```

3. Load image into kind cluster

```sh
kind load docker-image flyteorg/executor:nightly --name flytev2
```

4. Install CRD into the cluster

```sh
make install
```

5. Deploy the Manager to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=flyteorg/executor:nightly
```

6. Clean up

```sh
kind delete cluster --name flytev2
```

### Run executor outside the cluster 

1. Create a kind cluster

```sh
kind create cluster --image=kindest/node:v1.26.0 --name flytev2
```

2. Install CRD into the cluster

```sh
make install
```

3. Compile the source code and run

```sh
# Compile the source code
make build
# Run the executor
./bin/manager
```

## Tests

### Manual test

You can apply the samples from the config/sample using the following command:

```sh
kubectl apply -k config/samples/
```

You can see the `TaskAction` CRD created:

```sh
❯ kubectl get taskactions
NAME                RUN          ACTION        STATUS      AGE
taskaction-sample   sample-run   sample-task   Completed   59m
```

Or use `-o wide` to view more details:

```sh
❯ kubectl get taskactions -o wide
NAME                RUN          ACTION        STATUS      AGE   PROGRESSING   SUCCEEDED   FAILED
taskaction-sample   sample-run   sample-task   Completed   59m   False         True
```

## Modify CRD

### Quick Steps

1. Modify the types file: Edit `api/v1/taskaction_types.go`
   - Add/modify fields in `TaskActionSpec` or `TaskActionStatus`
   - Add/modify printcolumns (the `// +kubebuilder:printcolumn` comments above the `TaskAction` struct)
   - Add validation rules using `// +kubebuilder:validation:` markers

2. Generate CRD manifests and DeepCopy code

```sh
make manifests generate
```

This runs two commands:
- `make manifests`: Updates YAML manifests:
  - `config/crd/bases/flyte.org_taskactions.yaml` (CRD with schema, validation, printcolumns)
  - `config/rbac/role.yaml` (RBAC permissions)
- `make generate`: Updates Go code:
  - `api/v1/zz_generated.deepcopy.go` (DeepCopy methods required by Kubernetes)

3. Update CRD in the cluster

```sh
make install
```
