.. _deployment:

#############
Deployment
#############

These *Deployment Guides* are primarily for platform and devops engineers to learn how to deploy and administer Flyte.

The sections below walk through how to create a Flyte cluster and cover topics related to enabling and configuring
plugins, authentication, performance tuning, and maintaining Flyte as a production-grade service.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: deployment-deployment
       :type: ref
       :text: Deployment Paths
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Walkthroughs for deploying Flyte, from the most basic to a fully-featured, multi-cluster production system.

    ---

    .. link-button:: deployment-plugin-setup
       :type: ref
       :text: Plugin Setup
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Learn how to enable backend plugins to extend Flyte's capabilities, such as hooks for K8s, AWS, GCP, and Web API
    services.

    ---

    .. link-button:: deployment-configuration
       :type: ref
       :text: Cluster Configuration
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    How to configure the various components of your cluster.

    ---

    .. link-button:: deployment-security-overview
       :type: ref
       :text: Security Overview
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Read for comments on security in Flyte.


.. toctree::
    :maxdepth: 1
    :name: deployment guide toc
    :hidden:

    deployment/index
    plugins/index
    configuration/index
    security/index
