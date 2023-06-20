// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct OAuth2MetadataRequest {
}
/// OAuth2MetadataResponse defines an RFC-Compliant response for /.well-known/oauth-authorization-server metadata
/// as defined in <https://tools.ietf.org/html/rfc8414>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct OAuth2MetadataResponse {
    /// Defines the issuer string in all JWT tokens this server issues. The issuer can be admin itself or an external
    /// issuer.
    #[prost(string, tag="1")]
    pub issuer: ::prost::alloc::string::String,
    /// URL of the authorization server's authorization endpoint \[RFC6749\]. This is REQUIRED unless no grant types are
    /// supported that use the authorization endpoint.
    #[prost(string, tag="2")]
    pub authorization_endpoint: ::prost::alloc::string::String,
    /// URL of the authorization server's token endpoint \[RFC6749\].
    #[prost(string, tag="3")]
    pub token_endpoint: ::prost::alloc::string::String,
    /// Array containing a list of the OAuth 2.0 response_type values that this authorization server supports.
    #[prost(string, repeated, tag="4")]
    pub response_types_supported: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// JSON array containing a list of the OAuth 2.0 \[RFC6749\] scope values that this authorization server supports.
    #[prost(string, repeated, tag="5")]
    pub scopes_supported: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// JSON array containing a list of client authentication methods supported by this token endpoint.
    #[prost(string, repeated, tag="6")]
    pub token_endpoint_auth_methods_supported: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// URL of the authorization server's JWK Set \[JWK\] document. The referenced document contains the signing key(s) the
    /// client uses to validate signatures from the authorization server.
    #[prost(string, tag="7")]
    pub jwks_uri: ::prost::alloc::string::String,
    /// JSON array containing a list of Proof Key for Code Exchange (PKCE) \[RFC7636\] code challenge methods supported by
    /// this authorization server.
    #[prost(string, repeated, tag="8")]
    pub code_challenge_methods_supported: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// JSON array containing a list of the OAuth 2.0 grant type values that this authorization server supports.
    #[prost(string, repeated, tag="9")]
    pub grant_types_supported: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// URL of the authorization server's device authorization endpoint, as defined in Section 3.1 of \[RFC8628\]
    #[prost(string, tag="10")]
    pub device_authorization_endpoint: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PublicClientAuthConfigRequest {
}
/// FlyteClientResponse encapsulates public information that flyte clients (CLIs... etc.) can use to authenticate users.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PublicClientAuthConfigResponse {
    /// client_id to use when initiating OAuth2 authorization requests.
    #[prost(string, tag="1")]
    pub client_id: ::prost::alloc::string::String,
    /// redirect uri to use when initiating OAuth2 authorization requests.
    #[prost(string, tag="2")]
    pub redirect_uri: ::prost::alloc::string::String,
    /// scopes to request when initiating OAuth2 authorization requests.
    #[prost(string, repeated, tag="3")]
    pub scopes: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// Authorization Header to use when passing Access Tokens to the server. If not provided, the client should use the
    /// default http `Authorization` header.
    #[prost(string, tag="4")]
    pub authorization_metadata_key: ::prost::alloc::string::String,
    /// ServiceHttpEndpoint points to the http endpoint for the backend. If empty, clients can assume the endpoint used
    /// to configure the gRPC connection can be used for the http one respecting the insecure flag to choose between
    /// SSL or no SSL connections.
    #[prost(string, tag="5")]
    pub service_http_endpoint: ::prost::alloc::string::String,
    /// audience to use when initiating OAuth2 authorization requests.
    #[prost(string, tag="6")]
    pub audience: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateUploadLocationResponse {
    /// SignedUrl specifies the url to use to upload content to (e.g. <https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...>)
    #[prost(string, tag="1")]
    pub signed_url: ::prost::alloc::string::String,
    /// NativeUrl specifies the url in the format of the configured storage provider (e.g. s3://my-bucket/randomstring/suffix.tar)
    #[prost(string, tag="2")]
    pub native_url: ::prost::alloc::string::String,
    /// ExpiresAt defines when will the signed URL expires.
    #[prost(message, optional, tag="3")]
    pub expires_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// CreateUploadLocationRequest specified request for the CreateUploadLocation API.
/// The implementation in data proxy service will create the s3 location with some server side configured prefixes,
/// and then:
///    - project/domain/(a deterministic str representation of the content_md5)/filename (if present); OR
///    - project/domain/filename_root (if present)/filename (if present).
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateUploadLocationRequest {
    /// Project to create the upload location for
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Domain to create the upload location for.
    /// +required
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// Filename specifies a desired suffix for the generated location. E.g. `file.py` or `pre/fix/file.zip`.
    /// +optional. By default, the service will generate a consistent name based on the provided parameters.
    #[prost(string, tag="3")]
    pub filename: ::prost::alloc::string::String,
    /// ExpiresIn defines a requested expiration duration for the generated url. The request will be rejected if this
    /// exceeds the platform allowed max.
    /// +optional. The default value comes from a global config.
    #[prost(message, optional, tag="4")]
    pub expires_in: ::core::option::Option<::prost_types::Duration>,
    /// ContentMD5 restricts the upload location to the specific MD5 provided. The ContentMD5 will also appear in the
    /// generated path.
    /// +required
    #[prost(bytes="vec", tag="5")]
    pub content_md5: ::prost::alloc::vec::Vec<u8>,
    /// If present, data proxy will use this string in lieu of the md5 hash in the path. When the filename is also included
    /// this makes the upload location deterministic. The native url will still be prefixed by the upload location prefix
    /// in data proxy config. This option is useful when uploading multiple files.
    /// +optional
    #[prost(string, tag="6")]
    pub filename_root: ::prost::alloc::string::String,
}
/// CreateDownloadLocationRequest specified request for the CreateDownloadLocation API.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateDownloadLocationRequest {
    /// NativeUrl specifies the url in the format of the configured storage provider (e.g. s3://my-bucket/randomstring/suffix.tar)
    #[prost(string, tag="1")]
    pub native_url: ::prost::alloc::string::String,
    /// ExpiresIn defines a requested expiration duration for the generated url. The request will be rejected if this
    /// exceeds the platform allowed max.
    /// +optional. The default value comes from a global config.
    #[prost(message, optional, tag="2")]
    pub expires_in: ::core::option::Option<::prost_types::Duration>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateDownloadLocationResponse {
    /// SignedUrl specifies the url to use to download content from (e.g. <https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...>)
    #[prost(string, tag="1")]
    pub signed_url: ::prost::alloc::string::String,
    /// ExpiresAt defines when will the signed URL expires.
    #[prost(message, optional, tag="2")]
    pub expires_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// CreateDownloadLinkRequest defines the request parameters to create a download link (signed url)
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateDownloadLinkRequest {
    /// ArtifactType of the artifact requested.
    #[prost(enumeration="ArtifactType", tag="1")]
    pub artifact_type: i32,
    /// ExpiresIn defines a requested expiration duration for the generated url. The request will be rejected if this
    /// exceeds the platform allowed max.
    /// +optional. The default value comes from a global config.
    #[prost(message, optional, tag="2")]
    pub expires_in: ::core::option::Option<::prost_types::Duration>,
    #[prost(oneof="create_download_link_request::Source", tags="3")]
    pub source: ::core::option::Option<create_download_link_request::Source>,
}
/// Nested message and enum types in `CreateDownloadLinkRequest`.
pub mod create_download_link_request {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Source {
        /// NodeId is the unique identifier for the node execution. For a task node, this will retrieve the output of the
        /// most recent attempt of the task.
        #[prost(message, tag="3")]
        NodeExecutionId(super::super::core::NodeExecutionIdentifier),
    }
}
/// CreateDownloadLinkResponse defines the response for the generated links
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateDownloadLinkResponse {
    /// SignedUrl specifies the url to use to download content from (e.g. <https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...>)
    #[deprecated]
    #[prost(string, repeated, tag="1")]
    pub signed_url: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// ExpiresAt defines when will the signed URL expire.
    #[deprecated]
    #[prost(message, optional, tag="2")]
    pub expires_at: ::core::option::Option<::prost_types::Timestamp>,
    /// New wrapper object containing the signed urls and expiration time
    #[prost(message, optional, tag="3")]
    pub pre_signed_urls: ::core::option::Option<PreSignedUrLs>,
}
/// Wrapper object since the message is shared across this and the GetDataResponse
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PreSignedUrLs {
    /// SignedUrl specifies the url to use to download content from (e.g. <https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...>)
    #[prost(string, repeated, tag="1")]
    pub signed_url: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// ExpiresAt defines when will the signed URL expire.
    #[prost(message, optional, tag="2")]
    pub expires_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// General request artifact to retrieve data from a Flyte artifact url.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDataRequest {
    /// A unique identifier in the form of flyte://<something> that uniquely, for a given Flyte
    /// backend, identifies a Flyte artifact (\[i\]nput, \[o\]utput, flyte \[d\]eck, etc.).
    /// e.g. flyte://v1/proj/development/execid/n2/0/i (for 0th task execution attempt input)
    ///       flyte://v1/proj/development/execid/n2/i (for node execution input)
    ///       flyte://v1/proj/development/execid/n2/o/o3 (the o3 output of the second node)
    #[prost(string, tag="1")]
    pub flyte_url: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDataResponse {
    #[prost(oneof="get_data_response::Data", tags="1, 2, 3")]
    pub data: ::core::option::Option<get_data_response::Data>,
}
/// Nested message and enum types in `GetDataResponse`.
pub mod get_data_response {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Data {
        /// literal map data will be returned
        #[prost(message, tag="1")]
        LiteralMap(super::super::core::LiteralMap),
        /// Flyte deck html will be returned as a signed url users can download
        #[prost(message, tag="2")]
        PreSignedUrls(super::PreSignedUrLs),
        /// Single literal will be returned. This is returned when the user/url requests a specific output or input
        /// by name. See the o3 example above.
        #[prost(message, tag="3")]
        Literal(super::super::core::Literal),
    }
}
/// ArtifactType
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum ArtifactType {
    /// ARTIFACT_TYPE_UNDEFINED is the default, often invalid, value for the enum.
    Undefined = 0,
    /// ARTIFACT_TYPE_DECK refers to the deck html file optionally generated after a task, a workflow or a launch plan
    /// finishes executing.
    Deck = 1,
}
impl ArtifactType {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            ArtifactType::Undefined => "ARTIFACT_TYPE_UNDEFINED",
            ArtifactType::Deck => "ARTIFACT_TYPE_DECK",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "ARTIFACT_TYPE_UNDEFINED" => Some(Self::Undefined),
            "ARTIFACT_TYPE_DECK" => Some(Self::Deck),
            _ => None,
        }
    }
}
/// Represents a request structure to create task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskCreateRequest {
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
}
/// Represents a create response structure.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskCreateResponse {
    #[prost(string, tag="1")]
    pub job_id: ::prost::alloc::string::String,
}
/// A message used to fetch a job state from backend plugin server.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskGetRequest {
    /// A predefined yet extensible Task type identifier.
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// The unique id identifying the job.
    #[prost(string, tag="2")]
    pub job_id: ::prost::alloc::string::String,
}
/// Response to get an individual task state.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskGetResponse {
    /// The state of the execution is used to control its visibility in the UI/CLI.
    #[prost(enumeration="State", tag="1")]
    pub state: i32,
    /// The outputs of the execution. It's typically used by sql task. Flyteplugins service will create a
    /// Structured dataset pointing to the query result table.
    /// +optional
    #[prost(message, optional, tag="2")]
    pub outputs: ::core::option::Option<super::core::LiteralMap>,
}
/// A message used to delete a task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskDeleteRequest {
    /// A predefined yet extensible Task type identifier.
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// The unique id identifying the job.
    #[prost(string, tag="2")]
    pub job_id: ::prost::alloc::string::String,
}
/// Response to delete a task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskDeleteResponse {
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
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UserInfoRequest {
}
/// See the OpenID Connect spec at <https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse> for more information.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UserInfoResponse {
    /// Locally unique and never reassigned identifier within the Issuer for the End-User, which is intended to be consumed
    /// by the Client.
    #[prost(string, tag="1")]
    pub subject: ::prost::alloc::string::String,
    /// Full name
    #[prost(string, tag="2")]
    pub name: ::prost::alloc::string::String,
    /// Shorthand name by which the End-User wishes to be referred to
    #[prost(string, tag="3")]
    pub preferred_username: ::prost::alloc::string::String,
    /// Given name(s) or first name(s)
    #[prost(string, tag="4")]
    pub given_name: ::prost::alloc::string::String,
    /// Surname(s) or last name(s)
    #[prost(string, tag="5")]
    pub family_name: ::prost::alloc::string::String,
    /// Preferred e-mail address
    #[prost(string, tag="6")]
    pub email: ::prost::alloc::string::String,
    /// Profile picture URL
    #[prost(string, tag="7")]
    pub picture: ::prost::alloc::string::String,
    /// Additional claims
    #[prost(message, optional, tag="8")]
    pub additional_claims: ::core::option::Option<::prost_types::Struct>,
}
// @@protoc_insertion_point(module)
