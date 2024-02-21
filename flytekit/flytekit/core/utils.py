import datetime
import os as _os
import shutil as _shutil
import tempfile as _tempfile
import time as _time
from abc import ABC, abstractmethod
from functools import wraps
from hashlib import sha224 as _sha224
from pathlib import Path
from typing import TYPE_CHECKING, Any, Callable, Dict, List, Optional, cast

from flyteidl.core import tasks_pb2 as _core_task

from flytekit.configuration import SerializationSettings
from flytekit.core.pod_template import PodTemplate
from flytekit.loggers import logger

if TYPE_CHECKING:
    from flytekit.models import task as task_models


def _dnsify(value: str) -> str:
    """
    Converts value into a DNS-compliant (RFC1035/RFC1123 DNS_LABEL). The resulting string must only consist of
    alphanumeric (lower-case a-z, and 0-9) and not exceed 63 characters. It's permitted to have '-' character as long
    as it's not in the first or last positions.

    :param Text value:
    :rtype: Text
    """
    res = ""
    MAX = 63
    HASH_LEN = 10
    if len(value) >= MAX:
        h = _sha224(value.encode("utf-8")).hexdigest()[:HASH_LEN]
        value = "{}-{}".format(h, value[-(MAX - HASH_LEN - 1) :])
    for ch in value:
        if ch == "_" or ch == "-" or ch == ".":
            # Convert '_' to '-' unless it's the first character, in which case we drop it.
            if res != "" and len(res) < 62:
                res += "-"
        elif not ch.isalnum():
            # Trim non-alphanumeric letters.
            pass
        elif ch.islower() or ch.isdigit():
            # Character is already compliant, just append it.
            res += ch
        else:
            # Character is upper-case. Add a '-' before it for better readability.
            if res != "" and res[-1] != "-" and len(res) < 62:
                res += "-"
            res += ch.lower()

    if len(res) > 0 and res[-1] == "-":
        res = res[: len(res) - 1]

    return res


def _get_container_definition(
    image: str,
    command: List[str],
    args: Optional[List[str]] = None,
    data_loading_config: Optional["task_models.DataLoadingConfig"] = None,
    ephemeral_storage_request: Optional[str] = None,
    cpu_request: Optional[str] = None,
    gpu_request: Optional[str] = None,
    memory_request: Optional[str] = None,
    ephemeral_storage_limit: Optional[str] = None,
    cpu_limit: Optional[str] = None,
    gpu_limit: Optional[str] = None,
    memory_limit: Optional[str] = None,
    environment: Optional[Dict[str, str]] = None,
) -> "task_models.Container":
    ephemeral_storage_limit = ephemeral_storage_limit
    ephemeral_storage_request = ephemeral_storage_request
    cpu_limit = cpu_limit
    cpu_request = cpu_request
    gpu_limit = gpu_limit
    gpu_request = gpu_request
    memory_limit = memory_limit
    memory_request = memory_request

    from flytekit.models import task as task_models

    # TODO: Use convert_resources_to_resource_model instead of manually fixing the resources.
    requests = []
    if ephemeral_storage_request:
        requests.append(
            task_models.Resources.ResourceEntry(
                task_models.Resources.ResourceName.EPHEMERAL_STORAGE,
                ephemeral_storage_request,
            )
        )
    if cpu_request:
        requests.append(task_models.Resources.ResourceEntry(task_models.Resources.ResourceName.CPU, cpu_request))
    if gpu_request:
        requests.append(task_models.Resources.ResourceEntry(task_models.Resources.ResourceName.GPU, gpu_request))
    if memory_request:
        requests.append(task_models.Resources.ResourceEntry(task_models.Resources.ResourceName.MEMORY, memory_request))

    limits = []
    if ephemeral_storage_limit:
        limits.append(
            task_models.Resources.ResourceEntry(
                task_models.Resources.ResourceName.EPHEMERAL_STORAGE,
                ephemeral_storage_limit,
            )
        )
    if cpu_limit:
        limits.append(task_models.Resources.ResourceEntry(task_models.Resources.ResourceName.CPU, cpu_limit))
    if gpu_limit:
        limits.append(task_models.Resources.ResourceEntry(task_models.Resources.ResourceName.GPU, gpu_limit))
    if memory_limit:
        limits.append(task_models.Resources.ResourceEntry(task_models.Resources.ResourceName.MEMORY, memory_limit))

    if environment is None:
        environment = {}

    return task_models.Container(
        image=image,
        command=command,
        args=args,
        resources=task_models.Resources(limits=limits, requests=requests),
        env=environment,
        config={},
        data_loading_config=data_loading_config,
    )


def _sanitize_resource_name(resource: "task_models.Resources.ResourceEntry") -> str:
    return _core_task.Resources.ResourceName.Name(resource.name).lower().replace("_", "-")


def _serialize_pod_spec(
    pod_template: "PodTemplate",
    primary_container: "task_models.Container",
    settings: SerializationSettings,
) -> Dict[str, Any]:
    # import here to avoid circular import
    from kubernetes.client import ApiClient, V1PodSpec
    from kubernetes.client.models import V1Container, V1EnvVar, V1ResourceRequirements

    from flytekit.core.python_auto_container import get_registerable_container_image

    if pod_template.pod_spec is None:
        return {}
    containers = cast(V1PodSpec, pod_template.pod_spec).containers
    primary_exists = False

    for container in containers:
        if container.name == cast(PodTemplate, pod_template).primary_container_name:
            primary_exists = True
            break

    if not primary_exists:
        # insert a placeholder primary container if it is not defined in the pod spec.
        containers.append(V1Container(name=cast(PodTemplate, pod_template).primary_container_name))
    final_containers = []

    for container in containers:
        # In the case of the primary container, we overwrite specific container attributes
        # with the values given to ContainerTask.
        # The attributes include: image, command, args, resource, and env (env is unioned)

        # resolve the image name if it is image spec or placeholder
        resolved_image = get_registerable_container_image(container.image, settings.image_config)

        if container.name == cast(PodTemplate, pod_template).primary_container_name:
            if container.image is None:
                # Copy the image from primary_container only if the image is not specified in the pod spec.
                container.image = primary_container.image
            else:
                container.image = resolved_image

            container.command = primary_container.command
            container.args = primary_container.args

            limits, requests = {}, {}
            for resource in primary_container.resources.limits:
                limits[_sanitize_resource_name(resource)] = resource.value
            for resource in primary_container.resources.requests:
                requests[_sanitize_resource_name(resource)] = resource.value
            resource_requirements = V1ResourceRequirements(limits=limits, requests=requests)
            if len(limits) > 0 or len(requests) > 0:
                # Important! Only copy over resource requirements if they are non-empty.
                container.resources = resource_requirements
            if primary_container.env is not None:
                container.env = [V1EnvVar(name=key, value=val) for key, val in primary_container.env.items()] + (
                    container.env or []
                )
        else:
            container.image = resolved_image

        final_containers.append(container)
    cast(V1PodSpec, pod_template.pod_spec).containers = final_containers

    return ApiClient().sanitize_for_serialization(cast(PodTemplate, pod_template).pod_spec)


def load_proto_from_file(pb2_type, path):
    with open(path, "rb") as reader:
        out = pb2_type()
        out.ParseFromString(reader.read())
        return out


def write_proto_to_file(proto, path):
    Path(_os.path.dirname(path)).mkdir(parents=True, exist_ok=True)
    with open(path, "wb") as writer:
        writer.write(proto.SerializeToString())


class Directory(object):
    def __init__(self, path):
        """
        :param Text path: local path of directory
        """
        self._name = path

    @property
    def name(self):
        """
        :rtype: Text
        """
        return self._name

    def list_dir(self):
        """
        The list of absolute filepaths for all immediate sub-paths
        :rtype: list[Text]
        """
        return [_os.path.join(self.name, f) for f in _os.listdir(self.name)]

    def __enter__(self):
        pass

    def __exit__(self, exc_type, exc_val, exc_tb):
        pass


class AutoDeletingTempDir(Directory):
    """
    Creates a posix safe tempdir which is auto deleted once out of scope
    """

    def __init__(self, working_dir_prefix=None, tmp_dir=None, cleanup=True):
        """
        :param Text working_dir_prefix: A prefix to help identify temporary directories
        :param Text tmp_dir: Path to desired temporary directory
        :param bool cleanup: Whether the directory should be cleaned up upon exit
        """
        self._tmp_dir = tmp_dir
        self._working_dir_prefix = (working_dir_prefix + "_") if working_dir_prefix else ""
        self._cleanup = cleanup
        super(AutoDeletingTempDir, self).__init__(None)

    def __enter__(self):
        self._name = _tempfile.mkdtemp(dir=self._tmp_dir, prefix=self._working_dir_prefix)
        return self

    def get_named_tempfile(self, name):
        return _os.path.join(self.name, name)

    def _cleanup_dir(self):
        if self.name and self._cleanup:
            if _os.path.exists(self.name):
                _shutil.rmtree(self.name)
            self._name = None

    def force_cleanup(self):
        self._cleanup_dir()

    def __exit__(self, exc_type, exc_val, exc_tb):
        self._cleanup_dir()

    def __repr__(self):
        return "Auto-Deleting Tmp Directory @ {}".format(self.name)

    def __str__(self):
        return self.__repr__()


class timeit:
    """
    A context manager and a decorator that measures the execution time of the wrapped code block or functions.
    It will append a timing information to TimeLineDeck. For instance:

    @timeit("Function description")
    def function()

    with timeit("Wrapped code block description"):
        # your code
    """

    def __init__(self, name: str = ""):
        """
        :param name: A string that describes the wrapped code block or function being executed.
        """
        self._name = name
        self.start_time = None
        self._start_wall_time = None
        self._start_process_time = None

    def __call__(self, func: Callable):
        @wraps(func)
        def wrapper(*args, **kwargs):
            with self:
                return func(*args, **kwargs)

        return wrapper

    def __enter__(self):
        self.start_time = datetime.datetime.utcnow()
        self._start_wall_time = _time.perf_counter()
        self._start_process_time = _time.process_time()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """
        The exception, if any, will propagate outside the context manager, as the purpose of this context manager
        is solely to measure the execution time of the wrapped code block.
        """
        from flytekit.core.context_manager import FlyteContextManager

        end_time = datetime.datetime.utcnow()
        end_wall_time = _time.perf_counter()
        end_process_time = _time.process_time()

        timeline_deck = FlyteContextManager.current_context().user_space_params.timeline_deck
        timeline_deck.append_time_info(
            dict(
                Name=self._name,
                Start=self.start_time,
                Finish=end_time,
                WallTime=end_wall_time - self._start_wall_time,
                ProcessTime=end_process_time - self._start_process_time,
            )
        )

        logger.info(
            "{}. [Wall Time: {}s, Process Time: {}s]".format(
                self._name,
                end_wall_time - self._start_wall_time,
                end_process_time - self._start_process_time,
            )
        )


class ClassDecorator(ABC):
    """
    Abstract class for class decorators.
    We can attach config on the decorator class and use it in the upper level.
    """

    LINK_TYPE_KEY = "link_type"
    PORT_KEY = "port"

    def __init__(self, task_function=None, **kwargs):
        """
        If the decorator is called with arguments, func will be None.
        If the decorator is called without arguments, func will be function to be decorated.
        """
        self.task_function = task_function
        self.decorator_kwargs = kwargs
        if task_function:
            # wraps preserve the function metadata, including type annotations, from the original function to the decorator.
            wraps(task_function)(self)

    def __call__(self, *args, **kwargs):
        if self.task_function:
            # Where the actual execution happens.
            return self.execute(*args, **kwargs)
        else:
            # If self.func is None, it means decorator was called with arguments.
            # Therefore, __call__ received the actual function to be decorated.
            # We return a new instance of ClassDecorator with the function and stored arguments.
            return self.__class__(args[0], **self.decorator_kwargs)

    @abstractmethod
    def execute(self, *args, **kwargs):
        """
        This method will be called when the decorated function is called.
        """
        pass

    @abstractmethod
    def get_extra_config(self):
        """
        Get the config of the decorator.
        """
        pass
