from typing import List

from flyteidl.plugins.sagemaker import training_job_pb2 as _training_job_pb2

from flytekit.models import common as _common


class DistributedProtocol(object):
    """
    The distribution framework is used for determining which underlying distributed training mechanism to use.
    This is only required for use cases where the user wants to train its custom training job in a distributed manner
    """

    UNSPECIFIED = _training_job_pb2.DistributedProtocol.UNSPECIFIED
    MPI = _training_job_pb2.DistributedProtocol.MPI


class TrainingJobResourceConfig(_common.FlyteIdlEntity):
    """
    TrainingJobResourceConfig is a pass-through, specifying the instance type to use for the training job, the
    number of instances to launch, and the size of the ML storage volume the user wants to provision
    Refer to SageMaker official doc for more details: https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_CreateTrainingJob.html
    """

    def __init__(
        self,
        instance_type: str,
        volume_size_in_gb: int,
        instance_count: int = 1,
        distributed_protocol: int = DistributedProtocol.UNSPECIFIED,
    ):
        self._instance_count = instance_count
        self._instance_type = instance_type
        self._volume_size_in_gb = volume_size_in_gb
        self._distributed_protocol = distributed_protocol

    @property
    def instance_count(self) -> int:
        """
        The number of ML compute instances to use. For distributed training, provide a value greater than 1.
        :rtype: int
        """
        return self._instance_count

    @property
    def instance_type(self) -> str:
        """
        The ML compute instance type.
        :rtype: str
        """
        return self._instance_type

    @property
    def volume_size_in_gb(self) -> int:
        """
        The size of the ML storage volume that you want to provision to store the data and intermediate artifacts, etc.
        :rtype: int
        """
        return self._volume_size_in_gb

    @property
    def distributed_protocol(self) -> int:
        """
        The distribution framework is used to determine through which mechanism the distributed training is done.
        enum value from DistributionFramework.
        :rtype: int
        """
        return self._distributed_protocol

    def to_flyte_idl(self) -> _training_job_pb2.TrainingJobResourceConfig:
        """

        :rtype: _training_job_pb2.TrainingJobResourceConfig
        """
        return _training_job_pb2.TrainingJobResourceConfig(
            instance_count=self.instance_count,
            instance_type=self.instance_type,
            volume_size_in_gb=self.volume_size_in_gb,
            distributed_protocol=self.distributed_protocol,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _training_job_pb2.TrainingJobResourceConfig):
        """

        :param pb2_object:
        :rtype: TrainingJobResourceConfig
        """
        return cls(
            instance_count=pb2_object.instance_count,
            instance_type=pb2_object.instance_type,
            volume_size_in_gb=pb2_object.volume_size_in_gb,
            distributed_protocol=pb2_object.distributed_protocol,
        )


class MetricDefinition(_common.FlyteIdlEntity):
    def __init__(
        self,
        name: str,
        regex: str,
    ):
        self._name = name
        self._regex = regex

    @property
    def name(self) -> str:
        """
        The user-defined name of the metric
        :rtype: str
        """
        return self._name

    @property
    def regex(self) -> str:
        """
        SageMaker hyperparameter tuning using this regex to parses your algorithm’s stdout and stderr
        streams to find the algorithm metrics on which the users want to track
        :rtype: str
        """
        return self._regex

    def to_flyte_idl(self) -> _training_job_pb2.MetricDefinition:
        """

        :rtype: _training_job_pb2.MetricDefinition
        """
        return _training_job_pb2.MetricDefinition(
            name=self.name,
            regex=self.regex,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _training_job_pb2.MetricDefinition):
        """

        :param pb2_object: _training_job_pb2.MetricDefinition
        :rtype: MetricDefinition
        """
        return cls(
            name=pb2_object.name,
            regex=pb2_object.regex,
        )


# TODO Convert to Enum
class InputMode(object):
    """
    When using FILE input mode, different SageMaker built-in algorithms require different file types of input data
    See https://docs.aws.amazon.com/sagemaker/latest/dg/cdf-training.html
    https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
    """

    PIPE = _training_job_pb2.InputMode.PIPE
    FILE = _training_job_pb2.InputMode.FILE


# TODO Convert to enum
class AlgorithmName(object):
    """
    The algorithm name is used for deciding which pre-built image to point to.
    This is only required for use cases where SageMaker's built-in algorithm mode is used.
    While we currently only support a subset of the algorithms, more will be added to the list.
    See: https://docs.aws.amazon.com/sagemaker/latest/dg/algos.html
    """

    CUSTOM = _training_job_pb2.AlgorithmName.CUSTOM
    XGBOOST = _training_job_pb2.AlgorithmName.XGBOOST


# TODO convert to enum
class InputContentType(object):
    """
    Specifies the type of content for input data. Different SageMaker built-in algorithms require different content types of input data
    See https://docs.aws.amazon.com/sagemaker/latest/dg/cdf-training.html
    https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
    """

    TEXT_CSV = _training_job_pb2.InputContentType.TEXT_CSV


class AlgorithmSpecification(_common.FlyteIdlEntity):
    """
    Specifies the training algorithm to be used in the training job
    This object is mostly a pass-through, with a couple of exceptions include: (1) in Flyte, users don't need to specify
    TrainingImage; either use the built-in algorithm mode by using Flytekit's Simple Training Job and specifying an algorithm
    name and an algorithm version or (2) when users want to supply custom algorithms they should set algorithm_name field to
    CUSTOM. In this case, the value of the algorithm_version field has no effect
    For pass-through use cases: refer to this AWS official document for more details
    https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_AlgorithmSpecification.html
    """

    def __init__(
        self,
        algorithm_name: int = AlgorithmName.CUSTOM,
        algorithm_version: str = "",
        input_mode: int = InputMode.FILE,
        metric_definitions: List[MetricDefinition] = None,
        input_content_type: int = InputContentType.TEXT_CSV,
    ):
        self._input_mode = input_mode
        self._input_content_type = input_content_type
        self._algorithm_name = algorithm_name
        self._algorithm_version = algorithm_version
        self._metric_definitions = metric_definitions or []

    @property
    def input_mode(self) -> int:
        """
        enum value from InputMode. The input mode can be either PIPE or FILE
        :rtype: int
        """
        return self._input_mode

    @property
    def input_content_type(self) -> int:
        """
        enum value from InputContentType. The content type of the input data
        See https://docs.aws.amazon.com/sagemaker/latest/dg/cdf-training.html
        https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
        :rtype: int
        """
        return self._input_content_type

    @property
    def algorithm_name(self) -> int:
        """
        The algorithm name is used for deciding which pre-built image to point to.
        enum value from AlgorithmName.
        :rtype: int
        """
        return self._algorithm_name

    @property
    def algorithm_version(self) -> str:
        """
        version of the algorithm (if using built-in algorithm mode).
        :rtype: str
        """
        return self._algorithm_version

    @property
    def metric_definitions(self) -> List[MetricDefinition]:
        """
        A list of metric definitions for SageMaker to evaluate/track on the progress of the training job
        See this: https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_AlgorithmSpecification.html

        Note that, when you use one of the Amazon SageMaker built-in algorithms, you cannot define custom metrics.
        If you are doing hyperparameter tuning, built-in algorithms automatically send metrics to hyperparameter tuning.
        When using hyperparameter tuning, you do need to choose one of the metrics that the built-in algorithm emits as
        the objective metric for the tuning job.
        See this: https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-metrics.html
        :rtype: List[MetricDefinition]
        """
        return self._metric_definitions

    def to_flyte_idl(self) -> _training_job_pb2.AlgorithmSpecification:
        return _training_job_pb2.AlgorithmSpecification(
            input_mode=self.input_mode,
            algorithm_name=self.algorithm_name,
            algorithm_version=self.algorithm_version,
            metric_definitions=[m.to_flyte_idl() for m in self.metric_definitions],
            input_content_type=self.input_content_type,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _training_job_pb2.AlgorithmSpecification):
        return cls(
            input_mode=pb2_object.input_mode,
            algorithm_name=pb2_object.algorithm_name,
            algorithm_version=pb2_object.algorithm_version,
            metric_definitions=[MetricDefinition.from_flyte_idl(m) for m in pb2_object.metric_definitions],
            input_content_type=pb2_object.input_content_type,
        )


class TrainingJob(_common.FlyteIdlEntity):
    def __init__(
        self,
        algorithm_specification: AlgorithmSpecification,
        training_job_resource_config: TrainingJobResourceConfig,
    ):
        self._algorithm_specification = algorithm_specification
        self._training_job_resource_config = training_job_resource_config

    @property
    def algorithm_specification(self) -> AlgorithmSpecification:
        """
        Contains the information related to the algorithm to use in the training job
        :rtype: AlgorithmSpecification
        """
        return self._algorithm_specification

    @property
    def training_job_resource_config(self) -> TrainingJobResourceConfig:
        """
        Specifies the information around the instances that will be used to run the training job.
        :rtype: TrainingJobResourceConfig
        """
        return self._training_job_resource_config

    def to_flyte_idl(self) -> _training_job_pb2.TrainingJob:
        """
        :rtype: _training_job_pb2.TrainingJob
        """

        return _training_job_pb2.TrainingJob(
            algorithm_specification=self.algorithm_specification.to_flyte_idl()
            if self.algorithm_specification
            else None,
            training_job_resource_config=self.training_job_resource_config.to_flyte_idl()
            if self.training_job_resource_config
            else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _training_job_pb2.TrainingJob):
        """

        :param pb2_object:
        :rtype: TrainingJob
        """
        return cls(
            algorithm_specification=pb2_object.algorithm_specification,
            training_job_resource_config=pb2_object.training_job_resource_config,
        )
