from flyteidl.core import identifier_pb2 as identifier_pb2

from flytekit.models import common as _common_models


class ResourceType(object):
    UNSPECIFIED = identifier_pb2.UNSPECIFIED
    TASK = identifier_pb2.TASK
    WORKFLOW = identifier_pb2.WORKFLOW
    LAUNCH_PLAN = identifier_pb2.LAUNCH_PLAN


class Identifier(_common_models.FlyteIdlEntity):
    def __init__(self, resource_type, project, domain, name, version):
        """
        :param int resource_type: enum value from ResourceType
        :param Text project:
        :param Text domain:
        :param Text name:
        :param Text version:
        """
        self._resource_type = resource_type
        self._project = project
        self._domain = domain
        self._name = name
        self._version = version

    @property
    def resource_type(self):
        """
        enum value from ResourceType
        :rtype: int
        """
        return self._resource_type

    def resource_type_name(self) -> str:
        return identifier_pb2.ResourceType.Name(self.resource_type)

    @property
    def project(self):
        """
        :rtype: Text
        """
        return self._project

    @property
    def domain(self):
        """
        :rtype: Text
        """
        return self._domain

    @property
    def name(self):
        """
        :rtype: Text
        """
        return self._name

    @property
    def version(self):
        """
        :rtype: Text
        """
        return self._version

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.identifier_pb2.Identifier
        """
        return identifier_pb2.Identifier(
            resource_type=self.resource_type,
            project=self.project,
            domain=self.domain,
            name=self.name,
            version=self.version,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.core.identifier_pb2.Identifier p:
        :rtype: Identifier
        """
        return cls(
            resource_type=p.resource_type,
            project=p.project,
            domain=p.domain,
            name=p.name,
            version=p.version,
        )

    def __repr__(self):
        return self.__str__()

    def __str__(self):
        return f"{self.resource_type_name()}:{self.project}:{self.domain}:{self.name}:{self.version}"


class WorkflowExecutionIdentifier(_common_models.FlyteIdlEntity):
    def __init__(self, project, domain, name):
        """
        :param Text project:
        :param Text domain:
        :param Text name:
        """
        self._project = project
        self._domain = domain
        self._name = name

    @property
    def project(self):
        """
        :rtype: Text
        """
        return self._project

    @property
    def domain(self):
        """
        :rtype: Text
        """
        return self._domain

    @property
    def name(self):
        """
        :rtype: Text
        """
        return self._name

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.identifier_pb2.WorkflowExecutionIdentifier
        """
        return identifier_pb2.WorkflowExecutionIdentifier(
            project=self.project,
            domain=self.domain,
            name=self.name,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.core.identifier_pb2.WorkflowExecutionIdentifier p:
        :rtype: WorkflowExecutionIdentifier
        """
        return cls(
            project=p.project,
            domain=p.domain,
            name=p.name,
        )


class NodeExecutionIdentifier(_common_models.FlyteIdlEntity):
    def __init__(self, node_id, execution_id):
        """
        :param Text node_id:
        :param WorkflowExecutionIdentifier execution_id:
        """
        self._node_id = node_id
        self._execution_id = execution_id

    @property
    def node_id(self):
        """
        :rtype: Text
        """
        return self._node_id

    @property
    def execution_id(self):
        """
        :rtype: WorkflowExecutionIdentifier
        """
        return self._execution_id

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.identifier_pb2.NodeExecutionIdentifier
        """
        return identifier_pb2.NodeExecutionIdentifier(
            node_id=self.node_id,
            execution_id=self.execution_id.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.core.identifier_pb2.NodeExecutionIdentifier p:
        :rtype: NodeExecutionIdentifier
        """
        return cls(
            node_id=p.node_id,
            execution_id=WorkflowExecutionIdentifier.from_flyte_idl(p.execution_id),
        )


class TaskExecutionIdentifier(_common_models.FlyteIdlEntity):
    def __init__(self, task_id, node_execution_id, retry_attempt):
        """
        :param Identifier task_id: The identifier for the task that is executing
        :param NodeExecutionIdentifier node_execution_id: The identifier for the node that owns this execution.
        :param int retry_attempt: The attempt for executing this task by the owning node.
        """
        self._task_id = task_id
        self._node_execution_id = node_execution_id
        self._retry_attempt = retry_attempt

    @property
    def task_id(self):
        """
        :rtype: Identifier
        """
        return self._task_id

    @property
    def node_execution_id(self):
        """
        :rtype: NodeExecutionIdentifier
        """
        return self._node_execution_id

    @property
    def retry_attempt(self):
        """
        :rtype: int
        """
        return self._retry_attempt

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.identifier_pb2.TaskExecutionIdentifier
        """
        return identifier_pb2.TaskExecutionIdentifier(
            task_id=self.task_id.to_flyte_idl(),
            node_execution_id=self.node_execution_id.to_flyte_idl(),
            retry_attempt=self.retry_attempt,
        )

    @classmethod
    def from_flyte_idl(cls, proto):
        """
        :param flyteidl.core.identifier_pb2.TaskExecutionIdentifier proto:
        :rtype: TaskExecutionIdentifier
        """
        return cls(
            task_id=Identifier.from_flyte_idl(proto.task_id),
            node_execution_id=NodeExecutionIdentifier.from_flyte_idl(proto.node_execution_id),
            retry_attempt=proto.retry_attempt,
        )


class SignalIdentifier(_common_models.FlyteIdlEntity):
    def __init__(self, signal_id: str, execution_id: WorkflowExecutionIdentifier):
        """
        :param signal_id: User provided name for the gate node.
        :param execution_id: The workflow execution id this signal is for.
        """
        self._signal_id = signal_id
        self._execution_id = execution_id

    @property
    def signal_id(self) -> str:
        return self._signal_id

    @property
    def execution_id(self) -> WorkflowExecutionIdentifier:
        return self._execution_id

    def to_flyte_idl(self) -> identifier_pb2.SignalIdentifier:
        return identifier_pb2.SignalIdentifier(
            signal_id=self.signal_id,
            execution_id=self.execution_id.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, proto: identifier_pb2.SignalIdentifier) -> "SignalIdentifier":
        return cls(
            signal_id=proto.signal_id,
            execution_id=WorkflowExecutionIdentifier.from_flyte_idl(proto.execution_id),
        )
