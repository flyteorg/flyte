.. _api_file_flyteidl/core/tasks.proto:

tasks.proto
=========================

For interruptible we will populate it at the node level but require it be part of TaskMetadata
for a user to set the value.
We are using oneof instead of bool because otherwise we would be unable to distinguish between value being
set by the user or defaulting to false.
The logic of handling precedence will be done as part of flytepropeller.

.. _api_msg_flyteidl.core.Resources:

flyteidl.core.Resources
-----------------------

`[flyteidl.core.Resources proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L15>`_

A customizable interface to convey resources requested for a container. This can be interpretted differently for different
container engines.

.. code-block:: json

  {
    "requests": [],
    "limits": []
  }

.. _api_field_flyteidl.core.Resources.requests:

requests
  (:ref:`flyteidl.core.Resources.ResourceEntry <api_msg_flyteidl.core.Resources.ResourceEntry>`) The desired set of resources requested. ResourceNames must be unique within the list.
  
  
.. _api_field_flyteidl.core.Resources.limits:

limits
  (:ref:`flyteidl.core.Resources.ResourceEntry <api_msg_flyteidl.core.Resources.ResourceEntry>`) Defines a set of bounds (e.g. min/max) within which the task can reliably run. ResourceNames must be unique
  within the list.
  
  
.. _api_msg_flyteidl.core.Resources.ResourceEntry:

flyteidl.core.Resources.ResourceEntry
-------------------------------------

`[flyteidl.core.Resources.ResourceEntry proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L26>`_

Encapsulates a resource name and value.

.. code-block:: json

  {
    "name": "...",
    "value": "..."
  }

.. _api_field_flyteidl.core.Resources.ResourceEntry.name:

name
  (:ref:`flyteidl.core.Resources.ResourceName <api_enum_flyteidl.core.Resources.ResourceName>`) Resource name.
  
  
.. _api_field_flyteidl.core.Resources.ResourceEntry.value:

value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Value must be a valid k8s quantity. See
  https://github.com/kubernetes/apimachinery/blob/master/pkg/api/resource/quantity.go#L30-L80
  
  


.. _api_enum_flyteidl.core.Resources.ResourceName:

Enum flyteidl.core.Resources.ResourceName
-----------------------------------------

`[flyteidl.core.Resources.ResourceName proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L17>`_

Known resource names.

.. _api_enum_value_flyteidl.core.Resources.ResourceName.UNKNOWN:

UNKNOWN
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.Resources.ResourceName.CPU:

CPU
  ⁣
  
.. _api_enum_value_flyteidl.core.Resources.ResourceName.GPU:

GPU
  ⁣
  
.. _api_enum_value_flyteidl.core.Resources.ResourceName.MEMORY:

MEMORY
  ⁣
  
.. _api_enum_value_flyteidl.core.Resources.ResourceName.STORAGE:

STORAGE
  ⁣
  

.. _api_msg_flyteidl.core.RuntimeMetadata:

flyteidl.core.RuntimeMetadata
-----------------------------

`[flyteidl.core.RuntimeMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L44>`_

Runtime information. This is losely defined to allow for extensibility.

.. code-block:: json

  {
    "type": "...",
    "version": "...",
    "flavor": "..."
  }

.. _api_field_flyteidl.core.RuntimeMetadata.type:

type
  (:ref:`flyteidl.core.RuntimeMetadata.RuntimeType <api_enum_flyteidl.core.RuntimeMetadata.RuntimeType>`) Type of runtime.
  
  
.. _api_field_flyteidl.core.RuntimeMetadata.version:

version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Version of the runtime. All versions should be backward compatible. However, certain cases call for version
  checks to ensure tighter validation or setting expectations.
  
  
.. _api_field_flyteidl.core.RuntimeMetadata.flavor:

flavor
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) optional It can be used to provide extra information about the runtime (e.g. python, golang... etc.).
  
  

.. _api_enum_flyteidl.core.RuntimeMetadata.RuntimeType:

Enum flyteidl.core.RuntimeMetadata.RuntimeType
----------------------------------------------

`[flyteidl.core.RuntimeMetadata.RuntimeType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L45>`_


.. _api_enum_value_flyteidl.core.RuntimeMetadata.RuntimeType.OTHER:

OTHER
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.RuntimeMetadata.RuntimeType.FLYTE_SDK:

FLYTE_SDK
  ⁣
  

.. _api_msg_flyteidl.core.TaskMetadata:

flyteidl.core.TaskMetadata
--------------------------

`[flyteidl.core.TaskMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L62>`_

Task Metadata

.. code-block:: json

  {
    "discoverable": "...",
    "runtime": "{...}",
    "timeout": "{...}",
    "retries": "{...}",
    "discovery_version": "...",
    "deprecated_error_message": "...",
    "interruptible": "..."
  }

.. _api_field_flyteidl.core.TaskMetadata.discoverable:

discoverable
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates whether the system should attempt to lookup this task's output to avoid duplication of work.
  
  
.. _api_field_flyteidl.core.TaskMetadata.runtime:

runtime
  (:ref:`flyteidl.core.RuntimeMetadata <api_msg_flyteidl.core.RuntimeMetadata>`) Runtime information about the task.
  
  
.. _api_field_flyteidl.core.TaskMetadata.timeout:

timeout
  (:ref:`google.protobuf.Duration <api_msg_google.protobuf.Duration>`) The overall timeout of a task including user-triggered retries.
  
  
.. _api_field_flyteidl.core.TaskMetadata.retries:

retries
  (:ref:`flyteidl.core.RetryStrategy <api_msg_flyteidl.core.RetryStrategy>`) Number of retries per task.
  
  
.. _api_field_flyteidl.core.TaskMetadata.discovery_version:

discovery_version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates a logical version to apply to this task for the purpose of discovery.
  
  
.. _api_field_flyteidl.core.TaskMetadata.deprecated_error_message:

deprecated_error_message
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) If set, this indicates that this task is deprecated.  This will enable owners of tasks to notify consumers
  of the ending of support for a given task.
  
  
.. _api_field_flyteidl.core.TaskMetadata.interruptible:

interruptible
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  Identify whether task is interruptible
  
  


.. _api_msg_flyteidl.core.TaskTemplate:

flyteidl.core.TaskTemplate
--------------------------

`[flyteidl.core.TaskTemplate proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L96>`_

A Task structure that uniquely identifies a task in the system
Tasks are registered as a first step in the system.

.. code-block:: json

  {
    "id": "{...}",
    "type": "...",
    "metadata": "{...}",
    "interface": "{...}",
    "custom": "{...}",
    "container": "{...}",
    "task_type_version": "...",
    "security_context": "{...}"
  }

.. _api_field_flyteidl.core.TaskTemplate.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) Auto generated taskId by the system. Task Id uniquely identifies this task globally.
  
  
.. _api_field_flyteidl.core.TaskTemplate.type:

type
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A predefined yet extensible Task type identifier. This can be used to customize any of the components. If no
  extensions are provided in the system, Flyte will resolve the this task to its TaskCategory and default the
  implementation registered for the TaskCategory.
  
  
.. _api_field_flyteidl.core.TaskTemplate.metadata:

metadata
  (:ref:`flyteidl.core.TaskMetadata <api_msg_flyteidl.core.TaskMetadata>`) Extra metadata about the task.
  
  
.. _api_field_flyteidl.core.TaskTemplate.interface:

interface
  (:ref:`flyteidl.core.TypedInterface <api_msg_flyteidl.core.TypedInterface>`) A strongly typed interface for the task. This enables others to use this task within a workflow and gauarantees
  compile-time validation of the workflow to avoid costly runtime failures.
  
  
.. _api_field_flyteidl.core.TaskTemplate.custom:

custom
  (:ref:`google.protobuf.Struct <api_msg_google.protobuf.Struct>`) Custom data about the task. This is extensible to allow various plugins in the system.
  
  
.. _api_field_flyteidl.core.TaskTemplate.container:

container
  (:ref:`flyteidl.core.Container <api_msg_flyteidl.core.Container>`) 
  Known target types that the system will guarantee plugins for. Custom SDK plugins are allowed to set these if needed.
  If no corresponding execution-layer plugins are found, the system will default to handling these using built-in
  handlers.
  
  
.. _api_field_flyteidl.core.TaskTemplate.task_type_version:

task_type_version
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) This can be used to customize task handling at execution time for the same task type.
  
  
.. _api_field_flyteidl.core.TaskTemplate.security_context:

security_context
  (:ref:`flyteidl.core.SecurityContext <api_msg_flyteidl.core.SecurityContext>`) security_context encapsulates security attributes requested to run this task.
  
  


.. _api_msg_flyteidl.core.ContainerPort:

flyteidl.core.ContainerPort
---------------------------

`[flyteidl.core.ContainerPort proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L132>`_

Defines port properties for a container.

.. code-block:: json

  {
    "container_port": "..."
  }

.. _api_field_flyteidl.core.ContainerPort.container_port:

container_port
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Number of port to expose on the pod's IP address.
  This must be a valid port number, 0 < x < 65536.
  
  


.. _api_msg_flyteidl.core.Container:

flyteidl.core.Container
-----------------------

`[flyteidl.core.Container proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L138>`_


.. code-block:: json

  {
    "image": "...",
    "command": [],
    "args": [],
    "resources": "{...}",
    "env": [],
    "config": [],
    "ports": [],
    "data_config": "{...}"
  }

.. _api_field_flyteidl.core.Container.image:

image
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Container image url. Eg: docker/redis:latest
  
  
.. _api_field_flyteidl.core.Container.command:

command
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Command to be executed, if not provided, the default entrypoint in the container image will be used.
  
  
.. _api_field_flyteidl.core.Container.args:

args
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) These will default to Flyte given paths. If provided, the system will not append known paths. If the task still
  needs flyte's inputs and outputs path, add $(FLYTE_INPUT_FILE), $(FLYTE_OUTPUT_FILE) wherever makes sense and the
  system will populate these before executing the container.
  
  
.. _api_field_flyteidl.core.Container.resources:

resources
  (:ref:`flyteidl.core.Resources <api_msg_flyteidl.core.Resources>`) Container resources requirement as specified by the container engine.
  
  
.. _api_field_flyteidl.core.Container.env:

env
  (:ref:`flyteidl.core.KeyValuePair <api_msg_flyteidl.core.KeyValuePair>`) Environment variables will be set as the container is starting up.
  
  
.. _api_field_flyteidl.core.Container.config:

config
  (:ref:`flyteidl.core.KeyValuePair <api_msg_flyteidl.core.KeyValuePair>`) Allows extra configs to be available for the container.
  TODO: elaborate on how configs will become available.
  
  
.. _api_field_flyteidl.core.Container.ports:

ports
  (:ref:`flyteidl.core.ContainerPort <api_msg_flyteidl.core.ContainerPort>`) Ports to open in the container. This feature is not supported by all execution engines. (e.g. supported on K8s but
  not supported on AWS Batch)
  Only K8s
  
  
.. _api_field_flyteidl.core.Container.data_config:

data_config
  (:ref:`flyteidl.core.DataLoadingConfig <api_msg_flyteidl.core.DataLoadingConfig>`) BETA: Optional configuration for DataLoading. If not specified, then default values are used.
  This makes it possible to to run a completely portable container, that uses inputs and outputs
  only from the local file-system and without having any reference to flyteidl. This is supported only on K8s at the moment.
  If data loading is enabled, then data will be mounted in accompanying directories specified in the DataLoadingConfig. If the directories
  are not specified, inputs will be mounted onto and outputs will be uploaded from a pre-determined file-system path. Refer to the documentation
  to understand the default paths.
  Only K8s
  
  


.. _api_msg_flyteidl.core.IOStrategy:

flyteidl.core.IOStrategy
------------------------

`[flyteidl.core.IOStrategy proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L176>`_

Strategy to use when dealing with Blob, Schema, or multipart blob data (large datasets)

.. code-block:: json

  {
    "download_mode": "...",
    "upload_mode": "..."
  }

.. _api_field_flyteidl.core.IOStrategy.download_mode:

download_mode
  (:ref:`flyteidl.core.IOStrategy.DownloadMode <api_enum_flyteidl.core.IOStrategy.DownloadMode>`) Mode to use to manage downloads
  
  
.. _api_field_flyteidl.core.IOStrategy.upload_mode:

upload_mode
  (:ref:`flyteidl.core.IOStrategy.UploadMode <api_enum_flyteidl.core.IOStrategy.UploadMode>`) Mode to use to manage uploads
  
  

.. _api_enum_flyteidl.core.IOStrategy.DownloadMode:

Enum flyteidl.core.IOStrategy.DownloadMode
------------------------------------------

`[flyteidl.core.IOStrategy.DownloadMode proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L178>`_

Mode to use for downloading

.. _api_enum_value_flyteidl.core.IOStrategy.DownloadMode.DOWNLOAD_EAGER:

DOWNLOAD_EAGER
  *(DEFAULT)* ⁣All data will be downloaded before the main container is executed
  
  
.. _api_enum_value_flyteidl.core.IOStrategy.DownloadMode.DOWNLOAD_STREAM:

DOWNLOAD_STREAM
  ⁣Data will be downloaded as a stream and an End-Of-Stream marker will be written to indicate all data has been downloaded. Refer to protocol for details
  
  
.. _api_enum_value_flyteidl.core.IOStrategy.DownloadMode.DO_NOT_DOWNLOAD:

DO_NOT_DOWNLOAD
  ⁣Large objects (offloaded) will not be downloaded
  
  

.. _api_enum_flyteidl.core.IOStrategy.UploadMode:

Enum flyteidl.core.IOStrategy.UploadMode
----------------------------------------

`[flyteidl.core.IOStrategy.UploadMode proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L187>`_

Mode to use for uploading

.. _api_enum_value_flyteidl.core.IOStrategy.UploadMode.UPLOAD_ON_EXIT:

UPLOAD_ON_EXIT
  *(DEFAULT)* ⁣All data will be uploaded after the main container exits
  
  
.. _api_enum_value_flyteidl.core.IOStrategy.UploadMode.UPLOAD_EAGER:

UPLOAD_EAGER
  ⁣Data will be uploaded as it appears. Refer to protocol specification for details
  
  
.. _api_enum_value_flyteidl.core.IOStrategy.UploadMode.DO_NOT_UPLOAD:

DO_NOT_UPLOAD
  ⁣Data will not be uploaded, only references will be written
  
  

.. _api_msg_flyteidl.core.DataLoadingConfig:

flyteidl.core.DataLoadingConfig
-------------------------------

`[flyteidl.core.DataLoadingConfig proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L204>`_

This configuration allows executing raw containers in Flyte using the Flyte CoPilot system.
Flyte CoPilot, eliminates the needs of flytekit or sdk inside the container. Any inputs required by the users container are side-loaded in the input_path
Any outputs generated by the user container - within output_path are automatically uploaded.

.. code-block:: json

  {
    "enabled": "...",
    "input_path": "...",
    "output_path": "...",
    "format": "...",
    "io_strategy": "{...}"
  }

.. _api_field_flyteidl.core.DataLoadingConfig.enabled:

enabled
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Flag enables DataLoading Config. If this is not set, data loading will not be used!
  
  
.. _api_field_flyteidl.core.DataLoadingConfig.input_path:

input_path
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) File system path (start at root). This folder will contain all the inputs exploded to a separate file.
  Example, if the input interface needs (x: int, y: blob, z: multipart_blob) and the input path is "/var/flyte/inputs", then the file system will look like
  /var/flyte/inputs/inputs.<metadata format dependent -> .pb .json .yaml> -> Format as defined previously. The Blob and Multipart blob will reference local filesystem instead of remote locations
  /var/flyte/inputs/x -> X is a file that contains the value of x (integer) in string format
  /var/flyte/inputs/y -> Y is a file in Binary format
  /var/flyte/inputs/z/... -> Note Z itself is a directory
  More information about the protocol - refer to docs #TODO reference docs here
  
  
.. _api_field_flyteidl.core.DataLoadingConfig.output_path:

output_path
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) File system path (start at root). This folder should contain all the outputs for the task as individual files and/or an error text file
  
  
.. _api_field_flyteidl.core.DataLoadingConfig.format:

format
  (:ref:`flyteidl.core.DataLoadingConfig.LiteralMapFormat <api_enum_flyteidl.core.DataLoadingConfig.LiteralMapFormat>`) In the inputs folder, there will be an additional summary/metadata file that contains references to all files or inlined primitive values.
  This format decides the actual encoding for the data. Refer to the encoding to understand the specifics of the contents and the encoding
  
  
.. _api_field_flyteidl.core.DataLoadingConfig.io_strategy:

io_strategy
  (:ref:`flyteidl.core.IOStrategy <api_msg_flyteidl.core.IOStrategy>`) 
  

.. _api_enum_flyteidl.core.DataLoadingConfig.LiteralMapFormat:

Enum flyteidl.core.DataLoadingConfig.LiteralMapFormat
-----------------------------------------------------

`[flyteidl.core.DataLoadingConfig.LiteralMapFormat proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L209>`_

LiteralMapFormat decides the encoding format in which the input metadata should be made available to the containers.
If the user has access to the protocol buffer definitions, it is recommended to use the PROTO format.
JSON and YAML do not need any protobuf definitions to read it
All remote references in core.LiteralMap are replaced with local filesystem references (the data is downloaded to local filesystem)

.. _api_enum_value_flyteidl.core.DataLoadingConfig.LiteralMapFormat.JSON:

JSON
  *(DEFAULT)* ⁣JSON / YAML for the metadata (which contains inlined primitive values). The representation is inline with the standard json specification as specified - https://www.json.org/json-en.html
  
  
.. _api_enum_value_flyteidl.core.DataLoadingConfig.LiteralMapFormat.YAML:

YAML
  ⁣
  
.. _api_enum_value_flyteidl.core.DataLoadingConfig.LiteralMapFormat.PROTO:

PROTO
  ⁣Proto is a serialized binary of `core.LiteralMap` defined in flyteidl/core
  
  
