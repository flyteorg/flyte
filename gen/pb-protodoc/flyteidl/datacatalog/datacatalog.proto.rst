.. _api_file_flyteidl/datacatalog/datacatalog.proto:

datacatalog.proto
======================================

.. _api_msg_datacatalog.CreateDatasetRequest:

datacatalog.CreateDatasetRequest
--------------------------------

`[datacatalog.CreateDatasetRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L69>`_


Request message for creating a Dataset.

.. code-block:: json

  {
    "dataset": "{...}"
  }

.. _api_field_datacatalog.CreateDatasetRequest.dataset:

dataset
  (:ref:`datacatalog.Dataset <api_msg_datacatalog.Dataset>`) 
  


.. _api_msg_datacatalog.CreateDatasetResponse:

datacatalog.CreateDatasetResponse
---------------------------------

`[datacatalog.CreateDatasetResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L76>`_


Response message for creating a Dataset

.. code-block:: json

  {}




.. _api_msg_datacatalog.GetDatasetRequest:

datacatalog.GetDatasetRequest
-----------------------------

`[datacatalog.GetDatasetRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L84>`_


Request message for retrieving a Dataset. The Dataset is retrieved by it's unique identifier
which is a combination of several fields.

.. code-block:: json

  {
    "dataset": "{...}"
  }

.. _api_field_datacatalog.GetDatasetRequest.dataset:

dataset
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) 
  


.. _api_msg_datacatalog.GetDatasetResponse:

datacatalog.GetDatasetResponse
------------------------------

`[datacatalog.GetDatasetResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L92>`_


Response message for retrieving a Dataset. The response will include the metadata for the
Dataset.

.. code-block:: json

  {
    "dataset": "{...}"
  }

.. _api_field_datacatalog.GetDatasetResponse.dataset:

dataset
  (:ref:`datacatalog.Dataset <api_msg_datacatalog.Dataset>`) 
  


.. _api_msg_datacatalog.GetArtifactRequest:

datacatalog.GetArtifactRequest
------------------------------

`[datacatalog.GetArtifactRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L101>`_


Request message for retrieving an Artifact. Retrieve an artifact based on a query handle that
can be one of artifact_id or tag. The result returned will include the artifact data and metadata
associated with the artifact.

.. code-block:: json

  {
    "dataset": "{...}",
    "artifact_id": "...",
    "tag_name": "..."
  }

.. _api_field_datacatalog.GetArtifactRequest.dataset:

dataset
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) 
  
.. _api_field_datacatalog.GetArtifactRequest.artifact_id:

artifact_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`artifact_id <api_field_datacatalog.GetArtifactRequest.artifact_id>`, :ref:`tag_name <api_field_datacatalog.GetArtifactRequest.tag_name>` may be set.
  
.. _api_field_datacatalog.GetArtifactRequest.tag_name:

tag_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`artifact_id <api_field_datacatalog.GetArtifactRequest.artifact_id>`, :ref:`tag_name <api_field_datacatalog.GetArtifactRequest.tag_name>` may be set.
  


.. _api_msg_datacatalog.GetArtifactResponse:

datacatalog.GetArtifactResponse
-------------------------------

`[datacatalog.GetArtifactResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L114>`_


Response message for retrieving an Artifact. The result returned will include the artifact data
and metadata associated with the artifact.

.. code-block:: json

  {
    "artifact": "{...}"
  }

.. _api_field_datacatalog.GetArtifactResponse.artifact:

artifact
  (:ref:`datacatalog.Artifact <api_msg_datacatalog.Artifact>`) 
  


.. _api_msg_datacatalog.CreateArtifactRequest:

datacatalog.CreateArtifactRequest
---------------------------------

`[datacatalog.CreateArtifactRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L121>`_


Request message for creating an Artifact and its associated artifact Data.

.. code-block:: json

  {
    "artifact": "{...}"
  }

.. _api_field_datacatalog.CreateArtifactRequest.artifact:

artifact
  (:ref:`datacatalog.Artifact <api_msg_datacatalog.Artifact>`) 
  


.. _api_msg_datacatalog.CreateArtifactResponse:

datacatalog.CreateArtifactResponse
----------------------------------

`[datacatalog.CreateArtifactResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L128>`_


Response message for creating an Artifact.

.. code-block:: json

  {}




.. _api_msg_datacatalog.AddTagRequest:

datacatalog.AddTagRequest
-------------------------

`[datacatalog.AddTagRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L135>`_


Request message for tagging an Artifact.

.. code-block:: json

  {
    "tag": "{...}"
  }

.. _api_field_datacatalog.AddTagRequest.tag:

tag
  (:ref:`datacatalog.Tag <api_msg_datacatalog.Tag>`) 
  


.. _api_msg_datacatalog.AddTagResponse:

datacatalog.AddTagResponse
--------------------------

`[datacatalog.AddTagResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L142>`_


Response message for tagging an Artifact.

.. code-block:: json

  {}




.. _api_msg_datacatalog.ListArtifactsRequest:

datacatalog.ListArtifactsRequest
--------------------------------

`[datacatalog.ListArtifactsRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L147>`_

List the artifacts that belong to the Dataset, optionally filtered using filtered expression.

.. code-block:: json

  {
    "dataset": "{...}",
    "filter": "{...}",
    "pagination": "{...}"
  }

.. _api_field_datacatalog.ListArtifactsRequest.dataset:

dataset
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) Use a datasetID for which you want to retrieve the artifacts
  
  
.. _api_field_datacatalog.ListArtifactsRequest.filter:

filter
  (:ref:`datacatalog.FilterExpression <api_msg_datacatalog.FilterExpression>`) Apply the filter expression to this query
  
  
.. _api_field_datacatalog.ListArtifactsRequest.pagination:

pagination
  (:ref:`datacatalog.PaginationOptions <api_msg_datacatalog.PaginationOptions>`) Pagination options to get a page of artifacts
  
  


.. _api_msg_datacatalog.ListArtifactsResponse:

datacatalog.ListArtifactsResponse
---------------------------------

`[datacatalog.ListArtifactsResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L158>`_

Response to list artifacts

.. code-block:: json

  {
    "artifacts": [],
    "next_token": "..."
  }

.. _api_field_datacatalog.ListArtifactsResponse.artifacts:

artifacts
  (:ref:`datacatalog.Artifact <api_msg_datacatalog.Artifact>`) The list of artifacts
  
  
.. _api_field_datacatalog.ListArtifactsResponse.next_token:

next_token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Token to use to request the next page, pass this into the next requests PaginationOptions
  
  


.. _api_msg_datacatalog.ListDatasetsRequest:

datacatalog.ListDatasetsRequest
-------------------------------

`[datacatalog.ListDatasetsRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L166>`_

List the datasets for the given query

.. code-block:: json

  {
    "filter": "{...}",
    "pagination": "{...}"
  }

.. _api_field_datacatalog.ListDatasetsRequest.filter:

filter
  (:ref:`datacatalog.FilterExpression <api_msg_datacatalog.FilterExpression>`) Apply the filter expression to this query
  
  
.. _api_field_datacatalog.ListDatasetsRequest.pagination:

pagination
  (:ref:`datacatalog.PaginationOptions <api_msg_datacatalog.PaginationOptions>`) Pagination options to get a page of datasets
  
  


.. _api_msg_datacatalog.ListDatasetsResponse:

datacatalog.ListDatasetsResponse
--------------------------------

`[datacatalog.ListDatasetsResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L174>`_

List the datasets response with token for next pagination

.. code-block:: json

  {
    "datasets": [],
    "next_token": "..."
  }

.. _api_field_datacatalog.ListDatasetsResponse.datasets:

datasets
  (:ref:`datacatalog.Dataset <api_msg_datacatalog.Dataset>`) The list of datasets
  
  
.. _api_field_datacatalog.ListDatasetsResponse.next_token:

next_token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Token to use to request the next page, pass this into the next requests PaginationOptions
  
  


.. _api_msg_datacatalog.GetOrReserveArtifactRequest:

datacatalog.GetOrReserveArtifactRequest
---------------------------------------

`[datacatalog.GetOrReserveArtifactRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L182>`_

Get the Artifact or try to reserve a spot if the Artifact does not exist.

.. code-block:: json

  {
    "dataset_id": "{...}",
    "tag_name": "...",
    "owner_id": "..."
  }

.. _api_field_datacatalog.GetOrReserveArtifactRequest.dataset_id:

dataset_id
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) 
  
.. _api_field_datacatalog.GetOrReserveArtifactRequest.tag_name:

tag_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.GetOrReserveArtifactRequest.owner_id:

owner_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_datacatalog.ReservationStatus:

datacatalog.ReservationStatus
-----------------------------

`[datacatalog.ReservationStatus proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L189>`_

Whether we successfully reserve a spot.

.. code-block:: json

  {
    "state": "...",
    "metadata": "{...}",
    "owner_id": "..."
  }

.. _api_field_datacatalog.ReservationStatus.state:

state
  (:ref:`datacatalog.ReservationStatus.State <api_enum_datacatalog.ReservationStatus.State>`) 
  
.. _api_field_datacatalog.ReservationStatus.metadata:

metadata
  (:ref:`datacatalog.Metadata <api_msg_datacatalog.Metadata>`) 
  
.. _api_field_datacatalog.ReservationStatus.owner_id:

owner_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  

.. _api_enum_datacatalog.ReservationStatus.State:

Enum datacatalog.ReservationStatus.State
----------------------------------------

`[datacatalog.ReservationStatus.State proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L190>`_


.. _api_enum_value_datacatalog.ReservationStatus.State.ACQUIRED:

ACQUIRED
  *(DEFAULT)* ⁣Acquired the reservation successfully.
  
  
.. _api_enum_value_datacatalog.ReservationStatus.State.ALREADY_IN_PROGRESS:

ALREADY_IN_PROGRESS
  ⁣Indicates an existing active reservation exist for a different owner_id.
  
  

.. _api_msg_datacatalog.GetOrReserveArtifactResponse:

datacatalog.GetOrReserveArtifactResponse
----------------------------------------

`[datacatalog.GetOrReserveArtifactResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L204>`_

Response to get artifact or reserve spot.

.. code-block:: json

  {
    "artifact": "{...}",
    "reservation_status": "{...}"
  }

.. _api_field_datacatalog.GetOrReserveArtifactResponse.artifact:

artifact
  (:ref:`datacatalog.Artifact <api_msg_datacatalog.Artifact>`) 
  
  
  Only one of :ref:`artifact <api_field_datacatalog.GetOrReserveArtifactResponse.artifact>`, :ref:`reservation_status <api_field_datacatalog.GetOrReserveArtifactResponse.reservation_status>` may be set.
  
.. _api_field_datacatalog.GetOrReserveArtifactResponse.reservation_status:

reservation_status
  (:ref:`datacatalog.ReservationStatus <api_msg_datacatalog.ReservationStatus>`) 
  
  
  Only one of :ref:`artifact <api_field_datacatalog.GetOrReserveArtifactResponse.artifact>`, :ref:`reservation_status <api_field_datacatalog.GetOrReserveArtifactResponse.reservation_status>` may be set.
  


.. _api_msg_datacatalog.ExtendReservationRequest:

datacatalog.ExtendReservationRequest
------------------------------------

`[datacatalog.ExtendReservationRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L212>`_

Request to extend reservation

.. code-block:: json

  {
    "dataset_id": "{...}",
    "tag_name": "...",
    "owner_id": "..."
  }

.. _api_field_datacatalog.ExtendReservationRequest.dataset_id:

dataset_id
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) 
  
.. _api_field_datacatalog.ExtendReservationRequest.tag_name:

tag_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.ExtendReservationRequest.owner_id:

owner_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_datacatalog.ExtendReservationResponse:

datacatalog.ExtendReservationResponse
-------------------------------------

`[datacatalog.ExtendReservationResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L219>`_

Response to extend reservation

.. code-block:: json

  {}




.. _api_msg_datacatalog.ReleaseReservationRequest:

datacatalog.ReleaseReservationRequest
-------------------------------------

`[datacatalog.ReleaseReservationRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L224>`_

Request to release reservation

.. code-block:: json

  {
    "dataset_id": "{...}",
    "tag_name": "...",
    "owner_id": "..."
  }

.. _api_field_datacatalog.ReleaseReservationRequest.dataset_id:

dataset_id
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) 
  
.. _api_field_datacatalog.ReleaseReservationRequest.tag_name:

tag_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.ReleaseReservationRequest.owner_id:

owner_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_datacatalog.ReleaseReservationResponse:

datacatalog.ReleaseReservationResponse
--------------------------------------

`[datacatalog.ReleaseReservationResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L231>`_

Response to release reservation

.. code-block:: json

  {}




.. _api_msg_datacatalog.Dataset:

datacatalog.Dataset
-------------------

`[datacatalog.Dataset proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L238>`_


Dataset message. It is uniquely identified by DatasetID.

.. code-block:: json

  {
    "id": "{...}",
    "metadata": "{...}",
    "partitionKeys": []
  }

.. _api_field_datacatalog.Dataset.id:

id
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) 
  
.. _api_field_datacatalog.Dataset.metadata:

metadata
  (:ref:`datacatalog.Metadata <api_msg_datacatalog.Metadata>`) 
  
.. _api_field_datacatalog.Dataset.partitionKeys:

partitionKeys
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_datacatalog.Partition:

datacatalog.Partition
---------------------

`[datacatalog.Partition proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L247>`_


An artifact could have multiple partitions and each partition can have an arbitrary string key/value pair

.. code-block:: json

  {
    "key": "...",
    "value": "..."
  }

.. _api_field_datacatalog.Partition.key:

key
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.Partition.value:

value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_datacatalog.DatasetID:

datacatalog.DatasetID
---------------------

`[datacatalog.DatasetID proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L255>`_


DatasetID message that is composed of several string fields.

.. code-block:: json

  {
    "project": "...",
    "name": "...",
    "domain": "...",
    "version": "...",
    "UUID": "..."
  }

.. _api_field_datacatalog.DatasetID.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.DatasetID.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.DatasetID.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.DatasetID.version:

version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.DatasetID.UUID:

UUID
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_datacatalog.Artifact:

datacatalog.Artifact
--------------------

`[datacatalog.Artifact proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L266>`_


Artifact message. It is composed of several string fields.

.. code-block:: json

  {
    "id": "...",
    "dataset": "{...}",
    "data": [],
    "metadata": "{...}",
    "partitions": [],
    "tags": [],
    "created_at": "{...}"
  }

.. _api_field_datacatalog.Artifact.id:

id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.Artifact.dataset:

dataset
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) 
  
.. _api_field_datacatalog.Artifact.data:

data
  (:ref:`datacatalog.ArtifactData <api_msg_datacatalog.ArtifactData>`) 
  
.. _api_field_datacatalog.Artifact.metadata:

metadata
  (:ref:`datacatalog.Metadata <api_msg_datacatalog.Metadata>`) 
  
.. _api_field_datacatalog.Artifact.partitions:

partitions
  (:ref:`datacatalog.Partition <api_msg_datacatalog.Partition>`) 
  
.. _api_field_datacatalog.Artifact.tags:

tags
  (:ref:`datacatalog.Tag <api_msg_datacatalog.Tag>`) 
  
.. _api_field_datacatalog.Artifact.created_at:

created_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) 
  


.. _api_msg_datacatalog.ArtifactData:

datacatalog.ArtifactData
------------------------

`[datacatalog.ArtifactData proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L279>`_


ArtifactData that belongs to an artifact

.. code-block:: json

  {
    "name": "...",
    "value": "{...}"
  }

.. _api_field_datacatalog.ArtifactData.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.ArtifactData.value:

value
  (:ref:`flyteidl.core.Literal <api_msg_flyteidl.core.Literal>`) 
  


.. _api_msg_datacatalog.Tag:

datacatalog.Tag
---------------

`[datacatalog.Tag proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L288>`_


Tag message that is unique to a Dataset. It is associated to a single artifact and
can be retrieved by name later.

.. code-block:: json

  {
    "name": "...",
    "artifact_id": "...",
    "dataset": "{...}"
  }

.. _api_field_datacatalog.Tag.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.Tag.artifact_id:

artifact_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.Tag.dataset:

dataset
  (:ref:`datacatalog.DatasetID <api_msg_datacatalog.DatasetID>`) 
  


.. _api_msg_datacatalog.Metadata:

datacatalog.Metadata
--------------------

`[datacatalog.Metadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L297>`_


Metadata representation for artifacts and datasets

.. code-block:: json

  {
    "key_map": "{...}"
  }

.. _api_field_datacatalog.Metadata.key_map:

key_map
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, `string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_>) 
  


.. _api_msg_datacatalog.FilterExpression:

datacatalog.FilterExpression
----------------------------

`[datacatalog.FilterExpression proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L302>`_

Filter expression that is composed of a combination of single filters

.. code-block:: json

  {
    "filters": []
  }

.. _api_field_datacatalog.FilterExpression.filters:

filters
  (:ref:`datacatalog.SinglePropertyFilter <api_msg_datacatalog.SinglePropertyFilter>`) 
  


.. _api_msg_datacatalog.SinglePropertyFilter:

datacatalog.SinglePropertyFilter
--------------------------------

`[datacatalog.SinglePropertyFilter proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L307>`_

A single property to filter on.

.. code-block:: json

  {
    "tag_filter": "{...}",
    "partition_filter": "{...}",
    "artifact_filter": "{...}",
    "dataset_filter": "{...}",
    "operator": "..."
  }

.. _api_field_datacatalog.SinglePropertyFilter.tag_filter:

tag_filter
  (:ref:`datacatalog.TagPropertyFilter <api_msg_datacatalog.TagPropertyFilter>`) 
  
  
  Only one of :ref:`tag_filter <api_field_datacatalog.SinglePropertyFilter.tag_filter>`, :ref:`partition_filter <api_field_datacatalog.SinglePropertyFilter.partition_filter>`, :ref:`artifact_filter <api_field_datacatalog.SinglePropertyFilter.artifact_filter>`, :ref:`dataset_filter <api_field_datacatalog.SinglePropertyFilter.dataset_filter>` may be set.
  
.. _api_field_datacatalog.SinglePropertyFilter.partition_filter:

partition_filter
  (:ref:`datacatalog.PartitionPropertyFilter <api_msg_datacatalog.PartitionPropertyFilter>`) 
  
  
  Only one of :ref:`tag_filter <api_field_datacatalog.SinglePropertyFilter.tag_filter>`, :ref:`partition_filter <api_field_datacatalog.SinglePropertyFilter.partition_filter>`, :ref:`artifact_filter <api_field_datacatalog.SinglePropertyFilter.artifact_filter>`, :ref:`dataset_filter <api_field_datacatalog.SinglePropertyFilter.dataset_filter>` may be set.
  
.. _api_field_datacatalog.SinglePropertyFilter.artifact_filter:

artifact_filter
  (:ref:`datacatalog.ArtifactPropertyFilter <api_msg_datacatalog.ArtifactPropertyFilter>`) 
  
  
  Only one of :ref:`tag_filter <api_field_datacatalog.SinglePropertyFilter.tag_filter>`, :ref:`partition_filter <api_field_datacatalog.SinglePropertyFilter.partition_filter>`, :ref:`artifact_filter <api_field_datacatalog.SinglePropertyFilter.artifact_filter>`, :ref:`dataset_filter <api_field_datacatalog.SinglePropertyFilter.dataset_filter>` may be set.
  
.. _api_field_datacatalog.SinglePropertyFilter.dataset_filter:

dataset_filter
  (:ref:`datacatalog.DatasetPropertyFilter <api_msg_datacatalog.DatasetPropertyFilter>`) 
  
  
  Only one of :ref:`tag_filter <api_field_datacatalog.SinglePropertyFilter.tag_filter>`, :ref:`partition_filter <api_field_datacatalog.SinglePropertyFilter.partition_filter>`, :ref:`artifact_filter <api_field_datacatalog.SinglePropertyFilter.artifact_filter>`, :ref:`dataset_filter <api_field_datacatalog.SinglePropertyFilter.dataset_filter>` may be set.
  
.. _api_field_datacatalog.SinglePropertyFilter.operator:

operator
  (:ref:`datacatalog.SinglePropertyFilter.ComparisonOperator <api_enum_datacatalog.SinglePropertyFilter.ComparisonOperator>`) 
  

.. _api_enum_datacatalog.SinglePropertyFilter.ComparisonOperator:

Enum datacatalog.SinglePropertyFilter.ComparisonOperator
--------------------------------------------------------

`[datacatalog.SinglePropertyFilter.ComparisonOperator proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L316>`_

as use-cases come up we can add more operators, ex: gte, like, not eq etc.

.. _api_enum_value_datacatalog.SinglePropertyFilter.ComparisonOperator.EQUALS:

EQUALS
  *(DEFAULT)* ⁣
  

.. _api_msg_datacatalog.ArtifactPropertyFilter:

datacatalog.ArtifactPropertyFilter
----------------------------------

`[datacatalog.ArtifactPropertyFilter proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L325>`_

Artifact properties we can filter by

.. code-block:: json

  {
    "artifact_id": "..."
  }

.. _api_field_datacatalog.ArtifactPropertyFilter.artifact_id:

artifact_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  oneof because we can add more properties in the future
  
  


.. _api_msg_datacatalog.TagPropertyFilter:

datacatalog.TagPropertyFilter
-----------------------------

`[datacatalog.TagPropertyFilter proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L333>`_

Tag properties we can filter by

.. code-block:: json

  {
    "tag_name": "..."
  }

.. _api_field_datacatalog.TagPropertyFilter.tag_name:

tag_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  


.. _api_msg_datacatalog.PartitionPropertyFilter:

datacatalog.PartitionPropertyFilter
-----------------------------------

`[datacatalog.PartitionPropertyFilter proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L340>`_

Partition properties we can filter by

.. code-block:: json

  {
    "key_val": "{...}"
  }

.. _api_field_datacatalog.PartitionPropertyFilter.key_val:

key_val
  (:ref:`datacatalog.KeyValuePair <api_msg_datacatalog.KeyValuePair>`) 
  
  


.. _api_msg_datacatalog.KeyValuePair:

datacatalog.KeyValuePair
------------------------

`[datacatalog.KeyValuePair proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L346>`_


.. code-block:: json

  {
    "key": "...",
    "value": "..."
  }

.. _api_field_datacatalog.KeyValuePair.key:

key
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_datacatalog.KeyValuePair.value:

value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_datacatalog.DatasetPropertyFilter:

datacatalog.DatasetPropertyFilter
---------------------------------

`[datacatalog.DatasetPropertyFilter proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L352>`_

Dataset properties we can filter by

.. code-block:: json

  {
    "project": "...",
    "name": "...",
    "domain": "...",
    "version": "..."
  }

.. _api_field_datacatalog.DatasetPropertyFilter.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`project <api_field_datacatalog.DatasetPropertyFilter.project>`, :ref:`name <api_field_datacatalog.DatasetPropertyFilter.name>`, :ref:`domain <api_field_datacatalog.DatasetPropertyFilter.domain>`, :ref:`version <api_field_datacatalog.DatasetPropertyFilter.version>` may be set.
  
.. _api_field_datacatalog.DatasetPropertyFilter.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`project <api_field_datacatalog.DatasetPropertyFilter.project>`, :ref:`name <api_field_datacatalog.DatasetPropertyFilter.name>`, :ref:`domain <api_field_datacatalog.DatasetPropertyFilter.domain>`, :ref:`version <api_field_datacatalog.DatasetPropertyFilter.version>` may be set.
  
.. _api_field_datacatalog.DatasetPropertyFilter.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`project <api_field_datacatalog.DatasetPropertyFilter.project>`, :ref:`name <api_field_datacatalog.DatasetPropertyFilter.name>`, :ref:`domain <api_field_datacatalog.DatasetPropertyFilter.domain>`, :ref:`version <api_field_datacatalog.DatasetPropertyFilter.version>` may be set.
  
.. _api_field_datacatalog.DatasetPropertyFilter.version:

version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`project <api_field_datacatalog.DatasetPropertyFilter.project>`, :ref:`name <api_field_datacatalog.DatasetPropertyFilter.name>`, :ref:`domain <api_field_datacatalog.DatasetPropertyFilter.domain>`, :ref:`version <api_field_datacatalog.DatasetPropertyFilter.version>` may be set.
  


.. _api_msg_datacatalog.PaginationOptions:

datacatalog.PaginationOptions
-----------------------------

`[datacatalog.PaginationOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L362>`_

Pagination options for making list requests

.. code-block:: json

  {
    "limit": "...",
    "token": "...",
    "sortKey": "...",
    "sortOrder": "..."
  }

.. _api_field_datacatalog.PaginationOptions.limit:

limit
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) the max number of results to return
  
  
.. _api_field_datacatalog.PaginationOptions.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) the token to pass to fetch the next page
  
  
.. _api_field_datacatalog.PaginationOptions.sortKey:

sortKey
  (:ref:`datacatalog.PaginationOptions.SortKey <api_enum_datacatalog.PaginationOptions.SortKey>`) the property that we want to sort the results by
  
  
.. _api_field_datacatalog.PaginationOptions.sortOrder:

sortOrder
  (:ref:`datacatalog.PaginationOptions.SortOrder <api_enum_datacatalog.PaginationOptions.SortOrder>`) the sort order of the results
  
  

.. _api_enum_datacatalog.PaginationOptions.SortOrder:

Enum datacatalog.PaginationOptions.SortOrder
--------------------------------------------

`[datacatalog.PaginationOptions.SortOrder proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L376>`_


.. _api_enum_value_datacatalog.PaginationOptions.SortOrder.DESCENDING:

DESCENDING
  *(DEFAULT)* ⁣
  
.. _api_enum_value_datacatalog.PaginationOptions.SortOrder.ASCENDING:

ASCENDING
  ⁣
  

.. _api_enum_datacatalog.PaginationOptions.SortKey:

Enum datacatalog.PaginationOptions.SortKey
------------------------------------------

`[datacatalog.PaginationOptions.SortKey proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/datacatalog/datacatalog.proto#L381>`_


.. _api_enum_value_datacatalog.PaginationOptions.SortKey.CREATION_TIME:

CREATION_TIME
  *(DEFAULT)* ⁣
  
