pub mod auth;

pub mod google {
    pub mod protobuf {
        include!("../gen/pb_rust/google.protobuf.rs");
    }
}

pub mod flyteidl {
    pub mod admin {
        include!("../gen/pb_rust/flyteidl.admin.rs");
    }

    // Recursive types, such as `ArrayNode`, can hold references to other Nodes and are converted to `ArrayNode {node: Option<Box<Node>>, ...}`.
    // This poses a challenge for PyO3, as it requires a PyClass to hold a reference to another PyClass. This is a Rust Non-lifetime-free issue.
    pub mod core {
        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<Node> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<Node> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(Node::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<LiteralType> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<LiteralType> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(LiteralType::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<Literal> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<Literal> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(Literal::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<Union> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<Union> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(Union::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<Scalar> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<Scalar> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(Scalar::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<ConjunctionExpression> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<ConjunctionExpression> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(ConjunctionExpression::extract_bound(obj).unwrap()))
            }
        }
        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<BooleanExpression> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<BooleanExpression> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(BooleanExpression::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<IfBlock> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<IfBlock> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(IfBlock::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<IfElseBlock> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<IfElseBlock> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(IfElseBlock::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<BranchNode> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<BranchNode> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(BranchNode::extract_bound(obj).unwrap()))
            }
        }

        impl pyo3::conversion::IntoPy<pyo3::PyObject> for Box<ArrayNode> {
            fn into_py(self, py: pyo3::marker::Python<'_>) -> pyo3::PyObject {
                self.as_ref().clone().into_py(py)
            }
        }
        impl<'py> pyo3::conversion::FromPyObject<'py> for Box<ArrayNode> {
            fn extract_bound(obj: &pyo3::Bound<'py, pyo3::types::PyAny>) -> pyo3::PyResult<Self> {
                Ok(Box::new(ArrayNode::extract_bound(obj).unwrap()))
            }
        }

        include!("../gen/pb_rust/flyteidl.core.rs");
        #[pyo3::pymethods]
        impl literal_type::Type {
            fn __repr__(&self) -> String {
                match self {
                    Self::Simple(value) => format!("{}", value),
                    _ => todo!(),
                }
            }
        }
    }
    pub mod cache {
        include!("../gen/pb_rust/flyteidl.cacheservice.rs");
    }
    pub mod event {
        include!("../gen/pb_rust/flyteidl.event.rs");
    }
    pub mod plugins {
        include!("../gen/pb_rust/flyteidl.plugins.rs");
        pub mod kubeflow {
            // TODO:
            // include!("../gen/pb_rust/flyteidl.plugins.kubeflow.rs");
        }
    }
    pub mod service {
        include!("../gen/pb_rust/flyteidl.service.rs");
    }
}

use pyo3::prelude::*;

// Set up maturin configuration in `pyproject.toml`
// python-source = "python"
// module-name = "flyteidl_rust._flyteidl_rust"
#[pymodule]
pub mod _flyteidl_rust {
    use std::collections::HashMap;
    use std::fmt;
    use std::fs::File;

    use keyring::Entry;
    use prost::{DecodeError, EncodeError, Message};
    use pyo3::{
        exceptions::{PyException, PyValueError},
        prelude::*,
        types::PyBytes,
        types::{PyDict, PyTuple},
        PyErr,
    };
    use serde_json;
    use tonic::{
        metadata::MetadataValue,
        service::interceptor::InterceptedService,
        service::Interceptor,
        transport::{Certificate, Channel, ClientTlsConfig, Endpoint, Uri},
        Request, Response, Status,
    };

    use crate::auth;

    // Foreign Rust error types: https://pyo3.rs/main/function/error-handling#foreign-rust-error-types
    // Create a newtype wrapper, e.g. MyOtherError. Then implement From<MyOtherError> for PyErr (or PyErrArguments), as well as From<OtherError> for MyOtherError.
    pub struct MessageJsonError(serde_json::Error);
    impl fmt::Display for MessageJsonError {
        fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
            write!(f, "")
        }
    }
    impl std::convert::From<MessageJsonError> for PyErr {
        fn from(err: MessageJsonError) -> PyErr {
            PyException::new_err(err.to_string())
        }
    }
    impl std::convert::From<serde_json::Error> for MessageJsonError {
        fn from(other: serde_json::Error) -> Self {
            Self(other)
        }
    }

    // An error indicates taht failing at serializing object to bytes string, like `SerializToString()` for python protos.
    pub struct MessageEncodeError(EncodeError);
    // An error indicates taht failing at deserializing object from bytes string, like `ParseFromString()` for python protos.
    pub struct MessageDecodeError(DecodeError);

    // TODO: Do we need this formatting (to string)?
    impl fmt::Display for MessageEncodeError {
        fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
            write!(f, "")
        }
    }

    // TODO: Do we need this formatting (to string)?
    impl fmt::Display for MessageDecodeError {
        fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
            write!(f, "")
        }
    }

    impl std::convert::From<MessageEncodeError> for PyErr {
        fn from(err: MessageEncodeError) -> PyErr {
            PyException::new_err(err.to_string())
        }
    }

    impl std::convert::From<MessageDecodeError> for PyErr {
        fn from(err: MessageDecodeError) -> PyErr {
            PyException::new_err(err.to_string())
        }
    }

    impl std::convert::From<EncodeError> for MessageEncodeError {
        fn from(other: EncodeError) -> Self {
            Self(other)
        }
    }

    impl std::convert::From<DecodeError> for MessageDecodeError {
        fn from(other: DecodeError) -> Self {
            Self(other)
        }
    }

    pub trait ProtobufDecoder<T>
    where
        T: Message + Default,
    {
        fn ParseFromString(&self, bytes_obj: &PyBytes) -> Result<T, MessageDecodeError>;
    }

    pub trait ProtobufEncoder<T>
    where
        T: Message + Default,
    {
        fn SerializeToString(&self, res: T) -> Result<Vec<u8>, MessageEncodeError>;
    }

    impl<T> ProtobufDecoder<T> for T
    where
        T: Message + Default,
    {
        fn ParseFromString(&self, bytes_obj: &PyBytes) -> Result<T, MessageDecodeError> {
            let bytes = bytes_obj.as_bytes();
            let de = Message::decode(&bytes.to_vec()[..]);
            Ok(de?)
        }
    }

    impl<T> ProtobufEncoder<T> for T
    where
        T: Message + Default,
    {
        fn SerializeToString(&self, res: T) -> Result<Vec<u8>, MessageEncodeError> {
            let mut buf = vec![];
            res.encode(&mut buf)?;
            Ok(buf)
        }
    }

    #[pyclass]
    #[pyo3(name = "FlyteUserException", subclass, extends = pyo3::exceptions::PyException)] // or PyBaseException (https://github.com/PyO3/pyo3/discussions/4165#discussioncomment-10073433)
    pub struct PyFlyteUserException {}
    #[pymethods]
    impl PyFlyteUserException {
        #[new]
        #[pyo3(signature = (*_args, **_kwargs))]
        fn new(
            _args: Bound<'_, pyo3::types::PyTuple>,
            _kwargs: Option<Bound<'_, pyo3::types::PyDict>>,
        ) -> Self {
            Self {}
        }
    }
    #[pyclass(extends = pyo3::exceptions::PyException)]
    pub struct FlyteEntityNotExistException {}
    #[pymethods]
    impl FlyteEntityNotExistException {
        #[new]
        #[pyo3(signature = (*_args, **_kwargs))]
        fn new(
            _args: Bound<'_, pyo3::types::PyTuple>,
            _kwargs: Option<Bound<'_, pyo3::types::PyDict>>,
        ) -> Self {
            Self {}
        }
    }
    #[pyclass(extends = pyo3::exceptions::PyException)]
    pub struct FlyteEntityAlreadyExistsException {}
    #[pymethods]
    impl FlyteEntityAlreadyExistsException {
        #[new]
        #[pyo3(signature = (*_args, **_kwargs))]
        fn new(
            _args: Bound<'_, pyo3::types::PyTuple>,
            _kwargs: Option<Bound<'_, pyo3::types::PyDict>>,
        ) -> Self {
            Self {}
        }
    }
    #[pyclass(extends = pyo3::exceptions::PyException)]
    pub struct FlyteAuthenticationException {}
    #[pymethods]
    impl FlyteAuthenticationException {
        #[new]
        #[pyo3(signature = (*_args, **_kwargs))]
        fn new(
            _args: Bound<'_, pyo3::types::PyTuple>,
            _kwargs: Option<Bound<'_, pyo3::types::PyDict>>,
        ) -> Self {
            Self {}
        }
    }

    #[pyclass(extends = pyo3::exceptions::PyException)]
    pub struct FlyteInvalidInputException {}
    #[pymethods]
    impl FlyteInvalidInputException {
        #[new]
        #[pyo3(signature = (*_args, **_kwargs))]
        fn new(
            _args: Bound<'_, pyo3::types::PyTuple>,
            _kwargs: Option<Bound<'_, pyo3::types::PyDict>>,
        ) -> Self {
            Self {}
        }
    }

    // A wrapped error indicates failure at communicating between gRPC clients and servers.
    pub struct GRPCError(Status);

    // TODO: Do we need this formatting (to string)?
    impl fmt::Display for GRPCError {
        fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
            write!(f, "{}", self.0)
        }
    }

    impl std::convert::From<Status> for GRPCError {
        fn from(other: Status) -> Self {
            Self(other)
        }
    }

    impl std::convert::From<GRPCError> for PyErr {
        fn from(err: GRPCError) -> Self {
            // Optional: Raise Python primitive error via `PyException::new_err(err.to_string())`
            // Pattern match tonic status code to riase correspond Rust binding error here, and user can catch them later in Python side.
            match err.0.code() {
                tonic::Code::Unauthenticated => {
                    return PyErr::new::<FlyteAuthenticationException, _>(err.to_string())
                }
                tonic::Code::AlreadyExists => {
                    return PyErr::new::<FlyteEntityAlreadyExistsException, _>(err.to_string())
                }
                tonic::Code::NotFound => {
                    return PyErr::new::<FlyteEntityNotExistException, _>(err.to_string())
                }
                tonic::Code::InvalidArgument => {
                    return PyErr::new::<FlyteInvalidInputException, _>(err.to_string())
                }
                _ => return PyErr::new::<PyFlyteUserException, _>(err.to_string()),
            }
        }
    }

    #[pymodule]
    pub mod protobuf {
        #[pymodule_export]
        pub use crate::google::protobuf::{Any, Duration, Struct, Timestamp, Value};
    }

    #[pymodule]
    pub mod core {
        #[pymodule_export]
        use crate::flyteidl::core::{
            Alias, ApproveCondition, ArrayNode, ArtifactId, ArtifactKey, ArtifactQuery,
            ArtifactTag, Binary, Binding, BindingData, BindingDataCollection, BindingDataMap, Blob,
            BlobMetadata, BlobType, BooleanExpression, BranchNode, CatalogArtifactTag,
            CatalogMetadata, ComparisonExpression, CompiledLaunchPlan, CompiledTask,
            CompiledWorkflow, CompiledWorkflowClosure, ConnectionSet, Container, ContainerError,
            ContainerPort, DataLoadingConfig, DynamicJobSpec, EnumType, Error, ErrorDocument,
            ExecutionEnv, ExecutionEnvAssignment, ExecutionError, ExtendedResources, GateNode,
            GpuAccelerator, Granularity, Identifier, Identity, IfBlock, IfElseBlock,
            InputBindingData, IoStrategy, K8sObjectMetadata, K8sPod, KeyValuePair, LabelValue,
            Literal, LiteralCollection, LiteralMap, LiteralType, Node, NodeExecutionIdentifier,
            NodeMetadata, OAuth2Client, OAuth2TokenRequest, Operand, Operator, OutputReference,
            Parameter, ParameterMap, Partitions, Primitive, PromiseAttribute, ResourceType,
            Resources, RetryStrategy, RuntimeBinding, RuntimeMetadata, Scalar, SchemaType, Secret,
            SecurityContext, SignalCondition, SignalIdentifier, SimpleType, SleepCondition, Sql,
            StructuredDataset, StructuredDatasetMetadata, StructuredDatasetType,
            TaskExecutionIdentifier, TaskLog, TaskMetadata, TaskNode, TaskNodeOverrides,
            TaskTemplate, TimePartition, TypeAnnotation, TypeStructure, TypedInterface, Union,
            UnionInfo, UnionType, Variable, VariableMap, Void, WorkflowExecution,
            WorkflowExecutionIdentifier, WorkflowMetadata, WorkflowMetadataDefaults, WorkflowNode,
            WorkflowTemplate,
        };
    }
    #[pymodule]
    pub mod boolean_expression {
        #[pymodule_export]
        use crate::flyteidl::core::boolean_expression::Expr;
    }
    #[pymodule]
    pub mod gpu_accelerator {
        #[pymodule_export]
        use crate::flyteidl::core::gpu_accelerator::PartitionSizeValue;
    }
    #[pymodule]
    pub mod artifact_query {
        #[pymodule_export]
        use crate::flyteidl::core::artifact_query::Identifier;
    }
    #[pymodule]
    pub mod task_log {
        #[pymodule_export]
        use crate::flyteidl::core::task_log::MessageFormat;
    }
    #[pymodule]
    pub mod conjunction_expression {
        #[pymodule_export]
        use crate::flyteidl::core::conjunction_expression::LogicalOperator;
    }
    #[pymodule]
    pub mod operand {
        #[pymodule_export]
        use crate::flyteidl::core::operand::Val;
    }
    #[pymodule]
    pub mod comparison_expression {
        #[pymodule_export]
        use crate::flyteidl::core::comparison_expression::Operator;
    }
    #[pymodule]
    pub mod connection_set {
        #[pymodule_export]
        use crate::flyteidl::core::connection_set::IdList;
    }
    #[pymodule]
    pub mod io_strategy {
        #[pymodule_export]
        use crate::flyteidl::core::io_strategy::{DownloadMode, UploadMode};
    }
    #[pymodule]
    pub mod secret {
        #[pymodule_export]
        use crate::flyteidl::core::secret::MountType;
    }
    #[pymodule]
    pub mod o_auth2_token_request {
        #[pymodule_export]
        use crate::flyteidl::core::o_auth2_token_request::Type;
    }
    #[pymodule]
    pub mod label_value {
        #[pymodule_export]
        use crate::flyteidl::core::label_value::Value;
    }
    #[pymodule]
    pub mod container_error {
        #[pymodule_export]
        use crate::flyteidl::core::container_error::Kind;
    }
    #[pymodule]
    pub mod execution_error {
        #[pymodule_export]
        use crate::flyteidl::core::execution_error::ErrorKind;
    }
    #[pymodule]
    pub mod array_node {
        #[pymodule_export]
        use crate::flyteidl::core::array_node::{
            ExecutionMode, ParallelismOption, SuccessCriteria,
        };
    }
    #[pymodule]
    pub mod catalog_metadata {
        #[pymodule_export]
        use crate::flyteidl::core::catalog_metadata::SourceExecution;
    }
    #[pymodule]
    pub mod parameter {
        #[pymodule_export]
        use crate::flyteidl::core::parameter::Behavior;
    }
    #[pymodule]
    pub mod workflow_metadata {
        #[pymodule_export]
        use crate::flyteidl::core::workflow_metadata::OnFailurePolicy;
    }
    #[pymodule]
    pub mod promise_attribute {
        #[pymodule_export]
        use crate::flyteidl::core::promise_attribute::Value;
    }
    #[pymodule]
    pub mod binding_data {
        #[pymodule_export]
        use crate::flyteidl::core::binding_data::Value;
    }
    #[pymodule]
    pub mod gate_node {
        #[pymodule_export]
        use crate::flyteidl::core::gate_node::Condition;
    }
    #[pymodule]
    pub mod if_else_block {
        #[pymodule_export]
        use crate::flyteidl::core::if_else_block::Default;
    }
    #[pymodule]
    pub mod workflow_node {
        #[pymodule_export]
        use crate::flyteidl::core::workflow_node::Reference;
    }
    #[pymodule]
    pub mod node {
        #[pymodule_export]
        use crate::flyteidl::core::node::Target;
    }
    #[pymodule]
    pub mod task_node {
        #[pymodule_export]
        use crate::flyteidl::core::task_node::Reference;
    }
    #[pymodule]
    pub mod literal {
        #[pymodule_export]
        use crate::flyteidl::core::literal::Value;
    }
    #[pymodule]
    pub mod resources {
        #[pymodule_export]
        use crate::flyteidl::core::resources::{ResourceEntry, ResourceName};
    }
    #[pymodule]
    pub mod primitive {
        #[pymodule_export]
        use crate::flyteidl::core::primitive::Value;
    }
    #[pymodule]
    pub mod literal_type {
        #[pymodule_export]
        use crate::flyteidl::core::literal_type::Type;
    }
    #[pymodule]
    pub mod scalar {
        #[pymodule_export]
        use crate::flyteidl::core::scalar::Value;
    }
    #[pymodule]
    pub mod runtime_metadata {
        #[pymodule_export]
        use crate::flyteidl::core::runtime_metadata::RuntimeType;
    }
    #[pymodule]
    pub mod task_metadata {
        #[pymodule_export]
        use crate::flyteidl::core::task_metadata::InterruptibleValue;
    }
    #[pymodule]
    pub mod node_metadata {
        #[pymodule_export]
        use crate::flyteidl::core::node_metadata::InterruptibleValue;
    }
    #[pymodule]
    pub mod node_execution {
        #[pymodule_export]
        use crate::flyteidl::core::node_execution::Phase;
    }
    #[pymodule]
    pub mod task_execution {
        #[pymodule_export]
        use crate::flyteidl::core::task_execution::Phase;
    }
    #[pymodule]
    pub mod data_loading_config {
        #[pymodule_export]
        use crate::flyteidl::core::data_loading_config::LiteralMapFormat;
    }
    #[pymodule]
    pub mod task_template {
        #[pymodule_export]
        use crate::flyteidl::core::task_template::Target;
    }
    #[pymodule]
    pub mod workflow_execution {
        #[pymodule_export]
        use crate::flyteidl::core::workflow_execution::Phase;
    }
    #[pymodule]
    pub mod schema_type {
        #[pymodule_export]
        use crate::flyteidl::core::schema_type::SchemaColumn;
    }
    #[pymodule]
    pub mod schema_column {
        #[pymodule_export]
        use crate::flyteidl::core::schema_type::schema_column::SchemaColumnType;
    }
    #[pymodule]
    pub mod structured_dataset_type {
        #[pymodule_export]
        use crate::flyteidl::core::structured_dataset_type::DatasetColumn;
    }
    #[pymodule]
    pub mod blob_type {
        #[pymodule_export]
        use crate::flyteidl::core::blob_type::BlobDimensionality;
    }

    #[pymodule]
    pub mod admin {
        #[pymodule_export]
        use crate::flyteidl::admin::{
            AbortMetadata, ActiveLaunchPlanRequest, Annotations, Auth, AuthRole, ClusterAssignment,
            ClusterResourceAttributes, CronSchedule, Description, DescriptionEntity,
            EmailNotification, Envs, Execution, ExecutionClosure, ExecutionClusterLabel,
            ExecutionCreateRequest, ExecutionCreateResponse, ExecutionMetadata,
            ExecutionQueueAttributes, ExecutionRecoverRequest, ExecutionRelaunchRequest,
            ExecutionSpec, ExecutionTerminateRequest, FixedRate, FixedRateUnit, Labels, LaunchPlan,
            LaunchPlanClosure, LaunchPlanCreateRequest, LaunchPlanCreateResponse,
            LaunchPlanMetadata, LaunchPlanSpec, LaunchPlanState, LaunchPlanUpdateRequest,
            ListMatchableAttributesRequest, LiteralMapBlob, MatchableResource,
            NamedEntityIdentifierList, NamedEntityIdentifierListRequest, NamedEntityState,
            NodeExecution, NodeExecutionClosure, NodeExecutionForTaskListRequest,
            NodeExecutionGetDataRequest, NodeExecutionGetDataResponse, NodeExecutionList,
            NodeExecutionListRequest, NodeExecutionMetaData, Notification, NotificationList,
            ObjectGetRequest, PagerDutyNotification, PluginOverride, Project,
            ProjectDomainAttributes, ProjectDomainAttributesGetRequest,
            ProjectDomainAttributesUpdateRequest, ProjectListRequest, ProjectRegisterRequest,
            RawOutputDataConfig, ResourceListRequest, Schedule, Signal, SignalList,
            SignalListRequest, SignalSetRequest, SlackNotification, Sort, SourceCode,
            SystemMetadata, Task, TaskClosure, TaskCreateRequest, TaskCreateResponse,
            TaskExecution, TaskExecutionClosure, TaskExecutionGetDataRequest,
            TaskExecutionGetDataResponse, TaskExecutionGetRequest, TaskExecutionList,
            TaskExecutionListRequest, TaskExecutionMetadata, TaskSpec, UrlBlob, Workflow,
            WorkflowAttributes, WorkflowAttributesGetRequest, WorkflowAttributesUpdateRequest,
            WorkflowClosure, WorkflowCreateRequest, WorkflowCreateResponse,
            WorkflowExecutionGetDataRequest, WorkflowExecutionGetDataResponse,
            WorkflowExecutionGetRequest, WorkflowList, WorkflowNodeMetadata, WorkflowSpec,
        };
    }
    #[pymodule]
    pub mod task_execution_closure {
        #[pymodule_export]
        use crate::flyteidl::admin::task_execution_closure::OutputResult;
    }
    #[pymodule]
    pub mod schedule {
        #[pymodule_export]
        use crate::flyteidl::admin::schedule::ScheduleExpression;
    }
    #[pymodule]
    pub mod description {
        #[pymodule_export]
        use crate::flyteidl::admin::description::Content;
    }
    #[pymodule]
    pub mod execution_metadata {
        #[pymodule_export]
        use crate::flyteidl::admin::execution_metadata::ExecutionMode;
    }
    #[pymodule]
    pub mod node_execution_closure {
        #[pymodule_export]
        use crate::flyteidl::admin::node_execution_closure::{OutputResult, TargetMetadata};
    }
    #[pymodule]
    pub mod execution_spec {
        #[pymodule_export]
        use crate::flyteidl::admin::execution_spec::NotificationOverrides;
    }
    #[pymodule]
    pub mod execution_closure {
        #[pymodule_export]
        use crate::flyteidl::admin::execution_closure::OutputResult;
    }
    #[pymodule]
    pub mod literal_map_blob {
        #[pymodule_export]
        use crate::flyteidl::admin::literal_map_blob::Data;
    }
    #[pymodule]
    pub mod notification {
        #[pymodule_export]
        use crate::flyteidl::admin::notification::Type;
    }
    #[pymodule]
    pub mod plugin_override {
        #[pymodule_export]
        use crate::flyteidl::admin::plugin_override::MissingPluginBehavior;
    }
    #[pymodule]
    pub mod project {
        #[pymodule_export]
        use crate::flyteidl::admin::project::ProjectState;
    }
    #[pymodule]
    pub mod sort {
        #[pymodule_export]
        use crate::flyteidl::admin::sort::Direction;
    }

    #[pymodule]
    pub mod service {
        #[pymodule_export]
        use crate::flyteidl::service::{
            CreateDownloadLocationRequest, CreateDownloadLocationResponse,
            CreateUploadLocationRequest, CreateUploadLocationResponse, GetDataRequest,
            GetDataResponse,
        };
    }

    #[pymodule]
    pub mod get_data_response {
        #[pymodule_export]
        use crate::flyteidl::service::get_data_response::Data;
    }

    #[pymodule]
    pub mod plugins {
        #[pymodule_export]
        use crate::flyteidl::plugins::{
            ArrayJob, HiveQuery, HiveQueryCollection, PrestoQuery, QuboleHiveJob,
        };
    }
    #[pymodule]
    pub mod array_job {
        #[pymodule_export]
        use crate::flyteidl::plugins::array_job::SuccessCriteria;
    }

    // A simple implementation for parsing google protobuf types `Struct` and `Value` from json string after deriving `serde::Deserialize` for structures.
    // This should equivalent to `google.protobuf._json_format.Parse()`.
    // For instance, `google.protobuf._json_format.Parse(_json.dumps(self.custom), flyteidl.protobuf.Struct()) if self.custom else None`

    pub fn parse_json_string(
        json_value: &serde_json::Value,
    ) -> PyResult<super::google::protobuf::Value> {
        use super::google::protobuf::value::Kind;

        let kind = match json_value {
            serde_json::Value::Null => super::google::protobuf::value::Kind::NullValue(0),
            serde_json::Value::Bool(b) => super::google::protobuf::value::Kind::BoolValue(*b),
            serde_json::Value::Number(num) => {
                if let Some(i) = num.as_i64() {
                    super::google::protobuf::value::Kind::NumberValue(i as f64)
                } else if let Some(f) = num.as_f64() {
                    super::google::protobuf::value::Kind::NumberValue(f)
                } else {
                    return Err(PyValueError::new_err("Invalid number type"));
                }
            }
            serde_json::Value::String(s) => Kind::StringValue(s.clone()),
            serde_json::Value::Array(arr) => {
                let values = arr
                    .iter()
                    .map(parse_json_string)
                    .collect::<Result<Vec<super::google::protobuf::Value>, _>>()?;
                super::google::protobuf::value::Kind::ListValue(
                    super::google::protobuf::ListValue { values },
                )
            }
            serde_json::Value::Object(obj) => {
                let fields = obj
                    .iter()
                    .map(|(k, v)| {
                        let value = parse_json_string(v)?;
                        Ok((k.clone(), value))
                    })
                    .collect::<Result<HashMap<String, super::google::protobuf::Value>, PyErr>>()?;
                super::google::protobuf::value::Kind::StructValue(super::google::protobuf::Struct {
                    fields,
                })
            }
        };

        Ok(super::google::protobuf::Value { kind: Some(kind) })
    }

    pub fn dump_protobuf_value(value: &super::google::protobuf::Value) -> PyResult<String> {
        let json_value = match value.kind.as_ref() {
            Some(super::google::protobuf::value::Kind::NullValue(_)) => serde_json::Value::Null,
            Some(super::google::protobuf::value::Kind::NumberValue(v)) => {
                serde_json::Value::Number(serde_json::Number::from_f64(*v).ok_or_else(|| {
                    PyValueError::new_err("Failed to convert number to JSON number")
                })?)
            }
            Some(super::google::protobuf::value::Kind::StringValue(v)) => {
                serde_json::Value::String(v.clone())
            }
            Some(super::google::protobuf::value::Kind::BoolValue(v)) => serde_json::Value::Bool(*v),
            Some(super::google::protobuf::value::Kind::StructValue(v)) => {
                let json_str = dump_struct(v)?;
                serde_json::from_str(&json_str)
                    .map_err(|e| PyValueError::new_err(format!("Failed to parse JSON: {}", e)))?
            }
            Some(super::google::protobuf::value::Kind::ListValue(v)) => {
                let json_values: Result<Vec<serde_json::Value>, PyErr> = v
                    .values
                    .iter()
                    .map(|val| {
                        serde_json::from_str(&dump_protobuf_value(val)?).map_err(|e| {
                            PyValueError::new_err(format!("Failed to parse JSON: {}", e))
                        })
                    })
                    .collect();
                serde_json::Value::Array(json_values?)
            }
            None => return Err(PyValueError::new_err("Invalid or unsupported value type")),
        };

        // Convert the JSON value to a string
        serde_json::to_string(&json_value)
            .map_err(|e| PyValueError::new_err(format!("Failed to serialize JSON: {}", e)))
    }

    // #[pyfunction]
    // #[pyo3(name = "ParseValue")]
    // pub fn parse_value(json_str: &str) -> PyResult<super::google::protobuf::Value> {
    //     let parsed: serde_json::Value = serde_json::from_str(json_str)
    //         .map_err(|e| PyValueError::new_err(format!("Invalid JSON: {}", e)))?;

    //     let value = parse_json_value(&parsed)?;
    //     Ok(value)
    // }

    #[pyfunction]
    #[pyo3(name = "DumpStruct")]
    pub fn dump_struct(proto_struct: &super::google::protobuf::Struct) -> PyResult<String> {
        let mut json_map = serde_json::Map::new();

        for (key, value) in &proto_struct.fields {
            let json_value = serde_json::from_str(&dump_protobuf_value(value)?)
                .map_err(|e| PyValueError::new_err(format!("Failed to parse JSON: {}", e)))?;
            json_map.insert(key.clone(), json_value);
        }

        let json_value = serde_json::Value::Object(json_map);
        serde_json::to_string(&json_value)
            .map_err(|e| PyValueError::new_err(format!("Failed to serialize JSON: {}", e)))
    }

    #[pyfunction]
    #[pyo3(name = "ParseStruct")]
    pub fn parse_struct(json_str: &str) -> PyResult<super::google::protobuf::Struct> {
        let parsed: serde_json::Value = serde_json::from_str(json_str)
            .map_err(|e| PyValueError::new_err(format!("Invalid JSON: {}", e)))?;

        if let serde_json::Value::Object(obj) = parsed {
            let fields = obj
                .iter()
                .map(|(k, v)| {
                    let value = parse_json_string(v)?;
                    Ok((k.clone(), value))
                })
                .collect::<Result<HashMap<String, super::google::protobuf::Value>, PyErr>>()?;
            Ok(super::google::protobuf::Struct { fields })
        } else {
            Err(PyValueError::new_err("Expected a JSON object"))
        }
    }

    // You can also use the `Interceptor` trait to create an interceptor type
    // that is easy to name
    #[derive(Clone)]
    pub struct UnaryAuthInterceptor {
        _access_token: String,
    }

    impl Interceptor for UnaryAuthInterceptor {
        fn call(&mut self, mut request: Request<()>) -> Result<Request<()>, Status> {
            // println!("access_token:\t{}\n", self._access_token.clone());
            let metadata_token: MetadataValue<_> =
                match format!("Bearer {}", self._access_token).parse::<MetadataValue<_>>() {
                    Ok(metadata_token) => metadata_token,
                    Err(error) => panic!("{}", error),
                };
            // println!("metadata_token:\t{:?}\n", metadata_token.clone());
            request
                .metadata_mut()
                .insert("flyte-authorization", metadata_token.clone());
            Ok(request)
        }
    }

    #[pyclass(subclass, name = "RawSynchronousFlyteClient")]
    pub struct RawSynchronousFlyteClient {
        admin_service: crate::flyteidl::service::admin_service_client::AdminServiceClient<
            InterceptedService<Channel, UnaryAuthInterceptor>,
        >,
        data_proxy_service:
            crate::flyteidl::service::data_proxy_service_client::DataProxyServiceClient<
                InterceptedService<Channel, UnaryAuthInterceptor>,
            >,
        runtime: tokio::runtime::Runtime,
    }

    #[pymethods]
    impl RawSynchronousFlyteClient {
        // It's necessary `new` attribute to construct the `RawSynchronousFlyteClient` in Python.
        #[new]
        // TODO: Instead of accepting endpoint and kwargs dict as arguments, we should take path as input that reads platform configuration file.
        #[pyo3(signature = (endpoint, insecure, auth_mode, client_id, client_credentials_secret, **kwargs))]
        pub fn new(
            endpoint: &str,
            insecure: bool,
            auth_mode: &str,
            client_id: &str,
            client_credentials_secret: &str,
            kwargs: Option<&Bound<'_, pyo3::types::PyDict>>,
        ) -> PyResult<RawSynchronousFlyteClient> {
            // Use Atomic Reference Counting abstractions as a cheap way to pass string reference into another thread that outlives the scope.
            let s = std::sync::Arc::new(endpoint);
            // Check details for constructing Tokio asynchronous `runtime`: https://docs.rs/tokio/latest/tokio/runtime/struct.Builder.html#method.new_current_thread
            let rt = match tokio::runtime::Builder::new_current_thread()
                .enable_all()
                .build()
            {
                Ok(rt) => rt,
                Err(error) => panic!("Failed to initiate Tokio multi-thread runtime: {:?}", error),
            };

            let endpoint_uri: Uri = auth::auth::bootstrap_uri_from_endpoint(endpoint, &insecure);
            // println!("RawSynchronousFlyteClient.endpoint: {}", endpoint);
            // let credentials_for_endpoint: &str = endpoint; // TODO: The default key in flytekit is `flyte-default``
            // let credentials_access_token_key: &str = "access_token";
            // println!("{:?}", credentials_for_endpoint);
            // let entry: Entry =
            //     match Entry::new(credentials_for_endpoint, credentials_access_token_key) {
            //         Ok(entry) => entry,
            //         Err(err) => {
            //             println!("{}", credentials_access_token_key);
            //             panic!("Failed at initializing keyring, not available.");
            //         }
            //     };
            // dummy write access_token when unauthenticated
            // match entry.set_password("") {
            //     Ok(()) => println!("KeyRing set successfully."),
            //     Err(err) => println!("KeyRing set not available."),
            // };
            let mut access_token: String = "".to_string();
            if !insecure {
                let cert: Certificate = auth::auth::bootstrap_creds_from_server(&endpoint_uri);
                // let tls: ClientTlsConfig = ClientTlsConfig::new()
                //     .ca_certificate(cert)
                //     .domain_name((*endpoint).to_string());

                // let channel = match rt.block_on(
                //     Channel::builder(endpoint_uri.clone())
                //         .tls_config(tls)
                //         .unwrap()
                //         .connect(),
                // ) {
                //     Ok(ch) => ch,
                //     Err(error) => panic!(
                //         "Failed at connecting to endpoint when constructing secured channel: {:?}",
                //         error
                //     ),
                // };

                let mut oauth_client: auth::auth::OAuthClient = auth::auth::OAuthClient::new(
                    endpoint,
                    &insecure,
                    Some(client_id),
                    Some(client_credentials_secret),
                );
                // TODO: swithch by AuthMode flag
                println!("AuthMode: {:?}", auth_mode);
                if auth_mode == "Pkce" {
                    access_token = oauth_client.pkce_authenticate().unwrap();
                } else if auth_mode == "client_credentials" || auth_mode == "ClientSecret" {
                    access_token = oauth_client.client_secret_authenticate().unwrap();
                } else {
                    todo!();
                }
            }

            // Check details on constructing `channel`: https://docs.rs/tonic/latest/tonic/transport/struct.Channel.html#method.builder
            // `Channel::builder(endpoint_uri)` returns type `tonic::transport::Endpoint`.

            let channel = match rt.block_on(Channel::builder(endpoint_uri.clone()).connect()) {
                Ok(ch) => ch,
                Err(error) => panic!(
                    "Failed at connecting to endpoint when constructing insecured channel: {:?}",
                    error
                ),
            };

            // let access_token: String = match entry.get_password() {
            //     Ok(access_token) => {
            //         println!("keyring retrieved successfully.{}", access_token);
            //         access_token
            //     }
            //     Err(error) => {
            //         println!("Failed at retrieving keyring: {:?}", error);
            //         "".to_string()
            //     }
            // };

            let interceptor: UnaryAuthInterceptor = UnaryAuthInterceptor {
                _access_token: access_token,
            };

            // Binding established channel to the service client stubs.
            let admin_stub =
                crate::flyteidl::service::admin_service_client::AdminServiceClient::with_interceptor(channel.clone(), interceptor.clone());
            let data_proxy_stub =
                crate::flyteidl::service::data_proxy_service_client::DataProxyServiceClient::with_interceptor(channel.clone(), interceptor.clone());

            Ok(RawSynchronousFlyteClient {
                runtime: rt, // The tokio runtime is used in a blocking manner for now.
                admin_service: admin_stub,
                data_proxy_service: data_proxy_stub,
            })
        }

        pub fn get_task(
            &mut self,
            req: crate::flyteidl::admin::ObjectGetRequest,
        ) -> Result<crate::flyteidl::admin::Task, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.get_task(req))?
                .into_inner();
            Ok(res)
        }

        pub fn create_task(
            &mut self,
            req: crate::flyteidl::admin::TaskCreateRequest,
        ) -> Result<crate::flyteidl::admin::TaskCreateResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.create_task(req))?
                .into_inner();
            Ok(res)
        }

        pub fn create_upload_location(
            &mut self,
            req: crate::flyteidl::service::CreateUploadLocationRequest,
        ) -> Result<crate::flyteidl::service::CreateUploadLocationResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.data_proxy_service.create_upload_location(req))?
                .into_inner();
            Ok(res)
        }

        pub fn get_data(
            &mut self,
            req: crate::flyteidl::service::GetDataRequest,
        ) -> Result<crate::flyteidl::service::GetDataResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.data_proxy_service.get_data(req))?
                .into_inner();
            Ok(res)
        }

        pub fn create_execution(
            &mut self,
            req: crate::flyteidl::admin::ExecutionCreateRequest,
        ) -> Result<crate::flyteidl::admin::ExecutionCreateResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.create_execution(req))?
                .into_inner();
            Ok(res)
        }

        pub fn get_execution(
            &mut self,
            req: crate::flyteidl::admin::WorkflowExecutionGetRequest,
        ) -> Result<crate::flyteidl::admin::Execution, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.get_execution(req))?
                .into_inner();
            Ok(res)
        }

        pub fn get_execution_data(
            &mut self,
            req: crate::flyteidl::admin::WorkflowExecutionGetDataRequest,
        ) -> Result<crate::flyteidl::admin::WorkflowExecutionGetDataResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.get_execution_data(req))?
                .into_inner();
            Ok(res)
        }

        pub fn list_node_executions(
            &mut self,
            req: crate::flyteidl::admin::NodeExecutionListRequest,
        ) -> Result<crate::flyteidl::admin::NodeExecutionList, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.list_node_executions(req))?
                .into_inner();
            Ok(res)
        }

        pub fn get_node_execution_data(
            &mut self,
            req: crate::flyteidl::admin::NodeExecutionGetDataRequest,
        ) -> Result<crate::flyteidl::admin::NodeExecutionGetDataResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.get_node_execution_data(req))?
                .into_inner();
            Ok(res)
        }

        pub fn get_task_execution_data(
            &mut self,
            req: crate::flyteidl::admin::TaskExecutionGetDataRequest,
        ) -> Result<crate::flyteidl::admin::TaskExecutionGetDataResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.get_task_execution_data(req))?
                .into_inner();
            Ok(res)
        }

        pub fn list_task_executions(
            &mut self,
            req: crate::flyteidl::admin::TaskExecutionListRequest,
        ) -> Result<crate::flyteidl::admin::TaskExecutionList, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.list_task_executions(req))?
                .into_inner();
            Ok(res)
        }

        pub fn get_task_execution(
            &mut self,
            req: crate::flyteidl::admin::TaskExecutionGetRequest,
        ) -> Result<crate::flyteidl::admin::TaskExecution, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.get_task_execution(req))?
                .into_inner();
            Ok(res)
        }

        pub fn create_workflow(
            &mut self,
            req: crate::flyteidl::admin::WorkflowCreateRequest,
        ) -> Result<crate::flyteidl::admin::WorkflowCreateResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.create_workflow(req))?
                .into_inner();
            Ok(res)
        }

        pub fn list_workflow_ids(
            &mut self,
            req: crate::flyteidl::admin::NamedEntityIdentifierListRequest,
        ) -> Result<crate::flyteidl::admin::NamedEntityIdentifierList, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.list_workflow_ids(req))?
                .into_inner();
            Ok(res)
        }

        pub fn list_workflows(
            &mut self,
            req: crate::flyteidl::admin::ResourceListRequest,
        ) -> Result<crate::flyteidl::admin::WorkflowList, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.list_workflows(req))?
                .into_inner();
            Ok(res)
        }

        pub fn get_workflow(
            &mut self,
            req: crate::flyteidl::admin::ObjectGetRequest,
        ) -> Result<crate::flyteidl::admin::Workflow, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.get_workflow(req))?
                .into_inner();
            Ok(res)
        }

        pub fn create_launch_plan(
            &mut self,
            req: crate::flyteidl::admin::LaunchPlanCreateRequest,
        ) -> Result<crate::flyteidl::admin::LaunchPlanCreateResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.create_launch_plan(req))?
                .into_inner();
            Ok(res)
        }

        pub fn get_launch_plan(
            &mut self,
            req: crate::flyteidl::admin::ObjectGetRequest,
        ) -> Result<crate::flyteidl::admin::LaunchPlan, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.get_launch_plan(req))?
                .into_inner();
            Ok(res)
        }
        pub fn terminate_execution(
            &mut self,
            req: crate::flyteidl::admin::ExecutionTerminateRequest,
        ) -> Result<crate::flyteidl::admin::ExecutionTerminateResponse, GRPCError> {
            let res = self
                .runtime
                .block_on(self.admin_service.terminate_execution(req))?
                .into_inner();
            Ok(res)
        }
    }
}

// At present, we don't need to specify the concrete type for generics parameterization.
// Since we output the whole wrapped `FlyteRemote` as binding class instead of each gRPC stubs.
// So the macro `concrete_generic_structure` is commented out.

// #[macro_export]
// macro_rules! concrete_generic_structure {
//     ($name:ident, $generic:ident, $type:ty, $( ($method_name:ident, $request_type:ty, $response_type:ty) ),* ) => {

//         #[pyo3::pyclass]
//         #[derive(Debug)]
//         pub struct $name {
//             pub inner: $generic<$type>,
//             pub rt:  Runtime,
//         }
//         // If we trying to expose `tonic::transport::Channel` through PyO3 so that we can create a tonic channel in Python and take it as input for our gRPC clients bindings,
//         // It'll violate the orphan ruls in terms of Rust's traits consisitency, because we were trying to implement a external trait for another external crate's structure in our local crates.
//         // Se We may need to define a local structure like `PyClientlWrapper` and it's not necessary to get or set all of its member fields.

//         // The macro mechanically implement our gRPC client stubs.
//         #[pyo3::pymethods]
//         impl $name {

//             // Attempt to create a new client by connecting to a given endpoint.
//             #[new]
//             #[pyo3(signature = (dst))]
//             pub fn new(dst: String) -> Self {
//                 let endpoint = tonic::transport::Endpoint::from_shared(dst).unwrap();
//                 let rt = Builder::new_current_thread().enable_all().build().unwrap();
//                 let client = rt.block_on($generic::<$type>::connect(endpoint)).unwrap();
//                 $name {
//                     inner: client,
//                     rt: rt,
//                 }
//             }

//             // Generate methods for each provided services
//             // Currently in a blocking manner.
//             $(
//                 #[pyo3(signature = (request))]
//                 pub fn $method_name(
//                     &mut self,
//                     request: $request_type,
//                 ) -> Result<$response_type, GRPCError> {
//                     let res = self.rt
//                         .block_on(self.inner.$method_name(request))?.into_inner();
//                 Ok(res)
//                 }
//             )*
//         }
//     };
// }

// A Python module implemented in Rust.
// #[pymodule]
// pub fn _flyteidl_rust(_py: Python, m: &Bound<'_, PyModule>) -> PyResult<()> {

// concrete_generic_structure!(
//     AdminStub,
//     AdminServiceClient,
//     tonic::transport::Channel,
//     (
//         create_task,
//         flyteidl::admin::TaskCreateRequest,
//         flyteidl::admin::TaskCreateResponse
//     ),
//     (
//         get_task,
//         flyteidl::admin::ObjectGetRequest,
//         flyteidl::admin::Task
//     ),
//     (
//         create_execution,
//         flyteidl::admin::ExecutionCreateRequest,
//         flyteidl::admin::ExecutionCreateResponse
//     ),
//     (
//         get_execution,
//         flyteidl::admin::WorkflowExecutionGetRequest,
//         flyteidl::admin::Execution
//     ),
//     (
//         get_execution_data,
//         flyteidl::admin::WorkflowExecutionGetDataRequest,
//         flyteidl::admin::WorkflowExecutionGetDataResponse
//     ),
//     (
//         list_node_executions,
//         flyteidl::admin::NodeExecutionListRequest,
//         flyteidl::admin::NodeExecutionList
//     ),
//     (
//         get_node_execution_data,
//         flyteidl::admin::NodeExecutionGetDataRequest,
//         flyteidl::admin::NodeExecutionGetDataResponse
//     ),
//     (
//         list_task_executions,
//         flyteidl::admin::TaskExecutionListRequest,
//         flyteidl::admin::TaskExecutionList
//     ),
//     (
//         get_task_execution,
//         flyteidl::admin::TaskExecutionGetRequest,
//         flyteidl::admin::TaskExecution
//     ),
//     (
//         get_task_execution_data,
//         flyteidl::admin::TaskExecutionGetDataRequest,
//         flyteidl::admin::TaskExecutionGetDataResponse
//     )
// );

// concrete_generic_structure!(
//     DataProxyStub,
//     DataProxyServiceClient,
//     tonic::transport::Channel,
//     (
//         create_upload_location,
//         flyteidl::service::CreateUploadLocationRequest,
//         flyteidl::service::CreateUploadLocationResponse
//     )
// );
// m.add_class::<AdminStub>();
// Ok(())
// }

// TODO: add Rust level tests
