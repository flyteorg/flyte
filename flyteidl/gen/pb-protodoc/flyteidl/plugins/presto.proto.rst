.. _api_file_flyteidl/plugins/presto.proto:

presto.proto
=============================

.. _api_msg_flyteidl.plugins.PrestoQuery:

flyteidl.plugins.PrestoQuery
----------------------------

`[flyteidl.plugins.PrestoQuery proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/presto.proto#L10>`_

This message works with the 'presto' task type in the SDK and is the object that will be in the 'custom' field
of a Presto task's TaskTemplate

.. code-block:: json

  {
    "routing_group": "...",
    "catalog": "...",
    "schema": "...",
    "statement": "..."
  }

.. _api_field_flyteidl.plugins.PrestoQuery.routing_group:

routing_group
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.PrestoQuery.catalog:

catalog
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.PrestoQuery.schema:

schema
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.PrestoQuery.statement:

statement
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  

