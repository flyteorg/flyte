from __future__ import annotations

from typing import Any, Generic, Optional, Type, TypeVar, Union, cast

from flytekit.core.context_manager import ExecutionParameters, ExecutionState, FlyteContext, FlyteContextManager
from flytekit.core.tracker import TrackedInstance
from flytekit.core.type_engine import TypeEngine
from flytekit.loggers import logger
from flytekit.models import dynamic_job as _dynamic_job
from flytekit.models import literals as _literal_models
from flytekit.models import task as _task_model


class ExecutableTemplateShimTask(object):
    """
    The canonical ``@task`` decorated Python function task is pretty simple to reason about. At execution time (either
    locally or on a Flyte cluster), the function runs.

    This class, along with the ``ShimTaskExecutor`` class below, represents another execution pattern. This pattern,
    has two components:

    * The ``TaskTemplate``, or something like it like a ``FlyteTask``.
    * An executor, which can use information from the task template (including the ``custom`` field)

    Basically at execution time (both locally and on a Flyte cluster), the task template is given to the executor,
    which is responsible for computing and returning the results.

    .. note::

       The interface at execution time will have to derived from the Flyte IDL interface, which means it may be lossy.
       This is because when a task is serialized from Python into the ``TaskTemplate`` some information is lost because
       Flyte IDL can't keep track of every single Python type (or Java type if writing in the Java flytekit).

    This class also implements the ``dispatch_execute`` and ``execute`` functions to make it look like a ``PythonTask``
    that the ``entrypoint.py`` can execute, even though this class doesn't inherit from ``PythonTask``.
    """

    def __init__(self, tt: _task_model.TaskTemplate, executor_type: Type[ShimTaskExecutor], *args, **kwargs):
        self._executor_type = executor_type
        self._executor = executor_type()
        self._task_template = tt
        super().__init__(*args, **kwargs)

    @property
    def name(self) -> str:
        """Return the name of the underlying task."""
        if self._task_template is not None:
            return self._task_template.id.name
        # if not access the subclass's name
        return self._name  # type: ignore

    @property
    def task_template(self) -> _task_model.TaskTemplate:
        return self._task_template

    @property
    def executor(self) -> ShimTaskExecutor:
        return self._executor

    @property
    def executor_type(self) -> Type[ShimTaskExecutor]:
        return self._executor_type

    def execute(self, **kwargs) -> Any:
        """
        Rather than running here, send everything to the executor.
        """
        return self.executor.execute_from_model(self.task_template, **kwargs)

    def pre_execute(self, user_params: Optional[ExecutionParameters]) -> Optional[ExecutionParameters]:
        """
        This function is a stub, just here to keep dispatch_execute compatibility between this class and PythonTask.
        """
        return user_params

    def post_execute(self, _: Optional[ExecutionParameters], rval: Any) -> Any:
        """
        This function is a stub, just here to keep dispatch_execute compatibility between this class and PythonTask.
        """
        return rval

    def dispatch_execute(
        self, ctx: FlyteContext, input_literal_map: _literal_models.LiteralMap
    ) -> Union[_literal_models.LiteralMap, _dynamic_job.DynamicJobSpec]:
        """
        This function is largely similar to the base PythonTask, with the exception that we have to infer the Python
        interface before executing. Also, we refer to ``self.task_template`` rather than just ``self`` similar to task
        classes that derive from the base ``PythonTask``.
        """
        # Invoked before the task is executed
        new_user_params = self.pre_execute(ctx.user_space_params)

        # Create another execution context with the new user params, but let's keep the same working dir
        with FlyteContextManager.with_context(
            ctx.with_execution_state(
                cast(ExecutionState, ctx.execution_state).with_params(user_space_params=new_user_params)
            )
        ) as exec_ctx:
            # Added: Have to reverse the Python interface from the task template Flyte interface
            # See docstring for more details.
            guessed_python_input_types = TypeEngine.guess_python_types(self.task_template.interface.inputs)
            native_inputs = TypeEngine.literal_map_to_kwargs(exec_ctx, input_literal_map, guessed_python_input_types)

            logger.info(f"Invoking FlyteTask executor {self.task_template.id.name} with inputs: {native_inputs}")
            try:
                native_outputs = self.execute(**native_inputs)
            except Exception as e:
                logger.exception(f"Exception when executing {e}")
                raise e

            logger.debug("Task executed successfully in user level")
            # Lets run the post_execute method. This may result in a IgnoreOutputs Exception, which is
            # bubbled up to be handled at the callee layer.
            native_outputs = self.post_execute(new_user_params, native_outputs)

            # Short circuit the translation to literal map because what's returned may be a dj spec (or an
            # already-constructed LiteralMap if the dynamic task was a no-op), not python native values
            if isinstance(native_outputs, _literal_models.LiteralMap) or isinstance(
                native_outputs, _dynamic_job.DynamicJobSpec
            ):
                return native_outputs

            expected_output_names = list(self.task_template.interface.outputs.keys())
            if len(expected_output_names) == 1:
                # Here we have to handle the fact that the task could've been declared with a typing.NamedTuple of
                # length one. That convention is used for naming outputs - and single-length-NamedTuples are
                # particularly troublesome but elegant handling of them is not a high priority
                # Again, we're using the output_tuple_name as a proxy.
                # Deleted some stuff
                native_outputs_as_map = {expected_output_names[0]: native_outputs}
            elif len(expected_output_names) == 0:
                native_outputs_as_map = {}
            else:
                native_outputs_as_map = {
                    expected_output_names[i]: native_outputs[i] for i, _ in enumerate(native_outputs)
                }

            # We manually construct a LiteralMap here because task inputs and outputs actually violate the assumption
            # built into the IDL that all the values of a literal map are of the same type.
            literals = {}
            for k, v in native_outputs_as_map.items():
                literal_type = self.task_template.interface.outputs[k].type
                py_type = type(v)

                if isinstance(v, tuple):
                    raise AssertionError(
                        f"Output({k}) in task{self.task_template.id.name} received a tuple {v}, instead of {py_type}"
                    )
                try:
                    literals[k] = TypeEngine.to_literal(exec_ctx, v, py_type, literal_type)
                except Exception as e:
                    raise AssertionError(f"failed to convert return value for var {k}") from e

            outputs_literal_map = _literal_models.LiteralMap(literals=literals)
            # After the execute has been successfully completed
            return outputs_literal_map


T = TypeVar("T")


class ShimTaskExecutor(TrackedInstance, Generic[T]):
    def execute_from_model(self, tt: _task_model.TaskTemplate, **kwargs) -> Any:
        """
        This function must be overridden and is where all the business logic for running a task should live. Keep in
        mind that you're only working with the ``TaskTemplate``. You won't have access to any information in the task
        that wasn't serialized into the template.

        :param tt: This is the template, the serialized form of the task.
        :param kwargs: These are the Python native input values to the task.
        :return: Python native output values from the task.
        """
        raise NotImplementedError
