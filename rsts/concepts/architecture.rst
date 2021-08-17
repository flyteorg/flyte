.. _divedeep-architecture-overview:

######################
Component Architecture
######################

This document aims to demystify how Flyte's major components ``FlyteIDL``, ``FlyteKit``, ``FlyteCLI``, ``FlyteConsole``, ``FlyteAdmin``, ``FlytePropeller``, and ``FlytePlugins`` fit together at a high level.

FlyteIDL
========

In Flyte, entities like "Workflows", "Tasks", "Launch Plans", and "Schedules" are recognized by multiple system components. In order for components to communicate effectively, they need a shared understanding about the structure of these entities.

The Flyte IDL (Interface Definition Language) is where shared Flyte entities are defined. This IDL also defines the RPC service definition for the :std:ref:`core Flyte API <protos/docs/service/index:rest/and/grpc/interface/for/the/flyte/admin/service>`.

FlyteIDL uses the `protobuf <https://developers.google.com/protocol-buffers/>`_ schema to describe entities. Clients are generated for Python, Golang, and JavaScript and imported by Flyte components.


Planes
======

Flyte components are separated into 3 logical planes. The planes are summarized here and explained in further detail below. The goal is that any of these planes can be replaced by an alternate implementation.

+-------------------+---------------------------------------------------------------------------------------------------------------+
| **User Plane**    | The User Plane consists of all user tools that assist in interacting with the core Flyte API.                 |
|                   | These tools include the FlyteConsole, FlyteKit, and FlyteCLI.                                                 |
+-------------------+---------------------------------------------------------------------------------------------------------------+
| **Control Plane** | The Control Plane implements the core Flyte API.                                                              |
|                   | It serves all client requests coming from the User Plane.                                                     |
|                   | It stores information such as current and past running workflows, and provides that information upon request. |
|                   | It also accepts requests to execute workflows, but offloads the work to the Data Plane.                       |
+-------------------+---------------------------------------------------------------------------------------------------------------+
| **Data Plane**    | The sole responsibility of the the Data Plane is to fulfill workflows.                                        |
|                   | It accepts workflow requests from the Control Plane and guides the workflow to completion,                    |
|                   | launching tasks on a cluster of machines as necessary based on the workflow graph.                            |
|                   | It sends status events back to the control plane so the information can be stored and surfaced to end-users.  |
+-------------------+---------------------------------------------------------------------------------------------------------------+

User Plane
----------

In Flyte, workflows are represented as a Directed Acyclic Graph (DAG) of tasks. While this representation is logical for services, managing workflow DAGs in this format is a tedious exercise for humans. The Flyte User Plane provides tools to create, manage, and visualize workflows in a format that is easily digestible to users.

These tools include: 

FlyteKit
  FlyteKit is an SDK that helps users design new workflows using the Python programming language. FlyteKit can parse the python code, compile it into a valid Workflow DAG, and submit it to Flyte to be executed.

FlyteConsole
  Flyte console provides the Web interface for Flyte. Users and administrators can use the console to view workflows, launch plans, schedules, tasks, and individual task executions. The console provides tools to visualize workflows, and surfaces relevant logs for debugging failed tasks.

FlyteCLI
  Flyte Command Line Interface provides interactive access to Flyte to launch and access Flyte workflows via terminal.


Control Plane
-------------

The Control Plane supports the core REST/gRPC API defined in FlyteIDL. User Plane tools like FlyteConsole and FlyteKit contact the control plane on behalf of users to store and retrieve information. 

Currently, the entire control plane is handled by a single service called **FlyteAdmin**.

FlyteAdmin is stateless. It processes requests to create entities like Tasks, Workflows, and Schedules by persisting data in a relational database.

While FlyteAdmin serves the Workflow Exeuction API, it does not, itself, execute workflows. To launch workflow executions, FlyteAdmin sends the workflow DAG off to the DataPlane. For added scalability and fault-tolerance, FlyteAdmin can be configured to load-balance workflows across multiple isolated data-plane clusters.


Data Plane
----------

The Data Plane is the engine that accepts DAGs, and fulfills workflow executions by launching tasks in the order defined by the graph. Requests to the Data Plane generally come via the control plane, and not from end-users.

In order to support compute-intensive workflows at massive scale, the Data Plane needs to launch containers on a cluster of machines. The current implementation leverages `kubernetes <https://kubernetes.io/>`_ for cluster management.

Unlike the user-facing control-plane, the Data Plane does not expose a traditional REST/gRPC API. To launch an execution in the Data Plane, you create a “flyteworkflow” resource in kubernetes.
A “flyteworkflow” is a kubernetes `Custom Resource <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ (CRD) created by our team. This custom resource represents the flyte workflow DAG.

The core state machine that processes flyteworkflows is worker we call **FlytePropeller**.

FlytePropeller leverages the kubernetes `operator pattern <https://kubernetes.io/docs/concepts/extend-kubernetes/operator/>`_. It polls the kubernetes API, looking for newly created flyteworkflow resources. FlytePropeller understands the workflow DAG, and launches the appropriate kubernetes pods as needed to complete tasks. It periodically checks for completed tasks, launching downstream tasks until the workflow is complete.

**Plugins**

Each task in a flyteworkflow DAG has a specified **type**. The logic for fulfilling a task is determined by its task type.
In the most basic case, FlytePropeller launches a single kubernetes pod to fulfill a task.
More complex task types require workloads to be distributed across hundreds of pods.

The type-specific task logic is separated into isolated code modules that we call **plugins**.
Each task type has an associated plugin that is responsible for handling tasks of its type.
For each task in a workflow, FlytePropeller activates the appropriate plugin based on the task type in order to fullfill the task.

The Flyte team has pre-built plugins for Hive, Spark, and AWS Batch, and more.
To support new use-cases, developers can create their own plugins and bundle them in their FlytePropeller deployment.

Component Code References
=========================

.. panels::
  :container: container-lg pb-4
  :column: col-lg-4 col-md-4 col-sm-4 col-xs-12 p-2
  :body: text-center

  .. link-button:: https://pkg.go.dev/mod/github.com/flyteorg/flyteadmin
     :type: url
     :text: FlyteAdmin
     :classes: btn-block stretched-link

  ---

  .. link-button:: https://pkg.go.dev/mod/github.com/flyteorg/flytepropeller
     :type: url
     :text: FlytePropeller
     :classes: btn-block stretched-link

  ---

  .. link-button:: https://pkg.go.dev/mod/github.com/flyteorg/datacatalog
     :type: url
     :text: DataCatalog
     :classes: btn-block stretched-link

  ---

  .. link-button:: https://pkg.go.dev/mod/github.com/flyteorg/flyteplugins
     :type: url
     :text: FlytePlugins
     :classes: btn-block stretched-link
