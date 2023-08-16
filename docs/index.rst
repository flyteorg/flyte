*************************************************
Production-grade Data and ML Workflows Made Easy
*************************************************

.. raw:: html

   <p style="color: #808080; font-weight: 350; font-size: 25px; padding-top: 10px; padding-bottom: 10px;">
   Highly scalable and flexible workflow orchestration for prototyping and production
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

   <br />

`Tags <_tags/tagsindex.html>`__

`Flyte <https://github.com/flyteorg/flyte>`__ is an open-source, Kubernetes-native
workflow orchestrator implemented in `Go <https://go.dev/>`__. It enables highly
concurrent, scalable and reproducible workflows for data processing, machine
learning and analytics.

Created at `Lyft <https://www.lyft.com/>`__ in collaboration with Spotify,
Freenome, and many others, Flyte provides first-class support for
`Python <https://docs.flyte.org/projects/flytekit/en/latest/>`__,
`Java, and Scala <https://github.com/flyteorg/flytekit-java>`__. Data Scientists
and ML Engineers in the industry use Flyte to create:

- ETL pipelines for petabyte-scale data processing.
- Analytics workflows for business and finance use cases.
- Machine learning pipelines for logistics, image processsing, and cancer diagnostics.

Explore Flyte
=============

Get a birds-eye view 🦅 of Flyte at the `official website <https://flyte.org/>`__:

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: https://flyte.org/features
       :type: url
       :text: ⭐️ Core features
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    From strongly typed interfaces to container-native DAGs, Flyte mitigates the
    trade-off between scalability and usability.

    ---

    .. link-button:: https://flyte.org/integrations
       :type: url
       :text: 🤝 Integrations
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    From strongly typed interfaces to container-native DAGs, Flyte mitigates the
    trade-off between scalability and usability.

    ---

    .. link-button:: https://flyte.org/airflow-alternative
       :type: url
       :text: 💨 Flyte vs Airflow
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Say goodbye to Airflow versioning pain and stepping over your teammate's toes
    when you change your package versions. Ouch!

    ---

    .. link-button:: https://flyte.org/kubeflow-alternative
       :type: url
       :text: 🔁 Flyte vs Kubeflow
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Unintuitive Python DSL boilerplate got you down? With ``flytekit`` you just
    write Python code and Flyte compiles down to type-safe execution graphs.

    ---

    .. link-button:: https://flyte.org/blog
       :type: url
       :text: 📝 Blog
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Learn more about orchestration, Flyte, and everything in between.

    ---

    .. link-button:: https://flyte.org/events
       :type: url
       :text: 🗓 Events
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Keep up-to-date with Flyte's upcoming talks, conferences, and more.


Learn Flyte
===========

The following main sections in the documentation will guide you through your
Flyte journey, whether you want to write Flyte workflows, deploy the Flyte
platform to your K8s cluster, or extend and contribute its architecture and
design.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: cookbook:getting_started_index
       :type: ref
       :text: 🔤 Getting Started
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Get your first workflow running, learn about the Flyte development lifecycle,
    and see the core use cases that Flyte enables.

    ---

    .. link-button:: cookbook:userguide
       :type: ref
       :text: 📖 User Guide
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    A comprehensive view of Flyte's functionality for data scientists, ML engineers,
    data engineers, and data analysts.

    ---

    .. link-button:: cookbook:tutorials
       :type: ref
       :text: 📚 Tutorials
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    End-to-end examples of Flyte for data/feature engineering, machine learning,
    bioinformatics, and more.

    ---

    .. link-button:: cookbook:integrations
       :type: ref
       :text: 🤝 Integrations
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Learn how to leverage a rich ecosystem of third-party tools and libraries
    to make your Flyte workflows even more effective.

    ---

    .. link-button:: deployment
       :type: ref
       :text: 🚀 Deployment Guide
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guides for platform engineers to deploy and maintain a Flyte cluster on your
    own infrastructure.

    ---

    .. link-button:: reference
       :type: ref
       :text: 📒 API Reference
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Reference for all of Flyte's component libraries.

    ---

    .. link-button:: divedeep
       :type: ref
       :text: 🧠 Concepts
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Dive deep into all of Flyte's concepts, from tasks and workflows to the underlying Flyte scheduler.

    ---

    .. link-button:: community
       :type: ref
       :text: 🤗 Community
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Join the fast-growing Flyte community to get help, ask questions, and contribute!

Deploy Flyte
============

The *Deployment Guides* are primarily for platform and devops engineers to learn how to deploy and administer Flyte.

The sections below walk through how to create a Flyte cluster and cover topics related to enabling and configuring
plugins, authentication, performance tuning, and maintaining Flyte as a production-grade service.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: deployment-deployment
       :type: ref
       :text: 🛣 Deployment Paths
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Walkthroughs for deploying Flyte, from the most basic to a fully-featured, multi-cluster production system.

    ---

    .. link-button:: deployment-plugin-setup
       :type: ref
       :text: 🔌 Plugin Setup
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Enable backend plugins to extend Flyte's capabilities, such as hooks for K8s, AWS, GCP, and Web API services.

    ---

    .. link-button:: deployment-configuration
       :type: ref
       :text: 🎛 Cluster Configuration
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    How to configure the various components of your cluster.

    ---

    .. link-button:: deployment-configuration-generated
       :type: ref
       :text: 📖 Configuration Reference
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Reference docs for configuration settings for Flyte's backend services.

    ---

    .. link-button:: deployment-security-overview
       :type: ref
       :text: 🔒 Security Overview
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Read for comments on security in Flyte.


Get Help
========

Have questions or need support? The best way to reach us is through Slack:

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: https://flyte-org.slack.com/archives/CP2HDHKE1
       :type: url
       :text: 🤔 Ask the Community
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Ask anything related to Flyte and get a response within a few hours.

    ---

    .. link-button:: https://flyte-org.slack.com/archives/C01RXBFV1M5
       :type: url
       :text: 👋 Introduce yourself
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Tell us about yourself. We'd love to know about you and what brings you to Flyte.

    ---

    .. link-button:: https://flyte-org.slack.com/archives/CPQ3ZFQ84
       :type: url
       :text: 💭 Share ideas
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Share any suggestions or feedback you have on how to make Flyte better.

    ---

    .. link-button:: https://flyte-org.slack.com/archives/C01P3B761A6
       :type: url
       :text: 🛠 Get help with deploment
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    If you need any help with Flyte deployment, hit us up.


.. toctree::
   :maxdepth: 1
   :caption: Overview
   :name: overview
   :hidden:

   Introduction <introduction>
   Flyte Fundamentals <_repos/flytesnacks/docs/getting_started/flyte_fundamentals>
   Core Use Cases <_repos/flytesnacks/docs/getting_started/core_use_cases>

.. toctree::
   :maxdepth: 1
   :caption: Examples
   :name: examples
   :hidden:

   User Guide <userguide>
   Tutorials <tutorials>
   Integrations <integrations>

.. toctree::
   :caption: Deployment
   :maxdepth: -1
   :name: deploymenttoc
   :hidden:

   deployment/deployment/index
   deployment/plugins/index
   deployment/configuration/index
   deployment/configuration/generated/index
   deployment/security/index
   reference/swagger

.. toctree::
   :maxdepth: 1
   :caption: API Reference
   :name: apitoc
   :hidden:

   flytekit <https://flytekit.readthedocs.io>
   flytekit-java <https://github.com/spotify/flytekit-java>
   flytectl <https://flytectl.readthedocs.io>
   flyteidl <https://flyteidl.readthedocs.io>
   flyteadmin <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/service/service.html>

.. toctree::
   :caption: Community
   :maxdepth: -1
   :name: roadmaptoc
   :hidden:

   Community <community/index>
   community/contribute
   community/roadmap
   Frequently Asked Questions <https://github.com/flyteorg/flyte/discussions/categories/q-a>
   community/troubleshoot

.. toctree::
   :caption: Glossary
   :maxdepth: -1
   :name: divedeeptoc
   :hidden:

   concepts/basics
   concepts/control_plane
   concepts/architecture
