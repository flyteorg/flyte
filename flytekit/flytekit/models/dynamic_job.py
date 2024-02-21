from flyteidl.core import dynamic_job_pb2 as _dynamic_job

from flytekit.models import common as _common
from flytekit.models import literals as _literals
from flytekit.models import task as _task
from flytekit.models.core import workflow as _workflow


class DynamicJobSpec(_common.FlyteIdlEntity):
    def __init__(self, tasks, nodes, min_successes, outputs, subworkflows):
        """
        Initializes a new FutureTaskDocument.

        :param list[flytekit.models.task.TaskTemplate] tasks: A collection of unique tasks to execute.
        :param list[flytekit.models.core.workflow.Node] nodes: A collection of task nodes.
        :param int min_successes: An absolute number of the minimum number of successful completions of subtasks. As
            soon as this criteria is met, the future job will be marked as successful and outputs will be computed.
        :param list[flytekit.models.literals.Binding] outputs: Describes how to bind the final output of the future
            task from the
            outputs of executed nodes. The referenced ids in bindings should have the generated id for the subtask.
        :param list[flytekit.models.workflow.WorkflowTemplate] subworkflows: A complete list of task specs referenced
            in nodes.
        """

        self._tasks = tasks
        self._nodes = nodes
        self._min_successes = min_successes
        self._outputs = outputs
        self._subworkflows = subworkflows

    @property
    def tasks(self):
        """
        A collection of tasks to execute.
        :rtype: list[_task.TaskTemplate]
        """
        return self._tasks

    @property
    def nodes(self):
        """
        A collection of dynamic nodes.
        :rtype: list[_workflow.Node]
        """
        return self._nodes

    @property
    def min_successes(self):
        """
        An absolute number of the minimum number of successful completions of subtasks. As
            soon as this criteria is met, the future job will be marked as successful and outputs will be computed.
        :rtype: int
        """
        return self._min_successes

    @property
    def outputs(self):
        """
        Describes how to bind the final output of the future task from the outputs of executed nodes.
            The referenced ids in bindings should have the generated id for the subtask.
        :rtype: list[flytekit.models.literals.Binding]
        """
        return self._outputs

    @property
    def subworkflows(self):
        """
        A collection of subworkflows to execute.
        :rtype: list[flytekit.models.core.workflow.WorkflowTemplate]
        """
        return self._subworkflows

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.dynamic_job.DynamicJobSpec
        """
        return _dynamic_job.DynamicJobSpec(
            tasks=[task.to_flyte_idl() for task in self.tasks] if self.tasks else None,
            nodes=[node.to_flyte_idl() for node in self.nodes] if self.nodes else None,
            min_successes=self.min_successes,
            outputs=[output.to_flyte_idl() for output in self.outputs],
            subworkflows=[workflow.to_flyte_idl() for workflow in self.subworkflows],
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.dynamic_job_pb2.DynamicJobSpec pb2_object:
        :return: DynamicJobSpec
        """
        return cls(
            tasks=[_task.TaskTemplate.from_flyte_idl(task) for task in pb2_object.tasks] if pb2_object.tasks else None,
            nodes=[_workflow.Node.from_flyte_idl(n) for n in pb2_object.nodes],
            min_successes=pb2_object.min_successes,
            outputs=[_literals.Binding.from_flyte_idl(output) for output in pb2_object.outputs]
            if pb2_object.outputs
            else None,
            subworkflows=[_workflow.WorkflowTemplate.from_flyte_idl(w) for w in pb2_object.subworkflows],
        )
