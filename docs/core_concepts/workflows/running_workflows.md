# Running workflows (TODO merge content from "Running and scheduling workflows" doc below)

Workflows on their own are not runnable directly. However, a launchplan is always bound to a workflow and you can use
launchplans to **launch** a workflow. For cases in which you want the launchplan to have the same arguments as a workflow,
if you are using one of the SDK's to author your workflows - like flytekit, flytekit-java etc, then they should
automatically create a `default launchplan` for the workflow.

A `default launchplan` has the same name as the workflow and all argument defaults are similar. See
{ref}`Launch Plans` to run a workflow via the default launchplan.

{ref}`Tasks also can be executed <task>` using the launch command.
One difference between running a task and a workflow via launchplans is that launchplans cannot be associated with a
task. This is to avoid triggers and scheduling.

## FlyteRemote

Workflows can be executed with FlyteRemote because under the hood it fetches and triggers a default launch plan.

```python
from flytekit.remote import FlyteRemote
from flytekit.configuration import Config

# FlyteRemote object is the main entrypoint to API
remote = FlyteRemote(
    config=Config.for_endpoint(endpoint="flyte.example.net"),
    default_project="flytesnacks",
    default_domain="development",
)

# Fetch workflow
flyte_workflow = remote.fetch_workflow(name="workflows.example.wf", version="v1")

# Execute
execution = remote.execute(
    flyte_workflow, inputs={"mean": 1}, execution_name="workflow-execution", wait=True
)

# Or use execution_name_prefix to avoid repeated execution names
execution = remote.execute(
    flyte_workflow, inputs={"mean": 1}, execution_name_prefix="flyte", wait=True
)
```


## Create a `FlyteRemote` object

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

## Running a workflow

You can run workflows using the `FlyteRemote` {py:meth}`~flytekit.remote.remote.FlyteRemote.execute`
method, where you need to pass in a dictionary of `inputs` that adhere to the
interface defined by the workflow.

`````{tabs}

````{group-tab} Locally imported

If you have access to the `@workflow`-decorated function in your Python runtime
environment, you can import and execute it directly:

Before execute it directly, you need to register the workflow first.

```{prompt} bash $
pyflyte register wf.py
```

```{code-block} python

from workflows.example import wf

execution = remote.execute(
    wf,
    inputs={"name": "Kermit"},
)
```

````

````{group-tab} Remotely fetched

Execute a workflow by fetching a `FlyteWorkflow` object from the remote
**FlyteAdmin** service, which essentially contains the metadata representing a
Flyte workflow that exists on a Flyte cluster backend.

```{code-block} python
flyte_wf = remote.fetch_workflow(name="workflows.example.wf")
execution = remote.execute(flyte_wf, inputs={"name": "Kermit"})
```

````

`````

```{note}
You can also launch workflows via `flytectl` which you can learn more about in
the {ref}`User Guide <remote_launchplan>`.
```