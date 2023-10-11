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
    /// Outputs of tasks will have this.
    #[prost(message, optional, tag="5")]
    pub task_execution: ::core::option::Option<super::core::TaskExecutionIdentifier>,
    /// Workflow outputs will have this.
    #[prost(message, optional, tag="6")]
    pub execution: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// Uploads, either from the UI or from the CLI, or FlyteRemote, will have this.
    #[prost(string, tag="7")]
    pub principal: ::prost::alloc::string::String,
    #[prost(string, tag="8")]
    pub short_description: ::prost::alloc::string::String,
    #[prost(string, tag="9")]
    pub long_description: ::prost::alloc::string::String,
    /// Additional user metadata
    #[prost(message, optional, tag="10")]
    pub user_metadata: ::core::option::Option<::prost_types::Any>,
    #[prost(string, tag="11")]
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
pub struct ListArtifactNamesRequest {
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListArtifactNamesResponse {
    #[prost(message, repeated, tag="1")]
    pub artifact_keys: ::prost::alloc::vec::Vec<super::core::ArtifactKey>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListArtifactsRequest {
    #[prost(message, optional, tag="1")]
    pub artifact_key: ::core::option::Option<super::core::ArtifactKey>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListArtifactsResponse {
    #[prost(message, repeated, tag="1")]
    pub artifacts: ::prost::alloc::vec::Vec<Artifact>,
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
// @@protoc_insertion_point(module)
