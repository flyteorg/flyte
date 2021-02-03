.. _api_file_flyteidl/datacatalog/datacatalog.proto:

datacatalog.proto
======================================

.. _api_msg_pb.lyft.datacatalog.Parameter:

pb.lyft.datacatalog.Parameter
-----------------------------

`[pb.lyft.datacatalog.Parameter proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L7>`_


.. code-block:: json

  {
    "name": "...",
    "value": "..."
  }

.. _api_field_pb.lyft.datacatalog.Parameter.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.Parameter.value:

value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_pb.lyft.datacatalog.Artifact:

pb.lyft.datacatalog.Artifact
----------------------------

`[pb.lyft.datacatalog.Artifact proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L33>`_

Before jumping to message definition, lets go over the expected flow-
  An Artifact represents an unit-of-work identified by (task, version, inputs). This is
  encoded as unique hash for faster queries(called provenance). An artifact is persisted with some other
  attributes (revision, createdAt, reference_id, outputs).
  Only Discovery service knows about the hashing algorithm; one can use the closure (task, version, inputs)
  to query an artifact if it doesnt have the provenance value.

  Before starting the work on a task, programming-model first checks if the task has been done.
    Request:   GET (task, version, inputs)
    Response:  (Exists, Artifact) or (NotFound, nil)
  if not found, Task executor goes ahead with the execution and at the end of execution creates a new entry in
  the discovery service
    Request:  CREATE (task, version, inputs) + (revision, reference_id, outputs)
    Response: (Exists, Artifact) or (Created, Artifact)

  One can also Query all the artifacts by querying any subset of properties.
Message Artifact represents the complete information of an artifact- field that unique define the artifact +
properties.
Message ArtifactInternal is our storage model where we create an additional derived column for faster queries.
Message ArtifactId only represents field that uniquely define the artifact.

.. code-block:: json

  {
    "provenance": "...",
    "name": "...",
    "version": "...",
    "revision": "...",
    "created_at": "...",
    "reference_id": "...",
    "inputs": [],
    "outputs": []
  }

.. _api_field_pb.lyft.datacatalog.Artifact.provenance:

provenance
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.Artifact.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.Artifact.version:

version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.Artifact.revision:

revision
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.Artifact.created_at:

created_at
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.Artifact.reference_id:

reference_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.Artifact.inputs:

inputs
  (:ref:`pb.lyft.datacatalog.Parameter <api_msg_pb.lyft.datacatalog.Parameter>`) 
  
.. _api_field_pb.lyft.datacatalog.Artifact.outputs:

outputs
  (:ref:`pb.lyft.datacatalog.Parameter <api_msg_pb.lyft.datacatalog.Parameter>`) 
  


.. _api_msg_pb.lyft.datacatalog.ArtifactId:

pb.lyft.datacatalog.ArtifactId
------------------------------

`[pb.lyft.datacatalog.ArtifactId proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L44>`_


.. code-block:: json

  {
    "name": "...",
    "version": "...",
    "inputs": []
  }

.. _api_field_pb.lyft.datacatalog.ArtifactId.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.ArtifactId.version:

version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.ArtifactId.inputs:

inputs
  (:ref:`pb.lyft.datacatalog.Parameter <api_msg_pb.lyft.datacatalog.Parameter>`) 
  


.. _api_msg_pb.lyft.datacatalog.GetRequest:

pb.lyft.datacatalog.GetRequest
------------------------------

`[pb.lyft.datacatalog.GetRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L50>`_


.. code-block:: json

  {
    "provenance": "...",
    "artifact_id": "{...}"
  }

.. _api_field_pb.lyft.datacatalog.GetRequest.provenance:

provenance
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`provenance <api_field_pb.lyft.datacatalog.GetRequest.provenance>`, :ref:`artifact_id <api_field_pb.lyft.datacatalog.GetRequest.artifact_id>` may be set.
  
.. _api_field_pb.lyft.datacatalog.GetRequest.artifact_id:

artifact_id
  (:ref:`pb.lyft.datacatalog.ArtifactId <api_msg_pb.lyft.datacatalog.ArtifactId>`) 
  
  
  Only one of :ref:`provenance <api_field_pb.lyft.datacatalog.GetRequest.provenance>`, :ref:`artifact_id <api_field_pb.lyft.datacatalog.GetRequest.artifact_id>` may be set.
  


.. _api_msg_pb.lyft.datacatalog.GetResponse:

pb.lyft.datacatalog.GetResponse
-------------------------------

`[pb.lyft.datacatalog.GetResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L57>`_


.. code-block:: json

  {
    "artifact": "{...}"
  }

.. _api_field_pb.lyft.datacatalog.GetResponse.artifact:

artifact
  (:ref:`pb.lyft.datacatalog.Artifact <api_msg_pb.lyft.datacatalog.Artifact>`) 
  


.. _api_msg_pb.lyft.datacatalog.IntFilter:

pb.lyft.datacatalog.IntFilter
-----------------------------

`[pb.lyft.datacatalog.IntFilter proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L67>`_


.. code-block:: json

  {
    "value": "...",
    "operator": "..."
  }

.. _api_field_pb.lyft.datacatalog.IntFilter.value:

value
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.IntFilter.operator:

operator
  (:ref:`pb.lyft.datacatalog.QueryOperator <api_enum_pb.lyft.datacatalog.QueryOperator>`) 
  


.. _api_msg_pb.lyft.datacatalog.IntRangeFilter:

pb.lyft.datacatalog.IntRangeFilter
----------------------------------

`[pb.lyft.datacatalog.IntRangeFilter proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L72>`_


.. code-block:: json

  {
    "min": "...",
    "max": "..."
  }

.. _api_field_pb.lyft.datacatalog.IntRangeFilter.min:

min
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.IntRangeFilter.max:

max
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_pb.lyft.datacatalog.IntQueryKey:

pb.lyft.datacatalog.IntQueryKey
-------------------------------

`[pb.lyft.datacatalog.IntQueryKey proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L77>`_


.. code-block:: json

  {
    "val": "{...}",
    "range": "{...}"
  }

.. _api_field_pb.lyft.datacatalog.IntQueryKey.val:

val
  (:ref:`pb.lyft.datacatalog.IntFilter <api_msg_pb.lyft.datacatalog.IntFilter>`) 
  
  
  Only one of :ref:`val <api_field_pb.lyft.datacatalog.IntQueryKey.val>`, :ref:`range <api_field_pb.lyft.datacatalog.IntQueryKey.range>` may be set.
  
.. _api_field_pb.lyft.datacatalog.IntQueryKey.range:

range
  (:ref:`pb.lyft.datacatalog.IntRangeFilter <api_msg_pb.lyft.datacatalog.IntRangeFilter>`) 
  
  
  Only one of :ref:`val <api_field_pb.lyft.datacatalog.IntQueryKey.val>`, :ref:`range <api_field_pb.lyft.datacatalog.IntQueryKey.range>` may be set.
  


.. _api_msg_pb.lyft.datacatalog.QueryRequest:

pb.lyft.datacatalog.QueryRequest
--------------------------------

`[pb.lyft.datacatalog.QueryRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L86>`_

QueryRequest allows queries on a range of values for revision column and point queries on created_at
and reference_id

.. code-block:: json

  {
    "name": "...",
    "version": "...",
    "revision": "{...}",
    "created_at": "...",
    "reference_id": "..."
  }

.. _api_field_pb.lyft.datacatalog.QueryRequest.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.QueryRequest.version:

version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.QueryRequest.revision:

revision
  (:ref:`pb.lyft.datacatalog.IntQueryKey <api_msg_pb.lyft.datacatalog.IntQueryKey>`) 
  
.. _api_field_pb.lyft.datacatalog.QueryRequest.created_at:

created_at
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.QueryRequest.reference_id:

reference_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_pb.lyft.datacatalog.QueryResponse:

pb.lyft.datacatalog.QueryResponse
---------------------------------

`[pb.lyft.datacatalog.QueryResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L94>`_


.. code-block:: json

  {
    "artifact": []
  }

.. _api_field_pb.lyft.datacatalog.QueryResponse.artifact:

artifact
  (:ref:`pb.lyft.datacatalog.Artifact <api_msg_pb.lyft.datacatalog.Artifact>`) 
  


.. _api_msg_pb.lyft.datacatalog.CreateRequest:

pb.lyft.datacatalog.CreateRequest
---------------------------------

`[pb.lyft.datacatalog.CreateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L98>`_


.. code-block:: json

  {
    "ref": "{...}",
    "reference_id": "...",
    "revision": "...",
    "outputs": []
  }

.. _api_field_pb.lyft.datacatalog.CreateRequest.ref:

ref
  (:ref:`pb.lyft.datacatalog.ArtifactId <api_msg_pb.lyft.datacatalog.ArtifactId>`) 
  
.. _api_field_pb.lyft.datacatalog.CreateRequest.reference_id:

reference_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.CreateRequest.revision:

revision
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_pb.lyft.datacatalog.CreateRequest.outputs:

outputs
  (:ref:`pb.lyft.datacatalog.Parameter <api_msg_pb.lyft.datacatalog.Parameter>`) 
  


.. _api_msg_pb.lyft.datacatalog.CreateResponse:

pb.lyft.datacatalog.CreateResponse
----------------------------------

`[pb.lyft.datacatalog.CreateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L105>`_


.. code-block:: json

  {
    "artifact": "{...}",
    "status": "..."
  }

.. _api_field_pb.lyft.datacatalog.CreateResponse.artifact:

artifact
  (:ref:`pb.lyft.datacatalog.Artifact <api_msg_pb.lyft.datacatalog.Artifact>`) 
  
.. _api_field_pb.lyft.datacatalog.CreateResponse.status:

status
  (:ref:`pb.lyft.datacatalog.CreateResponse.Status <api_enum_pb.lyft.datacatalog.CreateResponse.Status>`) 
  

.. _api_enum_pb.lyft.datacatalog.CreateResponse.Status:

Enum pb.lyft.datacatalog.CreateResponse.Status
----------------------------------------------

`[pb.lyft.datacatalog.CreateResponse.Status proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L107>`_


.. _api_enum_value_pb.lyft.datacatalog.CreateResponse.Status.ALREADY_EXISTS:

ALREADY_EXISTS
  *(DEFAULT)* ⁣
  
.. _api_enum_value_pb.lyft.datacatalog.CreateResponse.Status.CREATED:

CREATED
  ⁣
  

.. _api_msg_pb.lyft.datacatalog.GenerateProvenanceRequest:

pb.lyft.datacatalog.GenerateProvenanceRequest
---------------------------------------------

`[pb.lyft.datacatalog.GenerateProvenanceRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L114>`_


.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_pb.lyft.datacatalog.GenerateProvenanceRequest.id:

id
  (:ref:`pb.lyft.datacatalog.ArtifactId <api_msg_pb.lyft.datacatalog.ArtifactId>`) 
  


.. _api_msg_pb.lyft.datacatalog.GenerateProvenanceResponse:

pb.lyft.datacatalog.GenerateProvenanceResponse
----------------------------------------------

`[pb.lyft.datacatalog.GenerateProvenanceResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L118>`_


.. code-block:: json

  {
    "provenance": "..."
  }

.. _api_field_pb.lyft.datacatalog.GenerateProvenanceResponse.provenance:

provenance
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  

.. _api_enum_pb.lyft.datacatalog.QueryOperator:

Enum pb.lyft.datacatalog.QueryOperator
--------------------------------------

`[pb.lyft.datacatalog.QueryOperator proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L61>`_


.. _api_enum_value_pb.lyft.datacatalog.QueryOperator.EQUAL:

EQUAL
  *(DEFAULT)* ⁣
  
.. _api_enum_value_pb.lyft.datacatalog.QueryOperator.GREATER_THAN:

GREATER_THAN
  ⁣
  
.. _api_enum_value_pb.lyft.datacatalog.QueryOperator.LESSER_THAN:

LESSER_THAN
  ⁣
  
