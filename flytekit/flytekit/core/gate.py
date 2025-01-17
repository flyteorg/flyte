from __future__ import annotations

import datetime
import typing
from typing import Tuple, Union

import click

from flytekit.core import constants
from flytekit.core import interface as flyte_interface
from flytekit.core.context_manager import ExecutionState, FlyteContext, FlyteContextManager
from flytekit.core.promise import Promise, VoidPromise, flyte_entity_call_handler
from flytekit.core.type_engine import TypeEngine
from flytekit.exceptions.user import FlyteDisapprovalException
from flytekit.interaction.parse_stdin import parse_stdin_to_literal
from flytekit.interaction.string_literals import scalar_to_string
from flytekit.models.core import workflow as _workflow_model
from flytekit.models.literals import Scalar
from flytekit.models.types import LiteralType

DEFAULT_TIMEOUT = datetime.timedelta(hours=1)


class Gate(object):
    """
    A node type that waits for user input before proceeding with a workflow.
    A gate is a type of node that behaves like a task, but instead of running code, it either needs to wait
    for user input to proceed or wait for a timer to complete running.
    """

    def __init__(
        self,
        name: str,
        input_type: typing.Optional[typing.Type] = None,
        upstream_item: typing.Optional[typing.Any] = None,
        sleep_duration: typing.Optional[datetime.timedelta] = None,
        timeout: typing.Optional[datetime.timedelta] = None,
    ):
        self._name = name
        self._input_type = input_type
        self._sleep_duration = sleep_duration
        self._timeout = timeout or DEFAULT_TIMEOUT
        self._upstream_item = upstream_item
        self._literal_type = TypeEngine.to_literal_type(input_type) if input_type else None

        # Determine the python interface if we can
        if self._sleep_duration:
            # Just a sleep so there is no interface
            self._python_interface = flyte_interface.Interface()
        elif input_type:
            # Waiting for user input, so the output of the node is whatever input the user provides.
            self._python_interface = flyte_interface.Interface(
                outputs={
                    "o0": self.input_type,
                }
            )
        else:
            # We don't know how to find the python interface here, approve() sets it below, See the code.
            self._python_interface = None  # type: ignore

    @property
    def name(self) -> str:
        # Part of SupportsNodeCreation interface
        return self._name

    @property
    def input_type(self) -> typing.Optional[typing.Type]:
        return self._input_type

    @property
    def literal_type(self) -> typing.Optional[LiteralType]:
        return self._literal_type

    @property
    def sleep_duration(self) -> typing.Optional[datetime.timedelta]:
        return self._sleep_duration

    @property
    def python_interface(self) -> flyte_interface.Interface:
        """
        This will not be valid during local execution
        Part of SupportsNodeCreation interface
        """
        # If this is just a sleep node, or user input node, then it will have a Python interface upon construction.
        if self._python_interface:
            return self._python_interface

        raise ValueError("You can't check for a Python interface for an approval node outside of compilation")

    def construct_node_metadata(self) -> _workflow_model.NodeMetadata:
        # Part of SupportsNodeCreation interface
        return _workflow_model.NodeMetadata(
            name=self.name,
            timeout=self._timeout,
        )

    # This is to satisfy the LocallyExecutable protocol
    def local_execute(self, ctx: FlyteContext, **kwargs) -> Union[Tuple[Promise], Promise, VoidPromise]:
        if self.sleep_duration:
            click.echo(
                f'{click.style("[Sleep Gate]", fg="yellow")} '
                f'{click.style(f"Simulating Sleep for {self.sleep_duration}", fg="cyan")}'
            )
            return VoidPromise(self.name)

        # Trigger stdin
        if self.input_type:
            msg = click.style("[Input Gate] ", fg="yellow") + click.style(
                f"Waiting for input @{self.name} of type {self.input_type}", fg="cyan"
            )
            literal = parse_stdin_to_literal(ctx, self.input_type, msg)
            p = Promise(var="o0", val=literal)
            return p

        # Assume this is an approval operation since that's the only remaining option.
        v = typing.cast(Promise, self._upstream_item).val.value
        if isinstance(v, Scalar):
            v = scalar_to_string(v)
        msg = click.style("[Approval Gate] ", fg="yellow") + click.style(
            f"@{self.name} Approve {click.style(v, fg='green')}?", fg="cyan"
        )
        proceed = click.confirm(msg, default=True)
        if proceed:
            # We need to return a promise here, and a promise is what should've been passed in by the call in approve()
            # Only one element should be in this map. Rely on kwargs instead of the stored _upstream_item even though
            # they should be the same to be cleaner
            output_name = list(kwargs.keys())[0]
            return kwargs[output_name]
        else:
            raise FlyteDisapprovalException(f"User did not approve the transaction for gate node {self.name}")

    def local_execution_mode(self):
        return ExecutionState.Mode.LOCAL_TASK_EXECUTION


def wait_for_input(name: str, timeout: datetime.timedelta, expected_type: typing.Type):
    """Create a Gate object that waits for user input of the specified type.

    Create a Gate object. This object will function like a task. Note that unlike a task,
    each time this function is called, a new Python object is created. If a workflow
    calls a subworkflow twice, and the subworkflow has a signal, then two Gate
    objects are created. This shouldn't be a problem as long as the objects are identical.

    :param name: The name of the gate node.
    :param timeout: How long to wait for before Flyte fails the workflow.
    :param expected_type: What is the type that the user will be inputting?
    :return:
    """

    g = Gate(name, input_type=expected_type, timeout=timeout)

    return flyte_entity_call_handler(g)


def sleep(duration: datetime.timedelta):
    """Create a sleep Gate object.

    :param duration: How long to sleep for
    :return:
    """
    g = Gate("sleep-gate", sleep_duration=duration)

    return flyte_entity_call_handler(g)


def approve(upstream_item: Union[Tuple[Promise], Promise, VoidPromise], name: str, timeout: datetime.timedelta):
    """Create a Gate object for binary approval.

    Create a Gate object. This object will function like a task. Note that unlike a task,
    each time this function is called, a new Python object is created. If a workflow
    calls a subworkflow twice, and the subworkflow has a signal, then two Gate
    objects are created. This shouldn't be a problem as long as the objects are identical.

    :param upstream_item: This should be the output, one output, of a previous task, that you want to gate execution
      on. This is the value that you want a human to check before moving on.
    :param name: The name of the gate node.
    :param timeout: How long to wait before Flyte fails the workflow.
    :return:
    """
    g = Gate(name, upstream_item=upstream_item, timeout=timeout)

    if upstream_item is None or isinstance(upstream_item, VoidPromise):
        raise ValueError("You can't use approval on a task that doesn't return anything.")

    ctx = FlyteContextManager.current_context()
    upstream_item = typing.cast(Promise, upstream_item)
    if ctx.compilation_state is not None and ctx.compilation_state.mode == 1:
        if upstream_item.ref.node_id == constants.GLOBAL_INPUT_NODE_ID:
            raise ValueError("Workflow inputs cannot be passed to approval nodes.")
        if not upstream_item.ref.node.flyte_entity.python_interface:
            raise ValueError(
                f"Upstream node doesn't have a Python interface. Node entity is: "
                f"{upstream_item.ref.node.flyte_entity}"
            )

        # We have reach back up to the entity that this promise came from, to get the python type, since
        # the approve function itself doesn't have a python interface.
        io_type = upstream_item.ref.node.flyte_entity.python_interface.outputs[upstream_item.var]
        io_var_name = upstream_item.var
    else:
        # We don't know the python type here. in local execution, downstream doesn't really use the type
        # so we should be okay. But use None instead of type() so that errors are more obvious hopefully.
        io_type = None
        io_var_name = "o0"

    # In either case, we need a python interface
    g._python_interface = flyte_interface.Interface(
        inputs={
            io_var_name: io_type,
        },
        outputs={
            io_var_name: io_type,
        },
    )
    kwargs = {io_var_name: upstream_item}

    return flyte_entity_call_handler(g, **kwargs)
