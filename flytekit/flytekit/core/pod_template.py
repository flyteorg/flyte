from dataclasses import dataclass
from typing import TYPE_CHECKING, Dict, Optional

from flytekit.exceptions import user as _user_exceptions

if TYPE_CHECKING:
    from kubernetes.client import V1PodSpec

PRIMARY_CONTAINER_DEFAULT_NAME = "primary"


@dataclass(init=True, repr=True, eq=True, frozen=False)
class PodTemplate(object):
    """Custom PodTemplate specification for a Task."""

    pod_spec: Optional["V1PodSpec"] = None
    primary_container_name: str = PRIMARY_CONTAINER_DEFAULT_NAME
    labels: Optional[Dict[str, str]] = None
    annotations: Optional[Dict[str, str]] = None

    def __post_init__(self):
        if self.pod_spec is None:
            from kubernetes.client import V1PodSpec

            self.pod_spec = V1PodSpec(containers=[])
        if not self.primary_container_name:
            raise _user_exceptions.FlyteValidationException("A primary container name cannot be undefined")
