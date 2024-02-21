from abc import ABC, abstractmethod
from typing import Any, Dict, Optional, Tuple, Type, Union

from flytekit.core.context_manager import BranchEvalMode, ExecutionState, FlyteContext
from flytekit.core.promise import Promise, VoidPromise, create_and_link_node_from_remote, extract_obj_name
from flytekit.exceptions import user as user_exceptions
from flytekit.loggers import logger
from flytekit.models.core.workflow import NodeMetadata


class RemoteEntity(ABC):
    def __init__(self, *args, **kwargs):
        # In cases where we make a FlyteTask/Workflow/LaunchPlan from a locally created Python object (i.e. an @task
        # or an @workflow decorated function), we actually have the Python interface, so
        self._python_interface: Optional[Dict[str, Type]] = None

        super().__init__(*args, **kwargs)

    @property
    @abstractmethod
    def name(self) -> str:
        ...

    def construct_node_metadata(self) -> NodeMetadata:
        """
        Used when constructing the node that encapsulates this task as part of a broader workflow definition.
        """
        return NodeMetadata(
            name=extract_obj_name(self.name),
        )

    def compile(self, ctx: FlyteContext, *args, **kwargs):
        return create_and_link_node_from_remote(ctx, entity=self, **kwargs)  # noqa

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
            raise user_exceptions.FlyteAssertion(
                f"Cannot call remotely fetched entity with args - detected {len(args)} positional args {args}"
            )

        ctx = FlyteContext.current_context()
        if ctx.compilation_state is not None and ctx.compilation_state.mode == 1:
            return self.compile(ctx, *args, **kwargs)
        elif (
            ctx.execution_state is not None and ctx.execution_state.mode == ExecutionState.Mode.LOCAL_WORKFLOW_EXECUTION
        ):
            if ctx.execution_state.branch_eval_mode == BranchEvalMode.BRANCH_SKIPPED:
                return
            return self.local_execute(ctx, **kwargs)
        else:
            logger.debug("Fetched entity, running raw execute.")
            return self.execute(**kwargs)

    def local_execute(self, ctx: FlyteContext, **kwargs) -> Optional[Union[Tuple[Promise], Promise, VoidPromise]]:
        return self.execute(**kwargs)

    def local_execution_mode(self) -> ExecutionState.Mode:
        return ExecutionState.Mode.LOCAL_TASK_EXECUTION

    def execute(self, **kwargs) -> Any:
        raise AssertionError(f"Remotely fetched entities cannot be run locally. Please mock the {self.name}.execute.")

    @property
    def python_interface(self) -> Optional[Dict[str, Type]]:
        return self._python_interface
