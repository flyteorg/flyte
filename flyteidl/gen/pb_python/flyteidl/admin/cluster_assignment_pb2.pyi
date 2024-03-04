from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ClusterAssignment(_message.Message):
    __slots__ = ["cluster_pool_name", "execution_cluster_label_name"]
    CLUSTER_POOL_NAME_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_CLUSTER_LABEL_NAME_FIELD_NUMBER: _ClassVar[int]
    cluster_pool_name: str
    execution_cluster_label_name: str
    def __init__(self, cluster_pool_name: _Optional[str] = ..., execution_cluster_label_name: _Optional[str] = ...) -> None: ...
