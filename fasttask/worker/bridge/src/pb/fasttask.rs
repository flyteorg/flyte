// @generated
/// The current execution status of a specific fasttask execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskStatus {
    /// The unique identifier for the fasttask execution instance.
    #[prost(string, tag = "1")]
    pub task_id: ::prost::alloc::string::String,
    /// The namespace for the fasttask execution.
    #[prost(string, tag = "2")]
    pub namespace: ::prost::alloc::string::String,
    /// The workflow identifier that triggered the fasttask execution.
    #[prost(string, tag = "3")]
    pub workflow_id: ::prost::alloc::string::String,
    /// The current phase of the fasttask execution.
    #[prost(int32, tag = "4")]
    pub phase: i32,
    /// A brief description to understand why this fasttask execution is in the specified phase.
    /// This is notably useful for explaining failure scenarios.
    #[prost(string, tag = "5")]
    pub reason: ::prost::alloc::string::String,
}
/// The current execution capacity for a fasttask worker replia.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Capacity {
    /// The number of currently active fasttask executions.
    #[prost(int32, tag = "1")]
    pub execution_count: i32,
    /// The total number of acceptable fasttask executions.
    #[prost(int32, tag = "2")]
    pub execution_limit: i32,
    /// The number of currently backlogged fasttask executions.
    #[prost(int32, tag = "3")]
    pub backlog_count: i32,
    /// The total number of acceptable backlogged fasttask executions.
    #[prost(int32, tag = "4")]
    pub backlog_limit: i32,
}
/// Information sent from a fasttask worker replica to the fasttask service reporting the current
/// execution state including available capacity and status' of fasttask executions.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HeartbeatRequest {
    /// The unique identifier for this specific worker replica.
    #[prost(string, tag = "1")]
    pub worker_id: ::prost::alloc::string::String,
    /// The queue that this worker replica queries for fasttask executions.
    #[prost(string, tag = "2")]
    pub queue_id: ::prost::alloc::string::String,
    /// The current capacity including active and backlogged execution assignments.
    #[prost(message, optional, tag = "3")]
    pub capacity: ::core::option::Option<Capacity>,
    /// The status' of currently assigned fasttask executions.
    #[prost(message, repeated, tag = "4")]
    pub task_statuses: ::prost::alloc::vec::Vec<TaskStatus>,
}
/// Information sent from the fasttask service to a fasttask worker replica. This includes all
/// fasttask execution lifecycle management operations.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HeartbeatResponse {
    /// The unique identifier for the fasttask execution instance.
    #[prost(string, tag = "1")]
    pub task_id: ::prost::alloc::string::String,
    /// The namespace for the fasttask execution.
    #[prost(string, tag = "2")]
    pub namespace: ::prost::alloc::string::String,
    /// The workflow identifier that triggered the fasttask execution.
    #[prost(string, tag = "3")]
    pub workflow_id: ::prost::alloc::string::String,
    /// A string array representing the command to be evaluated for the fasttask execution.
    #[prost(string, repeated, tag = "4")]
    pub cmd: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// The operation to perform on this fasttask execution.
    #[prost(enumeration = "heartbeat_response::Operation", tag = "5")]
    pub operation: i32,
}
/// Nested message and enum types in `HeartbeatResponse`.
pub mod heartbeat_response {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Operation {
        /// Acknowledges that the worker replica is still processing the fasttask execution. This is
        /// useful for mitigating failure scenarios where multiple worker replicas may be assigned
        /// the same fasttask execution.
        Ack = 0,
        /// Assigns the fasttask execution to this specific worker replica.
        Assign = 1,
        /// Delete the current execution. For active executions this will kill the child process and
        /// effectively abort the fasttask executions, for completed executions this acknowledges
        /// that a terminal phase has been successfully processed by the fasttask service.
        Delete = 2,
    }
    impl Operation {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Operation::Ack => "ACK",
                Operation::Assign => "ASSIGN",
                Operation::Delete => "DELETE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "ACK" => Some(Self::Ack),
                "ASSIGN" => Some(Self::Assign),
                "DELETE" => Some(Self::Delete),
                _ => None,
            }
        }
    }
}
/// The metadata defining an active fasttask environment.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FastTaskEnvironment {
    /// The queue that replicas of the fasttask environment will query for task executions.
    #[prost(string, tag = "1")]
    pub queue_id: ::prost::alloc::string::String,
}
/// A definition of a fasttask environment.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FastTaskEnvironmentSpec {
    /// The number of items to keep in each replicas backlog. The backlog is used to effectively
    /// queue work on each replica so that it can begin executing another task immediately after
    /// the current one completes. This improves overall throughput at the cost of uneven
    /// distribution.
    #[prost(int32, tag = "1")]
    pub backlog_length: i32,
    /// The number of tasks that can be executed on a replica at the same time.
    #[prost(int32, tag = "2")]
    pub parallelism: i32,
    /// A byte serialized `PodTemplateSpec` that will be used to create replica Pods. If not
    /// provided Flyte will build a `PodTemplateSpec` using the `PodPlugin` as if this fasttask
    /// instance were executed in a singular Pod.
    #[prost(bytes = "vec", tag = "3")]
    pub pod_template_spec: ::prost::alloc::vec::Vec<u8>,
    /// The name of the primary container in the defined `pod_template_spec`. If a
    /// `pod_template_spec` is not provided, this field is unnecessary.
    #[prost(string, tag = "4")]
    pub primary_container_name: ::prost::alloc::string::String,
    /// The number of replicas to initialize for this environment.
    #[prost(int32, tag = "5")]
    pub replica_count: i32,
    /// The criteria to determine how this environment should be deleted.
    #[prost(oneof = "fast_task_environment_spec::TerminationCriteria", tags = "6")]
    pub termination_criteria:
        ::core::option::Option<fast_task_environment_spec::TerminationCriteria>,
}
/// Nested message and enum types in `FastTaskEnvironmentSpec`.
pub mod fast_task_environment_spec {
    /// The criteria to determine how this environment should be deleted.
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum TerminationCriteria {
        /// Indicates the minimum number of seconds after becoming idle (ie. no active task
        /// executions) that this environment will be GCed.
        #[prost(int32, tag = "6")]
        TtlSeconds(i32),
    }
}
/// Represents an fasttask worker replica assignment.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FastTaskAssignment {
    /// Environment organization for this fast task, if available.
    #[prost(string, tag = "1")]
    pub environment_org: ::prost::alloc::string::String,
    /// Environment project for this fast task.
    #[prost(string, tag = "2")]
    pub environment_project: ::prost::alloc::string::String,
    /// Environment domain for this fast task.
    #[prost(string, tag = "3")]
    pub environment_domain: ::prost::alloc::string::String,
    /// Environment name for this fast task.
    #[prost(string, tag = "4")]
    pub environment_name: ::prost::alloc::string::String,
    /// Environment version for this fast task.
    #[prost(string, tag = "5")]
    pub environment_version: ::prost::alloc::string::String,
    /// The assigned worker pod name for this fast task, if available.
    #[prost(string, tag = "6")]
    pub assigned_worker: ::prost::alloc::string::String,
}
include!("fasttask.tonic.rs");
// @@protoc_insertion_point(module)
