from collections import OrderedDict
from functools import partial

import numpy as np
import pytest
from sklearn.base import BaseEstimator
from sklearn.linear_model import LinearRegression
from sklearn.svm import SVC

import flytekit
from flytekit import task
from flytekit.configuration import Image, ImageConfig
from flytekit.core import context_manager
from flytekit.extras.sklearn import SklearnEstimatorTransformer
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


def get_model(model_type: str) -> BaseEstimator:
    models_map = {
        "lr": LinearRegression,
        "svc": partial(SVC, kernel="linear"),
    }
    x = np.random.normal(size=(10, 2))
    y = np.random.randint(2, size=(10,))
    while len(set(y)) < 2:
        y = np.random.randint(2, size=(10,))

    model = models_map[model_type]()
    model.fit(x, y)
    return model


@pytest.mark.parametrize(
    "transformer,python_type,format",
    [
        (SklearnEstimatorTransformer(), BaseEstimator, SklearnEstimatorTransformer.SKLEARN_FORMAT),
    ],
)
def test_get_literal_type(transformer, python_type, format):
    tf = transformer
    lt = tf.get_literal_type(python_type)
    assert lt == LiteralType(blob=BlobType(format=format, dimensionality=BlobType.BlobDimensionality.SINGLE))


@pytest.mark.parametrize(
    "transformer,python_type,format,python_val",
    [
        (
            SklearnEstimatorTransformer(),
            BaseEstimator,
            SklearnEstimatorTransformer.SKLEARN_FORMAT,
            get_model("lr"),
        ),
        (
            SklearnEstimatorTransformer(),
            BaseEstimator,
            SklearnEstimatorTransformer.SKLEARN_FORMAT,
            get_model("svc"),
        ),
    ],
)
def test_to_python_value_and_literal(transformer, python_type, format, python_val):
    ctx = context_manager.FlyteContext.current_context()
    tf = transformer
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

    np.testing.assert_array_equal(output.coef_, python_val.coef_)


def test_example_estimator():
    @task
    def t1() -> BaseEstimator:
        return get_model("lr")

    task_spec = get_serializable(OrderedDict(), serialization_settings, t1)
    assert task_spec.template.interface.outputs["o0"].type.blob.format is SklearnEstimatorTransformer.SKLEARN_FORMAT
