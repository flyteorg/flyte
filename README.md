Flyteadmin
=============
[![Current Release](https://img.shields.io/github/release/lyft/flyteadmin.svg)](https://github.com/lyft/flyteadmin/releases/latest)
![Master](https://github.com/lyft/flyteadmin/workflows/Master/badge.svg)
[![GoDoc](https://godoc.org/github.com/lyft/flyteadmin?status.svg)](https://pkg.go.dev/mod/github.com/lyft/flyteadmin)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![CodeCoverage](https://img.shields.io/codecov/c/github/lyft/flyteadmin.svg)](https://codecov.io/gh/lyft/flyteadmin)
[![Go Report Card](https://goreportcard.com/badge/github.com/lyft/flyteadmin)](https://goreportcard.com/report/github.com/lyft/flyteadmin)
![Commit activity](https://img.shields.io/github/commit-activity/w/lyft/flyteadmin.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/lyft/flyteadmin/latest.svg?style=plastic)

Flyteadmin is the control plane for Flyte responsible for managing entities (task, workflows, launch plans) and
administering workflow executions. Flyteadmin implements the
`AdminService <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/admin.proto>`_ which
defines a stateless REST/gRPC service for interacting with registered Flyte entities and executions.
Flyteadmin uses a relational style Metadata Store abstracted by `GORM <http://gorm.io/>`_ ORM library.

Before Check-In
~~~~~~~~~~~~~~~

Flyte Admin has a few useful make targets for linting and testing. Please use these before checking in to help suss out
minor bugs and linting errors.

  .. code-block:: console

    make goimports

  .. code-block:: console

    make test_unit

  .. code-block:: console

    make lint
