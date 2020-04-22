.. _api_file_flyteidl/core/literals.proto:

literals.proto
============================

.. _api_msg_flyteidl.core.Primitive:

flyteidl.core.Primitive
-----------------------

`[flyteidl.core.Primitive proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L12>`_

Primitive Types

.. code-block:: json

  {
    "integer": "...",
    "float_value": "...",
    "string_value": "...",
    "boolean": "...",
    "datetime": "{...}",
    "duration": "{...}"
  }

.. _api_field_flyteidl.core.Primitive.integer:

integer
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  Defines one of simple primitive types. These types will get translated into different programming languages as
  described in https://developers.google.com/protocol-buffers/docs/proto#scalar.
  
  
  Only one of :ref:`integer <api_field_flyteidl.core.Primitive.integer>`, :ref:`float_value <api_field_flyteidl.core.Primitive.float_value>`, :ref:`string_value <api_field_flyteidl.core.Primitive.string_value>`, :ref:`boolean <api_field_flyteidl.core.Primitive.boolean>`, :ref:`datetime <api_field_flyteidl.core.Primitive.datetime>`, :ref:`duration <api_field_flyteidl.core.Primitive.duration>` may be set.
  
.. _api_field_flyteidl.core.Primitive.float_value:

float_value
  (`double <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  Defines one of simple primitive types. These types will get translated into different programming languages as
  described in https://developers.google.com/protocol-buffers/docs/proto#scalar.
  
  
  Only one of :ref:`integer <api_field_flyteidl.core.Primitive.integer>`, :ref:`float_value <api_field_flyteidl.core.Primitive.float_value>`, :ref:`string_value <api_field_flyteidl.core.Primitive.string_value>`, :ref:`boolean <api_field_flyteidl.core.Primitive.boolean>`, :ref:`datetime <api_field_flyteidl.core.Primitive.datetime>`, :ref:`duration <api_field_flyteidl.core.Primitive.duration>` may be set.
  
.. _api_field_flyteidl.core.Primitive.string_value:

string_value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  Defines one of simple primitive types. These types will get translated into different programming languages as
  described in https://developers.google.com/protocol-buffers/docs/proto#scalar.
  
  
  Only one of :ref:`integer <api_field_flyteidl.core.Primitive.integer>`, :ref:`float_value <api_field_flyteidl.core.Primitive.float_value>`, :ref:`string_value <api_field_flyteidl.core.Primitive.string_value>`, :ref:`boolean <api_field_flyteidl.core.Primitive.boolean>`, :ref:`datetime <api_field_flyteidl.core.Primitive.datetime>`, :ref:`duration <api_field_flyteidl.core.Primitive.duration>` may be set.
  
.. _api_field_flyteidl.core.Primitive.boolean:

boolean
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  Defines one of simple primitive types. These types will get translated into different programming languages as
  described in https://developers.google.com/protocol-buffers/docs/proto#scalar.
  
  
  Only one of :ref:`integer <api_field_flyteidl.core.Primitive.integer>`, :ref:`float_value <api_field_flyteidl.core.Primitive.float_value>`, :ref:`string_value <api_field_flyteidl.core.Primitive.string_value>`, :ref:`boolean <api_field_flyteidl.core.Primitive.boolean>`, :ref:`datetime <api_field_flyteidl.core.Primitive.datetime>`, :ref:`duration <api_field_flyteidl.core.Primitive.duration>` may be set.
  
.. _api_field_flyteidl.core.Primitive.datetime:

datetime
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) 
  Defines one of simple primitive types. These types will get translated into different programming languages as
  described in https://developers.google.com/protocol-buffers/docs/proto#scalar.
  
  
  Only one of :ref:`integer <api_field_flyteidl.core.Primitive.integer>`, :ref:`float_value <api_field_flyteidl.core.Primitive.float_value>`, :ref:`string_value <api_field_flyteidl.core.Primitive.string_value>`, :ref:`boolean <api_field_flyteidl.core.Primitive.boolean>`, :ref:`datetime <api_field_flyteidl.core.Primitive.datetime>`, :ref:`duration <api_field_flyteidl.core.Primitive.duration>` may be set.
  
.. _api_field_flyteidl.core.Primitive.duration:

duration
  (:ref:`google.protobuf.Duration <api_msg_google.protobuf.Duration>`) 
  Defines one of simple primitive types. These types will get translated into different programming languages as
  described in https://developers.google.com/protocol-buffers/docs/proto#scalar.
  
  
  Only one of :ref:`integer <api_field_flyteidl.core.Primitive.integer>`, :ref:`float_value <api_field_flyteidl.core.Primitive.float_value>`, :ref:`string_value <api_field_flyteidl.core.Primitive.string_value>`, :ref:`boolean <api_field_flyteidl.core.Primitive.boolean>`, :ref:`datetime <api_field_flyteidl.core.Primitive.datetime>`, :ref:`duration <api_field_flyteidl.core.Primitive.duration>` may be set.
  


.. _api_msg_flyteidl.core.Void:

flyteidl.core.Void
------------------

`[flyteidl.core.Void proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L27>`_

Used to denote a nil/null/None assignment to a scalar value. The underlying LiteralType for Void is intentionally
undefined since it can be assigned to a scalar of any LiteralType.

.. code-block:: json

  {}




.. _api_msg_flyteidl.core.Blob:

flyteidl.core.Blob
------------------

`[flyteidl.core.Blob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L32>`_

Refers to an offloaded set of files. It encapsulates the type of the store and a unique uri for where the data is.
There are no restrictions on how the uri is formatted since it will depend on how to interact with the store.

.. code-block:: json

  {
    "metadata": "{...}",
    "uri": "..."
  }

.. _api_field_flyteidl.core.Blob.metadata:

metadata
  (:ref:`flyteidl.core.BlobMetadata <api_msg_flyteidl.core.BlobMetadata>`) 
  
.. _api_field_flyteidl.core.Blob.uri:

uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_flyteidl.core.BlobMetadata:

flyteidl.core.BlobMetadata
--------------------------

`[flyteidl.core.BlobMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L37>`_


.. code-block:: json

  {
    "type": "{...}"
  }

.. _api_field_flyteidl.core.BlobMetadata.type:

type
  (:ref:`flyteidl.core.BlobType <api_msg_flyteidl.core.BlobType>`) 
  


.. _api_msg_flyteidl.core.Binary:

flyteidl.core.Binary
--------------------

`[flyteidl.core.Binary proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L43>`_

A simple byte array with a tag to help different parts of the system communicate about what is in the byte array.
It's strongly advisable that consumers of this type define a unique tag and validate the tag before parsing the data.

.. code-block:: json

  {
    "value": "...",
    "tag": "..."
  }

.. _api_field_flyteidl.core.Binary.value:

value
  (`bytes <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.core.Binary.tag:

tag
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_flyteidl.core.Schema:

flyteidl.core.Schema
--------------------

`[flyteidl.core.Schema proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L49>`_

A strongly typed schema that defines the interface of data retrieved from the underlying storage medium.

.. code-block:: json

  {
    "uri": "...",
    "type": "{...}"
  }

.. _api_field_flyteidl.core.Schema.uri:

uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.core.Schema.type:

type
  (:ref:`flyteidl.core.SchemaType <api_msg_flyteidl.core.SchemaType>`) 
  


.. _api_msg_flyteidl.core.Scalar:

flyteidl.core.Scalar
--------------------

`[flyteidl.core.Scalar proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L54>`_


.. code-block:: json

  {
    "primitive": "{...}",
    "blob": "{...}",
    "binary": "{...}",
    "schema": "{...}",
    "none_type": "{...}",
    "error": "{...}",
    "generic": "{...}"
  }

.. _api_field_flyteidl.core.Scalar.primitive:

primitive
  (:ref:`flyteidl.core.Primitive <api_msg_flyteidl.core.Primitive>`) 
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Scalar.primitive>`, :ref:`blob <api_field_flyteidl.core.Scalar.blob>`, :ref:`binary <api_field_flyteidl.core.Scalar.binary>`, :ref:`schema <api_field_flyteidl.core.Scalar.schema>`, :ref:`none_type <api_field_flyteidl.core.Scalar.none_type>`, :ref:`error <api_field_flyteidl.core.Scalar.error>`, :ref:`generic <api_field_flyteidl.core.Scalar.generic>` may be set.
  
.. _api_field_flyteidl.core.Scalar.blob:

blob
  (:ref:`flyteidl.core.Blob <api_msg_flyteidl.core.Blob>`) 
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Scalar.primitive>`, :ref:`blob <api_field_flyteidl.core.Scalar.blob>`, :ref:`binary <api_field_flyteidl.core.Scalar.binary>`, :ref:`schema <api_field_flyteidl.core.Scalar.schema>`, :ref:`none_type <api_field_flyteidl.core.Scalar.none_type>`, :ref:`error <api_field_flyteidl.core.Scalar.error>`, :ref:`generic <api_field_flyteidl.core.Scalar.generic>` may be set.
  
.. _api_field_flyteidl.core.Scalar.binary:

binary
  (:ref:`flyteidl.core.Binary <api_msg_flyteidl.core.Binary>`) 
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Scalar.primitive>`, :ref:`blob <api_field_flyteidl.core.Scalar.blob>`, :ref:`binary <api_field_flyteidl.core.Scalar.binary>`, :ref:`schema <api_field_flyteidl.core.Scalar.schema>`, :ref:`none_type <api_field_flyteidl.core.Scalar.none_type>`, :ref:`error <api_field_flyteidl.core.Scalar.error>`, :ref:`generic <api_field_flyteidl.core.Scalar.generic>` may be set.
  
.. _api_field_flyteidl.core.Scalar.schema:

schema
  (:ref:`flyteidl.core.Schema <api_msg_flyteidl.core.Schema>`) 
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Scalar.primitive>`, :ref:`blob <api_field_flyteidl.core.Scalar.blob>`, :ref:`binary <api_field_flyteidl.core.Scalar.binary>`, :ref:`schema <api_field_flyteidl.core.Scalar.schema>`, :ref:`none_type <api_field_flyteidl.core.Scalar.none_type>`, :ref:`error <api_field_flyteidl.core.Scalar.error>`, :ref:`generic <api_field_flyteidl.core.Scalar.generic>` may be set.
  
.. _api_field_flyteidl.core.Scalar.none_type:

none_type
  (:ref:`flyteidl.core.Void <api_msg_flyteidl.core.Void>`) 
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Scalar.primitive>`, :ref:`blob <api_field_flyteidl.core.Scalar.blob>`, :ref:`binary <api_field_flyteidl.core.Scalar.binary>`, :ref:`schema <api_field_flyteidl.core.Scalar.schema>`, :ref:`none_type <api_field_flyteidl.core.Scalar.none_type>`, :ref:`error <api_field_flyteidl.core.Scalar.error>`, :ref:`generic <api_field_flyteidl.core.Scalar.generic>` may be set.
  
.. _api_field_flyteidl.core.Scalar.error:

error
  (:ref:`flyteidl.core.Error <api_msg_flyteidl.core.Error>`) 
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Scalar.primitive>`, :ref:`blob <api_field_flyteidl.core.Scalar.blob>`, :ref:`binary <api_field_flyteidl.core.Scalar.binary>`, :ref:`schema <api_field_flyteidl.core.Scalar.schema>`, :ref:`none_type <api_field_flyteidl.core.Scalar.none_type>`, :ref:`error <api_field_flyteidl.core.Scalar.error>`, :ref:`generic <api_field_flyteidl.core.Scalar.generic>` may be set.
  
.. _api_field_flyteidl.core.Scalar.generic:

generic
  (:ref:`google.protobuf.Struct <api_msg_google.protobuf.Struct>`) 
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Scalar.primitive>`, :ref:`blob <api_field_flyteidl.core.Scalar.blob>`, :ref:`binary <api_field_flyteidl.core.Scalar.binary>`, :ref:`schema <api_field_flyteidl.core.Scalar.schema>`, :ref:`none_type <api_field_flyteidl.core.Scalar.none_type>`, :ref:`error <api_field_flyteidl.core.Scalar.error>`, :ref:`generic <api_field_flyteidl.core.Scalar.generic>` may be set.
  


.. _api_msg_flyteidl.core.Literal:

flyteidl.core.Literal
---------------------

`[flyteidl.core.Literal proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L67>`_

A simple value. This supports any level of nesting (e.g. array of array of array of Blobs) as well as simple primitives.

.. code-block:: json

  {
    "scalar": "{...}",
    "collection": "{...}",
    "map": "{...}"
  }

.. _api_field_flyteidl.core.Literal.scalar:

scalar
  (:ref:`flyteidl.core.Scalar <api_msg_flyteidl.core.Scalar>`) A simple value.
  
  
  
  Only one of :ref:`scalar <api_field_flyteidl.core.Literal.scalar>`, :ref:`collection <api_field_flyteidl.core.Literal.collection>`, :ref:`map <api_field_flyteidl.core.Literal.map>` may be set.
  
.. _api_field_flyteidl.core.Literal.collection:

collection
  (:ref:`flyteidl.core.LiteralCollection <api_msg_flyteidl.core.LiteralCollection>`) A collection of literals to allow nesting.
  
  
  
  Only one of :ref:`scalar <api_field_flyteidl.core.Literal.scalar>`, :ref:`collection <api_field_flyteidl.core.Literal.collection>`, :ref:`map <api_field_flyteidl.core.Literal.map>` may be set.
  
.. _api_field_flyteidl.core.Literal.map:

map
  (:ref:`flyteidl.core.LiteralMap <api_msg_flyteidl.core.LiteralMap>`) A map of strings to literals.
  
  
  
  Only one of :ref:`scalar <api_field_flyteidl.core.Literal.scalar>`, :ref:`collection <api_field_flyteidl.core.Literal.collection>`, :ref:`map <api_field_flyteidl.core.Literal.map>` may be set.
  


.. _api_msg_flyteidl.core.LiteralCollection:

flyteidl.core.LiteralCollection
-------------------------------

`[flyteidl.core.LiteralCollection proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L81>`_

A collection of literals. This is a workaround since oneofs in proto messages cannot contain a repeated field.

.. code-block:: json

  {
    "literals": []
  }

.. _api_field_flyteidl.core.LiteralCollection.literals:

literals
  (:ref:`flyteidl.core.Literal <api_msg_flyteidl.core.Literal>`) 
  


.. _api_msg_flyteidl.core.LiteralMap:

flyteidl.core.LiteralMap
------------------------

`[flyteidl.core.LiteralMap proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L86>`_

A map of literals. This is a workaround since oneofs in proto messages cannot contain a repeated field.

.. code-block:: json

  {
    "literals": "{...}"
  }

.. _api_field_flyteidl.core.LiteralMap.literals:

literals
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`flyteidl.core.Literal <api_msg_flyteidl.core.Literal>`>) 
  


.. _api_msg_flyteidl.core.BindingDataCollection:

flyteidl.core.BindingDataCollection
-----------------------------------

`[flyteidl.core.BindingDataCollection proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L91>`_

A collection of BindingData items.

.. code-block:: json

  {
    "bindings": []
  }

.. _api_field_flyteidl.core.BindingDataCollection.bindings:

bindings
  (:ref:`flyteidl.core.BindingData <api_msg_flyteidl.core.BindingData>`) 
  


.. _api_msg_flyteidl.core.BindingDataMap:

flyteidl.core.BindingDataMap
----------------------------

`[flyteidl.core.BindingDataMap proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L96>`_

A map of BindingData items.

.. code-block:: json

  {
    "bindings": "{...}"
  }

.. _api_field_flyteidl.core.BindingDataMap.bindings:

bindings
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`flyteidl.core.BindingData <api_msg_flyteidl.core.BindingData>`>) 
  


.. _api_msg_flyteidl.core.BindingData:

flyteidl.core.BindingData
-------------------------

`[flyteidl.core.BindingData proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L101>`_

Specifies either a simple value or a reference to another output.

.. code-block:: json

  {
    "scalar": "{...}",
    "collection": "{...}",
    "promise": "{...}",
    "map": "{...}"
  }

.. _api_field_flyteidl.core.BindingData.scalar:

scalar
  (:ref:`flyteidl.core.Scalar <api_msg_flyteidl.core.Scalar>`) A simple scalar value.
  
  
  
  Only one of :ref:`scalar <api_field_flyteidl.core.BindingData.scalar>`, :ref:`collection <api_field_flyteidl.core.BindingData.collection>`, :ref:`promise <api_field_flyteidl.core.BindingData.promise>`, :ref:`map <api_field_flyteidl.core.BindingData.map>` may be set.
  
.. _api_field_flyteidl.core.BindingData.collection:

collection
  (:ref:`flyteidl.core.BindingDataCollection <api_msg_flyteidl.core.BindingDataCollection>`) A collection of binding data. This allows nesting of binding data to any number
  of levels.
  
  
  
  Only one of :ref:`scalar <api_field_flyteidl.core.BindingData.scalar>`, :ref:`collection <api_field_flyteidl.core.BindingData.collection>`, :ref:`promise <api_field_flyteidl.core.BindingData.promise>`, :ref:`map <api_field_flyteidl.core.BindingData.map>` may be set.
  
.. _api_field_flyteidl.core.BindingData.promise:

promise
  (:ref:`flyteidl.core.OutputReference <api_msg_flyteidl.core.OutputReference>`) References an output promised by another node.
  
  
  
  Only one of :ref:`scalar <api_field_flyteidl.core.BindingData.scalar>`, :ref:`collection <api_field_flyteidl.core.BindingData.collection>`, :ref:`promise <api_field_flyteidl.core.BindingData.promise>`, :ref:`map <api_field_flyteidl.core.BindingData.map>` may be set.
  
.. _api_field_flyteidl.core.BindingData.map:

map
  (:ref:`flyteidl.core.BindingDataMap <api_msg_flyteidl.core.BindingDataMap>`) A map of bindings. The key is always a string.
  
  
  
  Only one of :ref:`scalar <api_field_flyteidl.core.BindingData.scalar>`, :ref:`collection <api_field_flyteidl.core.BindingData.collection>`, :ref:`promise <api_field_flyteidl.core.BindingData.promise>`, :ref:`map <api_field_flyteidl.core.BindingData.map>` may be set.
  


.. _api_msg_flyteidl.core.Binding:

flyteidl.core.Binding
---------------------

`[flyteidl.core.Binding proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L119>`_

An input/output binding of a variable to either static value or a node output.

.. code-block:: json

  {
    "var": "...",
    "binding": "{...}"
  }

.. _api_field_flyteidl.core.Binding.var:

var
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Variable name must match an input/output variable of the node.
  
  
.. _api_field_flyteidl.core.Binding.binding:

binding
  (:ref:`flyteidl.core.BindingData <api_msg_flyteidl.core.BindingData>`) Data to use to bind this variable.
  
  


.. _api_msg_flyteidl.core.KeyValuePair:

flyteidl.core.KeyValuePair
--------------------------

`[flyteidl.core.KeyValuePair proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L128>`_

A generic key value pair.

.. code-block:: json

  {
    "key": "...",
    "value": "..."
  }

.. _api_field_flyteidl.core.KeyValuePair.key:

key
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) equired.
  
  
.. _api_field_flyteidl.core.KeyValuePair.value:

value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) optional.
  
  


.. _api_msg_flyteidl.core.RetryStrategy:

flyteidl.core.RetryStrategy
---------------------------

`[flyteidl.core.RetryStrategy proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L137>`_

Retry strategy associated with an executable unit.

.. code-block:: json

  {
    "retries": "..."
  }

.. _api_field_flyteidl.core.RetryStrategy.retries:

retries
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Number of retries. Retries will be consumed when the job fails with a recoverable error.
  The number of retries must be less than or equals to 10.
  
  

