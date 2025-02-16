from typing import List, NamedTuple

import numpy
import onnxruntime as rt
import pandas as pd
from flytekitplugins.onnxscikitlearn import ScikitLearn2ONNX, ScikitLearn2ONNXConfig
from skl2onnx.common._apply_operation import apply_mul
from skl2onnx.common.data_types import FloatTensorType
from skl2onnx.proto import onnx_proto
from sklearn.base import BaseEstimator, TransformerMixin
from sklearn.datasets import load_iris
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
from typing_extensions import Annotated

from flytekit import task, workflow
from flytekit.types.file import ONNXFile


def test_onnx_scikitlearn_simple():
    TrainOutput = NamedTuple(
        "TrainOutput",
        [
            (
                "model",
                Annotated[
                    ScikitLearn2ONNX,
                    ScikitLearn2ONNXConfig(
                        initial_types=[("float_input", FloatTensorType([None, 4]))],
                        target_opset=12,
                    ),
                ],
            ),
            ("test", pd.DataFrame),
        ],
    )

    @task
    def train() -> TrainOutput:
        iris = load_iris(as_frame=True)
        X, y = iris.data, iris.target
        X_train, X_test, y_train, _ = train_test_split(X, y)
        model = RandomForestClassifier()
        model.fit(X_train, y_train)

        return TrainOutput(test=X_test, model=ScikitLearn2ONNX(model))

    @task
    def predict(
        model: ONNXFile,
        X_test: pd.DataFrame,
    ) -> List[int]:
        sess = rt.InferenceSession(model.download())
        input_name = sess.get_inputs()[0].name
        label_name = sess.get_outputs()[0].name
        pred_onx = sess.run([label_name], {input_name: X_test.to_numpy(dtype=numpy.float32)})[0]
        return pred_onx.tolist()

    @workflow
    def wf() -> List[int]:
        train_output = train()
        return predict(model=train_output.model, X_test=train_output.test)

    print(wf())


class CustomTransform(BaseEstimator, TransformerMixin):
    def __init__(self):
        TransformerMixin.__init__(self)
        BaseEstimator.__init__(self)

    def fit(self, X, y, sample_weight=None):
        pass

    def transform(self, X):
        return X * numpy.array([[0.5, 0.1, 10], [0.5, 0.1, 10]]).T


def custom_transform_shape_calculator(operator):
    operator.outputs[0].type = FloatTensorType([3, 2])


def custom_tranform_converter(scope, operator, container):
    input = operator.inputs[0]
    output = operator.outputs[0]

    weights_name = scope.get_unique_variable_name("weights")
    atype = onnx_proto.TensorProto.FLOAT
    weights = [0.5, 0.1, 10]
    shape = [len(weights), 1]
    container.add_initializer(weights_name, atype, shape, weights)
    apply_mul(scope, [input.full_name, weights_name], output.full_name, container)


def test_onnx_scikitlearn():
    @task
    def get_model() -> (
        Annotated[
            ScikitLearn2ONNX,
            ScikitLearn2ONNXConfig(
                initial_types=[("input", FloatTensorType([None, numpy.array([[1, 2], [3, 4], [4, 5]]).shape[1]]))],
                custom_shape_calculators={CustomTransform: custom_transform_shape_calculator},
                custom_conversion_functions={CustomTransform: custom_tranform_converter},
                target_opset=12,
            ),
        ]
    ):
        model = CustomTransform()
        return ScikitLearn2ONNX(model)

    @workflow
    def wf() -> ONNXFile:
        return get_model()

    print(wf())
