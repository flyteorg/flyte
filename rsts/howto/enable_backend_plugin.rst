.. _howto-enable-backend-plugins:

#################################
How do I enable backend plugins?
#################################

.. tip:: Flyte Backend plugins are awesome, but are not required to extend Flyte! You can always write a `flytekit-only plugin`. Refer to :std:ref:`sphx_glr_auto_core_advanced_custom_task_plugin.py`. Also refer to :ref:`howto-create-plugins`.

Flyte has a unique capability of adding backend plugins. Backend plugins enable Flyte platform to add new capabilities. This has several advantages,

#. Advanced introspection capabilities - ways to improve logging etc
#. Service oriented architecture - ability to bugfix, deploy plugins without releasing new libraries and forcing all users to update their libraries
#. Better management of the system communication - For example in case of aborts, Flyte can guarantee cleanup of the remote resources
#. Reduced cost overhead, for many plugins which launch jobs on a remote service or cluster, the plugins are essentially just polling. This has a huge compute cost in traditional architectures like Airflow etc. Flyte on the other hand, can run these operations in its own control plane.
#. Potential to create drastically new interfaces, that work across multiple languages and platforms.

Ok, How do I enable the backend plugins?
=========================================

To enable a backend plugin you have to add the ``ID`` of the plugin to the enabled plugins list. The ``enabled-plugins`` is available under the ``tasks > task-plugins`` section of FlytePropeller's configuration.
The `plugin configuration structure is defined here <https://pkg.go.dev/github.com/lyft/flytepropeller@v0.6.1/pkg/controller/nodes/task/config#TaskPluginConfig>`_. An example of the config follows,

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/overlays/sandbox/config/propeller/enabled_plugins.yaml
    :language: yaml

How do I find the ``ID`` of the backend plugin?
===============================================
This is a little tricky and sadly at the moment you have to look at the source code of the plugin to figure out the ``ID``. In the case of Spark, for example, the value of ``ID`` is `used <https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L424>`_ here, defined as `spark <https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L41>`_.

Enable a specific Backend Plugin in your own Kustomize generator
=================================================================
Flyte uses Kustomize to generate the the deployment configuration and it can be leveraged to `kustomize your own deployment <https://github.com/flyteorg/flyte/tree/master/kustomize>`_.

We will soon be supporting helm or a better deployment model - See issue :issue:`299`.
