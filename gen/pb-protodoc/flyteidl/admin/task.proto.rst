.. _api_file_flyteidl/admin/task.proto:

task.proto
=========================

.. _api_msg_flyteidl.admin.TaskCreateRequest:

flyteidl.admin.TaskCreateRequest
--------------------------------

`[flyteidl.admin.TaskCreateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task.proto#L11>`_

Represents a request structure to create a revision of a task.

.. code-block:: json

  {
    "id": "{...}",
    "spec": "{...}"
  }

.. _api_field_flyteidl.admin.TaskCreateRequest.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) id represents the unique identifier of the task.
  
  
.. _api_field_flyteidl.admin.TaskCreateRequest.spec:

spec
  (:ref:`flyteidl.admin.TaskSpec <api_msg_flyteidl.admin.TaskSpec>`) Represents the specification for task.
  
  


.. _api_msg_flyteidl.admin.TaskCreateResponse:

flyteidl.admin.TaskCreateResponse
---------------------------------

`[flyteidl.admin.TaskCreateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task.proto#L20>`_

Represents a response structure if task creation succeeds.

.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.Task:

flyteidl.admin.Task
-------------------

`[flyteidl.admin.Task proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task.proto#L27>`_

Flyte workflows are composed of many ordered tasks. That is small, reusable, self-contained logical blocks
arranged to process workflow inputs and produce a deterministic set of outputs.
Tasks can come in many varieties tuned for specialized behavior. 

.. code-block:: json

  {
    "id": "{...}",
    "closure": "{...}"
  }

.. _api_field_flyteidl.admin.Task.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) id represents the unique identifier of the task.
  
  
.. _api_field_flyteidl.admin.Task.closure:

closure
  (:ref:`flyteidl.admin.TaskClosure <api_msg_flyteidl.admin.TaskClosure>`) closure encapsulates all the fields that maps to a compiled version of the task.
  
  


.. _api_msg_flyteidl.admin.TaskList:

flyteidl.admin.TaskList
-----------------------

`[flyteidl.admin.TaskList proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task.proto#L37>`_

Represents a list of tasks returned from the admin.

.. code-block:: json

  {
    "tasks": [],
    "token": "..."
  }

.. _api_field_flyteidl.admin.TaskList.tasks:

tasks
  (:ref:`flyteidl.admin.Task <api_msg_flyteidl.admin.Task>`) A list of tasks returned based on the request.
  
  
.. _api_field_flyteidl.admin.TaskList.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query. If there are no more results, this value will be empty.
  
  


.. _api_msg_flyteidl.admin.TaskSpec:

flyteidl.admin.TaskSpec
-----------------------

`[flyteidl.admin.TaskSpec proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task.proto#L47>`_

Represents a structure that encapsulates the user-configured specification of the task.

.. code-block:: json

  {
    "template": "{...}"
  }

.. _api_field_flyteidl.admin.TaskSpec.template:

template
  (:ref:`flyteidl.core.TaskTemplate <api_msg_flyteidl.core.TaskTemplate>`) Template of the task that encapsulates all the metadata of the task.
  
  


.. _api_msg_flyteidl.admin.TaskClosure:

flyteidl.admin.TaskClosure
--------------------------

`[flyteidl.admin.TaskClosure proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task.proto#L54>`_

Compute task attributes which include values derived from the TaskSpec, as well as plugin-specific data
and task metadata.

.. code-block:: json

  {
    "compiled_task": "{...}",
    "created_at": "{...}"
  }

.. _api_field_flyteidl.admin.TaskClosure.compiled_task:

compiled_task
  (:ref:`flyteidl.core.CompiledTask <api_msg_flyteidl.core.CompiledTask>`) Represents the compiled representation of the task from the specification provided.
  
  
.. _api_field_flyteidl.admin.TaskClosure.created_at:

created_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Time at which the task was created.
  
  

