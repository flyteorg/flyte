import mlflow
import tensorflow as tf
from flytekitplugins.mlflow import mlflow_autolog

import flytekit
from flytekit import task


@task(enable_deck=True)
@mlflow_autolog(framework=mlflow.keras)
def train_model(epochs: int):
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
        optimizer="adam", loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True), metrics=["accuracy"]
    )
    model.fit(train_images, train_labels, epochs=epochs)


def test_local_exec():
    train_model(epochs=1)
    assert len(flytekit.current_context().decks) == 5  # mlflow metrics, params, timeline, input, and output
