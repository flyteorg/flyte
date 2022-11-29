from flyteidl.plugins.sagemaker import parameter_ranges_pb2 as _parameter_ranges_pb2
from flyteidl.plugins.sagemaker import training_job_pb2 as _training_job_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class HyperparameterTuningJob(_message.Message):
    __slots__ = ["max_number_of_training_jobs", "max_parallel_training_jobs", "training_job"]
    MAX_NUMBER_OF_TRAINING_JOBS_FIELD_NUMBER: _ClassVar[int]
    MAX_PARALLEL_TRAINING_JOBS_FIELD_NUMBER: _ClassVar[int]
    TRAINING_JOB_FIELD_NUMBER: _ClassVar[int]
    max_number_of_training_jobs: int
    max_parallel_training_jobs: int
    training_job: _training_job_pb2.TrainingJob
    def __init__(self, training_job: _Optional[_Union[_training_job_pb2.TrainingJob, _Mapping]] = ..., max_number_of_training_jobs: _Optional[int] = ..., max_parallel_training_jobs: _Optional[int] = ...) -> None: ...

class HyperparameterTuningJobConfig(_message.Message):
    __slots__ = ["hyperparameter_ranges", "training_job_early_stopping_type", "tuning_objective", "tuning_strategy"]
    HYPERPARAMETER_RANGES_FIELD_NUMBER: _ClassVar[int]
    TRAINING_JOB_EARLY_STOPPING_TYPE_FIELD_NUMBER: _ClassVar[int]
    TUNING_OBJECTIVE_FIELD_NUMBER: _ClassVar[int]
    TUNING_STRATEGY_FIELD_NUMBER: _ClassVar[int]
    hyperparameter_ranges: _parameter_ranges_pb2.ParameterRanges
    training_job_early_stopping_type: TrainingJobEarlyStoppingType.Value
    tuning_objective: HyperparameterTuningObjective
    tuning_strategy: HyperparameterTuningStrategy.Value
    def __init__(self, hyperparameter_ranges: _Optional[_Union[_parameter_ranges_pb2.ParameterRanges, _Mapping]] = ..., tuning_strategy: _Optional[_Union[HyperparameterTuningStrategy.Value, str]] = ..., tuning_objective: _Optional[_Union[HyperparameterTuningObjective, _Mapping]] = ..., training_job_early_stopping_type: _Optional[_Union[TrainingJobEarlyStoppingType.Value, str]] = ...) -> None: ...

class HyperparameterTuningObjective(_message.Message):
    __slots__ = ["metric_name", "objective_type"]
    METRIC_NAME_FIELD_NUMBER: _ClassVar[int]
    OBJECTIVE_TYPE_FIELD_NUMBER: _ClassVar[int]
    metric_name: str
    objective_type: HyperparameterTuningObjectiveType.Value
    def __init__(self, objective_type: _Optional[_Union[HyperparameterTuningObjectiveType.Value, str]] = ..., metric_name: _Optional[str] = ...) -> None: ...

class HyperparameterTuningObjectiveType(_message.Message):
    __slots__ = []
    class Value(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    MAXIMIZE: HyperparameterTuningObjectiveType.Value
    MINIMIZE: HyperparameterTuningObjectiveType.Value
    def __init__(self) -> None: ...

class HyperparameterTuningStrategy(_message.Message):
    __slots__ = []
    class Value(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    BAYESIAN: HyperparameterTuningStrategy.Value
    RANDOM: HyperparameterTuningStrategy.Value
    def __init__(self) -> None: ...

class TrainingJobEarlyStoppingType(_message.Message):
    __slots__ = []
    class Value(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    AUTO: TrainingJobEarlyStoppingType.Value
    OFF: TrainingJobEarlyStoppingType.Value
    def __init__(self) -> None: ...
