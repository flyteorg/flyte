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

+++ {"lines_to_next_cell": 0}

# Pod Example

Pod configuration for a Flyte task allows you to run multiple containers within a single task.
They provide access to a fully customizable [Kubernetes pod spec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#podspec-v1-core),
which can be used to modify the runtime of the task execution.

The primary container is the main driver for Flyte task execution and is responsible for producing inputs and outputs.

Pod tasks accept arguments that are commonly used with container tasks, such as resource specifications.
However, these arguments only apply to the primary container.
To customize the other containers that are brought up during task execution, you can define a complete pod spec using the
[Kubernetes Python client library](https://github.com/kubernetes-client/python)'s,
[V1PodSpec](https://github.com/kubernetes-client/python/blob/master/kubernetes/client/models/v1_pod_spec.py).

+++ {"lines_to_next_cell": 0}

First, we import the necessary libraries for use in the following examples.

```{code-cell}
import os
import time
from typing import List

from flytekit import Resources, TaskMetadata, dynamic, map_task, task, workflow
from flytekitplugins.pod import Pod
from kubernetes.client.models import (
    V1Container,
    V1EmptyDirVolumeSource,
    V1PodSpec,
    V1ResourceRequirements,
    V1Volume,
    V1VolumeMount,
)
```

+++ {"lines_to_next_cell": 0}

## Add additional properties to the task container

In this example, we define a simple pod specification.
The `containers` field is set to an empty list, signaling to Flyte to insert a placeholder primary container.
The `node_selector` field specifies the nodes on which the task pod should be run.

```{code-cell}
@task(
    task_config=Pod(
        pod_spec=V1PodSpec(
            containers=[],
            node_selector={"node_group": "memory"},
        ),
    ),
    requests=Resources(
        mem="1G",
    ),
)
def pod_task() -> str:
    return "Hello from pod task!"


@workflow
def pod_workflow() -> str:
    return pod_task()
```

:::{note}
To configure default settings for all pods Flyte creates, including tasks for pods, containers, PyTorch, Spark, Ray, and Dask,
[configure a default Kubernetes pod template](https://docs.flyte.org/en/latest/deployment/cluster_config/general.html#using-default-k8s-podtemplates).
:::

+++ {"lines_to_next_cell": 0}

## Multiple containers

We define a simple pod spec with a shared volume that is mounted in both the primary and secondary containers.
The secondary container writes a file that the primary container waits for and reads before completing.

First, we define the shared data path:

```{code-cell}
_SHARED_DATA_PATH = "/data/message.txt"
```

+++ {"lines_to_next_cell": 0}

We define a pod spec with two containers.
While pod tasks generally allow you to customize Kubernetes container attributes, you can still use Flyte directives to specify resources and the image.
Unless you specify the `container_image` task attribute, the default image built for Flyte tasks will be used.

```{code-cell}
@task(
    task_config=Pod(
        pod_spec=V1PodSpec(
            containers=[
                V1Container(
                    name="primary",
                    resources=V1ResourceRequirements(
                        requests={"cpu": "1", "memory": "100Mi"},
                        limits={"cpu": "1", "memory": "100Mi"},
                    ),
                    volume_mounts=[
                        V1VolumeMount(
                            name="shared-data",
                            mount_path="/data",
                        )
                    ],
                ),
                V1Container(
                    name="secondary",
                    image="alpine",
                    command=["/bin/sh"],
                    args=[
                        "-c",
                        "echo hi pod world > {}".format(_SHARED_DATA_PATH),
                    ],
                    resources=V1ResourceRequirements(
                        requests={"cpu": "1", "memory": "100Mi"},
                        limits={"cpu": "1", "memory": "100Mi"},
                    ),
                    volume_mounts=[
                        V1VolumeMount(
                            name="shared-data",
                            mount_path="/data",
                        )
                    ],
                ),
            ],
            volumes=[
                V1Volume(
                    name="shared-data",
                    empty_dir=V1EmptyDirVolumeSource(medium="Memory"),
                )
            ],
        ),
    ),
    requests=Resources(
        mem="1G",
    ),
)
def multiple_containers_pod_task() -> str:
    # The code defined in this task will get injected into the primary container.
    while not os.path.isfile(_SHARED_DATA_PATH):
        time.sleep(5)

    with open(_SHARED_DATA_PATH, "r") as shared_message_file:
        return shared_message_file.read()


@workflow
def multiple_containers_pod_workflow() -> str:
    txt = multiple_containers_pod_task()
    return txt
```

+++ {"lines_to_next_cell": 0}

## Pod configuration in a map task

To use a pod task as part of a map task, you can send the pod task definition to the `map_task`.
This will run the pod task across a collection of inputs.

```{code-cell}
@task(
    task_config=Pod(
        pod_spec=V1PodSpec(
            containers=[
                V1Container(
                    name="primary",
                    resources=V1ResourceRequirements(
                        requests={"cpu": ".5", "memory": "500Mi"},
                        limits={"cpu": ".5", "memory": "500Mi"},
                    ),
                )
            ],
            init_containers=[
                V1Container(
                    image="alpine",
                    name="init",
                    command=["/bin/sh"],
                    args=["-c", 'echo "I\'m a customizable init container"'],
                    resources=V1ResourceRequirements(
                        limits={"cpu": ".5", "memory": "500Mi"},
                    ),
                )
            ],
        ),
    )
)
def map_pod_task(int_val: int) -> str:
    return str(int_val)


@task
def coalesce(list_of_strings: List[str]) -> str:
    coalesced = ", ".join(list_of_strings)
    return coalesced


@workflow
def map_pod_workflow(list_of_ints: List[int]) -> str:
    mapped_out = map_task(map_pod_task, metadata=TaskMetadata(retries=1))(int_val=list_of_ints)
    coalesced = coalesce(list_of_strings=mapped_out)
    return coalesced
```

+++ {"lines_to_next_cell": 0}

## Pod configuration in a dynamic workflow

To use a pod task in a dynamic workflow, simply pass the pod task config to the annotated dynamic workflow.

```{code-cell}
@task
def stringify(val: int) -> str:
    return f"{val} served courtesy of a dynamic pod task!"


@dynamic(
    task_config=Pod(
        pod_spec=V1PodSpec(
            containers=[
                V1Container(
                    name="primary",
                    resources=V1ResourceRequirements(
                        requests={"cpu": ".5", "memory": "450Mi"},
                        limits={"cpu": ".5", "memory": "500Mi"},
                    ),
                )
            ],
        ),
    )
)
def dynamic_pod_task(val: int) -> str:
    return stringify(val=val)


@workflow
def dynamic_pod_workflow(val: int = 6) -> str:
    txt = dynamic_pod_task(val=val)
    return txt
```

+++ {"lines_to_next_cell": 0}

You can execute workflows locally as follows:

```{code-cell}
if __name__ == "__main__":
    print(f"Running {__file__}...")
    print(f"Calling pod_workflow()... {pod_workflow()}")
    print(f"Calling multiple_containers_pod_workflow()... {multiple_containers_pod_workflow()}")
    print(f"Calling map_pod_workflow()... {map_pod_workflow()}")
    print(f"Calling dynamic_pod_workflow()... {dynamic_pod_workflow()}")
```
