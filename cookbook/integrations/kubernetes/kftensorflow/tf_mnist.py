"""
Distributed TensorFlow Training
-------------------------------

When you need to scale up model training using TensorFlow, you can use :py:class:`~tensorflow:tf.distribute.Strategy` to distribute your training across multiple devices.
There are various strategies available under this API and you can use any of them. In this example, we will use :py:class:`~tensorflow:tf.distribute.MirroredStrategy` to train an MNIST model using a convolutional network.

:py:class:`~tensorflow:tf.distribute.MirroredStrategy` supports synchronous distributed training on multiple GPUs on one machine.
To learn more about distributed training with TensorFlow, refer to the `Distributed training with TensorFlow <https://www.tensorflow.org/guide/distributed_training>`__ in the TensorFlow documentation.

Let's get started with an example!
"""

# %%
# First, we load the libraries.
import os
from dataclasses import dataclass
from typing import NamedTuple, Tuple

import tensorflow as tf
import tensorflow_datasets as tfds
from dataclasses_json import dataclass_json
from flytekit import Resources, task, workflow
from flytekit.types.directory import FlyteDirectory
from flytekitplugins.kftensorflow import TfJob

# %%
# We define ``MODEL_FILE_PATH`` indicating where to store the model file.
MODEL_FILE_PATH = "saved_model/"

# %%
# We initialize a data class to store the hyperparameters.


@dataclass_json
@dataclass
class Hyperparameters(object):
    batch_size_per_replica: int = 64
    buffer_size: int = 10000
    epochs: int = 10


# %%
# Loading the Data
# ================
#
# We use the `MNIST <https://www.tensorflow.org/datasets/catalog/mnist>`__ dataset to train our model.
def load_data(
    hyperparameters: Hyperparameters,
) -> Tuple[tf.data.Dataset, tf.data.Dataset, tf.distribute.Strategy]:
    datasets, _ = tfds.load(name="mnist", with_info=True, as_supervised=True)
    mnist_train, mnist_test = datasets["train"], datasets["test"]

    strategy = tf.distribute.MirroredStrategy()
    print("Number of devices: {}".format(strategy.num_replicas_in_sync))

    # strategy.num_replicas_in_sync returns the number of replicas; helpful to utilize the extra compute power by increasing the batch size
    BATCH_SIZE = hyperparameters.batch_size_per_replica * strategy.num_replicas_in_sync

    def scale(image, label):
        image = tf.cast(image, tf.float32)
        image /= 255

        return image, label

    # fetch train and evaluation datasets
    train_dataset = (
        mnist_train.map(scale).shuffle(hyperparameters.buffer_size).batch(BATCH_SIZE)
    )
    eval_dataset = mnist_test.map(scale).batch(BATCH_SIZE)

    return train_dataset, eval_dataset, strategy


# %%
# Compiling the Model
# ===================
#
# We create and compile a model in the context of `Strategy.scope <https://www.tensorflow.org/api_docs/python/tf/distribute/MirroredStrategy#scope>`__.
def get_compiled_model(strategy: tf.distribute.Strategy) -> tf.keras.Model:
    with strategy.scope():
        model = tf.keras.Sequential(
            [
                tf.keras.layers.Conv2D(
                    32, 3, activation="relu", input_shape=(28, 28, 1)
                ),
                tf.keras.layers.MaxPooling2D(),
                tf.keras.layers.Flatten(),
                tf.keras.layers.Dense(64, activation="relu"),
                tf.keras.layers.Dense(10),
            ]
        )

        model.compile(
            loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True),
            optimizer=tf.keras.optimizers.Adam(),
            metrics=["accuracy"],
        )

    return model


# %%
# Training
# ========
#
# We define a function for decaying the learning rate.
def decay(epoch: int):
    if epoch < 3:
        return 1e-3
    elif epoch >= 3 and epoch < 7:
        return 1e-4
    else:
        return 1e-5


# %%
# Next, we define ``train_model`` to train the model with three callbacks:
#
# * :py:class:`~tensorflow:tf.keras.callbacks.TensorBoard` to log the training metrics
# * :py:class:`~tensorflow:tf.keras.callbacks.ModelCheckpoint` to save the model after every epoch
# * :py:class:`~tensorflow:tf.keras.callbacks.LearningRateScheduler` to decay the learning rate
def train_model(
    model: tf.keras.Model,
    train_dataset: tf.data.Dataset,
    hyperparameters: Hyperparameters,
) -> Tuple[tf.keras.Model, str]:
    # define the checkpoint directory to store the checkpoints
    checkpoint_dir = "./training_checkpoints"

    # define the name of the checkpoint files
    checkpoint_prefix = os.path.join(checkpoint_dir, "ckpt_{epoch}")

    # define a callback for printing the learning rate at the end of each epoch
    class PrintLR(tf.keras.callbacks.Callback):
        def on_epoch_end(self, epoch, logs=None):
            print(
                "\nLearning rate for epoch {} is {}".format(
                    epoch + 1, model.optimizer.lr.numpy()
                )
            )

    # put all the callbacks together
    callbacks = [
        tf.keras.callbacks.TensorBoard(log_dir="./logs"),
        tf.keras.callbacks.ModelCheckpoint(
            filepath=checkpoint_prefix, save_weights_only=True
        ),
        tf.keras.callbacks.LearningRateScheduler(decay),
        PrintLR(),
    ]

    # train the model
    model.fit(train_dataset, epochs=hyperparameters.epochs, callbacks=callbacks)

    # save the model
    model.save(MODEL_FILE_PATH, save_format="tf")

    return model, checkpoint_dir


# %%
# Evaluation
# ==========
#
# We define ``test_model`` to evaluate loss and accuracy on the test dataset.
def test_model(
    model: tf.keras.Model, checkpoint_dir: str, eval_dataset: tf.data.Dataset
) -> Tuple[float, float]:
    model.load_weights(tf.train.latest_checkpoint(checkpoint_dir))

    eval_loss, eval_acc = model.evaluate(eval_dataset)

    return eval_loss, eval_acc


# %%
# Defining an MNIST TensorFlow Task
# ==================================
#
# We initialize compute requirements and task output signature. Next, we define a ``mnist_tensorflow_job`` to kick off the training and evaluation process.
# The task is initialized with ``TFJob`` with certain values set:
#
# * ``num_workers``: integer determining the number of worker replicas to be spawned in the cluster for this job
# * ``num_ps_replicas``: number of parameter server replicas to use
# * ``num_chief_replicas``: number of chief replicas to use
#
# MirroredStrategy uses an all-reduce algorithm to communicate the variable updates across the devices.
# Hence, ``num_ps_replicas`` is not useful in our example.
#
# .. note::
#   If you'd like to understand the various Tensorflow strategies in distributed training, refer to the `Types of strategies <https://www.tensorflow.org/guide/distributed_training#types_of_strategies>`__ section in the TensorFlow documentation.
training_outputs = NamedTuple(
    "TrainingOutputs", accuracy=float, loss=float, model_state=FlyteDirectory
)

if os.getenv("SANDBOX") != "":
    resources = Resources(
        gpu="0", mem="1000Mi", storage="500Mi", ephemeral_storage="500Mi"
    )
else:
    resources = Resources(
        gpu="2", mem="10Gi", storage="10Gi", ephemeral_storage="500Mi"
    )


@task(
    task_config=TfJob(num_workers=2, num_ps_replicas=1, num_chief_replicas=1),
    retries=2,
    cache=True,
    cache_version="1.0",
    requests=resources,
    limits=resources,
)
def mnist_tensorflow_job(hyperparameters: Hyperparameters) -> training_outputs:
    train_dataset, eval_dataset, strategy = load_data(hyperparameters=hyperparameters)
    model = get_compiled_model(strategy=strategy)
    model, checkpoint_dir = train_model(
        model=model, train_dataset=train_dataset, hyperparameters=hyperparameters
    )
    eval_loss, eval_accuracy = test_model(
        model=model, checkpoint_dir=checkpoint_dir, eval_dataset=eval_dataset
    )
    return training_outputs(
        accuracy=eval_accuracy, loss=eval_loss, model_state=MODEL_FILE_PATH
    )


# %%
# Workflow
# ========
#
# Finally we define a workflow to call the ``mnist_tensorflow_job`` task.
@workflow
def mnist_tensorflow_workflow(
    hyperparameters: Hyperparameters = Hyperparameters(),
) -> training_outputs:
    return mnist_tensorflow_job(hyperparameters=hyperparameters)


# %%
# We can also run the code locally.
if __name__ == "__main__":
    print(mnist_tensorflow_workflow())
