.. _userguide:

##############
User Guide
##############

If this is your first time using Flyte, check out the `Getting Started <https://docs.flyte.org/en/latest/getting_started.html>`_ guide.

This *User Guide*, the :doc:`Tutorials <tutorials>`, and the :doc:`Integrations <integrations>` examples cover all of
the key features of Flyte for data analytics, data science and machine learning practitioners, organized by topic. Each
section below introduces a core feature of Flyte and how you can use it to address specific use cases. Code for all
of the examples can be found in the `flytesnacks repo <https://github.com/flyteorg/flytesnacks>`_.

`Flytesnacks <https://github.com/flyteorg/flytesnacks>`_ comes with a specific environment to make running, documenting
and contributing samples easy. If this is your first time running these examples, follow the
:doc:`environment setup guide <userguide_setup>` to get started.

.. tip::

   To learn about how to spin up and manage a Flyte cluster in the cloud, see the
   :doc:`Deployment Guides <flyte:deployment/index>`.

******************
Table of Contents
******************

.. panels::
   :header: text-center
   :column: col-lg-12 p-2

   .. link-button:: userguide_setup
      :type: ref
      :text: 🌳 Environment Setup
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Setup your development environment to run the User Guide examples.

   ---

   .. link-button:: auto/core/flyte_basics/index
      :type: ref
      :text: 🔤 Flyte Basics
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Learn about tasks, workflows, launch plans, caching, and working with files and directories.

   ---

   .. link-button:: auto/core/control_flow/index
      :type: ref
      :text: 🚰 Control Flow
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Implement conditionals, nested and dynamic workflows, map tasks, and even recursion!

   ---

   .. link-button:: auto/core/type_system/index
      :type: ref
      :text: ⌨️ Type System
      :classes: btn-block stretched-link
   ^^^^^^^
   Improve pipeline robustness with Flyte's portable and extensible type system.

   ---

   .. link-button:: auto/core/scheduled_workflows/index
      :type: ref
      :text: ⏱ Scheduled Workflows
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Learn about scheduled workflows.

   ---

   .. link-button:: auto/testing/index
      :type: ref
      :text: ⚗️ Testing
      :classes: btn-block stretched-link
   ^^^^^^^
   Test tasks and workflows with Flyte's testing utilities.

   ---

   .. link-button:: auto/core/containerization/index
      :type: ref
      :text: 📦  Containerization
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^^^^^
   Easily manage the complexity of configuring the containers that run Flyte tasks.

   ---

   .. link-button:: auto/deployment/index
      :type: ref
      :text: 🚢  Production Config
      :classes: btn-block stretched-link
   ^^^^^^^^^^
   Ship and configure your machine learning pipelines on a production Flyte installation.

   ---

   .. link-button:: auto/remote_access/index
      :type: ref
      :text: 🎮 Remote Access
      :classes: btn-block stretched-link
   ^^^^^^^^^^
   Register, inspect, and monitor tasks and workflows on a Flyte backend.

   ---

   .. link-button:: auto/core/extend_flyte/index
      :type: ref
      :text: 🏗 Extending Flyte
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^^^^
   Define custom plugins that aren't currently supported in the Flyte ecosystem.

   ---

   .. link-button:: auto/larger_apps/index
      :type: ref
      :text: ⛰ Building Large Apps
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^^^^
   Build, deploy, and iterate on large projects by organizing your Flyte app.

.. toctree::
   :maxdepth: 1
   :hidden:

   |plane| Getting Started <https://docs.flyte.org/en/latest/getting_started.html>
   |book-reader| User Guide <index>
   |chalkboard| Tutorials <tutorials>
   |project-diagram| Concepts <https://docs.flyte.org/en/latest/concepts/basics.html>
   |rocket| Deployment <https://docs.flyte.org/en/latest/deployment/index.html>
   |book| API Reference <https://docs.flyte.org/en/latest/reference/index.html>
   |hands-helping| Community <https://docs.flyte.org/en/latest/community/index.html>

.. toctree::
   :maxdepth: -1
   :caption: User Guide
   :hidden:

   User Guide <self>
   Environment Setup <userguide_setup>
   Basics <auto/core/flyte_basics/index>
   Control Flow <auto/core/control_flow/index>
   Type System <auto/core/type_system/index>
   Testing <auto/testing/index>
   Containerization <auto/core/containerization/index>
   Remote Access <auto/remote_access/index>
   Production Config <auto/deployment/index>
   Scheduling Workflows <auto/core/scheduled_workflows/index>
   Extending Flyte <auto/core/extend_flyte/index>
   Building Large Apps <auto/larger_apps/index>
   contribute

.. toctree::
   :maxdepth: -1
   :caption: Tutorials
   :hidden:

   Tutorials <tutorials>
   ml_training
   feature_engineering
   bioinformatics
   flytelab

.. toctree::
   :maxdepth: -1
   :caption: Integrations
   :hidden:

   Integrations <integrations>
   auto/integrations/flytekit_plugins/sql/index
   auto/integrations/flytekit_plugins/greatexpectations/index
   auto/integrations/flytekit_plugins/papermilltasks/index
   auto/integrations/flytekit_plugins/pandera_examples/index
   auto/integrations/flytekit_plugins/modin_examples/index
   auto/integrations/flytekit_plugins/dolt/index
   auto/integrations/kubernetes/pod/index
   auto/integrations/kubernetes/k8s_spark/index
   auto/integrations/kubernetes/kfpytorch/index
   auto/integrations/kubernetes/kftensorflow/index
   auto/integrations/kubernetes/kfmpi/index
   auto/integrations/aws/sagemaker_training/index
   auto/integrations/aws/sagemaker_pytorch/index
   auto/integrations/aws/athena/index
   auto/integrations/external_services/hive/index
   auto/integrations/external_services/snowflake/index
   auto/integrations/gcp/bigquery/index
