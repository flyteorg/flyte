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

`[flyteidl.core.Resources proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L14>`_

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

`[flyteidl.core.Resources.ResourceEntry proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L25>`_

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

`[flyteidl.core.Resources.ResourceName proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L16>`_

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

`[flyteidl.core.RuntimeMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L43>`_

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

`[flyteidl.core.RuntimeMetadata.RuntimeType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L44>`_


.. _api_enum_value_flyteidl.core.RuntimeMetadata.RuntimeType.OTHER:

OTHER
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.RuntimeMetadata.RuntimeType.FLYTE_SDK:

FLYTE_SDK
  ⁣
  

.. _api_msg_flyteidl.core.TaskMetadata:

flyteidl.core.TaskMetadata
--------------------------

`[flyteidl.core.TaskMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L61>`_

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

`[flyteidl.core.TaskTemplate proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L95>`_

A Task structure that uniquely identifies a task in the system
Tasks are registered as a first step in the system.

.. code-block:: json

  {
    "id": "{...}",
    "type": "...",
    "metadata": "{...}",
    "interface": "{...}",
    "custom": "{...}",
    "container": "{...}"
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
  
  


.. _api_msg_flyteidl.core.ContainerPort:

flyteidl.core.ContainerPort
---------------------------

`[flyteidl.core.ContainerPort proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L125>`_

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

`[flyteidl.core.Container proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L131>`_


.. code-block:: json

  {
    "image": "...",
    "command": [],
    "args": [],
    "resources": "{...}",
    "env": [],
    "config": [],
    "ports": []
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
  
  

