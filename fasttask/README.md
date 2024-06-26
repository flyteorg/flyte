# fast task
FastTask is a Flyte plugin to execute tasks quickly using a persistent, external worker; reducing the overhead of cold-starting a k8s Pod as an execution environment. This enables Flyte to execute short tasks in parallel, without the overhead of starting a new Pod for each task; instead sharing a pool of workers across multiple tasks and workflow executions.

## development setup
    # get flytekit in terminal
    source ~/Development/flytekit/.venv-union/bin/activate

    # create temporary python dir for fast registration
    mkdir /tmp/fasttask-test
    export PYTHONPATH=$PYTHONPATH:/tmp/fasttask-test

    # allow flytekit to connect to minio cluster
    export FLYTE_AWS_ACCESS_KEY_ID=minio
    export FLYTE_AWS_SECRET_ACCESS_KEY=miniostorage
    export FLYTE_AWS_ENDPOINT=http://localhost:30084

    ./target/debug/worker bridge --queue-id=bar --fast-register-dir-override /tmp/fasttask-test
## build fast task worker image
    docker build -t hamersaw/fasttask:<tag> -f Dockerfile .

# DevX (worker only) guide using the flytectl demo sandbox

## Flyte Single Binary Environment Setup
### Build sandbox image
```bash
cd docker/sandbox-bundled
make build
```
This should build a new local image called `flyte-sandbox-private` (as opposed to the oss flyte, which should produce a local image just called `flyte-sandbox`). 

### Run the sandbox
Now with the image built, we can use it to run both the private fork, as well as all the dependencies (minio, postgres, etc.).
But since we added a new port, we need to run the custom version of flytectl in this private fork that forwards that new port.

From the root of this private fork
```bash
go run flytectl/main.go demo start --image flyte-sandbox-private:latest --dev
```
Running in dev mode allows you to then run the private fork single binary in an IDE, which you can start with the `flyte-single-binary-local.yaml` file in this repository. You'll need to search and replace all instances of `$HOME`.

## Python Setup
tbd...
Check out the repo with the sample file.
Create a python virtual env.


