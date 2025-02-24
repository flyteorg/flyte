import datetime
import typing
from datetime import timezone as _timezone

import flyteidl.admin.node_execution_pb2 as _node_execution_pb2

from flytekit.models import common as _common_models
from flytekit.models.core import catalog as catalog_models
from flytekit.models.core import compiler as core_compiler_models
from flytekit.models.core import execution as _core_execution
from flytekit.models.core import identifier as _identifier


class WorkflowNodeMetadata(_common_models.FlyteIdlEntity):
    def __init__(self, execution_id: _identifier.WorkflowExecutionIdentifier):
        self._execution_id = execution_id

    @property
    def execution_id(self) -> _identifier.WorkflowExecutionIdentifier:
        return self._execution_id

    def to_flyte_idl(self) -> _node_execution_pb2.WorkflowNodeMetadata:
        return _node_execution_pb2.WorkflowNodeMetadata(
            executionId=self.execution_id.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, p: _node_execution_pb2.WorkflowNodeMetadata) -> "WorkflowNodeMetadata":
        return cls(
            execution_id=_identifier.WorkflowExecutionIdentifier.from_flyte_idl(p.executionId),
        )


class DynamicWorkflowNodeMetadata(_common_models.FlyteIdlEntity):
    def __init__(self, id: _identifier.Identifier, compiled_workflow: core_compiler_models.CompiledWorkflowClosure):
        self._id = id
        self._compiled_workflow = compiled_workflow

    @property
    def id(self) -> _identifier.Identifier:
        return self._id

    @property
    def compiled_workflow(self) -> core_compiler_models.CompiledWorkflowClosure:
        return self._compiled_workflow

    def to_flyte_idl(self) -> _node_execution_pb2.DynamicWorkflowNodeMetadata:
        return _node_execution_pb2.DynamicWorkflowNodeMetadata(
            id=self.id.to_flyte_idl(),
            compiled_workflow=self.compiled_workflow.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, p: _node_execution_pb2.DynamicWorkflowNodeMetadata) -> "DynamicWorkflowNodeMetadata":
        yy = cls(
            id=_identifier.Identifier.from_flyte_idl(p.id),
            compiled_workflow=core_compiler_models.CompiledWorkflowClosure.from_flyte_idl(p.compiled_workflow),
        )
        return yy


class TaskNodeMetadata(_common_models.FlyteIdlEntity):
    def __init__(self, cache_status: int, catalog_key: catalog_models.CatalogMetadata):
        self._cache_status = cache_status
        self._catalog_key = catalog_key

    @property
    def cache_status(self) -> int:
        return self._cache_status

    @property
    def catalog_key(self) -> catalog_models.CatalogMetadata:
        return self._catalog_key

    def to_flyte_idl(self) -> _node_execution_pb2.TaskNodeMetadata:
        return _node_execution_pb2.TaskNodeMetadata(
            cache_status=self.cache_status,
            catalog_key=self.catalog_key.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, p: _node_execution_pb2.TaskNodeMetadata) -> "TaskNodeMetadata":
        return cls(
            cache_status=p.cache_status,
            catalog_key=catalog_models.CatalogMetadata.from_flyte_idl(p.catalog_key),
        )


class NodeExecutionClosure(_common_models.FlyteIdlEntity):
    def __init__(
        self,
        phase,
        started_at,
        duration,
        output_uri=None,
        deck_uri=None,
        error=None,
        workflow_node_metadata: typing.Optional[WorkflowNodeMetadata] = None,
        task_node_metadata: typing.Optional[TaskNodeMetadata] = None,
        created_at: typing.Optional[datetime.datetime] = None,
        updated_at: typing.Optional[datetime.datetime] = None,
    ):
        """
        :param int phase:
        :param datetime.datetime started_at:
        :param datetime.timedelta duration:
        :param Text output_uri:
        :param flytekit.models.core.execution.ExecutionError error:
        """
        self._phase = phase
        self._started_at = started_at
        self._duration = duration
        self._output_uri = output_uri
        self._deck_uri = deck_uri
        self._error = error
        self._workflow_node_metadata = workflow_node_metadata
        self._task_node_metadata = task_node_metadata
        # TODO: Add output_data field as well.
        self._created_at = created_at
        self._updated_at = updated_at

    @property
    def phase(self):
        """
        :rtype: int
        """
        return self._phase

    @property
    def started_at(self):
        """
        :rtype: datetime.datetime
        """
        return self._started_at

    @property
    def duration(self):
        """
        :rtype: datetime.timedelta
        """
        return self._duration

    @property
    def created_at(self) -> typing.Optional[datetime.datetime]:
        return self._created_at

    @property
    def updated_at(self) -> typing.Optional[datetime.datetime]:
        return self._updated_at

    @property
    def output_uri(self):
        """
        :rtype: Text
        """
        return self._output_uri

    @property
    def deck_uri(self):
        """
        :rtype: str
        """
        return self._deck_uri

    @property
    def error(self):
        """
        :rtype: flytekit.models.core.execution.ExecutionError
        """
        return self._error

    @property
    def workflow_node_metadata(self) -> typing.Optional[WorkflowNodeMetadata]:
        return self._workflow_node_metadata

    @property
    def task_node_metadata(self) -> typing.Optional[TaskNodeMetadata]:
        return self._task_node_metadata

    @property
    def target_metadata(self) -> typing.Union[WorkflowNodeMetadata, TaskNodeMetadata]:
        return self.workflow_node_metadata or self.task_node_metadata

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.node_execution_pb2.NodeExecutionClosure
        """
        obj = _node_execution_pb2.NodeExecutionClosure(
            phase=self.phase,
            output_uri=self.output_uri,
            deck_uri=self.deck_uri,
            error=self.error.to_flyte_idl() if self.error is not None else None,
            workflow_node_metadata=self.workflow_node_metadata.to_flyte_idl()
            if self.workflow_node_metadata is not None
            else None,
            task_node_metadata=self.task_node_metadata.to_flyte_idl() if self.task_node_metadata is not None else None,
        )
        obj.started_at.FromDatetime(self.started_at.astimezone(_timezone.utc).replace(tzinfo=None))
        obj.duration.FromTimedelta(self.duration)
        if self.created_at:
            obj.created_at.FromDatetime(self.created_at.astimezone(_timezone.utc).replace(tzinfo=None))
        if self.updated_at:
            obj.updated_at.FromDatetime(self.updated_at.astimezone(_timezone.utc).replace(tzinfo=None))
        return obj

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.admin.node_execution_pb2.NodeExecutionClosure p:
        :rtype: NodeExecutionClosure
        """
        return cls(
            phase=p.phase,
            output_uri=p.output_uri if p.HasField("output_uri") else None,
            deck_uri=p.deck_uri,
            error=_core_execution.ExecutionError.from_flyte_idl(p.error) if p.HasField("error") else None,
            started_at=p.started_at.ToDatetime().replace(tzinfo=_timezone.utc),
            duration=p.duration.ToTimedelta(),
            workflow_node_metadata=WorkflowNodeMetadata.from_flyte_idl(p.workflow_node_metadata)
            if p.HasField("workflow_node_metadata")
            else None,
            task_node_metadata=TaskNodeMetadata.from_flyte_idl(p.task_node_metadata)
            if p.HasField("task_node_metadata")
            else None,
            created_at=p.created_at.ToDatetime().replace(tzinfo=_timezone.utc) if p.HasField("created_at") else None,
            updated_at=p.updated_at.ToDatetime().replace(tzinfo=_timezone.utc) if p.HasField("updated_at") else None,
        )


class NodeExecutionMetaData(_common_models.FlyteIdlEntity):
    def __init__(self, retry_group: str, is_parent_node: bool, spec_node_id: str):
        self._retry_group = retry_group
        self._is_parent_node = is_parent_node
        self._spec_node_id = spec_node_id

    @property
    def retry_group(self) -> str:
        return self._retry_group

    @property
    def is_parent_node(self) -> bool:
        return self._is_parent_node

    @property
    def spec_node_id(self) -> str:
        return self._spec_node_id

    def to_flyte_idl(self) -> _node_execution_pb2.NodeExecutionMetaData:
        return _node_execution_pb2.NodeExecutionMetaData(
            retry_group=self.retry_group,
            is_parent_node=self.is_parent_node,
            spec_node_id=self.spec_node_id,
        )

    @classmethod
    def from_flyte_idl(cls, p: _node_execution_pb2.NodeExecutionMetaData) -> "NodeExecutionMetaData":
        return cls(
            retry_group=p.retry_group,
            is_parent_node=p.is_parent_node,
            spec_node_id=p.spec_node_id,
        )


class NodeExecution(_common_models.FlyteIdlEntity):
    def __init__(self, id, input_uri, closure, metadata):
        """
        :param flytekit.models.core.identifier.NodeExecutionIdentifier id:
        :param Text input_uri:
        :param NodeExecutionClosure closure:
        :param NodeExecutionMetaData metadata:
        """
        self._id = id
        self._input_uri = input_uri
        self._closure = closure
        self._metadata = metadata

    @property
    def id(self):
        """
        :rtype: flytekit.models.core.identifier.NodeExecutionIdentifier
        """
        return self._id

    @property
    def input_uri(self):
        """
        :rtype: Text
        """
        return self._input_uri

    @property
    def closure(self):
        """
        :rtype: NodeExecutionClosure
        """
        return self._closure

    @property
    def metadata(self) -> NodeExecutionMetaData:
        return self._metadata

    def to_flyte_idl(self) -> _node_execution_pb2.NodeExecution:
        return _node_execution_pb2.NodeExecution(
            id=self.id.to_flyte_idl(),
            input_uri=self.input_uri,
            closure=self.closure.to_flyte_idl(),
            metadata=self.metadata.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, p: _node_execution_pb2.NodeExecution) -> "NodeExecution":
        return cls(
            id=_identifier.NodeExecutionIdentifier.from_flyte_idl(p.id),
            input_uri=p.input_uri,
            closure=NodeExecutionClosure.from_flyte_idl(p.closure),
            metadata=NodeExecutionMetaData.from_flyte_idl(p.metadata),
        )
