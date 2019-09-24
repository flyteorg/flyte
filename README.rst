Flyteadmin
=============

Flyteadmin is the control plane for Flyte responsible for managing entities (task, workflows, launch plans) and
administering workflow executions. Flyteadmin implements the
`AdminService <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/admin.proto>`_ which
defines a stateless REST/gRPC service for interacting with registered Flyte entities and executions.
Flyteadmin uses a relational style Metadata Store abstracted by `GORM <http://gorm.io/>`_ ORM library.
