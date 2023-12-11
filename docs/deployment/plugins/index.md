(deployment-plugin-setup)=

# Plugin Setup

Flyte integrates with a wide variety of [data, ML and analytical tools](https://flyte.org/integrations)
Some of these plugins, such as Databricks, Kubeflow, and Ray integrations, require the Flyte cluster administrator to enable them.

This section of the *Deployment Guides* will cover how to configure your cluster
to use these plugins in your workflows written in `flytekit`.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {ref}`K8s Plugins <deployment-plugin-setup-k8s>`
  - Guide to setting up the K8s Operator Plugins.
* - {ref}`Web API Plugin <deployment-plugin-setup-webapi>`
  - Guide to setting up the Web API Plugins.
* - {ref}`AWS Plugins <deployment-plugin-setup-aws>`
  - Guide to setting up AWS-specific Plugins.
* - {ref}`GCP Plugins <deployment-plugin-setup-gcp>`
  - Guide to setting up GCP-specific Plugins.
```

```{toctree}
:maxdepth: 1
:name: Plugin Setup
:hidden:

k8s/index
aws/index
gcp/index
webapi/index
```