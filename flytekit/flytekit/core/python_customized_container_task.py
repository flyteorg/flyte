from __future__ import annotations

import os
from typing import Any, Dict, List, Optional, Type, TypeVar

from flyteidl.core import tasks_pb2 as _tasks_pb2

from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.base_task import PythonTask, Task, TaskResolverMixin
from flytekit.core.context_manager import FlyteContext
from flytekit.core.resources import Resources, ResourceSpec
from flytekit.core.shim_task import ExecutableTemplateShimTask, ShimTaskExecutor
from flytekit.core.tracker import TrackedInstance
from flytekit.core.utils import _get_container_definition, load_proto_from_file
from flytekit.loggers import logger
from flytekit.models import task as _task_model
from flytekit.models.core import identifier as identifier_models
from flytekit.models.security import Secret, SecurityContext
from flytekit.tools.module_loader import load_object_from_module

TC = TypeVar("TC")


class PythonCustomizedContainerTask(ExecutableTemplateShimTask, PythonTask[TC]):  # type: ignore
    """
    Please take a look at the comments for :py:class`flytekit.extend.ExecutableTemplateShimTask` as well. This class
    should be subclassed and a custom Executor provided as a default to this parent class constructor
    when building a new external-container flytekit-only plugin.

    This class provides authors of new task types the basic scaffolding to create task-template based tasks. In order
    to write such a task, authors need to

    * subclass the ``ShimTaskExecutor`` class  and override the ``execute_from_model`` function. This function is
      where all the business logic should go. Keep in mind though that you, the plugin author, will not have access
      to anything that's not serialized within the ``TaskTemplate`` which is why you'll also need to
    * subclass this class, and override the ``get_custom`` function to include all the information the executor
      will need to run.
    * Also pass the executor you created as the ``executor_type`` argument of this class's constructor.

    Keep in mind that the total size of the ``TaskTemplate`` still needs to be small, since these will be accessed
    frequently by the Flyte engine.
    """

    SERIALIZE_SETTINGS = SerializationSettings(
        project="PLACEHOLDER_PROJECT",
        domain="LOCAL",
        version="PLACEHOLDER_VERSION",
        env=None,
        image_config=ImageConfig(
            default_image=Image(name="custom_container_task", fqn="flyteorg.io/placeholder", tag="image")
        ),
    )

    def __init__(
        self,
        name: str,
        task_config: TC,
        container_image: str,
        executor_type: Type[ShimTaskExecutor],
        task_resolver: Optional[TaskTemplateResolver] = None,
        task_type="python-task",
        requests: Optional[Resources] = None,
        limits: Optional[Resources] = None,
        environment: Optional[Dict[str, str]] = None,
        secret_requests: Optional[List[Secret]] = None,
        **kwargs,
    ):
        """
        :param name: unique name for the task, usually the function's module and name.
        :param task_config: Configuration object for Task. Should be a unique type for that specific Task
        :param container_image: This is the external container image the task should run at platform-run-time.
        :param executor: This is an executor which will actually provide the business logic.
        :param task_resolver: Custom resolver - if you don't make one, use the default task template resolver.
        :param task_type: String task type to be associated with this Task.
        :param requests: custom resource request settings.
        :param limits: custom resource limit settings.
        :param environment: Environment variables you want the task to have when run.
        :param List[Secret] secret_requests: Secrets that are requested by this container execution. These secrets will
           be mounted based on the configuration in the Secret and available through
           the SecretManager using the name of the secret as the group
           Ideally the secret keys should also be semi-descriptive.
           The key values will be available from runtime, if the backend is configured to provide secrets and
           if secrets are available in the configured secrets store. Possible options for secret stores are

           - `Vault <https://www.vaultproject.io/>`__
           - `Confidant <https://lyft.github.io/confidant/>`__
           - `Kube secrets <https://kubernetes.io/docs/concepts/configuration/secret/>`__
           - `AWS Parameter store <https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html>`__
        """
        sec_ctx = None
        if secret_requests:
            for s in secret_requests:
                if not isinstance(s, Secret):
                    raise AssertionError(f"Secret {s} should be of type flytekit.Secret, received {type(s)}")
            sec_ctx = SecurityContext(secrets=secret_requests)
        super().__init__(
            tt=None,
            executor_type=executor_type,
            task_type=task_type,
            name=name,
            task_config=task_config,
            security_ctx=sec_ctx,
            **kwargs,
        )
        self._resources = ResourceSpec(
            requests=requests if requests else Resources(), limits=limits if limits else Resources()
        )
        self._environment = environment or {}
        self._container_image = container_image
        self._task_resolver = task_resolver or default_task_template_resolver

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        # Overriding base implementation to raise an error, force task author to implement
        raise NotImplementedError

    def get_config(self, settings: SerializationSettings) -> Dict[str, str]:
        # Overriding base implementation but not doing anything. Technically this should be the task config,
        # but the IDL limitation that the value also has to be a string is very limiting.
        # Recommend putting information you need in the config into custom instead, because when serializing
        # the custom field, we jsonify custom and the place it into a protobuf struct. This config field
        # just gets put into a Dict[str, str]
        return {}

    @property
    def resources(self) -> ResourceSpec:
        return self._resources

    @property
    def task_resolver(self) -> TaskTemplateResolver:
        return self._task_resolver

    @property
    def task_template(self) -> Optional[_task_model.TaskTemplate]:
        """
        Override the base class implementation to serialize on first call.
        """
        return self._task_template or self.serialize_to_model(settings=PythonCustomizedContainerTask.SERIALIZE_SETTINGS)

    @property
    def container_image(self) -> str:
        return self._container_image

    def get_command(self, settings: SerializationSettings) -> List[str]:
        container_args = [
            "pyflyte-execute",
            "--inputs",
            "{{.input}}",
            "--output-prefix",
            "{{.outputPrefix}}",
            "--raw-output-data-prefix",
            "{{.rawOutputDataPrefix}}",
            "--resolver",
            self.task_resolver.location,
            "--",
            *self.task_resolver.loader_args(settings, self),
        ]

        return container_args

    def get_container(self, settings: SerializationSettings) -> _task_model.Container:
        env = {**settings.env, **self.environment} if self.environment else settings.env
        return _get_container_definition(
            image=self.container_image,
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

    def serialize_to_model(self, settings: SerializationSettings) -> _task_model.TaskTemplate:
        # This doesn't get called from translator unfortunately. Will need to move the translator to use the model
        # objects directly first.
        # Note: This doesn't settle the issue of duplicate registrations. We'll need to figure that out somehow.
        # TODO: After new control plane classes are in, promote the template to a FlyteTask, so that authors of
        #  customized-container tasks have a familiar thing to work with.
        obj = _task_model.TaskTemplate(
            identifier_models.Identifier(
                identifier_models.ResourceType.TASK, settings.project, settings.domain, self.name, settings.version
            ),
            self.task_type,
            self.metadata.to_taskmetadata_model(),
            self.interface,
            self.get_custom(settings),
            container=self.get_container(settings),
            config=self.get_config(settings),
        )
        self._task_template = obj
        return obj


class TaskTemplateResolver(TrackedInstance, TaskResolverMixin):
    """
    This is a special resolver that resolves the task above at execution time, using only the ``TaskTemplate``,
    meaning it should only be used for tasks that contain all pertinent information within the template itself.

    This class differs from some TaskResolverMixin pattern a bit. Most of the other resolvers you'll find,

    * restores the same task when ``load_task`` is called as the object that ``loader_args`` was called on.
      That is, even though at run time it's in a container on a cluster and is obviously a different Python process,
      the Python object in memory should look the same.
    * offers a one-to-one mapping between the list of strings returned by the ``loader_args`` function, an the task,
      at least within the container.

    This resolver differs in that,
    * when loading a task, the task that is a loaded is always an ``ExecutableTemplateShimTask``, regardless of what
      kind of task it was originally. It will only ever have what's available to it from the ``TaskTemplate``. No
      information that wasn't serialized into the template will be available.
    * all tasks will result in the same list of strings for a given subclass of the ``ShimTaskExecutor``
      executor. The strings will be ``["{{.taskTemplatePath}}", "path.to.your.executor"]``

    Also, ``get_all_tasks`` will always return an empty list, at least for now.
    """

    def __init__(self):
        super(TaskTemplateResolver, self).__init__()

    def name(self) -> str:
        return "task template resolver"

    # The return type of this function is different, it should be a Task, but it's not because it doesn't make
    # sense for ExecutableTemplateShimTask to inherit from Task.
    def load_task(self, loader_args: List[str]) -> ExecutableTemplateShimTask:  # type: ignore
        logger.info(f"Task template loader args: {loader_args}")
        ctx = FlyteContext.current_context()
        task_template_local_path = os.path.join(ctx.execution_state.working_dir, "task_template.pb")  # type: ignore
        ctx.file_access.get_data(loader_args[0], task_template_local_path)
        task_template_proto = load_proto_from_file(_tasks_pb2.TaskTemplate, task_template_local_path)
        task_template_model = _task_model.TaskTemplate.from_flyte_idl(task_template_proto)

        executor_class = load_object_from_module(loader_args[1])
        return ExecutableTemplateShimTask(task_template_model, executor_class)

    def loader_args(self, settings: SerializationSettings, t: PythonCustomizedContainerTask) -> List[str]:  # type: ignore
        return ["{{.taskTemplatePath}}", f"{t.executor_type.__module__}.{t.executor_type.__name__}"]

    def get_all_tasks(self) -> List[Task]:
        return []


default_task_template_resolver = TaskTemplateResolver()
