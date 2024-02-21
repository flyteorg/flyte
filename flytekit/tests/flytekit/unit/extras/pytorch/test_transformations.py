from collections import OrderedDict

import pytest
import torch

import flytekit
from flytekit import task
from flytekit.configuration import Image, ImageConfig
from flytekit.core import context_manager
from flytekit.extras.pytorch import (
    PyTorchCheckpoint,
    PyTorchCheckpointTransformer,
    PyTorchModuleTransformer,
    PyTorchTensorTransformer,
)
from flytekit.models.core.types import BlobType
from flytekit.models.literals import BlobMetadata
from flytekit.models.types import LiteralType
from flytekit.tools.translator import get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


@pytest.mark.parametrize(
    "transformer,python_type,format",
    [
        (PyTorchTensorTransformer(), torch.Tensor, PyTorchTensorTransformer.PYTORCH_FORMAT),
        (PyTorchModuleTransformer(), torch.nn.Module, PyTorchModuleTransformer.PYTORCH_FORMAT),
        (PyTorchCheckpointTransformer(), PyTorchCheckpoint, PyTorchCheckpointTransformer.PYTORCH_CHECKPOINT_FORMAT),
    ],
)
def test_get_literal_type(transformer, python_type, format):
    tf = transformer
    lt = tf.get_literal_type(python_type)
    assert lt == LiteralType(blob=BlobType(format=format, dimensionality=BlobType.BlobDimensionality.SINGLE))
    assert tf.guess_python_type(lt) == python_type


@pytest.mark.parametrize(
    "transformer,python_type,format,python_val",
    [
        (
            PyTorchTensorTransformer(),
            torch.Tensor,
            PyTorchTensorTransformer.PYTORCH_FORMAT,
            torch.tensor([[1, 2], [3, 4]]),
        ),
        (
            PyTorchModuleTransformer(),
            torch.nn.Module,
            PyTorchModuleTransformer.PYTORCH_FORMAT,
            torch.nn.Linear(2, 2),
        ),
        (
            PyTorchCheckpointTransformer(),
            PyTorchCheckpoint,
            PyTorchCheckpointTransformer.PYTORCH_CHECKPOINT_FORMAT,
            PyTorchCheckpoint(
                module=torch.nn.Linear(2, 2),
                hyperparameters={"epochs": 10, "batch_size": 32},
                optimizer=torch.optim.Adam(torch.nn.Linear(2, 2).parameters()),
            ),
        ),
    ],
)
def test_to_python_value_and_literal(transformer, python_type, format, python_val):
    ctx = context_manager.FlyteContext.current_context()
    tf = transformer
    device = torch.device("cuda:0" if torch.cuda.is_available() else "cpu")
    python_val = python_val.to(device) if hasattr(python_val, "to") else python_val
    lt = tf.get_literal_type(python_type)

    lv = tf.to_literal(ctx, python_val, type(python_val), lt)  # type: ignore
    assert lv.scalar.blob.metadata == BlobMetadata(
        type=BlobType(
            format=format,
            dimensionality=BlobType.BlobDimensionality.SINGLE,
        )
    )
    assert lv.scalar.blob.uri is not None

    output = tf.to_python_value(ctx, lv, python_type)
    if isinstance(python_val, torch.Tensor):
        assert torch.equal(output, python_val)
    elif isinstance(python_val, torch.nn.Module):
        for p1, p2 in zip(output.parameters(), python_val.parameters()):
            if p1.data.ne(p2.data).sum() > 0:
                assert False
        assert True
    else:
        assert isinstance(output, dict)


def test_example_tensor():
    @task
    def t1(array: torch.Tensor) -> torch.Tensor:
        return torch.flatten(array)

    task_spec = get_serializable(OrderedDict(), serialization_settings, t1)
    assert task_spec.template.interface.outputs["o0"].type.blob.format is PyTorchTensorTransformer.PYTORCH_FORMAT


def test_example_module():
    @task
    def t1() -> torch.nn.Module:
        return torch.nn.BatchNorm1d(3, track_running_stats=True)

    task_spec = get_serializable(OrderedDict(), serialization_settings, t1)
    assert task_spec.template.interface.outputs["o0"].type.blob.format is PyTorchModuleTransformer.PYTORCH_FORMAT


def test_example_checkpoint():
    @task
    def t1() -> PyTorchCheckpoint:
        return PyTorchCheckpoint(
            module=torch.nn.Linear(2, 2),
            hyperparameters={"epochs": 10, "batch_size": 32},
            optimizer=torch.optim.Adam(torch.nn.Linear(2, 2).parameters()),
        )

    task_spec = get_serializable(OrderedDict(), serialization_settings, t1)
    assert (
        task_spec.template.interface.outputs["o0"].type.blob.format
        is PyTorchCheckpointTransformer.PYTORCH_CHECKPOINT_FORMAT
    )
