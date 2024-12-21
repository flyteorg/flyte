from google.protobuf import struct_pb2 as _struct_pb2
from flyteidl.core import tasks_pb2 as _tasks_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SparkApplication(_message.Message):
    __slots__ = []
    class Type(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        PYTHON: _ClassVar[SparkApplication.Type]
        JAVA: _ClassVar[SparkApplication.Type]
        SCALA: _ClassVar[SparkApplication.Type]
        R: _ClassVar[SparkApplication.Type]
    PYTHON: SparkApplication.Type
    JAVA: SparkApplication.Type
    SCALA: SparkApplication.Type
    R: SparkApplication.Type
    def __init__(self) -> None: ...

class SparkJob(_message.Message):
    __slots__ = ["applicationType", "mainApplicationFile", "mainClass", "sparkConf", "hadoopConf", "executorPath", "databricksConf", "databricksToken", "databricksInstance", "driverPod", "executorPod"]
    class SparkConfEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    class HadoopConfEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    APPLICATIONTYPE_FIELD_NUMBER: _ClassVar[int]
    MAINAPPLICATIONFILE_FIELD_NUMBER: _ClassVar[int]
    MAINCLASS_FIELD_NUMBER: _ClassVar[int]
    SPARKCONF_FIELD_NUMBER: _ClassVar[int]
    HADOOPCONF_FIELD_NUMBER: _ClassVar[int]
    EXECUTORPATH_FIELD_NUMBER: _ClassVar[int]
    DATABRICKSCONF_FIELD_NUMBER: _ClassVar[int]
    DATABRICKSTOKEN_FIELD_NUMBER: _ClassVar[int]
    DATABRICKSINSTANCE_FIELD_NUMBER: _ClassVar[int]
    DRIVERPOD_FIELD_NUMBER: _ClassVar[int]
    EXECUTORPOD_FIELD_NUMBER: _ClassVar[int]
    applicationType: SparkApplication.Type
    mainApplicationFile: str
    mainClass: str
    sparkConf: _containers.ScalarMap[str, str]
    hadoopConf: _containers.ScalarMap[str, str]
    executorPath: str
    databricksConf: _struct_pb2.Struct
    databricksToken: str
    databricksInstance: str
    driverPod: _tasks_pb2.K8sPod
    executorPod: _tasks_pb2.K8sPod
    def __init__(self, applicationType: _Optional[_Union[SparkApplication.Type, str]] = ..., mainApplicationFile: _Optional[str] = ..., mainClass: _Optional[str] = ..., sparkConf: _Optional[_Mapping[str, str]] = ..., hadoopConf: _Optional[_Mapping[str, str]] = ..., executorPath: _Optional[str] = ..., databricksConf: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., databricksToken: _Optional[str] = ..., databricksInstance: _Optional[str] = ..., driverPod: _Optional[_Union[_tasks_pb2.K8sPod, _Mapping]] = ..., executorPod: _Optional[_Union[_tasks_pb2.K8sPod, _Mapping]] = ...) -> None: ...
