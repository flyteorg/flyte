import typing

from flyteidl.admin import workflow_pb2 as _admin_workflow

from flytekit.models import common as _common
from flytekit.models.core import compiler as _compiler_models
from flytekit.models.core import identifier as _identifier
from flytekit.models.core import workflow as _core_workflow
from flytekit.models.documentation import Documentation


class WorkflowSpec(_common.FlyteIdlEntity):
    def __init__(
        self,
        template: _core_workflow.WorkflowTemplate,
        sub_workflows: typing.List[_core_workflow.WorkflowTemplate],
        docs: typing.Optional[Documentation] = None,
    ):
        """
        This object fully encapsulates the specification of a workflow
        :param flytekit.models.core.workflow.WorkflowTemplate template:
        :param list[flytekit.models.core.workflow.WorkflowTemplate] sub_workflows:
        """
        self._template = template
        self._sub_workflows = sub_workflows
        self._docs = docs

    @property
    def template(self):
        """
        :rtype: flytekit.models.core.workflow.WorkflowTemplate
        """
        return self._template

    @property
    def sub_workflows(self):
        """
        :rtype: list[flytekit.models.core.workflow.WorkflowTemplate]
        """
        return self._sub_workflows

    @property
    def docs(self):
        """
        :rtype: Description entity for the workflow
        """
        return self._docs

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.workflow_pb2.WorkflowSpec
        """
        return _admin_workflow.WorkflowSpec(
            template=self._template.to_flyte_idl(),
            sub_workflows=[s.to_flyte_idl() for s in self._sub_workflows],
            description=self._docs.to_flyte_idl() if self._docs else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param pb2_object: flyteidl.admin.workflow_pb2.WorkflowSpec
        :rtype: WorkflowSpec
        """
        return cls(
            _core_workflow.WorkflowTemplate.from_flyte_idl(pb2_object.template),
            [_core_workflow.WorkflowTemplate.from_flyte_idl(s) for s in pb2_object.sub_workflows],
            Documentation.from_flyte_idl(pb2_object.description) if pb2_object.description else None,
        )


class Workflow(_common.FlyteIdlEntity):
    def __init__(self, id, closure):
        """
        :param flytekit.models.core.identifier.Identifier id:
        :param WorkflowClosure closure:
        """
        self._id = id
        self._closure = closure

    @property
    def id(self):
        """
        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._id

    @property
    def closure(self):
        """
        :rtype: WorkflowClosure
        """
        return self._closure

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.workflow_pb2.Workflow
        """
        return _admin_workflow.Workflow(id=self.id.to_flyte_idl(), closure=self.closure.to_flyte_idl())

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.workflow_pb2.Workflow pb2_object:
        :return: Workflow
        """
        return cls(
            id=_identifier.Identifier.from_flyte_idl(pb2_object.id),
            closure=WorkflowClosure.from_flyte_idl(pb2_object.closure),
        )


class WorkflowClosure(_common.FlyteIdlEntity):
    def __init__(self, compiled_workflow):
        """
        :param flytekit.models.core.compiler.CompiledWorkflowClosure compiled_workflow:
        """
        self._compiled_workflow = compiled_workflow

    @property
    def compiled_workflow(self):
        """
        :rtype: flytekit.models.core.compiler.CompiledWorkflowClosure
        """
        return self._compiled_workflow

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.workflow_pb2.WorkflowClosure
        """
        return _admin_workflow.WorkflowClosure(compiled_workflow=self.compiled_workflow.to_flyte_idl())

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.admin.workflow_pb2.WorkflowClosure p:
        :rtype: WorkflowClosure
        """
        return cls(compiled_workflow=_compiler_models.CompiledWorkflowClosure.from_flyte_idl(p.compiled_workflow))
