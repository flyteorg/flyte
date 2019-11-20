.. _api_file_flyteidl/core/types.proto:

types.proto
=========================

.. _api_msg_flyteidl.core.SchemaType:

flyteidl.core.SchemaType
------------------------

`[flyteidl.core.SchemaType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L23>`_

Defines schema columns and types to strongly type-validate schemas interoperability.

.. code-block:: json

  {
    "columns": []
  }

.. _api_field_flyteidl.core.SchemaType.columns:

columns
  (:ref:`flyteidl.core.SchemaType.SchemaColumn <api_msg_flyteidl.core.SchemaType.SchemaColumn>`) A list of ordered columns this schema comprises of.
  
  
.. _api_msg_flyteidl.core.SchemaType.SchemaColumn:

flyteidl.core.SchemaType.SchemaColumn
-------------------------------------

`[flyteidl.core.SchemaType.SchemaColumn proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L24>`_


.. code-block:: json

  {
    "name": "...",
    "type": "..."
  }

.. _api_field_flyteidl.core.SchemaType.SchemaColumn.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A unique name -within the schema type- for the column
  
  
.. _api_field_flyteidl.core.SchemaType.SchemaColumn.type:

type
  (:ref:`flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType <api_enum_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType>`) The column type. This allows a limited set of types currently.
  
  

.. _api_enum_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType:

Enum flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType
-----------------------------------------------------------

`[flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L28>`_


.. _api_enum_value_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType.INTEGER:

INTEGER
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType.FLOAT:

FLOAT
  ⁣
  
.. _api_enum_value_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType.STRING:

STRING
  ⁣
  
.. _api_enum_value_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType.BOOLEAN:

BOOLEAN
  ⁣
  
.. _api_enum_value_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType.DATETIME:

DATETIME
  ⁣
  
.. _api_enum_value_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType.DURATION:

DURATION
  ⁣
  


.. _api_msg_flyteidl.core.BlobType:

flyteidl.core.BlobType
----------------------

`[flyteidl.core.BlobType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L46>`_

Defines type behavior for blob objects

.. code-block:: json

  {
    "format": "...",
    "dimensionality": "..."
  }

.. _api_field_flyteidl.core.BlobType.format:

format
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Format can be a free form string understood by SDK/UI etc like
  csv, parquet etc
  
  
.. _api_field_flyteidl.core.BlobType.dimensionality:

dimensionality
  (:ref:`flyteidl.core.BlobType.BlobDimensionality <api_enum_flyteidl.core.BlobType.BlobDimensionality>`) 
  

.. _api_enum_flyteidl.core.BlobType.BlobDimensionality:

Enum flyteidl.core.BlobType.BlobDimensionality
----------------------------------------------

`[flyteidl.core.BlobType.BlobDimensionality proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L47>`_


.. _api_enum_value_flyteidl.core.BlobType.BlobDimensionality.SINGLE:

SINGLE
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.BlobType.BlobDimensionality.MULTIPART:

MULTIPART
  ⁣
  

.. _api_msg_flyteidl.core.LiteralType:

flyteidl.core.LiteralType
-------------------------

`[flyteidl.core.LiteralType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L59>`_

Defines a strong type to allow type checking between interfaces.

.. code-block:: json

  {
    "simple": "...",
    "schema": "{...}",
    "collection_type": "{...}",
    "map_value_type": "{...}",
    "blob": "{...}",
    "metadata": "{...}"
  }

.. _api_field_flyteidl.core.LiteralType.simple:

simple
  (:ref:`flyteidl.core.SimpleType <api_enum_flyteidl.core.SimpleType>`) A simple type that can be compared one-to-one with another.
  
  
  
  Only one of :ref:`simple <api_field_flyteidl.core.LiteralType.simple>`, :ref:`schema <api_field_flyteidl.core.LiteralType.schema>`, :ref:`collection_type <api_field_flyteidl.core.LiteralType.collection_type>`, :ref:`map_value_type <api_field_flyteidl.core.LiteralType.map_value_type>`, :ref:`blob <api_field_flyteidl.core.LiteralType.blob>` may be set.
  
.. _api_field_flyteidl.core.LiteralType.schema:

schema
  (:ref:`flyteidl.core.SchemaType <api_msg_flyteidl.core.SchemaType>`) A complex type that requires matching of inner fields.
  
  
  
  Only one of :ref:`simple <api_field_flyteidl.core.LiteralType.simple>`, :ref:`schema <api_field_flyteidl.core.LiteralType.schema>`, :ref:`collection_type <api_field_flyteidl.core.LiteralType.collection_type>`, :ref:`map_value_type <api_field_flyteidl.core.LiteralType.map_value_type>`, :ref:`blob <api_field_flyteidl.core.LiteralType.blob>` may be set.
  
.. _api_field_flyteidl.core.LiteralType.collection_type:

collection_type
  (:ref:`flyteidl.core.LiteralType <api_msg_flyteidl.core.LiteralType>`) Defines the type of the value of a collection. Only homogeneous collections are allowed.
  
  
  
  Only one of :ref:`simple <api_field_flyteidl.core.LiteralType.simple>`, :ref:`schema <api_field_flyteidl.core.LiteralType.schema>`, :ref:`collection_type <api_field_flyteidl.core.LiteralType.collection_type>`, :ref:`map_value_type <api_field_flyteidl.core.LiteralType.map_value_type>`, :ref:`blob <api_field_flyteidl.core.LiteralType.blob>` may be set.
  
.. _api_field_flyteidl.core.LiteralType.map_value_type:

map_value_type
  (:ref:`flyteidl.core.LiteralType <api_msg_flyteidl.core.LiteralType>`) Defines the type of the value of a map type. The type of the key is always a string.
  
  
  
  Only one of :ref:`simple <api_field_flyteidl.core.LiteralType.simple>`, :ref:`schema <api_field_flyteidl.core.LiteralType.schema>`, :ref:`collection_type <api_field_flyteidl.core.LiteralType.collection_type>`, :ref:`map_value_type <api_field_flyteidl.core.LiteralType.map_value_type>`, :ref:`blob <api_field_flyteidl.core.LiteralType.blob>` may be set.
  
.. _api_field_flyteidl.core.LiteralType.blob:

blob
  (:ref:`flyteidl.core.BlobType <api_msg_flyteidl.core.BlobType>`) A blob might have specialized implementation details depending on associated metadata.
  
  
  
  Only one of :ref:`simple <api_field_flyteidl.core.LiteralType.simple>`, :ref:`schema <api_field_flyteidl.core.LiteralType.schema>`, :ref:`collection_type <api_field_flyteidl.core.LiteralType.collection_type>`, :ref:`map_value_type <api_field_flyteidl.core.LiteralType.map_value_type>`, :ref:`blob <api_field_flyteidl.core.LiteralType.blob>` may be set.
  
.. _api_field_flyteidl.core.LiteralType.metadata:

metadata
  (:ref:`google.protobuf.Struct <api_msg_google.protobuf.Struct>`) This field contains type metadata that is descriptive of the type, but is NOT considered in type-checking.  This might be used by
  consumers to identify special behavior or display extended information for the type.
  
  


.. _api_msg_flyteidl.core.OutputReference:

flyteidl.core.OutputReference
-----------------------------

`[flyteidl.core.OutputReference proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L84>`_

A reference to an output produced by a node. The type can be retrieved -and validated- from
the underlying interface of the node.

.. code-block:: json

  {
    "node_id": "...",
    "var": "..."
  }

.. _api_field_flyteidl.core.OutputReference.node_id:

node_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Node id must exist at the graph layer.
  
  
.. _api_field_flyteidl.core.OutputReference.var:

var
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Variable name must refer to an output variable for the node.
  
  


.. _api_msg_flyteidl.core.Error:

flyteidl.core.Error
-------------------

`[flyteidl.core.Error proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L93>`_

Represents an error thrown from a node.

.. code-block:: json

  {
    "failed_node_id": "...",
    "message": "..."
  }

.. _api_field_flyteidl.core.Error.failed_node_id:

failed_node_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The node id that threw the error.
  
  
.. _api_field_flyteidl.core.Error.message:

message
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Error message thrown.
  
  

.. _api_enum_flyteidl.core.SimpleType:

Enum flyteidl.core.SimpleType
-----------------------------

`[flyteidl.core.SimpleType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/types.proto#L9>`_

Define a set of simple types.

.. _api_enum_value_flyteidl.core.SimpleType.NONE:

NONE
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.INTEGER:

INTEGER
  ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.FLOAT:

FLOAT
  ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.STRING:

STRING
  ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.BOOLEAN:

BOOLEAN
  ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.DATETIME:

DATETIME
  ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.DURATION:

DURATION
  ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.BINARY:

BINARY
  ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.ERROR:

ERROR
  ⁣
  
.. _api_enum_value_flyteidl.core.SimpleType.STRUCT:

STRUCT
  ⁣
  
