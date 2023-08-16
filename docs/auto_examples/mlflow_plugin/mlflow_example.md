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

(mlflow_example)=

# MLflow Example

MLflow is a platform to streamline machine learning development, including tracking experiments, packaging code into reproducible runs, and sharing and deploying models.

Flyte provides an easy-to-use interface to log the task's metrics and parameters to either Flyte Deck or MLflow server.

```{code-cell}
import mlflow.keras
import tensorflow as tf
```

+++ {"lines_to_next_cell": 0}

Let's first import the libraries.

```{code-cell}
from flytekit import task, workflow
from flytekitplugins.mlflow import mlflow_autolog
```

+++ {"lines_to_next_cell": 0}

Run a model training here and generate metrics and parameters.
Add `mlflow_autolog` to the task, then flyte will automatically log the metric to the Flyte Deck.

```{code-cell}
@task(disable_deck=False)
@mlflow_autolog(framework=mlflow.keras)
def train_model(epochs: int):
    # Refer to https://www.tensorflow.org/tutorials/keras/classification
    fashion_mnist = tf.keras.datasets.fashion_mnist
    (train_images, train_labels), (_, _) = fashion_mnist.load_data()
    train_images = train_images / 255.0

    model = tf.keras.Sequential(
        [
            tf.keras.layers.Flatten(input_shape=(28, 28)),
            tf.keras.layers.Dense(128, activation="relu"),
            tf.keras.layers.Dense(10),
        ]
    )

    model.compile(
        optimizer="adam",
        loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True),
        metrics=["accuracy"],
    )
    model.fit(train_images, train_labels, epochs=epochs)
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/f4b53a550bed70d9d7722d523e0b7568b781fc7d/flytesnacks/integrations/mlflow/metrics.png
:alt: Model Metrics
:class: with-shadow
:::

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/f4b53a550bed70d9d7722d523e0b7568b781fc7d/flytesnacks/integrations/mlflow/params.png
:alt: Model Parameters
:class: with-shadow
:::

+++ {"lines_to_next_cell": 0}

Finally, we put everything together into a workflow:

```{code-cell}
@workflow
def ml_pipeline(epochs: int):
    train_model(epochs=epochs)


if __name__ == "__main__":
    print(f"Running {__file__} main...")
    ml_pipeline(epochs=5)
```
