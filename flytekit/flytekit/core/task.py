from __future__ import annotations

import datetime as _datetime
from functools import update_wrapper
from typing import Any, Callable, Dict, Iterable, List, Optional, Type, TypeVar, Union, overload

from flytekit.core import launch_plan as _annotated_launchplan
from flytekit.core import workflow as _annotated_workflow
from flytekit.core.base_task import TaskMetadata, TaskResolverMixin
from flytekit.core.interface import transform_function_to_interface
from flytekit.core.pod_template import PodTemplate
from flytekit.core.python_function_task import PythonFunctionTask
from flytekit.core.reference_entity import ReferenceEntity, TaskReference
from flytekit.core.resources import Resources
from flytekit.extras.accelerators import BaseAccelerator
from flytekit.image_spec.image_spec import ImageSpec
from flytekit.models.documentation import Documentation
from flytekit.models.security import Secret


class TaskPlugins(object):
    """
    This is the TaskPlugins factory for task types that are derivative of PythonFunctionTask.
    Every task that the user wishes to use should be available in this factory.
    Usage

    .. code-block:: python

        TaskPlugins.register_pythontask_plugin(config_object_type, plugin_object_type)
        # config_object_type is any class that will be passed to the plugin_object as task_config
        # Plugin_object_type is a derivative of ``PythonFunctionTask``

    Examples of available task plugins include different query-based plugins such as
    :py:class:`flytekitplugins.athena.task.AthenaTask` and :py:class:`flytekitplugins.hive.task.HiveTask`, ML tools like
    :py:class:`plugins.awssagemaker.flytekitplugins.awssagemaker.training.SagemakerBuiltinAlgorithmsTask`, kubeflow
    operators like :py:class:`plugins.kfpytorch.flytekitplugins.kfpytorch.task.PyTorchFunctionTask` and
    :py:class:`plugins.kftensorflow.flytekitplugins.kftensorflow.task.TensorflowFunctionTask`, and generic plugins like
    :py:class:`flytekitplugins.pod.task.PodFunctionTask` which doesn't integrate with third party tools or services.

    The `task_config` is different for every task plugin type. This is filled out by users when they define a task to
    specify plugin-specific behavior and features.  For example, with a query type task plugin, the config might store
    information related to which database to query.

    The `plugin_object_type` can be used to customize execution behavior and task serialization properties in tandem
    with the `task_config`.
    """

    _PYTHONFUNCTION_TASK_PLUGINS: Dict[type, Type[PythonFunctionTask]] = {}

    @classmethod
    def register_pythontask_plugin(cls, plugin_config_type: type, plugin: Type[PythonFunctionTask]):
        """
        Use this method to register a new plugin into Flytekit. Usage ::

        .. code-block:: python

            TaskPlugins.register_pythontask_plugin(config_object_type, plugin_object_type)
            # config_object_type is any class that will be passed to the plugin_object as task_config
            # Plugin_object_type is a derivative of ``PythonFunctionTask``
        """
        if plugin_config_type in cls._PYTHONFUNCTION_TASK_PLUGINS:
            found = cls._PYTHONFUNCTION_TASK_PLUGINS[plugin_config_type]
            if found == plugin:
                return
            raise TypeError(
                f"Requesting to register plugin {plugin} - collides with existing plugin {found}"
                f" for type {plugin_config_type}"
            )

        cls._PYTHONFUNCTION_TASK_PLUGINS[plugin_config_type] = plugin

    @classmethod
    def find_pythontask_plugin(cls, plugin_config_type: type) -> Type[PythonFunctionTask]:
        """
        Returns a PluginObjectType if found or returns the base PythonFunctionTask
        """
        if plugin_config_type in cls._PYTHONFUNCTION_TASK_PLUGINS:
            return cls._PYTHONFUNCTION_TASK_PLUGINS[plugin_config_type]
        # Defaults to returning Base PythonFunctionTask
        return PythonFunctionTask


T = TypeVar("T")
FuncOut = TypeVar("FuncOut")


@overload
def task(
    _task_function: None = ...,
    task_config: Optional[T] = ...,
    cache: bool = ...,
    cache_serialize: bool = ...,
    cache_version: str = ...,
    retries: int = ...,
    interruptible: Optional[bool] = ...,
    deprecated: str = ...,
    timeout: Union[_datetime.timedelta, int] = ...,
    container_image: Optional[Union[str, ImageSpec]] = ...,
    environment: Optional[Dict[str, str]] = ...,
    requests: Optional[Resources] = ...,
    limits: Optional[Resources] = ...,
    secret_requests: Optional[List[Secret]] = ...,
    execution_mode: PythonFunctionTask.ExecutionBehavior = ...,
    node_dependency_hints: Optional[
        Iterable[Union[PythonFunctionTask, _annotated_launchplan.LaunchPlan, _annotated_workflow.WorkflowBase]]
    ] = ...,
    task_resolver: Optional[TaskResolverMixin] = ...,
    docs: Optional[Documentation] = ...,
    disable_deck: Optional[bool] = ...,
    enable_deck: Optional[bool] = ...,
    pod_template: Optional["PodTemplate"] = ...,
    pod_template_name: Optional[str] = ...,
    accelerator: Optional[BaseAccelerator] = ...,
) -> Callable[[Callable[..., FuncOut]], PythonFunctionTask[T]]:
    ...


@overload
def task(
    _task_function: Callable[..., FuncOut],
    task_config: Optional[T] = ...,
    cache: bool = ...,
    cache_serialize: bool = ...,
    cache_version: str = ...,
    retries: int = ...,
    interruptible: Optional[bool] = ...,
    deprecated: str = ...,
    timeout: Union[_datetime.timedelta, int] = ...,
    container_image: Optional[Union[str, ImageSpec]] = ...,
    environment: Optional[Dict[str, str]] = ...,
    requests: Optional[Resources] = ...,
    limits: Optional[Resources] = ...,
    secret_requests: Optional[List[Secret]] = ...,
    execution_mode: PythonFunctionTask.ExecutionBehavior = ...,
    node_dependency_hints: Optional[
        Iterable[Union[PythonFunctionTask, _annotated_launchplan.LaunchPlan, _annotated_workflow.WorkflowBase]]
    ] = ...,
    task_resolver: Optional[TaskResolverMixin] = ...,
    docs: Optional[Documentation] = ...,
    disable_deck: Optional[bool] = ...,
    enable_deck: Optional[bool] = ...,
    pod_template: Optional["PodTemplate"] = ...,
    pod_template_name: Optional[str] = ...,
    accelerator: Optional[BaseAccelerator] = ...,
) -> Union[PythonFunctionTask[T], Callable[..., FuncOut]]:
    ...


def task(
    _task_function: Optional[Callable[..., FuncOut]] = None,
    task_config: Optional[T] = None,
    cache: bool = False,
    cache_serialize: bool = False,
    cache_version: str = "",
    retries: int = 0,
    interruptible: Optional[bool] = None,
    deprecated: str = "",
    timeout: Union[_datetime.timedelta, int] = 0,
    container_image: Optional[Union[str, ImageSpec]] = None,
    environment: Optional[Dict[str, str]] = None,
    requests: Optional[Resources] = None,
    limits: Optional[Resources] = None,
    secret_requests: Optional[List[Secret]] = None,
    execution_mode: PythonFunctionTask.ExecutionBehavior = PythonFunctionTask.ExecutionBehavior.DEFAULT,
    node_dependency_hints: Optional[
        Iterable[Union[PythonFunctionTask, _annotated_launchplan.LaunchPlan, _annotated_workflow.WorkflowBase]]
    ] = None,
    task_resolver: Optional[TaskResolverMixin] = None,
    docs: Optional[Documentation] = None,
    disable_deck: Optional[bool] = None,
    enable_deck: Optional[bool] = None,
    pod_template: Optional["PodTemplate"] = None,
    pod_template_name: Optional[str] = None,
    accelerator: Optional[BaseAccelerator] = None,
) -> Union[Callable[[Callable[..., FuncOut]], PythonFunctionTask[T]], PythonFunctionTask[T], Callable[..., FuncOut]]:
    """
    This is the core decorator to use for any task type in flytekit.

    Tasks are the building blocks of Flyte. They represent users code. Tasks have the following properties

    * Versioned (usually tied to the git revision SHA1)
    * Strong interfaces (specified inputs and outputs)
    * Declarative
    * Independently executable
    * Unit testable

    For a simple python task,

    .. code-block:: python

        @task
        def my_task(x: int, y: typing.Dict[str, str]) -> str:
            ...

    For specific task types

    .. code-block:: python

        @task(task_config=Spark(), retries=3)
        def my_task(x: int, y: typing.Dict[str, str]) -> str:
            ...

    Please see some cookbook :std:ref:`task examples <cookbook:tasks>` for additional information.

    :param _task_function: This argument is implicitly passed and represents the decorated function
    :param task_config: This argument provides configuration for a specific task types.
                        Please refer to the plugins documentation for the right object to use.
    :param cache: Boolean that indicates if caching should be enabled
    :param cache_serialize: Boolean that indicates if identical (ie. same inputs) instances of this task should be
          executed in serial when caching is enabled. This means that given multiple concurrent executions over
          identical inputs, only a single instance executes and the rest wait to reuse the cached results. This
          parameter does nothing without also setting the cache parameter.
    :param cache_version: Cache version to use. Changes to the task signature will automatically trigger a cache miss,
           but you can always manually update this field as well to force a cache miss. You should also manually bump
           this version if the function body/business logic has changed, but the signature hasn't.
    :param retries: Number of times to retry this task during a workflow execution.
    :param interruptible: [Optional] Boolean that indicates that this task can be interrupted and/or scheduled on nodes
                          with lower QoS guarantees. This will directly reduce the `$`/`execution cost` associated,
                          at the cost of performance penalties due to potential interruptions. Requires additional
                          Flyte platform level configuration. If no value is provided, the task will inherit this
                          attribute from its workflow, as follows:
                          No values set for interruptible at the task or workflow level - task is not interruptible
                          Task has interruptible=True, but workflow has no value set - task is interruptible
                          Workflow has interruptible=True, but task has no value set - task is interruptible
                          Workflow has interruptible=False, but task has interruptible=True - task is interruptible
                          Workflow has interruptible=True, but task has interruptible=False - task is not interruptible
    :param deprecated: A string that can be used to provide a warning message for deprecated task. Absence / empty str
                       indicates that the task is active and not deprecated
    :param timeout: the max amount of time for which one execution of this task should be executed for. The execution
                    will be terminated if the runtime exceeds the given timeout (approximately).
    :param container_image: By default the configured FLYTE_INTERNAL_IMAGE is used for every task. This directive can be
                used to provide an alternate image for a specific task. This is useful for the cases in which images
                bloat because of various dependencies and a dependency is only required for this or a set of tasks,
                and they vary from the default.

                .. code-block:: python

                    # Use default image name `fqn` and alter the tag to `tag-{{default.tag}}` tag of the default image
                    # with a prefix. In this case, it is assumed that the image like
                    # flytecookbook:tag-gitsha is published alongwith the default of flytecookbook:gitsha
                    @task(container_image='{{.images.default.fqn}}:tag-{{images.default.tag}}')
                    def foo():
                        ...

                    # Refer to configurations to configure fqns for other images besides default. In this case it will
                    # lookup for an image named xyz
                    @task(container_image='{{.images.xyz.fqn}}:{{images.default.tag}}')
                    def foo2():
                        ...
    :param environment: Environment variables that should be added for this tasks execution
    :param requests: Specify compute resource requests for your task. For Pod-plugin tasks, these values will apply only
      to the primary container.
    :param limits: Compute limits. Specify compute resource limits for your task. For Pod-plugin tasks, these values
      will apply only to the primary container. For more information, please see :py:class:`flytekit.Resources`.
    :param secret_requests: Keys that can identify the secrets supplied at runtime. Ideally the secret keys should also be
                     semi-descriptive. The key values will be available from runtime, if the backend is configured
                     to provide secrets and if secrets are available in the configured secrets store.
                     Possible options for secret stores are - Vault, Confidant, Kube secrets, AWS KMS etc
                     Refer to :py:class:`Secret` to understand how to specify the request for a secret. It
                     may change based on the backend provider.
    :param execution_mode: This is mainly for internal use. Please ignore. It is filled in automatically.
    :param node_dependency_hints: A list of tasks, launchplans, or workflows that this task depends on. This is only
        for dynamic tasks/workflows, where flyte cannot automatically determine the dependencies prior to runtime.
        Even on dynamic tasks this is optional, but in some scenarios it will make registering the workflow easier,
        because it allows registration to be done the same as for static tasks/workflows.

        For example this is useful to run launchplans dynamically, because launchplans must be registered on flyteadmin
        before they can be run. Tasks and workflows do not have this requirement.

        .. code-block:: python

            @workflow
            def workflow0():
                ...

            launchplan0 = LaunchPlan.get_or_create(workflow0)

            # Specify node_dependency_hints so that launchplan0 will be registered on flyteadmin, despite this being a
            # dynamic task.
            @dynamic(node_dependency_hints=[launchplan0])
            def launch_dynamically():
                # To run a sub-launchplan it must have previously been registered on flyteadmin.
                return [launchplan0]*10
    :param task_resolver: Provide a custom task resolver.
    :param disable_deck: (deprecated) If true, this task will not output deck html file
    :param enable_deck: If true, this task will output deck html file
    :param docs: Documentation about this task
    :param pod_template: Custom PodTemplate for this task.
    :param pod_template_name: The name of the existing PodTemplate resource which will be used in this task.
    :param accelerator: The accelerator to use for this task.
    """

    def wrapper(fn: Callable[..., Any]) -> PythonFunctionTask[T]:
        _metadata = TaskMetadata(
            cache=cache,
            cache_serialize=cache_serialize,
            cache_version=cache_version,
            retries=retries,
            interruptible=interruptible,
            deprecated=deprecated,
            timeout=timeout,
        )

        task_instance = TaskPlugins.find_pythontask_plugin(type(task_config))(
            task_config,
            fn,
            metadata=_metadata,
            container_image=container_image,
            environment=environment,
            requests=requests,
            limits=limits,
            secret_requests=secret_requests,
            execution_mode=execution_mode,
            node_dependency_hints=node_dependency_hints,
            task_resolver=task_resolver,
            disable_deck=disable_deck,
            enable_deck=enable_deck,
            docs=docs,
            pod_template=pod_template,
            pod_template_name=pod_template_name,
            accelerator=accelerator,
        )
        update_wrapper(task_instance, fn)
        return task_instance

    if _task_function:
        return wrapper(_task_function)
    else:
        return wrapper


class ReferenceTask(ReferenceEntity, PythonFunctionTask):  # type: ignore
    """
    This is a reference task, the body of the function passed in through the constructor will never be used, only the
    signature of the function will be. The signature should also match the signature of the task you're referencing,
    as stored by Flyte Admin, if not, workflows using this will break upon compilation.
    """

    def __init__(
        self, project: str, domain: str, name: str, version: str, inputs: Dict[str, type], outputs: Dict[str, Type]
    ):
        super().__init__(TaskReference(project, domain, name, version), inputs, outputs)

        # Reference tasks shouldn't call the parent constructor, but the parent constructor is what sets the resolver
        self._task_resolver = None  # type: ignore


def reference_task(
    project: str,
    domain: str,
    name: str,
    version: str,
) -> Callable[[Callable[..., Any]], ReferenceTask]:
    """
    A reference task is a pointer to a task that already exists on your Flyte installation. This
    object will not initiate a network call to Admin, which is why the user is asked to provide the expected interface.
    If at registration time the interface provided causes an issue with compilation, an error will be returned.

    Example:

    .. literalinclude:: ../../../tests/flytekit/unit/core/test_references.py
       :pyobject: ref_t1

    """

    def wrapper(fn) -> ReferenceTask:
        interface = transform_function_to_interface(fn)
        return ReferenceTask(project, domain, name, version, interface.inputs, interface.outputs)

    return wrapper
