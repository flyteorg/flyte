use core::str;
pub use pyo3::prelude::*;

#[macro_export]
macro_rules! convert_foreign_tonic_error {
    // Foreign Rust error types: https://pyo3.rs/main/function/error-handling#foreign-rust-error-types
    // Create a newtype wrapper, e.g. MyOtherError. Then implement From<MyOtherError> for PyErr (or PyErrArguments), as well as From<OtherError> for MyOtherError.
    () => {
        use tonic::Status;
        // An error indicates taht failing at communicating between gRPC clients and servers.
        pub struct GRPCError(Status);
        use pyo3::exceptions::PyOSError;
        use pyo3::PyErr;
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
        use std::fmt;

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

#[macro_export]
macro_rules! concrete_generic_structure {
    ($name:ident, $generic:ident, $type:ty, $( ($method_name:ident, $request_type:ty, $response_type:ty) ),* ) => {

        #[pyo3::pyclass]
        #[derive(Debug)]
        pub struct $name {
            pub inner: $generic<$type>,
            pub rt:  Runtime,
        }
        // If we trying to expose `tonic::transport::Channel` through PyO3 so that we can create a tonic channel in Python and take it as input for our gRPC clients bindings,
        // It'll violate the orphan ruls in terms of Rust's traits consisitency, because we were trying to implement a external trait for another external crate's structure in our local crates.
        // Se We may need to define a local structure like `PyClientlWrapper` and it's not necessary to get or set all of its member fields.

        // The macro mechanically implement our gRPC client stubs.
        #[pyo3::pymethods]
        impl $name {

            // Attempt to create a new client by connecting to a given endpoint.
            #[new]
            #[pyo3(signature = (dst))]
            pub fn new(dst: String) -> Self {
                let endpoint = tonic::transport::Endpoint::from_shared(dst).unwrap();
                let rt = Builder::new_current_thread().enable_all().build().unwrap();
                let client = rt.block_on($generic::<$type>::connect(endpoint)).unwrap();
                $name {
                    inner: client,
                    rt: rt,
                }
            }

            // Generate methods for each provided services
            // Currently in a blocking manner.
            $(
                #[pyo3(signature = (request))]
                pub fn $method_name(
                    &mut self,
                    request: $request_type,
                ) -> Result<$response_type, GRPCError> {
                    let res = self.rt
                        .block_on(self.inner.$method_name(request))?.into_inner();
                Ok(res)
                }
            )*
        }
    };
}

pub mod google {
    use pyo3::prelude::*;
    pub mod protobuf {
        include!("../gen/pb_rust/google.protobuf.rs");
    }
}

pub mod flyteidl {
    convert_foreign_prost_error!();

    use pyo3::prelude::*;
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
        #[pymethods]
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
        use pyo3::prelude::*;
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

fn register_wkt_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "wkt")?;
    child_module.add_class::<google::protobuf::Duration>();
    child_module.add_class::<google::protobuf::StringValue>();
    child_module.add_class::<google::protobuf::value::Kind>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_scalar_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "scalar")?;
    child_module.add_class::<flyteidl::core::scalar::Value>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_literal_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "literal")?;
    child_module.add_class::<flyteidl::core::literal::Value>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_primitive_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "primitive")?;
    child_module.add_class::<flyteidl::core::primitive::Value>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_literal_type_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "literal_type")?;
    child_module.add_class::<flyteidl::core::literal_type::Type>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_workflow_execution_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "workflow_execution")?;
    child_module.add_class::<flyteidl::core::workflow_execution::Phase>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_node_execution_closure_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "node_execution_closure")?;
    child_module.add_class::<flyteidl::admin::node_execution_closure::TargetMetadata>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_node_execution_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "node_execution")?;
    child_module.add_class::<flyteidl::core::node_execution::Phase>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_task_execution_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let child_module = PyModule::new_bound(parent_module.py(), "task_execution")?;
    child_module.add_class::<flyteidl::core::task_execution::Phase>();
    parent_module.add_submodule(&child_module)?;
    Ok(())
}

fn register_core_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let m = PyModule::new_bound(parent_module.py(), "core")?;

    m.add_class::<flyteidl::core::literal_type::Type>();
    m.add_class::<flyteidl::core::LiteralType>();
    m.add_class::<flyteidl::core::Literal>();
    m.add_class::<flyteidl::core::LiteralMap>();
    m.add_class::<flyteidl::core::ExtendedResources>();
    m.add_class::<flyteidl::core::Resources>();
    m.add_class::<flyteidl::core::resources::ResourceName>();
    m.add_class::<flyteidl::core::resources::ResourceEntry>();
    m.add_class::<flyteidl::core::ResourceType>();
    m.add_class::<flyteidl::core::ContainerPort>();
    m.add_class::<flyteidl::core::KeyValuePair>();
    m.add_class::<flyteidl::core::task_template::Target>();
    m.add_class::<flyteidl::core::Container>();
    m.add_class::<flyteidl::core::task_metadata::InterruptibleValue>();
    m.add_class::<flyteidl::core::SecurityContext>();
    m.add_class::<flyteidl::core::WorkflowExecutionIdentifier>();
    m.add_class::<flyteidl::core::task_node::Reference>();
    m.add_class::<flyteidl::core::TaskNodeOverrides>();
    m.add_class::<flyteidl::core::TaskNode>();
    m.add_class::<flyteidl::core::NodeMetadata>();
    m.add_class::<flyteidl::core::Variable>();
    m.add_class::<flyteidl::core::VariableMap>();
    m.add_class::<flyteidl::core::TypedInterface>();
    m.add_class::<flyteidl::core::RetryStrategy>();
    m.add_class::<flyteidl::core::RuntimeMetadata>();
    m.add_class::<flyteidl::core::runtime_metadata::RuntimeType>();
    m.add_class::<flyteidl::core::TaskMetadata>();
    m.add_class::<flyteidl::core::Identifier>();
    m.add_class::<flyteidl::core::TaskTemplate>();
    m.add_class::<flyteidl::core::ParameterMap>();
    m.add_class::<flyteidl::core::Parameter>();
    m.add_class::<flyteidl::core::NodeExecutionIdentifier>();
    m.add_class::<flyteidl::core::Primitive>();
    m.add_class::<flyteidl::core::SimpleType>();
    m.add_class::<flyteidl::core::Scalar>();
    m.add_class::<flyteidl::core::WorkflowExecution>();
    m.add_class::<flyteidl::core::ExecutionError>();
    m.add_class::<flyteidl::core::workflow_execution::Phase>();
    m.add_class::<flyteidl::core::scalar::Value>();
    m.add_class::<flyteidl::core::primitive::Value>();
    m.add_class::<flyteidl::core::Node>();
    m.add_class::<flyteidl::core::ArrayNode>();

    parent_module.add_submodule(&m)?;
    Ok(())
}

fn register_admin_submodule(parent_module: &Bound<'_, PyModule>) -> PyResult<()> {
    let m = PyModule::new_bound(parent_module.py(), "admin")?;

    m.add_class::<flyteidl::admin::Labels>();
    m.add_class::<flyteidl::admin::Annotations>();
    m.add_class::<flyteidl::admin::AuthRole>();
    m.add_class::<flyteidl::admin::RawOutputDataConfig>();
    m.add_class::<flyteidl::admin::Envs>();
    m.add_class::<flyteidl::admin::ClusterAssignment>();
    m.add_class::<flyteidl::admin::ExecutionMetadata>();
    m.add_class::<flyteidl::admin::ExecutionSpec>();
    m.add_class::<flyteidl::admin::ExecutionCreateRequest>();
    m.add_class::<flyteidl::admin::ExecutionCreateResponse>();
    m.add_class::<flyteidl::admin::Description>();
    m.add_class::<flyteidl::admin::DescriptionEntity>();
    m.add_class::<flyteidl::admin::description::Content>();
    m.add_class::<flyteidl::admin::TaskSpec>();
    m.add_class::<flyteidl::admin::ObjectGetRequest>();
    m.add_class::<flyteidl::admin::Task>();
    m.add_class::<flyteidl::admin::TaskCreateRequest>();
    m.add_class::<flyteidl::admin::TaskCreateResponse>();
    m.add_class::<flyteidl::admin::SourceCode>();
    m.add_class::<flyteidl::admin::Execution>();
    m.add_class::<flyteidl::admin::execution_metadata::ExecutionMode>();
    m.add_class::<flyteidl::admin::WorkflowExecutionGetRequest>();
    m.add_class::<flyteidl::admin::ExecutionClusterLabel>();
    m.add_class::<flyteidl::admin::WorkflowExecutionGetDataRequest>();
    m.add_class::<flyteidl::admin::WorkflowExecutionGetDataResponse>();
    m.add_class::<flyteidl::admin::NodeExecutionListRequest>();
    m.add_class::<flyteidl::admin::NodeExecutionList>();
    m.add_class::<flyteidl::admin::NodeExecution>();
    m.add_class::<flyteidl::admin::NodeExecutionGetDataRequest>();
    m.add_class::<flyteidl::admin::NodeExecutionGetDataResponse>();
    m.add_class::<flyteidl::admin::TaskExecutionListRequest>();
    m.add_class::<flyteidl::admin::TaskExecutionList>();
    m.add_class::<flyteidl::admin::TaskExecutionGetRequest>();
    m.add_class::<flyteidl::admin::TaskExecution>();
    m.add_class::<flyteidl::admin::TaskExecutionGetDataRequest>();
    m.add_class::<flyteidl::admin::TaskExecutionGetDataResponse>();

    parent_module.add_submodule(&m)?;
    Ok(())
}
// A Python module implemented in Rust.
#[pymodule]
pub fn _flyteidl_rust(_py: Python, m: &Bound<'_, PyModule>) -> PyResult<()> {
    register_wkt_submodule(m);
    register_core_submodule(m);
    register_admin_submodule(m);
    register_scalar_submodule(m);
    register_literal_submodule(m);
    register_primitive_submodule(m);
    register_literal_type_submodule(m);
    register_workflow_execution_submodule(m);
    register_node_execution_closure_submodule(m);
    register_node_execution_submodule(m);
    register_task_execution_submodule(m);

    use tokio::runtime::{Builder, Runtime};

    use flyteidl::service::admin_service_client::AdminServiceClient;
    use flyteidl::service::data_proxy_service_client::DataProxyServiceClient;

    convert_foreign_tonic_error!();

    concrete_generic_structure!(
        AdminStub,
        AdminServiceClient,
        tonic::transport::Channel,
        (
            create_task,
            flyteidl::admin::TaskCreateRequest,
            flyteidl::admin::TaskCreateResponse
        ),
        (
            get_task,
            flyteidl::admin::ObjectGetRequest,
            flyteidl::admin::Task
        ),
        (
            create_execution,
            flyteidl::admin::ExecutionCreateRequest,
            flyteidl::admin::ExecutionCreateResponse
        ),
        (
            get_execution,
            flyteidl::admin::WorkflowExecutionGetRequest,
            flyteidl::admin::Execution
        ),
        (
            get_execution_data,
            flyteidl::admin::WorkflowExecutionGetDataRequest,
            flyteidl::admin::WorkflowExecutionGetDataResponse
        ),
        (
            list_node_executions,
            flyteidl::admin::NodeExecutionListRequest,
            flyteidl::admin::NodeExecutionList
        ),
        (
            get_node_execution_data,
            flyteidl::admin::NodeExecutionGetDataRequest,
            flyteidl::admin::NodeExecutionGetDataResponse
        ),
        (
            list_task_executions,
            flyteidl::admin::TaskExecutionListRequest,
            flyteidl::admin::TaskExecutionList
        ),
        (
            get_task_execution,
            flyteidl::admin::TaskExecutionGetRequest,
            flyteidl::admin::TaskExecution
        ),
        (
            get_task_execution_data,
            flyteidl::admin::TaskExecutionGetDataRequest,
            flyteidl::admin::TaskExecutionGetDataResponse
        )
    );

    concrete_generic_structure!(
        DataProxyStub,
        DataProxyServiceClient,
        tonic::transport::Channel,
        (
            create_upload_location,
            flyteidl::service::CreateUploadLocationRequest,
            flyteidl::service::CreateUploadLocationResponse
        )
    );

    m.add_class::<AdminStub>();
    m.add_class::<DataProxyStub>();

    m.add_class::<flyteidl::service::CreateUploadLocationRequest>();
    m.add_class::<flyteidl::service::CreateUploadLocationResponse>();

    Ok(())
}
