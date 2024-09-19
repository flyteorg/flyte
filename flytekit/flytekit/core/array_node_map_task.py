# TODO: has to support the SupportsNodeCreation protocol
import functools
import hashlib
import logging
import math
import os  # TODO: use flytekit logger
from contextlib import contextmanager
from typing import Any, Dict, List, Optional, Set, Union, cast

from flytekit.configuration import SerializationSettings
from flytekit.core import tracker
from flytekit.core.base_task import PythonTask, TaskResolverMixin
from flytekit.core.context_manager import ExecutionState, FlyteContext, FlyteContextManager
from flytekit.core.interface import transform_interface_to_list_interface
from flytekit.core.python_function_task import PythonFunctionTask, PythonInstanceTask
from flytekit.core.utils import timeit
from flytekit.exceptions import scopes as exception_scopes
from flytekit.loggers import logger
from flytekit.models.array_job import ArrayJob
from flytekit.models.core.workflow import NodeMetadata
from flytekit.models.interface import Variable
from flytekit.models.task import Container, K8sPod, Sql, Task
from flytekit.tools.module_loader import load_object_from_module


class ArrayNodeMapTask(PythonTask):
    def __init__(
        self,
        # TODO: add support for other Flyte entities
        python_function_task: Union[PythonFunctionTask, PythonInstanceTask, functools.partial],
        concurrency: Optional[int] = None,
        min_successes: Optional[int] = None,
        min_success_ratio: Optional[float] = None,
        bound_inputs: Optional[Set[str]] = None,
        **kwargs,
    ):
        """
        :param python_function_task: The task to be executed in parallel
        :param concurrency: The number of parallel executions to run
        :param min_successes: The minimum number of successful executions
        :param min_success_ratio: The minimum ratio of successful executions
        :param bound_inputs: The set of inputs that should be bound to the map task
        :param kwargs: Additional keyword arguments to pass to the base class
        """
        self._partial = None
        if isinstance(python_function_task, functools.partial):
            # TODO: We should be able to support partial tasks with lists as inputs
            for arg in python_function_task.keywords.values():
                if isinstance(arg, list):
                    raise ValueError("Cannot use a partial task with lists as inputs")
            self._partial = python_function_task
            actual_task = self._partial.func
        else:
            actual_task = python_function_task

        # TODO: add support for other Flyte entities
        if not (isinstance(actual_task, PythonFunctionTask) or isinstance(actual_task, PythonInstanceTask)):
            raise ValueError("Only PythonFunctionTask and PythonInstanceTask are supported in map tasks.")

        n_outputs = len(actual_task.python_interface.outputs)
        if n_outputs > 1:
            raise ValueError("Only tasks with a single output are supported in map tasks.")

        self._bound_inputs: Set[str] = bound_inputs or set(bound_inputs) if bound_inputs else set()
        if self._partial:
            self._bound_inputs.update(self._partial.keywords.keys())

        # Transform the interface to List[Optional[T]] in case `min_success_ratio` is set
        output_as_list_of_optionals = min_success_ratio is not None and min_success_ratio != 1 and n_outputs == 1
        collection_interface = transform_interface_to_list_interface(
            actual_task.python_interface, self._bound_inputs, output_as_list_of_optionals
        )

        self._run_task: Union[PythonFunctionTask, PythonInstanceTask] = actual_task  # type: ignore
        if isinstance(actual_task, PythonInstanceTask):
            mod = actual_task.task_type
            f = actual_task.lhs
        else:
            _, mod, f, _ = tracker.extract_task_module(cast(PythonFunctionTask, actual_task).task_function)
        sorted_bounded_inputs = ",".join(sorted(self._bound_inputs))
        h = hashlib.md5(
            f"{sorted_bounded_inputs}{concurrency}{min_successes}{min_success_ratio}".encode("utf-8")
        ).hexdigest()
        self._name = f"{mod}.map_{f}_{h}-arraynode"

        self._cmd_prefix: Optional[List[str]] = None
        self._concurrency: Optional[int] = concurrency
        self._min_successes: Optional[int] = min_successes
        self._min_success_ratio: Optional[float] = min_success_ratio
        self._collection_interface = collection_interface

        if "metadata" not in kwargs and actual_task.metadata:
            kwargs["metadata"] = actual_task.metadata
        if "security_ctx" not in kwargs and actual_task.security_context:
            kwargs["security_ctx"] = actual_task.security_context

        super().__init__(
            name=self.name,
            interface=collection_interface,
            task_type=self._run_task.task_type,
            task_config=None,
            task_type_version=1,
            **kwargs,
        )

    @property
    def name(self) -> str:
        return self._name

    @property
    def python_interface(self):
        return self._collection_interface

    def construct_node_metadata(self) -> NodeMetadata:
        # TODO: add support for other Flyte entities
        return NodeMetadata(
            name=self.name,
        )

    @property
    def min_success_ratio(self) -> Optional[float]:
        return self._min_success_ratio

    @property
    def min_successes(self) -> Optional[int]:
        return self._min_successes

    @property
    def concurrency(self) -> Optional[int]:
        return self._concurrency

    @property
    def python_function_task(self) -> Union[PythonFunctionTask, PythonInstanceTask]:
        return self._run_task

    @property
    def bound_inputs(self) -> Set[str]:
        return self._bound_inputs

    @contextmanager
    def prepare_target(self):
        """
        Alters the underlying run_task command to modify it for map task execution and then resets it after.
        """
        self.python_function_task.set_command_fn(self.get_command)
        try:
            yield
        finally:
            self.python_function_task.reset_command_fn()

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        return ArrayJob(parallelism=self._concurrency, min_success_ratio=self._min_success_ratio).to_dict()

    def get_config(self, settings: SerializationSettings) -> Optional[Dict[str, str]]:
        return self.python_function_task.get_config(settings)

    def get_container(self, settings: SerializationSettings) -> Container:
        with self.prepare_target():
            return self.python_function_task.get_container(settings)

    def get_k8s_pod(self, settings: SerializationSettings) -> K8sPod:
        with self.prepare_target():
            return self.python_function_task.get_k8s_pod(settings)

    def get_sql(self, settings: SerializationSettings) -> Sql:
        with self.prepare_target():
            return self.python_function_task.get_sql(settings)

    def get_command(self, settings: SerializationSettings) -> List[str]:
        """
        TODO ADD bound variables to the resolver. Maybe we need a different resolver?
        """
        mt = ArrayNodeMapTaskResolver()
        container_args = [
            "pyflyte-map-execute",
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
            "--experimental",
            "--resolver",
            mt.name(),
            "--",
            *mt.loader_args(settings, self),
        ]

        if self._cmd_prefix:
            return self._cmd_prefix + container_args
        return container_args

    def set_command_prefix(self, cmd: Optional[List[str]]):
        self._cmd_prefix = cmd

    def __call__(self, *args, **kwargs):
        """
        This call method modifies the kwargs and adds kwargs from partial.
        This is mostly done in the local_execute and compilation only.
        At runtime, the map_task is created with all the inputs filled in. to support this, we have modified
        the map_task interface in the constructor.
        """
        if self._partial:
            """If partial exists, then mix-in all partial values"""
            kwargs = {**self._partial.keywords, **kwargs}
        return super().__call__(*args, **kwargs)

    def execute(self, **kwargs) -> Any:
        ctx = FlyteContextManager.current_context()
        if ctx.execution_state and ctx.execution_state.mode == ExecutionState.Mode.TASK_EXECUTION:
            return self._execute_map_task(ctx, **kwargs)

        return self._raw_execute(**kwargs)

    def _execute_map_task(self, _: FlyteContext, **kwargs) -> Any:
        task_index = self._compute_array_job_index()
        map_task_inputs = {}
        for k in self.interface.inputs.keys():
            v = kwargs[k]
            if isinstance(v, list) and k not in self.bound_inputs:
                map_task_inputs[k] = v[task_index]
            else:
                map_task_inputs[k] = v
        return exception_scopes.user_entry_point(self.python_function_task.execute)(**map_task_inputs)

    @staticmethod
    def _compute_array_job_index() -> int:
        """
        Computes the absolute index of the current array job. This is determined by summing the compute-environment-specific
        environment variable and the offset (if one's set). The offset will be set and used when the user request that the
        job runs in a number of slots less than the size of the input.
        """
        return int(os.environ.get("BATCH_JOB_ARRAY_INDEX_OFFSET", "0")) + int(
            os.environ.get(os.environ.get("BATCH_JOB_ARRAY_INDEX_VAR_NAME", "0"), "0")
        )

    @property
    def _outputs_interface(self) -> Dict[Any, Variable]:
        """
        We override this method from PythonTask because the dispatch_execute method uses this
        interface to construct outputs. Each instance of an container_array task will however produce outputs
        according to the underlying run_task interface and the array plugin handler will actually create a collection
        from these individual outputs as the final output value.
        """

        ctx = FlyteContextManager.current_context()
        if ctx.execution_state is not None and ctx.execution_state.mode == ExecutionState.Mode.LOCAL_WORKFLOW_EXECUTION:
            # In workflow execution mode we actually need to use the parent (mapper) task output interface.
            return self.interface.outputs
        return self.python_function_task.interface.outputs

    def get_type_for_output_var(self, k: str, v: Any) -> type:
        """
        We override this method from flytekit.core.base_task Task because the dispatch_execute method uses this
        interface to construct outputs. Each instance of an container_array task will however produce outputs
        according to the underlying run_task interface and the array plugin handler will actually create a collection
        from these individual outputs as the final output value.
        """
        ctx = FlyteContextManager.current_context()
        if ctx.execution_state and ctx.execution_state.is_local_execution():
            # In workflow execution mode we actually need to use the parent (mapper) task output interface.
            return self._python_interface.outputs[k]
        return self.python_function_task.python_interface.outputs[k]

    def _raw_execute(self, **kwargs) -> Any:
        """
        This is called during locally run executions. Unlike array task execution on the Flyte platform, _raw_execute
        produces the full output collection.
        """
        outputs_expected = True
        if not self.interface.outputs:
            outputs_expected = False
        outputs = []

        mapped_tasks_count = 0
        if self._run_task.interface.inputs.items():
            for k in self._run_task.interface.inputs.keys():
                v = kwargs[k]
                if isinstance(v, list) and k not in self.bound_inputs:
                    mapped_tasks_count = len(v)
                    break

        failed_count = 0
        min_successes = mapped_tasks_count
        if self._min_successes:
            min_successes = self._min_successes
        elif self._min_success_ratio:
            min_successes = math.ceil(min_successes * self._min_success_ratio)

        for i in range(mapped_tasks_count):
            single_instance_inputs = {}
            for k in self.interface.inputs.keys():
                v = kwargs[k]
                if isinstance(v, list) and k not in self._bound_inputs:
                    single_instance_inputs[k] = kwargs[k][i]
                else:
                    single_instance_inputs[k] = kwargs[k]
            try:
                o = exception_scopes.user_entry_point(self._run_task.execute)(**single_instance_inputs)
                if outputs_expected:
                    outputs.append(o)
            except Exception as exc:
                outputs.append(None)
                failed_count += 1
                if mapped_tasks_count - failed_count < min_successes:
                    logger.error("The number of successful tasks is lower than the minimum ratio")
                    raise exc

        return outputs


def map_task(
    task_function: PythonFunctionTask,
    concurrency: int = 0,
    # TODO why no min_successes?
    min_success_ratio: float = 1.0,
    **kwargs,
):
    """Map task that uses the ``ArrayNode`` construct..

    .. important::

       This is an experimental drop-in replacement for :py:func:`~flytekit.map_task`.

    :param task_function: This argument is implicitly passed and represents the repeatable function
    :param concurrency: If specified, this limits the number of mapped tasks than can run in parallel to the given batch
        size. If the size of the input exceeds the concurrency value, then multiple batches will be run serially until
        all inputs are processed. If left unspecified, this means unbounded concurrency.
    :param min_success_ratio: If specified, this determines the minimum fraction of total jobs which can complete
        successfully before terminating this task and marking it successful.
    """
    return ArrayNodeMapTask(task_function, concurrency=concurrency, min_success_ratio=min_success_ratio, **kwargs)


class ArrayNodeMapTaskResolver(tracker.TrackedInstance, TaskResolverMixin):
    """
    Special resolver that is used for ArrayNodeMapTasks.
    This exists because it is possible that ArrayNodeMapTasks are created using nested "partial" subtasks.
    When a maptask is created its interface is interpolated from the interface of the subtask - the interpolation,
    simply converts every input into a list/collection input.

    For example:
      interface -> (i: int, j: str) -> str  => map_task interface -> (i: List[int], j: List[str]) -> List[str]

    But in cases in which `j` is bound to a fixed value by using `functools.partial` we need a way to ensure that
    the interface is not simply interpolated, but only the unbound inputs are interpolated.

        .. code-block:: python

            def foo((i: int, j: str) -> str:
                ...

            mt = map_task(functools.partial(foo, j=10))

            print(mt.interface)

    output:

            (i: List[int], j: str) -> List[str]

    But, at runtime this information is lost. To reconstruct this, we use ArrayNodeMapTaskResolver that records the "bound vars"
    and then at runtime reconstructs the interface with this knowledge
    """

    def name(self) -> str:
        return "ArrayNodeMapTaskResolver"

    @timeit("Load map task")
    def load_task(self, loader_args: List[str], max_concurrency: int = 0) -> ArrayNodeMapTask:
        """
        Loader args should be of the form
        vars "var1,var2,.." resolver "resolver" [resolver_args]
        """
        _, bound_vars, _, resolver, *resolver_args = loader_args
        logging.info(f"MapTask found task resolver {resolver} and arguments {resolver_args}")
        resolver_obj = load_object_from_module(resolver)
        # Use the resolver to load the actual task object
        _task_def = resolver_obj.load_task(loader_args=resolver_args)
        bound_inputs = set(bound_vars.split(","))
        return ArrayNodeMapTask(
            python_function_task=_task_def, max_concurrency=max_concurrency, bound_inputs=bound_inputs
        )

    def loader_args(self, settings: SerializationSettings, t: ArrayNodeMapTask) -> List[str]:  # type:ignore
        return [
            "vars",
            f'{",".join(sorted(t.bound_inputs))}',
            "resolver",
            t.python_function_task.task_resolver.location,
            *t.python_function_task.task_resolver.loader_args(settings, t.python_function_task),
        ]

    def get_all_tasks(self) -> List[Task]:
        raise NotImplementedError("MapTask resolver cannot return every instance of the map task")
