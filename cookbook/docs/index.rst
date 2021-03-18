Flyte Tutorials
==============================================
These tutorials are intended to help the user learn by example. We start with various concepts in flyte and flytekit with examples and
then introduce some of the core plugins. This cookbook is designed to get you running both locally, and on a Flyte cluster using
`Flytekit Python <https://github.com/lyft/flytekit>`__

All the examples are authored using `Flytekit Python <https://github.com/lyft/flytekit>`__ and are designed to be run-locally (this may not be possible for some plugins) or on a Flyte cluster.

The tutorial is divided into 3 sections:

* :ref:`Core Language Examples <flyte-core>` - This covers the basics of Flytekit, as well as working with a deployed Flyte platform.
* :ref:`Case Studies <sphx_glr_auto_case_studies>` - These are examples that showcase the power of Flyte.
* :ref:`Plugins Examples <plugins_main>` - This section details flytekit extensibility

For more details on Flyte refer to the `Homepage <https://flyte.org>`__. For documentation on how to install, configure, administer Flyte, please refer to the
`platform documentation <https://lyft.github.io/flyte/>`__

Please report bugs, suggestions, or comments by creating a `GitHub issue <https://github.com/lyft/flyte/issues/new/choose>`_ or just by opening a PR against this or any Flyte repo.
It is our hope new contributors find the community welcoming!

Please join our Slack channel as well through this `form <https://docs.google.com/forms/d/e/1FAIpQLScWPJZncyL-bBCWbuQ9HWuMzA8-r2RJc__CSWJoAYUQADN-BQ/viewform>`__.

.. toctree::
   :maxdepth: 4
   :caption: Examples

   tutorial
   core
   auto_core_basic/index
   auto_core_intermediate/index
   auto_core_advanced/index
   auto_core_remote_flyte/index

.. toctree::
   :maxdepth: 4
   :caption: Case Studies

   case_studies

.. toctree::
   :maxdepth: 4
   :caption: Flytekit Plugins

   plugins
   auto_plugins_pod/index
   auto_plugins_k8s_spark/index
   auto_plugins_papermilltasks/index
   auto_plugins_hive/index
   auto_plugins_sagemaker_training/index
   auto_plugins_kfpytorch/index
   auto_plugins_sagemaker_pytorch/index


Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`
