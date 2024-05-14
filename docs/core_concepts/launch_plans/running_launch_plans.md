(remote_launchplan)=

# Running launch plans

## Flytectl

This is multi-steps process where we create an execution spec file, update the spec file and then create the execution.
More details can be found [here](https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html).

**Generate an execution spec file**

```
flytectl get launchplan -p flytesnacks -d development myapp.workflows.example.my_wf  --execFile exec_spec.yaml
```

**Update the input spec file for arguments to the workflow**

```
....
inputs:
    name: "adam"
....
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

A launch plan can be launched via FlyteRemote programmatically.

```python
from flytekit.remote import FlyteRemote
from flytekit.configuration import Config
from flytekit import LaunchPlan

# FlyteRemote object is the main entrypoint to API
remote = FlyteRemote(
    config=Config.for_endpoint(endpoint="flyte.example.net"),
    default_project="flytesnacks",
    default_domain="development",
)

# Fetch launch plan
flyte_lp = remote.fetch_launch_plan(
    name="workflows.example.wf", version="v1", project="flytesnacks", domain="development"
)

# Execute
execution = remote.execute(
    flyte_lp, inputs={"mean": 1}, execution_name="lp-execution", wait=True
)

# Or use execution_name_prefix to avoid repeated execution names
execution = remote.execute(
    flyte_lp, inputs={"mean": 1}, execution_name_prefix="flyte", wait=True
)
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

## Running a launch plan (TODO merge in)

Similar to workflows, you can run launch plans with `FlyteRemote`:

`````{tabs}

````{group-tab} Locally imported

If you have a `LaunchPlan` defined in your Python runtime environment, you can
execute it directly:

```{code-block} python

from workflows.example import wf

launch_plan = LaunchPlan.get_or_create(
    wf, name="launch_plan", default_inputs={"name": "Elmo"},
)

execution = remote.execute(launch_plan, inputs={})
```

````

````{group-tab} Remotely fetched

Execute a task by fetching a `FlyteLaunchPlan` object from the remote
**FlyteAdmin** service, which essentially contains the metadata representing a
Flyte task that exists on a Flyte cluster backend.

This example assumes that you've added a `launch_plan` with some default inputs
to the `example.py` script and registered it to the backend:

```{code-block} python
flyte_launchplan = remote.fetch_launch_plan(name="workflows.example.launch_plan")
execution = remote.execute(flyte_launchplan, inputs={})
```

````

`````
