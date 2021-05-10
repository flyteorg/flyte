"""
Pod plugin example
--------------------------

Pod tasks can be used anytime you need to bring up multiple containers within a single task. They expose a fully
modifable kubernetes `pod spec
<https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#podspec-v1-core>`_ you can use to customize
your task execution runtime.

All you need to use Pod tasks are to 1) define a pod spec and 2) specify the primary container name
The primary container is the driver for the flyte task execution for example, producing inputs and outputs.
"""


# %%
# Pod tasks accept all the same arguments that ordinary container tasks accept, such as resource specifications.
# However, these are only applied to the primary container. To customize other containers brought up during execution
# we define a fully-fledged pod spec. This is done using the
# `kubernetes python client library <https://github.com/kubernetes-client/python>`__
# specifically with the
# `V1PodSpec <https://github.com/kubernetes-client/python/blob/master/kubernetes/client/models/v1_pod_spec.py>`__.
#
# In this example, we define a simple pod spec in which a shared volume is mounted in both the primary and secondary
# containers. The secondary writes a file that the primary waits on before completing.
import os
import time

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


def generate_pod_spec_for_task():

    # Primary containers do not require us to specify an image, the default image built for flyte tasks will get used.
    primary_container = V1Container(name="primary")

    # Note: for non-primary containers we must specify an image.
    secondary_container = V1Container(name="secondary", image="alpine",)
    secondary_container.command = ["/bin/sh"]
    secondary_container.args = ["-c", "echo hi pod world > {}".format(_SHARED_DATA_PATH)]

    resources = V1ResourceRequirements(
        requests={"cpu": "1", "memory": "100Mi"}, limits={"cpu": "1", "memory": "100Mi"}
    )
    primary_container.resources = resources
    secondary_container.resources = resources

    shared_volume_mount = V1VolumeMount(name="shared-data", mount_path="/data",)
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
# Although Pod tasks for the most part allow you to customize kubernetes container attributes
# you can still use flyte directives to specify resources and even the image. The default image built for
# flyte tasks will get used unless you specify the `container_image` task attribute.
@task(
    task_config=Pod(
        pod_spec=generate_pod_spec_for_task(), primary_container_name="primary"
    ),
)
def my_pod_task() -> str:
    # The code defined in this task will get injected into the primary container.
    while not os.path.isfile(_SHARED_DATA_PATH):
        time.sleep(5)

    with open(_SHARED_DATA_PATH, "r") as shared_message_file:
        return shared_message_file.read()


@workflow
def PodWorkflow() -> str:
    s = my_pod_task()
    return s


if __name__ == "__main__":
    pass
