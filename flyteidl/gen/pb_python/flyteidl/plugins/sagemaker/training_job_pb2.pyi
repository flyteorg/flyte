from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class InputMode(_message.Message):
    __slots__ = []
    class Value(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        FILE: _ClassVar[InputMode.Value]
        PIPE: _ClassVar[InputMode.Value]
    FILE: InputMode.Value
    PIPE: InputMode.Value
    def __init__(self) -> None: ...

class AlgorithmName(_message.Message):
    __slots__ = []
    class Value(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        CUSTOM: _ClassVar[AlgorithmName.Value]
        XGBOOST: _ClassVar[AlgorithmName.Value]
    CUSTOM: AlgorithmName.Value
    XGBOOST: AlgorithmName.Value
    def __init__(self) -> None: ...

class InputContentType(_message.Message):
    __slots__ = []
    class Value(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        TEXT_CSV: _ClassVar[InputContentType.Value]
    TEXT_CSV: InputContentType.Value
    def __init__(self) -> None: ...

class MetricDefinition(_message.Message):
    __slots__ = ["name", "regex"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    REGEX_FIELD_NUMBER: _ClassVar[int]
    name: str
    regex: str
    def __init__(self, name: _Optional[str] = ..., regex: _Optional[str] = ...) -> None: ...

class AlgorithmSpecification(_message.Message):
    __slots__ = ["input_mode", "algorithm_name", "algorithm_version", "metric_definitions", "input_content_type"]
    INPUT_MODE_FIELD_NUMBER: _ClassVar[int]
    ALGORITHM_NAME_FIELD_NUMBER: _ClassVar[int]
    ALGORITHM_VERSION_FIELD_NUMBER: _ClassVar[int]
    METRIC_DEFINITIONS_FIELD_NUMBER: _ClassVar[int]
    INPUT_CONTENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    input_mode: InputMode.Value
    algorithm_name: AlgorithmName.Value
    algorithm_version: str
    metric_definitions: _containers.RepeatedCompositeFieldContainer[MetricDefinition]
    input_content_type: InputContentType.Value
    def __init__(self, input_mode: _Optional[_Union[InputMode.Value, str]] = ..., algorithm_name: _Optional[_Union[AlgorithmName.Value, str]] = ..., algorithm_version: _Optional[str] = ..., metric_definitions: _Optional[_Iterable[_Union[MetricDefinition, _Mapping]]] = ..., input_content_type: _Optional[_Union[InputContentType.Value, str]] = ...) -> None: ...

class DistributedProtocol(_message.Message):
    __slots__ = []
    class Value(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNSPECIFIED: _ClassVar[DistributedProtocol.Value]
        MPI: _ClassVar[DistributedProtocol.Value]
    UNSPECIFIED: DistributedProtocol.Value
    MPI: DistributedProtocol.Value
    def __init__(self) -> None: ...

class TrainingJobResourceConfig(_message.Message):
    __slots__ = ["instance_count", "instance_type", "volume_size_in_gb", "distributed_protocol"]
    INSTANCE_COUNT_FIELD_NUMBER: _ClassVar[int]
    INSTANCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    VOLUME_SIZE_IN_GB_FIELD_NUMBER: _ClassVar[int]
    DISTRIBUTED_PROTOCOL_FIELD_NUMBER: _ClassVar[int]
    instance_count: int
    instance_type: str
    volume_size_in_gb: int
    distributed_protocol: DistributedProtocol.Value
    def __init__(self, instance_count: _Optional[int] = ..., instance_type: _Optional[str] = ..., volume_size_in_gb: _Optional[int] = ..., distributed_protocol: _Optional[_Union[DistributedProtocol.Value, str]] = ...) -> None: ...

class TrainingJob(_message.Message):
    __slots__ = ["algorithm_specification", "training_job_resource_config"]
    ALGORITHM_SPECIFICATION_FIELD_NUMBER: _ClassVar[int]
    TRAINING_JOB_RESOURCE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    algorithm_specification: AlgorithmSpecification
    training_job_resource_config: TrainingJobResourceConfig
    def __init__(self, algorithm_specification: _Optional[_Union[AlgorithmSpecification, _Mapping]] = ..., training_job_resource_config: _Optional[_Union[TrainingJobResourceConfig, _Mapping]] = ...) -> None: ...
