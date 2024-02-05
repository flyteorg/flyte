---
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(getting_started_optimizing_tasks)=

# Optimizing Tasks

There are many ways to optimize your tasks and workflows in Flyte, and this guide
will take you through some of the common methods for doing so.

## Caching

Caching allows you to avoid re-running potentially expensive tasks. You can
specify which tasks to cache with the `cache` and `cache_version` arguments:

```{code-cell} ipython3
from typing import List
from flytekit import task, workflow


@task(cache=True, cache_version="1")
def compute_mean(data: List[float]) -> float:
    return sum(data) / len(data)

@workflow
def wf(data: List[float]) -> float:
    return compute_mean(data=data)
```

Caching works both locally and on a Flyte backend.

```{code-cell} ipython3
%timeit -n 1 -r 1 wf(data=[float(x) for x in range(100_000)])
```

```{code-cell} ipython3
%timeit -n 1 -r 1 wf(data=[float(x) for x in range(100_000)])
```

As you can see, the second call to the `wf` workflow takes less time because
Flyte simply hits the cache to obtain the result.

```{note}
For file-like data types like {py:class}`flytekit.types.file.FlyteFile` and offloaded
data types like `pandas.DataFrame` objects, you can provide a hash function that
represents the cache key. Learn more in the {ref}`User Guide <cache-offloaded-objects>`.
```

## Retries

Flyte also allows you to automatically retry failing tasks in the case of
system-level or catastrophic errors that may arise from issues that don't have
anything to do with user-defined code, like network issues and data center
outages.

The following version of the `compute_mean` task simulates these kinds of
errors by randomly throwing a `RuntimeError` 5% of the time:

```{code-cell} ipython3
import random

@task(retries=3)
def compute_mean(data: List[float]) -> float:
    if random() < 0.05:
        raise RuntimeError("Something bad happened 🔥")
    return sum(data) / len(data)
```

```{note}
Retries only take effect when running a task on a Flyte cluster.
```

## Timeouts

To protect against zombie tasks that hang due to system-level issues, you
can supply the `timeout` argument to the `@task` decorator to make sure that
problematic tasks adhere to a maximum runtime.

In this example, we make sure that the task is terminated after it's been
running for more that one hour.

```{code-cell} ipython3
from datetime import timedelta

@task(timeout=timedelta(hours=1))
def compute_mean(data: List[float]) -> float:
    return sum(data) / len(data)
```

Notice that the `timeout` argument takes a built-in Python
{py:class}`~datetime.timedelta` object.

## Map Tasks

If you need to parallelize a task, you can use the {py:func}`~flytekit.map_task`
construct. A mappable task is one that takes in a single argument and produces
some output.

In this example, we partition our data into chunks so that we can horizontally
scale our workload:

```{code-cell} ipython3
import math
from typing import Tuple

from flytekit import map_task


@task
def sum_and_length(data: List[float]) -> List[float]:
    """Return the sum and length of a dataset of numbers."""
    return [sum(data), float(len(data))]


@task
def prepare_partitions(data: List[float], n_partitions: int) -> List[List[float]]:
    """Create partitions from the full dataset."""
    size = math.ceil(len(data) / n_partitions)
    return [data[size * i: size * (i + 1)] for i in range(n_partitions)]

@task
def reduce(results: List[List[float]]) -> float:
    """Combine results from the map task."""
    total, length = 0.0, 0.0
    for sub_total, sub_length in results:
        total += sub_total
        length += sub_length
    return total / length


@workflow
def parallelized_compute_mean(data: List[float], n_partitions: int = 10) -> float:
    """An embarrassingly parallel implementation to compute the mean from data."""
    partitioned_data = prepare_partitions(data=data, n_partitions=n_partitions)

    # use map_task to apply the sum_and_length task to the partitions
    results = map_task(sum_and_length)(data=partitioned_data)
    return reduce(results=results)


parallelized_compute_mean(data=[float(x) for x in range(10_000)])
```

## Resource Allocation

As one of the core features of Flyte, workflows can be composed of tasks that
potentially have heterogeneous resource requirements. You can express this with
the {py:class}`~flytekit.Resources` object:

```{code-cell} ipython3
from flytekit import Resources


@task(requests=Resources(cpu="2", mem="100Mi"))
def light_task() -> float:
    ...


@task(requests=Resources(cpu="16", mem="16Gi"))
def heavy_task() -> float:
    ...
```

## Multi-image Workflows

In addition to task-level resource configuration, you can
also specify different images per task. This is particularly useful if some
tasks in your workflow have a different set of dependencies (e.g. require CUDA
to be installed for model training) where most of the other tasks can use
another image.

In this example we specify two tasks: one that uses CPUs and another that uses
GPUs. For the former task, we use the default image that ships with `flytekit`
and for the latter task, we specify a pre-built image that the core Flyte team
maintains that enables distributed training with the
{ref}`Kubeflow Pytorch integration <kf-pytorch-op>`.

```{code-cell} ipython3
import numpy as np
import torch.nn as nn

@task(
    requests=Resources(cpu="2", mem="16Gi"),
    container_image="ghcr.io/flyteorg/flytekit:py3.9-latest",
)
def get_data() -> Tuple[np.ndarray, np.ndarray]:
    ...  # get dataset as numpy ndarrays


@task(
    requests=Resources(cpu="4", gpu="1", mem="16Gi"),
    container_image="ghcr.io/flyteorg/flytecookbook:kfpytorch-latest",
)
def train_model(features: np.ndarray, target: np.ndarray) -> nn.Module:
    ...  # train a model using gpus
```

These tasks assume that we're going to use the {ref}`pyflyte register <getting_started_pyflyte_register>`
command to register these tasks, since these static images will not contain
the code that we defined above. Using `pyflyte register` ensures that `get_data`
and `train_model` are zipped up and Flyte has access to it when they're executed
on a Flyte backend.

```{important}
You can also configure the container images dynamically. See the
{ref}`User Guide <multi_images>` for more details.
```


## Declarative Infrastructure

Finally, staying with the theme of Flyte's ability to handle heterogeneous
workloads at the most granular level, you can configure tasks to leverage
third-party infrastructure via the extensible task plugin system.

As we saw with the `train_model` task example, we're using a CUDA-enabled image,
but in order to do distributed training, we'll have to leverage the
{py:class}`~flytekitplugins.kfpytorch.PyTorch` plugin:

```{code-cell} ipython3
from flytekitplugins.kfpytorch import PyTorch

@task(
    task_config=PyTorch(num_workers=2),
    requests=Resources(cpu="2", gpu="1", mem="8Gi"),
    limits=Resources(cpu="4", gpu="2", mem="16Gi"),
    container_image="ghcr.io/flyteorg/flytecookbook:kfpytorch-latest",
)
def train_model(features: np.ndarray, target: np.ndarray) -> nn.Module:
    ...  # train a model using gpus

```

This plugin highlights one of the most powerful abilities that you gain with
Flyte: the ability to *declaratively specify infrastructure requirements*
at the most granular level of your workflow!

When this task is executed on a Flyte cluster, it automatically provisions all of
the resources that you need. In this case, that need is distributed
training, but Flyte also provides integrations for {ref}`Spark <plugins-spark-k8s>`,
{ref}`Ray <kube-ray-op>`, {ref}`MPI <kf-mpi-op>`, {ref}`Sagemaker <aws-sagemaker>`,
{ref}`Snowflake <plugins-snowflake>`, and more.

Even though Flyte itself is a powerful compute engine and orchestrator for
data engineering, machine learning, and analytics, perhaps you have existing
code that leverages other platforms. Flyte recognizes the pain of migrating code,
so its plugin system enables you to call out to third-party services and
infrastructure when needed so that you can embed existing workloads into the
Flyte programming paradigm.

## What's Next?

In this guide, you learned the various ways in which you can optimize your
tasks and workflows to make them more scalable and robust. In the final stop of
the Flyte Fundamentals tour, we'll see how to extend Flyte in the cases where
the built-in functionality doesn't quite fit your needs.
