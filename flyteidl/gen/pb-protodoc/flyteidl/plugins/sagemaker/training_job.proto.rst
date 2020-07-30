.. _api_file_flyteidl/plugins/sagemaker/training_job.proto:

training_job.proto
=============================================

.. _api_msg_flyteidl.plugins.sagemaker.InputMode:

flyteidl.plugins.sagemaker.InputMode
------------------------------------

`[flyteidl.plugins.sagemaker.InputMode proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L15>`_

The input mode that the algorithm supports. When using the File input mode, SageMaker downloads
the training data from S3 to the provisioned ML storage Volume, and mounts the directory to docker
volume for training container. When using Pipe input mode, Amazon SageMaker streams data directly
from S3 to the container.
See: https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_AlgorithmSpecification.html
For the input modes that different SageMaker algorithms support, see:
https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html

.. code-block:: json

  {}



.. _api_enum_flyteidl.plugins.sagemaker.InputMode.Value:

Enum flyteidl.plugins.sagemaker.InputMode.Value
-----------------------------------------------

`[flyteidl.plugins.sagemaker.InputMode.Value proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L16>`_


.. _api_enum_value_flyteidl.plugins.sagemaker.InputMode.Value.FILE:

FILE
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.plugins.sagemaker.InputMode.Value.PIPE:

PIPE
  ⁣
  

.. _api_msg_flyteidl.plugins.sagemaker.AlgorithmName:

flyteidl.plugins.sagemaker.AlgorithmName
----------------------------------------

`[flyteidl.plugins.sagemaker.AlgorithmName proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L26>`_

The algorithm name is used for deciding which pre-built image to point to.
This is only required for use cases where SageMaker's built-in algorithm mode is used.
While we currently only support a subset of the algorithms, more will be added to the list.
See: https://docs.aws.amazon.com/sagemaker/latest/dg/algos.html

.. code-block:: json

  {}



.. _api_enum_flyteidl.plugins.sagemaker.AlgorithmName.Value:

Enum flyteidl.plugins.sagemaker.AlgorithmName.Value
---------------------------------------------------

`[flyteidl.plugins.sagemaker.AlgorithmName.Value proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L27>`_


.. _api_enum_value_flyteidl.plugins.sagemaker.AlgorithmName.Value.CUSTOM:

CUSTOM
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.plugins.sagemaker.AlgorithmName.Value.XGBOOST:

XGBOOST
  ⁣
  

.. _api_msg_flyteidl.plugins.sagemaker.InputContentType:

flyteidl.plugins.sagemaker.InputContentType
-------------------------------------------

`[flyteidl.plugins.sagemaker.InputContentType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L37>`_

Specifies the type of file for input data. Different SageMaker built-in algorithms require different file types of input data
See https://docs.aws.amazon.com/sagemaker/latest/dg/cdf-training.html
https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html

.. code-block:: json

  {}



.. _api_enum_flyteidl.plugins.sagemaker.InputContentType.Value:

Enum flyteidl.plugins.sagemaker.InputContentType.Value
------------------------------------------------------

`[flyteidl.plugins.sagemaker.InputContentType.Value proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L38>`_


.. _api_enum_value_flyteidl.plugins.sagemaker.InputContentType.Value.TEXT_CSV:

TEXT_CSV
  *(DEFAULT)* ⁣
  

.. _api_msg_flyteidl.plugins.sagemaker.MetricDefinition:

flyteidl.plugins.sagemaker.MetricDefinition
-------------------------------------------

`[flyteidl.plugins.sagemaker.MetricDefinition proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L46>`_

Specifies a metric that the training algorithm writes to stderr or stdout.
This object is a pass-through.
See this for details: https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_MetricDefinition.html

.. code-block:: json

  {
    "name": "...",
    "regex": "..."
  }

.. _api_field_flyteidl.plugins.sagemaker.MetricDefinition.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) User-defined name of the metric
  
  
.. _api_field_flyteidl.plugins.sagemaker.MetricDefinition.regex:

regex
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) SageMaker hyperparameter tuning parses your algorithm’s stdout and stderr streams to find algorithm metrics
  
  


.. _api_msg_flyteidl.plugins.sagemaker.AlgorithmSpecification:

flyteidl.plugins.sagemaker.AlgorithmSpecification
-------------------------------------------------

`[flyteidl.plugins.sagemaker.AlgorithmSpecification proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L61>`_

Specifies the training algorithm to be used in the training job
This object is mostly a pass-through, with a couple of exceptions include: (1) in Flyte, users don't need to specify
TrainingImage; either use the built-in algorithm mode by using Flytekit's Simple Training Job and specifying an algorithm
name and an algorithm version or (2) when users want to supply custom algorithms they should set algorithm_name field to
CUSTOM. In this case, the value of the algorithm_version field has no effect
For pass-through use cases: refer to this AWS official document for more details
https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_AlgorithmSpecification.html

.. code-block:: json

  {
    "input_mode": "...",
    "algorithm_name": "...",
    "algorithm_version": "...",
    "metric_definitions": [],
    "input_content_type": "..."
  }

.. _api_field_flyteidl.plugins.sagemaker.AlgorithmSpecification.input_mode:

input_mode
  (:ref:`flyteidl.plugins.sagemaker.InputMode.Value <api_enum_flyteidl.plugins.sagemaker.InputMode.Value>`) The input mode can be either PIPE or FILE
  
  
.. _api_field_flyteidl.plugins.sagemaker.AlgorithmSpecification.algorithm_name:

algorithm_name
  (:ref:`flyteidl.plugins.sagemaker.AlgorithmName.Value <api_enum_flyteidl.plugins.sagemaker.AlgorithmName.Value>`) The algorithm name is used for deciding which pre-built image to point to
  
  
.. _api_field_flyteidl.plugins.sagemaker.AlgorithmSpecification.algorithm_version:

algorithm_version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The algorithm version field is used for deciding which pre-built image to point to
  This is only needed for use cases where SageMaker's built-in algorithm mode is chosen
  
  
.. _api_field_flyteidl.plugins.sagemaker.AlgorithmSpecification.metric_definitions:

metric_definitions
  (:ref:`flyteidl.plugins.sagemaker.MetricDefinition <api_msg_flyteidl.plugins.sagemaker.MetricDefinition>`) A list of metric definitions for SageMaker to evaluate/track on the progress of the training job
  See this: https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_AlgorithmSpecification.html
  and this: https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-metrics.html
  
  
.. _api_field_flyteidl.plugins.sagemaker.AlgorithmSpecification.input_content_type:

input_content_type
  (:ref:`flyteidl.plugins.sagemaker.InputContentType.Value <api_enum_flyteidl.plugins.sagemaker.InputContentType.Value>`) The content type of the input
  See https://docs.aws.amazon.com/sagemaker/latest/dg/cdf-training.html
  https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
  
  


.. _api_msg_flyteidl.plugins.sagemaker.TrainingJobResourceConfig:

flyteidl.plugins.sagemaker.TrainingJobResourceConfig
----------------------------------------------------

`[flyteidl.plugins.sagemaker.TrainingJobResourceConfig proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L85>`_

TrainingJobResourceConfig is a pass-through, specifying the instance type to use for the training job, the
number of instances to launch, and the size of the ML storage volume the user wants to provision
Refer to SageMaker official doc for more details: https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_CreateTrainingJob.html

.. code-block:: json

  {
    "instance_count": "...",
    "instance_type": "...",
    "volume_size_in_gb": "..."
  }

.. _api_field_flyteidl.plugins.sagemaker.TrainingJobResourceConfig.instance_count:

instance_count
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The number of ML compute instances to use. For distributed training, provide a value greater than 1.
  
  
.. _api_field_flyteidl.plugins.sagemaker.TrainingJobResourceConfig.instance_type:

instance_type
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The ML compute instance type
  
  
.. _api_field_flyteidl.plugins.sagemaker.TrainingJobResourceConfig.volume_size_in_gb:

volume_size_in_gb
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The size of the ML storage volume that you want to provision.
  
  


.. _api_msg_flyteidl.plugins.sagemaker.TrainingJob:

flyteidl.plugins.sagemaker.TrainingJob
--------------------------------------

`[flyteidl.plugins.sagemaker.TrainingJob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sagemaker/training_job.proto#L96>`_

The spec of a training job. This is mostly a pass-through object
https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_CreateTrainingJob.html

.. code-block:: json

  {
    "algorithm_specification": "{...}",
    "training_job_resource_config": "{...}"
  }

.. _api_field_flyteidl.plugins.sagemaker.TrainingJob.algorithm_specification:

algorithm_specification
  (:ref:`flyteidl.plugins.sagemaker.AlgorithmSpecification <api_msg_flyteidl.plugins.sagemaker.AlgorithmSpecification>`) 
  
.. _api_field_flyteidl.plugins.sagemaker.TrainingJob.training_job_resource_config:

training_job_resource_config
  (:ref:`flyteidl.plugins.sagemaker.TrainingJobResourceConfig <api_msg_flyteidl.plugins.sagemaker.TrainingJobResourceConfig>`) 
  

