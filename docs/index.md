# Welcome to Flyte!

```{eval-rst}
.. raw:: html

   <p style="color: #808080; font-weight: 350; font-size: 25px; padding-top: 10px; padding-bottom: 10px;">
   The highly scalable and flexible workflow orchestrator that unifies data, ML and analytics.
   </p>

.. image:: https://img.shields.io/badge/Graduate%20Project-Linux%20Foundation-purple?style=for-the-badge
    :target: https://lfaidata.foundation/projects/flyte/
    :alt: Linux Foundation

.. image:: https://img.shields.io/github/stars/flyteorg/flyte?label=github&logo=github&style=for-the-badge
   :target: https://github.com/flyteorg/flyte
   :alt: GitHub Repo stars

.. image:: https://img.shields.io/github/release/flyteorg/flyte.svg?style=for-the-badge&color=blue
   :target: https://github.com/flyteorg/flyte/releases/latest
   :alt: Flyte Release

.. image:: https://img.shields.io/github/actions/workflow/status/flyteorg/flyte/tests.yml?label=tests&style=for-the-badge
   :target: https://github.com/flyteorg/flyte/actions/workflows/tests.yml
   :alt: GitHub Test Status

.. image:: https://img.shields.io/github/actions/workflow/status/flyteorg/flyte/sandbox.yml?label=Sandbox%20docker%20image&style=for-the-badge
   :target: https://github.com/flyteorg/flyte/actions/workflows/sandbox.yml
   :alt: GitHub Sandbox Status

.. image:: https://img.shields.io/github/milestones/closed/flyteorg/flyte?style=for-the-badge
    :target: https://github.com/flyteorg/flyte/milestones?state=closed
    :alt: Completed Milestones

.. image:: https://img.shields.io/pypi/dm/flytekit?color=blue&label=flytekit%20downloads&style=for-the-badge&logo=pypi&logoColor=white
   :target: https://github.com/flyteorg/flytekit
   :alt: Flytekit Downloads

.. image:: https://img.shields.io/badge/Slack-Chat-pink?style=for-the-badge&logo=slack
    :target: https://slack.flyte.org
    :alt: Flyte Slack

.. image:: https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg?style=for-the-badge
    :target: http://www.apache.org/licenses/LICENSE-2.0.html
    :alt: License

.. |br| raw:: html

   <br>
   <br>

```

[Flyte](https://github.com/flyteorg/flyte) is an open-source, Kubernetes-native
workflow orchestrator implemented in [Go](https://go.dev/). It enables highly
concurrent, scalable and reproducible workflows for data processing, machine
learning and analytics.

Flyte provides first-class support for
{doc}`Python <api_reference/flytekit/index>` and has a community-driven [Java and Scala SDK](https://github.com/flyteorg/flytekit-java). 
Data Scientists and ML Engineers in the industry use Flyte to create:

- Data pipelines for processing petabyte-scale data.
- Analytics workflows for business and finance use cases.
- Machine learning pipelines for logistics, image processing, and cancer diagnostics.

## Learn Flyte

The following guides will take you through Flyte, whether you want to write
workflows, deploy the Flyte platform to your K8s cluster, or extend and
contribute its architecture and design. You can also access the
{ref}`docs pages by tag <tagoverview>`.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`📖 User Guide <user_guide/index>`
  - Get your first workflow running, learn about the Flyte development lifecycle, and get a comprehensive view of Flyte's functionality for data and ML practitioners.
* - {doc}`📚 Tutorials <flytesnacks/tutorials>`
  - End-to-end examples of Flyte for data/feature engineering, machine learning,
    bioinformatics, and more.
* - {doc}`🔌 Integrations <integrations/index>`
  - Leverage a rich ecosystem of third-party tools and libraries to make your
    Flyte workflows even more effective.
* - {ref}`🚀 Deployment Guide <deployment>`
  - Guides for platform engineers to deploy and maintain a Flyte cluster on your
    own infrastructure.
* - {ref}`🧠 Concepts <divedeep>`
  - Dive deep into all of Flyte's concepts, from tasks and workflows to the underlying Flyte scheduler.
```

## API Reference

Below are the API reference to the different components of Flyte:

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Flytekit <api_reference/flytekit/index>`
  - Flyte's official Python SDK.
* - {doc}`FlyteCTL <api_reference/flytectl/index>`
  - Flyte's command-line interface for interacting with a Flyte cluster.
* - {doc}`FlyteIDL <api_reference/flyteidl/index>`
  - Flyte's core specification language.
```

## Get Help

Have questions or need support? The best way to reach us is through Slack:

```{list-table}
:header-rows: 0
:widths: 20 30

* - {ref}`🗓️ Resources <community>`
  - Find resources for office hours, newsletter, and slack.
* - [🤔 Ask the Community](https://flyte-org.slack.com/archives/CP2HDHKE1)
  - Ask anything related to Flyte and get a response within a few hours.
* - [👋 Introduce yourself](https://flyte-org.slack.com/archives/C01RXBFV1M5)
  - Tell us about yourself. We'd love to know about you and what brings you to Flyte.
* - [💭 Share ideas](https://flyte-org.slack.com/archives/CPQ3ZFQ84>)
  - Share any suggestions or feedback you have on how to make Flyte better.
* - [🛠 Get help with deploment](https://flyte-org.slack.com/archives/C01P3B761A6>)
  - If you need any help with Flyte deployment, hit us up.
```


```{toctree}
:maxdepth: 1
:hidden:

user_guide/index
concepts/index
tutorials/index
integrations/index
cluster_deployment/index
api_reference/index
ecosystem/index
community/index
```