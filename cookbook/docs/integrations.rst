############
Integrations
############

Flyte is designed to be highly extensible and can be customized in multiple ways:

.. panels::
    :header: text-center

    .. link-button:: flytekit_plugins
       :type: ref
       :text: Flytekit Plugins
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    These are Flytekit (python) plugins that are like executing a python function in a container.

    ---

    .. link-button:: native_backend_plugins
       :type: ref
       :text: Native Backend Plugins
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Plugins that enable backend capabilities in Flyte and are independent of external services.

    ---

    .. link-button:: external_services
       :type: ref
       :text: External Services Backend Plugins
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Plugins that enable backend capabilities in Flyte and rely on external services like
    AWS Sagemaker and Hive.

    ---

    Custom Container Executions
    ^^^^^^^^^^^^
    Execute arbitrary containers: You can write c++ code, bash scripts and any containerized program.
    See the :ref:`raw container <raw_container>` as an example.

    ---

    Bring Your Own SDK
    ^^^^^^^^^^^^
    The community would love to help you with your own ideas of building a new SDK. Currently the available SDKs are:

    - `flytekit <https://github.com/flyteorg/flytekit>`_: Flyte Python SDK
    - `flytekit-java <https://github.com/spotify/flytekit-java>`_: Flyte Java/SCALA SDK

Enabling Backend Plugins
^^^^^^^^^^^^^^^^^^^^^^^^^
To enable a backend plugin you have to add the ``ID`` of the plugin to the enabled plugins list. The ``enabled-plugins`` is available under the ``tasks > task-plugins`` section of FlytePropeller's configuration.
The `plugin configuration structure is defined here <https://pkg.go.dev/github.com/flyteorg/flytepropeller@v0.6.1/pkg/controller/nodes/task/config#TaskPluginConfig>`_. An example of the config follows,

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/overlays/sandbox/flyte/config/propeller/enabled_plugins.yaml
    :language: yaml

Finding the ``ID`` of the Backend Plugin
""""""""""""""""""""""""""""""""""""""""
This is a little tricky since you have to look at the source code of the plugin to figure out the ``ID``. In the case of Spark, for example, the value of ``ID`` is `used <https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L424>`_ here, defined as `spark <https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L41>`_.

Enabling a Specific Backend Plugin in Your Own Kustomize Generator
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""
Flyte uses Kustomize to generate the the deployment configuration which can be leveraged to `kustomize your own deployment <https://github.com/flyteorg/flyte/tree/master/kustomize>`_.

.. admonition:: Coming Soon!

    We will soon be supporting Helm. Track the `GitHub Issue <https://github.com/flyteorg/flyte/issues/299>`__ to stay tuned!


.. toctree::
    :maxdepth: -1
    :caption: Integrations
    :hidden:
 
    flytekit_plugins
    native_backend_plugins
    external_services
