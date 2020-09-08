from flytekit.sdk.tasks import (
    python_task,
    inputs,
    outputs,
)
from flytekit.sdk.types import Types
from flytekit.sdk.sagemaker.task import custom_training_job_task
from flytekit.sdk.workflow import workflow_class, Input, Output
from flytekit.models.sagemaker import training_job as training_job_models

import tensorflow as tf
from flytekit.common import utils as _common_utils


@inputs(dummy_train_dataset=Types.Blob, dummy_validation_dataset=Types.Blob, my_input=Types.String)
@outputs(out_model=Types.Blob, out=Types.Integer, out_extra_output_file=Types.Blob)
@custom_training_job_task(
    algorithm_specification=training_job_models.AlgorithmSpecification(
        input_mode=training_job_models.InputMode.FILE,
        algorithm_name=training_job_models.AlgorithmName.CUSTOM,
        algorithm_version="",
        input_content_type=training_job_models.InputContentType.TEXT_CSV,
    ),
    training_job_resource_config=training_job_models.TrainingJobResourceConfig(
        instance_type="ml.m4.xlarge",
        instance_count=1,
        volume_size_in_gb=25,
    )
)
def custom_training_task(wf_params, dummy_train_dataset, dummy_validation_dataset, my_input, out_model, out,
                         out_extra_output_file):
    with _common_utils.AutoDeletingTempDir("output_dir") as output_dir:
        wf_params.logging.info("My printed value: {}".format(my_input))
        wf_params.logging.info("My dummy train_dataset: {}".format(dummy_train_dataset))
        wf_params.logging.info("My dummy train_dataset: {}".format(dummy_validation_dataset))

        mnist = tf.keras.datasets.mnist

        (x_train, y_train), (x_test, y_test) = mnist.load_data()
        x_train, x_test = x_train / 255.0, x_test / 255.0

        model = tf.keras.models.Sequential([
            tf.keras.layers.Flatten(input_shape=(28, 28)),
            tf.keras.layers.Dense(128, activation='relu'),
            tf.keras.layers.Dropout(0.2),
            tf.keras.layers.Dense(10, activation='softmax')
        ])

        model.compile(
            optimizer='adam',
            loss='sparse_categorical_crossentropy',
            metrics=['accuracy'])

        model.fit(x_train, y_train, epochs=1)

        model.evaluate(x_test, y_test)

        print("Saving weights")
        tempfile_name = "my_model.h5"
        local_model_file = output_dir.get_named_tempfile(tempfile_name)
        model.save_weights(local_model_file)

        print("Outputting extra output to extra_output_file.txt")
        extra_tempfile_name = "my_extra_output_file.txt"
        local_extra_output_file = output_dir.get_named_tempfile(extra_tempfile_name)
        f = open(local_extra_output_file, "w")
        f.write("Trying to pass back the extra outputs")
        f.close()

        out_model.set(local_model_file)
        out_extra_output_file.set(local_extra_output_file)
        out.set(5)


@workflow_class
class SimpleWorkflow(object):
    train_dataset = Input(Types.Blob)
    validation_dataset = Input(Types.Blob)

    custom = custom_training_task(dummy_train_dataset=train_dataset,
                                  dummy_validation_dataset=validation_dataset,
                                  my_input="hello world")

    final_model = Output(custom.outputs.out_model, sdk_type=Types.Blob)
    final_extra_output = Output(custom.outputs.out_extra_output_file, sdk_type=Types.Blob)
    final_value = Output(custom.outputs.out, sdk_type=Types.Integer)
