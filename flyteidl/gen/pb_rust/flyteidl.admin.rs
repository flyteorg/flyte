// @generated
/// Represents a subset of runtime task execution metadata that are relevant to external plugins.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionMetadata {
    /// ID of the task execution
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
    /// Metadata is created by the agent. It could be a string (jobId) or a dict (more complex metadata).
    #[prost(bytes="vec", tag="1")]
    pub resource_meta: ::prost::alloc::vec::Vec<u8>,
}
/// A message used to fetch a job resource from flyte agent server.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTaskRequest {
    /// A predefined yet extensible Task type identifier.
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// Metadata about the resource to be pass to the agent.
    #[prost(bytes="vec", tag="2")]
    pub resource_meta: ::prost::alloc::vec::Vec<u8>,
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
    /// The state of the execution is used to control its visibility in the UI/CLI.
    #[prost(enumeration="State", tag="1")]
    pub state: i32,
    /// The outputs of the execution. It's typically used by sql task. Agent service will create a
    /// Structured dataset pointing to the query result table.
    /// +optional
    #[prost(message, optional, tag="2")]
    pub outputs: ::core::option::Option<super::core::LiteralMap>,
}
/// A message used to delete a task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeleteTaskRequest {
    /// A predefined yet extensible Task type identifier.
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// Metadata about the resource to be pass to the agent.
    #[prost(bytes="vec", tag="2")]
    pub resource_meta: ::prost::alloc::vec::Vec<u8>,
}
/// Response to delete a task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeleteTaskResponse {
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
/// Encapsulates specifications for routing an execution onto a specific cluster.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ClusterAssignment {
    #[prost(string, tag="3")]
    pub cluster_pool_name: ::prost::alloc::string::String,
}
/// Encapsulation of fields that identifies a Flyte resource.
/// A Flyte resource can be a task, workflow or launch plan.
/// A resource can internally have multiple versions and is uniquely identified
/// by project, domain, and name.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityIdentifier {
    /// Name of the project the resource belongs to.
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Name of the domain the resource belongs to.
    /// A domain can be considered as a subset within a specific project.
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// User provided value for the resource.
    /// The combination of project + domain + name uniquely identifies the resource.
    /// +optional - in certain contexts - like 'List API', 'Launch plans'
    #[prost(string, tag="3")]
    pub name: ::prost::alloc::string::String,
}
/// Additional metadata around a named entity.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityMetadata {
    /// Common description across all versions of the entity
    /// +optional
    #[prost(string, tag="1")]
    pub description: ::prost::alloc::string::String,
    /// Shared state across all version of the entity
    /// At this point in time, only workflow entities can have their state archived.
    #[prost(enumeration="NamedEntityState", tag="2")]
    pub state: i32,
}
/// Encapsulates information common to a NamedEntity, a Flyte resource such as a task,
/// workflow or launch plan. A NamedEntity is exclusively identified by its resource type
/// and identifier.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntity {
    /// Resource type of the named entity. One of Task, Workflow or LaunchPlan.
    #[prost(enumeration="super::core::ResourceType", tag="1")]
    pub resource_type: i32,
    #[prost(message, optional, tag="2")]
    pub id: ::core::option::Option<NamedEntityIdentifier>,
    /// Additional metadata around a named entity.
    #[prost(message, optional, tag="3")]
    pub metadata: ::core::option::Option<NamedEntityMetadata>,
}
/// Specifies sort ordering in a list request.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Sort {
    /// Indicates an attribute to sort the response values.
    /// +required
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    /// Indicates the direction to apply sort key for response values.
    /// +optional
    #[prost(enumeration="sort::Direction", tag="2")]
    pub direction: i32,
}
/// Nested message and enum types in `Sort`.
pub mod sort {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Direction {
        /// By default, fields are sorted in descending order.
        Descending = 0,
        Ascending = 1,
    }
    impl Direction {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Direction::Descending => "DESCENDING",
                Direction::Ascending => "ASCENDING",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "DESCENDING" => Some(Self::Descending),
                "ASCENDING" => Some(Self::Ascending),
                _ => None,
            }
        }
    }
}
/// Represents a request structure to list NamedEntityIdentifiers.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityIdentifierListRequest {
    /// Name of the project that contains the identifiers.
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Name of the domain the identifiers belongs to within the project.
    /// +required
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// Indicates the number of resources to be returned.
    /// +required
    #[prost(uint32, tag="3")]
    pub limit: u32,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="4")]
    pub token: ::prost::alloc::string::String,
    /// Specifies how listed entities should be sorted in the response.
    /// +optional
    #[prost(message, optional, tag="5")]
    pub sort_by: ::core::option::Option<Sort>,
    /// Indicates a list of filters passed as string.
    /// +optional
    #[prost(string, tag="6")]
    pub filters: ::prost::alloc::string::String,
}
/// Represents a request structure to list NamedEntity objects
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityListRequest {
    /// Resource type of the metadata to query. One of Task, Workflow or LaunchPlan.
    /// +required
    #[prost(enumeration="super::core::ResourceType", tag="1")]
    pub resource_type: i32,
    /// Name of the project that contains the identifiers.
    /// +required
    #[prost(string, tag="2")]
    pub project: ::prost::alloc::string::String,
    /// Name of the domain the identifiers belongs to within the project.
    #[prost(string, tag="3")]
    pub domain: ::prost::alloc::string::String,
    /// Indicates the number of resources to be returned.
    #[prost(uint32, tag="4")]
    pub limit: u32,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="5")]
    pub token: ::prost::alloc::string::String,
    /// Specifies how listed entities should be sorted in the response.
    /// +optional
    #[prost(message, optional, tag="6")]
    pub sort_by: ::core::option::Option<Sort>,
    /// Indicates a list of filters passed as string.
    /// +optional
    #[prost(string, tag="7")]
    pub filters: ::prost::alloc::string::String,
}
/// Represents a list of NamedEntityIdentifiers.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityIdentifierList {
    /// A list of identifiers.
    #[prost(message, repeated, tag="1")]
    pub entities: ::prost::alloc::vec::Vec<NamedEntityIdentifier>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Represents a list of NamedEntityIdentifiers.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityList {
    /// A list of NamedEntity objects
    #[prost(message, repeated, tag="1")]
    pub entities: ::prost::alloc::vec::Vec<NamedEntity>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// A request to retrieve the metadata associated with a NamedEntityIdentifier
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityGetRequest {
    /// Resource type of the metadata to get. One of Task, Workflow or LaunchPlan.
    /// +required
    #[prost(enumeration="super::core::ResourceType", tag="1")]
    pub resource_type: i32,
    /// The identifier for the named entity for which to fetch metadata.
    /// +required
    #[prost(message, optional, tag="2")]
    pub id: ::core::option::Option<NamedEntityIdentifier>,
}
/// Request to set the referenced named entity state to the configured value.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityUpdateRequest {
    /// Resource type of the metadata to update
    /// +required
    #[prost(enumeration="super::core::ResourceType", tag="1")]
    pub resource_type: i32,
    /// Identifier of the metadata to update
    /// +required
    #[prost(message, optional, tag="2")]
    pub id: ::core::option::Option<NamedEntityIdentifier>,
    /// Metadata object to set as the new value
    /// +required
    #[prost(message, optional, tag="3")]
    pub metadata: ::core::option::Option<NamedEntityMetadata>,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NamedEntityUpdateResponse {
}
/// Shared request structure to fetch a single resource.
/// Resources include: Task, Workflow, LaunchPlan
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ObjectGetRequest {
    /// Indicates a unique version of resource.
    /// +required
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
}
/// Shared request structure to retrieve a list of resources.
/// Resources include: Task, Workflow, LaunchPlan
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ResourceListRequest {
    /// id represents the unique identifier of the resource.
    /// +required
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<NamedEntityIdentifier>,
    /// Indicates the number of resources to be returned.
    /// +required
    #[prost(uint32, tag="2")]
    pub limit: u32,
    /// In the case of multiple pages of results, this server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="3")]
    pub token: ::prost::alloc::string::String,
    /// Indicates a list of filters passed as string.
    /// More info on constructing filters : <Link>
    /// +optional
    #[prost(string, tag="4")]
    pub filters: ::prost::alloc::string::String,
    /// Sort ordering.
    /// +optional
    #[prost(message, optional, tag="5")]
    pub sort_by: ::core::option::Option<Sort>,
}
/// Defines an email notification specification.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EmailNotification {
    /// The list of email addresses recipients for this notification.
    /// +required
    #[prost(string, repeated, tag="1")]
    pub recipients_email: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// Defines a pager duty notification specification.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PagerDutyNotification {
    /// Currently, PagerDuty notifications leverage email to trigger a notification.
    /// +required
    #[prost(string, repeated, tag="1")]
    pub recipients_email: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// Defines a slack notification specification.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SlackNotification {
    /// Currently, Slack notifications leverage email to trigger a notification.
    /// +required
    #[prost(string, repeated, tag="1")]
    pub recipients_email: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// Represents a structure for notifications based on execution status.
/// The notification content is configured within flyte admin but can be templatized.
/// Future iterations could expose configuring notifications with custom content.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Notification {
    /// A list of phases to which users can associate the notifications to.
    /// +required
    #[prost(enumeration="super::core::workflow_execution::Phase", repeated, tag="1")]
    pub phases: ::prost::alloc::vec::Vec<i32>,
    /// The type of notification to trigger.
    /// +required
    #[prost(oneof="notification::Type", tags="2, 3, 4")]
    pub r#type: ::core::option::Option<notification::Type>,
}
/// Nested message and enum types in `Notification`.
pub mod notification {
    /// The type of notification to trigger.
    /// +required
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Type {
        #[prost(message, tag="2")]
        Email(super::EmailNotification),
        #[prost(message, tag="3")]
        PagerDuty(super::PagerDutyNotification),
        #[prost(message, tag="4")]
        Slack(super::SlackNotification),
    }
}
/// Represents a string url and associated metadata used throughout the platform.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UrlBlob {
    /// Actual url value.
    #[prost(string, tag="1")]
    pub url: ::prost::alloc::string::String,
    /// Represents the size of the file accessible at the above url.
    #[prost(int64, tag="2")]
    pub bytes: i64,
}
/// Label values to be applied to an execution resource.
/// In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
/// to specify how to merge labels defined at registration and execution time.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Labels {
    /// Map of custom labels to be applied to the execution resource.
    #[prost(map="string, string", tag="1")]
    pub values: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
/// Annotation values to be applied to an execution resource.
/// In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
/// to specify how to merge annotations defined at registration and execution time.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Annotations {
    /// Map of custom annotations to be applied to the execution resource.
    #[prost(map="string, string", tag="1")]
    pub values: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
/// Environment variable values to be applied to an execution resource.
/// In the future a mode (e.g. OVERRIDE, APPEND, etc) can be defined
/// to specify how to merge environment variables defined at registration and execution time.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Envs {
    /// Map of custom environment variables to be applied to the execution resource.
    #[prost(message, repeated, tag="1")]
    pub values: ::prost::alloc::vec::Vec<super::core::KeyValuePair>,
}
/// Defines permissions associated with executions created by this launch plan spec.
/// Use either of these roles when they have permissions required by your workflow execution.
/// Deprecated.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AuthRole {
    /// Defines an optional iam role which will be used for tasks run in executions created with this launch plan.
    #[prost(string, tag="1")]
    pub assumable_iam_role: ::prost::alloc::string::String,
    /// Defines an optional kubernetes service account which will be used for tasks run in executions created with this launch plan.
    #[prost(string, tag="2")]
    pub kubernetes_service_account: ::prost::alloc::string::String,
}
/// Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).
/// See <https://github.com/flyteorg/flyte/issues/211> for more background information.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RawOutputDataConfig {
    /// Prefix for where offloaded data from user workflows will be written
    /// e.g. s3://bucket/key or s3://bucket/
    #[prost(string, tag="1")]
    pub output_location_prefix: ::prost::alloc::string::String,
}
/// These URLs are returned as part of node and task execution data requests.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FlyteUrLs {
    #[prost(string, tag="1")]
    pub inputs: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub outputs: ::prost::alloc::string::String,
    #[prost(string, tag="3")]
    pub deck: ::prost::alloc::string::String,
}
/// The status of the named entity is used to control its visibility in the UI.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum NamedEntityState {
    /// By default, all named entities are considered active and under development.
    NamedEntityActive = 0,
    /// Archived named entities are no longer visible in the UI.
    NamedEntityArchived = 1,
    /// System generated entities that aren't explicitly created or managed by a user.
    SystemGenerated = 2,
}
impl NamedEntityState {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            NamedEntityState::NamedEntityActive => "NAMED_ENTITY_ACTIVE",
            NamedEntityState::NamedEntityArchived => "NAMED_ENTITY_ARCHIVED",
            NamedEntityState::SystemGenerated => "SYSTEM_GENERATED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "NAMED_ENTITY_ACTIVE" => Some(Self::NamedEntityActive),
            "NAMED_ENTITY_ARCHIVED" => Some(Self::NamedEntityArchived),
            "SYSTEM_GENERATED" => Some(Self::SystemGenerated),
            _ => None,
        }
    }
}
/// DescriptionEntity contains detailed description for the task/workflow.
/// Documentation could provide insight into the algorithms, business use case, etc.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DescriptionEntity {
    /// id represents the unique identifier of the description entity.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// One-liner overview of the entity.
    #[prost(string, tag="2")]
    pub short_description: ::prost::alloc::string::String,
    /// Full user description with formatting preserved.
    #[prost(message, optional, tag="3")]
    pub long_description: ::core::option::Option<Description>,
    /// Optional link to source code used to define this entity.
    #[prost(message, optional, tag="4")]
    pub source_code: ::core::option::Option<SourceCode>,
    /// User-specified tags. These are arbitrary and can be used for searching
    /// filtering and discovering tasks.
    #[prost(string, repeated, tag="5")]
    pub tags: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// Full user description with formatting preserved. This can be rendered
/// by clients, such as the console or command line tools with in-tact
/// formatting.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Description {
    /// Format of the long description
    #[prost(enumeration="DescriptionFormat", tag="3")]
    pub format: i32,
    /// Optional link to an icon for the entity
    #[prost(string, tag="4")]
    pub icon_link: ::prost::alloc::string::String,
    #[prost(oneof="description::Content", tags="1, 2")]
    pub content: ::core::option::Option<description::Content>,
}
/// Nested message and enum types in `Description`.
pub mod description {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Content {
        /// long description - no more than 4KB
        #[prost(string, tag="1")]
        Value(::prost::alloc::string::String),
        /// if the description sizes exceed some threshold we can offload the entire
        /// description proto altogether to an external data store, like S3 rather than store inline in the db
        #[prost(string, tag="2")]
        Uri(::prost::alloc::string::String),
    }
}
/// Link to source code used to define this entity
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SourceCode {
    #[prost(string, tag="1")]
    pub link: ::prost::alloc::string::String,
}
/// Represents a list of DescriptionEntities returned from the admin.
/// See :ref:`ref_flyteidl.admin.DescriptionEntity` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DescriptionEntityList {
    /// A list of DescriptionEntities returned based on the request.
    #[prost(message, repeated, tag="1")]
    pub description_entities: ::prost::alloc::vec::Vec<DescriptionEntity>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Represents a request structure to retrieve a list of DescriptionEntities.
/// See :ref:`ref_flyteidl.admin.DescriptionEntity` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DescriptionEntityListRequest {
    /// Identifies the specific type of resource that this identifier corresponds to.
    #[prost(enumeration="super::core::ResourceType", tag="1")]
    pub resource_type: i32,
    /// The identifier for the description entity.
    /// +required
    #[prost(message, optional, tag="2")]
    pub id: ::core::option::Option<NamedEntityIdentifier>,
    /// Indicates the number of resources to be returned.
    /// +required
    #[prost(uint32, tag="3")]
    pub limit: u32,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="4")]
    pub token: ::prost::alloc::string::String,
    /// Indicates a list of filters passed as string.
    /// More info on constructing filters : <Link>
    /// +optional
    #[prost(string, tag="5")]
    pub filters: ::prost::alloc::string::String,
    /// Sort ordering for returned list.
    /// +optional
    #[prost(message, optional, tag="6")]
    pub sort_by: ::core::option::Option<Sort>,
}
/// The format of the long description
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum DescriptionFormat {
    Unknown = 0,
    Markdown = 1,
    Html = 2,
    /// python default documentation - comments is rst
    Rst = 3,
}
impl DescriptionFormat {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            DescriptionFormat::Unknown => "DESCRIPTION_FORMAT_UNKNOWN",
            DescriptionFormat::Markdown => "DESCRIPTION_FORMAT_MARKDOWN",
            DescriptionFormat::Html => "DESCRIPTION_FORMAT_HTML",
            DescriptionFormat::Rst => "DESCRIPTION_FORMAT_RST",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "DESCRIPTION_FORMAT_UNKNOWN" => Some(Self::Unknown),
            "DESCRIPTION_FORMAT_MARKDOWN" => Some(Self::Markdown),
            "DESCRIPTION_FORMAT_HTML" => Some(Self::Html),
            "DESCRIPTION_FORMAT_RST" => Some(Self::Rst),
            _ => None,
        }
    }
}
/// Indicates that a sent event was not used to update execution state due to
/// the referenced execution already being terminated (and therefore ineligible
/// for further state transitions).
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EventErrorAlreadyInTerminalState {
    /// +required
    #[prost(string, tag="1")]
    pub current_phase: ::prost::alloc::string::String,
}
/// Indicates an event was rejected because it came from a different cluster than 
/// is on record as running the execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EventErrorIncompatibleCluster {
    /// The cluster which has been recorded as processing the execution.
    /// +required
    #[prost(string, tag="1")]
    pub cluster: ::prost::alloc::string::String,
}
/// Indicates why a sent event was not used to update execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EventFailureReason {
    /// +required
    #[prost(oneof="event_failure_reason::Reason", tags="1, 2")]
    pub reason: ::core::option::Option<event_failure_reason::Reason>,
}
/// Nested message and enum types in `EventFailureReason`.
pub mod event_failure_reason {
    /// +required
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Reason {
        #[prost(message, tag="1")]
        AlreadyInTerminalState(super::EventErrorAlreadyInTerminalState),
        #[prost(message, tag="2")]
        IncompatibleCluster(super::EventErrorIncompatibleCluster),
    }
}
/// Request to send a notification that a workflow execution event has occurred.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionEventRequest {
    /// Unique ID for this request that can be traced between services
    #[prost(string, tag="1")]
    pub request_id: ::prost::alloc::string::String,
    /// Details about the event that occurred.
    #[prost(message, optional, tag="2")]
    pub event: ::core::option::Option<super::event::WorkflowExecutionEvent>,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionEventResponse {
}
/// Request to send a notification that a node execution event has occurred.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionEventRequest {
    /// Unique ID for this request that can be traced between services
    #[prost(string, tag="1")]
    pub request_id: ::prost::alloc::string::String,
    /// Details about the event that occurred.
    #[prost(message, optional, tag="2")]
    pub event: ::core::option::Option<super::event::NodeExecutionEvent>,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionEventResponse {
}
/// Request to send a notification that a task execution event has occurred.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionEventRequest {
    /// Unique ID for this request that can be traced between services
    #[prost(string, tag="1")]
    pub request_id: ::prost::alloc::string::String,
    /// Details about the event that occurred.
    #[prost(message, optional, tag="2")]
    pub event: ::core::option::Option<super::event::TaskExecutionEvent>,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionEventResponse {
}
/// Request to launch an execution with the given project, domain and optionally-assigned name.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionCreateRequest {
    /// Name of the project the execution belongs to.
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Name of the domain the execution belongs to.
    /// A domain can be considered as a subset within a specific project.
    /// +required
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// User provided value for the resource.
    /// If none is provided the system will generate a unique string.
    /// +optional
    #[prost(string, tag="3")]
    pub name: ::prost::alloc::string::String,
    /// Additional fields necessary to launch the execution.
    /// +optional
    #[prost(message, optional, tag="4")]
    pub spec: ::core::option::Option<ExecutionSpec>,
    /// The inputs required to start the execution. All required inputs must be
    /// included in this map. If not required and not provided, defaults apply.
    /// +optional
    #[prost(message, optional, tag="5")]
    pub inputs: ::core::option::Option<super::core::LiteralMap>,
}
/// Request to relaunch the referenced execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionRelaunchRequest {
    /// Identifier of the workflow execution to relaunch.
    /// +required
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// User provided value for the relaunched execution.
    /// If none is provided the system will generate a unique string.
    /// +optional
    #[prost(string, tag="3")]
    pub name: ::prost::alloc::string::String,
    /// Allows for all cached values of a workflow and its tasks to be overwritten for a single execution.
    /// If enabled, all calculations are performed even if cached results would be available, overwriting the stored
    /// data once execution finishes successfully.
    #[prost(bool, tag="4")]
    pub overwrite_cache: bool,
}
/// Request to recover the referenced execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionRecoverRequest {
    /// Identifier of the workflow execution to recover.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// User provided value for the recovered execution.
    /// If none is provided the system will generate a unique string.
    /// +optional
    #[prost(string, tag="2")]
    pub name: ::prost::alloc::string::String,
    /// Additional metadata which will be used to overwrite any metadata in the reference execution when triggering a recovery execution.
    #[prost(message, optional, tag="3")]
    pub metadata: ::core::option::Option<ExecutionMetadata>,
}
/// The unique identifier for a successfully created execution.
/// If the name was *not* specified in the create request, this identifier will include a generated name.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionCreateResponse {
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
}
/// A message used to fetch a single workflow execution entity.
/// See :ref:`ref_flyteidl.admin.Execution` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionGetRequest {
    /// Uniquely identifies an individual workflow execution.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
}
/// A workflow execution represents an instantiated workflow, including all inputs and additional
/// metadata as well as computed results included state, outputs, and duration-based attributes.
/// Used as a response object used in Get and List execution requests.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Execution {
    /// Unique identifier of the workflow execution.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// User-provided configuration and inputs for launching the execution.
    #[prost(message, optional, tag="2")]
    pub spec: ::core::option::Option<ExecutionSpec>,
    /// Execution results.
    #[prost(message, optional, tag="3")]
    pub closure: ::core::option::Option<ExecutionClosure>,
}
/// Used as a response for request to list executions.
/// See :ref:`ref_flyteidl.admin.Execution` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionList {
    #[prost(message, repeated, tag="1")]
    pub executions: ::prost::alloc::vec::Vec<Execution>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Input/output data can represented by actual values or a link to where values are stored
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LiteralMapBlob {
    #[prost(oneof="literal_map_blob::Data", tags="1, 2")]
    pub data: ::core::option::Option<literal_map_blob::Data>,
}
/// Nested message and enum types in `LiteralMapBlob`.
pub mod literal_map_blob {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Data {
        /// Data in LiteralMap format
        #[prost(message, tag="1")]
        Values(super::super::core::LiteralMap),
        /// In the event that the map is too large, we return a uri to the data
        #[prost(string, tag="2")]
        Uri(::prost::alloc::string::String),
    }
}
/// Specifies metadata around an aborted workflow execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AbortMetadata {
    /// In the case of a user-specified abort, this will pass along the user-supplied cause.
    #[prost(string, tag="1")]
    pub cause: ::prost::alloc::string::String,
    /// Identifies the entity (if any) responsible for terminating the execution
    #[prost(string, tag="2")]
    pub principal: ::prost::alloc::string::String,
}
/// Encapsulates the results of the Execution
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionClosure {
    /// Inputs computed and passed for execution.
    /// computed_inputs depends on inputs in ExecutionSpec, fixed and default inputs in launch plan
    #[deprecated]
    #[prost(message, optional, tag="3")]
    pub computed_inputs: ::core::option::Option<super::core::LiteralMap>,
    /// Most recent recorded phase for the execution.
    #[prost(enumeration="super::core::workflow_execution::Phase", tag="4")]
    pub phase: i32,
    /// Reported time at which the execution began running.
    #[prost(message, optional, tag="5")]
    pub started_at: ::core::option::Option<::prost_types::Timestamp>,
    /// The amount of time the execution spent running.
    #[prost(message, optional, tag="6")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    /// Reported time at which the execution was created.
    #[prost(message, optional, tag="7")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Reported time at which the execution was last updated.
    #[prost(message, optional, tag="8")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
    /// The notification settings to use after merging the CreateExecutionRequest and the launch plan
    /// notification settings. An execution launched with notifications will always prefer that definition
    /// to notifications defined statically in a launch plan.
    #[prost(message, repeated, tag="9")]
    pub notifications: ::prost::alloc::vec::Vec<Notification>,
    /// Identifies the workflow definition for this execution.
    #[prost(message, optional, tag="11")]
    pub workflow_id: ::core::option::Option<super::core::Identifier>,
    /// Provides the details of the last stage change
    #[prost(message, optional, tag="14")]
    pub state_change_details: ::core::option::Option<ExecutionStateChangeDetails>,
    /// A result produced by a terminated execution.
    /// A pending (non-terminal) execution will not have any output result.
    #[prost(oneof="execution_closure::OutputResult", tags="1, 2, 10, 12, 13")]
    pub output_result: ::core::option::Option<execution_closure::OutputResult>,
}
/// Nested message and enum types in `ExecutionClosure`.
pub mod execution_closure {
    /// A result produced by a terminated execution.
    /// A pending (non-terminal) execution will not have any output result.
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum OutputResult {
        /// Output URI in the case of a successful execution.
        /// DEPRECATED. Use GetExecutionData to fetch output data instead.
        #[prost(message, tag="1")]
        Outputs(super::LiteralMapBlob),
        /// Error information in the case of a failed execution.
        #[prost(message, tag="2")]
        Error(super::super::core::ExecutionError),
        /// In the case of a user-specified abort, this will pass along the user-supplied cause.
        #[prost(string, tag="10")]
        AbortCause(::prost::alloc::string::String),
        /// In the case of a user-specified abort, this will pass along the user and their supplied cause.
        #[prost(message, tag="12")]
        AbortMetadata(super::AbortMetadata),
        /// Raw output data produced by this execution.
        /// DEPRECATED. Use GetExecutionData to fetch output data instead.
        #[prost(message, tag="13")]
        OutputData(super::super::core::LiteralMap),
    }
}
/// Represents system, rather than user-facing, metadata about an execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SystemMetadata {
    /// Which execution cluster this execution ran on.
    #[prost(string, tag="1")]
    pub execution_cluster: ::prost::alloc::string::String,
    /// Which kubernetes namespace the execution ran under.
    #[prost(string, tag="2")]
    pub namespace: ::prost::alloc::string::String,
}
/// Represents attributes about an execution which are not required to launch the execution but are useful to record.
/// These attributes are assigned at launch time and do not change.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionMetadata {
    #[prost(enumeration="execution_metadata::ExecutionMode", tag="1")]
    pub mode: i32,
    /// Identifier of the entity that triggered this execution.
    /// For systems using back-end authentication any value set here will be discarded in favor of the
    /// authenticated user context.
    #[prost(string, tag="2")]
    pub principal: ::prost::alloc::string::String,
    /// Indicates the nestedness of this execution.
    /// If a user launches a workflow execution, the default nesting is 0.
    /// If this execution further launches a workflow (child workflow), the nesting level is incremented by 0 => 1
    /// Generally, if workflow at nesting level k launches a workflow then the child workflow will have
    /// nesting = k + 1.
    #[prost(uint32, tag="3")]
    pub nesting: u32,
    /// For scheduled executions, the requested time for execution for this specific schedule invocation.
    #[prost(message, optional, tag="4")]
    pub scheduled_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Which subworkflow node (if any) launched this execution
    #[prost(message, optional, tag="5")]
    pub parent_node_execution: ::core::option::Option<super::core::NodeExecutionIdentifier>,
    /// Optional, a reference workflow execution related to this execution.
    /// In the case of a relaunch, this references the original workflow execution.
    #[prost(message, optional, tag="16")]
    pub reference_execution: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// Optional, platform-specific metadata about the execution.
    /// In this the future this may be gated behind an ACL or some sort of authorization.
    #[prost(message, optional, tag="17")]
    pub system_metadata: ::core::option::Option<SystemMetadata>,
}
/// Nested message and enum types in `ExecutionMetadata`.
pub mod execution_metadata {
    /// The method by which this execution was launched.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum ExecutionMode {
        /// The default execution mode, MANUAL implies that an execution was launched by an individual.
        Manual = 0,
        /// A schedule triggered this execution launch.
        Scheduled = 1,
        /// A system process was responsible for launching this execution rather an individual.
        System = 2,
        /// This execution was launched with identical inputs as a previous execution.
        Relaunch = 3,
        /// This execution was triggered by another execution.
        ChildWorkflow = 4,
        /// This execution was recovered from another execution.
        Recovered = 5,
    }
    impl ExecutionMode {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                ExecutionMode::Manual => "MANUAL",
                ExecutionMode::Scheduled => "SCHEDULED",
                ExecutionMode::System => "SYSTEM",
                ExecutionMode::Relaunch => "RELAUNCH",
                ExecutionMode::ChildWorkflow => "CHILD_WORKFLOW",
                ExecutionMode::Recovered => "RECOVERED",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "MANUAL" => Some(Self::Manual),
                "SCHEDULED" => Some(Self::Scheduled),
                "SYSTEM" => Some(Self::System),
                "RELAUNCH" => Some(Self::Relaunch),
                "CHILD_WORKFLOW" => Some(Self::ChildWorkflow),
                "RECOVERED" => Some(Self::Recovered),
                _ => None,
            }
        }
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NotificationList {
    #[prost(message, repeated, tag="1")]
    pub notifications: ::prost::alloc::vec::Vec<Notification>,
}
/// An ExecutionSpec encompasses all data used to launch this execution. The Spec does not change over the lifetime
/// of an execution as it progresses across phase changes.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionSpec {
    /// Launch plan to be executed
    #[prost(message, optional, tag="1")]
    pub launch_plan: ::core::option::Option<super::core::Identifier>,
    /// Input values to be passed for the execution
    #[deprecated]
    #[prost(message, optional, tag="2")]
    pub inputs: ::core::option::Option<super::core::LiteralMap>,
    /// Metadata for the execution
    #[prost(message, optional, tag="3")]
    pub metadata: ::core::option::Option<ExecutionMetadata>,
    /// Labels to apply to the execution resource.
    #[prost(message, optional, tag="7")]
    pub labels: ::core::option::Option<Labels>,
    /// Annotations to apply to the execution resource.
    #[prost(message, optional, tag="8")]
    pub annotations: ::core::option::Option<Annotations>,
    /// Optional: security context override to apply this execution.
    #[prost(message, optional, tag="10")]
    pub security_context: ::core::option::Option<super::core::SecurityContext>,
    /// Optional: auth override to apply this execution.
    #[deprecated]
    #[prost(message, optional, tag="16")]
    pub auth_role: ::core::option::Option<AuthRole>,
    /// Indicates the runtime priority of the execution.
    #[prost(message, optional, tag="17")]
    pub quality_of_service: ::core::option::Option<super::core::QualityOfService>,
    /// Controls the maximum number of task nodes that can be run in parallel for the entire workflow.
    /// This is useful to achieve fairness. Note: MapTasks are regarded as one unit,
    /// and parallelism/concurrency of MapTasks is independent from this.
    #[prost(int32, tag="18")]
    pub max_parallelism: i32,
    /// User setting to configure where to store offloaded data (i.e. Blobs, structured datasets, query data, etc.).
    /// This should be a prefix like s3://my-bucket/my-data
    #[prost(message, optional, tag="19")]
    pub raw_output_data_config: ::core::option::Option<RawOutputDataConfig>,
    /// Controls how to select an available cluster on which this execution should run.
    #[prost(message, optional, tag="20")]
    pub cluster_assignment: ::core::option::Option<ClusterAssignment>,
    /// Allows for the interruptible flag of a workflow to be overwritten for a single execution.
    /// Omitting this field uses the workflow's value as a default.
    /// As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper
    /// around the bool field.
    #[prost(message, optional, tag="21")]
    pub interruptible: ::core::option::Option<bool>,
    /// Allows for all cached values of a workflow and its tasks to be overwritten for a single execution.
    /// If enabled, all calculations are performed even if cached results would be available, overwriting the stored
    /// data once execution finishes successfully.
    #[prost(bool, tag="22")]
    pub overwrite_cache: bool,
    /// Environment variables to be set for the execution.
    #[prost(message, optional, tag="23")]
    pub envs: ::core::option::Option<Envs>,
    /// Tags to be set for the execution.
    #[prost(string, repeated, tag="24")]
    pub tags: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(oneof="execution_spec::NotificationOverrides", tags="5, 6")]
    pub notification_overrides: ::core::option::Option<execution_spec::NotificationOverrides>,
}
/// Nested message and enum types in `ExecutionSpec`.
pub mod execution_spec {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum NotificationOverrides {
        /// List of notifications based on Execution status transitions
        /// When this list is not empty it is used rather than any notifications defined in the referenced launch plan.
        /// When this list is empty, the notifications defined for the launch plan will be applied.
        #[prost(message, tag="5")]
        Notifications(super::NotificationList),
        /// This should be set to true if all notifications are intended to be disabled for this execution.
        #[prost(bool, tag="6")]
        DisableAll(bool),
    }
}
/// Request to terminate an in-progress execution.  This action is irreversible.
/// If an execution is already terminated, this request will simply be a no-op.
/// This request will fail if it references a non-existent execution.
/// If the request succeeds the phase "ABORTED" will be recorded for the termination
/// with the optional cause added to the output_result.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionTerminateRequest {
    /// Uniquely identifies the individual workflow execution to be terminated.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// Optional reason for aborting.
    #[prost(string, tag="2")]
    pub cause: ::prost::alloc::string::String,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionTerminateResponse {
}
/// Request structure to fetch inputs, output and other data produced by an execution.
/// By default this data is not returned inline in :ref:`ref_flyteidl.admin.WorkflowExecutionGetRequest`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionGetDataRequest {
    /// The identifier of the execution for which to fetch inputs and outputs.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
}
/// Response structure for WorkflowExecutionGetDataRequest which contains inputs and outputs for an execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionGetDataResponse {
    /// Signed url to fetch a core.LiteralMap of execution outputs.
    /// Deprecated: Please use full_outputs instead.
    #[deprecated]
    #[prost(message, optional, tag="1")]
    pub outputs: ::core::option::Option<UrlBlob>,
    /// Signed url to fetch a core.LiteralMap of execution inputs.
    /// Deprecated: Please use full_inputs instead.
    #[deprecated]
    #[prost(message, optional, tag="2")]
    pub inputs: ::core::option::Option<UrlBlob>,
    /// Full_inputs will only be populated if they are under a configured size threshold.
    #[prost(message, optional, tag="3")]
    pub full_inputs: ::core::option::Option<super::core::LiteralMap>,
    /// Full_outputs will only be populated if they are under a configured size threshold.
    #[prost(message, optional, tag="4")]
    pub full_outputs: ::core::option::Option<super::core::LiteralMap>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionUpdateRequest {
    /// Identifier of the execution to update
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// State to set as the new value active/archive
    #[prost(enumeration="ExecutionState", tag="2")]
    pub state: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionStateChangeDetails {
    /// The state of the execution is used to control its visibility in the UI/CLI.
    #[prost(enumeration="ExecutionState", tag="1")]
    pub state: i32,
    /// This timestamp represents when the state changed.
    #[prost(message, optional, tag="2")]
    pub occurred_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Identifies the entity (if any) responsible for causing the state change of the execution
    #[prost(string, tag="3")]
    pub principal: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionUpdateResponse {
}
/// WorkflowExecutionGetMetricsRequest represents a request to retrieve metrics for the specified workflow execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionGetMetricsRequest {
    /// id defines the workflow execution to query for.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// depth defines the number of Flyte entity levels to traverse when breaking down execution details.
    #[prost(int32, tag="2")]
    pub depth: i32,
}
/// WorkflowExecutionGetMetricsResponse represents the response containing metrics for the specified workflow execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionGetMetricsResponse {
    /// Span defines the top-level breakdown of the workflows execution. More precise information is nested in a
    /// hierarchical structure using Flyte entity references.
    #[prost(message, optional, tag="1")]
    pub span: ::core::option::Option<super::core::Span>,
}
/// The state of the execution is used to control its visibility in the UI/CLI.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum ExecutionState {
    /// By default, all executions are considered active.
    ExecutionActive = 0,
    /// Archived executions are no longer visible in the UI.
    ExecutionArchived = 1,
}
impl ExecutionState {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            ExecutionState::ExecutionActive => "EXECUTION_ACTIVE",
            ExecutionState::ExecutionArchived => "EXECUTION_ARCHIVED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "EXECUTION_ACTIVE" => Some(Self::ExecutionActive),
            "EXECUTION_ARCHIVED" => Some(Self::ExecutionArchived),
            _ => None,
        }
    }
}
/// Option for schedules run at a certain frequency e.g. every 2 minutes.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FixedRate {
    #[prost(uint32, tag="1")]
    pub value: u32,
    #[prost(enumeration="FixedRateUnit", tag="2")]
    pub unit: i32,
}
/// Options for schedules to run according to a cron expression.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CronSchedule {
    /// Standard/default cron implementation as described by <https://en.wikipedia.org/wiki/Cron#CRON_expression;>
    /// Also supports nonstandard predefined scheduling definitions
    /// as described by <https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions>
    /// except @reboot
    #[prost(string, tag="1")]
    pub schedule: ::prost::alloc::string::String,
    /// ISO 8601 duration as described by <https://en.wikipedia.org/wiki/ISO_8601#Durations>
    #[prost(string, tag="2")]
    pub offset: ::prost::alloc::string::String,
}
/// Defines complete set of information required to trigger an execution on a schedule.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Schedule {
    /// Name of the input variable that the kickoff time will be supplied to when the workflow is kicked off.
    #[prost(string, tag="3")]
    pub kickoff_time_input_arg: ::prost::alloc::string::String,
    #[prost(oneof="schedule::ScheduleExpression", tags="1, 2, 4")]
    pub schedule_expression: ::core::option::Option<schedule::ScheduleExpression>,
}
/// Nested message and enum types in `Schedule`.
pub mod schedule {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum ScheduleExpression {
        /// Uses AWS syntax: Minutes Hours Day-of-month Month Day-of-week Year
        /// e.g. for a schedule that runs every 15 minutes: 0/15 * * * ? *
        #[prost(string, tag="1")]
        CronExpression(::prost::alloc::string::String),
        #[prost(message, tag="2")]
        Rate(super::FixedRate),
        #[prost(message, tag="4")]
        CronSchedule(super::CronSchedule),
    }
}
/// Represents a frequency at which to run a schedule.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum FixedRateUnit {
    Minute = 0,
    Hour = 1,
    Day = 2,
}
impl FixedRateUnit {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            FixedRateUnit::Minute => "MINUTE",
            FixedRateUnit::Hour => "HOUR",
            FixedRateUnit::Day => "DAY",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "MINUTE" => Some(Self::Minute),
            "HOUR" => Some(Self::Hour),
            "DAY" => Some(Self::Day),
            _ => None,
        }
    }
}
/// Request to register a launch plan. The included LaunchPlanSpec may have a complete or incomplete set of inputs required
/// to launch a workflow execution. By default all launch plans are registered in state INACTIVE. If you wish to
/// set the state to ACTIVE, you must submit a LaunchPlanUpdateRequest, after you have successfully created a launch plan.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlanCreateRequest {
    /// Uniquely identifies a launch plan entity.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// User-provided launch plan details, including reference workflow, inputs and other metadata.
    #[prost(message, optional, tag="2")]
    pub spec: ::core::option::Option<LaunchPlanSpec>,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlanCreateResponse {
}
/// A LaunchPlan provides the capability to templatize workflow executions.
/// Launch plans simplify associating one or more schedules, inputs and notifications with your workflows.
/// Launch plans can be shared and used to trigger executions with predefined inputs even when a workflow
/// definition doesn't necessarily have a default value for said input.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlan {
    /// Uniquely identifies a launch plan entity.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// User-provided launch plan details, including reference workflow, inputs and other metadata.
    #[prost(message, optional, tag="2")]
    pub spec: ::core::option::Option<LaunchPlanSpec>,
    /// Values computed by the flyte platform after launch plan registration.
    #[prost(message, optional, tag="3")]
    pub closure: ::core::option::Option<LaunchPlanClosure>,
}
/// Response object for list launch plan requests.
/// See :ref:`ref_flyteidl.admin.LaunchPlan` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlanList {
    #[prost(message, repeated, tag="1")]
    pub launch_plans: ::prost::alloc::vec::Vec<LaunchPlan>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Defines permissions associated with executions created by this launch plan spec.
/// Use either of these roles when they have permissions required by your workflow execution.
/// Deprecated.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Auth {
    /// Defines an optional iam role which will be used for tasks run in executions created with this launch plan.
    #[prost(string, tag="1")]
    pub assumable_iam_role: ::prost::alloc::string::String,
    /// Defines an optional kubernetes service account which will be used for tasks run in executions created with this launch plan.
    #[prost(string, tag="2")]
    pub kubernetes_service_account: ::prost::alloc::string::String,
}
/// User-provided launch plan definition and configuration values.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlanSpec {
    /// Reference to the Workflow template that the launch plan references
    #[prost(message, optional, tag="1")]
    pub workflow_id: ::core::option::Option<super::core::Identifier>,
    /// Metadata for the Launch Plan
    #[prost(message, optional, tag="2")]
    pub entity_metadata: ::core::option::Option<LaunchPlanMetadata>,
    /// Input values to be passed for the execution.
    /// These can be overriden when an execution is created with this launch plan.
    #[prost(message, optional, tag="3")]
    pub default_inputs: ::core::option::Option<super::core::ParameterMap>,
    /// Fixed, non-overridable inputs for the Launch Plan.
    /// These can not be overriden when an execution is created with this launch plan.
    #[prost(message, optional, tag="4")]
    pub fixed_inputs: ::core::option::Option<super::core::LiteralMap>,
    /// String to indicate the role to use to execute the workflow underneath
    #[deprecated]
    #[prost(string, tag="5")]
    pub role: ::prost::alloc::string::String,
    /// Custom labels to be applied to the execution resource.
    #[prost(message, optional, tag="6")]
    pub labels: ::core::option::Option<Labels>,
    /// Custom annotations to be applied to the execution resource.
    #[prost(message, optional, tag="7")]
    pub annotations: ::core::option::Option<Annotations>,
    /// Indicates the permission associated with workflow executions triggered with this launch plan.
    #[deprecated]
    #[prost(message, optional, tag="8")]
    pub auth: ::core::option::Option<Auth>,
    #[deprecated]
    #[prost(message, optional, tag="9")]
    pub auth_role: ::core::option::Option<AuthRole>,
    /// Indicates security context for permissions triggered with this launch plan
    #[prost(message, optional, tag="10")]
    pub security_context: ::core::option::Option<super::core::SecurityContext>,
    /// Indicates the runtime priority of the execution.
    #[prost(message, optional, tag="16")]
    pub quality_of_service: ::core::option::Option<super::core::QualityOfService>,
    /// Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).
    #[prost(message, optional, tag="17")]
    pub raw_output_data_config: ::core::option::Option<RawOutputDataConfig>,
    /// Controls the maximum number of tasknodes that can be run in parallel for the entire workflow.
    /// This is useful to achieve fairness. Note: MapTasks are regarded as one unit,
    /// and parallelism/concurrency of MapTasks is independent from this.
    #[prost(int32, tag="18")]
    pub max_parallelism: i32,
    /// Allows for the interruptible flag of a workflow to be overwritten for a single execution.
    /// Omitting this field uses the workflow's value as a default.
    /// As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper
    /// around the bool field.
    #[prost(message, optional, tag="19")]
    pub interruptible: ::core::option::Option<bool>,
    /// Allows for all cached values of a workflow and its tasks to be overwritten for a single execution.
    /// If enabled, all calculations are performed even if cached results would be available, overwriting the stored
    /// data once execution finishes successfully.
    #[prost(bool, tag="20")]
    pub overwrite_cache: bool,
    /// Environment variables to be set for the execution.
    #[prost(message, optional, tag="21")]
    pub envs: ::core::option::Option<Envs>,
}
/// Values computed by the flyte platform after launch plan registration.
/// These include expected_inputs required to be present in a CreateExecutionRequest
/// to launch the reference workflow as well timestamp values associated with the launch plan.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlanClosure {
    /// Indicate the Launch plan state. 
    #[prost(enumeration="LaunchPlanState", tag="1")]
    pub state: i32,
    /// Indicates the set of inputs expected when creating an execution with the Launch plan
    #[prost(message, optional, tag="2")]
    pub expected_inputs: ::core::option::Option<super::core::ParameterMap>,
    /// Indicates the set of outputs expected to be produced by creating an execution with the Launch plan
    #[prost(message, optional, tag="3")]
    pub expected_outputs: ::core::option::Option<super::core::VariableMap>,
    /// Time at which the launch plan was created.
    #[prost(message, optional, tag="4")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Time at which the launch plan was last updated.
    #[prost(message, optional, tag="5")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// Additional launch plan attributes included in the LaunchPlanSpec not strictly required to launch
/// the reference workflow.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlanMetadata {
    /// Schedule to execute the Launch Plan
    #[prost(message, optional, tag="1")]
    pub schedule: ::core::option::Option<Schedule>,
    /// List of notifications based on Execution status transitions
    #[prost(message, repeated, tag="2")]
    pub notifications: ::prost::alloc::vec::Vec<Notification>,
}
/// Request to set the referenced launch plan state to the configured value.
/// See :ref:`ref_flyteidl.admin.LaunchPlan` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlanUpdateRequest {
    /// Identifier of launch plan for which to change state.
    /// +required.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// Desired state to apply to the launch plan.
    /// +required.
    #[prost(enumeration="LaunchPlanState", tag="2")]
    pub state: i32,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LaunchPlanUpdateResponse {
}
/// Represents a request struct for finding an active launch plan for a given NamedEntityIdentifier
/// See :ref:`ref_flyteidl.admin.LaunchPlan` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ActiveLaunchPlanRequest {
    /// +required.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<NamedEntityIdentifier>,
}
/// Represents a request structure to list active launch plans within a project/domain.
/// See :ref:`ref_flyteidl.admin.LaunchPlan` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ActiveLaunchPlanListRequest {
    /// Name of the project that contains the identifiers.
    /// +required.
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Name of the domain the identifiers belongs to within the project.
    /// +required.
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// Indicates the number of resources to be returned.
    /// +required.
    #[prost(uint32, tag="3")]
    pub limit: u32,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="4")]
    pub token: ::prost::alloc::string::String,
    /// Sort ordering.
    /// +optional
    #[prost(message, optional, tag="5")]
    pub sort_by: ::core::option::Option<Sort>,
}
/// By default any launch plan regardless of state can be used to launch a workflow execution.
/// However, at most one version of a launch plan
/// (e.g. a NamedEntityIdentifier set of shared project, domain and name values) can be
/// active at a time in regards to *schedules*. That is, at most one schedule in a NamedEntityIdentifier
/// group will be observed and trigger executions at a defined cadence.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum LaunchPlanState {
    Inactive = 0,
    Active = 1,
}
impl LaunchPlanState {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            LaunchPlanState::Inactive => "INACTIVE",
            LaunchPlanState::Active => "ACTIVE",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "INACTIVE" => Some(Self::Inactive),
            "ACTIVE" => Some(Self::Active),
            _ => None,
        }
    }
}
/// Defines a set of overridable task resource attributes set during task registration.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskResourceSpec {
    #[prost(string, tag="1")]
    pub cpu: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub gpu: ::prost::alloc::string::String,
    #[prost(string, tag="3")]
    pub memory: ::prost::alloc::string::String,
    #[prost(string, tag="4")]
    pub storage: ::prost::alloc::string::String,
    #[prost(string, tag="5")]
    pub ephemeral_storage: ::prost::alloc::string::String,
}
/// Defines task resource defaults and limits that will be applied at task registration.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskResourceAttributes {
    #[prost(message, optional, tag="1")]
    pub defaults: ::core::option::Option<TaskResourceSpec>,
    #[prost(message, optional, tag="2")]
    pub limits: ::core::option::Option<TaskResourceSpec>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ClusterResourceAttributes {
    /// Custom resource attributes which will be applied in cluster resource creation (e.g. quotas).
    /// Map keys are the *case-sensitive* names of variables in templatized resource files.
    /// Map values should be the custom values which get substituted during resource creation.
    #[prost(map="string, string", tag="1")]
    pub attributes: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionQueueAttributes {
    /// Tags used for assigning execution queues for tasks defined within this project.
    #[prost(string, repeated, tag="1")]
    pub tags: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionClusterLabel {
    /// Label value to determine where the execution will be run
    #[prost(string, tag="1")]
    pub value: ::prost::alloc::string::String,
}
/// This MatchableAttribute configures selecting alternate plugin implementations for a given task type.
/// In addition to an override implementation a selection of fallbacks can be provided or other modes
/// for handling cases where the desired plugin override is not enabled in a given Flyte deployment.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PluginOverride {
    /// A predefined yet extensible Task type identifier.
    #[prost(string, tag="1")]
    pub task_type: ::prost::alloc::string::String,
    /// A set of plugin ids which should handle tasks of this type instead of the default registered plugin. The list will be tried in order until a plugin is found with that id.
    #[prost(string, repeated, tag="2")]
    pub plugin_id: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// Defines the behavior when no plugin from the plugin_id list is not found.
    #[prost(enumeration="plugin_override::MissingPluginBehavior", tag="4")]
    pub missing_plugin_behavior: i32,
}
/// Nested message and enum types in `PluginOverride`.
pub mod plugin_override {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum MissingPluginBehavior {
        /// By default, if this plugin is not enabled for a Flyte deployment then execution will fail.
        Fail = 0,
        /// Uses the system-configured default implementation.
        UseDefault = 1,
    }
    impl MissingPluginBehavior {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                MissingPluginBehavior::Fail => "FAIL",
                MissingPluginBehavior::UseDefault => "USE_DEFAULT",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "FAIL" => Some(Self::Fail),
                "USE_DEFAULT" => Some(Self::UseDefault),
                _ => None,
            }
        }
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PluginOverrides {
    #[prost(message, repeated, tag="1")]
    pub overrides: ::prost::alloc::vec::Vec<PluginOverride>,
}
/// Adds defaults for customizable workflow-execution specifications and overrides.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionConfig {
    /// Can be used to control the number of parallel nodes to run within the workflow. This is useful to achieve fairness.
    #[prost(int32, tag="1")]
    pub max_parallelism: i32,
    /// Indicates security context permissions for executions triggered with this matchable attribute. 
    #[prost(message, optional, tag="2")]
    pub security_context: ::core::option::Option<super::core::SecurityContext>,
    /// Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).
    #[prost(message, optional, tag="3")]
    pub raw_output_data_config: ::core::option::Option<RawOutputDataConfig>,
    /// Custom labels to be applied to a triggered execution resource.
    #[prost(message, optional, tag="4")]
    pub labels: ::core::option::Option<Labels>,
    /// Custom annotations to be applied to a triggered execution resource.
    #[prost(message, optional, tag="5")]
    pub annotations: ::core::option::Option<Annotations>,
    /// Allows for the interruptible flag of a workflow to be overwritten for a single execution.
    /// Omitting this field uses the workflow's value as a default.
    /// As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper
    /// around the bool field.
    #[prost(message, optional, tag="6")]
    pub interruptible: ::core::option::Option<bool>,
    /// Allows for all cached values of a workflow and its tasks to be overwritten for a single execution.
    /// If enabled, all calculations are performed even if cached results would be available, overwriting the stored
    /// data once execution finishes successfully.
    #[prost(bool, tag="7")]
    pub overwrite_cache: bool,
    /// Environment variables to be set for the execution.
    #[prost(message, optional, tag="8")]
    pub envs: ::core::option::Option<Envs>,
}
/// Generic container for encapsulating all types of the above attributes messages.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MatchingAttributes {
    #[prost(oneof="matching_attributes::Target", tags="1, 2, 3, 4, 5, 6, 7, 8")]
    pub target: ::core::option::Option<matching_attributes::Target>,
}
/// Nested message and enum types in `MatchingAttributes`.
pub mod matching_attributes {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Target {
        #[prost(message, tag="1")]
        TaskResourceAttributes(super::TaskResourceAttributes),
        #[prost(message, tag="2")]
        ClusterResourceAttributes(super::ClusterResourceAttributes),
        #[prost(message, tag="3")]
        ExecutionQueueAttributes(super::ExecutionQueueAttributes),
        #[prost(message, tag="4")]
        ExecutionClusterLabel(super::ExecutionClusterLabel),
        #[prost(message, tag="5")]
        QualityOfService(super::super::core::QualityOfService),
        #[prost(message, tag="6")]
        PluginOverrides(super::PluginOverrides),
        #[prost(message, tag="7")]
        WorkflowExecutionConfig(super::WorkflowExecutionConfig),
        #[prost(message, tag="8")]
        ClusterAssignment(super::ClusterAssignment),
    }
}
/// Represents a custom set of attributes applied for either a domain; a domain and project; or
/// domain, project and workflow name.
/// These are used to override system level defaults for kubernetes cluster resource management,
/// default execution values, and more all across different levels of specificity.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MatchableAttributesConfiguration {
    #[prost(message, optional, tag="1")]
    pub attributes: ::core::option::Option<MatchingAttributes>,
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    #[prost(string, tag="3")]
    pub project: ::prost::alloc::string::String,
    #[prost(string, tag="4")]
    pub workflow: ::prost::alloc::string::String,
    #[prost(string, tag="5")]
    pub launch_plan: ::prost::alloc::string::String,
}
/// Request all matching resource attributes for a resource type.
/// See :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListMatchableAttributesRequest {
    /// +required
    #[prost(enumeration="MatchableResource", tag="1")]
    pub resource_type: i32,
}
/// Response for a request for all matching resource attributes for a resource type.
/// See :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListMatchableAttributesResponse {
    #[prost(message, repeated, tag="1")]
    pub configurations: ::prost::alloc::vec::Vec<MatchableAttributesConfiguration>,
}
/// Defines a resource that can be configured by customizable Project-, ProjectDomain- or WorkflowAttributes
/// based on matching tags.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum MatchableResource {
    /// Applies to customizable task resource requests and limits.
    TaskResource = 0,
    /// Applies to configuring templated kubernetes cluster resources.
    ClusterResource = 1,
    /// Configures task and dynamic task execution queue assignment.
    ExecutionQueue = 2,
    /// Configures the K8s cluster label to be used for execution to be run
    ExecutionClusterLabel = 3,
    /// Configures default quality of service when undefined in an execution spec.
    QualityOfServiceSpecification = 4,
    /// Selects configurable plugin implementation behavior for a given task type.
    PluginOverride = 5,
    /// Adds defaults for customizable workflow-execution specifications and overrides.
    WorkflowExecutionConfig = 6,
    /// Controls how to select an available cluster on which this execution should run.
    ClusterAssignment = 7,
}
impl MatchableResource {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            MatchableResource::TaskResource => "TASK_RESOURCE",
            MatchableResource::ClusterResource => "CLUSTER_RESOURCE",
            MatchableResource::ExecutionQueue => "EXECUTION_QUEUE",
            MatchableResource::ExecutionClusterLabel => "EXECUTION_CLUSTER_LABEL",
            MatchableResource::QualityOfServiceSpecification => "QUALITY_OF_SERVICE_SPECIFICATION",
            MatchableResource::PluginOverride => "PLUGIN_OVERRIDE",
            MatchableResource::WorkflowExecutionConfig => "WORKFLOW_EXECUTION_CONFIG",
            MatchableResource::ClusterAssignment => "CLUSTER_ASSIGNMENT",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "TASK_RESOURCE" => Some(Self::TaskResource),
            "CLUSTER_RESOURCE" => Some(Self::ClusterResource),
            "EXECUTION_QUEUE" => Some(Self::ExecutionQueue),
            "EXECUTION_CLUSTER_LABEL" => Some(Self::ExecutionClusterLabel),
            "QUALITY_OF_SERVICE_SPECIFICATION" => Some(Self::QualityOfServiceSpecification),
            "PLUGIN_OVERRIDE" => Some(Self::PluginOverride),
            "WORKFLOW_EXECUTION_CONFIG" => Some(Self::WorkflowExecutionConfig),
            "CLUSTER_ASSIGNMENT" => Some(Self::ClusterAssignment),
            _ => None,
        }
    }
}
/// A message used to fetch a single node execution entity.
/// See :ref:`ref_flyteidl.admin.NodeExecution` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionGetRequest {
    /// Uniquely identifies an individual node execution.
    /// +required
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::NodeExecutionIdentifier>,
}
/// Represents a request structure to retrieve a list of node execution entities.
/// See :ref:`ref_flyteidl.admin.NodeExecution` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionListRequest {
    /// Indicates the workflow execution to filter by.
    /// +required
    #[prost(message, optional, tag="1")]
    pub workflow_execution_id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// Indicates the number of resources to be returned.
    /// +required
    #[prost(uint32, tag="2")]
    pub limit: u32,
    // In the case of multiple pages of results, the, server-provided token can be used to fetch the next page
    // in a query.
    // +optional

    #[prost(string, tag="3")]
    pub token: ::prost::alloc::string::String,
    /// Indicates a list of filters passed as string.
    /// More info on constructing filters : <Link>
    /// +optional
    #[prost(string, tag="4")]
    pub filters: ::prost::alloc::string::String,
    /// Sort ordering.
    /// +optional
    #[prost(message, optional, tag="5")]
    pub sort_by: ::core::option::Option<Sort>,
    /// Unique identifier of the parent node in the execution
    /// +optional
    #[prost(string, tag="6")]
    pub unique_parent_id: ::prost::alloc::string::String,
}
/// Represents a request structure to retrieve a list of node execution entities launched by a specific task.
/// This can arise when a task yields a subworkflow.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionForTaskListRequest {
    /// Indicates the node execution to filter by.
    /// +required
    #[prost(message, optional, tag="1")]
    pub task_execution_id: ::core::option::Option<super::core::TaskExecutionIdentifier>,
    /// Indicates the number of resources to be returned.
    /// +required
    #[prost(uint32, tag="2")]
    pub limit: u32,
    /// In the case of multiple pages of results, the, server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="3")]
    pub token: ::prost::alloc::string::String,
    /// Indicates a list of filters passed as string.
    /// More info on constructing filters : <Link>
    /// +optional
    #[prost(string, tag="4")]
    pub filters: ::prost::alloc::string::String,
    /// Sort ordering.
    /// +optional
    #[prost(message, optional, tag="5")]
    pub sort_by: ::core::option::Option<Sort>,
}
/// Encapsulates all details for a single node execution entity.
/// A node represents a component in the overall workflow graph. A node launch a task, multiple tasks, an entire nested
/// sub-workflow, or even a separate child-workflow execution.
/// The same task can be called repeatedly in a single workflow but each node is unique.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecution {
    /// Uniquely identifies an individual node execution.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::NodeExecutionIdentifier>,
    /// Path to remote data store where input blob is stored.
    #[prost(string, tag="2")]
    pub input_uri: ::prost::alloc::string::String,
    /// Computed results associated with this node execution.
    #[prost(message, optional, tag="3")]
    pub closure: ::core::option::Option<NodeExecutionClosure>,
    /// Metadata for Node Execution
    #[prost(message, optional, tag="4")]
    pub metadata: ::core::option::Option<NodeExecutionMetaData>,
}
/// Represents additional attributes related to a Node Execution
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionMetaData {
    /// Node executions are grouped depending on retries of the parent
    /// Retry group is unique within the context of a parent node.
    #[prost(string, tag="1")]
    pub retry_group: ::prost::alloc::string::String,
    /// Boolean flag indicating if the node has child nodes under it
    /// This can be true when a node contains a dynamic workflow which then produces
    /// child nodes.
    #[prost(bool, tag="2")]
    pub is_parent_node: bool,
    /// Node id of the node in the original workflow
    /// This maps to value of WorkflowTemplate.nodes\[X\].id
    #[prost(string, tag="3")]
    pub spec_node_id: ::prost::alloc::string::String,
    /// Boolean flag indicating if the node has contains a dynamic workflow which then produces child nodes.
    /// This is to distinguish between subworkflows and dynamic workflows which can both have is_parent_node as true.
    #[prost(bool, tag="4")]
    pub is_dynamic: bool,
}
/// Request structure to retrieve a list of node execution entities.
/// See :ref:`ref_flyteidl.admin.NodeExecution` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionList {
    #[prost(message, repeated, tag="1")]
    pub node_executions: ::prost::alloc::vec::Vec<NodeExecution>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Container for node execution details and results.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionClosure {
    /// The last recorded phase for this node execution.
    #[prost(enumeration="super::core::node_execution::Phase", tag="3")]
    pub phase: i32,
    /// Time at which the node execution began running.
    #[prost(message, optional, tag="4")]
    pub started_at: ::core::option::Option<::prost_types::Timestamp>,
    /// The amount of time the node execution spent running.
    #[prost(message, optional, tag="5")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    /// Time at which the node execution was created.
    #[prost(message, optional, tag="6")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Time at which the node execution was last updated.
    #[prost(message, optional, tag="7")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
    /// String location uniquely identifying where the deck HTML file is.
    /// NativeUrl specifies the url in the format of the configured storage provider (e.g. s3://my-bucket/randomstring/suffix.tar)
    #[prost(string, tag="11")]
    pub deck_uri: ::prost::alloc::string::String,
    /// dynamic_job_spec_uri is the location of the DynamicJobSpec proto message for a DynamicWorkflow. This is required
    /// to correctly recover partially completed executions where the subworkflow has already been compiled.
    #[prost(string, tag="12")]
    pub dynamic_job_spec_uri: ::prost::alloc::string::String,
    /// Only a node in a terminal state will have a non-empty output_result.
    #[prost(oneof="node_execution_closure::OutputResult", tags="1, 2, 10")]
    pub output_result: ::core::option::Option<node_execution_closure::OutputResult>,
    /// Store metadata for what the node launched.
    /// for ex: if this is a workflow node, we store information for the launched workflow.
    #[prost(oneof="node_execution_closure::TargetMetadata", tags="8, 9")]
    pub target_metadata: ::core::option::Option<node_execution_closure::TargetMetadata>,
}
/// Nested message and enum types in `NodeExecutionClosure`.
pub mod node_execution_closure {
    /// Only a node in a terminal state will have a non-empty output_result.
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum OutputResult {
        /// Links to a remotely stored, serialized core.LiteralMap of node execution outputs.
        /// DEPRECATED. Use GetNodeExecutionData to fetch output data instead.
        #[prost(string, tag="1")]
        OutputUri(::prost::alloc::string::String),
        /// Error information for the Node
        #[prost(message, tag="2")]
        Error(super::super::core::ExecutionError),
        /// Raw output data produced by this node execution.
        /// DEPRECATED. Use GetNodeExecutionData to fetch output data instead.
        #[prost(message, tag="10")]
        OutputData(super::super::core::LiteralMap),
    }
    /// Store metadata for what the node launched.
    /// for ex: if this is a workflow node, we store information for the launched workflow.
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum TargetMetadata {
        #[prost(message, tag="8")]
        WorkflowNodeMetadata(super::WorkflowNodeMetadata),
        #[prost(message, tag="9")]
        TaskNodeMetadata(super::TaskNodeMetadata),
    }
}
/// Metadata for a WorkflowNode
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowNodeMetadata {
    /// The identifier for a workflow execution launched by a node.
    #[prost(message, optional, tag="1")]
    pub execution_id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
}
/// Metadata for the case in which the node is a TaskNode
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskNodeMetadata {
    /// Captures the status of caching for this execution.
    #[prost(enumeration="super::core::CatalogCacheStatus", tag="1")]
    pub cache_status: i32,
    /// This structure carries the catalog artifact information
    #[prost(message, optional, tag="2")]
    pub catalog_key: ::core::option::Option<super::core::CatalogMetadata>,
    /// The latest checkpoint location
    #[prost(string, tag="4")]
    pub checkpoint_uri: ::prost::alloc::string::String,
}
/// For dynamic workflow nodes we capture information about the dynamic workflow definition that gets generated.
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
    /// required to correctly recover partially completed executions where the subworkflow has already been compiled.
    #[prost(string, tag="3")]
    pub dynamic_job_spec_uri: ::prost::alloc::string::String,
}
/// Request structure to fetch inputs and output for a node execution.
/// By default, these are not returned in :ref:`ref_flyteidl.admin.NodeExecutionGetRequest`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionGetDataRequest {
    /// The identifier of the node execution for which to fetch inputs and outputs.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::NodeExecutionIdentifier>,
}
/// Response structure for NodeExecutionGetDataRequest which contains inputs and outputs for a node execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionGetDataResponse {
    /// Signed url to fetch a core.LiteralMap of node execution inputs.
    /// Deprecated: Please use full_inputs instead.
    #[deprecated]
    #[prost(message, optional, tag="1")]
    pub inputs: ::core::option::Option<UrlBlob>,
    /// Signed url to fetch a core.LiteralMap of node execution outputs.
    /// Deprecated: Please use full_outputs instead.
    #[deprecated]
    #[prost(message, optional, tag="2")]
    pub outputs: ::core::option::Option<UrlBlob>,
    /// Full_inputs will only be populated if they are under a configured size threshold.
    #[prost(message, optional, tag="3")]
    pub full_inputs: ::core::option::Option<super::core::LiteralMap>,
    /// Full_outputs will only be populated if they are under a configured size threshold. 
    #[prost(message, optional, tag="4")]
    pub full_outputs: ::core::option::Option<super::core::LiteralMap>,
    /// Optional Workflow closure for a dynamically generated workflow, in the case this node yields a dynamic workflow we return its structure here.
    #[prost(message, optional, tag="16")]
    pub dynamic_workflow: ::core::option::Option<DynamicWorkflowNodeMetadata>,
    #[prost(message, optional, tag="17")]
    pub flyte_urls: ::core::option::Option<FlyteUrLs>,
}
/// Represents the Email object that is sent to a publisher/subscriber
/// to forward the notification.
/// Note: This is internal to Admin and doesn't need to be exposed to other components.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EmailMessage {
    /// The list of email addresses to receive an email with the content populated in the other fields.
    /// Currently, each email recipient will receive its own email.
    /// This populates the TO field.
    #[prost(string, repeated, tag="1")]
    pub recipients_email: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// The email of the sender.
    /// This populates the FROM field.
    #[prost(string, tag="2")]
    pub sender_email: ::prost::alloc::string::String,
    /// The content of the subject line.
    /// This populates the SUBJECT field.
    #[prost(string, tag="3")]
    pub subject_line: ::prost::alloc::string::String,
    /// The content of the email body.
    /// This populates the BODY field.
    #[prost(string, tag="4")]
    pub body: ::prost::alloc::string::String,
}
/// Namespace within a project commonly used to differentiate between different service instances.
/// e.g. "production", "development", etc.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Domain {
    /// Globally unique domain name.
    #[prost(string, tag="1")]
    pub id: ::prost::alloc::string::String,
    /// Display name.
    #[prost(string, tag="2")]
    pub name: ::prost::alloc::string::String,
}
/// Top-level namespace used to classify different entities like workflows and executions.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Project {
    /// Globally unique project name.
    #[prost(string, tag="1")]
    pub id: ::prost::alloc::string::String,
    /// Display name.
    #[prost(string, tag="2")]
    pub name: ::prost::alloc::string::String,
    #[prost(message, repeated, tag="3")]
    pub domains: ::prost::alloc::vec::Vec<Domain>,
    #[prost(string, tag="4")]
    pub description: ::prost::alloc::string::String,
    /// Leverage Labels from flyteidl.admin.common.proto to
    /// tag projects with ownership information.
    #[prost(message, optional, tag="5")]
    pub labels: ::core::option::Option<Labels>,
    #[prost(enumeration="project::ProjectState", tag="6")]
    pub state: i32,
}
/// Nested message and enum types in `Project`.
pub mod project {
    /// The state of the project is used to control its visibility in the UI and validity.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum ProjectState {
        /// By default, all projects are considered active.
        Active = 0,
        /// Archived projects are no longer visible in the UI and no longer valid.
        Archived = 1,
        /// System generated projects that aren't explicitly created or managed by a user.
        SystemGenerated = 2,
    }
    impl ProjectState {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                ProjectState::Active => "ACTIVE",
                ProjectState::Archived => "ARCHIVED",
                ProjectState::SystemGenerated => "SYSTEM_GENERATED",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "ACTIVE" => Some(Self::Active),
                "ARCHIVED" => Some(Self::Archived),
                "SYSTEM_GENERATED" => Some(Self::SystemGenerated),
                _ => None,
            }
        }
    }
}
/// Represents a list of projects.
/// See :ref:`ref_flyteidl.admin.Project` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Projects {
    #[prost(message, repeated, tag="1")]
    pub projects: ::prost::alloc::vec::Vec<Project>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Request to retrieve a list of projects matching specified filters. 
/// See :ref:`ref_flyteidl.admin.Project` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectListRequest {
    /// Indicates the number of projects to be returned.
    /// +required
    #[prost(uint32, tag="1")]
    pub limit: u32,
    /// In the case of multiple pages of results, this server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
    /// Indicates a list of filters passed as string.
    /// More info on constructing filters : <Link>
    /// +optional
    #[prost(string, tag="3")]
    pub filters: ::prost::alloc::string::String,
    /// Sort ordering.
    /// +optional
    #[prost(message, optional, tag="4")]
    pub sort_by: ::core::option::Option<Sort>,
}
/// Adds a new user-project within the Flyte deployment.
/// See :ref:`ref_flyteidl.admin.Project` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectRegisterRequest {
    /// +required
    #[prost(message, optional, tag="1")]
    pub project: ::core::option::Option<Project>,
}
/// Purposefully empty, may be updated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectRegisterResponse {
}
/// Purposefully empty, may be updated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectUpdateResponse {
}
/// Defines a set of custom matching attributes at the project level.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectAttributes {
    /// Unique project id for which this set of attributes will be applied.
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    #[prost(message, optional, tag="2")]
    pub matching_attributes: ::core::option::Option<MatchingAttributes>,
}
/// Sets custom attributes for a project
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectAttributesUpdateRequest {
    /// +required
    #[prost(message, optional, tag="1")]
    pub attributes: ::core::option::Option<ProjectAttributes>,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectAttributesUpdateResponse {
}
/// Request to get an individual project level attribute override.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectAttributesGetRequest {
    /// Unique project id which this set of attributes references.
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Which type of matchable attributes to return.
    /// +required
    #[prost(enumeration="MatchableResource", tag="2")]
    pub resource_type: i32,
}
/// Response to get an individual project level attribute override.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectAttributesGetResponse {
    #[prost(message, optional, tag="1")]
    pub attributes: ::core::option::Option<ProjectAttributes>,
}
/// Request to delete a set matchable project level attribute override.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectAttributesDeleteRequest {
    /// Unique project id which this set of attributes references.
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Which type of matchable attributes to delete.
    /// +required
    #[prost(enumeration="MatchableResource", tag="2")]
    pub resource_type: i32,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectAttributesDeleteResponse {
}
/// Defines a set of custom matching attributes which defines resource defaults for a project and domain.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectDomainAttributes {
    /// Unique project id for which this set of attributes will be applied.
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Unique domain id for which this set of attributes will be applied.
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    #[prost(message, optional, tag="3")]
    pub matching_attributes: ::core::option::Option<MatchingAttributes>,
}
/// Sets custom attributes for a project-domain combination.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectDomainAttributesUpdateRequest {
    /// +required
    #[prost(message, optional, tag="1")]
    pub attributes: ::core::option::Option<ProjectDomainAttributes>,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectDomainAttributesUpdateResponse {
}
/// Request to get an individual project domain attribute override.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectDomainAttributesGetRequest {
    /// Unique project id which this set of attributes references.
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Unique domain id which this set of attributes references.
    /// +required
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// Which type of matchable attributes to return.
    /// +required
    #[prost(enumeration="MatchableResource", tag="3")]
    pub resource_type: i32,
}
/// Response to get an individual project domain attribute override.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectDomainAttributesGetResponse {
    #[prost(message, optional, tag="1")]
    pub attributes: ::core::option::Option<ProjectDomainAttributes>,
}
/// Request to delete a set matchable project domain attribute override.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectDomainAttributesDeleteRequest {
    /// Unique project id which this set of attributes references.
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Unique domain id which this set of attributes references.
    /// +required
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// Which type of matchable attributes to delete.
    /// +required
    #[prost(enumeration="MatchableResource", tag="3")]
    pub resource_type: i32,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProjectDomainAttributesDeleteResponse {
}
/// SignalGetOrCreateRequest represents a request structure to retrive or create a signal.
/// See :ref:`ref_flyteidl.admin.Signal` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignalGetOrCreateRequest {
    /// A unique identifier for the requested signal.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::SignalIdentifier>,
    /// A type denoting the required value type for this signal.
    #[prost(message, optional, tag="2")]
    pub r#type: ::core::option::Option<super::core::LiteralType>,
}
/// SignalListRequest represents a request structure to retrieve a collection of signals.
/// See :ref:`ref_flyteidl.admin.Signal` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignalListRequest {
    /// Indicates the workflow execution to filter by.
    /// +required
    #[prost(message, optional, tag="1")]
    pub workflow_execution_id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    /// Indicates the number of resources to be returned.
    /// +required
    #[prost(uint32, tag="2")]
    pub limit: u32,
    /// In the case of multiple pages of results, the, server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="3")]
    pub token: ::prost::alloc::string::String,
    /// Indicates a list of filters passed as string.
    /// +optional
    #[prost(string, tag="4")]
    pub filters: ::prost::alloc::string::String,
    /// Sort ordering.
    /// +optional
    #[prost(message, optional, tag="5")]
    pub sort_by: ::core::option::Option<Sort>,
}
/// SignalList represents collection of signals along with the token of the last result.
/// See :ref:`ref_flyteidl.admin.Signal` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignalList {
    /// A list of signals matching the input filters.
    #[prost(message, repeated, tag="1")]
    pub signals: ::prost::alloc::vec::Vec<Signal>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// SignalSetRequest represents a request structure to set the value on a signal. Setting a signal
/// effetively satisfies the signal condition within a Flyte workflow.
/// See :ref:`ref_flyteidl.admin.Signal` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignalSetRequest {
    /// A unique identifier for the requested signal.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::SignalIdentifier>,
    /// The value of this signal, must match the defining signal type.
    #[prost(message, optional, tag="2")]
    pub value: ::core::option::Option<super::core::Literal>,
}
/// SignalSetResponse represents a response structure if signal setting succeeds.
///
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignalSetResponse {
}
/// Signal encapsulates a unique identifier, associated metadata, and a value for a single Flyte
/// signal. Signals may exist either without a set value (representing a signal request) or with a
/// populated value (indicating the signal has been given).
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Signal {
    /// A unique identifier for the requested signal.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::SignalIdentifier>,
    /// A type denoting the required value type for this signal.
    #[prost(message, optional, tag="2")]
    pub r#type: ::core::option::Option<super::core::LiteralType>,
    /// The value of the signal. This is only available if the signal has been "set" and must match
    /// the defined the type.
    #[prost(message, optional, tag="3")]
    pub value: ::core::option::Option<super::core::Literal>,
}
/// Represents a request structure to create a revision of a task.
/// See :ref:`ref_flyteidl.admin.Task` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskCreateRequest {
    /// id represents the unique identifier of the task.
    /// +required
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// Represents the specification for task.
    /// +required
    #[prost(message, optional, tag="2")]
    pub spec: ::core::option::Option<TaskSpec>,
}
/// Represents a response structure if task creation succeeds.
///
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskCreateResponse {
}
/// Flyte workflows are composed of many ordered tasks. That is small, reusable, self-contained logical blocks
/// arranged to process workflow inputs and produce a deterministic set of outputs.
/// Tasks can come in many varieties tuned for specialized behavior. 
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Task {
    /// id represents the unique identifier of the task.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// closure encapsulates all the fields that maps to a compiled version of the task.
    #[prost(message, optional, tag="2")]
    pub closure: ::core::option::Option<TaskClosure>,
    /// One-liner overview of the entity.
    #[prost(string, tag="3")]
    pub short_description: ::prost::alloc::string::String,
}
/// Represents a list of tasks returned from the admin.
/// See :ref:`ref_flyteidl.admin.Task` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskList {
    /// A list of tasks returned based on the request.
    #[prost(message, repeated, tag="1")]
    pub tasks: ::prost::alloc::vec::Vec<Task>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Represents a structure that encapsulates the user-configured specification of the task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskSpec {
    /// Template of the task that encapsulates all the metadata of the task.
    #[prost(message, optional, tag="1")]
    pub template: ::core::option::Option<super::core::TaskTemplate>,
    /// Represents the specification for description entity.
    #[prost(message, optional, tag="2")]
    pub description: ::core::option::Option<DescriptionEntity>,
}
/// Compute task attributes which include values derived from the TaskSpec, as well as plugin-specific data
/// and task metadata.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskClosure {
    /// Represents the compiled representation of the task from the specification provided.
    #[prost(message, optional, tag="1")]
    pub compiled_task: ::core::option::Option<super::core::CompiledTask>,
    /// Time at which the task was created.
    #[prost(message, optional, tag="2")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// A message used to fetch a single task execution entity.
/// See :ref:`ref_flyteidl.admin.TaskExecution` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionGetRequest {
    /// Unique identifier for the task execution.
    /// +required
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::TaskExecutionIdentifier>,
}
/// Represents a request structure to retrieve a list of task execution entities yielded by a specific node execution.
/// See :ref:`ref_flyteidl.admin.TaskExecution` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionListRequest {
    /// Indicates the node execution to filter by.
    /// +required
    #[prost(message, optional, tag="1")]
    pub node_execution_id: ::core::option::Option<super::core::NodeExecutionIdentifier>,
    /// Indicates the number of resources to be returned.
    /// +required
    #[prost(uint32, tag="2")]
    pub limit: u32,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query.
    /// +optional
    #[prost(string, tag="3")]
    pub token: ::prost::alloc::string::String,
    /// Indicates a list of filters passed as string.
    /// More info on constructing filters : <Link>
    /// +optional
    #[prost(string, tag="4")]
    pub filters: ::prost::alloc::string::String,
    /// Sort ordering for returned list.
    /// +optional
    #[prost(message, optional, tag="5")]
    pub sort_by: ::core::option::Option<Sort>,
}
/// Encapsulates all details for a single task execution entity.
/// A task execution represents an instantiated task, including all inputs and additional
/// metadata as well as computed results included state, outputs, and duration-based attributes.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecution {
    /// Unique identifier for the task execution.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::TaskExecutionIdentifier>,
    /// Path to remote data store where input blob is stored.
    #[prost(string, tag="2")]
    pub input_uri: ::prost::alloc::string::String,
    /// Task execution details and results.
    #[prost(message, optional, tag="3")]
    pub closure: ::core::option::Option<TaskExecutionClosure>,
    /// Whether this task spawned nodes.
    #[prost(bool, tag="4")]
    pub is_parent: bool,
}
/// Response structure for a query to list of task execution entities.
/// See :ref:`ref_flyteidl.admin.TaskExecution` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionList {
    #[prost(message, repeated, tag="1")]
    pub task_executions: ::prost::alloc::vec::Vec<TaskExecution>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Container for task execution details and results.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionClosure {
    /// The last recorded phase for this task execution.
    #[prost(enumeration="super::core::task_execution::Phase", tag="3")]
    pub phase: i32,
    /// Detailed log information output by the task execution.
    #[prost(message, repeated, tag="4")]
    pub logs: ::prost::alloc::vec::Vec<super::core::TaskLog>,
    /// Time at which the task execution began running.
    #[prost(message, optional, tag="5")]
    pub started_at: ::core::option::Option<::prost_types::Timestamp>,
    /// The amount of time the task execution spent running.
    #[prost(message, optional, tag="6")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    /// Time at which the task execution was created.
    #[prost(message, optional, tag="7")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Time at which the task execution was last updated.
    #[prost(message, optional, tag="8")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Custom data specific to the task plugin.
    #[prost(message, optional, tag="9")]
    pub custom_info: ::core::option::Option<::prost_types::Struct>,
    /// If there is an explanation for the most recent phase transition, the reason will capture it.
    #[prost(string, tag="10")]
    pub reason: ::prost::alloc::string::String,
    /// A predefined yet extensible Task type identifier.
    #[prost(string, tag="11")]
    pub task_type: ::prost::alloc::string::String,
    /// Metadata around how a task was executed.
    #[prost(message, optional, tag="16")]
    pub metadata: ::core::option::Option<super::event::TaskExecutionMetadata>,
    /// The event version is used to indicate versioned changes in how data is maintained using this
    /// proto message. For example, event_verison > 0 means that maps tasks logs use the
    /// TaskExecutionMetadata ExternalResourceInfo fields for each subtask rather than the TaskLog
    /// in this message.
    #[prost(int32, tag="17")]
    pub event_version: i32,
    /// A time-series of the phase transition or update explanations. This, when compared to storing a singular reason
    /// as previously done, is much more valuable in visualizing and understanding historical evaluations.
    #[prost(message, repeated, tag="18")]
    pub reasons: ::prost::alloc::vec::Vec<Reason>,
    #[prost(oneof="task_execution_closure::OutputResult", tags="1, 2, 12")]
    pub output_result: ::core::option::Option<task_execution_closure::OutputResult>,
}
/// Nested message and enum types in `TaskExecutionClosure`.
pub mod task_execution_closure {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum OutputResult {
        /// Path to remote data store where output blob is stored if the execution succeeded (and produced outputs).
        /// DEPRECATED. Use GetTaskExecutionData to fetch output data instead.
        #[prost(string, tag="1")]
        OutputUri(::prost::alloc::string::String),
        /// Error information for the task execution. Populated if the execution failed.
        #[prost(message, tag="2")]
        Error(super::super::core::ExecutionError),
        /// Raw output data produced by this task execution.
        /// DEPRECATED. Use GetTaskExecutionData to fetch output data instead.
        #[prost(message, tag="12")]
        OutputData(super::super::core::LiteralMap),
    }
}
/// Reason is a single message annotated with a timestamp to indicate the instant the reason occurred.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Reason {
    /// occurred_at is the timestamp indicating the instant that this reason happened.
    #[prost(message, optional, tag="1")]
    pub occurred_at: ::core::option::Option<::prost_types::Timestamp>,
    /// message is the explanation for the most recent phase transition or status update.
    #[prost(string, tag="2")]
    pub message: ::prost::alloc::string::String,
}
/// Request structure to fetch inputs and output for a task execution.
/// By default this data is not returned inline in :ref:`ref_flyteidl.admin.TaskExecutionGetRequest`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionGetDataRequest {
    /// The identifier of the task execution for which to fetch inputs and outputs.
    /// +required
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::TaskExecutionIdentifier>,
}
/// Response structure for TaskExecutionGetDataRequest which contains inputs and outputs for a task execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionGetDataResponse {
    /// Signed url to fetch a core.LiteralMap of task execution inputs.
    /// Deprecated: Please use full_inputs instead.
    #[deprecated]
    #[prost(message, optional, tag="1")]
    pub inputs: ::core::option::Option<UrlBlob>,
    /// Signed url to fetch a core.LiteralMap of task execution outputs.
    /// Deprecated: Please use full_outputs instead.
    #[deprecated]
    #[prost(message, optional, tag="2")]
    pub outputs: ::core::option::Option<UrlBlob>,
    /// Full_inputs will only be populated if they are under a configured size threshold.
    #[prost(message, optional, tag="3")]
    pub full_inputs: ::core::option::Option<super::core::LiteralMap>,
    /// Full_outputs will only be populated if they are under a configured size threshold.
    #[prost(message, optional, tag="4")]
    pub full_outputs: ::core::option::Option<super::core::LiteralMap>,
    /// flyte tiny url to fetch a core.LiteralMap of task execution's IO
    /// Deck will be empty for task
    #[prost(message, optional, tag="5")]
    pub flyte_urls: ::core::option::Option<FlyteUrLs>,
}
/// Response for the GetVersion API
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetVersionResponse {
    /// The control plane version information. FlyteAdmin and related components
    /// form the control plane of Flyte
    #[prost(message, optional, tag="1")]
    pub control_plane_version: ::core::option::Option<Version>,
}
/// Provides Version information for a component
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Version {
    /// Specifies the GIT sha of the build
    #[prost(string, tag="1")]
    pub build: ::prost::alloc::string::String,
    /// Version for the build, should follow a semver
    #[prost(string, tag="2")]
    pub version: ::prost::alloc::string::String,
    /// Build timestamp
    #[prost(string, tag="3")]
    pub build_time: ::prost::alloc::string::String,
}
/// Empty request for GetVersion
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetVersionRequest {
}
/// Represents a request structure to create a revision of a workflow.
/// See :ref:`ref_flyteidl.admin.Workflow` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowCreateRequest {
    /// id represents the unique identifier of the workflow.
    /// +required
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// Represents the specification for workflow.
    /// +required
    #[prost(message, optional, tag="2")]
    pub spec: ::core::option::Option<WorkflowSpec>,
}
/// Purposefully empty, may be populated in the future. 
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowCreateResponse {
}
/// Represents the workflow structure stored in the Admin
/// A workflow is created by ordering tasks and associating outputs to inputs
/// in order to produce a directed-acyclic execution graph.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Workflow {
    /// id represents the unique identifier of the workflow.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
    /// closure encapsulates all the fields that maps to a compiled version of the workflow.
    #[prost(message, optional, tag="2")]
    pub closure: ::core::option::Option<WorkflowClosure>,
    /// One-liner overview of the entity.
    #[prost(string, tag="3")]
    pub short_description: ::prost::alloc::string::String,
}
/// Represents a list of workflows returned from the admin.
/// See :ref:`ref_flyteidl.admin.Workflow` for more details
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowList {
    /// A list of workflows returned based on the request.
    #[prost(message, repeated, tag="1")]
    pub workflows: ::prost::alloc::vec::Vec<Workflow>,
    /// In the case of multiple pages of results, the server-provided token can be used to fetch the next page
    /// in a query. If there are no more results, this value will be empty.
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
}
/// Represents a structure that encapsulates the specification of the workflow.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowSpec {
    /// Template of the task that encapsulates all the metadata of the workflow.
    #[prost(message, optional, tag="1")]
    pub template: ::core::option::Option<super::core::WorkflowTemplate>,
    /// Workflows that are embedded into other workflows need to be passed alongside the parent workflow to the
    /// propeller compiler (since the compiler doesn't have any knowledge of other workflows - ie, it doesn't reach out
    /// to Admin to see other registered workflows).  In fact, subworkflows do not even need to be registered.
    #[prost(message, repeated, tag="2")]
    pub sub_workflows: ::prost::alloc::vec::Vec<super::core::WorkflowTemplate>,
    /// Represents the specification for description entity.
    #[prost(message, optional, tag="3")]
    pub description: ::core::option::Option<DescriptionEntity>,
}
/// A container holding the compiled workflow produced from the WorkflowSpec and additional metadata.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowClosure {
    /// Represents the compiled representation of the workflow from the specification provided.
    #[prost(message, optional, tag="1")]
    pub compiled_workflow: ::core::option::Option<super::core::CompiledWorkflowClosure>,
    /// Time at which the workflow was created.
    #[prost(message, optional, tag="2")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// The workflow id is already used and the structure is different
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowErrorExistsDifferentStructure {
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
}
/// The workflow id is already used with an identical sctructure
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowErrorExistsIdenticalStructure {
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<super::core::Identifier>,
}
/// When a CreateWorkflowRequest failes due to matching id
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateWorkflowFailureReason {
    #[prost(oneof="create_workflow_failure_reason::Reason", tags="1, 2")]
    pub reason: ::core::option::Option<create_workflow_failure_reason::Reason>,
}
/// Nested message and enum types in `CreateWorkflowFailureReason`.
pub mod create_workflow_failure_reason {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Reason {
        #[prost(message, tag="1")]
        ExistsDifferentStructure(super::WorkflowErrorExistsDifferentStructure),
        #[prost(message, tag="2")]
        ExistsIdenticalStructure(super::WorkflowErrorExistsIdenticalStructure),
    }
}
/// Defines a set of custom matching attributes which defines resource defaults for a project, domain and workflow.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowAttributes {
    /// Unique project id for which this set of attributes will be applied.
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Unique domain id for which this set of attributes will be applied.
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// Workflow name for which this set of attributes will be applied.
    #[prost(string, tag="3")]
    pub workflow: ::prost::alloc::string::String,
    #[prost(message, optional, tag="4")]
    pub matching_attributes: ::core::option::Option<MatchingAttributes>,
}
/// Sets custom attributes for a project, domain and workflow combination.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowAttributesUpdateRequest {
    #[prost(message, optional, tag="1")]
    pub attributes: ::core::option::Option<WorkflowAttributes>,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowAttributesUpdateResponse {
}
/// Request to get an individual workflow attribute override.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowAttributesGetRequest {
    /// Unique project id which this set of attributes references.
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Unique domain id which this set of attributes references.
    /// +required
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// Workflow name which this set of attributes references.
    /// +required
    #[prost(string, tag="3")]
    pub workflow: ::prost::alloc::string::String,
    /// Which type of matchable attributes to return.
    /// +required
    #[prost(enumeration="MatchableResource", tag="4")]
    pub resource_type: i32,
}
/// Response to get an individual workflow attribute override.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowAttributesGetResponse {
    #[prost(message, optional, tag="1")]
    pub attributes: ::core::option::Option<WorkflowAttributes>,
}
/// Request to delete a set matchable workflow attribute override.
/// For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowAttributesDeleteRequest {
    /// Unique project id which this set of attributes references.
    /// +required
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Unique domain id which this set of attributes references.
    /// +required
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// Workflow name which this set of attributes references.
    /// +required
    #[prost(string, tag="3")]
    pub workflow: ::prost::alloc::string::String,
    /// Which type of matchable attributes to delete.
    /// +required
    #[prost(enumeration="MatchableResource", tag="4")]
    pub resource_type: i32,
}
/// Purposefully empty, may be populated in the future.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowAttributesDeleteResponse {
}
// @@protoc_insertion_point(module)
