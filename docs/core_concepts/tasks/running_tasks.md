(remote_task)=

# Running tasks

## Flytectl

This is a multi-step process where we create an execution spec file, update the spec file, and then create the execution.
More details can be found in the [Flytectl API reference](https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html).

**Generate execution spec file**

```
flytectl get tasks -d development -p flytesnacks workflows.example.generate_normal_df  --latest --execFile exec_spec.yaml
```

**Update the input spec file for arguments to the workflow**

```
iamRoleARN: 'arn:aws:iam::12345678:role/defaultrole'
inputs:
  n: 200
  mean: 0.0
  sigma: 1.0
kubeServiceAcct: ""
targetDomain: ""
targetProject: ""
task: workflows.example.generate_normal_df
version: "v1"
```

**Create execution using the exec spec file**

```
flytectl create execution -p flytesnacks -d development --execFile exec_spec.yaml
```

**Monitor the execution by providing the execution id from create command**

```
flytectl get execution -p flytesnacks -d development <execid>
```

## FlyteRemote

A task can be launched via FlyteRemote programmatically.

```python
from flytekit.remote import FlyteRemote
from flytekit.configuration import Config, SerializationSettings

# FlyteRemote object is the main entrypoint to API
remote = FlyteRemote(
    config=Config.for_endpoint(endpoint="flyte.example.net"),
    default_project="flytesnacks",
    default_domain="development",
)

# Get Task
flyte_task = remote.fetch_task(name="workflows.example.generate_normal_df", version="v1")

flyte_task = remote.register_task(
    entity=flyte_task,
    serialization_settings=SerializationSettings(image_config=None),
    version="v2",
)

# Run Task
execution = remote.execute(
     flyte_task, inputs={"n": 200, "mean": 0.0, "sigma": 1.0}, execution_name="task-execution", wait=True
)

# Or use execution_name_prefix to avoid repeated execution names
execution = remote.execute(
     flyte_task, inputs={"n": 200, "mean": 0.0, "sigma": 1.0}, execution_name_prefix="flyte", wait=True
)

# Inspecting execution
# The 'inputs' and 'outputs' correspond to the task execution.
input_keys = execution.inputs.keys()
output_keys = execution.outputs.keys()
```


## Create a `FlyteRemote` object (TODO merge in)

In {doc}`"Running a workflow locally" <../getting_started/running_a_workflow_locally>`, you saw
how to run Flyte workflows with `pyflyte run` in the case that you're working
with standalone scripts.

Once you're working with larger projects where you've registered workflows
to a Flyte cluster, we recommend using the {py:class}`~flytekit.remote.remote.FlyteRemote`
client to run workflows from a Python runtime. First, let's create a `FlyteRemote`
object:

```{code-cell} ipython3
:tags: [remove-output]

from flytekit.configuration import Config
from flytekit.remote import FlyteRemote

remote = FlyteRemote(
    config=Config.auto(),
    default_project="flytesnacks",
    default_domain="development",
)
```


## Running a task (TODO merge in)

You can also run individual tasks on a Flyte cluster using `FlyteRemote`:

`````{tabs}

````{group-tab} Locally imported

If you have access to the `@task`-decorated function in your Python runtime
environment, you can import and execute it directly:

```{code-block} python

from workflows.example import say_hello

execution = remote.execute(say_hello, inputs={"name": "Kermit"})
```

````

````{group-tab} Remotely fetched

Execute a task by fetching a `FlyteWorkflow` object from the remote
**FlyteAdmin** service, which essentially contains the metadata representing a
Flyte task that exists on a Flyte cluster backend.

```{code-block} python
flyte_task = remote.fetch_task(name="workflows.example.say_hello")
execution = remote.execute(flyte_task, inputs={"name": "Kermit"})
```

````

`````

```{note}
You can also launch tasks via `flytectl`, learn more in the {ref}`User Guide <remote_task>`
```
