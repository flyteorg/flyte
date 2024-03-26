# Sane overridable defaults
## Background
Flyte comes with a rich set of overridable defaults ([matchable resources](https://docs.flyte.org/en/latest/deployment/configuration/general.html#deployment-configuration-general)) from everything to default task resource requests, project resource quota, kubernetes labels and annotations for executions and default service accounts.

These can be configured to have default levels across a Flyte installation, on a per-domain basis, on a per-project basis and even at a per-workflow basis.

Managing these requires setting the defaults at a level of specificity using a variety of [message types](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto#L14,L38). Furthermore it's difficult to holistically reason about what overrides exist at say, the project level. Updating these overrides is error-prone because there is no existing merge logic and requires understanding the data model to even know what specific matchable resource to specifically update.

Importantly, the current mechanism for applying application overrides is driven imperatively using flytectl and hard to reason about across revisions or manage using [Infrastructure as Code](https://learn.microsoft.com/en-us/devops/deliver/what-is-infrastructure-as-code).

## Proposal
We should be able to make it simple to view, edit and update defaults for the entire platform, domains, projects and workflows. It should be difficult to unintentionally overwrite changes and easy to understand what is currently applied.

## Overview

### Proposed User Experience
It should be simple to view overridable defaults

```
flytectl get overrides -o overrides.yaml
...

flytectl get overrides -d development -p flytesnacks -o overrides.yaml
...

```
For
* global/per-domain defaults (immutable application settings)
* project defaults
* per-workflow defaults


#### Proposed document structure
A single call to get all overrides should return the collated document for all overrides set at a level (e.g. per project)

```
$ flytectl get overrides -p flytesnacks -d development --workflow MyExampleWorkflow

apiVersion: 1  # Version of this document structure
domain: development
project: flytesnacks
workflow: MyExampleWorkflow
kind: overrides
spec:
  taskResources:
    defaults:
      cpu: "1"
      memory: "500Mi"
    limits:
      cpu: "3"
      memory: "2Gi"
      gpu: "4"
  workflowExecutionConfig:
    inherited: true
    source: project
    labels:
      my-key: "my-value"
    securityContext:
      coreIdentity:
        iamRole: "my_iam_role"
status:
  version: 1 # Version for the entire overrides document
```

Note a few details
* The spec is the yaml formatted list of all available [matchable resource messages](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto) defined
* Specific matchable resources can be inherited, that is, not set at the workflow level used to get the document. In this example, the workflow execution config is set at the project level
* The status includes a version. This is managed on the backend.


### Implementation Proposal
Define a new set of APIs for returning overridable defaults


#### Read Experience
```
rpc GetOverrides(GetOverridesRequest) returns (GetOverridesResponse)

...

message OverridesDocument {
  string domain = 1;
  
  string project = 2; // optional
  
  string workflow = 3; // optional
  
  int32 api_version = 4;
  
  repeated ResourceOverrides overrides = 5;
  
  OverridesStatus status = 6;
}

message ResourceOverrides {
  MatchingAttributes attributes = 1; # Oneof for existing matchable resources
  
  bool inherited = 2;
  
  string domain = 3;  // optional, if inherited
  
  string project = 4;  // optional, if inherited
  
  ...
}

message OverridesStatus {
  int32 version = 1;
}

```

##### Write Experience

Traditionally overrides are managed programmatically through `flytectl` however declarative, [GitOps](https://about.gitlab.com/topics/gitops/) style management can be very useful for managing infrastructure as code. To that end, this proposal includes using a static config map for managing configurable overrides as the source of truth.

A migration strategy defined below outlines how existing users of overridable resources who have built up overrides in the database using flytectl can cut-over to the declarative style. FlyteAdmin deployments can be made configurable, so that overrides are either read from the existing database (for backwards compatibility with the current implementation) or a configmap.

In FlyteAdmin config

```
flyteadmin:
  overrides:
    # Points to deprecated matchable overrides written by flytectl
    sourceType: db 
```
OR 

In FlyteAdmin config

```
flyteadmin:
  overrides:
    sourceType: configmap
    sourcePath: /etc/flyte/overrides/

```

And the corresponding ConfigMap which gets mounted as a volume mount in the FlyteAdmin deployment
```
kind: ConfigMap
metadata:
  name: defaults
  namespace: flyte
  apiVersion: v1
  data:
    defaults.yaml: |
      # MyExampleWorkflow overrides
      - project: flytesnacks
        domain: development
        workflow: MyExampleWorkflow
        overrides:
          taskResources:
            defaults:
              cpu: "1" 
              memory: "500Mi"
            limits:
              cpu: "3" 
              memory: "2Gi"
              gpu: "4" 
            ...    
      # Flytesnacks/development level overrides    
      - project: flytesnacks
        domain: development
        overrides:
          workflowExecutionConfig:
            labels:
              my-key: "my-value"
            securityContext:
              coreIdentity:
                iamRole: "my_iam_role"
        ...                             

```
This can work the same way cluster resource templates are mounted as ConfigMaps in the FlyteAdmin deployment but are [fully customizable](https://github.com/flyteorg/flyte/blob/5ab8cb887a5a7051070cb93fca603ed1f22f2f74/charts/flyte-core/values.yaml#L752-L810).

This does require restarting the FlyteAdmin deployment for individual overrides changes but allows for declaratively defining the complete set of overrides in a reproducible manner. See more discussion on the trade-offs [here](https://github.com/flyteorg/flyte/discussions/3731#discussioncomment-6053743).

#### FlyteAdmin changes

##### Deprecated DB: Get Overrides
Getting overrides at a super specific level, for example at the workflow level, should use the same fallback logic currently applied at execution time to determine all applicable overrides across all matchable resource types (task resources, workflow execution config, etc).

That is, the workflow-level document should include all applicable workflow overrides, any project overrides and any domain overrides in one document with inherited overrides marked as such.

##### Single ConfigMap
When the FlyteAdmin application starts, it can load the overridable defaults configmap and create an in-memory mapping of per-project, per-domain and per-workflow overrides.

At execution time when overrides need to be applied or when a reader calls the `GetOverrides` API directly, the mapping can be queried for the applicable per-project, per-domain and per-workflow overrides, much like the logic currently applied at execution time.

##### Updating overrides
When FlyteAdmin is configured to use the database as the source of overrides, the existing flytectl CLI methods can be used to set matchable overrides from the command line.

When FlyteAdmin is configured to use the configmap as the source, any attempts to use the MatchableResource write API should fail (that is calls to `Update*Attributes` and `Delete*Attributes` endpoints)

### Migration Strategy

Eventually the single, entire document stored in the ConfigMap should be the source of truth instead of the current resources table in the database.

This allows us to do a finite number of in-memory queries at execution time to fetch the entire list of realized overrides for the global, domain, project and workflow level for all overridable attributes

Because this RFC proposes one source of truth for overrides, either the deprecated `db` or `configmap` types, Flyteadmin could ship with an additional endpoint to spit out a complete ConfigMap that captures the full set of all overrides

```
rpc GetAllOverrides(GetAllOverridesRequest) returns (GetAllOverridesResponse)
```

When overrides are stored in the current (to be deprecated) database, Flyte administrators can call the above endpoint to return the fully fleshed out ConfigMap they can use to customize their FlyteAdmin deployment
