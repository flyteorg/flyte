# SageMaker on Flyte -- Launching SageMaker TrainingJob and HPOJob from Flyte [Alpha]

<!-- vscode-markdown-toc -->
* 1. [Disclaimer](#Disclaimer)
* 2. [Overview](#Overview)
* 3. [Flyte Component Statuses](#FlyteComponentStatuses)
* 4. [SDK design in Flytekit](#SDKdesigninFlytekit)
	* 4.1. [Defining a Simple Training Job](#DefiningaSimpleTrainingJobstatus)
	* 4.2. [Defining a Custom Training Job (Work-in-progress)](#DefiningaCustomTrainingJobWork-in-progress)
	* 4.3. [Defining a simple Hyperparameter Tuning Job](#DefiningasimpleHyperparameterTuningJob)
	* 4.4. [Invoking Training Jobs Task and Hyperparameter Tuning Jobs Task](#InvokingTrainingJobsTaskandHyperparameterTuningJobsTask)
	* 4.5. [Examples](#Examples)
* 5. [Related Pull Requests](#RelatedPullRequests)
	* 5.1. [`flyteidl`](#flyteidl)
	* 5.2. [`flytekit`](#flytekit)
	* 5.3. [`flyteplugins`](#flyteplugins)
	* 5.4. [`flytepropeller`](#flytepropeller)

<!-- vscode-markdown-toc-config
	numbering=true
	autoSave=true
	/vscode-markdown-toc-config -->
<!-- /vscode-markdown-toc -->

##  1. <a name='Disclaimer'></a>Disclaimer
🚧 ** This is currently an Alpha version of the proposal and implementation. The
interface and the functionalities are not finalized and are still subject to
change.** 🚧

##  2. <a name='Overview'></a>Overview

AWS SageMaker provides an elastic infrastructure for users to train models for a
wide spectrum of state-of-the-art machine learning algorithms and frameworks on
the AWS cloud. To enable seamless and powerful machine learning use cases on
Flyte, in this proposal, we aim to implement a plugin for Flyte to allow users
to leverage some of AWS SageMaker's key functionalities directly from within
their Flyte workflows and tasks, so that they can enjoy the excellent
data-processing and orchestration capability of Flyte at the same time.

We aim to build a unified pipeline to cover two end-to-end user flow for further
evaluation. These two e2e flows include:

* Calling SageMaker XGBoost from a Jupyter Notebook through Flyte
* Calling SageMaker XGBoost from Flyte Workflow

To achieve this goal, we have been iterating on designing an intuitive and
flexible interface in Flytekit, and actively implementing the various backend
components and protobuf message spec in Flyteplugins and Flyteidl to realize
these functionalities. The rest of the document aims to provide the status of
each of the critical components, as well as demonstrating the current interface
design that we have.

This proposal is still a work in progress. We welcome any kind of constructive
feedback that will improve the integration between Flyte and SageMaker and make
Flyte more powerful and easier to use. If you have one, please feel free to
write it in Flyte's [issue](https://github.com/lyft/flyte/issues) page.

##  3. <a name='FlyteComponentStatuses'></a>Flyte Component Statuses

Statuses of the *alpha version* of the important Flyte components:

| Flyte component \ Functionality  | simple TrainingJob task | simple HPOJob task | custom TrainingJob task |
| -------------------------------- | ----------------------- | ------------------ | ----------------------- |
| FlyteIDL (Proto Msg definitions) | :heavy_check_mark:      | :heavy_check_mark: | :heavy_check_mark:      |
| Flytekit (Python SDK)            | :heavy_check_mark:      | :heavy_check_mark: | :construction:          |
| Flyteplugins (GO Plugins)        | :heavy_check_mark:      | :heavy_check_mark: | :construction:          |

##  4. <a name='SDKdesigninFlytekit'></a>SDK design in Flytekit

In this section, we will demonstrate our Flytekit interface that users will use
in Flyte to compose tasks running SageMaker jobs.

###  4.1. <a name='DefiningaSimpleTrainingJobstatus'></a>Defining a Simple Training Job

Users can leverage SageMaker's powerful built-in algorithms easily without
needing to write any function or logic. They can simply define a
`SdkSimpleTrainingJobTask` in Flytekit and supplies the settings and the spec of
the built-in algorithm as follows:

```python

alg_spec = training_job_models.AlgorithmSpecification(
    input_mode=_sdk_sagemaker_types.InputMode.FILE,
    algorithm_name=_sdk_sagemaker_types.AlgorithmName.XGBOOST,
    algorithm_version="0.72",
    metric_definitions=[training_job_models.MetricDefinition(name="Minimize", regex="validation:error")]
)

# Definition of a SageMaker training job using SageMaker's built-in algorithm mode of XGBoost
xgboost_simple_train_task = training_job_task.SdkSimpleTrainingJobTask(
    training_job_config=training_job_models.TrainingJobConfig(
        instance_type="ml.m4.xlarge",
        instance_count=1,
        volume_size_in_gb=25,
    ),
    algorithm_specification=alg_spec,
    cache_version='v1',
    cacheable=True,
)


```

###  4.2. <a name='DefiningaCustomTrainingJobWork-in-progress'></a>Defining a Custom Training Job (Work-in-progress)

Users can also define a custom training job by using the
`@custom_training_job_task` decorator. With a custom training job, users can write complicated
training code freely with the frameworks/algorithms of their choose.

```python

# Definition of a custom training job using XGBoost. This corresponds to the Framework mode and the Bring-your-own-container mode in SageMaker training
@inputs(
    custom_input1=Types.Integer,
    custom_input2=Types.Integer,
)
@outputs(
    custom_output1=Types.Blob,
)
@custom_training_job_task(
    trainingjob_conf=training_job_models.TrainingJobConfig(
        instance_type="ml.m4.xlarge",
        instance_count=1,
        volume_size_in_gb=25,
    },
    algorithm_specification={
        input_mode=_sdk_sagemaker_types.InputMode.FILE,
        algorithm_name=_sdk_sagemaker_types.AlgorithmName.CUSTOM,
        metric_definitions=[training_job_models.MetricDefinition(name="Minimize", regex="validation:error")]
    },
    cache_version='v1',
    retries=2,
    cachable=True
)
def custom_xgboost_training_job_task(
        wf_params,
        train,
        validation,
        hyperparameters,
        stopping_condition,
        custom_input1,
        custom_input2,
        model,
        custom_output1,
    ):

    with train as reader:
        train_df = reader.read(concat=True)
        dtrain_x = xgb.DMatrix(train_df[:-1])
        dtrain_y = xgb.DMatrix(train_df[-1])
    with validation as reader:
        validation_df = reader.read(concat=True)
        dvalidation_x = xgb.DMatrix(validation_df[:-1])
        dvalidation_y = xgb.DMatrix(validation_df[-1])

    my_model = xgb.XGBModel(**hyperparameters)

    my_model.fit(dtrain_x,
                 dtrain_y,
                 eval_set=[(dvalidation_x, dvalidation_y)],
                 eval_metric=sample_eval_function)

    model.set(my_model)
    custom_output1.set(my_model.evals_result())
```

###  4.3. <a name='DefiningasimpleHyperparameterTuningJob'></a>Defining a simple Hyperparameter Tuning Job

SageMaker-on-Flyte also supports easy chaining between a TrainingJob task and a
hpo job. After users define a TrainingJob task, he/she may want to apply
hyperparameter tuning to the the training, while also maintaining the
flexibility to run the TrainingJob task standalone. This should be easily doable
by using the `SdkSimpleHPOJobTask` in our SDK. `SdkSimpleHPOJobTask` accepts the
definition of a TrainingJob task as a part of the spec.

```python
# Definition of a SageMaker hyperparameter-tuning job.
# Note that is chained behind the simple TrainingJob task defined above
xgboost_simple_hpo_task = hpo_job_task.SdkSimpleHPOJobTask(
    training_job=xgboost_simple_train_task,
    max_number_of_training_jobs=10,
    max_parallel_training_jobs=5,
    cache_version='2',
    retries=2,
    cacheable=True,
)


# Definition of a SageMaker hyperparameter-tuning job.
# Note that is chained behind the custom TrainingJob task defined above
xgboost_hpo_task = hpo_job_task.SdkSimpleHPOJobTask(
    training_job=custom_xgboost_training_job_task,
    max_number_of_training_jobs=10,
    max_parallel_training_jobs=5,
    cache_version='2',
    retries=2,
    cacheable=True,
)
```

###  4.4. <a name='InvokingTrainingJobsTaskandHyperparameterTuningJobsTask'></a>Invoking Training Jobs Task and Hyperparameter Tuning Jobs Task

Invoking Training Job Tasks and HPO Job Tasks from inside a Flyte workflow is
pretty much the same as invoking other types of tasks. You should be able to
achieve this by simply supplying the *required inputs* to the task definition.
For Training Job Tasks and HPO Job Tasks , *required inputs* are pre-defined
inputs that we think is needed for every training job. That's why you don't see
the declaration of these inputs in the task definition -- we added them for you
in our SDK.

```python

@workflow_class
class SageMakerSimpleWorkflow(object):
    static_hyperparameters = Input(Types.Generic, required=True, help="Sample hyperparameter input")

    my_simple_trianing_task_exec = xgboost_simple_train_task(
        train="s3://path/to/train/data",
        validation="s3://path/to/validation/data",
        static_hyperparameters=static_hyperparameters,
        stopping_condition=StoppingCondition(
            max_runtime_in_seconds=43200,
        ).to_flyte_idl(),
    )
    ...
```

With Flyte's Single Task Execution capability available, it is even easier now
to invoke the SageMaker tasks. That is, users do not need a workflow to launch
the SageMaker tasks; instead, they can simply define the tasks, and then
register and launch the tasks standalone, which enables faster iterations.

What makes this more useful is that, with this pattern, users do not need to
copy-and-paste their code from their Jupyter notebooks to a Flyte python file.
Users can now feel free to focus on doing fast iterations on the task
definitions; after they are satisfied with the TrainingJob task and HPOJob task
definitions and want to see how it plays with the other components in their pipelines,
they can just directly invoke the same tasks they've been iterating
on inside their workflows. With the handy caching capability provided by Flyte,
if the last standalone task execution produced the results or models they wanted
to use for other downstream tasks, and they have enabled caching on that
execution, when they invoke the task in the workflow, they can skip the
re-execution (because the result is cached) and save cost significantly.

```python
xgboost_simple_train_task = training_job_task.SdkSimpleTrainingJobTask( ... )

xgboost_simple_hpo_task = hpo_job_task.SdkSimpleHPOJobTask(
    training_job=xgboost_simple_train_task,
    ...
)

# Users can register and launch the task in a standalone fashion with Flyte's
# Single Task Execution functionality, which allows users to do fast iterations
# and focus on the most important logic
exc = xgboost_simple_hpo_task.register_and_launch("flyteexamples", "development", inputs=...)

@workflow_class
class SageMakerSimpleWorkflow(object):
    ...

    # When users are satisfied with their task definitions and want to see the bigger
    # picture by running some end-to-end pipelines, they can directly invoke the same
    # task definition from within a workflow. If they had enabled caching in the
    # standalone execution, their workflow can happily skip the re-execution of the
    # task and save compute cost.
    my_simple_hpo_task_exec_in_wf = xgboost_simple_hpo_task(
        ...
    )
    ...

```

###  4.5. <a name='Examples'></a>Examples

A working alpha can be found in this following [jupyter notebook example](https://github.com/lyft/flytekit/blob/345e057e5840af39a1d156d157d008bd65d23451/sample-notebooks/sagemaker-hpo.ipynb).

This example mainly demonstrates how users can invoke the task definitions
standalone using the Single Task Execution capability, without needing to write
a Flyte workflow

It is NOT required to run this notebook in SageMaker Studio environment. You
should be able to run it as long as your enviornment satisfies the following two
requirements:

1. you have a proper config file that points to your Flytekit to your Flyte
   deployment, and
2. you have proper roles and policies set up and have them associated with your
   Flyte deployment


##  5. <a name='RelatedPullRequests'></a>Related Pull Requests
###  5.1. <a name='flyteidl'></a>`flyteidl`

- https://github.com/lyft/flyteidl/pull/66

###  5.2. <a name='flytekit'></a>`flytekit`

- https://github.com/lyft/flytekit/pull/120
- https://github.com/lyft/flytekit/pull/123

###  5.3. <a name='flyteplugins'></a>`flyteplugins`

- https://github.com/lyft/flyteplugins/pull/100

###  5.4. <a name='flytepropeller'></a>`flytepropeller`

- https://github.com/lyft/flytepropeller/pull/163