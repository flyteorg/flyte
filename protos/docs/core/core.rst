######################
Protocol Documentation
######################




.. _ref_flyteidl/core/catalog.proto:

flyteidl/core/catalog.proto
==================================================================





.. _ref_flyteidl.core.CatalogArtifactTag:

CatalogArtifactTag
------------------------------------------------------------------





.. csv-table:: CatalogArtifactTag type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "artifact_id", ":ref:`ref_string`", "", "Artifact ID is generated name"
   "name", ":ref:`ref_string`", "", "Flyte computes the tag automatically, as the hash of the values"







.. _ref_flyteidl.core.CatalogMetadata:

CatalogMetadata
------------------------------------------------------------------

Catalog artifact information with specific metadata



.. csv-table:: CatalogMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "dataset_id", ":ref:`ref_flyteidl.core.Identifier`", "", "Dataset ID in the catalog"
   "artifact_tag", ":ref:`ref_flyteidl.core.CatalogArtifactTag`", "", "Artifact tag in the catalog"
   "source_task_execution", ":ref:`ref_flyteidl.core.TaskExecutionIdentifier`", "", "Today we only support TaskExecutionIdentifier as a source, as catalog caching only works for task executions"







.. _ref_flyteidl.core.CatalogReservation:

CatalogReservation
------------------------------------------------------------------









..
   end messages



.. _ref_flyteidl.core.CatalogCacheStatus:

CatalogCacheStatus
------------------------------------------------------------------

Indicates the status of CatalogCaching. The reason why this is not embedded in TaskNodeMetadata is, that we may use for other types of nodes as well in the future

.. csv-table:: Enum CatalogCacheStatus values
   :header: "Name", "Number", "Description"
   :widths: auto

   "CACHE_DISABLED", "0", "Used to indicate that caching was disabled"
   "CACHE_MISS", "1", "Used to indicate that the cache lookup resulted in no matches"
   "CACHE_HIT", "2", "used to indicate that the associated artifact was a result of a previous execution"
   "CACHE_POPULATED", "3", "used to indicate that the resultant artifact was added to the cache"
   "CACHE_LOOKUP_FAILURE", "4", "Used to indicate that cache lookup failed because of an error"
   "CACHE_PUT_FAILURE", "5", "Used to indicate that cache lookup failed because of an error"
   "CACHE_SKIPPED", "6", "Used to indicate the cache lookup was skipped"



.. _ref_flyteidl.core.CatalogReservation.Status:

CatalogReservation.Status
------------------------------------------------------------------

Indicates the status of a catalog reservation operation.

.. csv-table:: Enum CatalogReservation.Status values
   :header: "Name", "Number", "Description"
   :widths: auto

   "RESERVATION_DISABLED", "0", "Used to indicate that reservations are disabled"
   "RESERVATION_ACQUIRED", "1", "Used to indicate that a reservation was successfully acquired or extended"
   "RESERVATION_EXISTS", "2", "Used to indicate that an active reservation currently exists"
   "RESERVATION_RELEASED", "3", "Used to indicate that the reservation has been successfully released"
   "RESERVATION_FAILURE", "4", "Used to indicate that a reservation operation resulted in failure"


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/compiler.proto:

flyteidl/core/compiler.proto
==================================================================





.. _ref_flyteidl.core.CompiledTask:

CompiledTask
------------------------------------------------------------------

Output of the Compilation step. This object represent one Task. We store more metadata at this layer



.. csv-table:: CompiledTask type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "template", ":ref:`ref_flyteidl.core.TaskTemplate`", "", "Completely contained TaskTemplate"







.. _ref_flyteidl.core.CompiledWorkflow:

CompiledWorkflow
------------------------------------------------------------------

Output of the compilation Step. This object represents one workflow. We store more metadata at this layer



.. csv-table:: CompiledWorkflow type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "template", ":ref:`ref_flyteidl.core.WorkflowTemplate`", "", "Completely contained Workflow Template"
   "connections", ":ref:`ref_flyteidl.core.ConnectionSet`", "", "For internal use only! This field is used by the system and must not be filled in. Any values set will be ignored."







.. _ref_flyteidl.core.CompiledWorkflowClosure:

CompiledWorkflowClosure
------------------------------------------------------------------

A Compiled Workflow Closure contains all the information required to start a new execution, or to visualize a workflow
and its details. The CompiledWorkflowClosure should always contain a primary workflow, that is the main workflow that
will being the execution. All subworkflows are denormalized. WorkflowNodes refer to the workflow identifiers of
compiled subworkflows.



.. csv-table:: CompiledWorkflowClosure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "primary", ":ref:`ref_flyteidl.core.CompiledWorkflow`", "", "+required"
   "sub_workflows", ":ref:`ref_flyteidl.core.CompiledWorkflow`", "repeated", "Guaranteed that there will only exist one and only one workflow with a given id, i.e., every sub workflow has a unique identifier. Also every enclosed subworkflow is used either by a primary workflow or by a subworkflow as an inlined workflow +optional"
   "tasks", ":ref:`ref_flyteidl.core.CompiledTask`", "repeated", "Guaranteed that there will only exist one and only one task with a given id, i.e., every task has a unique id +required (at least 1)"







.. _ref_flyteidl.core.ConnectionSet:

ConnectionSet
------------------------------------------------------------------

Adjacency list for the workflow. This is created as part of the compilation process. Every process after the compilation
step uses this created ConnectionSet



.. csv-table:: ConnectionSet type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "downstream", ":ref:`ref_flyteidl.core.ConnectionSet.DownstreamEntry`", "repeated", "A list of all the node ids that are downstream from a given node id"
   "upstream", ":ref:`ref_flyteidl.core.ConnectionSet.UpstreamEntry`", "repeated", "A list of all the node ids, that are upstream of this node id"







.. _ref_flyteidl.core.ConnectionSet.DownstreamEntry:

ConnectionSet.DownstreamEntry
------------------------------------------------------------------





.. csv-table:: ConnectionSet.DownstreamEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_flyteidl.core.ConnectionSet.IdList`", "", ""







.. _ref_flyteidl.core.ConnectionSet.IdList:

ConnectionSet.IdList
------------------------------------------------------------------





.. csv-table:: ConnectionSet.IdList type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "ids", ":ref:`ref_string`", "repeated", ""







.. _ref_flyteidl.core.ConnectionSet.UpstreamEntry:

ConnectionSet.UpstreamEntry
------------------------------------------------------------------





.. csv-table:: ConnectionSet.UpstreamEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_flyteidl.core.ConnectionSet.IdList`", "", ""






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/condition.proto:

flyteidl/core/condition.proto
==================================================================





.. _ref_flyteidl.core.BooleanExpression:

BooleanExpression
------------------------------------------------------------------

Defines a boolean expression tree. It can be a simple or a conjunction expression.
Multiple expressions can be combined using a conjunction or a disjunction to result in a final boolean result.



.. csv-table:: BooleanExpression type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "conjunction", ":ref:`ref_flyteidl.core.ConjunctionExpression`", "", ""
   "comparison", ":ref:`ref_flyteidl.core.ComparisonExpression`", "", ""







.. _ref_flyteidl.core.ComparisonExpression:

ComparisonExpression
------------------------------------------------------------------

Defines a 2-level tree where the root is a comparison operator and Operands are primitives or known variables.
Each expression results in a boolean result.



.. csv-table:: ComparisonExpression type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "operator", ":ref:`ref_flyteidl.core.ComparisonExpression.Operator`", "", ""
   "left_value", ":ref:`ref_flyteidl.core.Operand`", "", ""
   "right_value", ":ref:`ref_flyteidl.core.Operand`", "", ""







.. _ref_flyteidl.core.ConjunctionExpression:

ConjunctionExpression
------------------------------------------------------------------

Defines a conjunction expression of two boolean expressions.



.. csv-table:: ConjunctionExpression type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "operator", ":ref:`ref_flyteidl.core.ConjunctionExpression.LogicalOperator`", "", ""
   "left_expression", ":ref:`ref_flyteidl.core.BooleanExpression`", "", ""
   "right_expression", ":ref:`ref_flyteidl.core.BooleanExpression`", "", ""







.. _ref_flyteidl.core.Operand:

Operand
------------------------------------------------------------------

Defines an operand to a comparison expression.



.. csv-table:: Operand type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "primitive", ":ref:`ref_flyteidl.core.Primitive`", "", "Can be a constant"
   "var", ":ref:`ref_string`", "", "Or one of this node's input variables"






..
   end messages



.. _ref_flyteidl.core.ComparisonExpression.Operator:

ComparisonExpression.Operator
------------------------------------------------------------------

Binary Operator for each expression

.. csv-table:: Enum ComparisonExpression.Operator values
   :header: "Name", "Number", "Description"
   :widths: auto

   "EQ", "0", ""
   "NEQ", "1", ""
   "GT", "2", "Greater Than"
   "GTE", "3", ""
   "LT", "4", "Less Than"
   "LTE", "5", ""



.. _ref_flyteidl.core.ConjunctionExpression.LogicalOperator:

ConjunctionExpression.LogicalOperator
------------------------------------------------------------------

Nested conditions. They can be conjoined using AND / OR
Order of evaluation is not important as the operators are Commutative

.. csv-table:: Enum ConjunctionExpression.LogicalOperator values
   :header: "Name", "Number", "Description"
   :widths: auto

   "AND", "0", "Conjunction"
   "OR", "1", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/dynamic_job.proto:

flyteidl/core/dynamic_job.proto
==================================================================





.. _ref_flyteidl.core.DynamicJobSpec:

DynamicJobSpec
------------------------------------------------------------------

Describes a set of tasks to execute and how the final outputs are produced.



.. csv-table:: DynamicJobSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "nodes", ":ref:`ref_flyteidl.core.Node`", "repeated", "A collection of nodes to execute."
   "min_successes", ":ref:`ref_int64`", "", "An absolute number of successful completions of nodes required to mark this job as succeeded. As soon as this criteria is met, the dynamic job will be marked as successful and outputs will be computed. If this number becomes impossible to reach (e.g. number of currently running tasks + number of already succeeded tasks < min_successes) the task will be aborted immediately and marked as failed. The default value of this field, if not specified, is the count of nodes repeated field."
   "outputs", ":ref:`ref_flyteidl.core.Binding`", "repeated", "Describes how to bind the final output of the dynamic job from the outputs of executed nodes. The referenced ids in bindings should have the generated id for the subtask."
   "tasks", ":ref:`ref_flyteidl.core.TaskTemplate`", "repeated", "[Optional] A complete list of task specs referenced in nodes."
   "subworkflows", ":ref:`ref_flyteidl.core.WorkflowTemplate`", "repeated", "[Optional] A complete list of task specs referenced in nodes."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/errors.proto:

flyteidl/core/errors.proto
==================================================================





.. _ref_flyteidl.core.ContainerError:

ContainerError
------------------------------------------------------------------

Error message to propagate detailed errors from container executions to the execution
engine.



.. csv-table:: ContainerError type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "code", ":ref:`ref_string`", "", "A simplified code for errors, so that we can provide a glossary of all possible errors."
   "message", ":ref:`ref_string`", "", "A detailed error message."
   "kind", ":ref:`ref_flyteidl.core.ContainerError.Kind`", "", "An abstract error kind for this error. Defaults to Non_Recoverable if not specified."
   "origin", ":ref:`ref_flyteidl.core.ExecutionError.ErrorKind`", "", "Defines the origin of the error (system, user, unknown)."







.. _ref_flyteidl.core.ErrorDocument:

ErrorDocument
------------------------------------------------------------------

Defines the errors.pb file format the container can produce to communicate
failure reasons to the execution engine.



.. csv-table:: ErrorDocument type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "error", ":ref:`ref_flyteidl.core.ContainerError`", "", "The error raised during execution."






..
   end messages



.. _ref_flyteidl.core.ContainerError.Kind:

ContainerError.Kind
------------------------------------------------------------------

Defines a generic error type that dictates the behavior of the retry strategy.

.. csv-table:: Enum ContainerError.Kind values
   :header: "Name", "Number", "Description"
   :widths: auto

   "NON_RECOVERABLE", "0", ""
   "RECOVERABLE", "1", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/execution.proto:

flyteidl/core/execution.proto
==================================================================





.. _ref_flyteidl.core.ExecutionError:

ExecutionError
------------------------------------------------------------------

Represents the error message from the execution.



.. csv-table:: ExecutionError type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "code", ":ref:`ref_string`", "", "Error code indicates a grouping of a type of error. More Info: <Link>"
   "message", ":ref:`ref_string`", "", "Detailed description of the error - including stack trace."
   "error_uri", ":ref:`ref_string`", "", "Full error contents accessible via a URI"
   "kind", ":ref:`ref_flyteidl.core.ExecutionError.ErrorKind`", "", ""







.. _ref_flyteidl.core.NodeExecution:

NodeExecution
------------------------------------------------------------------

Indicates various phases of Node Execution that only include the time spent to run the nodes/workflows








.. _ref_flyteidl.core.QualityOfService:

QualityOfService
------------------------------------------------------------------

Indicates the priority of an execution.



.. csv-table:: QualityOfService type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "tier", ":ref:`ref_flyteidl.core.QualityOfService.Tier`", "", ""
   "spec", ":ref:`ref_flyteidl.core.QualityOfServiceSpec`", "", ""







.. _ref_flyteidl.core.QualityOfServiceSpec:

QualityOfServiceSpec
------------------------------------------------------------------

Represents customized execution run-time attributes.



.. csv-table:: QualityOfServiceSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "queueing_budget", ":ref:`ref_google.protobuf.Duration`", "", "Indicates how much queueing delay an execution can tolerate."







.. _ref_flyteidl.core.TaskExecution:

TaskExecution
------------------------------------------------------------------

Phases that task plugins can go through. Not all phases may be applicable to a specific plugin task,
but this is the cumulative list that customers may want to know about for their task.








.. _ref_flyteidl.core.TaskLog:

TaskLog
------------------------------------------------------------------

Log information for the task that is specific to a log sink
When our log story is flushed out, we may have more metadata here like log link expiry



.. csv-table:: TaskLog type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "uri", ":ref:`ref_string`", "", ""
   "name", ":ref:`ref_string`", "", ""
   "message_format", ":ref:`ref_flyteidl.core.TaskLog.MessageFormat`", "", ""
   "ttl", ":ref:`ref_google.protobuf.Duration`", "", ""







.. _ref_flyteidl.core.WorkflowExecution:

WorkflowExecution
------------------------------------------------------------------

Indicates various phases of Workflow Execution







..
   end messages



.. _ref_flyteidl.core.ExecutionError.ErrorKind:

ExecutionError.ErrorKind
------------------------------------------------------------------

Error type: System or User

.. csv-table:: Enum ExecutionError.ErrorKind values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNKNOWN", "0", ""
   "USER", "1", ""
   "SYSTEM", "2", ""



.. _ref_flyteidl.core.NodeExecution.Phase:

NodeExecution.Phase
------------------------------------------------------------------



.. csv-table:: Enum NodeExecution.Phase values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNDEFINED", "0", ""
   "QUEUED", "1", ""
   "RUNNING", "2", ""
   "SUCCEEDED", "3", ""
   "FAILING", "4", ""
   "FAILED", "5", ""
   "ABORTED", "6", ""
   "SKIPPED", "7", ""
   "TIMED_OUT", "8", ""
   "DYNAMIC_RUNNING", "9", ""
   "RECOVERED", "10", ""



.. _ref_flyteidl.core.QualityOfService.Tier:

QualityOfService.Tier
------------------------------------------------------------------



.. csv-table:: Enum QualityOfService.Tier values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNDEFINED", "0", "Default: no quality of service specified."
   "HIGH", "1", ""
   "MEDIUM", "2", ""
   "LOW", "3", ""



.. _ref_flyteidl.core.TaskExecution.Phase:

TaskExecution.Phase
------------------------------------------------------------------



.. csv-table:: Enum TaskExecution.Phase values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNDEFINED", "0", ""
   "QUEUED", "1", ""
   "RUNNING", "2", ""
   "SUCCEEDED", "3", ""
   "ABORTED", "4", ""
   "FAILED", "5", ""
   "INITIALIZING", "6", "To indicate cases where task is initializing, like: ErrImagePull, ContainerCreating, PodInitializing"
   "WAITING_FOR_RESOURCES", "7", "To address cases, where underlying resource is not available: Backoff error, Resource quota exceeded"



.. _ref_flyteidl.core.TaskLog.MessageFormat:

TaskLog.MessageFormat
------------------------------------------------------------------



.. csv-table:: Enum TaskLog.MessageFormat values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNKNOWN", "0", ""
   "CSV", "1", ""
   "JSON", "2", ""



.. _ref_flyteidl.core.WorkflowExecution.Phase:

WorkflowExecution.Phase
------------------------------------------------------------------



.. csv-table:: Enum WorkflowExecution.Phase values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNDEFINED", "0", ""
   "QUEUED", "1", ""
   "RUNNING", "2", ""
   "SUCCEEDING", "3", ""
   "SUCCEEDED", "4", ""
   "FAILING", "5", ""
   "FAILED", "6", ""
   "ABORTED", "7", ""
   "TIMED_OUT", "8", ""
   "ABORTING", "9", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/identifier.proto:

flyteidl/core/identifier.proto
==================================================================





.. _ref_flyteidl.core.Identifier:

Identifier
------------------------------------------------------------------

Encapsulation of fields that uniquely identifies a Flyte resource.



.. csv-table:: Identifier type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "resource_type", ":ref:`ref_flyteidl.core.ResourceType`", "", "Identifies the specific type of resource that this identifier corresponds to."
   "project", ":ref:`ref_string`", "", "Name of the project the resource belongs to."
   "domain", ":ref:`ref_string`", "", "Name of the domain the resource belongs to. A domain can be considered as a subset within a specific project."
   "name", ":ref:`ref_string`", "", "User provided value for the resource."
   "version", ":ref:`ref_string`", "", "Specific version of the resource."







.. _ref_flyteidl.core.NodeExecutionIdentifier:

NodeExecutionIdentifier
------------------------------------------------------------------

Encapsulation of fields that identify a Flyte node execution entity.



.. csv-table:: NodeExecutionIdentifier type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "node_id", ":ref:`ref_string`", "", ""
   "execution_id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", ""







.. _ref_flyteidl.core.SignalIdentifier:

SignalIdentifier
------------------------------------------------------------------

Encapsulation of fields the uniquely identify a signal.



.. csv-table:: SignalIdentifier type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signal_id", ":ref:`ref_string`", "", "Unique identifier for a signal."
   "execution_id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", "Identifies the Flyte workflow execution this signal belongs to."







.. _ref_flyteidl.core.TaskExecutionIdentifier:

TaskExecutionIdentifier
------------------------------------------------------------------

Encapsulation of fields that identify a Flyte task execution entity.



.. csv-table:: TaskExecutionIdentifier type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "task_id", ":ref:`ref_flyteidl.core.Identifier`", "", ""
   "node_execution_id", ":ref:`ref_flyteidl.core.NodeExecutionIdentifier`", "", ""
   "retry_attempt", ":ref:`ref_uint32`", "", ""







.. _ref_flyteidl.core.WorkflowExecutionIdentifier:

WorkflowExecutionIdentifier
------------------------------------------------------------------

Encapsulation of fields that uniquely identifies a Flyte workflow execution



.. csv-table:: WorkflowExecutionIdentifier type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Name of the project the resource belongs to."
   "domain", ":ref:`ref_string`", "", "Name of the domain the resource belongs to. A domain can be considered as a subset within a specific project."
   "name", ":ref:`ref_string`", "", "User or system provided value for the resource."






..
   end messages



.. _ref_flyteidl.core.ResourceType:

ResourceType
------------------------------------------------------------------

Indicates a resource type within Flyte.

.. csv-table:: Enum ResourceType values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNSPECIFIED", "0", ""
   "TASK", "1", ""
   "WORKFLOW", "2", ""
   "LAUNCH_PLAN", "3", ""
   "DATASET", "4", "A dataset represents an entity modeled in Flyte DataCatalog. A Dataset is also a versioned entity and can be a compilation of multiple individual objects. Eventually all Catalog objects should be modeled similar to Flyte Objects. The Dataset entities makes it possible for the UI and CLI to act on the objects in a similar manner to other Flyte objects"


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/interface.proto:

flyteidl/core/interface.proto
==================================================================





.. _ref_flyteidl.core.Parameter:

Parameter
------------------------------------------------------------------

A parameter is used as input to a launch plan and has
the special ability to have a default value or mark itself as required.



.. csv-table:: Parameter type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "var", ":ref:`ref_flyteidl.core.Variable`", "", "+required Variable. Defines the type of the variable backing this parameter."
   "default", ":ref:`ref_flyteidl.core.Literal`", "", "Defines a default value that has to match the variable type defined."
   "required", ":ref:`ref_bool`", "", "+optional, is this value required to be filled."







.. _ref_flyteidl.core.ParameterMap:

ParameterMap
------------------------------------------------------------------

A map of Parameters.



.. csv-table:: ParameterMap type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "parameters", ":ref:`ref_flyteidl.core.ParameterMap.ParametersEntry`", "repeated", "Defines a map of parameter names to parameters."







.. _ref_flyteidl.core.ParameterMap.ParametersEntry:

ParameterMap.ParametersEntry
------------------------------------------------------------------





.. csv-table:: ParameterMap.ParametersEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_flyteidl.core.Parameter`", "", ""







.. _ref_flyteidl.core.TypedInterface:

TypedInterface
------------------------------------------------------------------

Defines strongly typed inputs and outputs.



.. csv-table:: TypedInterface type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "inputs", ":ref:`ref_flyteidl.core.VariableMap`", "", ""
   "outputs", ":ref:`ref_flyteidl.core.VariableMap`", "", ""







.. _ref_flyteidl.core.Variable:

Variable
------------------------------------------------------------------

Defines a strongly typed variable.



.. csv-table:: Variable type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "type", ":ref:`ref_flyteidl.core.LiteralType`", "", "Variable literal type."
   "description", ":ref:`ref_string`", "", "+optional string describing input variable"







.. _ref_flyteidl.core.VariableMap:

VariableMap
------------------------------------------------------------------

A map of Variables



.. csv-table:: VariableMap type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "variables", ":ref:`ref_flyteidl.core.VariableMap.VariablesEntry`", "repeated", "Defines a map of variable names to variables."







.. _ref_flyteidl.core.VariableMap.VariablesEntry:

VariableMap.VariablesEntry
------------------------------------------------------------------





.. csv-table:: VariableMap.VariablesEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_flyteidl.core.Variable`", "", ""






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/literals.proto:

flyteidl/core/literals.proto
==================================================================





.. _ref_flyteidl.core.Binary:

Binary
------------------------------------------------------------------

A simple byte array with a tag to help different parts of the system communicate about what is in the byte array.
It's strongly advisable that consumers of this type define a unique tag and validate the tag before parsing the data.



.. csv-table:: Binary type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_bytes`", "", ""
   "tag", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.core.Binding:

Binding
------------------------------------------------------------------

An input/output binding of a variable to either static value or a node output.



.. csv-table:: Binding type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "var", ":ref:`ref_string`", "", "Variable name must match an input/output variable of the node."
   "binding", ":ref:`ref_flyteidl.core.BindingData`", "", "Data to use to bind this variable."







.. _ref_flyteidl.core.BindingData:

BindingData
------------------------------------------------------------------

Specifies either a simple value or a reference to another output.



.. csv-table:: BindingData type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "scalar", ":ref:`ref_flyteidl.core.Scalar`", "", "A simple scalar value."
   "collection", ":ref:`ref_flyteidl.core.BindingDataCollection`", "", "A collection of binding data. This allows nesting of binding data to any number of levels."
   "promise", ":ref:`ref_flyteidl.core.OutputReference`", "", "References an output promised by another node."
   "map", ":ref:`ref_flyteidl.core.BindingDataMap`", "", "A map of bindings. The key is always a string."
   "union", ":ref:`ref_flyteidl.core.UnionInfo`", "", ""







.. _ref_flyteidl.core.BindingDataCollection:

BindingDataCollection
------------------------------------------------------------------

A collection of BindingData items.



.. csv-table:: BindingDataCollection type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "bindings", ":ref:`ref_flyteidl.core.BindingData`", "repeated", ""







.. _ref_flyteidl.core.BindingDataMap:

BindingDataMap
------------------------------------------------------------------

A map of BindingData items.



.. csv-table:: BindingDataMap type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "bindings", ":ref:`ref_flyteidl.core.BindingDataMap.BindingsEntry`", "repeated", ""







.. _ref_flyteidl.core.BindingDataMap.BindingsEntry:

BindingDataMap.BindingsEntry
------------------------------------------------------------------





.. csv-table:: BindingDataMap.BindingsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_flyteidl.core.BindingData`", "", ""







.. _ref_flyteidl.core.Blob:

Blob
------------------------------------------------------------------

Refers to an offloaded set of files. It encapsulates the type of the store and a unique uri for where the data is.
There are no restrictions on how the uri is formatted since it will depend on how to interact with the store.



.. csv-table:: Blob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "metadata", ":ref:`ref_flyteidl.core.BlobMetadata`", "", ""
   "uri", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.core.BlobMetadata:

BlobMetadata
------------------------------------------------------------------





.. csv-table:: BlobMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "type", ":ref:`ref_flyteidl.core.BlobType`", "", ""







.. _ref_flyteidl.core.KeyValuePair:

KeyValuePair
------------------------------------------------------------------

A generic key value pair.



.. csv-table:: KeyValuePair type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", "required."
   "value", ":ref:`ref_string`", "", "+optional."







.. _ref_flyteidl.core.Literal:

Literal
------------------------------------------------------------------

A simple value. This supports any level of nesting (e.g. array of array of array of Blobs) as well as simple primitives.



.. csv-table:: Literal type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "scalar", ":ref:`ref_flyteidl.core.Scalar`", "", "A simple value."
   "collection", ":ref:`ref_flyteidl.core.LiteralCollection`", "", "A collection of literals to allow nesting."
   "map", ":ref:`ref_flyteidl.core.LiteralMap`", "", "A map of strings to literals."
   "hash", ":ref:`ref_string`", "", "A hash representing this literal. This is used for caching purposes. For more details refer to RFC 1893 (https://github.com/flyteorg/flyte/blob/master/rfc/system/1893-caching-of-offloaded-objects.md)"







.. _ref_flyteidl.core.LiteralCollection:

LiteralCollection
------------------------------------------------------------------

A collection of literals. This is a workaround since oneofs in proto messages cannot contain a repeated field.



.. csv-table:: LiteralCollection type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "literals", ":ref:`ref_flyteidl.core.Literal`", "repeated", ""







.. _ref_flyteidl.core.LiteralMap:

LiteralMap
------------------------------------------------------------------

A map of literals. This is a workaround since oneofs in proto messages cannot contain a repeated field.



.. csv-table:: LiteralMap type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "literals", ":ref:`ref_flyteidl.core.LiteralMap.LiteralsEntry`", "repeated", ""







.. _ref_flyteidl.core.LiteralMap.LiteralsEntry:

LiteralMap.LiteralsEntry
------------------------------------------------------------------





.. csv-table:: LiteralMap.LiteralsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_flyteidl.core.Literal`", "", ""







.. _ref_flyteidl.core.Primitive:

Primitive
------------------------------------------------------------------

Primitive Types



.. csv-table:: Primitive type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "integer", ":ref:`ref_int64`", "", ""
   "float_value", ":ref:`ref_double`", "", ""
   "string_value", ":ref:`ref_string`", "", ""
   "boolean", ":ref:`ref_bool`", "", ""
   "datetime", ":ref:`ref_google.protobuf.Timestamp`", "", ""
   "duration", ":ref:`ref_google.protobuf.Duration`", "", ""







.. _ref_flyteidl.core.RetryStrategy:

RetryStrategy
------------------------------------------------------------------

Retry strategy associated with an executable unit.



.. csv-table:: RetryStrategy type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "retries", ":ref:`ref_uint32`", "", "Number of retries. Retries will be consumed when the job fails with a recoverable error. The number of retries must be less than or equals to 10."







.. _ref_flyteidl.core.Scalar:

Scalar
------------------------------------------------------------------





.. csv-table:: Scalar type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "primitive", ":ref:`ref_flyteidl.core.Primitive`", "", ""
   "blob", ":ref:`ref_flyteidl.core.Blob`", "", ""
   "binary", ":ref:`ref_flyteidl.core.Binary`", "", ""
   "schema", ":ref:`ref_flyteidl.core.Schema`", "", ""
   "none_type", ":ref:`ref_flyteidl.core.Void`", "", ""
   "error", ":ref:`ref_flyteidl.core.Error`", "", ""
   "generic", ":ref:`ref_google.protobuf.Struct`", "", ""
   "structured_dataset", ":ref:`ref_flyteidl.core.StructuredDataset`", "", ""
   "union", ":ref:`ref_flyteidl.core.Union`", "", ""







.. _ref_flyteidl.core.Schema:

Schema
------------------------------------------------------------------

A strongly typed schema that defines the interface of data retrieved from the underlying storage medium.



.. csv-table:: Schema type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "uri", ":ref:`ref_string`", "", ""
   "type", ":ref:`ref_flyteidl.core.SchemaType`", "", ""







.. _ref_flyteidl.core.StructuredDataset:

StructuredDataset
------------------------------------------------------------------





.. csv-table:: StructuredDataset type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "uri", ":ref:`ref_string`", "", "String location uniquely identifying where the data is. Should start with the storage location (e.g. s3://, gs://, bq://, etc.)"
   "metadata", ":ref:`ref_flyteidl.core.StructuredDatasetMetadata`", "", ""







.. _ref_flyteidl.core.StructuredDatasetMetadata:

StructuredDatasetMetadata
------------------------------------------------------------------





.. csv-table:: StructuredDatasetMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "structured_dataset_type", ":ref:`ref_flyteidl.core.StructuredDatasetType`", "", "Bundle the type information along with the literal. This is here because StructuredDatasets can often be more defined at run time than at compile time. That is, at compile time you might only declare a task to return a pandas dataframe or a StructuredDataset, without any column information, but at run time, you might have that column information. flytekit python will copy this type information into the literal, from the type information, if not provided by the various plugins (encoders). Since this field is run time generated, it's not used for any type checking."







.. _ref_flyteidl.core.Union:

Union
------------------------------------------------------------------

The runtime representation of a tagged union value. See `UnionType` for more details.



.. csv-table:: Union type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "value", ":ref:`ref_flyteidl.core.Literal`", "", ""
   "type", ":ref:`ref_flyteidl.core.LiteralType`", "", ""







.. _ref_flyteidl.core.UnionInfo:

UnionInfo
------------------------------------------------------------------





.. csv-table:: UnionInfo type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "targetType", ":ref:`ref_flyteidl.core.LiteralType`", "", ""







.. _ref_flyteidl.core.Void:

Void
------------------------------------------------------------------

Used to denote a nil/null/None assignment to a scalar value. The underlying LiteralType for Void is intentionally
undefined since it can be assigned to a scalar of any LiteralType.







..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/security.proto:

flyteidl/core/security.proto
==================================================================





.. _ref_flyteidl.core.Identity:

Identity
------------------------------------------------------------------

Identity encapsulates the various security identities a task can run as. It's up to the underlying plugin to pick the
right identity for the execution environment.



.. csv-table:: Identity type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "iam_role", ":ref:`ref_string`", "", "iam_role references the fully qualified name of Identity & Access Management role to impersonate."
   "k8s_service_account", ":ref:`ref_string`", "", "k8s_service_account references a kubernetes service account to impersonate."
   "oauth2_client", ":ref:`ref_flyteidl.core.OAuth2Client`", "", "oauth2_client references an oauth2 client. Backend plugins can use this information to impersonate the client when making external calls."







.. _ref_flyteidl.core.OAuth2Client:

OAuth2Client
------------------------------------------------------------------

OAuth2Client encapsulates OAuth2 Client Credentials to be used when making calls on behalf of that task.



.. csv-table:: OAuth2Client type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "client_id", ":ref:`ref_string`", "", "client_id is the public id for the client to use. The system will not perform any pre-auth validation that the secret requested matches the client_id indicated here. +required"
   "client_secret", ":ref:`ref_flyteidl.core.Secret`", "", "client_secret is a reference to the secret used to authenticate the OAuth2 client. +required"







.. _ref_flyteidl.core.OAuth2TokenRequest:

OAuth2TokenRequest
------------------------------------------------------------------

OAuth2TokenRequest encapsulates information needed to request an OAuth2 token.
FLYTE_TOKENS_ENV_PREFIX will be passed to indicate the prefix of the environment variables that will be present if
tokens are passed through environment variables.
FLYTE_TOKENS_PATH_PREFIX will be passed to indicate the prefix of the path where secrets will be mounted if tokens
are passed through file mounts.



.. csv-table:: OAuth2TokenRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "name", ":ref:`ref_string`", "", "name indicates a unique id for the token request within this task token requests. It'll be used as a suffix for environment variables and as a filename for mounting tokens as files. +required"
   "type", ":ref:`ref_flyteidl.core.OAuth2TokenRequest.Type`", "", "type indicates the type of the request to make. Defaults to CLIENT_CREDENTIALS. +required"
   "client", ":ref:`ref_flyteidl.core.OAuth2Client`", "", "client references the client_id/secret to use to request the OAuth2 token. +required"
   "idp_discovery_endpoint", ":ref:`ref_string`", "", "idp_discovery_endpoint references the discovery endpoint used to retrieve token endpoint and other related information. +optional"
   "token_endpoint", ":ref:`ref_string`", "", "token_endpoint references the token issuance endpoint. If idp_discovery_endpoint is not provided, this parameter is mandatory. +optional"







.. _ref_flyteidl.core.Secret:

Secret
------------------------------------------------------------------

Secret encapsulates information about the secret a task needs to proceed. An environment variable
FLYTE_SECRETS_ENV_PREFIX will be passed to indicate the prefix of the environment variables that will be present if
secrets are passed through environment variables.
FLYTE_SECRETS_DEFAULT_DIR will be passed to indicate the prefix of the path where secrets will be mounted if secrets
are passed through file mounts.



.. csv-table:: Secret type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "group", ":ref:`ref_string`", "", "The name of the secret group where to find the key referenced below. For K8s secrets, this should be the name of the v1/secret object. For Confidant, this should be the Credential name. For Vault, this should be the secret name. For AWS Secret Manager, this should be the name of the secret. +required"
   "group_version", ":ref:`ref_string`", "", "The group version to fetch. This is not supported in all secret management systems. It'll be ignored for the ones that do not support it. +optional"
   "key", ":ref:`ref_string`", "", "The name of the secret to mount. This has to match an existing secret in the system. It's up to the implementation of the secret management system to require case sensitivity. For K8s secrets, Confidant and Vault, this should match one of the keys inside the secret. For AWS Secret Manager, it's ignored. +optional"
   "mount_requirement", ":ref:`ref_flyteidl.core.Secret.MountType`", "", "mount_requirement is optional. Indicates where the secret has to be mounted. If provided, the execution will fail if the underlying key management system cannot satisfy that requirement. If not provided, the default location will depend on the key management system. +optional"







.. _ref_flyteidl.core.SecurityContext:

SecurityContext
------------------------------------------------------------------

SecurityContext holds security attributes that apply to tasks.



.. csv-table:: SecurityContext type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "run_as", ":ref:`ref_flyteidl.core.Identity`", "", "run_as encapsulates the identity a pod should run as. If the task fills in multiple fields here, it'll be up to the backend plugin to choose the appropriate identity for the execution engine the task will run on."
   "secrets", ":ref:`ref_flyteidl.core.Secret`", "repeated", "secrets indicate the list of secrets the task needs in order to proceed. Secrets will be mounted/passed to the pod as it starts. If the plugin responsible for kicking of the task will not run it on a flyte cluster (e.g. AWS Batch), it's the responsibility of the plugin to fetch the secret (which means propeller identity will need access to the secret) and to pass it to the remote execution engine."
   "tokens", ":ref:`ref_flyteidl.core.OAuth2TokenRequest`", "repeated", "tokens indicate the list of token requests the task needs in order to proceed. Tokens will be mounted/passed to the pod as it starts. If the plugin responsible for kicking of the task will not run it on a flyte cluster (e.g. AWS Batch), it's the responsibility of the plugin to fetch the secret (which means propeller identity will need access to the secret) and to pass it to the remote execution engine."






..
   end messages



.. _ref_flyteidl.core.OAuth2TokenRequest.Type:

OAuth2TokenRequest.Type
------------------------------------------------------------------

Type of the token requested.

.. csv-table:: Enum OAuth2TokenRequest.Type values
   :header: "Name", "Number", "Description"
   :widths: auto

   "CLIENT_CREDENTIALS", "0", "CLIENT_CREDENTIALS indicates a 2-legged OAuth token requested using client credentials."



.. _ref_flyteidl.core.Secret.MountType:

Secret.MountType
------------------------------------------------------------------



.. csv-table:: Enum Secret.MountType values
   :header: "Name", "Number", "Description"
   :widths: auto

   "ANY", "0", "Default case, indicates the client can tolerate either mounting options."
   "ENV_VAR", "1", "ENV_VAR indicates the secret needs to be mounted as an environment variable."
   "FILE", "2", "FILE indicates the secret needs to be mounted as a file."


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/tasks.proto:

flyteidl/core/tasks.proto
==================================================================





.. _ref_flyteidl.core.Container:

Container
------------------------------------------------------------------





.. csv-table:: Container type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "image", ":ref:`ref_string`", "", "Container image url. Eg: docker/redis:latest"
   "command", ":ref:`ref_string`", "repeated", "Command to be executed, if not provided, the default entrypoint in the container image will be used."
   "args", ":ref:`ref_string`", "repeated", "These will default to Flyte given paths. If provided, the system will not append known paths. If the task still needs flyte's inputs and outputs path, add $(FLYTE_INPUT_FILE), $(FLYTE_OUTPUT_FILE) wherever makes sense and the system will populate these before executing the container."
   "resources", ":ref:`ref_flyteidl.core.Resources`", "", "Container resources requirement as specified by the container engine."
   "env", ":ref:`ref_flyteidl.core.KeyValuePair`", "repeated", "Environment variables will be set as the container is starting up."
   "config", ":ref:`ref_flyteidl.core.KeyValuePair`", "repeated", "**Deprecated.** Allows extra configs to be available for the container. TODO: elaborate on how configs will become available. Deprecated, please use TaskTemplate.config instead."
   "ports", ":ref:`ref_flyteidl.core.ContainerPort`", "repeated", "Ports to open in the container. This feature is not supported by all execution engines. (e.g. supported on K8s but not supported on AWS Batch) Only K8s"
   "data_config", ":ref:`ref_flyteidl.core.DataLoadingConfig`", "", "BETA: Optional configuration for DataLoading. If not specified, then default values are used. This makes it possible to to run a completely portable container, that uses inputs and outputs only from the local file-system and without having any reference to flyteidl. This is supported only on K8s at the moment. If data loading is enabled, then data will be mounted in accompanying directories specified in the DataLoadingConfig. If the directories are not specified, inputs will be mounted onto and outputs will be uploaded from a pre-determined file-system path. Refer to the documentation to understand the default paths. Only K8s"
   "architecture", ":ref:`ref_flyteidl.core.Container.Architecture`", "", ""







.. _ref_flyteidl.core.ContainerPort:

ContainerPort
------------------------------------------------------------------

Defines port properties for a container.



.. csv-table:: ContainerPort type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "container_port", ":ref:`ref_uint32`", "", "Number of port to expose on the pod's IP address. This must be a valid port number, 0 < x < 65536."







.. _ref_flyteidl.core.DataLoadingConfig:

DataLoadingConfig
------------------------------------------------------------------

This configuration allows executing raw containers in Flyte using the Flyte CoPilot system.
Flyte CoPilot, eliminates the needs of flytekit or sdk inside the container. Any inputs required by the users container are side-loaded in the input_path
Any outputs generated by the user container - within output_path are automatically uploaded.



.. csv-table:: DataLoadingConfig type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "enabled", ":ref:`ref_bool`", "", "Flag enables DataLoading Config. If this is not set, data loading will not be used!"
   "input_path", ":ref:`ref_string`", "", "File system path (start at root). This folder will contain all the inputs exploded to a separate file. Example, if the input interface needs (x: int, y: blob, z: multipart_blob) and the input path is '/var/flyte/inputs', then the file system will look like /var/flyte/inputs/inputs.<metadata format dependent -> .pb .json .yaml> -> Format as defined previously. The Blob and Multipart blob will reference local filesystem instead of remote locations /var/flyte/inputs/x -> X is a file that contains the value of x (integer) in string format /var/flyte/inputs/y -> Y is a file in Binary format /var/flyte/inputs/z/... -> Note Z itself is a directory More information about the protocol - refer to docs #TODO reference docs here"
   "output_path", ":ref:`ref_string`", "", "File system path (start at root). This folder should contain all the outputs for the task as individual files and/or an error text file"
   "format", ":ref:`ref_flyteidl.core.DataLoadingConfig.LiteralMapFormat`", "", "In the inputs folder, there will be an additional summary/metadata file that contains references to all files or inlined primitive values. This format decides the actual encoding for the data. Refer to the encoding to understand the specifics of the contents and the encoding"
   "io_strategy", ":ref:`ref_flyteidl.core.IOStrategy`", "", ""







.. _ref_flyteidl.core.IOStrategy:

IOStrategy
------------------------------------------------------------------

Strategy to use when dealing with Blob, Schema, or multipart blob data (large datasets)



.. csv-table:: IOStrategy type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "download_mode", ":ref:`ref_flyteidl.core.IOStrategy.DownloadMode`", "", "Mode to use to manage downloads"
   "upload_mode", ":ref:`ref_flyteidl.core.IOStrategy.UploadMode`", "", "Mode to use to manage uploads"







.. _ref_flyteidl.core.K8sObjectMetadata:

K8sObjectMetadata
------------------------------------------------------------------

Metadata for building a kubernetes object when a task is executed.



.. csv-table:: K8sObjectMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "labels", ":ref:`ref_flyteidl.core.K8sObjectMetadata.LabelsEntry`", "repeated", "Optional labels to add to the pod definition."
   "annotations", ":ref:`ref_flyteidl.core.K8sObjectMetadata.AnnotationsEntry`", "repeated", "Optional annotations to add to the pod definition."







.. _ref_flyteidl.core.K8sObjectMetadata.AnnotationsEntry:

K8sObjectMetadata.AnnotationsEntry
------------------------------------------------------------------





.. csv-table:: K8sObjectMetadata.AnnotationsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.core.K8sObjectMetadata.LabelsEntry:

K8sObjectMetadata.LabelsEntry
------------------------------------------------------------------





.. csv-table:: K8sObjectMetadata.LabelsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.core.K8sPod:

K8sPod
------------------------------------------------------------------

Defines a pod spec and additional pod metadata that is created when a task is executed.



.. csv-table:: K8sPod type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "metadata", ":ref:`ref_flyteidl.core.K8sObjectMetadata`", "", "Contains additional metadata for building a kubernetes pod."
   "pod_spec", ":ref:`ref_google.protobuf.Struct`", "", "Defines the primary pod spec created when a task is executed. This should be a JSON-marshalled pod spec, which can be defined in - go, using: https://github.com/kubernetes/api/blob/release-1.21/core/v1/types.go#L2936 - python: using https://github.com/kubernetes-client/python/blob/release-19.0/kubernetes/client/models/v1_pod_spec.py"







.. _ref_flyteidl.core.Resources:

Resources
------------------------------------------------------------------

A customizable interface to convey resources requested for a container. This can be interpreted differently for different
container engines.



.. csv-table:: Resources type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "requests", ":ref:`ref_flyteidl.core.Resources.ResourceEntry`", "repeated", "The desired set of resources requested. ResourceNames must be unique within the list."
   "limits", ":ref:`ref_flyteidl.core.Resources.ResourceEntry`", "repeated", "Defines a set of bounds (e.g. min/max) within which the task can reliably run. ResourceNames must be unique within the list."







.. _ref_flyteidl.core.Resources.ResourceEntry:

Resources.ResourceEntry
------------------------------------------------------------------

Encapsulates a resource name and value.



.. csv-table:: Resources.ResourceEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "name", ":ref:`ref_flyteidl.core.Resources.ResourceName`", "", "Resource name."
   "value", ":ref:`ref_string`", "", "Value must be a valid k8s quantity. See https://github.com/kubernetes/apimachinery/blob/master/pkg/api/resource/quantity.go#L30-L80"







.. _ref_flyteidl.core.RuntimeMetadata:

RuntimeMetadata
------------------------------------------------------------------

Runtime information. This is loosely defined to allow for extensibility.



.. csv-table:: RuntimeMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "type", ":ref:`ref_flyteidl.core.RuntimeMetadata.RuntimeType`", "", "Type of runtime."
   "version", ":ref:`ref_string`", "", "Version of the runtime. All versions should be backward compatible. However, certain cases call for version checks to ensure tighter validation or setting expectations."
   "flavor", ":ref:`ref_string`", "", "+optional It can be used to provide extra information about the runtime (e.g. python, golang... etc.)."







.. _ref_flyteidl.core.Sql:

Sql
------------------------------------------------------------------

Sql represents a generic sql workload with a statement and dialect.



.. csv-table:: Sql type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "statement", ":ref:`ref_string`", "", "The actual query to run, the query can have templated parameters. We use Flyte's Golang templating format for Query templating. Refer to the templating documentation. https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/external_services/hive/hive.html#sphx-glr-auto-integrations-external-services-hive-hive-py For example, insert overwrite directory '{{ .rawOutputDataPrefix }}' stored as parquet select * from my_table where ds = '{{ .Inputs.ds }}'"
   "dialect", ":ref:`ref_flyteidl.core.Sql.Dialect`", "", ""







.. _ref_flyteidl.core.TaskMetadata:

TaskMetadata
------------------------------------------------------------------

Task Metadata



.. csv-table:: TaskMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "discoverable", ":ref:`ref_bool`", "", "Indicates whether the system should attempt to lookup this task's output to avoid duplication of work."
   "runtime", ":ref:`ref_flyteidl.core.RuntimeMetadata`", "", "Runtime information about the task."
   "timeout", ":ref:`ref_google.protobuf.Duration`", "", "The overall timeout of a task including user-triggered retries."
   "retries", ":ref:`ref_flyteidl.core.RetryStrategy`", "", "Number of retries per task."
   "discovery_version", ":ref:`ref_string`", "", "Indicates a logical version to apply to this task for the purpose of discovery."
   "deprecated_error_message", ":ref:`ref_string`", "", "If set, this indicates that this task is deprecated. This will enable owners of tasks to notify consumers of the ending of support for a given task."
   "interruptible", ":ref:`ref_bool`", "", ""
   "cache_serializable", ":ref:`ref_bool`", "", "Indicates whether the system should attempt to execute discoverable instances in serial to avoid duplicate work"
   "generates_deck", ":ref:`ref_bool`", "", "Indicates whether the task will generate a Deck URI when it finishes executing."
   "tags", ":ref:`ref_flyteidl.core.TaskMetadata.TagsEntry`", "repeated", "Arbitrary tags that allow users and the platform to store small but arbitrary labels"







.. _ref_flyteidl.core.TaskMetadata.TagsEntry:

TaskMetadata.TagsEntry
------------------------------------------------------------------





.. csv-table:: TaskMetadata.TagsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.core.TaskTemplate:

TaskTemplate
------------------------------------------------------------------

A Task structure that uniquely identifies a task in the system
Tasks are registered as a first step in the system.



.. csv-table:: TaskTemplate type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "Auto generated taskId by the system. Task Id uniquely identifies this task globally."
   "type", ":ref:`ref_string`", "", "A predefined yet extensible Task type identifier. This can be used to customize any of the components. If no extensions are provided in the system, Flyte will resolve the this task to its TaskCategory and default the implementation registered for the TaskCategory."
   "metadata", ":ref:`ref_flyteidl.core.TaskMetadata`", "", "Extra metadata about the task."
   "interface", ":ref:`ref_flyteidl.core.TypedInterface`", "", "A strongly typed interface for the task. This enables others to use this task within a workflow and guarantees compile-time validation of the workflow to avoid costly runtime failures."
   "custom", ":ref:`ref_google.protobuf.Struct`", "", "Custom data about the task. This is extensible to allow various plugins in the system."
   "container", ":ref:`ref_flyteidl.core.Container`", "", ""
   "k8s_pod", ":ref:`ref_flyteidl.core.K8sPod`", "", ""
   "sql", ":ref:`ref_flyteidl.core.Sql`", "", ""
   "task_type_version", ":ref:`ref_int32`", "", "This can be used to customize task handling at execution time for the same task type."
   "security_context", ":ref:`ref_flyteidl.core.SecurityContext`", "", "security_context encapsulates security attributes requested to run this task."
   "config", ":ref:`ref_flyteidl.core.TaskTemplate.ConfigEntry`", "repeated", "Metadata about the custom defined for this task. This is extensible to allow various plugins in the system to use as required. reserve the field numbers 1 through 15 for very frequently occurring message elements"







.. _ref_flyteidl.core.TaskTemplate.ConfigEntry:

TaskTemplate.ConfigEntry
------------------------------------------------------------------





.. csv-table:: TaskTemplate.ConfigEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""






..
   end messages



.. _ref_flyteidl.core.Container.Architecture:

Container.Architecture
------------------------------------------------------------------

Architecture-type the container image supports.

.. csv-table:: Enum Container.Architecture values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNKNOWN", "0", ""
   "AMD64", "1", ""
   "ARM64", "2", ""
   "ARM_V6", "3", ""
   "ARM_V7", "4", ""



.. _ref_flyteidl.core.DataLoadingConfig.LiteralMapFormat:

DataLoadingConfig.LiteralMapFormat
------------------------------------------------------------------

LiteralMapFormat decides the encoding format in which the input metadata should be made available to the containers.
If the user has access to the protocol buffer definitions, it is recommended to use the PROTO format.
JSON and YAML do not need any protobuf definitions to read it
All remote references in core.LiteralMap are replaced with local filesystem references (the data is downloaded to local filesystem)

.. csv-table:: Enum DataLoadingConfig.LiteralMapFormat values
   :header: "Name", "Number", "Description"
   :widths: auto

   "JSON", "0", "JSON / YAML for the metadata (which contains inlined primitive values). The representation is inline with the standard json specification as specified - https://www.json.org/json-en.html"
   "YAML", "1", ""
   "PROTO", "2", "Proto is a serialized binary of `core.LiteralMap` defined in flyteidl/core"



.. _ref_flyteidl.core.IOStrategy.DownloadMode:

IOStrategy.DownloadMode
------------------------------------------------------------------

Mode to use for downloading

.. csv-table:: Enum IOStrategy.DownloadMode values
   :header: "Name", "Number", "Description"
   :widths: auto

   "DOWNLOAD_EAGER", "0", "All data will be downloaded before the main container is executed"
   "DOWNLOAD_STREAM", "1", "Data will be downloaded as a stream and an End-Of-Stream marker will be written to indicate all data has been downloaded. Refer to protocol for details"
   "DO_NOT_DOWNLOAD", "2", "Large objects (offloaded) will not be downloaded"



.. _ref_flyteidl.core.IOStrategy.UploadMode:

IOStrategy.UploadMode
------------------------------------------------------------------

Mode to use for uploading

.. csv-table:: Enum IOStrategy.UploadMode values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UPLOAD_ON_EXIT", "0", "All data will be uploaded after the main container exits"
   "UPLOAD_EAGER", "1", "Data will be uploaded as it appears. Refer to protocol specification for details"
   "DO_NOT_UPLOAD", "2", "Data will not be uploaded, only references will be written"



.. _ref_flyteidl.core.Resources.ResourceName:

Resources.ResourceName
------------------------------------------------------------------

Known resource names.

.. csv-table:: Enum Resources.ResourceName values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNKNOWN", "0", ""
   "CPU", "1", ""
   "GPU", "2", ""
   "MEMORY", "3", ""
   "STORAGE", "4", ""
   "EPHEMERAL_STORAGE", "5", "For Kubernetes-based deployments, pods use ephemeral local storage for scratch space, caching, and for logs."



.. _ref_flyteidl.core.RuntimeMetadata.RuntimeType:

RuntimeMetadata.RuntimeType
------------------------------------------------------------------



.. csv-table:: Enum RuntimeMetadata.RuntimeType values
   :header: "Name", "Number", "Description"
   :widths: auto

   "OTHER", "0", ""
   "FLYTE_SDK", "1", ""



.. _ref_flyteidl.core.Sql.Dialect:

Sql.Dialect
------------------------------------------------------------------

The dialect of the SQL statement. This is used to validate and parse SQL statements at compilation time to avoid
expensive runtime operations. If set to an unsupported dialect, no validation will be done on the statement.
We support the following dialect: ansi, hive.

.. csv-table:: Enum Sql.Dialect values
   :header: "Name", "Number", "Description"
   :widths: auto

   "UNDEFINED", "0", ""
   "ANSI", "1", ""
   "HIVE", "2", ""
   "OTHER", "3", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/types.proto:

flyteidl/core/types.proto
==================================================================





.. _ref_flyteidl.core.BlobType:

BlobType
------------------------------------------------------------------

Defines type behavior for blob objects



.. csv-table:: BlobType type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "format", ":ref:`ref_string`", "", "Format can be a free form string understood by SDK/UI etc like csv, parquet etc"
   "dimensionality", ":ref:`ref_flyteidl.core.BlobType.BlobDimensionality`", "", ""







.. _ref_flyteidl.core.EnumType:

EnumType
------------------------------------------------------------------

Enables declaring enum types, with predefined string values
For len(values) > 0, the first value in the ordered list is regarded as the default value. If you wish
To provide no defaults, make the first value as undefined.



.. csv-table:: EnumType type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "values", ":ref:`ref_string`", "repeated", "Predefined set of enum values."







.. _ref_flyteidl.core.Error:

Error
------------------------------------------------------------------

Represents an error thrown from a node.



.. csv-table:: Error type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "failed_node_id", ":ref:`ref_string`", "", "The node id that threw the error."
   "message", ":ref:`ref_string`", "", "Error message thrown."







.. _ref_flyteidl.core.LiteralType:

LiteralType
------------------------------------------------------------------

Defines a strong type to allow type checking between interfaces.



.. csv-table:: LiteralType type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "simple", ":ref:`ref_flyteidl.core.SimpleType`", "", "A simple type that can be compared one-to-one with another."
   "schema", ":ref:`ref_flyteidl.core.SchemaType`", "", "A complex type that requires matching of inner fields."
   "collection_type", ":ref:`ref_flyteidl.core.LiteralType`", "", "Defines the type of the value of a collection. Only homogeneous collections are allowed."
   "map_value_type", ":ref:`ref_flyteidl.core.LiteralType`", "", "Defines the type of the value of a map type. The type of the key is always a string."
   "blob", ":ref:`ref_flyteidl.core.BlobType`", "", "A blob might have specialized implementation details depending on associated metadata."
   "enum_type", ":ref:`ref_flyteidl.core.EnumType`", "", "Defines an enum with pre-defined string values."
   "structured_dataset_type", ":ref:`ref_flyteidl.core.StructuredDatasetType`", "", "Generalized schema support"
   "union_type", ":ref:`ref_flyteidl.core.UnionType`", "", "Defines an union type with pre-defined LiteralTypes."
   "metadata", ":ref:`ref_google.protobuf.Struct`", "", "This field contains type metadata that is descriptive of the type, but is NOT considered in type-checking. This might be used by consumers to identify special behavior or display extended information for the type."
   "annotation", ":ref:`ref_flyteidl.core.TypeAnnotation`", "", "This field contains arbitrary data that might have special semantic meaning for the client but does not effect internal flyte behavior."
   "structure", ":ref:`ref_flyteidl.core.TypeStructure`", "", "Hints to improve type matching."







.. _ref_flyteidl.core.OutputReference:

OutputReference
------------------------------------------------------------------

A reference to an output produced by a node. The type can be retrieved -and validated- from
the underlying interface of the node.



.. csv-table:: OutputReference type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "node_id", ":ref:`ref_string`", "", "Node id must exist at the graph layer."
   "var", ":ref:`ref_string`", "", "Variable name must refer to an output variable for the node."







.. _ref_flyteidl.core.SchemaType:

SchemaType
------------------------------------------------------------------

Defines schema columns and types to strongly type-validate schemas interoperability.



.. csv-table:: SchemaType type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "columns", ":ref:`ref_flyteidl.core.SchemaType.SchemaColumn`", "repeated", "A list of ordered columns this schema comprises of."







.. _ref_flyteidl.core.SchemaType.SchemaColumn:

SchemaType.SchemaColumn
------------------------------------------------------------------





.. csv-table:: SchemaType.SchemaColumn type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "name", ":ref:`ref_string`", "", "A unique name -within the schema type- for the column"
   "type", ":ref:`ref_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType`", "", "The column type. This allows a limited set of types currently."







.. _ref_flyteidl.core.StructuredDatasetType:

StructuredDatasetType
------------------------------------------------------------------





.. csv-table:: StructuredDatasetType type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "columns", ":ref:`ref_flyteidl.core.StructuredDatasetType.DatasetColumn`", "repeated", "A list of ordered columns this schema comprises of."
   "format", ":ref:`ref_string`", "", "This is the storage format, the format of the bits at rest parquet, feather, csv, etc. For two types to be compatible, the format will need to be an exact match."
   "external_schema_type", ":ref:`ref_string`", "", "This is a string representing the type that the bytes in external_schema_bytes are formatted in. This is an optional field that will not be used for type checking."
   "external_schema_bytes", ":ref:`ref_bytes`", "", "The serialized bytes of a third-party schema library like Arrow. This is an optional field that will not be used for type checking."







.. _ref_flyteidl.core.StructuredDatasetType.DatasetColumn:

StructuredDatasetType.DatasetColumn
------------------------------------------------------------------





.. csv-table:: StructuredDatasetType.DatasetColumn type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "name", ":ref:`ref_string`", "", "A unique name within the schema type for the column."
   "literal_type", ":ref:`ref_flyteidl.core.LiteralType`", "", "The column type."







.. _ref_flyteidl.core.TypeAnnotation:

TypeAnnotation
------------------------------------------------------------------

TypeAnnotation encapsulates registration time information about a type. This can be used for various control-plane operations. TypeAnnotation will not be available at runtime when a task runs.



.. csv-table:: TypeAnnotation type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "annotations", ":ref:`ref_google.protobuf.Struct`", "", "A arbitrary JSON payload to describe a type."







.. _ref_flyteidl.core.TypeStructure:

TypeStructure
------------------------------------------------------------------

Hints to improve type matching
e.g. allows distinguishing output from custom type transformers
even if the underlying IDL serialization matches.



.. csv-table:: TypeStructure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "tag", ":ref:`ref_string`", "", "Must exactly match for types to be castable"







.. _ref_flyteidl.core.UnionType:

UnionType
------------------------------------------------------------------

Defines a tagged union type, also known as a variant (and formally as the sum type).

A sum type S is defined by a sequence of types (A, B, C, ...), each tagged by a string tag
A value of type S is constructed from a value of any of the variant types. The specific choice of type is recorded by
storing the varaint's tag with the literal value and can be examined in runtime.

Type S is typically written as
S := Apple A | Banana B | Cantaloupe C | ...

Notably, a nullable (optional) type is a sum type between some type X and the singleton type representing a null-value:
Optional X := X | Null

See also: https://en.wikipedia.org/wiki/Tagged_union



.. csv-table:: UnionType type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "variants", ":ref:`ref_flyteidl.core.LiteralType`", "repeated", "Predefined set of variants in union."






..
   end messages



.. _ref_flyteidl.core.BlobType.BlobDimensionality:

BlobType.BlobDimensionality
------------------------------------------------------------------



.. csv-table:: Enum BlobType.BlobDimensionality values
   :header: "Name", "Number", "Description"
   :widths: auto

   "SINGLE", "0", ""
   "MULTIPART", "1", ""



.. _ref_flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType:

SchemaType.SchemaColumn.SchemaColumnType
------------------------------------------------------------------



.. csv-table:: Enum SchemaType.SchemaColumn.SchemaColumnType values
   :header: "Name", "Number", "Description"
   :widths: auto

   "INTEGER", "0", ""
   "FLOAT", "1", ""
   "STRING", "2", ""
   "BOOLEAN", "3", ""
   "DATETIME", "4", ""
   "DURATION", "5", ""



.. _ref_flyteidl.core.SimpleType:

SimpleType
------------------------------------------------------------------

Define a set of simple types.

.. csv-table:: Enum SimpleType values
   :header: "Name", "Number", "Description"
   :widths: auto

   "NONE", "0", ""
   "INTEGER", "1", ""
   "FLOAT", "2", ""
   "STRING", "3", ""
   "BOOLEAN", "4", ""
   "DATETIME", "5", ""
   "DURATION", "6", ""
   "BINARY", "7", ""
   "ERROR", "8", ""
   "STRUCT", "9", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/workflow.proto:

flyteidl/core/workflow.proto
==================================================================





.. _ref_flyteidl.core.Alias:

Alias
------------------------------------------------------------------

Links a variable to an alias.



.. csv-table:: Alias type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "var", ":ref:`ref_string`", "", "Must match one of the output variable names on a node."
   "alias", ":ref:`ref_string`", "", "A workflow-level unique alias that downstream nodes can refer to in their input."







.. _ref_flyteidl.core.ApproveCondition:

ApproveCondition
------------------------------------------------------------------

ApproveCondition represents a dependency on an external approval. During execution, this will manifest as a boolean
signal with the provided signal_id.



.. csv-table:: ApproveCondition type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signal_id", ":ref:`ref_string`", "", "A unique identifier for the requested boolean signal."







.. _ref_flyteidl.core.BranchNode:

BranchNode
------------------------------------------------------------------

BranchNode is a special node that alter the flow of the workflow graph. It allows the control flow to branch at
runtime based on a series of conditions that get evaluated on various parameters (e.g. inputs, primitives).



.. csv-table:: BranchNode type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "if_else", ":ref:`ref_flyteidl.core.IfElseBlock`", "", "+required"







.. _ref_flyteidl.core.GateNode:

GateNode
------------------------------------------------------------------

GateNode refers to the condition that is required for the gate to successfully complete.



.. csv-table:: GateNode type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "approve", ":ref:`ref_flyteidl.core.ApproveCondition`", "", "ApproveCondition represents a dependency on an external approval provided by a boolean signal."
   "signal", ":ref:`ref_flyteidl.core.SignalCondition`", "", "SignalCondition represents a dependency on an signal."
   "sleep", ":ref:`ref_flyteidl.core.SleepCondition`", "", "SleepCondition represents a dependency on waiting for the specified duration."







.. _ref_flyteidl.core.IfBlock:

IfBlock
------------------------------------------------------------------

Defines a condition and the execution unit that should be executed if the condition is satisfied.



.. csv-table:: IfBlock type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "condition", ":ref:`ref_flyteidl.core.BooleanExpression`", "", ""
   "then_node", ":ref:`ref_flyteidl.core.Node`", "", ""







.. _ref_flyteidl.core.IfElseBlock:

IfElseBlock
------------------------------------------------------------------

Defines a series of if/else blocks. The first branch whose condition evaluates to true is the one to execute.
If no conditions were satisfied, the else_node or the error will execute.



.. csv-table:: IfElseBlock type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "case", ":ref:`ref_flyteidl.core.IfBlock`", "", "+required. First condition to evaluate."
   "other", ":ref:`ref_flyteidl.core.IfBlock`", "repeated", "+optional. Additional branches to evaluate."
   "else_node", ":ref:`ref_flyteidl.core.Node`", "", "The node to execute in case none of the branches were taken."
   "error", ":ref:`ref_flyteidl.core.Error`", "", "An error to throw in case none of the branches were taken."







.. _ref_flyteidl.core.Node:

Node
------------------------------------------------------------------

A Workflow graph Node. One unit of execution in the graph. Each node can be linked to a Task, a Workflow or a branch
node.



.. csv-table:: Node type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_string`", "", "A workflow-level unique identifier that identifies this node in the workflow. 'inputs' and 'outputs' are reserved node ids that cannot be used by other nodes."
   "metadata", ":ref:`ref_flyteidl.core.NodeMetadata`", "", "Extra metadata about the node."
   "inputs", ":ref:`ref_flyteidl.core.Binding`", "repeated", "Specifies how to bind the underlying interface's inputs. All required inputs specified in the underlying interface must be fulfilled."
   "upstream_node_ids", ":ref:`ref_string`", "repeated", "+optional Specifies execution dependency for this node ensuring it will only get scheduled to run after all its upstream nodes have completed. This node will have an implicit dependency on any node that appears in inputs field."
   "output_aliases", ":ref:`ref_flyteidl.core.Alias`", "repeated", "+optional. A node can define aliases for a subset of its outputs. This is particularly useful if different nodes need to conform to the same interface (e.g. all branches in a branch node). Downstream nodes must refer to this nodes outputs using the alias if one's specified."
   "task_node", ":ref:`ref_flyteidl.core.TaskNode`", "", "Information about the Task to execute in this node."
   "workflow_node", ":ref:`ref_flyteidl.core.WorkflowNode`", "", "Information about the Workflow to execute in this mode."
   "branch_node", ":ref:`ref_flyteidl.core.BranchNode`", "", "Information about the branch node to evaluate in this node."
   "gate_node", ":ref:`ref_flyteidl.core.GateNode`", "", "Information about the condition to evaluate in this node."







.. _ref_flyteidl.core.NodeMetadata:

NodeMetadata
------------------------------------------------------------------

Defines extra information about the Node.



.. csv-table:: NodeMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "name", ":ref:`ref_string`", "", "A friendly name for the Node"
   "timeout", ":ref:`ref_google.protobuf.Duration`", "", "The overall timeout of a task."
   "retries", ":ref:`ref_flyteidl.core.RetryStrategy`", "", "Number of retries per task."
   "interruptible", ":ref:`ref_bool`", "", ""







.. _ref_flyteidl.core.SignalCondition:

SignalCondition
------------------------------------------------------------------

SignalCondition represents a dependency on an signal.



.. csv-table:: SignalCondition type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signal_id", ":ref:`ref_string`", "", "A unique identifier for the requested signal."
   "type", ":ref:`ref_flyteidl.core.LiteralType`", "", "A type denoting the required value type for this signal."
   "output_variable_name", ":ref:`ref_string`", "", "The variable name for the signal value in this nodes outputs."







.. _ref_flyteidl.core.SleepCondition:

SleepCondition
------------------------------------------------------------------

SleepCondition represents a dependency on waiting for the specified duration.



.. csv-table:: SleepCondition type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "duration", ":ref:`ref_google.protobuf.Duration`", "", "The overall duration for this sleep."







.. _ref_flyteidl.core.TaskNode:

TaskNode
------------------------------------------------------------------

Refers to the task that the Node is to execute.



.. csv-table:: TaskNode type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "reference_id", ":ref:`ref_flyteidl.core.Identifier`", "", "A globally unique identifier for the task."
   "overrides", ":ref:`ref_flyteidl.core.TaskNodeOverrides`", "", "Optional overrides applied at task execution time."







.. _ref_flyteidl.core.TaskNodeOverrides:

TaskNodeOverrides
------------------------------------------------------------------

Optional task node overrides that will be applied at task execution time.



.. csv-table:: TaskNodeOverrides type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "resources", ":ref:`ref_flyteidl.core.Resources`", "", "A customizable interface to convey resources requested for a task container."







.. _ref_flyteidl.core.WorkflowMetadata:

WorkflowMetadata
------------------------------------------------------------------

This is workflow layer metadata. These settings are only applicable to the workflow as a whole, and do not
percolate down to child entities (like tasks) launched by the workflow.



.. csv-table:: WorkflowMetadata type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "quality_of_service", ":ref:`ref_flyteidl.core.QualityOfService`", "", "Indicates the runtime priority of workflow executions."
   "on_failure", ":ref:`ref_flyteidl.core.WorkflowMetadata.OnFailurePolicy`", "", "Defines how the system should behave when a failure is detected in the workflow execution."
   "tags", ":ref:`ref_flyteidl.core.WorkflowMetadata.TagsEntry`", "repeated", "Arbitrary tags that allow users and the platform to store small but arbitrary labels"







.. _ref_flyteidl.core.WorkflowMetadata.TagsEntry:

WorkflowMetadata.TagsEntry
------------------------------------------------------------------





.. csv-table:: WorkflowMetadata.TagsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.core.WorkflowMetadataDefaults:

WorkflowMetadataDefaults
------------------------------------------------------------------

The difference between these settings and the WorkflowMetadata ones is that these are meant to be passed down to
a workflow's underlying entities (like tasks). For instance, 'interruptible' has no meaning at the workflow layer, it
is only relevant when a task executes. The settings here are the defaults that are passed to all nodes
unless explicitly overridden at the node layer.
If you are adding a setting that applies to both the Workflow itself, and everything underneath it, it should be
added to both this object and the WorkflowMetadata object above.



.. csv-table:: WorkflowMetadataDefaults type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "interruptible", ":ref:`ref_bool`", "", "Whether child nodes of the workflow are interruptible."







.. _ref_flyteidl.core.WorkflowNode:

WorkflowNode
------------------------------------------------------------------

Refers to a the workflow the node is to execute.



.. csv-table:: WorkflowNode type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "launchplan_ref", ":ref:`ref_flyteidl.core.Identifier`", "", "A globally unique identifier for the launch plan."
   "sub_workflow_ref", ":ref:`ref_flyteidl.core.Identifier`", "", "Reference to a subworkflow, that should be defined with the compiler context"







.. _ref_flyteidl.core.WorkflowTemplate:

WorkflowTemplate
------------------------------------------------------------------

Flyte Workflow Structure that encapsulates task, branch and subworkflow nodes to form a statically analyzable,
directed acyclic graph.



.. csv-table:: WorkflowTemplate type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "id", ":ref:`ref_flyteidl.core.Identifier`", "", "A globally unique identifier for the workflow."
   "metadata", ":ref:`ref_flyteidl.core.WorkflowMetadata`", "", "Extra metadata about the workflow."
   "interface", ":ref:`ref_flyteidl.core.TypedInterface`", "", "Defines a strongly typed interface for the Workflow. This can include some optional parameters."
   "nodes", ":ref:`ref_flyteidl.core.Node`", "repeated", "A list of nodes. In addition, 'globals' is a special reserved node id that can be used to consume workflow inputs."
   "outputs", ":ref:`ref_flyteidl.core.Binding`", "repeated", "A list of output bindings that specify how to construct workflow outputs. Bindings can pull node outputs or specify literals. All workflow outputs specified in the interface field must be bound in order for the workflow to be validated. A workflow has an implicit dependency on all of its nodes to execute successfully in order to bind final outputs. Most of these outputs will be Binding's with a BindingData of type OutputReference. That is, your workflow can just have an output of some constant (`Output(5)`), but usually, the workflow will be pulling outputs from the output of a task."
   "failure_node", ":ref:`ref_flyteidl.core.Node`", "", "+optional A catch-all node. This node is executed whenever the execution engine determines the workflow has failed. The interface of this node must match the Workflow interface with an additional input named 'error' of type pb.lyft.flyte.core.Error."
   "metadata_defaults", ":ref:`ref_flyteidl.core.WorkflowMetadataDefaults`", "", "workflow defaults"






..
   end messages



.. _ref_flyteidl.core.WorkflowMetadata.OnFailurePolicy:

WorkflowMetadata.OnFailurePolicy
------------------------------------------------------------------

Failure Handling Strategy

.. csv-table:: Enum WorkflowMetadata.OnFailurePolicy values
   :header: "Name", "Number", "Description"
   :widths: auto

   "FAIL_IMMEDIATELY", "0", "FAIL_IMMEDIATELY instructs the system to fail as soon as a node fails in the workflow. It'll automatically abort all currently running nodes and clean up resources before finally marking the workflow executions as failed."
   "FAIL_AFTER_EXECUTABLE_NODES_COMPLETE", "1", "FAIL_AFTER_EXECUTABLE_NODES_COMPLETE instructs the system to make as much progress as it can. The system will not alter the dependencies of the execution graph so any node that depend on the failed node will not be run. Other nodes that will be executed to completion before cleaning up resources and marking the workflow execution as failed."


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/core/workflow_closure.proto:

flyteidl/core/workflow_closure.proto
==================================================================





.. _ref_flyteidl.core.WorkflowClosure:

WorkflowClosure
------------------------------------------------------------------

Defines an enclosed package of workflow and tasks it references.



.. csv-table:: WorkflowClosure type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "workflow", ":ref:`ref_flyteidl.core.WorkflowTemplate`", "", "required. Workflow template."
   "tasks", ":ref:`ref_flyteidl.core.TaskTemplate`", "repeated", "optional. A collection of tasks referenced by the workflow. Only needed if the workflow references tasks."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_google/protobuf/timestamp.proto:

google/protobuf/timestamp.proto
==================================================================





.. _ref_google.protobuf.Timestamp:

Timestamp
------------------------------------------------------------------

A Timestamp represents a point in time independent of any time zone or local
calendar, encoded as a count of seconds and fractions of seconds at
nanosecond resolution. The count is relative to an epoch at UTC midnight on
January 1, 1970, in the proleptic Gregorian calendar which extends the
Gregorian calendar backwards to year one.

All minutes are 60 seconds long. Leap seconds are "smeared" so that no leap
second table is needed for interpretation, using a [24-hour linear
smear](https://developers.google.com/time/smear).

The range is from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59.999999999Z. By
restricting to that range, we ensure that we can convert to and from [RFC
3339](https://www.ietf.org/rfc/rfc3339.txt) date strings.

# Examples

Example 1: Compute Timestamp from POSIX `time()`.

    Timestamp timestamp;
    timestamp.set_seconds(time(NULL));
    timestamp.set_nanos(0);

Example 2: Compute Timestamp from POSIX `gettimeofday()`.

    struct timeval tv;
    gettimeofday(&tv, NULL);

    Timestamp timestamp;
    timestamp.set_seconds(tv.tv_sec);
    timestamp.set_nanos(tv.tv_usec * 1000);

Example 3: Compute Timestamp from Win32 `GetSystemTimeAsFileTime()`.

    FILETIME ft;
    GetSystemTimeAsFileTime(&ft);
    UINT64 ticks = (((UINT64)ft.dwHighDateTime) << 32) | ft.dwLowDateTime;

    // A Windows tick is 100 nanoseconds. Windows epoch 1601-01-01T00:00:00Z
    // is 11644473600 seconds before Unix epoch 1970-01-01T00:00:00Z.
    Timestamp timestamp;
    timestamp.set_seconds((INT64) ((ticks / 10000000) - 11644473600LL));
    timestamp.set_nanos((INT32) ((ticks % 10000000) * 100));

Example 4: Compute Timestamp from Java `System.currentTimeMillis()`.

    long millis = System.currentTimeMillis();

    Timestamp timestamp = Timestamp.newBuilder().setSeconds(millis / 1000)
        .setNanos((int) ((millis % 1000) * 1000000)).build();

Example 5: Compute Timestamp from Java `Instant.now()`.

    Instant now = Instant.now();

    Timestamp timestamp =
        Timestamp.newBuilder().setSeconds(now.getEpochSecond())
            .setNanos(now.getNano()).build();

Example 6: Compute Timestamp from current time in Python.

    timestamp = Timestamp()
    timestamp.GetCurrentTime()

# JSON Mapping

In JSON format, the Timestamp type is encoded as a string in the
[RFC 3339](https://www.ietf.org/rfc/rfc3339.txt) format. That is, the
format is "{year}-{month}-{day}T{hour}:{min}:{sec}[.{frac_sec}]Z"
where {year} is always expressed using four digits while {month}, {day},
{hour}, {min}, and {sec} are zero-padded to two digits each. The fractional
seconds, which can go up to 9 digits (i.e. up to 1 nanosecond resolution),
are optional. The "Z" suffix indicates the timezone ("UTC"); the timezone
is required. A proto3 JSON serializer should always use UTC (as indicated by
"Z") when printing the Timestamp type and a proto3 JSON parser should be
able to accept both UTC and other timezones (as indicated by an offset).

For example, "2017-01-15T01:30:15.01Z" encodes 15.01 seconds past
01:30 UTC on January 15, 2017.

In JavaScript, one can convert a Date object to this format using the
standard
[toISOString()](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date/toISOString)
method. In Python, a standard `datetime.datetime` object can be converted
to this format using
[`strftime`](https://docs.python.org/2/library/time.html#time.strftime) with
the time format spec '%Y-%m-%dT%H:%M:%S.%fZ'. Likewise, in Java, one can use
the Joda Time's [`ISODateTimeFormat.dateTime()`](
http://www.joda.org/joda-time/apidocs/org/joda/time/format/ISODateTimeFormat.html#dateTime%2D%2D
) to obtain a formatter capable of generating timestamps in this format.



.. csv-table:: Timestamp type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "seconds", ":ref:`ref_int64`", "", "Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive."
   "nanos", ":ref:`ref_int32`", "", "Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive."






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




.. _ref_google/protobuf/struct.proto:

google/protobuf/struct.proto
==================================================================





.. _ref_google.protobuf.ListValue:

ListValue
------------------------------------------------------------------

`ListValue` is a wrapper around a repeated field of values.

The JSON representation for `ListValue` is JSON array.



.. csv-table:: ListValue type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "values", ":ref:`ref_google.protobuf.Value`", "repeated", "Repeated field of dynamically typed values."







.. _ref_google.protobuf.Struct:

Struct
------------------------------------------------------------------

`Struct` represents a structured data value, consisting of fields
which map to dynamically typed values. In some languages, `Struct`
might be supported by a native representation. For example, in
scripting languages like JS a struct is represented as an
object. The details of that representation are described together
with the proto support for the language.

The JSON representation for `Struct` is JSON object.



.. csv-table:: Struct type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "fields", ":ref:`ref_google.protobuf.Struct.FieldsEntry`", "repeated", "Unordered map of dynamically typed values."







.. _ref_google.protobuf.Struct.FieldsEntry:

Struct.FieldsEntry
------------------------------------------------------------------





.. csv-table:: Struct.FieldsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_google.protobuf.Value`", "", ""







.. _ref_google.protobuf.Value:

Value
------------------------------------------------------------------

`Value` represents a dynamically typed value which can be either
null, a number, a string, a boolean, a recursive struct value, or a
list of values. A producer of value is expected to set one of these
variants. Absence of any variant indicates an error.

The JSON representation for `Value` is JSON value.



.. csv-table:: Value type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "null_value", ":ref:`ref_google.protobuf.NullValue`", "", "Represents a null value."
   "number_value", ":ref:`ref_double`", "", "Represents a double value."
   "string_value", ":ref:`ref_string`", "", "Represents a string value."
   "bool_value", ":ref:`ref_bool`", "", "Represents a boolean value."
   "struct_value", ":ref:`ref_google.protobuf.Struct`", "", "Represents a structured value."
   "list_value", ":ref:`ref_google.protobuf.ListValue`", "", "Represents a repeated `Value`."






..
   end messages



.. _ref_google.protobuf.NullValue:

NullValue
------------------------------------------------------------------

`NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

 The JSON representation for `NullValue` is JSON `null`.

.. csv-table:: Enum NullValue values
   :header: "Name", "Number", "Description"
   :widths: auto

   "NULL_VALUE", "0", "Null value."


..
   end enums


..
   end HasExtensions


..
   end services



.. _ref_scala_types:

Scalar Value Types
==================



.. _ref_double:

double
-----------------------------



.. csv-table:: double language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "double", "double", "double", "float", "float64", "double", "float", "Float"



.. _ref_float:

float
-----------------------------



.. csv-table:: float language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "float", "float", "float", "float", "float32", "float", "float", "Float"



.. _ref_int32:

int32
-----------------------------

Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead.

.. csv-table:: int32 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "int32", "int32", "int", "int", "int32", "int", "integer", "Bignum or Fixnum (as required)"



.. _ref_int64:

int64
-----------------------------

Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead.

.. csv-table:: int64 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "int64", "int64", "long", "int/long", "int64", "long", "integer/string", "Bignum"



.. _ref_uint32:

uint32
-----------------------------

Uses variable-length encoding.

.. csv-table:: uint32 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "uint32", "uint32", "int", "int/long", "uint32", "uint", "integer", "Bignum or Fixnum (as required)"



.. _ref_uint64:

uint64
-----------------------------

Uses variable-length encoding.

.. csv-table:: uint64 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "uint64", "uint64", "long", "int/long", "uint64", "ulong", "integer/string", "Bignum or Fixnum (as required)"



.. _ref_sint32:

sint32
-----------------------------

Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s.

.. csv-table:: sint32 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "sint32", "int32", "int", "int", "int32", "int", "integer", "Bignum or Fixnum (as required)"



.. _ref_sint64:

sint64
-----------------------------

Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s.

.. csv-table:: sint64 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "sint64", "int64", "long", "int/long", "int64", "long", "integer/string", "Bignum"



.. _ref_fixed32:

fixed32
-----------------------------

Always four bytes. More efficient than uint32 if values are often greater than 2^28.

.. csv-table:: fixed32 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "fixed32", "uint32", "int", "int", "uint32", "uint", "integer", "Bignum or Fixnum (as required)"



.. _ref_fixed64:

fixed64
-----------------------------

Always eight bytes. More efficient than uint64 if values are often greater than 2^56.

.. csv-table:: fixed64 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "fixed64", "uint64", "long", "int/long", "uint64", "ulong", "integer/string", "Bignum"



.. _ref_sfixed32:

sfixed32
-----------------------------

Always four bytes.

.. csv-table:: sfixed32 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "sfixed32", "int32", "int", "int", "int32", "int", "integer", "Bignum or Fixnum (as required)"



.. _ref_sfixed64:

sfixed64
-----------------------------

Always eight bytes.

.. csv-table:: sfixed64 language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "sfixed64", "int64", "long", "int/long", "int64", "long", "integer/string", "Bignum"



.. _ref_bool:

bool
-----------------------------



.. csv-table:: bool language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "bool", "bool", "boolean", "boolean", "bool", "bool", "boolean", "TrueClass/FalseClass"



.. _ref_string:

string
-----------------------------

A string must always contain UTF-8 encoded or 7-bit ASCII text.

.. csv-table:: string language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "string", "string", "String", "str/unicode", "string", "string", "string", "String (UTF-8)"



.. _ref_bytes:

bytes
-----------------------------

May contain any arbitrary sequence of bytes.

.. csv-table:: bytes language representation
   :header: ".proto Type", "C++", "Java", "Python", "Go", "C#", "PHP", "Ruby"
   :widths: auto

   "bytes", "string", "ByteString", "str", "[]byte", "ByteString", "string", "String (ASCII-8BIT)"


..
   end scalars