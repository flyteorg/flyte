from __future__ import annotations

from typing import TYPE_CHECKING, Union

from flytekit.core.base_task import PythonTask
from flytekit.core.context_manager import BranchEvalMode, FlyteContext
from flytekit.core.launch_plan import LaunchPlan
from flytekit.core.node import Node
from flytekit.core.promise import VoidPromise
from flytekit.core.workflow import WorkflowBase
from flytekit.exceptions import user as _user_exceptions
from flytekit.loggers import logger

if TYPE_CHECKING:
    from flytekit.remote.remote_callable import RemoteEntity


# This file exists instead of moving to node.py because it needs Task/Workflow/LaunchPlan and those depend on Node


def create_node(
    entity: Union[PythonTask, LaunchPlan, WorkflowBase, RemoteEntity], *args, **kwargs
) -> Union[Node, VoidPromise]:
    """
    This is the function you want to call if you need to specify dependencies between tasks that don't consume and/or
    don't produce outputs. For example, if you have t1() and t2(), both of which do not take in nor produce any
    outputs, how do you specify that t2 should run before t1? ::

        t1_node = create_node(t1)
        t2_node = create_node(t2)

        t2_node.runs_before(t1_node)
        # OR
        t2_node >> t1_node

    This works for tasks that take inputs as well, say a ``t3(in1: int)`` ::

        t3_node = create_node(t3, in1=some_int)  # basically calling t3(in1=some_int)

    You can still use this method to handle setting certain overrides ::

        t3_node = create_node(t3, in1=some_int).with_overrides(...)

    Outputs, if there are any, will be accessible. A `t4() -> (int, str)` ::

        t4_node = create_node(t4)

    In compilation node.o0 has the promise. ::
        t5(in1=t4_node.o0)

    If t1 produces only one output, note that in local execution, you still get a wrapper object that
    needs to be dereferenced by the output name. ::

        t1_node = create_node(t1)
        t2(t1_node.o0)

    """
    from flytekit.remote.remote_callable import RemoteEntity

    if len(args) > 0:
        raise _user_exceptions.FlyteAssertion(
            f"Only keyword args are supported to pass inputs to workflows and tasks."
            f"Aborting execution as detected {len(args)} positional args {args}"
        )

    if (
        not isinstance(entity, PythonTask)
        and not isinstance(entity, WorkflowBase)
        and not isinstance(entity, LaunchPlan)
        and not isinstance(entity, RemoteEntity)
    ):
        raise AssertionError(f"Should be a callable Flyte entity (either local or fetched) but is {type(entity)}")

    # This function is only called from inside workflows and dynamic tasks.
    # That means there are two scenarios we need to take care of, compilation and local workflow execution.

    # When compiling, calling the entity will create a node.
    ctx = FlyteContext.current_context()
    if ctx.compilation_state is not None and ctx.compilation_state.mode == 1:
        outputs = entity(**kwargs)
        # This is always the output of create_and_link_node which returns create_task_output, which can be
        # VoidPromise, Promise, or our custom namedtuple of Promises.
        node = ctx.compilation_state.nodes[-1]

        # In addition to storing the outputs on the object itself, we also want to set them in a map. When used by
        # the imperative workflow patterns, users will probably find themselves doing things like
        #   n = create_node(...)  # then
        #   output_name = "o0"
        #   n.outputs[output_name]  # rather than
        #   n.o0
        # That is, they'll likely have the name of the output stored as a string variable, and dicts provide cleaner
        # access than getattr
        node._outputs = {}

        # If a VoidPromise, just return the node.
        if isinstance(outputs, VoidPromise):
            return node

        # If a Promise or custom namedtuple of Promises, we need to attach each output as an attribute to the node.
        # todo: fix the noqas below somehow... can't add abstract property to RemoteEntity because it has to come
        #  before the model Template classes in FlyteTask/Workflow/LaunchPlan
        if entity.interface.outputs:  # noqa
            if isinstance(outputs, tuple):
                for output_name in entity.interface.outputs.keys():  # noqa
                    attr = getattr(outputs, output_name)
                    if attr is None:
                        raise _user_exceptions.FlyteAssertion(
                            f"Output {output_name} in outputs when calling {entity.name} is empty {attr}."
                        )
                    if hasattr(node, output_name):
                        raise _user_exceptions.FlyteAssertion(
                            f"Node {node} already has attribute {output_name}, change the name of output."
                        )
                    setattr(node, output_name, attr)
                    node.outputs[output_name] = attr
            else:
                output_names = [k for k in entity.interface.outputs.keys()]  # noqa
                if len(output_names) != 1:
                    raise _user_exceptions.FlyteAssertion(f"Output of length 1 expected but {len(output_names)} found")

                if hasattr(node, output_names[0]):
                    raise _user_exceptions.FlyteAssertion(
                        f"Node {node} already has attribute {output_names[0]}, change the name of output."
                    )

                setattr(node, output_names[0], outputs)  # This should be a singular Promise
                node.outputs[output_names[0]] = outputs

        return node

    # Handling local execution
    # Note: execution state is set to TASK_EXECUTION when running dynamic task locally
    # https://github.com/flyteorg/flytekit/blob/0815345faf0fae5dc26746a43d4bda4cc2cdf830/flytekit/core/python_function_task.py#L262
    elif ctx.execution_state and ctx.execution_state.is_local_execution():
        if isinstance(entity, RemoteEntity):
            raise AssertionError(f"Remote entities are not yet runnable locally {entity.name}")

        if ctx.execution_state.branch_eval_mode == BranchEvalMode.BRANCH_SKIPPED:
            logger.warning(f"Manual node creation cannot be used in branch logic {entity.name}")
            raise Exception("Being more restrictive for now and disallowing manual node creation in branch logic")

        # This the output of __call__ under local execute conditions which means this is the output of local_execute
        # which means this is the output of create_task_output with Promises containing values (or a VoidPromise)
        results = entity(**kwargs)

        # If it's a VoidPromise, let's just return it, it shouldn't get used anywhere and if it does, we want an error
        # The reason we return it if it's a tuple is to handle the case where the task returns a typing.NamedTuple.
        # In that case, it's already a tuple and we don't need to further tupletize.
        if isinstance(results, VoidPromise) or isinstance(results, tuple):
            return results  # type: ignore

        output_names = entity.python_interface.output_names  # type: ignore

        if not output_names:
            raise Exception(f"Non-VoidPromise received {results} but interface for {entity.name} doesn't have outputs")

        if len(output_names) == 1:
            # See explanation above for why we still tupletize a single element.
            return entity.python_interface.output_tuple(results)  # type: ignore

        return entity.python_interface.output_tuple(*results)  # type: ignore

    else:
        raise Exception(f"Cannot use explicit run to call Flyte entities {entity.name}")
