#[macro_export]
macro_rules! convert_foreign_tonic_error {
    // Foreign Rust error types: https://pyo3.rs/main/function/error-handling#foreign-rust-error-types
    // Create a newtype wrapper, e.g. MyOtherError. Then implement From<MyOtherError> for PyErr (or PyErrArguments), as well as From<OtherError> for MyOtherError.
    () => {
        use tonic::Status;
        // An error indicates taht failing at communicating between gRPC clients and servers.
        pub struct GRPCError(Status);
        use std::fmt;

        impl fmt::Display for GRPCError {
            fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
                write!(f, "GRPCError: {}", self.0)
            }
        }

        impl std::convert::From<Status> for GRPCError {
            fn from(other: Status) -> Self {
                Self(other)
            }
        }

        impl std::convert::From<GRPCError> for PyErr {
            fn from(err: GRPCError) -> PyErr {
                PyOSError::new_err(err.to_string())
            }
        }
    };
}

#[macro_export]
macro_rules! convert_foreign_prost_error {
    // Foreign Rust error types: https://pyo3.rs/main/function/error-handling#foreign-rust-error-types
    // Create a newtype wrapper, e.g. MyOtherError. Then implement From<MyOtherError> for PyErr (or PyErrArguments), as well as From<OtherError> for MyOtherError.
    () => {
        use prost::{DecodeError, EncodeError, Message};
        use pyo3::exceptions::PyOSError;
        use pyo3::types::PyBytes;
        use pyo3::PyErr;

        // An error indicates taht failing at serializing object to bytes string, like `SerializTOString()` for python protos.
        pub struct MessageEncodeError(EncodeError);
        // An error indicates taht failing at deserializing object from bytes string, like `ParseFromString()` for python protos.
        pub struct MessageDecodeError(DecodeError);

        impl fmt::Display for MessageEncodeError {
            fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
                write!(f, "")
            }
        }

        impl fmt::Display for MessageDecodeError {
            fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
                write!(f, "")
            }
        }

        impl std::convert::From<MessageEncodeError> for PyErr {
            fn from(err: MessageEncodeError) -> PyErr {
                PyOSError::new_err(err.to_string())
            }
        }

        impl std::convert::From<MessageDecodeError> for PyErr {
            fn from(err: MessageDecodeError) -> PyErr {
                PyOSError::new_err(err.to_string())
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
    };
}

pub mod google {
    pub mod protobuf {
        include!("../gen/pb_rust/google.protobuf.rs");
    }
}

pub mod flyteidl {
    pub mod admin {
        include!("../gen/pb_rust/flyteidl.admin.rs");
    }
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

#[pymodule]
// #[pyo3(name="_flyteidl_rust")]
pub mod _flyteidl_rust {
    pub use pyo3::prelude::*;

    convert_foreign_tonic_error!();
    convert_foreign_prost_error!();

    #[pymodule]
    pub mod protobuf {
        #[pymodule_export]
        use crate::google::protobuf::{Any, Duration, Struct, Timestamp, Value};
    }

    #[pymodule]
    pub mod core {
        #[pymodule_export]
        use crate::flyteidl::core::{
            Alias, ApproveCondition, ArrayNode, Binary, Binding, BindingData,
            BindingDataCollection, BlobMetadata, BlobType, BooleanExpression, BranchNode,
            CompiledLaunchPlan, CompiledTask, CompiledWorkflow, CompiledWorkflowClosure, Container,
            ContainerPort, DataLoadingConfig, Error, ExecutionEnv, ExecutionEnvAssignment,
            ExecutionError, ExtendedResources, GateNode, Identifier, IfBlock, IfElseBlock,
            IoStrategy, KeyValuePair, Literal, LiteralMap, LiteralType, Node,
            NodeExecutionIdentifier, NodeMetadata, OutputReference, Parameter, ParameterMap,
            Primitive, PromiseAttribute, ResourceType, Resources, RetryStrategy, RuntimeMetadata,
            Scalar, SchemaType, SecurityContext, SignalCondition, SimpleType, SleepCondition,
            StructuredDataset, StructuredDatasetMetadata, StructuredDatasetType, TaskMetadata,
            TaskNode, TaskNodeOverrides, TaskTemplate, TypeAnnotation, TypeStructure,
            TypedInterface, Union, UnionInfo, Variable, VariableMap, WorkflowExecution,
            WorkflowExecutionIdentifier, WorkflowMetadata, WorkflowMetadataDefaults, WorkflowNode,
            WorkflowTemplate,
        };
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
        use crate::flyteidl::core::resources::ResourceName;
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
            AbortMetadata, Annotations, AuthRole, ClusterAssignment, Description,
            DescriptionEntity, Envs, Execution, ExecutionClosure, ExecutionClusterLabel,
            ExecutionCreateRequest, ExecutionCreateResponse, ExecutionMetadata, ExecutionSpec,
            Labels, LaunchPlan, LaunchPlanCreateRequest, LaunchPlanCreateResponse,
            LaunchPlanMetadata, LaunchPlanSpec, LiteralMapBlob, NamedEntityIdentifierList,
            NamedEntityIdentifierListRequest, NodeExecution, NodeExecutionGetDataRequest,
            NodeExecutionGetDataResponse, NodeExecutionList, NodeExecutionListRequest,
            Notification, NotificationList, ObjectGetRequest, RawOutputDataConfig,
            ResourceListRequest, Schedule, SourceCode, SystemMetadata, Task, TaskClosure,
            TaskCreateRequest, TaskCreateResponse, TaskExecution, TaskExecutionGetDataRequest,
            TaskExecutionGetDataResponse, TaskExecutionGetRequest, TaskExecutionList,
            TaskExecutionListRequest, TaskSpec, Workflow, WorkflowClosure, WorkflowCreateRequest,
            WorkflowCreateResponse, WorkflowExecutionGetDataRequest,
            WorkflowExecutionGetDataResponse, WorkflowExecutionGetRequest, WorkflowList,
            WorkflowSpec,
        };
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
        use crate::flyteidl::admin::node_execution_closure::TargetMetadata;
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
    pub mod service {
        #[pymodule_export]
        use crate::flyteidl::service::{CreateUploadLocationRequest, CreateUploadLocationResponse};
    }

    #[pyclass(subclass, name = "RawSynchronousFlyteClient")]
    pub struct RawSynchronousFlyteClient {
        admin_service: crate::flyteidl::service::admin_service_client::AdminServiceClient<
            tonic::transport::Channel,
        >,
        data_proxy_service:
            crate::flyteidl::service::data_proxy_service_client::DataProxyServiceClient<
                tonic::transport::Channel,
            >,
        runtime: tokio::runtime::Runtime,
    }

    #[pymethods]
    impl RawSynchronousFlyteClient {
        // We need this attribute to construct the `RawSynchronousFlyteClient` in Python.
        #[new]
        // TODO: Instead of accepting endpoint and kwargs dict as arguments, we should accept a path that reads platform configuration file.
        #[pyo3(signature = (endpoint, **kwargs))]
        pub fn new(
            endpoint: &str,
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
            // Check details for constructing `channel`: https://docs.rs/tonic/latest/tonic/transport/struct.Channel.html#method.builder
            // TODO: generally handle more protocols, like the secured one, i.e., `https://`
            let endpoint_uri =
                match format!("http://{}", *s.clone()).parse::<tonic::transport::Uri>() {
                    Ok(uri) => uri,
                    Err(error) => panic!(
                        "Got invalid endpoint when parsing endpoint_uri: {:?}",
                        error
                    ),
                };
            // `Channel::builder(endpoint_uri)` returns type `tonic::transport::Endpoint`.
            let channel =
                match rt.block_on(tonic::transport::Channel::builder(endpoint_uri).connect()) {
                    Ok(ch) => ch,
                    Err(error) => panic!(
                        "Failed at connecting to endpoint when constructing channel: {:?}",
                        error
                    ),
                };
            // Binding connected channel into service client stubs.
            let admin_stub =
                crate::flyteidl::service::admin_service_client::AdminServiceClient::new(
                    channel.clone(),
                );
            let data_proxy_stub =
                crate::flyteidl::service::data_proxy_service_client::DataProxyServiceClient::new(
                    channel.clone(),
                );
            Ok(RawSynchronousFlyteClient {
                runtime: rt, // The tokio runtime is used in a blocking manner for now.
                admin_service: admin_stub,
                data_proxy_service: data_proxy_stub,
            })
        }

        pub fn get_task(
            &mut self,
            req: crate::flyteidl::admin::ObjectGetRequest,
        ) -> PyResult<crate::flyteidl::admin::Task> {
            let res = (match self.runtime.block_on(self.admin_service.get_task(req)) {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn create_task(
            &mut self,
            req: crate::flyteidl::admin::TaskCreateRequest,
        ) -> PyResult<crate::flyteidl::admin::TaskCreateResponse> {
            let res = (match self.runtime.block_on(self.admin_service.create_task(req)) {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn create_upload_location(
            &mut self,
            req: crate::flyteidl::service::CreateUploadLocationRequest,
        ) -> PyResult<crate::flyteidl::service::CreateUploadLocationResponse> {
            let res = (match self
                .runtime
                .block_on(self.data_proxy_service.create_upload_location(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn create_execution(
            &mut self,
            req: crate::flyteidl::admin::ExecutionCreateRequest,
        ) -> PyResult<crate::flyteidl::admin::ExecutionCreateResponse> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.create_execution(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn get_execution(
            &mut self,
            req: crate::flyteidl::admin::WorkflowExecutionGetRequest,
        ) -> PyResult<crate::flyteidl::admin::Execution> {
            let res = (match self.runtime.block_on(self.admin_service.get_execution(req)) {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn get_execution_data(
            &mut self,
            req: crate::flyteidl::admin::WorkflowExecutionGetDataRequest,
        ) -> PyResult<crate::flyteidl::admin::WorkflowExecutionGetDataResponse> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.get_execution_data(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn list_node_executions(
            &mut self,
            req: crate::flyteidl::admin::NodeExecutionListRequest,
        ) -> PyResult<crate::flyteidl::admin::NodeExecutionList> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.list_node_executions(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn get_node_execution_data(
            &mut self,
            req: crate::flyteidl::admin::NodeExecutionGetDataRequest,
        ) -> PyResult<crate::flyteidl::admin::NodeExecutionGetDataResponse> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.get_node_execution_data(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn list_task_executions(
            &mut self,
            req: crate::flyteidl::admin::TaskExecutionListRequest,
        ) -> PyResult<crate::flyteidl::admin::TaskExecutionList> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.list_task_executions(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn get_task_execution(
            &mut self,
            req: crate::flyteidl::admin::TaskExecutionGetRequest,
        ) -> PyResult<crate::flyteidl::admin::TaskExecution> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.get_task_execution(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn create_workflow(
            &mut self,
            req: crate::flyteidl::admin::WorkflowCreateRequest,
        ) -> PyResult<crate::flyteidl::admin::WorkflowCreateResponse> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.create_workflow(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn list_workflow_ids(
            &mut self,
            req: crate::flyteidl::admin::NamedEntityIdentifierListRequest,
        ) -> PyResult<crate::flyteidl::admin::NamedEntityIdentifierList> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.list_workflow_ids(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn list_workflows(
            &mut self,
            req: crate::flyteidl::admin::ResourceListRequest,
        ) -> PyResult<crate::flyteidl::admin::WorkflowList> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.list_workflows(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn get_workflow(
            &mut self,
            req: crate::flyteidl::admin::ObjectGetRequest,
        ) -> PyResult<crate::flyteidl::admin::Workflow> {
            let res = (match self.runtime.block_on(self.admin_service.get_workflow(req)) {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn create_launch_plan(
            &mut self,
            req: crate::flyteidl::admin::LaunchPlanCreateRequest,
        ) -> PyResult<crate::flyteidl::admin::LaunchPlanCreateResponse> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.create_launch_plan(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }

        pub fn get_launch_plan(
            &mut self,
            req: crate::flyteidl::admin::ObjectGetRequest,
        ) -> PyResult<crate::flyteidl::admin::LaunchPlan> {
            let res = (match self
                .runtime
                .block_on(self.admin_service.get_launch_plan(req))
            {
                Ok(res) => res,
                Err(error) => panic!("Error responsed from gRPC server: {:?}", error),
            })
            .into_inner();
            Ok(res)
        }
    }
}

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
