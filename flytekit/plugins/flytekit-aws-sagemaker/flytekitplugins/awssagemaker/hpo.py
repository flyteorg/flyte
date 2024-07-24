import json
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Type, Union

from flyteidl.plugins.sagemaker import hyperparameter_tuning_job_pb2 as _pb2_hpo_job
from flyteidl.plugins.sagemaker import parameter_ranges_pb2 as _pb2_params
from flytekitplugins.awssagemaker.training import SagemakerBuiltinAlgorithmsTask, SagemakerCustomTrainingTask
from google.protobuf import json_format
from google.protobuf.json_format import MessageToDict

from flytekit import FlyteContext
from flytekit.configuration import SerializationSettings
from flytekit.extend import DictTransformer, PythonTask, TypeEngine, TypeTransformer
from flytekit.models.literals import Literal
from flytekit.models.types import LiteralType, SimpleType

from .models import hpo_job as _hpo_job_model
from .models import parameter_ranges as _params
from .models import training_job as _training_job_model


@dataclass
class HPOJob(object):
    """
    HPOJob Configuration should be used to configure the HPO Job.

    Args:
        max_number_of_training_jobs: maximum number of jobs to run for a training round
        max_parallel_training_jobs: limits the concurrency of the training jobs
        tunable_params: [optional] should be a list of parameters for which we want to provide the tuning ranges
    """

    max_number_of_training_jobs: int
    max_parallel_training_jobs: int
    # TODO. we could make the tunable params a tuple of name and type of range?
    tunable_params: Optional[List[str]] = None


# TODO Not done yet, but once we clean this up, the client interface should be simplified. The interface should
# Just take a list of Union of different types of Parameter Ranges. Lets see how simplify that
class SagemakerHPOTask(PythonTask[HPOJob]):
    _SAGEMAKER_HYPERPARAMETER_TUNING_JOB_TASK = "sagemaker_hyperparameter_tuning_job_task"

    def __init__(
        self,
        name: str,
        task_config: HPOJob,
        training_task: Union[SagemakerCustomTrainingTask, SagemakerBuiltinAlgorithmsTask],
        **kwargs,
    ):
        if training_task is None or not (
            isinstance(training_task, SagemakerCustomTrainingTask)
            or isinstance(training_task, SagemakerBuiltinAlgorithmsTask)
        ):
            raise ValueError(
                "Training Task of type SagemakerCustomTrainingTask/SagemakerBuiltinAlgorithmsTask is required to work"
                " with Sagemaker HPO"
            )

        self._task_config = task_config
        self._training_task = training_task

        extra_inputs = {"hyperparameter_tuning_job_config": _hpo_job_model.HyperparameterTuningJobConfig}

        if task_config.tunable_params:
            extra_inputs.update({param: _params.ParameterRangeOneOf for param in task_config.tunable_params})

        iface = training_task.python_interface
        updated_iface = iface.with_inputs(extra_inputs)
        super().__init__(
            task_type=self._SAGEMAKER_HYPERPARAMETER_TUNING_JOB_TASK,
            name=name,
            interface=updated_iface,
            task_config=task_config,
            **kwargs,
        )

    def execute(self, **kwargs) -> Any:
        raise NotImplementedError("Sagemaker HPO Task cannot be executed locally, to execute locally mock it!")

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        training_job = _training_job_model.TrainingJob(
            algorithm_specification=self._training_task.task_config.algorithm_specification,
            training_job_resource_config=self._training_task.task_config.training_job_resource_config,
        )
        return MessageToDict(
            _hpo_job_model.HyperparameterTuningJob(
                max_number_of_training_jobs=self.task_config.max_number_of_training_jobs,
                max_parallel_training_jobs=self.task_config.max_parallel_training_jobs,
                training_job=training_job,
            ).to_flyte_idl()
        )


# %%
# HPO Task allows ParameterRangeOneOf and HyperparameterTuningJobConfig as inputs. In flytekit this is possible
# to allow these two types to be registered as valid input / output types and provide a custom transformer
# We will create custom transformers for them as follows and provide them once a user loads HPO task


class HPOTuningJobConfigTransformer(TypeTransformer[_hpo_job_model.HyperparameterTuningJobConfig]):
    """
    Transformer to make ``HyperparameterTuningJobConfig`` an accepted value, for which a transformer is registered
    """

    def __init__(self):
        super().__init__("sagemaker-hpojobconfig-transformer", _hpo_job_model.HyperparameterTuningJobConfig)

    def get_literal_type(self, t: Type[_hpo_job_model.HyperparameterTuningJobConfig]) -> LiteralType:
        return LiteralType(simple=SimpleType.STRUCT, metadata=None)

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: _hpo_job_model.HyperparameterTuningJobConfig,
        python_type: Type[_hpo_job_model.HyperparameterTuningJobConfig],
        expected: LiteralType,
    ) -> Literal:
        d = MessageToDict(python_val.to_flyte_idl())
        return DictTransformer.dict_to_generic_literal(d)

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[_hpo_job_model.HyperparameterTuningJobConfig]
    ) -> _hpo_job_model.HyperparameterTuningJobConfig:
        if lv and lv.scalar and lv.scalar.generic is not None:
            d = json.loads(json_format.MessageToJson(lv.scalar.generic))
            o = _pb2_hpo_job.HyperparameterTuningJobConfig()
            o = json_format.ParseDict(d, o)
            return _hpo_job_model.HyperparameterTuningJobConfig.from_flyte_idl(o)
        return None


class ParameterRangesTransformer(TypeTransformer[_params.ParameterRangeOneOf]):
    """
    Transformer to make ``ParameterRange`` an accepted value, for which a transformer is registered
    """

    def __init__(self):
        super().__init__("sagemaker-paramrange-transformer", _params.ParameterRangeOneOf)

    def get_literal_type(self, t: Type[_params.ParameterRangeOneOf]) -> LiteralType:
        return LiteralType(simple=SimpleType.STRUCT, metadata=None)

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: _params.ParameterRangeOneOf,
        python_type: Type[_hpo_job_model.HyperparameterTuningJobConfig],
        expected: LiteralType,
    ) -> Literal:
        d = MessageToDict(python_val.to_flyte_idl())
        return DictTransformer.dict_to_generic_literal(d)

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[_params.ParameterRangeOneOf]
    ) -> _params.ParameterRangeOneOf:
        if lv and lv.scalar and lv.scalar.generic is not None:
            d = json.loads(json_format.MessageToJson(lv.scalar.generic))
            o = _pb2_params.ParameterRangeOneOf()
            o = json_format.ParseDict(d, o)
            return _params.ParameterRangeOneOf.from_flyte_idl(o)
        return None


# %%
# Register the types
TypeEngine.register(HPOTuningJobConfigTransformer())
TypeEngine.register(ParameterRangesTransformer())
