# Project isolation

**Authors:**

- @robert-ulbrich-mercedes-benz


## 1 Executive Summary

This feature enables admins of Flyte to assign users of Flyte only to certain projects. That way users can only access and submit workflows to Flyte projects they are assigned to. This is a security feature to prevent unauthorized access to projects and their data. This RFC proposes a mechanism to isolate projects from each other. 

For the feature to work, authentication needs to be enabled. Flyte will evaluate the identity token issued by the IDP and extract the values of a configured claim. The claim values are mapped to projects in Flyte. If the claim value matches a project, the user is allowed to access the project. If the claim value does not match any project, the user is denied access.

## 2 Motivation

Having Flyte projects isolated from each other is a security feature to prevent unauthorized access to projects and their data. This is especially important when Flyte is used in a multi-tenant environment. This feature is also required to comply with data protection regulations. Especially in enterprise contexts with many different teams each having their own Flyte project, isolation is essential.

## 3 Proposed Implementation

The Flyte operators can enable or disable the project isolation feature. When enabled, Flyte will evaluate the identity token issued by the IDP and extract the values of a configured claim. The user can then 
access a project if the claim value matches the project. If the project isolation feature is enabled and the claim value does not match any project, the user is denied access to all projects. So it is a whitelist mechanism.

The only external interfaces of Flyte is the API that is being used by the Flyte console, the flytectl cli and pyflyte. So it is sufficient to add a check if the user interacting with the API for a certain project has the required permission.

Since the API endpoints are very different from one another, the check needs to be implemented in each endpoint. Because of the different formats of the endpoint holding the project it will be hard to define a middleware centrally, which performs the check.

The mapping between claim value and project to authorize can be configured in the following way and put in the Helm chart of the Flyte deployment:

```yaml
auth:
  project-authorization:
    enabled: true
    claim: "entitlements"
    mapping:
      project1: "project1"
      project2: "project2"
```

The identity token is evaluated by Flyte already and data can be accessed in the Flyte code. The following example golang code shows how the project isolation can be implemented in the Flyte code:


```go
package auth

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/auth/config"
)

func ProjectAccessPermitted(ctx context.Context, project string) bool {
	authConfig := config.GetConfig()

	if authConfig.ProjectAuth.isEnabled() {
		identityContext := IdentityContextFromContext(ctx)

		claimValues := identityContext.UserInfo().GetAdditionalClaims().GetFields()[authConfig.ProjectAuth.Claim].GetListValue().GetValues()

		for claimValue := range claimValues {
			if claimValue.GetStringValue() == authConfig.ProjectAuth.mapping[project] {
				return true
			}
		}
		
		return false
	} else {
		return true
	}

}
```

The following piece of code also shows how the check is integrated into existing API endpoints:

```go
func (m *AdminService) GetWorkflow(ctx context.Context, request *admin.ObjectGetRequest) (*admin.Workflow, error) {
	defer m.interceptPanic(ctx, request)

	if ProjectAccessPermitted(ctx, request.GetId().Project) {
		return nil, status.Errorf(codes.PermissionDenied, "Access denied for project %s.", request.GetId().Project)
	}

	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
```


## 4 Metrics & Dashboards

There are no metrics or dashboards needed for this feature.

## 5 Drawbacks

The feature needs to be thoughtfully be implemented and tested in order to avoid introducing security vulnerabilities which allows a user to access Flyte's project she is not meant to see information about.

## 6 Alternatives

In order to provide tenant isolation it is also possible to have multiple Flyte deployments. This is a more complex setup and requires more resources. It is also more difficult to manage. It will also only be feasible if the number of different projects is relatively small.

## 7 Potential Impact and Dependencies

Existing Flyte deployments will not be affected by this feature. The feature can be disabled by default to not impact existing Flyte deployments when updating to a Flyte version with the feature. To use this feature, Flyte needs to have authentication enabled. The feature will not work without authentication with OIDC.

## 8 Unresolved questions

This feature will not introduce an RBAC concept within Flyte's single projects. So there will not be different roles within a project. Any person with access to a project can do whatever she deserves without restrictions but has no permissions for a project that she is not assigned to.

## 9 Conclusion

With the introduction of project isolation, tenants can be efficiently be isolated from each other. This is a security feature that is essential for enterprise use cases and multi-tenant environments. It is a feature that is required to comply with data protection regulations. The feature is easy to configure and does not require any changes to the Flyte API. It is a feature that can be enabled or disabled by the Flyte operators. The feature will not impact existing Flyte deployments. The feature will not introduce an RBAC concept within Flyte's single projects. So there will not be different roles within a project. Any person with access to a project can do whatever she deserves without restrictions but has no permissions for a project that she is not assigned to.
