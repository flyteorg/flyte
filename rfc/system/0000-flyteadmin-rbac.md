# [RFC] FlyteAdmin RBAC

**Authors:**

- @sovietaced

## 1 Executive Summary

We propose support for role based access control in Flyte Admin with support for project and domain level isolation/filtering.

## 2 Motivation

Support for authorization in Flyte Admin is minimal and may be unsuitable for production deployments. Flyte Admin only 
supports blanket authorization which does not align with information security best practices like the  
[principle of least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege). Flyte's project and domain 
concepts seem like strong constructs for supporting multi-tenancy but there are no mechanisms in place to prevent users 
of one tenant from accidentally or maliciously maniuplating another tenant's executions, tasks, launch plans, etc.

At Stack AV we have solved this problem in our fork and are looking to contribute our work back after several requests 
from the community. 

## 3 Proposed Implementation

### High Level Design 

![High Level Design](/rfc/images/rbac-high-level-black.png)

This proposal introduces an authorization interceptor that will be executed after the authentication interceptor. The 
authorization interceptor will consult the auth or authorization config to understand how it should resolve roles from 
the identity context along with what authorization policies are tied to the user's roles. If the user has an authorization 
policy that permits the incoming request the project and domain level scope will be extracted from the authorization 
policy and injected into the authorization context for use in the service layer. The request will hit the service layer which 
will make requests to the repository layer. The repository layer will leverage some utility functions to 1) authorize 
whether resources can be created in a project/domain and 2) generate some SQL filters to be injected into queries that 
read/update/delete resources. 

Ultimately the authorization interceptor will provide RPC level authorization and the changes to the repository layer 
will provide resource level authorization within the target RPC. 

#### Authorization Config 
The authorization config will be used to configure the behavior of RBAC. The first thing it will need is a way to
resolve roles from the user's identity context. Ideally this should be flexible enough to resolve a role from different
elements of a token to support various use cases. 

```yaml
authorization:
  roleResolutionStrategies: # one or more can be configured
    - scopes # attempts to use token scopes as roles
    - claims # attempts to use token claim values as roles
    - userID # attempts to use the user ID as the role
    
  claimRoleResolver: # claim based role resolution requires additional configuration
    - key: groups # this is the key to look for in the token claims
      type: list # this declares the structure of the value. parse value as comma separated string
    - key: group # this is the key to look for in the token claims
      type: string # this declares the structure of the value. parse value as single string
```

The second part of the authorization will include exceptions. There are some RPCs which should not be authorized. 
```yaml
authorization:
  methodBypassPatterns: # ideally these should be enabled by default
    - "/grpc.health.v1.Health/.*"
    - "/flyteidl.service.AuthMetadataService/.*"
```

The last part of the authorization will declare the authorization policies. The authorization policies will include the 
name of the role and will include a list of permitting rules. Each rule includes a mandatory method pattern which is
regex statement that matches the target gRPC method. Each rules includes optional project and domain isolation/filtering 
fields. If these fields are not included it is implied that the rule applies to all projects and domains. 
```yaml
authorization:
  policies: # policies is a list instead of a map since viper map keys are case insensitive
    - name: admin
      rules: # list of rules
        - name: "allow all" # name / description of the rule
          methodPattern: ".*" # regex pattern to match on gRPC method
    - name: read-only
      rules:
        - name: "read everything"
          methodPattern: "Get.*|List.*"
    - name: mapping-team
      rules:
        - name: "r/w for the mapping project in dev only"
          methodPattern: ".*"
          project: mapping # access to only the mapping project
          domain: development # access to only the development domain within the mapping project
    - name: ci
      rules:
        - name: "r/w for every project in production"
          methodPattern: ".*"
          domain: production # you can wildcard project and declare domain level access
    - name: 0oahjhk34aUxGnWcZ0h7 # the names can even include things like okta app IDs
      rules:
        - name: "flyte propeller"
          methodPattern: ".*"
```

#### Authorization Interceptor 

The authorization interceptor logic will behave like so: 
1. The interceptor checks to see if the target RPC method is configured to be bypassed. Call the next handler if bypassed. 
2. The interceptor consults the role resolver to obtain a list of role names for the user. 
3. The interceptor iterates over the list of role names and searches for any authorization policies tied to the role names. If no matching authorization policies are found return Permission Denied.
4. The interceptor iterates over the matching authorization policies to see if they permit the target RPC. Each rule that matches is added to a list. If no matching rules are found return Permission Denied. 
5. The interceptor extracts the project and domain scope out of each of the permitting rules and sets authorized resource scope on the authorization context. 
6. The interceptor calls the next handler.

Below is a high level overview of what the authorization context would look like. 

```go
type ResourceScope struct {
    Project string
    Domain string
}

type AuthorizationContext struct {
    resourceScopes []ResourceScope
}

// TargetResourceScope represents the scope of an individual resource. Sometimes only project level or domain level scope
// is applicable to a resource. In such cases, a user's resource scope may need to be truncated to matcht he depth 
// of the target resource scope.
type TargetResourceScope = int

const (
    ProjectTargetResourceScope = 0
    DomainTargetResourceScope = 1
)
```

#### Authorization Utils + DB Layer

The final piece of the puzzle is what performs resource level authorization and filtering. Historically, I have found 
that the best (albeit challenging) way to do this is at the database layer for a few reasons: 
* Filtering resources at the database works natively with any pagination
* Filtering resources at the database is the most performant since the database engine is fast and ends up returning a smaller data set over the network
* Filtering resources at the database plays nicely with logic to handle records that aren't found and does not leak data about which resources may or may not be present.

My proposal is to have utility functions for two different workflows: 
1. Creating Resources
   * This is a dedicated utility function used in repository code to create resource since you cannot add a WHERE clause filter for records that don't exist yet :) 
   * ```go
     func (r *ExecutionRepo) Create(ctx context.Context, input models.Execution, executionTagModel []*models.ExecutionTag) error {
         if err := util.AuthorizeResourceCreation(ctx, input.Project, input.Domain); err != nil {
             return err
         }
     
         ...
     }
     ```
2. Reading, Updating, Deleting resources with util
   * This is a utility function used in repository code that translates project/domain level scope into gorm where clause operations that can be attached to existing gorm queries.
   * ```go 
     var (
        executionColumnNames = util.ResourceColumns{Project: "executions.execution_project", Domain: "executions.execution_domain"}
     )
     
     func (r *ExecutionRepo) Get(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
         authzFilter := util.GetAuthzFilter(ctx, DomainTargetResourceScope, executionColumnNames)
         var execution models.Execution
         tx := r.db.WithContext(ctx).Where(&models.Execution{....})
         
         if authzFilter != nil {
             cleanSession := tx.Session(&gorm.Session{NewDB: true})
             tx = tx.Where(cleanSession.Scopes(authzFilter.GetScopes()...))
         }
     
         tx = tx.Take(&execution)
     
         ...
     }

     ```

## 4 Metrics & Dashboards

Existing metrics should measure RPC response codes. 

## 5 Drawbacks

*Are there any reasons why we should not do this? Here we aim to evaluate risk and check ourselves.*

## 6 Alternatives

*What are other ways of achieving the same outcome?*

## 7 Potential Impact and Dependencies

*Here, we aim to be mindful of our environment and generate empathy towards others who may be impacted by our decisions.*

- *What other systems or teams are affected by this proposal?*
- *How could this be exploited by malicious attackers?*

## 8 Unresolved questions

*What parts of the proposal are still being defined or not covered by this proposal?*

## 9 Conclusion

*Here, we briefly outline why this is the right decision to make at this time and move forward!*
