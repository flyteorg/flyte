# Flytekit Cookbook

This is a collection of short "how to" articles demonstrating the various capabilities and idiomatic usage of Flytekit.
Note this currently does not include articles/tid-bits on how to use the Flyte platform at large, though in the future we may expand it to that.

## Contents   
1. [Write and Execute a Task](recipes/task)
1. [Create a workflow from Tasks](recipes/workflows)
1. [Launch Plans](recipes/launchplans)
1. [Multiple Schedules for a Workflow](recipes/multi_schedules)
1. [Executing pre-created tasks & workflows](recipes/fetch)
1. [Interact with Past Workflow / Task Executions](recipes/interaction)
1. [Working with Types](recipes/types)
1. [Composing a Workflow from shared tasks and workflows](recipes/shared)
1. [Map/Array Tasks](recipes/map_tasks)
1. [Compose a Workflow from other workflows](recipes/compose)
1. [Dynamically Generate a Workflow at Runtime](recipes/dynamic_wfs)
1. [(WIP)Dynamic Tasks](recipes/dynamictasks)
1. [Tasks without flytekit or Using arbitrary containers](recipes/rawcontainers)
1. [(WIP)Using Papermill & Jupyter notebook to author tasks](recipes/papermill)
1. [(WIP)Different container per task](recipes/differentcontainers) 

Each example is organized into a separate folder at this layer, and each has the Readme file linked to in the Contents, as well as supporting .py and .ipynb files, some of which contain information in a lot more depth.

## Setup

Careful care has been taken to ensure that all the .py files contained in this guide are fully usable and compilable into Flyte entities. Readers that want to run these workflows for themselves should

1. Follow instructions in the main Flyte documentation to set up a local Flyte cluster. All the commands in this book assume that you are using Docker Desktop. If you are using minikube or another K8s deployment, the commands will need to be modified.
1. Please also ensure that you have a Python virtual environment installed and activated, and have pip installed the requirements.
1. Use flyte-cli to create a project named `flytesnacks` (the name the Makefile is set up to use).

## Using the Cookbook

The examples written in this cookbook are meant to used in two ways, as a reference to read, and a live project to iterate and test on your local (or EKS/GCP/etc K8s cluster). To make some simple changes to the code in this cookbook to just try things out, or to just see it run on your local cluster, the iteration cycle should be

1. Make your changes and commit (the steps below require a clean git tree).
1. Activate your Python virtual environment with the requirements installed.
1. `make docker_build` This will build the image tagged with just `flytecookbook:<sha>`, no registry will be prefixed.
1. `make register_sandbox` This will register the Flyte entities in this cookbook against your local installation.

If you are just iterating locally, there is no need to push your Docker image. For Docker for Desktop at least, locally built images will be available for use in its K8s cluster.

If you would like to later push your image to a registry (Dockerhub, ECR, etc.), you can

```bash
REGISTRY=docker.io/corp make docker_push
``` 
