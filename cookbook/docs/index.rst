Flyte Cookbook
==============================================
Flytesnacks is intended to be a ``learn by example`` style cookbook.  We cover all the various concepts in Flyte along with most of the maintained
plugins. This cookbook is designed to get you running both locally, and on a Flyte cluster using `Flyte Python SDK (flytekit) <https://github.com/lyft/flytekit>`__

All the examples in this book are executable locally and we recommend the best way to try out Flyte is to run these examples in your terminal or IDE.
If you have access to a whole Flyte platform deployment, this cookbook can also be used as a fully functional workflow repo.

The tutorial is divided into 3 sections:

* :ref:`Getting Started <sphx_glr_auto_core>` - This covers the basics of Flytekit, as well as working with a deployed Flyte platform.
* :ref:`Case Studies <sphx_glr_auto_case_studies>` - These are examples that showcase the power of Flyte.
* :ref:`Plugins <sphx_glr_auto_plugins>` - This section details flytekit extensibility

For more details on Flyte refer to the `Homepage <https://flyte.org>`__. For documentation on how to install, configure, administer Flyte, please refer to the
`platform documentation <https://lyft.github.io/flyte/>`__

Please report bugs, suggestions, or comments by creating a `GitHub issue <https://github.com/lyft/flyte/issues/new/choose>`_ or just by opening a PR against this or any Flyte repo.
It is our hope new contributors find the community welcoming!

Please join our Slack channel as well through this `form <https://docs.google.com/forms/d/e/1FAIpQLScWPJZncyL-bBCWbuQ9HWuMzA8-r2RJc__CSWJoAYUQADN-BQ/viewform>`__.

.. toctree::
   :maxdepth: 4
   :caption: Getting Started

   core
   auto_core_basic/index
   auto_core_intermediate/index
   auto_core_advanced/index

.. toctree::
   :maxdepth: 4
   :caption: Case Studies

   case_studies

.. toctree::
   :maxdepth: 4
   :caption: Flytekit Plugins

   plugins
   auto_plugins_hive/index
   auto_plugins_sagemaker_training/index
   auto_plugins_k8s_spark/index
   auto_plugins_kfpytorch/index
   auto_plugins_pod/index
   auto_plugins_sagemaker_pytorch/index

.. toctree::
   :maxdepth: 4
   :caption: Additional Flyte Documentation

   Flyte Project <https://flyte.readthedocs.io/en/latest//>
   Flytekit Python <https://flytekit.readthedocs.io/en/latest/">


Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`
