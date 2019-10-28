.. _api_file_google/protobuf/descriptor.proto:

descriptor.proto
================================

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
Author: kenton@google.com (Kenton Varda)
 Based on original Protocol Buffers design by
 Sanjay Ghemawat, Jeff Dean, and others.

The messages in this file describe the definitions found in .proto files.
A valid .proto file can be translated directly to a FileDescriptorProto
without any other information (e.g. without reading its imports).

.. _api_msg_google.protobuf.FileDescriptorSet:

google.protobuf.FileDescriptorSet
---------------------------------

`[google.protobuf.FileDescriptorSet proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L55>`_

The protocol compiler can output a FileDescriptorSet containing the .proto
files it parses.

.. code-block:: json

  {
    "file": []
  }

.. _api_field_google.protobuf.FileDescriptorSet.file:

file
  (:ref:`google.protobuf.FileDescriptorProto <api_msg_google.protobuf.FileDescriptorProto>`) 
  


.. _api_msg_google.protobuf.FileDescriptorProto:

google.protobuf.FileDescriptorProto
-----------------------------------

`[google.protobuf.FileDescriptorProto proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L60>`_

Describes a complete .proto file.

.. code-block:: json

  {
    "name": "...",
    "package": "...",
    "dependency": [],
    "public_dependency": [],
    "weak_dependency": [],
    "message_type": [],
    "enum_type": [],
    "service": [],
    "extension": [],
    "options": "{...}",
    "source_code_info": "{...}",
    "syntax": "..."
  }

.. _api_field_google.protobuf.FileDescriptorProto.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.FileDescriptorProto.package:

package
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.FileDescriptorProto.dependency:

dependency
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Names of files imported by this file.
  
  
.. _api_field_google.protobuf.FileDescriptorProto.public_dependency:

public_dependency
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indexes of the public imported files in the dependency list above.
  
  
.. _api_field_google.protobuf.FileDescriptorProto.weak_dependency:

weak_dependency
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indexes of the weak imported files in the dependency list.
  For Google-internal migration only. Do not use.
  
  
.. _api_field_google.protobuf.FileDescriptorProto.message_type:

message_type
  (:ref:`google.protobuf.DescriptorProto <api_msg_google.protobuf.DescriptorProto>`) All top-level definitions in this file.
  
  
.. _api_field_google.protobuf.FileDescriptorProto.enum_type:

enum_type
  (:ref:`google.protobuf.EnumDescriptorProto <api_msg_google.protobuf.EnumDescriptorProto>`) 
  
.. _api_field_google.protobuf.FileDescriptorProto.service:

service
  (:ref:`google.protobuf.ServiceDescriptorProto <api_msg_google.protobuf.ServiceDescriptorProto>`) 
  
.. _api_field_google.protobuf.FileDescriptorProto.extension:

extension
  (:ref:`google.protobuf.FieldDescriptorProto <api_msg_google.protobuf.FieldDescriptorProto>`) 
  
.. _api_field_google.protobuf.FileDescriptorProto.options:

options
  (:ref:`google.protobuf.FileOptions <api_msg_google.protobuf.FileOptions>`) 
  
.. _api_field_google.protobuf.FileDescriptorProto.source_code_info:

source_code_info
  (:ref:`google.protobuf.SourceCodeInfo <api_msg_google.protobuf.SourceCodeInfo>`) This field contains optional information about the original source code.
  You may safely remove this entire field without harming runtime
  functionality of the descriptors -- the information is needed only by
  development tools.
  
  
.. _api_field_google.protobuf.FileDescriptorProto.syntax:

syntax
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The syntax of the proto file.
  The supported values are "proto2" and "proto3".
  
  


.. _api_msg_google.protobuf.DescriptorProto:

google.protobuf.DescriptorProto
-------------------------------

`[google.protobuf.DescriptorProto proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L92>`_

Describes a message type.

.. code-block:: json

  {
    "name": "...",
    "field": [],
    "extension": [],
    "nested_type": [],
    "enum_type": [],
    "extension_range": [],
    "oneof_decl": [],
    "options": "{...}",
    "reserved_range": [],
    "reserved_name": []
  }

.. _api_field_google.protobuf.DescriptorProto.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.DescriptorProto.field:

field
  (:ref:`google.protobuf.FieldDescriptorProto <api_msg_google.protobuf.FieldDescriptorProto>`) 
  
.. _api_field_google.protobuf.DescriptorProto.extension:

extension
  (:ref:`google.protobuf.FieldDescriptorProto <api_msg_google.protobuf.FieldDescriptorProto>`) 
  
.. _api_field_google.protobuf.DescriptorProto.nested_type:

nested_type
  (:ref:`google.protobuf.DescriptorProto <api_msg_google.protobuf.DescriptorProto>`) 
  
.. _api_field_google.protobuf.DescriptorProto.enum_type:

enum_type
  (:ref:`google.protobuf.EnumDescriptorProto <api_msg_google.protobuf.EnumDescriptorProto>`) 
  
.. _api_field_google.protobuf.DescriptorProto.extension_range:

extension_range
  (:ref:`google.protobuf.DescriptorProto.ExtensionRange <api_msg_google.protobuf.DescriptorProto.ExtensionRange>`) 
  
.. _api_field_google.protobuf.DescriptorProto.oneof_decl:

oneof_decl
  (:ref:`google.protobuf.OneofDescriptorProto <api_msg_google.protobuf.OneofDescriptorProto>`) 
  
.. _api_field_google.protobuf.DescriptorProto.options:

options
  (:ref:`google.protobuf.MessageOptions <api_msg_google.protobuf.MessageOptions>`) 
  
.. _api_field_google.protobuf.DescriptorProto.reserved_range:

reserved_range
  (:ref:`google.protobuf.DescriptorProto.ReservedRange <api_msg_google.protobuf.DescriptorProto.ReservedRange>`) 
  
.. _api_field_google.protobuf.DescriptorProto.reserved_name:

reserved_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Reserved field names, which may not be used by fields in the same message.
  A given name may only be reserved once.
  
  
.. _api_msg_google.protobuf.DescriptorProto.ExtensionRange:

google.protobuf.DescriptorProto.ExtensionRange
----------------------------------------------

`[google.protobuf.DescriptorProto.ExtensionRange proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L101>`_


.. code-block:: json

  {
    "start": "...",
    "end": "...",
    "options": "{...}"
  }

.. _api_field_google.protobuf.DescriptorProto.ExtensionRange.start:

start
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.DescriptorProto.ExtensionRange.end:

end
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.DescriptorProto.ExtensionRange.options:

options
  (:ref:`google.protobuf.ExtensionRangeOptions <api_msg_google.protobuf.ExtensionRangeOptions>`) 
  


.. _api_msg_google.protobuf.DescriptorProto.ReservedRange:

google.protobuf.DescriptorProto.ReservedRange
---------------------------------------------

`[google.protobuf.DescriptorProto.ReservedRange proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L116>`_

Range of reserved tag numbers. Reserved tag numbers may not be used by
fields or extension ranges in the same message. Reserved ranges may
not overlap.

.. code-block:: json

  {
    "start": "...",
    "end": "..."
  }

.. _api_field_google.protobuf.DescriptorProto.ReservedRange.start:

start
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.DescriptorProto.ReservedRange.end:

end
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  



.. _api_msg_google.protobuf.ExtensionRangeOptions:

google.protobuf.ExtensionRangeOptions
-------------------------------------

`[google.protobuf.ExtensionRangeOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L126>`_


.. code-block:: json

  {
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.ExtensionRangeOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here. See above.
  
  


.. _api_msg_google.protobuf.FieldDescriptorProto:

google.protobuf.FieldDescriptorProto
------------------------------------

`[google.protobuf.FieldDescriptorProto proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L135>`_

Describes a field within a message.

.. code-block:: json

  {
    "name": "...",
    "number": "...",
    "label": "...",
    "type": "...",
    "type_name": "...",
    "extendee": "...",
    "default_value": "...",
    "oneof_index": "...",
    "json_name": "...",
    "options": "{...}"
  }

.. _api_field_google.protobuf.FieldDescriptorProto.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.FieldDescriptorProto.number:

number
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.FieldDescriptorProto.label:

label
  (:ref:`google.protobuf.FieldDescriptorProto.Label <api_enum_google.protobuf.FieldDescriptorProto.Label>`) 
  
.. _api_field_google.protobuf.FieldDescriptorProto.type:

type
  (:ref:`google.protobuf.FieldDescriptorProto.Type <api_enum_google.protobuf.FieldDescriptorProto.Type>`) If type_name is set, this need not be set.  If both this and type_name
  are set, this must be one of TYPE_ENUM, TYPE_MESSAGE or TYPE_GROUP.
  
  
.. _api_field_google.protobuf.FieldDescriptorProto.type_name:

type_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) For message and enum types, this is the name of the type.  If the name
  starts with a '.', it is fully-qualified.  Otherwise, C++-like scoping
  rules are used to find the type (i.e. first the nested types within this
  message are searched, then within the parent, on up to the root
  namespace).
  
  
.. _api_field_google.protobuf.FieldDescriptorProto.extendee:

extendee
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) For extensions, this is the name of the type being extended.  It is
  resolved in the same manner as type_name.
  
  
.. _api_field_google.protobuf.FieldDescriptorProto.default_value:

default_value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) For numeric types, contains the original text representation of the value.
  For booleans, "true" or "false".
  For strings, contains the default text contents (not escaped in any way).
  For bytes, contains the C escaped value.  All bytes >= 128 are escaped.
  TODO(kenton):  Base-64 encode?
  
  
.. _api_field_google.protobuf.FieldDescriptorProto.oneof_index:

oneof_index
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) If set, gives the index of a oneof in the containing type's oneof_decl
  list.  This field is a member of that oneof.
  
  
.. _api_field_google.protobuf.FieldDescriptorProto.json_name:

json_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) JSON name of this field. The value is set by protocol compiler. If the
  user has set a "json_name" option on this field, that option's value
  will be used. Otherwise, it's deduced from the field's name by converting
  it to camelCase.
  
  
.. _api_field_google.protobuf.FieldDescriptorProto.options:

options
  (:ref:`google.protobuf.FieldOptions <api_msg_google.protobuf.FieldOptions>`) 
  

.. _api_enum_google.protobuf.FieldDescriptorProto.Type:

Enum google.protobuf.FieldDescriptorProto.Type
----------------------------------------------

`[google.protobuf.FieldDescriptorProto.Type proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L136>`_


.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_DOUBLE:

TYPE_DOUBLE
  ⁣0 is reserved for errors.
  Order is weird for historical reasons.
  
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_FLOAT:

TYPE_FLOAT
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_INT64:

TYPE_INT64
  ⁣Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT64 if
  negative values are likely.
  
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_UINT64:

TYPE_UINT64
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_INT32:

TYPE_INT32
  ⁣Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT32 if
  negative values are likely.
  
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_FIXED64:

TYPE_FIXED64
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_FIXED32:

TYPE_FIXED32
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_BOOL:

TYPE_BOOL
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_STRING:

TYPE_STRING
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_GROUP:

TYPE_GROUP
  ⁣Tag-delimited aggregate.
  Group type is deprecated and not supported in proto3. However, Proto3
  implementations should still be able to parse the group wire format and
  treat group fields as unknown fields.
  
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_MESSAGE:

TYPE_MESSAGE
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_BYTES:

TYPE_BYTES
  ⁣New in version 2.
  
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_UINT32:

TYPE_UINT32
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_ENUM:

TYPE_ENUM
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_SFIXED32:

TYPE_SFIXED32
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_SFIXED64:

TYPE_SFIXED64
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_SINT32:

TYPE_SINT32
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Type.TYPE_SINT64:

TYPE_SINT64
  ⁣
  

.. _api_enum_google.protobuf.FieldDescriptorProto.Label:

Enum google.protobuf.FieldDescriptorProto.Label
-----------------------------------------------

`[google.protobuf.FieldDescriptorProto.Label proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L169>`_


.. _api_enum_value_google.protobuf.FieldDescriptorProto.Label.LABEL_OPTIONAL:

LABEL_OPTIONAL
  ⁣0 is reserved for errors
  
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Label.LABEL_REQUIRED:

LABEL_REQUIRED
  ⁣
  
.. _api_enum_value_google.protobuf.FieldDescriptorProto.Label.LABEL_REPEATED:

LABEL_REPEATED
  ⁣
  

.. _api_msg_google.protobuf.OneofDescriptorProto:

google.protobuf.OneofDescriptorProto
------------------------------------

`[google.protobuf.OneofDescriptorProto proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L216>`_

Describes a oneof.

.. code-block:: json

  {
    "name": "...",
    "options": "{...}"
  }

.. _api_field_google.protobuf.OneofDescriptorProto.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.OneofDescriptorProto.options:

options
  (:ref:`google.protobuf.OneofOptions <api_msg_google.protobuf.OneofOptions>`) 
  


.. _api_msg_google.protobuf.EnumDescriptorProto:

google.protobuf.EnumDescriptorProto
-----------------------------------

`[google.protobuf.EnumDescriptorProto proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L222>`_

Describes an enum type.

.. code-block:: json

  {
    "name": "...",
    "value": [],
    "options": "{...}",
    "reserved_range": [],
    "reserved_name": []
  }

.. _api_field_google.protobuf.EnumDescriptorProto.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.EnumDescriptorProto.value:

value
  (:ref:`google.protobuf.EnumValueDescriptorProto <api_msg_google.protobuf.EnumValueDescriptorProto>`) 
  
.. _api_field_google.protobuf.EnumDescriptorProto.options:

options
  (:ref:`google.protobuf.EnumOptions <api_msg_google.protobuf.EnumOptions>`) 
  
.. _api_field_google.protobuf.EnumDescriptorProto.reserved_range:

reserved_range
  (:ref:`google.protobuf.EnumDescriptorProto.EnumReservedRange <api_msg_google.protobuf.EnumDescriptorProto.EnumReservedRange>`) Range of reserved numeric values. Reserved numeric values may not be used
  by enum values in the same enum declaration. Reserved ranges may not
  overlap.
  
  
.. _api_field_google.protobuf.EnumDescriptorProto.reserved_name:

reserved_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Reserved enum value names, which may not be reused. A given name may only
  be reserved once.
  
  
.. _api_msg_google.protobuf.EnumDescriptorProto.EnumReservedRange:

google.protobuf.EnumDescriptorProto.EnumReservedRange
-----------------------------------------------------

`[google.protobuf.EnumDescriptorProto.EnumReservedRange proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L235>`_

Range of reserved numeric values. Reserved values may not be used by
entries in the same enum. Reserved ranges may not overlap.

Note that this is distinct from DescriptorProto.ReservedRange in that it
is inclusive such that it can appropriately represent the entire int32
domain.

.. code-block:: json

  {
    "start": "...",
    "end": "..."
  }

.. _api_field_google.protobuf.EnumDescriptorProto.EnumReservedRange.start:

start
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.EnumDescriptorProto.EnumReservedRange.end:

end
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  



.. _api_msg_google.protobuf.EnumValueDescriptorProto:

google.protobuf.EnumValueDescriptorProto
----------------------------------------

`[google.protobuf.EnumValueDescriptorProto proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L251>`_

Describes a value within an enum.

.. code-block:: json

  {
    "name": "...",
    "number": "...",
    "options": "{...}"
  }

.. _api_field_google.protobuf.EnumValueDescriptorProto.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.EnumValueDescriptorProto.number:

number
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.EnumValueDescriptorProto.options:

options
  (:ref:`google.protobuf.EnumValueOptions <api_msg_google.protobuf.EnumValueOptions>`) 
  


.. _api_msg_google.protobuf.ServiceDescriptorProto:

google.protobuf.ServiceDescriptorProto
--------------------------------------

`[google.protobuf.ServiceDescriptorProto proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L259>`_

Describes a service.

.. code-block:: json

  {
    "name": "...",
    "method": [],
    "options": "{...}"
  }

.. _api_field_google.protobuf.ServiceDescriptorProto.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.ServiceDescriptorProto.method:

method
  (:ref:`google.protobuf.MethodDescriptorProto <api_msg_google.protobuf.MethodDescriptorProto>`) 
  
.. _api_field_google.protobuf.ServiceDescriptorProto.options:

options
  (:ref:`google.protobuf.ServiceOptions <api_msg_google.protobuf.ServiceOptions>`) 
  


.. _api_msg_google.protobuf.MethodDescriptorProto:

google.protobuf.MethodDescriptorProto
-------------------------------------

`[google.protobuf.MethodDescriptorProto proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L267>`_

Describes a method of a service.

.. code-block:: json

  {
    "name": "...",
    "input_type": "...",
    "output_type": "...",
    "options": "{...}",
    "client_streaming": "...",
    "server_streaming": "..."
  }

.. _api_field_google.protobuf.MethodDescriptorProto.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.MethodDescriptorProto.input_type:

input_type
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Input and output type names.  These are resolved in the same way as
  FieldDescriptorProto.type_name, but must refer to a message type.
  
  
.. _api_field_google.protobuf.MethodDescriptorProto.output_type:

output_type
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.MethodDescriptorProto.options:

options
  (:ref:`google.protobuf.MethodOptions <api_msg_google.protobuf.MethodOptions>`) 
  
.. _api_field_google.protobuf.MethodDescriptorProto.client_streaming:

client_streaming
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifies if client streams multiple client messages
  
  
.. _api_field_google.protobuf.MethodDescriptorProto.server_streaming:

server_streaming
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifies if server streams multiple server messages
  
  


.. _api_msg_google.protobuf.FileOptions:

google.protobuf.FileOptions
---------------------------

`[google.protobuf.FileOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L317>`_


.. code-block:: json

  {
    "java_package": "...",
    "java_outer_classname": "...",
    "java_multiple_files": "...",
    "java_generate_equals_and_hash": "...",
    "java_string_check_utf8": "...",
    "optimize_for": "...",
    "go_package": "...",
    "cc_generic_services": "...",
    "java_generic_services": "...",
    "py_generic_services": "...",
    "php_generic_services": "...",
    "deprecated": "...",
    "cc_enable_arenas": "...",
    "objc_class_prefix": "...",
    "csharp_namespace": "...",
    "swift_prefix": "...",
    "php_class_prefix": "...",
    "php_namespace": "...",
    "php_metadata_namespace": "...",
    "ruby_package": "...",
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.FileOptions.java_package:

java_package
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Sets the Java package where classes generated from this .proto will be
  placed.  By default, the proto package is used, but this is often
  inappropriate because proto packages do not normally start with backwards
  domain names.
  
  
.. _api_field_google.protobuf.FileOptions.java_outer_classname:

java_outer_classname
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) If set, all the classes from the .proto file are wrapped in a single
  outer class with the given name.  This applies to both Proto1
  (equivalent to the old "--one_java_file" option) and Proto2 (where
  a .proto always translates to a single class, but you may want to
  explicitly choose the class name).
  
  
.. _api_field_google.protobuf.FileOptions.java_multiple_files:

java_multiple_files
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) If set true, then the Java code generator will generate a separate .java
  file for each top-level message, enum, and service defined in the .proto
  file.  Thus, these types will *not* be nested inside the outer class
  named by java_outer_classname.  However, the outer class will still be
  generated to contain the file's getDescriptor() method as well as any
  top-level extensions defined in the file.
  
  
.. _api_field_google.protobuf.FileOptions.java_generate_equals_and_hash:

java_generate_equals_and_hash
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) This option does nothing.
  
  
.. _api_field_google.protobuf.FileOptions.java_string_check_utf8:

java_string_check_utf8
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) If set true, then the Java2 code generator will generate code that
  throws an exception whenever an attempt is made to assign a non-UTF-8
  byte sequence to a string field.
  Message reflection will do the same.
  However, an extension field still accepts non-UTF-8 byte sequences.
  This option has no effect on when used with the lite runtime.
  
  
.. _api_field_google.protobuf.FileOptions.optimize_for:

optimize_for
  (:ref:`google.protobuf.FileOptions.OptimizeMode <api_enum_google.protobuf.FileOptions.OptimizeMode>`) 
  
.. _api_field_google.protobuf.FileOptions.go_package:

go_package
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Sets the Go package where structs generated from this .proto will be
  placed. If omitted, the Go package will be derived from the following:
    - The basename of the package import path, if provided.
    - Otherwise, the package statement in the .proto file, if present.
    - Otherwise, the basename of the .proto file, without extension.
  
  
.. _api_field_google.protobuf.FileOptions.cc_generic_services:

cc_generic_services
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Should generic services be generated in each language?  "Generic" services
  are not specific to any particular RPC system.  They are generated by the
  main code generators in each language (without additional plugins).
  Generic services were the only kind of service generation supported by
  early versions of google.protobuf.
  
  Generic services are now considered deprecated in favor of using plugins
  that generate code specific to your particular RPC system.  Therefore,
  these default to false.  Old code which depends on generic services should
  explicitly set them to true.
  
  
.. _api_field_google.protobuf.FileOptions.java_generic_services:

java_generic_services
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.FileOptions.py_generic_services:

py_generic_services
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.FileOptions.php_generic_services:

php_generic_services
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.FileOptions.deprecated:

deprecated
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Is this file deprecated?
  Depending on the target platform, this can emit Deprecated annotations
  for everything in the file, or it will be completely ignored; in the very
  least, this is a formalization for deprecating files.
  
  
.. _api_field_google.protobuf.FileOptions.cc_enable_arenas:

cc_enable_arenas
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Enables the use of arenas for the proto messages in this file. This applies
  only to generated classes for C++.
  
  
.. _api_field_google.protobuf.FileOptions.objc_class_prefix:

objc_class_prefix
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Sets the objective c class prefix which is prepended to all objective c
  generated classes from this .proto. There is no default.
  
  
.. _api_field_google.protobuf.FileOptions.csharp_namespace:

csharp_namespace
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Namespace for generated classes; defaults to the package.
  
  
.. _api_field_google.protobuf.FileOptions.swift_prefix:

swift_prefix
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) By default Swift generators will take the proto package and CamelCase it
  replacing '.' with underscore and use that to prefix the types/symbols
  defined. When this options is provided, they will use this value instead
  to prefix the types/symbols defined.
  
  
.. _api_field_google.protobuf.FileOptions.php_class_prefix:

php_class_prefix
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Sets the php class prefix which is prepended to all php generated classes
  from this .proto. Default is empty.
  
  
.. _api_field_google.protobuf.FileOptions.php_namespace:

php_namespace
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Use this option to change the namespace of php generated classes. Default
  is empty. When this option is empty, the package name will be used for
  determining the namespace.
  
  
.. _api_field_google.protobuf.FileOptions.php_metadata_namespace:

php_metadata_namespace
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Use this option to change the namespace of php generated metadata classes.
  Default is empty. When this option is empty, the proto file name will be used
  for determining the namespace.
  
  
.. _api_field_google.protobuf.FileOptions.ruby_package:

ruby_package
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Use this option to change the package of ruby generated classes. Default
  is empty. When this option is not set, the package name will be used for
  determining the ruby package.
  
  
.. _api_field_google.protobuf.FileOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here.
  See the documentation for the "Options" section above.
  
  

.. _api_enum_google.protobuf.FileOptions.OptimizeMode:

Enum google.protobuf.FileOptions.OptimizeMode
---------------------------------------------

`[google.protobuf.FileOptions.OptimizeMode proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L354>`_

Generated classes can be optimized for speed or code size.

.. _api_enum_value_google.protobuf.FileOptions.OptimizeMode.SPEED:

SPEED
  ⁣
  
.. _api_enum_value_google.protobuf.FileOptions.OptimizeMode.CODE_SIZE:

CODE_SIZE
  ⁣etc.
  
  
.. _api_enum_value_google.protobuf.FileOptions.OptimizeMode.LITE_RUNTIME:

LITE_RUNTIME
  ⁣
  

.. _api_msg_google.protobuf.MessageOptions:

google.protobuf.MessageOptions
------------------------------

`[google.protobuf.MessageOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L441>`_


.. code-block:: json

  {
    "message_set_wire_format": "...",
    "no_standard_descriptor_accessor": "...",
    "deprecated": "...",
    "map_entry": "...",
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.MessageOptions.message_set_wire_format:

message_set_wire_format
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Set true to use the old proto1 MessageSet wire format for extensions.
  This is provided for backwards-compatibility with the MessageSet wire
  format.  You should not use this for any other reason:  It's less
  efficient, has fewer features, and is more complicated.
  
  The message must be defined exactly as follows:
    message Foo {
      option message_set_wire_format = true;
      extensions 4 to max;
    }
  Note that the message cannot have any defined fields; MessageSets only
  have extensions.
  
  All extensions of your type must be singular messages; e.g. they cannot
  be int32s, enums, or repeated messages.
  
  Because this is an option, the above two restrictions are not enforced by
  the protocol compiler.
  
  
.. _api_field_google.protobuf.MessageOptions.no_standard_descriptor_accessor:

no_standard_descriptor_accessor
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Disables the generation of the standard "descriptor()" accessor, which can
  conflict with a field of the same name.  This is meant to make migration
  from proto1 easier; new code should avoid fields named "descriptor".
  
  
.. _api_field_google.protobuf.MessageOptions.deprecated:

deprecated
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Is this message deprecated?
  Depending on the target platform, this can emit Deprecated annotations
  for the message, or it will be completely ignored; in the very least,
  this is a formalization for deprecating messages.
  
  
.. _api_field_google.protobuf.MessageOptions.map_entry:

map_entry
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Whether the message is an automatically generated map entry type for the
  maps field.
  
  For maps fields:
      map<KeyType, ValueType> map_field = 1;
  The parsed descriptor looks like:
      message MapFieldEntry {
          option map_entry = true;
          optional KeyType key = 1;
          optional ValueType value = 2;
      }
      repeated MapFieldEntry map_field = 1;
  
  Implementations may choose not to generate the map_entry=true message, but
  use a native map in the target language to hold the keys and values.
  The reflection APIs in such implementions still need to work as
  if the field is a repeated message field.
  
  NOTE: Do not set the option in .proto files. Always use the maps syntax
  instead. The option should only be implicitly set by the proto compiler
  parser.
  
  
.. _api_field_google.protobuf.MessageOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here. See above.
  
  


.. _api_msg_google.protobuf.FieldOptions:

google.protobuf.FieldOptions
----------------------------

`[google.protobuf.FieldOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L506>`_


.. code-block:: json

  {
    "ctype": "...",
    "packed": "...",
    "jstype": "...",
    "lazy": "...",
    "deprecated": "...",
    "weak": "...",
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.FieldOptions.ctype:

ctype
  (:ref:`google.protobuf.FieldOptions.CType <api_enum_google.protobuf.FieldOptions.CType>`) The ctype option instructs the C++ code generator to use a different
  representation of the field than it normally would.  See the specific
  options below.  This option is not yet implemented in the open source
  release -- sorry, we'll try to include it in a future version!
  
  
.. _api_field_google.protobuf.FieldOptions.packed:

packed
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The packed option can be enabled for repeated primitive fields to enable
  a more efficient representation on the wire. Rather than repeatedly
  writing the tag and type for each element, the entire array is encoded as
  a single length-delimited blob. In proto3, only explicit setting it to
  false will avoid using packed encoding.
  
  
.. _api_field_google.protobuf.FieldOptions.jstype:

jstype
  (:ref:`google.protobuf.FieldOptions.JSType <api_enum_google.protobuf.FieldOptions.JSType>`) The jstype option determines the JavaScript type used for values of the
  field.  The option is permitted only for 64 bit integral and fixed types
  (int64, uint64, sint64, fixed64, sfixed64).  A field with jstype JS_STRING
  is represented as JavaScript string, which avoids loss of precision that
  can happen when a large value is converted to a floating point JavaScript.
  Specifying JS_NUMBER for the jstype causes the generated JavaScript code to
  use the JavaScript "number" type.  The behavior of the default option
  JS_NORMAL is implementation dependent.
  
  This option is an enum to permit additional types to be added, e.g.
  goog.math.Integer.
  
  
.. _api_field_google.protobuf.FieldOptions.lazy:

lazy
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Should this field be parsed lazily?  Lazy applies only to message-type
  fields.  It means that when the outer message is initially parsed, the
  inner message's contents will not be parsed but instead stored in encoded
  form.  The inner message will actually be parsed when it is first accessed.
  
  This is only a hint.  Implementations are free to choose whether to use
  eager or lazy parsing regardless of the value of this option.  However,
  setting this option true suggests that the protocol author believes that
  using lazy parsing on this field is worth the additional bookkeeping
  overhead typically needed to implement it.
  
  This option does not affect the public interface of any generated code;
  all method signatures remain the same.  Furthermore, thread-safety of the
  interface is not affected by this option; const methods remain safe to
  call from multiple threads concurrently, while non-const methods continue
  to require exclusive access.
  
  
  Note that implementations may choose not to check required fields within
  a lazy sub-message.  That is, calling IsInitialized() on the outer message
  may return true even if the inner message has missing required fields.
  This is necessary because otherwise the inner message would have to be
  parsed in order to perform the check, defeating the purpose of lazy
  parsing.  An implementation which chooses not to check required fields
  must be consistent about it.  That is, for any particular sub-message, the
  implementation must either *always* check its required fields, or *never*
  check its required fields, regardless of whether or not the message has
  been parsed.
  
  
.. _api_field_google.protobuf.FieldOptions.deprecated:

deprecated
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Is this field deprecated?
  Depending on the target platform, this can emit Deprecated annotations
  for accessors, or it will be completely ignored; in the very least, this
  is a formalization for deprecating fields.
  
  
.. _api_field_google.protobuf.FieldOptions.weak:

weak
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) For Google-internal migration only. Do not use.
  
  
.. _api_field_google.protobuf.FieldOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here. See above.
  
  

.. _api_enum_google.protobuf.FieldOptions.CType:

Enum google.protobuf.FieldOptions.CType
---------------------------------------

`[google.protobuf.FieldOptions.CType proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L512>`_


.. _api_enum_value_google.protobuf.FieldOptions.CType.STRING:

STRING
  *(DEFAULT)* ⁣Default mode.
  
  
.. _api_enum_value_google.protobuf.FieldOptions.CType.CORD:

CORD
  ⁣
  
.. _api_enum_value_google.protobuf.FieldOptions.CType.STRING_PIECE:

STRING_PIECE
  ⁣
  

.. _api_enum_google.protobuf.FieldOptions.JSType:

Enum google.protobuf.FieldOptions.JSType
----------------------------------------

`[google.protobuf.FieldOptions.JSType proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L539>`_


.. _api_enum_value_google.protobuf.FieldOptions.JSType.JS_NORMAL:

JS_NORMAL
  *(DEFAULT)* ⁣Use the default type.
  
  
.. _api_enum_value_google.protobuf.FieldOptions.JSType.JS_STRING:

JS_STRING
  ⁣Use JavaScript strings.
  
  
.. _api_enum_value_google.protobuf.FieldOptions.JSType.JS_NUMBER:

JS_NUMBER
  ⁣Use JavaScript numbers.
  
  

.. _api_msg_google.protobuf.OneofOptions:

google.protobuf.OneofOptions
----------------------------

`[google.protobuf.OneofOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L599>`_


.. code-block:: json

  {
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.OneofOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here. See above.
  
  


.. _api_msg_google.protobuf.EnumOptions:

google.protobuf.EnumOptions
---------------------------

`[google.protobuf.EnumOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L607>`_


.. code-block:: json

  {
    "allow_alias": "...",
    "deprecated": "...",
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.EnumOptions.allow_alias:

allow_alias
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Set this option to true to allow mapping different tag names to the same
  value.
  
  
.. _api_field_google.protobuf.EnumOptions.deprecated:

deprecated
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Is this enum deprecated?
  Depending on the target platform, this can emit Deprecated annotations
  for the enum, or it will be completely ignored; in the very least, this
  is a formalization for deprecating enums.
  
  
.. _api_field_google.protobuf.EnumOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here. See above.
  
  


.. _api_msg_google.protobuf.EnumValueOptions:

google.protobuf.EnumValueOptions
--------------------------------

`[google.protobuf.EnumValueOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L628>`_


.. code-block:: json

  {
    "deprecated": "...",
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.EnumValueOptions.deprecated:

deprecated
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Is this enum value deprecated?
  Depending on the target platform, this can emit Deprecated annotations
  for the enum value, or it will be completely ignored; in the very least,
  this is a formalization for deprecating enum values.
  
  
.. _api_field_google.protobuf.EnumValueOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here. See above.
  
  


.. _api_msg_google.protobuf.ServiceOptions:

google.protobuf.ServiceOptions
------------------------------

`[google.protobuf.ServiceOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L642>`_


.. code-block:: json

  {
    "deprecated": "...",
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.ServiceOptions.deprecated:

deprecated
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Is this service deprecated?
  Depending on the target platform, this can emit Deprecated annotations
  for the service, or it will be completely ignored; in the very least,
  this is a formalization for deprecating services.
  
  
.. _api_field_google.protobuf.ServiceOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here. See above.
  
  


.. _api_msg_google.protobuf.MethodOptions:

google.protobuf.MethodOptions
-----------------------------

`[google.protobuf.MethodOptions proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L662>`_


.. code-block:: json

  {
    "deprecated": "...",
    "idempotency_level": "...",
    "uninterpreted_option": []
  }

.. _api_field_google.protobuf.MethodOptions.deprecated:

deprecated
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Is this method deprecated?
  Depending on the target platform, this can emit Deprecated annotations
  for the method, or it will be completely ignored; in the very least,
  this is a formalization for deprecating methods.
  
  
.. _api_field_google.protobuf.MethodOptions.idempotency_level:

idempotency_level
  (:ref:`google.protobuf.MethodOptions.IdempotencyLevel <api_enum_google.protobuf.MethodOptions.IdempotencyLevel>`) 
  
.. _api_field_google.protobuf.MethodOptions.uninterpreted_option:

uninterpreted_option
  (:ref:`google.protobuf.UninterpretedOption <api_msg_google.protobuf.UninterpretedOption>`) The parser stores options it doesn't recognize here. See above.
  
  

.. _api_enum_google.protobuf.MethodOptions.IdempotencyLevel:

Enum google.protobuf.MethodOptions.IdempotencyLevel
---------------------------------------------------

`[google.protobuf.MethodOptions.IdempotencyLevel proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L678>`_

Is this method side-effect-free (or safe in HTTP parlance), or idempotent,
or neither? HTTP based RPC implementation may choose GET verb for safe
methods, and PUT verb for idempotent methods instead of the default POST.

.. _api_enum_value_google.protobuf.MethodOptions.IdempotencyLevel.IDEMPOTENCY_UNKNOWN:

IDEMPOTENCY_UNKNOWN
  *(DEFAULT)* ⁣
  
.. _api_enum_value_google.protobuf.MethodOptions.IdempotencyLevel.NO_SIDE_EFFECTS:

NO_SIDE_EFFECTS
  ⁣
  
.. _api_enum_value_google.protobuf.MethodOptions.IdempotencyLevel.IDEMPOTENT:

IDEMPOTENT
  ⁣
  

.. _api_msg_google.protobuf.UninterpretedOption:

google.protobuf.UninterpretedOption
-----------------------------------

`[google.protobuf.UninterpretedOption proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L700>`_

A message representing a option the parser does not recognize. This only
appears in options protos created by the compiler::Parser class.
DescriptorPool resolves these when building Descriptor objects. Therefore,
options protos in descriptor objects (e.g. returned by Descriptor::options(),
or produced by Descriptor::CopyTo()) will never have UninterpretedOptions
in them.

.. code-block:: json

  {
    "name": [],
    "identifier_value": "...",
    "positive_int_value": "...",
    "negative_int_value": "...",
    "double_value": "...",
    "string_value": "...",
    "aggregate_value": "..."
  }

.. _api_field_google.protobuf.UninterpretedOption.name:

name
  (:ref:`google.protobuf.UninterpretedOption.NamePart <api_msg_google.protobuf.UninterpretedOption.NamePart>`) 
  
.. _api_field_google.protobuf.UninterpretedOption.identifier_value:

identifier_value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The value of the uninterpreted option, in whatever type the tokenizer
  identified it as during parsing. Exactly one of these should be set.
  
  
.. _api_field_google.protobuf.UninterpretedOption.positive_int_value:

positive_int_value
  (`uint64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.UninterpretedOption.negative_int_value:

negative_int_value
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.UninterpretedOption.double_value:

double_value
  (`double <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.UninterpretedOption.string_value:

string_value
  (`bytes <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.UninterpretedOption.aggregate_value:

aggregate_value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_msg_google.protobuf.UninterpretedOption.NamePart:

google.protobuf.UninterpretedOption.NamePart
--------------------------------------------

`[google.protobuf.UninterpretedOption.NamePart proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L706>`_

The name of the uninterpreted option.  Each string represents a segment in
a dot-separated name.  is_extension is true iff a segment represents an
extension (denoted with parentheses in options specs in .proto files).
E.g.,{ ["foo", false], ["bar.baz", true], ["qux", false] } represents
"foo.(bar.baz).qux".

.. code-block:: json

  {
    "name_part": "...",
    "is_extension": "..."
  }

.. _api_field_google.protobuf.UninterpretedOption.NamePart.name_part:

name_part
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.UninterpretedOption.NamePart.is_extension:

is_extension
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  



.. _api_msg_google.protobuf.SourceCodeInfo:

google.protobuf.SourceCodeInfo
------------------------------

`[google.protobuf.SourceCodeInfo proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L727>`_

Encapsulates information about the original source file from which a
FileDescriptorProto was generated.

.. code-block:: json

  {
    "location": []
  }

.. _api_field_google.protobuf.SourceCodeInfo.location:

location
  (:ref:`google.protobuf.SourceCodeInfo.Location <api_msg_google.protobuf.SourceCodeInfo.Location>`) A Location identifies a piece of source code in a .proto file which
  corresponds to a particular definition.  This information is intended
  to be useful to IDEs, code indexers, documentation generators, and similar
  tools.
  
  For example, say we have a file like:
    message Foo {
      optional string foo = 1;
    }
  Let's look at just the field definition:
    optional string foo = 1;
    ^       ^^     ^^  ^  ^^^
    a       bc     de  f  ghi
  We have the following locations:
    span   path               represents
    [a,i)  [ 4, 0, 2, 0 ]     The whole field definition.
    [a,b)  [ 4, 0, 2, 0, 4 ]  The label (optional).
    [c,d)  [ 4, 0, 2, 0, 5 ]  The type (string).
    [e,f)  [ 4, 0, 2, 0, 1 ]  The name (foo).
    [g,h)  [ 4, 0, 2, 0, 3 ]  The number (1).
  
  Notes:
  - A location may refer to a repeated field itself (i.e. not to any
    particular index within it).  This is used whenever a set of elements are
    logically enclosed in a single code segment.  For example, an entire
    extend block (possibly containing multiple extension definitions) will
    have an outer location whose path refers to the "extensions" repeated
    field without an index.
  - Multiple locations may have the same path.  This happens when a single
    logical declaration is spread out across multiple places.  The most
    obvious example is the "extend" block again -- there may be multiple
    extend blocks in the same scope, each of which will have the same path.
  - A location's span is not always a subset of its parent's span.  For
    example, the "extendee" of an extension declaration appears at the
    beginning of the "extend" block and is shared by all extensions within
    the block.
  - Just because a location's span is a subset of some other location's span
    does not mean that it is a descendent.  For example, a "group" defines
    both a type and a field in a single declaration.  Thus, the locations
    corresponding to the type and field and their components will overlap.
  - Code which tries to interpret locations should probably be designed to
    ignore those that it doesn't understand, as more types of locations could
    be recorded in the future.
  
  
.. _api_msg_google.protobuf.SourceCodeInfo.Location:

google.protobuf.SourceCodeInfo.Location
---------------------------------------

`[google.protobuf.SourceCodeInfo.Location proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L772>`_


.. code-block:: json

  {
    "path": [],
    "span": [],
    "leading_comments": "...",
    "trailing_comments": "...",
    "leading_detached_comments": []
  }

.. _api_field_google.protobuf.SourceCodeInfo.Location.path:

path
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifies which part of the FileDescriptorProto was defined at this
  location.
  
  Each element is a field number or an index.  They form a path from
  the root FileDescriptorProto to the place where the definition.  For
  example, this path:
    [ 4, 3, 2, 7, 1 ]
  refers to:
    file.message_type(3)  // 4, 3
        .field(7)         // 2, 7
        .name()           // 1
  This is because FileDescriptorProto.message_type has field number 4:
    repeated DescriptorProto message_type = 4;
  and DescriptorProto.field has field number 2:
    repeated FieldDescriptorProto field = 2;
  and FieldDescriptorProto.name has field number 1:
    optional string name = 1;
  
  Thus, the above path gives the location of a field name.  If we removed
  the last element:
    [ 4, 3, 2, 7 ]
  this path refers to the whole field declaration (from the beginning
  of the label to the terminating semicolon).
  
  
.. _api_field_google.protobuf.SourceCodeInfo.Location.span:

span
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Always has exactly three or four elements: start line, start column,
  end line (optional, otherwise assumed same as start line), end column.
  These are packed into a single field for efficiency.  Note that line
  and column numbers are zero-based -- typically you will want to add
  1 to each before displaying to a user.
  
  
.. _api_field_google.protobuf.SourceCodeInfo.Location.leading_comments:

leading_comments
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) If this SourceCodeInfo represents a complete declaration, these are any
  comments appearing before and after the declaration which appear to be
  attached to the declaration.
  
  A series of line comments appearing on consecutive lines, with no other
  tokens appearing on those lines, will be treated as a single comment.
  
  leading_detached_comments will keep paragraphs of comments that appear
  before (but not connected to) the current element. Each paragraph,
  separated by empty lines, will be one comment element in the repeated
  field.
  
  Only the comment content is provided; comment markers (e.g. //) are
  stripped out.  For block comments, leading whitespace and an asterisk
  will be stripped from the beginning of each line other than the first.
  Newlines are included in the output.
  
  Examples:
  
    optional int32 foo = 1;  // Comment attached to foo.
    // Comment attached to bar.
    optional int32 bar = 2;
  
    optional string baz = 3;
    // Comment attached to baz.
    // Another line attached to baz.
  
    // Comment attached to qux.
    //
    // Another line attached to qux.
    optional double qux = 4;
  
    // Detached comment for corge. This is not leading or trailing comments
    // to qux or corge because there are blank lines separating it from
    // both.
  
    // Detached comment for corge paragraph 2.
  
    optional string corge = 5;
    /* Block comment attached
     * to corge.  Leading asterisks
     * will be removed. */
    /* Block comment attached to
     * grault. */
    optional int32 grault = 6;
  
    // ignored detached comments.
  
  
.. _api_field_google.protobuf.SourceCodeInfo.Location.trailing_comments:

trailing_comments
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_google.protobuf.SourceCodeInfo.Location.leading_detached_comments:

leading_detached_comments
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  



.. _api_msg_google.protobuf.GeneratedCodeInfo:

google.protobuf.GeneratedCodeInfo
---------------------------------

`[google.protobuf.GeneratedCodeInfo proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L861>`_

Describes the relationship between generated code and its original source
file. A GeneratedCodeInfo message is associated with only one generated
source file, but may contain references to different source .proto files.

.. code-block:: json

  {
    "annotation": []
  }

.. _api_field_google.protobuf.GeneratedCodeInfo.annotation:

annotation
  (:ref:`google.protobuf.GeneratedCodeInfo.Annotation <api_msg_google.protobuf.GeneratedCodeInfo.Annotation>`) An Annotation connects some span of text in generated code to an element
  of its generating .proto file.
  
  
.. _api_msg_google.protobuf.GeneratedCodeInfo.Annotation:

google.protobuf.GeneratedCodeInfo.Annotation
--------------------------------------------

`[google.protobuf.GeneratedCodeInfo.Annotation proto] <https://github.com/lyft/flyteidl/blob/master/protos/google/protobuf/descriptor.proto#L865>`_


.. code-block:: json

  {
    "path": [],
    "source_file": "...",
    "begin": "...",
    "end": "..."
  }

.. _api_field_google.protobuf.GeneratedCodeInfo.Annotation.path:

path
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifies the element in the original source .proto file. This field
  is formatted the same as SourceCodeInfo.Location.path.
  
  
.. _api_field_google.protobuf.GeneratedCodeInfo.Annotation.source_file:

source_file
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifies the filesystem path to the original source .proto.
  
  
.. _api_field_google.protobuf.GeneratedCodeInfo.Annotation.begin:

begin
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifies the starting offset in bytes in the generated code
  that relates to the identified object.
  
  
.. _api_field_google.protobuf.GeneratedCodeInfo.Annotation.end:

end
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifies the ending offset in bytes in the generated code that
  relates to the identified offset. The end offset should be one past
  the last relevant byte (so the length of the text = end - begin).
  
  


