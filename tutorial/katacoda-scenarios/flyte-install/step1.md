flyte Components

<img src="https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/executions/flyte_wf_execution_overview.svg?sanitize=true" >

# Component Repos
Repo | Language | Purpose | Status
--- | --- | --- | ---
[flytepropeller](https://github.com/lyft/flytepropeller) | Go | execution engine | Production-grade
[flyteadmin](https://github.com/lyft/flyteadmin) | Go | control plane | Production-grade
[flyteconsole](https://github.com/lyft/flyteconsole) | Typescript | admin console | Production-grade

### flytepropeller

Kubernetes operator to executes Flyte graphs natively on kubernetes

### flyteadmin
Flyteadmin is the control plane for Flyte responsible for managing entities (task, workflows, launch plans) and
administering workflow executions. Flyteadmin implements the
`AdminService <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/admin.proto>`_ which
defines a stateless REST/gRPC service for interacting with registered Flyte entities and executions.
Flyteadmin uses a relational style Metadata Store abstracted by `GORM <http://gorm.io/>`_ ORM library.

### flyteconsole

This is the web UI for the Flyte platform.
