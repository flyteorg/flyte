.. _api_file_flyteidl/plugins/qubole.proto:

qubole.proto
=============================

.. _api_msg_flyteidl.plugins.HiveQuery:

flyteidl.plugins.HiveQuery
--------------------------

`[flyteidl.plugins.HiveQuery proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/qubole.proto#L9>`_

Defines a query to execute on a hive cluster.

.. code-block:: json

  {
    "query": "...",
    "timeout_sec": "...",
    "retryCount": "..."
  }

.. _api_field_flyteidl.plugins.HiveQuery.query:

query
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.HiveQuery.timeout_sec:

timeout_sec
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.HiveQuery.retryCount:

retryCount
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_flyteidl.plugins.HiveQueryCollection:

flyteidl.plugins.HiveQueryCollection
------------------------------------

`[flyteidl.plugins.HiveQueryCollection proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/qubole.proto#L16>`_

Defines a collection of hive queries.

.. code-block:: json

  {
    "queries": []
  }

.. _api_field_flyteidl.plugins.HiveQueryCollection.queries:

queries
  (:ref:`flyteidl.plugins.HiveQuery <api_msg_flyteidl.plugins.HiveQuery>`) 
  


.. _api_msg_flyteidl.plugins.QuboleHiveJob:

flyteidl.plugins.QuboleHiveJob
------------------------------

`[flyteidl.plugins.QuboleHiveJob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/qubole.proto#L22>`_

This message works with the 'hive' task type in the SDK and is the object that will be in the 'custom' field
of a hive task's TaskTemplate

.. code-block:: json

  {
    "cluster_label": "...",
    "query_collection": "{...}",
    "tags": [],
    "query": "{...}"
  }

.. _api_field_flyteidl.plugins.QuboleHiveJob.cluster_label:

cluster_label
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.QuboleHiveJob.query_collection:

query_collection
  (:ref:`flyteidl.plugins.HiveQueryCollection <api_msg_flyteidl.plugins.HiveQueryCollection>`) 
  
.. _api_field_flyteidl.plugins.QuboleHiveJob.tags:

tags
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.QuboleHiveJob.query:

query
  (:ref:`flyteidl.plugins.HiveQuery <api_msg_flyteidl.plugins.HiveQuery>`) 
  

