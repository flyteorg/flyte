######################
Protocol Documentation
######################




.. _ref_flyteidl/service/admin.proto:

flyteidl/service/admin.proto
==================================================================




..
   end messages


..
   end enums


..
   end HasExtensions



.. _ref_flyteidl.service.AdminService:

AdminService
------------------------------------------------------------------

The following defines an RPC service that is also served over HTTP via grpc-gateway.
Standard response codes for both are defined here: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go

.. csv-table:: AdminService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "CreateTask", ":ref:`ref_flyteidl.admin.TaskCreateRequest`", ":ref:`ref_flyteidl.admin.TaskCreateResponse`", "Create and upload a :ref:`ref_flyteidl.admin.Task` definition"
   "GetTask", ":ref:`ref_flyteidl.admin.ObjectGetRequest`", ":ref:`ref_flyteidl.admin.Task`", "Fetch a :ref:`ref_flyteidl.admin.Task` definition."
   "ListTaskIds", ":ref:`ref_flyteidl.admin.NamedEntityIdentifierListRequest`", ":ref:`ref_flyteidl.admin.NamedEntityIdentifierList`", "Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of task objects."
   "ListTasks", ":ref:`ref_flyteidl.admin.ResourceListRequest`", ":ref:`ref_flyteidl.admin.TaskList`", "Fetch a list of :ref:`ref_flyteidl.admin.Task` definitions."
   "CreateWorkflow", ":ref:`ref_flyteidl.admin.WorkflowCreateRequest`", ":ref:`ref_flyteidl.admin.WorkflowCreateResponse`", "Create and upload a :ref:`ref_flyteidl.admin.Workflow` definition"
   "GetWorkflow", ":ref:`ref_flyteidl.admin.ObjectGetRequest`", ":ref:`ref_flyteidl.admin.Workflow`", "Fetch a :ref:`ref_flyteidl.admin.Workflow` definition."
   "ListWorkflowIds", ":ref:`ref_flyteidl.admin.NamedEntityIdentifierListRequest`", ":ref:`ref_flyteidl.admin.NamedEntityIdentifierList`", "Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of workflow objects."
   "ListWorkflows", ":ref:`ref_flyteidl.admin.ResourceListRequest`", ":ref:`ref_flyteidl.admin.WorkflowList`", "Fetch a list of :ref:`ref_flyteidl.admin.Workflow` definitions."
   "CreateLaunchPlan", ":ref:`ref_flyteidl.admin.LaunchPlanCreateRequest`", ":ref:`ref_flyteidl.admin.LaunchPlanCreateResponse`", "Create and upload a :ref:`ref_flyteidl.admin.LaunchPlan` definition"
   "GetLaunchPlan", ":ref:`ref_flyteidl.admin.ObjectGetRequest`", ":ref:`ref_flyteidl.admin.LaunchPlan`", "Fetch a :ref:`ref_flyteidl.admin.LaunchPlan` definition."
   "GetActiveLaunchPlan", ":ref:`ref_flyteidl.admin.ActiveLaunchPlanRequest`", ":ref:`ref_flyteidl.admin.LaunchPlan`", "Fetch the active version of a :ref:`ref_flyteidl.admin.LaunchPlan`."
   "ListActiveLaunchPlans", ":ref:`ref_flyteidl.admin.ActiveLaunchPlanListRequest`", ":ref:`ref_flyteidl.admin.LaunchPlanList`", "List active versions of :ref:`ref_flyteidl.admin.LaunchPlan`."
   "ListLaunchPlanIds", ":ref:`ref_flyteidl.admin.NamedEntityIdentifierListRequest`", ":ref:`ref_flyteidl.admin.NamedEntityIdentifierList`", "Fetch a list of :ref:`ref_flyteidl.admin.NamedEntityIdentifier` of launch plan objects."
   "ListLaunchPlans", ":ref:`ref_flyteidl.admin.ResourceListRequest`", ":ref:`ref_flyteidl.admin.LaunchPlanList`", "Fetch a list of :ref:`ref_flyteidl.admin.LaunchPlan` definitions."
   "UpdateLaunchPlan", ":ref:`ref_flyteidl.admin.LaunchPlanUpdateRequest`", ":ref:`ref_flyteidl.admin.LaunchPlanUpdateResponse`", "Updates the status of a registered :ref:`ref_flyteidl.admin.LaunchPlan`."
   "CreateExecution", ":ref:`ref_flyteidl.admin.ExecutionCreateRequest`", ":ref:`ref_flyteidl.admin.ExecutionCreateResponse`", "Triggers the creation of a :ref:`ref_flyteidl.admin.Execution`"
   "RelaunchExecution", ":ref:`ref_flyteidl.admin.ExecutionRelaunchRequest`", ":ref:`ref_flyteidl.admin.ExecutionCreateResponse`", "Triggers the creation of an identical :ref:`ref_flyteidl.admin.Execution`"
   "RecoverExecution", ":ref:`ref_flyteidl.admin.ExecutionRecoverRequest`", ":ref:`ref_flyteidl.admin.ExecutionCreateResponse`", "Recreates a previously-run workflow execution that will only start executing from the last known failure point. In Recover mode, users cannot change any input parameters or update the version of the execution. This is extremely useful to recover from system errors and byzantine faults like - Loss of K8s cluster, bugs in platform or instability, machine failures, downstream system failures (downstream services), or simply to recover executions that failed because of retry exhaustion and should complete if tried again. See :ref:`ref_flyteidl.admin.ExecutionRecoverRequest` for more details."
   "GetExecution", ":ref:`ref_flyteidl.admin.WorkflowExecutionGetRequest`", ":ref:`ref_flyteidl.admin.Execution`", "Fetches a :ref:`ref_flyteidl.admin.Execution`."
   "UpdateExecution", ":ref:`ref_flyteidl.admin.ExecutionUpdateRequest`", ":ref:`ref_flyteidl.admin.ExecutionUpdateResponse`", "Update execution belonging to project domain :ref:`ref_flyteidl.admin.Execution`."
   "GetExecutionData", ":ref:`ref_flyteidl.admin.WorkflowExecutionGetDataRequest`", ":ref:`ref_flyteidl.admin.WorkflowExecutionGetDataResponse`", "Fetches input and output data for a :ref:`ref_flyteidl.admin.Execution`."
   "ListExecutions", ":ref:`ref_flyteidl.admin.ResourceListRequest`", ":ref:`ref_flyteidl.admin.ExecutionList`", "Fetch a list of :ref:`ref_flyteidl.admin.Execution`."
   "TerminateExecution", ":ref:`ref_flyteidl.admin.ExecutionTerminateRequest`", ":ref:`ref_flyteidl.admin.ExecutionTerminateResponse`", "Terminates an in-progress :ref:`ref_flyteidl.admin.Execution`."
   "GetNodeExecution", ":ref:`ref_flyteidl.admin.NodeExecutionGetRequest`", ":ref:`ref_flyteidl.admin.NodeExecution`", "Fetches a :ref:`ref_flyteidl.admin.NodeExecution`."
   "ListNodeExecutions", ":ref:`ref_flyteidl.admin.NodeExecutionListRequest`", ":ref:`ref_flyteidl.admin.NodeExecutionList`", "Fetch a list of :ref:`ref_flyteidl.admin.NodeExecution`."
   "ListNodeExecutionsForTask", ":ref:`ref_flyteidl.admin.NodeExecutionForTaskListRequest`", ":ref:`ref_flyteidl.admin.NodeExecutionList`", "Fetch a list of :ref:`ref_flyteidl.admin.NodeExecution` launched by the reference :ref:`ref_flyteidl.admin.TaskExecution`."
   "GetNodeExecutionData", ":ref:`ref_flyteidl.admin.NodeExecutionGetDataRequest`", ":ref:`ref_flyteidl.admin.NodeExecutionGetDataResponse`", "Fetches input and output data for a :ref:`ref_flyteidl.admin.NodeExecution`."
   "RegisterProject", ":ref:`ref_flyteidl.admin.ProjectRegisterRequest`", ":ref:`ref_flyteidl.admin.ProjectRegisterResponse`", "Registers a :ref:`ref_flyteidl.admin.Project` with the Flyte deployment."
   "UpdateProject", ":ref:`ref_flyteidl.admin.Project`", ":ref:`ref_flyteidl.admin.ProjectUpdateResponse`", "Updates an existing :ref:`ref_flyteidl.admin.Project` flyteidl.admin.Project should be passed but the domains property should be empty; it will be ignored in the handler as domains cannot be updated via this API."
   "ListProjects", ":ref:`ref_flyteidl.admin.ProjectListRequest`", ":ref:`ref_flyteidl.admin.Projects`", "Fetches a list of :ref:`ref_flyteidl.admin.Project`"
   "CreateWorkflowEvent", ":ref:`ref_flyteidl.admin.WorkflowExecutionEventRequest`", ":ref:`ref_flyteidl.admin.WorkflowExecutionEventResponse`", "Indicates a :ref:`ref_flyteidl.event.WorkflowExecutionEvent` has occurred."
   "CreateNodeEvent", ":ref:`ref_flyteidl.admin.NodeExecutionEventRequest`", ":ref:`ref_flyteidl.admin.NodeExecutionEventResponse`", "Indicates a :ref:`ref_flyteidl.event.NodeExecutionEvent` has occurred."
   "CreateTaskEvent", ":ref:`ref_flyteidl.admin.TaskExecutionEventRequest`", ":ref:`ref_flyteidl.admin.TaskExecutionEventResponse`", "Indicates a :ref:`ref_flyteidl.event.TaskExecutionEvent` has occurred."
   "GetTaskExecution", ":ref:`ref_flyteidl.admin.TaskExecutionGetRequest`", ":ref:`ref_flyteidl.admin.TaskExecution`", "Fetches a :ref:`ref_flyteidl.admin.TaskExecution`."
   "ListTaskExecutions", ":ref:`ref_flyteidl.admin.TaskExecutionListRequest`", ":ref:`ref_flyteidl.admin.TaskExecutionList`", "Fetches a list of :ref:`ref_flyteidl.admin.TaskExecution`."
   "GetTaskExecutionData", ":ref:`ref_flyteidl.admin.TaskExecutionGetDataRequest`", ":ref:`ref_flyteidl.admin.TaskExecutionGetDataResponse`", "Fetches input and output data for a :ref:`ref_flyteidl.admin.TaskExecution`."
   "UpdateProjectDomainAttributes", ":ref:`ref_flyteidl.admin.ProjectDomainAttributesUpdateRequest`", ":ref:`ref_flyteidl.admin.ProjectDomainAttributesUpdateResponse`", "Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain."
   "GetProjectDomainAttributes", ":ref:`ref_flyteidl.admin.ProjectDomainAttributesGetRequest`", ":ref:`ref_flyteidl.admin.ProjectDomainAttributesGetResponse`", "Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain."
   "DeleteProjectDomainAttributes", ":ref:`ref_flyteidl.admin.ProjectDomainAttributesDeleteRequest`", ":ref:`ref_flyteidl.admin.ProjectDomainAttributesDeleteResponse`", "Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain."
   "UpdateProjectAttributes", ":ref:`ref_flyteidl.admin.ProjectAttributesUpdateRequest`", ":ref:`ref_flyteidl.admin.ProjectAttributesUpdateResponse`", "Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` at the project level"
   "GetProjectAttributes", ":ref:`ref_flyteidl.admin.ProjectAttributesGetRequest`", ":ref:`ref_flyteidl.admin.ProjectAttributesGetResponse`", "Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain."
   "DeleteProjectAttributes", ":ref:`ref_flyteidl.admin.ProjectAttributesDeleteRequest`", ":ref:`ref_flyteidl.admin.ProjectAttributesDeleteResponse`", "Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project and domain."
   "UpdateWorkflowAttributes", ":ref:`ref_flyteidl.admin.WorkflowAttributesUpdateRequest`", ":ref:`ref_flyteidl.admin.WorkflowAttributesUpdateResponse`", "Creates or updates custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow."
   "GetWorkflowAttributes", ":ref:`ref_flyteidl.admin.WorkflowAttributesGetRequest`", ":ref:`ref_flyteidl.admin.WorkflowAttributesGetResponse`", "Fetches custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow."
   "DeleteWorkflowAttributes", ":ref:`ref_flyteidl.admin.WorkflowAttributesDeleteRequest`", ":ref:`ref_flyteidl.admin.WorkflowAttributesDeleteResponse`", "Deletes custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a project, domain and workflow."
   "ListMatchableAttributes", ":ref:`ref_flyteidl.admin.ListMatchableAttributesRequest`", ":ref:`ref_flyteidl.admin.ListMatchableAttributesResponse`", "Lists custom :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration` for a specific resource type."
   "ListNamedEntities", ":ref:`ref_flyteidl.admin.NamedEntityListRequest`", ":ref:`ref_flyteidl.admin.NamedEntityList`", "Returns a list of :ref:`ref_flyteidl.admin.NamedEntity` objects."
   "GetNamedEntity", ":ref:`ref_flyteidl.admin.NamedEntityGetRequest`", ":ref:`ref_flyteidl.admin.NamedEntity`", "Returns a :ref:`ref_flyteidl.admin.NamedEntity` object."
   "UpdateNamedEntity", ":ref:`ref_flyteidl.admin.NamedEntityUpdateRequest`", ":ref:`ref_flyteidl.admin.NamedEntityUpdateResponse`", "Updates a :ref:`ref_flyteidl.admin.NamedEntity` object."
   "GetVersion", ":ref:`ref_flyteidl.admin.GetVersionRequest`", ":ref:`ref_flyteidl.admin.GetVersionResponse`", ""
   "GetDescriptionEntity", ":ref:`ref_flyteidl.admin.ObjectGetRequest`", ":ref:`ref_flyteidl.admin.DescriptionEntity`", "Fetch a :ref:`ref_flyteidl.admin.DescriptionEntity` object."
   "ListDescriptionEntities", ":ref:`ref_flyteidl.admin.DescriptionEntityListRequest`", ":ref:`ref_flyteidl.admin.DescriptionEntityList`", "Fetch a list of :ref:`ref_flyteidl.admin.DescriptionEntity` definitions."
   "GetExecutionMetrics", ":ref:`ref_flyteidl.admin.WorkflowExecutionGetMetricsRequest`", ":ref:`ref_flyteidl.admin.WorkflowExecutionGetMetricsResponse`", "Fetches runtime metrics for a :ref:`ref_flyteidl.admin.Execution`."

..
   end services




.. _ref_flyteidl/service/agent.proto:

flyteidl/service/agent.proto
==================================================================




..
   end messages


..
   end enums


..
   end HasExtensions



.. _ref_flyteidl.service.AgentMetadataService:

AgentMetadataService
------------------------------------------------------------------

AgentMetadataService defines an RPC service that is also served over HTTP via grpc-gateway.
This service allows propeller or users to get the metadata of agents.

.. csv-table:: AgentMetadataService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "GetAgent", ":ref:`ref_flyteidl.admin.GetAgentRequest`", ":ref:`ref_flyteidl.admin.GetAgentResponse`", "Fetch a :ref:`ref_flyteidl.admin.Agent` definition."
   "ListAgents", ":ref:`ref_flyteidl.admin.ListAgentsRequest`", ":ref:`ref_flyteidl.admin.ListAgentsResponse`", "Fetch a list of :ref:`ref_flyteidl.admin.Agent` definitions."


.. _ref_flyteidl.service.AsyncAgentService:

AsyncAgentService
------------------------------------------------------------------

AsyncAgentService defines an RPC Service that allows propeller to send the request to the agent server.

.. csv-table:: AsyncAgentService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "CreateTask", ":ref:`ref_flyteidl.admin.CreateTaskRequest`", ":ref:`ref_flyteidl.admin.CreateTaskResponse`", "Send a task create request to the agent server."
   "GetTask", ":ref:`ref_flyteidl.admin.GetTaskRequest`", ":ref:`ref_flyteidl.admin.GetTaskResponse`", "Get job status."
   "DeleteTask", ":ref:`ref_flyteidl.admin.DeleteTaskRequest`", ":ref:`ref_flyteidl.admin.DeleteTaskResponse`", "Delete the task resource."

..
   end services




.. _ref_flyteidl/service/auth.proto:

flyteidl/service/auth.proto
==================================================================





.. _ref_flyteidl.service.OAuth2MetadataRequest:

OAuth2MetadataRequest
------------------------------------------------------------------










.. _ref_flyteidl.service.OAuth2MetadataResponse:

OAuth2MetadataResponse
------------------------------------------------------------------

OAuth2MetadataResponse defines an RFC-Compliant response for /.well-known/oauth-authorization-server metadata
as defined in https://tools.ietf.org/html/rfc8414



.. csv-table:: OAuth2MetadataResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "issuer", ":ref:`ref_string`", "", "Defines the issuer string in all JWT tokens this server issues. The issuer can be admin itself or an external issuer."
   "authorization_endpoint", ":ref:`ref_string`", "", "URL of the authorization server's authorization endpoint [RFC6749]. This is REQUIRED unless no grant types are supported that use the authorization endpoint."
   "token_endpoint", ":ref:`ref_string`", "", "URL of the authorization server's token endpoint [RFC6749]."
   "response_types_supported", ":ref:`ref_string`", "repeated", "Array containing a list of the OAuth 2.0 response_type values that this authorization server supports."
   "scopes_supported", ":ref:`ref_string`", "repeated", "JSON array containing a list of the OAuth 2.0 [RFC6749] scope values that this authorization server supports."
   "token_endpoint_auth_methods_supported", ":ref:`ref_string`", "repeated", "JSON array containing a list of client authentication methods supported by this token endpoint."
   "jwks_uri", ":ref:`ref_string`", "", "URL of the authorization server's JWK Set [JWK] document. The referenced document contains the signing key(s) the client uses to validate signatures from the authorization server."
   "code_challenge_methods_supported", ":ref:`ref_string`", "repeated", "JSON array containing a list of Proof Key for Code Exchange (PKCE) [RFC7636] code challenge methods supported by this authorization server."
   "grant_types_supported", ":ref:`ref_string`", "repeated", "JSON array containing a list of the OAuth 2.0 grant type values that this authorization server supports."
   "device_authorization_endpoint", ":ref:`ref_string`", "", "URL of the authorization server's device authorization endpoint, as defined in Section 3.1 of [RFC8628]"







.. _ref_flyteidl.service.PublicClientAuthConfigRequest:

PublicClientAuthConfigRequest
------------------------------------------------------------------










.. _ref_flyteidl.service.PublicClientAuthConfigResponse:

PublicClientAuthConfigResponse
------------------------------------------------------------------

FlyteClientResponse encapsulates public information that flyte clients (CLIs... etc.) can use to authenticate users.



.. csv-table:: PublicClientAuthConfigResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "client_id", ":ref:`ref_string`", "", "client_id to use when initiating OAuth2 authorization requests."
   "redirect_uri", ":ref:`ref_string`", "", "redirect uri to use when initiating OAuth2 authorization requests."
   "scopes", ":ref:`ref_string`", "repeated", "scopes to request when initiating OAuth2 authorization requests."
   "authorization_metadata_key", ":ref:`ref_string`", "", "Authorization Header to use when passing Access Tokens to the server. If not provided, the client should use the default http `Authorization` header."
   "service_http_endpoint", ":ref:`ref_string`", "", "ServiceHttpEndpoint points to the http endpoint for the backend. If empty, clients can assume the endpoint used to configure the gRPC connection can be used for the http one respecting the insecure flag to choose between SSL or no SSL connections."
   "audience", ":ref:`ref_string`", "", "audience to use when initiating OAuth2 authorization requests."






..
   end messages


..
   end enums


..
   end HasExtensions



.. _ref_flyteidl.service.AuthMetadataService:

AuthMetadataService
------------------------------------------------------------------

The following defines an RPC service that is also served over HTTP via grpc-gateway.
Standard response codes for both are defined here: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go
RPCs defined in this service must be anonymously accessible.

.. csv-table:: AuthMetadataService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "GetOAuth2Metadata", ":ref:`ref_flyteidl.service.OAuth2MetadataRequest`", ":ref:`ref_flyteidl.service.OAuth2MetadataResponse`", "Anonymously accessible. Retrieves local or external oauth authorization server metadata."
   "GetPublicClientConfig", ":ref:`ref_flyteidl.service.PublicClientAuthConfigRequest`", ":ref:`ref_flyteidl.service.PublicClientAuthConfigResponse`", "Anonymously accessible. Retrieves the client information clients should use when initiating OAuth2 authorization requests."

..
   end services




.. _ref_flyteidl/service/cache.proto:

flyteidl/service/cache.proto
==================================================================





.. _ref_flyteidl.service.EvictCacheResponse:

EvictCacheResponse
------------------------------------------------------------------





.. csv-table:: EvictCacheResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "errors", ":ref:`ref_flyteidl.core.CacheEvictionErrorList`", "", "List of errors encountered during cache eviction (if any)."







.. _ref_flyteidl.service.EvictTaskExecutionCacheRequest:

EvictTaskExecutionCacheRequest
------------------------------------------------------------------





.. csv-table:: EvictTaskExecutionCacheRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "task_execution_id", ":ref:`ref_flyteidl.core.TaskExecutionIdentifier`", "", "Identifier of :ref:`ref_flyteidl.admin.TaskExecution` to evict cache for."






..
   end messages


..
   end enums


..
   end HasExtensions



.. _ref_flyteidl.service.CacheService:

CacheService
------------------------------------------------------------------

CacheService defines an RPC Service for interacting with cached data in Flyte on a high level basis.

.. csv-table:: CacheService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "EvictTaskExecutionCache", ":ref:`ref_flyteidl.service.EvictTaskExecutionCacheRequest`", ":ref:`ref_flyteidl.service.EvictCacheResponse`", "Evicts all cached data for the referenced :ref:`ref_flyteidl.admin.TaskExecution`."

..
   end services




.. _ref_flyteidl/service/dataproxy.proto:

flyteidl/service/dataproxy.proto
==================================================================





.. _ref_flyteidl.service.CreateDownloadLinkRequest:

CreateDownloadLinkRequest
------------------------------------------------------------------

CreateDownloadLinkRequest defines the request parameters to create a download link (signed url)



.. csv-table:: CreateDownloadLinkRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "artifact_type", ":ref:`ref_flyteidl.service.ArtifactType`", "", "ArtifactType of the artifact requested."
   "expires_in", ":ref:`ref_google.protobuf.Duration`", "", "ExpiresIn defines a requested expiration duration for the generated url. The request will be rejected if this exceeds the platform allowed max. +optional. The default value comes from a global config."
   "node_execution_id", ":ref:`ref_flyteidl.core.NodeExecutionIdentifier`", "", "NodeId is the unique identifier for the node execution. For a task node, this will retrieve the output of the most recent attempt of the task."







.. _ref_flyteidl.service.CreateDownloadLinkResponse:

CreateDownloadLinkResponse
------------------------------------------------------------------

CreateDownloadLinkResponse defines the response for the generated links



.. csv-table:: CreateDownloadLinkResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signed_url", ":ref:`ref_string`", "repeated", "**Deprecated.** SignedUrl specifies the url to use to download content from (e.g. https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...)"
   "expires_at", ":ref:`ref_google.protobuf.Timestamp`", "", "**Deprecated.** ExpiresAt defines when will the signed URL expire."
   "pre_signed_urls", ":ref:`ref_flyteidl.service.PreSignedURLs`", "", "New wrapper object containing the signed urls and expiration time"







.. _ref_flyteidl.service.CreateDownloadLocationRequest:

CreateDownloadLocationRequest
------------------------------------------------------------------

CreateDownloadLocationRequest specified request for the CreateDownloadLocation API.



.. csv-table:: CreateDownloadLocationRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "native_url", ":ref:`ref_string`", "", "NativeUrl specifies the url in the format of the configured storage provider (e.g. s3://my-bucket/randomstring/suffix.tar)"
   "expires_in", ":ref:`ref_google.protobuf.Duration`", "", "ExpiresIn defines a requested expiration duration for the generated url. The request will be rejected if this exceeds the platform allowed max. +optional. The default value comes from a global config."







.. _ref_flyteidl.service.CreateDownloadLocationResponse:

CreateDownloadLocationResponse
------------------------------------------------------------------





.. csv-table:: CreateDownloadLocationResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signed_url", ":ref:`ref_string`", "", "SignedUrl specifies the url to use to download content from (e.g. https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...)"
   "expires_at", ":ref:`ref_google.protobuf.Timestamp`", "", "ExpiresAt defines when will the signed URL expires."







.. _ref_flyteidl.service.CreateUploadLocationRequest:

CreateUploadLocationRequest
------------------------------------------------------------------

CreateUploadLocationRequest specified request for the CreateUploadLocation API.
The implementation in data proxy service will create the s3 location with some server side configured prefixes,
and then:
  - project/domain/(a deterministic str representation of the content_md5)/filename (if present); OR
  - project/domain/filename_root (if present)/filename (if present).



.. csv-table:: CreateUploadLocationRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Project to create the upload location for +required"
   "domain", ":ref:`ref_string`", "", "Domain to create the upload location for. +required"
   "filename", ":ref:`ref_string`", "", "Filename specifies a desired suffix for the generated location. E.g. `file.py` or `pre/fix/file.zip`. +optional. By default, the service will generate a consistent name based on the provided parameters."
   "expires_in", ":ref:`ref_google.protobuf.Duration`", "", "ExpiresIn defines a requested expiration duration for the generated url. The request will be rejected if this exceeds the platform allowed max. +optional. The default value comes from a global config."
   "content_md5", ":ref:`ref_bytes`", "", "ContentMD5 restricts the upload location to the specific MD5 provided. The ContentMD5 will also appear in the generated path. +required"
   "filename_root", ":ref:`ref_string`", "", "If present, data proxy will use this string in lieu of the md5 hash in the path. When the filename is also included this makes the upload location deterministic. The native url will still be prefixed by the upload location prefix in data proxy config. This option is useful when uploading multiple files. +optional"







.. _ref_flyteidl.service.CreateUploadLocationResponse:

CreateUploadLocationResponse
------------------------------------------------------------------





.. csv-table:: CreateUploadLocationResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signed_url", ":ref:`ref_string`", "", "SignedUrl specifies the url to use to upload content to (e.g. https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...)"
   "native_url", ":ref:`ref_string`", "", "NativeUrl specifies the url in the format of the configured storage provider (e.g. s3://my-bucket/randomstring/suffix.tar)"
   "expires_at", ":ref:`ref_google.protobuf.Timestamp`", "", "ExpiresAt defines when will the signed URL expires."







.. _ref_flyteidl.service.GetDataRequest:

GetDataRequest
------------------------------------------------------------------

General request artifact to retrieve data from a Flyte artifact url.



.. csv-table:: GetDataRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "flyte_url", ":ref:`ref_string`", "", "A unique identifier in the form of flyte://<something> that uniquely, for a given Flyte backend, identifies a Flyte artifact ([i]nput, [o]output, flyte [d]eck, etc.). e.g. flyte://v1/proj/development/execid/n2/0/i (for 0th task execution attempt input) flyte://v1/proj/development/execid/n2/i (for node execution input) flyte://v1/proj/development/execid/n2/o/o3 (the o3 output of the second node)"







.. _ref_flyteidl.service.GetDataResponse:

GetDataResponse
------------------------------------------------------------------





.. csv-table:: GetDataResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "literal_map", ":ref:`ref_flyteidl.core.LiteralMap`", "", "literal map data will be returned"
   "pre_signed_urls", ":ref:`ref_flyteidl.service.PreSignedURLs`", "", "Flyte deck html will be returned as a signed url users can download"
   "literal", ":ref:`ref_flyteidl.core.Literal`", "", "Single literal will be returned. This is returned when the user/url requests a specific output or input by name. See the o3 example above."







.. _ref_flyteidl.service.PreSignedURLs:

PreSignedURLs
------------------------------------------------------------------

Wrapper object since the message is shared across this and the GetDataResponse



.. csv-table:: PreSignedURLs type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signed_url", ":ref:`ref_string`", "repeated", "SignedUrl specifies the url to use to download content from (e.g. https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...)"
   "expires_at", ":ref:`ref_google.protobuf.Timestamp`", "", "ExpiresAt defines when will the signed URL expire."






..
   end messages



.. _ref_flyteidl.service.ArtifactType:

ArtifactType
------------------------------------------------------------------

ArtifactType

.. csv-table:: Enum ArtifactType values
   :header: "Name", "Number", "Description"
   :widths: auto

   "ARTIFACT_TYPE_UNDEFINED", "0", "ARTIFACT_TYPE_UNDEFINED is the default, often invalid, value for the enum."
   "ARTIFACT_TYPE_DECK", "1", "ARTIFACT_TYPE_DECK refers to the deck html file optionally generated after a task, a workflow or a launch plan finishes executing."


..
   end enums


..
   end HasExtensions



.. _ref_flyteidl.service.DataProxyService:

DataProxyService
------------------------------------------------------------------

DataProxyService defines an RPC Service that allows access to user-data in a controlled manner.

.. csv-table:: DataProxyService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "CreateUploadLocation", ":ref:`ref_flyteidl.service.CreateUploadLocationRequest`", ":ref:`ref_flyteidl.service.CreateUploadLocationResponse`", "CreateUploadLocation creates a signed url to upload artifacts to for a given project/domain."
   "CreateDownloadLocation", ":ref:`ref_flyteidl.service.CreateDownloadLocationRequest`", ":ref:`ref_flyteidl.service.CreateDownloadLocationResponse`", "CreateDownloadLocation creates a signed url to download artifacts."
   "CreateDownloadLink", ":ref:`ref_flyteidl.service.CreateDownloadLinkRequest`", ":ref:`ref_flyteidl.service.CreateDownloadLinkResponse`", "CreateDownloadLocation creates a signed url to download artifacts."
   "GetData", ":ref:`ref_flyteidl.service.GetDataRequest`", ":ref:`ref_flyteidl.service.GetDataResponse`", ""

..
   end services




.. _ref_flyteidl/service/external_plugin_service.proto:

flyteidl/service/external_plugin_service.proto
==================================================================





.. _ref_flyteidl.service.TaskCreateRequest:

TaskCreateRequest
------------------------------------------------------------------

Represents a request structure to create task.



.. csv-table:: TaskCreateRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "inputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "The inputs required to start the execution. All required inputs must be included in this map. If not required and not provided, defaults apply. +optional"
   "template", ":ref:`ref_flyteidl.core.TaskTemplate`", "", "Template of the task that encapsulates all the metadata of the task."
   "output_prefix", ":ref:`ref_string`", "", "Prefix for where task output data will be written. (e.g. s3://my-bucket/randomstring)"







.. _ref_flyteidl.service.TaskCreateResponse:

TaskCreateResponse
------------------------------------------------------------------

Represents a create response structure.



.. csv-table:: TaskCreateResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "job_id", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.service.TaskDeleteRequest:

TaskDeleteRequest
------------------------------------------------------------------

A message used to delete a task.



.. csv-table:: TaskDeleteRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "task_type", ":ref:`ref_string`", "", "A predefined yet extensible Task type identifier."
   "job_id", ":ref:`ref_string`", "", "The unique id identifying the job."







.. _ref_flyteidl.service.TaskDeleteResponse:

TaskDeleteResponse
------------------------------------------------------------------

Response to delete a task.








.. _ref_flyteidl.service.TaskGetRequest:

TaskGetRequest
------------------------------------------------------------------

A message used to fetch a job state from backend plugin server.



.. csv-table:: TaskGetRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "task_type", ":ref:`ref_string`", "", "A predefined yet extensible Task type identifier."
   "job_id", ":ref:`ref_string`", "", "The unique id identifying the job."







.. _ref_flyteidl.service.TaskGetResponse:

TaskGetResponse
------------------------------------------------------------------

Response to get an individual task state.



.. csv-table:: TaskGetResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "state", ":ref:`ref_flyteidl.service.State`", "", "The state of the execution is used to control its visibility in the UI/CLI."
   "outputs", ":ref:`ref_flyteidl.core.LiteralMap`", "", "The outputs of the execution. It's typically used by sql task. Flyteplugins service will create a Structured dataset pointing to the query result table. +optional"






..
   end messages



.. _ref_flyteidl.service.State:

State
------------------------------------------------------------------

The state of the execution is used to control its visibility in the UI/CLI.

.. csv-table:: Enum State values
   :header: "Name", "Number", "Description"
   :widths: auto

   "RETRYABLE_FAILURE", "0", ""
   "PERMANENT_FAILURE", "1", ""
   "PENDING", "2", ""
   "RUNNING", "3", ""
   "SUCCEEDED", "4", ""


..
   end enums


..
   end HasExtensions



.. _ref_flyteidl.service.ExternalPluginService:

ExternalPluginService
------------------------------------------------------------------

ExternalPluginService defines an RPC Service that allows propeller to send the request to the backend plugin server.

.. csv-table:: ExternalPluginService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "CreateTask", ":ref:`ref_flyteidl.service.TaskCreateRequest`", ":ref:`ref_flyteidl.service.TaskCreateResponse`", "Send a task create request to the backend plugin server."
   "GetTask", ":ref:`ref_flyteidl.service.TaskGetRequest`", ":ref:`ref_flyteidl.service.TaskGetResponse`", "Get job status."
   "DeleteTask", ":ref:`ref_flyteidl.service.TaskDeleteRequest`", ":ref:`ref_flyteidl.service.TaskDeleteResponse`", "Delete the task resource."

..
   end services




.. _ref_flyteidl/service/identity.proto:

flyteidl/service/identity.proto
==================================================================





.. _ref_flyteidl.service.UserInfoRequest:

UserInfoRequest
------------------------------------------------------------------










.. _ref_flyteidl.service.UserInfoResponse:

UserInfoResponse
------------------------------------------------------------------

See the OpenID Connect spec at https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse for more information.



.. csv-table:: UserInfoResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "subject", ":ref:`ref_string`", "", "Locally unique and never reassigned identifier within the Issuer for the End-User, which is intended to be consumed by the Client."
   "name", ":ref:`ref_string`", "", "Full name"
   "preferred_username", ":ref:`ref_string`", "", "Shorthand name by which the End-User wishes to be referred to"
   "given_name", ":ref:`ref_string`", "", "Given name(s) or first name(s)"
   "family_name", ":ref:`ref_string`", "", "Surname(s) or last name(s)"
   "email", ":ref:`ref_string`", "", "Preferred e-mail address"
   "picture", ":ref:`ref_string`", "", "Profile picture URL"
   "additional_claims", ":ref:`ref_google.protobuf.Struct`", "", "Additional claims"






..
   end messages


..
   end enums


..
   end HasExtensions



.. _ref_flyteidl.service.IdentityService:

IdentityService
------------------------------------------------------------------

IdentityService defines an RPC Service that interacts with user/app identities.

.. csv-table:: IdentityService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "UserInfo", ":ref:`ref_flyteidl.service.UserInfoRequest`", ":ref:`ref_flyteidl.service.UserInfoResponse`", "Retrieves user information about the currently logged in user."

..
   end services




.. _ref_flyteidl/service/signal.proto:

flyteidl/service/signal.proto
==================================================================




..
   end messages


..
   end enums


..
   end HasExtensions



.. _ref_flyteidl.service.SignalService:

SignalService
------------------------------------------------------------------

SignalService defines an RPC Service that may create, update, and retrieve signal(s).

.. csv-table:: SignalService service methods
   :header: "Method Name", "Request Type", "Response Type", "Description"
   :widths: auto

   "GetOrCreateSignal", ":ref:`ref_flyteidl.admin.SignalGetOrCreateRequest`", ":ref:`ref_flyteidl.admin.Signal`", "Fetches or creates a :ref:`ref_flyteidl.admin.Signal`.

Purposefully left out an HTTP API for this RPC call. This is meant to idempotently retrieve a signal, meaning the first call will create the signal and all subsequent calls will fetch the existing signal. This is only useful during Flyte Workflow execution and therefore is not exposed to mitigate unintended behavior. option (grpc.gateway.protoc_gen_swagger.options.openapiv2_operation) = { description: "Retrieve a signal, creating it if it does not exist." };"
   "ListSignals", ":ref:`ref_flyteidl.admin.SignalListRequest`", ":ref:`ref_flyteidl.admin.SignalList`", "Fetch a list of :ref:`ref_flyteidl.admin.Signal` definitions."
   "SetSignal", ":ref:`ref_flyteidl.admin.SignalSetRequest`", ":ref:`ref_flyteidl.admin.SignalSetResponse`", "Sets the value on a :ref:`ref_flyteidl.admin.Signal` definition"

..
   end services


