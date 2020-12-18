.. _working_hosted_service:

Working with a Hosted Flyte Service
====================================

Flytekit provides a python SDK for authoring and executing workflows and tasks in python.
Flytekit comes with a simplistic local scheduler that executes code in a local environment.
But, to leverage the full power of Flyte, we recommend using a deployed backend of Flyte. Flyte can be run
on a kubernetes cluster - locally, in a cloud environment or on-prem.

Please refer to the `Installing Flyte <https://lyft.github.io/flyte/administrator/install/index.html>`_ for details on getting started with a Flyte installation.
This section walks through steps on deploying your local workflow to a distributed Flyte environment, with ``NO CODE CHANGES``.