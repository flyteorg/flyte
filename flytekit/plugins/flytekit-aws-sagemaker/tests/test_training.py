import os
import pathlib
import tempfile

import pytest
from flytekitplugins.awssagemaker.distributed_training import setup_envars_for_testing
from flytekitplugins.awssagemaker.models.training_job import (
    AlgorithmName,
    AlgorithmSpecification,
    DistributedProtocol,
    TrainingJobResourceConfig,
)
from flytekitplugins.awssagemaker.training import SagemakerBuiltinAlgorithmsTask, SagemakerTrainingJobConfig

import flytekit
from flytekit import task
from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.context_manager import ExecutionParameters


def _get_reg_settings():
    default_img = Image(name="default", fqn="test", tag="tag")
    settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
    )
    return settings


def test_builtin_training():
    trainer = SagemakerBuiltinAlgorithmsTask(
        name="builtin-trainer",
        task_config=SagemakerTrainingJobConfig(
            training_job_resource_config=TrainingJobResourceConfig(
                instance_count=1,
                instance_type="ml-xlarge",
                volume_size_in_gb=1,
            ),
            algorithm_specification=AlgorithmSpecification(
                algorithm_name=AlgorithmName.XGBOOST,
            ),
        ),
    )

    assert trainer.python_interface.inputs.keys() == {"static_hyperparameters", "train", "validation"}
    assert trainer.python_interface.outputs.keys() == {"model"}

    with tempfile.TemporaryDirectory() as tmp:
        x = pathlib.Path(os.path.join(tmp, "x"))
        y = pathlib.Path(os.path.join(tmp, "y"))
        x.mkdir(parents=True, exist_ok=True)
        y.mkdir(parents=True, exist_ok=True)
        with pytest.raises(NotImplementedError):
            # Type engine doesn't support pathlib.Path yet
            trainer(static_hyperparameters={}, train=f"{x}", validation=f"{y}")

    assert trainer.get_custom(_get_reg_settings()) == {
        "algorithmSpecification": {"algorithmName": "XGBOOST"},
        "trainingJobResourceConfig": {"instanceCount": "1", "instanceType": "ml-xlarge", "volumeSizeInGb": "1"},
    }


def test_custom_training():
    @task(
        task_config=SagemakerTrainingJobConfig(
            training_job_resource_config=TrainingJobResourceConfig(
                instance_type="ml-xlarge",
                volume_size_in_gb=1,
            ),
            algorithm_specification=AlgorithmSpecification(
                algorithm_name=AlgorithmName.CUSTOM,
            ),
        )
    )
    def my_custom_trainer(x: int) -> int:
        return x

    assert my_custom_trainer.python_interface.inputs == {"x": int}
    assert my_custom_trainer.python_interface.outputs == {"o0": int}

    assert my_custom_trainer(x=10) == 10

    assert my_custom_trainer.get_custom(_get_reg_settings()) == {
        "algorithmSpecification": {},
        "trainingJobResourceConfig": {"instanceCount": "1", "instanceType": "ml-xlarge", "volumeSizeInGb": "1"},
    }


def test_distributed_custom_training():
    setup_envars_for_testing()

    @task(
        task_config=SagemakerTrainingJobConfig(
            training_job_resource_config=TrainingJobResourceConfig(
                instance_type="ml-xlarge",
                volume_size_in_gb=1,
                instance_count=2,  # Indicates distributed training
                distributed_protocol=DistributedProtocol.MPI,
            ),
            algorithm_specification=AlgorithmSpecification(
                algorithm_name=AlgorithmName.CUSTOM,
            ),
        )
    )
    def my_custom_trainer(x: int) -> int:
        assert flytekit.current_context().distributed_training_context is not None
        return x

    assert my_custom_trainer.python_interface.inputs == {"x": int}
    assert my_custom_trainer.python_interface.outputs == {"o0": int}

    assert my_custom_trainer(x=10) == 10

    assert my_custom_trainer._is_distributed() is True

    pb = ExecutionParameters.new_builder()
    pb.working_dir = "/tmp"
    p = pb.build()
    new_p = my_custom_trainer.pre_execute(p)
    assert new_p is not None
    assert new_p.has_attr("distributed_training_context")

    assert my_custom_trainer.get_custom(_get_reg_settings()) == {
        "algorithmSpecification": {},
        "trainingJobResourceConfig": {
            "distributedProtocol": "MPI",
            "instanceCount": "2",
            "instanceType": "ml-xlarge",
            "volumeSizeInGb": "1",
        },
    }
