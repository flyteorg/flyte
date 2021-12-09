.. _deployment-plugin-setup-aws-athena:

Athena Plugin Setup
-------------------

This guide gives an overview of how to set up Athena in your Flyte deployment. Athena plugin needs Flyte deployment in AWS cloud; sandbox/GCP/Azure wouldn't work.

1. Setup the AWS Flyte cluster

.. tabbed:: AWS cluster setup

  * Make sure you have up and running flyte cluster in `AWS <https://docs.flyte.org/en/latest/deployment/aws/index.html#deployment-aws>`__
  * Make sure you have correct kubeconfig and selected the correct kubernetes context
  * make sure you have the correct FlyteCTL config at ~/.flyte/config.yaml


2. Create a file named ``values-override.yaml`` and add the following config to it. Please make sure that the propeller has the correct service account for Athena.

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


3. Upgrade the Flyte Helm release.

.. code-block:: bash

  helm upgrade -n flyte -f values-override.yaml flyteorg/flyte-core


4. Register the Athena plugin example.

.. code-block:: bash

  flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-aws-athena.tar.gz --archive -p flytesnacks -d development


5. Launch an execution

.. tabbed:: Flyte Console

  * Navigate to Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the workflow.
  * Click on `Launch` to open up the launch form.
  * Submit the form.

.. tabbed:: FlyteCTL

  * Retrieve an execution form in the form of a YAML file:

    .. code-block:: bash

       flytectl get launchplan --config ~/.flyte/flytectl.yaml --project flytesnacks --domain development athena.athena.full_hive_demo_wf  --latest --execFile exec_spec.yaml --config ~/.flyte/flytectl.yaml
       flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml


  * Launch! 🚀

    .. code-block:: bash

       flytectl --config ~/.flyte/flytectl.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml