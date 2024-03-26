Flyte Propeller
===============

[![Current Release](https://img.shields.io/github/release/flyteorg/flytepropeller.svg)](https://github.com/flyteorg/flytepropeller/releases/latest)
![Master](https://github.com/flyteorg/flytepropeller/workflows/Master/badge.svg)
[![GoDoc](https://godoc.org/github.com/flyteorg/flytepropeller?status.svg)](https://pkg.go.dev/mod/github.com/flyteorg/flytepropeller)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![CodeCoverage](https://img.shields.io/codecov/c/github/flyteorg/flytepropeller.svg)](https://codecov.io/gh/flyteorg/flytepropeller)
[![Go Report Card](https://goreportcard.com/badge/github.com/flyteorg/flytepropeller)](https://goreportcard.com/report/github.com/flyteorg/flytepropeller)
![Commit activity](https://img.shields.io/github/commit-activity/w/flyteorg/flytepropeller.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/flyteorg/flytepropeller/latest.svg?style=plastic)
[![Slack](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://slack.flyte.org)

Kubernetes operator to executes Flyte graphs natively on kubernetes

Components
==========

Propeller
---------
Propeller is a K8s native operator that executes Flyte workflows. Workflow Spec is written in Protobuf for
cross-compatibility.

Propeller Webhook
-----------------
A Mutating Webhook that can be optionally deployed to extend Flyte Propeller's functionality. It currently supports
enables injecting secrets into pods launched directly or indirectly through Flyte backend plugins.

kubectl-flyte
-------------
A Kubectl-plugin to interact with Flyte Workflow CRDs. It enables retrieving and rendering Flyte Workflows in CLI as
well as safely aborting running workflows.

Getting Started
===============
kubectl-flyte tool
------------------
kubectl-flyte is an command line tool that can be used as an extension to kubectl. It is a separate binary that is built
from the propeller repo.

Install
-------
This command will install kubectl-flyte and flytepropeller to `~/go/bin`

```
   $ make compile
```

You can also use [Krew](https://github.com/kubernetes-sigs/krew) to install the kubectl-flyte CLI:

```
   $ kubectl krew install flyte
```

Use
---
Two ways to execute the command, either standalone *kubectl-flyte* or as a subcommand of *kubectl*

```
    $ kubectl-flyte --help
    OR
    $ kubectl flyte --help
    Flyte is a serverless workflow processing platform built for native execution on K8s.
          It is extensible and flexible to allow adding new operators and comes with many operators built in

    Usage:
      kubectl-flyte [flags]
      kubectl-flyte [command]

    Available Commands:
      compile     Compile a workflow from core proto-buffer files and output a closure.
      config      Runs various config commands, look at the help of this command to get a list of available commands..
      create      Creates a new workflow from proto-buffer files.
      delete      delete a workflow
      get         Gets a single workflow or lists all workflows currently in execution
      help        Help about any command
      visualize   Get GraphViz dot-formatted output.
```

Observing running workflows
---------------------------

To retrieve all workflows in a namespace use the --namespace option, --namespace = "" implies all namespaces.

```
   $ kubectl-flyte get --namespace flytekit-development
    workflows
    ├── flytekit-development/flytekit-development-f01c74085110840b8827 [ExecId: ... ] (2m34s Succeeded) - Time SinceCreation(30h1m39.683602s)
    ...
    Found 19 workflows
    Success: 19, Failed: 0, Running: 0, Waiting: 0
```

To retrieve a specific workflow, namespace can either be provided in the format namespace/name or using the --namespace
argument

```
   $ kubectl-flyte get flytekit-development/flytekit-development-ff806e973581f4508bf1
    Workflow
    └── flytekit-development/flytekit-development-ff806e973581f4508bf1 [ExecId: project:"flytekit" domain:"development" name:"ff806e973581f4508bf1" ] (2m32s Succeeded )
        ├── start-node start 0s Succeeded
        ├── c task 0s Succeeded
        ├── b task 0s Succeeded
        ├── a task 0s Succeeded
        └── end-node end 0s Succeeded
```

Deleting workflows
------------------
To delete a specific workflow

```
   $ kubectl-flyte delete --namespace flytekit-development flytekit-development-ff806e973581f4508bf1
```

To delete all completed workflows - they have to be either success/failed with a special isCompleted label set on them.
The Label is set [here](https://github.com/flyteorg/flytepropeller/blob/master/pkg/controller/controller.go#L247)

```
   $ kubectl-flyte delete --namespace flytekit-development --all-completed
```

Running propeller locally
-------------------------
use the config.yaml in root found [here](https://github.com/flyteorg/flytepropeller/blob/master/config.yaml). Cd into
this folder and then run

```
   $ flytepropeller
```

Following dependencies need to be met

1. Blob store (you can forward minio port to localhost)
2. Admin Service endpoint (can be forwarded) OR *Disable* events to admin and launchplans
3. access to kubeconfig and kubeapi

Running webhook
---------------

API Server requires the webhook to serve traffic over SSL. To issue self-signed certs to be used for serving traffic,
use:

```
    $ flytepropeller webhook init-certs
```

This will create a ca.crt, tls.crt and key.crt and store them to flyte-pod-webhook secret. If a secret of the same name
already exist, it'll not override it.

Starting the webhook can be done by running:

```
    $ flytepropeller webhook
```

The secret should be mounted and accessible to this command. It'll then create a MutatingWebhookConfiguration object
with the details of the webhook and that registers the webhook with ApiServer.

Making changes to CRD
=====================
*Remember* changes to CRD should be carefully done, they should be backwards compatible or else you should use proper
operator versioning system. Once you do the changes, you have to follow the following steps.

- Ensure the propeller code is checked out in `$GOPATH/github.com/flyteorg/flytepropeller`
- Uncomment https://github.com/flyteorg/flytepropeller/blob/master/hack/tools.go#L5

```bash
    $ go mod vendor
```

- Now generate the code

```bash
    $ make op_code_generate
```

**Why do we have to do this?**
Flytepropeller uses old way of writing Custom controllers for K8s. The k8s.io/code-generator only works in the GOPATH
relative code path (sadly). So you have checkout the code in the right place. Also, `go mod vendor` is needed to get
code-generator in a discoverable path.

**TODO**

1. We may be able to avoid needing the old style go-path
2. Migrate to using controller runtime

