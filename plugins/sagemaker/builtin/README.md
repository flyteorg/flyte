[Back to Sagemaker Top menu](..)

# AWS Sagemaker Builtin Algorithms
Amazon SageMaker provides several built-in machine learning algorithms that you can use for a variety of problem types. 
Built-in algorithms are the fastest to get started with, as they are already pre-built and optimized on Sagemaker. To understand how they work and the various options available please refer to [Amazon
Sagemaker Official Documentation](https://docs.aws.amazon.com/sagemaker/latest/dg/algos.html)

Flyte Sagemaker plugin intends to greatly simplify using Sagemaker for training. We have tried to distill the API into a meaningful subset that makes it easier for users to adopt and run with Sagemaker.
Due to the nature of the Sagemaker built-in algorithms, it is possible to run them completely from a local notebook using Flyte. This is because, Flyte will automatically use a pre-built Image for the
given algorithm.

The Algorithm Images are configured in FlytePlugins in the plugin Configuration [here](https://github.com/lyft/flyteplugins/blob/master/go/tasks/plugins/k8s/sagemaker/config/config.go#L40). In the
default setting, we have configured XGBoost.

## Define a Training Job Task of SageMaker's built-in algorithm

Define a training job of SageMaker's built-in algorithm using Flytekit's `SdkBuiltinAlgorithmTrainingJobTask`:

```python
from flytekit.models.sagemaker import training_job as training_job_models
from flytekit.common.tasks.sagemaker import training_job_task

alg_spec = training_job_models.AlgorithmSpecification(
    input_mode=training_job_models.InputMode.FILE,
    algorithm_name=training_job_models.AlgorithmName.XGBOOST,
    algorithm_version="0.72",
    input_content_type=training_job_models.InputContentType.TEXT_CSV,
)

xgboost_train_task = training_job_task.SdkBuiltinAlgorithmTrainingJobTask(
    training_job_resource_config=training_job_models.TrainingJobResourceConfig(
        instance_type="ml.m4.xlarge",
        instance_count=1,
        volume_size_in_gb=25,
    ),
    algorithm_specification=alg_spec,
    cache_version='1',
    cacheable=True,
)
```
## Launch a Training Job Task
You can launch the training job standalone or from within a workflow.

```python
# Launching the training job standalone

xgboost_hyperparameters = {
    "num_round": "100",  # num_round is a required hyperparameter for XGBoost
    "base_score": "0.5",  
    "booster": "gbtree",  
}

training_inputs={
    "train": "s3://path/to/your/train/data/csv/folder",
    "validation": "s3://path/to/your/validation/data/csv/folder",
    "static_hyperparameters": xgboost_hyperparameters,
}

# Invoking the SdkBuiltinAlgorithmTrainingJobTask
training_exc = xgboost_train_task.register_and_launch("project", "domain", inputs=training_inputs)


# Invoking the training job from within a workflow

@workflow_class()
class TrainingWorkflow(object):
    ... 
    # The following two lines are just for demonstration purpose
    train_data = get_train_data(...)    
    validation_data = get_validation_data(...)

    # Invoking the training job 
    train = xgboost_train_task(
        train=train_data.outputs.output_csv,
        validation=validation_data.outputs.output_csv,
        static_hyperparameters=xgboost_hyperparameters,
    )

    ...
```

## Define hyperparameter tuning job 

### Wrapping a SageMaker Hyperparameter Tuning Job around a SageMaker Training Job

To define the hyperparameter tuning job, a training job is required.
The hyperparameter tuning job wraps around a training job:

```python
from flytekit.common.tasks.sagemaker import hpo_job_task
xgboost_hpo_task = hpo_job_task.SdkSimpleHyperparameterTuningJobTask(
    training_job=xgboost_train_task,
    max_number_of_training_jobs=10,
    max_parallel_training_jobs=5,
    cache_version='1',
    retries=2,
    cacheable=True,
)
```

### Invoking the hyperparameter tuning job using single-task execution
```python
from flytekit.models.sagemaker.hpo_job import HyperparameterTuningJobConfig, \
    HyperparameterTuningObjectiveType, HyperparameterTuningStrategy, \
    TrainingJobEarlyStoppingType, HyperparameterTuningObjective
from flytekit.models.sagemaker.parameter_ranges import HyperparameterScalingType, \ 
    ParameterRanges, ContinuousParameterRange, IntegerParameterRange

hpo_inputs={
    "train": "s3://path/to/your/train/data/csv/folder",
    "validation": "s3://path/to/your/validation/data/csv/folder",
    "static_hyperparameters": xgboost_hyperparameters,
    "hyperparameter_tuning_job_config": HyperparameterTuningJobConfig(
        hyperparameter_ranges=ParameterRanges(
            parameter_range_map={
                "num_round": IntegerParameterRange(min_value=3, max_value=10, 
                                                   scaling_type=HyperparameterScalingType.LINEAR),
                "gamma": ContinuousParameterRange(min_value=0.0, max_value=0.3,
                                                  scaling_type=HyperparameterScalingType.LINEAR),
            }
        ),
        tuning_strategy=HyperparameterTuningStrategy.BAYESIAN,
        tuning_objective=HyperparameterTuningObjective(
            objective_type=HyperparameterTuningObjectiveType.MINIMIZE,
            metric_name="validation:error",
        ),
        training_job_early_stopping_type=TrainingJobEarlyStoppingType.AUTO
    ).to_flyte_idl(),
}

hpo_exc = xgboost_hpo_task.register_and_launch("project", "domain", inputs=hpo_inputs)
```

### Invoking the hyperparameter tuning job from a workflow.

```python
from flytekit.models.sagemaker.hpo_job import HyperparameterTuningJobConfig, \
    HyperparameterTuningObjectiveType, HyperparameterTuningStrategy, \
    TrainingJobEarlyStoppingType, HyperparameterTuningObjective
from flytekit.models.sagemaker.parameter_ranges import HyperparameterScalingType, \ 
    ParameterRanges, ContinuousParameterRange, IntegerParameterRange

@workflow_class()
class TrainingWorkflow(object):    
    # Retrieve data
    train_data = get_train_data(...)
    validation_data = get_validation_data(...)

    train = xgboost_hpo_task(
        # Using the input we got from the Presto tasks
        train=train_data.outputs.output_csv,
        validation=validation_data.outputs.output_csv,
        
        static_hyperparameters=xgboost_hyperparameters,
        hyperparameter_tuning_job_config=HyperparameterTuningJobConfig(    
            hyperparameter_ranges=ParameterRanges(
                parameter_range_map={
                    "num_round": IntegerParameterRange(min_value=3, max_value=10, 
                                                       scaling_type=HyperparameterScalingType.LINEAR),
                    "gamma": ContinuousParameterRange(min_value=0.0, max_value=0.3,
                                                      scaling_type=HyperparameterScalingType.LINEAR),
                }
            ),
            tuning_strategy=HyperparameterTuningStrategy.BAYESIAN,
            tuning_objective=HyperparameterTuningObjective(
                objective_type=HyperparameterTuningObjectiveType.MINIMIZE,
                metric_name="validation:error",
            ),
            training_job_early_stopping_type=TrainingJobEarlyStoppingType.AUTO
        ).to_flyte_idl(),
    )
```
