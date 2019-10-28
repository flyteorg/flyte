.. _api_file_protoc-gen-swagger/options/openapiv2.proto:

openapiv2.proto
==========================================

.. _api_msg_grpc.gateway.protoc_gen_swagger.options.Swagger:

grpc.gateway.protoc_gen_swagger.options.Swagger
-----------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Swagger proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L13>`_

`Swagger` is a representation of OpenAPI v2 specification's Swagger object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#swaggerObject

TODO(ivucica): document fields

.. code-block:: json

  {
    "swagger": "...",
    "info": "{...}",
    "host": "...",
    "base_path": "...",
    "schemes": [],
    "consumes": [],
    "produces": [],
    "responses": "{...}",
    "security_definitions": "{...}",
    "security": [],
    "external_docs": "{...}"
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.swagger:

swagger
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.info:

info
  (:ref:`grpc.gateway.protoc_gen_swagger.options.Info <api_msg_grpc.gateway.protoc_gen_swagger.options.Info>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.host:

host
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.base_path:

base_path
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.schemes:

schemes
  (:ref:`grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme <api_enum_grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.consumes:

consumes
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.produces:

produces
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.responses:

responses
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`grpc.gateway.protoc_gen_swagger.options.Response <api_msg_grpc.gateway.protoc_gen_swagger.options.Response>`>) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.security_definitions:

security_definitions
  (:ref:`grpc.gateway.protoc_gen_swagger.options.SecurityDefinitions <api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityDefinitions>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.security:

security
  (:ref:`grpc.gateway.protoc_gen_swagger.options.SecurityRequirement <api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityRequirement>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Swagger.external_docs:

external_docs
  (:ref:`grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation <api_msg_grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation>`) 
  

.. _api_enum_grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme:

Enum grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme
------------------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L18>`_


.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme.UNKNOWN:

UNKNOWN
  *(DEFAULT)* ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme.HTTP:

HTTP
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme.HTTPS:

HTTPS
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme.WS:

WS
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.Swagger.SwaggerScheme.WSS:

WSS
  ⁣
  

.. _api_msg_grpc.gateway.protoc_gen_swagger.options.Operation:

grpc.gateway.protoc_gen_swagger.options.Operation
-------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Operation proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L48>`_

`Operation` is a representation of OpenAPI v2 specification's Operation object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#operationObject

TODO(ivucica): document fields

.. code-block:: json

  {
    "tags": [],
    "summary": "...",
    "description": "...",
    "external_docs": "{...}",
    "operation_id": "...",
    "consumes": [],
    "produces": [],
    "responses": "{...}",
    "schemes": [],
    "deprecated": "...",
    "security": []
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.tags:

tags
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.summary:

summary
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.description:

description
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.external_docs:

external_docs
  (:ref:`grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation <api_msg_grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.operation_id:

operation_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.consumes:

consumes
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.produces:

produces
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.responses:

responses
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`grpc.gateway.protoc_gen_swagger.options.Response <api_msg_grpc.gateway.protoc_gen_swagger.options.Response>`>) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.schemes:

schemes
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.deprecated:

deprecated
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Operation.security:

security
  (:ref:`grpc.gateway.protoc_gen_swagger.options.SecurityRequirement <api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityRequirement>`) 
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.Response:

grpc.gateway.protoc_gen_swagger.options.Response
------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Response proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L68>`_

`Response` is a representation of OpenAPI v2 specification's Response object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#responseObject


.. code-block:: json

  {
    "description": "...",
    "schema": "{...}"
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.Response.description:

description
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) `Description` is a short description of the response.
  GFM syntax can be used for rich text representation.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Response.schema:

schema
  (:ref:`grpc.gateway.protoc_gen_swagger.options.Schema <api_msg_grpc.gateway.protoc_gen_swagger.options.Schema>`) `Schema` optionally defines the structure of the response.
  If `Schema` is not provided, it means there is no content to the response.
  
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.Info:

grpc.gateway.protoc_gen_swagger.options.Info
--------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Info proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L86>`_

`Info` is a representation of OpenAPI v2 specification's Info object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#infoObject

TODO(ivucica): document fields

.. code-block:: json

  {
    "title": "...",
    "description": "...",
    "terms_of_service": "...",
    "contact": "{...}",
    "license": "{...}",
    "version": "..."
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.Info.title:

title
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Info.description:

description
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Info.terms_of_service:

terms_of_service
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Info.contact:

contact
  (:ref:`grpc.gateway.protoc_gen_swagger.options.Contact <api_msg_grpc.gateway.protoc_gen_swagger.options.Contact>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Info.license:

license
  (:ref:`grpc.gateway.protoc_gen_swagger.options.License <api_msg_grpc.gateway.protoc_gen_swagger.options.License>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Info.version:

version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.Contact:

grpc.gateway.protoc_gen_swagger.options.Contact
-----------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Contact proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L100>`_

`Contact` is a representation of OpenAPI v2 specification's Contact object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#contactObject

TODO(ivucica): document fields

.. code-block:: json

  {
    "name": "...",
    "url": "...",
    "email": "..."
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.Contact.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Contact.url:

url
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Contact.email:

email
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.License:

grpc.gateway.protoc_gen_swagger.options.License
-----------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.License proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L110>`_

`License` is a representation of OpenAPI v2 specification's License object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#licenseObject


.. code-block:: json

  {
    "name": "...",
    "url": "..."
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.License.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Required. The license name used for the API.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.License.url:

url
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A URL to the license used for the API.
  
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation:

grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation
-------------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L123>`_

`ExternalDocumentation` is a representation of OpenAPI v2 specification's
ExternalDocumentation object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#externalDocumentationObject

TODO(ivucica): document fields

.. code-block:: json

  {
    "description": "...",
    "url": "..."
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation.description:

description
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation.url:

url
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.Schema:

grpc.gateway.protoc_gen_swagger.options.Schema
----------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Schema proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L133>`_

`Schema` is a representation of OpenAPI v2 specification's Schema object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#schemaObject

TODO(ivucica): document fields

.. code-block:: json

  {
    "json_schema": "{...}",
    "discriminator": "...",
    "read_only": "...",
    "external_docs": "{...}",
    "example": "{...}"
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.Schema.json_schema:

json_schema
  (:ref:`grpc.gateway.protoc_gen_swagger.options.JSONSchema <api_msg_grpc.gateway.protoc_gen_swagger.options.JSONSchema>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Schema.discriminator:

discriminator
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Schema.read_only:

read_only
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Schema.external_docs:

external_docs
  (:ref:`grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation <api_msg_grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation>`) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Schema.example:

example
  (:ref:`google.protobuf.Any <api_msg_google.protobuf.Any>`) 
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.JSONSchema:

grpc.gateway.protoc_gen_swagger.options.JSONSchema
--------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.JSONSchema proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L154>`_

`JSONSchema` represents properties from JSON Schema taken, and as used, in
the OpenAPI v2 spec.

This includes changes made by OpenAPI v2.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#schemaObject

See also: https://cswr.github.io/JsonSchema/spec/basic_types/,
https://github.com/json-schema-org/json-schema-spec/blob/master/schema.json

TODO(ivucica): document fields

.. code-block:: json

  {
    "ref": "...",
    "title": "...",
    "description": "...",
    "default": "...",
    "read_only": "...",
    "multiple_of": "...",
    "maximum": "...",
    "exclusive_maximum": "...",
    "minimum": "...",
    "exclusive_minimum": "...",
    "max_length": "...",
    "min_length": "...",
    "pattern": "...",
    "max_items": "...",
    "min_items": "...",
    "unique_items": "...",
    "max_properties": "...",
    "min_properties": "...",
    "required": [],
    "array": [],
    "type": []
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.ref:

ref
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Ref is used to define an external reference to include in the message.
  This could be a fully qualified proto message reference, and that type must be imported
  into the protofile. If no message is identified, the Ref will be used verbatim in
  the output.
  For example:
   `ref: ".google.protobuf.Timestamp"`.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.title:

title
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.description:

description
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.default:

default
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.read_only:

read_only
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.multiple_of:

multiple_of
  (`double <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.maximum:

maximum
  (`double <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.exclusive_maximum:

exclusive_maximum
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.minimum:

minimum
  (`double <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.exclusive_minimum:

exclusive_minimum
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.max_length:

max_length
  (`uint64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.min_length:

min_length
  (`uint64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.pattern:

pattern
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.max_items:

max_items
  (`uint64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.min_items:

min_items
  (`uint64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.unique_items:

unique_items
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.max_properties:

max_properties
  (`uint64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.min_properties:

min_properties
  (`uint64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.required:

required
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.array:

array
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Items in 'array' must be unique.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.JSONSchema.type:

type
  (:ref:`grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes <api_enum_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes>`) 
  

.. _api_enum_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes:

Enum grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes
-----------------------------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L206>`_


.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes.UNKNOWN:

UNKNOWN
  *(DEFAULT)* ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes.ARRAY:

ARRAY
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes.BOOLEAN:

BOOLEAN
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes.INTEGER:

INTEGER
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes.NULL:

NULL
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes.NUMBER:

NUMBER
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes.OBJECT:

OBJECT
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.JSONSchema.JSONSchemaSimpleTypes.STRING:

STRING
  ⁣
  

.. _api_msg_grpc.gateway.protoc_gen_swagger.options.Tag:

grpc.gateway.protoc_gen_swagger.options.Tag
-------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Tag proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L233>`_

`Tag` is a representation of OpenAPI v2 specification's Tag object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#tagObject

TODO(ivucica): document fields

.. code-block:: json

  {
    "description": "...",
    "external_docs": "{...}"
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.Tag.description:

description
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) TODO(ivucica): Description should be extracted from comments on the proto
  service object.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.Tag.external_docs:

external_docs
  (:ref:`grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation <api_msg_grpc.gateway.protoc_gen_swagger.options.ExternalDocumentation>`) 
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityDefinitions:

grpc.gateway.protoc_gen_swagger.options.SecurityDefinitions
-----------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.SecurityDefinitions proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L257>`_

`SecurityDefinitions` is a representation of OpenAPI v2 specification's
Security Definitions object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#securityDefinitionsObject

A declaration of the security schemes available to be used in the
specification. This does not enforce the security schemes on the operations
and only serves to provide the relevant details for each scheme.

.. code-block:: json

  {
    "security": "{...}"
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityDefinitions.security:

security
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`grpc.gateway.protoc_gen_swagger.options.SecurityScheme <api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityScheme>`>) A single security scheme definition, mapping a "name" to the scheme it defines.
  
  


.. _api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityScheme:

grpc.gateway.protoc_gen_swagger.options.SecurityScheme
------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.SecurityScheme proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L271>`_

`SecurityScheme` is a representation of OpenAPI v2 specification's
Security Scheme object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#securitySchemeObject

Allows the definition of a security scheme that can be used by the
operations. Supported schemes are basic authentication, an API key (either as
a header or as a query parameter) and OAuth2's common flows (implicit,
password, application and access code).

.. code-block:: json

  {
    "type": "...",
    "description": "...",
    "name": "...",
    "in": "...",
    "flow": "...",
    "authorization_url": "...",
    "token_url": "...",
    "scopes": "{...}"
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.type:

type
  (:ref:`grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type <api_enum_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type>`) Required. The type of the security scheme. Valid values are "basic",
  "apiKey" or "oauth2".
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.description:

description
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A short description for security scheme.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Required. The name of the header or query parameter to be used.
  
  Valid for apiKey.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.in:

in
  (:ref:`grpc.gateway.protoc_gen_swagger.options.SecurityScheme.In <api_enum_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.In>`) Required. The location of the API key. Valid values are "query" or "header".
  
  Valid for apiKey.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.flow:

flow
  (:ref:`grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow <api_enum_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow>`) Required. The flow used by the OAuth2 security scheme. Valid values are
  "implicit", "password", "application" or "accessCode".
  
  Valid for oauth2.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.authorization_url:

authorization_url
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Required. The authorization URL to be used for this flow. This SHOULD be in
  the form of a URL.
  
  Valid for oauth2/implicit and oauth2/accessCode.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.token_url:

token_url
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Required. The token URL to be used for this flow. This SHOULD be in the
  form of a URL.
  
  Valid for oauth2/password, oauth2/application and oauth2/accessCode.
  
  
.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.scopes:

scopes
  (:ref:`grpc.gateway.protoc_gen_swagger.options.Scopes <api_msg_grpc.gateway.protoc_gen_swagger.options.Scopes>`) Required. The available scopes for the OAuth2 security scheme.
  
  Valid for oauth2.
  
  

.. _api_enum_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type:

Enum grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type
----------------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L274>`_

Required. The type of the security scheme. Valid values are "basic",
"apiKey" or "oauth2".

.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type.TYPE_INVALID:

TYPE_INVALID
  *(DEFAULT)* ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type.TYPE_BASIC:

TYPE_BASIC
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type.TYPE_API_KEY:

TYPE_API_KEY
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Type.TYPE_OAUTH2:

TYPE_OAUTH2
  ⁣
  

.. _api_enum_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.In:

Enum grpc.gateway.protoc_gen_swagger.options.SecurityScheme.In
--------------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.SecurityScheme.In proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L282>`_

Required. The location of the API key. Valid values are "query" or "header".

.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.In.IN_INVALID:

IN_INVALID
  *(DEFAULT)* ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.In.IN_QUERY:

IN_QUERY
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.In.IN_HEADER:

IN_HEADER
  ⁣
  

.. _api_enum_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow:

Enum grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow
----------------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L290>`_

Required. The flow used by the OAuth2 security scheme. Valid values are
"implicit", "password", "application" or "accessCode".

.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow.FLOW_INVALID:

FLOW_INVALID
  *(DEFAULT)* ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow.FLOW_IMPLICIT:

FLOW_IMPLICIT
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow.FLOW_PASSWORD:

FLOW_PASSWORD
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow.FLOW_APPLICATION:

FLOW_APPLICATION
  ⁣
  
.. _api_enum_value_grpc.gateway.protoc_gen_swagger.options.SecurityScheme.Flow.FLOW_ACCESS_CODE:

FLOW_ACCESS_CODE
  ⁣
  

.. _api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityRequirement:

grpc.gateway.protoc_gen_swagger.options.SecurityRequirement
-----------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.SecurityRequirement proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L343>`_

`SecurityRequirement` is a representation of OpenAPI v2 specification's
Security Requirement object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#securityRequirementObject

Lists the required security schemes to execute this operation. The object can
have multiple security schemes declared in it which are all required (that
is, there is a logical AND between the schemes).

The name used for each property MUST correspond to a security scheme
declared in the Security Definitions.

.. code-block:: json

  {
    "security_requirement": "{...}"
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityRequirement.security_requirement:

security_requirement
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`grpc.gateway.protoc_gen_swagger.options.SecurityRequirement.SecurityRequirementValue <api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityRequirement.SecurityRequirementValue>`>) Each name must correspond to a security scheme which is declared in
  the Security Definitions. If the security scheme is of type "oauth2",
  then the value is a list of scope names required for the execution.
  For other security scheme types, the array MUST be empty.
  
  
.. _api_msg_grpc.gateway.protoc_gen_swagger.options.SecurityRequirement.SecurityRequirementValue:

grpc.gateway.protoc_gen_swagger.options.SecurityRequirement.SecurityRequirementValue
------------------------------------------------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.SecurityRequirement.SecurityRequirementValue proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L347>`_

If the security scheme is of type "oauth2", then the value is a list of
scope names required for the execution. For other security scheme types,
the array MUST be empty.

.. code-block:: json

  {
    "scope": []
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.SecurityRequirement.SecurityRequirementValue.scope:

scope
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  




.. _api_msg_grpc.gateway.protoc_gen_swagger.options.Scopes:

grpc.gateway.protoc_gen_swagger.options.Scopes
----------------------------------------------

`[grpc.gateway.protoc_gen_swagger.options.Scopes proto] <https://github.com/lyft/flyteidl/blob/master/protos/protoc-gen-swagger/options/openapiv2.proto#L362>`_

`Scopes` is a representation of OpenAPI v2 specification's Scopes object.

See: https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#scopesObject

Lists the available scopes for an OAuth2 security scheme.

.. code-block:: json

  {
    "scope": "{...}"
  }

.. _api_field_grpc.gateway.protoc_gen_swagger.options.Scopes.scope:

scope
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, `string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_>) Maps between a name of a scope to a short description of it (as the value
  of the property).
  
  

