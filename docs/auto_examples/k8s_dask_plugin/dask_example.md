---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(intermediate_using_dask_tasks)=

# Writing a Dask Task

Flyte has an optional plugin that makes it possible to run [Dask](https://www.dask.org/) jobs natively on your
kubernetes cluster. It makes it extremely easy to run your `dask` code as a task. The plugin creates a new
virtual/ephemeral cluster for each `dask` task, where Flyte manages the cluster lifecycle.

## Dask in flytekit

For a more complete example refer to the {std:ref}`example-dask`

1. Enable the `dask` plugin in the backend, following the steps from the previous section

2. Install the [flytekit dask plugin](https://pypi.org/project/flytekitplugins-dask/)

   ```
   pip install flytekitplugins-dask
   ```

3. Write regular `dask` code - with one change in the `@task` decorator. Refer to the example below:

   > ```python
   > @task(
   >     task_config=Dask(
   >         job_pod_spec=JobPodSpec(
   >             limits=Resources(cpu="1", mem="2Gi"),
   >         ),
   >         cluster=DaskCluster(
   >             n_workers=10,
   >             limits=Resources(cpu="4", mem="10Gi"),
   >         ),
   >     ),
   >     cache_version="1",
   >     cache=True,
   > )
   > def hello_dask(size: int) -> float:
   >     ...
   >     client = Client()  # Create a client as you would in local code
   >     # Regular dask code
   >     ...
   > ```

4. Run it locally

   ```python
   hello_dask(size=10)
   ```

5. Use it in a workflow

6. Run it on a remote cluster

## Examples

(example-dask)=

### How Flytekit Simplifies the Usage of dask in User Code

The task `hello_dask` runs a new dask cluster, which when run locally starts a local `dask` cluster, but when run
remote spins up an arbitrarily sized cluster depending on the specified dask configuration.

```{code-cell}
from dask import array as da
from flytekit import Resources, task
```

+++ {"lines_to_next_cell": 0}

The following imports are required to configure the Dask cluster in Flyte

```{code-cell}
:lines_to_next_cell: 1

from flytekitplugins.dask import Dask, WorkerGroup
```

### Dask Task Sample

This example shows how a Dask task can be written simply by adding a `@task(task_config=Dask(...), ...)` decorator.
Refer to the [Dask](https://github.com/flyteorg/flytekit/blob/4b1675ffb85648dc5742e9a6dea98b94714963e1/plugins/flytekit-dask/flytekitplugins/dask/task.py#L54-L63)
class to understand the various configuration options.

```{code-cell}
@task(
    task_config=Dask(
        workers=WorkerGroup(
            number_of_workers=10,
            limits=Resources(cpu="4", mem="10Gi"),
        ),
    ),
    limits=Resources(cpu="1", mem="2Gi"),
    cache_version="1",
    cache=True,
)
def hello_dask(size: int) -> float:
    # Dask will implicitly create a Client in the background by calling Client(). When executing
    # remotely, this Client() will use the deployed ``dask`` cluster.
    array = da.random.random(size)
    return float(array.mean().compute())
```

+++ {"lines_to_next_cell": 0}

The function can be executed locally:
Guard with:

> if \_\_name\_\_ == '\_\_main\_\_':

in a local Python script

```{code-cell}
print(hello_dask(size=1000))
```
