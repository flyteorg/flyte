from __future__ import annotations

import datetime
import typing
from datetime import timezone as _timezone
from typing import Optional

import flyteidl
import flyteidl.admin.cluster_assignment_pb2 as _cluster_assignment_pb2
import flyteidl.admin.execution_pb2 as _execution_pb2
import flyteidl.admin.node_execution_pb2 as _node_execution_pb2
import flyteidl.admin.task_execution_pb2 as _task_execution_pb2

import flytekit
from flytekit.models import common as _common_models
from flytekit.models import literals as _literals_models
from flytekit.models import security
from flytekit.models.core import execution as _core_execution
from flytekit.models.core import identifier as _identifier
from flytekit.models.node_execution import DynamicWorkflowNodeMetadata


class SystemMetadata(_common_models.FlyteIdlEntity):
    def __init__(self, execution_cluster: str):
        self._execution_cluster = execution_cluster

    @property
    def execution_cluster(self) -> str:
        return self._execution_cluster

    def to_flyte_idl(self) -> flyteidl.admin.execution_pb2.SystemMetadata:
        return _execution_pb2.SystemMetadata(execution_cluster=self.execution_cluster)

    @classmethod
    def from_flyte_idl(cls, pb2_object: flyteidl.admin.execution_pb2.SystemMetadata) -> SystemMetadata:
        return cls(
            execution_cluster=pb2_object.execution_cluster,
        )


class ExecutionMetadata(_common_models.FlyteIdlEntity):
    class ExecutionMode(object):
        MANUAL = 0
        SCHEDULED = 1
        SYSTEM = 2

    def __init__(
        self,
        mode: int,
        principal: str,
        nesting: int,
        scheduled_at: Optional[datetime.datetime] = None,
        parent_node_execution: Optional[_identifier.NodeExecutionIdentifier] = None,
        reference_execution: Optional[_identifier.WorkflowExecutionIdentifier] = None,
        system_metadata: Optional[SystemMetadata] = None,
    ):
        """
        :param mode: An enum value from ExecutionMetadata.ExecutionMode which specifies how the job started.
        :param principal: The entity that triggered the execution
        :param nesting: An integer representing how deeply nested the workflow is (i.e. was it triggered by a parent
            workflow)
        :param scheduled_at: For scheduled executions, the requested time for execution for this specific schedule invocation.
        :param parent_node_execution: Which subworkflow node (if any) launched this execution
        :param reference_execution: Optional, reference workflow execution related to this execution
        :param system_metadata: Optional, platform-specific metadata about the execution.
        """
        self._mode = mode
        self._principal = principal
        self._nesting = nesting
        self._scheduled_at = scheduled_at
        self._parent_node_execution = parent_node_execution
        self._reference_execution = reference_execution
        self._system_metadata = system_metadata

    @property
    def mode(self) -> int:
        """
        An enum value from ExecutionMetadata.ExecutionMode which specifies how the job started.
        """
        return self._mode

    @property
    def principal(self) -> str:
        """
        The entity that triggered the execution
        """
        return self._principal

    @property
    def nesting(self) -> int:
        """
        An integer representing how deeply nested the workflow is (i.e. was it triggered by a parent workflow)
        """
        return self._nesting

    @property
    def scheduled_at(self) -> datetime.datetime:
        """
        For scheduled executions, the requested time for execution for this specific schedule invocation.
        """
        return self._scheduled_at

    @property
    def parent_node_execution(self) -> _identifier.NodeExecutionIdentifier:
        """
        Which subworkflow node (if any) launched this execution
        """
        return self._parent_node_execution

    @property
    def reference_execution(self) -> _identifier.WorkflowExecutionIdentifier:
        """
        Optional, reference workflow execution related to this execution
        """
        return self._reference_execution

    @property
    def system_metadata(self) -> SystemMetadata:
        """
        Optional, platform-specific metadata about the execution.
        """
        return self._system_metadata

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.execution_pb2.ExecutionMetadata
        """
        p = _execution_pb2.ExecutionMetadata(
            mode=self.mode,
            principal=self.principal,
            nesting=self.nesting,
            parent_node_execution=self.parent_node_execution.to_flyte_idl()
            if self.parent_node_execution is not None
            else None,
            reference_execution=self.reference_execution.to_flyte_idl()
            if self.reference_execution is not None
            else None,
            system_metadata=self.system_metadata.to_flyte_idl() if self.system_metadata is not None else None,
        )
        if self.scheduled_at is not None:
            p.scheduled_at.FromDatetime(self.scheduled_at)
        return p

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.execution_pb2.ExecutionMetadata pb2_object:
        :return: ExecutionMetadata
        """
        return cls(
            mode=pb2_object.mode,
            principal=pb2_object.principal,
            nesting=pb2_object.nesting,
            scheduled_at=pb2_object.scheduled_at.ToDatetime() if pb2_object.HasField("scheduled_at") else None,
            parent_node_execution=_identifier.NodeExecutionIdentifier.from_flyte_idl(pb2_object.parent_node_execution)
            if pb2_object.HasField("parent_node_execution")
            else None,
            reference_execution=_identifier.WorkflowExecutionIdentifier.from_flyte_idl(pb2_object.reference_execution)
            if pb2_object.HasField("reference_execution")
            else None,
            system_metadata=SystemMetadata.from_flyte_idl(pb2_object.system_metadata)
            if pb2_object.HasField("system_metadata")
            else None,
        )


class ExecutionSpec(_common_models.FlyteIdlEntity):
    def __init__(
        self,
        launch_plan,
        metadata,
        notifications=None,
        disable_all=None,
        labels=None,
        annotations=None,
        auth_role=None,
        raw_output_data_config=None,
        max_parallelism: Optional[int] = None,
        security_context: Optional[security.SecurityContext] = None,
        overwrite_cache: Optional[bool] = None,
        envs: Optional[_common_models.Envs] = None,
        tags: Optional[typing.List[str]] = None,
        cluster_assignment: Optional[ClusterAssignment] = None,
    ):
        """
        :param flytekit.models.core.identifier.Identifier launch_plan: Launch plan unique identifier to execute
        :param ExecutionMetadata metadata: The metadata to be associated with this execution
        :param NotificationList notifications: List of notifications for this execution.
        :param bool disable_all: If true, all notifications should be disabled.
        :param flytekit.models.common.Labels labels: Labels to apply to the execution.
        :param flytekit.models.common.Annotations annotations: Annotations to apply to the execution
        :param flytekit.models.common.AuthRole auth_role: The authorization method with which to execute the workflow.
        :param raw_output_data_config: Optional location of offloaded data for things like S3, etc.
        :param max_parallelism int: Controls the maximum number of tasknodes that can be run in parallel for the entire
            workflow. This is useful to achieve fairness. Note: MapTasks are regarded as one unit, and
            parallelism/concurrency of MapTasks is independent from this.
        :param security_context: Optional security context to use for this execution.
        :param overwrite_cache: Optional flag to overwrite the cache for this execution.
        :param envs: flytekit.models.common.Envs environment variables to set for this execution.
        :param tags: Optional list of tags to apply to the execution.
        """
        self._launch_plan = launch_plan
        self._metadata = metadata
        self._notifications = notifications
        self._disable_all = disable_all
        self._labels = labels or _common_models.Labels({})
        self._annotations = annotations or _common_models.Annotations({})
        self._auth_role = auth_role or _common_models.AuthRole()
        self._raw_output_data_config = raw_output_data_config
        self._max_parallelism = max_parallelism
        self._security_context = security_context
        self._overwrite_cache = overwrite_cache
        self._envs = envs
        self._tags = tags
        self._cluster_assignment = cluster_assignment

    @property
    def launch_plan(self):
        """
        If the values were too large, this is the URI where the values were offloaded.
        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._launch_plan

    @property
    def metadata(self):
        """
        :rtype: ExecutionMetadata
        """
        return self._metadata

    @property
    def notifications(self):
        """
        :rtype: Optional[NotificationList]
        """
        return self._notifications

    @property
    def disable_all(self):
        """
        :rtype: Optional[bool]
        """
        return self._disable_all

    @property
    def labels(self):
        """
        :rtype: flytekit.models.common.Labels
        """
        return self._labels

    @property
    def annotations(self):
        """
        :rtype: flytekit.models.common.Annotations
        """
        return self._annotations

    @property
    def auth_role(self):
        """
        :rtype: flytekit.models.common.AuthRole
        """
        return self._auth_role

    @property
    def raw_output_data_config(self):
        """
        :rtype: flytekit.models.common.RawOutputDataConfig
        """
        return self._raw_output_data_config

    @property
    def max_parallelism(self) -> int:
        return self._max_parallelism

    @property
    def security_context(self) -> typing.Optional[security.SecurityContext]:
        return self._security_context

    @property
    def overwrite_cache(self) -> Optional[bool]:
        return self._overwrite_cache

    @property
    def envs(self) -> Optional[_common_models.Envs]:
        return self._envs

    @property
    def tags(self) -> Optional[typing.List[str]]:
        return self._tags

    @property
    def cluster_assignment(self) -> Optional[ClusterAssignment]:
        return self._cluster_assignment

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.execution_pb2.ExecutionSpec
        """
        return _execution_pb2.ExecutionSpec(
            launch_plan=self.launch_plan.to_flyte_idl(),
            metadata=self.metadata.to_flyte_idl(),
            notifications=self.notifications.to_flyte_idl() if self.notifications else None,
            disable_all=self.disable_all,  # type: ignore
            labels=self.labels.to_flyte_idl(),
            annotations=self.annotations.to_flyte_idl(),
            auth_role=self._auth_role.to_flyte_idl() if self.auth_role else None,
            raw_output_data_config=self._raw_output_data_config.to_flyte_idl()
            if self._raw_output_data_config
            else None,
            max_parallelism=self.max_parallelism,
            security_context=self.security_context.to_flyte_idl() if self.security_context else None,
            overwrite_cache=self.overwrite_cache,
            envs=self.envs.to_flyte_idl() if self.envs else None,
            tags=self.tags,
            cluster_assignment=self._cluster_assignment.to_flyte_idl() if self._cluster_assignment else None,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.admin.execution_pb2.ExecutionSpec p:
        :return: ExecutionSpec
        """
        return cls(
            launch_plan=_identifier.Identifier.from_flyte_idl(p.launch_plan),
            metadata=ExecutionMetadata.from_flyte_idl(p.metadata),
            notifications=NotificationList.from_flyte_idl(p.notifications) if p.HasField("notifications") else None,
            disable_all=p.disable_all if p.HasField("disable_all") else None,
            labels=_common_models.Labels.from_flyte_idl(p.labels),
            annotations=_common_models.Annotations.from_flyte_idl(p.annotations),
            auth_role=_common_models.AuthRole.from_flyte_idl(p.auth_role),
            raw_output_data_config=_common_models.RawOutputDataConfig.from_flyte_idl(p.raw_output_data_config)
            if p.HasField("raw_output_data_config")
            else None,
            max_parallelism=p.max_parallelism,
            security_context=security.SecurityContext.from_flyte_idl(p.security_context)
            if p.security_context
            else None,
            overwrite_cache=p.overwrite_cache,
            envs=_common_models.Envs.from_flyte_idl(p.envs) if p.HasField("envs") else None,
            tags=p.tags,
            cluster_assignment=ClusterAssignment.from_flyte_idl(p.cluster_assignment)
            if p.HasField("cluster_assignment")
            else None,
        )


class ClusterAssignment(_common_models.FlyteIdlEntity):
    def __init__(self, cluster_pool=None):
        """
        :param Text cluster_pool:
        """
        self._cluster_pool = cluster_pool

    @property
    def cluster_pool(self):
        """
        :rtype: Text
        """
        return self._cluster_pool

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin._cluster_assignment_pb2.ClusterAssignment
        """
        return _cluster_assignment_pb2.ClusterAssignment(
            cluster_pool_name=self._cluster_pool,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.admin._cluster_assignment_pb2.ClusterAssignment p:
        :rtype: flyteidl.admin.ClusterAssignment
        """
        return cls(cluster_pool=p.cluster_pool_name)


class LiteralMapBlob(_common_models.FlyteIdlEntity):
    def __init__(self, values=None, uri=None):
        """
        :param flytekit.models.literals.LiteralMap values:
        :param Text uri:
        """
        self._values = values
        self._uri = uri

    @property
    def values(self):
        """
        :rtype: flytekit.models.literals.LiteralMap
        """
        return self._values

    @property
    def uri(self):
        """
        :rtype: Text
        """
        return self._uri

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.execution_pb2.LiteralMapBlob
        """
        return _execution_pb2.LiteralMapBlob(
            values=self.values.to_flyte_idl() if self.values is not None else None,
            uri=self.uri,
        )

    @classmethod
    def from_flyte_idl(cls, pb):
        """
        :param flyteidl.admin.execution_pb2.LiteralMapBlob pb:
        :rtype: LiteralMapBlob
        """
        values = None
        if pb.HasField("values"):
            values = LiteralMapBlob.from_flyte_idl(pb.values)
        return cls(values=values, uri=pb.uri if pb.HasField("uri") else None)


class Execution(_common_models.FlyteIdlEntity):
    def __init__(self, id, spec, closure):
        """
        :param flytekit.models.core.identifier.WorkflowExecutionIdentifier id:
        :param ExecutionSpec spec:
        :param ExecutionClosure closure:
        """
        self._id = id
        self._spec = spec
        self._closure = closure

    @property
    def id(self):
        """
        :rtype: flytekit.models.core.identifier.WorkflowExecutionIdentifier
        """
        return self._id

    @property
    def closure(self):
        """
        :rtype: ExecutionClosure
        """
        return self._closure

    @property
    def spec(self):
        """
        :rtype: ExecutionSpec
        """
        return self._spec

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.execution_pb2.Execution
        """
        return _execution_pb2.Execution(
            id=self.id.to_flyte_idl(),
            closure=self.closure.to_flyte_idl(),
            spec=self.spec.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, pb):
        """
        :param flyteidl.admin.execution_pb2.Execution pb:
        :rtype: Execution
        """
        return cls(
            id=_identifier.WorkflowExecutionIdentifier.from_flyte_idl(pb.id),
            closure=ExecutionClosure.from_flyte_idl(pb.closure),
            spec=ExecutionSpec.from_flyte_idl(pb.spec),
        )


class AbortMetadata(_common_models.FlyteIdlEntity):
    def __init__(self, cause: str, principal: str):
        self._cause = cause
        self._principal = principal

    @property
    def cause(self) -> str:
        return self._cause

    @property
    def principal(self) -> str:
        return self._principal

    def to_flyte_idl(self) -> flyteidl.admin.execution_pb2.AbortMetadata:
        return _execution_pb2.AbortMetadata(cause=self.cause, principal=self.principal)

    @classmethod
    def from_flyte_idl(cls, pb2_object: flyteidl.admin.execution_pb2.AbortMetadata) -> AbortMetadata:
        return cls(
            cause=pb2_object.cause,
            principal=pb2_object.principal,
        )


class ExecutionClosure(_common_models.FlyteIdlEntity):
    def __init__(
        self,
        phase: int,
        started_at: datetime.datetime,
        duration: datetime.timedelta,
        error: typing.Optional[flytekit.models.core.execution.ExecutionError] = None,
        outputs: typing.Optional[LiteralMapBlob] = None,
        abort_metadata: typing.Optional[AbortMetadata] = None,
        created_at: typing.Optional[datetime.datetime] = None,
        updated_at: typing.Optional[datetime.datetime] = None,
    ):
        """
        :param phase: From the flytekit.models.core.execution.WorkflowExecutionPhase enum
        :param started_at:
        :param duration: Duration for which the execution has been running.
        :param error:
        :param outputs:
        :param abort_metadata: Specifies metadata around an aborted workflow execution.
        """
        self._phase = phase
        self._started_at = started_at
        self._duration = duration
        self._error = error
        self._outputs = outputs
        self._abort_metadata = abort_metadata
        self._created_at = created_at
        self._updated_at = updated_at

    @property
    def error(self) -> flytekit.models.core.execution.ExecutionError:
        return self._error

    @property
    def phase(self) -> int:
        """
        From the flytekit.models.core.execution.WorkflowExecutionPhase enum
        """
        return self._phase

    @property
    def started_at(self) -> datetime.datetime:
        return self._started_at

    @property
    def duration(self) -> datetime.timedelta:
        return self._duration

    @property
    def created_at(self) -> typing.Optional[datetime.datetime]:
        return self._created_at

    @property
    def updated_at(self) -> typing.Optional[datetime.datetime]:
        return self._updated_at

    @property
    def outputs(self) -> LiteralMapBlob:
        return self._outputs

    @property
    def abort_metadata(self) -> AbortMetadata:
        return self._abort_metadata

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.execution_pb2.ExecutionClosure
        """
        obj = _execution_pb2.ExecutionClosure(
            phase=self.phase,
            error=self.error.to_flyte_idl() if self.error is not None else None,
            outputs=self.outputs.to_flyte_idl() if self.outputs is not None else None,
            abort_metadata=self.abort_metadata.to_flyte_idl() if self.abort_metadata is not None else None,
        )
        obj.started_at.FromDatetime(self.started_at.astimezone(_timezone.utc).replace(tzinfo=None))
        obj.duration.FromTimedelta(self.duration)
        if self.created_at:
            obj.created_at.FromDatetime(self.created_at.astimezone(_timezone.utc).replace(tzinfo=None))
        if self.updated_at:
            obj.updated_at.FromDatetime(self.updated_at.astimezone(_timezone.utc).replace(tzinfo=None))
        return obj

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.execution_pb2.ExecutionClosure pb2_object:
        :rtype: ExecutionClosure
        """
        error = None
        if pb2_object.HasField("error"):
            error = _core_execution.ExecutionError.from_flyte_idl(pb2_object.error)
        outputs = None
        if pb2_object.HasField("outputs"):
            outputs = LiteralMapBlob.from_flyte_idl(pb2_object.outputs)
        abort_metadata = None
        if pb2_object.HasField("abort_metadata"):
            abort_metadata = AbortMetadata.from_flyte_idl(pb2_object.abort_metadata)
        return cls(
            error=error,
            outputs=outputs,
            phase=pb2_object.phase,
            started_at=pb2_object.started_at.ToDatetime().replace(tzinfo=_timezone.utc),
            duration=pb2_object.duration.ToTimedelta(),
            abort_metadata=abort_metadata,
            created_at=pb2_object.created_at.ToDatetime().replace(tzinfo=_timezone.utc)
            if pb2_object.HasField("created_at")
            else None,
            updated_at=pb2_object.updated_at.ToDatetime().replace(tzinfo=_timezone.utc)
            if pb2_object.HasField("updated_at")
            else None,
        )


class NotificationList(_common_models.FlyteIdlEntity):
    def __init__(self, notifications):
        """
        :param list[flytekit.models.common.Notification] notifications: A simple list of notifications.
        """
        self._notifications = notifications

    @property
    def notifications(self):
        """
        :rtype: list[flytekit.models.common.Notification]
        """
        return self._notifications

    def to_flyte_idl(self):
        """
        :rtype:  flyteidl.admin.execution_pb2.NotificationList
        """
        return _execution_pb2.NotificationList(notifications=[n.to_flyte_idl() for n in self.notifications])

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.execution_pb2.NotificationList pb2_object:
        :rtype: NotificationList
        """
        return cls([_common_models.Notification.from_flyte_idl(p) for p in pb2_object.notifications])


class _CommonDataResponse(_common_models.FlyteIdlEntity):
    """
    Currently, node, task, and workflow execution all have the same get data response. So we'll create this common
    superclass to reduce code duplication until things diverge in the future.
    """

    def __init__(self, inputs, outputs, full_inputs, full_outputs):
        """
        :param _common_models.UrlBlob inputs:
        :param _common_models.UrlBlob outputs:
        :param _literals_models.LiteralMap full_inputs:
        :param _literals_models.LiteralMap full_outputs:
        """
        self._inputs = inputs
        self._outputs = outputs
        self._full_inputs = full_inputs
        self._full_outputs = full_outputs

    @property
    def inputs(self):
        """
        :rtype: _common_models.UrlBlob
        """
        return self._inputs

    @property
    def outputs(self):
        """
        :rtype: _common_models.UrlBlob
        """
        return self._outputs

    @property
    def full_inputs(self):
        """
        :rtype: _literals_models.LiteralMap
        """
        return self._full_inputs

    @property
    def full_outputs(self):
        """
        :rtype: _literals_models.LiteralMap
        """
        return self._full_outputs


class WorkflowExecutionGetDataResponse(_CommonDataResponse):
    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param _execution_pb2.WorkflowExecutionGetDataResponse pb2_object:
        :rtype: WorkflowExecutionGetDataResponse
        """
        return cls(
            inputs=_common_models.UrlBlob.from_flyte_idl(pb2_object.inputs),
            outputs=_common_models.UrlBlob.from_flyte_idl(pb2_object.outputs),
            full_inputs=_literals_models.LiteralMap.from_flyte_idl(pb2_object.full_inputs),
            full_outputs=_literals_models.LiteralMap.from_flyte_idl(pb2_object.full_outputs),
        )

    def to_flyte_idl(self):
        """
        :rtype: _execution_pb2.WorkflowExecutionGetDataResponse
        """
        return _execution_pb2.WorkflowExecutionGetDataResponse(
            inputs=self.inputs.to_flyte_idl(),
            outputs=self.outputs.to_flyte_idl(),
            full_inputs=self.full_inputs.to_flyte_idl(),
            full_outputs=self.full_outputs.to_flyte_idl(),
        )


class TaskExecutionGetDataResponse(_CommonDataResponse):
    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param _task_execution_pb2.TaskExecutionGetDataResponse pb2_object:
        :rtype: TaskExecutionGetDataResponse
        """
        return cls(
            inputs=_common_models.UrlBlob.from_flyte_idl(pb2_object.inputs),
            outputs=_common_models.UrlBlob.from_flyte_idl(pb2_object.outputs),
            full_inputs=_literals_models.LiteralMap.from_flyte_idl(pb2_object.full_inputs),
            full_outputs=_literals_models.LiteralMap.from_flyte_idl(pb2_object.full_outputs),
        )

    def to_flyte_idl(self):
        """
        :rtype: _task_execution_pb2.TaskExecutionGetDataResponse
        """
        return _task_execution_pb2.TaskExecutionGetDataResponse(
            inputs=self.inputs.to_flyte_idl(),
            outputs=self.outputs.to_flyte_idl(),
            full_inputs=self.full_inputs.to_flyte_idl(),
            full_outputs=self.full_outputs.to_flyte_idl(),
        )


class NodeExecutionGetDataResponse(_CommonDataResponse):
    def __init__(self, *args, dynamic_workflow: typing.Optional[DynamicWorkflowNodeMetadata] = None, **kwargs):
        super().__init__(*args, **kwargs)
        self._dynamic_workflow = dynamic_workflow

    @property
    def dynamic_workflow(self) -> typing.Optional[DynamicWorkflowNodeMetadata]:
        return self._dynamic_workflow

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param _node_execution_pb2.NodeExecutionGetDataResponse pb2_object:
        :rtype: NodeExecutionGetDataResponse
        """
        return cls(
            inputs=_common_models.UrlBlob.from_flyte_idl(pb2_object.inputs),
            outputs=_common_models.UrlBlob.from_flyte_idl(pb2_object.outputs),
            full_inputs=_literals_models.LiteralMap.from_flyte_idl(pb2_object.full_inputs),
            full_outputs=_literals_models.LiteralMap.from_flyte_idl(pb2_object.full_outputs),
            dynamic_workflow=DynamicWorkflowNodeMetadata.from_flyte_idl(pb2_object.dynamic_workflow)
            if pb2_object.HasField("dynamic_workflow")
            else None,
        )

    def to_flyte_idl(self):
        """
        :rtype: _node_execution_pb2.NodeExecutionGetDataResponse
        """
        return _node_execution_pb2.NodeExecutionGetDataResponse(
            inputs=self.inputs.to_flyte_idl(),
            outputs=self.outputs.to_flyte_idl(),
            full_inputs=self.full_inputs.to_flyte_idl(),
            full_outputs=self.full_outputs.to_flyte_idl(),
            dynamic_workflow=self.dynamic_workflow.to_flyte_idl() if self.dynamic_workflow else None,
        )
