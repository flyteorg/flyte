// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskStatus {
    #[prost(string, tag = "1")]
    pub task_id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub namespace: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub workflow_id: ::prost::alloc::string::String,
    #[prost(int32, tag = "4")]
    pub phase: i32,
    #[prost(string, tag = "5")]
    pub reason: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Capacity {
    #[prost(int32, tag = "1")]
    pub execution_count: i32,
    #[prost(int32, tag = "2")]
    pub execution_limit: i32,
    #[prost(int32, tag = "3")]
    pub backlog_count: i32,
    #[prost(int32, tag = "4")]
    pub backlog_limit: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HeartbeatRequest {
    #[prost(string, tag = "1")]
    pub worker_id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub queue_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "3")]
    pub capacity: ::core::option::Option<Capacity>,
    #[prost(message, repeated, tag = "4")]
    pub task_statuses: ::prost::alloc::vec::Vec<TaskStatus>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HeartbeatResponse {
    #[prost(string, tag = "1")]
    pub task_id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub namespace: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub workflow_id: ::prost::alloc::string::String,
    #[prost(string, repeated, tag = "4")]
    pub cmd: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(enumeration = "heartbeat_response::Operation", tag = "5")]
    pub operation: i32,
}
/// Nested message and enum types in `HeartbeatResponse`.
pub mod heartbeat_response {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Operation {
        Ack = 0,
        Assign = 1,
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
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FastTaskEnvironment {
    #[prost(string, tag = "1")]
    pub queue_id: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FastTaskEnvironmentSpec {
    #[prost(int32, tag = "1")]
    pub backlog_length: i32,
    #[prost(int32, tag = "2")]
    pub parallelism: i32,
    #[prost(bytes = "vec", tag = "3")]
    pub pod_template_spec: ::prost::alloc::vec::Vec<u8>,
    #[prost(string, tag = "4")]
    pub primary_container_name: ::prost::alloc::string::String,
    #[prost(int32, tag = "5")]
    pub replica_count: i32,
    #[prost(oneof = "fast_task_environment_spec::TerminationCriteria", tags = "6")]
    pub termination_criteria:
        ::core::option::Option<fast_task_environment_spec::TerminationCriteria>,
}
/// Nested message and enum types in `FastTaskEnvironmentSpec`.
pub mod fast_task_environment_spec {
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum TerminationCriteria {
        #[prost(int32, tag = "6")]
        TtlSeconds(i32),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FastTaskAssignment {
    /// Environment ID for this fast task
    #[prost(string, tag = "1")]
    pub environment_id: ::prost::alloc::string::String,
    /// The assigned worker pod name for this fast task, if available
    #[prost(string, tag = "2")]
    pub assigned_worker: ::prost::alloc::string::String,
}
include!("fasttask.tonic.rs");
// @@protoc_insertion_point(module)
