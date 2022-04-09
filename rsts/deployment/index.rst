.. _deployment:

#############
Deployment
#############

******************
Cluster Deployment
******************

The guides in this section are geared towards those who have found the limits of the small scale single-container
version of Flyte, and are looking to deploy it as a platform on a cloud provider (or an equivalent on-premise Kubernetes
solution). The following pages will help you effectively deploy and manage an enterprise-ready Flyte platform.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: deployment-overview
       :type: ref
       :text: Overview
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    A high-level look into the Flyte components that we'll need to move around.

    ---

    .. link-button:: sandbox
       :type: ref
       :text: Sandbox Deployment
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    A stand-alone minimal environment for running a Flyte backend.

    ---

    .. link-button:: deployment-aws
       :type: ref
       :text: AWS
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Deployment guides with detailed instructions specific to AWS.

    ---

    .. link-button:: deployment-gcp
       :type: ref
       :text: GCP
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Deployment guides with detailed instructions specific to GCP.

    ---

    .. link-button:: deployment-sandbox
       :type: ref
       :text: Sandbox
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Deployment guides with detailed instructions for sandbox deployment.

    ---

    .. link-button:: multicluster
       :type: ref
       :text: Multi-cluster setup
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Detailed instructions on multiple clusters' setup.

    ---

    .. link-button:: deployment-plugin-setup
       :type: ref
       :text: Plugin Setup
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    How to set up a plugin for your deployment.

    ---
    .. link-button:: deployment-cluster-config
       :type: ref
       :text: Cluster Configuration
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Flyte comes with a lot of things you can configure. These pages will walk you through the various components.

    ---
    .. link-button:: ideal-flow
       :type: ref
       :text: How to Streamline Your Flyte Workflows
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    An example use case leveraging DevOps practices to run Flyte pipelines.


.. toctree::
    :maxdepth: 1
    :name: deployment breakdown toc
    :hidden:

    overview
    aws/index
    gcp/index
    sandbox/index
    multicluster
    cluster_config/index
    sandbox
    plugin_setup/index
    security/security
    ideal_flow
