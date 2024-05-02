// @generated
///
/// Additional metadata as key-value pairs
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct KeyMapMetadata {
    /// Additional metadata as key-value pairs
    #[prost(map="string, string", tag="1")]
    pub values: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
///
/// Metadata for cached outputs, including the source identifier and timestamps.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Metadata {
    /// Source task or workflow identifier
    #[prost(message, optional, tag="1")]
    pub source_identifier: ::core::option::Option<super::core::Identifier>,
    /// Additional metadata as key-value pairs
    #[prost(message, optional, tag="2")]
    pub key_map: ::core::option::Option<KeyMapMetadata>,
    /// Creation timestamp
    #[prost(message, optional, tag="3")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    /// Last update timestamp
    #[prost(message, optional, tag="4")]
    pub last_updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
///
/// Represents cached output, either as literals or an URI, with associated metadata.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CachedOutput {
    /// Associated metadata
    #[prost(message, optional, tag="3")]
    pub metadata: ::core::option::Option<Metadata>,
    #[prost(oneof="cached_output::Output", tags="1, 2")]
    pub output: ::core::option::Option<cached_output::Output>,
}
/// Nested message and enum types in `CachedOutput`.
pub mod cached_output {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Output {
        /// Output literals
        #[prost(message, tag="1")]
        OutputLiterals(super::super::core::LiteralMap),
        /// URI to output data
        #[prost(string, tag="2")]
        OutputUri(::prost::alloc::string::String),
    }
}
///
/// Request to retrieve cached data by key.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetCacheRequest {
    /// Cache key
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
}
///
/// Response with cached data for a given key.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetCacheResponse {
    /// Cached output
    #[prost(message, optional, tag="1")]
    pub output: ::core::option::Option<CachedOutput>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct OverwriteOutput {
    /// Overwrite flag
    #[prost(bool, tag="1")]
    pub overwrite: bool,
    /// Delete existing blob
    #[prost(bool, tag="2")]
    pub delete_blob: bool,
}
///
/// Request to store/update cached data by key.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PutCacheRequest {
    /// Cache key
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    /// Output to cache
    #[prost(message, optional, tag="2")]
    pub output: ::core::option::Option<CachedOutput>,
    /// Overwrite flag if exists
    #[prost(message, optional, tag="3")]
    pub overwrite: ::core::option::Option<OverwriteOutput>,
}
///
/// Response message of cache store/update operation.
///
/// Empty, success indicated by no errors
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PutCacheResponse {
}
///
/// Request to delete cached data by key.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeleteCacheRequest {
    /// Cache key
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
}
///
/// Response message of cache deletion operation.
///
/// Empty, success indicated by no errors
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeleteCacheResponse {
}
/// A reservation including owner, heartbeat interval, expiration timestamp, and various metadata.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Reservation {
    /// The unique ID for the reservation - same as the cache key
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    /// The unique ID of the owner for the reservation
    #[prost(string, tag="2")]
    pub owner_id: ::prost::alloc::string::String,
    /// Requested reservation extension heartbeat interval
    #[prost(message, optional, tag="3")]
    pub heartbeat_interval: ::core::option::Option<::prost_types::Duration>,
    /// Expiration timestamp of this reservation
    #[prost(message, optional, tag="4")]
    pub expires_at: ::core::option::Option<::prost_types::Timestamp>,
}
///
/// Request to get or extend a reservation for a cache key
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetOrExtendReservationRequest {
    /// The unique ID for the reservation - same as the cache key
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    /// The unique ID of the owner for the reservation
    #[prost(string, tag="2")]
    pub owner_id: ::prost::alloc::string::String,
    /// Requested reservation extension heartbeat interval
    #[prost(message, optional, tag="3")]
    pub heartbeat_interval: ::core::option::Option<::prost_types::Duration>,
}
///
/// Request to get or extend a reservation for a cache key
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetOrExtendReservationResponse {
    /// The reservation that was created or extended
    #[prost(message, optional, tag="1")]
    pub reservation: ::core::option::Option<Reservation>,
}
///
/// Request to release the reservation for a cache key
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ReleaseReservationRequest {
    /// The unique ID for the reservation - same as the cache key
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    /// The unique ID of the owner for the reservation
    #[prost(string, tag="2")]
    pub owner_id: ::prost::alloc::string::String,
}
///
/// Response message of release reservation operation.
///
/// Empty, success indicated by no errors
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ReleaseReservationResponse {
}
// @@protoc_insertion_point(module)
