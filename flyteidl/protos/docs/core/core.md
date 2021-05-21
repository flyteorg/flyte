# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [flyteidl/core/catalog.proto](#flyteidl/core/catalog.proto)
    - [CatalogArtifactTag](#flyteidl.core.CatalogArtifactTag)
    - [CatalogMetadata](#flyteidl.core.CatalogMetadata)
  
    - [CatalogCacheStatus](#flyteidl.core.CatalogCacheStatus)
  
- [flyteidl/core/compiler.proto](#flyteidl/core/compiler.proto)
    - [CompiledTask](#flyteidl.core.CompiledTask)
    - [CompiledWorkflow](#flyteidl.core.CompiledWorkflow)
    - [CompiledWorkflowClosure](#flyteidl.core.CompiledWorkflowClosure)
    - [ConnectionSet](#flyteidl.core.ConnectionSet)
    - [ConnectionSet.DownstreamEntry](#flyteidl.core.ConnectionSet.DownstreamEntry)
    - [ConnectionSet.IdList](#flyteidl.core.ConnectionSet.IdList)
    - [ConnectionSet.UpstreamEntry](#flyteidl.core.ConnectionSet.UpstreamEntry)
  
- [flyteidl/core/condition.proto](#flyteidl/core/condition.proto)
    - [BooleanExpression](#flyteidl.core.BooleanExpression)
    - [ComparisonExpression](#flyteidl.core.ComparisonExpression)
    - [ConjunctionExpression](#flyteidl.core.ConjunctionExpression)
    - [Operand](#flyteidl.core.Operand)
  
    - [ComparisonExpression.Operator](#flyteidl.core.ComparisonExpression.Operator)
    - [ConjunctionExpression.LogicalOperator](#flyteidl.core.ConjunctionExpression.LogicalOperator)
  
- [flyteidl/core/dynamic_job.proto](#flyteidl/core/dynamic_job.proto)
    - [DynamicJobSpec](#flyteidl.core.DynamicJobSpec)
  
- [flyteidl/core/errors.proto](#flyteidl/core/errors.proto)
    - [ContainerError](#flyteidl.core.ContainerError)
    - [ErrorDocument](#flyteidl.core.ErrorDocument)
  
    - [ContainerError.Kind](#flyteidl.core.ContainerError.Kind)
  
- [flyteidl/core/execution.proto](#flyteidl/core/execution.proto)
    - [ExecutionError](#flyteidl.core.ExecutionError)
    - [NodeExecution](#flyteidl.core.NodeExecution)
    - [QualityOfService](#flyteidl.core.QualityOfService)
    - [QualityOfServiceSpec](#flyteidl.core.QualityOfServiceSpec)
    - [TaskExecution](#flyteidl.core.TaskExecution)
    - [TaskLog](#flyteidl.core.TaskLog)
    - [WorkflowExecution](#flyteidl.core.WorkflowExecution)
  
    - [ExecutionError.ErrorKind](#flyteidl.core.ExecutionError.ErrorKind)
    - [NodeExecution.Phase](#flyteidl.core.NodeExecution.Phase)
    - [QualityOfService.Tier](#flyteidl.core.QualityOfService.Tier)
    - [TaskExecution.Phase](#flyteidl.core.TaskExecution.Phase)
    - [TaskLog.MessageFormat](#flyteidl.core.TaskLog.MessageFormat)
    - [WorkflowExecution.Phase](#flyteidl.core.WorkflowExecution.Phase)
  
- [flyteidl/core/identifier.proto](#flyteidl/core/identifier.proto)
    - [Identifier](#flyteidl.core.Identifier)
    - [NodeExecutionIdentifier](#flyteidl.core.NodeExecutionIdentifier)
    - [TaskExecutionIdentifier](#flyteidl.core.TaskExecutionIdentifier)
    - [WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier)
  
    - [ResourceType](#flyteidl.core.ResourceType)
  
- [flyteidl/core/interface.proto](#flyteidl/core/interface.proto)
    - [Parameter](#flyteidl.core.Parameter)
    - [ParameterMap](#flyteidl.core.ParameterMap)
    - [ParameterMap.ParametersEntry](#flyteidl.core.ParameterMap.ParametersEntry)
    - [TypedInterface](#flyteidl.core.TypedInterface)
    - [Variable](#flyteidl.core.Variable)
    - [VariableMap](#flyteidl.core.VariableMap)
    - [VariableMap.VariablesEntry](#flyteidl.core.VariableMap.VariablesEntry)
  
- [flyteidl/core/literals.proto](#flyteidl/core/literals.proto)
    - [Binary](#flyteidl.core.Binary)
    - [Binding](#flyteidl.core.Binding)
    - [BindingData](#flyteidl.core.BindingData)
    - [BindingDataCollection](#flyteidl.core.BindingDataCollection)
    - [BindingDataMap](#flyteidl.core.BindingDataMap)
    - [BindingDataMap.BindingsEntry](#flyteidl.core.BindingDataMap.BindingsEntry)
    - [Blob](#flyteidl.core.Blob)
    - [BlobMetadata](#flyteidl.core.BlobMetadata)
    - [KeyValuePair](#flyteidl.core.KeyValuePair)
    - [Literal](#flyteidl.core.Literal)
    - [LiteralCollection](#flyteidl.core.LiteralCollection)
    - [LiteralMap](#flyteidl.core.LiteralMap)
    - [LiteralMap.LiteralsEntry](#flyteidl.core.LiteralMap.LiteralsEntry)
    - [Primitive](#flyteidl.core.Primitive)
    - [RetryStrategy](#flyteidl.core.RetryStrategy)
    - [Scalar](#flyteidl.core.Scalar)
    - [Schema](#flyteidl.core.Schema)
    - [Void](#flyteidl.core.Void)
  
- [flyteidl/core/security.proto](#flyteidl/core/security.proto)
    - [Identity](#flyteidl.core.Identity)
    - [OAuth2Client](#flyteidl.core.OAuth2Client)
    - [OAuth2TokenRequest](#flyteidl.core.OAuth2TokenRequest)
    - [Secret](#flyteidl.core.Secret)
    - [SecurityContext](#flyteidl.core.SecurityContext)
  
    - [OAuth2TokenRequest.Type](#flyteidl.core.OAuth2TokenRequest.Type)
    - [Secret.MountType](#flyteidl.core.Secret.MountType)
  
- [flyteidl/core/tasks.proto](#flyteidl/core/tasks.proto)
    - [Container](#flyteidl.core.Container)
    - [ContainerPort](#flyteidl.core.ContainerPort)
    - [DataLoadingConfig](#flyteidl.core.DataLoadingConfig)
    - [IOStrategy](#flyteidl.core.IOStrategy)
    - [Resources](#flyteidl.core.Resources)
    - [Resources.ResourceEntry](#flyteidl.core.Resources.ResourceEntry)
    - [RuntimeMetadata](#flyteidl.core.RuntimeMetadata)
    - [TaskMetadata](#flyteidl.core.TaskMetadata)
    - [TaskTemplate](#flyteidl.core.TaskTemplate)
    - [TaskTemplate.ConfigEntry](#flyteidl.core.TaskTemplate.ConfigEntry)
  
    - [DataLoadingConfig.LiteralMapFormat](#flyteidl.core.DataLoadingConfig.LiteralMapFormat)
    - [IOStrategy.DownloadMode](#flyteidl.core.IOStrategy.DownloadMode)
    - [IOStrategy.UploadMode](#flyteidl.core.IOStrategy.UploadMode)
    - [Resources.ResourceName](#flyteidl.core.Resources.ResourceName)
    - [RuntimeMetadata.RuntimeType](#flyteidl.core.RuntimeMetadata.RuntimeType)
  
- [flyteidl/core/types.proto](#flyteidl/core/types.proto)
    - [BlobType](#flyteidl.core.BlobType)
    - [Error](#flyteidl.core.Error)
    - [LiteralType](#flyteidl.core.LiteralType)
    - [OutputReference](#flyteidl.core.OutputReference)
    - [SchemaType](#flyteidl.core.SchemaType)
    - [SchemaType.SchemaColumn](#flyteidl.core.SchemaType.SchemaColumn)
  
    - [BlobType.BlobDimensionality](#flyteidl.core.BlobType.BlobDimensionality)
    - [SchemaType.SchemaColumn.SchemaColumnType](#flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType)
    - [SimpleType](#flyteidl.core.SimpleType)
  
- [flyteidl/core/workflow.proto](#flyteidl/core/workflow.proto)
    - [Alias](#flyteidl.core.Alias)
    - [BranchNode](#flyteidl.core.BranchNode)
    - [IfBlock](#flyteidl.core.IfBlock)
    - [IfElseBlock](#flyteidl.core.IfElseBlock)
    - [Node](#flyteidl.core.Node)
    - [NodeMetadata](#flyteidl.core.NodeMetadata)
    - [TaskNode](#flyteidl.core.TaskNode)
    - [WorkflowMetadata](#flyteidl.core.WorkflowMetadata)
    - [WorkflowMetadataDefaults](#flyteidl.core.WorkflowMetadataDefaults)
    - [WorkflowNode](#flyteidl.core.WorkflowNode)
    - [WorkflowTemplate](#flyteidl.core.WorkflowTemplate)
  
    - [WorkflowMetadata.OnFailurePolicy](#flyteidl.core.WorkflowMetadata.OnFailurePolicy)
  
- [flyteidl/core/workflow_closure.proto](#flyteidl/core/workflow_closure.proto)
    - [WorkflowClosure](#flyteidl.core.WorkflowClosure)
  
- [Scalar Value Types](#scalar-value-types)



<a name="flyteidl/core/catalog.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/catalog.proto



<a name="flyteidl.core.CatalogArtifactTag"></a>

### CatalogArtifactTag



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| artifact_id | [string](#string) |  | Artifact ID is generated name |
| name | [string](#string) |  | Flyte computes the tag automatically, as the hash of the values |






<a name="flyteidl.core.CatalogMetadata"></a>

### CatalogMetadata
Catalog artifact information with specific metadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dataset_id | [Identifier](#flyteidl.core.Identifier) |  | Dataset ID in the catalog |
| artifact_tag | [CatalogArtifactTag](#flyteidl.core.CatalogArtifactTag) |  | Artifact tag in the catalog |
| source_task_execution | [TaskExecutionIdentifier](#flyteidl.core.TaskExecutionIdentifier) |  | Today we only support TaskExecutionIdentifier as a source, as catalog caching only works for task executions |





 


<a name="flyteidl.core.CatalogCacheStatus"></a>

### CatalogCacheStatus
Indicates the status of CatalogCaching. The reason why this is not embeded in TaskNodeMetadata is, that we may use for other types of nodes as well in the future

| Name | Number | Description |
| ---- | ------ | ----------- |
| CACHE_DISABLED | 0 | Used to indicate that caching was disabled |
| CACHE_MISS | 1 | Used to indicate that the cache lookup resulted in no matches |
| CACHE_HIT | 2 | used to indicate that the associated artifact was a result of a previous execution |
| CACHE_POPULATED | 3 | used to indicate that the resultant artifact was added to the cache |
| CACHE_LOOKUP_FAILURE | 4 | Used to indicate that cache lookup failed because of an error |
| CACHE_PUT_FAILURE | 5 | Used to indicate that cache lookup failed because of an error |


 

 

 



<a name="flyteidl/core/compiler.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/compiler.proto



<a name="flyteidl.core.CompiledTask"></a>

### CompiledTask
Output of the Compilation step. This object represent one Task. We store more metadata at this layer


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [TaskTemplate](#flyteidl.core.TaskTemplate) |  | Completely contained TaskTemplate |






<a name="flyteidl.core.CompiledWorkflow"></a>

### CompiledWorkflow
Output of the compilation Step. This object represents one workflow. We store more metadata at this layer


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [WorkflowTemplate](#flyteidl.core.WorkflowTemplate) |  | Completely contained Workflow Template |
| connections | [ConnectionSet](#flyteidl.core.ConnectionSet) |  | For internal use only! This field is used by the system and must not be filled in. Any values set will be ignored. |






<a name="flyteidl.core.CompiledWorkflowClosure"></a>

### CompiledWorkflowClosure
A Compiled Workflow Closure contains all the information required to start a new execution, or to visualize a workflow
and its details. The CompiledWorkflowClosure should always contain a primary workflow, that is the main workflow that
will being the execution. All subworkflows are denormalized. WorkflowNodes refer to the workflow identifiers of
compiled subworkflows.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primary | [CompiledWorkflow](#flyteidl.core.CompiledWorkflow) |  | &#43;required |
| sub_workflows | [CompiledWorkflow](#flyteidl.core.CompiledWorkflow) | repeated | Guaranteed that there will only exist one and only one workflow with a given id, i.e., every sub workflow has a unique identifier. Also every enclosed subworkflow is used either by a primary workflow or by a subworkflow as an inlined workflow &#43;optional |
| tasks | [CompiledTask](#flyteidl.core.CompiledTask) | repeated | Guaranteed that there will only exist one and only one task with a given id, i.e., every task has a unique id &#43;required (atleast 1) |






<a name="flyteidl.core.ConnectionSet"></a>

### ConnectionSet
Adjacency list for the workflow. This is created as part of the compilation process. Every process after the compilation
step uses this created ConnectionSet


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| downstream | [ConnectionSet.DownstreamEntry](#flyteidl.core.ConnectionSet.DownstreamEntry) | repeated | A list of all the node ids that are downstream from a given node id |
| upstream | [ConnectionSet.UpstreamEntry](#flyteidl.core.ConnectionSet.UpstreamEntry) | repeated | A list of all the node ids, that are upstream of this node id |






<a name="flyteidl.core.ConnectionSet.DownstreamEntry"></a>

### ConnectionSet.DownstreamEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ConnectionSet.IdList](#flyteidl.core.ConnectionSet.IdList) |  |  |






<a name="flyteidl.core.ConnectionSet.IdList"></a>

### ConnectionSet.IdList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [string](#string) | repeated |  |






<a name="flyteidl.core.ConnectionSet.UpstreamEntry"></a>

### ConnectionSet.UpstreamEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ConnectionSet.IdList](#flyteidl.core.ConnectionSet.IdList) |  |  |





 

 

 

 



<a name="flyteidl/core/condition.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/condition.proto



<a name="flyteidl.core.BooleanExpression"></a>

### BooleanExpression
Defines a boolean expression tree. It can be a simple or a conjunction expression.
Multiple expressions can be combined using a conjunction or a disjunction to result in a final boolean result.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| conjunction | [ConjunctionExpression](#flyteidl.core.ConjunctionExpression) |  |  |
| comparison | [ComparisonExpression](#flyteidl.core.ComparisonExpression) |  |  |






<a name="flyteidl.core.ComparisonExpression"></a>

### ComparisonExpression
Defines a 2-level tree where the root is a comparison operator and Operands are primitives or known variables.
Each expression results in a boolean result.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operator | [ComparisonExpression.Operator](#flyteidl.core.ComparisonExpression.Operator) |  |  |
| left_value | [Operand](#flyteidl.core.Operand) |  |  |
| right_value | [Operand](#flyteidl.core.Operand) |  |  |






<a name="flyteidl.core.ConjunctionExpression"></a>

### ConjunctionExpression
Defines a conjunction expression of two boolean expressions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operator | [ConjunctionExpression.LogicalOperator](#flyteidl.core.ConjunctionExpression.LogicalOperator) |  |  |
| left_expression | [BooleanExpression](#flyteidl.core.BooleanExpression) |  |  |
| right_expression | [BooleanExpression](#flyteidl.core.BooleanExpression) |  |  |






<a name="flyteidl.core.Operand"></a>

### Operand
Defines an operand to a comparison expression.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive | [Primitive](#flyteidl.core.Primitive) |  | Can be a constant |
| var | [string](#string) |  | Or one of this node&#39;s input variables |





 


<a name="flyteidl.core.ComparisonExpression.Operator"></a>

### ComparisonExpression.Operator
Binary Operator for each expression

| Name | Number | Description |
| ---- | ------ | ----------- |
| EQ | 0 |  |
| NEQ | 1 |  |
| GT | 2 | Greater Than |
| GTE | 3 |  |
| LT | 4 | Less Than |
| LTE | 5 |  |



<a name="flyteidl.core.ConjunctionExpression.LogicalOperator"></a>

### ConjunctionExpression.LogicalOperator
Nested conditions. They can be conjoined using AND / OR
Order of evaluation is not important as the operators are Commutative

| Name | Number | Description |
| ---- | ------ | ----------- |
| AND | 0 | Conjunction |
| OR | 1 |  |


 

 

 



<a name="flyteidl/core/dynamic_job.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/dynamic_job.proto



<a name="flyteidl.core.DynamicJobSpec"></a>

### DynamicJobSpec
Describes a set of tasks to execute and how the final outputs are produced.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nodes | [Node](#flyteidl.core.Node) | repeated | A collection of nodes to execute. |
| min_successes | [int64](#int64) |  | An absolute number of successful completions of nodes required to mark this job as succeeded. As soon as this criteria is met, the dynamic job will be marked as successful and outputs will be computed. If this number becomes impossible to reach (e.g. number of currently running tasks &#43; number of already succeeded tasks &lt; min_successes) the task will be aborted immediately and marked as failed. The default value of this field, if not specified, is the count of nodes repeated field. |
| outputs | [Binding](#flyteidl.core.Binding) | repeated | Describes how to bind the final output of the dynamic job from the outputs of executed nodes. The referenced ids in bindings should have the generated id for the subtask. |
| tasks | [TaskTemplate](#flyteidl.core.TaskTemplate) | repeated | [Optional] A complete list of task specs referenced in nodes. |
| subworkflows | [WorkflowTemplate](#flyteidl.core.WorkflowTemplate) | repeated | [Optional] A complete list of task specs referenced in nodes. |





 

 

 

 



<a name="flyteidl/core/errors.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/errors.proto



<a name="flyteidl.core.ContainerError"></a>

### ContainerError
Error message to propagate detailed errors from container executions to the execution
engine.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  | A simplified code for errors, so that we can provide a glossary of all possible errors. |
| message | [string](#string) |  | A detailed error message. |
| kind | [ContainerError.Kind](#flyteidl.core.ContainerError.Kind) |  | An abstract error kind for this error. Defaults to Non_Recoverable if not specified. |
| origin | [ExecutionError.ErrorKind](#flyteidl.core.ExecutionError.ErrorKind) |  | Defines the origin of the error (system, user, unknown). |






<a name="flyteidl.core.ErrorDocument"></a>

### ErrorDocument
Defines the errors.pb file format the container can produce to communicate
failure reasons to the execution engine.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| error | [ContainerError](#flyteidl.core.ContainerError) |  | The error raised during execution. |





 


<a name="flyteidl.core.ContainerError.Kind"></a>

### ContainerError.Kind
Defines a generic error type that dictates the behavior of the retry strategy.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NON_RECOVERABLE | 0 |  |
| RECOVERABLE | 1 |  |


 

 

 



<a name="flyteidl/core/execution.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/execution.proto



<a name="flyteidl.core.ExecutionError"></a>

### ExecutionError
Represents the error message from the execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  | Error code indicates a grouping of a type of error. More Info: &lt;Link&gt; |
| message | [string](#string) |  | Detailed description of the error - including stack trace. |
| error_uri | [string](#string) |  | Full error contents accessible via a URI |
| kind | [ExecutionError.ErrorKind](#flyteidl.core.ExecutionError.ErrorKind) |  |  |






<a name="flyteidl.core.NodeExecution"></a>

### NodeExecution
Indicates various phases of Node Execution






<a name="flyteidl.core.QualityOfService"></a>

### QualityOfService
Indicates the priority of an execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tier | [QualityOfService.Tier](#flyteidl.core.QualityOfService.Tier) |  |  |
| spec | [QualityOfServiceSpec](#flyteidl.core.QualityOfServiceSpec) |  |  |






<a name="flyteidl.core.QualityOfServiceSpec"></a>

### QualityOfServiceSpec
Represents customized execution run-time attributes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| queueing_budget | [google.protobuf.Duration](#google.protobuf.Duration) |  | Indicates how much queueing delay an execution can tolerate. |






<a name="flyteidl.core.TaskExecution"></a>

### TaskExecution
Phases that task plugins can go through. Not all phases may be applicable to a specific plugin task,
but this is the cumulative list that customers may want to know about for their task.






<a name="flyteidl.core.TaskLog"></a>

### TaskLog
Log information for the task that is specific to a log sink
When our log story is flushed out, we may have more metadata here like log link expiry


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | [string](#string) |  |  |
| name | [string](#string) |  |  |
| message_format | [TaskLog.MessageFormat](#flyteidl.core.TaskLog.MessageFormat) |  |  |
| ttl | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="flyteidl.core.WorkflowExecution"></a>

### WorkflowExecution
Indicates various phases of Workflow Execution





 


<a name="flyteidl.core.ExecutionError.ErrorKind"></a>

### ExecutionError.ErrorKind
Error type: System or User

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| USER | 1 |  |
| SYSTEM | 2 |  |



<a name="flyteidl.core.NodeExecution.Phase"></a>

### NodeExecution.Phase


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNDEFINED | 0 |  |
| QUEUED | 1 |  |
| RUNNING | 2 |  |
| SUCCEEDED | 3 |  |
| FAILING | 4 |  |
| FAILED | 5 |  |
| ABORTED | 6 |  |
| SKIPPED | 7 |  |
| TIMED_OUT | 8 |  |
| DYNAMIC_RUNNING | 9 |  |



<a name="flyteidl.core.QualityOfService.Tier"></a>

### QualityOfService.Tier


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNDEFINED | 0 | Default: no quality of service specified. |
| HIGH | 1 |  |
| MEDIUM | 2 |  |
| LOW | 3 |  |



<a name="flyteidl.core.TaskExecution.Phase"></a>

### TaskExecution.Phase


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNDEFINED | 0 |  |
| QUEUED | 1 |  |
| RUNNING | 2 |  |
| SUCCEEDED | 3 |  |
| ABORTED | 4 |  |
| FAILED | 5 |  |
| INITIALIZING | 6 | To indicate cases where task is initializing, like: ErrImagePull, ContainerCreating, PodInitializing |
| WAITING_FOR_RESOURCES | 7 | To address cases, where underlying resource is not available: Backoff error, Resource quota exceeded |



<a name="flyteidl.core.TaskLog.MessageFormat"></a>

### TaskLog.MessageFormat


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| CSV | 1 |  |
| JSON | 2 |  |



<a name="flyteidl.core.WorkflowExecution.Phase"></a>

### WorkflowExecution.Phase


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNDEFINED | 0 |  |
| QUEUED | 1 |  |
| RUNNING | 2 |  |
| SUCCEEDING | 3 |  |
| SUCCEEDED | 4 |  |
| FAILING | 5 |  |
| FAILED | 6 |  |
| ABORTED | 7 |  |
| TIMED_OUT | 8 |  |


 

 

 



<a name="flyteidl/core/identifier.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/identifier.proto



<a name="flyteidl.core.Identifier"></a>

### Identifier
Encapsulation of fields that uniquely identifies a Flyte resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_type | [ResourceType](#flyteidl.core.ResourceType) |  | Identifies the specific type of resource that this identifer corresponds to. |
| project | [string](#string) |  | Name of the project the resource belongs to. |
| domain | [string](#string) |  | Name of the domain the resource belongs to. A domain can be considered as a subset within a specific project. |
| name | [string](#string) |  | User provided value for the resource. |
| version | [string](#string) |  | Specific version of the resource. |






<a name="flyteidl.core.NodeExecutionIdentifier"></a>

### NodeExecutionIdentifier
Encapsulation of fields that identify a Flyte node execution entity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node_id | [string](#string) |  |  |
| execution_id | [WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  |  |






<a name="flyteidl.core.TaskExecutionIdentifier"></a>

### TaskExecutionIdentifier
Encapsulation of fields that identify a Flyte task execution entity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_id | [Identifier](#flyteidl.core.Identifier) |  |  |
| node_execution_id | [NodeExecutionIdentifier](#flyteidl.core.NodeExecutionIdentifier) |  |  |
| retry_attempt | [uint32](#uint32) |  |  |






<a name="flyteidl.core.WorkflowExecutionIdentifier"></a>

### WorkflowExecutionIdentifier
Encapsulation of fields that uniquely identifies a Flyte workflow execution


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | Name of the project the resource belongs to. |
| domain | [string](#string) |  | Name of the domain the resource belongs to. A domain can be considered as a subset within a specific project. |
| name | [string](#string) |  | User or system provided value for the resource. |





 


<a name="flyteidl.core.ResourceType"></a>

### ResourceType
Indicates a resource type within Flyte.

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNSPECIFIED | 0 |  |
| TASK | 1 |  |
| WORKFLOW | 2 |  |
| LAUNCH_PLAN | 3 |  |
| DATASET | 4 | A dataset represents an entity modeled in Flyte DataCatalog. A Dataset is also a versioned entity and can be a compilation of multiple individual objects. Eventually all Catalog objects should be modeled similar to Flyte Objects. The Dataset entities makes it possible for the UI and CLI to act on the objects in a similar manner to other Flyte objects |


 

 

 



<a name="flyteidl/core/interface.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/interface.proto



<a name="flyteidl.core.Parameter"></a>

### Parameter
A parameter is used as input to a launch plan and has
the special ability to have a default value or mark itself as required.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| var | [Variable](#flyteidl.core.Variable) |  | &#43;required Variable. Defines the type of the variable backing this parameter. |
| default | [Literal](#flyteidl.core.Literal) |  | Defines a default value that has to match the variable type defined. |
| required | [bool](#bool) |  | &#43;optional, is this value required to be filled. |






<a name="flyteidl.core.ParameterMap"></a>

### ParameterMap
A map of Parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| parameters | [ParameterMap.ParametersEntry](#flyteidl.core.ParameterMap.ParametersEntry) | repeated | Defines a map of parameter names to parameters. |






<a name="flyteidl.core.ParameterMap.ParametersEntry"></a>

### ParameterMap.ParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Parameter](#flyteidl.core.Parameter) |  |  |






<a name="flyteidl.core.TypedInterface"></a>

### TypedInterface
Defines strongly typed inputs and outputs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inputs | [VariableMap](#flyteidl.core.VariableMap) |  |  |
| outputs | [VariableMap](#flyteidl.core.VariableMap) |  |  |






<a name="flyteidl.core.Variable"></a>

### Variable
Defines a strongly typed variable.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [LiteralType](#flyteidl.core.LiteralType) |  | Variable literal type. |
| description | [string](#string) |  | &#43;optional string describing input variable |






<a name="flyteidl.core.VariableMap"></a>

### VariableMap
A map of Variables


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| variables | [VariableMap.VariablesEntry](#flyteidl.core.VariableMap.VariablesEntry) | repeated | Defines a map of variable names to variables. |






<a name="flyteidl.core.VariableMap.VariablesEntry"></a>

### VariableMap.VariablesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Variable](#flyteidl.core.Variable) |  |  |





 

 

 

 



<a name="flyteidl/core/literals.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/literals.proto



<a name="flyteidl.core.Binary"></a>

### Binary
A simple byte array with a tag to help different parts of the system communicate about what is in the byte array.
It&#39;s strongly advisable that consumers of this type define a unique tag and validate the tag before parsing the data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| tag | [string](#string) |  |  |






<a name="flyteidl.core.Binding"></a>

### Binding
An input/output binding of a variable to either static value or a node output.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| var | [string](#string) |  | Variable name must match an input/output variable of the node. |
| binding | [BindingData](#flyteidl.core.BindingData) |  | Data to use to bind this variable. |






<a name="flyteidl.core.BindingData"></a>

### BindingData
Specifies either a simple value or a reference to another output.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| scalar | [Scalar](#flyteidl.core.Scalar) |  | A simple scalar value. |
| collection | [BindingDataCollection](#flyteidl.core.BindingDataCollection) |  | A collection of binding data. This allows nesting of binding data to any number of levels. |
| promise | [OutputReference](#flyteidl.core.OutputReference) |  | References an output promised by another node. |
| map | [BindingDataMap](#flyteidl.core.BindingDataMap) |  | A map of bindings. The key is always a string. |






<a name="flyteidl.core.BindingDataCollection"></a>

### BindingDataCollection
A collection of BindingData items.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bindings | [BindingData](#flyteidl.core.BindingData) | repeated |  |






<a name="flyteidl.core.BindingDataMap"></a>

### BindingDataMap
A map of BindingData items.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bindings | [BindingDataMap.BindingsEntry](#flyteidl.core.BindingDataMap.BindingsEntry) | repeated |  |






<a name="flyteidl.core.BindingDataMap.BindingsEntry"></a>

### BindingDataMap.BindingsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [BindingData](#flyteidl.core.BindingData) |  |  |






<a name="flyteidl.core.Blob"></a>

### Blob
Refers to an offloaded set of files. It encapsulates the type of the store and a unique uri for where the data is.
There are no restrictions on how the uri is formatted since it will depend on how to interact with the store.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [BlobMetadata](#flyteidl.core.BlobMetadata) |  |  |
| uri | [string](#string) |  |  |






<a name="flyteidl.core.BlobMetadata"></a>

### BlobMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [BlobType](#flyteidl.core.BlobType) |  |  |






<a name="flyteidl.core.KeyValuePair"></a>

### KeyValuePair
A generic key value pair.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  | required. |
| value | [string](#string) |  | &#43;optional. |






<a name="flyteidl.core.Literal"></a>

### Literal
A simple value. This supports any level of nesting (e.g. array of array of array of Blobs) as well as simple primitives.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| scalar | [Scalar](#flyteidl.core.Scalar) |  | A simple value. |
| collection | [LiteralCollection](#flyteidl.core.LiteralCollection) |  | A collection of literals to allow nesting. |
| map | [LiteralMap](#flyteidl.core.LiteralMap) |  | A map of strings to literals. |






<a name="flyteidl.core.LiteralCollection"></a>

### LiteralCollection
A collection of literals. This is a workaround since oneofs in proto messages cannot contain a repeated field.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| literals | [Literal](#flyteidl.core.Literal) | repeated |  |






<a name="flyteidl.core.LiteralMap"></a>

### LiteralMap
A map of literals. This is a workaround since oneofs in proto messages cannot contain a repeated field.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| literals | [LiteralMap.LiteralsEntry](#flyteidl.core.LiteralMap.LiteralsEntry) | repeated |  |






<a name="flyteidl.core.LiteralMap.LiteralsEntry"></a>

### LiteralMap.LiteralsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Literal](#flyteidl.core.Literal) |  |  |






<a name="flyteidl.core.Primitive"></a>

### Primitive
Primitive Types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| integer | [int64](#int64) |  |  |
| float_value | [double](#double) |  |  |
| string_value | [string](#string) |  |  |
| boolean | [bool](#bool) |  |  |
| datetime | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| duration | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="flyteidl.core.RetryStrategy"></a>

### RetryStrategy
Retry strategy associated with an executable unit.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| retries | [uint32](#uint32) |  | Number of retries. Retries will be consumed when the job fails with a recoverable error. The number of retries must be less than or equals to 10. |






<a name="flyteidl.core.Scalar"></a>

### Scalar



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive | [Primitive](#flyteidl.core.Primitive) |  |  |
| blob | [Blob](#flyteidl.core.Blob) |  |  |
| binary | [Binary](#flyteidl.core.Binary) |  |  |
| schema | [Schema](#flyteidl.core.Schema) |  |  |
| none_type | [Void](#flyteidl.core.Void) |  |  |
| error | [Error](#flyteidl.core.Error) |  |  |
| generic | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |






<a name="flyteidl.core.Schema"></a>

### Schema
A strongly typed schema that defines the interface of data retrieved from the underlying storage medium.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | [string](#string) |  |  |
| type | [SchemaType](#flyteidl.core.SchemaType) |  |  |






<a name="flyteidl.core.Void"></a>

### Void
Used to denote a nil/null/None assignment to a scalar value. The underlying LiteralType for Void is intentionally
undefined since it can be assigned to a scalar of any LiteralType.





 

 

 

 



<a name="flyteidl/core/security.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/security.proto



<a name="flyteidl.core.Identity"></a>

### Identity
Identity encapsulates the various security identities a task can run as. It&#39;s up to the underlying plugin to pick the
right identity for the execution environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| iam_role | [string](#string) |  | iam_role references the fully qualified name of Identity &amp; Access Management role to impersonate. |
| k8s_service_account | [string](#string) |  | k8s_service_account references a kubernetes service account to impersonate. |
| oauth2_client | [OAuth2Client](#flyteidl.core.OAuth2Client) |  | oauth2_client references an oauth2 client. Backend plugins can use this information to impersonate the client when making external calls. |






<a name="flyteidl.core.OAuth2Client"></a>

### OAuth2Client
OAuth2Client encapsulates OAuth2 Client Credentials to be used when making calls on behalf of that task.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_id | [string](#string) |  | client_id is the public id for the client to use. The system will not perform any pre-auth validation that the secret requested matches the client_id indicated here. &#43;required |
| client_secret | [Secret](#flyteidl.core.Secret) |  | client_secret is a reference to the secret used to authenticate the OAuth2 client. &#43;required |






<a name="flyteidl.core.OAuth2TokenRequest"></a>

### OAuth2TokenRequest
OAuth2TokenRequest encapsulates information needed to request an OAuth2 token.
FLYTE_TOKENS_ENV_PREFIX will be passed to indicate the prefix of the environment variables that will be present if
tokens are passed through environment variables.
FLYTE_TOKENS_PATH_PREFIX will be passed to indicate the prefix of the path where secrets will be mounted if tokens
are passed through file mounts.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name indicates a unique id for the token request within this task token requests. It&#39;ll be used as a suffix for environment variables and as a filename for mounting tokens as files. &#43;required |
| type | [OAuth2TokenRequest.Type](#flyteidl.core.OAuth2TokenRequest.Type) |  | type indicates the type of the request to make. Defaults to CLIENT_CREDENTIALS. &#43;required |
| client | [OAuth2Client](#flyteidl.core.OAuth2Client) |  | client references the client_id/secret to use to request the OAuth2 token. &#43;required |
| idp_discovery_endpoint | [string](#string) |  | idp_discovery_endpoint references the discovery endpoint used to retrieve token endpoint and other related information. &#43;optional |
| token_endpoint | [string](#string) |  | token_endpoint references the token issuance endpoint. If idp_discovery_endpoint is not provided, this parameter is mandatory. &#43;optional |






<a name="flyteidl.core.Secret"></a>

### Secret
Secret encapsulates information about the secret a task needs to proceed. An environment variable
FLYTE_SECRETS_ENV_PREFIX will be passed to indicate the prefix of the environment variables that will be present if
secrets are passed through environment variables.
FLYTE_SECRETS_DEFAULT_DIR will be passed to indicate the prefix of the path where secrets will be mounted if secrets
are passed through file mounts.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) |  | The name of the secret group where to find the key referenced below. For K8s secrets, this should be the name of the v1/secret object. For Confidant, this should be the Credential name. For Vault, this should be the secret name. For AWS Secret Manager, this should be the name of the secret. &#43;required |
| group_version | [string](#string) |  | The group version to fetch. This is not supported in all secret management systems. It&#39;ll be ignored for the ones that do not support it. &#43;optional |
| key | [string](#string) |  | The name of the secret to mount. This has to match an existing secret in the system. It&#39;s up to the implementation of the secret management system to require case sensitivity. For K8s secrets, Confidant and Vault, this should match one of the keys inside the secret. For AWS Secret Manager, it&#39;s ignored. &#43;optional |
| mount_requirement | [Secret.MountType](#flyteidl.core.Secret.MountType) |  | mount_requirement is optional. Indicates where the secret has to be mounted. If provided, the execution will fail if the underlying key management system cannot satisfy that requirement. If not provided, the default location will depend on the key management system. &#43;optional |






<a name="flyteidl.core.SecurityContext"></a>

### SecurityContext
SecurityContext holds security attributes that apply to tasks.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| run_as | [Identity](#flyteidl.core.Identity) |  | run_as encapsulates the identity a pod should run as. If the task fills in multiple fields here, it&#39;ll be up to the backend plugin to choose the appropriate identity for the execution engine the task will run on. |
| secrets | [Secret](#flyteidl.core.Secret) | repeated | secrets indicate the list of secrets the task needs in order to proceed. Secrets will be mounted/passed to the pod as it starts. If the plugin responsible for kicking of the task will not run it on a flyte cluster (e.g. AWS Batch), it&#39;s the responsibility of the plugin to fetch the secret (which means propeller identity will need access to the secret) and to pass it to the remote execution engine. |
| tokens | [OAuth2TokenRequest](#flyteidl.core.OAuth2TokenRequest) | repeated | tokens indicate the list of token requests the task needs in order to proceed. Tokens will be mounted/passed to the pod as it starts. If the plugin responsible for kicking of the task will not run it on a flyte cluster (e.g. AWS Batch), it&#39;s the responsibility of the plugin to fetch the secret (which means propeller identity will need access to the secret) and to pass it to the remote execution engine. |





 


<a name="flyteidl.core.OAuth2TokenRequest.Type"></a>

### OAuth2TokenRequest.Type
Type of the token requested.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CLIENT_CREDENTIALS | 0 | CLIENT_CREDENTIALS indicates a 2-legged OAuth token requested using client credentials. |



<a name="flyteidl.core.Secret.MountType"></a>

### Secret.MountType


| Name | Number | Description |
| ---- | ------ | ----------- |
| ANY | 0 | Default case, indicates the client can tolerate either mounting options. |
| ENV_VAR | 1 | ENV_VAR indicates the secret needs to be mounted as an environment variable. |
| FILE | 2 | FILE indicates the secret needs to be mounted as a file. |


 

 

 



<a name="flyteidl/core/tasks.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/tasks.proto



<a name="flyteidl.core.Container"></a>

### Container



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image | [string](#string) |  | Container image url. Eg: docker/redis:latest |
| command | [string](#string) | repeated | Command to be executed, if not provided, the default entrypoint in the container image will be used. |
| args | [string](#string) | repeated | These will default to Flyte given paths. If provided, the system will not append known paths. If the task still needs flyte&#39;s inputs and outputs path, add $(FLYTE_INPUT_FILE), $(FLYTE_OUTPUT_FILE) wherever makes sense and the system will populate these before executing the container. |
| resources | [Resources](#flyteidl.core.Resources) |  | Container resources requirement as specified by the container engine. |
| env | [KeyValuePair](#flyteidl.core.KeyValuePair) | repeated | Environment variables will be set as the container is starting up. |
| config | [KeyValuePair](#flyteidl.core.KeyValuePair) | repeated | **Deprecated.** Allows extra configs to be available for the container. TODO: elaborate on how configs will become available. Deprecated, please use TaskTemplate.config instead. |
| ports | [ContainerPort](#flyteidl.core.ContainerPort) | repeated | Ports to open in the container. This feature is not supported by all execution engines. (e.g. supported on K8s but not supported on AWS Batch) Only K8s |
| data_config | [DataLoadingConfig](#flyteidl.core.DataLoadingConfig) |  | BETA: Optional configuration for DataLoading. If not specified, then default values are used. This makes it possible to to run a completely portable container, that uses inputs and outputs only from the local file-system and without having any reference to flyteidl. This is supported only on K8s at the moment. If data loading is enabled, then data will be mounted in accompanying directories specified in the DataLoadingConfig. If the directories are not specified, inputs will be mounted onto and outputs will be uploaded from a pre-determined file-system path. Refer to the documentation to understand the default paths. Only K8s |






<a name="flyteidl.core.ContainerPort"></a>

### ContainerPort
Defines port properties for a container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| container_port | [uint32](#uint32) |  | Number of port to expose on the pod&#39;s IP address. This must be a valid port number, 0 &lt; x &lt; 65536. |






<a name="flyteidl.core.DataLoadingConfig"></a>

### DataLoadingConfig
This configuration allows executing raw containers in Flyte using the Flyte CoPilot system.
Flyte CoPilot, eliminates the needs of flytekit or sdk inside the container. Any inputs required by the users container are side-loaded in the input_path
Any outputs generated by the user container - within output_path are automatically uploaded.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | Flag enables DataLoading Config. If this is not set, data loading will not be used! |
| input_path | [string](#string) |  | File system path (start at root). This folder will contain all the inputs exploded to a separate file. Example, if the input interface needs (x: int, y: blob, z: multipart_blob) and the input path is &#34;/var/flyte/inputs&#34;, then the file system will look like /var/flyte/inputs/inputs.&lt;metadata format dependent -&gt; .pb .json .yaml&gt; -&gt; Format as defined previously. The Blob and Multipart blob will reference local filesystem instead of remote locations /var/flyte/inputs/x -&gt; X is a file that contains the value of x (integer) in string format /var/flyte/inputs/y -&gt; Y is a file in Binary format /var/flyte/inputs/z/... -&gt; Note Z itself is a directory More information about the protocol - refer to docs #TODO reference docs here |
| output_path | [string](#string) |  | File system path (start at root). This folder should contain all the outputs for the task as individual files and/or an error text file |
| format | [DataLoadingConfig.LiteralMapFormat](#flyteidl.core.DataLoadingConfig.LiteralMapFormat) |  | In the inputs folder, there will be an additional summary/metadata file that contains references to all files or inlined primitive values. This format decides the actual encoding for the data. Refer to the encoding to understand the specifics of the contents and the encoding |
| io_strategy | [IOStrategy](#flyteidl.core.IOStrategy) |  |  |






<a name="flyteidl.core.IOStrategy"></a>

### IOStrategy
Strategy to use when dealing with Blob, Schema, or multipart blob data (large datasets)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| download_mode | [IOStrategy.DownloadMode](#flyteidl.core.IOStrategy.DownloadMode) |  | Mode to use to manage downloads |
| upload_mode | [IOStrategy.UploadMode](#flyteidl.core.IOStrategy.UploadMode) |  | Mode to use to manage uploads |






<a name="flyteidl.core.Resources"></a>

### Resources
A customizable interface to convey resources requested for a container. This can be interpretted differently for different
container engines.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| requests | [Resources.ResourceEntry](#flyteidl.core.Resources.ResourceEntry) | repeated | The desired set of resources requested. ResourceNames must be unique within the list. |
| limits | [Resources.ResourceEntry](#flyteidl.core.Resources.ResourceEntry) | repeated | Defines a set of bounds (e.g. min/max) within which the task can reliably run. ResourceNames must be unique within the list. |






<a name="flyteidl.core.Resources.ResourceEntry"></a>

### Resources.ResourceEntry
Encapsulates a resource name and value.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [Resources.ResourceName](#flyteidl.core.Resources.ResourceName) |  | Resource name. |
| value | [string](#string) |  | Value must be a valid k8s quantity. See https://github.com/kubernetes/apimachinery/blob/master/pkg/api/resource/quantity.go#L30-L80 |






<a name="flyteidl.core.RuntimeMetadata"></a>

### RuntimeMetadata
Runtime information. This is losely defined to allow for extensibility.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [RuntimeMetadata.RuntimeType](#flyteidl.core.RuntimeMetadata.RuntimeType) |  | Type of runtime. |
| version | [string](#string) |  | Version of the runtime. All versions should be backward compatible. However, certain cases call for version checks to ensure tighter validation or setting expectations. |
| flavor | [string](#string) |  | &#43;optional It can be used to provide extra information about the runtime (e.g. python, golang... etc.). |






<a name="flyteidl.core.TaskMetadata"></a>

### TaskMetadata
Task Metadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| discoverable | [bool](#bool) |  | Indicates whether the system should attempt to lookup this task&#39;s output to avoid duplication of work. |
| runtime | [RuntimeMetadata](#flyteidl.core.RuntimeMetadata) |  | Runtime information about the task. |
| timeout | [google.protobuf.Duration](#google.protobuf.Duration) |  | The overall timeout of a task including user-triggered retries. |
| retries | [RetryStrategy](#flyteidl.core.RetryStrategy) |  | Number of retries per task. |
| discovery_version | [string](#string) |  | Indicates a logical version to apply to this task for the purpose of discovery. |
| deprecated_error_message | [string](#string) |  | If set, this indicates that this task is deprecated. This will enable owners of tasks to notify consumers of the ending of support for a given task. |
| interruptible | [bool](#bool) |  |  |






<a name="flyteidl.core.TaskTemplate"></a>

### TaskTemplate
A Task structure that uniquely identifies a task in the system
Tasks are registered as a first step in the system.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [Identifier](#flyteidl.core.Identifier) |  | Auto generated taskId by the system. Task Id uniquely identifies this task globally. |
| type | [string](#string) |  | A predefined yet extensible Task type identifier. This can be used to customize any of the components. If no extensions are provided in the system, Flyte will resolve the this task to its TaskCategory and default the implementation registered for the TaskCategory. |
| metadata | [TaskMetadata](#flyteidl.core.TaskMetadata) |  | Extra metadata about the task. |
| interface | [TypedInterface](#flyteidl.core.TypedInterface) |  | A strongly typed interface for the task. This enables others to use this task within a workflow and gauarantees compile-time validation of the workflow to avoid costly runtime failures. |
| custom | [google.protobuf.Struct](#google.protobuf.Struct) |  | Custom data about the task. This is extensible to allow various plugins in the system. |
| container | [Container](#flyteidl.core.Container) |  |  |
| task_type_version | [int32](#int32) |  | This can be used to customize task handling at execution time for the same task type. |
| security_context | [SecurityContext](#flyteidl.core.SecurityContext) |  | security_context encapsulates security attributes requested to run this task. |
| config | [TaskTemplate.ConfigEntry](#flyteidl.core.TaskTemplate.ConfigEntry) | repeated | Metadata about the custom defined for this task. This is extensible to allow various plugins in the system to use as required. reserve the field numbers 1 through 15 for very frequently occurring message elements |






<a name="flyteidl.core.TaskTemplate.ConfigEntry"></a>

### TaskTemplate.ConfigEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


<a name="flyteidl.core.DataLoadingConfig.LiteralMapFormat"></a>

### DataLoadingConfig.LiteralMapFormat
LiteralMapFormat decides the encoding format in which the input metadata should be made available to the containers.
If the user has access to the protocol buffer definitions, it is recommended to use the PROTO format.
JSON and YAML do not need any protobuf definitions to read it
All remote references in core.LiteralMap are replaced with local filesystem references (the data is downloaded to local filesystem)

| Name | Number | Description |
| ---- | ------ | ----------- |
| JSON | 0 | JSON / YAML for the metadata (which contains inlined primitive values). The representation is inline with the standard json specification as specified - https://www.json.org/json-en.html |
| YAML | 1 |  |
| PROTO | 2 | Proto is a serialized binary of `core.LiteralMap` defined in flyteidl/core |



<a name="flyteidl.core.IOStrategy.DownloadMode"></a>

### IOStrategy.DownloadMode
Mode to use for downloading

| Name | Number | Description |
| ---- | ------ | ----------- |
| DOWNLOAD_EAGER | 0 | All data will be downloaded before the main container is executed |
| DOWNLOAD_STREAM | 1 | Data will be downloaded as a stream and an End-Of-Stream marker will be written to indicate all data has been downloaded. Refer to protocol for details |
| DO_NOT_DOWNLOAD | 2 | Large objects (offloaded) will not be downloaded |



<a name="flyteidl.core.IOStrategy.UploadMode"></a>

### IOStrategy.UploadMode
Mode to use for uploading

| Name | Number | Description |
| ---- | ------ | ----------- |
| UPLOAD_ON_EXIT | 0 | All data will be uploaded after the main container exits |
| UPLOAD_EAGER | 1 | Data will be uploaded as it appears. Refer to protocol specification for details |
| DO_NOT_UPLOAD | 2 | Data will not be uploaded, only references will be written |



<a name="flyteidl.core.Resources.ResourceName"></a>

### Resources.ResourceName
Known resource names.

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| CPU | 1 |  |
| GPU | 2 |  |
| MEMORY | 3 |  |
| STORAGE | 4 |  |



<a name="flyteidl.core.RuntimeMetadata.RuntimeType"></a>

### RuntimeMetadata.RuntimeType


| Name | Number | Description |
| ---- | ------ | ----------- |
| OTHER | 0 |  |
| FLYTE_SDK | 1 |  |


 

 

 



<a name="flyteidl/core/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/types.proto



<a name="flyteidl.core.BlobType"></a>

### BlobType
Defines type behavior for blob objects


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| format | [string](#string) |  | Format can be a free form string understood by SDK/UI etc like csv, parquet etc |
| dimensionality | [BlobType.BlobDimensionality](#flyteidl.core.BlobType.BlobDimensionality) |  |  |






<a name="flyteidl.core.Error"></a>

### Error
Represents an error thrown from a node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| failed_node_id | [string](#string) |  | The node id that threw the error. |
| message | [string](#string) |  | Error message thrown. |






<a name="flyteidl.core.LiteralType"></a>

### LiteralType
Defines a strong type to allow type checking between interfaces.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| simple | [SimpleType](#flyteidl.core.SimpleType) |  | A simple type that can be compared one-to-one with another. |
| schema | [SchemaType](#flyteidl.core.SchemaType) |  | A complex type that requires matching of inner fields. |
| collection_type | [LiteralType](#flyteidl.core.LiteralType) |  | Defines the type of the value of a collection. Only homogeneous collections are allowed. |
| map_value_type | [LiteralType](#flyteidl.core.LiteralType) |  | Defines the type of the value of a map type. The type of the key is always a string. |
| blob | [BlobType](#flyteidl.core.BlobType) |  | A blob might have specialized implementation details depending on associated metadata. |
| metadata | [google.protobuf.Struct](#google.protobuf.Struct) |  | This field contains type metadata that is descriptive of the type, but is NOT considered in type-checking. This might be used by consumers to identify special behavior or display extended information for the type. |






<a name="flyteidl.core.OutputReference"></a>

### OutputReference
A reference to an output produced by a node. The type can be retrieved -and validated- from
the underlying interface of the node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node_id | [string](#string) |  | Node id must exist at the graph layer. |
| var | [string](#string) |  | Variable name must refer to an output variable for the node. |






<a name="flyteidl.core.SchemaType"></a>

### SchemaType
Defines schema columns and types to strongly type-validate schemas interoperability.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| columns | [SchemaType.SchemaColumn](#flyteidl.core.SchemaType.SchemaColumn) | repeated | A list of ordered columns this schema comprises of. |






<a name="flyteidl.core.SchemaType.SchemaColumn"></a>

### SchemaType.SchemaColumn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | A unique name -within the schema type- for the column |
| type | [SchemaType.SchemaColumn.SchemaColumnType](#flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType) |  | The column type. This allows a limited set of types currently. |





 


<a name="flyteidl.core.BlobType.BlobDimensionality"></a>

### BlobType.BlobDimensionality


| Name | Number | Description |
| ---- | ------ | ----------- |
| SINGLE | 0 |  |
| MULTIPART | 1 |  |



<a name="flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType"></a>

### SchemaType.SchemaColumn.SchemaColumnType


| Name | Number | Description |
| ---- | ------ | ----------- |
| INTEGER | 0 |  |
| FLOAT | 1 |  |
| STRING | 2 |  |
| BOOLEAN | 3 |  |
| DATETIME | 4 |  |
| DURATION | 5 |  |



<a name="flyteidl.core.SimpleType"></a>

### SimpleType
Define a set of simple types.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 |  |
| INTEGER | 1 |  |
| FLOAT | 2 |  |
| STRING | 3 |  |
| BOOLEAN | 4 |  |
| DATETIME | 5 |  |
| DURATION | 6 |  |
| BINARY | 7 |  |
| ERROR | 8 |  |
| STRUCT | 9 |  |


 

 

 



<a name="flyteidl/core/workflow.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/workflow.proto



<a name="flyteidl.core.Alias"></a>

### Alias
Links a variable to an alias.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| var | [string](#string) |  | Must match one of the output variable names on a node. |
| alias | [string](#string) |  | A workflow-level unique alias that downstream nodes can refer to in their input. |






<a name="flyteidl.core.BranchNode"></a>

### BranchNode
BranchNode is a special node that alter the flow of the workflow graph. It allows the control flow to branch at
runtime based on a series of conditions that get evaluated on various parameters (e.g. inputs, primtives).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| if_else | [IfElseBlock](#flyteidl.core.IfElseBlock) |  | &#43;required |






<a name="flyteidl.core.IfBlock"></a>

### IfBlock
Defines a condition and the execution unit that should be executed if the condition is satisfied.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| condition | [BooleanExpression](#flyteidl.core.BooleanExpression) |  |  |
| then_node | [Node](#flyteidl.core.Node) |  |  |






<a name="flyteidl.core.IfElseBlock"></a>

### IfElseBlock
Defines a series of if/else blocks. The first branch whose condition evaluates to true is the one to execute.
If no conditions were satisfied, the else_node or the error will execute.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| case | [IfBlock](#flyteidl.core.IfBlock) |  | &#43;required. First condition to evaluate. |
| other | [IfBlock](#flyteidl.core.IfBlock) | repeated | &#43;optional. Additional branches to evaluate. |
| else_node | [Node](#flyteidl.core.Node) |  | The node to execute in case none of the branches were taken. |
| error | [Error](#flyteidl.core.Error) |  | An error to throw in case none of the branches were taken. |






<a name="flyteidl.core.Node"></a>

### Node
A Workflow graph Node. One unit of execution in the graph. Each node can be linked to a Task, a Workflow or a branch
node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | A workflow-level unique identifier that identifies this node in the workflow. &#34;inputs&#34; and &#34;outputs&#34; are reserved node ids that cannot be used by other nodes. |
| metadata | [NodeMetadata](#flyteidl.core.NodeMetadata) |  | Extra metadata about the node. |
| inputs | [Binding](#flyteidl.core.Binding) | repeated | Specifies how to bind the underlying interface&#39;s inputs. All required inputs specified in the underlying interface must be fullfilled. |
| upstream_node_ids | [string](#string) | repeated | &#43;optional Specifies execution depdendency for this node ensuring it will only get scheduled to run after all its upstream nodes have completed. This node will have an implicit depdendency on any node that appears in inputs field. |
| output_aliases | [Alias](#flyteidl.core.Alias) | repeated | &#43;optional. A node can define aliases for a subset of its outputs. This is particularly useful if different nodes need to conform to the same interface (e.g. all branches in a branch node). Downstream nodes must refer to this nodes outputs using the alias if one&#39;s specified. |
| task_node | [TaskNode](#flyteidl.core.TaskNode) |  | Information about the Task to execute in this node. |
| workflow_node | [WorkflowNode](#flyteidl.core.WorkflowNode) |  | Information about the Workflow to execute in this mode. |
| branch_node | [BranchNode](#flyteidl.core.BranchNode) |  | Information about the branch node to evaluate in this node. |






<a name="flyteidl.core.NodeMetadata"></a>

### NodeMetadata
Defines extra information about the Node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | A friendly name for the Node |
| timeout | [google.protobuf.Duration](#google.protobuf.Duration) |  | The overall timeout of a task. |
| retries | [RetryStrategy](#flyteidl.core.RetryStrategy) |  | Number of retries per task. |
| interruptible | [bool](#bool) |  |  |






<a name="flyteidl.core.TaskNode"></a>

### TaskNode
Refers to the task that the Node is to execute.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| reference_id | [Identifier](#flyteidl.core.Identifier) |  | A globally unique identifier for the task. |






<a name="flyteidl.core.WorkflowMetadata"></a>

### WorkflowMetadata
This is workflow layer metadata. These settings are only applicable to the workflow as a whole, and do not
percolate down to child entities (like tasks) launched by the workflow.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| quality_of_service | [QualityOfService](#flyteidl.core.QualityOfService) |  | Indicates the runtime priority of workflow executions. |
| on_failure | [WorkflowMetadata.OnFailurePolicy](#flyteidl.core.WorkflowMetadata.OnFailurePolicy) |  | Defines how the system should behave when a failure is detected in the workflow execution. |






<a name="flyteidl.core.WorkflowMetadataDefaults"></a>

### WorkflowMetadataDefaults
The difference between these settings and the WorkflowMetadata ones is that these are meant to be passed down to
a workflow&#39;s underlying entities (like tasks). For instance, &#39;interruptible&#39; has no meaning at the workflow layer, it
is only relevant when a task executes. The settings here are the defaults that are passed to all nodes
unless explicitly overridden at the node layer.
If you are adding a setting that applies to both the Workflow itself, and everything underneath it, it should be
added to both this object and the WorkflowMetadata object above.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| interruptible | [bool](#bool) |  | Whether child nodes of the workflow are interruptible. |






<a name="flyteidl.core.WorkflowNode"></a>

### WorkflowNode
Refers to a the workflow the node is to execute.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| launchplan_ref | [Identifier](#flyteidl.core.Identifier) |  | A globally unique identifier for the launch plan. |
| sub_workflow_ref | [Identifier](#flyteidl.core.Identifier) |  | Reference to a subworkflow, that should be defined with the compiler context |






<a name="flyteidl.core.WorkflowTemplate"></a>

### WorkflowTemplate
Flyte Workflow Structure that encapsulates task, branch and subworkflow nodes to form a statically analyzable,
directed acyclic graph.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [Identifier](#flyteidl.core.Identifier) |  | A globally unique identifier for the workflow. |
| metadata | [WorkflowMetadata](#flyteidl.core.WorkflowMetadata) |  | Extra metadata about the workflow. |
| interface | [TypedInterface](#flyteidl.core.TypedInterface) |  | Defines a strongly typed interface for the Workflow. This can include some optional parameters. |
| nodes | [Node](#flyteidl.core.Node) | repeated | A list of nodes. In addition, &#34;globals&#34; is a special reserved node id that can be used to consume workflow inputs. |
| outputs | [Binding](#flyteidl.core.Binding) | repeated | A list of output bindings that specify how to construct workflow outputs. Bindings can pull node outputs or specify literals. All workflow outputs specified in the interface field must be bound in order for the workflow to be validated. A workflow has an implicit dependency on all of its nodes to execute successfully in order to bind final outputs. Most of these outputs will be Binding&#39;s with a BindingData of type OutputReference. That is, your workflow can just have an output of some constant (`Output(5)`), but usually, the workflow will be pulling outputs from the output of a task. |
| failure_node | [Node](#flyteidl.core.Node) |  | &#43;optional A catch-all node. This node is executed whenever the execution engine determines the workflow has failed. The interface of this node must match the Workflow interface with an additional input named &#34;error&#34; of type pb.lyft.flyte.core.Error. |
| metadata_defaults | [WorkflowMetadataDefaults](#flyteidl.core.WorkflowMetadataDefaults) |  | workflow defaults |





 


<a name="flyteidl.core.WorkflowMetadata.OnFailurePolicy"></a>

### WorkflowMetadata.OnFailurePolicy
Failure Handling Strategy

| Name | Number | Description |
| ---- | ------ | ----------- |
| FAIL_IMMEDIATELY | 0 | FAIL_IMMEDIATELY instructs the system to fail as soon as a node fails in the workflow. It&#39;ll automatically abort all currently running nodes and clean up resources before finally marking the workflow executions as failed. |
| FAIL_AFTER_EXECUTABLE_NODES_COMPLETE | 1 | FAIL_AFTER_EXECUTABLE_NODES_COMPLETE instructs the system to make as much progress as it can. The system will not alter the dependencies of the execution graph so any node that depend on the failed node will not be run. Other nodes that will be executed to completion before cleaning up resources and marking the workflow execution as failed. |


 

 

 



<a name="flyteidl/core/workflow_closure.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/core/workflow_closure.proto



<a name="flyteidl.core.WorkflowClosure"></a>

### WorkflowClosure
Defines an enclosed package of workflow and tasks it references.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workflow | [WorkflowTemplate](#flyteidl.core.WorkflowTemplate) |  | required. Workflow template. |
| tasks | [TaskTemplate](#flyteidl.core.TaskTemplate) | repeated | optional. A collection of tasks referenced by the workflow. Only needed if the workflow references tasks. |





 

 

 

 



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

