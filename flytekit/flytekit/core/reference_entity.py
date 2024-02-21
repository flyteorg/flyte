from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Any, Dict, Optional, Tuple, Type, Union

from flytekit.core.context_manager import BranchEvalMode, ExecutionState, FlyteContext
from flytekit.core.interface import Interface, transform_interface_to_typed_interface
from flytekit.core.promise import (
    Promise,
    VoidPromise,
    create_and_link_node,
    create_task_output,
    extract_obj_name,
    translate_inputs_to_literals,
)
from flytekit.core.type_engine import TypeEngine
from flytekit.exceptions import user as _user_exceptions
from flytekit.loggers import logger
from flytekit.models import interface as _interface_models
from flytekit.models import literals as _literal_models
from flytekit.models.core import identifier as _identifier_model
from flytekit.models.core import workflow as _workflow_model


@dataclass  # type: ignore
class Reference(ABC):
    project: str
    domain: str
    name: str
    version: str

    def __post_init__(self):
        self._id = _identifier_model.Identifier(self.resource_type, self.project, self.domain, self.name, self.version)

    @property
    def id(self) -> _identifier_model.Identifier:
        return self._id

    @property
    @abstractmethod
    def resource_type(self) -> int:
        ...


@dataclass
class TaskReference(Reference):
    """A reference object containing metadata that points to a remote task."""

    @property
    def resource_type(self) -> int:
        return _identifier_model.ResourceType.TASK


@dataclass
class LaunchPlanReference(Reference):
    """A reference object containing metadata that points to a remote launch plan."""

    @property
    def resource_type(self) -> int:
        return _identifier_model.ResourceType.LAUNCH_PLAN


@dataclass
class WorkflowReference(Reference):
    """A reference object containing metadata that points to a remote workflow."""

    @property
    def resource_type(self) -> int:
        return _identifier_model.ResourceType.WORKFLOW


class ReferenceEntity(object):
    def __init__(
        self,
        reference: Union[WorkflowReference, TaskReference, LaunchPlanReference],
        inputs: Dict[str, Type],
        outputs: Dict[str, Type],
    ):
        if (
            not isinstance(reference, WorkflowReference)
            and not isinstance(reference, TaskReference)
            and not isinstance(reference, LaunchPlanReference)
        ):
            raise Exception("Must be one of task, workflow, or launch plan")
        self._reference = reference
        self._native_interface = Interface(inputs=inputs, outputs=outputs)
        self._interface = transform_interface_to_typed_interface(self._native_interface)

    def execute(self, **kwargs) -> Any:
        raise Exception("Remote reference entities cannot be run locally. You must mock this out.")

    @property
    def python_interface(self) -> Interface:
        return self._native_interface

    @property
    def interface(self) -> _interface_models.TypedInterface:
        return self._interface

    @property
    def reference(self) -> Reference:
        return self._reference

    @property
    def name(self):
        return self._reference.id.name

    @property
    def id(self) -> _identifier_model.Identifier:
        return self.reference.id

    def unwrap_literal_map_and_execute(
        self, ctx: FlyteContext, input_literal_map: _literal_models.LiteralMap
    ) -> _literal_models.LiteralMap:
        """
        Please see the implementation of the dispatch_execute function in the real task.
        """

        # Invoked before the task is executed
        # Translate the input literals to Python native
        native_inputs = TypeEngine.literal_map_to_kwargs(ctx, input_literal_map, self.python_interface.inputs)

        logger.info(f"Invoking {self.name} with inputs: {native_inputs}")
        try:
            native_outputs = self.execute(**native_inputs)
        except Exception as e:
            logger.exception(f"Exception when executing {e}")
            raise e
        logger.debug("Task executed successfully in user level")

        expected_output_names = list(self.python_interface.outputs.keys())
        if len(expected_output_names) == 1:
            native_outputs_as_map = {expected_output_names[0]: native_outputs}
        elif len(expected_output_names) == 0:
            native_outputs_as_map = {}
        else:
            native_outputs_as_map = {expected_output_names[i]: native_outputs[i] for i, _ in enumerate(native_outputs)}

        # We manually construct a LiteralMap here because task inputs and outputs actually violate the assumption
        # built into the IDL that all the values of a literal map are of the same type.
        literals = {}
        for k, v in native_outputs_as_map.items():
            literal_type = self.interface.outputs[k].type
            py_type = self.python_interface.outputs[k]
            if isinstance(v, tuple):
                raise AssertionError(f"Output({k}) in task{self.name} received a tuple {v}, instead of {py_type}")
            literals[k] = TypeEngine.to_literal(ctx, v, py_type, literal_type)
        outputs_literal_map = _literal_models.LiteralMap(literals=literals)
        # After the execute has been successfully completed
        return outputs_literal_map

    def local_execute(self, ctx: FlyteContext, **kwargs) -> Optional[Union[Tuple[Promise], Promise, VoidPromise]]:
        """
        Please see the local_execute comments in the main task.
        """
        # Unwrap the kwargs values. After this, we essentially have a LiteralMap
        # The reason why we need to do this is because the inputs during local execute can be of 2 types
        #  - Promises or native constants
        #  Promises as essentially inputs from previous task executions
        #  native constants are just bound to this specific task (default values for a task input)
        #  Also alongwith promises and constants, there could be dictionary or list of promises or constants
        kwargs = translate_inputs_to_literals(
            ctx,
            incoming_values=kwargs,
            flyte_interface_types=self.interface.inputs,
            native_types=self.python_interface.inputs,
        )
        input_literal_map = _literal_models.LiteralMap(literals=kwargs)

        outputs_literal_map = self.unwrap_literal_map_and_execute(ctx, input_literal_map)

        # After running, we again have to wrap the outputs, if any, back into Promise objects
        outputs_literals = outputs_literal_map.literals
        output_names = list(self.python_interface.outputs.keys())
        if len(output_names) != len(outputs_literals):
            # Length check, clean up exception
            raise AssertionError(f"Length difference {len(output_names)} {len(outputs_literals)}")

        # Tasks that don't return anything still return a VoidPromise
        if len(output_names) == 0:
            return VoidPromise(self.name)

        vals = [Promise(var, outputs_literals[var]) for var in output_names]
        return create_task_output(vals, self.python_interface)

    def local_execution_mode(self):
        return ExecutionState.Mode.LOCAL_TASK_EXECUTION

    def construct_node_metadata(self) -> _workflow_model.NodeMetadata:
        return _workflow_model.NodeMetadata(name=extract_obj_name(self.name))

    def compile(self, ctx: FlyteContext, *args, **kwargs):
        return create_and_link_node(ctx, entity=self, **kwargs)

    def __call__(self, *args, **kwargs):
        # When a Task is () aka __called__, there are three things we may do:
        #  a. Plain execution Mode - just run the execute function. If not overridden, we should raise an exception
        #  b. Compilation Mode - this happens when the function is called as part of a workflow (potentially
        #     dynamic task). Produce promise objects and create a node.
        #  c. Workflow Execution Mode - when a workflow is being run locally. Even though workflows are functions
        #     and everything should be able to be passed through naturally, we'll want to wrap output values of the
        #     function into objects, so that potential .with_cpu or other ancillary functions can be attached to do
        #     nothing. Subsequent tasks will have to know how to unwrap these. If by chance a non-Flyte task uses a
        #     task output as an input, things probably will fail pretty obviously.
        #     Since this is a reference entity, it still needs to be mocked otherwise an exception will be raised.
        if len(args) > 0:
            raise _user_exceptions.FlyteAssertion(
                f"Cannot call reference entity with args - detected {len(args)} positional args {args}"
            )

        ctx = FlyteContext.current_context()
        if ctx.compilation_state is not None and ctx.compilation_state.mode == 1:
            return self.compile(ctx, *args, **kwargs)
        elif ctx.execution_state and ctx.execution_state.is_local_execution():
            if ctx.execution_state.branch_eval_mode == BranchEvalMode.BRANCH_SKIPPED:
                return
            return self.local_execute(ctx, **kwargs)
        else:
            logger.debug("Reference entity - running raw execute")
            return self.execute(**kwargs)


# ReferenceEntity is not a registerable entity and therefore the below classes do not need to inherit from
# flytekit.models.common.FlyteIdlEntity.
class ReferenceTemplate(object):
    def __init__(self, id: _identifier_model.Identifier, resource_type: int) -> None:
        """
        A reference template encapsulates all the information necessary to use reference entities within other
        workflows or dynamic tasks.

        :param flytekit.models.core.identifier.Identifier id: User-specified information that uniquely
            identifies this reference.
        :param int resource_type: The type of reference. See: flytekit.models.core.identifier.ResourceType
        """
        self._id = id
        self._resource_type = resource_type

    @property
    def id(self) -> _identifier_model.Identifier:
        """
        User-specified information that uniquely identifies this reference.
        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._id

    @property
    def resource_type(self) -> int:
        """
        The type of reference.
        :rtype: flytekit.models.core.identifier.ResourceType
        """
        return self._resource_type


class ReferenceSpec(object):
    def __init__(self, template: ReferenceTemplate) -> None:
        """
        :param ReferenceTemplate template:
        """
        self._template = template

    @property
    def template(self) -> ReferenceTemplate:
        """
        :rtype: ReferenceTemplate
        """
        return self._template
