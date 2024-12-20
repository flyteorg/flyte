from collections import OrderedDict

import numpy as np
import pytest
import tensorflow as tf

import flytekit
from flytekit import task
from flytekit.configuration import Image, ImageConfig
from flytekit.core import context_manager
from flytekit.extras.tensorflow import TensorFlowModelTransformer
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


def get_tf_model() -> tf.keras.Model:
    inputs = tf.keras.Input(shape=(32,))
    outputs = tf.keras.layers.Dense(1)(inputs)
    tf_model = tf.keras.Model(inputs, outputs)
    return tf_model


@pytest.mark.parametrize(
    "transformer,python_type,format",
    [
        (TensorFlowModelTransformer(), tf.keras.Model, TensorFlowModelTransformer.TENSORFLOW_FORMAT),
    ],
)
def test_get_literal_type(transformer, python_type, format):
    lt = transformer.get_literal_type(python_type)
    assert lt == LiteralType(blob=BlobType(format=format, dimensionality=BlobType.BlobDimensionality.MULTIPART))


@pytest.mark.parametrize(
    "transformer,python_type,format,python_val",
    [
        (TensorFlowModelTransformer(), tf.keras.Model, TensorFlowModelTransformer.TENSORFLOW_FORMAT, get_tf_model()),
    ],
)
def test_to_python_value_and_literal(transformer, python_type, format, python_val):
    ctx = context_manager.FlyteContext.current_context()
    lt = transformer.get_literal_type(python_type)

    lv = transformer.to_literal(ctx, python_val, type(python_val), lt)  # type: ignore
    output = transformer.to_python_value(ctx, lv, python_type)

    assert lv.scalar.blob.metadata == BlobMetadata(
        type=BlobType(
            format=format,
            dimensionality=BlobType.BlobDimensionality.MULTIPART,
        )
    )
    assert lv.scalar.blob.uri is not None
    for w1, w2 in zip(output.weights, python_val.weights):
        np.testing.assert_allclose(w1.numpy(), w2.numpy())


def test_example_model():
    @task
    def t1() -> tf.keras.Model:
        return get_tf_model()

    task_spec = get_serializable(OrderedDict(), serialization_settings, t1)
    assert task_spec.template.interface.outputs["o0"].type.blob.format is TensorFlowModelTransformer.TENSORFLOW_FORMAT
