import datetime

from flytekit.models.core import execution


def test_task_logs():
    obj = execution.TaskLog("uri", "name", execution.TaskLog.MessageFormat.CSV, datetime.timedelta(days=30))
    assert obj.message_format == execution.TaskLog.MessageFormat.CSV
    assert obj.uri == "uri"
    assert obj.name == "name"
    assert obj.ttl == datetime.timedelta(days=30)

    obj2 = execution.TaskLog.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
    assert obj2.message_format == execution.TaskLog.MessageFormat.CSV
    assert obj2.uri == "uri"
    assert obj2.name == "name"
    assert obj2.ttl == datetime.timedelta(days=30)


def test_execution_error():
    obj = execution.ExecutionError("code", "message", "uri", execution.ExecutionError.ErrorKind.UNKNOWN)
    assert obj.code == "code"
    assert obj.message == "message"
    assert obj.error_uri == "uri"
    assert obj.kind == 0

    obj2 = execution.ExecutionError.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.code == "code"
    assert obj2.message == "message"
    assert obj2.error_uri == "uri"
    assert obj2.kind == 0
