// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Artifact {
    #[prost(message, optional, tag="1")]
    pub artifact_id: ::core::option::Option<super::core::ArtifactId>,
    #[prost(message, optional, tag="2")]
    pub spec: ::core::option::Option<ArtifactSpec>,
    /// references the tag field in ArtifactTag
    #[prost(string, repeated, tag="3")]
    pub tags: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(message, optional, tag="4")]
    pub source: ::core::option::Option<ArtifactSource>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateArtifactRequest {
    /// Specify just project/domain on creation
    #[prost(message, optional, tag="1")]
    pub artifact_key: ::core::option::Option<super::core::ArtifactKey>,
    #[prost(string, tag="3")]
    pub version: ::prost::alloc::string::String,
    #[prost(message, optional, tag="2")]
    pub spec: ::core::option::Option<ArtifactSpec>,
    #[prost(map="string, string", tag="4")]
    pub partitions: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    #[prost(string, tag="5")]
    pub tag: ::prost::alloc::string::String,
    #[prost(message, optional, tag="6")]
    pub source: ::core::option::Option<ArtifactSource>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ArtifactSource {
    #[prost(message, optional, tag="1")]
    pub workflow_execution: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    #[prost(string, tag="2")]
    pub node_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag="3")]
    pub task_id: ::core::option::Option<super::core::Identifier>,
    #[prost(uint32, tag="4")]
    pub retry_attempt: u32,
    /// Uploads, either from the UI or from the CLI, or FlyteRemote, will have this.
    #[prost(string, tag="5")]
    pub principal: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ArtifactSpec {
    #[prost(message, optional, tag="1")]
    pub value: ::core::option::Option<super::core::Literal>,
    /// This type will not form part of the artifact key, so for user-named artifacts, if the user changes the type, but
    /// forgets to change the name, that is okay. And the reason why this is a separate field is because adding the
    /// type to all Literals is a lot of work.
    #[prost(message, optional, tag="2")]
    pub r#type: ::core::option::Option<super::core::LiteralType>,
    #[prost(string, tag="3")]
    pub short_description: ::prost::alloc::string::String,
    /// Additional user metadata
    #[prost(message, optional, tag="4")]
    pub user_metadata: ::core::option::Option<::prost_types::Any>,
    #[prost(string, tag="5")]
    pub metadata_type: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateArtifactResponse {
    #[prost(message, optional, tag="1")]
    pub artifact: ::core::option::Option<Artifact>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetArtifactRequest {
    #[prost(message, optional, tag="1")]
    pub query: ::core::option::Option<super::core::ArtifactQuery>,
    /// If false, then long_description is not returned.
    #[prost(bool, tag="2")]
    pub details: bool,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetArtifactResponse {
    #[prost(message, optional, tag="1")]
    pub artifact: ::core::option::Option<Artifact>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SearchOptions {
    /// If true, this means a strict partition search. meaning if you don't specify the partition
    /// field, that will mean, non-partitioned, rather than any partition.
    #[prost(bool, tag="1")]
    pub strict_partitions: bool,
    /// If true, only one artifact per key will be returned. It will be the latest one by creation time.
    #[prost(bool, tag="2")]
    pub latest_by_key: bool,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SearchArtifactsRequest {
    #[prost(message, optional, tag="1")]
    pub artifact_key: ::core::option::Option<super::core::ArtifactKey>,
    #[prost(message, optional, tag="2")]
    pub partitions: ::core::option::Option<super::core::Partitions>,
    #[prost(string, tag="3")]
    pub principal: ::prost::alloc::string::String,
    #[prost(string, tag="4")]
    pub version: ::prost::alloc::string::String,
    #[prost(message, optional, tag="5")]
    pub options: ::core::option::Option<SearchOptions>,
    #[prost(string, tag="6")]
    pub token: ::prost::alloc::string::String,
    #[prost(int32, tag="7")]
    pub limit: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SearchArtifactsResponse {
    /// If artifact specs are not requested, the resultant artifacts may be empty.
    #[prost(message, repeated, tag="1")]
    pub artifacts: ::prost::alloc::vec::Vec<Artifact>,
    /// continuation token if relevant.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FindByWorkflowExecRequest {
    #[prost(message, optional, tag="1")]
    pub exec_id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    #[prost(enumeration="find_by_workflow_exec_request::Direction", tag="2")]
    pub direction: i32,
}
/// Nested message and enum types in `FindByWorkflowExecRequest`.
pub mod find_by_workflow_exec_request {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Direction {
        Inputs = 0,
        Outputs = 1,
    }
    impl Direction {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Direction::Inputs => "INPUTS",
                Direction::Outputs => "OUTPUTS",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "INPUTS" => Some(Self::Inputs),
                "OUTPUTS" => Some(Self::Outputs),
                _ => None,
            }
        }
    }
}
/// Aliases identify a particular version of an artifact. They are different than tags in that they
/// have to be unique for a given artifact project/domain/name. That is, for a given project/domain/name/kind,
/// at most one version can have any given value at any point.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AddTagRequest {
    #[prost(message, optional, tag="1")]
    pub artifact_id: ::core::option::Option<super::core::ArtifactId>,
    #[prost(string, tag="2")]
    pub value: ::prost::alloc::string::String,
    /// If true, and another version already has the specified kind/value, set this version instead
    #[prost(bool, tag="3")]
    pub overwrite: bool,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AddTagResponse {
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateTriggerRequest {
    #[prost(message, optional, tag="1")]
    pub trigger_launch_plan: ::core::option::Option<super::admin::LaunchPlan>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateTriggerResponse {
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeleteTriggerRequest {
    #[prost(message, optional, tag="1")]
    pub trigger_id: ::core::option::Option<super::core::Identifier>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeleteTriggerResponse {
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ArtifactProducer {
    /// These can be tasks, and workflows. Keeping track of the launch plans that a given workflow has is purely in
    /// Admin's domain.
    #[prost(message, optional, tag="1")]
    pub entity_id: ::core::option::Option<super::core::Identifier>,
    #[prost(message, optional, tag="2")]
    pub outputs: ::core::option::Option<super::core::VariableMap>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RegisterProducerRequest {
    #[prost(message, repeated, tag="1")]
    pub producers: ::prost::alloc::vec::Vec<ArtifactProducer>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ArtifactConsumer {
    /// These should all be launch plan IDs
    #[prost(message, optional, tag="1")]
    pub entity_id: ::core::option::Option<super::core::Identifier>,
    #[prost(message, optional, tag="2")]
    pub inputs: ::core::option::Option<super::core::ParameterMap>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RegisterConsumerRequest {
    #[prost(message, repeated, tag="1")]
    pub consumers: ::prost::alloc::vec::Vec<ArtifactConsumer>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RegisterResponse {
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionInputsRequest {
    #[prost(message, optional, tag="1")]
    pub execution_id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// can make this a map in the future, currently no need.
    #[prost(message, repeated, tag="2")]
    pub inputs: ::prost::alloc::vec::Vec<super::core::ArtifactId>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionInputsResponse {
}
// @@protoc_insertion_point(module)
