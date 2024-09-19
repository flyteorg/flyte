from flyteidl.core import workflow_closure_pb2 as _workflow_closure_pb2

from flytekit.models import common as _common
from flytekit.models import task as _task_models
from flytekit.models.core import workflow as _core_workflow_models


class WorkflowClosure(_common.FlyteIdlEntity):
    def __init__(self, workflow, tasks=None):
        """
        :param flytekit.models.core.workflow.WorkflowTemplate workflow: Workflow template
        :param list[flytekit.models.task.TaskTemplate] tasks: [Optional]
        """
        self._workflow = workflow
        self._tasks = tasks

    @property
    def workflow(self):
        """
        :rtype: flytekit.models.core.workflow.WorkflowTemplate
        """
        return self._workflow

    @property
    def tasks(self):
        """
        :rtype: list[flytekit.models.task.TaskTemplate]
        """
        return self._tasks

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_closure_pb2.WorkflowClosure
        """
        return _workflow_closure_pb2.WorkflowClosure(
            workflow=self.workflow.to_flyte_idl(),
            tasks=[t.to_flyte_idl() for t in self.tasks],
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.workflow_closure_pb2.WorkflowClosure pb2_object
        :rtype: WorkflowClosure
        """
        return cls(
            workflow=_core_workflow_models.WorkflowTemplate.from_flyte_idl(pb2_object.workflow),
            tasks=[_task_models.TaskTemplate.from_flyte_idl(t) for t in pb2_object.tasks],
        )
