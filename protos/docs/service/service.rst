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

   "signed_url", ":ref:`ref_string`", "repeated", "SignedUrl specifies the url to use to download content from (e.g. https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...)"
   "expires_at", ":ref:`ref_google.protobuf.Timestamp`", "", "ExpiresAt defines when will the signed URL expire."







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



.. csv-table:: CreateUploadLocationRequest type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "project", ":ref:`ref_string`", "", "Project to create the upload location for +required"
   "domain", ":ref:`ref_string`", "", "Domain to create the upload location for. +required"
   "filename", ":ref:`ref_string`", "", "Filename specifies a desired suffix for the generated location. E.g. `file.py` or `pre/fix/file.zip`. +optional. By default, the service will generate a consistent name based on the provided parameters."
   "expires_in", ":ref:`ref_google.protobuf.Duration`", "", "ExpiresIn defines a requested expiration duration for the generated url. The request will be rejected if this exceeds the platform allowed max. +optional. The default value comes from a global config."
   "content_md5", ":ref:`ref_bytes`", "", "ContentMD5 restricts the upload location to the specific MD5 provided. The ContentMD5 will also appear in the generated path. +required"







.. _ref_flyteidl.service.CreateUploadLocationResponse:

CreateUploadLocationResponse
------------------------------------------------------------------





.. csv-table:: CreateUploadLocationResponse type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "signed_url", ":ref:`ref_string`", "", "SignedUrl specifies the url to use to upload content to (e.g. https://my-bucket.s3.amazonaws.com/randomstring/suffix.tar?X-...)"
   "native_url", ":ref:`ref_string`", "", "NativeUrl specifies the url in the format of the configured storage provider (e.g. s3://my-bucket/randomstring/suffix.tar)"
   "expires_at", ":ref:`ref_google.protobuf.Timestamp`", "", "ExpiresAt defines when will the signed URL expires."






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

   "GetOrCreateSignal", ":ref:`ref_flyteidl.admin.SignalGetOrCreateRequest`", ":ref:`ref_flyteidl.admin.Signal`", "Fetches or creates a :ref:`ref_flyteidl.admin.Signal`."
   "ListSignals", ":ref:`ref_flyteidl.admin.SignalListRequest`", ":ref:`ref_flyteidl.admin.SignalList`", "Fetch a list of :ref:`ref_flyteidl.admin.Signal` definitions."
   "SetSignal", ":ref:`ref_flyteidl.admin.SignalSetRequest`", ":ref:`ref_flyteidl.admin.SignalSetResponse`", "Sets the value on a :ref:`ref_flyteidl.admin.Signal` definition"

..
   end services


