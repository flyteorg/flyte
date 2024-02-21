from flytekit.models.core import errors, execution


def test_container_error():
    obj = errors.ContainerError(
        "code", "my message", errors.ContainerError.Kind.RECOVERABLE, execution.ExecutionError.ErrorKind.SYSTEM
    )
    assert obj.code == "code"
    assert obj.message == "my message"
    assert obj.kind == errors.ContainerError.Kind.RECOVERABLE
    assert obj.origin == 2

    obj2 = errors.ContainerError.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.code == "code"
    assert obj2.message == "my message"
    assert obj2.kind == errors.ContainerError.Kind.RECOVERABLE
    assert obj2.origin == 2


def test_error_document():
    ce = errors.ContainerError(
        "code", "my message", errors.ContainerError.Kind.RECOVERABLE, execution.ExecutionError.ErrorKind.USER
    )
    obj = errors.ErrorDocument(ce)
    assert obj.error == ce

    obj2 = errors.ErrorDocument.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.error == ce
