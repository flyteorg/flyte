.. _deployment-plugin-setup-webapi-athena:

Athena Plugin Setup
---------------------

This guide gives an overview of how to set up the Athena in your AWS Flyte deployment. Athena plugin need flyte deployment in aws cloud, It will not work in sandbox/GCP/Azure flyte deployment. Please provide correct service account to flyte propeller for athena access.


1. Create a file named ``values-athena.yaml`` and add the following config to it. Please make sure that the propeller has the correct service account for Athena.

.. code-block::

    configmap:
      enabled_plugins:
        # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
        tasks:
          # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
          task-plugins:
            # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend
            # plugins
            enabled-plugins:
              - container
              - sidecar
              - k8s-array
              - athena
            default-for-task-types:
              container: container
              sidecar: sidecar
              container_array: k8s-array
              athena: athena

2. Upgrade the Flyte Helm release.

   .. code-block:: bash

      helm upgrade -n flyte -f values-athena.yaml flyteorg/flyte

3. Register the Athena plugin example.

   .. code-block:: bash

      flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-aws-athena.tar.gz --archive -p flytesnacks -d development

4. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development  athena.athena.full_hive_demo_wf  --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execution_id>
