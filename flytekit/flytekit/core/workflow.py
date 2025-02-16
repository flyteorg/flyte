from __future__ import annotations

import asyncio
import inspect
import typing
from dataclasses import dataclass
from enum import Enum
from functools import update_wrapper
from typing import Any, Callable, Coroutine, Dict, List, Optional, Tuple, Type, Union, cast, overload

from flytekit.core import constants as _common_constants
from flytekit.core import launch_plan as _annotated_launch_plan
from flytekit.core.base_task import PythonTask, Task
from flytekit.core.class_based_resolver import ClassStorageTaskResolver
from flytekit.core.condition import ConditionalSection, conditional
from flytekit.core.context_manager import (
    CompilationState,
    ExecutionState,
    FlyteContext,
    FlyteContextManager,
    FlyteEntities,
)
from flytekit.core.docstring import Docstring
from flytekit.core.interface import (
    Interface,
    transform_function_to_interface,
    transform_interface_to_typed_interface,
)
from flytekit.core.node import Node
from flytekit.core.promise import (
    NodeOutput,
    Promise,
    VoidPromise,
    binding_from_python_std,
    create_task_output,
    extract_obj_name,
    flyte_entity_call_handler,
    translate_inputs_to_literals,
)
from flytekit.core.python_auto_container import PythonAutoContainerTask
from flytekit.core.reference_entity import ReferenceEntity, WorkflowReference
from flytekit.core.tracker import extract_task_module
from flytekit.core.type_engine import TypeEngine
from flytekit.exceptions import scopes as exception_scopes
from flytekit.exceptions.user import FlyteValidationException, FlyteValueException
from flytekit.loggers import logger
from flytekit.models import interface as _interface_models
from flytekit.models import literals as _literal_models
from flytekit.models.core import workflow as _workflow_model
from flytekit.models.documentation import Description, Documentation
from flytekit.types.error import FlyteError

GLOBAL_START_NODE = Node(
    id=_common_constants.GLOBAL_INPUT_NODE_ID,
    metadata=None,
    bindings=[],
    upstream_nodes=[],
    flyte_entity=None,
)

T = typing.TypeVar("T")
FuncOut = typing.TypeVar("FuncOut")


class WorkflowFailurePolicy(Enum):
    """
    Defines the behavior for a workflow execution in the case of an observed node execution failure. By default, a
    workflow execution will immediately enter a failed state if a component node fails.
    """

    #: Causes the entire workflow execution to fail once a component node fails.
    FAIL_IMMEDIATELY = _workflow_model.WorkflowMetadata.OnFailurePolicy.FAIL_IMMEDIATELY

    #: Will proceed to run any remaining runnable nodes once a component node fails.
    FAIL_AFTER_EXECUTABLE_NODES_COMPLETE = (
        _workflow_model.WorkflowMetadata.OnFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE
    )


@dataclass
class WorkflowMetadata(object):
    on_failure: WorkflowFailurePolicy

    def __post_init__(self):
        if (
            self.on_failure != WorkflowFailurePolicy.FAIL_IMMEDIATELY
            and self.on_failure != WorkflowFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE
        ):
            raise FlyteValidationException(f"Failure policy {self.on_failure} not acceptable")

    def to_flyte_model(self):
        if self.on_failure == WorkflowFailurePolicy.FAIL_IMMEDIATELY:
            on_failure = 0
        else:
            on_failure = 1
        return _workflow_model.WorkflowMetadata(on_failure=on_failure)


@dataclass
class WorkflowMetadataDefaults(object):
    """
    This class is similarly named to the one above. Please see the IDL for more information but essentially, this
    WorkflowMetadataDefaults class represents the defaults that are handed down to a workflow's tasks, whereas
    WorkflowMetadata represents metadata about the workflow itself.
    """

    interruptible: bool

    def __post_init__(self):
        # TODO: Get mypy working so we don't have to worry about these checks
        if self.interruptible is not True and self.interruptible is not False:
            raise FlyteValidationException(f"Interruptible must be boolean, {self.interruptible} invalid")

    def to_flyte_model(self):
        return _workflow_model.WorkflowMetadataDefaults(interruptible=self.interruptible)


def construct_input_promises(inputs: List[str]) -> Dict[str, Promise]:
    return {
        input_name: Promise(var=input_name, val=NodeOutput(node=GLOBAL_START_NODE, var=input_name))
        for input_name in inputs
    }


def get_promise(binding_data: _literal_models.BindingData, outputs_cache: Dict[Node, Dict[str, Promise]]) -> Promise:
    """
    This is a helper function that will turn a binding into a Promise object, using a lookup map. Please see
    get_promise_map for the rest of the details.
    """
    if binding_data.promise is not None:
        if not isinstance(binding_data.promise, NodeOutput):
            raise FlyteValidationException(
                f"Binding data Promises have to be of the NodeOutput type {type(binding_data.promise)} found"
            )
        # b.var is the name of the input to the task
        # binding_data.promise.var is the name of the upstream node's output we want
        return outputs_cache[binding_data.promise.node][binding_data.promise.var]
    elif binding_data.scalar is not None:
        return Promise(var="placeholder", val=_literal_models.Literal(scalar=binding_data.scalar))
    elif binding_data.collection is not None:
        literals = []
        for bd in binding_data.collection.bindings:
            p = get_promise(bd, outputs_cache)
            literals.append(p.val)
        return Promise(
            var="placeholder",
            val=_literal_models.Literal(collection=_literal_models.LiteralCollection(literals=literals)),
        )
    elif binding_data.map is not None:
        literals = {}  # type: ignore
        for k, bd in binding_data.map.bindings.items():
            p = get_promise(bd, outputs_cache)
            literals[k] = p.val
        return Promise(
            var="placeholder", val=_literal_models.Literal(map=_literal_models.LiteralMap(literals=literals))
        )

    raise FlyteValidationException("Binding type unrecognized.")


def get_promise_map(
    bindings: List[_literal_models.Binding], outputs_cache: Dict[Node, Dict[str, Promise]]
) -> Dict[str, Promise]:
    """
    Local execution of imperatively defined workflows is done node by node. This function will fill in the node's
    entity's input arguments, which are specified using the bindings list, and a map of nodes to its outputs.
    Basically this takes the place of propeller in resolving bindings, pulling in outputs from previously completed
    nodes and filling in the necessary inputs.
    """
    entity_kwargs = {}
    for b in bindings:
        entity_kwargs[b.var] = get_promise(b.binding, outputs_cache)

    return entity_kwargs


class WorkflowBase(object):
    def __init__(
        self,
        name: str,
        workflow_metadata: WorkflowMetadata,
        workflow_metadata_defaults: WorkflowMetadataDefaults,
        python_interface: Interface,
        on_failure: Optional[Union[WorkflowBase, Task]] = None,
        docs: Optional[Documentation] = None,
        **kwargs,
    ):
        self._name = name
        self._workflow_metadata = workflow_metadata
        self._workflow_metadata_defaults = workflow_metadata_defaults
        self._python_interface = python_interface
        self._interface = transform_interface_to_typed_interface(python_interface)
        self._inputs: Dict[str, Promise] = {}
        self._unbound_inputs: typing.Set[Promise] = set()
        self._nodes: List[Node] = []
        self._output_bindings: List[_literal_models.Binding] = []
        self._on_failure = on_failure
        self._failure_node = None
        self._docs = docs

        if self._python_interface.docstring:
            if self.docs is None:
                self._docs = Documentation(
                    short_description=self._python_interface.docstring.short_description,
                    long_description=Description(value=self._python_interface.docstring.long_description),
                )
            else:
                if self._python_interface.docstring.short_description:
                    cast(
                        Documentation, self._docs
                    ).short_description = self._python_interface.docstring.short_description
                if self._python_interface.docstring.long_description:
                    self._docs = Description(value=self._python_interface.docstring.long_description)

        FlyteEntities.entities.append(self)
        super().__init__(**kwargs)

    @property
    def name(self) -> str:
        return self._name

    @property
    def docs(self):
        return self._docs

    @property
    def short_name(self) -> str:
        return extract_obj_name(self._name)

    @property
    def workflow_metadata(self) -> WorkflowMetadata:
        return self._workflow_metadata

    @property
    def workflow_metadata_defaults(self) -> WorkflowMetadataDefaults:
        return self._workflow_metadata_defaults

    @property
    def python_interface(self) -> Interface:
        return self._python_interface

    @property
    def interface(self) -> _interface_models.TypedInterface:
        return self._interface

    @property
    def output_bindings(self) -> List[_literal_models.Binding]:
        self.compile()
        return self._output_bindings

    @property
    def nodes(self) -> List[Node]:
        self.compile()
        return self._nodes

    @property
    def on_failure(self) -> Optional[Union[WorkflowBase, Task]]:
        return self._on_failure

    @property
    def failure_node(self) -> Optional[Node]:
        return self._failure_node

    def __repr__(self):
        return (
            f"WorkflowBase - {self._name} && "
            f"Inputs ({len(self._python_interface.inputs)}): {self._python_interface.inputs} && "
            f"Outputs ({len(self._python_interface.outputs)}): {self._python_interface.outputs} && "
            f"Output bindings: {self._output_bindings} && "
        )

    def construct_node_metadata(self) -> _workflow_model.NodeMetadata:
        return _workflow_model.NodeMetadata(
            name=extract_obj_name(self.name),
            interruptible=self.workflow_metadata_defaults.interruptible,
        )

    def __call__(self, *args, **kwargs) -> Union[Tuple[Promise], Promise, VoidPromise, Tuple, Coroutine, None]:
        """
        Workflow needs to fill in default arguments before invoking the call handler.
        """
        # Get default arguments and override with kwargs passed in
        input_kwargs = self.python_interface.default_inputs_as_kwargs
        input_kwargs.update(kwargs)
        self.compile()
        try:
            return flyte_entity_call_handler(self, *args, **input_kwargs)
        except Exception as exc:
            if self.on_failure:
                if self.on_failure.python_interface and "err" in self.on_failure.python_interface.inputs:
                    input_kwargs["err"] = FlyteError(failed_node_id="", message=str(exc))
                self.on_failure(**input_kwargs)
            raise exc

    def execute(self, **kwargs):
        raise Exception("Should not be called")

    def compile(self, **kwargs):
        pass

    def local_execute(self, ctx: FlyteContext, **kwargs) -> Union[Tuple[Promise], Promise, VoidPromise, None]:
        # This is done to support the invariant that Workflow local executions always work with Promise objects
        # holding Flyte literal values. Even in a wf, a user can call a sub-workflow with a Python native value.
        literal_map = translate_inputs_to_literals(
            ctx,
            incoming_values=kwargs,
            flyte_interface_types=self.interface.inputs,
            native_types=self.python_interface.inputs,
        )
        kwargs_literals = {k: Promise(var=k, val=v) for k, v in literal_map.items()}
        self.compile()
        function_outputs = self.execute(**kwargs_literals)

        if inspect.iscoroutine(function_outputs):
            # handle coroutines for eager workflows
            function_outputs = asyncio.run(function_outputs)

        # First handle the empty return case.
        # A workflow function may return a task that doesn't return anything
        #   def wf():
        #       return t1()
        # or it may not return at all
        #   def wf():
        #       t1()
        # In the former case we get the task's VoidPromise, in the latter we get None
        if isinstance(function_outputs, VoidPromise) or function_outputs is None:
            if len(self.python_interface.outputs) != 0:
                raise FlyteValueException(
                    function_outputs,
                    f"Interface has {len(self.python_interface.outputs)} outputs.",
                )
            return VoidPromise(self.name)

        # Because we should've already returned in the above check, we just raise an error here.
        if len(self.python_interface.outputs) == 0:
            raise FlyteValueException(function_outputs, "Interface output should've been VoidPromise or None.")

        expected_output_names = list(self.python_interface.outputs.keys())
        if len(expected_output_names) == 1:
            # Here we have to handle the fact that the wf could've been declared with a typing.NamedTuple of
            # length one. That convention is used for naming outputs - and single-length-NamedTuples are
            # particularly troublesome but elegant handling of them is not a high priority
            # Again, we're using the output_tuple_name as a proxy.
            if self.python_interface.output_tuple_name and isinstance(function_outputs, tuple):
                wf_outputs_as_map = {expected_output_names[0]: function_outputs[0]}
            else:
                wf_outputs_as_map = {expected_output_names[0]: function_outputs}
        else:
            wf_outputs_as_map = {expected_output_names[i]: function_outputs[i] for i, _ in enumerate(function_outputs)}

        # Basically we need to repackage the promises coming from the tasks into Promises that match the workflow's
        # interface. We do that by extracting out the literals, and creating new Promises
        wf_outputs_as_literal_dict = translate_inputs_to_literals(
            ctx,
            wf_outputs_as_map,
            flyte_interface_types=self.interface.outputs,
            native_types=self.python_interface.outputs,
        )
        # Recreate new promises that use the workflow's output names.
        new_promises = [Promise(var, wf_outputs_as_literal_dict[var]) for var in expected_output_names]

        return create_task_output(new_promises, self.python_interface)

    def local_execution_mode(self) -> ExecutionState.Mode:
        """ """
        return ExecutionState.Mode.LOCAL_WORKFLOW_EXECUTION


class ImperativeWorkflow(WorkflowBase):
    """
    An imperative workflow is a programmatic analogue to the typical ``@workflow`` function-based workflow and is
    better suited to programmatic applications.

    Assuming you have some tasks like so

    .. literalinclude:: ../../../tests/flytekit/unit/core/test_imperative.py
       :start-after: # docs_tasks_start
       :end-before: # docs_tasks_end
       :language: python
       :dedent: 4

    You could create a workflow imperatively like so

    .. literalinclude:: ../../../tests/flytekit/unit/core/test_imperative.py
       :start-after: # docs_start
       :end-before: # docs_end
       :language: python
       :dedent: 4

    This workflow would be identical on the back-end to

    .. literalinclude:: ../../../tests/flytekit/unit/core/test_imperative.py
       :start-after: # docs_equivalent_start
       :end-before: # docs_equivalent_end
       :language: python
       :dedent: 4

    Note that the only reason we need the ``NamedTuple`` is so we can name the output the same thing as in the
    imperative example. The imperative paradigm makes the naming of workflow outputs easier, but this isn't a big
    deal in function-workflows because names tend to not be necessary.
    """

    def __init__(
        self,
        name: str,
        failure_policy: Optional[WorkflowFailurePolicy] = None,
        interruptible: bool = False,
    ):
        metadata = WorkflowMetadata(on_failure=failure_policy or WorkflowFailurePolicy.FAIL_IMMEDIATELY)
        workflow_metadata_defaults = WorkflowMetadataDefaults(interruptible)
        self._compilation_state = CompilationState(prefix="")
        self._inputs = {}
        # This unbound inputs construct is just here to help workflow authors detect issues a bit earlier. It just
        # keeps track of workflow inputs that you've declared with add_workflow_input but haven't yet consumed. This
        # is an error that Admin would return at compile time anyways, but this allows flytekit to raise
        # the error earlier.
        self._unbound_inputs = set()
        super().__init__(
            name=name,
            workflow_metadata=metadata,
            workflow_metadata_defaults=workflow_metadata_defaults,
            python_interface=Interface(),
        )

    @property
    def compilation_state(self) -> CompilationState:
        """
        Compilation is done a bit at a time, one task or other entity call at a time. This is why this workflow
        class has to keep track of its own compilation state.
        """
        return self._compilation_state

    @property
    def nodes(self) -> List[Node]:
        return self._compilation_state.nodes

    @property
    def inputs(self) -> Dict[str, Promise]:
        """
        This holds the input promises to the workflow. The nodes in these Promise objects should always point to
        the global start node.
        """
        return self._inputs

    def __repr__(self):
        return super().__repr__() + f"Nodes ({len(self.compilation_state.nodes)}): {self.compilation_state.nodes}"

    def execute(self, **kwargs):
        """
        Called by local_execute. This function is how local execution for imperative workflows runs. Because when an
        entity is added using the add_entity function, all inputs to that entity should've been already declared, we
        can just iterate through the nodes in order and we shouldn't run into any dependency issues. That is, we force
        the user to declare entities already in a topological sort. To keep track of outputs, we create a map to
        start things off, filled in only with the workflow inputs (if any). As things are run, their outputs are stored
        in this map.
        After all nodes are run, we fill in workflow level outputs the same way as any other previous node.
        """
        if not self.ready():
            raise FlyteValidationException(f"Workflow not ready, wf is currently {self}")

        # Create a map that holds the outputs of each node.
        intermediate_node_outputs: Dict[Node, Dict[str, Promise]] = {GLOBAL_START_NODE: {}}  # type: ignore

        # Start things off with the outputs of the global input node, i.e. the inputs to the workflow.
        # local_execute should've already ensured that all the values in kwargs are Promise objects
        for k, v in kwargs.items():
            intermediate_node_outputs[GLOBAL_START_NODE][k] = v

        # Next iterate through the nodes in order.
        for node in self.compilation_state.nodes:
            if node not in intermediate_node_outputs.keys():
                intermediate_node_outputs[node] = {}

            # Retrieve the entity from the node, and call it by looking up the promises the node's bindings require,
            # and then fill them in using the node output tracker map we have.
            entity = node.flyte_entity
            entity_kwargs = get_promise_map(node.bindings, intermediate_node_outputs)

            # Handle the calling and outputs of each node's entity
            results = entity(**entity_kwargs)
            expected_output_names = list(entity.python_interface.outputs.keys())

            if isinstance(results, VoidPromise) or results is None:
                continue  # pragma: no cover # Move along, nothing to assign

            # Because we should've already returned in the above check, we just raise an Exception here.
            if len(entity.python_interface.outputs) == 0:
                raise FlyteValueException(results, "Interface output should've been VoidPromise or None.")

            # if there's only one output,
            if len(expected_output_names) == 1:
                if entity.python_interface.output_tuple_name and isinstance(results, tuple):
                    intermediate_node_outputs[node][expected_output_names[0]] = results[0]
                else:
                    intermediate_node_outputs[node][expected_output_names[0]] = results

            else:
                if len(results) != len(expected_output_names):
                    raise FlyteValueException(results, f"Different lengths {results} {expected_output_names}")
                for idx, r in enumerate(results):
                    intermediate_node_outputs[node][expected_output_names[idx]] = r

        # The rest of this function looks like the above but now we're doing it for the workflow as a whole rather
        # than just one node at a time.
        if len(self.python_interface.outputs) == 0:
            return VoidPromise(self.name)

        # The values that we return below from the output have to be pulled by fulfilling all of the
        # workflow's output bindings.
        # The return style here has to match what 1) what the workflow would've returned had it been declared
        # functionally, and 2) what a user would return in mock function. That is, if it's a tuple, then it
        # should be a tuple here, if it's a one element named tuple, then we do a one-element non-named tuple,
        # if it's a single element then we return a single element
        if len(self.output_bindings) == 1:
            # Again use presence of output_tuple_name to understand that we're dealing with a one-element
            # named tuple
            if self.python_interface.output_tuple_name:
                return (get_promise(self.output_bindings[0].binding, intermediate_node_outputs),)
            # Just a normal single element
            return get_promise(self.output_bindings[0].binding, intermediate_node_outputs)
        return tuple([get_promise(b.binding, intermediate_node_outputs) for b in self.output_bindings])

    def create_conditional(self, name: str) -> ConditionalSection:
        ctx = FlyteContext.current_context()
        if ctx.compilation_state is not None:
            raise Exception("Can't already be compiling")
        FlyteContextManager.with_context(ctx.with_compilation_state(self.compilation_state))
        return conditional(name=name)

    def add_entity(self, entity: Union[PythonTask, _annotated_launch_plan.LaunchPlan, WorkflowBase], **kwargs) -> Node:
        """
        Anytime you add an entity, all the inputs to the entity must be bound.
        """
        # circular import
        from flytekit.core.node_creation import create_node

        ctx = FlyteContext.current_context()
        if ctx.compilation_state is not None:
            raise Exception("Can't already be compiling")
        with FlyteContextManager.with_context(ctx.with_compilation_state(self.compilation_state)) as ctx:
            n = create_node(entity=entity, **kwargs)

            def get_input_values(input_value):
                if isinstance(input_value, list):
                    input_promises = []
                    for x in input_value:
                        input_promises.extend(get_input_values(x))
                    return input_promises
                if isinstance(input_value, dict):
                    input_promises = []
                    for _, v in input_value.items():
                        input_promises.extend(get_input_values(v))
                    return input_promises
                else:
                    return [input_value]

            # Every time an entity is added, mark it as used. The above function though will gather all the input
            # values but we're only interested in the ones that are Promises so let's filter for those.
            # There's probably a way to clean this up, maybe key off of the name instead of value?
            all_input_values = get_input_values(kwargs)
            for input_value in filter(lambda x: isinstance(x, Promise), all_input_values):
                if input_value in self._unbound_inputs:
                    self._unbound_inputs.remove(input_value)
            return n  # type: ignore

    def add_workflow_input(self, input_name: str, python_type: Type) -> Promise:
        """
        Adds an input to the workflow.
        """
        if input_name in self._inputs:
            raise FlyteValidationException(f"Input {input_name} has already been specified for wf {self.name}.")
        self._python_interface = self._python_interface.with_inputs(extra_inputs={input_name: python_type})
        self._interface = transform_interface_to_typed_interface(self._python_interface)
        self._inputs[input_name] = Promise(var=input_name, val=NodeOutput(node=GLOBAL_START_NODE, var=input_name))
        self._unbound_inputs.add(self._inputs[input_name])
        return self._inputs[input_name]

    def add_workflow_output(
        self, output_name: str, p: Union[Promise, List[Promise], Dict[str, Promise]], python_type: Optional[Type] = None
    ):
        """
        Add an output with the given name from the given node output.
        """
        if output_name in self._python_interface.outputs:
            raise FlyteValidationException(f"Output {output_name} already exists in workflow {self.name}")

        if python_type is None:
            if type(p) == list or type(p) == dict:
                raise FlyteValidationException(
                    f"If specifying a list or dict of Promises, you must specify the python_type type for {output_name}"
                    f" starting with the container type (e.g. List[int]"
                )
            promise = cast(Promise, p)
            python_type = promise.ref.node.flyte_entity.python_interface.outputs[promise.var]
            logger.debug(f"Inferring python type for wf output {output_name} from Promise provided {python_type}")

        flyte_type = TypeEngine.to_literal_type(python_type=python_type)

        ctx = FlyteContext.current_context()
        if ctx.compilation_state is not None:
            raise Exception("Can't already be compiling")
        with FlyteContextManager.with_context(ctx.with_compilation_state(self.compilation_state)) as ctx:
            b, _ = binding_from_python_std(
                ctx, output_name, expected_literal_type=flyte_type, t_value=p, t_value_type=python_type
            )
            self._output_bindings.append(b)
            self._python_interface = self._python_interface.with_outputs(extra_outputs={output_name: python_type})
            self._interface = transform_interface_to_typed_interface(self._python_interface)

    def add_task(self, task: PythonTask, **kwargs) -> Node:
        return self.add_entity(task, **kwargs)

    def add_launch_plan(self, launch_plan: _annotated_launch_plan.LaunchPlan, **kwargs) -> Node:
        return self.add_entity(launch_plan, **kwargs)

    def add_subwf(self, sub_wf: WorkflowBase, **kwargs) -> Node:
        return self.add_entity(sub_wf, **kwargs)

    def ready(self) -> bool:
        """
        This function returns whether or not the workflow is in a ready state, which means
          * Has at least one node
          * All workflow inputs are bound

        These conditions assume that all nodes and workflow i/o changes were done with the functions above, which
        do additional checking.
        """
        if len(self.compilation_state.nodes) == 0:
            return False

        if len(self._unbound_inputs) > 0:
            return False

        return True


class PythonFunctionWorkflow(WorkflowBase, ClassStorageTaskResolver):
    """
    Please read :std:ref:`flyte:divedeep-workflows` first for a high-level understanding of what workflows are in Flyte.
    This Python object represents a workflow  defined by a function and decorated with the
    :py:func:`@workflow <flytekit.workflow>` decorator. Please see notes on that object for additional information.
    """

    def __init__(
        self,
        workflow_function: Callable,
        metadata: WorkflowMetadata,
        default_metadata: WorkflowMetadataDefaults,
        docstring: Optional[Docstring] = None,
        on_failure: Optional[Union[WorkflowBase, Task]] = None,
        docs: Optional[Documentation] = None,
    ):
        name, _, _, _ = extract_task_module(workflow_function)
        self._workflow_function = workflow_function
        native_interface = transform_function_to_interface(workflow_function, docstring=docstring)

        # TODO do we need this - can this not be in launchplan only?
        #    This can be in launch plan only, but is here only so that we don't have to re-evaluate. Or
        #    we can re-evaluate.
        self._input_parameters = None
        super().__init__(
            name=name,
            workflow_metadata=metadata,
            workflow_metadata_defaults=default_metadata,
            python_interface=native_interface,
            on_failure=on_failure,
            docs=docs,
        )
        self.compiled = False

    @property
    def function(self):
        return self._workflow_function

    def task_name(self, t: PythonAutoContainerTask) -> str:  # type: ignore
        return f"{self.name}.{t.__module__}.{t.name}"

    def _validate_add_on_failure_handler(self, ctx: FlyteContext, prefix: str, wf_args: Dict[str, Promise]):
        # Compare
        with FlyteContextManager.with_context(
            ctx.with_compilation_state(CompilationState(prefix=prefix, task_resolver=self))
        ) as inner_comp_ctx:
            # Now lets compile the failure-node if it exists
            if self.on_failure:
                c = wf_args.copy()
                exception_scopes.user_entry_point(self.on_failure)(**c)
                inner_nodes = None
                if inner_comp_ctx.compilation_state and inner_comp_ctx.compilation_state.nodes:
                    inner_nodes = inner_comp_ctx.compilation_state.nodes
                if not inner_nodes or len(inner_nodes) > 1:
                    raise AssertionError("Unable to compile failure node, only either a task or a workflow can be used")
                self._failure_node = inner_nodes[0]

    def compile(self, **kwargs):
        """
        Supply static Python native values in the kwargs if you want them to be used in the compilation. This mimics
        a 'closure' in the traditional sense of the word.
        """
        if self.compiled:
            return

        self.compiled = True
        ctx = FlyteContextManager.current_context()
        all_nodes = []
        prefix = ctx.compilation_state.prefix if ctx.compilation_state is not None else ""

        with FlyteContextManager.with_context(
            ctx.with_compilation_state(CompilationState(prefix=prefix, task_resolver=self))
        ) as comp_ctx:
            # Construct the default input promise bindings, but then override with the provided inputs, if any
            input_kwargs = construct_input_promises([k for k in self.interface.inputs.keys()])
            input_kwargs.update(kwargs)
            workflow_outputs = exception_scopes.user_entry_point(self._workflow_function)(**input_kwargs)
            all_nodes.extend(comp_ctx.compilation_state.nodes)

            # This little loop was added as part of the task resolver change. The task resolver interface itself is
            # more or less stateless (the future-proofing get_all_tasks function notwithstanding). However the
            # implementation of the TaskResolverMixin that this workflow class inherits from (ClassStorageTaskResolver)
            # does store state. This loop adds Tasks that are defined within the body of the workflow to the workflow
            # object itself.
            for n in comp_ctx.compilation_state.nodes:
                if isinstance(n.flyte_entity, PythonAutoContainerTask) and n.flyte_entity.task_resolver == self:
                    logger.debug(f"WF {self.name} saving task {n.flyte_entity.name}")
                    self.add(n.flyte_entity)

            self._validate_add_on_failure_handler(comp_ctx, comp_ctx.compilation_state.prefix + "f", input_kwargs)

        # Iterate through the workflow outputs
        bindings = []
        output_names = list(self.interface.outputs.keys())
        # The reason the length 1 case is separate is because the one output might be a list. We don't want to
        # iterate through the list here, instead we should let the binding creation unwrap it and make a binding
        # collection/map out of it.
        if len(output_names) == 1:
            if isinstance(workflow_outputs, tuple):
                if len(workflow_outputs) != 1:
                    raise AssertionError(
                        f"The Workflow specification indicates only one return value, received {len(workflow_outputs)}"
                    )
                if self.python_interface.output_tuple_name is None:
                    raise AssertionError(
                        "Outputs specification for Workflow does not define a tuple, but return value is a tuple"
                    )
                workflow_outputs = workflow_outputs[0]
            t = self.python_interface.outputs[output_names[0]]
            try:
                b, _ = binding_from_python_std(
                    ctx,
                    output_names[0],
                    self.interface.outputs[output_names[0]].type,
                    workflow_outputs,
                    t,
                )
                bindings.append(b)
            except Exception as e:
                raise FlyteValidationException(
                    f"Failed to bind output {output_names[0]} for function {self.name}: {e}"
                ) from e
        elif len(output_names) > 1:
            if not isinstance(workflow_outputs, tuple):
                raise AssertionError("The Workflow specification indicates multiple return values, received only one")
            if len(output_names) != len(workflow_outputs):
                raise Exception(f"Length mismatch {len(output_names)} vs {len(workflow_outputs)}")
            for i, out in enumerate(output_names):
                if isinstance(workflow_outputs[i], ConditionalSection):
                    raise AssertionError("A Conditional block (if-else) should always end with an `else_()` clause")
                t = self.python_interface.outputs[out]
                try:
                    b, _ = binding_from_python_std(
                        ctx,
                        out,
                        self.interface.outputs[out].type,
                        workflow_outputs[i],
                        t,
                    )
                    bindings.append(b)
                except Exception as e:
                    raise FlyteValidationException(f"Failed to bind output {out} for function {self.name}: {e}") from e

        # Save all the things necessary to create an WorkflowTemplate, except for the missing project and domain
        self._nodes = all_nodes
        self._output_bindings = bindings

        if not output_names:
            return None
        if len(output_names) == 1:
            return bindings[0]
        return tuple(bindings)

    def execute(self, **kwargs):
        """
        This function is here only to try to streamline the pattern between workflows and tasks. Since tasks
        call execute from dispatch_execute which is in local_execute, workflows should also call an execute inside
        local_execute. This makes mocking cleaner.
        """
        return exception_scopes.user_entry_point(self._workflow_function)(**kwargs)


@overload
def workflow(
    _workflow_function: None = ...,
    failure_policy: Optional[WorkflowFailurePolicy] = ...,
    interruptible: bool = ...,
    on_failure: Optional[Union[WorkflowBase, Task]] = ...,
    docs: Optional[Documentation] = ...,
) -> Callable[[Callable[..., FuncOut]], PythonFunctionWorkflow]:
    ...


@overload
def workflow(
    _workflow_function: Callable[..., FuncOut],
    failure_policy: Optional[WorkflowFailurePolicy] = ...,
    interruptible: bool = ...,
    on_failure: Optional[Union[WorkflowBase, Task]] = ...,
    docs: Optional[Documentation] = ...,
) -> Union[PythonFunctionWorkflow, Callable[..., FuncOut]]:
    ...


def workflow(
    _workflow_function: Optional[Callable[..., Any]] = None,
    failure_policy: Optional[WorkflowFailurePolicy] = None,
    interruptible: bool = False,
    on_failure: Optional[Union[WorkflowBase, Task]] = None,
    docs: Optional[Documentation] = None,
) -> Union[Callable[[Callable[..., FuncOut]], PythonFunctionWorkflow], PythonFunctionWorkflow, Callable[..., FuncOut]]:
    """
    This decorator declares a function to be a Flyte workflow. Workflows are declarative entities that construct a DAG
    of tasks using the data flow between tasks.

    Unlike a task, the function body of a workflow is evaluated at serialization-time (aka compile-time). This is
    because while we can determine the entire structure of a task by looking at the function's signature, workflows need
    to run through the function itself because the body of the function is what expresses the workflow structure. It's
    also important to note that, local execution notwithstanding, it is not evaluated again when the workflow runs on
    Flyte.
    That is, workflows should not call non-Flyte entities since they are only run once (again, this is with respect to
    the platform, local runs notwithstanding).

    Example:

    .. literalinclude:: ../../../tests/flytekit/unit/core/test_workflows.py
       :pyobject: my_wf_example

    Again, users should keep in mind that even though the body of the function looks like regular Python, it is
    actually not. When flytekit scans the workflow function, the objects being passed around between the tasks are not
    your typical Python values. So even though you may have a task ``t1() -> int``, when ``a = t1()`` is called, ``a``
    will not be an integer so if you try to ``range(a)`` you'll get an error.

    Please see the :ref:`user guide <cookbook:workflow>` for more usage examples.

    :param _workflow_function: This argument is implicitly passed and represents the decorated function.
    :param failure_policy: Use the options in flytekit.WorkflowFailurePolicy
    :param interruptible: Whether or not tasks launched from this workflow are by default interruptible
    :param on_failure: Invoke this workflow or task on failure. The Workflow / task has to match the signature of
         the current workflow, with an additional parameter called `error` Error
    :param docs: Description entity for the workflow
    """

    def wrapper(fn: Callable[..., Any]) -> PythonFunctionWorkflow:
        workflow_metadata = WorkflowMetadata(on_failure=failure_policy or WorkflowFailurePolicy.FAIL_IMMEDIATELY)

        workflow_metadata_defaults = WorkflowMetadataDefaults(interruptible)

        workflow_instance = PythonFunctionWorkflow(
            fn,
            metadata=workflow_metadata,
            default_metadata=workflow_metadata_defaults,
            docstring=Docstring(callable_=fn),
            on_failure=on_failure,
            docs=docs,
        )
        update_wrapper(workflow_instance, fn)
        return workflow_instance

    if _workflow_function is not None:
        return wrapper(_workflow_function)
    else:
        return wrapper


class ReferenceWorkflow(ReferenceEntity, PythonFunctionWorkflow):  # type: ignore
    """
    A reference workflow is a pointer to a workflow that already exists on your Flyte installation. This
    object will not initiate a network call to Admin, which is why the user is asked to provide the expected interface.
    If at registration time the interface provided causes an issue with compilation, an error will be returned.
    """

    def __init__(
        self, project: str, domain: str, name: str, version: str, inputs: Dict[str, Type], outputs: Dict[str, Type]
    ):
        super().__init__(WorkflowReference(project, domain, name, version), inputs, outputs)


def reference_workflow(
    project: str,
    domain: str,
    name: str,
    version: str,
) -> Callable[[Callable[..., Any]], ReferenceWorkflow]:
    """
    A reference workflow is a pointer to a workflow that already exists on your Flyte installation. This
    object will not initiate a network call to Admin, which is why the user is asked to provide the expected interface.
    If at registration time the interface provided causes an issue with compilation, an error will be returned.

    Example:

    .. literalinclude:: ../../../tests/flytekit/unit/core/test_references.py
       :pyobject: ref_wf1
    """

    def wrapper(fn) -> ReferenceWorkflow:
        interface = transform_function_to_interface(fn)
        return ReferenceWorkflow(project, domain, name, version, interface.inputs, interface.outputs)

    return wrapper
