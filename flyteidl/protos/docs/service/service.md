# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [flyteidl/service/admin.proto](#flyteidl/service/admin.proto)
    - [AdminService](#flyteidl.service.AdminService)
  
- [flyteidl/service/auth.proto](#flyteidl/service/auth.proto)
    - [OAuth2MetadataRequest](#flyteidl.service.OAuth2MetadataRequest)
    - [OAuth2MetadataResponse](#flyteidl.service.OAuth2MetadataResponse)
    - [PublicClientAuthConfigRequest](#flyteidl.service.PublicClientAuthConfigRequest)
    - [PublicClientAuthConfigResponse](#flyteidl.service.PublicClientAuthConfigResponse)
  
    - [AuthMetadataService](#flyteidl.service.AuthMetadataService)
  
- [flyteidl/service/identity.proto](#flyteidl/service/identity.proto)
    - [UserInfoRequest](#flyteidl.service.UserInfoRequest)
    - [UserInfoResponse](#flyteidl.service.UserInfoResponse)
  
    - [IdentityService](#flyteidl.service.IdentityService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="flyteidl/service/admin.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/service/admin.proto


 

 

 


<a name="flyteidl.service.AdminService"></a>

### AdminService
The following defines an RPC service that is also served over HTTP via grpc-gateway.
Standard response codes for both are defined here: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateTask | [.flyteidl.admin.TaskCreateRequest](#flyteidl.admin.TaskCreateRequest) | [.flyteidl.admin.TaskCreateResponse](#flyteidl.admin.TaskCreateResponse) |  |
| GetTask | [.flyteidl.admin.ObjectGetRequest](#flyteidl.admin.ObjectGetRequest) | [.flyteidl.admin.Task](#flyteidl.admin.Task) |  |
| ListTaskIds | [.flyteidl.admin.NamedEntityIdentifierListRequest](#flyteidl.admin.NamedEntityIdentifierListRequest) | [.flyteidl.admin.NamedEntityIdentifierList](#flyteidl.admin.NamedEntityIdentifierList) |  |
| ListTasks | [.flyteidl.admin.ResourceListRequest](#flyteidl.admin.ResourceListRequest) | [.flyteidl.admin.TaskList](#flyteidl.admin.TaskList) |  |
| CreateWorkflow | [.flyteidl.admin.WorkflowCreateRequest](#flyteidl.admin.WorkflowCreateRequest) | [.flyteidl.admin.WorkflowCreateResponse](#flyteidl.admin.WorkflowCreateResponse) |  |
| GetWorkflow | [.flyteidl.admin.ObjectGetRequest](#flyteidl.admin.ObjectGetRequest) | [.flyteidl.admin.Workflow](#flyteidl.admin.Workflow) |  |
| ListWorkflowIds | [.flyteidl.admin.NamedEntityIdentifierListRequest](#flyteidl.admin.NamedEntityIdentifierListRequest) | [.flyteidl.admin.NamedEntityIdentifierList](#flyteidl.admin.NamedEntityIdentifierList) |  |
| ListWorkflows | [.flyteidl.admin.ResourceListRequest](#flyteidl.admin.ResourceListRequest) | [.flyteidl.admin.WorkflowList](#flyteidl.admin.WorkflowList) |  |
| CreateLaunchPlan | [.flyteidl.admin.LaunchPlanCreateRequest](#flyteidl.admin.LaunchPlanCreateRequest) | [.flyteidl.admin.LaunchPlanCreateResponse](#flyteidl.admin.LaunchPlanCreateResponse) |  |
| GetLaunchPlan | [.flyteidl.admin.ObjectGetRequest](#flyteidl.admin.ObjectGetRequest) | [.flyteidl.admin.LaunchPlan](#flyteidl.admin.LaunchPlan) |  |
| GetActiveLaunchPlan | [.flyteidl.admin.ActiveLaunchPlanRequest](#flyteidl.admin.ActiveLaunchPlanRequest) | [.flyteidl.admin.LaunchPlan](#flyteidl.admin.LaunchPlan) |  |
| ListActiveLaunchPlans | [.flyteidl.admin.ActiveLaunchPlanListRequest](#flyteidl.admin.ActiveLaunchPlanListRequest) | [.flyteidl.admin.LaunchPlanList](#flyteidl.admin.LaunchPlanList) |  |
| ListLaunchPlanIds | [.flyteidl.admin.NamedEntityIdentifierListRequest](#flyteidl.admin.NamedEntityIdentifierListRequest) | [.flyteidl.admin.NamedEntityIdentifierList](#flyteidl.admin.NamedEntityIdentifierList) |  |
| ListLaunchPlans | [.flyteidl.admin.ResourceListRequest](#flyteidl.admin.ResourceListRequest) | [.flyteidl.admin.LaunchPlanList](#flyteidl.admin.LaunchPlanList) |  |
| UpdateLaunchPlan | [.flyteidl.admin.LaunchPlanUpdateRequest](#flyteidl.admin.LaunchPlanUpdateRequest) | [.flyteidl.admin.LaunchPlanUpdateResponse](#flyteidl.admin.LaunchPlanUpdateResponse) |  |
| CreateExecution | [.flyteidl.admin.ExecutionCreateRequest](#flyteidl.admin.ExecutionCreateRequest) | [.flyteidl.admin.ExecutionCreateResponse](#flyteidl.admin.ExecutionCreateResponse) |  |
| RelaunchExecution | [.flyteidl.admin.ExecutionRelaunchRequest](#flyteidl.admin.ExecutionRelaunchRequest) | [.flyteidl.admin.ExecutionCreateResponse](#flyteidl.admin.ExecutionCreateResponse) |  |
| GetExecution | [.flyteidl.admin.WorkflowExecutionGetRequest](#flyteidl.admin.WorkflowExecutionGetRequest) | [.flyteidl.admin.Execution](#flyteidl.admin.Execution) |  |
| GetExecutionData | [.flyteidl.admin.WorkflowExecutionGetDataRequest](#flyteidl.admin.WorkflowExecutionGetDataRequest) | [.flyteidl.admin.WorkflowExecutionGetDataResponse](#flyteidl.admin.WorkflowExecutionGetDataResponse) |  |
| ListExecutions | [.flyteidl.admin.ResourceListRequest](#flyteidl.admin.ResourceListRequest) | [.flyteidl.admin.ExecutionList](#flyteidl.admin.ExecutionList) |  |
| TerminateExecution | [.flyteidl.admin.ExecutionTerminateRequest](#flyteidl.admin.ExecutionTerminateRequest) | [.flyteidl.admin.ExecutionTerminateResponse](#flyteidl.admin.ExecutionTerminateResponse) |  |
| GetNodeExecution | [.flyteidl.admin.NodeExecutionGetRequest](#flyteidl.admin.NodeExecutionGetRequest) | [.flyteidl.admin.NodeExecution](#flyteidl.admin.NodeExecution) |  |
| ListNodeExecutions | [.flyteidl.admin.NodeExecutionListRequest](#flyteidl.admin.NodeExecutionListRequest) | [.flyteidl.admin.NodeExecutionList](#flyteidl.admin.NodeExecutionList) |  |
| ListNodeExecutionsForTask | [.flyteidl.admin.NodeExecutionForTaskListRequest](#flyteidl.admin.NodeExecutionForTaskListRequest) | [.flyteidl.admin.NodeExecutionList](#flyteidl.admin.NodeExecutionList) |  |
| GetNodeExecutionData | [.flyteidl.admin.NodeExecutionGetDataRequest](#flyteidl.admin.NodeExecutionGetDataRequest) | [.flyteidl.admin.NodeExecutionGetDataResponse](#flyteidl.admin.NodeExecutionGetDataResponse) |  |
| RegisterProject | [.flyteidl.admin.ProjectRegisterRequest](#flyteidl.admin.ProjectRegisterRequest) | [.flyteidl.admin.ProjectRegisterResponse](#flyteidl.admin.ProjectRegisterResponse) |  |
| UpdateProject | [.flyteidl.admin.Project](#flyteidl.admin.Project) | [.flyteidl.admin.ProjectUpdateResponse](#flyteidl.admin.ProjectUpdateResponse) | flyteidl.admin.Project should be passed but the domains property should be empty; it will be ignored in the handler as domains cannot be updated via this API. |
| ListProjects | [.flyteidl.admin.ProjectListRequest](#flyteidl.admin.ProjectListRequest) | [.flyteidl.admin.Projects](#flyteidl.admin.Projects) |  |
| CreateWorkflowEvent | [.flyteidl.admin.WorkflowExecutionEventRequest](#flyteidl.admin.WorkflowExecutionEventRequest) | [.flyteidl.admin.WorkflowExecutionEventResponse](#flyteidl.admin.WorkflowExecutionEventResponse) |  |
| CreateNodeEvent | [.flyteidl.admin.NodeExecutionEventRequest](#flyteidl.admin.NodeExecutionEventRequest) | [.flyteidl.admin.NodeExecutionEventResponse](#flyteidl.admin.NodeExecutionEventResponse) |  |
| CreateTaskEvent | [.flyteidl.admin.TaskExecutionEventRequest](#flyteidl.admin.TaskExecutionEventRequest) | [.flyteidl.admin.TaskExecutionEventResponse](#flyteidl.admin.TaskExecutionEventResponse) |  |
| GetTaskExecution | [.flyteidl.admin.TaskExecutionGetRequest](#flyteidl.admin.TaskExecutionGetRequest) | [.flyteidl.admin.TaskExecution](#flyteidl.admin.TaskExecution) |  |
| ListTaskExecutions | [.flyteidl.admin.TaskExecutionListRequest](#flyteidl.admin.TaskExecutionListRequest) | [.flyteidl.admin.TaskExecutionList](#flyteidl.admin.TaskExecutionList) |  |
| GetTaskExecutionData | [.flyteidl.admin.TaskExecutionGetDataRequest](#flyteidl.admin.TaskExecutionGetDataRequest) | [.flyteidl.admin.TaskExecutionGetDataResponse](#flyteidl.admin.TaskExecutionGetDataResponse) |  |
| UpdateProjectDomainAttributes | [.flyteidl.admin.ProjectDomainAttributesUpdateRequest](#flyteidl.admin.ProjectDomainAttributesUpdateRequest) | [.flyteidl.admin.ProjectDomainAttributesUpdateResponse](#flyteidl.admin.ProjectDomainAttributesUpdateResponse) |  |
| GetProjectDomainAttributes | [.flyteidl.admin.ProjectDomainAttributesGetRequest](#flyteidl.admin.ProjectDomainAttributesGetRequest) | [.flyteidl.admin.ProjectDomainAttributesGetResponse](#flyteidl.admin.ProjectDomainAttributesGetResponse) |  |
| DeleteProjectDomainAttributes | [.flyteidl.admin.ProjectDomainAttributesDeleteRequest](#flyteidl.admin.ProjectDomainAttributesDeleteRequest) | [.flyteidl.admin.ProjectDomainAttributesDeleteResponse](#flyteidl.admin.ProjectDomainAttributesDeleteResponse) |  |
| UpdateWorkflowAttributes | [.flyteidl.admin.WorkflowAttributesUpdateRequest](#flyteidl.admin.WorkflowAttributesUpdateRequest) | [.flyteidl.admin.WorkflowAttributesUpdateResponse](#flyteidl.admin.WorkflowAttributesUpdateResponse) |  |
| GetWorkflowAttributes | [.flyteidl.admin.WorkflowAttributesGetRequest](#flyteidl.admin.WorkflowAttributesGetRequest) | [.flyteidl.admin.WorkflowAttributesGetResponse](#flyteidl.admin.WorkflowAttributesGetResponse) |  |
| DeleteWorkflowAttributes | [.flyteidl.admin.WorkflowAttributesDeleteRequest](#flyteidl.admin.WorkflowAttributesDeleteRequest) | [.flyteidl.admin.WorkflowAttributesDeleteResponse](#flyteidl.admin.WorkflowAttributesDeleteResponse) |  |
| ListMatchableAttributes | [.flyteidl.admin.ListMatchableAttributesRequest](#flyteidl.admin.ListMatchableAttributesRequest) | [.flyteidl.admin.ListMatchableAttributesResponse](#flyteidl.admin.ListMatchableAttributesResponse) |  |
| ListNamedEntities | [.flyteidl.admin.NamedEntityListRequest](#flyteidl.admin.NamedEntityListRequest) | [.flyteidl.admin.NamedEntityList](#flyteidl.admin.NamedEntityList) |  |
| GetNamedEntity | [.flyteidl.admin.NamedEntityGetRequest](#flyteidl.admin.NamedEntityGetRequest) | [.flyteidl.admin.NamedEntity](#flyteidl.admin.NamedEntity) |  |
| UpdateNamedEntity | [.flyteidl.admin.NamedEntityUpdateRequest](#flyteidl.admin.NamedEntityUpdateRequest) | [.flyteidl.admin.NamedEntityUpdateResponse](#flyteidl.admin.NamedEntityUpdateResponse) |  |
| GetVersion | [.flyteidl.admin.GetVersionRequest](#flyteidl.admin.GetVersionRequest) | [.flyteidl.admin.GetVersionResponse](#flyteidl.admin.GetVersionResponse) |  |

 



<a name="flyteidl/service/auth.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/service/auth.proto



<a name="flyteidl.service.OAuth2MetadataRequest"></a>

### OAuth2MetadataRequest







<a name="flyteidl.service.OAuth2MetadataResponse"></a>

### OAuth2MetadataResponse
OAuth2MetadataResponse defines an RFC-Compliant response for /.well-known/oauth-authorization-server metadata
as defined in https://tools.ietf.org/html/rfc8414


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| issuer | [string](#string) |  | Defines the issuer string in all JWT tokens this server issues. The issuer can be admin itself or an external issuer. |
| authorization_endpoint | [string](#string) |  | URL of the authorization server&#39;s authorization endpoint [RFC6749]. This is REQUIRED unless no grant types are supported that use the authorization endpoint. |
| token_endpoint | [string](#string) |  | URL of the authorization server&#39;s token endpoint [RFC6749]. |
| response_types_supported | [string](#string) | repeated | Array containing a list of the OAuth 2.0 &#34;response_type&#34; values that this authorization server supports. |
| scopes_supported | [string](#string) | repeated | JSON array containing a list of the OAuth 2.0 [RFC6749] &#34;scope&#34; values that this authorization server supports. |
| token_endpoint_auth_methods_supported | [string](#string) | repeated | JSON array containing a list of client authentication methods supported by this token endpoint. |
| jwks_uri | [string](#string) |  | URL of the authorization server&#39;s JWK Set [JWK] document. The referenced document contains the signing key(s) the client uses to validate signatures from the authorization server. |
| code_challenge_methods_supported | [string](#string) | repeated | JSON array containing a list of Proof Key for Code Exchange (PKCE) [RFC7636] code challenge methods supported by this authorization server. |
| grant_types_supported | [string](#string) | repeated | JSON array containing a list of the OAuth 2.0 grant type values that this authorization server supports. |






<a name="flyteidl.service.PublicClientAuthConfigRequest"></a>

### PublicClientAuthConfigRequest







<a name="flyteidl.service.PublicClientAuthConfigResponse"></a>

### PublicClientAuthConfigResponse
FlyteClientResponse encapsulates public information that flyte clients (CLIs... etc.) can use to authenticate users.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_id | [string](#string) |  | client_id to use when initiating OAuth2 authorization requests. |
| redirect_uri | [string](#string) |  | redirect uri to use when initiating OAuth2 authorization requests. |
| scopes | [string](#string) | repeated | scopes to request when initiating OAuth2 authorization requests. |
| authorization_metadata_key | [string](#string) |  | Authorization Header to use when passing Access Tokens to the server. If not provided, the client should use the default http `Authorization` header. |





 

 

 


<a name="flyteidl.service.AuthMetadataService"></a>

### AuthMetadataService
The following defines an RPC service that is also served over HTTP via grpc-gateway.
Standard response codes for both are defined here: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go
RPCs defined in this service must be anonymously accessible.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetOAuth2Metadata | [OAuth2MetadataRequest](#flyteidl.service.OAuth2MetadataRequest) | [OAuth2MetadataResponse](#flyteidl.service.OAuth2MetadataResponse) | Anonymously accessible. Retrieves local or external oauth authorization server metadata. |
| GetPublicClientConfig | [PublicClientAuthConfigRequest](#flyteidl.service.PublicClientAuthConfigRequest) | [PublicClientAuthConfigResponse](#flyteidl.service.PublicClientAuthConfigResponse) | Anonymously accessible. Retrieves the client information clients should use when initiating OAuth2 authorization requests. |

 



<a name="flyteidl/service/identity.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/service/identity.proto



<a name="flyteidl.service.UserInfoRequest"></a>

### UserInfoRequest







<a name="flyteidl.service.UserInfoResponse"></a>

### UserInfoResponse
See the OpenID Connect spec at https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse for more information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| subject | [string](#string) |  | Locally unique and never reassigned identifier within the Issuer for the End-User, which is intended to be consumed by the Client. |
| name | [string](#string) |  | Full name |
| preferred_username | [string](#string) |  | Shorthand name by which the End-User wishes to be referred to |
| given_name | [string](#string) |  | Given name(s) or first name(s) |
| family_name | [string](#string) |  | Surname(s) or last name(s) |
| email | [string](#string) |  | Preferred e-mail address |
| picture | [string](#string) |  | Profile picture URL |





 

 

 


<a name="flyteidl.service.IdentityService"></a>

### IdentityService
IdentityService defines an RPC Service that interacts with user/app identities.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| UserInfo | [UserInfoRequest](#flyteidl.service.UserInfoRequest) | [UserInfoResponse](#flyteidl.service.UserInfoResponse) | Retrieves user information about the currently logged in user. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

