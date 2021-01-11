"""
2. Custom training algorithms on Amazon Sagemaker
###################################################
This script shows an example of how you can simply convert your tensorflow training scripts to run on Amazon Sagemaker
with very few modifications
"""
import typing

import matplotlib.pyplot as plt
import tensorflow as tf
import tensorflow_datasets as tfds
from flytekit import task, workflow
from flytekit.taskplugins.sagemaker import (
    SagemakerTrainingJobConfig, TrainingJobResourceConfig, AlgorithmSpecification, InputMode, AlgorithmName,
    InputContentType)
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile

# %%
# Training Algorithm
# -------------------
# In this custom algorithm we will train MNIST using tensorflow.
#
# The trained model will be serialized using HDF5 encoding. Lets create a special type alias to denote this model
HDF5EncodedModelFile = FlyteFile[typing.TypeVar("hdf5")]

# %%
# The Training will produce 2 outputs
#
# #. The serialized model in HDF5 format
# #. And a log dictionary which is the Keras - `History.history`. This contains the accuracies and loss values
# #. Tensorboard Logs: We will also output A Directory that contains Tensorboard compatible logs. Flyte will collect
#    these logs and make them available for visualization in tensorboard - locally or if running remote.
#
# Refer to section :ref:`sagemaker_tensorboard` to visualize the outputs of this example.
TensorboardLogs = FlyteDirectory[typing.TypeVar("tensorboard")]
TrainingOutputs = typing.NamedTuple("TrainingOutputs", model=HDF5EncodedModelFile, epoch_logs=dict,
                                    logs=TensorboardLogs)


# %%
# Actual Algorithm
# ------------------
# To ensure that the code runs on Sagemaker, create a sagemaker task config using the class
# ``SagemakerTrainingJobConfig``
#
#  .. code::python
#
#       @task(
#        task_config=SagemakerTrainingJobConfig(
#         algorithm_specification=...,
#         training_job_resource_config=...,
#        )
def normalize_img(image, label):
    """Normalizes images: `uint8` -> `float32`."""
    return tf.cast(image, tf.float32) / 255., label


@task(
    task_config=SagemakerTrainingJobConfig(
        algorithm_specification=AlgorithmSpecification(
            input_mode=InputMode.FILE,
            algorithm_name=AlgorithmName.CUSTOM,
            algorithm_version="",
            input_content_type=InputContentType.TEXT_CSV,
        ),
        training_job_resource_config=TrainingJobResourceConfig(
            instance_type="ml.m4.xlarge",
            instance_count=1,
            volume_size_in_gb=25,
        ),
    ),
    cache_version="1.0",
    cache=True,
    container_image="{{.image.sagemaker.fqn}}:sagemaker-{{.image.default.version}}"
)
def custom_training_task(epochs: int, batch_size: int) -> TrainingOutputs:
    (ds_train, ds_test), ds_info = tfds.load(
        'mnist',
        split=['train', 'test'],
        shuffle_files=True,
        as_supervised=True,
        with_info=True,
    )

    ds_train = ds_train.map(
        normalize_img, num_parallel_calls=tf.data.experimental.AUTOTUNE)
    ds_train = ds_train.cache()
    ds_train = ds_train.shuffle(ds_info.splits['train'].num_examples)
    ds_train = ds_train.batch(batch_size)
    ds_train = ds_train.prefetch(tf.data.experimental.AUTOTUNE)

    ds_test = ds_test.map(
        normalize_img, num_parallel_calls=tf.data.experimental.AUTOTUNE)
    ds_test = ds_test.batch(batch_size)
    ds_test = ds_test.cache()
    ds_test = ds_test.prefetch(tf.data.experimental.AUTOTUNE)

    model = tf.keras.models.Sequential([
        tf.keras.layers.Flatten(input_shape=(28, 28)),
        tf.keras.layers.Dense(128, activation='relu'),
        tf.keras.layers.Dense(10)
    ])
    model.compile(
        optimizer=tf.keras.optimizers.Adam(0.001),
        loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True),
        metrics=[tf.keras.metrics.SparseCategoricalAccuracy()],
    )

    log_dir = "/tmp/training-logs"
    tb_callback = tf.keras.callbacks.TensorBoard(log_dir=log_dir)

    history = model.fit(
        ds_train,
        epochs=epochs,
        validation_data=ds_test,
        callbacks=[tb_callback],
    )

    serialized_model = "my_model.h5"
    model.save(serialized_model, overwrite=True)

    return TrainingOutputs(model=HDF5EncodedModelFile(serialized_model), epoch_logs=history.history,
                           logs=TensorboardLogs(log_dir))


# %%
# Plot the metrics
# -----------------
# In the following task we will use the history logs from the training in the previous step and plot the curves using
# matplotlib. Images will be output as png.
PNGImage = FlyteFile[typing.TypeVar("png")]
PlotOutputs = typing.NamedTuple("PlotOutputs", accuracy=PNGImage, loss=PNGImage)


@task
def plot_loss_and_accuracy(epoch_logs: dict) -> PlotOutputs:
    # summarize history for accuracy
    plt.plot(epoch_logs['sparse_categorical_accuracy'])
    plt.plot(epoch_logs['val_sparse_categorical_accuracy'])
    plt.title('Sparse Categorical accuracy')
    plt.ylabel('accuracy')
    plt.xlabel('epoch')
    plt.legend(['train', 'test'], loc='upper left')
    accuracy_plot = "accuracy.png"
    plt.savefig(accuracy_plot)
    # summarize history for loss
    plt.plot(epoch_logs['loss'])
    plt.plot(epoch_logs['val_loss'])
    plt.title('model loss')
    plt.ylabel('loss')
    plt.xlabel('epoch')
    plt.legend(['train', 'test'], loc='upper left')
    loss_plot = "loss.png"
    plt.savefig(loss_plot)

    return PlotOutputs(accuracy=FlyteFile(accuracy_plot), loss=FlyteFile(loss_plot))


# %%
# The workflow takes in the hyperparams - in this case just the epochs and the batch_size and outputs the trained model
# and the plotted curves
@workflow
def mnist_trainer(epochs: int = 5, batch_size: int = 128) -> (
        HDF5EncodedModelFile, PNGImage, PNGImage, TensorboardLogs):
    model, history, logs = custom_training_task(epochs=epochs, batch_size=batch_size)
    accuracy, loss = plot_loss_and_accuracy(epoch_logs=history)
    return model, accuracy, loss, logs


# %%
# As long as you have tensorflow setup locally, it will run like a regular python script
if __name__ == "__main__":
    model, accurracy, loss, logs = mnist_trainer()
    print(f"Model: {model}, Accuracy PNG: {accurracy}, loss PNG: {loss}, Tensorboard Log Dir: {logs}")

# %%
#
# .. _sagemaker_tensorboard:
#
# Rendering the output logs in tensorboard
# -----------------------------------------
# When running locally, the output of execution looks like
#
# .. code-block::
#
#   Model: /tmp/flyte/20210110_214129/mock_remote/8421ae4d041f76488e245edf3f4360d5/my_model.h5, Accuracy PNG: /tmp/flyte/20210110_214129/mock_remote/cf6a2cd9d3ded89ed814278a8fb3678c/accuracy.png, loss PNG: /tmp/flyte/20210110_214129/mock_remote/267c9dd17d4d4e7c9c8bb8b12ef1e3d2/loss.png, Tensorboard Log Dir: /tmp/flyte/20210110_214129/mock_remote/a4b04e58e21f26f08f81df24094d6446/
#
# You can use the ``Tensorboard Log Dir: /tmp/flyte/20210110_214129/mock_remote/a4b04e58e21f26f08f81df24094d6446/`` as
# an input to tensorboard to visualize the training as follows
#
# .. prompt:: bash
#
#   tensorboard --logdir /tmp/flyte/20210110_214129/mock_remote/a4b04e58e21f26f08f81df24094d6446/
#
#
# If running remotely (executing on Flyte hosted environment), the workflow execution outputs can be retrieved.
# Refer to .. TODO.
# You can retrieve the outputs - which will be a path to a blob store like S3, GCS, minio, etc. Tensorboad can be
# pointed to on your local laptop to visualize the results.
