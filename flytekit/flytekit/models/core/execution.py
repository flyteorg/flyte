import datetime
import typing

from flyteidl.core import execution_pb2 as _execution_pb2

from flytekit.models import common as _common


class WorkflowExecutionPhase(object):
    """
    This class holds enum values used for setting notifications. See :py:class:`flytekit.Email`
    for sample usage.
    """

    UNDEFINED = _execution_pb2.WorkflowExecution.UNDEFINED
    QUEUED = _execution_pb2.WorkflowExecution.QUEUED
    RUNNING = _execution_pb2.WorkflowExecution.RUNNING
    SUCCEEDING = _execution_pb2.WorkflowExecution.SUCCEEDING
    SUCCEEDED = _execution_pb2.WorkflowExecution.SUCCEEDED
    FAILING = _execution_pb2.WorkflowExecution.FAILING
    FAILED = _execution_pb2.WorkflowExecution.FAILED
    ABORTED = _execution_pb2.WorkflowExecution.ABORTED
    TIMED_OUT = _execution_pb2.WorkflowExecution.TIMED_OUT
    ABORTING = _execution_pb2.WorkflowExecution.ABORTING

    @classmethod
    def enum_to_string(cls, int_value):
        """
        :param int_value:
        :rtype: Text
        """
        for name, value in cls.__dict__.items():
            if value == int_value:
                return name
        return str(int_value)


class NodeExecutionPhase(object):
    UNDEFINED = _execution_pb2.NodeExecution.UNDEFINED
    QUEUED = _execution_pb2.NodeExecution.QUEUED
    RUNNING = _execution_pb2.NodeExecution.RUNNING
    SUCCEEDED = _execution_pb2.NodeExecution.SUCCEEDED
    FAILING = _execution_pb2.NodeExecution.FAILING
    FAILED = _execution_pb2.NodeExecution.FAILED
    ABORTED = _execution_pb2.NodeExecution.ABORTED
    SKIPPED = _execution_pb2.NodeExecution.SKIPPED
    TIMED_OUT = _execution_pb2.NodeExecution.TIMED_OUT
    DYNAMIC_RUNNING = _execution_pb2.NodeExecution.DYNAMIC_RUNNING
    RECOVERED = _execution_pb2.NodeExecution.RECOVERED

    @classmethod
    def enum_to_string(cls, int_value):
        """
        :param int_value:
        :rtype: Text
        """
        for name, value in cls.__dict__.items():
            if value == int_value:
                return name
        return str(int_value)


class TaskExecutionPhase(object):
    UNDEFINED = _execution_pb2.TaskExecution.UNDEFINED
    RUNNING = _execution_pb2.TaskExecution.RUNNING
    SUCCEEDED = _execution_pb2.TaskExecution.SUCCEEDED
    FAILED = _execution_pb2.TaskExecution.FAILED
    ABORTED = _execution_pb2.TaskExecution.ABORTED
    QUEUED = _execution_pb2.TaskExecution.QUEUED
    INITIALIZING = _execution_pb2.TaskExecution.INITIALIZING
    WAITING_FOR_RESOURCES = _execution_pb2.TaskExecution.WAITING_FOR_RESOURCES

    @classmethod
    def enum_to_string(cls, int_value):
        """
        :param int_value:
        :rtype: Text
        """
        for name, value in cls.__dict__.items():
            if value == int_value:
                return name
        return str(int_value)


class ExecutionError(_common.FlyteIdlEntity):
    class ErrorKind(object):
        UNKNOWN = _execution_pb2.ExecutionError.ErrorKind.UNKNOWN
        USER = _execution_pb2.ExecutionError.ErrorKind.USER
        SYSTEM = _execution_pb2.ExecutionError.ErrorKind.SYSTEM

    def __init__(self, code: str, message: str, error_uri: str, kind: int):
        """
        :param code:
        :param message:
        :param uri:
        :param kind:
        """
        self._code = code
        self._message = message
        self._error_uri = error_uri
        self._kind = kind

    @property
    def code(self):
        """
        :rtype: Text
        """
        return self._code

    @property
    def message(self):
        """
        :rtype: Text
        """
        return self._message

    @property
    def error_uri(self):
        """
        :rtype: Text
        """
        return self._error_uri

    @property
    def kind(self) -> int:
        """
        Enum value from ErrorKind
        """
        return self._kind

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.execution_pb2.ExecutionError
        """
        return _execution_pb2.ExecutionError(
            code=self.code,
            message=self.message,
            error_uri=self.error_uri,
            kind=self.kind,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.core.execution_pb2.ExecutionError p:
        :rtype: ExecutionError
        """
        return cls(code=p.code, message=p.message, error_uri=p.error_uri, kind=p.kind)


class TaskLog(_common.FlyteIdlEntity):
    class MessageFormat(object):
        UNKNOWN = _execution_pb2.TaskLog.UNKNOWN
        CSV = _execution_pb2.TaskLog.CSV
        JSON = _execution_pb2.TaskLog.JSON

    def __init__(
        self,
        uri: str,
        name: str,
        message_format: typing.Optional[MessageFormat] = None,
        ttl: typing.Optional[datetime.timedelta] = None,
    ):
        """
        :param Text uri:
        :param Text name:
        :param MessageFormat message_format: Enum value from TaskLog.MessageFormat
        :param datetime.timedelta ttl: The time the log will persist for.  0 represents unknown or ephemeral in nature.
        """
        self._uri = uri
        self._name = name
        self._message_format = message_format
        self._ttl = ttl

    @property
    def uri(self):
        """
        :rtype: Text
        """
        return self._uri

    @property
    def name(self):
        """
        :rtype: Text
        """
        return self._name

    @property
    def message_format(self):
        """
        Enum value from TaskLog.MessageFormat
        :rtype: MessageFormat
        """
        return self._message_format

    @property
    def ttl(self):
        """
        :rtype: datetime.timedelta
        """
        return self._ttl

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.execution_pb2.TaskLog
        """
        p = _execution_pb2.TaskLog(uri=self.uri, name=self.name, message_format=self.message_format)
        if self.ttl is not None:
            p.ttl.FromTimedelta(self.ttl)
        return p

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.core.execution_pb2.TaskLog p:
        :rtype: TaskLog
        """
        return cls(
            uri=p.uri,
            name=p.name,
            message_format=p.message_format,
            ttl=p.ttl.ToTimedelta() if p.ttl else None,
        )
