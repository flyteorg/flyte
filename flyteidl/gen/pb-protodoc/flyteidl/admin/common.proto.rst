.. _api_file_flyteidl/admin/common.proto:

common.proto
===========================

.. _api_msg_flyteidl.admin.NamedEntityIdentifier:

flyteidl.admin.NamedEntityIdentifier
------------------------------------

`[flyteidl.admin.NamedEntityIdentifier proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L11>`_

Encapsulation of fields that identifies a Flyte resource.
A resource can internally have multiple versions.

.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "name": "..."
  }

.. _api_field_flyteidl.admin.NamedEntityIdentifier.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the project the resource belongs to.
  
  
.. _api_field_flyteidl.admin.NamedEntityIdentifier.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the domain the resource belongs to.
  A domain can be considered as a subset within a specific project.
  
  
.. _api_field_flyteidl.admin.NamedEntityIdentifier.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) User provided value for the resource.
  The combination of project + domain + name uniquely identifies the resource.
  +optional - in certain contexts - like 'List API', 'Launch plans'
  
  


.. _api_msg_flyteidl.admin.Sort:

flyteidl.admin.Sort
-------------------

`[flyteidl.admin.Sort proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L24>`_

Species sort ordering in a list request.

.. code-block:: json

  {
    "key": "...",
    "direction": "..."
  }

.. _api_field_flyteidl.admin.Sort.key:

key
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates an attribute to sort the response values.
  TODO(katrogan): Add string validation here. This should never be empty.
  
  
.. _api_field_flyteidl.admin.Sort.direction:

direction
  (:ref:`flyteidl.admin.Sort.Direction <api_enum_flyteidl.admin.Sort.Direction>`) Indicates the direction to apply sort key for response values.
  +optional
  
  

.. _api_enum_flyteidl.admin.Sort.Direction:

Enum flyteidl.admin.Sort.Direction
----------------------------------

`[flyteidl.admin.Sort.Direction proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L25>`_


.. _api_enum_value_flyteidl.admin.Sort.Direction.DESCENDING:

DESCENDING
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.admin.Sort.Direction.ASCENDING:

ASCENDING
  ⁣
  

.. _api_msg_flyteidl.admin.NamedEntityIdentifierListRequest:

flyteidl.admin.NamedEntityIdentifierListRequest
-----------------------------------------------

`[flyteidl.admin.NamedEntityIdentifierListRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L39>`_

Represents a request structure to list identifiers.

.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "limit": "...",
    "token": "...",
    "sort_by": "{...}"
  }

.. _api_field_flyteidl.admin.NamedEntityIdentifierListRequest.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the project that contains the identifiers.
  
  
.. _api_field_flyteidl.admin.NamedEntityIdentifierListRequest.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the domain the identifiers belongs to within the project.
  
  
.. _api_field_flyteidl.admin.NamedEntityIdentifierListRequest.limit:

limit
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates the number of resources to be returned.
  
  
.. _api_field_flyteidl.admin.NamedEntityIdentifierListRequest.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query.
  +optional
  
  
.. _api_field_flyteidl.admin.NamedEntityIdentifierListRequest.sort_by:

sort_by
  (:ref:`flyteidl.admin.Sort <api_msg_flyteidl.admin.Sort>`) Sort ordering.
  +optional
  
  


.. _api_msg_flyteidl.admin.NamedEntityIdentifierList:

flyteidl.admin.NamedEntityIdentifierList
----------------------------------------

`[flyteidl.admin.NamedEntityIdentifierList proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L57>`_

Represents a list of Identifiers.

.. code-block:: json

  {
    "entities": [],
    "token": "..."
  }

.. _api_field_flyteidl.admin.NamedEntityIdentifierList.entities:

entities
  (:ref:`flyteidl.admin.NamedEntityIdentifier <api_msg_flyteidl.admin.NamedEntityIdentifier>`) A list of identifiers.
  
  
.. _api_field_flyteidl.admin.NamedEntityIdentifierList.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query. If there are no more results, this value will be empty.
  
  


.. _api_msg_flyteidl.admin.ObjectGetRequest:

flyteidl.admin.ObjectGetRequest
-------------------------------

`[flyteidl.admin.ObjectGetRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L67>`_

Represents a structure to fetch a single resource.

.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_flyteidl.admin.ObjectGetRequest.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) Indicates a unique version of resource.
  
  


.. _api_msg_flyteidl.admin.ResourceListRequest:

flyteidl.admin.ResourceListRequest
----------------------------------

`[flyteidl.admin.ResourceListRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L74>`_

Represents a request structure to retrieve a list of resources.
Resources include: Task, Workflow, LaunchPlan

.. code-block:: json

  {
    "id": "{...}",
    "limit": "...",
    "token": "...",
    "filters": "...",
    "sort_by": "{...}"
  }

.. _api_field_flyteidl.admin.ResourceListRequest.id:

id
  (:ref:`flyteidl.admin.NamedEntityIdentifier <api_msg_flyteidl.admin.NamedEntityIdentifier>`) id represents the unique identifier of the resource.
  
  
.. _api_field_flyteidl.admin.ResourceListRequest.limit:

limit
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates the number of resources to be returned.
  
  
.. _api_field_flyteidl.admin.ResourceListRequest.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, this server-provided token can be used to fetch the next page
  in a query.
  +optional
  
  
.. _api_field_flyteidl.admin.ResourceListRequest.filters:

filters
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates a list of filters passed as string.
  More info on constructing filters : <Link>
  +optional
  
  
.. _api_field_flyteidl.admin.ResourceListRequest.sort_by:

sort_by
  (:ref:`flyteidl.admin.Sort <api_msg_flyteidl.admin.Sort>`) Sort ordering.
  +optional
  
  


.. _api_msg_flyteidl.admin.EmailNotification:

flyteidl.admin.EmailNotification
--------------------------------

`[flyteidl.admin.EmailNotification proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L93>`_


.. code-block:: json

  {
    "recipients_email": []
  }

.. _api_field_flyteidl.admin.EmailNotification.recipients_email:

recipients_email
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The list of email addresses recipients for this notification.
  
  


.. _api_msg_flyteidl.admin.PagerDutyNotification:

flyteidl.admin.PagerDutyNotification
------------------------------------

`[flyteidl.admin.PagerDutyNotification proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L98>`_


.. code-block:: json

  {
    "recipients_email": []
  }

.. _api_field_flyteidl.admin.PagerDutyNotification.recipients_email:

recipients_email
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Currently, PagerDuty notifications leverage email to trigger a notification.
  
  


.. _api_msg_flyteidl.admin.SlackNotification:

flyteidl.admin.SlackNotification
--------------------------------

`[flyteidl.admin.SlackNotification proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L103>`_


.. code-block:: json

  {
    "recipients_email": []
  }

.. _api_field_flyteidl.admin.SlackNotification.recipients_email:

recipients_email
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Currently, Slack notifications leverage email to trigger a notification.
  
  


.. _api_msg_flyteidl.admin.Notification:

flyteidl.admin.Notification
---------------------------

`[flyteidl.admin.Notification proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L112>`_

Represents a structure for notifications based on execution status.
The Notification content is configured within Admin. Future iterations could
expose configuring notifications with custom content.

.. code-block:: json

  {
    "phases": [],
    "email": "{...}",
    "pager_duty": "{...}",
    "slack": "{...}"
  }

.. _api_field_flyteidl.admin.Notification.phases:

phases
  (:ref:`flyteidl.core.WorkflowExecution.Phase <api_enum_flyteidl.core.WorkflowExecution.Phase>`) A list of phases to which users can associate the notifications to.
  
  
.. _api_field_flyteidl.admin.Notification.email:

email
  (:ref:`flyteidl.admin.EmailNotification <api_msg_flyteidl.admin.EmailNotification>`) option (validate.required) = true;
  
  
  
  Only one of :ref:`email <api_field_flyteidl.admin.Notification.email>`, :ref:`pager_duty <api_field_flyteidl.admin.Notification.pager_duty>`, :ref:`slack <api_field_flyteidl.admin.Notification.slack>` may be set.
  
.. _api_field_flyteidl.admin.Notification.pager_duty:

pager_duty
  (:ref:`flyteidl.admin.PagerDutyNotification <api_msg_flyteidl.admin.PagerDutyNotification>`) 
  
  
  Only one of :ref:`email <api_field_flyteidl.admin.Notification.email>`, :ref:`pager_duty <api_field_flyteidl.admin.Notification.pager_duty>`, :ref:`slack <api_field_flyteidl.admin.Notification.slack>` may be set.
  
.. _api_field_flyteidl.admin.Notification.slack:

slack
  (:ref:`flyteidl.admin.SlackNotification <api_msg_flyteidl.admin.SlackNotification>`) 
  
  
  Only one of :ref:`email <api_field_flyteidl.admin.Notification.email>`, :ref:`pager_duty <api_field_flyteidl.admin.Notification.pager_duty>`, :ref:`slack <api_field_flyteidl.admin.Notification.slack>` may be set.
  


.. _api_msg_flyteidl.admin.UrlBlob:

flyteidl.admin.UrlBlob
----------------------

`[flyteidl.admin.UrlBlob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L126>`_

Represents a string url and associated metadata used throughout the platform.

.. code-block:: json

  {
    "url": "...",
    "bytes": "..."
  }

.. _api_field_flyteidl.admin.UrlBlob.url:

url
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Actual url value.
  
  
.. _api_field_flyteidl.admin.UrlBlob.bytes:

bytes
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Represents the size of the file accessible at the above url.
  
  


.. _api_msg_flyteidl.admin.Labels:

flyteidl.admin.Labels
---------------------

`[flyteidl.admin.Labels proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L137>`_

Label values to be applied to an execution resource.
In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
to specify how to merge labels defined at registration and execution time.

.. code-block:: json

  {
    "values": "{...}"
  }

.. _api_field_flyteidl.admin.Labels.values:

values
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, `string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_>) Map of custom labels to be applied to the execution resource.
  
  


.. _api_msg_flyteidl.admin.Annotations:

flyteidl.admin.Annotations
--------------------------

`[flyteidl.admin.Annotations proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/common.proto#L145>`_

Annotation values to be applied to an execution resource.
In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
to specify how to merge annotations defined at registration and execution time.

.. code-block:: json

  {
    "values": "{...}"
  }

.. _api_field_flyteidl.admin.Annotations.values:

values
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, `string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_>) Map of custom annotations to be applied to the execution resource.
  
  

