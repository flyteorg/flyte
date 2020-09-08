[Back to Sagemaker Top menu](..)

# :construction: AWS Sagemaker Custom Training algorithms
Flyte Sagemaker plugin intends to greatly simplify using Sagemaker for training. We have tried to distill the API into a meaningful subset that makes it easier for users to adopt and run with Sagemaker. Training code that runs on Sagemaker looks almost identical to writing any other task on Flyte.
Once a custom job is defined, hyper parameter optimization for pre-built algorithms or custom jobs is identical. Users need to wrap their training tasks into an HPO Task and launch it.

NOTE: Sagemaker custom algorithms work by building our own Docker images. These images need to be pushed to ECR for Sagemaker to access them. Thus, this examples need to be compiled and pushed to your
own AWS ECR docker registry to actually execute on Sagemaker.

## Define a Training Job Task of SageMaker's built-in algorithm
To define a job that can be run on Sagemaker training platform using flyte-kit, it should essentially look like any other containerized task declaration - **e.g. - @python_task**.

```python
from flytekit.sdk.sagemaker.task import custom_training_job_task

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
  ...
```
The example above defines a custom training task, that takes in **dummy_train_dataset, dummy_validation_dataset, my_input** as inputs and produces, **out_model, out_extra_output_file, out** as
outputs.

**@custom_training_job_task** declares this function as sagemaker executable and automatically configures it correctly. When an execution is triggered, the Sagemaker API is invoked to launch a job in
and users function **custom_training_task** is invoked and all parameters are passed to it.

