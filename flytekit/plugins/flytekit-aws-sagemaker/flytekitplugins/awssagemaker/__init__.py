"""
.. currentmodule:: flytekitplugins.awssagemaker

This package contains things that are useful when extending Flytekit.

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   AlgorithmName
   AlgorithmSpecification
   CategoricalParameterRange
   ContinuousParameterRange
   DISTRIBUTED_TRAINING_CONTEXT_KEY
   DistributedProtocol
   DistributedTrainingContext
   HPOJob
   HyperparameterScalingType
   HyperparameterTuningJobConfig
   HyperparameterTuningObjective
   HyperparameterTuningObjectiveType
   HyperparameterTuningStrategy
   InputContentType
   InputMode
   IntegerParameterRange
   ParameterRangeOneOf
   SagemakerCustomTrainingTask
   SagemakerHPOTask
   SagemakerTrainingJobConfig
   TrainingJobEarlyStoppingType
   TrainingJobResourceConfig
"""

__all__ = [
    "AlgorithmName",
    "AlgorithmSpecification",
    "CategoricalParameterRange",
    "ContinuousParameterRange",
    "DISTRIBUTED_TRAINING_CONTEXT_KEY",
    "DistributedProtocol",
    "DistributedTrainingContext",
    "HPOJob",
    "HyperparameterScalingType",
    "HyperparameterTuningJobConfig",
    "HyperparameterTuningObjective",
    "HyperparameterTuningObjectiveType",
    "HyperparameterTuningStrategy",
    "InputContentType",
    "InputMode",
    "IntegerParameterRange",
    "ParameterRangeOneOf",
    "SagemakerBuiltinAlgorithmsTask",
    "SagemakerCustomTrainingTask",
    "SagemakerHPOTask",
    "SagemakerTrainingJobConfig",
    "TrainingJobEarlyStoppingType",
    "TrainingJobResourceConfig",
]

from flytekitplugins.awssagemaker.models.hpo_job import (
    HyperparameterTuningJobConfig,
    HyperparameterTuningObjective,
    HyperparameterTuningObjectiveType,
    HyperparameterTuningStrategy,
    TrainingJobEarlyStoppingType,
)
from flytekitplugins.awssagemaker.models.parameter_ranges import (
    CategoricalParameterRange,
    ContinuousParameterRange,
    HyperparameterScalingType,
    IntegerParameterRange,
    ParameterRangeOneOf,
)
from flytekitplugins.awssagemaker.models.training_job import (
    AlgorithmName,
    AlgorithmSpecification,
    DistributedProtocol,
    InputContentType,
    InputMode,
    TrainingJobResourceConfig,
)

from .distributed_training import DISTRIBUTED_TRAINING_CONTEXT_KEY, DistributedTrainingContext
from .hpo import HPOJob, SagemakerHPOTask
from .training import SagemakerBuiltinAlgorithmsTask, SagemakerCustomTrainingTask, SagemakerTrainingJobConfig
