from __future__ import annotations

from abc import abstractmethod
from typing import Dict, List, Optional, Union

from flytekit.core.type_engine import LiteralsResolver
from flytekit.exceptions import user as user_exceptions
from flytekit.models import execution as execution_models
from flytekit.models import node_execution as node_execution_models
from flytekit.models.admin import task_execution as admin_task_execution_models
from flytekit.models.core import execution as core_execution_models
from flytekit.remote.entities import FlyteTask, FlyteWorkflow


class RemoteExecutionBase(object):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._inputs: Optional[LiteralsResolver] = None
        self._outputs: Optional[LiteralsResolver] = None

    @property
    def inputs(self) -> Optional[LiteralsResolver]:
        return self._inputs

    @property
    @abstractmethod
    def error(self) -> core_execution_models.ExecutionError:
        ...

    @property
    @abstractmethod
    def is_done(self) -> bool:
        ...

    @property
    def outputs(self) -> Optional[LiteralsResolver]:
        """
        :return: Returns the outputs LiteralsResolver to the execution
        :raises: ``FlyteAssertion`` error if execution is in progress or execution ended in error.
        """
        if not self.is_done:
            raise user_exceptions.FlyteAssertion(
                "Please wait until the execution has completed before requesting the outputs."
            )
        if self.error:
            raise user_exceptions.FlyteAssertion("Outputs could not be found because the execution ended in failure.")

        return self._outputs


class FlyteTaskExecution(RemoteExecutionBase, admin_task_execution_models.TaskExecution):
    """A class encapsulating a task execution being run on a Flyte remote backend."""

    def __init__(self, *args, **kwargs):
        super(FlyteTaskExecution, self).__init__(*args, **kwargs)
        self._flyte_task = None

    @property
    def task(self) -> Optional[FlyteTask]:
        return self._flyte_task

    @property
    def is_done(self) -> bool:
        """Whether or not the execution is complete."""
        return self.closure.phase in {
            core_execution_models.TaskExecutionPhase.ABORTED,
            core_execution_models.TaskExecutionPhase.FAILED,
            core_execution_models.TaskExecutionPhase.SUCCEEDED,
        }

    @property
    def error(self) -> Optional[core_execution_models.ExecutionError]:
        """
        If execution is in progress, raise an exception. Otherwise, return None if no error was present upon
        reaching completion.
        """
        if not self.is_done:
            raise user_exceptions.FlyteAssertion(
                "Please what until the task execution has completed before requesting error information."
            )
        return self.closure.error

    @classmethod
    def promote_from_model(cls, base_model: admin_task_execution_models.TaskExecution) -> "FlyteTaskExecution":
        return cls(
            closure=base_model.closure,
            id=base_model.id,
            input_uri=base_model.input_uri,
            is_parent=base_model.is_parent,
        )


class FlyteWorkflowExecution(RemoteExecutionBase, execution_models.Execution):
    """A class encapsulating a workflow execution being run on a Flyte remote backend."""

    def __init__(self, *args, **kwargs):
        super(FlyteWorkflowExecution, self).__init__(*args, **kwargs)
        self._node_executions = None
        self._flyte_workflow: Optional[FlyteWorkflow] = None

    @property
    def flyte_workflow(self) -> Optional[FlyteWorkflow]:
        return self._flyte_workflow

    @property
    def node_executions(self) -> Dict[str, "FlyteNodeExecution"]:
        """Get a dictionary of node executions that are a part of this workflow execution."""
        return self._node_executions or {}

    @property
    def error(self) -> core_execution_models.ExecutionError:
        """
        If execution is in progress, raise an exception.  Otherwise, return None if no error was present upon
        reaching completion.
        """
        if not self.is_done:
            raise user_exceptions.FlyteAssertion(
                "Please wait until a workflow has completed before checking for an error."
            )
        return self.closure.error

    @property
    def is_done(self) -> bool:
        """
        Whether or not the execution is complete.
        """
        return self.closure.phase in {
            core_execution_models.WorkflowExecutionPhase.ABORTED,
            core_execution_models.WorkflowExecutionPhase.FAILED,
            core_execution_models.WorkflowExecutionPhase.SUCCEEDED,
            core_execution_models.WorkflowExecutionPhase.TIMED_OUT,
        }

    @classmethod
    def promote_from_model(cls, base_model: execution_models.Execution) -> "FlyteWorkflowExecution":
        return cls(
            closure=base_model.closure,
            id=base_model.id,
            spec=base_model.spec,
        )


class FlyteNodeExecution(RemoteExecutionBase, node_execution_models.NodeExecution):
    """A class encapsulating a node execution being run on a Flyte remote backend."""

    def __init__(self, *args, **kwargs):
        super(FlyteNodeExecution, self).__init__(*args, **kwargs)
        self._task_executions = None
        self._workflow_executions = []
        self._underlying_node_executions = None
        self._interface = None
        self._flyte_node = None

    @property
    def task_executions(self) -> List[FlyteTaskExecution]:
        return self._task_executions or []

    @property
    def workflow_executions(self) -> List[FlyteWorkflowExecution]:
        return self._workflow_executions

    @property
    def subworkflow_node_executions(self) -> Dict[str, FlyteNodeExecution]:
        """
        This returns underlying node executions in instances where the current node execution is
        a parent node. This happens when it's either a static or dynamic subworkflow.
        """
        return (
            {}
            if self._underlying_node_executions is None
            else {n.id.node_id: n for n in self._underlying_node_executions}
        )

    @property
    def executions(self) -> List[Union[FlyteTaskExecution, FlyteWorkflowExecution]]:
        return self.task_executions or self._underlying_node_executions or []

    @property
    def error(self) -> core_execution_models.ExecutionError:
        """
        If execution is in progress, raise an exception. Otherwise, return None if no error was present upon
        reaching completion.
        """
        if not self.is_done:
            raise user_exceptions.FlyteAssertion(
                "Please wait until the node execution has completed before requesting error information."
            )
        return self.closure.error

    @property
    def is_done(self) -> bool:
        """Whether or not the execution is complete."""
        return self.closure.phase in {
            core_execution_models.NodeExecutionPhase.ABORTED,
            core_execution_models.NodeExecutionPhase.FAILED,
            core_execution_models.NodeExecutionPhase.SKIPPED,
            core_execution_models.NodeExecutionPhase.SUCCEEDED,
            core_execution_models.NodeExecutionPhase.TIMED_OUT,
        }

    @classmethod
    def promote_from_model(cls, base_model: node_execution_models.NodeExecution) -> "FlyteNodeExecution":
        return cls(
            closure=base_model.closure, id=base_model.id, input_uri=base_model.input_uri, metadata=base_model.metadata
        )

    @property
    def interface(self) -> "flytekit.remote.interface.TypedInterface":
        """
        Return the interface of the task or subworkflow associated with this node execution.
        """
        return self._interface
