import urllib
from io import BytesIO
from typing import List, NamedTuple

import numpy as np
import onnxruntime as rt
import tensorflow as tf
from flytekitplugins.onnxtensorflow import TensorFlow2ONNX, TensorFlow2ONNXConfig
from tensorflow.keras.applications.resnet50 import ResNet50, preprocess_input
from tensorflow.keras.preprocessing import image
from typing_extensions import Annotated

from flytekit import task, workflow
from flytekit.types.file import ONNXFile


def test_tf_onnx():
    @task
    def load_test_img() -> np.ndarray:
        with urllib.request.urlopen(
            "https://raw.githubusercontent.com/flyteorg/static-resources/main/flytekit/onnx/ade20k.jpg"
        ) as url:
            img = image.load_img(
                BytesIO(url.read()),
                target_size=(224, 224),
            )

        x = image.img_to_array(img)
        x = np.expand_dims(x, axis=0)
        x = preprocess_input(x)
        return x

    TrainPredictOutput = NamedTuple(
        "TrainPredictOutput",
        [
            ("predictions", np.ndarray),
            (
                "model",
                Annotated[
                    TensorFlow2ONNX,
                    TensorFlow2ONNXConfig(
                        input_signature=(tf.TensorSpec((None, 224, 224, 3), tf.float32, name="input"),), opset=13
                    ),
                ],
            ),
        ],
    )

    @task
    def train_and_predict(img: np.ndarray) -> TrainPredictOutput:
        model = ResNet50(weights="imagenet")

        preds = model.predict(img)
        return TrainPredictOutput(predictions=preds, model=TensorFlow2ONNX(model))

    @task
    def onnx_predict(
        model: ONNXFile,
        img: np.ndarray,
    ) -> List[np.ndarray]:
        m = rt.InferenceSession(model.download(), providers=["CPUExecutionProvider"])
        onnx_pred = m.run([n.name for n in m.get_outputs()], {"input": img})

        return onnx_pred

    WorkflowOutput = NamedTuple(
        "WorkflowOutput", [("keras_predictions", np.ndarray), ("onnx_predictions", List[np.ndarray])]
    )

    @workflow
    def wf() -> WorkflowOutput:
        img = load_test_img()
        train_predict_output = train_and_predict(img=img)
        onnx_preds = onnx_predict(model=train_predict_output.model, img=img)
        return WorkflowOutput(keras_predictions=train_predict_output.predictions, onnx_predictions=onnx_preds)

    predictions = wf()
    np.testing.assert_allclose(predictions.keras_predictions, predictions.onnx_predictions[0], rtol=1e-5)
