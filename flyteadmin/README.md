FlyteAdmin
==========

[![Current Release](https://img.shields.io/github/release/flyteorg/flyteadmin.svg)](https://github.com/flyteorg/flyteadmin/releases/latest)
![Master](https://github.com/flyteorg/flyteadmin/workflows/Master/badge.svg)
[![GoDoc](https://godoc.org/github.com/flyteorg/flyteadmin?status.svg)](https://pkg.go.dev/mod/github.com/flyteorg/flyteadmin)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![CodeCoverage](https://img.shields.io/codecov/c/github/flyteorg/flyteadmin.svg)](https://codecov.io/gh/flyteorg/flyteadmin)
[![Go Report Card](https://goreportcard.com/badge/github.com/flyteorg/flyteadmin)](https://goreportcard.com/report/github.com/flyteorg/flyteadmin)
![Commit activity](https://img.shields.io/github/commit-activity/w/flyteorg/flyteadmin.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/flyteorg/flyteadmin/latest.svg?style=plastic)
[![Slack](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://slack.flyte.org)

FlyteAdmin is the control plane for Flyte responsible for managing entities (task, workflows, launch plans) and
administering workflow executions. FlyteAdmin implements the
[AdminService](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/service/admin.proto) which
defines a stateless REST/gRPC service for interacting with registered Flyte entities and executions.
FlyteAdmin uses a relational style Metadata Store abstracted by [GORM](http://gorm.io/) ORM library.

For more background on Flyte, check out the official [website](https://flyte.org/) and the [docs](https://docs.flyte.org/en/latest/index.html)

Before Check-In
---------------

Flyte Admin has a few useful make targets for linting and testing. Please use these before checking in to help suss out
minor bugs and linting errors.

```
  # Please make sure you have all the dependencies installed:
  $ make install
  
  # In case you are only missing goimports:
  $ go install golang.org/x/tools/cmd/goimports@latest
  $ make goimports
```

```
  $ make test_unit
```

```
  $ make lint
```
