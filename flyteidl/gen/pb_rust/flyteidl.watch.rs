// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WatchExecutionStatusUpdatesRequest {
    /// In a multi-cluster setup, propeller should only request executions that were assigned to a given cluster
    #[prost(string, tag="1")]
    pub cluster: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WatchExecutionStatusUpdatesResponse {
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    #[prost(enumeration="super::core::workflow_execution::Phase", tag="2")]
    pub phase: i32,
    /// May only be returned when phase is SUCCEEDED
    #[prost(string, tag="3")]
    pub output_uri: ::prost::alloc::string::String,
    /// May only be returned when phase is FAILED
    #[prost(message, optional, tag="4")]
    pub error: ::core::option::Option<super::core::ExecutionError>,
}
// @@protoc_insertion_point(module)
