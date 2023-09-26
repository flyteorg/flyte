.. _deployment-plugin-setup:

Plugin Setup
============

Flyte integrates with a wide variety of `data, ML and analytical tools <https://flyte.org/integrations>`__.
Some of these plugins, such as Databricks, Kubeflow, and Ray integrations, require the Flyte cluster administrator to enable them.

This section of the *Deployment Guides* will cover how to configure your cluster
to use these plugins in your workflows written in ``flytekit``.


.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: deployment-plugin-setup-k8s
       :type: ref
       :text: K8s Plugins
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up the K8s Operator Plugins.

    ---

    .. link-button:: deployment-plugin-setup-webapi
       :type: ref
       :text: Web API Plugin
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up the Web API Plugins.

    ---

    .. link-button:: deployment-plugin-setup-aws
       :type: ref
       :text: AWS Plugins
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up AWS-specific Plugins.

    ---

    .. link-button:: deployment-plugin-setup-gcp
       :type: ref
       :text: GCP Plugins
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up GCP-specific Plugins.


.. toctree::
    :maxdepth: 1
    :name: Plugin Setup
    :hidden:

    k8s/index
    aws/index
    gcp/index
    webapi/index
