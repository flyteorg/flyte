"""
Pod Example
-----------

Pod tasks can be used whenever multiple containers need to spin up within a single task.
They expose a fully modifiable Kubernetes `pod spec <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#podspec-v1-core>`__ which can be used to customize the task execution runtime.

All we need to do to use pod tasks is:

1. Define a pod spec
2. Specify the name of the primary container

`The primary container is the driver for Flyte task execution, for example, producing inputs and outputs.`

Pod tasks accept arguments that ordinary container tasks usually accept, such as resource specifications, etc.
However, these are only applied to the primary container.
To customize other containers brought up during the execution, we can define a full-fledged pod spec.
This is done using the `Kubernetes Python client library <https://github.com/kubernetes-client/python>`__,
specifically with the
`V1PodSpec <https://github.com/kubernetes-client/python/blob/master/kubernetes/client/models/v1_pod_spec.py>`__.
"""

# %%
# Simple Pod Task
# ===============
#
# In this example, let's define a simple pod spec in which a shared volume is mounted in both primary and secondary containers.
# The secondary writes a file that the primary waits on before completing.
#
# First, we import the necessary libraries and define the shared data path.
import os
import time
from typing import List

from flytekit import task, workflow
from flytekitplugins.pod import Pod
from kubernetes.client.models import (
    V1Container,
    V1EmptyDirVolumeSource,
    V1PodSpec,
    V1ResourceRequirements,
    V1Volume,
    V1VolumeMount,
)

_SHARED_DATA_PATH = "/data/message.txt"

# %%
# We define a simple pod spec with two containers.
def generate_pod_spec_for_task():

    # primary containers do not require us to specify an image, the default image built for Flyte tasks will get used
    primary_container = V1Container(name="primary")

    # NOTE: for non-primary containers, we must specify the image
    secondary_container = V1Container(
        name="secondary",
        image="alpine",
    )
    secondary_container.command = ["/bin/sh"]
    secondary_container.args = [
        "-c",
        "echo hi pod world > {}".format(_SHARED_DATA_PATH),
    ]

    resources = V1ResourceRequirements(
        requests={"cpu": "1", "memory": "100Mi"}, limits={"cpu": "1", "memory": "100Mi"}
    )
    primary_container.resources = resources
    secondary_container.resources = resources

    shared_volume_mount = V1VolumeMount(
        name="shared-data",
        mount_path="/data",
    )
    secondary_container.volume_mounts = [shared_volume_mount]
    primary_container.volume_mounts = [shared_volume_mount]

    pod_spec = V1PodSpec(
        containers=[primary_container, secondary_container],
        volumes=[
            V1Volume(
                name="shared-data", empty_dir=V1EmptyDirVolumeSource(medium="Memory")
            )
        ],
    )

    return pod_spec


# %%
# Although pod tasks for the most part allow us to customize Kubernetes container attributes, we can still use Flyte directives to specify resources and even the image.
# The default image built for Flyte tasks will be used unless ``container_image`` task attribute is specified.
@task(
    task_config=Pod(
        pod_spec=generate_pod_spec_for_task(), primary_container_name="primary"
    ),
)
def my_pod_task() -> str:
    # the code defined in this task will get injected into the primary container.
    while not os.path.isfile(_SHARED_DATA_PATH):
        time.sleep(5)

    with open(_SHARED_DATA_PATH, "r") as shared_message_file:
        return shared_message_file.read()


@workflow
def PodWorkflow() -> str:
    s = my_pod_task()
    return s


# %%
# Pod Task in Map Task
# ====================
#
# To use pod task as part of map task, we send pod task definition to :py:func:`~flytekit:flytekit.map_task`.
# This will run pod task across a collection of inputs.
from flytekit import map_task, TaskMetadata


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
                    name="init",
                    command=["/bin/sh"],
                    args=["-c", 'echo "I\'m a customizable init container"'],
                )
            ],
        ),
        primary_container_name="primary",
    )
)
def my_pod_map_task(stringify: int) -> str:
    return str(stringify)


@task
def coalesce(b: List[str]) -> str:
    coalesced = ", ".join(b)
    return coalesced


@workflow
def my_map_workflow(a: List[int]) -> str:
    mapped_out = map_task(my_pod_map_task, metadata=TaskMetadata(retries=1))(
        stringify=a
    )
    coalesced = coalesce(b=mapped_out)
    return coalesced

# %%
# Since pod tasks cannot be run locally, we use the ``pass`` keyword to skip running the tasks.
if __name__ == "__main__":
    pass
