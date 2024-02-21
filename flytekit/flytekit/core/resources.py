from dataclasses import dataclass
from typing import List, Optional

from flytekit.models import task as task_models


@dataclass
class Resources(object):
    """
    This class is used to specify both resource requests and resource limits.

    .. code-block:: python

        Resources(cpu="1", mem="2048")  # This is 1 CPU and 2 KB of memory
        Resources(cpu="100m", mem="2Gi")  # This is 1/10th of a CPU and 2 gigabytes of memory

        # For Kubernetes-based tasks, pods use ephemeral local storage for scratch space, caching, and for logs.
        # This allocates 1Gi of such local storage.
        Resources(ephemeral_storage="1Gi")

    .. note::

        Persistent storage is not currently supported on the Flyte backend.

    Please see the :std:ref:`User Guide <cookbook:customizing task resources>` for detailed examples.
    Also refer to the `K8s conventions. <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes>`__
    """

    cpu: Optional[str] = None
    mem: Optional[str] = None
    gpu: Optional[str] = None
    ephemeral_storage: Optional[str] = None

    def __post_init__(self):
        def _check_none_or_str(value):
            if value is None:
                return
            if not isinstance(value, str):
                raise AssertionError(f"{value} should be a string")

        _check_none_or_str(self.cpu)
        _check_none_or_str(self.mem)
        _check_none_or_str(self.gpu)
        _check_none_or_str(self.ephemeral_storage)


@dataclass
class ResourceSpec(object):
    requests: Resources
    limits: Resources


_ResourceName = task_models.Resources.ResourceName
_ResourceEntry = task_models.Resources.ResourceEntry


def _convert_resources_to_resource_entries(resources: Resources) -> List[_ResourceEntry]:  # type: ignore
    resource_entries = []
    if resources.cpu is not None:
        resource_entries.append(_ResourceEntry(name=_ResourceName.CPU, value=resources.cpu))
    if resources.mem is not None:
        resource_entries.append(_ResourceEntry(name=_ResourceName.MEMORY, value=resources.mem))
    if resources.gpu is not None:
        resource_entries.append(_ResourceEntry(name=_ResourceName.GPU, value=resources.gpu))
    if resources.ephemeral_storage is not None:
        resource_entries.append(_ResourceEntry(name=_ResourceName.EPHEMERAL_STORAGE, value=resources.ephemeral_storage))
    return resource_entries


def convert_resources_to_resource_model(
    requests: Optional[Resources] = None,
    limits: Optional[Resources] = None,
) -> task_models.Resources:
    """
    Convert flytekit ``Resources`` objects to a Resources model

    :param requests: Resource requests. Optional, defaults to ``None``
    :param limits: Resource limits. Optional, defaults to ``None``
    :return: The given resources as requests and limits
    """
    request_entries = []
    limit_entries = []
    if requests is not None:
        request_entries = _convert_resources_to_resource_entries(requests)
    if limits is not None:
        limit_entries = _convert_resources_to_resource_entries(limits)
    return task_models.Resources(requests=request_entries, limits=limit_entries)
