# Execute Tasks

We will demonstrate how tasks are being executed in flyte v2.

Quick recap on the components that will mainly be invoked when executing tasks:
- flyte-sdk: SDK for Flyte. This is used in client side for user to run and deploy task, and it's also packed
  to the flyte backend for executing actions/sub-actions, update status, etc...
- run service: An API that receive CreateRun gRPC request from SDK and submit to queue service.
- queue service: queue and submit action CR to the K8s etcd.
- state service: store the action status.
- executor: Watching actions CR from the K8s etcd and execute task with suitable handler based on the task type

The Terminologies:
- task: A flyte task that defined by the user, which can be deployed to the flyte backend.
- action: We call every task excution as an action
- run: We call the root action a run, with the name "a0"

## Run the task

When you try to run a task with `flyte run`, flyte will go through following process to execute your task:

### On local client

Image building stage:

1. If the image is already built before, directly use the image from image cache and no need to build again

2. Invoke the image builder. The image builder will generate a Dockerfile based on user's environment settings
   (e.g.  pip packages, apt packages, ...), then use Docker buildx to build the image and push to the remote
registry.

Serialize task

1. Create the tarball of the source file and upload to the storage in data plane. We will get the remote
   path to the code bundle.
2. Serialize task to wire format (task spec). The image cache and the code bundle info will be passed into the
   task spec by injecting into the container arguments. Container arguments are as below. Note that those are
   the arguments of the entrypoint for the pod.

```sh
[
    "a0",
    "--inputs", "{{.input}}",
    "--outputs-path", "{{.outputPrefix}}",
    "--version", "abc123def456",
    "--raw-data-path", "{{.rawOutputDataPrefix}}",
    ...
    "--image-cache", "H4sIAAAAA...",           # Compressed ImageCache
    "--tgz", "s3://bucket/code/abc123.tar.gz", # Code bundle path
    "--dest", ".",                              # Where to extract the code bundle to
    "--module-name", "my_module",
    "--task-name", "my_task"
]
```

3. Send `CreateRunRequest` RPC with task spec and serialized inputs to remote run service to create the run.

### On remote cluster

1. After run service received `CreateRunRequest` RPC request, the run service will store the run info into the DB and
   call `EnqueueAction` RPC to the queue service.
2. Queue service receives the `EnqueueAction` request and create a `TaskAction` CR in Kubernetes.
3. Then, we will create pod for the root action `a0` that will execute its entrypoint command, which will be
   something like following to execute flyte-sdk for running the action.


```sh
python -m flyte._internal.entrypoint \
    a0 \
    --inputs {{.input}} \
    --outputs-path {{.outputPrefix}} \
    --version abc123def456 \
    --tgz s3://bucket/code/abc123.tar.gz \
    --dest . \
    --resolver flyte._internal.resolvers.default.DefaultTaskResolver \
    --module-name my_module \
    --task-name my_task
```

4. By executing the above entrypoint command, we will execute the action `a0` in the pod with flyte-sdk.
   During the exeuction of actions, when a task function is being called, flyte will invoke the
`controller` in flyte-sdk to submit a new action through sending `EnqueueAction` RPC to queue service.

For example, when running to the say_hello() line in main(), flyte-sdk will submit a new action `say_hello()`
to the queue service throuch `controller`.

```py
@env.task
async def say_hello(data: str, lt: List[int]) -> str:
    print(f"Hello, world! - {flyte.ctx().action}")
    return f"Hello {data} {lt}"

@env.task
async def main(data: str = "default string", n: int = 3) -> str:
    print(f"Hello, nested! - {flyte.ctx().action}")

    return await say_hello(data=data, lt=vals)
```


5. When submitting an sub-action, we will also create an `informer` that will link to the action and watch the
   state service for the action status update.
