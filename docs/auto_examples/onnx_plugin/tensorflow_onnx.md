---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

# TensorFlow Example

In this example, we will see how to convert a tensorflow model to an ONNX model.

First import the necessary libraries.

```{code-cell}
from typing import List, NamedTuple

import numpy as np
import onnxruntime as rt
import tensorflow as tf
from flytekit import Resources, task, workflow
from flytekit.types.file import ONNXFile
from flytekitplugins.onnxtensorflow import TensorFlow2ONNX, TensorFlow2ONNXConfig
from tensorflow.keras import datasets, layers, models
from typing_extensions import Annotated
```

+++ {"lines_to_next_cell": 0}

Define a `NamedTuple` to define the data schema.

```{code-cell}
:lines_to_next_cell: 1

DataOutput = NamedTuple(
    "DataOutput",
    [
        ("train_images", np.ndarray),
        ("train_labels", np.ndarray),
        ("test_images", np.ndarray),
    ],
)
```

+++ {"lines_to_next_cell": 0}

Define a `load_data` task to load CIFAR10 data.

```{code-cell}
@task(cache=True, cache_version="0.0.2", requests=Resources(mem="1000Mi", cpu="2"))
def load_data() -> DataOutput:
    (train_images, train_labels), (test_images, _) = datasets.cifar10.load_data()

    # Normalize pixel values to be between 0 and 1
    train_images, test_images = train_images[:1000] / 255.0, test_images[:100] / 255.0
    train_labels = train_labels[:1000]

    return DataOutput(train_images=train_images, train_labels=train_labels, test_images=test_images)
```

+++ {"lines_to_next_cell": 0}

Define a `train` task to train a CNN model on the CIFAR10 dataset.
Note the annotated output type.
This is a special annotation that tells Flytekit that this parameter is to be converted to an ONNX model with the given metadata.

```{code-cell}
@task(requests=Resources(mem="1000Mi", cpu="2"))
def train(
    train_images: np.ndarray, train_labels: np.ndarray
) -> Annotated[
    TensorFlow2ONNX,
    TensorFlow2ONNXConfig(
        input_signature=(tf.TensorSpec((None, 32, 32, 3), tf.double, name="input"),),
        opset=13,
    ),
]:
    model = models.Sequential()
    model.add(layers.Conv2D(32, (3, 3), activation="relu", input_shape=(32, 32, 3)))
    model.add(layers.MaxPooling2D((2, 2)))
    model.add(layers.Conv2D(64, (3, 3), activation="relu"))
    model.add(layers.Flatten())
    model.add(layers.Dense(64, activation="relu"))
    model.add(layers.Dense(10))

    model.compile(
        optimizer="adam",
        loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True),
        metrics=["accuracy"],
    )

    model.fit(train_images, train_labels, epochs=2)

    return TensorFlow2ONNX(model=model)
```

+++ {"lines_to_next_cell": 0}

Define an `onnx_predict` task to generate predictions for the test data using the ONNX model.

```{code-cell}
@task(requests=Resources(mem="1000Mi", cpu="2"))
def onnx_predict(
    model: ONNXFile,
    test_images: np.ndarray,
) -> List[np.ndarray]:
    m = rt.InferenceSession(model.download(), providers=["CPUExecutionProvider"])
    onnx_pred = m.run([n.name for n in m.get_outputs()], {"input": test_images})

    return onnx_pred
```

+++ {"lines_to_next_cell": 0}

Define a workflow to run the tasks.

```{code-cell}
@workflow
def wf() -> List[np.ndarray]:
    load_data_output = load_data()
    model = train(
        train_images=load_data_output.train_images,
        train_labels=load_data_output.train_labels,
    )
    onnx_preds = onnx_predict(model=model, test_images=load_data_output.test_images)
    return onnx_preds
```

+++ {"lines_to_next_cell": 0}

Run the workflow locally.

```{code-cell}
if __name__ == "__main__":
    print(f"Predictions: {wf()}")
```
