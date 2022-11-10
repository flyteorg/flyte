######################
Protocol Documentation
######################




.. _ref_flyteidl/admin/cluster_assignment.proto:

flyteidl/admin/cluster_assignment.proto
==================================================================





.. _ref_flyteidl.admin.ClusterAssignment:

ClusterAssignment
------------------------------------------------------------------

Encapsulates specifications for routing an execution onto a specific cluster.



.. csv-table:: ClusterAssignment type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "cluster_pool_name", ":ref:`ref_string`", "", ""






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/common.proto:

flyteidl/admin/common.proto
==================================================================





.. _ref_flyteidl.admin.Annotations:

Annotations
------------------------------------------------------------------

Annotation values to be applied to an execution resource.
In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
to specify how to merge annotations defined at registration and execution time.



.. csv-table:: Annotations type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "values", ":ref:`ref_flyteidl.admin.Annotations.ValuesEntry`", "repeated", "Map of custom annotations to be applied to the execution resource."







.. _ref_flyteidl.admin.Annotations.ValuesEntry:

Annotations.ValuesEntry
------------------------------------------------------------------





.. csv-table:: Annotations.ValuesEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.admin.AuthRole:

AuthRole
------------------------------------------------------------------

Defines permissions associated with executions created by this launch plan spec.
Use either of these roles when they have permissions required by your workflow execution.
Deprecated.



.. csv-table:: AuthRole type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "assumable_iam_role", ":ref:`ref_string`", "", "Defines an optional iam role which will be used for tasks run in executions created with this launch plan."
   "kubernetes_service_account", ":ref:`ref_string`", "", "Defines an optional kubernetes service account which will be used for tasks run in executions created with this launch plan."







.. _ref_flyteidl.admin.EmailNotification:

EmailNotification
------------------------------------------------------------------

Defines an email notification specification.



.. csv-table:: EmailNotification type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "recipients_email", ":ref:`ref_string`", "repeated", "The list of email addresses recipients for this notification. +required"







.. _ref_flyteidl.admin.Labels:

Labels
------------------------------------------------------------------

Label values to be applied to an execution resource.
In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
to specify how to merge labels defined at registration and execution time.



.. csv-table:: Labels type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "values", ":ref:`ref_flyteidl.admin.Labels.ValuesEntry`", "repeated", "Map of custom labels to be applied to the execution resource."







.. _ref_flyteidl.admin.Labels.ValuesEntry:

Labels.ValuesEntry
------------------------------------------------------------------





.. csv-table:: Labels.ValuesEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.admin.NamedEntity:

NamedEntity
------------------------------------------------------------------

Encapsulates information common to a NamedEntity, a Flyte resource such as a task,
workflow or launch plan. A NamedEntity is exclusively identified by its resource type
and identifier.



.. csv-table:: NamedEntity type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "resource_type", ":ref:`ref_flyteidl.core.ResourceType`", "", "Resource type of the named entity. One of Task, Workflow or LaunchPlan."
   "id", ":ref:`ref_flyteidl.admin.NamedEntityIdentifier`", "", ""
   "metadata", ":ref:`ref_flyteidl.admin.NamedEntityMetadata`", "", "Additional metadata around a named entity."







.. _ref_flyteidl.admin.NamedEntityGetRequest:

NamedEntityGetRequest
------------------------------------------------------------------

A request to retrieve the metadata associated with a NamedEntityIdentifier



.. csv-table:: NamedEntityGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "resource_type", ":ref:`ref_flyteidl.core.ResourceType`", "", "Resource type of the metadata to get. One of Task, Workflow or LaunchPlan. +required"
   "id", ":ref:`ref_flyteidl.admin.NamedEntityIdentifier`", "", "The identifier for the named entity for which to fetch metadata. +required"







.. _ref_flyteidl.admin.NamedEntityIdentifier:

NamedEntityIdentifier
------------------------------------------------------------------

Encapsulation of fields that identifies a Flyte resource.
A Flyte resource can be a task, workflow or launch plan.
A resource can internally have multiple versions and is uniquely identified
by project, domain, and name.



.. csv-table:: NamedEntityIdentifier type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Name of the project the resource belongs to."
   "domain", ":ref:`ref_string`", "", "Name of the domain the resource belongs to. A domain can be considered as a subset within a specific project."
   "name", ":ref:`ref_string`", "", "User provided value for the resource. The combination of project + domain + name uniquely identifies the resource. +optional - in certain contexts - like 'List API', 'Launch plans'"







.. _ref_flyteidl.admin.NamedEntityIdentifierList:

NamedEntityIdentifierList
------------------------------------------------------------------

Represents a list of NamedEntityIdentifiers.



.. csv-table:: NamedEntityIdentifierList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "entities", ":ref:`ref_flyteidl.admin.NamedEntityIdentifier`", "repeated", "A list of identifiers."
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.NamedEntityIdentifierListRequest:

NamedEntityIdentifierListRequest
------------------------------------------------------------------

Represents a request structure to list NamedEntityIdentifiers.



.. csv-table:: NamedEntityIdentifierListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Name of the project that contains the identifiers. +required"
   "domain", ":ref:`ref_string`", "", "Name of the domain the identifiers belongs to within the project. +required"
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned. +required"
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Specifies how listed entities should be sorted in the response. +optional"
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. +optional"







.. _ref_flyteidl.admin.NamedEntityList:

NamedEntityList
------------------------------------------------------------------

Represents a list of NamedEntityIdentifiers.



.. csv-table:: NamedEntityList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "entities", ":ref:`ref_flyteidl.admin.NamedEntity`", "repeated", "A list of NamedEntity objects"
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.NamedEntityListRequest:

NamedEntityListRequest
------------------------------------------------------------------

Represents a request structure to list NamedEntity objects



.. csv-table:: NamedEntityListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "resource_type", ":ref:`ref_flyteidl.core.ResourceType`", "", "Resource type of the metadata to query. One of Task, Workflow or LaunchPlan. +required"
   "project", ":ref:`ref_string`", "", "Name of the project that contains the identifiers. +required"
   "domain", ":ref:`ref_string`", "", "Name of the domain the identifiers belongs to within the project."
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned."
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Specifies how listed entities should be sorted in the response. +optional"
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. +optional"







.. _ref_flyteidl.admin.NamedEntityMetadata:

NamedEntityMetadata
------------------------------------------------------------------

Additional metadata around a named entity.



.. csv-table:: NamedEntityMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "description", ":ref:`ref_string`", "", "Common description across all versions of the entity +optional"
   "state", ":ref:`ref_flyteidl.admin.NamedEntityState`", "", "Shared state across all version of the entity At this point in time, only workflow entities can have their state archived."







.. _ref_flyteidl.admin.NamedEntityUpdateRequest:

NamedEntityUpdateRequest
------------------------------------------------------------------

Request to set the referenced named entity state to the configured value.



.. csv-table:: NamedEntityUpdateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "resource_type", ":ref:`ref_flyteidl.core.ResourceType`", "", "Resource type of the metadata to update +required"
   "id", ":ref:`ref_flyteidl.admin.NamedEntityIdentifier`", "", "Identifier of the metadata to update +required"
   "metadata", ":ref:`ref_flyteidl.admin.NamedEntityMetadata`", "", "Metadata object to set as the new value +required"







.. _ref_flyteidl.admin.NamedEntityUpdateResponse:

NamedEntityUpdateResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.Notification:

Notification
------------------------------------------------------------------

Represents a structure for notifications based on execution status.
The notification content is configured within flyte admin but can be templatized.
Future iterations could expose configuring notifications with custom content.



.. csv-table:: Notification type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "phases", ":ref:`ref_flyteidl.core.WorkflowExecution.Phase`", "repeated", "A list of phases to which users can associate the notifications to. +required"
   "email", ":ref:`ref_flyteidl.admin.EmailNotification`", "", ""
   "pager_duty", ":ref:`ref_flyteidl.admin.PagerDutyNotification`", "", ""
   "slack", ":ref:`ref_flyteidl.admin.SlackNotification`", "", ""







.. _ref_flyteidl.admin.ObjectGetRequest:

ObjectGetRequest
------------------------------------------------------------------

Shared request structure to fetch a single resource.
Resources include: Task, Workflow, LaunchPlan



.. csv-table:: ObjectGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "Indicates a unique version of resource. +required"







.. _ref_flyteidl.admin.PagerDutyNotification:

PagerDutyNotification
------------------------------------------------------------------

Defines a pager duty notification specification.



.. csv-table:: PagerDutyNotification type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "recipients_email", ":ref:`ref_string`", "repeated", "Currently, PagerDuty notifications leverage email to trigger a notification. +required"







.. _ref_flyteidl.admin.RawOutputDataConfig:

RawOutputDataConfig
------------------------------------------------------------------

Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).
See https://github.com/flyteorg/flyte/issues/211 for more background information.



.. csv-table:: RawOutputDataConfig type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "output_location_prefix", ":ref:`ref_string`", "", "Prefix for where offloaded data from user workflows will be written e.g. s3://bucket/key or s3://bucket/"







.. _ref_flyteidl.admin.ResourceListRequest:

ResourceListRequest
------------------------------------------------------------------

Shared request structure to retrieve a list of resources.
Resources include: Task, Workflow, LaunchPlan



.. csv-table:: ResourceListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.admin.NamedEntityIdentifier`", "", "id represents the unique identifier of the resource. +required"
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned. +required"
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, this server-provided token can be used to fetch the next page in a query. +optional"
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. More info on constructing filters : <Link> +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Sort ordering. +optional"







.. _ref_flyteidl.admin.SlackNotification:

SlackNotification
------------------------------------------------------------------

Defines a slack notification specification.



.. csv-table:: SlackNotification type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "recipients_email", ":ref:`ref_string`", "repeated", "Currently, Slack notifications leverage email to trigger a notification. +required"







.. _ref_flyteidl.admin.Sort:

Sort
------------------------------------------------------------------

Specifies sort ordering in a list request.



.. csv-table:: Sort type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", "Indicates an attribute to sort the response values. +required"
   "direction", ":ref:`ref_flyteidl.admin.Sort.Direction`", "", "Indicates the direction to apply sort key for response values. +optional"







.. _ref_flyteidl.admin.UrlBlob:

UrlBlob
------------------------------------------------------------------

Represents a string url and associated metadata used throughout the platform.



.. csv-table:: UrlBlob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "url", ":ref:`ref_string`", "", "Actual url value."
   "bytes", ":ref:`ref_int64`", "", "Represents the size of the file accessible at the above url."






..
   end messages



.. _ref_flyteidl.admin.NamedEntityState:

NamedEntityState
------------------------------------------------------------------

The status of the named entity is used to control its visibility in the UI.

.. csv-table:: Enum NamedEntityState values
   :header: "Name", "Number", "Description"
   :widths: auto

   "NAMED_ENTITY_ACTIVE", "0", "By default, all named entities are considered active and under development."
   "NAMED_ENTITY_ARCHIVED", "1", "Archived named entities are no longer visible in the UI."
   "SYSTEM_GENERATED", "2", "System generated entities that aren't explicitly created or managed by a user."



.. _ref_flyteidl.admin.Sort.Direction:

Sort.Direction
------------------------------------------------------------------



.. csv-table:: Enum Sort.Direction values
   :header: "Name", "Number", "Description"
   :widths: auto

   "DESCENDING", "0", "By default, fields are sorted in descending order."
   "ASCENDING", "1", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/description_entity.proto:

flyteidl/admin/description_entity.proto
==================================================================





.. _ref_flyteidl.admin.Description:

Description
------------------------------------------------------------------

Full user description with formatting preserved. This can be rendered
by clients, such as the console or command line tools with in-tact
formatting.



.. csv-table:: Description type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_string`", "", "long description - no more than 4KB"
   "uri", ":ref:`ref_string`", "", "if the description sizes exceed some threshold we can offload the entire description proto altogether to an external data store, like S3 rather than store inline in the db"
   "format", ":ref:`ref_flyteidl.admin.DescriptionFormat`", "", "Format of the long description"
   "icon_link", ":ref:`ref_string`", "", "Optional link to an icon for the entity"







.. _ref_flyteidl.admin.DescriptionEntity:

DescriptionEntity
------------------------------------------------------------------

DescriptionEntity contains detailed description for the task/workflow.
Documentation could provide insight into the algorithms, business use case, etc.



.. csv-table:: DescriptionEntity type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "id represents the unique identifier of the description entity."
   "short_description", ":ref:`ref_string`", "", "One-liner overview of the entity."
   "long_description", ":ref:`ref_flyteidl.admin.Description`", "", "Full user description with formatting preserved."
   "source_code", ":ref:`ref_flyteidl.admin.SourceCode`", "", "Optional link to source code used to define this entity."
   "tags", ":ref:`ref_string`", "repeated", "User-specified tags. These are arbitrary and can be used for searching filtering and discovering tasks."







.. _ref_flyteidl.admin.DescriptionEntityList:

DescriptionEntityList
------------------------------------------------------------------

Represents a list of DescriptionEntities returned from the admin.
See :ref:`ref_flyteidl.admin.DescriptionEntity` for more details



.. csv-table:: DescriptionEntityList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "descriptionEntities", ":ref:`ref_flyteidl.admin.DescriptionEntity`", "repeated", "A list of DescriptionEntities returned based on the request."
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.DescriptionEntityListRequest:

DescriptionEntityListRequest
------------------------------------------------------------------

Represents a request structure to retrieve a list of DescriptionEntities.
See :ref:`ref_flyteidl.admin.DescriptionEntity` for more details



.. csv-table:: DescriptionEntityListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "resource_type", ":ref:`ref_flyteidl.core.ResourceType`", "", "Identifies the specific type of resource that this identifier corresponds to."
   "id", ":ref:`ref_flyteidl.admin.NamedEntityIdentifier`", "", "The identifier for the description entity. +required"
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned. +required"
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. +optional"
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. More info on constructing filters : <Link> +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Sort ordering for returned list. +optional"







.. _ref_flyteidl.admin.SourceCode:

SourceCode
------------------------------------------------------------------

Link to source code used to define this entity



.. csv-table:: SourceCode type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "link", ":ref:`ref_string`", "", ""






..
   end messages



.. _ref_flyteidl.admin.DescriptionFormat:

DescriptionFormat
------------------------------------------------------------------

The format of the long description

.. csv-table:: Enum DescriptionFormat values
   :header: "Name", "Number", "Description"
   :widths: auto

   "DESCRIPTION_FORMAT_UNKNOWN", "0", ""
   "DESCRIPTION_FORMAT_MARKDOWN", "1", ""
   "DESCRIPTION_FORMAT_HTML", "2", ""
   "DESCRIPTION_FORMAT_RST", "3", "python default documentation - comments is rst"


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/event.proto:

flyteidl/admin/event.proto
==================================================================





.. _ref_flyteidl.admin.EventErrorAlreadyInTerminalState:

EventErrorAlreadyInTerminalState
------------------------------------------------------------------

Indicates that a sent event was not used to update execution state due to
the referenced execution already being terminated (and therefore ineligible
for further state transitions).



.. csv-table:: EventErrorAlreadyInTerminalState type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "current_phase", ":ref:`ref_string`", "", "+required"







.. _ref_flyteidl.admin.EventErrorIncompatibleCluster:

EventErrorIncompatibleCluster
------------------------------------------------------------------

Indicates an event was rejected because it came from a different cluster than 
is on record as running the execution.



.. csv-table:: EventErrorIncompatibleCluster type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "cluster", ":ref:`ref_string`", "", "The cluster which has been recorded as processing the execution. +required"







.. _ref_flyteidl.admin.EventFailureReason:

EventFailureReason
------------------------------------------------------------------

Indicates why a sent event was not used to update execution.



.. csv-table:: EventFailureReason type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "already_in_terminal_state", ":ref:`ref_flyteidl.admin.EventErrorAlreadyInTerminalState`", "", ""
   "incompatible_cluster", ":ref:`ref_flyteidl.admin.EventErrorIncompatibleCluster`", "", ""







.. _ref_flyteidl.admin.NodeExecutionEventRequest:

NodeExecutionEventRequest
------------------------------------------------------------------

Request to send a notification that a node execution event has occurred.



.. csv-table:: NodeExecutionEventRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "request_id", ":ref:`ref_string`", "", "Unique ID for this request that can be traced between services"
   "event", ":ref:`ref_flyteidl.event.NodeExecutionEvent`", "", "Details about the event that occurred."







.. _ref_flyteidl.admin.NodeExecutionEventResponse:

NodeExecutionEventResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.TaskExecutionEventRequest:

TaskExecutionEventRequest
------------------------------------------------------------------

Request to send a notification that a task execution event has occurred.



.. csv-table:: TaskExecutionEventRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "request_id", ":ref:`ref_string`", "", "Unique ID for this request that can be traced between services"
   "event", ":ref:`ref_flyteidl.event.TaskExecutionEvent`", "", "Details about the event that occurred."







.. _ref_flyteidl.admin.TaskExecutionEventResponse:

TaskExecutionEventResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.WorkflowExecutionEventRequest:

WorkflowExecutionEventRequest
------------------------------------------------------------------

Request to send a notification that a workflow execution event has occurred.



.. csv-table:: WorkflowExecutionEventRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "request_id", ":ref:`ref_string`", "", "Unique ID for this request that can be traced between services"
   "event", ":ref:`ref_flyteidl.event.WorkflowExecutionEvent`", "", "Details about the event that occurred."







.. _ref_flyteidl.admin.WorkflowExecutionEventResponse:

WorkflowExecutionEventResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.







..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/execution.proto:

flyteidl/admin/execution.proto
==================================================================





.. _ref_flyteidl.admin.AbortMetadata:

AbortMetadata
------------------------------------------------------------------

Specifies metadata around an aborted workflow execution.



.. csv-table:: AbortMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "cause", ":ref:`ref_string`", "", "In the case of a user-specified abort, this will pass along the user-supplied cause."
   "principal", ":ref:`ref_string`", "", "Identifies the entity (if any) responsible for terminating the execution"







.. _ref_flyteidl.admin.Execution:

Execution
------------------------------------------------------------------

A workflow execution represents an instantiated workflow, including all inputs and additional
metadata as well as computed results included state, outputs, and duration-based attributes.
Used as a response object used in Get and List execution requests.



.. csv-table:: Execution type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Unique identifier of the workflow execution."
   "spec", ":ref:`ref_flyteidl.admin.ExecutionSpec`", "", "User-provided configuration and inputs for launching the execution."
   "closure", ":ref:`ref_flyteidl.admin.ExecutionClosure`", "", "Execution results."







.. _ref_flyteidl.admin.ExecutionClosure:

ExecutionClosure
------------------------------------------------------------------

Encapsulates the results of the Execution



.. csv-table:: ExecutionClosure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "outputs", ":ref:`ref_flyteidl.admin.LiteralMapBlob`", "", "**Deprecated.** Output URI in the case of a successful execution. DEPRECATED. Use GetExecutionData to fetch output data instead."
   "error", ":ref:`ref_flyteidl.core.ExecutionError`", "", "Error information in the case of a failed execution."
   "abort_cause", ":ref:`ref_string`", "", "**Deprecated.** In the case of a user-specified abort, this will pass along the user-supplied cause."
   "abort_metadata", ":ref:`ref_flyteidl.admin.AbortMetadata`", "", "In the case of a user-specified abort, this will pass along the user and their supplied cause."
   "output_data", ":ref:`ref_flyteidl.core.LiteralMap`", "", "**Deprecated.** Raw output data produced by this execution. DEPRECATED. Use GetExecutionData to fetch output data instead."
   "computed_inputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "**Deprecated.** Inputs computed and passed for execution. computed_inputs depends on inputs in ExecutionSpec, fixed and default inputs in launch plan"
   "phase", ":ref:`ref_flyteidl.core.WorkflowExecution.Phase`", "", "Most recent recorded phase for the execution."
   "started_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Reported time at which the execution began running."
   "duration", ":ref:`ref_google.protobuf.Duration`", "", "The amount of time the execution spent running."
   "created_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Reported time at which the execution was created."
   "updated_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Reported time at which the execution was last updated."
   "notifications", ":ref:`ref_flyteidl.admin.Notification`", "repeated", "The notification settings to use after merging the CreateExecutionRequest and the launch plan notification settings. An execution launched with notifications will always prefer that definition to notifications defined statically in a launch plan."
   "workflow_id", ":ref:`ref_flyteidl.core.Identifier`", "", "Identifies the workflow definition for this execution."
   "state_change_details", ":ref:`ref_flyteidl.admin.ExecutionStateChangeDetails`", "", "Provides the details of the last stage change"







.. _ref_flyteidl.admin.ExecutionCreateRequest:

ExecutionCreateRequest
------------------------------------------------------------------

Request to launch an execution with the given project, domain and optionally-assigned name.



.. csv-table:: ExecutionCreateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Name of the project the execution belongs to. +required"
   "domain", ":ref:`ref_string`", "", "Name of the domain the execution belongs to. A domain can be considered as a subset within a specific project. +required"
   "name", ":ref:`ref_string`", "", "User provided value for the resource. If none is provided the system will generate a unique string. +optional"
   "spec", ":ref:`ref_flyteidl.admin.ExecutionSpec`", "", "Additional fields necessary to launch the execution. +optional"
   "inputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "The inputs required to start the execution. All required inputs must be included in this map. If not required and not provided, defaults apply. +optional"







.. _ref_flyteidl.admin.ExecutionCreateResponse:

ExecutionCreateResponse
------------------------------------------------------------------

The unique identifier for a successfully created execution.
If the name was *not* specified in the create request, this identifier will include a generated name.



.. csv-table:: ExecutionCreateResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", ""







.. _ref_flyteidl.admin.ExecutionList:

ExecutionList
------------------------------------------------------------------

Used as a response for request to list executions.
See :ref:`ref_flyteidl.admin.Execution` for more details



.. csv-table:: ExecutionList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "executions", ":ref:`ref_flyteidl.admin.Execution`", "repeated", ""
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.ExecutionMetadata:

ExecutionMetadata
------------------------------------------------------------------

Represents attributes about an execution which are not required to launch the execution but are useful to record.
These attributes are assigned at launch time and do not change.



.. csv-table:: ExecutionMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "mode", ":ref:`ref_flyteidl.admin.ExecutionMetadata.ExecutionMode`", "", ""
   "principal", ":ref:`ref_string`", "", "Identifier of the entity that triggered this execution. For systems using back-end authentication any value set here will be discarded in favor of the authenticated user context."
   "nesting", ":ref:`ref_uint32`", "", "Indicates the nestedness of this execution. If a user launches a workflow execution, the default nesting is 0. If this execution further launches a workflow (child workflow), the nesting level is incremented by 0 => 1 Generally, if workflow at nesting level k launches a workflow then the child workflow will have nesting = k + 1."
   "scheduled_at", ":ref:`ref_google.protobuf.Timestamp`", "", "For scheduled executions, the requested time for execution for this specific schedule invocation."
   "parent_node_execution", ":ref:`ref_flyteidl.core.NodeExecutionIdentifier`", "", "Which subworkflow node (if any) launched this execution"
   "reference_execution", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Optional, a reference workflow execution related to this execution. In the case of a relaunch, this references the original workflow execution."
   "system_metadata", ":ref:`ref_flyteidl.admin.SystemMetadata`", "", "Optional, platform-specific metadata about the execution. In this the future this may be gated behind an ACL or some sort of authorization."







.. _ref_flyteidl.admin.ExecutionRecoverRequest:

ExecutionRecoverRequest
------------------------------------------------------------------

Request to recover the referenced execution.



.. csv-table:: ExecutionRecoverRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Identifier of the workflow execution to recover."
   "name", ":ref:`ref_string`", "", "User provided value for the recovered execution. If none is provided the system will generate a unique string. +optional"
   "metadata", ":ref:`ref_flyteidl.admin.ExecutionMetadata`", "", "Additional metadata which will be used to overwrite any metadata in the reference execution when triggering a recovery execution."







.. _ref_flyteidl.admin.ExecutionRelaunchRequest:

ExecutionRelaunchRequest
------------------------------------------------------------------

Request to relaunch the referenced execution.



.. csv-table:: ExecutionRelaunchRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Identifier of the workflow execution to relaunch. +required"
   "name", ":ref:`ref_string`", "", "User provided value for the relaunched execution. If none is provided the system will generate a unique string. +optional"
   "overwrite_cache", ":ref:`ref_bool`", "", "Allows for all cached values of a workflow and its tasks to be overwritten for a single execution. If enabled, all calculations are performed even if cached results would be available, overwriting the stored data once execution finishes successfully."







.. _ref_flyteidl.admin.ExecutionSpec:

ExecutionSpec
------------------------------------------------------------------

An ExecutionSpec encompasses all data used to launch this execution. The Spec does not change over the lifetime
of an execution as it progresses across phase changes.



.. csv-table:: ExecutionSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "launch_plan", ":ref:`ref_flyteidl.core.Identifier`", "", "Launch plan to be executed"
   "inputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "**Deprecated.** Input values to be passed for the execution"
   "metadata", ":ref:`ref_flyteidl.admin.ExecutionMetadata`", "", "Metadata for the execution"
   "notifications", ":ref:`ref_flyteidl.admin.NotificationList`", "", "List of notifications based on Execution status transitions When this list is not empty it is used rather than any notifications defined in the referenced launch plan. When this list is empty, the notifications defined for the launch plan will be applied."
   "disable_all", ":ref:`ref_bool`", "", "This should be set to true if all notifications are intended to be disabled for this execution."
   "labels", ":ref:`ref_flyteidl.admin.Labels`", "", "Labels to apply to the execution resource."
   "annotations", ":ref:`ref_flyteidl.admin.Annotations`", "", "Annotations to apply to the execution resource."
   "security_context", ":ref:`ref_flyteidl.core.SecurityContext`", "", "Optional: security context override to apply this execution."
   "auth_role", ":ref:`ref_flyteidl.admin.AuthRole`", "", "**Deprecated.** Optional: auth override to apply this execution."
   "quality_of_service", ":ref:`ref_flyteidl.core.QualityOfService`", "", "Indicates the runtime priority of the execution."
   "max_parallelism", ":ref:`ref_int32`", "", "Controls the maximum number of task nodes that can be run in parallel for the entire workflow. This is useful to achieve fairness. Note: MapTasks are regarded as one unit, and parallelism/concurrency of MapTasks is independent from this."
   "raw_output_data_config", ":ref:`ref_flyteidl.admin.RawOutputDataConfig`", "", "User setting to configure where to store offloaded data (i.e. Blobs, structured datasets, query data, etc.). This should be a prefix like s3://my-bucket/my-data"
   "cluster_assignment", ":ref:`ref_flyteidl.admin.ClusterAssignment`", "", "Controls how to select an available cluster on which this execution should run."
   "interruptible", ":ref:`ref_google.protobuf.BoolValue`", "", "Allows for the interruptible flag of a workflow to be overwritten for a single execution. Omitting this field uses the workflow's value as a default. As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper around the bool field."
   "overwrite_cache", ":ref:`ref_bool`", "", "Allows for all cached values of a workflow and its tasks to be overwritten for a single execution. If enabled, all calculations are performed even if cached results would be available, overwriting the stored data once execution finishes successfully."







.. _ref_flyteidl.admin.ExecutionStateChangeDetails:

ExecutionStateChangeDetails
------------------------------------------------------------------





.. csv-table:: ExecutionStateChangeDetails type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "state", ":ref:`ref_flyteidl.admin.ExecutionState`", "", "The state of the execution is used to control its visibility in the UI/CLI."
   "occurred_at", ":ref:`ref_google.protobuf.Timestamp`", "", "This timestamp represents when the state changed."
   "principal", ":ref:`ref_string`", "", "Identifies the entity (if any) responsible for causing the state change of the execution"







.. _ref_flyteidl.admin.ExecutionTerminateRequest:

ExecutionTerminateRequest
------------------------------------------------------------------

Request to terminate an in-progress execution.  This action is irreversible.
If an execution is already terminated, this request will simply be a no-op.
This request will fail if it references a non-existent execution.
If the request succeeds the phase "ABORTED" will be recorded for the termination
with the optional cause added to the output_result.



.. csv-table:: ExecutionTerminateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Uniquely identifies the individual workflow execution to be terminated."
   "cause", ":ref:`ref_string`", "", "Optional reason for aborting."







.. _ref_flyteidl.admin.ExecutionTerminateResponse:

ExecutionTerminateResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.ExecutionUpdateRequest:

ExecutionUpdateRequest
------------------------------------------------------------------





.. csv-table:: ExecutionUpdateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Identifier of the execution to update"
   "state", ":ref:`ref_flyteidl.admin.ExecutionState`", "", "State to set as the new value active/archive"







.. _ref_flyteidl.admin.ExecutionUpdateResponse:

ExecutionUpdateResponse
------------------------------------------------------------------










.. _ref_flyteidl.admin.LiteralMapBlob:

LiteralMapBlob
------------------------------------------------------------------

Input/output data can represented by actual values or a link to where values are stored



.. csv-table:: LiteralMapBlob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "values", ":ref:`ref_flyteidl.core.LiteralMap`", "", "**Deprecated.** Data in LiteralMap format"
   "uri", ":ref:`ref_string`", "", "In the event that the map is too large, we return a uri to the data"







.. _ref_flyteidl.admin.NotificationList:

NotificationList
------------------------------------------------------------------





.. csv-table:: NotificationList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "notifications", ":ref:`ref_flyteidl.admin.Notification`", "repeated", ""







.. _ref_flyteidl.admin.SystemMetadata:

SystemMetadata
------------------------------------------------------------------

Represents system, rather than user-facing, metadata about an execution.



.. csv-table:: SystemMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "execution_cluster", ":ref:`ref_string`", "", "Which execution cluster this execution ran on."







.. _ref_flyteidl.admin.WorkflowExecutionGetDataRequest:

WorkflowExecutionGetDataRequest
------------------------------------------------------------------

Request structure to fetch inputs, output and other data produced by an execution.
By default this data is not returned inline in :ref:`ref_flyteidl.admin.WorkflowExecutionGetRequest`



.. csv-table:: WorkflowExecutionGetDataRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "The identifier of the execution for which to fetch inputs and outputs."







.. _ref_flyteidl.admin.WorkflowExecutionGetDataResponse:

WorkflowExecutionGetDataResponse
------------------------------------------------------------------

Response structure for WorkflowExecutionGetDataRequest which contains inputs and outputs for an execution.



.. csv-table:: WorkflowExecutionGetDataResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "outputs", ":ref:`ref_flyteidl.admin.UrlBlob`", "", "**Deprecated.** Signed url to fetch a core.LiteralMap of execution outputs. Deprecated: Please use full_outputs instead."
   "inputs", ":ref:`ref_flyteidl.admin.UrlBlob`", "", "**Deprecated.** Signed url to fetch a core.LiteralMap of execution inputs. Deprecated: Please use full_inputs instead."
   "full_inputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "Full_inputs will only be populated if they are under a configured size threshold."
   "full_outputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "Full_outputs will only be populated if they are under a configured size threshold."







.. _ref_flyteidl.admin.WorkflowExecutionGetRequest:

WorkflowExecutionGetRequest
------------------------------------------------------------------

A message used to fetch a single workflow execution entity.
See :ref:`ref_flyteidl.admin.Execution` for more details



.. csv-table:: WorkflowExecutionGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Uniquely identifies an individual workflow execution."






..
   end messages



.. _ref_flyteidl.admin.ExecutionMetadata.ExecutionMode:

ExecutionMetadata.ExecutionMode
------------------------------------------------------------------

The method by which this execution was launched.

.. csv-table:: Enum ExecutionMetadata.ExecutionMode values
   :header: "Name", "Number", "Description"
   :widths: auto

   "MANUAL", "0", "The default execution mode, MANUAL implies that an execution was launched by an individual."
   "SCHEDULED", "1", "A schedule triggered this execution launch."
   "SYSTEM", "2", "A system process was responsible for launching this execution rather an individual."
   "RELAUNCH", "3", "This execution was launched with identical inputs as a previous execution."
   "CHILD_WORKFLOW", "4", "This execution was triggered by another execution."
   "RECOVERED", "5", "This execution was recovered from another execution."



.. _ref_flyteidl.admin.ExecutionState:

ExecutionState
------------------------------------------------------------------

The state of the execution is used to control its visibility in the UI/CLI.

.. csv-table:: Enum ExecutionState values
   :header: "Name", "Number", "Description"
   :widths: auto

   "EXECUTION_ACTIVE", "0", "By default, all executions are considered active."
   "EXECUTION_ARCHIVED", "1", "Archived executions are no longer visible in the UI."


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/launch_plan.proto:

flyteidl/admin/launch_plan.proto
==================================================================





.. _ref_flyteidl.admin.ActiveLaunchPlanListRequest:

ActiveLaunchPlanListRequest
------------------------------------------------------------------

Represents a request structure to list active launch plans within a project/domain.
See :ref:`ref_flyteidl.admin.LaunchPlan` for more details



.. csv-table:: ActiveLaunchPlanListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Name of the project that contains the identifiers. +required."
   "domain", ":ref:`ref_string`", "", "Name of the domain the identifiers belongs to within the project. +required."
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned. +required."
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Sort ordering. +optional"







.. _ref_flyteidl.admin.ActiveLaunchPlanRequest:

ActiveLaunchPlanRequest
------------------------------------------------------------------

Represents a request struct for finding an active launch plan for a given NamedEntityIdentifier
See :ref:`ref_flyteidl.admin.LaunchPlan` for more details



.. csv-table:: ActiveLaunchPlanRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.admin.NamedEntityIdentifier`", "", "+required."







.. _ref_flyteidl.admin.Auth:

Auth
------------------------------------------------------------------

Defines permissions associated with executions created by this launch plan spec.
Use either of these roles when they have permissions required by your workflow execution.
Deprecated.



.. csv-table:: Auth type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "assumable_iam_role", ":ref:`ref_string`", "", "Defines an optional iam role which will be used for tasks run in executions created with this launch plan."
   "kubernetes_service_account", ":ref:`ref_string`", "", "Defines an optional kubernetes service account which will be used for tasks run in executions created with this launch plan."







.. _ref_flyteidl.admin.LaunchPlan:

LaunchPlan
------------------------------------------------------------------

A LaunchPlan provides the capability to templatize workflow executions.
Launch plans simplify associating one or more schedules, inputs and notifications with your workflows.
Launch plans can be shared and used to trigger executions with predefined inputs even when a workflow
definition doesn't necessarily have a default value for said input.



.. csv-table:: LaunchPlan type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "Uniquely identifies a launch plan entity."
   "spec", ":ref:`ref_flyteidl.admin.LaunchPlanSpec`", "", "User-provided launch plan details, including reference workflow, inputs and other metadata."
   "closure", ":ref:`ref_flyteidl.admin.LaunchPlanClosure`", "", "Values computed by the flyte platform after launch plan registration."







.. _ref_flyteidl.admin.LaunchPlanClosure:

LaunchPlanClosure
------------------------------------------------------------------

Values computed by the flyte platform after launch plan registration.
These include expected_inputs required to be present in a CreateExecutionRequest
to launch the reference workflow as well timestamp values associated with the launch plan.



.. csv-table:: LaunchPlanClosure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "state", ":ref:`ref_flyteidl.admin.LaunchPlanState`", "", "Indicate the Launch plan state."
   "expected_inputs", ":ref:`ref_flyteidl.core.ParameterMap`", "", "Indicates the set of inputs expected when creating an execution with the Launch plan"
   "expected_outputs", ":ref:`ref_flyteidl.core.VariableMap`", "", "Indicates the set of outputs expected to be produced by creating an execution with the Launch plan"
   "created_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the launch plan was created."
   "updated_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the launch plan was last updated."







.. _ref_flyteidl.admin.LaunchPlanCreateRequest:

LaunchPlanCreateRequest
------------------------------------------------------------------

Request to register a launch plan. The included LaunchPlanSpec may have a complete or incomplete set of inputs required
to launch a workflow execution. By default all launch plans are registered in state INACTIVE. If you wish to
set the state to ACTIVE, you must submit a LaunchPlanUpdateRequest, after you have successfully created a launch plan.



.. csv-table:: LaunchPlanCreateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "Uniquely identifies a launch plan entity."
   "spec", ":ref:`ref_flyteidl.admin.LaunchPlanSpec`", "", "User-provided launch plan details, including reference workflow, inputs and other metadata."







.. _ref_flyteidl.admin.LaunchPlanCreateResponse:

LaunchPlanCreateResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.LaunchPlanList:

LaunchPlanList
------------------------------------------------------------------

Response object for list launch plan requests.
See :ref:`ref_flyteidl.admin.LaunchPlan` for more details



.. csv-table:: LaunchPlanList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "launch_plans", ":ref:`ref_flyteidl.admin.LaunchPlan`", "repeated", ""
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.LaunchPlanMetadata:

LaunchPlanMetadata
------------------------------------------------------------------

Additional launch plan attributes included in the LaunchPlanSpec not strictly required to launch
the reference workflow.



.. csv-table:: LaunchPlanMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "schedule", ":ref:`ref_flyteidl.admin.Schedule`", "", "Schedule to execute the Launch Plan"
   "notifications", ":ref:`ref_flyteidl.admin.Notification`", "repeated", "List of notifications based on Execution status transitions"







.. _ref_flyteidl.admin.LaunchPlanSpec:

LaunchPlanSpec
------------------------------------------------------------------

User-provided launch plan definition and configuration values.



.. csv-table:: LaunchPlanSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "workflow_id", ":ref:`ref_flyteidl.core.Identifier`", "", "Reference to the Workflow template that the launch plan references"
   "entity_metadata", ":ref:`ref_flyteidl.admin.LaunchPlanMetadata`", "", "Metadata for the Launch Plan"
   "default_inputs", ":ref:`ref_flyteidl.core.ParameterMap`", "", "Input values to be passed for the execution. These can be overriden when an execution is created with this launch plan."
   "fixed_inputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "Fixed, non-overridable inputs for the Launch Plan. These can not be overriden when an execution is created with this launch plan."
   "role", ":ref:`ref_string`", "", "**Deprecated.** String to indicate the role to use to execute the workflow underneath"
   "labels", ":ref:`ref_flyteidl.admin.Labels`", "", "Custom labels to be applied to the execution resource."
   "annotations", ":ref:`ref_flyteidl.admin.Annotations`", "", "Custom annotations to be applied to the execution resource."
   "auth", ":ref:`ref_flyteidl.admin.Auth`", "", "**Deprecated.** Indicates the permission associated with workflow executions triggered with this launch plan."
   "auth_role", ":ref:`ref_flyteidl.admin.AuthRole`", "", "**Deprecated.** "
   "security_context", ":ref:`ref_flyteidl.core.SecurityContext`", "", "Indicates security context for permissions triggered with this launch plan"
   "quality_of_service", ":ref:`ref_flyteidl.core.QualityOfService`", "", "Indicates the runtime priority of the execution."
   "raw_output_data_config", ":ref:`ref_flyteidl.admin.RawOutputDataConfig`", "", "Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.)."
   "max_parallelism", ":ref:`ref_int32`", "", "Controls the maximum number of tasknodes that can be run in parallel for the entire workflow. This is useful to achieve fairness. Note: MapTasks are regarded as one unit, and parallelism/concurrency of MapTasks is independent from this."
   "interruptible", ":ref:`ref_google.protobuf.BoolValue`", "", "Allows for the interruptible flag of a workflow to be overwritten for a single execution. Omitting this field uses the workflow's value as a default. As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper around the bool field."
   "overwrite_cache", ":ref:`ref_bool`", "", "Allows for all cached values of a workflow and its tasks to be overwritten for a single execution. If enabled, all calculations are performed even if cached results would be available, overwriting the stored data once execution finishes successfully."







.. _ref_flyteidl.admin.LaunchPlanUpdateRequest:

LaunchPlanUpdateRequest
------------------------------------------------------------------

Request to set the referenced launch plan state to the configured value.
See :ref:`ref_flyteidl.admin.LaunchPlan` for more details



.. csv-table:: LaunchPlanUpdateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "Identifier of launch plan for which to change state. +required."
   "state", ":ref:`ref_flyteidl.admin.LaunchPlanState`", "", "Desired state to apply to the launch plan. +required."







.. _ref_flyteidl.admin.LaunchPlanUpdateResponse:

LaunchPlanUpdateResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.







..
   end messages



.. _ref_flyteidl.admin.LaunchPlanState:

LaunchPlanState
------------------------------------------------------------------

By default any launch plan regardless of state can be used to launch a workflow execution.
However, at most one version of a launch plan
(e.g. a NamedEntityIdentifier set of shared project, domain and name values) can be
active at a time in regards to *schedules*. That is, at most one schedule in a NamedEntityIdentifier
group will be observed and trigger executions at a defined cadence.

.. csv-table:: Enum LaunchPlanState values
   :header: "Name", "Number", "Description"
   :widths: auto

   "INACTIVE", "0", ""
   "ACTIVE", "1", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/matchable_resource.proto:

flyteidl/admin/matchable_resource.proto
==================================================================





.. _ref_flyteidl.admin.ClusterResourceAttributes:

ClusterResourceAttributes
------------------------------------------------------------------





.. csv-table:: ClusterResourceAttributes type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "attributes", ":ref:`ref_flyteidl.admin.ClusterResourceAttributes.AttributesEntry`", "repeated", "Custom resource attributes which will be applied in cluster resource creation (e.g. quotas). Map keys are the *case-sensitive* names of variables in templatized resource files. Map values should be the custom values which get substituted during resource creation."







.. _ref_flyteidl.admin.ClusterResourceAttributes.AttributesEntry:

ClusterResourceAttributes.AttributesEntry
------------------------------------------------------------------





.. csv-table:: ClusterResourceAttributes.AttributesEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.admin.ExecutionClusterLabel:

ExecutionClusterLabel
------------------------------------------------------------------





.. csv-table:: ExecutionClusterLabel type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_string`", "", "Label value to determine where the execution will be run"







.. _ref_flyteidl.admin.ExecutionQueueAttributes:

ExecutionQueueAttributes
------------------------------------------------------------------





.. csv-table:: ExecutionQueueAttributes type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "tags", ":ref:`ref_string`", "repeated", "Tags used for assigning execution queues for tasks defined within this project."







.. _ref_flyteidl.admin.ListMatchableAttributesRequest:

ListMatchableAttributesRequest
------------------------------------------------------------------

Request all matching resource attributes for a resource type.
See :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for more details



.. csv-table:: ListMatchableAttributesRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "resource_type", ":ref:`ref_flyteidl.admin.MatchableResource`", "", "+required"







.. _ref_flyteidl.admin.ListMatchableAttributesResponse:

ListMatchableAttributesResponse
------------------------------------------------------------------

Response for a request for all matching resource attributes for a resource type.
See :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for more details



.. csv-table:: ListMatchableAttributesResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "configurations", ":ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`", "repeated", ""







.. _ref_flyteidl.admin.MatchableAttributesConfiguration:

MatchableAttributesConfiguration
------------------------------------------------------------------

Represents a custom set of attributes applied for either a domain; a domain and project; or
domain, project and workflow name.
These are used to override system level defaults for kubernetes cluster resource management,
default execution values, and more all across different levels of specificity.



.. csv-table:: MatchableAttributesConfiguration type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "attributes", ":ref:`ref_flyteidl.admin.MatchingAttributes`", "", ""
   "domain", ":ref:`ref_string`", "", ""
   "project", ":ref:`ref_string`", "", ""
   "workflow", ":ref:`ref_string`", "", ""
   "launch_plan", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.admin.MatchingAttributes:

MatchingAttributes
------------------------------------------------------------------

Generic container for encapsulating all types of the above attributes messages.



.. csv-table:: MatchingAttributes type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "task_resource_attributes", ":ref:`ref_flyteidl.admin.TaskResourceAttributes`", "", ""
   "cluster_resource_attributes", ":ref:`ref_flyteidl.admin.ClusterResourceAttributes`", "", ""
   "execution_queue_attributes", ":ref:`ref_flyteidl.admin.ExecutionQueueAttributes`", "", ""
   "execution_cluster_label", ":ref:`ref_flyteidl.admin.ExecutionClusterLabel`", "", ""
   "quality_of_service", ":ref:`ref_flyteidl.core.QualityOfService`", "", ""
   "plugin_overrides", ":ref:`ref_flyteidl.admin.PluginOverrides`", "", ""
   "workflow_execution_config", ":ref:`ref_flyteidl.admin.WorkflowExecutionConfig`", "", ""
   "cluster_assignment", ":ref:`ref_flyteidl.admin.ClusterAssignment`", "", ""







.. _ref_flyteidl.admin.PluginOverride:

PluginOverride
------------------------------------------------------------------

This MatchableAttribute configures selecting alternate plugin implementations for a given task type.
In addition to an override implementation a selection of fallbacks can be provided or other modes
for handling cases where the desired plugin override is not enabled in a given Flyte deployment.



.. csv-table:: PluginOverride type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "task_type", ":ref:`ref_string`", "", "A predefined yet extensible Task type identifier."
   "plugin_id", ":ref:`ref_string`", "repeated", "A set of plugin ids which should handle tasks of this type instead of the default registered plugin. The list will be tried in order until a plugin is found with that id."
   "missing_plugin_behavior", ":ref:`ref_flyteidl.admin.PluginOverride.MissingPluginBehavior`", "", "Defines the behavior when no plugin from the plugin_id list is not found."







.. _ref_flyteidl.admin.PluginOverrides:

PluginOverrides
------------------------------------------------------------------





.. csv-table:: PluginOverrides type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "overrides", ":ref:`ref_flyteidl.admin.PluginOverride`", "repeated", ""







.. _ref_flyteidl.admin.TaskResourceAttributes:

TaskResourceAttributes
------------------------------------------------------------------

Defines task resource defaults and limits that will be applied at task registration.



.. csv-table:: TaskResourceAttributes type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "defaults", ":ref:`ref_flyteidl.admin.TaskResourceSpec`", "", ""
   "limits", ":ref:`ref_flyteidl.admin.TaskResourceSpec`", "", ""







.. _ref_flyteidl.admin.TaskResourceSpec:

TaskResourceSpec
------------------------------------------------------------------

Defines a set of overridable task resource attributes set during task registration.



.. csv-table:: TaskResourceSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "cpu", ":ref:`ref_string`", "", ""
   "gpu", ":ref:`ref_string`", "", ""
   "memory", ":ref:`ref_string`", "", ""
   "storage", ":ref:`ref_string`", "", ""
   "ephemeral_storage", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.admin.WorkflowExecutionConfig:

WorkflowExecutionConfig
------------------------------------------------------------------

Adds defaults for customizable workflow-execution specifications and overrides.



.. csv-table:: WorkflowExecutionConfig type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "max_parallelism", ":ref:`ref_int32`", "", "Can be used to control the number of parallel nodes to run within the workflow. This is useful to achieve fairness."
   "security_context", ":ref:`ref_flyteidl.core.SecurityContext`", "", "Indicates security context permissions for executions triggered with this matchable attribute."
   "raw_output_data_config", ":ref:`ref_flyteidl.admin.RawOutputDataConfig`", "", "Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.)."
   "labels", ":ref:`ref_flyteidl.admin.Labels`", "", "Custom labels to be applied to a triggered execution resource."
   "annotations", ":ref:`ref_flyteidl.admin.Annotations`", "", "Custom annotations to be applied to a triggered execution resource."
   "interruptible", ":ref:`ref_google.protobuf.BoolValue`", "", "Allows for the interruptible flag of a workflow to be overwritten for a single execution. Omitting this field uses the workflow's value as a default. As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper around the bool field."
   "overwrite_cache", ":ref:`ref_bool`", "", "Allows for all cached values of a workflow and its tasks to be overwritten for a single execution. If enabled, all calculations are performed even if cached results would be available, overwriting the stored data once execution finishes successfully."






..
   end messages



.. _ref_flyteidl.admin.MatchableResource:

MatchableResource
------------------------------------------------------------------

Defines a resource that can be configured by customizable Project-, ProjectDomain- or WorkflowAttributes
based on matching tags.

.. csv-table:: Enum MatchableResource values
   :header: "Name", "Number", "Description"
   :widths: auto

   "TASK_RESOURCE", "0", "Applies to customizable task resource requests and limits."
   "CLUSTER_RESOURCE", "1", "Applies to configuring templated kubernetes cluster resources."
   "EXECUTION_QUEUE", "2", "Configures task and dynamic task execution queue assignment."
   "EXECUTION_CLUSTER_LABEL", "3", "Configures the K8s cluster label to be used for execution to be run"
   "QUALITY_OF_SERVICE_SPECIFICATION", "4", "Configures default quality of service when undefined in an execution spec."
   "PLUGIN_OVERRIDE", "5", "Selects configurable plugin implementation behavior for a given task type."
   "WORKFLOW_EXECUTION_CONFIG", "6", "Adds defaults for customizable workflow-execution specifications and overrides."
   "CLUSTER_ASSIGNMENT", "7", "Controls how to select an available cluster on which this execution should run."



.. _ref_flyteidl.admin.PluginOverride.MissingPluginBehavior:

PluginOverride.MissingPluginBehavior
------------------------------------------------------------------



.. csv-table:: Enum PluginOverride.MissingPluginBehavior values
   :header: "Name", "Number", "Description"
   :widths: auto

   "FAIL", "0", "By default, if this plugin is not enabled for a Flyte deployment then execution will fail."
   "USE_DEFAULT", "1", "Uses the system-configured default implementation."


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/node_execution.proto:

flyteidl/admin/node_execution.proto
==================================================================





.. _ref_flyteidl.admin.DynamicWorkflowNodeMetadata:

DynamicWorkflowNodeMetadata
------------------------------------------------------------------

For dynamic workflow nodes we capture information about the dynamic workflow definition that gets generated.



.. csv-table:: DynamicWorkflowNodeMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "id represents the unique identifier of the workflow."
   "compiled_workflow", ":ref:`ref_flyteidl.core.CompiledWorkflowClosure`", "", "Represents the compiled representation of the embedded dynamic workflow."







.. _ref_flyteidl.admin.NodeExecution:

NodeExecution
------------------------------------------------------------------

Encapsulates all details for a single node execution entity.
A node represents a component in the overall workflow graph. A node launch a task, multiple tasks, an entire nested
sub-workflow, or even a separate child-workflow execution.
The same task can be called repeatedly in a single workflow but each node is unique.



.. csv-table:: NodeExecution type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.NodeExecutionIdentifier`", "", "Uniquely identifies an individual node execution."
   "input_uri", ":ref:`ref_string`", "", "Path to remote data store where input blob is stored."
   "closure", ":ref:`ref_flyteidl.admin.NodeExecutionClosure`", "", "Computed results associated with this node execution."
   "metadata", ":ref:`ref_flyteidl.admin.NodeExecutionMetaData`", "", "Metadata for Node Execution"







.. _ref_flyteidl.admin.NodeExecutionClosure:

NodeExecutionClosure
------------------------------------------------------------------

Container for node execution details and results.



.. csv-table:: NodeExecutionClosure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "output_uri", ":ref:`ref_string`", "", "**Deprecated.** Links to a remotely stored, serialized core.LiteralMap of node execution outputs. DEPRECATED. Use GetNodeExecutionData to fetch output data instead."
   "error", ":ref:`ref_flyteidl.core.ExecutionError`", "", "Error information for the Node"
   "output_data", ":ref:`ref_flyteidl.core.LiteralMap`", "", "**Deprecated.** Raw output data produced by this node execution. DEPRECATED. Use GetNodeExecutionData to fetch output data instead."
   "phase", ":ref:`ref_flyteidl.core.NodeExecution.Phase`", "", "The last recorded phase for this node execution."
   "started_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the node execution began running."
   "duration", ":ref:`ref_google.protobuf.Duration`", "", "The amount of time the node execution spent running."
   "created_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the node execution was created."
   "updated_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the node execution was last updated."
   "workflow_node_metadata", ":ref:`ref_flyteidl.admin.WorkflowNodeMetadata`", "", ""
   "task_node_metadata", ":ref:`ref_flyteidl.admin.TaskNodeMetadata`", "", ""
   "deck_uri", ":ref:`ref_string`", "", "String location uniquely identifying where the deck HTML file is. NativeUrl specifies the url in the format of the configured storage provider (e.g. s3://my-bucket/randomstring/suffix.tar)"







.. _ref_flyteidl.admin.NodeExecutionForTaskListRequest:

NodeExecutionForTaskListRequest
------------------------------------------------------------------

Represents a request structure to retrieve a list of node execution entities launched by a specific task.
This can arise when a task yields a subworkflow.



.. csv-table:: NodeExecutionForTaskListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "task_execution_id", ":ref:`ref_flyteidl.core.TaskExecutionIdentifier`", "", "Indicates the node execution to filter by. +required"
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned. +required"
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the, server-provided token can be used to fetch the next page in a query. +optional"
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. More info on constructing filters : <Link> +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Sort ordering. +optional"







.. _ref_flyteidl.admin.NodeExecutionGetDataRequest:

NodeExecutionGetDataRequest
------------------------------------------------------------------

Request structure to fetch inputs and output for a node execution.
By default, these are not returned in :ref:`ref_flyteidl.admin.NodeExecutionGetRequest`



.. csv-table:: NodeExecutionGetDataRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.NodeExecutionIdentifier`", "", "The identifier of the node execution for which to fetch inputs and outputs."







.. _ref_flyteidl.admin.NodeExecutionGetDataResponse:

NodeExecutionGetDataResponse
------------------------------------------------------------------

Response structure for NodeExecutionGetDataRequest which contains inputs and outputs for a node execution.



.. csv-table:: NodeExecutionGetDataResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "inputs", ":ref:`ref_flyteidl.admin.UrlBlob`", "", "**Deprecated.** Signed url to fetch a core.LiteralMap of node execution inputs. Deprecated: Please use full_inputs instead."
   "outputs", ":ref:`ref_flyteidl.admin.UrlBlob`", "", "**Deprecated.** Signed url to fetch a core.LiteralMap of node execution outputs. Deprecated: Please use full_outputs instead."
   "full_inputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "Full_inputs will only be populated if they are under a configured size threshold."
   "full_outputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "Full_outputs will only be populated if they are under a configured size threshold."
   "dynamic_workflow", ":ref:`ref_flyteidl.admin.DynamicWorkflowNodeMetadata`", "", "Optional Workflow closure for a dynamically generated workflow, in the case this node yields a dynamic workflow we return its structure here."







.. _ref_flyteidl.admin.NodeExecutionGetRequest:

NodeExecutionGetRequest
------------------------------------------------------------------

A message used to fetch a single node execution entity.
See :ref:`ref_flyteidl.admin.NodeExecution` for more details



.. csv-table:: NodeExecutionGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.NodeExecutionIdentifier`", "", "Uniquely identifies an individual node execution. +required"







.. _ref_flyteidl.admin.NodeExecutionList:

NodeExecutionList
------------------------------------------------------------------

Request structure to retrieve a list of node execution entities.
See :ref:`ref_flyteidl.admin.NodeExecution` for more details



.. csv-table:: NodeExecutionList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "node_executions", ":ref:`ref_flyteidl.admin.NodeExecution`", "repeated", ""
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.NodeExecutionListRequest:

NodeExecutionListRequest
------------------------------------------------------------------

Represents a request structure to retrieve a list of node execution entities.
See :ref:`ref_flyteidl.admin.NodeExecution` for more details



.. csv-table:: NodeExecutionListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "workflow_execution_id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Indicates the workflow execution to filter by. +required"
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned. +required"
   "token", ":ref:`ref_string`", "", ""
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. More info on constructing filters : <Link> +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Sort ordering. +optional"
   "unique_parent_id", ":ref:`ref_string`", "", "Unique identifier of the parent node in the execution +optional"







.. _ref_flyteidl.admin.NodeExecutionMetaData:

NodeExecutionMetaData
------------------------------------------------------------------

Represents additional attributes related to a Node Execution



.. csv-table:: NodeExecutionMetaData type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "retry_group", ":ref:`ref_string`", "", "Node executions are grouped depending on retries of the parent Retry group is unique within the context of a parent node."
   "is_parent_node", ":ref:`ref_bool`", "", "Boolean flag indicating if the node has child nodes under it This can be true when a node contains a dynamic workflow which then produces child nodes."
   "spec_node_id", ":ref:`ref_string`", "", "Node id of the node in the original workflow This maps to value of WorkflowTemplate.nodes[X].id"
   "is_dynamic", ":ref:`ref_bool`", "", "Boolean flag indicating if the node has contains a dynamic workflow which then produces child nodes. This is to distinguish between subworkflows and dynamic workflows which can both have is_parent_node as true."







.. _ref_flyteidl.admin.TaskNodeMetadata:

TaskNodeMetadata
------------------------------------------------------------------

Metadata for the case in which the node is a TaskNode



.. csv-table:: TaskNodeMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "cache_status", ":ref:`ref_flyteidl.core.CatalogCacheStatus`", "", "Captures the status of caching for this execution."
   "catalog_key", ":ref:`ref_flyteidl.core.CatalogMetadata`", "", "This structure carries the catalog artifact information"
   "checkpoint_uri", ":ref:`ref_string`", "", "The latest checkpoint location"







.. _ref_flyteidl.admin.WorkflowNodeMetadata:

WorkflowNodeMetadata
------------------------------------------------------------------

Metadata for a WorkflowNode



.. csv-table:: WorkflowNodeMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "executionId", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "The identifier for a workflow execution launched by a node."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/notification.proto:

flyteidl/admin/notification.proto
==================================================================





.. _ref_flyteidl.admin.EmailMessage:

EmailMessage
------------------------------------------------------------------

Represents the Email object that is sent to a publisher/subscriber
to forward the notification.
Note: This is internal to Admin and doesn't need to be exposed to other components.



.. csv-table:: EmailMessage type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "recipients_email", ":ref:`ref_string`", "repeated", "The list of email addresses to receive an email with the content populated in the other fields. Currently, each email recipient will receive its own email. This populates the TO field."
   "sender_email", ":ref:`ref_string`", "", "The email of the sender. This populates the FROM field."
   "subject_line", ":ref:`ref_string`", "", "The content of the subject line. This populates the SUBJECT field."
   "body", ":ref:`ref_string`", "", "The content of the email body. This populates the BODY field."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/project.proto:

flyteidl/admin/project.proto
==================================================================





.. _ref_flyteidl.admin.Domain:

Domain
------------------------------------------------------------------

Namespace within a project commonly used to differentiate between different service instances.
e.g. "production", "development", etc.



.. csv-table:: Domain type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_string`", "", "Globally unique domain name."
   "name", ":ref:`ref_string`", "", "Display name."







.. _ref_flyteidl.admin.Project:

Project
------------------------------------------------------------------

Top-level namespace used to classify different entities like workflows and executions.



.. csv-table:: Project type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_string`", "", "Globally unique project name."
   "name", ":ref:`ref_string`", "", "Display name."
   "domains", ":ref:`ref_flyteidl.admin.Domain`", "repeated", ""
   "description", ":ref:`ref_string`", "", ""
   "labels", ":ref:`ref_flyteidl.admin.Labels`", "", "Leverage Labels from flyteidl.admin.common.proto to tag projects with ownership information."
   "state", ":ref:`ref_flyteidl.admin.Project.ProjectState`", "", ""







.. _ref_flyteidl.admin.ProjectListRequest:

ProjectListRequest
------------------------------------------------------------------

Request to retrieve a list of projects matching specified filters. 
See :ref:`ref_flyteidl.admin.Project` for more details



.. csv-table:: ProjectListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "limit", ":ref:`ref_uint32`", "", "Indicates the number of projects to be returned. +required"
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, this server-provided token can be used to fetch the next page in a query. +optional"
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. More info on constructing filters : <Link> +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Sort ordering. +optional"







.. _ref_flyteidl.admin.ProjectRegisterRequest:

ProjectRegisterRequest
------------------------------------------------------------------

Adds a new user-project within the Flyte deployment.
See :ref:`ref_flyteidl.admin.Project` for more details



.. csv-table:: ProjectRegisterRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_flyteidl.admin.Project`", "", "+required"







.. _ref_flyteidl.admin.ProjectRegisterResponse:

ProjectRegisterResponse
------------------------------------------------------------------

Purposefully empty, may be updated in the future.








.. _ref_flyteidl.admin.ProjectUpdateResponse:

ProjectUpdateResponse
------------------------------------------------------------------

Purposefully empty, may be updated in the future.








.. _ref_flyteidl.admin.Projects:

Projects
------------------------------------------------------------------

Represents a list of projects.
See :ref:`ref_flyteidl.admin.Project` for more details



.. csv-table:: Projects type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "projects", ":ref:`ref_flyteidl.admin.Project`", "repeated", ""
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."






..
   end messages



.. _ref_flyteidl.admin.Project.ProjectState:

Project.ProjectState
------------------------------------------------------------------

The state of the project is used to control its visibility in the UI and validity.

.. csv-table:: Enum Project.ProjectState values
   :header: "Name", "Number", "Description"
   :widths: auto

   "ACTIVE", "0", "By default, all projects are considered active."
   "ARCHIVED", "1", "Archived projects are no longer visible in the UI and no longer valid."
   "SYSTEM_GENERATED", "2", "System generated projects that aren't explicitly created or managed by a user."


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/project_attributes.proto:

flyteidl/admin/project_attributes.proto
==================================================================





.. _ref_flyteidl.admin.ProjectAttributes:

ProjectAttributes
------------------------------------------------------------------

Defines a set of custom matching attributes at the project level.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectAttributes type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id for which this set of attributes will be applied."
   "matching_attributes", ":ref:`ref_flyteidl.admin.MatchingAttributes`", "", ""







.. _ref_flyteidl.admin.ProjectAttributesDeleteRequest:

ProjectAttributesDeleteRequest
------------------------------------------------------------------

Request to delete a set matchable project level attribute override.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectAttributesDeleteRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id which this set of attributes references. +required"
   "resource_type", ":ref:`ref_flyteidl.admin.MatchableResource`", "", "Which type of matchable attributes to delete. +required"







.. _ref_flyteidl.admin.ProjectAttributesDeleteResponse:

ProjectAttributesDeleteResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.ProjectAttributesGetRequest:

ProjectAttributesGetRequest
------------------------------------------------------------------

Request to get an individual project level attribute override.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectAttributesGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id which this set of attributes references. +required"
   "resource_type", ":ref:`ref_flyteidl.admin.MatchableResource`", "", "Which type of matchable attributes to return. +required"







.. _ref_flyteidl.admin.ProjectAttributesGetResponse:

ProjectAttributesGetResponse
------------------------------------------------------------------

Response to get an individual project level attribute override.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectAttributesGetResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "attributes", ":ref:`ref_flyteidl.admin.ProjectAttributes`", "", ""







.. _ref_flyteidl.admin.ProjectAttributesUpdateRequest:

ProjectAttributesUpdateRequest
------------------------------------------------------------------

Sets custom attributes for a project
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectAttributesUpdateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "attributes", ":ref:`ref_flyteidl.admin.ProjectAttributes`", "", "+required"







.. _ref_flyteidl.admin.ProjectAttributesUpdateResponse:

ProjectAttributesUpdateResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.







..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/project_domain_attributes.proto:

flyteidl/admin/project_domain_attributes.proto
==================================================================





.. _ref_flyteidl.admin.ProjectDomainAttributes:

ProjectDomainAttributes
------------------------------------------------------------------

Defines a set of custom matching attributes which defines resource defaults for a project and domain.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectDomainAttributes type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id for which this set of attributes will be applied."
   "domain", ":ref:`ref_string`", "", "Unique domain id for which this set of attributes will be applied."
   "matching_attributes", ":ref:`ref_flyteidl.admin.MatchingAttributes`", "", ""







.. _ref_flyteidl.admin.ProjectDomainAttributesDeleteRequest:

ProjectDomainAttributesDeleteRequest
------------------------------------------------------------------

Request to delete a set matchable project domain attribute override.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectDomainAttributesDeleteRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id which this set of attributes references. +required"
   "domain", ":ref:`ref_string`", "", "Unique domain id which this set of attributes references. +required"
   "resource_type", ":ref:`ref_flyteidl.admin.MatchableResource`", "", "Which type of matchable attributes to delete. +required"







.. _ref_flyteidl.admin.ProjectDomainAttributesDeleteResponse:

ProjectDomainAttributesDeleteResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.ProjectDomainAttributesGetRequest:

ProjectDomainAttributesGetRequest
------------------------------------------------------------------

Request to get an individual project domain attribute override.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectDomainAttributesGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id which this set of attributes references. +required"
   "domain", ":ref:`ref_string`", "", "Unique domain id which this set of attributes references. +required"
   "resource_type", ":ref:`ref_flyteidl.admin.MatchableResource`", "", "Which type of matchable attributes to return. +required"







.. _ref_flyteidl.admin.ProjectDomainAttributesGetResponse:

ProjectDomainAttributesGetResponse
------------------------------------------------------------------

Response to get an individual project domain attribute override.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectDomainAttributesGetResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "attributes", ":ref:`ref_flyteidl.admin.ProjectDomainAttributes`", "", ""







.. _ref_flyteidl.admin.ProjectDomainAttributesUpdateRequest:

ProjectDomainAttributesUpdateRequest
------------------------------------------------------------------

Sets custom attributes for a project-domain combination.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: ProjectDomainAttributesUpdateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "attributes", ":ref:`ref_flyteidl.admin.ProjectDomainAttributes`", "", "+required"







.. _ref_flyteidl.admin.ProjectDomainAttributesUpdateResponse:

ProjectDomainAttributesUpdateResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.







..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/schedule.proto:

flyteidl/admin/schedule.proto
==================================================================





.. _ref_flyteidl.admin.CronSchedule:

CronSchedule
------------------------------------------------------------------

Options for schedules to run according to a cron expression.



.. csv-table:: CronSchedule type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "schedule", ":ref:`ref_string`", "", "Standard/default cron implementation as described by https://en.wikipedia.org/wiki/Cron#CRON_expression; Also supports nonstandard predefined scheduling definitions as described by https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions except @reboot"
   "offset", ":ref:`ref_string`", "", "ISO 8601 duration as described by https://en.wikipedia.org/wiki/ISO_8601#Durations"







.. _ref_flyteidl.admin.FixedRate:

FixedRate
------------------------------------------------------------------

Option for schedules run at a certain frequency e.g. every 2 minutes.



.. csv-table:: FixedRate type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_uint32`", "", ""
   "unit", ":ref:`ref_flyteidl.admin.FixedRateUnit`", "", ""







.. _ref_flyteidl.admin.Schedule:

Schedule
------------------------------------------------------------------

Defines complete set of information required to trigger an execution on a schedule.



.. csv-table:: Schedule type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "cron_expression", ":ref:`ref_string`", "", "**Deprecated.** Uses AWS syntax: Minutes Hours Day-of-month Month Day-of-week Year e.g. for a schedule that runs every 15 minutes: 0/15 * * * ? *"
   "rate", ":ref:`ref_flyteidl.admin.FixedRate`", "", ""
   "cron_schedule", ":ref:`ref_flyteidl.admin.CronSchedule`", "", ""
   "kickoff_time_input_arg", ":ref:`ref_string`", "", "Name of the input variable that the kickoff time will be supplied to when the workflow is kicked off."






..
   end messages



.. _ref_flyteidl.admin.FixedRateUnit:

FixedRateUnit
------------------------------------------------------------------

Represents a frequency at which to run a schedule.

.. csv-table:: Enum FixedRateUnit values
   :header: "Name", "Number", "Description"
   :widths: auto

   "MINUTE", "0", ""
   "HOUR", "1", ""
   "DAY", "2", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/signal.proto:

flyteidl/admin/signal.proto
==================================================================





.. _ref_flyteidl.admin.Signal:

Signal
------------------------------------------------------------------

Signal encapsulates a unique identifier, associated metadata, and a value for a single Flyte
signal. Signals may exist either without a set value (representing a signal request) or with a
populated value (indicating the signal has been given).



.. csv-table:: Signal type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.SignalIdentifier`", "", "A unique identifier for the requested signal."
   "type", ":ref:`ref_flyteidl.core.LiteralType`", "", "A type denoting the required value type for this signal."
   "value", ":ref:`ref_flyteidl.core.Literal`", "", "The value of the signal. This is only available if the signal has been "set" and must match the defined the type."







.. _ref_flyteidl.admin.SignalGetOrCreateRequest:

SignalGetOrCreateRequest
------------------------------------------------------------------

SignalGetOrCreateRequest represents a request structure to retrive or create a signal.
See :ref:`ref_flyteidl.admin.Signal` for more details



.. csv-table:: SignalGetOrCreateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.SignalIdentifier`", "", "A unique identifier for the requested signal."
   "type", ":ref:`ref_flyteidl.core.LiteralType`", "", "A type denoting the required value type for this signal."







.. _ref_flyteidl.admin.SignalList:

SignalList
------------------------------------------------------------------

SignalList represents collection of signals along with the token of the last result.
See :ref:`ref_flyteidl.admin.Signal` for more details



.. csv-table:: SignalList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signals", ":ref:`ref_flyteidl.admin.Signal`", "repeated", "A list of signals matching the input filters."
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.SignalListRequest:

SignalListRequest
------------------------------------------------------------------

SignalListRequest represents a request structure to retrieve a collection of signals.
See :ref:`ref_flyteidl.admin.Signal` for more details



.. csv-table:: SignalListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "workflow_execution_id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Indicates the workflow execution to filter by. +required"
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned. +required"
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the, server-provided token can be used to fetch the next page in a query. +optional"
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Sort ordering. +optional"







.. _ref_flyteidl.admin.SignalSetRequest:

SignalSetRequest
------------------------------------------------------------------

SignalSetRequest represents a request structure to set the value on a signal. Setting a signal
effetively satisfies the signal condition within a Flyte workflow.
See :ref:`ref_flyteidl.admin.Signal` for more details



.. csv-table:: SignalSetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.SignalIdentifier`", "", "A unique identifier for the requested signal."
   "value", ":ref:`ref_flyteidl.core.Literal`", "", "The value of this signal, must match the defining signal type."







.. _ref_flyteidl.admin.SignalSetResponse:

SignalSetResponse
------------------------------------------------------------------

SignalSetResponse represents a response structure if signal setting succeeds.

Purposefully empty, may be populated in the future.







..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/task.proto:

flyteidl/admin/task.proto
==================================================================





.. _ref_flyteidl.admin.Task:

Task
------------------------------------------------------------------

Flyte workflows are composed of many ordered tasks. That is small, reusable, self-contained logical blocks
arranged to process workflow inputs and produce a deterministic set of outputs.
Tasks can come in many varieties tuned for specialized behavior.



.. csv-table:: Task type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "id represents the unique identifier of the task."
   "closure", ":ref:`ref_flyteidl.admin.TaskClosure`", "", "closure encapsulates all the fields that maps to a compiled version of the task."
   "short_description", ":ref:`ref_string`", "", "One-liner overview of the entity."







.. _ref_flyteidl.admin.TaskClosure:

TaskClosure
------------------------------------------------------------------

Compute task attributes which include values derived from the TaskSpec, as well as plugin-specific data
and task metadata.



.. csv-table:: TaskClosure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "compiled_task", ":ref:`ref_flyteidl.core.CompiledTask`", "", "Represents the compiled representation of the task from the specification provided."
   "created_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the task was created."







.. _ref_flyteidl.admin.TaskCreateRequest:

TaskCreateRequest
------------------------------------------------------------------

Represents a request structure to create a revision of a task.
See :ref:`ref_flyteidl.admin.Task` for more details



.. csv-table:: TaskCreateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "id represents the unique identifier of the task. +required"
   "spec", ":ref:`ref_flyteidl.admin.TaskSpec`", "", "Represents the specification for task. +required"







.. _ref_flyteidl.admin.TaskCreateResponse:

TaskCreateResponse
------------------------------------------------------------------

Represents a response structure if task creation succeeds.

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.TaskList:

TaskList
------------------------------------------------------------------

Represents a list of tasks returned from the admin.
See :ref:`ref_flyteidl.admin.Task` for more details



.. csv-table:: TaskList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "tasks", ":ref:`ref_flyteidl.admin.Task`", "repeated", "A list of tasks returned based on the request."
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.TaskSpec:

TaskSpec
------------------------------------------------------------------

Represents a structure that encapsulates the user-configured specification of the task.



.. csv-table:: TaskSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "template", ":ref:`ref_flyteidl.core.TaskTemplate`", "", "Template of the task that encapsulates all the metadata of the task."
   "description", ":ref:`ref_flyteidl.admin.DescriptionEntity`", "", "Represents the specification for description entity."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/task_execution.proto:

flyteidl/admin/task_execution.proto
==================================================================





.. _ref_flyteidl.admin.TaskExecution:

TaskExecution
------------------------------------------------------------------

Encapsulates all details for a single task execution entity.
A task execution represents an instantiated task, including all inputs and additional
metadata as well as computed results included state, outputs, and duration-based attributes.



.. csv-table:: TaskExecution type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.TaskExecutionIdentifier`", "", "Unique identifier for the task execution."
   "input_uri", ":ref:`ref_string`", "", "Path to remote data store where input blob is stored."
   "closure", ":ref:`ref_flyteidl.admin.TaskExecutionClosure`", "", "Task execution details and results."
   "is_parent", ":ref:`ref_bool`", "", "Whether this task spawned nodes."







.. _ref_flyteidl.admin.TaskExecutionClosure:

TaskExecutionClosure
------------------------------------------------------------------

Container for task execution details and results.



.. csv-table:: TaskExecutionClosure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "output_uri", ":ref:`ref_string`", "", "**Deprecated.** Path to remote data store where output blob is stored if the execution succeeded (and produced outputs). DEPRECATED. Use GetTaskExecutionData to fetch output data instead."
   "error", ":ref:`ref_flyteidl.core.ExecutionError`", "", "Error information for the task execution. Populated if the execution failed."
   "output_data", ":ref:`ref_flyteidl.core.LiteralMap`", "", "**Deprecated.** Raw output data produced by this task execution. DEPRECATED. Use GetTaskExecutionData to fetch output data instead."
   "phase", ":ref:`ref_flyteidl.core.TaskExecution.Phase`", "", "The last recorded phase for this task execution."
   "logs", ":ref:`ref_flyteidl.core.TaskLog`", "repeated", "Detailed log information output by the task execution."
   "started_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the task execution began running."
   "duration", ":ref:`ref_google.protobuf.Duration`", "", "The amount of time the task execution spent running."
   "created_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the task execution was created."
   "updated_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the task execution was last updated."
   "custom_info", ":ref:`ref_google.protobuf.Struct`", "", "Custom data specific to the task plugin."
   "reason", ":ref:`ref_string`", "", "If there is an explanation for the most recent phase transition, the reason will capture it."
   "task_type", ":ref:`ref_string`", "", "A predefined yet extensible Task type identifier."
   "metadata", ":ref:`ref_flyteidl.event.TaskExecutionMetadata`", "", "Metadata around how a task was executed."
   "event_version", ":ref:`ref_int32`", "", "The event version is used to indicate versioned changes in how data is maintained using this proto message. For example, event_verison > 0 means that maps tasks logs use the TaskExecutionMetadata ExternalResourceInfo fields for each subtask rather than the TaskLog in this message."







.. _ref_flyteidl.admin.TaskExecutionGetDataRequest:

TaskExecutionGetDataRequest
------------------------------------------------------------------

Request structure to fetch inputs and output for a task execution.
By default this data is not returned inline in :ref:`ref_flyteidl.admin.TaskExecutionGetRequest`



.. csv-table:: TaskExecutionGetDataRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.TaskExecutionIdentifier`", "", "The identifier of the task execution for which to fetch inputs and outputs. +required"







.. _ref_flyteidl.admin.TaskExecutionGetDataResponse:

TaskExecutionGetDataResponse
------------------------------------------------------------------

Response structure for TaskExecutionGetDataRequest which contains inputs and outputs for a task execution.



.. csv-table:: TaskExecutionGetDataResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "inputs", ":ref:`ref_flyteidl.admin.UrlBlob`", "", "**Deprecated.** Signed url to fetch a core.LiteralMap of task execution inputs. Deprecated: Please use full_inputs instead."
   "outputs", ":ref:`ref_flyteidl.admin.UrlBlob`", "", "**Deprecated.** Signed url to fetch a core.LiteralMap of task execution outputs. Deprecated: Please use full_outputs instead."
   "full_inputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "Full_inputs will only be populated if they are under a configured size threshold."
   "full_outputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "Full_outputs will only be populated if they are under a configured size threshold."







.. _ref_flyteidl.admin.TaskExecutionGetRequest:

TaskExecutionGetRequest
------------------------------------------------------------------

A message used to fetch a single task execution entity.
See :ref:`ref_flyteidl.admin.TaskExecution` for more details



.. csv-table:: TaskExecutionGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.TaskExecutionIdentifier`", "", "Unique identifier for the task execution. +required"







.. _ref_flyteidl.admin.TaskExecutionList:

TaskExecutionList
------------------------------------------------------------------

Response structure for a query to list of task execution entities.
See :ref:`ref_flyteidl.admin.TaskExecution` for more details



.. csv-table:: TaskExecutionList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "task_executions", ":ref:`ref_flyteidl.admin.TaskExecution`", "repeated", ""
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.TaskExecutionListRequest:

TaskExecutionListRequest
------------------------------------------------------------------

Represents a request structure to retrieve a list of task execution entities yielded by a specific node execution.
See :ref:`ref_flyteidl.admin.TaskExecution` for more details



.. csv-table:: TaskExecutionListRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "node_execution_id", ":ref:`ref_flyteidl.core.NodeExecutionIdentifier`", "", "Indicates the node execution to filter by. +required"
   "limit", ":ref:`ref_uint32`", "", "Indicates the number of resources to be returned. +required"
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. +optional"
   "filters", ":ref:`ref_string`", "", "Indicates a list of filters passed as string. More info on constructing filters : <Link> +optional"
   "sort_by", ":ref:`ref_flyteidl.admin.Sort`", "", "Sort ordering for returned list. +optional"






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/version.proto:

flyteidl/admin/version.proto
==================================================================





.. _ref_flyteidl.admin.GetVersionRequest:

GetVersionRequest
------------------------------------------------------------------

Empty request for GetVersion








.. _ref_flyteidl.admin.GetVersionResponse:

GetVersionResponse
------------------------------------------------------------------

Response for the GetVersion API



.. csv-table:: GetVersionResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "control_plane_version", ":ref:`ref_flyteidl.admin.Version`", "", "The control plane version information. FlyteAdmin and related components form the control plane of Flyte"







.. _ref_flyteidl.admin.Version:

Version
------------------------------------------------------------------

Provides Version information for a component



.. csv-table:: Version type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "Build", ":ref:`ref_string`", "", "Specifies the GIT sha of the build"
   "Version", ":ref:`ref_string`", "", "Version for the build, should follow a semver"
   "BuildTime", ":ref:`ref_string`", "", "Build timestamp"






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/workflow.proto:

flyteidl/admin/workflow.proto
==================================================================





.. _ref_flyteidl.admin.CreateWorkflowFailureReason:

CreateWorkflowFailureReason
------------------------------------------------------------------

When a CreateWorkflowRequest failes due to matching id



.. csv-table:: CreateWorkflowFailureReason type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "exists_different_structure", ":ref:`ref_flyteidl.admin.WorkflowErrorExistsDifferentStructure`", "", ""
   "exists_identical_structure", ":ref:`ref_flyteidl.admin.WorkflowErrorExistsIdenticalStructure`", "", ""







.. _ref_flyteidl.admin.Workflow:

Workflow
------------------------------------------------------------------

Represents the workflow structure stored in the Admin
A workflow is created by ordering tasks and associating outputs to inputs
in order to produce a directed-acyclic execution graph.



.. csv-table:: Workflow type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "id represents the unique identifier of the workflow."
   "closure", ":ref:`ref_flyteidl.admin.WorkflowClosure`", "", "closure encapsulates all the fields that maps to a compiled version of the workflow."
   "short_description", ":ref:`ref_string`", "", "One-liner overview of the entity."







.. _ref_flyteidl.admin.WorkflowClosure:

WorkflowClosure
------------------------------------------------------------------

A container holding the compiled workflow produced from the WorkflowSpec and additional metadata.



.. csv-table:: WorkflowClosure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "compiled_workflow", ":ref:`ref_flyteidl.core.CompiledWorkflowClosure`", "", "Represents the compiled representation of the workflow from the specification provided."
   "created_at", ":ref:`ref_google.protobuf.Timestamp`", "", "Time at which the workflow was created."







.. _ref_flyteidl.admin.WorkflowCreateRequest:

WorkflowCreateRequest
------------------------------------------------------------------

Represents a request structure to create a revision of a workflow.
See :ref:`ref_flyteidl.admin.Workflow` for more details



.. csv-table:: WorkflowCreateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "id represents the unique identifier of the workflow. +required"
   "spec", ":ref:`ref_flyteidl.admin.WorkflowSpec`", "", "Represents the specification for workflow. +required"







.. _ref_flyteidl.admin.WorkflowCreateResponse:

WorkflowCreateResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.WorkflowErrorExistsDifferentStructure:

WorkflowErrorExistsDifferentStructure
------------------------------------------------------------------

The workflow id is already used and the structure is different



.. csv-table:: WorkflowErrorExistsDifferentStructure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", ""







.. _ref_flyteidl.admin.WorkflowErrorExistsIdenticalStructure:

WorkflowErrorExistsIdenticalStructure
------------------------------------------------------------------

The workflow id is already used with an identical sctructure



.. csv-table:: WorkflowErrorExistsIdenticalStructure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", ""







.. _ref_flyteidl.admin.WorkflowList:

WorkflowList
------------------------------------------------------------------

Represents a list of workflows returned from the admin.
See :ref:`ref_flyteidl.admin.Workflow` for more details



.. csv-table:: WorkflowList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "workflows", ":ref:`ref_flyteidl.admin.Workflow`", "repeated", "A list of workflows returned based on the request."
   "token", ":ref:`ref_string`", "", "In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty."







.. _ref_flyteidl.admin.WorkflowSpec:

WorkflowSpec
------------------------------------------------------------------

Represents a structure that encapsulates the specification of the workflow.



.. csv-table:: WorkflowSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "template", ":ref:`ref_flyteidl.core.WorkflowTemplate`", "", "Template of the task that encapsulates all the metadata of the workflow."
   "sub_workflows", ":ref:`ref_flyteidl.core.WorkflowTemplate`", "repeated", "Workflows that are embedded into other workflows need to be passed alongside the parent workflow to the propeller compiler (since the compiler doesn't have any knowledge of other workflows - ie, it doesn't reach out to Admin to see other registered workflows). In fact, subworkflows do not even need to be registered."
   "description", ":ref:`ref_flyteidl.admin.DescriptionEntity`", "", "Represents the specification for description entity."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/admin/workflow_attributes.proto:

flyteidl/admin/workflow_attributes.proto
==================================================================





.. _ref_flyteidl.admin.WorkflowAttributes:

WorkflowAttributes
------------------------------------------------------------------

Defines a set of custom matching attributes which defines resource defaults for a project, domain and workflow.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: WorkflowAttributes type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id for which this set of attributes will be applied."
   "domain", ":ref:`ref_string`", "", "Unique domain id for which this set of attributes will be applied."
   "workflow", ":ref:`ref_string`", "", "Workflow name for which this set of attributes will be applied."
   "matching_attributes", ":ref:`ref_flyteidl.admin.MatchingAttributes`", "", ""







.. _ref_flyteidl.admin.WorkflowAttributesDeleteRequest:

WorkflowAttributesDeleteRequest
------------------------------------------------------------------

Request to delete a set matchable workflow attribute override.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: WorkflowAttributesDeleteRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id which this set of attributes references. +required"
   "domain", ":ref:`ref_string`", "", "Unique domain id which this set of attributes references. +required"
   "workflow", ":ref:`ref_string`", "", "Workflow name which this set of attributes references. +required"
   "resource_type", ":ref:`ref_flyteidl.admin.MatchableResource`", "", "Which type of matchable attributes to delete. +required"







.. _ref_flyteidl.admin.WorkflowAttributesDeleteResponse:

WorkflowAttributesDeleteResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.








.. _ref_flyteidl.admin.WorkflowAttributesGetRequest:

WorkflowAttributesGetRequest
------------------------------------------------------------------

Request to get an individual workflow attribute override.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: WorkflowAttributesGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Unique project id which this set of attributes references. +required"
   "domain", ":ref:`ref_string`", "", "Unique domain id which this set of attributes references. +required"
   "workflow", ":ref:`ref_string`", "", "Workflow name which this set of attributes references. +required"
   "resource_type", ":ref:`ref_flyteidl.admin.MatchableResource`", "", "Which type of matchable attributes to return. +required"







.. _ref_flyteidl.admin.WorkflowAttributesGetResponse:

WorkflowAttributesGetResponse
------------------------------------------------------------------

Response to get an individual workflow attribute override.



.. csv-table:: WorkflowAttributesGetResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "attributes", ":ref:`ref_flyteidl.admin.WorkflowAttributes`", "", ""







.. _ref_flyteidl.admin.WorkflowAttributesUpdateRequest:

WorkflowAttributesUpdateRequest
------------------------------------------------------------------

Sets custom attributes for a project, domain and workflow combination.
For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`



.. csv-table:: WorkflowAttributesUpdateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "attributes", ":ref:`ref_flyteidl.admin.WorkflowAttributes`", "", ""







.. _ref_flyteidl.admin.WorkflowAttributesUpdateResponse:

WorkflowAttributesUpdateResponse
------------------------------------------------------------------

Purposefully empty, may be populated in the future.







..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_google/protobuf/duration.proto:

google/protobuf/duration.proto
==================================================================





.. _ref_google.protobuf.Duration:

Duration
------------------------------------------------------------------

A Duration represents a signed, fixed-length span of time represented
as a count of seconds and fractions of seconds at nanosecond
resolution. It is independent of any calendar and concepts like "day"
or "month". It is related to Timestamp in that the difference between
two Timestamp values is a Duration and it can be added or subtracted
from a Timestamp. Range is approximately +-10,000 years.

# Examples

Example 1: Compute Duration from two Timestamps in pseudo code.

    Timestamp start = ...;
    Timestamp end = ...;
    Duration duration = ...;

    duration.seconds = end.seconds - start.seconds;
    duration.nanos = end.nanos - start.nanos;

    if (duration.seconds < 0 && duration.nanos > 0) {
      duration.seconds += 1;
      duration.nanos -= 1000000000;
    } else if (duration.seconds > 0 && duration.nanos < 0) {
      duration.seconds -= 1;
      duration.nanos += 1000000000;
    }

Example 2: Compute Timestamp from Timestamp + Duration in pseudo code.

    Timestamp start = ...;
    Duration duration = ...;
    Timestamp end = ...;

    end.seconds = start.seconds + duration.seconds;
    end.nanos = start.nanos + duration.nanos;

    if (end.nanos < 0) {
      end.seconds -= 1;
      end.nanos += 1000000000;
    } else if (end.nanos >= 1000000000) {
      end.seconds += 1;
      end.nanos -= 1000000000;
    }

Example 3: Compute Duration from datetime.timedelta in Python.

    td = datetime.timedelta(days=3, minutes=10)
    duration = Duration()
    duration.FromTimedelta(td)

# JSON Mapping

In JSON format, the Duration type is encoded as a string rather than an
object, where the string ends in the suffix "s" (indicating seconds) and
is preceded by the number of seconds, with nanoseconds expressed as
fractional seconds. For example, 3 seconds with 0 nanoseconds should be
encoded in JSON format as "3s", while 3 seconds and 1 nanosecond should
be expressed in JSON format as "3.000000001s", and 3 seconds and 1
microsecond should be expressed in JSON format as "3.000001s".



.. csv-table:: Duration type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "seconds", ":ref:`ref_int64`", "", "Signed seconds of the span of time. Must be from -315,576,000,000 to +315,576,000,000 inclusive. Note: these bounds are computed from: 60 sec/min * 60 min/hr * 24 hr/day * 365.25 days/year * 10000 years"
   "nanos", ":ref:`ref_int32`", "", "Signed fractions of a second at nanosecond resolution of the span of time. Durations less than one second are represented with a 0 `seconds` field and a positive or negative `nanos` field. For durations of one second or more, a non-zero value for the `nanos` field must be of the same sign as the `seconds` field. Must be from -999,999,999 to +999,999,999 inclusive."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_google/protobuf/wrappers.proto:

google/protobuf/wrappers.proto
==================================================================





.. _ref_google.protobuf.BoolValue:

BoolValue
------------------------------------------------------------------

Wrapper message for `bool`.

The JSON representation for `BoolValue` is JSON `true` and `false`.



.. csv-table:: BoolValue type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_bool`", "", "The bool value."







.. _ref_google.protobuf.BytesValue:

BytesValue
------------------------------------------------------------------

Wrapper message for `bytes`.

The JSON representation for `BytesValue` is JSON string.



.. csv-table:: BytesValue type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_bytes`", "", "The bytes value."







.. _ref_google.protobuf.DoubleValue:

DoubleValue
------------------------------------------------------------------

Wrapper message for `double`.

The JSON representation for `DoubleValue` is JSON number.



.. csv-table:: DoubleValue type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_double`", "", "The double value."







.. _ref_google.protobuf.FloatValue:

FloatValue
------------------------------------------------------------------

Wrapper message for `float`.

The JSON representation for `FloatValue` is JSON number.



.. csv-table:: FloatValue type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_float`", "", "The float value."







.. _ref_google.protobuf.Int32Value:

Int32Value
------------------------------------------------------------------

Wrapper message for `int32`.

The JSON representation for `Int32Value` is JSON number.



.. csv-table:: Int32Value type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_int32`", "", "The int32 value."







.. _ref_google.protobuf.Int64Value:

Int64Value
------------------------------------------------------------------

Wrapper message for `int64`.

The JSON representation for `Int64Value` is JSON string.



.. csv-table:: Int64Value type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_int64`", "", "The int64 value."







.. _ref_google.protobuf.StringValue:

StringValue
------------------------------------------------------------------

Wrapper message for `string`.

The JSON representation for `StringValue` is JSON string.



.. csv-table:: StringValue type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_string`", "", "The string value."







.. _ref_google.protobuf.UInt32Value:

UInt32Value
------------------------------------------------------------------

Wrapper message for `uint32`.

The JSON representation for `UInt32Value` is JSON number.



.. csv-table:: UInt32Value type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_uint32`", "", "The uint32 value."







.. _ref_google.protobuf.UInt64Value:

UInt64Value
------------------------------------------------------------------

Wrapper message for `uint64`.

The JSON representation for `UInt64Value` is JSON string.



.. csv-table:: UInt64Value type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_uint64`", "", "The uint64 value."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services


