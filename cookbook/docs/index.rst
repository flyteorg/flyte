.. _userguide:

##############
User Guide
##############

The user guide covers all of the key features of Flyte organized by topic. Each of the sections
below introduces a topic and discusses how you can use Flyte to address a specific problem.

Code for all of the examples in the user guide be found `here <https://github.com/flyteorg/flytesnacks>`__.

If this is your first time using Flyte, check out the
`Getting Started <https://docs.flyte.org/en/latest/getting_started.html>`__ guide.


.. TODO: add control plane section to the panels

.. panels::
   :header: text-center

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

   .. link-button:: deployment
      :type: ref
      :text: 🚢  Deployment
      :classes: btn-block stretched-link
   ^^^^^^^^^^
   Ship data processing and machine learning pipelines to production.

   ---

   .. link-button:: integrations
      :type: ref
      :text: 🔌  Integrations
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Leverage a rich ecosystem of plugins from compute infrastructure to jupyter notebooks.

   ---

   .. link-button:: auto/core/extend_flyte/index
      :type: ref
      :text: 🏗 Extending Flyte
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^^^^
   Define custom plugins that aren't currently supported in the Flyte ecosystem.


.. toctree::
   :maxdepth: 1
   :hidden:

   |plane| Getting Started <https://docs.flyte.org/en/latest/getting_started.html>
   |book-reader| User Guide <index>
   |chalkboard| Tutorials <tutorials>
   |project-diagram| Concepts <https://docs.flyte.org/en/latest/concepts/basics.html>
   |book| API Reference <https://docs.flyte.org/en/latest/reference/index.html>
   |hands-helping| Community <https://docs.flyte.org/en/latest/community/index.html>

.. TODO: add control plane entry in the toctree once examples are written
.. Control Plane <auto/control_plane/index>

.. toctree::
   :maxdepth: -1
   :caption: User Guide
   :hidden:

   Introduction <self>
   Basics <auto/core/flyte_basics/index>
   Control Flow <auto/core/control_flow/index>
   Type System <auto/core/type_system/index>
   Testing <auto/testing/index>
   Containerization <auto/core/containerization/index>
   deployment
   integrations
   Extending flyte <auto/core/extend_flyte/index>

.. TODO: add the following sections when they are complete:
   - data_processing
   - ml_monitoring
   - feature_engineering
   - batch_prediction

.. toctree::
   :maxdepth: -1
   :caption: Tutorials
   :hidden:

   Introduction <tutorials>
   ml_training
   feature_engineering
