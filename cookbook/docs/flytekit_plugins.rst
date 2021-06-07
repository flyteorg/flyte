Flytekit Plugins
================

Flytekit plugins are simple plugins that can be implemented purely in python, unit tested locally and allow extending
Flytekit functionality. These plugins can be anything and for comparison can be thought of like
`Airflow Operators <https://airflow.apache.org/docs/apache-airflow/stable/howto/operator/index.html>`__.

Data is automatically marshalled and unmarshalled into and out of the plugin and mostly users should implement the
:py:class:`~flytekit.core.base_task.PythonTask` API, defined in flytekit. This tutorial will illustrate how a plugin can
be implemented with the help of an example.

Flytekit Plugins are lazily loaded and can be released independently like libraries. We follow a convention to name the
plugin like ``flytekitplugins-*``, where * indicates the package to be integrated into flytekit. For example
``flytekitplugins-papermill`` enables users to author flytekit tasks using `Papermill <https://papermill.readthedocs.io/en/latest/>`__.

You can find the plugins maintained by the core flyte team `here <https://github.com/flyteorg/flytekit/tree/master/plugins>`__.

.. panels::
   :header: text-center

   .. link-button:: auto/integrations/flytekit_plugins/papermilltasks/index
      :type: ref
      :text: Papermill
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Execute Jupyter Notebooks with ``papermill``.

   ---

   .. link-button:: auto/integrations/flytekit_plugins/pandera/index
      :type: ref
      :text: Pandera
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Validate pandas dataframes with ``pandera``.


.. TODO: add the following items to the TOC when the content is written.
.. - auto/integrations/flytekit_plugins/sqllite3/index
.. - auto/integrations/flytekit_plugins/sqlalchemy/index

.. toctree::
   :maxdepth: -1
   :caption: Contents
   :hidden:

   auto/integrations/flytekit_plugins/papermilltasks/index
   auto/integrations/flytekit_plugins/pandera/index
   auto/integrations/flytekit_plugins/dolt/index
