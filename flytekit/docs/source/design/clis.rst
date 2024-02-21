.. _design-clis:

###################################
Command Line Interfaces and Clients
###################################

.. tags:: CLI, Basic

Flytekit currently ships with two CLIs, both of which rely on the same client implementation code.

*******
Clients
*******
The client code is located in ``flytekit/clients`` and there are two.

* Similar to the :ref:`design-models` files, but a bit more complex, the ``raw`` one is basically a wrapper around the protobuf generated code, with some handling for authentication in place, and acts as a mechanism for autocompletion and comments.
* The ``friendly`` client uses the ``raw`` client, adds handling of things like pagination, and is structurally more aligned with the functionality and call pattern of the CLI itself.

:py:class:`clients.friendly.SynchronousFlyteClient`

:py:class:`clients.raw.RawSynchronousFlyteClient`

***********************
Command Line Interfaces
***********************

Flytectl
=========

`Flytectl <https://pypi.org/project/yt-flyte-playground-flytectl/>`__ is the general CLI to communicate with the Flyte control plane (FlyteAdmin). Think of this as the ``kubectl`` for Flyte.

Think of this as a network-aware (can talk to FlyteAdmin) but not code-aware (no need to have user code checked out) CLI. In the registration flow, this CLI is responsible for shipping the compiled Protobuf files off to FlyteAdmin.

Pyflyte
========

Unlike ``flytectl``, think of this CLI as code-aware, which is responsible for the serialization (compilation) step in the registration flow. It will parse through the user code, looking for tasks, workflows, and launch plans, and compile them to `protobuf files <https://github.com/flyteorg/flyteidl/blob/0b20c5c99f9e964370d4f4ca663990ed56a14c7c/protos/flyteidl/core/workflow_closure.proto#L11-L18>`__.

.. _pyflyte-run:

What is ``pyflyte run``?
========================

The ``pyflyte run`` command is a light-weight, convenience command that incorporates packaging, registering, and launching a workflow into a single command.

It is not a fully featured production scale mode of operation, because it is designed to be a quick and easy iteration tool to get started with Flyte or test small self-contained scripts. The caveat here is it operates on a single file, and this file will have to contain all the required Flyte entities. Let’s take an example so that you can understand it better.

Suppose you execute a script that defines 10 tasks and a workflow that calls only 2 out of the 10 tasks. The remaining 8 tasks don’t get registered at that point.

It is considered fast registration because when a script is executed using ``pyflyte run``, the script is bundled up and uploaded to FlyteAdmin. When the task is executed in the backend, this zipped file is extracted and used.

.. _pyflyte-register:

What is ``pyflyte register``?
=============================

``pyflyte register`` is a command that registers all the workflows present in the repository/directory using fast-registration. It is equivalent to using two commands (``pyflyte package`` and ``flytectl register``) to perform the same operation (registration). It compiles the Python code into protobuf objects and uploads the files directly to FlyteAdmin. In the process, the protobuf objects are not written to the local disk, making it difficult to introspect these objects since they are lost.

The ``pyflyte package`` command parses and compiles the user’s Python code into Flyte protobuf objects. These compiled objects are stored as protobuf files and are available locally after you run the ``pyflyte package``.

The ``flytectl register`` command then ships the protobuf objects over the network to the Flyte control plane. In the process, ``flytectl`` also allows you to set run-time attributes such as IAM roles, K8s service accounts, etc.

``pyflyte package + flytectl register`` produces a **portable** package (a .tgz file) of Flyte entities (compiled protobuf files that are stored on the local disk), which makes it easy to introspect the objects at a later time if required. You can use register this package with multiple Flyte backends. You can save this package, use it for an audit, register with different FlyteAdmins, etc.

Why should you use ``pyflyte register``?
========================================

The ``pyflyte register`` command bridges the gap between ``pyflyte package`` + ``flytectl register`` and ``pyflyte run`` commands. It offers the functionality of the ``pyflyte package`` (with smarter naming semantics and combining the network call into one step).

.. note ::

   You can't use ``pyflyte register`` if you are unaware of the run-time options yet (IAM role, service account, and so on).

Usage
=====

.. prompt:: bash $

  pyflyte register --image ghcr.io/flyteorg/flytecookbook:core-latest --image trainer=ghcr.io/flyteorg/flytecookbook:core-latest --image predictor=ghcr.io/flyteorg/flytecookbook:core-latest --raw-data-prefix s3://development-service-flyte/reltsts flyte_basics

In a broad way, ``pyflyte register`` is equivalent to ``pyflyte run`` minus launching workflows, with the exception that ``pyflyte run`` can only register a single workflow, whereas ``pyflyte register`` can register all workflows in a repository.

What is the difference between ``pyflyte package + flytectl register`` and ``pyflyte register``?
================================================================================================

``pyflyte package + flytectl register`` works well with multiple FlyteAdmins since it produces a portable package. You can also use it to run scripts in CI.

``pyflyte register`` works well in single FlyteAdmin use-cases and cases where you are iterating locally.

Should you use ``pyflyte run`` or ``pyflyte package + flytectl register``?
==========================================================================

Both the commands have their own place in a production Flyte setting.

``pyflyte run`` is useful when you are getting started with Flyte, testing small scripts, or iterating over local scripts.

``pyflyte package + flytectl register`` is useful when you wish to work with multiple FlyteAdmins, wherein you can package the script, compile it into protobuf objects, write it to local disk, and upload this zipped package to different FlyteAdmins.

.. note ::

   Neither ``pyflyte register`` nor ``pyflyte run`` commands work on Python namespace packages since both the tools traverse the filesystem to find the first folder that doesn't have an __init__.py file, which is interpreted as the root of the project. Both the commands use this root as the basis to name the Flyte entities.
