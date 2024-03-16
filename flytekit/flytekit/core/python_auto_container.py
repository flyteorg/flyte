from __future__ import annotations

import importlib
import re
from abc import ABC
from typing import Callable, Dict, List, Optional, TypeVar, Union

from flyteidl.core import tasks_pb2

from flytekit.configuration import ImageConfig, SerializationSettings
from flytekit.core.base_task import PythonTask, TaskMetadata, TaskResolverMixin
from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.pod_template import PodTemplate
from flytekit.core.resources import Resources, ResourceSpec
from flytekit.core.tracked_abc import FlyteTrackedABC
from flytekit.core.tracker import TrackedInstance, extract_task_module
from flytekit.core.utils import _get_container_definition, _serialize_pod_spec, timeit
from flytekit.extras.accelerators import BaseAccelerator
from flytekit.image_spec.image_spec import ImageBuildEngine, ImageSpec
from flytekit.loggers import logger
from flytekit.models import task as _task_model
from flytekit.models.security import Secret, SecurityContext

T = TypeVar("T")
_PRIMARY_CONTAINER_NAME_FIELD = "primary_container_name"


class PythonAutoContainerTask(PythonTask[T], ABC, metaclass=FlyteTrackedABC):
    """
    A Python AutoContainer task should be used as the base for all extensions that want the user's code to be in the
    container and the container information to be automatically captured.
    This base will auto configure the image and image version to be used for all its derivatives.

    If you are looking to extend, you might prefer to use ``PythonFunctionTask`` or ``PythonInstanceTask``
    """

    def __init__(
        self,
        name: str,
        task_config: T,
        task_type="python-task",
        container_image: Optional[Union[str, ImageSpec]] = None,
        requests: Optional[Resources] = None,
        limits: Optional[Resources] = None,
        environment: Optional[Dict[str, str]] = None,
        task_resolver: Optional[TaskResolverMixin] = None,
        secret_requests: Optional[List[Secret]] = None,
        pod_template: Optional[PodTemplate] = None,
        pod_template_name: Optional[str] = None,
        accelerator: Optional[BaseAccelerator] = None,
        **kwargs,
    ):
        """
        :param name: unique name for the task, usually the function's module and name.
        :param task_config: Configuration object for Task. Should be a unique type for that specific Task.
        :param task_type: String task type to be associated with this Task
        :param container_image: String FQN for the image.
        :param requests: custom resource request settings.
        :param limits: custom resource limit settings.
        :param environment: Environment variables you want the task to have when run.
        :param task_resolver: Custom resolver - will pick up the default resolver if empty, or the resolver set
          in the compilation context if one is set.
        :param List[Secret] secret_requests: Secrets that are requested by this container execution. These secrets will
           be mounted based on the configuration in the Secret and available through
           the SecretManager using the name of the secret as the group
           Ideally the secret keys should also be semi-descriptive.
           The key values will be available from runtime, if the backend is configured
           to provide secrets and if secrets are available in the configured secrets store.
           Possible options for secret stores are

           - `Vault <https://www.vaultproject.io/>`__
           - `Confidant <https://lyft.github.io/confidant/>`__
           - `Kube secrets <https://kubernetes.io/docs/concepts/configuration/secret/>`__
           - `AWS Parameter store <https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html>`__
        :param pod_template: Custom PodTemplate for this task.
        :param pod_template_name: The name of the existing PodTemplate resource which will be used in this task.
        :param accelerator: The accelerator to use for this task.
        """
        sec_ctx = None
        if secret_requests:
            for s in secret_requests:
                if not isinstance(s, Secret):
                    raise AssertionError(f"Secret {s} should be of type flytekit.Secret, received {type(s)}")
            sec_ctx = SecurityContext(secrets=secret_requests)

        # pod_template_name overwrites the metadata.pod_template_name
        kwargs["metadata"] = kwargs["metadata"] if "metadata" in kwargs else TaskMetadata()
        kwargs["metadata"].pod_template_name = pod_template_name

        super().__init__(
            task_type=task_type,
            name=name,
            task_config=task_config,
            security_ctx=sec_ctx,
            **kwargs,
        )
        self._container_image = container_image
        # TODO(katrogan): Implement resource overrides
        self._resources = ResourceSpec(
            requests=requests if requests else Resources(), limits=limits if limits else Resources()
        )
        self._environment = environment or {}

        compilation_state = FlyteContextManager.current_context().compilation_state
        if compilation_state and compilation_state.task_resolver:
            if task_resolver:
                logger.info(
                    f"Not using the passed in task resolver {task_resolver} because one found in compilation context"
                )
            self._task_resolver = compilation_state.task_resolver
            if self._task_resolver.task_name(self) is not None:
                self._name = self._task_resolver.task_name(self) or ""
        else:
            self._task_resolver = task_resolver or default_task_resolver
        self._get_command_fn = self.get_default_command

        self.pod_template = pod_template
        self.accelerator = accelerator

    @property
    def task_resolver(self) -> TaskResolverMixin:
        return self._task_resolver

    @property
    def container_image(self) -> Optional[Union[str, ImageSpec]]:
        return self._container_image

    @property
    def resources(self) -> ResourceSpec:
        return self._resources

    def get_default_command(self, settings: SerializationSettings) -> List[str]:
        """
        Returns the default pyflyte-execute command used to run this on hosted Flyte platforms.
        """
        container_args = [
            "pyflyte-execute",
            "--inputs",
            "{{.input}}",
            "--output-prefix",
            "{{.outputPrefix}}",
            "--raw-output-data-prefix",
            "{{.rawOutputDataPrefix}}",
            "--checkpoint-path",
            "{{.checkpointOutputPrefix}}",
            "--prev-checkpoint",
            "{{.prevCheckpointPrefix}}",
            "--resolver",
            self.task_resolver.location,
            "--",
            *self.task_resolver.loader_args(settings, self),
        ]

        return container_args

    def set_command_fn(self, get_command_fn: Optional[Callable[[SerializationSettings], List[str]]] = None):
        """
        By default, the task will run on the Flyte platform using the pyflyte-execute command.
        However, it can be useful to update the command with which the task is serialized for specific cases like
        running map tasks ("pyflyte-map-execute") or for fast-executed tasks.
        """
        self._get_command_fn = get_command_fn  # type: ignore

    def reset_command_fn(self):
        """
        Resets the command which should be used in the container definition of this task to the default arguments.
        This is useful when the command line is overridden at serialization time.
        """
        self._get_command_fn = self.get_default_command

    def get_command(self, settings: SerializationSettings) -> List[str]:
        """
        Returns the command which should be used in the container definition for the serialized version of this task
        registered on a hosted Flyte platform.
        """
        return self._get_command_fn(settings)

    def get_container(self, settings: SerializationSettings) -> _task_model.Container:
        # if pod_template is not None, return None here but in get_k8s_pod, return pod_template merged with container
        if self.pod_template is not None:
            return None
        else:
            return self._get_container(settings)

    def _get_container(self, settings: SerializationSettings) -> _task_model.Container:
        env = {}
        for elem in (settings.env, self.environment):
            if elem:
                env.update(elem)
        if settings.fast_serialization_settings is None or not settings.fast_serialization_settings.enabled:
            if isinstance(self.container_image, ImageSpec):
                self.container_image.source_root = settings.source_root
        return _get_container_definition(
            image=get_registerable_container_image(self.container_image, settings.image_config),
            command=[],
            args=self.get_command(settings=settings),
            data_loading_config=None,
            environment=env,
            ephemeral_storage_request=self.resources.requests.ephemeral_storage,
            cpu_request=self.resources.requests.cpu,
            gpu_request=self.resources.requests.gpu,
            memory_request=self.resources.requests.mem,
            ephemeral_storage_limit=self.resources.limits.ephemeral_storage,
            cpu_limit=self.resources.limits.cpu,
            gpu_limit=self.resources.limits.gpu,
            memory_limit=self.resources.limits.mem,
        )

    def get_k8s_pod(self, settings: SerializationSettings) -> _task_model.K8sPod:
        if self.pod_template is None:
            return None
        return _task_model.K8sPod(
            pod_spec=_serialize_pod_spec(self.pod_template, self._get_container(settings), settings),
            metadata=_task_model.K8sObjectMetadata(
                labels=self.pod_template.labels,
                annotations=self.pod_template.annotations,
            ),
        )

    # need to call super in all its children tasks
    def get_config(self, settings: SerializationSettings) -> Optional[Dict[str, str]]:
        if self.pod_template is None:
            return {}
        return {_PRIMARY_CONTAINER_NAME_FIELD: self.pod_template.primary_container_name}

    def get_extended_resources(self, settings: SerializationSettings) -> Optional[tasks_pb2.ExtendedResources]:
        """
        Returns the extended resources to allocate to the task on hosted Flyte.
        """
        if self.accelerator is None:
            return None

        return tasks_pb2.ExtendedResources(gpu_accelerator=self.accelerator.to_flyte_idl())


class DefaultTaskResolver(TrackedInstance, TaskResolverMixin):
    """
    Please see the notes in the TaskResolverMixin as it describes this default behavior.
    """

    def name(self) -> str:
        return "DefaultTaskResolver"

    @timeit("Load task")
    def load_task(self, loader_args: List[str]) -> PythonAutoContainerTask:
        _, task_module, _, task_name, *_ = loader_args

        task_module = importlib.import_module(name=task_module)  # type: ignore
        task_def = getattr(task_module, task_name)
        return task_def

    def loader_args(self, settings: SerializationSettings, task: PythonAutoContainerTask) -> List[str]:  # type:ignore
        _, m, t, _ = extract_task_module(task)
        return ["task-module", m, "task-name", t]

    def get_all_tasks(self) -> List[PythonAutoContainerTask]:  # type: ignore
        raise Exception("should not be needed")


default_task_resolver = DefaultTaskResolver()


def get_registerable_container_image(img: Optional[Union[str, ImageSpec]], cfg: ImageConfig) -> str:
    """
    Resolve the image to the real image name that should be used for registration.
    1. If img is a ImageSpec, it will be built and the image name will be returned
    2. If img is a placeholder string (e.g. {{.image.default.fqn}}:{{.image.default.version}}),
        it will be resolved using the cfg and the image name will be returned

    :param img: Configured image or image spec
    :param cfg: Registration configuration
    :return:
    """
    if isinstance(img, ImageSpec):
        ImageBuildEngine.build(img)
        return img.image_name()

    if img is not None and img != "":
        matches = _IMAGE_REPLACE_REGEX.findall(img)
        if matches is None or len(matches) == 0:
            return img
        for m in matches:
            if len(m) < 3:
                raise AssertionError(
                    "Image specification should be of the form <fqn>:<tag> OR <fqn>:{{.image.default.version}} OR "
                    f"{{.image.xyz.fqn}}:{{.image.xyz.version}} OR {{.image.xyz}} - Received {m}"
                )
            replace_group, name, attr = m
            if name is None or name == "":
                raise AssertionError(f"Image format is incorrect {m}")
            img_cfg = cfg.find_image(name)
            if img_cfg is None:
                raise AssertionError(f"Image Config with name {name} not found in the configuration")
            if attr == "version":
                if img_cfg.tag is not None:
                    img = img.replace(replace_group, img_cfg.tag)
                else:
                    img = img.replace(replace_group, cfg.default_image.tag)
            elif attr == "fqn":
                img = img.replace(replace_group, img_cfg.fqn)
            elif attr == "":
                img = img.replace(replace_group, img_cfg.full)
            else:
                raise AssertionError(f"Only fqn and version are supported replacements, {attr} is not supported")
        return img
    if cfg.default_image is None:
        raise ValueError("An image is required for PythonAutoContainer tasks")
    return f"{cfg.default_image.fqn}:{cfg.default_image.tag}"


# Matches {{.image.<name>.<attr>}}. A name can be either 'default' indicating the default image passed during
# serialization or it can be a custom name for an image that must be defined in the config section Images. An attribute
# can be either 'fqn', 'version' or non-existent.
# fqn will access the fully qualified name of the image (e.g. registry/imagename:version -> registry/imagename)
# version will access the version part of the image (e.g. registry/imagename:version -> version)
# With empty attribute, it'll access the full image path (e.g. registry/imagename:version -> registry/imagename:version)
_IMAGE_REPLACE_REGEX = re.compile(r"({{\s*\.image[s]?(?:\.([a-zA-Z0-9_]+))(?:\.([a-zA-Z0-9_]+))?\s*}})", re.IGNORECASE)
