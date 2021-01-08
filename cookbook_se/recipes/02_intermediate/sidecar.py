"""
10: Sidecar plugin example
--------------------------

Sidecar tasks can be used anytime you need to bring up multiple containers within a single task. They expose a fully
modifable kubernetes `pod spec
<https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#podspec-v1-core>`_ you can use to customize
your task execution runtime.

All you need to use sidecar tasks are to 1) define a pod spec and 2) specify the primary container name
The primary container is the driver for the flyte task execution for example, producing inputs and outputs.
"""


# %%
# Sidecar tasks accept all the same arguments that ordinary container tasks accept, such as resource specifications.
# However, these are only applied to the primary container. To customize other containers brought up during execution
# we define a fully-fledged pod spec.
#
# In this example, we define a simple pod spec in which a shared volume is mounted in both the primary and secondary
# containers. The secondary writes a file that the primary waits on before completing.
import os
import time

from flytekit import task, workflow
from flytekit.taskplugins.pod import Pod
from k8s.io.api.core.v1 import generated_pb2
from k8s.io.apimachinery.pkg.api.resource.generated_pb2 import Quantity

_SHARED_DATA_PATH = "/data/message.txt"


def generate_pod_spec_for_task():
    pod_spec = generated_pb2.PodSpec()

    # Primary containers do not require us to specify an image, the default image built for flyte tasks will get used.
    primary_container = generated_pb2.Container(name="primary")

    # Note: for non-primary containers we must specify an image.
    secondary_container = generated_pb2.Container(name="secondary", image="alpine",)
    secondary_container.command.extend(["/bin/sh"])
    secondary_container.args.extend(
        ["-c", "echo hi sidecar world > {}".format(_SHARED_DATA_PATH)]
    )

    resources = generated_pb2.ResourceRequirements()
    resources.limits["cpu"].CopyFrom(Quantity(string="1"))
    resources.requests["cpu"].CopyFrom(Quantity(string="1"))
    resources.limits["memory"].CopyFrom(Quantity(string="100Mi"))
    resources.requests["memory"].CopyFrom(Quantity(string="100Mi"))
    primary_container.resources.CopyFrom(resources)
    secondary_container.resources.CopyFrom(resources)

    shared_volume_mount = generated_pb2.VolumeMount(
        name="shared-data", mountPath="/data",
    )
    secondary_container.volumeMounts.extend([shared_volume_mount])
    primary_container.volumeMounts.extend([shared_volume_mount])

    pod_spec.volumes.extend(
        [
            generated_pb2.Volume(
                name="shared-data",
                volumeSource=generated_pb2.VolumeSource(
                    emptyDir=generated_pb2.EmptyDirVolumeSource(medium="Memory",)
                ),
            )
        ]
    )
    pod_spec.containers.extend([primary_container, secondary_container])
    return pod_spec


# %%
# Although sidecar tasks for the most part allow you to customize kubernetes container attributes
# you can still use flyte directives to specify resources and even the image. The default image built for
# flyte tasks will get used unless you specify the `container_image` task attribute.
@task(
    task_config=Pod(
        pod_spec=generate_pod_spec_for_task(), primary_container_name="primary"
    )
)
def my_sidecar_task() -> str:
    # The code defined in this task will get injected into the primary container.
    while not os.path.isfile(_SHARED_DATA_PATH):
        time.sleep(5)

    with open(_SHARED_DATA_PATH, "r") as shared_message_file:
        return shared_message_file.read()


@workflow
def SidecarWorkflow() -> str:
    s = my_sidecar_task()
    return s


if __name__ == "__main__":
    pass
