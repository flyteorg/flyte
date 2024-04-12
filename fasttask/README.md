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
## build fast task worker image
    docker build -t hamersaw/fasttask:<tag> -f Dockerfile .
