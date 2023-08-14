// @generated
/// Defines schema columns and types to strongly type-validate schemas interoperability.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SchemaType {
    /// A list of ordered columns this schema comprises of.
    #[prost(message, repeated, tag="3")]
    pub columns: ::prost::alloc::vec::Vec<schema_type::SchemaColumn>,
}
/// Nested message and enum types in `SchemaType`.
pub mod schema_type {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
    pub struct SchemaColumn {
        /// A unique name -within the schema type- for the column
        #[prost(string, tag="1")]
        pub name: ::prost::alloc::string::String,
        /// The column type. This allows a limited set of types currently.
        #[prost(enumeration="schema_column::SchemaColumnType", tag="2")]
        pub r#type: i32,
    }
    /// Nested message and enum types in `SchemaColumn`.
    pub mod schema_column {
        #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
        #[repr(i32)]
        pub enum SchemaColumnType {
            Integer = 0,
            Float = 1,
            String = 2,
            Boolean = 3,
            Datetime = 4,
            Duration = 5,
        }
        impl SchemaColumnType {
            /// String value of the enum field names used in the ProtoBuf definition.
            ///
            /// The values are not transformed in any way and thus are considered stable
            /// (if the ProtoBuf definition does not change) and safe for programmatic use.
            pub fn as_str_name(&self) -> &'static str {
                match self {
                    SchemaColumnType::Integer => "INTEGER",
                    SchemaColumnType::Float => "FLOAT",
                    SchemaColumnType::String => "STRING",
                    SchemaColumnType::Boolean => "BOOLEAN",
                    SchemaColumnType::Datetime => "DATETIME",
                    SchemaColumnType::Duration => "DURATION",
                }
            }
            /// Creates an enum from field names used in the ProtoBuf definition.
            pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
                match value {
                    "INTEGER" => Some(Self::Integer),
                    "FLOAT" => Some(Self::Float),
                    "STRING" => Some(Self::String),
                    "BOOLEAN" => Some(Self::Boolean),
                    "DATETIME" => Some(Self::Datetime),
                    "DURATION" => Some(Self::Duration),
                    _ => None,
                }
            }
        }
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StructuredDatasetType {
    /// A list of ordered columns this schema comprises of.
    #[prost(message, repeated, tag="1")]
    pub columns: ::prost::alloc::vec::Vec<structured_dataset_type::DatasetColumn>,
    /// This is the storage format, the format of the bits at rest
    /// parquet, feather, csv, etc.
    /// For two types to be compatible, the format will need to be an exact match.
    #[prost(string, tag="2")]
    pub format: ::prost::alloc::string::String,
    /// This is a string representing the type that the bytes in external_schema_bytes are formatted in.
    /// This is an optional field that will not be used for type checking.
    #[prost(string, tag="3")]
    pub external_schema_type: ::prost::alloc::string::String,
    /// The serialized bytes of a third-party schema library like Arrow.
    /// This is an optional field that will not be used for type checking.
    #[prost(bytes="vec", tag="4")]
    pub external_schema_bytes: ::prost::alloc::vec::Vec<u8>,
}
/// Nested message and enum types in `StructuredDatasetType`.
pub mod structured_dataset_type {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
    pub struct DatasetColumn {
        /// A unique name within the schema type for the column.
        #[prost(string, tag="1")]
        pub name: ::prost::alloc::string::String,
        /// The column type.
        #[prost(message, optional, tag="2")]
        pub literal_type: ::core::option::Option<super::LiteralType>,
    }
}
/// Defines type behavior for blob objects
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BlobType {
    /// Format can be a free form string understood by SDK/UI etc like
    /// csv, parquet etc
    #[prost(string, tag="1")]
    pub format: ::prost::alloc::string::String,
    #[prost(enumeration="blob_type::BlobDimensionality", tag="2")]
    pub dimensionality: i32,
}
/// Nested message and enum types in `BlobType`.
pub mod blob_type {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum BlobDimensionality {
        Single = 0,
        Multipart = 1,
    }
    impl BlobDimensionality {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                BlobDimensionality::Single => "SINGLE",
                BlobDimensionality::Multipart => "MULTIPART",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "SINGLE" => Some(Self::Single),
                "MULTIPART" => Some(Self::Multipart),
                _ => None,
            }
        }
    }
}
/// Enables declaring enum types, with predefined string values
/// For len(values) > 0, the first value in the ordered list is regarded as the default value. If you wish
/// To provide no defaults, make the first value as undefined.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EnumType {
    /// Predefined set of enum values.
    #[prost(string, repeated, tag="1")]
    pub values: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// Defines a tagged union type, also known as a variant (and formally as the sum type).
///
/// A sum type S is defined by a sequence of types (A, B, C, ...), each tagged by a string tag
/// A value of type S is constructed from a value of any of the variant types. The specific choice of type is recorded by
/// storing the varaint's tag with the literal value and can be examined in runtime.
///
/// Type S is typically written as
/// S := Apple A | Banana B | Cantaloupe C | ...
///
/// Notably, a nullable (optional) type is a sum type between some type X and the singleton type representing a null-value:
/// Optional X := X | Null
///
/// See also: <https://en.wikipedia.org/wiki/Tagged_union>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnionType {
    /// Predefined set of variants in union.
    #[prost(message, repeated, tag="1")]
    pub variants: ::prost::alloc::vec::Vec<LiteralType>,
}
/// Hints to improve type matching
/// e.g. allows distinguishing output from custom type transformers
/// even if the underlying IDL serialization matches.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TypeStructure {
    /// Must exactly match for types to be castable
    #[prost(string, tag="1")]
    pub tag: ::prost::alloc::string::String,
}
/// TypeAnnotation encapsulates registration time information about a type. This can be used for various control-plane operations. TypeAnnotation will not be available at runtime when a task runs.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TypeAnnotation {
    /// A arbitrary JSON payload to describe a type.
    #[prost(message, optional, tag="1")]
    pub annotations: ::core::option::Option<::prost_types::Struct>,
}
/// Defines a strong type to allow type checking between interfaces.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LiteralType {
    /// This field contains type metadata that is descriptive of the type, but is NOT considered in type-checking.  This might be used by
    /// consumers to identify special behavior or display extended information for the type.
    #[prost(message, optional, tag="6")]
    pub metadata: ::core::option::Option<::prost_types::Struct>,
    /// This field contains arbitrary data that might have special semantic
    /// meaning for the client but does not effect internal flyte behavior.
    #[prost(message, optional, tag="9")]
    pub annotation: ::core::option::Option<TypeAnnotation>,
    /// Hints to improve type matching.
    #[prost(message, optional, tag="11")]
    pub structure: ::core::option::Option<TypeStructure>,
    #[prost(oneof="literal_type::Type", tags="1, 2, 3, 4, 5, 7, 8, 10")]
    pub r#type: ::core::option::Option<literal_type::Type>,
}
/// Nested message and enum types in `LiteralType`.
pub mod literal_type {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Type {
        /// A simple type that can be compared one-to-one with another.
        #[prost(enumeration="super::SimpleType", tag="1")]
        Simple(i32),
        /// A complex type that requires matching of inner fields.
        #[prost(message, tag="2")]
        Schema(super::SchemaType),
        /// Defines the type of the value of a collection. Only homogeneous collections are allowed.
        #[prost(message, tag="3")]
        CollectionType(::prost::alloc::boxed::Box<super::LiteralType>),
        /// Defines the type of the value of a map type. The type of the key is always a string.
        #[prost(message, tag="4")]
        MapValueType(::prost::alloc::boxed::Box<super::LiteralType>),
        /// A blob might have specialized implementation details depending on associated metadata.
        #[prost(message, tag="5")]
        Blob(super::BlobType),
        /// Defines an enum with pre-defined string values.
        #[prost(message, tag="7")]
        EnumType(super::EnumType),
        /// Generalized schema support
        #[prost(message, tag="8")]
        StructuredDatasetType(super::StructuredDatasetType),
        /// Defines an union type with pre-defined LiteralTypes.
        #[prost(message, tag="10")]
        UnionType(super::UnionType),
    }
}
/// A reference to an output produced by a node. The type can be retrieved -and validated- from
/// the underlying interface of the node.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct OutputReference {
    /// Node id must exist at the graph layer.
    #[prost(string, tag="1")]
    pub node_id: ::prost::alloc::string::String,
    /// Variable name must refer to an output variable for the node.
    #[prost(string, tag="2")]
    pub var: ::prost::alloc::string::String,
}
/// Represents an error thrown from a node.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Error {
    /// The node id that threw the error.
    #[prost(string, tag="1")]
    pub failed_node_id: ::prost::alloc::string::String,
    /// Error message thrown.
    #[prost(string, tag="2")]
    pub message: ::prost::alloc::string::String,
}
/// Define a set of simple types.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum SimpleType {
    None = 0,
    Integer = 1,
    Float = 2,
    String = 3,
    Boolean = 4,
    Datetime = 5,
    Duration = 6,
    Binary = 7,
    Error = 8,
    Struct = 9,
}
impl SimpleType {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            SimpleType::None => "NONE",
            SimpleType::Integer => "INTEGER",
            SimpleType::Float => "FLOAT",
            SimpleType::String => "STRING",
            SimpleType::Boolean => "BOOLEAN",
            SimpleType::Datetime => "DATETIME",
            SimpleType::Duration => "DURATION",
            SimpleType::Binary => "BINARY",
            SimpleType::Error => "ERROR",
            SimpleType::Struct => "STRUCT",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "NONE" => Some(Self::None),
            "INTEGER" => Some(Self::Integer),
            "FLOAT" => Some(Self::Float),
            "STRING" => Some(Self::String),
            "BOOLEAN" => Some(Self::Boolean),
            "DATETIME" => Some(Self::Datetime),
            "DURATION" => Some(Self::Duration),
            "BINARY" => Some(Self::Binary),
            "ERROR" => Some(Self::Error),
            "STRUCT" => Some(Self::Struct),
            _ => None,
        }
    }
}
/// Primitive Types
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Primitive {
    /// Defines one of simple primitive types. These types will get translated into different programming languages as
    /// described in <https://developers.google.com/protocol-buffers/docs/proto#scalar.>
    #[prost(oneof="primitive::Value", tags="1, 2, 3, 4, 5, 6")]
    pub value: ::core::option::Option<primitive::Value>,
}
/// Nested message and enum types in `Primitive`.
pub mod primitive {
    /// Defines one of simple primitive types. These types will get translated into different programming languages as
    /// described in <https://developers.google.com/protocol-buffers/docs/proto#scalar.>
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Value {
        #[prost(int64, tag="1")]
        Integer(i64),
        #[prost(double, tag="2")]
        FloatValue(f64),
        #[prost(string, tag="3")]
        StringValue(::prost::alloc::string::String),
        #[prost(bool, tag="4")]
        Boolean(bool),
        #[prost(message, tag="5")]
        Datetime(::prost_types::Timestamp),
        #[prost(message, tag="6")]
        Duration(::prost_types::Duration),
    }
}
/// Used to denote a nil/null/None assignment to a scalar value. The underlying LiteralType for Void is intentionally
/// undefined since it can be assigned to a scalar of any LiteralType.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Void {
}
/// Refers to an offloaded set of files. It encapsulates the type of the store and a unique uri for where the data is.
/// There are no restrictions on how the uri is formatted since it will depend on how to interact with the store.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Blob {
    #[prost(message, optional, tag="1")]
    pub metadata: ::core::option::Option<BlobMetadata>,
    #[prost(string, tag="3")]
    pub uri: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BlobMetadata {
    #[prost(message, optional, tag="1")]
    pub r#type: ::core::option::Option<BlobType>,
}
/// A simple byte array with a tag to help different parts of the system communicate about what is in the byte array.
/// It's strongly advisable that consumers of this type define a unique tag and validate the tag before parsing the data.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Binary {
    #[prost(bytes="vec", tag="1")]
    pub value: ::prost::alloc::vec::Vec<u8>,
    #[prost(string, tag="2")]
    pub tag: ::prost::alloc::string::String,
}
/// A strongly typed schema that defines the interface of data retrieved from the underlying storage medium.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Schema {
    #[prost(string, tag="1")]
    pub uri: ::prost::alloc::string::String,
    #[prost(message, optional, tag="3")]
    pub r#type: ::core::option::Option<SchemaType>,
}
/// The runtime representation of a tagged union value. See `UnionType` for more details.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Union {
    #[prost(message, optional, boxed, tag="1")]
    pub value: ::core::option::Option<::prost::alloc::boxed::Box<Literal>>,
    #[prost(message, optional, tag="2")]
    pub r#type: ::core::option::Option<LiteralType>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StructuredDatasetMetadata {
    /// Bundle the type information along with the literal.
    /// This is here because StructuredDatasets can often be more defined at run time than at compile time.
    /// That is, at compile time you might only declare a task to return a pandas dataframe or a StructuredDataset,
    /// without any column information, but at run time, you might have that column information.
    /// flytekit python will copy this type information into the literal, from the type information, if not provided by
    /// the various plugins (encoders).
    /// Since this field is run time generated, it's not used for any type checking.
    #[prost(message, optional, tag="1")]
    pub structured_dataset_type: ::core::option::Option<StructuredDatasetType>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StructuredDataset {
    /// String location uniquely identifying where the data is.
    /// Should start with the storage location (e.g. s3://, gs://, bq://, etc.)
    #[prost(string, tag="1")]
    pub uri: ::prost::alloc::string::String,
    #[prost(message, optional, tag="2")]
    pub metadata: ::core::option::Option<StructuredDatasetMetadata>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Scalar {
    #[prost(oneof="scalar::Value", tags="1, 2, 3, 4, 5, 6, 7, 8, 9")]
    pub value: ::core::option::Option<scalar::Value>,
}
/// Nested message and enum types in `Scalar`.
pub mod scalar {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Value {
        #[prost(message, tag="1")]
        Primitive(super::Primitive),
        #[prost(message, tag="2")]
        Blob(super::Blob),
        #[prost(message, tag="3")]
        Binary(super::Binary),
        #[prost(message, tag="4")]
        Schema(super::Schema),
        #[prost(message, tag="5")]
        NoneType(super::Void),
        #[prost(message, tag="6")]
        Error(super::Error),
        #[prost(message, tag="7")]
        Generic(::prost_types::Struct),
        #[prost(message, tag="8")]
        StructuredDataset(super::StructuredDataset),
        #[prost(message, tag="9")]
        Union(::prost::alloc::boxed::Box<super::Union>),
    }
}
/// A simple value. This supports any level of nesting (e.g. array of array of array of Blobs) as well as simple primitives.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Literal {
    /// A hash representing this literal.
    /// This is used for caching purposes. For more details refer to RFC 1893
    /// (<https://github.com/flyteorg/flyte/blob/master/rfc/system/1893-caching-of-offloaded-objects.md>)
    #[prost(string, tag="4")]
    pub hash: ::prost::alloc::string::String,
    #[prost(oneof="literal::Value", tags="1, 2, 3")]
    pub value: ::core::option::Option<literal::Value>,
}
/// Nested message and enum types in `Literal`.
pub mod literal {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Value {
        /// A simple value.
        #[prost(message, tag="1")]
        Scalar(::prost::alloc::boxed::Box<super::Scalar>),
        /// A collection of literals to allow nesting.
        #[prost(message, tag="2")]
        Collection(super::LiteralCollection),
        /// A map of strings to literals.
        #[prost(message, tag="3")]
        Map(super::LiteralMap),
    }
}
/// A collection of literals. This is a workaround since oneofs in proto messages cannot contain a repeated field.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LiteralCollection {
    #[prost(message, repeated, tag="1")]
    pub literals: ::prost::alloc::vec::Vec<Literal>,
}
/// A map of literals. This is a workaround since oneofs in proto messages cannot contain a repeated field.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LiteralMap {
    #[prost(map="string, message", tag="1")]
    pub literals: ::std::collections::HashMap<::prost::alloc::string::String, Literal>,
}
/// A collection of BindingData items.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BindingDataCollection {
    #[prost(message, repeated, tag="1")]
    pub bindings: ::prost::alloc::vec::Vec<BindingData>,
}
/// A map of BindingData items.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BindingDataMap {
    #[prost(map="string, message", tag="1")]
    pub bindings: ::std::collections::HashMap<::prost::alloc::string::String, BindingData>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnionInfo {
    #[prost(message, optional, tag="1")]
    pub target_type: ::core::option::Option<LiteralType>,
}
/// Specifies either a simple value or a reference to another output.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BindingData {
    #[prost(message, optional, tag="5")]
    pub union: ::core::option::Option<UnionInfo>,
    #[prost(oneof="binding_data::Value", tags="1, 2, 3, 4")]
    pub value: ::core::option::Option<binding_data::Value>,
}
/// Nested message and enum types in `BindingData`.
pub mod binding_data {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Value {
        /// A simple scalar value.
        #[prost(message, tag="1")]
        Scalar(super::Scalar),
        /// A collection of binding data. This allows nesting of binding data to any number
        /// of levels.
        #[prost(message, tag="2")]
        Collection(super::BindingDataCollection),
        /// References an output promised by another node.
        #[prost(message, tag="3")]
        Promise(super::OutputReference),
        /// A map of bindings. The key is always a string.
        #[prost(message, tag="4")]
        Map(super::BindingDataMap),
    }
}
/// An input/output binding of a variable to either static value or a node output.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Binding {
    /// Variable name must match an input/output variable of the node.
    #[prost(string, tag="1")]
    pub var: ::prost::alloc::string::String,
    /// Data to use to bind this variable.
    #[prost(message, optional, tag="2")]
    pub binding: ::core::option::Option<BindingData>,
}
/// A generic key value pair.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct KeyValuePair {
    /// required.
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    /// +optional.
    #[prost(string, tag="2")]
    pub value: ::prost::alloc::string::String,
}
/// Retry strategy associated with an executable unit.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RetryStrategy {
    /// Number of retries. Retries will be consumed when the job fails with a recoverable error.
    /// The number of retries must be less than or equals to 10.
    #[prost(uint32, tag="5")]
    pub retries: u32,
}
/// Encapsulation of fields that uniquely identifies a Flyte resource.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Identifier {
    /// Identifies the specific type of resource that this identifier corresponds to.
    #[prost(enumeration="ResourceType", tag="1")]
    pub resource_type: i32,
    /// Name of the project the resource belongs to.
    #[prost(string, tag="2")]
    pub project: ::prost::alloc::string::String,
    /// Name of the domain the resource belongs to.
    /// A domain can be considered as a subset within a specific project.
    #[prost(string, tag="3")]
    pub domain: ::prost::alloc::string::String,
    /// User provided value for the resource.
    #[prost(string, tag="4")]
    pub name: ::prost::alloc::string::String,
    /// Specific version of the resource.
    #[prost(string, tag="5")]
    pub version: ::prost::alloc::string::String,
}
/// Encapsulation of fields that uniquely identifies a Flyte workflow execution
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecutionIdentifier {
    /// Name of the project the resource belongs to.
    #[prost(string, tag="1")]
    pub project: ::prost::alloc::string::String,
    /// Name of the domain the resource belongs to.
    /// A domain can be considered as a subset within a specific project.
    #[prost(string, tag="2")]
    pub domain: ::prost::alloc::string::String,
    /// User or system provided value for the resource.
    #[prost(string, tag="4")]
    pub name: ::prost::alloc::string::String,
}
/// Encapsulation of fields that identify a Flyte node execution entity.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecutionIdentifier {
    #[prost(string, tag="1")]
    pub node_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag="2")]
    pub execution_id: ::core::option::Option<WorkflowExecutionIdentifier>,
}
/// Encapsulation of fields that identify a Flyte task execution entity.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecutionIdentifier {
    #[prost(message, optional, tag="1")]
    pub task_id: ::core::option::Option<Identifier>,
    #[prost(message, optional, tag="2")]
    pub node_execution_id: ::core::option::Option<NodeExecutionIdentifier>,
    #[prost(uint32, tag="3")]
    pub retry_attempt: u32,
}
/// Encapsulation of fields the uniquely identify a signal.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignalIdentifier {
    /// Unique identifier for a signal.
    #[prost(string, tag="1")]
    pub signal_id: ::prost::alloc::string::String,
    /// Identifies the Flyte workflow execution this signal belongs to.
    #[prost(message, optional, tag="2")]
    pub execution_id: ::core::option::Option<WorkflowExecutionIdentifier>,
}
/// Indicates a resource type within Flyte.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum ResourceType {
    Unspecified = 0,
    Task = 1,
    Workflow = 2,
    LaunchPlan = 3,
    /// A dataset represents an entity modeled in Flyte DataCatalog. A Dataset is also a versioned entity and can be a compilation of multiple individual objects.
    /// Eventually all Catalog objects should be modeled similar to Flyte Objects. The Dataset entities makes it possible for the UI  and CLI to act on the objects 
    /// in a similar manner to other Flyte objects
    Dataset = 4,
}
impl ResourceType {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            ResourceType::Unspecified => "UNSPECIFIED",
            ResourceType::Task => "TASK",
            ResourceType::Workflow => "WORKFLOW",
            ResourceType::LaunchPlan => "LAUNCH_PLAN",
            ResourceType::Dataset => "DATASET",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "UNSPECIFIED" => Some(Self::Unspecified),
            "TASK" => Some(Self::Task),
            "WORKFLOW" => Some(Self::Workflow),
            "LAUNCH_PLAN" => Some(Self::LaunchPlan),
            "DATASET" => Some(Self::Dataset),
            _ => None,
        }
    }
}
/// Defines a strongly typed variable.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Variable {
    /// Variable literal type.
    #[prost(message, optional, tag="1")]
    pub r#type: ::core::option::Option<LiteralType>,
    /// +optional string describing input variable
    #[prost(string, tag="2")]
    pub description: ::prost::alloc::string::String,
}
/// A map of Variables
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct VariableMap {
    /// Defines a map of variable names to variables.
    #[prost(map="string, message", tag="1")]
    pub variables: ::std::collections::HashMap<::prost::alloc::string::String, Variable>,
}
/// Defines strongly typed inputs and outputs.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TypedInterface {
    #[prost(message, optional, tag="1")]
    pub inputs: ::core::option::Option<VariableMap>,
    #[prost(message, optional, tag="2")]
    pub outputs: ::core::option::Option<VariableMap>,
}
/// A parameter is used as input to a launch plan and has
/// the special ability to have a default value or mark itself as required.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Parameter {
    /// +required Variable. Defines the type of the variable backing this parameter.
    #[prost(message, optional, tag="1")]
    pub var: ::core::option::Option<Variable>,
    /// +optional
    #[prost(oneof="parameter::Behavior", tags="2, 3")]
    pub behavior: ::core::option::Option<parameter::Behavior>,
}
/// Nested message and enum types in `Parameter`.
pub mod parameter {
    /// +optional
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Behavior {
        /// Defines a default value that has to match the variable type defined.
        #[prost(message, tag="2")]
        Default(super::Literal),
        /// +optional, is this value required to be filled.
        #[prost(bool, tag="3")]
        Required(bool),
    }
}
/// A map of Parameters.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ParameterMap {
    /// Defines a map of parameter names to parameters.
    #[prost(map="string, message", tag="1")]
    pub parameters: ::std::collections::HashMap<::prost::alloc::string::String, Parameter>,
}
/// Secret encapsulates information about the secret a task needs to proceed. An environment variable
/// FLYTE_SECRETS_ENV_PREFIX will be passed to indicate the prefix of the environment variables that will be present if
/// secrets are passed through environment variables.
/// FLYTE_SECRETS_DEFAULT_DIR will be passed to indicate the prefix of the path where secrets will be mounted if secrets
/// are passed through file mounts.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Secret {
    /// The name of the secret group where to find the key referenced below. For K8s secrets, this should be the name of
    /// the v1/secret object. For Confidant, this should be the Credential name. For Vault, this should be the secret name.
    /// For AWS Secret Manager, this should be the name of the secret.
    /// +required
    #[prost(string, tag="1")]
    pub group: ::prost::alloc::string::String,
    /// The group version to fetch. This is not supported in all secret management systems. It'll be ignored for the ones
    /// that do not support it.
    /// +optional
    #[prost(string, tag="2")]
    pub group_version: ::prost::alloc::string::String,
    /// The name of the secret to mount. This has to match an existing secret in the system. It's up to the implementation
    /// of the secret management system to require case sensitivity. For K8s secrets, Confidant and Vault, this should
    /// match one of the keys inside the secret. For AWS Secret Manager, it's ignored.
    /// +optional
    #[prost(string, tag="3")]
    pub key: ::prost::alloc::string::String,
    /// mount_requirement is optional. Indicates where the secret has to be mounted. If provided, the execution will fail
    /// if the underlying key management system cannot satisfy that requirement. If not provided, the default location
    /// will depend on the key management system.
    /// +optional
    #[prost(enumeration="secret::MountType", tag="4")]
    pub mount_requirement: i32,
}
/// Nested message and enum types in `Secret`.
pub mod secret {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum MountType {
        /// Default case, indicates the client can tolerate either mounting options.
        Any = 0,
        /// ENV_VAR indicates the secret needs to be mounted as an environment variable.
        EnvVar = 1,
        /// FILE indicates the secret needs to be mounted as a file.
        File = 2,
    }
    impl MountType {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                MountType::Any => "ANY",
                MountType::EnvVar => "ENV_VAR",
                MountType::File => "FILE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "ANY" => Some(Self::Any),
                "ENV_VAR" => Some(Self::EnvVar),
                "FILE" => Some(Self::File),
                _ => None,
            }
        }
    }
}
/// OAuth2Client encapsulates OAuth2 Client Credentials to be used when making calls on behalf of that task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct OAuth2Client {
    /// client_id is the public id for the client to use. The system will not perform any pre-auth validation that the
    /// secret requested matches the client_id indicated here.
    /// +required
    #[prost(string, tag="1")]
    pub client_id: ::prost::alloc::string::String,
    /// client_secret is a reference to the secret used to authenticate the OAuth2 client.
    /// +required
    #[prost(message, optional, tag="2")]
    pub client_secret: ::core::option::Option<Secret>,
}
/// Identity encapsulates the various security identities a task can run as. It's up to the underlying plugin to pick the
/// right identity for the execution environment.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Identity {
    /// iam_role references the fully qualified name of Identity & Access Management role to impersonate.
    #[prost(string, tag="1")]
    pub iam_role: ::prost::alloc::string::String,
    /// k8s_service_account references a kubernetes service account to impersonate.
    #[prost(string, tag="2")]
    pub k8s_service_account: ::prost::alloc::string::String,
    /// oauth2_client references an oauth2 client. Backend plugins can use this information to impersonate the client when
    /// making external calls.
    #[prost(message, optional, tag="3")]
    pub oauth2_client: ::core::option::Option<OAuth2Client>,
    /// execution_identity references the subject who makes the execution
    #[prost(string, tag="4")]
    pub execution_identity: ::prost::alloc::string::String,
}
/// OAuth2TokenRequest encapsulates information needed to request an OAuth2 token.
/// FLYTE_TOKENS_ENV_PREFIX will be passed to indicate the prefix of the environment variables that will be present if
/// tokens are passed through environment variables.
/// FLYTE_TOKENS_PATH_PREFIX will be passed to indicate the prefix of the path where secrets will be mounted if tokens
/// are passed through file mounts.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct OAuth2TokenRequest {
    /// name indicates a unique id for the token request within this task token requests. It'll be used as a suffix for
    /// environment variables and as a filename for mounting tokens as files.
    /// +required
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    /// type indicates the type of the request to make. Defaults to CLIENT_CREDENTIALS.
    /// +required
    #[prost(enumeration="o_auth2_token_request::Type", tag="2")]
    pub r#type: i32,
    /// client references the client_id/secret to use to request the OAuth2 token.
    /// +required
    #[prost(message, optional, tag="3")]
    pub client: ::core::option::Option<OAuth2Client>,
    /// idp_discovery_endpoint references the discovery endpoint used to retrieve token endpoint and other related
    /// information.
    /// +optional
    #[prost(string, tag="4")]
    pub idp_discovery_endpoint: ::prost::alloc::string::String,
    /// token_endpoint references the token issuance endpoint. If idp_discovery_endpoint is not provided, this parameter is
    /// mandatory.
    /// +optional
    #[prost(string, tag="5")]
    pub token_endpoint: ::prost::alloc::string::String,
}
/// Nested message and enum types in `OAuth2TokenRequest`.
pub mod o_auth2_token_request {
    /// Type of the token requested.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Type {
        /// CLIENT_CREDENTIALS indicates a 2-legged OAuth token requested using client credentials.
        ClientCredentials = 0,
    }
    impl Type {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Type::ClientCredentials => "CLIENT_CREDENTIALS",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "CLIENT_CREDENTIALS" => Some(Self::ClientCredentials),
                _ => None,
            }
        }
    }
}
/// SecurityContext holds security attributes that apply to tasks.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SecurityContext {
    /// run_as encapsulates the identity a pod should run as. If the task fills in multiple fields here, it'll be up to the
    /// backend plugin to choose the appropriate identity for the execution engine the task will run on.
    #[prost(message, optional, tag="1")]
    pub run_as: ::core::option::Option<Identity>,
    /// secrets indicate the list of secrets the task needs in order to proceed. Secrets will be mounted/passed to the
    /// pod as it starts. If the plugin responsible for kicking of the task will not run it on a flyte cluster (e.g. AWS
    /// Batch), it's the responsibility of the plugin to fetch the secret (which means propeller identity will need access
    /// to the secret) and to pass it to the remote execution engine.
    #[prost(message, repeated, tag="2")]
    pub secrets: ::prost::alloc::vec::Vec<Secret>,
    /// tokens indicate the list of token requests the task needs in order to proceed. Tokens will be mounted/passed to the
    /// pod as it starts. If the plugin responsible for kicking of the task will not run it on a flyte cluster (e.g. AWS
    /// Batch), it's the responsibility of the plugin to fetch the secret (which means propeller identity will need access
    /// to the secret) and to pass it to the remote execution engine.
    #[prost(message, repeated, tag="3")]
    pub tokens: ::prost::alloc::vec::Vec<OAuth2TokenRequest>,
}
/// A customizable interface to convey resources requested for a container. This can be interpreted differently for different
/// container engines.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Resources {
    /// The desired set of resources requested. ResourceNames must be unique within the list.
    #[prost(message, repeated, tag="1")]
    pub requests: ::prost::alloc::vec::Vec<resources::ResourceEntry>,
    /// Defines a set of bounds (e.g. min/max) within which the task can reliably run. ResourceNames must be unique
    /// within the list.
    #[prost(message, repeated, tag="2")]
    pub limits: ::prost::alloc::vec::Vec<resources::ResourceEntry>,
}
/// Nested message and enum types in `Resources`.
pub mod resources {
    /// Encapsulates a resource name and value.
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
    pub struct ResourceEntry {
        /// Resource name.
        #[prost(enumeration="ResourceName", tag="1")]
        pub name: i32,
        /// Value must be a valid k8s quantity. See
        /// <https://github.com/kubernetes/apimachinery/blob/master/pkg/api/resource/quantity.go#L30-L80>
        #[prost(string, tag="2")]
        pub value: ::prost::alloc::string::String,
    }
    /// Known resource names.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum ResourceName {
        Unknown = 0,
        Cpu = 1,
        Gpu = 2,
        Memory = 3,
        Storage = 4,
        /// For Kubernetes-based deployments, pods use ephemeral local storage for scratch space, caching, and for logs.
        EphemeralStorage = 5,
    }
    impl ResourceName {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                ResourceName::Unknown => "UNKNOWN",
                ResourceName::Cpu => "CPU",
                ResourceName::Gpu => "GPU",
                ResourceName::Memory => "MEMORY",
                ResourceName::Storage => "STORAGE",
                ResourceName::EphemeralStorage => "EPHEMERAL_STORAGE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNKNOWN" => Some(Self::Unknown),
                "CPU" => Some(Self::Cpu),
                "GPU" => Some(Self::Gpu),
                "MEMORY" => Some(Self::Memory),
                "STORAGE" => Some(Self::Storage),
                "EPHEMERAL_STORAGE" => Some(Self::EphemeralStorage),
                _ => None,
            }
        }
    }
}
/// Runtime information. This is loosely defined to allow for extensibility.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RuntimeMetadata {
    /// Type of runtime.
    #[prost(enumeration="runtime_metadata::RuntimeType", tag="1")]
    pub r#type: i32,
    /// Version of the runtime. All versions should be backward compatible. However, certain cases call for version
    /// checks to ensure tighter validation or setting expectations.
    #[prost(string, tag="2")]
    pub version: ::prost::alloc::string::String,
    /// +optional It can be used to provide extra information about the runtime (e.g. python, golang... etc.).
    #[prost(string, tag="3")]
    pub flavor: ::prost::alloc::string::String,
}
/// Nested message and enum types in `RuntimeMetadata`.
pub mod runtime_metadata {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum RuntimeType {
        Other = 0,
        FlyteSdk = 1,
    }
    impl RuntimeType {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                RuntimeType::Other => "OTHER",
                RuntimeType::FlyteSdk => "FLYTE_SDK",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "OTHER" => Some(Self::Other),
                "FLYTE_SDK" => Some(Self::FlyteSdk),
                _ => None,
            }
        }
    }
}
/// Task Metadata
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskMetadata {
    /// Indicates whether the system should attempt to lookup this task's output to avoid duplication of work.
    #[prost(bool, tag="1")]
    pub discoverable: bool,
    /// Runtime information about the task.
    #[prost(message, optional, tag="2")]
    pub runtime: ::core::option::Option<RuntimeMetadata>,
    /// The overall timeout of a task including user-triggered retries.
    #[prost(message, optional, tag="4")]
    pub timeout: ::core::option::Option<::prost_types::Duration>,
    /// Number of retries per task.
    #[prost(message, optional, tag="5")]
    pub retries: ::core::option::Option<RetryStrategy>,
    /// Indicates a logical version to apply to this task for the purpose of discovery.
    #[prost(string, tag="6")]
    pub discovery_version: ::prost::alloc::string::String,
    /// If set, this indicates that this task is deprecated.  This will enable owners of tasks to notify consumers
    /// of the ending of support for a given task.
    #[prost(string, tag="7")]
    pub deprecated_error_message: ::prost::alloc::string::String,
    /// Indicates whether the system should attempt to execute discoverable instances in serial to avoid duplicate work
    #[prost(bool, tag="9")]
    pub cache_serializable: bool,
    /// Indicates whether the task will generate a Deck URI when it finishes executing.
    #[prost(bool, tag="10")]
    pub generates_deck: bool,
    /// Arbitrary tags that allow users and the platform to store small but arbitrary labels
    #[prost(map="string, string", tag="11")]
    pub tags: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// pod_template_name is the unique name of a PodTemplate k8s resource to be used as the base configuration if this
    /// task creates a k8s Pod. If this value is set, the specified PodTemplate will be used instead of, but applied
    /// identically as, the default PodTemplate configured in FlytePropeller.
    #[prost(string, tag="12")]
    pub pod_template_name: ::prost::alloc::string::String,
    // For interruptible we will populate it at the node level but require it be part of TaskMetadata
    // for a user to set the value.
    // We are using oneof instead of bool because otherwise we would be unable to distinguish between value being
    // set by the user or defaulting to false.
    // The logic of handling precedence will be done as part of flytepropeller.

    /// Identify whether task is interruptible
    #[prost(oneof="task_metadata::InterruptibleValue", tags="8")]
    pub interruptible_value: ::core::option::Option<task_metadata::InterruptibleValue>,
}
/// Nested message and enum types in `TaskMetadata`.
pub mod task_metadata {
    // For interruptible we will populate it at the node level but require it be part of TaskMetadata
    // for a user to set the value.
    // We are using oneof instead of bool because otherwise we would be unable to distinguish between value being
    // set by the user or defaulting to false.
    // The logic of handling precedence will be done as part of flytepropeller.

    /// Identify whether task is interruptible
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum InterruptibleValue {
        #[prost(bool, tag="8")]
        Interruptible(bool),
    }
}
/// A Task structure that uniquely identifies a task in the system
/// Tasks are registered as a first step in the system.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskTemplate {
    /// Auto generated taskId by the system. Task Id uniquely identifies this task globally.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<Identifier>,
    /// A predefined yet extensible Task type identifier. This can be used to customize any of the components. If no
    /// extensions are provided in the system, Flyte will resolve the this task to its TaskCategory and default the
    /// implementation registered for the TaskCategory.
    #[prost(string, tag="2")]
    pub r#type: ::prost::alloc::string::String,
    /// Extra metadata about the task.
    #[prost(message, optional, tag="3")]
    pub metadata: ::core::option::Option<TaskMetadata>,
    /// A strongly typed interface for the task. This enables others to use this task within a workflow and guarantees
    /// compile-time validation of the workflow to avoid costly runtime failures.
    #[prost(message, optional, tag="4")]
    pub interface: ::core::option::Option<TypedInterface>,
    /// Custom data about the task. This is extensible to allow various plugins in the system.
    #[prost(message, optional, tag="5")]
    pub custom: ::core::option::Option<::prost_types::Struct>,
    /// This can be used to customize task handling at execution time for the same task type.
    #[prost(int32, tag="7")]
    pub task_type_version: i32,
    /// security_context encapsulates security attributes requested to run this task.
    #[prost(message, optional, tag="8")]
    pub security_context: ::core::option::Option<SecurityContext>,
    /// Metadata about the custom defined for this task. This is extensible to allow various plugins in the system
    /// to use as required.
    /// reserve the field numbers 1 through 15 for very frequently occurring message elements
    #[prost(map="string, string", tag="16")]
    pub config: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// Known target types that the system will guarantee plugins for. Custom SDK plugins are allowed to set these if needed.
    /// If no corresponding execution-layer plugins are found, the system will default to handling these using built-in
    /// handlers.
    #[prost(oneof="task_template::Target", tags="6, 17, 18")]
    pub target: ::core::option::Option<task_template::Target>,
}
/// Nested message and enum types in `TaskTemplate`.
pub mod task_template {
    /// Known target types that the system will guarantee plugins for. Custom SDK plugins are allowed to set these if needed.
    /// If no corresponding execution-layer plugins are found, the system will default to handling these using built-in
    /// handlers.
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Target {
        #[prost(message, tag="6")]
        Container(super::Container),
        #[prost(message, tag="17")]
        K8sPod(super::K8sPod),
        #[prost(message, tag="18")]
        Sql(super::Sql),
    }
}
// ----------------- First class Plugins

/// Defines port properties for a container.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContainerPort {
    /// Number of port to expose on the pod's IP address.
    /// This must be a valid port number, 0 < x < 65536.
    #[prost(uint32, tag="1")]
    pub container_port: u32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Container {
    /// Container image url. Eg: docker/redis:latest
    #[prost(string, tag="1")]
    pub image: ::prost::alloc::string::String,
    /// Command to be executed, if not provided, the default entrypoint in the container image will be used.
    #[prost(string, repeated, tag="2")]
    pub command: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// These will default to Flyte given paths. If provided, the system will not append known paths. If the task still
    /// needs flyte's inputs and outputs path, add $(FLYTE_INPUT_FILE), $(FLYTE_OUTPUT_FILE) wherever makes sense and the
    /// system will populate these before executing the container.
    #[prost(string, repeated, tag="3")]
    pub args: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// Container resources requirement as specified by the container engine.
    #[prost(message, optional, tag="4")]
    pub resources: ::core::option::Option<Resources>,
    /// Environment variables will be set as the container is starting up.
    #[prost(message, repeated, tag="5")]
    pub env: ::prost::alloc::vec::Vec<KeyValuePair>,
    /// Allows extra configs to be available for the container.
    /// TODO: elaborate on how configs will become available.
    /// Deprecated, please use TaskTemplate.config instead.
    #[deprecated]
    #[prost(message, repeated, tag="6")]
    pub config: ::prost::alloc::vec::Vec<KeyValuePair>,
    /// Ports to open in the container. This feature is not supported by all execution engines. (e.g. supported on K8s but
    /// not supported on AWS Batch)
    /// Only K8s
    #[prost(message, repeated, tag="7")]
    pub ports: ::prost::alloc::vec::Vec<ContainerPort>,
    /// BETA: Optional configuration for DataLoading. If not specified, then default values are used.
    /// This makes it possible to to run a completely portable container, that uses inputs and outputs
    /// only from the local file-system and without having any reference to flyteidl. This is supported only on K8s at the moment.
    /// If data loading is enabled, then data will be mounted in accompanying directories specified in the DataLoadingConfig. If the directories
    /// are not specified, inputs will be mounted onto and outputs will be uploaded from a pre-determined file-system path. Refer to the documentation
    /// to understand the default paths.
    /// Only K8s
    #[prost(message, optional, tag="9")]
    pub data_config: ::core::option::Option<DataLoadingConfig>,
    #[prost(enumeration="container::Architecture", tag="10")]
    pub architecture: i32,
}
/// Nested message and enum types in `Container`.
pub mod container {
    /// Architecture-type the container image supports.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Architecture {
        Unknown = 0,
        Amd64 = 1,
        Arm64 = 2,
        ArmV6 = 3,
        ArmV7 = 4,
    }
    impl Architecture {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Architecture::Unknown => "UNKNOWN",
                Architecture::Amd64 => "AMD64",
                Architecture::Arm64 => "ARM64",
                Architecture::ArmV6 => "ARM_V6",
                Architecture::ArmV7 => "ARM_V7",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNKNOWN" => Some(Self::Unknown),
                "AMD64" => Some(Self::Amd64),
                "ARM64" => Some(Self::Arm64),
                "ARM_V6" => Some(Self::ArmV6),
                "ARM_V7" => Some(Self::ArmV7),
                _ => None,
            }
        }
    }
}
/// Strategy to use when dealing with Blob, Schema, or multipart blob data (large datasets)
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct IoStrategy {
    /// Mode to use to manage downloads
    #[prost(enumeration="io_strategy::DownloadMode", tag="1")]
    pub download_mode: i32,
    /// Mode to use to manage uploads
    #[prost(enumeration="io_strategy::UploadMode", tag="2")]
    pub upload_mode: i32,
}
/// Nested message and enum types in `IOStrategy`.
pub mod io_strategy {
    /// Mode to use for downloading
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum DownloadMode {
        /// All data will be downloaded before the main container is executed
        DownloadEager = 0,
        /// Data will be downloaded as a stream and an End-Of-Stream marker will be written to indicate all data has been downloaded. Refer to protocol for details
        DownloadStream = 1,
        /// Large objects (offloaded) will not be downloaded
        DoNotDownload = 2,
    }
    impl DownloadMode {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                DownloadMode::DownloadEager => "DOWNLOAD_EAGER",
                DownloadMode::DownloadStream => "DOWNLOAD_STREAM",
                DownloadMode::DoNotDownload => "DO_NOT_DOWNLOAD",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "DOWNLOAD_EAGER" => Some(Self::DownloadEager),
                "DOWNLOAD_STREAM" => Some(Self::DownloadStream),
                "DO_NOT_DOWNLOAD" => Some(Self::DoNotDownload),
                _ => None,
            }
        }
    }
    /// Mode to use for uploading
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum UploadMode {
        /// All data will be uploaded after the main container exits
        UploadOnExit = 0,
        /// Data will be uploaded as it appears. Refer to protocol specification for details
        UploadEager = 1,
        /// Data will not be uploaded, only references will be written
        DoNotUpload = 2,
    }
    impl UploadMode {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                UploadMode::UploadOnExit => "UPLOAD_ON_EXIT",
                UploadMode::UploadEager => "UPLOAD_EAGER",
                UploadMode::DoNotUpload => "DO_NOT_UPLOAD",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UPLOAD_ON_EXIT" => Some(Self::UploadOnExit),
                "UPLOAD_EAGER" => Some(Self::UploadEager),
                "DO_NOT_UPLOAD" => Some(Self::DoNotUpload),
                _ => None,
            }
        }
    }
}
/// This configuration allows executing raw containers in Flyte using the Flyte CoPilot system.
/// Flyte CoPilot, eliminates the needs of flytekit or sdk inside the container. Any inputs required by the users container are side-loaded in the input_path
/// Any outputs generated by the user container - within output_path are automatically uploaded.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DataLoadingConfig {
    /// Flag enables DataLoading Config. If this is not set, data loading will not be used!
    #[prost(bool, tag="1")]
    pub enabled: bool,
    /// File system path (start at root). This folder will contain all the inputs exploded to a separate file.
    /// Example, if the input interface needs (x: int, y: blob, z: multipart_blob) and the input path is '/var/flyte/inputs', then the file system will look like
    /// /var/flyte/inputs/inputs.<metadata format dependent -> .pb .json .yaml> -> Format as defined previously. The Blob and Multipart blob will reference local filesystem instead of remote locations
    /// /var/flyte/inputs/x -> X is a file that contains the value of x (integer) in string format
    /// /var/flyte/inputs/y -> Y is a file in Binary format
    /// /var/flyte/inputs/z/... -> Note Z itself is a directory
    /// More information about the protocol - refer to docs #TODO reference docs here
    #[prost(string, tag="2")]
    pub input_path: ::prost::alloc::string::String,
    /// File system path (start at root). This folder should contain all the outputs for the task as individual files and/or an error text file
    #[prost(string, tag="3")]
    pub output_path: ::prost::alloc::string::String,
    /// In the inputs folder, there will be an additional summary/metadata file that contains references to all files or inlined primitive values.
    /// This format decides the actual encoding for the data. Refer to the encoding to understand the specifics of the contents and the encoding
    #[prost(enumeration="data_loading_config::LiteralMapFormat", tag="4")]
    pub format: i32,
    #[prost(message, optional, tag="5")]
    pub io_strategy: ::core::option::Option<IoStrategy>,
}
/// Nested message and enum types in `DataLoadingConfig`.
pub mod data_loading_config {
    /// LiteralMapFormat decides the encoding format in which the input metadata should be made available to the containers.
    /// If the user has access to the protocol buffer definitions, it is recommended to use the PROTO format.
    /// JSON and YAML do not need any protobuf definitions to read it
    /// All remote references in core.LiteralMap are replaced with local filesystem references (the data is downloaded to local filesystem)
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum LiteralMapFormat {
        /// JSON / YAML for the metadata (which contains inlined primitive values). The representation is inline with the standard json specification as specified - <https://www.json.org/json-en.html>
        Json = 0,
        Yaml = 1,
        /// Proto is a serialized binary of `core.LiteralMap` defined in flyteidl/core
        Proto = 2,
    }
    impl LiteralMapFormat {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                LiteralMapFormat::Json => "JSON",
                LiteralMapFormat::Yaml => "YAML",
                LiteralMapFormat::Proto => "PROTO",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "JSON" => Some(Self::Json),
                "YAML" => Some(Self::Yaml),
                "PROTO" => Some(Self::Proto),
                _ => None,
            }
        }
    }
}
/// Defines a pod spec and additional pod metadata that is created when a task is executed.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct K8sPod {
    /// Contains additional metadata for building a kubernetes pod.
    #[prost(message, optional, tag="1")]
    pub metadata: ::core::option::Option<K8sObjectMetadata>,
    /// Defines the primary pod spec created when a task is executed.
    /// This should be a JSON-marshalled pod spec, which can be defined in
    /// - go, using: <https://github.com/kubernetes/api/blob/release-1.21/core/v1/types.go#L2936>
    /// - python: using <https://github.com/kubernetes-client/python/blob/release-19.0/kubernetes/client/models/v1_pod_spec.py>
    #[prost(message, optional, tag="2")]
    pub pod_spec: ::core::option::Option<::prost_types::Struct>,
    /// BETA: Optional configuration for DataLoading. If not specified, then default values are used.
    /// This makes it possible to to run a completely portable container, that uses inputs and outputs
    /// only from the local file-system and without having any reference to flytekit. This is supported only on K8s at the moment.
    /// If data loading is enabled, then data will be mounted in accompanying directories specified in the DataLoadingConfig. If the directories
    /// are not specified, inputs will be mounted onto and outputs will be uploaded from a pre-determined file-system path. Refer to the documentation
    /// to understand the default paths.
    /// Only K8s
    #[prost(message, optional, tag="3")]
    pub data_config: ::core::option::Option<DataLoadingConfig>,
}
/// Metadata for building a kubernetes object when a task is executed.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct K8sObjectMetadata {
    /// Optional labels to add to the pod definition.
    #[prost(map="string, string", tag="1")]
    pub labels: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// Optional annotations to add to the pod definition.
    #[prost(map="string, string", tag="2")]
    pub annotations: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
/// Sql represents a generic sql workload with a statement and dialect.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Sql {
    /// The actual query to run, the query can have templated parameters.
    /// We use Flyte's Golang templating format for Query templating.
    /// Refer to the templating documentation.
    /// <https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/external_services/hive/hive.html#sphx-glr-auto-integrations-external-services-hive-hive-py>
    /// For example,
    /// insert overwrite directory '{{ .rawOutputDataPrefix }}' stored as parquet
    /// select *
    /// from my_table
    /// where ds = '{{ .Inputs.ds }}'
    #[prost(string, tag="1")]
    pub statement: ::prost::alloc::string::String,
    #[prost(enumeration="sql::Dialect", tag="2")]
    pub dialect: i32,
}
/// Nested message and enum types in `Sql`.
pub mod sql {
    /// The dialect of the SQL statement. This is used to validate and parse SQL statements at compilation time to avoid
    /// expensive runtime operations. If set to an unsupported dialect, no validation will be done on the statement.
    /// We support the following dialect: ansi, hive.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Dialect {
        Undefined = 0,
        Ansi = 1,
        Hive = 2,
        Other = 3,
    }
    impl Dialect {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Dialect::Undefined => "UNDEFINED",
                Dialect::Ansi => "ANSI",
                Dialect::Hive => "HIVE",
                Dialect::Other => "OTHER",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNDEFINED" => Some(Self::Undefined),
                "ANSI" => Some(Self::Ansi),
                "HIVE" => Some(Self::Hive),
                "OTHER" => Some(Self::Other),
                _ => None,
            }
        }
    }
}
/// Indicates various phases of Workflow Execution
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowExecution {
}
/// Nested message and enum types in `WorkflowExecution`.
pub mod workflow_execution {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Phase {
        Undefined = 0,
        Queued = 1,
        Running = 2,
        Succeeding = 3,
        Succeeded = 4,
        Failing = 5,
        Failed = 6,
        Aborted = 7,
        TimedOut = 8,
        Aborting = 9,
    }
    impl Phase {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Phase::Undefined => "UNDEFINED",
                Phase::Queued => "QUEUED",
                Phase::Running => "RUNNING",
                Phase::Succeeding => "SUCCEEDING",
                Phase::Succeeded => "SUCCEEDED",
                Phase::Failing => "FAILING",
                Phase::Failed => "FAILED",
                Phase::Aborted => "ABORTED",
                Phase::TimedOut => "TIMED_OUT",
                Phase::Aborting => "ABORTING",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNDEFINED" => Some(Self::Undefined),
                "QUEUED" => Some(Self::Queued),
                "RUNNING" => Some(Self::Running),
                "SUCCEEDING" => Some(Self::Succeeding),
                "SUCCEEDED" => Some(Self::Succeeded),
                "FAILING" => Some(Self::Failing),
                "FAILED" => Some(Self::Failed),
                "ABORTED" => Some(Self::Aborted),
                "TIMED_OUT" => Some(Self::TimedOut),
                "ABORTING" => Some(Self::Aborting),
                _ => None,
            }
        }
    }
}
/// Indicates various phases of Node Execution that only include the time spent to run the nodes/workflows
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeExecution {
}
/// Nested message and enum types in `NodeExecution`.
pub mod node_execution {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Phase {
        Undefined = 0,
        Queued = 1,
        Running = 2,
        Succeeded = 3,
        Failing = 4,
        Failed = 5,
        Aborted = 6,
        Skipped = 7,
        TimedOut = 8,
        DynamicRunning = 9,
        Recovered = 10,
    }
    impl Phase {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Phase::Undefined => "UNDEFINED",
                Phase::Queued => "QUEUED",
                Phase::Running => "RUNNING",
                Phase::Succeeded => "SUCCEEDED",
                Phase::Failing => "FAILING",
                Phase::Failed => "FAILED",
                Phase::Aborted => "ABORTED",
                Phase::Skipped => "SKIPPED",
                Phase::TimedOut => "TIMED_OUT",
                Phase::DynamicRunning => "DYNAMIC_RUNNING",
                Phase::Recovered => "RECOVERED",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNDEFINED" => Some(Self::Undefined),
                "QUEUED" => Some(Self::Queued),
                "RUNNING" => Some(Self::Running),
                "SUCCEEDED" => Some(Self::Succeeded),
                "FAILING" => Some(Self::Failing),
                "FAILED" => Some(Self::Failed),
                "ABORTED" => Some(Self::Aborted),
                "SKIPPED" => Some(Self::Skipped),
                "TIMED_OUT" => Some(Self::TimedOut),
                "DYNAMIC_RUNNING" => Some(Self::DynamicRunning),
                "RECOVERED" => Some(Self::Recovered),
                _ => None,
            }
        }
    }
}
/// Phases that task plugins can go through. Not all phases may be applicable to a specific plugin task,
/// but this is the cumulative list that customers may want to know about for their task.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskExecution {
}
/// Nested message and enum types in `TaskExecution`.
pub mod task_execution {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Phase {
        Undefined = 0,
        Queued = 1,
        Running = 2,
        Succeeded = 3,
        Aborted = 4,
        Failed = 5,
        /// To indicate cases where task is initializing, like: ErrImagePull, ContainerCreating, PodInitializing
        Initializing = 6,
        /// To address cases, where underlying resource is not available: Backoff error, Resource quota exceeded
        WaitingForResources = 7,
    }
    impl Phase {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Phase::Undefined => "UNDEFINED",
                Phase::Queued => "QUEUED",
                Phase::Running => "RUNNING",
                Phase::Succeeded => "SUCCEEDED",
                Phase::Aborted => "ABORTED",
                Phase::Failed => "FAILED",
                Phase::Initializing => "INITIALIZING",
                Phase::WaitingForResources => "WAITING_FOR_RESOURCES",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNDEFINED" => Some(Self::Undefined),
                "QUEUED" => Some(Self::Queued),
                "RUNNING" => Some(Self::Running),
                "SUCCEEDED" => Some(Self::Succeeded),
                "ABORTED" => Some(Self::Aborted),
                "FAILED" => Some(Self::Failed),
                "INITIALIZING" => Some(Self::Initializing),
                "WAITING_FOR_RESOURCES" => Some(Self::WaitingForResources),
                _ => None,
            }
        }
    }
}
/// Represents the error message from the execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionError {
    /// Error code indicates a grouping of a type of error.
    /// More Info: <Link>
    #[prost(string, tag="1")]
    pub code: ::prost::alloc::string::String,
    /// Detailed description of the error - including stack trace.
    #[prost(string, tag="2")]
    pub message: ::prost::alloc::string::String,
    /// Full error contents accessible via a URI
    #[prost(string, tag="3")]
    pub error_uri: ::prost::alloc::string::String,
    #[prost(enumeration="execution_error::ErrorKind", tag="4")]
    pub kind: i32,
}
/// Nested message and enum types in `ExecutionError`.
pub mod execution_error {
    /// Error type: System or User
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum ErrorKind {
        Unknown = 0,
        User = 1,
        System = 2,
    }
    impl ErrorKind {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                ErrorKind::Unknown => "UNKNOWN",
                ErrorKind::User => "USER",
                ErrorKind::System => "SYSTEM",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNKNOWN" => Some(Self::Unknown),
                "USER" => Some(Self::User),
                "SYSTEM" => Some(Self::System),
                _ => None,
            }
        }
    }
}
/// Log information for the task that is specific to a log sink
/// When our log story is flushed out, we may have more metadata here like log link expiry
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskLog {
    #[prost(string, tag="1")]
    pub uri: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub name: ::prost::alloc::string::String,
    #[prost(enumeration="task_log::MessageFormat", tag="3")]
    pub message_format: i32,
    #[prost(message, optional, tag="4")]
    pub ttl: ::core::option::Option<::prost_types::Duration>,
}
/// Nested message and enum types in `TaskLog`.
pub mod task_log {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum MessageFormat {
        Unknown = 0,
        Csv = 1,
        Json = 2,
    }
    impl MessageFormat {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                MessageFormat::Unknown => "UNKNOWN",
                MessageFormat::Csv => "CSV",
                MessageFormat::Json => "JSON",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNKNOWN" => Some(Self::Unknown),
                "CSV" => Some(Self::Csv),
                "JSON" => Some(Self::Json),
                _ => None,
            }
        }
    }
}
/// Represents customized execution run-time attributes.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QualityOfServiceSpec {
    /// Indicates how much queueing delay an execution can tolerate.
    #[prost(message, optional, tag="1")]
    pub queueing_budget: ::core::option::Option<::prost_types::Duration>,
}
/// Indicates the priority of an execution.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QualityOfService {
    #[prost(oneof="quality_of_service::Designation", tags="1, 2")]
    pub designation: ::core::option::Option<quality_of_service::Designation>,
}
/// Nested message and enum types in `QualityOfService`.
pub mod quality_of_service {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Tier {
        /// Default: no quality of service specified.
        Undefined = 0,
        High = 1,
        Medium = 2,
        Low = 3,
    }
    impl Tier {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Tier::Undefined => "UNDEFINED",
                Tier::High => "HIGH",
                Tier::Medium => "MEDIUM",
                Tier::Low => "LOW",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNDEFINED" => Some(Self::Undefined),
                "HIGH" => Some(Self::High),
                "MEDIUM" => Some(Self::Medium),
                "LOW" => Some(Self::Low),
                _ => None,
            }
        }
    }
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Designation {
        #[prost(enumeration="Tier", tag="1")]
        Tier(i32),
        #[prost(message, tag="2")]
        Spec(super::QualityOfServiceSpec),
    }
}
/// Defines a 2-level tree where the root is a comparison operator and Operands are primitives or known variables.
/// Each expression results in a boolean result.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ComparisonExpression {
    #[prost(enumeration="comparison_expression::Operator", tag="1")]
    pub operator: i32,
    #[prost(message, optional, tag="2")]
    pub left_value: ::core::option::Option<Operand>,
    #[prost(message, optional, tag="3")]
    pub right_value: ::core::option::Option<Operand>,
}
/// Nested message and enum types in `ComparisonExpression`.
pub mod comparison_expression {
    /// Binary Operator for each expression
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Operator {
        Eq = 0,
        Neq = 1,
        /// Greater Than
        Gt = 2,
        Gte = 3,
        /// Less Than
        Lt = 4,
        Lte = 5,
    }
    impl Operator {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Operator::Eq => "EQ",
                Operator::Neq => "NEQ",
                Operator::Gt => "GT",
                Operator::Gte => "GTE",
                Operator::Lt => "LT",
                Operator::Lte => "LTE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "EQ" => Some(Self::Eq),
                "NEQ" => Some(Self::Neq),
                "GT" => Some(Self::Gt),
                "GTE" => Some(Self::Gte),
                "LT" => Some(Self::Lt),
                "LTE" => Some(Self::Lte),
                _ => None,
            }
        }
    }
}
/// Defines an operand to a comparison expression.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Operand {
    #[prost(oneof="operand::Val", tags="1, 2, 3")]
    pub val: ::core::option::Option<operand::Val>,
}
/// Nested message and enum types in `Operand`.
pub mod operand {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Val {
        /// Can be a constant
        #[prost(message, tag="1")]
        Primitive(super::Primitive),
        /// Or one of this node's input variables
        #[prost(string, tag="2")]
        Var(::prost::alloc::string::String),
        /// Replace the primitive field
        #[prost(message, tag="3")]
        Scalar(super::Scalar),
    }
}
/// Defines a boolean expression tree. It can be a simple or a conjunction expression.
/// Multiple expressions can be combined using a conjunction or a disjunction to result in a final boolean result.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BooleanExpression {
    #[prost(oneof="boolean_expression::Expr", tags="1, 2")]
    pub expr: ::core::option::Option<boolean_expression::Expr>,
}
/// Nested message and enum types in `BooleanExpression`.
pub mod boolean_expression {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Expr {
        #[prost(message, tag="1")]
        Conjunction(::prost::alloc::boxed::Box<super::ConjunctionExpression>),
        #[prost(message, tag="2")]
        Comparison(super::ComparisonExpression),
    }
}
/// Defines a conjunction expression of two boolean expressions.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConjunctionExpression {
    #[prost(enumeration="conjunction_expression::LogicalOperator", tag="1")]
    pub operator: i32,
    #[prost(message, optional, boxed, tag="2")]
    pub left_expression: ::core::option::Option<::prost::alloc::boxed::Box<BooleanExpression>>,
    #[prost(message, optional, boxed, tag="3")]
    pub right_expression: ::core::option::Option<::prost::alloc::boxed::Box<BooleanExpression>>,
}
/// Nested message and enum types in `ConjunctionExpression`.
pub mod conjunction_expression {
    /// Nested conditions. They can be conjoined using AND / OR
    /// Order of evaluation is not important as the operators are Commutative
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum LogicalOperator {
        /// Conjunction
        And = 0,
        Or = 1,
    }
    impl LogicalOperator {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                LogicalOperator::And => "AND",
                LogicalOperator::Or => "OR",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "AND" => Some(Self::And),
                "OR" => Some(Self::Or),
                _ => None,
            }
        }
    }
}
/// Defines a condition and the execution unit that should be executed if the condition is satisfied.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct IfBlock {
    #[prost(message, optional, tag="1")]
    pub condition: ::core::option::Option<BooleanExpression>,
    #[prost(message, optional, boxed, tag="2")]
    pub then_node: ::core::option::Option<::prost::alloc::boxed::Box<Node>>,
}
/// Defines a series of if/else blocks. The first branch whose condition evaluates to true is the one to execute.
/// If no conditions were satisfied, the else_node or the error will execute.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct IfElseBlock {
    /// +required. First condition to evaluate.
    #[prost(message, optional, boxed, tag="1")]
    pub case: ::core::option::Option<::prost::alloc::boxed::Box<IfBlock>>,
    /// +optional. Additional branches to evaluate.
    #[prost(message, repeated, tag="2")]
    pub other: ::prost::alloc::vec::Vec<IfBlock>,
    /// +required.
    #[prost(oneof="if_else_block::Default", tags="3, 4")]
    pub default: ::core::option::Option<if_else_block::Default>,
}
/// Nested message and enum types in `IfElseBlock`.
pub mod if_else_block {
    /// +required.
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Default {
        /// The node to execute in case none of the branches were taken.
        #[prost(message, tag="3")]
        ElseNode(::prost::alloc::boxed::Box<super::Node>),
        /// An error to throw in case none of the branches were taken.
        #[prost(message, tag="4")]
        Error(super::Error),
    }
}
/// BranchNode is a special node that alter the flow of the workflow graph. It allows the control flow to branch at
/// runtime based on a series of conditions that get evaluated on various parameters (e.g. inputs, primitives).
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BranchNode {
    /// +required
    #[prost(message, optional, boxed, tag="1")]
    pub if_else: ::core::option::Option<::prost::alloc::boxed::Box<IfElseBlock>>,
}
/// Refers to the task that the Node is to execute.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskNode {
    /// Optional overrides applied at task execution time.
    #[prost(message, optional, tag="2")]
    pub overrides: ::core::option::Option<TaskNodeOverrides>,
    #[prost(oneof="task_node::Reference", tags="1")]
    pub reference: ::core::option::Option<task_node::Reference>,
}
/// Nested message and enum types in `TaskNode`.
pub mod task_node {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Reference {
        /// A globally unique identifier for the task.
        #[prost(message, tag="1")]
        ReferenceId(super::Identifier),
    }
}
/// Refers to a the workflow the node is to execute.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowNode {
    #[prost(oneof="workflow_node::Reference", tags="1, 2")]
    pub reference: ::core::option::Option<workflow_node::Reference>,
}
/// Nested message and enum types in `WorkflowNode`.
pub mod workflow_node {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Reference {
        /// A globally unique identifier for the launch plan.
        #[prost(message, tag="1")]
        LaunchplanRef(super::Identifier),
        /// Reference to a subworkflow, that should be defined with the compiler context
        #[prost(message, tag="2")]
        SubWorkflowRef(super::Identifier),
    }
}
/// ApproveCondition represents a dependency on an external approval. During execution, this will manifest as a boolean
/// signal with the provided signal_id.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ApproveCondition {
    /// A unique identifier for the requested boolean signal.
    #[prost(string, tag="1")]
    pub signal_id: ::prost::alloc::string::String,
}
/// SignalCondition represents a dependency on an signal.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignalCondition {
    /// A unique identifier for the requested signal.
    #[prost(string, tag="1")]
    pub signal_id: ::prost::alloc::string::String,
    /// A type denoting the required value type for this signal.
    #[prost(message, optional, tag="2")]
    pub r#type: ::core::option::Option<LiteralType>,
    /// The variable name for the signal value in this nodes outputs.
    #[prost(string, tag="3")]
    pub output_variable_name: ::prost::alloc::string::String,
}
/// SleepCondition represents a dependency on waiting for the specified duration.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SleepCondition {
    /// The overall duration for this sleep.
    #[prost(message, optional, tag="1")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
}
/// GateNode refers to the condition that is required for the gate to successfully complete.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GateNode {
    #[prost(oneof="gate_node::Condition", tags="1, 2, 3")]
    pub condition: ::core::option::Option<gate_node::Condition>,
}
/// Nested message and enum types in `GateNode`.
pub mod gate_node {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Condition {
        /// ApproveCondition represents a dependency on an external approval provided by a boolean signal.
        #[prost(message, tag="1")]
        Approve(super::ApproveCondition),
        /// SignalCondition represents a dependency on an signal.
        #[prost(message, tag="2")]
        Signal(super::SignalCondition),
        /// SleepCondition represents a dependency on waiting for the specified duration.
        #[prost(message, tag="3")]
        Sleep(super::SleepCondition),
    }
}
/// ArrayNode is a Flyte node type that simplifies the execution of a sub-node over a list of input
/// values. An ArrayNode can be executed with configurable parallelism (separate from the parent
/// workflow) and can be configured to succeed when a certain number of sub-nodes succeed.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ArrayNode {
    /// node is the sub-node that will be executed for each element in the array.
    #[prost(message, optional, boxed, tag="1")]
    pub node: ::core::option::Option<::prost::alloc::boxed::Box<Node>>,
    /// parallelism defines the minimum number of instances to bring up concurrently at any given
    /// point. Note that this is an optimistic restriction and that, due to network partitioning or
    /// other failures, the actual number of currently running instances might be more. This has to
    /// be a positive number if assigned. Default value is size.
    #[prost(uint32, tag="2")]
    pub parallelism: u32,
    #[prost(oneof="array_node::SuccessCriteria", tags="3, 4")]
    pub success_criteria: ::core::option::Option<array_node::SuccessCriteria>,
}
/// Nested message and enum types in `ArrayNode`.
pub mod array_node {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum SuccessCriteria {
        /// min_successes is an absolute number of the minimum number of successful completions of
        /// sub-nodes. As soon as this criteria is met, the ArrayNode will be marked as successful
        /// and outputs will be computed. This has to be a non-negative number if assigned. Default
        /// value is size (if specified).
        #[prost(uint32, tag="3")]
        MinSuccesses(u32),
        /// If the array job size is not known beforehand, the min_success_ratio can instead be used
        /// to determine when an ArrayNode can be marked successful.
        #[prost(float, tag="4")]
        MinSuccessRatio(f32),
    }
}
/// Defines extra information about the Node.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct NodeMetadata {
    /// A friendly name for the Node
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    /// The overall timeout of a task.
    #[prost(message, optional, tag="4")]
    pub timeout: ::core::option::Option<::prost_types::Duration>,
    /// Number of retries per task.
    #[prost(message, optional, tag="5")]
    pub retries: ::core::option::Option<RetryStrategy>,
    /// Identify whether node is interruptible
    #[prost(oneof="node_metadata::InterruptibleValue", tags="6")]
    pub interruptible_value: ::core::option::Option<node_metadata::InterruptibleValue>,
}
/// Nested message and enum types in `NodeMetadata`.
pub mod node_metadata {
    /// Identify whether node is interruptible
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum InterruptibleValue {
        #[prost(bool, tag="6")]
        Interruptible(bool),
    }
}
/// Links a variable to an alias.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Alias {
    /// Must match one of the output variable names on a node.
    #[prost(string, tag="1")]
    pub var: ::prost::alloc::string::String,
    /// A workflow-level unique alias that downstream nodes can refer to in their input.
    #[prost(string, tag="2")]
    pub alias: ::prost::alloc::string::String,
}
/// A Workflow graph Node. One unit of execution in the graph. Each node can be linked to a Task, a Workflow or a branch
/// node.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Node {
    /// A workflow-level unique identifier that identifies this node in the workflow. 'inputs' and 'outputs' are reserved
    /// node ids that cannot be used by other nodes.
    #[prost(string, tag="1")]
    pub id: ::prost::alloc::string::String,
    /// Extra metadata about the node.
    #[prost(message, optional, tag="2")]
    pub metadata: ::core::option::Option<NodeMetadata>,
    /// Specifies how to bind the underlying interface's inputs. All required inputs specified in the underlying interface
    /// must be fulfilled.
    #[prost(message, repeated, tag="3")]
    pub inputs: ::prost::alloc::vec::Vec<Binding>,
    /// +optional Specifies execution dependency for this node ensuring it will only get scheduled to run after all its
    /// upstream nodes have completed. This node will have an implicit dependency on any node that appears in inputs
    /// field.
    #[prost(string, repeated, tag="4")]
    pub upstream_node_ids: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// +optional. A node can define aliases for a subset of its outputs. This is particularly useful if different nodes
    /// need to conform to the same interface (e.g. all branches in a branch node). Downstream nodes must refer to this
    /// nodes outputs using the alias if one's specified.
    #[prost(message, repeated, tag="5")]
    pub output_aliases: ::prost::alloc::vec::Vec<Alias>,
    /// Information about the target to execute in this node.
    #[prost(oneof="node::Target", tags="6, 7, 8, 9, 10")]
    pub target: ::core::option::Option<node::Target>,
}
/// Nested message and enum types in `Node`.
pub mod node {
    /// Information about the target to execute in this node.
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Target {
        /// Information about the Task to execute in this node.
        #[prost(message, tag="6")]
        TaskNode(super::TaskNode),
        /// Information about the Workflow to execute in this mode.
        #[prost(message, tag="7")]
        WorkflowNode(super::WorkflowNode),
        /// Information about the branch node to evaluate in this node.
        #[prost(message, tag="8")]
        BranchNode(::prost::alloc::boxed::Box<super::BranchNode>),
        /// Information about the condition to evaluate in this node.
        #[prost(message, tag="9")]
        GateNode(super::GateNode),
        /// Information about the sub-node executions for each value in the list of this nodes
        /// inputs values.
        #[prost(message, tag="10")]
        ArrayNode(::prost::alloc::boxed::Box<super::ArrayNode>),
    }
}
/// This is workflow layer metadata. These settings are only applicable to the workflow as a whole, and do not
/// percolate down to child entities (like tasks) launched by the workflow.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowMetadata {
    /// Indicates the runtime priority of workflow executions. 
    #[prost(message, optional, tag="1")]
    pub quality_of_service: ::core::option::Option<QualityOfService>,
    /// Defines how the system should behave when a failure is detected in the workflow execution.
    #[prost(enumeration="workflow_metadata::OnFailurePolicy", tag="2")]
    pub on_failure: i32,
    /// Arbitrary tags that allow users and the platform to store small but arbitrary labels
    #[prost(map="string, string", tag="3")]
    pub tags: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
/// Nested message and enum types in `WorkflowMetadata`.
pub mod workflow_metadata {
    /// Failure Handling Strategy
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum OnFailurePolicy {
        /// FAIL_IMMEDIATELY instructs the system to fail as soon as a node fails in the workflow. It'll automatically
        /// abort all currently running nodes and clean up resources before finally marking the workflow executions as
        /// failed.
        FailImmediately = 0,
        /// FAIL_AFTER_EXECUTABLE_NODES_COMPLETE instructs the system to make as much progress as it can. The system will
        /// not alter the dependencies of the execution graph so any node that depend on the failed node will not be run.
        /// Other nodes that will be executed to completion before cleaning up resources and marking the workflow
        /// execution as failed.
        FailAfterExecutableNodesComplete = 1,
    }
    impl OnFailurePolicy {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                OnFailurePolicy::FailImmediately => "FAIL_IMMEDIATELY",
                OnFailurePolicy::FailAfterExecutableNodesComplete => "FAIL_AFTER_EXECUTABLE_NODES_COMPLETE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "FAIL_IMMEDIATELY" => Some(Self::FailImmediately),
                "FAIL_AFTER_EXECUTABLE_NODES_COMPLETE" => Some(Self::FailAfterExecutableNodesComplete),
                _ => None,
            }
        }
    }
}
/// The difference between these settings and the WorkflowMetadata ones is that these are meant to be passed down to
/// a workflow's underlying entities (like tasks). For instance, 'interruptible' has no meaning at the workflow layer, it
/// is only relevant when a task executes. The settings here are the defaults that are passed to all nodes
/// unless explicitly overridden at the node layer.
/// If you are adding a setting that applies to both the Workflow itself, and everything underneath it, it should be
/// added to both this object and the WorkflowMetadata object above.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowMetadataDefaults {
    /// Whether child nodes of the workflow are interruptible.
    #[prost(bool, tag="1")]
    pub interruptible: bool,
}
/// Flyte Workflow Structure that encapsulates task, branch and subworkflow nodes to form a statically analyzable,
/// directed acyclic graph.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowTemplate {
    /// A globally unique identifier for the workflow.
    #[prost(message, optional, tag="1")]
    pub id: ::core::option::Option<Identifier>,
    /// Extra metadata about the workflow.
    #[prost(message, optional, tag="2")]
    pub metadata: ::core::option::Option<WorkflowMetadata>,
    /// Defines a strongly typed interface for the Workflow. This can include some optional parameters.
    #[prost(message, optional, tag="3")]
    pub interface: ::core::option::Option<TypedInterface>,
    /// A list of nodes. In addition, 'globals' is a special reserved node id that can be used to consume workflow inputs.
    #[prost(message, repeated, tag="4")]
    pub nodes: ::prost::alloc::vec::Vec<Node>,
    /// A list of output bindings that specify how to construct workflow outputs. Bindings can pull node outputs or
    /// specify literals. All workflow outputs specified in the interface field must be bound in order for the workflow
    /// to be validated. A workflow has an implicit dependency on all of its nodes to execute successfully in order to
    /// bind final outputs.
    /// Most of these outputs will be Binding's with a BindingData of type OutputReference.  That is, your workflow can
    /// just have an output of some constant (`Output(5)`), but usually, the workflow will be pulling
    /// outputs from the output of a task.
    #[prost(message, repeated, tag="5")]
    pub outputs: ::prost::alloc::vec::Vec<Binding>,
    /// +optional A catch-all node. This node is executed whenever the execution engine determines the workflow has failed.
    /// The interface of this node must match the Workflow interface with an additional input named 'error' of type
    /// pb.lyft.flyte.core.Error.
    #[prost(message, optional, tag="6")]
    pub failure_node: ::core::option::Option<Node>,
    /// workflow defaults
    #[prost(message, optional, tag="7")]
    pub metadata_defaults: ::core::option::Option<WorkflowMetadataDefaults>,
}
/// Optional task node overrides that will be applied at task execution time.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TaskNodeOverrides {
    /// A customizable interface to convey resources requested for a task container. 
    #[prost(message, optional, tag="1")]
    pub resources: ::core::option::Option<Resources>,
}
/// Adjacency list for the workflow. This is created as part of the compilation process. Every process after the compilation
/// step uses this created ConnectionSet
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConnectionSet {
    /// A list of all the node ids that are downstream from a given node id
    #[prost(map="string, message", tag="7")]
    pub downstream: ::std::collections::HashMap<::prost::alloc::string::String, connection_set::IdList>,
    /// A list of all the node ids, that are upstream of this node id
    #[prost(map="string, message", tag="8")]
    pub upstream: ::std::collections::HashMap<::prost::alloc::string::String, connection_set::IdList>,
}
/// Nested message and enum types in `ConnectionSet`.
pub mod connection_set {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
    pub struct IdList {
        #[prost(string, repeated, tag="1")]
        pub ids: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    }
}
/// Output of the compilation Step. This object represents one workflow. We store more metadata at this layer
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CompiledWorkflow {
    /// Completely contained Workflow Template
    #[prost(message, optional, tag="1")]
    pub template: ::core::option::Option<WorkflowTemplate>,
    /// For internal use only! This field is used by the system and must not be filled in. Any values set will be ignored.
    #[prost(message, optional, tag="2")]
    pub connections: ::core::option::Option<ConnectionSet>,
}
/// Output of the Compilation step. This object represent one Task. We store more metadata at this layer
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CompiledTask {
    /// Completely contained TaskTemplate
    #[prost(message, optional, tag="1")]
    pub template: ::core::option::Option<TaskTemplate>,
}
/// A Compiled Workflow Closure contains all the information required to start a new execution, or to visualize a workflow
/// and its details. The CompiledWorkflowClosure should always contain a primary workflow, that is the main workflow that
/// will being the execution. All subworkflows are denormalized. WorkflowNodes refer to the workflow identifiers of
/// compiled subworkflows.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CompiledWorkflowClosure {
    /// +required
    #[prost(message, optional, tag="1")]
    pub primary: ::core::option::Option<CompiledWorkflow>,
    /// Guaranteed that there will only exist one and only one workflow with a given id, i.e., every sub workflow has a
    /// unique identifier. Also every enclosed subworkflow is used either by a primary workflow or by a subworkflow
    /// as an inlined workflow
    /// +optional
    #[prost(message, repeated, tag="2")]
    pub sub_workflows: ::prost::alloc::vec::Vec<CompiledWorkflow>,
    /// Guaranteed that there will only exist one and only one task with a given id, i.e., every task has a unique id
    /// +required (at least 1)
    #[prost(message, repeated, tag="3")]
    pub tasks: ::prost::alloc::vec::Vec<CompiledTask>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CatalogArtifactTag {
    /// Artifact ID is generated name
    #[prost(string, tag="1")]
    pub artifact_id: ::prost::alloc::string::String,
    /// Flyte computes the tag automatically, as the hash of the values
    #[prost(string, tag="2")]
    pub name: ::prost::alloc::string::String,
}
/// Catalog artifact information with specific metadata
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CatalogMetadata {
    /// Dataset ID in the catalog
    #[prost(message, optional, tag="1")]
    pub dataset_id: ::core::option::Option<Identifier>,
    /// Artifact tag in the catalog
    #[prost(message, optional, tag="2")]
    pub artifact_tag: ::core::option::Option<CatalogArtifactTag>,
    /// Optional: Source Execution identifier, if this dataset was generated by another execution in Flyte. This is a one-of field and will depend on the caching context
    #[prost(oneof="catalog_metadata::SourceExecution", tags="3")]
    pub source_execution: ::core::option::Option<catalog_metadata::SourceExecution>,
}
/// Nested message and enum types in `CatalogMetadata`.
pub mod catalog_metadata {
    /// Optional: Source Execution identifier, if this dataset was generated by another execution in Flyte. This is a one-of field and will depend on the caching context
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum SourceExecution {
        /// Today we only support TaskExecutionIdentifier as a source, as catalog caching only works for task executions
        #[prost(message, tag="3")]
        SourceTaskExecution(super::TaskExecutionIdentifier),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CatalogReservation {
}
/// Nested message and enum types in `CatalogReservation`.
pub mod catalog_reservation {
    /// Indicates the status of a catalog reservation operation.
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Status {
        /// Used to indicate that reservations are disabled
        ReservationDisabled = 0,
        /// Used to indicate that a reservation was successfully acquired or extended
        ReservationAcquired = 1,
        /// Used to indicate that an active reservation currently exists
        ReservationExists = 2,
        /// Used to indicate that the reservation has been successfully released
        ReservationReleased = 3,
        /// Used to indicate that a reservation operation resulted in failure
        ReservationFailure = 4,
    }
    impl Status {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Status::ReservationDisabled => "RESERVATION_DISABLED",
                Status::ReservationAcquired => "RESERVATION_ACQUIRED",
                Status::ReservationExists => "RESERVATION_EXISTS",
                Status::ReservationReleased => "RESERVATION_RELEASED",
                Status::ReservationFailure => "RESERVATION_FAILURE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "RESERVATION_DISABLED" => Some(Self::ReservationDisabled),
                "RESERVATION_ACQUIRED" => Some(Self::ReservationAcquired),
                "RESERVATION_EXISTS" => Some(Self::ReservationExists),
                "RESERVATION_RELEASED" => Some(Self::ReservationReleased),
                "RESERVATION_FAILURE" => Some(Self::ReservationFailure),
                _ => None,
            }
        }
    }
}
/// Indicates the status of CatalogCaching. The reason why this is not embedded in TaskNodeMetadata is, that we may use for other types of nodes as well in the future
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum CatalogCacheStatus {
    /// Used to indicate that caching was disabled
    CacheDisabled = 0,
    /// Used to indicate that the cache lookup resulted in no matches
    CacheMiss = 1,
    /// used to indicate that the associated artifact was a result of a previous execution
    CacheHit = 2,
    /// used to indicate that the resultant artifact was added to the cache
    CachePopulated = 3,
    /// Used to indicate that cache lookup failed because of an error
    CacheLookupFailure = 4,
    /// Used to indicate that cache lookup failed because of an error
    CachePutFailure = 5,
    /// Used to indicate the cache lookup was skipped
    CacheSkipped = 6,
}
impl CatalogCacheStatus {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            CatalogCacheStatus::CacheDisabled => "CACHE_DISABLED",
            CatalogCacheStatus::CacheMiss => "CACHE_MISS",
            CatalogCacheStatus::CacheHit => "CACHE_HIT",
            CatalogCacheStatus::CachePopulated => "CACHE_POPULATED",
            CatalogCacheStatus::CacheLookupFailure => "CACHE_LOOKUP_FAILURE",
            CatalogCacheStatus::CachePutFailure => "CACHE_PUT_FAILURE",
            CatalogCacheStatus::CacheSkipped => "CACHE_SKIPPED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "CACHE_DISABLED" => Some(Self::CacheDisabled),
            "CACHE_MISS" => Some(Self::CacheMiss),
            "CACHE_HIT" => Some(Self::CacheHit),
            "CACHE_POPULATED" => Some(Self::CachePopulated),
            "CACHE_LOOKUP_FAILURE" => Some(Self::CacheLookupFailure),
            "CACHE_PUT_FAILURE" => Some(Self::CachePutFailure),
            "CACHE_SKIPPED" => Some(Self::CacheSkipped),
            _ => None,
        }
    }
}
/// Span represents a duration trace of Flyte execution. The id field denotes a Flyte execution entity or an operation
/// which uniquely identifies the Span. The spans attribute allows this Span to be further broken down into more
/// precise definitions.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Span {
    /// start_time defines the instance this span began.
    #[prost(message, optional, tag="1")]
    pub start_time: ::core::option::Option<::prost_types::Timestamp>,
    /// end_time defines the instance this span completed.
    #[prost(message, optional, tag="2")]
    pub end_time: ::core::option::Option<::prost_types::Timestamp>,
    /// spans defines a collection of Spans that breakdown this execution.
    #[prost(message, repeated, tag="7")]
    pub spans: ::prost::alloc::vec::Vec<Span>,
    #[prost(oneof="span::Id", tags="3, 4, 5, 6")]
    pub id: ::core::option::Option<span::Id>,
}
/// Nested message and enum types in `Span`.
pub mod span {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Id {
        /// workflow_id is the id of the workflow execution this Span represents.
        #[prost(message, tag="3")]
        WorkflowId(super::WorkflowExecutionIdentifier),
        /// node_id is the id of the node execution this Span represents.
        #[prost(message, tag="4")]
        NodeId(super::NodeExecutionIdentifier),
        /// task_id is the id of the task execution this Span represents.
        #[prost(message, tag="5")]
        TaskId(super::TaskExecutionIdentifier),
        /// operation_id is the id of a unique operation that this Span represents.
        #[prost(string, tag="6")]
        OperationId(::prost::alloc::string::String),
    }
}
/// Describes a set of tasks to execute and how the final outputs are produced.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DynamicJobSpec {
    /// A collection of nodes to execute.
    #[prost(message, repeated, tag="1")]
    pub nodes: ::prost::alloc::vec::Vec<Node>,
    /// An absolute number of successful completions of nodes required to mark this job as succeeded. As soon as this
    /// criteria is met, the dynamic job will be marked as successful and outputs will be computed. If this number
    /// becomes impossible to reach (e.g. number of currently running tasks + number of already succeeded tasks <
    /// min_successes) the task will be aborted immediately and marked as failed. The default value of this field, if not
    /// specified, is the count of nodes repeated field.
    #[prost(int64, tag="2")]
    pub min_successes: i64,
    /// Describes how to bind the final output of the dynamic job from the outputs of executed nodes. The referenced ids
    /// in bindings should have the generated id for the subtask.
    #[prost(message, repeated, tag="3")]
    pub outputs: ::prost::alloc::vec::Vec<Binding>,
    /// \[Optional\] A complete list of task specs referenced in nodes.
    #[prost(message, repeated, tag="4")]
    pub tasks: ::prost::alloc::vec::Vec<TaskTemplate>,
    /// \[Optional\] A complete list of task specs referenced in nodes.
    #[prost(message, repeated, tag="5")]
    pub subworkflows: ::prost::alloc::vec::Vec<WorkflowTemplate>,
}
/// Error message to propagate detailed errors from container executions to the execution
/// engine.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContainerError {
    /// A simplified code for errors, so that we can provide a glossary of all possible errors.
    #[prost(string, tag="1")]
    pub code: ::prost::alloc::string::String,
    /// A detailed error message.
    #[prost(string, tag="2")]
    pub message: ::prost::alloc::string::String,
    /// An abstract error kind for this error. Defaults to Non_Recoverable if not specified.
    #[prost(enumeration="container_error::Kind", tag="3")]
    pub kind: i32,
    /// Defines the origin of the error (system, user, unknown).
    #[prost(enumeration="execution_error::ErrorKind", tag="4")]
    pub origin: i32,
}
/// Nested message and enum types in `ContainerError`.
pub mod container_error {
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
/// Defines the errors.pb file format the container can produce to communicate
/// failure reasons to the execution engine.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ErrorDocument {
    /// The error raised during execution.
    #[prost(message, optional, tag="1")]
    pub error: ::core::option::Option<ContainerError>,
}
/// Defines an enclosed package of workflow and tasks it references.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkflowClosure {
    /// required. Workflow template.
    #[prost(message, optional, tag="1")]
    pub workflow: ::core::option::Option<WorkflowTemplate>,
    /// optional. A collection of tasks referenced by the workflow. Only needed if the workflow
    /// references tasks.
    #[prost(message, repeated, tag="2")]
    pub tasks: ::prost::alloc::vec::Vec<TaskTemplate>,
}
// @@protoc_insertion_point(module)
