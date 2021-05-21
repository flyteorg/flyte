# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [flyteidl/admin/common.proto](#flyteidl/admin/common.proto)
    - [Annotations](#flyteidl.admin.Annotations)
    - [Annotations.ValuesEntry](#flyteidl.admin.Annotations.ValuesEntry)
    - [AuthRole](#flyteidl.admin.AuthRole)
    - [EmailNotification](#flyteidl.admin.EmailNotification)
    - [Labels](#flyteidl.admin.Labels)
    - [Labels.ValuesEntry](#flyteidl.admin.Labels.ValuesEntry)
    - [NamedEntity](#flyteidl.admin.NamedEntity)
    - [NamedEntityGetRequest](#flyteidl.admin.NamedEntityGetRequest)
    - [NamedEntityIdentifier](#flyteidl.admin.NamedEntityIdentifier)
    - [NamedEntityIdentifierList](#flyteidl.admin.NamedEntityIdentifierList)
    - [NamedEntityIdentifierListRequest](#flyteidl.admin.NamedEntityIdentifierListRequest)
    - [NamedEntityList](#flyteidl.admin.NamedEntityList)
    - [NamedEntityListRequest](#flyteidl.admin.NamedEntityListRequest)
    - [NamedEntityMetadata](#flyteidl.admin.NamedEntityMetadata)
    - [NamedEntityUpdateRequest](#flyteidl.admin.NamedEntityUpdateRequest)
    - [NamedEntityUpdateResponse](#flyteidl.admin.NamedEntityUpdateResponse)
    - [Notification](#flyteidl.admin.Notification)
    - [ObjectGetRequest](#flyteidl.admin.ObjectGetRequest)
    - [PagerDutyNotification](#flyteidl.admin.PagerDutyNotification)
    - [RawOutputDataConfig](#flyteidl.admin.RawOutputDataConfig)
    - [ResourceListRequest](#flyteidl.admin.ResourceListRequest)
    - [SlackNotification](#flyteidl.admin.SlackNotification)
    - [Sort](#flyteidl.admin.Sort)
    - [UrlBlob](#flyteidl.admin.UrlBlob)
  
    - [NamedEntityState](#flyteidl.admin.NamedEntityState)
    - [Sort.Direction](#flyteidl.admin.Sort.Direction)
  
- [flyteidl/admin/event.proto](#flyteidl/admin/event.proto)
    - [EventErrorAlreadyInTerminalState](#flyteidl.admin.EventErrorAlreadyInTerminalState)
    - [EventFailureReason](#flyteidl.admin.EventFailureReason)
    - [NodeExecutionEventRequest](#flyteidl.admin.NodeExecutionEventRequest)
    - [NodeExecutionEventResponse](#flyteidl.admin.NodeExecutionEventResponse)
    - [TaskExecutionEventRequest](#flyteidl.admin.TaskExecutionEventRequest)
    - [TaskExecutionEventResponse](#flyteidl.admin.TaskExecutionEventResponse)
    - [WorkflowExecutionEventRequest](#flyteidl.admin.WorkflowExecutionEventRequest)
    - [WorkflowExecutionEventResponse](#flyteidl.admin.WorkflowExecutionEventResponse)
  
- [flyteidl/admin/execution.proto](#flyteidl/admin/execution.proto)
    - [AbortMetadata](#flyteidl.admin.AbortMetadata)
    - [Execution](#flyteidl.admin.Execution)
    - [ExecutionClosure](#flyteidl.admin.ExecutionClosure)
    - [ExecutionCreateRequest](#flyteidl.admin.ExecutionCreateRequest)
    - [ExecutionCreateResponse](#flyteidl.admin.ExecutionCreateResponse)
    - [ExecutionList](#flyteidl.admin.ExecutionList)
    - [ExecutionMetadata](#flyteidl.admin.ExecutionMetadata)
    - [ExecutionRelaunchRequest](#flyteidl.admin.ExecutionRelaunchRequest)
    - [ExecutionSpec](#flyteidl.admin.ExecutionSpec)
    - [ExecutionTerminateRequest](#flyteidl.admin.ExecutionTerminateRequest)
    - [ExecutionTerminateResponse](#flyteidl.admin.ExecutionTerminateResponse)
    - [LiteralMapBlob](#flyteidl.admin.LiteralMapBlob)
    - [NotificationList](#flyteidl.admin.NotificationList)
    - [SystemMetadata](#flyteidl.admin.SystemMetadata)
    - [WorkflowExecutionGetDataRequest](#flyteidl.admin.WorkflowExecutionGetDataRequest)
    - [WorkflowExecutionGetDataResponse](#flyteidl.admin.WorkflowExecutionGetDataResponse)
    - [WorkflowExecutionGetRequest](#flyteidl.admin.WorkflowExecutionGetRequest)
  
    - [ExecutionMetadata.ExecutionMode](#flyteidl.admin.ExecutionMetadata.ExecutionMode)
  
- [flyteidl/admin/launch_plan.proto](#flyteidl/admin/launch_plan.proto)
    - [ActiveLaunchPlanListRequest](#flyteidl.admin.ActiveLaunchPlanListRequest)
    - [ActiveLaunchPlanRequest](#flyteidl.admin.ActiveLaunchPlanRequest)
    - [Auth](#flyteidl.admin.Auth)
    - [LaunchPlan](#flyteidl.admin.LaunchPlan)
    - [LaunchPlanClosure](#flyteidl.admin.LaunchPlanClosure)
    - [LaunchPlanCreateRequest](#flyteidl.admin.LaunchPlanCreateRequest)
    - [LaunchPlanCreateResponse](#flyteidl.admin.LaunchPlanCreateResponse)
    - [LaunchPlanList](#flyteidl.admin.LaunchPlanList)
    - [LaunchPlanMetadata](#flyteidl.admin.LaunchPlanMetadata)
    - [LaunchPlanSpec](#flyteidl.admin.LaunchPlanSpec)
    - [LaunchPlanUpdateRequest](#flyteidl.admin.LaunchPlanUpdateRequest)
    - [LaunchPlanUpdateResponse](#flyteidl.admin.LaunchPlanUpdateResponse)
  
    - [LaunchPlanState](#flyteidl.admin.LaunchPlanState)
  
- [flyteidl/admin/matchable_resource.proto](#flyteidl/admin/matchable_resource.proto)
    - [ClusterResourceAttributes](#flyteidl.admin.ClusterResourceAttributes)
    - [ClusterResourceAttributes.AttributesEntry](#flyteidl.admin.ClusterResourceAttributes.AttributesEntry)
    - [ExecutionClusterLabel](#flyteidl.admin.ExecutionClusterLabel)
    - [ExecutionQueueAttributes](#flyteidl.admin.ExecutionQueueAttributes)
    - [ListMatchableAttributesRequest](#flyteidl.admin.ListMatchableAttributesRequest)
    - [ListMatchableAttributesResponse](#flyteidl.admin.ListMatchableAttributesResponse)
    - [MatchableAttributesConfiguration](#flyteidl.admin.MatchableAttributesConfiguration)
    - [MatchingAttributes](#flyteidl.admin.MatchingAttributes)
    - [PluginOverride](#flyteidl.admin.PluginOverride)
    - [PluginOverrides](#flyteidl.admin.PluginOverrides)
    - [TaskResourceAttributes](#flyteidl.admin.TaskResourceAttributes)
    - [TaskResourceSpec](#flyteidl.admin.TaskResourceSpec)
  
    - [MatchableResource](#flyteidl.admin.MatchableResource)
    - [PluginOverride.MissingPluginBehavior](#flyteidl.admin.PluginOverride.MissingPluginBehavior)
  
- [flyteidl/admin/node_execution.proto](#flyteidl/admin/node_execution.proto)
    - [DynamicWorkflowNodeMetadata](#flyteidl.admin.DynamicWorkflowNodeMetadata)
    - [NodeExecution](#flyteidl.admin.NodeExecution)
    - [NodeExecutionClosure](#flyteidl.admin.NodeExecutionClosure)
    - [NodeExecutionForTaskListRequest](#flyteidl.admin.NodeExecutionForTaskListRequest)
    - [NodeExecutionGetDataRequest](#flyteidl.admin.NodeExecutionGetDataRequest)
    - [NodeExecutionGetDataResponse](#flyteidl.admin.NodeExecutionGetDataResponse)
    - [NodeExecutionGetRequest](#flyteidl.admin.NodeExecutionGetRequest)
    - [NodeExecutionList](#flyteidl.admin.NodeExecutionList)
    - [NodeExecutionListRequest](#flyteidl.admin.NodeExecutionListRequest)
    - [NodeExecutionMetaData](#flyteidl.admin.NodeExecutionMetaData)
    - [TaskNodeMetadata](#flyteidl.admin.TaskNodeMetadata)
    - [WorkflowNodeMetadata](#flyteidl.admin.WorkflowNodeMetadata)
  
- [flyteidl/admin/notification.proto](#flyteidl/admin/notification.proto)
    - [EmailMessage](#flyteidl.admin.EmailMessage)
  
- [flyteidl/admin/project.proto](#flyteidl/admin/project.proto)
    - [Domain](#flyteidl.admin.Domain)
    - [Project](#flyteidl.admin.Project)
    - [ProjectListRequest](#flyteidl.admin.ProjectListRequest)
    - [ProjectRegisterRequest](#flyteidl.admin.ProjectRegisterRequest)
    - [ProjectRegisterResponse](#flyteidl.admin.ProjectRegisterResponse)
    - [ProjectUpdateResponse](#flyteidl.admin.ProjectUpdateResponse)
    - [Projects](#flyteidl.admin.Projects)
  
    - [Project.ProjectState](#flyteidl.admin.Project.ProjectState)
  
- [flyteidl/admin/project_domain_attributes.proto](#flyteidl/admin/project_domain_attributes.proto)
    - [ProjectDomainAttributes](#flyteidl.admin.ProjectDomainAttributes)
    - [ProjectDomainAttributesDeleteRequest](#flyteidl.admin.ProjectDomainAttributesDeleteRequest)
    - [ProjectDomainAttributesDeleteResponse](#flyteidl.admin.ProjectDomainAttributesDeleteResponse)
    - [ProjectDomainAttributesGetRequest](#flyteidl.admin.ProjectDomainAttributesGetRequest)
    - [ProjectDomainAttributesGetResponse](#flyteidl.admin.ProjectDomainAttributesGetResponse)
    - [ProjectDomainAttributesUpdateRequest](#flyteidl.admin.ProjectDomainAttributesUpdateRequest)
    - [ProjectDomainAttributesUpdateResponse](#flyteidl.admin.ProjectDomainAttributesUpdateResponse)
  
- [flyteidl/admin/schedule.proto](#flyteidl/admin/schedule.proto)
    - [CronSchedule](#flyteidl.admin.CronSchedule)
    - [FixedRate](#flyteidl.admin.FixedRate)
    - [Schedule](#flyteidl.admin.Schedule)
  
    - [FixedRateUnit](#flyteidl.admin.FixedRateUnit)
  
- [flyteidl/admin/task.proto](#flyteidl/admin/task.proto)
    - [Task](#flyteidl.admin.Task)
    - [TaskClosure](#flyteidl.admin.TaskClosure)
    - [TaskCreateRequest](#flyteidl.admin.TaskCreateRequest)
    - [TaskCreateResponse](#flyteidl.admin.TaskCreateResponse)
    - [TaskList](#flyteidl.admin.TaskList)
    - [TaskSpec](#flyteidl.admin.TaskSpec)
  
- [flyteidl/admin/task_execution.proto](#flyteidl/admin/task_execution.proto)
    - [TaskExecution](#flyteidl.admin.TaskExecution)
    - [TaskExecutionClosure](#flyteidl.admin.TaskExecutionClosure)
    - [TaskExecutionGetDataRequest](#flyteidl.admin.TaskExecutionGetDataRequest)
    - [TaskExecutionGetDataResponse](#flyteidl.admin.TaskExecutionGetDataResponse)
    - [TaskExecutionGetRequest](#flyteidl.admin.TaskExecutionGetRequest)
    - [TaskExecutionList](#flyteidl.admin.TaskExecutionList)
    - [TaskExecutionListRequest](#flyteidl.admin.TaskExecutionListRequest)
  
- [flyteidl/admin/version.proto](#flyteidl/admin/version.proto)
    - [GetVersionRequest](#flyteidl.admin.GetVersionRequest)
    - [GetVersionResponse](#flyteidl.admin.GetVersionResponse)
    - [Version](#flyteidl.admin.Version)
  
- [flyteidl/admin/workflow.proto](#flyteidl/admin/workflow.proto)
    - [Workflow](#flyteidl.admin.Workflow)
    - [WorkflowClosure](#flyteidl.admin.WorkflowClosure)
    - [WorkflowCreateRequest](#flyteidl.admin.WorkflowCreateRequest)
    - [WorkflowCreateResponse](#flyteidl.admin.WorkflowCreateResponse)
    - [WorkflowList](#flyteidl.admin.WorkflowList)
    - [WorkflowSpec](#flyteidl.admin.WorkflowSpec)
  
- [flyteidl/admin/workflow_attributes.proto](#flyteidl/admin/workflow_attributes.proto)
    - [WorkflowAttributes](#flyteidl.admin.WorkflowAttributes)
    - [WorkflowAttributesDeleteRequest](#flyteidl.admin.WorkflowAttributesDeleteRequest)
    - [WorkflowAttributesDeleteResponse](#flyteidl.admin.WorkflowAttributesDeleteResponse)
    - [WorkflowAttributesGetRequest](#flyteidl.admin.WorkflowAttributesGetRequest)
    - [WorkflowAttributesGetResponse](#flyteidl.admin.WorkflowAttributesGetResponse)
    - [WorkflowAttributesUpdateRequest](#flyteidl.admin.WorkflowAttributesUpdateRequest)
    - [WorkflowAttributesUpdateResponse](#flyteidl.admin.WorkflowAttributesUpdateResponse)
  
- [Scalar Value Types](#scalar-value-types)



<a name="flyteidl/admin/common.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/common.proto



<a name="flyteidl.admin.Annotations"></a>

### Annotations
Annotation values to be applied to an execution resource.
In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
to specify how to merge annotations defined at registration and execution time.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [Annotations.ValuesEntry](#flyteidl.admin.Annotations.ValuesEntry) | repeated | Map of custom annotations to be applied to the execution resource. |






<a name="flyteidl.admin.Annotations.ValuesEntry"></a>

### Annotations.ValuesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="flyteidl.admin.AuthRole"></a>

### AuthRole
Defines permissions associated with executions.
Deprecated, please use core.SecurityContext


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| assumable_iam_role | [string](#string) |  |  |
| kubernetes_service_account | [string](#string) |  |  |






<a name="flyteidl.admin.EmailNotification"></a>

### EmailNotification



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| recipients_email | [string](#string) | repeated | The list of email addresses recipients for this notification.

[(validate.rules).repeated = {min_items: 1, unique: true, items: {string: {email: true}}}]; |






<a name="flyteidl.admin.Labels"></a>

### Labels
Label values to be applied to an execution resource.
In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
to specify how to merge labels defined at registration and execution time.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [Labels.ValuesEntry](#flyteidl.admin.Labels.ValuesEntry) | repeated | Map of custom labels to be applied to the execution resource. |






<a name="flyteidl.admin.Labels.ValuesEntry"></a>

### Labels.ValuesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="flyteidl.admin.NamedEntity"></a>

### NamedEntity
Describes information common to a NamedEntity, identified by a project /
domain / name / resource type combination


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_type | [flyteidl.core.ResourceType](#flyteidl.core.ResourceType) |  |  |
| id | [NamedEntityIdentifier](#flyteidl.admin.NamedEntityIdentifier) |  |  |
| metadata | [NamedEntityMetadata](#flyteidl.admin.NamedEntityMetadata) |  |  |






<a name="flyteidl.admin.NamedEntityGetRequest"></a>

### NamedEntityGetRequest
A request to retrieve the metadata associated with a NamedEntityIdentifier


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_type | [flyteidl.core.ResourceType](#flyteidl.core.ResourceType) |  |  |
| id | [NamedEntityIdentifier](#flyteidl.admin.NamedEntityIdentifier) |  |  |






<a name="flyteidl.admin.NamedEntityIdentifier"></a>

### NamedEntityIdentifier
Encapsulation of fields that identifies a Flyte resource.
A resource can internally have multiple versions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Name of the project the resource belongs to.

[(validate.rules).string.min_bytes = 1]; |
| domain | [string](#string) |  | Name of the domain the resource belongs to. A domain can be considered as a subset within a specific project.

[(validate.rules).string.min_bytes = 1]; |
| name | [string](#string) |  | User provided value for the resource. The combination of project &#43; domain &#43; name uniquely identifies the resource. &#43;optional - in certain contexts - like &#39;List API&#39;, &#39;Launch plans&#39; |






<a name="flyteidl.admin.NamedEntityIdentifierList"></a>

### NamedEntityIdentifierList
Represents a list of NamedEntityIdentifiers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entities | [NamedEntityIdentifier](#flyteidl.admin.NamedEntityIdentifier) | repeated | A list of identifiers. |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |






<a name="flyteidl.admin.NamedEntityIdentifierListRequest"></a>

### NamedEntityIdentifierListRequest
Represents a request structure to list identifiers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Name of the project that contains the identifiers.

[(validate.rules).string.min_bytes = 1]; |
| domain | [string](#string) |  | Name of the domain the identifiers belongs to within the project.

[(validate.rules).string.min_bytes = 1]; |
| limit | [uint32](#uint32) |  | Indicates the number of resources to be returned. |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. &#43;optional |
| sort_by | [Sort](#flyteidl.admin.Sort) |  | Sort ordering. &#43;optional |
| filters | [string](#string) |  | Indicates a list of filters passed as string. &#43;optional |






<a name="flyteidl.admin.NamedEntityList"></a>

### NamedEntityList
Represents a list of NamedEntityIdentifiers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entities | [NamedEntity](#flyteidl.admin.NamedEntity) | repeated | A list of NamedEntity objects |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |






<a name="flyteidl.admin.NamedEntityListRequest"></a>

### NamedEntityListRequest
Represents a request structure to list NamedEntity objects


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_type | [flyteidl.core.ResourceType](#flyteidl.core.ResourceType) |  |  |
| project | [string](#string) |  | Name of the project that contains the identifiers.

[(validate.rules).string.min_bytes = 1]; |
| domain | [string](#string) |  | Name of the domain the identifiers belongs to within the project.

[(validate.rules).string.min_bytes = 1]; |
| limit | [uint32](#uint32) |  | Indicates the number of resources to be returned. |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. &#43;optional |
| sort_by | [Sort](#flyteidl.admin.Sort) |  | Sort ordering. &#43;optional |
| filters | [string](#string) |  | Indicates a list of filters passed as string. &#43;optional |






<a name="flyteidl.admin.NamedEntityMetadata"></a>

### NamedEntityMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| description | [string](#string) |  | Common description across all versions of the entity &#43;optional |
| state | [NamedEntityState](#flyteidl.admin.NamedEntityState) |  | Shared state across all version of the entity At this point in time, only workflow entities can have their state archived. |






<a name="flyteidl.admin.NamedEntityUpdateRequest"></a>

### NamedEntityUpdateRequest
Request to set the referenced launch plan state to the configured value.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_type | [flyteidl.core.ResourceType](#flyteidl.core.ResourceType) |  | Resource type of the metadata to update |
| id | [NamedEntityIdentifier](#flyteidl.admin.NamedEntityIdentifier) |  | Identifier of the metadata to update |
| metadata | [NamedEntityMetadata](#flyteidl.admin.NamedEntityMetadata) |  | Metadata object to set as the new value |






<a name="flyteidl.admin.NamedEntityUpdateResponse"></a>

### NamedEntityUpdateResponse
Purposefully empty, may be populated in the future.






<a name="flyteidl.admin.Notification"></a>

### Notification
Represents a structure for notifications based on execution status.
The Notification content is configured within Admin. Future iterations could
expose configuring notifications with custom content.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| phases | [flyteidl.core.WorkflowExecution.Phase](#flyteidl.core.WorkflowExecution.Phase) | repeated | A list of phases to which users can associate the notifications to. |
| email | [EmailNotification](#flyteidl.admin.EmailNotification) |  | option (validate.required) = true; |
| pager_duty | [PagerDutyNotification](#flyteidl.admin.PagerDutyNotification) |  |  |
| slack | [SlackNotification](#flyteidl.admin.SlackNotification) |  |  |






<a name="flyteidl.admin.ObjectGetRequest"></a>

### ObjectGetRequest
Represents a structure to fetch a single resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | Indicates a unique version of resource. |






<a name="flyteidl.admin.PagerDutyNotification"></a>

### PagerDutyNotification



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| recipients_email | [string](#string) | repeated | Currently, PagerDuty notifications leverage email to trigger a notification.

[(validate.rules).repeated = {min_items: 1, unique: true, items: {string: {email: true}}}]; |






<a name="flyteidl.admin.RawOutputDataConfig"></a>

### RawOutputDataConfig
Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).
See https://github.com/flyteorg/flyte/issues/211 for more background information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| output_location_prefix | [string](#string) |  | Prefix for where offloaded data from user workflows will be written e.g. s3://bucket/key or s3://bucket/ |






<a name="flyteidl.admin.ResourceListRequest"></a>

### ResourceListRequest
Represents a request structure to retrieve a list of resources.
Resources include: Task, Workflow, LaunchPlan


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [NamedEntityIdentifier](#flyteidl.admin.NamedEntityIdentifier) |  | id represents the unique identifier of the resource. |
| limit | [uint32](#uint32) |  | Indicates the number of resources to be returned. |
| token | [string](#string) |  | In the case of multiple pages of results, this server-provided token can be used to fetch the next page in a query. &#43;optional |
| filters | [string](#string) |  | Indicates a list of filters passed as string. More info on constructing filters : &lt;Link&gt; &#43;optional |
| sort_by | [Sort](#flyteidl.admin.Sort) |  | Sort ordering. &#43;optional |






<a name="flyteidl.admin.SlackNotification"></a>

### SlackNotification



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| recipients_email | [string](#string) | repeated | Currently, Slack notifications leverage email to trigger a notification.

[(validate.rules).repeated = {min_items: 1, unique: true, items: {string: {email: true}}}]; |






<a name="flyteidl.admin.Sort"></a>

### Sort
Species sort ordering in a list request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  | Indicates an attribute to sort the response values. TODO(katrogan): Add string validation here. This should never be empty. |
| direction | [Sort.Direction](#flyteidl.admin.Sort.Direction) |  | Indicates the direction to apply sort key for response values. &#43;optional |






<a name="flyteidl.admin.UrlBlob"></a>

### UrlBlob
Represents a string url and associated metadata used throughout the platform.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  | Actual url value. |
| bytes | [int64](#int64) |  | Represents the size of the file accessible at the above url. |





 


<a name="flyteidl.admin.NamedEntityState"></a>

### NamedEntityState
The status of the named entity is used to control its visibility in the UI.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NAMED_ENTITY_ACTIVE | 0 | By default, all named entities are considered active and under development. |
| NAMED_ENTITY_ARCHIVED | 1 | Archived named entities are no longer visible in the UI. |
| SYSTEM_GENERATED | 2 | System generated entities that aren&#39;t explicitly created or managed by a user. |



<a name="flyteidl.admin.Sort.Direction"></a>

### Sort.Direction


| Name | Number | Description |
| ---- | ------ | ----------- |
| DESCENDING | 0 |  |
| ASCENDING | 1 |  |


 

 

 



<a name="flyteidl/admin/event.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/event.proto



<a name="flyteidl.admin.EventErrorAlreadyInTerminalState"></a>

### EventErrorAlreadyInTerminalState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| current_phase | [string](#string) |  |  |






<a name="flyteidl.admin.EventFailureReason"></a>

### EventFailureReason



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| already_in_terminal_state | [EventErrorAlreadyInTerminalState](#flyteidl.admin.EventErrorAlreadyInTerminalState) |  |  |






<a name="flyteidl.admin.NodeExecutionEventRequest"></a>

### NodeExecutionEventRequest
Request to send a notification that a node execution event has occurred.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | Unique ID for this request that can be traced between services |
| event | [flyteidl.event.NodeExecutionEvent](#flyteidl.event.NodeExecutionEvent) |  | Details about the event that occurred. |






<a name="flyteidl.admin.NodeExecutionEventResponse"></a>

### NodeExecutionEventResponse
a placeholder for now






<a name="flyteidl.admin.TaskExecutionEventRequest"></a>

### TaskExecutionEventRequest
Request to send a notification that a task execution event has occurred.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | Unique ID for this request that can be traced between services |
| event | [flyteidl.event.TaskExecutionEvent](#flyteidl.event.TaskExecutionEvent) |  | Details about the event that occurred. |






<a name="flyteidl.admin.TaskExecutionEventResponse"></a>

### TaskExecutionEventResponse
a placeholder for now






<a name="flyteidl.admin.WorkflowExecutionEventRequest"></a>

### WorkflowExecutionEventRequest
Request to send a notification that a workflow execution event has occurred.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | Unique ID for this request that can be traced between services |
| event | [flyteidl.event.WorkflowExecutionEvent](#flyteidl.event.WorkflowExecutionEvent) |  | Details about the event that occurred. |






<a name="flyteidl.admin.WorkflowExecutionEventResponse"></a>

### WorkflowExecutionEventResponse
a placeholder for now





 

 

 

 



<a name="flyteidl/admin/execution.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/execution.proto



<a name="flyteidl.admin.AbortMetadata"></a>

### AbortMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cause | [string](#string) |  | In the case of a user-specified abort, this will pass along the user-supplied cause. |
| principal | [string](#string) |  | Identifies the entity (if any) responsible for terminating the execution |






<a name="flyteidl.admin.Execution"></a>

### Execution
A workflow execution represents an instantiated workflow, including all inputs and additional
metadata as well as computed results included state, outputs, and duration-based attributes.
Used as a response object used in Get and List execution requests.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  | Unique identifier of the workflow execution. |
| spec | [ExecutionSpec](#flyteidl.admin.ExecutionSpec) |  | User-provided configuration and inputs for launching the execution. |
| closure | [ExecutionClosure](#flyteidl.admin.ExecutionClosure) |  | Execution results. |






<a name="flyteidl.admin.ExecutionClosure"></a>

### ExecutionClosure
Encapsulates the results of the Execution


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| outputs | [LiteralMapBlob](#flyteidl.admin.LiteralMapBlob) |  | A map of outputs in the case of a successful execution. |
| error | [flyteidl.core.ExecutionError](#flyteidl.core.ExecutionError) |  | Error information in the case of a failed execution. |
| abort_cause | [string](#string) |  | **Deprecated.** In the case of a user-specified abort, this will pass along the user-supplied cause. |
| abort_metadata | [AbortMetadata](#flyteidl.admin.AbortMetadata) |  | In the case of a user-specified abort, this will pass along the user and their supplied cause. |
| computed_inputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | **Deprecated.** Inputs computed and passed for execution. computed_inputs depends on inputs in ExecutionSpec, fixed and default inputs in launch plan |
| phase | [flyteidl.core.WorkflowExecution.Phase](#flyteidl.core.WorkflowExecution.Phase) |  | Most recent recorded phase for the execution. |
| started_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Reported ime at which the execution began running. |
| duration | [google.protobuf.Duration](#google.protobuf.Duration) |  | The amount of time the execution spent running. |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Reported time at which the execution was created. |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Reported time at which the execution was last updated. |
| notifications | [Notification](#flyteidl.admin.Notification) | repeated | The notification settings to use after merging the CreateExecutionRequest and the launch plan notification settings. An execution launched with notifications will always prefer that definition to notifications defined statically in a launch plan. |
| workflow_id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | Identifies the workflow definition for this execution. |






<a name="flyteidl.admin.ExecutionCreateRequest"></a>

### ExecutionCreateRequest
Request to launch an execution with the given project, domain and optionally name.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Name of the project the execution belongs to. |
| domain | [string](#string) |  | Name of the domain the execution belongs to. A domain can be considered as a subset within a specific project. |
| name | [string](#string) |  | User provided value for the resource. If none is provided the system will generate a unique string. &#43;optional |
| spec | [ExecutionSpec](#flyteidl.admin.ExecutionSpec) |  | Additional fields necessary to launch the execution. |
| inputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | The inputs required to start the execution. All required inputs must be included in this map. If not required and not provided, defaults apply. |






<a name="flyteidl.admin.ExecutionCreateResponse"></a>

### ExecutionCreateResponse
The unique identifier for a successfully created execution.
If the name was *not* specified in the create request, this identifier will include a generated name.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  |  |






<a name="flyteidl.admin.ExecutionList"></a>

### ExecutionList
Used as a response for request to list executions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| executions | [Execution](#flyteidl.admin.Execution) | repeated |  |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |






<a name="flyteidl.admin.ExecutionMetadata"></a>

### ExecutionMetadata
Represents attributes about an execution which are not required to launch the execution but are useful to record.
These attributes are assigned at launch time and do not change.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mode | [ExecutionMetadata.ExecutionMode](#flyteidl.admin.ExecutionMetadata.ExecutionMode) |  | [(validate.rules).enum.defined_only = true]; |
| principal | [string](#string) |  | Identifier of the entity that triggered this execution. For systems using back-end authentication any value set here will be discarded in favor of the authenticated user context. |
| nesting | [uint32](#uint32) |  | Indicates the &#34;nestedness&#34; of this execution. If a user launches a workflow execution, the default nesting is 0. If this execution further launches a workflow (child workflow), the nesting level is incremented by 0 =&gt; 1 Generally, if workflow at nesting level k launches a workflow then the child workflow will have nesting = k &#43; 1. |
| scheduled_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | For scheduled executions, the requested time for execution for this specific schedule invocation. |
| parent_node_execution | [flyteidl.core.NodeExecutionIdentifier](#flyteidl.core.NodeExecutionIdentifier) |  | Which subworkflow node launched this execution |
| reference_execution | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  | Optional, a reference workflow execution related to this execution. In the case of a relaunch, this references the original workflow execution. |
| system_metadata | [SystemMetadata](#flyteidl.admin.SystemMetadata) |  | Optional, platform-specific metadata about the execution. In this the future this may be gated behind an ACL or some sort of authorization. |






<a name="flyteidl.admin.ExecutionRelaunchRequest"></a>

### ExecutionRelaunchRequest
Request to relaunch the referenced execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  | Identifier of the workflow execution to relaunch. |
| name | [string](#string) |  | User provided value for the relaunched execution. If none is provided the system will generate a unique string. &#43;optional |






<a name="flyteidl.admin.ExecutionSpec"></a>

### ExecutionSpec
An ExecutionSpec encompasses all data used to launch this execution. The Spec does not change over the lifetime
of an execution as it progresses across phase changes..


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| launch_plan | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | Launch plan to be executed |
| inputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | **Deprecated.** Input values to be passed for the execution |
| metadata | [ExecutionMetadata](#flyteidl.admin.ExecutionMetadata) |  | Metadata for the execution |
| notifications | [NotificationList](#flyteidl.admin.NotificationList) |  | List of notifications based on Execution status transitions When this list is not empty it is used rather than any notifications defined in the referenced launch plan. When this list is empty, the notifications defined for the launch plan will be applied. |
| disable_all | [bool](#bool) |  | This should be set to true if all notifications are intended to be disabled for this execution. |
| labels | [Labels](#flyteidl.admin.Labels) |  | Labels to apply to the execution resource. |
| annotations | [Annotations](#flyteidl.admin.Annotations) |  | Annotations to apply to the execution resource. |
| security_context | [flyteidl.core.SecurityContext](#flyteidl.core.SecurityContext) |  | Optional: security context override to apply this execution. |
| auth_role | [AuthRole](#flyteidl.admin.AuthRole) |  | **Deprecated.** Optional: auth override to apply this execution. |
| quality_of_service | [flyteidl.core.QualityOfService](#flyteidl.core.QualityOfService) |  | Indicates the runtime priority of the execution. |






<a name="flyteidl.admin.ExecutionTerminateRequest"></a>

### ExecutionTerminateRequest
Request to terminate an in-progress execution.  This action is irreversible.
If an execution is already terminated, this request will simply be a no-op.
This request will fail if it references a non-existent execution.
If the request succeeds the phase &#34;ABORTED&#34; will be recorded for the termination
with the optional cause added to the output_result.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  | Uniquely identifies the individual workflow execution to be terminated. |
| cause | [string](#string) |  | Optional reason for aborting. |






<a name="flyteidl.admin.ExecutionTerminateResponse"></a>

### ExecutionTerminateResponse
Purposefully empty, may be populated in the future.






<a name="flyteidl.admin.LiteralMapBlob"></a>

### LiteralMapBlob
Input/output data can represented by actual values or a link to where values are stored


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | Data in LiteralMap format |
| uri | [string](#string) |  | In the event that the map is too large, we return a uri to the data |






<a name="flyteidl.admin.NotificationList"></a>

### NotificationList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| notifications | [Notification](#flyteidl.admin.Notification) | repeated |  |






<a name="flyteidl.admin.SystemMetadata"></a>

### SystemMetadata
Represents system rather than user-facing metadata about an execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| execution_cluster | [string](#string) |  | Which execution cluster this execution ran on. |






<a name="flyteidl.admin.WorkflowExecutionGetDataRequest"></a>

### WorkflowExecutionGetDataRequest
Request structure to fetch inputs and output urls for an execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  | The identifier of the execution for which to fetch inputs and outputs. |






<a name="flyteidl.admin.WorkflowExecutionGetDataResponse"></a>

### WorkflowExecutionGetDataResponse
Response structure for WorkflowExecutionGetDataRequest which contains inputs and outputs for an execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| outputs | [UrlBlob](#flyteidl.admin.UrlBlob) |  | Signed url to fetch a core.LiteralMap of execution outputs. |
| inputs | [UrlBlob](#flyteidl.admin.UrlBlob) |  | Signed url to fetch a core.LiteralMap of execution inputs. |
| full_inputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | Optional, full_inputs will only be populated if they are under a configured size threshold. |
| full_outputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | Optional, full_outputs will only be populated if they are under a configured size threshold. |






<a name="flyteidl.admin.WorkflowExecutionGetRequest"></a>

### WorkflowExecutionGetRequest
A message used to fetch a single workflow execution entity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  | Uniquely identifies an individual workflow execution. |





 


<a name="flyteidl.admin.ExecutionMetadata.ExecutionMode"></a>

### ExecutionMetadata.ExecutionMode
The method by which this execution was launched.

| Name | Number | Description |
| ---- | ------ | ----------- |
| MANUAL | 0 | The default execution mode, MANUAL implies that an execution was launched by an individual. |
| SCHEDULED | 1 | A schedule triggered this execution launch. |
| SYSTEM | 2 | A system process was responsible for launching this execution rather an individual. |
| RELAUNCH | 3 | This execution was launched with identical inputs as a previous execution. |
| CHILD_WORKFLOW | 4 | This execution was triggered by another execution. |


 

 

 



<a name="flyteidl/admin/launch_plan.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/launch_plan.proto



<a name="flyteidl.admin.ActiveLaunchPlanListRequest"></a>

### ActiveLaunchPlanListRequest
Represents a request structure to list active launch plans within a project/domain.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Name of the project that contains the identifiers.

[(validate.rules).string.min_bytes = 1]; |
| domain | [string](#string) |  | Name of the domain the identifiers belongs to within the project.

[(validate.rules).string.min_bytes = 1]; |
| limit | [uint32](#uint32) |  | Indicates the number of resources to be returned. |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. &#43;optional |
| sort_by | [Sort](#flyteidl.admin.Sort) |  | Sort ordering. &#43;optional |






<a name="flyteidl.admin.ActiveLaunchPlanRequest"></a>

### ActiveLaunchPlanRequest
Represents a request struct for finding an active launch plan for a given NamedEntityIdentifier


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [NamedEntityIdentifier](#flyteidl.admin.NamedEntityIdentifier) |  |  |






<a name="flyteidl.admin.Auth"></a>

### Auth
Defines permissions associated with executions created by this launch plan spec.
Deprecated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| assumable_iam_role | [string](#string) |  |  |
| kubernetes_service_account | [string](#string) |  |  |






<a name="flyteidl.admin.LaunchPlan"></a>

### LaunchPlan
A LaunchPlan provides the capability to templatize workflow executions.
Launch plans simplify associating one or more schedules, inputs and notifications with your workflows.
Launch plans can be shared and used to trigger executions with predefined inputs even when a workflow
definition doesn&#39;t necessarily have a default value for said input.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  |  |
| spec | [LaunchPlanSpec](#flyteidl.admin.LaunchPlanSpec) |  |  |
| closure | [LaunchPlanClosure](#flyteidl.admin.LaunchPlanClosure) |  |  |






<a name="flyteidl.admin.LaunchPlanClosure"></a>

### LaunchPlanClosure
Values computed by the flyte platform after launch plan registration.
These include expected_inputs required to be present in a CreateExecutionRequest
to launch the reference workflow as well timestamp values associated with the launch plan.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [LaunchPlanState](#flyteidl.admin.LaunchPlanState) |  | Indicate the Launch plan phase |
| expected_inputs | [flyteidl.core.ParameterMap](#flyteidl.core.ParameterMap) |  | Indicates the set of inputs to execute the Launch plan |
| expected_outputs | [flyteidl.core.VariableMap](#flyteidl.core.VariableMap) |  | Indicates the set of outputs from the Launch plan |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the launch plan was created. |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the launch plan was last updated. |






<a name="flyteidl.admin.LaunchPlanCreateRequest"></a>

### LaunchPlanCreateRequest
Request to register a launch plan. A LaunchPlanSpec may include a complete or incomplete set of inputs required
to launch a workflow execution. By default all launch plans are registered in state INACTIVE. If you wish to
set the state to ACTIVE, you must submit a LaunchPlanUpdateRequest, after you have created a launch plan.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | Uniquely identifies a launch plan entity. |
| spec | [LaunchPlanSpec](#flyteidl.admin.LaunchPlanSpec) |  | User-provided launch plan details, including reference workflow, inputs and other metadata. |






<a name="flyteidl.admin.LaunchPlanCreateResponse"></a>

### LaunchPlanCreateResponse
Purposefully empty, may be populated in the future.






<a name="flyteidl.admin.LaunchPlanList"></a>

### LaunchPlanList
Response object for list launch plan requests.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| launch_plans | [LaunchPlan](#flyteidl.admin.LaunchPlan) | repeated |  |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |






<a name="flyteidl.admin.LaunchPlanMetadata"></a>

### LaunchPlanMetadata
Additional launch plan attributes included in the LaunchPlanSpec not strictly required to launch
the reference workflow.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| schedule | [Schedule](#flyteidl.admin.Schedule) |  | Schedule to execute the Launch Plan |
| notifications | [Notification](#flyteidl.admin.Notification) | repeated | List of notifications based on Execution status transitions |






<a name="flyteidl.admin.LaunchPlanSpec"></a>

### LaunchPlanSpec
User-provided launch plan definition and configuration values.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workflow_id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | Reference to the Workflow template that the launch plan references |
| entity_metadata | [LaunchPlanMetadata](#flyteidl.admin.LaunchPlanMetadata) |  | Metadata for the Launch Plan |
| default_inputs | [flyteidl.core.ParameterMap](#flyteidl.core.ParameterMap) |  | Input values to be passed for the execution |
| fixed_inputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | Fixed, non-overridable inputs for the Launch Plan |
| role | [string](#string) |  | **Deprecated.** String to indicate the role to use to execute the workflow underneath |
| labels | [Labels](#flyteidl.admin.Labels) |  | Custom labels to be applied to the execution resource. |
| annotations | [Annotations](#flyteidl.admin.Annotations) |  | Custom annotations to be applied to the execution resource. |
| auth | [Auth](#flyteidl.admin.Auth) |  | **Deprecated.** Indicates the permission associated with workflow executions triggered with this launch plan. |
| auth_role | [AuthRole](#flyteidl.admin.AuthRole) |  | **Deprecated.**  |
| security_context | [flyteidl.core.SecurityContext](#flyteidl.core.SecurityContext) |  | Indicates security context for permissions triggered with this launch plan |
| quality_of_service | [flyteidl.core.QualityOfService](#flyteidl.core.QualityOfService) |  | Indicates the runtime priority of the execution. |
| raw_output_data_config | [RawOutputDataConfig](#flyteidl.admin.RawOutputDataConfig) |  |  |






<a name="flyteidl.admin.LaunchPlanUpdateRequest"></a>

### LaunchPlanUpdateRequest
Request to set the referenced launch plan state to the configured value.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | Identifier of launch plan for which to change state. |
| state | [LaunchPlanState](#flyteidl.admin.LaunchPlanState) |  | Desired state to apply to the launch plan. |






<a name="flyteidl.admin.LaunchPlanUpdateResponse"></a>

### LaunchPlanUpdateResponse
Purposefully empty, may be populated in the future.





 


<a name="flyteidl.admin.LaunchPlanState"></a>

### LaunchPlanState
By default any launch plan regardless of state can be used to launch a workflow execution.
However, at most one version of a launch plan
(e.g. a NamedEntityIdentifier set of shared project, domain and name values) can be
active at a time in regards to *schedules*. That is, at most one schedule in a NamedEntityIdentifier
group will be observed and trigger executions at a defined cadence.

| Name | Number | Description |
| ---- | ------ | ----------- |
| INACTIVE | 0 |  |
| ACTIVE | 1 |  |


 

 

 



<a name="flyteidl/admin/matchable_resource.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/matchable_resource.proto



<a name="flyteidl.admin.ClusterResourceAttributes"></a>

### ClusterResourceAttributes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [ClusterResourceAttributes.AttributesEntry](#flyteidl.admin.ClusterResourceAttributes.AttributesEntry) | repeated | Custom resource attributes which will be applied in cluster resource creation (e.g. quotas). Map keys are the *case-sensitive* names of variables in templatized resource files. Map values should be the custom values which get substituted during resource creation. |






<a name="flyteidl.admin.ClusterResourceAttributes.AttributesEntry"></a>

### ClusterResourceAttributes.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="flyteidl.admin.ExecutionClusterLabel"></a>

### ExecutionClusterLabel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [string](#string) |  | Label value to determine where the execution will be run |






<a name="flyteidl.admin.ExecutionQueueAttributes"></a>

### ExecutionQueueAttributes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tags | [string](#string) | repeated | Tags used for assigning execution queues for tasks defined within this project. |






<a name="flyteidl.admin.ListMatchableAttributesRequest"></a>

### ListMatchableAttributesRequest
Request all matching resource attributes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_type | [MatchableResource](#flyteidl.admin.MatchableResource) |  |  |






<a name="flyteidl.admin.ListMatchableAttributesResponse"></a>

### ListMatchableAttributesResponse
Response for a request for all matching resource attributes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| configurations | [MatchableAttributesConfiguration](#flyteidl.admin.MatchableAttributesConfiguration) | repeated |  |






<a name="flyteidl.admin.MatchableAttributesConfiguration"></a>

### MatchableAttributesConfiguration
Represents a custom set of attributes applied for either a domain; a domain and project; or
domain, project and workflow name.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [MatchingAttributes](#flyteidl.admin.MatchingAttributes) |  |  |
| domain | [string](#string) |  |  |
| project | [string](#string) |  |  |
| workflow | [string](#string) |  |  |
| launch_plan | [string](#string) |  |  |






<a name="flyteidl.admin.MatchingAttributes"></a>

### MatchingAttributes
Generic container for encapsulating all types of the above attributes messages.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_resource_attributes | [TaskResourceAttributes](#flyteidl.admin.TaskResourceAttributes) |  |  |
| cluster_resource_attributes | [ClusterResourceAttributes](#flyteidl.admin.ClusterResourceAttributes) |  |  |
| execution_queue_attributes | [ExecutionQueueAttributes](#flyteidl.admin.ExecutionQueueAttributes) |  |  |
| execution_cluster_label | [ExecutionClusterLabel](#flyteidl.admin.ExecutionClusterLabel) |  |  |
| quality_of_service | [flyteidl.core.QualityOfService](#flyteidl.core.QualityOfService) |  |  |
| plugin_overrides | [PluginOverrides](#flyteidl.admin.PluginOverrides) |  |  |






<a name="flyteidl.admin.PluginOverride"></a>

### PluginOverride
This MatchableAttribute configures selecting alternate plugin implementations for a given task type.
In addition to an override implementation a selection of fallbacks can be provided or other modes
for handling cases where the desired plugin override is not enabled in a given Flyte deployment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_type | [string](#string) |  | A predefined yet extensible Task type identifier. |
| plugin_id | [string](#string) | repeated | A set of plugin ids which should handle tasks of this type instead of the default registered plugin. The list will be tried in order until a plugin is found with that id. |
| missing_plugin_behavior | [PluginOverride.MissingPluginBehavior](#flyteidl.admin.PluginOverride.MissingPluginBehavior) |  | Defines the behavior when no plugin from the plugin_id list is not found. |






<a name="flyteidl.admin.PluginOverrides"></a>

### PluginOverrides



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| overrides | [PluginOverride](#flyteidl.admin.PluginOverride) | repeated |  |






<a name="flyteidl.admin.TaskResourceAttributes"></a>

### TaskResourceAttributes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| defaults | [TaskResourceSpec](#flyteidl.admin.TaskResourceSpec) |  |  |
| limits | [TaskResourceSpec](#flyteidl.admin.TaskResourceSpec) |  |  |






<a name="flyteidl.admin.TaskResourceSpec"></a>

### TaskResourceSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cpu | [string](#string) |  |  |
| gpu | [string](#string) |  |  |
| memory | [string](#string) |  |  |
| storage | [string](#string) |  |  |





 


<a name="flyteidl.admin.MatchableResource"></a>

### MatchableResource
Defines a resource that can be configured by customizable Project-, ProjectDomain- or WorkflowAttributes
based on matching tags.

| Name | Number | Description |
| ---- | ------ | ----------- |
| TASK_RESOURCE | 0 | Applies to customizable task resource requests and limits. |
| CLUSTER_RESOURCE | 1 | Applies to configuring templated kubernetes cluster resources. |
| EXECUTION_QUEUE | 2 | Configures task and dynamic task execution queue assignment. |
| EXECUTION_CLUSTER_LABEL | 3 | Configures the K8s cluster label to be used for execution to be run |
| QUALITY_OF_SERVICE_SPECIFICATION | 4 | Configures default quality of service when undefined in an execution spec. |
| PLUGIN_OVERRIDE | 5 | Selects configurable plugin implementation behavior for a given task type. |



<a name="flyteidl.admin.PluginOverride.MissingPluginBehavior"></a>

### PluginOverride.MissingPluginBehavior


| Name | Number | Description |
| ---- | ------ | ----------- |
| FAIL | 0 |  |
| USE_DEFAULT | 1 | Uses the system-configured default implementation. |


 

 

 



<a name="flyteidl/admin/node_execution.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/node_execution.proto



<a name="flyteidl.admin.DynamicWorkflowNodeMetadata"></a>

### DynamicWorkflowNodeMetadata
For dynamic workflow nodes we capture information about the dynamic workflow definition that gets generated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | id represents the unique identifier of the workflow. |
| compiled_workflow | [flyteidl.core.CompiledWorkflowClosure](#flyteidl.core.CompiledWorkflowClosure) |  | Represents the compiled representation of the embedded dynamic workflow. |






<a name="flyteidl.admin.NodeExecution"></a>

### NodeExecution
Encapsulates all details for a single node execution entity.
A node represents a component in the overall workflow graph. A node launch a task, multiple tasks, an entire nested
sub-workflow, or even a separate child-workflow execution.
The same task can be called repeatedly in a single workflow but each node is unique.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.NodeExecutionIdentifier](#flyteidl.core.NodeExecutionIdentifier) |  | Uniquely identifies an individual node execution. |
| input_uri | [string](#string) |  | Path to remote data store where input blob is stored. |
| closure | [NodeExecutionClosure](#flyteidl.admin.NodeExecutionClosure) |  | Computed results associated with this node execution. |
| metadata | [NodeExecutionMetaData](#flyteidl.admin.NodeExecutionMetaData) |  | Metadata for Node Execution |






<a name="flyteidl.admin.NodeExecutionClosure"></a>

### NodeExecutionClosure
Container for node execution details and results.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| output_uri | [string](#string) |  |  |
| error | [flyteidl.core.ExecutionError](#flyteidl.core.ExecutionError) |  | Error information for the Node |
| phase | [flyteidl.core.NodeExecution.Phase](#flyteidl.core.NodeExecution.Phase) |  | The last recorded phase for this node execution. |
| started_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the node execution began running. |
| duration | [google.protobuf.Duration](#google.protobuf.Duration) |  | The amount of time the node execution spent running. |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the node execution was created. |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the node execution was last updated. |
| workflow_node_metadata | [WorkflowNodeMetadata](#flyteidl.admin.WorkflowNodeMetadata) |  |  |
| task_node_metadata | [TaskNodeMetadata](#flyteidl.admin.TaskNodeMetadata) |  |  |






<a name="flyteidl.admin.NodeExecutionForTaskListRequest"></a>

### NodeExecutionForTaskListRequest
Represents a request structure to retrieve a list of node execution entities launched by a specific task.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_execution_id | [flyteidl.core.TaskExecutionIdentifier](#flyteidl.core.TaskExecutionIdentifier) |  | Indicates the node execution to filter by. |
| limit | [uint32](#uint32) |  | Indicates the number of resources to be returned. |
| token | [string](#string) |  | In the case of multiple pages of results, the, server-provided token can be used to fetch the next page in a query. &#43;optional |
| filters | [string](#string) |  | Indicates a list of filters passed as string. More info on constructing filters : &lt;Link&gt; &#43;optional |
| sort_by | [Sort](#flyteidl.admin.Sort) |  | Sort ordering. &#43;optional |






<a name="flyteidl.admin.NodeExecutionGetDataRequest"></a>

### NodeExecutionGetDataRequest
Request structure to fetch inputs and output urls for a node execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.NodeExecutionIdentifier](#flyteidl.core.NodeExecutionIdentifier) |  | The identifier of the node execution for which to fetch inputs and outputs. |






<a name="flyteidl.admin.NodeExecutionGetDataResponse"></a>

### NodeExecutionGetDataResponse
Response structure for NodeExecutionGetDataRequest which contains inputs and outputs for a node execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inputs | [UrlBlob](#flyteidl.admin.UrlBlob) |  | Signed url to fetch a core.LiteralMap of node execution inputs. |
| outputs | [UrlBlob](#flyteidl.admin.UrlBlob) |  | Signed url to fetch a core.LiteralMap of node execution outputs. |
| full_inputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | Optional, full_inputs will only be populated if they are under a configured size threshold. |
| full_outputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | Optional, full_outputs will only be populated if they are under a configured size threshold. |
| dynamic_workflow | [DynamicWorkflowNodeMetadata](#flyteidl.admin.DynamicWorkflowNodeMetadata) |  | Optional Workflow closure for a dynamically generated workflow, in the case this node yields a dynamic workflow we return its structure here. |






<a name="flyteidl.admin.NodeExecutionGetRequest"></a>

### NodeExecutionGetRequest
A message used to fetch a single node execution entity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.NodeExecutionIdentifier](#flyteidl.core.NodeExecutionIdentifier) |  | Uniquely identifies an individual node execution. |






<a name="flyteidl.admin.NodeExecutionList"></a>

### NodeExecutionList
Request structure to retrieve a list of node execution entities.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node_executions | [NodeExecution](#flyteidl.admin.NodeExecution) | repeated |  |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |






<a name="flyteidl.admin.NodeExecutionListRequest"></a>

### NodeExecutionListRequest
Represents a request structure to retrieve a list of node execution entities.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workflow_execution_id | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  | Indicates the workflow execution to filter by. |
| limit | [uint32](#uint32) |  | Indicates the number of resources to be returned. |
| token | [string](#string) |  | In the case of multiple pages of results, the, server-provided token can be used to fetch the next page in a query. &#43;optional |
| filters | [string](#string) |  | Indicates a list of filters passed as string. More info on constructing filters : &lt;Link&gt; &#43;optional |
| sort_by | [Sort](#flyteidl.admin.Sort) |  | Sort ordering. &#43;optional |
| unique_parent_id | [string](#string) |  | Unique identifier of the parent node in the execution &#43;optional |






<a name="flyteidl.admin.NodeExecutionMetaData"></a>

### NodeExecutionMetaData
Represents additional attributes related to a Node Execution


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| retry_group | [string](#string) |  | Node executions are grouped depending on retries of the parent Retry group is unique within the context of a parent node. |
| is_parent_node | [bool](#bool) |  | Boolean flag indicating if the node has child nodes under it |
| spec_node_id | [string](#string) |  | Node id of the node in the original workflow This maps to value of WorkflowTemplate.nodes[X].id |






<a name="flyteidl.admin.TaskNodeMetadata"></a>

### TaskNodeMetadata
Metadata for the case in which the node is a TaskNode


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cache_status | [flyteidl.core.CatalogCacheStatus](#flyteidl.core.CatalogCacheStatus) |  | Captures the status of caching for this execution. |
| catalog_key | [flyteidl.core.CatalogMetadata](#flyteidl.core.CatalogMetadata) |  | This structure carries the catalog artifact information |






<a name="flyteidl.admin.WorkflowNodeMetadata"></a>

### WorkflowNodeMetadata
Metadata for a WorkflowNode


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| executionId | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  |  |





 

 

 

 



<a name="flyteidl/admin/notification.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/notification.proto



<a name="flyteidl.admin.EmailMessage"></a>

### EmailMessage
Represents the Email object that is sent to a publisher/subscriber
to forward the notification.
Note: This is internal to Admin and doesn&#39;t need to be exposed to other components.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| recipients_email | [string](#string) | repeated | The list of email addresses to receive an email with the content populated in the other fields. Currently, each email recipient will receive its own email. This populates the TO field.

[(validate.rules).repeated = {min_items: 1, unique: true, items: {string: {email: true}}}]; |
| sender_email | [string](#string) |  | The email of the sender. This populates the FROM field.

[(validate.rules).string.email = true]; |
| subject_line | [string](#string) |  | The content of the subject line. This populates the SUBJECT field.

[(validate.rules).string.min_len = 1]; |
| body | [string](#string) |  | The content of the email body. This populates the BODY field.

[(validate.rules).string.min_len = 1]; |





 

 

 

 



<a name="flyteidl/admin/project.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/project.proto



<a name="flyteidl.admin.Domain"></a>

### Domain
Namespace within a project commonly used to differentiate between different service instances.
e.g. &#34;production&#34;, &#34;development&#34;, etc.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  | Display name. |






<a name="flyteidl.admin.Project"></a>

### Project
Top-level namespace used to classify different entities like workflows and executions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  | Display name. |
| domains | [Domain](#flyteidl.admin.Domain) | repeated |  |
| description | [string](#string) |  |  |
| labels | [Labels](#flyteidl.admin.Labels) |  | Leverage Labels from flyteidel.admin.common.proto to tag projects with ownership information. |
| state | [Project.ProjectState](#flyteidl.admin.Project.ProjectState) |  |  |






<a name="flyteidl.admin.ProjectListRequest"></a>

### ProjectListRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| limit | [uint32](#uint32) |  | Indicates the number of projects to be returned. |
| token | [string](#string) |  | In the case of multiple pages of results, this server-provided token can be used to fetch the next page in a query. &#43;optional |
| filters | [string](#string) |  | Indicates a list of filters passed as string. More info on constructing filters : &lt;Link&gt; &#43;optional |
| sort_by | [Sort](#flyteidl.admin.Sort) |  | Sort ordering. &#43;optional |






<a name="flyteidl.admin.ProjectRegisterRequest"></a>

### ProjectRegisterRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#flyteidl.admin.Project) |  |  |






<a name="flyteidl.admin.ProjectRegisterResponse"></a>

### ProjectRegisterResponse







<a name="flyteidl.admin.ProjectUpdateResponse"></a>

### ProjectUpdateResponse







<a name="flyteidl.admin.Projects"></a>

### Projects



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| projects | [Project](#flyteidl.admin.Project) | repeated |  |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |





 


<a name="flyteidl.admin.Project.ProjectState"></a>

### Project.ProjectState
The state of the project is used to control its visibility in the UI and validity.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ACTIVE | 0 | By default, all projects are considered active. |
| ARCHIVED | 1 | Archived projects are no longer visible in the UI and no longer valid. |
| SYSTEM_GENERATED | 2 | System generated projects that aren&#39;t explicitly created or managed by a user. |


 

 

 



<a name="flyteidl/admin/project_domain_attributes.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/project_domain_attributes.proto



<a name="flyteidl.admin.ProjectDomainAttributes"></a>

### ProjectDomainAttributes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Unique project id for which this set of attributes will be applied. |
| domain | [string](#string) |  | Unique domain id for which this set of attributes will be applied. |
| matching_attributes | [MatchingAttributes](#flyteidl.admin.MatchingAttributes) |  |  |






<a name="flyteidl.admin.ProjectDomainAttributesDeleteRequest"></a>

### ProjectDomainAttributesDeleteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Unique project id which this set of attributes references. |
| domain | [string](#string) |  | Unique domain id which this set of attributes references. |
| resource_type | [MatchableResource](#flyteidl.admin.MatchableResource) |  |  |






<a name="flyteidl.admin.ProjectDomainAttributesDeleteResponse"></a>

### ProjectDomainAttributesDeleteResponse
Purposefully empty, may be populated in the future.






<a name="flyteidl.admin.ProjectDomainAttributesGetRequest"></a>

### ProjectDomainAttributesGetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Unique project id which this set of attributes references. |
| domain | [string](#string) |  | Unique domain id which this set of attributes references. |
| resource_type | [MatchableResource](#flyteidl.admin.MatchableResource) |  |  |






<a name="flyteidl.admin.ProjectDomainAttributesGetResponse"></a>

### ProjectDomainAttributesGetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [ProjectDomainAttributes](#flyteidl.admin.ProjectDomainAttributes) |  |  |






<a name="flyteidl.admin.ProjectDomainAttributesUpdateRequest"></a>

### ProjectDomainAttributesUpdateRequest
Sets custom attributes for a project-domain combination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [ProjectDomainAttributes](#flyteidl.admin.ProjectDomainAttributes) |  |  |






<a name="flyteidl.admin.ProjectDomainAttributesUpdateResponse"></a>

### ProjectDomainAttributesUpdateResponse
Purposefully empty, may be populated in the future.





 

 

 

 



<a name="flyteidl/admin/schedule.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/schedule.proto



<a name="flyteidl.admin.CronSchedule"></a>

### CronSchedule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| schedule | [string](#string) |  | Standard/default cron implementation as described by https://en.wikipedia.org/wiki/Cron#CRON_expression; Also supports nonstandard predefined scheduling definitions as described by https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions except @reboot |
| offset | [string](#string) |  | ISO 8601 duration as described by https://en.wikipedia.org/wiki/ISO_8601#Durations |






<a name="flyteidl.admin.FixedRate"></a>

### FixedRate
Option for schedules run at a certain frequency, e.g. every 2 minutes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [uint32](#uint32) |  |  |
| unit | [FixedRateUnit](#flyteidl.admin.FixedRateUnit) |  |  |






<a name="flyteidl.admin.Schedule"></a>

### Schedule
Defines complete set of information required to trigger an execution on a schedule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cron_expression | [string](#string) |  | **Deprecated.** Uses AWS syntax: &#34;Minutes Hours Day-of-month Month Day-of-week Year&#34; e.g. for a schedule that runs every 15 minutes: &#34;0/15 * * * ? *&#34; |
| rate | [FixedRate](#flyteidl.admin.FixedRate) |  |  |
| cron_schedule | [CronSchedule](#flyteidl.admin.CronSchedule) |  |  |
| kickoff_time_input_arg | [string](#string) |  | Name of the input variable that the kickoff time will be supplied to when the workflow is kicked off. |





 


<a name="flyteidl.admin.FixedRateUnit"></a>

### FixedRateUnit
Represents a frequency at which to run a schedule.

| Name | Number | Description |
| ---- | ------ | ----------- |
| MINUTE | 0 |  |
| HOUR | 1 |  |
| DAY | 2 |  |


 

 

 



<a name="flyteidl/admin/task.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/task.proto



<a name="flyteidl.admin.Task"></a>

### Task
Flyte workflows are composed of many ordered tasks. That is small, reusable, self-contained logical blocks
arranged to process workflow inputs and produce a deterministic set of outputs.
Tasks can come in many varieties tuned for specialized behavior.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | id represents the unique identifier of the task. |
| closure | [TaskClosure](#flyteidl.admin.TaskClosure) |  | closure encapsulates all the fields that maps to a compiled version of the task. |






<a name="flyteidl.admin.TaskClosure"></a>

### TaskClosure
Compute task attributes which include values derived from the TaskSpec, as well as plugin-specific data
and task metadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| compiled_task | [flyteidl.core.CompiledTask](#flyteidl.core.CompiledTask) |  | Represents the compiled representation of the task from the specification provided. |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the task was created. |






<a name="flyteidl.admin.TaskCreateRequest"></a>

### TaskCreateRequest
Represents a request structure to create a revision of a task.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | id represents the unique identifier of the task. |
| spec | [TaskSpec](#flyteidl.admin.TaskSpec) |  | Represents the specification for task. |






<a name="flyteidl.admin.TaskCreateResponse"></a>

### TaskCreateResponse
Represents a response structure if task creation succeeds.

Purposefully empty, may be populated in the future.






<a name="flyteidl.admin.TaskList"></a>

### TaskList
Represents a list of tasks returned from the admin.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tasks | [Task](#flyteidl.admin.Task) | repeated | A list of tasks returned based on the request. |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |






<a name="flyteidl.admin.TaskSpec"></a>

### TaskSpec
Represents a structure that encapsulates the user-configured specification of the task.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [flyteidl.core.TaskTemplate](#flyteidl.core.TaskTemplate) |  | Template of the task that encapsulates all the metadata of the task. |





 

 

 

 



<a name="flyteidl/admin/task_execution.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/task_execution.proto



<a name="flyteidl.admin.TaskExecution"></a>

### TaskExecution
Encapsulates all details for a single task execution entity.
A task execution represents an instantiated task, including all inputs and additional
metadata as well as computed results included state, outputs, and duration-based attributes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.TaskExecutionIdentifier](#flyteidl.core.TaskExecutionIdentifier) |  | Unique identifier for the task execution. |
| input_uri | [string](#string) |  | Path to remote data store where input blob is stored. |
| closure | [TaskExecutionClosure](#flyteidl.admin.TaskExecutionClosure) |  | Task execution details and results. |
| is_parent | [bool](#bool) |  | Whether this task spawned nodes. |






<a name="flyteidl.admin.TaskExecutionClosure"></a>

### TaskExecutionClosure
Container for task execution details and results.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| output_uri | [string](#string) |  | Path to remote data store where output blob is stored if the execution succeeded (and produced outputs). |
| error | [flyteidl.core.ExecutionError](#flyteidl.core.ExecutionError) |  | Error information for the task execution. Populated if the execution failed. |
| phase | [flyteidl.core.TaskExecution.Phase](#flyteidl.core.TaskExecution.Phase) |  | The last recorded phase for this task execution. |
| logs | [flyteidl.core.TaskLog](#flyteidl.core.TaskLog) | repeated | Detailed log information output by the task execution. |
| started_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the task execution began running. |
| duration | [google.protobuf.Duration](#google.protobuf.Duration) |  | The amount of time the task execution spent running. |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the task execution was created. |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the task execution was last updated. |
| custom_info | [google.protobuf.Struct](#google.protobuf.Struct) |  | Custom data specific to the task plugin. |
| reason | [string](#string) |  | If there is an explanation for the most recent phase transition, the reason will capture it. |
| task_type | [string](#string) |  | A predefined yet extensible Task type identifier. |
| metadata | [flyteidl.event.TaskExecutionMetadata](#flyteidl.event.TaskExecutionMetadata) |  | Metadata around how a task was executed. |






<a name="flyteidl.admin.TaskExecutionGetDataRequest"></a>

### TaskExecutionGetDataRequest
Request structure to fetch inputs and output urls for a task execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.TaskExecutionIdentifier](#flyteidl.core.TaskExecutionIdentifier) |  | The identifier of the task execution for which to fetch inputs and outputs. |






<a name="flyteidl.admin.TaskExecutionGetDataResponse"></a>

### TaskExecutionGetDataResponse
Response structure for TaskExecutionGetDataRequest which contains inputs and outputs for a task execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inputs | [UrlBlob](#flyteidl.admin.UrlBlob) |  | Signed url to fetch a core.LiteralMap of task execution inputs. |
| outputs | [UrlBlob](#flyteidl.admin.UrlBlob) |  | Signed url to fetch a core.LiteralMap of task execution outputs. |
| full_inputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | Optional, full_inputs will only be populated if they are under a configured size threshold. |
| full_outputs | [flyteidl.core.LiteralMap](#flyteidl.core.LiteralMap) |  | Optional, full_outputs will only be populated if they are under a configured size threshold. |






<a name="flyteidl.admin.TaskExecutionGetRequest"></a>

### TaskExecutionGetRequest
A message used to fetch a single task execution entity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.TaskExecutionIdentifier](#flyteidl.core.TaskExecutionIdentifier) |  | Unique identifier for the task execution. |






<a name="flyteidl.admin.TaskExecutionList"></a>

### TaskExecutionList
Response structure for a query to list of task execution entities.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_executions | [TaskExecution](#flyteidl.admin.TaskExecution) | repeated |  |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |






<a name="flyteidl.admin.TaskExecutionListRequest"></a>

### TaskExecutionListRequest
Represents a request structure to retrieve a list of task execution entities.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node_execution_id | [flyteidl.core.NodeExecutionIdentifier](#flyteidl.core.NodeExecutionIdentifier) |  | Indicates the node execution to filter by. |
| limit | [uint32](#uint32) |  | Indicates the number of resources to be returned. |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. &#43;optional |
| filters | [string](#string) |  | Indicates a list of filters passed as string. More info on constructing filters : &lt;Link&gt; &#43;optional |
| sort_by | [Sort](#flyteidl.admin.Sort) |  | Sort ordering for returned list. &#43;optional |





 

 

 

 



<a name="flyteidl/admin/version.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/version.proto



<a name="flyteidl.admin.GetVersionRequest"></a>

### GetVersionRequest
Empty request for GetVersion






<a name="flyteidl.admin.GetVersionResponse"></a>

### GetVersionResponse
Response for the GetVersion API


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| control_plane_version | [Version](#flyteidl.admin.Version) |  | The control plane version information. FlyteAdmin and related components form the control plane of Flyte |






<a name="flyteidl.admin.Version"></a>

### Version
Provides Version information for a component


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Build | [string](#string) |  | Specifies the GIT sha of the build |
| Version | [string](#string) |  | Version for the build, should follow a semver |
| BuildTime | [string](#string) |  | Build timestamp |





 

 

 

 



<a name="flyteidl/admin/workflow.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/workflow.proto



<a name="flyteidl.admin.Workflow"></a>

### Workflow
Represents the workflow structure stored in the Admin
A workflow is created by ordering tasks and associating outputs to inputs
in order to produce a directed-acyclic execution graph.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | id represents the unique identifier of the workflow. |
| closure | [WorkflowClosure](#flyteidl.admin.WorkflowClosure) |  | closure encapsulates all the fields that maps to a compiled version of the workflow. |






<a name="flyteidl.admin.WorkflowClosure"></a>

### WorkflowClosure
A container holding the compiled workflow produced from the WorkflowSpec and additional metadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| compiled_workflow | [flyteidl.core.CompiledWorkflowClosure](#flyteidl.core.CompiledWorkflowClosure) |  | Represents the compiled representation of the workflow from the specification provided. |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time at which the workflow was created. |






<a name="flyteidl.admin.WorkflowCreateRequest"></a>

### WorkflowCreateRequest
Represents a request structure to create a revision of a workflow.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [flyteidl.core.Identifier](#flyteidl.core.Identifier) |  | id represents the unique identifier of the workflow. |
| spec | [WorkflowSpec](#flyteidl.admin.WorkflowSpec) |  | Represents the specification for workflow. |






<a name="flyteidl.admin.WorkflowCreateResponse"></a>

### WorkflowCreateResponse
Purposefully empty, may be populated in the future.






<a name="flyteidl.admin.WorkflowList"></a>

### WorkflowList
Represents a list of workflows returned from the admin.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workflows | [Workflow](#flyteidl.admin.Workflow) | repeated | A list of workflows returned based on the request. |
| token | [string](#string) |  | In the case of multiple pages of results, the server-provided token can be used to fetch the next page in a query. If there are no more results, this value will be empty. |






<a name="flyteidl.admin.WorkflowSpec"></a>

### WorkflowSpec
Represents a structure that encapsulates the specification of the workflow.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [flyteidl.core.WorkflowTemplate](#flyteidl.core.WorkflowTemplate) |  | Template of the task that encapsulates all the metadata of the workflow. |
| sub_workflows | [flyteidl.core.WorkflowTemplate](#flyteidl.core.WorkflowTemplate) | repeated | Workflows that are embedded into other workflows need to be passed alongside the parent workflow to the propeller compiler (since the compiler doesn&#39;t have any knowledge of other workflows - ie, it doesn&#39;t reach out to Admin to see other registered workflows). In fact, subworkflows do not even need to be registered. |





 

 

 

 



<a name="flyteidl/admin/workflow_attributes.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/admin/workflow_attributes.proto



<a name="flyteidl.admin.WorkflowAttributes"></a>

### WorkflowAttributes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Unique project id for which this set of attributes will be applied. |
| domain | [string](#string) |  | Unique domain id for which this set of attributes will be applied. |
| workflow | [string](#string) |  | Workflow name for which this set of attributes will be applied. |
| matching_attributes | [MatchingAttributes](#flyteidl.admin.MatchingAttributes) |  |  |






<a name="flyteidl.admin.WorkflowAttributesDeleteRequest"></a>

### WorkflowAttributesDeleteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Unique project id which this set of attributes references. |
| domain | [string](#string) |  | Unique domain id which this set of attributes references. |
| workflow | [string](#string) |  | Workflow name which this set of attributes references. |
| resource_type | [MatchableResource](#flyteidl.admin.MatchableResource) |  |  |






<a name="flyteidl.admin.WorkflowAttributesDeleteResponse"></a>

### WorkflowAttributesDeleteResponse
Purposefully empty, may be populated in the future.






<a name="flyteidl.admin.WorkflowAttributesGetRequest"></a>

### WorkflowAttributesGetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Unique project id which this set of attributes references. |
| domain | [string](#string) |  | Unique domain id which this set of attributes references. |
| workflow | [string](#string) |  | Workflow name which this set of attributes references. |
| resource_type | [MatchableResource](#flyteidl.admin.MatchableResource) |  |  |






<a name="flyteidl.admin.WorkflowAttributesGetResponse"></a>

### WorkflowAttributesGetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [WorkflowAttributes](#flyteidl.admin.WorkflowAttributes) |  |  |






<a name="flyteidl.admin.WorkflowAttributesUpdateRequest"></a>

### WorkflowAttributesUpdateRequest
Sets custom attributes for a project, domain and workflow combination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [WorkflowAttributes](#flyteidl.admin.WorkflowAttributes) |  |  |






<a name="flyteidl.admin.WorkflowAttributesUpdateResponse"></a>

### WorkflowAttributesUpdateResponse
Purposefully empty, may be populated in the future.





 

 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

