from dataclasses import dataclass
from typing import Any, Callable, Dict, Optional, Tuple, Union

from flyteidl.core import tasks_pb2 as _core_task

from flytekit import FlyteContext, PythonFunctionTask, lazy_module
from flytekit.configuration import SerializationSettings
from flytekit.exceptions import user as _user_exceptions
from flytekit.extend import Promise, TaskPlugins
from flytekit.loggers import logger
from flytekit.models import task as _task_models

_PRIMARY_CONTAINER_NAME_FIELD = "primary_container_name"
PRIMARY_CONTAINER_DEFAULT_NAME = "primary"


k8s_client = lazy_module("kubernetes.client")
k8s_models = lazy_module("kubernetes.client.models")


def _sanitize_resource_name(resource: _task_models.Resources.ResourceEntry) -> str:
    return _core_task.Resources.ResourceName.Name(resource.name).lower().replace("_", "-")


@dataclass
class Pod(object):
    """
    Pod is a platform-wide configuration that uses pod templates. By default, every task is launched as a container in a pod.
    This plugin helps expose a fully modifiable Kubernetes pod spec to customize the task execution runtime.
    To use pod tasks: (1) Define a pod spec, and (2) Specify the primary container name.

    :param V1PodSpec pod_spec: Kubernetes pod spec. https://kubernetes.io/docs/concepts/workloads/pods
    :param str primary_container_name: the primary container name. If provided the pod-spec can contain a container whose name matches the primary_container_name. This will force Flyte to give up control of the primary
        container and will expect users to control setting up the container. If you expect your python function to run as is, simply create containers that do not match the default primary-container-name and Flyte will auto-inject a
        container for the python function based on the default image provided during serialization.
    :param Optional[Dict[str, str]] labels: Labels are key/value pairs that are attached to pod spec
    :param Optional[Dict[str, str]] annotations: Annotations are key/value pairs that are attached to arbitrary non-identifying metadata to pod spec.
    """

    pod_spec: k8s_models.V1PodSpec
    primary_container_name: str = PRIMARY_CONTAINER_DEFAULT_NAME
    labels: Optional[Dict[str, str]] = None
    annotations: Optional[Dict[str, str]] = None

    def __post_init__(self):
        if not self.pod_spec:
            raise _user_exceptions.FlyteValidationException("A pod spec cannot be undefined")
        if not self.primary_container_name:
            raise _user_exceptions.FlyteValidationException("A primary container name cannot be undefined")


class PodFunctionTask(PythonFunctionTask[Pod]):
    def __init__(self, task_config: Pod, task_function: Callable, **kwargs):
        super(PodFunctionTask, self).__init__(
            task_config=task_config,
            task_type="sidecar",
            task_function=task_function,
            task_type_version=2,
            **kwargs,
        )

    def _serialize_pod_spec(self, settings: SerializationSettings) -> Dict[str, Any]:
        containers = self.task_config.pod_spec.containers
        primary_exists = False
        for container in containers:
            if container.name == self.task_config.primary_container_name:
                primary_exists = True
                break
        if not primary_exists:
            # insert a placeholder primary container if it is not defined in the pod spec.
            containers.append(k8s_models.V1Container(name=self.task_config.primary_container_name))

        final_containers = []
        for container in containers:
            # In the case of the primary container, we overwrite specific container attributes with the default values
            # used in the regular Python task.
            if container.name == self.task_config.primary_container_name:
                sdk_default_container = super().get_container(settings)

                container.image = sdk_default_container.image
                # clear existing commands
                container.command = sdk_default_container.command
                # also clear existing args
                container.args = sdk_default_container.args

                limits, requests = {}, {}
                for resource in sdk_default_container.resources.limits:
                    limits[_sanitize_resource_name(resource)] = resource.value
                for resource in sdk_default_container.resources.requests:
                    requests[_sanitize_resource_name(resource)] = resource.value

                resource_requirements = k8s_models.V1ResourceRequirements(limits=limits, requests=requests)
                if len(limits) > 0 or len(requests) > 0:
                    # Important! Only copy over resource requirements if they are non-empty.
                    container.resources = resource_requirements

                container.env = [
                    k8s_models.V1EnvVar(name=key, value=val) for key, val in sdk_default_container.env.items()
                ] + (container.env or [])

            final_containers.append(container)

        self.task_config.pod_spec.containers = final_containers

        return k8s_client.ApiClient().sanitize_for_serialization(self.task_config.pod_spec)

    def get_k8s_pod(self, settings: SerializationSettings) -> _task_models.K8sPod:
        return _task_models.K8sPod(
            pod_spec=self._serialize_pod_spec(settings),
            metadata=_task_models.K8sObjectMetadata(
                labels=self.task_config.labels,
                annotations=self.task_config.annotations,
            ),
        )

    def get_container(self, settings: SerializationSettings) -> _task_models.Container:
        return None

    def get_config(self, settings: SerializationSettings) -> Dict[str, str]:
        return {_PRIMARY_CONTAINER_NAME_FIELD: self.task_config.primary_container_name}

    def local_execute(self, ctx: FlyteContext, **kwargs) -> Union[Tuple[Promise], Promise, None]:
        logger.warning(
            "Running pod task locally. Local environment may not match pod environment which may cause issues."
        )
        return super().local_execute(ctx=ctx, **kwargs)


TaskPlugins.register_pythontask_plugin(Pod, PodFunctionTask)
