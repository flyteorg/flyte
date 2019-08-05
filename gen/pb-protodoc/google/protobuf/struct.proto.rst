.. _api_file_google/protobuf/struct.proto:

struct.proto
============================

Protocol Buffers - Google's data interchange format
Copyright 2008 Google Inc.  All rights reserved.
https://developers.google.com/protocol-buffers/

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

    * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
    * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

.. _api_msg_google.protobuf.Struct:

google.protobuf.Struct
----------------------

`[google.protobuf.Struct proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/struct.proto#L51>`_

`Struct` represents a structured data value, consisting of fields
which map to dynamically typed values. In some languages, `Struct`
might be supported by a native representation. For example, in
scripting languages like JS a struct is represented as an
object. The details of that representation are described together
with the proto support for the language.

The JSON representation for `Struct` is JSON object.

.. code-block:: json

  {
    "fields": "{...}"
  }

.. _api_field_google.protobuf.Struct.fields:

fields
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`google.protobuf.Value <api_msg_google.protobuf.Value>`>) Unordered map of dynamically typed values.
  
  


.. _api_msg_google.protobuf.Value:

google.protobuf.Value
---------------------

`[google.protobuf.Value proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/struct.proto#L62>`_

`Value` represents a dynamically typed value which can be either
null, a number, a string, a boolean, a recursive struct value, or a
list of values. A producer of value is expected to set one of that
variants, absence of any variant indicates an error.

The JSON representation for `Value` is JSON value.

.. code-block:: json

  {
    "null_value": "...",
    "number_value": "...",
    "string_value": "...",
    "bool_value": "...",
    "struct_value": "{...}",
    "list_value": "{...}"
  }

.. _api_field_google.protobuf.Value.null_value:

null_value
  (:ref:`google.protobuf.NullValue <api_enum_google.protobuf.NullValue>`) Represents a null value.
  
  The kind of value.
  
  
  Only one of :ref:`null_value <api_field_google.protobuf.Value.null_value>`, :ref:`number_value <api_field_google.protobuf.Value.number_value>`, :ref:`string_value <api_field_google.protobuf.Value.string_value>`, :ref:`bool_value <api_field_google.protobuf.Value.bool_value>`, :ref:`struct_value <api_field_google.protobuf.Value.struct_value>`, :ref:`list_value <api_field_google.protobuf.Value.list_value>` may be set.
  
.. _api_field_google.protobuf.Value.number_value:

number_value
  (`double <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Represents a double value.
  
  The kind of value.
  
  
  Only one of :ref:`null_value <api_field_google.protobuf.Value.null_value>`, :ref:`number_value <api_field_google.protobuf.Value.number_value>`, :ref:`string_value <api_field_google.protobuf.Value.string_value>`, :ref:`bool_value <api_field_google.protobuf.Value.bool_value>`, :ref:`struct_value <api_field_google.protobuf.Value.struct_value>`, :ref:`list_value <api_field_google.protobuf.Value.list_value>` may be set.
  
.. _api_field_google.protobuf.Value.string_value:

string_value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Represents a string value.
  
  The kind of value.
  
  
  Only one of :ref:`null_value <api_field_google.protobuf.Value.null_value>`, :ref:`number_value <api_field_google.protobuf.Value.number_value>`, :ref:`string_value <api_field_google.protobuf.Value.string_value>`, :ref:`bool_value <api_field_google.protobuf.Value.bool_value>`, :ref:`struct_value <api_field_google.protobuf.Value.struct_value>`, :ref:`list_value <api_field_google.protobuf.Value.list_value>` may be set.
  
.. _api_field_google.protobuf.Value.bool_value:

bool_value
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Represents a boolean value.
  
  The kind of value.
  
  
  Only one of :ref:`null_value <api_field_google.protobuf.Value.null_value>`, :ref:`number_value <api_field_google.protobuf.Value.number_value>`, :ref:`string_value <api_field_google.protobuf.Value.string_value>`, :ref:`bool_value <api_field_google.protobuf.Value.bool_value>`, :ref:`struct_value <api_field_google.protobuf.Value.struct_value>`, :ref:`list_value <api_field_google.protobuf.Value.list_value>` may be set.
  
.. _api_field_google.protobuf.Value.struct_value:

struct_value
  (:ref:`google.protobuf.Struct <api_msg_google.protobuf.Struct>`) Represents a structured value.
  
  The kind of value.
  
  
  Only one of :ref:`null_value <api_field_google.protobuf.Value.null_value>`, :ref:`number_value <api_field_google.protobuf.Value.number_value>`, :ref:`string_value <api_field_google.protobuf.Value.string_value>`, :ref:`bool_value <api_field_google.protobuf.Value.bool_value>`, :ref:`struct_value <api_field_google.protobuf.Value.struct_value>`, :ref:`list_value <api_field_google.protobuf.Value.list_value>` may be set.
  
.. _api_field_google.protobuf.Value.list_value:

list_value
  (:ref:`google.protobuf.ListValue <api_msg_google.protobuf.ListValue>`) Represents a repeated `Value`.
  
  The kind of value.
  
  
  Only one of :ref:`null_value <api_field_google.protobuf.Value.null_value>`, :ref:`number_value <api_field_google.protobuf.Value.number_value>`, :ref:`string_value <api_field_google.protobuf.Value.string_value>`, :ref:`bool_value <api_field_google.protobuf.Value.bool_value>`, :ref:`struct_value <api_field_google.protobuf.Value.struct_value>`, :ref:`list_value <api_field_google.protobuf.Value.list_value>` may be set.
  


.. _api_msg_google.protobuf.ListValue:

google.protobuf.ListValue
-------------------------

`[google.protobuf.ListValue proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/struct.proto#L92>`_

`ListValue` is a wrapper around a repeated field of values.

The JSON representation for `ListValue` is JSON array.

.. code-block:: json

  {
    "values": []
  }

.. _api_field_google.protobuf.ListValue.values:

values
  (:ref:`google.protobuf.Value <api_msg_google.protobuf.Value>`) Repeated field of dynamically typed values.
  
  

.. _api_enum_google.protobuf.NullValue:

Enum google.protobuf.NullValue
------------------------------

`[google.protobuf.NullValue proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/struct.proto#L84>`_

`NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

 The JSON representation for `NullValue` is JSON `null`.

.. _api_enum_value_google.protobuf.NullValue.NULL_VALUE:

NULL_VALUE
  *(DEFAULT)* ⁣Null value.
  
  
