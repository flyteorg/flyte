// @generated
// This file is @generated by prost-build.
/// Represents a subset of runtime task execution metadata that are relevant to external plugins.
///
/// ID of the task execution
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionMetadata {
    #[prost(message, optional, tag="1")]
    pub task_execution_id: ::core::option::Option<super::core::TaskExecutionIdentifier>,
    /// k8s namespace where the task is executed in
    #[prost(string, tag="2")]
    pub namespace: ::prost::alloc::string::String,
    /// Labels attached to the task execution
    #[prost(map="string, string", tag="3")]
    pub labels: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// Annotations attached to the task execution
    #[prost(map="string, string", tag="4")]
    pub annotations: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// k8s service account associated with the task execution
    #[prost(string, tag="5")]
    pub k8s_service_account: ::prost::alloc::string::String,
    /// Environment variables attached to the task execution
    #[prost(map="string, string", tag="6")]
    pub environment_variables: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// Represents the maximum number of attempts allowed for a task.
    /// If a task fails, it can be retried up to this maximum number of attempts.
    #[prost(int32, tag="7")]
    pub max_attempts: i32,
    /// Indicates whether the task execution can be interrupted.
    /// If set to true, the task can be stopped before completion.
    #[prost(bool, tag="8")]
    pub interruptible: bool,
    /// Specifies the threshold for failure count at which the interruptible property
    /// will take effect. If the number of consecutive task failures exceeds this threshold,
    /// interruptible behavior will be activated.
    #[prost(int32, tag="9")]
    pub interruptible_failure_threshold: i32,
    /// Overrides for specific properties of the task node.
    /// These overrides can be used to customize the behavior of the task node.
    #[prost(message, optional, tag="10")]
    pub overrides: ::core::option::Option<super::core::TaskNodeOverrides>,
    /// Identity of user running this task execution
    #[prost(message, optional, tag="11")]
    pub identity: ::core::option::Option<super::core::Identity>,
}
/// Represents a request structure to create task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateTaskRequest {
    /// The inputs required to start the execution. All required inputs must be
    /// included in this map. If not required and not provided, defaults apply.
    /// +optional
    #[prost(message, optional, tag="1")]
    pub inputs: ::core::option::Option<super::core::LiteralMap>,
    /// Template of the task that encapsulates all the metadata of the task.
    #[prost(message, optional, tag="2")]
    pub template: ::core::option::Option<super::core::TaskTemplate>,
    /// Prefix for where task output data will be written. (e.g. s3://my-bucket/randomstring)
    #[prost(string, tag="3")]
    pub output_prefix: ::prost::alloc::string::String,
    /// subset of runtime task execution metadata.
    #[prost(message, optional, tag="4")]
    pub task_execution_metadata: ::core::option::Option<TaskExecutionMetadata>,
}
/// Represents a create response structure.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateTaskResponse {
    /// ResourceMeta is created by the agent. It could be a string (jobId) or a dict (more complex metadata).
    #[prost(bytes="vec", tag="1")]
    pub resource_meta: ::prost::alloc::vec::Vec<u8>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateRequestHeader {
    /// Template of the task that encapsulates all the metadata of the task.
    #[prost(message, optional, tag="1")]
    pub template: ::core::option::Option<super::core::TaskTemplate>,
    /// Prefix for where task output data will be written. (e.g. s3://my-bucket/randomstring)
    #[prost(string, tag="2")]
    pub output_prefix: ::prost::alloc::string::String,
    /// subset of runtime task execution metadata.
    #[prost(message, optional, tag="3")]
    pub task_execution_metadata: ::core::option::Option<TaskExecutionMetadata>,
    /// MaxDatasetSizeBytes is the maximum size of the dataset that can be generated by the task.
    #[prost(int64, tag="4")]
    pub max_dataset_size_bytes: i64,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecuteTaskSyncRequest {
    #[prost(oneof="execute_task_sync_request::Part", tags="1, 2")]
    pub part: ::core::option::Option<execute_task_sync_request::Part>,
}
/// Nested message and enum types in `ExecuteTaskSyncRequest`.
pub mod execute_task_sync_request {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Part {
        #[prost(message, tag="1")]
        Header(super::CreateRequestHeader),
        #[prost(message, tag="2")]
        Inputs(super::super::core::LiteralMap),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecuteTaskSyncResponseHeader {
    #[prost(message, optional, tag="1")]
    pub resource: ::core::option::Option<Resource>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecuteTaskSyncResponse {
    /// Metadata is created by the agent. It could be a string (jobId) or a dict (more complex metadata).
    /// Resource is for synchronous task execution.
    #[prost(oneof="execute_task_sync_response::Res", tags="1, 2")]
    pub res: ::core::option::Option<execute_task_sync_response::Res>,
}
/// Nested message and enum types in `ExecuteTaskSyncResponse`.
pub mod execute_task_sync_response {
    /// Metadata is created by the agent. It could be a string (jobId) or a dict (more complex metadata).
    /// Resource is for synchronous task execution.
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Res {
        #[prost(message, tag="1")]
        Header(super::ExecuteTaskSyncResponseHeader),
        #[prost(message, tag="2")]
        Outputs(super::super::core::LiteralMap),
    }
}
/// A message used to fetch a job resource from flyte agent server.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskRequest {
    /// A predefined yet extensible Task type identifier.
    #[deprecated]
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// Metadata about the resource to be pass to the agent.
    #[prost(bytes="vec", tag="2")]
    pub resource_meta: ::prost::alloc::vec::Vec<u8>,
    /// A predefined yet extensible Task type identifier.
    #[prost(message, optional, tag="3")]
    pub task_category: ::core::option::Option<TaskCategory>,
    /// Prefix for where task output data will be written. (e.g. s3://my-bucket/randomstring)
    #[prost(string, tag="4")]
    pub output_prefix: ::prost::alloc::string::String,
}
/// Response to get an individual task resource.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskResponse {
    #[prost(message, optional, tag="1")]
    pub resource: ::core::option::Option<Resource>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Resource {
    /// DEPRECATED. The state of the execution is used to control its visibility in the UI/CLI.
    #[deprecated]
    #[prost(enumeration="State", tag="1")]
    pub state: i32,
    /// The outputs of the execution. It's typically used by sql task. Agent service will create a
    /// Structured dataset pointing to the query result table.
    /// +optional
    #[prost(message, optional, tag="2")]
    pub outputs: ::core::option::Option<super::core::LiteralMap>,
    /// A descriptive message for the current state. e.g. waiting for cluster.
    #[prost(string, tag="3")]
    pub message: ::prost::alloc::string::String,
    /// log information for the task execution.
    #[prost(message, repeated, tag="4")]
    pub log_links: ::prost::alloc::vec::Vec<super::core::TaskLog>,
    /// The phase of the execution is used to determine the phase of the plugin's execution.
    #[prost(enumeration="super::core::task_execution::Phase", tag="5")]
    pub phase: i32,
    /// Custom data specific to the agent.
    #[prost(message, optional, tag="6")]
    pub custom_info: ::core::option::Option<::prost_types::Struct>,
    /// The error raised during execution
    #[prost(message, optional, tag="7")]
    pub agent_error: ::core::option::Option<AgentError>,
}
/// A message used to delete a task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeleteTaskRequest {
    /// A predefined yet extensible Task type identifier.
    #[deprecated]
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// Metadata about the resource to be pass to the agent.
    #[prost(bytes="vec", tag="2")]
    pub resource_meta: ::prost::alloc::vec::Vec<u8>,
    /// A predefined yet extensible Task type identifier.
    #[prost(message, optional, tag="3")]
    pub task_category: ::core::option::Option<TaskCategory>,
}
/// Response to delete a task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct DeleteTaskResponse {
}
/// A message containing the agent metadata.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Agent {
    /// Name is the developer-assigned name of the agent.
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    /// SupportedTaskTypes are the types of the tasks that the agent can handle.
    #[deprecated]
    #[prost(string, repeated, tag="2")]
    pub supported_task_types: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// IsSync indicates whether this agent is a sync agent. Sync agents are expected to return their
    /// results synchronously when called by propeller. Given that sync agents can affect the performance
    /// of the system, it's important to enforce strict timeout policies.
    /// An Async agent, on the other hand, is required to be able to identify jobs by an
    /// identifier and query for job statuses as jobs progress.
    #[prost(bool, tag="3")]
    pub is_sync: bool,
    /// Supported_task_categories are the categories of the tasks that the agent can handle.
    #[prost(message, repeated, tag="4")]
    pub supported_task_categories: ::prost::alloc::vec::Vec<TaskCategory>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskCategory {
    /// The name of the task type.
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    /// The version of the task type.
    #[prost(int32, tag="2")]
    pub version: i32,
}
/// A request to get an agent.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetAgentRequest {
    /// The name of the agent.
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
}
/// A response containing an agent.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetAgentResponse {
    #[prost(message, optional, tag="1")]
    pub agent: ::core::option::Option<Agent>,
}
/// A request to list all agents.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct ListAgentsRequest {
}
/// A response containing a list of agents.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListAgentsResponse {
    #[prost(message, repeated, tag="1")]
    pub agents: ::prost::alloc::vec::Vec<Agent>,
}
/// A request to get the metrics from a task execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskMetricsRequest {
    /// A predefined yet extensible Task type identifier.
    #[deprecated]
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// Metadata is created by the agent. It could be a string (jobId) or a dict (more complex metadata).
    #[prost(bytes="vec", tag="2")]
    pub resource_meta: ::prost::alloc::vec::Vec<u8>,
    /// The metrics to query. If empty, will return a default set of metrics.
    /// e.g. EXECUTION_METRIC_USED_CPU_AVG or EXECUTION_METRIC_USED_MEMORY_BYTES_AVG
    #[prost(string, repeated, tag="3")]
    pub queries: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// Start timestamp, inclusive.
    #[prost(message, optional, tag="4")]
    pub start_time: ::core::option::Option<::prost_types::Timestamp>,
    /// End timestamp, inclusive..
    #[prost(message, optional, tag="5")]
    pub end_time: ::core::option::Option<::prost_types::Timestamp>,
    /// Query resolution step width in duration format or float number of seconds.
    #[prost(message, optional, tag="6")]
    pub step: ::core::option::Option<::prost_types::Duration>,
    /// A predefined yet extensible Task type identifier.
    #[prost(message, optional, tag="7")]
    pub task_category: ::core::option::Option<TaskCategory>,
}
/// A response containing a list of metrics for a task execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskMetricsResponse {
    /// The execution metric results.
    #[prost(message, repeated, tag="1")]
    pub results: ::prost::alloc::vec::Vec<super::core::ExecutionMetricResult>,
}
/// A request to get the log from a task execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskLogsRequest {
    /// A predefined yet extensible Task type identifier.
    #[deprecated]
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// Metadata is created by the agent. It could be a string (jobId) or a dict (more complex metadata).
    #[prost(bytes="vec", tag="2")]
    pub resource_meta: ::prost::alloc::vec::Vec<u8>,
    /// Number of lines to return.
    #[prost(uint64, tag="3")]
    pub lines: u64,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="4")]
    pub token: ::prost::alloc::string::String,
    /// A predefined yet extensible Task type identifier.
    #[prost(message, optional, tag="5")]
    pub task_category: ::core::option::Option<TaskCategory>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskLogsResponseHeader {
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="1")]
    pub token: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskLogsResponseBody {
    /// The execution log results.
    #[prost(string, repeated, tag="1")]
    pub results: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// A response containing the logs for a task execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskLogsResponse {
    #[prost(oneof="get_task_logs_response::Part", tags="1, 2")]
    pub part: ::core::option::Option<get_task_logs_response::Part>,
}
/// Nested message and enum types in `GetTaskLogsResponse`.
pub mod get_task_logs_response {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Part {
        #[prost(message, tag="1")]
        Header(super::GetTaskLogsResponseHeader),
        #[prost(message, tag="2")]
        Body(super::GetTaskLogsResponseBody),
    }
}
/// Error message to propagate detailed errors from agent executions to the execution
/// engine.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AgentError {
    /// A simplified code for errors, so that we can provide a glossary of all possible errors.
    #[prost(string, tag="1")]
    pub code: ::prost::alloc::string::String,
    /// An abstract error kind for this error. Defaults to Non_Recoverable if not specified.
    #[prost(enumeration="agent_error::Kind", tag="3")]
    pub kind: i32,
    /// Defines the origin of the error (system, user, unknown).
    #[prost(enumeration="super::core::execution_error::ErrorKind", tag="4")]
    pub origin: i32,
}
/// Nested message and enum types in `AgentError`.
pub mod agent_error {
    /// Defines a generic error type that dictates the behavior of the retry strategy.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Kind {
        NonRecoverable = 0,
        Recoverable = 1,
    }
    impl Kind {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Kind::NonRecoverable => "NON_RECOVERABLE",
                Kind::Recoverable => "RECOVERABLE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "NON_RECOVERABLE" => Some(Self::NonRecoverable),
                "RECOVERABLE" => Some(Self::Recoverable),
                _ => None,
            }
        }
    }
}
/// The state of the execution is used to control its visibility in the UI/CLI.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum State {
    RetryableFailure = 0,
    PermanentFailure = 1,
    Pending = 2,
    Running = 3,
    Succeeded = 4,
}
impl State {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            State::RetryableFailure => "RETRYABLE_FAILURE",
            State::PermanentFailure => "PERMANENT_FAILURE",
            State::Pending => "PENDING",
            State::Running => "RUNNING",
            State::Succeeded => "SUCCEEDED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "RETRYABLE_FAILURE" => Some(Self::RetryableFailure),
            "PERMANENT_FAILURE" => Some(Self::PermanentFailure),
            "PENDING" => Some(Self::Pending),
            "RUNNING" => Some(Self::Running),
            "SUCCEEDED" => Some(Self::Succeeded),
            _ => None,
        }
    }
}
// @@protoc_insertion_point(module)
