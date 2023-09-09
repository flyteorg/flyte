// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionEvent {
    /// Workflow execution id
    #[prost(message, optional, tag="1")]
    pub execution_id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// the id of the originator (Propeller) of the event
    #[prost(string, tag="2")]
    pub producer_id: ::prost::alloc::string::String,
    #[prost(enumeration="super::core::workflow_execution::Phase", tag="3")]
    pub phase: i32,
    /// This timestamp represents when the original event occurred, it is generated
    /// by the executor of the workflow.
    #[prost(message, optional, tag="4")]
    pub occurred_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(oneof="workflow_execution_event::OutputResult", tags="5, 6, 7")]
    pub output_result: ::core::option::Option<workflow_execution_event::OutputResult>,
}
/// Nested message and enum types in `WorkflowExecutionEvent`.
pub mod workflow_execution_event {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum OutputResult {
        /// URL to the output of the execution, it encodes all the information
        /// including Cloud source provider. ie., s3://...
        #[prost(string, tag="5")]
        OutputUri(::prost::alloc::string::String),
        /// Error information for the execution
        #[prost(message, tag="6")]
        Error(super::super::core::ExecutionError),
        /// Raw output data produced by this workflow execution.
        #[prost(message, tag="7")]
        OutputData(super::super::core::LiteralMap),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionEvent {
    /// Unique identifier for this node execution
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::NodeExecutionIdentifier>,
    /// the id of the originator (Propeller) of the event
    #[prost(string, tag="2")]
    pub producer_id: ::prost::alloc::string::String,
    #[prost(enumeration="super::core::node_execution::Phase", tag="3")]
    pub phase: i32,
    /// This timestamp represents when the original event occurred, it is generated
    /// by the executor of the node.
    #[prost(message, optional, tag="4")]
    pub occurred_at: ::core::option::Option<::prost_types::Timestamp>,
    /// [To be deprecated] Specifies which task (if any) launched this node.
    #[prost(message, optional, tag="9")]
    pub parent_task_metadata: ::core::option::Option<ParentTaskExecutionMetadata>,
    /// Specifies the parent node of the current node execution. Node executions at level zero will not have a parent node.
    #[prost(message, optional, tag="10")]
    pub parent_node_metadata: ::core::option::Option<ParentNodeExecutionMetadata>,
    /// Retry group to indicate grouping of nodes by retries
    #[prost(string, tag="11")]
    pub retry_group: ::prost::alloc::string::String,
    /// Identifier of the node in the original workflow/graph
    /// This maps to value of WorkflowTemplate.nodes\[X\].id
    #[prost(string, tag="12")]
    pub spec_node_id: ::prost::alloc::string::String,
    /// Friendly readable name for the node
    #[prost(string, tag="13")]
    pub node_name: ::prost::alloc::string::String,
    #[prost(int32, tag="16")]
    pub event_version: i32,
    /// Whether this node launched a subworkflow.
    #[prost(bool, tag="17")]
    pub is_parent: bool,
    /// Whether this node yielded a dynamic workflow.
    #[prost(bool, tag="18")]
    pub is_dynamic: bool,
    /// String location uniquely identifying where the deck HTML file is
    /// NativeUrl specifies the url in the format of the configured storage provider (e.g. s3://my-bucket/randomstring/suffix.tar)
    #[prost(string, tag="19")]
    pub deck_uri: ::prost::alloc::string::String,
    /// This timestamp represents the instant when the event was reported by the executing framework. For example,
    /// when first processing a node the `occurred_at` timestamp should be the instant propeller makes progress, so when
    /// literal inputs are initially copied. The event however will not be sent until after the copy completes.
    /// Extracting both of these timestamps facilitates a more accurate portrayal of the evaluation time-series.
    #[prost(message, optional, tag="21")]
    pub reported_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(oneof="node_execution_event::InputValue", tags="5, 20")]
    pub input_value: ::core::option::Option<node_execution_event::InputValue>,
    #[prost(oneof="node_execution_event::OutputResult", tags="6, 7, 15")]
    pub output_result: ::core::option::Option<node_execution_event::OutputResult>,
    /// Additional metadata to do with this event's node target based
    /// on the node type
    #[prost(oneof="node_execution_event::TargetMetadata", tags="8, 14")]
    pub target_metadata: ::core::option::Option<node_execution_event::TargetMetadata>,
}
/// Nested message and enum types in `NodeExecutionEvent`.
pub mod node_execution_event {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum InputValue {
        #[prost(string, tag="5")]
        InputUri(::prost::alloc::string::String),
        /// Raw input data consumed by this node execution.
        #[prost(message, tag="20")]
        InputData(super::super::core::LiteralMap),
    }
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum OutputResult {
        /// URL to the output of the execution, it encodes all the information
        /// including Cloud source provider. ie., s3://...
        #[prost(string, tag="6")]
        OutputUri(::prost::alloc::string::String),
        /// Error information for the execution
        #[prost(message, tag="7")]
        Error(super::super::core::ExecutionError),
        /// Raw output data produced by this node execution.
        #[prost(message, tag="15")]
        OutputData(super::super::core::LiteralMap),
    }
    /// Additional metadata to do with this event's node target based
    /// on the node type
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum TargetMetadata {
        #[prost(message, tag="8")]
        WorkflowNodeMetadata(super::WorkflowNodeMetadata),
        #[prost(message, tag="14")]
        TaskNodeMetadata(super::TaskNodeMetadata),
    }
}
/// For Workflow Nodes we need to send information about the workflow that's launched
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowNodeMetadata {
    #[prost(message, optional, tag="1")]
    pub execution_id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskNodeMetadata {
    /// Captures the status of caching for this execution.
    #[prost(enumeration="super::core::CatalogCacheStatus", tag="1")]
    pub cache_status: i32,
    /// This structure carries the catalog artifact information
    #[prost(message, optional, tag="2")]
    pub catalog_key: ::core::option::Option<super::core::CatalogMetadata>,
    /// Captures the status of cache reservations for this execution.
    #[prost(enumeration="super::core::catalog_reservation::Status", tag="3")]
    pub reservation_status: i32,
    /// The latest checkpoint location
    #[prost(string, tag="4")]
    pub checkpoint_uri: ::prost::alloc::string::String,
    /// In the case this task launched a dynamic workflow we capture its structure here.
    #[prost(message, optional, tag="16")]
    pub dynamic_workflow: ::core::option::Option<DynamicWorkflowNodeMetadata>,
}
/// For dynamic workflow nodes we send information about the dynamic workflow definition that gets generated.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DynamicWorkflowNodeMetadata {
    /// id represents the unique identifier of the workflow.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// Represents the compiled representation of the embedded dynamic workflow.
    #[prost(message, optional, tag="2")]
    pub compiled_workflow: ::core::option::Option<super::core::CompiledWorkflowClosure>,
    /// dynamic_job_spec_uri is the location of the DynamicJobSpec proto message for this DynamicWorkflow. This is
    /// required to correctly recover partially completed executions where the workflow has already been compiled.
    #[prost(string, tag="3")]
    pub dynamic_job_spec_uri: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ParentTaskExecutionMetadata {
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::TaskExecutionIdentifier>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ParentNodeExecutionMetadata {
    /// Unique identifier of the parent node id within the execution
    /// This is value of core.NodeExecutionIdentifier.node_id of the parent node 
    #[prost(string, tag="1")]
    pub node_id: ::prost::alloc::string::String,
}
/// Plugin specific execution event information. For tasks like Python, Hive, Spark, DynamicJob.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionEvent {
    /// ID of the task. In combination with the retryAttempt this will indicate
    /// the task execution uniquely for a given parent node execution.
    #[prost(message, optional, tag="1")]
    pub task_id: ::core::option::Option<super::core::Identifier>,
    /// A task execution is always kicked off by a node execution, the event consumer
    /// will use the parent_id to relate the task to it's parent node execution
    #[prost(message, optional, tag="2")]
    pub parent_node_execution_id: ::core::option::Option<super::core::NodeExecutionIdentifier>,
    /// retry attempt number for this task, ie., 2 for the second attempt
    #[prost(uint32, tag="3")]
    pub retry_attempt: u32,
    /// Phase associated with the event
    #[prost(enumeration="super::core::task_execution::Phase", tag="4")]
    pub phase: i32,
    /// id of the process that sent this event, mainly for trace debugging
    #[prost(string, tag="5")]
    pub producer_id: ::prost::alloc::string::String,
    /// log information for the task execution
    #[prost(message, repeated, tag="6")]
    pub logs: ::prost::alloc::vec::Vec<super::core::TaskLog>,
    /// This timestamp represents when the original event occurred, it is generated
    /// by the executor of the task.
    #[prost(message, optional, tag="7")]
    pub occurred_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Custom data that the task plugin sends back. This is extensible to allow various plugins in the system.
    #[prost(message, optional, tag="11")]
    pub custom_info: ::core::option::Option<::prost_types::Struct>,
    /// Some phases, like RUNNING, can send multiple events with changed metadata (new logs, additional custom_info, etc)
    /// that should be recorded regardless of the lack of phase change.
    /// The version field should be incremented when metadata changes across the duration of an individual phase.
    #[prost(uint32, tag="12")]
    pub phase_version: u32,
    /// An optional explanation for the phase transition.
    #[prost(string, tag="13")]
    pub reason: ::prost::alloc::string::String,
    /// A predefined yet extensible Task type identifier. If the task definition is already registered in flyte admin
    /// this type will be identical, but not all task executions necessarily use pre-registered definitions and this
    /// type is useful to render the task in the UI, filter task executions, etc.
    #[prost(string, tag="14")]
    pub task_type: ::prost::alloc::string::String,
    /// Metadata around how a task was executed.
    #[prost(message, optional, tag="16")]
    pub metadata: ::core::option::Option<TaskExecutionMetadata>,
    /// The event version is used to indicate versioned changes in how data is reported using this
    /// proto message. For example, event_verison > 0 means that maps tasks report logs using the
    /// TaskExecutionMetadata ExternalResourceInfo fields for each subtask rather than the TaskLog
    /// in this message.
    #[prost(int32, tag="18")]
    pub event_version: i32,
    /// This timestamp represents the instant when the event was reported by the executing framework. For example, a k8s
    /// pod task may be marked completed at (ie. `occurred_at`) the instant the container running user code completes,
    /// but this event will not be reported until the pod is marked as completed. Extracting both of these timestamps
    /// facilitates a more accurate portrayal of the evaluation time-series. 
    #[prost(message, optional, tag="20")]
    pub reported_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(oneof="task_execution_event::InputValue", tags="8, 19")]
    pub input_value: ::core::option::Option<task_execution_event::InputValue>,
    #[prost(oneof="task_execution_event::OutputResult", tags="9, 10, 17")]
    pub output_result: ::core::option::Option<task_execution_event::OutputResult>,
}
/// Nested message and enum types in `TaskExecutionEvent`.
pub mod task_execution_event {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum InputValue {
        /// URI of the input file, it encodes all the information
        /// including Cloud source provider. ie., s3://...
        #[prost(string, tag="8")]
        InputUri(::prost::alloc::string::String),
        /// Raw input data consumed by this task execution.
        #[prost(message, tag="19")]
        InputData(super::super::core::LiteralMap),
    }
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum OutputResult {
        /// URI to the output of the execution, it will be in a format that encodes all the information
        /// including Cloud source provider. ie., s3://...
        #[prost(string, tag="9")]
        OutputUri(::prost::alloc::string::String),
        /// Error information for the execution
        #[prost(message, tag="10")]
        Error(super::super::core::ExecutionError),
        /// Raw output data produced by this task execution.
        #[prost(message, tag="17")]
        OutputData(super::super::core::LiteralMap),
    }
}
/// This message contains metadata about external resources produced or used by a specific task execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExternalResourceInfo {
    /// Identifier for an external resource created by this task execution, for example Qubole query ID or presto query ids.
    #[prost(string, tag="1")]
    pub external_id: ::prost::alloc::string::String,
    /// A unique index for the external resource with respect to all external resources for this task. Although the
    /// identifier may change between task reporting events or retries, this will remain the same to enable aggregating
    /// information from multiple reports.
    #[prost(uint32, tag="2")]
    pub index: u32,
    /// Retry attempt number for this external resource, ie., 2 for the second attempt
    #[prost(uint32, tag="3")]
    pub retry_attempt: u32,
    /// Phase associated with the external resource
    #[prost(enumeration="super::core::task_execution::Phase", tag="4")]
    pub phase: i32,
    /// Captures the status of caching for this external resource execution.
    #[prost(enumeration="super::core::CatalogCacheStatus", tag="5")]
    pub cache_status: i32,
    /// log information for the external resource execution
    #[prost(message, repeated, tag="6")]
    pub logs: ::prost::alloc::vec::Vec<super::core::TaskLog>,
}
/// This message holds task execution metadata specific to resource allocation used to manage concurrent
/// executions for a project namespace.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ResourcePoolInfo {
    /// Unique resource ID used to identify this execution when allocating a token.
    #[prost(string, tag="1")]
    pub allocation_token: ::prost::alloc::string::String,
    /// Namespace under which this task execution requested an allocation token.
    #[prost(string, tag="2")]
    pub namespace: ::prost::alloc::string::String,
}
/// Holds metadata around how a task was executed.
/// As a task transitions across event phases during execution some attributes, such its generated name, generated external resources,
/// and more may grow in size but not change necessarily based on the phase transition that sparked the event update.
/// Metadata is a container for these attributes across the task execution lifecycle.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionMetadata {
    /// Unique, generated name for this task execution used by the backend.
    #[prost(string, tag="1")]
    pub generated_name: ::prost::alloc::string::String,
    /// Additional data on external resources on other back-ends or platforms (e.g. Hive, Qubole, etc) launched by this task execution.
    #[prost(message, repeated, tag="2")]
    pub external_resources: ::prost::alloc::vec::Vec<ExternalResourceInfo>,
    /// Includes additional data on concurrent resource management used during execution..
    /// This is a repeated field because a plugin can request multiple resource allocations during execution.
    #[prost(message, repeated, tag="3")]
    pub resource_pool_info: ::prost::alloc::vec::Vec<ResourcePoolInfo>,
    /// The identifier of the plugin used to execute this task.
    #[prost(string, tag="4")]
    pub plugin_identifier: ::prost::alloc::string::String,
    #[prost(enumeration="task_execution_metadata::InstanceClass", tag="16")]
    pub instance_class: i32,
}
/// Nested message and enum types in `TaskExecutionMetadata`.
pub mod task_execution_metadata {
    /// Includes the broad category of machine used for this specific task execution.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum InstanceClass {
        /// The default instance class configured for the flyte application platform.
        Default = 0,
        /// The instance class configured for interruptible tasks.
        Interruptible = 1,
    }
    impl InstanceClass {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                InstanceClass::Default => "DEFAULT",
                InstanceClass::Interruptible => "INTERRUPTIBLE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "DEFAULT" => Some(Self::Default),
                "INTERRUPTIBLE" => Some(Self::Interruptible),
                _ => None,
            }
        }
    }
}
// @@protoc_insertion_point(module)
