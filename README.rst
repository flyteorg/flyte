Flyte Propeller
===============


..
.. image:: https://img.shields.io/github/release/lyft/flytepropeller.svg
:target: https://github.com/lyft/flytepropeller/releases/latest

.. image:: https://godoc.org/github.com/lyft/flytepropeller?status.svg
:target: https://godoc.org/github.com/lyft/flytepropeller

.. image:: https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg
:target: http://www.apache.org/licenses/LICENSE-2.0.html

.. image:: https://img.shields.io/codecov/c/github/lyft/flytepropeller.svg
:target: https://codecov.io/gh/lyft/flytepropeller

.. image:: https://goreportcard.com/badge/github.com/lyft/flytepropeller
:target: https://goreportcard.com/report/github.com/lyft/flytepropeller

.. image:: https://img.shields.io/github/commit-activity/w/lyft/flytepropeller.svg?style=plastic

.. image:: https://img.shields.io/github/commits-since/lyft/flytepropeller/latest.svg?style=plastic


Kubernetes operator to executes Flyte graphs natively on kubernetes

Getting Started
===============
kubectl-flyte tool
------------------
kubectl-flyte is an command line tool that can be used as an extension to kubectl. It is a separate binary that is built from the propeller repo.

Install
-------
This command will install kubectl-flyte and flytepropeller to `~/go/bin`
.. code-block:: make

   $make compile

Use
---
Two ways to execute the command, either standalone *kubectl-flyte* or as a subcommand of *kubectl*

.. code-block:: command

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


Observing running workflows
---------------------------

To retrieve all workflows in a namespace use the --namespace option, --namespace = "" implies all namespaces.

.. code-block:: command

   $ kubectl-flyte get --namespace flytekit-development
    workflows
    ├── flytekit-development/flytekit-development-f01c74085110840b8827 [ExecId: ... ] (2m34s Succeeded) - Time SinceCreation(30h1m39.683602s)
    ...
    Found 19 workflows
    Success: 19, Failed: 0, Running: 0, Waiting: 0


To retrieve a specific workflow, namespace can either be provided in the format namespace/name or using the --namespace argument

.. code-block:: command

   $ kubectl-flyte get flytekit-development/flytekit-development-ff806e973581f4508bf1
    Workflow
    └── flytekit-development/flytekit-development-ff806e973581f4508bf1 [ExecId: project:"flytekit" domain:"development" name:"ff806e973581f4508bf1" ] (2m32s Succeeded )
        ├── start-node start 0s Succeeded
        ├── c task 0s Succeeded
        ├── b task 0s Succeeded
        ├── a task 0s Succeeded
        └── end-node end 0s Succeeded

Deleting workflows
------------------
To delete a specific workflow

.. code-block:: command

   $ kubectl-flyte delete --namespace flytekit-development flytekit-development-ff806e973581f4508bf1

To delete all completed workflows - they have to be either success/failed with a special isCompleted label set on them. The Label is set `here <https://github.com/lyft/flytepropeller/blob/master/pkg/controller/controller.go#L247>`

.. code-block:: command

   $ kubectl-flyte delete --namespace flytekit-development --all-completed

Running propeller locally
-------------------------
use the config.yaml in root found `here <https://github.com/lyft/flytepropeller/blob/master/config.yaml>`. Cd into this folder and then run

.. code-block:: command

   $ flytepropeller --logtostderr

Following dependencies need to be met
1. Blob store (you can forward minio port to localhost)
2. Admin Service endpoint (can be forwarded) OR *Disable* events to admin and launchplans
3. access to kubeconfig and kubeapi

Making changes to CRD
=====================
*Remember* changes to CRD should be carefully done, they should be backwards compatible or else you should use proper
operator versioning system. Once you do the changes, remember to execute

.. code-block:: make

    $make op_code_generate
