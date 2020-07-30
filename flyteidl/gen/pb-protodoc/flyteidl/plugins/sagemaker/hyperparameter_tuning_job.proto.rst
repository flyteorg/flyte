.. _api_file_flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto:

hyperparameter_tuning_job.proto
==========================================================

.. _api_msg_flyteidl.plugins.sagemaker.HyperparameterTuningJob:

flyteidl.plugins.sagemaker.HyperparameterTuningJob
--------------------------------------------------

`[flyteidl.plugins.sagemaker.HyperparameterTuningJob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L10>`_

A pass-through for SageMaker's hyperparameter tuning job

.. code-block:: json

  {
    "training_job": "{...}",
    "max_number_of_training_jobs": "...",
    "max_parallel_training_jobs": "..."
  }

.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningJob.training_job:

training_job
  (:ref:`flyteidl.plugins.sagemaker.TrainingJob <api_msg_flyteidl.plugins.sagemaker.TrainingJob>`) The underlying training job that the hyperparameter tuning job will launch during the process
  
  
.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningJob.max_number_of_training_jobs:

max_number_of_training_jobs
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The maximum number of training jobs that an hpo job can launch. For resource limit purpose.
  https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_ResourceLimits.html
  
  
.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningJob.max_parallel_training_jobs:

max_parallel_training_jobs
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The maximum number of concurrent training job that an hpo job can launch
  https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_ResourceLimits.html
  
  


.. _api_msg_flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType:

flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType
------------------------------------------------------------

`[flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L25>`_

HyperparameterTuningObjectiveType determines the direction of the tuning of the Hyperparameter Tuning Job
with respect to the specified metric.

.. code-block:: json

  {}



.. _api_enum_flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType.Value:

Enum flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType.Value
-----------------------------------------------------------------------

`[flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType.Value proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L26>`_


.. _api_enum_value_flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType.Value.MINIMIZE:

MINIMIZE
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType.Value.MAXIMIZE:

MAXIMIZE
  ⁣
  

.. _api_msg_flyteidl.plugins.sagemaker.HyperparameterTuningObjective:

flyteidl.plugins.sagemaker.HyperparameterTuningObjective
--------------------------------------------------------

`[flyteidl.plugins.sagemaker.HyperparameterTuningObjective proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L34>`_

The target metric and the objective of the hyperparameter tuning.
https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-metrics.html

.. code-block:: json

  {
    "objective_type": "...",
    "metric_name": "..."
  }

.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningObjective.objective_type:

objective_type
  (:ref:`flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType.Value <api_enum_flyteidl.plugins.sagemaker.HyperparameterTuningObjectiveType.Value>`) HyperparameterTuningObjectiveType determines the direction of the tuning of the Hyperparameter Tuning Job
  with respect to the specified metric.
  
  
.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningObjective.metric_name:

metric_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The target metric name, which is the user-defined name of the metric specified in the
  training job's algorithm specification
  
  


.. _api_msg_flyteidl.plugins.sagemaker.HyperparameterTuningStrategy:

flyteidl.plugins.sagemaker.HyperparameterTuningStrategy
-------------------------------------------------------

`[flyteidl.plugins.sagemaker.HyperparameterTuningStrategy proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L49>`_

Setting the strategy used when searching in the hyperparameter space
Refer this doc for more details:
https://aws.amazon.com/blogs/machine-learning/amazon-sagemaker-automatic-model-tuning-now-supports-random-search-and-hyperparameter-scaling/

.. code-block:: json

  {}



.. _api_enum_flyteidl.plugins.sagemaker.HyperparameterTuningStrategy.Value:

Enum flyteidl.plugins.sagemaker.HyperparameterTuningStrategy.Value
------------------------------------------------------------------

`[flyteidl.plugins.sagemaker.HyperparameterTuningStrategy.Value proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L50>`_


.. _api_enum_value_flyteidl.plugins.sagemaker.HyperparameterTuningStrategy.Value.BAYESIAN:

BAYESIAN
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.plugins.sagemaker.HyperparameterTuningStrategy.Value.RANDOM:

RANDOM
  ⁣
  

.. _api_msg_flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType:

flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType
-------------------------------------------------------

`[flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L60>`_

When the training jobs launched by the hyperparameter tuning job are not improving significantly,
a hyperparameter tuning job can be stopping early.
Note that there's only a subset of built-in algorithms that supports early stopping.
see: https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-early-stopping.html

.. code-block:: json

  {}



.. _api_enum_flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType.Value:

Enum flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType.Value
------------------------------------------------------------------

`[flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType.Value proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L61>`_


.. _api_enum_value_flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType.Value.OFF:

OFF
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType.Value.AUTO:

AUTO
  ⁣
  

.. _api_msg_flyteidl.plugins.sagemaker.HyperparameterTuningJobConfig:

flyteidl.plugins.sagemaker.HyperparameterTuningJobConfig
--------------------------------------------------------

`[flyteidl.plugins.sagemaker.HyperparameterTuningJobConfig proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/hyperparameter_tuning_job.proto#L69>`_

The specification of the hyperparameter tuning process
https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-ex-tuning-job.html#automatic-model-tuning-ex-low-tuning-config

.. code-block:: json

  {
    "hyperparameter_ranges": "{...}",
    "tuning_strategy": "...",
    "tuning_objective": "{...}",
    "training_job_early_stopping_type": "..."
  }

.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningJobConfig.hyperparameter_ranges:

hyperparameter_ranges
  (:ref:`flyteidl.plugins.sagemaker.ParameterRanges <api_msg_flyteidl.plugins.sagemaker.ParameterRanges>`) ParameterRanges is a map that maps hyperparameter name to the corresponding hyperparameter range
  
  
.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningJobConfig.tuning_strategy:

tuning_strategy
  (:ref:`flyteidl.plugins.sagemaker.HyperparameterTuningStrategy.Value <api_enum_flyteidl.plugins.sagemaker.HyperparameterTuningStrategy.Value>`) Setting the strategy used when searching in the hyperparameter space
  
  
.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningJobConfig.tuning_objective:

tuning_objective
  (:ref:`flyteidl.plugins.sagemaker.HyperparameterTuningObjective <api_msg_flyteidl.plugins.sagemaker.HyperparameterTuningObjective>`) The target metric and the objective of the hyperparameter tuning.
  
  
.. _api_field_flyteidl.plugins.sagemaker.HyperparameterTuningJobConfig.training_job_early_stopping_type:

training_job_early_stopping_type
  (:ref:`flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType.Value <api_enum_flyteidl.plugins.sagemaker.TrainingJobEarlyStoppingType.Value>`) When the training jobs launched by the hyperparameter tuning job are not improving significantly,
  a hyperparameter tuning job can be stopping early.
  
  

