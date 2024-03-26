(getting_started_index)=

# Introduction to Flyte

Flyte is a workflow orchestrator that unifies machine learning, data engineering, and data analytics stacks for building robust and reliable applications. Flyte features:
* Reproducible, repeatable workflows
* Strongly typed interfaces
* Structured datasets to enable easy conversion of dataframes between types, and column-level type checking
* Easy movement of data between local and cloud storage
* Easy tracking of data lineages
* Built-in data and artifact visualization

For a full list of features, see the [Flyte features page](https://flyte.org/features).

## Basic Flyte components

Flyte is made up of a user plane, control plane, and data plane.
* The **user plane** contains the elements you need to develop the code that will implement your application's directed acyclic graph (DAG). These elements are FlyteKit and Flytectl. Data scientists and machine learning engineers primarily work in the user plane.
* The **control plane** is part of the Flyte backend that is configured by platform engineers or others tasked with setting up computing infrastructure. It consists of FlyteConsole and FlyteAdmin, which serves as the main Flyte API to process requests from clients in the user plane. The control plane sends workflow execution requests to the data plane for execution, and stores information such as current and past running workflows, and provides that information upon request.
* The **data plane** is another part of the Flyte backend that contains FlytePropeller, the core engine of Flyte that executes workflows. FlytePropeller is designed as a [Kubernetes Controller](https://kubernetes.io/docs/concepts/architecture/controller/). The data plane sends status events back to the control plane so that information can be stored and surfaced to end users.

## Next steps

* To quickly try out Flyte on your machine, follow the {ref}`Quickstart guide <getting_started_quickstart_guide>`.
* To create a Flyte project that can be used to package workflow code for deployment to a Flyte cluster, see {ref}`"Getting started with workflow development" <getting_started_workflow_development>`.
* To set up a Flyte cluster, see the [Deployment documentation](https://docs.flyte.org/en/latest/deployment/index.html).
