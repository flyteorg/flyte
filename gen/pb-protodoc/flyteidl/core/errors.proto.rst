.. _api_file_flyteidl/core/errors.proto:

errors.proto
==========================

.. _api_msg_flyteidl.core.ContainerError:

flyteidl.core.ContainerError
----------------------------

`[flyteidl.core.ContainerError proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/errors.proto#L10>`_

Error message to propagate detailed errors from container executions to the execution
engine.

.. code-block:: json

  {
    "code": "...",
    "message": "...",
    "kind": "...",
    "origin": "..."
  }

.. _api_field_flyteidl.core.ContainerError.code:

code
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A simplified code for errors, so that we can provide a glossary of all possible errors.
  
  
.. _api_field_flyteidl.core.ContainerError.message:

message
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A detailed error message.
  
  
.. _api_field_flyteidl.core.ContainerError.kind:

kind
  (:ref:`flyteidl.core.ContainerError.Kind <api_enum_flyteidl.core.ContainerError.Kind>`) An abstract error kind for this error. Defaults to Non_Recoverable if not specified.
  
  
.. _api_field_flyteidl.core.ContainerError.origin:

origin
  (:ref:`flyteidl.core.ExecutionError.ErrorKind <api_enum_flyteidl.core.ExecutionError.ErrorKind>`) Defines the origin of the error (system, user, unknown).
  
  

.. _api_enum_flyteidl.core.ContainerError.Kind:

Enum flyteidl.core.ContainerError.Kind
--------------------------------------

`[flyteidl.core.ContainerError.Kind proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/errors.proto#L17>`_

Defines a generic error type that dictates the behavior of the retry strategy.

.. _api_enum_value_flyteidl.core.ContainerError.Kind.NON_RECOVERABLE:

NON_RECOVERABLE
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.ContainerError.Kind.RECOVERABLE:

RECOVERABLE
  ⁣
  

.. _api_msg_flyteidl.core.ErrorDocument:

flyteidl.core.ErrorDocument
---------------------------

`[flyteidl.core.ErrorDocument proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/errors.proto#L31>`_

Defines the errors.pb file format the container can produce to communicate
failure reasons to the execution engine.

.. code-block:: json

  {
    "error": "{...}"
  }

.. _api_field_flyteidl.core.ErrorDocument.error:

error
  (:ref:`flyteidl.core.ContainerError <api_msg_flyteidl.core.ContainerError>`) The error raised during execution.
  
  

