// @generated
/// Generated client implementations.
pub mod admin_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** The following defines an RPC service that is also served over HTTP via grpc-gateway.
 Standard response codes for both are defined here: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go
*/
    #[derive(Debug, Clone)]
    pub struct AdminServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl AdminServiceClient<tonic::transport::Channel> {
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
    impl<T> AdminServiceClient<T>
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
        ) -> AdminServiceClient<InterceptedService<T, F>>
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
            AdminServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** Create and upload a :ref:`ref_flyteidl.admin.Task` definition
*/
        pub async fn create_task(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::TaskCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskCreateResponse>,
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
                "/flyteidl.service.AdminService/CreateTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.AdminService", "CreateTask"));
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a :ref:`ref_flyteidl.admin.Task` definition.
*/
        pub async fn get_task(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ObjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Task>,
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
                "/flyteidl.service.AdminService/GetTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.AdminService", "GetTask"));
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of task objects.
*/
        pub async fn list_task_ids(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NamedEntityIdentifierListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityIdentifierList>,
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
                "/flyteidl.service.AdminService/ListTaskIds",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.AdminService", "ListTaskIds"));
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.Task` definitions.
*/
        pub async fn list_tasks(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ResourceListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskList>,
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
                "/flyteidl.service.AdminService/ListTasks",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.AdminService", "ListTasks"));
            self.inner.unary(req, path, codec).await
        }
        /** Create and upload a :ref:`ref_flyteidl.admin.Workflow` definition
*/
        pub async fn create_workflow(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::WorkflowCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowCreateResponse>,
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
                "/flyteidl.service.AdminService/CreateWorkflow",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "CreateWorkflow"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a :ref:`ref_flyteidl.admin.Workflow` definition.
*/
        pub async fn get_workflow(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ObjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Workflow>,
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
                "/flyteidl.service.AdminService/GetWorkflow",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.AdminService", "GetWorkflow"));
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of workflow objects.
*/
        pub async fn list_workflow_ids(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NamedEntityIdentifierListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityIdentifierList>,
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
                "/flyteidl.service.AdminService/ListWorkflowIds",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "ListWorkflowIds"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.Workflow` definitions.
*/
        pub async fn list_workflows(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ResourceListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowList>,
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
                "/flyteidl.service.AdminService/ListWorkflows",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "ListWorkflows"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Create and upload a :ref:`ref_flyteidl.admin.LaunchPlan` definition
*/
        pub async fn create_launch_plan(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::LaunchPlanCreateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlanCreateResponse>,
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
                "/flyteidl.service.AdminService/CreateLaunchPlan",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "CreateLaunchPlan"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a :ref:`ref_flyteidl.admin.LaunchPlan` definition.
*/
        pub async fn get_launch_plan(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ObjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlan>,
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
                "/flyteidl.service.AdminService/GetLaunchPlan",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "GetLaunchPlan"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch the active version of a :ref:`ref_flyteidl.admin.LaunchPlan`.
*/
        pub async fn get_active_launch_plan(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ActiveLaunchPlanRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlan>,
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
                "/flyteidl.service.AdminService/GetActiveLaunchPlan",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetActiveLaunchPlan",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** List active versions of :ref:`ref_flyteidl.admin.LaunchPlan`.
*/
        pub async fn list_active_launch_plans(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ActiveLaunchPlanListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlanList>,
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
                "/flyteidl.service.AdminService/ListActiveLaunchPlans",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "ListActiveLaunchPlans",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of launch plan objects.
*/
        pub async fn list_launch_plan_ids(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NamedEntityIdentifierListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityIdentifierList>,
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
                "/flyteidl.service.AdminService/ListLaunchPlanIds",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "ListLaunchPlanIds"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.LaunchPlan` definitions.
*/
        pub async fn list_launch_plans(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ResourceListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlanList>,
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
                "/flyteidl.service.AdminService/ListLaunchPlans",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "ListLaunchPlans"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Updates the status of a registered :ref:`ref_flyteidl.admin.LaunchPlan`.
*/
        pub async fn update_launch_plan(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::LaunchPlanUpdateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlanUpdateResponse>,
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
                "/flyteidl.service.AdminService/UpdateLaunchPlan",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "UpdateLaunchPlan"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Triggers the creation of a :ref:`ref_flyteidl.admin.Execution`
*/
        pub async fn create_execution(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ExecutionCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionCreateResponse>,
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
                "/flyteidl.service.AdminService/CreateExecution",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "CreateExecution"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Triggers the creation of an identical :ref:`ref_flyteidl.admin.Execution`
*/
        pub async fn relaunch_execution(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ExecutionRelaunchRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionCreateResponse>,
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
                "/flyteidl.service.AdminService/RelaunchExecution",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "RelaunchExecution"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Recreates a previously-run workflow execution that will only start executing from the last known failure point.
 In Recover mode, users cannot change any input parameters or update the version of the execution.
 This is extremely useful to recover from system errors and byzantine faults like - Loss of K8s cluster, bugs in platform or instability, machine failures,
 downstream system failures (downstream services), or simply to recover executions that failed because of retry exhaustion and should complete if tried again.
 See :ref:`ref_flyteidl.admin.ExecutionRecoverRequest` for more details.
*/
        pub async fn recover_execution(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ExecutionRecoverRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionCreateResponse>,
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
                "/flyteidl.service.AdminService/RecoverExecution",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "RecoverExecution"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches a :ref:`ref_flyteidl.admin.Execution`.
*/
        pub async fn get_execution(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::WorkflowExecutionGetRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Execution>,
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
                "/flyteidl.service.AdminService/GetExecution",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "GetExecution"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Update execution belonging to project domain :ref:`ref_flyteidl.admin.Execution`.
*/
        pub async fn update_execution(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ExecutionUpdateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionUpdateResponse>,
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
                "/flyteidl.service.AdminService/UpdateExecution",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "UpdateExecution"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches input and output data for a :ref:`ref_flyteidl.admin.Execution`.
*/
        pub async fn get_execution_data(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::WorkflowExecutionGetDataRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowExecutionGetDataResponse>,
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
                "/flyteidl.service.AdminService/GetExecutionData",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "GetExecutionData"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.Execution`.
*/
        pub async fn list_executions(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ResourceListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionList>,
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
                "/flyteidl.service.AdminService/ListExecutions",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "ListExecutions"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Terminates an in-progress :ref:`ref_flyteidl.admin.Execution`.
*/
        pub async fn terminate_execution(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ExecutionTerminateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionTerminateResponse>,
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
                "/flyteidl.service.AdminService/TerminateExecution",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "TerminateExecution",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches a :ref:`ref_flyteidl.admin.NodeExecution`.
*/
        pub async fn get_node_execution(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NodeExecutionGetRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecution>,
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
                "/flyteidl.service.AdminService/GetNodeExecution",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "GetNodeExecution"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches a :ref:`ref_flyteidl.admin.DynamicNodeWorkflowResponse`.
*/
        pub async fn get_dynamic_node_workflow(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::GetDynamicNodeWorkflowRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::DynamicNodeWorkflowResponse>,
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
                "/flyteidl.service.AdminService/GetDynamicNodeWorkflow",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetDynamicNodeWorkflow",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.NodeExecution`.
*/
        pub async fn list_node_executions(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NodeExecutionListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecutionList>,
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
                "/flyteidl.service.AdminService/ListNodeExecutions",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "ListNodeExecutions",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.NodeExecution` launched by the reference :ref:`ref_flyteidl.admin.TaskExecution`.
*/
        pub async fn list_node_executions_for_task(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NodeExecutionForTaskListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecutionList>,
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
                "/flyteidl.service.AdminService/ListNodeExecutionsForTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "ListNodeExecutionsForTask",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches input and output data for a :ref:`ref_flyteidl.admin.NodeExecution`.
*/
        pub async fn get_node_execution_data(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NodeExecutionGetDataRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecutionGetDataResponse>,
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
                "/flyteidl.service.AdminService/GetNodeExecutionData",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetNodeExecutionData",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Registers a :ref:`ref_flyteidl.admin.Project` with the Flyte deployment.
*/
        pub async fn register_project(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ProjectRegisterRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectRegisterResponse>,
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
                "/flyteidl.service.AdminService/RegisterProject",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "RegisterProject"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Updates an existing :ref:`ref_flyteidl.admin.Project`
 flyteidl.admin.Project should be passed but the domains property should be empty;
 it will be ignored in the handler as domains cannot be updated via this API.
*/
        pub async fn update_project(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::Project>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectUpdateResponse>,
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
                "/flyteidl.service.AdminService/UpdateProject",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "UpdateProject"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches a :ref:`ref_flyteidl.admin.Project`
*/
        pub async fn get_project(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ProjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Project>,
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
                "/flyteidl.service.AdminService/GetProject",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.AdminService", "GetProject"));
            self.inner.unary(req, path, codec).await
        }
        /** Fetches a list of :ref:`ref_flyteidl.admin.Project`
*/
        pub async fn list_projects(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ProjectListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Projects>,
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
                "/flyteidl.service.AdminService/ListProjects",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "ListProjects"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_domains(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::GetDomainRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetDomainsResponse>,
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
                "/flyteidl.service.AdminService/GetDomains",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.AdminService", "GetDomains"));
            self.inner.unary(req, path, codec).await
        }
        /** Indicates a :ref:`ref_flyteidl.event.WorkflowExecutionEvent` has occurred.
*/
        pub async fn create_workflow_event(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::WorkflowExecutionEventRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowExecutionEventResponse>,
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
                "/flyteidl.service.AdminService/CreateWorkflowEvent",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "CreateWorkflowEvent",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Indicates a :ref:`ref_flyteidl.event.NodeExecutionEvent` has occurred.
*/
        pub async fn create_node_event(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NodeExecutionEventRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecutionEventResponse>,
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
                "/flyteidl.service.AdminService/CreateNodeEvent",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "CreateNodeEvent"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Indicates a :ref:`ref_flyteidl.event.TaskExecutionEvent` has occurred.
*/
        pub async fn create_task_event(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::TaskExecutionEventRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskExecutionEventResponse>,
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
                "/flyteidl.service.AdminService/CreateTaskEvent",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "CreateTaskEvent"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches a :ref:`ref_flyteidl.admin.TaskExecution`.
*/
        pub async fn get_task_execution(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::TaskExecutionGetRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskExecution>,
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
                "/flyteidl.service.AdminService/GetTaskExecution",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "GetTaskExecution"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches a list of :ref:`ref_flyteidl.admin.TaskExecution`.
*/
        pub async fn list_task_executions(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::TaskExecutionListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskExecutionList>,
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
                "/flyteidl.service.AdminService/ListTaskExecutions",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "ListTaskExecutions",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches input and output data for a :ref:`ref_flyteidl.admin.TaskExecution`.
*/
        pub async fn get_task_execution_data(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::TaskExecutionGetDataRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskExecutionGetDataResponse>,
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
                "/flyteidl.service.AdminService/GetTaskExecutionData",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetTaskExecutionData",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        pub async fn update_project_domain_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ProjectDomainAttributesUpdateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectDomainAttributesUpdateResponse>,
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
                "/flyteidl.service.AdminService/UpdateProjectDomainAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "UpdateProjectDomainAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        pub async fn get_project_domain_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ProjectDomainAttributesGetRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectDomainAttributesGetResponse>,
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
                "/flyteidl.service.AdminService/GetProjectDomainAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetProjectDomainAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        pub async fn delete_project_domain_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ProjectDomainAttributesDeleteRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectDomainAttributesDeleteResponse>,
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
                "/flyteidl.service.AdminService/DeleteProjectDomainAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "DeleteProjectDomainAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` at the project level
*/
        pub async fn update_project_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ProjectAttributesUpdateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectAttributesUpdateResponse>,
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
                "/flyteidl.service.AdminService/UpdateProjectAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "UpdateProjectAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        pub async fn get_project_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ProjectAttributesGetRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectAttributesGetResponse>,
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
                "/flyteidl.service.AdminService/GetProjectAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetProjectAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        pub async fn delete_project_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ProjectAttributesDeleteRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectAttributesDeleteResponse>,
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
                "/flyteidl.service.AdminService/DeleteProjectAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "DeleteProjectAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow.
*/
        pub async fn update_workflow_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::WorkflowAttributesUpdateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowAttributesUpdateResponse>,
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
                "/flyteidl.service.AdminService/UpdateWorkflowAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "UpdateWorkflowAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow.
*/
        pub async fn get_workflow_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::WorkflowAttributesGetRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowAttributesGetResponse>,
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
                "/flyteidl.service.AdminService/GetWorkflowAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetWorkflowAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow.
*/
        pub async fn delete_workflow_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::WorkflowAttributesDeleteRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowAttributesDeleteResponse>,
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
                "/flyteidl.service.AdminService/DeleteWorkflowAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "DeleteWorkflowAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Lists custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a specific resource type.
*/
        pub async fn list_matchable_attributes(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::ListMatchableAttributesRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ListMatchableAttributesResponse>,
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
                "/flyteidl.service.AdminService/ListMatchableAttributes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "ListMatchableAttributes",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Returns a list of :ref:`ref_flyteidl.admin.NamedEntity` objects.
*/
        pub async fn list_named_entities(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::NamedEntityListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityList>,
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
                "/flyteidl.service.AdminService/ListNamedEntities",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "ListNamedEntities"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Returns a :ref:`ref_flyteidl.admin.NamedEntity` object.
*/
        pub async fn get_named_entity(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::NamedEntityGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntity>,
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
                "/flyteidl.service.AdminService/GetNamedEntity",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "GetNamedEntity"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Updates a :ref:`ref_flyteidl.admin.NamedEntity` object.
*/
        pub async fn update_named_entity(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::NamedEntityUpdateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityUpdateResponse>,
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
                "/flyteidl.service.AdminService/UpdateNamedEntity",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AdminService", "UpdateNamedEntity"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_version(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::GetVersionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetVersionResponse>,
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
                "/flyteidl.service.AdminService/GetVersion",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.AdminService", "GetVersion"));
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a :ref:`ref_flyteidl.admin.DescriptionEntity` object.
*/
        pub async fn get_description_entity(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ObjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::DescriptionEntity>,
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
                "/flyteidl.service.AdminService/GetDescriptionEntity",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetDescriptionEntity",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.DescriptionEntity` definitions.
*/
        pub async fn list_description_entities(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::DescriptionEntityListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::DescriptionEntityList>,
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
                "/flyteidl.service.AdminService/ListDescriptionEntities",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "ListDescriptionEntities",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetches runtime metrics for a :ref:`ref_flyteidl.admin.Execution`.
*/
        pub async fn get_execution_metrics(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::WorkflowExecutionGetMetricsRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowExecutionGetMetricsResponse>,
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
                "/flyteidl.service.AdminService/GetExecutionMetrics",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AdminService",
                        "GetExecutionMetrics",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod admin_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with AdminServiceServer.
    #[async_trait]
    pub trait AdminService: Send + Sync + 'static {
        /** Create and upload a :ref:`ref_flyteidl.admin.Task` definition
*/
        async fn create_task(
            &self,
            request: tonic::Request<super::super::admin::TaskCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskCreateResponse>,
            tonic::Status,
        >;
        /** Fetch a :ref:`ref_flyteidl.admin.Task` definition.
*/
        async fn get_task(
            &self,
            request: tonic::Request<super::super::admin::ObjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Task>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of task objects.
*/
        async fn list_task_ids(
            &self,
            request: tonic::Request<
                super::super::admin::NamedEntityIdentifierListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityIdentifierList>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.Task` definitions.
*/
        async fn list_tasks(
            &self,
            request: tonic::Request<super::super::admin::ResourceListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskList>,
            tonic::Status,
        >;
        /** Create and upload a :ref:`ref_flyteidl.admin.Workflow` definition
*/
        async fn create_workflow(
            &self,
            request: tonic::Request<super::super::admin::WorkflowCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowCreateResponse>,
            tonic::Status,
        >;
        /** Fetch a :ref:`ref_flyteidl.admin.Workflow` definition.
*/
        async fn get_workflow(
            &self,
            request: tonic::Request<super::super::admin::ObjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Workflow>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of workflow objects.
*/
        async fn list_workflow_ids(
            &self,
            request: tonic::Request<
                super::super::admin::NamedEntityIdentifierListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityIdentifierList>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.Workflow` definitions.
*/
        async fn list_workflows(
            &self,
            request: tonic::Request<super::super::admin::ResourceListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowList>,
            tonic::Status,
        >;
        /** Create and upload a :ref:`ref_flyteidl.admin.LaunchPlan` definition
*/
        async fn create_launch_plan(
            &self,
            request: tonic::Request<super::super::admin::LaunchPlanCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlanCreateResponse>,
            tonic::Status,
        >;
        /** Fetch a :ref:`ref_flyteidl.admin.LaunchPlan` definition.
*/
        async fn get_launch_plan(
            &self,
            request: tonic::Request<super::super::admin::ObjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlan>,
            tonic::Status,
        >;
        /** Fetch the active version of a :ref:`ref_flyteidl.admin.LaunchPlan`.
*/
        async fn get_active_launch_plan(
            &self,
            request: tonic::Request<super::super::admin::ActiveLaunchPlanRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlan>,
            tonic::Status,
        >;
        /** List active versions of :ref:`ref_flyteidl.admin.LaunchPlan`.
*/
        async fn list_active_launch_plans(
            &self,
            request: tonic::Request<super::super::admin::ActiveLaunchPlanListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlanList>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of launch plan objects.
*/
        async fn list_launch_plan_ids(
            &self,
            request: tonic::Request<
                super::super::admin::NamedEntityIdentifierListRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityIdentifierList>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.LaunchPlan` definitions.
*/
        async fn list_launch_plans(
            &self,
            request: tonic::Request<super::super::admin::ResourceListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlanList>,
            tonic::Status,
        >;
        /** Updates the status of a registered :ref:`ref_flyteidl.admin.LaunchPlan`.
*/
        async fn update_launch_plan(
            &self,
            request: tonic::Request<super::super::admin::LaunchPlanUpdateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::LaunchPlanUpdateResponse>,
            tonic::Status,
        >;
        /** Triggers the creation of a :ref:`ref_flyteidl.admin.Execution`
*/
        async fn create_execution(
            &self,
            request: tonic::Request<super::super::admin::ExecutionCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionCreateResponse>,
            tonic::Status,
        >;
        /** Triggers the creation of an identical :ref:`ref_flyteidl.admin.Execution`
*/
        async fn relaunch_execution(
            &self,
            request: tonic::Request<super::super::admin::ExecutionRelaunchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionCreateResponse>,
            tonic::Status,
        >;
        /** Recreates a previously-run workflow execution that will only start executing from the last known failure point.
 In Recover mode, users cannot change any input parameters or update the version of the execution.
 This is extremely useful to recover from system errors and byzantine faults like - Loss of K8s cluster, bugs in platform or instability, machine failures,
 downstream system failures (downstream services), or simply to recover executions that failed because of retry exhaustion and should complete if tried again.
 See :ref:`ref_flyteidl.admin.ExecutionRecoverRequest` for more details.
*/
        async fn recover_execution(
            &self,
            request: tonic::Request<super::super::admin::ExecutionRecoverRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionCreateResponse>,
            tonic::Status,
        >;
        /** Fetches a :ref:`ref_flyteidl.admin.Execution`.
*/
        async fn get_execution(
            &self,
            request: tonic::Request<super::super::admin::WorkflowExecutionGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Execution>,
            tonic::Status,
        >;
        /** Update execution belonging to project domain :ref:`ref_flyteidl.admin.Execution`.
*/
        async fn update_execution(
            &self,
            request: tonic::Request<super::super::admin::ExecutionUpdateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionUpdateResponse>,
            tonic::Status,
        >;
        /** Fetches input and output data for a :ref:`ref_flyteidl.admin.Execution`.
*/
        async fn get_execution_data(
            &self,
            request: tonic::Request<super::super::admin::WorkflowExecutionGetDataRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowExecutionGetDataResponse>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.Execution`.
*/
        async fn list_executions(
            &self,
            request: tonic::Request<super::super::admin::ResourceListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionList>,
            tonic::Status,
        >;
        /** Terminates an in-progress :ref:`ref_flyteidl.admin.Execution`.
*/
        async fn terminate_execution(
            &self,
            request: tonic::Request<super::super::admin::ExecutionTerminateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ExecutionTerminateResponse>,
            tonic::Status,
        >;
        /** Fetches a :ref:`ref_flyteidl.admin.NodeExecution`.
*/
        async fn get_node_execution(
            &self,
            request: tonic::Request<super::super::admin::NodeExecutionGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecution>,
            tonic::Status,
        >;
        /** Fetches a :ref:`ref_flyteidl.admin.DynamicNodeWorkflowResponse`.
*/
        async fn get_dynamic_node_workflow(
            &self,
            request: tonic::Request<super::super::admin::GetDynamicNodeWorkflowRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::DynamicNodeWorkflowResponse>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.NodeExecution`.
*/
        async fn list_node_executions(
            &self,
            request: tonic::Request<super::super::admin::NodeExecutionListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecutionList>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.NodeExecution` launched by the reference :ref:`ref_flyteidl.admin.TaskExecution`.
*/
        async fn list_node_executions_for_task(
            &self,
            request: tonic::Request<super::super::admin::NodeExecutionForTaskListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecutionList>,
            tonic::Status,
        >;
        /** Fetches input and output data for a :ref:`ref_flyteidl.admin.NodeExecution`.
*/
        async fn get_node_execution_data(
            &self,
            request: tonic::Request<super::super::admin::NodeExecutionGetDataRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecutionGetDataResponse>,
            tonic::Status,
        >;
        /** Registers a :ref:`ref_flyteidl.admin.Project` with the Flyte deployment.
*/
        async fn register_project(
            &self,
            request: tonic::Request<super::super::admin::ProjectRegisterRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectRegisterResponse>,
            tonic::Status,
        >;
        /** Updates an existing :ref:`ref_flyteidl.admin.Project`
 flyteidl.admin.Project should be passed but the domains property should be empty;
 it will be ignored in the handler as domains cannot be updated via this API.
*/
        async fn update_project(
            &self,
            request: tonic::Request<super::super::admin::Project>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectUpdateResponse>,
            tonic::Status,
        >;
        /** Fetches a :ref:`ref_flyteidl.admin.Project`
*/
        async fn get_project(
            &self,
            request: tonic::Request<super::super::admin::ProjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Project>,
            tonic::Status,
        >;
        /** Fetches a list of :ref:`ref_flyteidl.admin.Project`
*/
        async fn list_projects(
            &self,
            request: tonic::Request<super::super::admin::ProjectListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Projects>,
            tonic::Status,
        >;
        ///
        async fn get_domains(
            &self,
            request: tonic::Request<super::super::admin::GetDomainRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetDomainsResponse>,
            tonic::Status,
        >;
        /** Indicates a :ref:`ref_flyteidl.event.WorkflowExecutionEvent` has occurred.
*/
        async fn create_workflow_event(
            &self,
            request: tonic::Request<super::super::admin::WorkflowExecutionEventRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowExecutionEventResponse>,
            tonic::Status,
        >;
        /** Indicates a :ref:`ref_flyteidl.event.NodeExecutionEvent` has occurred.
*/
        async fn create_node_event(
            &self,
            request: tonic::Request<super::super::admin::NodeExecutionEventRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NodeExecutionEventResponse>,
            tonic::Status,
        >;
        /** Indicates a :ref:`ref_flyteidl.event.TaskExecutionEvent` has occurred.
*/
        async fn create_task_event(
            &self,
            request: tonic::Request<super::super::admin::TaskExecutionEventRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskExecutionEventResponse>,
            tonic::Status,
        >;
        /** Fetches a :ref:`ref_flyteidl.admin.TaskExecution`.
*/
        async fn get_task_execution(
            &self,
            request: tonic::Request<super::super::admin::TaskExecutionGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskExecution>,
            tonic::Status,
        >;
        /** Fetches a list of :ref:`ref_flyteidl.admin.TaskExecution`.
*/
        async fn list_task_executions(
            &self,
            request: tonic::Request<super::super::admin::TaskExecutionListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskExecutionList>,
            tonic::Status,
        >;
        /** Fetches input and output data for a :ref:`ref_flyteidl.admin.TaskExecution`.
*/
        async fn get_task_execution_data(
            &self,
            request: tonic::Request<super::super::admin::TaskExecutionGetDataRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::TaskExecutionGetDataResponse>,
            tonic::Status,
        >;
        /** Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        async fn update_project_domain_attributes(
            &self,
            request: tonic::Request<
                super::super::admin::ProjectDomainAttributesUpdateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectDomainAttributesUpdateResponse>,
            tonic::Status,
        >;
        /** Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        async fn get_project_domain_attributes(
            &self,
            request: tonic::Request<
                super::super::admin::ProjectDomainAttributesGetRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectDomainAttributesGetResponse>,
            tonic::Status,
        >;
        /** Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        async fn delete_project_domain_attributes(
            &self,
            request: tonic::Request<
                super::super::admin::ProjectDomainAttributesDeleteRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectDomainAttributesDeleteResponse>,
            tonic::Status,
        >;
        /** Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` at the project level
*/
        async fn update_project_attributes(
            &self,
            request: tonic::Request<super::super::admin::ProjectAttributesUpdateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectAttributesUpdateResponse>,
            tonic::Status,
        >;
        /** Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        async fn get_project_attributes(
            &self,
            request: tonic::Request<super::super::admin::ProjectAttributesGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectAttributesGetResponse>,
            tonic::Status,
        >;
        /** Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain.
*/
        async fn delete_project_attributes(
            &self,
            request: tonic::Request<super::super::admin::ProjectAttributesDeleteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ProjectAttributesDeleteResponse>,
            tonic::Status,
        >;
        /** Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow.
*/
        async fn update_workflow_attributes(
            &self,
            request: tonic::Request<super::super::admin::WorkflowAttributesUpdateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowAttributesUpdateResponse>,
            tonic::Status,
        >;
        /** Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow.
*/
        async fn get_workflow_attributes(
            &self,
            request: tonic::Request<super::super::admin::WorkflowAttributesGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowAttributesGetResponse>,
            tonic::Status,
        >;
        /** Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow.
*/
        async fn delete_workflow_attributes(
            &self,
            request: tonic::Request<super::super::admin::WorkflowAttributesDeleteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowAttributesDeleteResponse>,
            tonic::Status,
        >;
        /** Lists custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a specific resource type.
*/
        async fn list_matchable_attributes(
            &self,
            request: tonic::Request<super::super::admin::ListMatchableAttributesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ListMatchableAttributesResponse>,
            tonic::Status,
        >;
        /** Returns a list of :ref:`ref_flyteidl.admin.NamedEntity` objects.
*/
        async fn list_named_entities(
            &self,
            request: tonic::Request<super::super::admin::NamedEntityListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityList>,
            tonic::Status,
        >;
        /** Returns a :ref:`ref_flyteidl.admin.NamedEntity` object.
*/
        async fn get_named_entity(
            &self,
            request: tonic::Request<super::super::admin::NamedEntityGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntity>,
            tonic::Status,
        >;
        /** Updates a :ref:`ref_flyteidl.admin.NamedEntity` object.
*/
        async fn update_named_entity(
            &self,
            request: tonic::Request<super::super::admin::NamedEntityUpdateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::NamedEntityUpdateResponse>,
            tonic::Status,
        >;
        ///
        async fn get_version(
            &self,
            request: tonic::Request<super::super::admin::GetVersionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetVersionResponse>,
            tonic::Status,
        >;
        /** Fetch a :ref:`ref_flyteidl.admin.DescriptionEntity` object.
*/
        async fn get_description_entity(
            &self,
            request: tonic::Request<super::super::admin::ObjectGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::DescriptionEntity>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.DescriptionEntity` definitions.
*/
        async fn list_description_entities(
            &self,
            request: tonic::Request<super::super::admin::DescriptionEntityListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::DescriptionEntityList>,
            tonic::Status,
        >;
        /** Fetches runtime metrics for a :ref:`ref_flyteidl.admin.Execution`.
*/
        async fn get_execution_metrics(
            &self,
            request: tonic::Request<
                super::super::admin::WorkflowExecutionGetMetricsRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::WorkflowExecutionGetMetricsResponse>,
            tonic::Status,
        >;
    }
    /** The following defines an RPC service that is also served over HTTP via grpc-gateway.
 Standard response codes for both are defined here: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go
*/
    #[derive(Debug)]
    pub struct AdminServiceServer<T: AdminService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: AdminService> AdminServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for AdminServiceServer<T>
    where
        T: AdminService,
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
                "/flyteidl.service.AdminService/CreateTask" => {
                    #[allow(non_camel_case_types)]
                    struct CreateTaskSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::TaskCreateRequest>
                    for CreateTaskSvc<T> {
                        type Response = super::super::admin::TaskCreateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::TaskCreateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::create_task(&inner, request).await
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
                        let method = CreateTaskSvc(inner);
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
                "/flyteidl.service.AdminService/GetTask" => {
                    #[allow(non_camel_case_types)]
                    struct GetTaskSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::ObjectGetRequest>
                    for GetTaskSvc<T> {
                        type Response = super::super::admin::Task;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ObjectGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_task(&inner, request).await
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
                        let method = GetTaskSvc(inner);
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
                "/flyteidl.service.AdminService/ListTaskIds" => {
                    #[allow(non_camel_case_types)]
                    struct ListTaskIdsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NamedEntityIdentifierListRequest,
                    > for ListTaskIdsSvc<T> {
                        type Response = super::super::admin::NamedEntityIdentifierList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NamedEntityIdentifierListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_task_ids(&inner, request).await
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
                        let method = ListTaskIdsSvc(inner);
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
                "/flyteidl.service.AdminService/ListTasks" => {
                    #[allow(non_camel_case_types)]
                    struct ListTasksSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ResourceListRequest,
                    > for ListTasksSvc<T> {
                        type Response = super::super::admin::TaskList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ResourceListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_tasks(&inner, request).await
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
                        let method = ListTasksSvc(inner);
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
                "/flyteidl.service.AdminService/CreateWorkflow" => {
                    #[allow(non_camel_case_types)]
                    struct CreateWorkflowSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::WorkflowCreateRequest,
                    > for CreateWorkflowSvc<T> {
                        type Response = super::super::admin::WorkflowCreateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::WorkflowCreateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::create_workflow(&inner, request).await
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
                        let method = CreateWorkflowSvc(inner);
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
                "/flyteidl.service.AdminService/GetWorkflow" => {
                    #[allow(non_camel_case_types)]
                    struct GetWorkflowSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::ObjectGetRequest>
                    for GetWorkflowSvc<T> {
                        type Response = super::super::admin::Workflow;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ObjectGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_workflow(&inner, request).await
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
                        let method = GetWorkflowSvc(inner);
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
                "/flyteidl.service.AdminService/ListWorkflowIds" => {
                    #[allow(non_camel_case_types)]
                    struct ListWorkflowIdsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NamedEntityIdentifierListRequest,
                    > for ListWorkflowIdsSvc<T> {
                        type Response = super::super::admin::NamedEntityIdentifierList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NamedEntityIdentifierListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_workflow_ids(&inner, request)
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
                        let method = ListWorkflowIdsSvc(inner);
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
                "/flyteidl.service.AdminService/ListWorkflows" => {
                    #[allow(non_camel_case_types)]
                    struct ListWorkflowsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ResourceListRequest,
                    > for ListWorkflowsSvc<T> {
                        type Response = super::super::admin::WorkflowList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ResourceListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_workflows(&inner, request).await
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
                        let method = ListWorkflowsSvc(inner);
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
                "/flyteidl.service.AdminService/CreateLaunchPlan" => {
                    #[allow(non_camel_case_types)]
                    struct CreateLaunchPlanSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::LaunchPlanCreateRequest,
                    > for CreateLaunchPlanSvc<T> {
                        type Response = super::super::admin::LaunchPlanCreateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::LaunchPlanCreateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::create_launch_plan(&inner, request)
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
                        let method = CreateLaunchPlanSvc(inner);
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
                "/flyteidl.service.AdminService/GetLaunchPlan" => {
                    #[allow(non_camel_case_types)]
                    struct GetLaunchPlanSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::ObjectGetRequest>
                    for GetLaunchPlanSvc<T> {
                        type Response = super::super::admin::LaunchPlan;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ObjectGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_launch_plan(&inner, request).await
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
                        let method = GetLaunchPlanSvc(inner);
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
                "/flyteidl.service.AdminService/GetActiveLaunchPlan" => {
                    #[allow(non_camel_case_types)]
                    struct GetActiveLaunchPlanSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ActiveLaunchPlanRequest,
                    > for GetActiveLaunchPlanSvc<T> {
                        type Response = super::super::admin::LaunchPlan;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ActiveLaunchPlanRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_active_launch_plan(&inner, request)
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
                        let method = GetActiveLaunchPlanSvc(inner);
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
                "/flyteidl.service.AdminService/ListActiveLaunchPlans" => {
                    #[allow(non_camel_case_types)]
                    struct ListActiveLaunchPlansSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ActiveLaunchPlanListRequest,
                    > for ListActiveLaunchPlansSvc<T> {
                        type Response = super::super::admin::LaunchPlanList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ActiveLaunchPlanListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_active_launch_plans(
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
                        let method = ListActiveLaunchPlansSvc(inner);
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
                "/flyteidl.service.AdminService/ListLaunchPlanIds" => {
                    #[allow(non_camel_case_types)]
                    struct ListLaunchPlanIdsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NamedEntityIdentifierListRequest,
                    > for ListLaunchPlanIdsSvc<T> {
                        type Response = super::super::admin::NamedEntityIdentifierList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NamedEntityIdentifierListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_launch_plan_ids(&inner, request)
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
                        let method = ListLaunchPlanIdsSvc(inner);
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
                "/flyteidl.service.AdminService/ListLaunchPlans" => {
                    #[allow(non_camel_case_types)]
                    struct ListLaunchPlansSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ResourceListRequest,
                    > for ListLaunchPlansSvc<T> {
                        type Response = super::super::admin::LaunchPlanList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ResourceListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_launch_plans(&inner, request)
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
                        let method = ListLaunchPlansSvc(inner);
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
                "/flyteidl.service.AdminService/UpdateLaunchPlan" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateLaunchPlanSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::LaunchPlanUpdateRequest,
                    > for UpdateLaunchPlanSvc<T> {
                        type Response = super::super::admin::LaunchPlanUpdateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::LaunchPlanUpdateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::update_launch_plan(&inner, request)
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
                        let method = UpdateLaunchPlanSvc(inner);
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
                "/flyteidl.service.AdminService/CreateExecution" => {
                    #[allow(non_camel_case_types)]
                    struct CreateExecutionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ExecutionCreateRequest,
                    > for CreateExecutionSvc<T> {
                        type Response = super::super::admin::ExecutionCreateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ExecutionCreateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::create_execution(&inner, request).await
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
                        let method = CreateExecutionSvc(inner);
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
                "/flyteidl.service.AdminService/RelaunchExecution" => {
                    #[allow(non_camel_case_types)]
                    struct RelaunchExecutionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ExecutionRelaunchRequest,
                    > for RelaunchExecutionSvc<T> {
                        type Response = super::super::admin::ExecutionCreateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ExecutionRelaunchRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::relaunch_execution(&inner, request)
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
                        let method = RelaunchExecutionSvc(inner);
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
                "/flyteidl.service.AdminService/RecoverExecution" => {
                    #[allow(non_camel_case_types)]
                    struct RecoverExecutionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ExecutionRecoverRequest,
                    > for RecoverExecutionSvc<T> {
                        type Response = super::super::admin::ExecutionCreateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ExecutionRecoverRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::recover_execution(&inner, request)
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
                        let method = RecoverExecutionSvc(inner);
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
                "/flyteidl.service.AdminService/GetExecution" => {
                    #[allow(non_camel_case_types)]
                    struct GetExecutionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::WorkflowExecutionGetRequest,
                    > for GetExecutionSvc<T> {
                        type Response = super::super::admin::Execution;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::WorkflowExecutionGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_execution(&inner, request).await
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
                        let method = GetExecutionSvc(inner);
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
                "/flyteidl.service.AdminService/UpdateExecution" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateExecutionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ExecutionUpdateRequest,
                    > for UpdateExecutionSvc<T> {
                        type Response = super::super::admin::ExecutionUpdateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ExecutionUpdateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::update_execution(&inner, request).await
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
                        let method = UpdateExecutionSvc(inner);
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
                "/flyteidl.service.AdminService/GetExecutionData" => {
                    #[allow(non_camel_case_types)]
                    struct GetExecutionDataSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::WorkflowExecutionGetDataRequest,
                    > for GetExecutionDataSvc<T> {
                        type Response = super::super::admin::WorkflowExecutionGetDataResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::WorkflowExecutionGetDataRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_execution_data(&inner, request)
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
                        let method = GetExecutionDataSvc(inner);
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
                "/flyteidl.service.AdminService/ListExecutions" => {
                    #[allow(non_camel_case_types)]
                    struct ListExecutionsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ResourceListRequest,
                    > for ListExecutionsSvc<T> {
                        type Response = super::super::admin::ExecutionList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ResourceListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_executions(&inner, request).await
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
                        let method = ListExecutionsSvc(inner);
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
                "/flyteidl.service.AdminService/TerminateExecution" => {
                    #[allow(non_camel_case_types)]
                    struct TerminateExecutionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ExecutionTerminateRequest,
                    > for TerminateExecutionSvc<T> {
                        type Response = super::super::admin::ExecutionTerminateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ExecutionTerminateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::terminate_execution(&inner, request)
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
                        let method = TerminateExecutionSvc(inner);
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
                "/flyteidl.service.AdminService/GetNodeExecution" => {
                    #[allow(non_camel_case_types)]
                    struct GetNodeExecutionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NodeExecutionGetRequest,
                    > for GetNodeExecutionSvc<T> {
                        type Response = super::super::admin::NodeExecution;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NodeExecutionGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_node_execution(&inner, request)
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
                        let method = GetNodeExecutionSvc(inner);
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
                "/flyteidl.service.AdminService/GetDynamicNodeWorkflow" => {
                    #[allow(non_camel_case_types)]
                    struct GetDynamicNodeWorkflowSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::GetDynamicNodeWorkflowRequest,
                    > for GetDynamicNodeWorkflowSvc<T> {
                        type Response = super::super::admin::DynamicNodeWorkflowResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::GetDynamicNodeWorkflowRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_dynamic_node_workflow(
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
                        let method = GetDynamicNodeWorkflowSvc(inner);
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
                "/flyteidl.service.AdminService/ListNodeExecutions" => {
                    #[allow(non_camel_case_types)]
                    struct ListNodeExecutionsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NodeExecutionListRequest,
                    > for ListNodeExecutionsSvc<T> {
                        type Response = super::super::admin::NodeExecutionList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NodeExecutionListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_node_executions(&inner, request)
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
                        let method = ListNodeExecutionsSvc(inner);
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
                "/flyteidl.service.AdminService/ListNodeExecutionsForTask" => {
                    #[allow(non_camel_case_types)]
                    struct ListNodeExecutionsForTaskSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NodeExecutionForTaskListRequest,
                    > for ListNodeExecutionsForTaskSvc<T> {
                        type Response = super::super::admin::NodeExecutionList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NodeExecutionForTaskListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_node_executions_for_task(
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
                        let method = ListNodeExecutionsForTaskSvc(inner);
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
                "/flyteidl.service.AdminService/GetNodeExecutionData" => {
                    #[allow(non_camel_case_types)]
                    struct GetNodeExecutionDataSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NodeExecutionGetDataRequest,
                    > for GetNodeExecutionDataSvc<T> {
                        type Response = super::super::admin::NodeExecutionGetDataResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NodeExecutionGetDataRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_node_execution_data(
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
                        let method = GetNodeExecutionDataSvc(inner);
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
                "/flyteidl.service.AdminService/RegisterProject" => {
                    #[allow(non_camel_case_types)]
                    struct RegisterProjectSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ProjectRegisterRequest,
                    > for RegisterProjectSvc<T> {
                        type Response = super::super::admin::ProjectRegisterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectRegisterRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::register_project(&inner, request).await
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
                        let method = RegisterProjectSvc(inner);
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
                "/flyteidl.service.AdminService/UpdateProject" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateProjectSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::Project>
                    for UpdateProjectSvc<T> {
                        type Response = super::super::admin::ProjectUpdateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::super::admin::Project>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::update_project(&inner, request).await
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
                        let method = UpdateProjectSvc(inner);
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
                "/flyteidl.service.AdminService/GetProject" => {
                    #[allow(non_camel_case_types)]
                    struct GetProjectSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::ProjectGetRequest>
                    for GetProjectSvc<T> {
                        type Response = super::super::admin::Project;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_project(&inner, request).await
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
                        let method = GetProjectSvc(inner);
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
                "/flyteidl.service.AdminService/ListProjects" => {
                    #[allow(non_camel_case_types)]
                    struct ListProjectsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ProjectListRequest,
                    > for ListProjectsSvc<T> {
                        type Response = super::super::admin::Projects;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_projects(&inner, request).await
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
                        let method = ListProjectsSvc(inner);
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
                "/flyteidl.service.AdminService/GetDomains" => {
                    #[allow(non_camel_case_types)]
                    struct GetDomainsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::GetDomainRequest>
                    for GetDomainsSvc<T> {
                        type Response = super::super::admin::GetDomainsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::GetDomainRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_domains(&inner, request).await
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
                        let method = GetDomainsSvc(inner);
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
                "/flyteidl.service.AdminService/CreateWorkflowEvent" => {
                    #[allow(non_camel_case_types)]
                    struct CreateWorkflowEventSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::WorkflowExecutionEventRequest,
                    > for CreateWorkflowEventSvc<T> {
                        type Response = super::super::admin::WorkflowExecutionEventResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::WorkflowExecutionEventRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::create_workflow_event(&inner, request)
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
                        let method = CreateWorkflowEventSvc(inner);
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
                "/flyteidl.service.AdminService/CreateNodeEvent" => {
                    #[allow(non_camel_case_types)]
                    struct CreateNodeEventSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NodeExecutionEventRequest,
                    > for CreateNodeEventSvc<T> {
                        type Response = super::super::admin::NodeExecutionEventResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NodeExecutionEventRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::create_node_event(&inner, request)
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
                        let method = CreateNodeEventSvc(inner);
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
                "/flyteidl.service.AdminService/CreateTaskEvent" => {
                    #[allow(non_camel_case_types)]
                    struct CreateTaskEventSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::TaskExecutionEventRequest,
                    > for CreateTaskEventSvc<T> {
                        type Response = super::super::admin::TaskExecutionEventResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::TaskExecutionEventRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::create_task_event(&inner, request)
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
                        let method = CreateTaskEventSvc(inner);
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
                "/flyteidl.service.AdminService/GetTaskExecution" => {
                    #[allow(non_camel_case_types)]
                    struct GetTaskExecutionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::TaskExecutionGetRequest,
                    > for GetTaskExecutionSvc<T> {
                        type Response = super::super::admin::TaskExecution;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::TaskExecutionGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_task_execution(&inner, request)
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
                        let method = GetTaskExecutionSvc(inner);
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
                "/flyteidl.service.AdminService/ListTaskExecutions" => {
                    #[allow(non_camel_case_types)]
                    struct ListTaskExecutionsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::TaskExecutionListRequest,
                    > for ListTaskExecutionsSvc<T> {
                        type Response = super::super::admin::TaskExecutionList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::TaskExecutionListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_task_executions(&inner, request)
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
                        let method = ListTaskExecutionsSvc(inner);
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
                "/flyteidl.service.AdminService/GetTaskExecutionData" => {
                    #[allow(non_camel_case_types)]
                    struct GetTaskExecutionDataSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::TaskExecutionGetDataRequest,
                    > for GetTaskExecutionDataSvc<T> {
                        type Response = super::super::admin::TaskExecutionGetDataResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::TaskExecutionGetDataRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_task_execution_data(
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
                        let method = GetTaskExecutionDataSvc(inner);
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
                "/flyteidl.service.AdminService/UpdateProjectDomainAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateProjectDomainAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ProjectDomainAttributesUpdateRequest,
                    > for UpdateProjectDomainAttributesSvc<T> {
                        type Response = super::super::admin::ProjectDomainAttributesUpdateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectDomainAttributesUpdateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::update_project_domain_attributes(
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
                        let method = UpdateProjectDomainAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/GetProjectDomainAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct GetProjectDomainAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ProjectDomainAttributesGetRequest,
                    > for GetProjectDomainAttributesSvc<T> {
                        type Response = super::super::admin::ProjectDomainAttributesGetResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectDomainAttributesGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_project_domain_attributes(
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
                        let method = GetProjectDomainAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/DeleteProjectDomainAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteProjectDomainAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ProjectDomainAttributesDeleteRequest,
                    > for DeleteProjectDomainAttributesSvc<T> {
                        type Response = super::super::admin::ProjectDomainAttributesDeleteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectDomainAttributesDeleteRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::delete_project_domain_attributes(
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
                        let method = DeleteProjectDomainAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/UpdateProjectAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateProjectAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ProjectAttributesUpdateRequest,
                    > for UpdateProjectAttributesSvc<T> {
                        type Response = super::super::admin::ProjectAttributesUpdateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectAttributesUpdateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::update_project_attributes(
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
                        let method = UpdateProjectAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/GetProjectAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct GetProjectAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ProjectAttributesGetRequest,
                    > for GetProjectAttributesSvc<T> {
                        type Response = super::super::admin::ProjectAttributesGetResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectAttributesGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_project_attributes(&inner, request)
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
                        let method = GetProjectAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/DeleteProjectAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteProjectAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ProjectAttributesDeleteRequest,
                    > for DeleteProjectAttributesSvc<T> {
                        type Response = super::super::admin::ProjectAttributesDeleteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ProjectAttributesDeleteRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::delete_project_attributes(
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
                        let method = DeleteProjectAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/UpdateWorkflowAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateWorkflowAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::WorkflowAttributesUpdateRequest,
                    > for UpdateWorkflowAttributesSvc<T> {
                        type Response = super::super::admin::WorkflowAttributesUpdateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::WorkflowAttributesUpdateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::update_workflow_attributes(
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
                        let method = UpdateWorkflowAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/GetWorkflowAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct GetWorkflowAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::WorkflowAttributesGetRequest,
                    > for GetWorkflowAttributesSvc<T> {
                        type Response = super::super::admin::WorkflowAttributesGetResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::WorkflowAttributesGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_workflow_attributes(
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
                        let method = GetWorkflowAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/DeleteWorkflowAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteWorkflowAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::WorkflowAttributesDeleteRequest,
                    > for DeleteWorkflowAttributesSvc<T> {
                        type Response = super::super::admin::WorkflowAttributesDeleteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::WorkflowAttributesDeleteRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::delete_workflow_attributes(
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
                        let method = DeleteWorkflowAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/ListMatchableAttributes" => {
                    #[allow(non_camel_case_types)]
                    struct ListMatchableAttributesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::ListMatchableAttributesRequest,
                    > for ListMatchableAttributesSvc<T> {
                        type Response = super::super::admin::ListMatchableAttributesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ListMatchableAttributesRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_matchable_attributes(
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
                        let method = ListMatchableAttributesSvc(inner);
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
                "/flyteidl.service.AdminService/ListNamedEntities" => {
                    #[allow(non_camel_case_types)]
                    struct ListNamedEntitiesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NamedEntityListRequest,
                    > for ListNamedEntitiesSvc<T> {
                        type Response = super::super::admin::NamedEntityList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NamedEntityListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_named_entities(&inner, request)
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
                        let method = ListNamedEntitiesSvc(inner);
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
                "/flyteidl.service.AdminService/GetNamedEntity" => {
                    #[allow(non_camel_case_types)]
                    struct GetNamedEntitySvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NamedEntityGetRequest,
                    > for GetNamedEntitySvc<T> {
                        type Response = super::super::admin::NamedEntity;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NamedEntityGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_named_entity(&inner, request).await
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
                        let method = GetNamedEntitySvc(inner);
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
                "/flyteidl.service.AdminService/UpdateNamedEntity" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateNamedEntitySvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::NamedEntityUpdateRequest,
                    > for UpdateNamedEntitySvc<T> {
                        type Response = super::super::admin::NamedEntityUpdateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::NamedEntityUpdateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::update_named_entity(&inner, request)
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
                        let method = UpdateNamedEntitySvc(inner);
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
                "/flyteidl.service.AdminService/GetVersion" => {
                    #[allow(non_camel_case_types)]
                    struct GetVersionSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::GetVersionRequest>
                    for GetVersionSvc<T> {
                        type Response = super::super::admin::GetVersionResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::GetVersionRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_version(&inner, request).await
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
                        let method = GetVersionSvc(inner);
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
                "/flyteidl.service.AdminService/GetDescriptionEntity" => {
                    #[allow(non_camel_case_types)]
                    struct GetDescriptionEntitySvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<super::super::admin::ObjectGetRequest>
                    for GetDescriptionEntitySvc<T> {
                        type Response = super::super::admin::DescriptionEntity;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ObjectGetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_description_entity(&inner, request)
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
                        let method = GetDescriptionEntitySvc(inner);
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
                "/flyteidl.service.AdminService/ListDescriptionEntities" => {
                    #[allow(non_camel_case_types)]
                    struct ListDescriptionEntitiesSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::DescriptionEntityListRequest,
                    > for ListDescriptionEntitiesSvc<T> {
                        type Response = super::super::admin::DescriptionEntityList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::DescriptionEntityListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::list_description_entities(
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
                        let method = ListDescriptionEntitiesSvc(inner);
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
                "/flyteidl.service.AdminService/GetExecutionMetrics" => {
                    #[allow(non_camel_case_types)]
                    struct GetExecutionMetricsSvc<T: AdminService>(pub Arc<T>);
                    impl<
                        T: AdminService,
                    > tonic::server::UnaryService<
                        super::super::admin::WorkflowExecutionGetMetricsRequest,
                    > for GetExecutionMetricsSvc<T> {
                        type Response = super::super::admin::WorkflowExecutionGetMetricsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::WorkflowExecutionGetMetricsRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AdminService>::get_execution_metrics(&inner, request)
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
                        let method = GetExecutionMetricsSvc(inner);
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
    impl<T: AdminService> Clone for AdminServiceServer<T> {
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
    impl<T: AdminService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: AdminService> tonic::server::NamedService for AdminServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.AdminService";
    }
}
/// Generated client implementations.
pub mod sync_agent_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    #[derive(Debug, Clone)]
    pub struct SyncAgentServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl SyncAgentServiceClient<tonic::transport::Channel> {
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
    impl<T> SyncAgentServiceClient<T>
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
        ) -> SyncAgentServiceClient<InterceptedService<T, F>>
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
            SyncAgentServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** ExecuteTaskSync streams the create request and inputs to the agent service and streams the outputs back.
*/
        pub async fn execute_task_sync(
            &mut self,
            request: impl tonic::IntoStreamingRequest<
                Message = super::super::admin::ExecuteTaskSyncRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<
                tonic::codec::Streaming<super::super::admin::ExecuteTaskSyncResponse>,
            >,
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
                "/flyteidl.service.SyncAgentService/ExecuteTaskSync",
            );
            let mut req = request.into_streaming_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.SyncAgentService",
                        "ExecuteTaskSync",
                    ),
                );
            self.inner.streaming(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod sync_agent_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with SyncAgentServiceServer.
    #[async_trait]
    pub trait SyncAgentService: Send + Sync + 'static {
        /// Server streaming response type for the ExecuteTaskSync method.
        type ExecuteTaskSyncStream: tonic::codegen::tokio_stream::Stream<
                Item = std::result::Result<
                    super::super::admin::ExecuteTaskSyncResponse,
                    tonic::Status,
                >,
            >
            + Send
            + 'static;
        /** ExecuteTaskSync streams the create request and inputs to the agent service and streams the outputs back.
*/
        async fn execute_task_sync(
            &self,
            request: tonic::Request<
                tonic::Streaming<super::super::admin::ExecuteTaskSyncRequest>,
            >,
        ) -> std::result::Result<
            tonic::Response<Self::ExecuteTaskSyncStream>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct SyncAgentServiceServer<T: SyncAgentService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: SyncAgentService> SyncAgentServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for SyncAgentServiceServer<T>
    where
        T: SyncAgentService,
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
                "/flyteidl.service.SyncAgentService/ExecuteTaskSync" => {
                    #[allow(non_camel_case_types)]
                    struct ExecuteTaskSyncSvc<T: SyncAgentService>(pub Arc<T>);
                    impl<
                        T: SyncAgentService,
                    > tonic::server::StreamingService<
                        super::super::admin::ExecuteTaskSyncRequest,
                    > for ExecuteTaskSyncSvc<T> {
                        type Response = super::super::admin::ExecuteTaskSyncResponse;
                        type ResponseStream = T::ExecuteTaskSyncStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                tonic::Streaming<
                                    super::super::admin::ExecuteTaskSyncRequest,
                                >,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as SyncAgentService>::execute_task_sync(&inner, request)
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
                        let method = ExecuteTaskSyncSvc(inner);
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
                        let res = grpc.streaming(method, req).await;
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
    impl<T: SyncAgentService> Clone for SyncAgentServiceServer<T> {
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
    impl<T: SyncAgentService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: SyncAgentService> tonic::server::NamedService for SyncAgentServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.SyncAgentService";
    }
}
/// Generated client implementations.
pub mod async_agent_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** AsyncAgentService defines an RPC Service that allows propeller to send the request to the agent server asynchronously.
*/
    #[derive(Debug, Clone)]
    pub struct AsyncAgentServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl AsyncAgentServiceClient<tonic::transport::Channel> {
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
    impl<T> AsyncAgentServiceClient<T>
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
        ) -> AsyncAgentServiceClient<InterceptedService<T, F>>
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
            AsyncAgentServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** CreateTask sends a task create request to the agent service.
*/
        pub async fn create_task(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::CreateTaskRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::CreateTaskResponse>,
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
                "/flyteidl.service.AsyncAgentService/CreateTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AsyncAgentService", "CreateTask"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Get job status.
*/
        pub async fn get_task(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::GetTaskRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetTaskResponse>,
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
                "/flyteidl.service.AsyncAgentService/GetTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AsyncAgentService", "GetTask"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Delete the task resource.
*/
        pub async fn delete_task(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::DeleteTaskRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::DeleteTaskResponse>,
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
                "/flyteidl.service.AsyncAgentService/DeleteTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AsyncAgentService", "DeleteTask"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** GetTaskMetrics returns one or more task execution metrics, if available.

 Errors include
  * OutOfRange if metrics are not available for the specified task time range
  * various other errors
*/
        pub async fn get_task_metrics(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::GetTaskMetricsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetTaskMetricsResponse>,
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
                "/flyteidl.service.AsyncAgentService/GetTaskMetrics",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AsyncAgentService",
                        "GetTaskMetrics",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** GetTaskLogs returns task execution logs, if available.
*/
        pub async fn get_task_logs(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::GetTaskLogsRequest>,
        ) -> std::result::Result<
            tonic::Response<
                tonic::codec::Streaming<super::super::admin::GetTaskLogsResponse>,
            >,
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
                "/flyteidl.service.AsyncAgentService/GetTaskLogs",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AsyncAgentService", "GetTaskLogs"),
                );
            self.inner.server_streaming(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod async_agent_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with AsyncAgentServiceServer.
    #[async_trait]
    pub trait AsyncAgentService: Send + Sync + 'static {
        /** CreateTask sends a task create request to the agent service.
*/
        async fn create_task(
            &self,
            request: tonic::Request<super::super::admin::CreateTaskRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::CreateTaskResponse>,
            tonic::Status,
        >;
        /** Get job status.
*/
        async fn get_task(
            &self,
            request: tonic::Request<super::super::admin::GetTaskRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetTaskResponse>,
            tonic::Status,
        >;
        /** Delete the task resource.
*/
        async fn delete_task(
            &self,
            request: tonic::Request<super::super::admin::DeleteTaskRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::DeleteTaskResponse>,
            tonic::Status,
        >;
        /** GetTaskMetrics returns one or more task execution metrics, if available.

 Errors include
  * OutOfRange if metrics are not available for the specified task time range
  * various other errors
*/
        async fn get_task_metrics(
            &self,
            request: tonic::Request<super::super::admin::GetTaskMetricsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetTaskMetricsResponse>,
            tonic::Status,
        >;
        /// Server streaming response type for the GetTaskLogs method.
        type GetTaskLogsStream: tonic::codegen::tokio_stream::Stream<
                Item = std::result::Result<
                    super::super::admin::GetTaskLogsResponse,
                    tonic::Status,
                >,
            >
            + Send
            + 'static;
        /** GetTaskLogs returns task execution logs, if available.
*/
        async fn get_task_logs(
            &self,
            request: tonic::Request<super::super::admin::GetTaskLogsRequest>,
        ) -> std::result::Result<
            tonic::Response<Self::GetTaskLogsStream>,
            tonic::Status,
        >;
    }
    /** AsyncAgentService defines an RPC Service that allows propeller to send the request to the agent server asynchronously.
*/
    #[derive(Debug)]
    pub struct AsyncAgentServiceServer<T: AsyncAgentService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: AsyncAgentService> AsyncAgentServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for AsyncAgentServiceServer<T>
    where
        T: AsyncAgentService,
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
                "/flyteidl.service.AsyncAgentService/CreateTask" => {
                    #[allow(non_camel_case_types)]
                    struct CreateTaskSvc<T: AsyncAgentService>(pub Arc<T>);
                    impl<
                        T: AsyncAgentService,
                    > tonic::server::UnaryService<super::super::admin::CreateTaskRequest>
                    for CreateTaskSvc<T> {
                        type Response = super::super::admin::CreateTaskResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::CreateTaskRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AsyncAgentService>::create_task(&inner, request).await
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
                        let method = CreateTaskSvc(inner);
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
                "/flyteidl.service.AsyncAgentService/GetTask" => {
                    #[allow(non_camel_case_types)]
                    struct GetTaskSvc<T: AsyncAgentService>(pub Arc<T>);
                    impl<
                        T: AsyncAgentService,
                    > tonic::server::UnaryService<super::super::admin::GetTaskRequest>
                    for GetTaskSvc<T> {
                        type Response = super::super::admin::GetTaskResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::super::admin::GetTaskRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AsyncAgentService>::get_task(&inner, request).await
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
                        let method = GetTaskSvc(inner);
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
                "/flyteidl.service.AsyncAgentService/DeleteTask" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteTaskSvc<T: AsyncAgentService>(pub Arc<T>);
                    impl<
                        T: AsyncAgentService,
                    > tonic::server::UnaryService<super::super::admin::DeleteTaskRequest>
                    for DeleteTaskSvc<T> {
                        type Response = super::super::admin::DeleteTaskResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::DeleteTaskRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AsyncAgentService>::delete_task(&inner, request).await
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
                        let method = DeleteTaskSvc(inner);
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
                "/flyteidl.service.AsyncAgentService/GetTaskMetrics" => {
                    #[allow(non_camel_case_types)]
                    struct GetTaskMetricsSvc<T: AsyncAgentService>(pub Arc<T>);
                    impl<
                        T: AsyncAgentService,
                    > tonic::server::UnaryService<
                        super::super::admin::GetTaskMetricsRequest,
                    > for GetTaskMetricsSvc<T> {
                        type Response = super::super::admin::GetTaskMetricsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::GetTaskMetricsRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AsyncAgentService>::get_task_metrics(&inner, request)
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
                        let method = GetTaskMetricsSvc(inner);
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
                "/flyteidl.service.AsyncAgentService/GetTaskLogs" => {
                    #[allow(non_camel_case_types)]
                    struct GetTaskLogsSvc<T: AsyncAgentService>(pub Arc<T>);
                    impl<
                        T: AsyncAgentService,
                    > tonic::server::ServerStreamingService<
                        super::super::admin::GetTaskLogsRequest,
                    > for GetTaskLogsSvc<T> {
                        type Response = super::super::admin::GetTaskLogsResponse;
                        type ResponseStream = T::GetTaskLogsStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::GetTaskLogsRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AsyncAgentService>::get_task_logs(&inner, request)
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
                        let method = GetTaskLogsSvc(inner);
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
                        let res = grpc.server_streaming(method, req).await;
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
    impl<T: AsyncAgentService> Clone for AsyncAgentServiceServer<T> {
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
    impl<T: AsyncAgentService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: AsyncAgentService> tonic::server::NamedService
    for AsyncAgentServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.AsyncAgentService";
    }
}
/// Generated client implementations.
pub mod agent_metadata_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** AgentMetadataService defines an RPC service that is also served over HTTP via grpc-gateway.
 This service allows propeller or users to get the metadata of agents.
*/
    #[derive(Debug, Clone)]
    pub struct AgentMetadataServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl AgentMetadataServiceClient<tonic::transport::Channel> {
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
    impl<T> AgentMetadataServiceClient<T>
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
        ) -> AgentMetadataServiceClient<InterceptedService<T, F>>
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
            AgentMetadataServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** Fetch a :ref:`ref_flyteidl.admin.Agent` definition.
*/
        pub async fn get_agent(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::GetAgentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetAgentResponse>,
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
                "/flyteidl.service.AgentMetadataService/GetAgent",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.AgentMetadataService", "GetAgent"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.Agent` definitions.
*/
        pub async fn list_agents(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::ListAgentsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ListAgentsResponse>,
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
                "/flyteidl.service.AgentMetadataService/ListAgents",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AgentMetadataService",
                        "ListAgents",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod agent_metadata_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with AgentMetadataServiceServer.
    #[async_trait]
    pub trait AgentMetadataService: Send + Sync + 'static {
        /** Fetch a :ref:`ref_flyteidl.admin.Agent` definition.
*/
        async fn get_agent(
            &self,
            request: tonic::Request<super::super::admin::GetAgentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::GetAgentResponse>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.Agent` definitions.
*/
        async fn list_agents(
            &self,
            request: tonic::Request<super::super::admin::ListAgentsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::ListAgentsResponse>,
            tonic::Status,
        >;
    }
    /** AgentMetadataService defines an RPC service that is also served over HTTP via grpc-gateway.
 This service allows propeller or users to get the metadata of agents.
*/
    #[derive(Debug)]
    pub struct AgentMetadataServiceServer<T: AgentMetadataService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: AgentMetadataService> AgentMetadataServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>>
    for AgentMetadataServiceServer<T>
    where
        T: AgentMetadataService,
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
                "/flyteidl.service.AgentMetadataService/GetAgent" => {
                    #[allow(non_camel_case_types)]
                    struct GetAgentSvc<T: AgentMetadataService>(pub Arc<T>);
                    impl<
                        T: AgentMetadataService,
                    > tonic::server::UnaryService<super::super::admin::GetAgentRequest>
                    for GetAgentSvc<T> {
                        type Response = super::super::admin::GetAgentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::super::admin::GetAgentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AgentMetadataService>::get_agent(&inner, request)
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
                        let method = GetAgentSvc(inner);
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
                "/flyteidl.service.AgentMetadataService/ListAgents" => {
                    #[allow(non_camel_case_types)]
                    struct ListAgentsSvc<T: AgentMetadataService>(pub Arc<T>);
                    impl<
                        T: AgentMetadataService,
                    > tonic::server::UnaryService<super::super::admin::ListAgentsRequest>
                    for ListAgentsSvc<T> {
                        type Response = super::super::admin::ListAgentsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::ListAgentsRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AgentMetadataService>::list_agents(&inner, request)
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
                        let method = ListAgentsSvc(inner);
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
    impl<T: AgentMetadataService> Clone for AgentMetadataServiceServer<T> {
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
    impl<T: AgentMetadataService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: AgentMetadataService> tonic::server::NamedService
    for AgentMetadataServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.AgentMetadataService";
    }
}
/// Generated client implementations.
pub mod auth_metadata_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** The following defines an RPC service that is also served over HTTP via grpc-gateway.
 Standard response codes for both are defined here: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go
 RPCs defined in this service must be anonymously accessible.
*/
    #[derive(Debug, Clone)]
    pub struct AuthMetadataServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl AuthMetadataServiceClient<tonic::transport::Channel> {
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
    impl<T> AuthMetadataServiceClient<T>
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
        ) -> AuthMetadataServiceClient<InterceptedService<T, F>>
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
            AuthMetadataServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** Anonymously accessible. Retrieves local or external oauth authorization server metadata.
*/
        pub async fn get_o_auth2_metadata(
            &mut self,
            request: impl tonic::IntoRequest<super::OAuth2MetadataRequest>,
        ) -> std::result::Result<
            tonic::Response<super::OAuth2MetadataResponse>,
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
                "/flyteidl.service.AuthMetadataService/GetOAuth2Metadata",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AuthMetadataService",
                        "GetOAuth2Metadata",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Anonymously accessible. Retrieves the client information clients should use when initiating OAuth2 authorization
 requests.
*/
        pub async fn get_public_client_config(
            &mut self,
            request: impl tonic::IntoRequest<super::PublicClientAuthConfigRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PublicClientAuthConfigResponse>,
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
                "/flyteidl.service.AuthMetadataService/GetPublicClientConfig",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.AuthMetadataService",
                        "GetPublicClientConfig",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod auth_metadata_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with AuthMetadataServiceServer.
    #[async_trait]
    pub trait AuthMetadataService: Send + Sync + 'static {
        /** Anonymously accessible. Retrieves local or external oauth authorization server metadata.
*/
        async fn get_o_auth2_metadata(
            &self,
            request: tonic::Request<super::OAuth2MetadataRequest>,
        ) -> std::result::Result<
            tonic::Response<super::OAuth2MetadataResponse>,
            tonic::Status,
        >;
        /** Anonymously accessible. Retrieves the client information clients should use when initiating OAuth2 authorization
 requests.
*/
        async fn get_public_client_config(
            &self,
            request: tonic::Request<super::PublicClientAuthConfigRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PublicClientAuthConfigResponse>,
            tonic::Status,
        >;
    }
    /** The following defines an RPC service that is also served over HTTP via grpc-gateway.
 Standard response codes for both are defined here: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go
 RPCs defined in this service must be anonymously accessible.
*/
    #[derive(Debug)]
    pub struct AuthMetadataServiceServer<T: AuthMetadataService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: AuthMetadataService> AuthMetadataServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for AuthMetadataServiceServer<T>
    where
        T: AuthMetadataService,
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
                "/flyteidl.service.AuthMetadataService/GetOAuth2Metadata" => {
                    #[allow(non_camel_case_types)]
                    struct GetOAuth2MetadataSvc<T: AuthMetadataService>(pub Arc<T>);
                    impl<
                        T: AuthMetadataService,
                    > tonic::server::UnaryService<super::OAuth2MetadataRequest>
                    for GetOAuth2MetadataSvc<T> {
                        type Response = super::OAuth2MetadataResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::OAuth2MetadataRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AuthMetadataService>::get_o_auth2_metadata(
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
                        let method = GetOAuth2MetadataSvc(inner);
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
                "/flyteidl.service.AuthMetadataService/GetPublicClientConfig" => {
                    #[allow(non_camel_case_types)]
                    struct GetPublicClientConfigSvc<T: AuthMetadataService>(pub Arc<T>);
                    impl<
                        T: AuthMetadataService,
                    > tonic::server::UnaryService<super::PublicClientAuthConfigRequest>
                    for GetPublicClientConfigSvc<T> {
                        type Response = super::PublicClientAuthConfigResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PublicClientAuthConfigRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AuthMetadataService>::get_public_client_config(
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
                        let method = GetPublicClientConfigSvc(inner);
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
    impl<T: AuthMetadataService> Clone for AuthMetadataServiceServer<T> {
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
    impl<T: AuthMetadataService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: AuthMetadataService> tonic::server::NamedService
    for AuthMetadataServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.AuthMetadataService";
    }
}
/// Generated client implementations.
pub mod data_proxy_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** DataProxyService defines an RPC Service that allows access to user-data in a controlled manner.
*/
    #[derive(Debug, Clone)]
    pub struct DataProxyServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl DataProxyServiceClient<tonic::transport::Channel> {
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
    impl<T> DataProxyServiceClient<T>
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
        ) -> DataProxyServiceClient<InterceptedService<T, F>>
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
            DataProxyServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** CreateUploadLocation creates a signed url to upload artifacts to for a given project/domain.
*/
        pub async fn create_upload_location(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateUploadLocationRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateUploadLocationResponse>,
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
                "/flyteidl.service.DataProxyService/CreateUploadLocation",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.DataProxyService",
                        "CreateUploadLocation",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** CreateDownloadLocation creates a signed url to download artifacts.
*/
        pub async fn create_download_location(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateDownloadLocationRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateDownloadLocationResponse>,
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
                "/flyteidl.service.DataProxyService/CreateDownloadLocation",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.DataProxyService",
                        "CreateDownloadLocation",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** CreateDownloadLocation creates a signed url to download artifacts.
*/
        pub async fn create_download_link(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateDownloadLinkRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateDownloadLinkResponse>,
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
                "/flyteidl.service.DataProxyService/CreateDownloadLink",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.DataProxyService",
                        "CreateDownloadLink",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_data(
            &mut self,
            request: impl tonic::IntoRequest<super::GetDataRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetDataResponse>,
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
                "/flyteidl.service.DataProxyService/GetData",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.DataProxyService", "GetData"));
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod data_proxy_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with DataProxyServiceServer.
    #[async_trait]
    pub trait DataProxyService: Send + Sync + 'static {
        /** CreateUploadLocation creates a signed url to upload artifacts to for a given project/domain.
*/
        async fn create_upload_location(
            &self,
            request: tonic::Request<super::CreateUploadLocationRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateUploadLocationResponse>,
            tonic::Status,
        >;
        /** CreateDownloadLocation creates a signed url to download artifacts.
*/
        async fn create_download_location(
            &self,
            request: tonic::Request<super::CreateDownloadLocationRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateDownloadLocationResponse>,
            tonic::Status,
        >;
        /** CreateDownloadLocation creates a signed url to download artifacts.
*/
        async fn create_download_link(
            &self,
            request: tonic::Request<super::CreateDownloadLinkRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateDownloadLinkResponse>,
            tonic::Status,
        >;
        ///
        async fn get_data(
            &self,
            request: tonic::Request<super::GetDataRequest>,
        ) -> std::result::Result<tonic::Response<super::GetDataResponse>, tonic::Status>;
    }
    /** DataProxyService defines an RPC Service that allows access to user-data in a controlled manner.
*/
    #[derive(Debug)]
    pub struct DataProxyServiceServer<T: DataProxyService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: DataProxyService> DataProxyServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for DataProxyServiceServer<T>
    where
        T: DataProxyService,
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
                "/flyteidl.service.DataProxyService/CreateUploadLocation" => {
                    #[allow(non_camel_case_types)]
                    struct CreateUploadLocationSvc<T: DataProxyService>(pub Arc<T>);
                    impl<
                        T: DataProxyService,
                    > tonic::server::UnaryService<super::CreateUploadLocationRequest>
                    for CreateUploadLocationSvc<T> {
                        type Response = super::CreateUploadLocationResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateUploadLocationRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as DataProxyService>::create_upload_location(
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
                        let method = CreateUploadLocationSvc(inner);
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
                "/flyteidl.service.DataProxyService/CreateDownloadLocation" => {
                    #[allow(non_camel_case_types)]
                    struct CreateDownloadLocationSvc<T: DataProxyService>(pub Arc<T>);
                    impl<
                        T: DataProxyService,
                    > tonic::server::UnaryService<super::CreateDownloadLocationRequest>
                    for CreateDownloadLocationSvc<T> {
                        type Response = super::CreateDownloadLocationResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateDownloadLocationRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as DataProxyService>::create_download_location(
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
                        let method = CreateDownloadLocationSvc(inner);
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
                "/flyteidl.service.DataProxyService/CreateDownloadLink" => {
                    #[allow(non_camel_case_types)]
                    struct CreateDownloadLinkSvc<T: DataProxyService>(pub Arc<T>);
                    impl<
                        T: DataProxyService,
                    > tonic::server::UnaryService<super::CreateDownloadLinkRequest>
                    for CreateDownloadLinkSvc<T> {
                        type Response = super::CreateDownloadLinkResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateDownloadLinkRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as DataProxyService>::create_download_link(
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
                        let method = CreateDownloadLinkSvc(inner);
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
                "/flyteidl.service.DataProxyService/GetData" => {
                    #[allow(non_camel_case_types)]
                    struct GetDataSvc<T: DataProxyService>(pub Arc<T>);
                    impl<
                        T: DataProxyService,
                    > tonic::server::UnaryService<super::GetDataRequest>
                    for GetDataSvc<T> {
                        type Response = super::GetDataResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetDataRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as DataProxyService>::get_data(&inner, request).await
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
                        let method = GetDataSvc(inner);
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
    impl<T: DataProxyService> Clone for DataProxyServiceServer<T> {
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
    impl<T: DataProxyService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: DataProxyService> tonic::server::NamedService for DataProxyServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.DataProxyService";
    }
}
/// Generated client implementations.
pub mod external_plugin_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    #[derive(Debug, Clone)]
    pub struct ExternalPluginServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl ExternalPluginServiceClient<tonic::transport::Channel> {
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
    impl<T> ExternalPluginServiceClient<T>
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
        ) -> ExternalPluginServiceClient<InterceptedService<T, F>>
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
            ExternalPluginServiceClient::new(InterceptedService::new(inner, interceptor))
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
        pub async fn create_task(
            &mut self,
            request: impl tonic::IntoRequest<super::TaskCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::TaskCreateResponse>,
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
                "/flyteidl.service.ExternalPluginService/CreateTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.ExternalPluginService",
                        "CreateTask",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn get_task(
            &mut self,
            request: impl tonic::IntoRequest<super::TaskGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::TaskGetResponse>,
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
                "/flyteidl.service.ExternalPluginService/GetTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.ExternalPluginService", "GetTask"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn delete_task(
            &mut self,
            request: impl tonic::IntoRequest<super::TaskDeleteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::TaskDeleteResponse>,
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
                "/flyteidl.service.ExternalPluginService/DeleteTask",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.ExternalPluginService",
                        "DeleteTask",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod external_plugin_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with ExternalPluginServiceServer.
    #[async_trait]
    pub trait ExternalPluginService: Send + Sync + 'static {
        async fn create_task(
            &self,
            request: tonic::Request<super::TaskCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::TaskCreateResponse>,
            tonic::Status,
        >;
        async fn get_task(
            &self,
            request: tonic::Request<super::TaskGetRequest>,
        ) -> std::result::Result<tonic::Response<super::TaskGetResponse>, tonic::Status>;
        async fn delete_task(
            &self,
            request: tonic::Request<super::TaskDeleteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::TaskDeleteResponse>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct ExternalPluginServiceServer<T: ExternalPluginService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: ExternalPluginService> ExternalPluginServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>>
    for ExternalPluginServiceServer<T>
    where
        T: ExternalPluginService,
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
                "/flyteidl.service.ExternalPluginService/CreateTask" => {
                    #[allow(non_camel_case_types)]
                    struct CreateTaskSvc<T: ExternalPluginService>(pub Arc<T>);
                    impl<
                        T: ExternalPluginService,
                    > tonic::server::UnaryService<super::TaskCreateRequest>
                    for CreateTaskSvc<T> {
                        type Response = super::TaskCreateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::TaskCreateRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ExternalPluginService>::create_task(&inner, request)
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
                        let method = CreateTaskSvc(inner);
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
                "/flyteidl.service.ExternalPluginService/GetTask" => {
                    #[allow(non_camel_case_types)]
                    struct GetTaskSvc<T: ExternalPluginService>(pub Arc<T>);
                    impl<
                        T: ExternalPluginService,
                    > tonic::server::UnaryService<super::TaskGetRequest>
                    for GetTaskSvc<T> {
                        type Response = super::TaskGetResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::TaskGetRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ExternalPluginService>::get_task(&inner, request)
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
                        let method = GetTaskSvc(inner);
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
                "/flyteidl.service.ExternalPluginService/DeleteTask" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteTaskSvc<T: ExternalPluginService>(pub Arc<T>);
                    impl<
                        T: ExternalPluginService,
                    > tonic::server::UnaryService<super::TaskDeleteRequest>
                    for DeleteTaskSvc<T> {
                        type Response = super::TaskDeleteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::TaskDeleteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ExternalPluginService>::delete_task(&inner, request)
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
                        let method = DeleteTaskSvc(inner);
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
    impl<T: ExternalPluginService> Clone for ExternalPluginServiceServer<T> {
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
    impl<T: ExternalPluginService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: ExternalPluginService> tonic::server::NamedService
    for ExternalPluginServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.ExternalPluginService";
    }
}
/// Generated client implementations.
pub mod identity_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** IdentityService defines an RPC Service that interacts with user/app identities.
*/
    #[derive(Debug, Clone)]
    pub struct IdentityServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl IdentityServiceClient<tonic::transport::Channel> {
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
    impl<T> IdentityServiceClient<T>
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
        ) -> IdentityServiceClient<InterceptedService<T, F>>
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
            IdentityServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** Retrieves user information about the currently logged in user.
*/
        pub async fn user_info(
            &mut self,
            request: impl tonic::IntoRequest<super::UserInfoRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UserInfoResponse>,
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
                "/flyteidl.service.IdentityService/UserInfo",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.IdentityService", "UserInfo"));
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod identity_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with IdentityServiceServer.
    #[async_trait]
    pub trait IdentityService: Send + Sync + 'static {
        /** Retrieves user information about the currently logged in user.
*/
        async fn user_info(
            &self,
            request: tonic::Request<super::UserInfoRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UserInfoResponse>,
            tonic::Status,
        >;
    }
    /** IdentityService defines an RPC Service that interacts with user/app identities.
*/
    #[derive(Debug)]
    pub struct IdentityServiceServer<T: IdentityService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: IdentityService> IdentityServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for IdentityServiceServer<T>
    where
        T: IdentityService,
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
                "/flyteidl.service.IdentityService/UserInfo" => {
                    #[allow(non_camel_case_types)]
                    struct UserInfoSvc<T: IdentityService>(pub Arc<T>);
                    impl<
                        T: IdentityService,
                    > tonic::server::UnaryService<super::UserInfoRequest>
                    for UserInfoSvc<T> {
                        type Response = super::UserInfoResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::UserInfoRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as IdentityService>::user_info(&inner, request).await
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
                        let method = UserInfoSvc(inner);
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
    impl<T: IdentityService> Clone for IdentityServiceServer<T> {
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
    impl<T: IdentityService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: IdentityService> tonic::server::NamedService for IdentityServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.IdentityService";
    }
}
/// Generated client implementations.
pub mod signal_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** SignalService defines an RPC Service that may create, update, and retrieve signal(s).
*/
    #[derive(Debug, Clone)]
    pub struct SignalServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl SignalServiceClient<tonic::transport::Channel> {
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
    impl<T> SignalServiceClient<T>
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
        ) -> SignalServiceClient<InterceptedService<T, F>>
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
            SignalServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** Fetches or creates a :ref:`ref_flyteidl.admin.Signal`.
*/
        pub async fn get_or_create_signal(
            &mut self,
            request: impl tonic::IntoRequest<
                super::super::admin::SignalGetOrCreateRequest,
            >,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Signal>,
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
                "/flyteidl.service.SignalService/GetOrCreateSignal",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "flyteidl.service.SignalService",
                        "GetOrCreateSignal",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Fetch a list of :ref:`ref_flyteidl.admin.Signal` definitions.
*/
        pub async fn list_signals(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::SignalListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::SignalList>,
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
                "/flyteidl.service.SignalService/ListSignals",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("flyteidl.service.SignalService", "ListSignals"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Sets the value on a :ref:`ref_flyteidl.admin.Signal` definition
*/
        pub async fn set_signal(
            &mut self,
            request: impl tonic::IntoRequest<super::super::admin::SignalSetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::SignalSetResponse>,
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
                "/flyteidl.service.SignalService/SetSignal",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("flyteidl.service.SignalService", "SetSignal"));
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod signal_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with SignalServiceServer.
    #[async_trait]
    pub trait SignalService: Send + Sync + 'static {
        /** Fetches or creates a :ref:`ref_flyteidl.admin.Signal`.
*/
        async fn get_or_create_signal(
            &self,
            request: tonic::Request<super::super::admin::SignalGetOrCreateRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::Signal>,
            tonic::Status,
        >;
        /** Fetch a list of :ref:`ref_flyteidl.admin.Signal` definitions.
*/
        async fn list_signals(
            &self,
            request: tonic::Request<super::super::admin::SignalListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::SignalList>,
            tonic::Status,
        >;
        /** Sets the value on a :ref:`ref_flyteidl.admin.Signal` definition
*/
        async fn set_signal(
            &self,
            request: tonic::Request<super::super::admin::SignalSetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::super::admin::SignalSetResponse>,
            tonic::Status,
        >;
    }
    /** SignalService defines an RPC Service that may create, update, and retrieve signal(s).
*/
    #[derive(Debug)]
    pub struct SignalServiceServer<T: SignalService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: SignalService> SignalServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for SignalServiceServer<T>
    where
        T: SignalService,
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
                "/flyteidl.service.SignalService/GetOrCreateSignal" => {
                    #[allow(non_camel_case_types)]
                    struct GetOrCreateSignalSvc<T: SignalService>(pub Arc<T>);
                    impl<
                        T: SignalService,
                    > tonic::server::UnaryService<
                        super::super::admin::SignalGetOrCreateRequest,
                    > for GetOrCreateSignalSvc<T> {
                        type Response = super::super::admin::Signal;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::SignalGetOrCreateRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as SignalService>::get_or_create_signal(&inner, request)
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
                        let method = GetOrCreateSignalSvc(inner);
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
                "/flyteidl.service.SignalService/ListSignals" => {
                    #[allow(non_camel_case_types)]
                    struct ListSignalsSvc<T: SignalService>(pub Arc<T>);
                    impl<
                        T: SignalService,
                    > tonic::server::UnaryService<super::super::admin::SignalListRequest>
                    for ListSignalsSvc<T> {
                        type Response = super::super::admin::SignalList;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::SignalListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as SignalService>::list_signals(&inner, request).await
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
                        let method = ListSignalsSvc(inner);
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
                "/flyteidl.service.SignalService/SetSignal" => {
                    #[allow(non_camel_case_types)]
                    struct SetSignalSvc<T: SignalService>(pub Arc<T>);
                    impl<
                        T: SignalService,
                    > tonic::server::UnaryService<super::super::admin::SignalSetRequest>
                    for SetSignalSvc<T> {
                        type Response = super::super::admin::SignalSetResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::super::admin::SignalSetRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as SignalService>::set_signal(&inner, request).await
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
                        let method = SetSignalSvc(inner);
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
    impl<T: SignalService> Clone for SignalServiceServer<T> {
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
    impl<T: SignalService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: SignalService> tonic::server::NamedService for SignalServiceServer<T> {
        const NAME: &'static str = "flyteidl.service.SignalService";
    }
}
