// @generated
///
/// Request message for creating a Dataset.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateDatasetRequest {
    #[prost(message, optional, tag="1")]
    pub dataset: ::core::option::Option<Dataset>,
}
///
/// Response message for creating a Dataset
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateDatasetResponse {
}
///
/// Request message for retrieving a Dataset. The Dataset is retrieved by it's unique identifier
/// which is a combination of several fields.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDatasetRequest {
    #[prost(message, optional, tag="1")]
    pub dataset: ::core::option::Option<DatasetId>,
}
///
/// Response message for retrieving a Dataset. The response will include the metadata for the
/// Dataset.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDatasetResponse {
    #[prost(message, optional, tag="1")]
    pub dataset: ::core::option::Option<Dataset>,
}
///
/// Request message for retrieving an Artifact. Retrieve an artifact based on a query handle that
/// can be one of artifact_id or tag. The result returned will include the artifact data and metadata
/// associated with the artifact.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetArtifactRequest {
    #[prost(message, optional, tag="1")]
    pub dataset: ::core::option::Option<DatasetId>,
    #[prost(oneof="get_artifact_request::QueryHandle", tags="2, 3")]
    pub query_handle: ::core::option::Option<get_artifact_request::QueryHandle>,
}
/// Nested message and enum types in `GetArtifactRequest`.
pub mod get_artifact_request {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum QueryHandle {
        #[prost(string, tag="2")]
        ArtifactId(::prost::alloc::string::String),
        #[prost(string, tag="3")]
        TagName(::prost::alloc::string::String),
    }
}
///
/// Response message for retrieving an Artifact. The result returned will include the artifact data
/// and metadata associated with the artifact.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetArtifactResponse {
    #[prost(message, optional, tag="1")]
    pub artifact: ::core::option::Option<Artifact>,
}
///
/// Request message for creating an Artifact and its associated artifact Data.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateArtifactRequest {
    #[prost(message, optional, tag="1")]
    pub artifact: ::core::option::Option<Artifact>,
}
///
/// Response message for creating an Artifact.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateArtifactResponse {
}
///
/// Request message for tagging an Artifact.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AddTagRequest {
    #[prost(message, optional, tag="1")]
    pub tag: ::core::option::Option<Tag>,
}
///
/// Response message for tagging an Artifact.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AddTagResponse {
}
/// List the artifacts that belong to the Dataset, optionally filtered using filtered expression.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListArtifactsRequest {
    /// Use a datasetID for which you want to retrieve the artifacts
    #[prost(message, optional, tag="1")]
    pub dataset: ::core::option::Option<DatasetId>,
    /// Apply the filter expression to this query
    #[prost(message, optional, tag="2")]
    pub filter: ::core::option::Option<FilterExpression>,
    /// Pagination options to get a page of artifacts
    #[prost(message, optional, tag="3")]
    pub pagination: ::core::option::Option<PaginationOptions>,
}
/// Response to list artifacts
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListArtifactsResponse {
    /// The list of artifacts
    #[prost(message, repeated, tag="1")]
    pub artifacts: ::prost::alloc::vec::Vec<Artifact>,
    /// Token to use to request the next page, pass this into the next requests PaginationOptions
    #[prost(string, tag="2")]
    pub next_token: ::prost::alloc::string::String,
}
/// List the datasets for the given query
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListDatasetsRequest {
    /// Apply the filter expression to this query
    #[prost(message, optional, tag="1")]
    pub filter: ::core::option::Option<FilterExpression>,
    /// Pagination options to get a page of datasets
    #[prost(message, optional, tag="2")]
    pub pagination: ::core::option::Option<PaginationOptions>,
}
/// List the datasets response with token for next pagination
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListDatasetsResponse {
    /// The list of datasets
    #[prost(message, repeated, tag="1")]
    pub datasets: ::prost::alloc::vec::Vec<Dataset>,
    /// Token to use to request the next page, pass this into the next requests PaginationOptions
    #[prost(string, tag="2")]
    pub next_token: ::prost::alloc::string::String,
}
///
/// Request message for updating an Artifact and overwriting its associated ArtifactData.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UpdateArtifactRequest {
    /// ID of dataset the artifact is associated with
    #[prost(message, optional, tag="1")]
    pub dataset: ::core::option::Option<DatasetId>,
    /// List of data to overwrite stored artifact data with. Must contain ALL data for updated Artifact as any missing
    /// ArtifactData entries will be removed from the underlying blob storage and database.
    #[prost(message, repeated, tag="4")]
    pub data: ::prost::alloc::vec::Vec<ArtifactData>,
    /// Either ID of artifact or name of tag to retrieve existing artifact from
    #[prost(oneof="update_artifact_request::QueryHandle", tags="2, 3")]
    pub query_handle: ::core::option::Option<update_artifact_request::QueryHandle>,
}
/// Nested message and enum types in `UpdateArtifactRequest`.
pub mod update_artifact_request {
    /// Either ID of artifact or name of tag to retrieve existing artifact from
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum QueryHandle {
        #[prost(string, tag="2")]
        ArtifactId(::prost::alloc::string::String),
        #[prost(string, tag="3")]
        TagName(::prost::alloc::string::String),
    }
}
///
/// Response message for updating an Artifact.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UpdateArtifactResponse {
    /// The unique ID of the artifact updated
    #[prost(string, tag="1")]
    pub artifact_id: ::prost::alloc::string::String,
}
///
/// ReservationID message that is composed of several string fields.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ReservationId {
    /// The unique ID for the reserved dataset
    #[prost(message, optional, tag="1")]
    pub dataset_id: ::core::option::Option<DatasetId>,
    /// The specific artifact tag for the reservation
    #[prost(string, tag="2")]
    pub tag_name: ::prost::alloc::string::String,
}
/// Try to acquire or extend an artifact reservation. If an active reservation exists, retreive that instance.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetOrExtendReservationRequest {
    /// The unique ID for the reservation
    #[prost(message, optional, tag="1")]
    pub reservation_id: ::core::option::Option<ReservationId>,
    /// The unique ID of the owner for the reservation
    #[prost(string, tag="2")]
    pub owner_id: ::prost::alloc::string::String,
    /// Requested reservation extension heartbeat interval
    #[prost(message, optional, tag="3")]
    pub heartbeat_interval: ::core::option::Option<::prost_types::Duration>,
}
/// A reservation including owner, heartbeat interval, expiration timestamp, and various metadata.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Reservation {
    /// The unique ID for the reservation
    #[prost(message, optional, tag="1")]
    pub reservation_id: ::core::option::Option<ReservationId>,
    /// The unique ID of the owner for the reservation
    #[prost(string, tag="2")]
    pub owner_id: ::prost::alloc::string::String,
    /// Recommended heartbeat interval to extend reservation
    #[prost(message, optional, tag="3")]
    pub heartbeat_interval: ::core::option::Option<::prost_types::Duration>,
    /// Expiration timestamp of this reservation
    #[prost(message, optional, tag="4")]
    pub expires_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Free-form metadata associated with the artifact
    #[prost(message, optional, tag="6")]
    pub metadata: ::core::option::Option<Metadata>,
}
/// Response including either a newly minted reservation or the existing reservation
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetOrExtendReservationResponse {
    /// The reservation to be acquired or extended
    #[prost(message, optional, tag="1")]
    pub reservation: ::core::option::Option<Reservation>,
}
/// Request to release reservation
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ReleaseReservationRequest {
    /// The unique ID for the reservation
    #[prost(message, optional, tag="1")]
    pub reservation_id: ::core::option::Option<ReservationId>,
    /// The unique ID of the owner for the reservation
    #[prost(string, tag="2")]
    pub owner_id: ::prost::alloc::string::String,
}
/// Response to release reservation
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ReleaseReservationResponse {
}
///
/// Dataset message. It is uniquely identified by DatasetID.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Dataset {
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<DatasetId>,
    #[prost(message, optional, tag="2")]
    pub metadata: ::core::option::Option<Metadata>,
    #[prost(string, repeated, tag="3")]
    pub partition_keys: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
///
/// An artifact could have multiple partitions and each partition can have an arbitrary string key/value pair
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Partition {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub value: ::prost::alloc::string::String,
}
///
/// DatasetID message that is composed of several string fields.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DatasetId {
    /// The name of the project
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// The name of the dataset
    #[prost(string, tag="2")]
    pub name: ::prost::alloc::string::String,
    /// The domain (eg. environment)
    #[prost(string, tag="3")]
    pub domain: ::prost::alloc::string::String,
    /// Version of the data schema
    #[prost(string, tag="4")]
    pub version: ::prost::alloc::string::String,
    /// UUID for the dataset (if set the above fields are optional)
    #[prost(string, tag="5")]
    pub uuid: ::prost::alloc::string::String,
}
///
/// Artifact message. It is composed of several string fields.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Artifact {
    /// The unique ID of the artifact
    #[prost(string, tag="1")]
    pub id: ::prost::alloc::string::String,
    /// The Dataset that the artifact belongs to
    #[prost(message, optional, tag="2")]
    pub dataset: ::core::option::Option<DatasetId>,
    /// A list of data that is associated with the artifact
    #[prost(message, repeated, tag="3")]
    pub data: ::prost::alloc::vec::Vec<ArtifactData>,
    /// Free-form metadata associated with the artifact
    #[prost(message, optional, tag="4")]
    pub metadata: ::core::option::Option<Metadata>,
    #[prost(message, repeated, tag="5")]
    pub partitions: ::prost::alloc::vec::Vec<Partition>,
    #[prost(message, repeated, tag="6")]
    pub tags: ::prost::alloc::vec::Vec<Tag>,
    /// creation timestamp of artifact, autogenerated by service
    #[prost(message, optional, tag="7")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
}
///
/// ArtifactData that belongs to an artifact
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ArtifactData {
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    #[prost(message, optional, tag="2")]
    pub value: ::core::option::Option<super::flyteidl::core::Literal>,
}
///
/// Tag message that is unique to a Dataset. It is associated to a single artifact and
/// can be retrieved by name later.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Tag {
    /// Name of tag
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    /// The tagged artifact
    #[prost(string, tag="2")]
    pub artifact_id: ::prost::alloc::string::String,
    /// The Dataset that this tag belongs to
    #[prost(message, optional, tag="3")]
    pub dataset: ::core::option::Option<DatasetId>,
}
///
/// Metadata representation for artifacts and datasets
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Metadata {
    /// key map is a dictionary of key/val strings that represent metadata
    #[prost(map="string, string", tag="1")]
    pub key_map: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
/// Filter expression that is composed of a combination of single filters
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FilterExpression {
    #[prost(message, repeated, tag="1")]
    pub filters: ::prost::alloc::vec::Vec<SinglePropertyFilter>,
}
/// A single property to filter on.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SinglePropertyFilter {
    /// field 10 in case we add more entities to query
    #[prost(enumeration="single_property_filter::ComparisonOperator", tag="10")]
    pub operator: i32,
    #[prost(oneof="single_property_filter::PropertyFilter", tags="1, 2, 3, 4")]
    pub property_filter: ::core::option::Option<single_property_filter::PropertyFilter>,
}
/// Nested message and enum types in `SinglePropertyFilter`.
pub mod single_property_filter {
    /// as use-cases come up we can add more operators, ex: gte, like, not eq etc.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum ComparisonOperator {
        Equals = 0,
    }
    impl ComparisonOperator {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                ComparisonOperator::Equals => "EQUALS",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "EQUALS" => Some(Self::Equals),
                _ => None,
            }
        }
    }
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum PropertyFilter {
        #[prost(message, tag="1")]
        TagFilter(super::TagPropertyFilter),
        #[prost(message, tag="2")]
        PartitionFilter(super::PartitionPropertyFilter),
        #[prost(message, tag="3")]
        ArtifactFilter(super::ArtifactPropertyFilter),
        #[prost(message, tag="4")]
        DatasetFilter(super::DatasetPropertyFilter),
    }
}
/// Artifact properties we can filter by
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ArtifactPropertyFilter {
    /// oneof because we can add more properties in the future
    #[prost(oneof="artifact_property_filter::Property", tags="1")]
    pub property: ::core::option::Option<artifact_property_filter::Property>,
}
/// Nested message and enum types in `ArtifactPropertyFilter`.
pub mod artifact_property_filter {
    /// oneof because we can add more properties in the future
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Property {
        #[prost(string, tag="1")]
        ArtifactId(::prost::alloc::string::String),
    }
}
/// Tag properties we can filter by
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TagPropertyFilter {
    #[prost(oneof="tag_property_filter::Property", tags="1")]
    pub property: ::core::option::Option<tag_property_filter::Property>,
}
/// Nested message and enum types in `TagPropertyFilter`.
pub mod tag_property_filter {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Property {
        #[prost(string, tag="1")]
        TagName(::prost::alloc::string::String),
    }
}
/// Partition properties we can filter by
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PartitionPropertyFilter {
    #[prost(oneof="partition_property_filter::Property", tags="1")]
    pub property: ::core::option::Option<partition_property_filter::Property>,
}
/// Nested message and enum types in `PartitionPropertyFilter`.
pub mod partition_property_filter {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Property {
        #[prost(message, tag="1")]
        KeyVal(super::KeyValuePair),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct KeyValuePair {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub value: ::prost::alloc::string::String,
}
/// Dataset properties we can filter by
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DatasetPropertyFilter {
    #[prost(oneof="dataset_property_filter::Property", tags="1, 2, 3, 4")]
    pub property: ::core::option::Option<dataset_property_filter::Property>,
}
/// Nested message and enum types in `DatasetPropertyFilter`.
pub mod dataset_property_filter {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Property {
        #[prost(string, tag="1")]
        Project(::prost::alloc::string::String),
        #[prost(string, tag="2")]
        Name(::prost::alloc::string::String),
        #[prost(string, tag="3")]
        Domain(::prost::alloc::string::String),
        #[prost(string, tag="4")]
        Version(::prost::alloc::string::String),
    }
}
/// Pagination options for making list requests
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PaginationOptions {
    /// the max number of results to return
    #[prost(uint32, tag="1")]
    pub limit: u32,
    /// the token to pass to fetch the next page
    #[prost(string, tag="2")]
    pub token: ::prost::alloc::string::String,
    /// the property that we want to sort the results by
    #[prost(enumeration="pagination_options::SortKey", tag="3")]
    pub sort_key: i32,
    /// the sort order of the results
    #[prost(enumeration="pagination_options::SortOrder", tag="4")]
    pub sort_order: i32,
}
/// Nested message and enum types in `PaginationOptions`.
pub mod pagination_options {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum SortOrder {
        Descending = 0,
        Ascending = 1,
    }
    impl SortOrder {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                SortOrder::Descending => "DESCENDING",
                SortOrder::Ascending => "ASCENDING",
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
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum SortKey {
        CreationTime = 0,
    }
    impl SortKey {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                SortKey::CreationTime => "CREATION_TIME",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "CREATION_TIME" => Some(Self::CreationTime),
                _ => None,
            }
        }
    }
}
// @@protoc_insertion_point(module)
