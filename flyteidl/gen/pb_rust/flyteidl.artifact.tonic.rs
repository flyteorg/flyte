// @generated
/// Generated client implementations.
pub mod artifact_registry_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    ///
    #[derive(Debug, Clone)]
    pub struct ArtifactRegistryClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl ArtifactRegistryClient<tonic::transport::Channel> {
        /// Attempt to create a new client by connecting to a given endpoint.
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> ArtifactRegistryClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::Error: Into<StdError>,
        T::ResponseBody: Body<Data = Bytes> + Send + 'static,
        <T::ResponseBody as Body>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_origin(inner: T, origin: Uri) -> Self {
            let inner = tonic::client::Grpc::with_origin(inner, origin);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> ArtifactRegistryClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T::ResponseBody: Default,
            T: tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
                Response = http::Response<
                    <T as tonic::client::GrpcService<tonic::body::BoxBody>>::ResponseBody,
                >,
            >,
            <T as tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
            >>::Error: Into<StdError> + Send + Sync,
        {
            ArtifactRegistryClient::new(InterceptedService::new(inner, interceptor))
        }
        /// Compress requests with the given encoding.
        ///
        /// This requires the server to support it otherwise it might respond with an
        /// error.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.send_compressed(encoding);
            self
        }
        /// Enable decompressing responses.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.accept_compressed(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_decoding_message_size(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_encoding_message_size(limit);
            self
        }
        ///
        pub async fn create_artifact(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateArtifactRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateArtifactResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/CreateArtifact",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "CreateArtifact",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_artifact(
            &mut self,
            request: impl tonic::IntoRequest<super::GetArtifactRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetArtifactResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/GetArtifact",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.artifact.ArtifactRegistry", "GetArtifact"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn search_artifacts(
            &mut self,
            request: impl tonic::IntoRequest<super::SearchArtifactsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SearchArtifactsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/SearchArtifacts",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "SearchArtifacts",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn create_trigger(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateTriggerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateTriggerResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/CreateTrigger",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "CreateTrigger",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn activate_trigger(
            &mut self,
            request: impl tonic::IntoRequest<super::ActivateTriggerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ActivateTriggerResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/ActivateTrigger",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "ActivateTrigger",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn deactivate_trigger(
            &mut self,
            request: impl tonic::IntoRequest<super::DeactivateTriggerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeactivateTriggerResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/DeactivateTrigger",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "DeactivateTrigger",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn deactivate_all_triggers(
            &mut self,
            request: impl tonic::IntoRequest<super::DeactivateAllTriggersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeactivateAllTriggersResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/DeactivateAllTriggers",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "DeactivateAllTriggers",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_card(
            &mut self,
            request: impl tonic::IntoRequest<super::GetCardRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCardResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/GetCard",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.artifact.ArtifactRegistry", "GetCard"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn add_tag(
            &mut self,
            request: impl tonic::IntoRequest<super::AddTagRequest>,
        ) -> std::result::Result<tonic::Response<super::AddTagResponse>, tonic::Status> {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/AddTag",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.artifact.ArtifactRegistry", "AddTag"));
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn register_producer(
            &mut self,
            request: impl tonic::IntoRequest<super::RegisterProducerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RegisterResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/RegisterProducer",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "RegisterProducer",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn register_consumer(
            &mut self,
            request: impl tonic::IntoRequest<super::RegisterConsumerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RegisterResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/RegisterConsumer",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "RegisterConsumer",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn set_execution_inputs(
            &mut self,
            request: impl tonic::IntoRequest<super::ExecutionInputsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ExecutionInputsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/SetExecutionInputs",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "SetExecutionInputs",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn find_by_workflow_exec(
            &mut self,
            request: impl tonic::IntoRequest<super::FindByWorkflowExecRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SearchArtifactsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/FindByWorkflowExec",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "FindByWorkflowExec",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn list_usage(
            &mut self,
            request: impl tonic::IntoRequest<super::ListUsageRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListUsageResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/ListUsage",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.artifact.ArtifactRegistry", "ListUsage"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_triggering_artifacts(
            &mut self,
            request: impl tonic::IntoRequest<super::GetTriggeringArtifactsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetTriggeringArtifactsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/GetTriggeringArtifacts",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "GetTriggeringArtifacts",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_triggered_executions_by_artifact(
            &mut self,
            request: impl tonic::IntoRequest<
                super::GetTriggeredExecutionsByArtifactRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::GetTriggeredExecutionsByArtifactResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/flyteidl.artifact.ArtifactRegistry/GetTriggeredExecutionsByArtifact",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.artifact.ArtifactRegistry",
                        "GetTriggeredExecutionsByArtifact",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod artifact_registry_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with ArtifactRegistryServer.
    #[async_trait]
    pub trait ArtifactRegistry: Send + Sync + 'static {
        ///
        async fn create_artifact(
            &self,
            request: tonic::Request<super::CreateArtifactRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateArtifactResponse>,
            tonic::Status,
        >;
        ///
        async fn get_artifact(
            &self,
            request: tonic::Request<super::GetArtifactRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetArtifactResponse>,
            tonic::Status,
        >;
        ///
        async fn search_artifacts(
            &self,
            request: tonic::Request<super::SearchArtifactsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SearchArtifactsResponse>,
            tonic::Status,
        >;
        ///
        async fn create_trigger(
            &self,
            request: tonic::Request<super::CreateTriggerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateTriggerResponse>,
            tonic::Status,
        >;
        ///
        async fn activate_trigger(
            &self,
            request: tonic::Request<super::ActivateTriggerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ActivateTriggerResponse>,
            tonic::Status,
        >;
        ///
        async fn deactivate_trigger(
            &self,
            request: tonic::Request<super::DeactivateTriggerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeactivateTriggerResponse>,
            tonic::Status,
        >;
        ///
        async fn deactivate_all_triggers(
            &self,
            request: tonic::Request<super::DeactivateAllTriggersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeactivateAllTriggersResponse>,
            tonic::Status,
        >;
        ///
        async fn get_card(
            &self,
            request: tonic::Request<super::GetCardRequest>,
        ) -> std::result::Result<tonic::Response<super::GetCardResponse>, tonic::Status>;
        ///
        async fn add_tag(
            &self,
            request: tonic::Request<super::AddTagRequest>,
        ) -> std::result::Result<tonic::Response<super::AddTagResponse>, tonic::Status>;
        ///
        async fn register_producer(
            &self,
            request: tonic::Request<super::RegisterProducerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RegisterResponse>,
            tonic::Status,
        >;
        ///
        async fn register_consumer(
            &self,
            request: tonic::Request<super::RegisterConsumerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RegisterResponse>,
            tonic::Status,
        >;
        ///
        async fn set_execution_inputs(
            &self,
            request: tonic::Request<super::ExecutionInputsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ExecutionInputsResponse>,
            tonic::Status,
        >;
        ///
        async fn find_by_workflow_exec(
            &self,
            request: tonic::Request<super::FindByWorkflowExecRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SearchArtifactsResponse>,
            tonic::Status,
        >;
        ///
        async fn list_usage(
            &self,
            request: tonic::Request<super::ListUsageRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListUsageResponse>,
            tonic::Status,
        >;
        ///
        async fn get_triggering_artifacts(
            &self,
            request: tonic::Request<super::GetTriggeringArtifactsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetTriggeringArtifactsResponse>,
            tonic::Status,
        >;
        ///
        async fn get_triggered_executions_by_artifact(
            &self,
            request: tonic::Request<super::GetTriggeredExecutionsByArtifactRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetTriggeredExecutionsByArtifactResponse>,
            tonic::Status,
        >;
    }
    ///
    #[derive(Debug)]
    pub struct ArtifactRegistryServer<T: ArtifactRegistry> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: ArtifactRegistry> ArtifactRegistryServer<T> {
        pub fn new(inner: T) -> Self {
            Self::from_arc(Arc::new(inner))
        }
        pub fn from_arc(inner: Arc<T>) -> Self {
            let inner = _Inner(inner);
            Self {
                inner,
                accept_compression_encodings: Default::default(),
                send_compression_encodings: Default::default(),
                max_decoding_message_size: None,
                max_encoding_message_size: None,
            }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> InterceptedService<Self, F>
        where
            F: tonic::service::Interceptor,
        {
            InterceptedService::new(Self::new(inner), interceptor)
        }
        /// Enable decompressing requests with the given encoding.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.accept_compression_encodings.enable(encoding);
            self
        }
        /// Compress responses with the given encoding, if the client supports it.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.send_compression_encodings.enable(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.max_decoding_message_size = Some(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.max_encoding_message_size = Some(limit);
            self
        }
    }
    impl<T, B> tonic::codegen::Service<http::Request<B>> for ArtifactRegistryServer<T>
    where
        T: ArtifactRegistry,
        B: Body + Send + 'static,
        B::Error: Into<StdError> + Send + 'static,
    {
        type Response = http::Response<tonic::body::BoxBody>;
        type Error = std::convert::Infallible;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(
            &mut self,
            _cx: &mut Context<'_>,
        ) -> Poll<std::result::Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            let inner = self.inner.clone();
            match req.uri().path() {
                "/flyteidl.artifact.ArtifactRegistry/CreateArtifact" => {
                    #[allow(non_camel_case_types)]
                    struct CreateArtifactSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::CreateArtifactRequest>
                    for CreateArtifactSvc<T> {
                        type Response = super::CreateArtifactResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateArtifactRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::create_artifact(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = CreateArtifactSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/GetArtifact" => {
                    #[allow(non_camel_case_types)]
                    struct GetArtifactSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::GetArtifactRequest>
                    for GetArtifactSvc<T> {
                        type Response = super::GetArtifactResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetArtifactRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::get_artifact(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetArtifactSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/SearchArtifacts" => {
                    #[allow(non_camel_case_types)]
                    struct SearchArtifactsSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::SearchArtifactsRequest>
                    for SearchArtifactsSvc<T> {
                        type Response = super::SearchArtifactsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::SearchArtifactsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::search_artifacts(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = SearchArtifactsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/CreateTrigger" => {
                    #[allow(non_camel_case_types)]
                    struct CreateTriggerSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::CreateTriggerRequest>
                    for CreateTriggerSvc<T> {
                        type Response = super::CreateTriggerResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateTriggerRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::create_trigger(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = CreateTriggerSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/ActivateTrigger" => {
                    #[allow(non_camel_case_types)]
                    struct ActivateTriggerSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::ActivateTriggerRequest>
                    for ActivateTriggerSvc<T> {
                        type Response = super::ActivateTriggerResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ActivateTriggerRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::activate_trigger(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ActivateTriggerSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/DeactivateTrigger" => {
                    #[allow(non_camel_case_types)]
                    struct DeactivateTriggerSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::DeactivateTriggerRequest>
                    for DeactivateTriggerSvc<T> {
                        type Response = super::DeactivateTriggerResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeactivateTriggerRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::deactivate_trigger(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = DeactivateTriggerSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/DeactivateAllTriggers" => {
                    #[allow(non_camel_case_types)]
                    struct DeactivateAllTriggersSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::DeactivateAllTriggersRequest>
                    for DeactivateAllTriggersSvc<T> {
                        type Response = super::DeactivateAllTriggersResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeactivateAllTriggersRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::deactivate_all_triggers(
                                        &inner,
                                        request,
                                    )
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = DeactivateAllTriggersSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/GetCard" => {
                    #[allow(non_camel_case_types)]
                    struct GetCardSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::GetCardRequest>
                    for GetCardSvc<T> {
                        type Response = super::GetCardResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetCardRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::get_card(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetCardSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/AddTag" => {
                    #[allow(non_camel_case_types)]
                    struct AddTagSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::AddTagRequest>
                    for AddTagSvc<T> {
                        type Response = super::AddTagResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::AddTagRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::add_tag(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = AddTagSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/RegisterProducer" => {
                    #[allow(non_camel_case_types)]
                    struct RegisterProducerSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::RegisterProducerRequest>
                    for RegisterProducerSvc<T> {
                        type Response = super::RegisterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RegisterProducerRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::register_producer(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = RegisterProducerSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/RegisterConsumer" => {
                    #[allow(non_camel_case_types)]
                    struct RegisterConsumerSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::RegisterConsumerRequest>
                    for RegisterConsumerSvc<T> {
                        type Response = super::RegisterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RegisterConsumerRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::register_consumer(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = RegisterConsumerSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/SetExecutionInputs" => {
                    #[allow(non_camel_case_types)]
                    struct SetExecutionInputsSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::ExecutionInputsRequest>
                    for SetExecutionInputsSvc<T> {
                        type Response = super::ExecutionInputsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ExecutionInputsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::set_execution_inputs(
                                        &inner,
                                        request,
                                    )
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = SetExecutionInputsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/FindByWorkflowExec" => {
                    #[allow(non_camel_case_types)]
                    struct FindByWorkflowExecSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::FindByWorkflowExecRequest>
                    for FindByWorkflowExecSvc<T> {
                        type Response = super::SearchArtifactsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::FindByWorkflowExecRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::find_by_workflow_exec(
                                        &inner,
                                        request,
                                    )
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = FindByWorkflowExecSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/ListUsage" => {
                    #[allow(non_camel_case_types)]
                    struct ListUsageSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::ListUsageRequest>
                    for ListUsageSvc<T> {
                        type Response = super::ListUsageResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListUsageRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::list_usage(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ListUsageSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/GetTriggeringArtifacts" => {
                    #[allow(non_camel_case_types)]
                    struct GetTriggeringArtifactsSvc<T: ArtifactRegistry>(pub Arc<T>);
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<super::GetTriggeringArtifactsRequest>
                    for GetTriggeringArtifactsSvc<T> {
                        type Response = super::GetTriggeringArtifactsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetTriggeringArtifactsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::get_triggering_artifacts(
                                        &inner,
                                        request,
                                    )
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetTriggeringArtifactsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/flyteidl.artifact.ArtifactRegistry/GetTriggeredExecutionsByArtifact" => {
                    #[allow(non_camel_case_types)]
                    struct GetTriggeredExecutionsByArtifactSvc<T: ArtifactRegistry>(
                        pub Arc<T>,
                    );
                    impl<
                        T: ArtifactRegistry,
                    > tonic::server::UnaryService<
                        super::GetTriggeredExecutionsByArtifactRequest,
                    > for GetTriggeredExecutionsByArtifactSvc<T> {
                        type Response = super::GetTriggeredExecutionsByArtifactResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::GetTriggeredExecutionsByArtifactRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ArtifactRegistry>::get_triggered_executions_by_artifact(
                                        &inner,
                                        request,
                                    )
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetTriggeredExecutionsByArtifactSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                _ => {
                    Box::pin(async move {
                        Ok(
                            http::Response::builder()
                                .status(200)
                                .header("grpc-status", "12")
                                .header("content-type", "application/grpc")
                                .body(empty_body())
                                .unwrap(),
                        )
                    })
                }
            }
        }
    }
    impl<T: ArtifactRegistry> Clone for ArtifactRegistryServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self {
                inner,
                accept_compression_encodings: self.accept_compression_encodings,
                send_compression_encodings: self.send_compression_encodings,
                max_decoding_message_size: self.max_decoding_message_size,
                max_encoding_message_size: self.max_encoding_message_size,
            }
        }
    }
    impl<T: ArtifactRegistry> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: ArtifactRegistry> tonic::server::NamedService for ArtifactRegistryServer<T> {
        const NAME: &'static str = "flyteidl.artifact.ArtifactRegistry";
    }
}
