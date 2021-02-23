.. _flyte-tutorials-firstrun:
.. currentmodule:: firstrun

############################################
Getting Started with Flyte
############################################

.. rubric:: Estimated time to complete: 2 minutes.

Flyte enables scalable, reproducable and reliable orchestration of massively large workflows. In order to get a sense of the product, we have packaged a minimalist version of the Flyte system into a Docker image.

With `docker installed <https://docs.docker.com/get-docker/>`__, run this command: ::

  docker run --rm --privileged -p 30081:30081 ghcr.io/flyteorg/flyte-sandbox

This creates a local Flyte sandbox. Once the sandbox is ready, you should see the following message: `Flyte is ready! Flyte UI is available at http://localhost:30081/console.`. Go ahead and visit that to check out the Flyte UI.

A quick visual tour for launching your first Workflow:

.. image:: https://github.com/flyteorg/flyte/raw/static-resources/img/first-run-console-2.gif
    :alt: A quick visual tour for launching your first Workflow.
