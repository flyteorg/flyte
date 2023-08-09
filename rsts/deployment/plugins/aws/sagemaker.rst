.. _deployment-plugin-setup-aws-sagemaker:

Sagemaker Plugin
================

This guide provides an overview of setting up Sagemaker in your Flyte deployment.

.. note::
   
   The Sagemaker plugin requires Flyte deployment in the AWS cloud;
   it is not compatible with demo/GCP/Azure.

Set up AWS Flyte cluster
------------------------

* Ensure you have a functional Flyte cluster running in `AWS <https://docs.flyte.org/en/latest/deployment/aws/index.html#deployment-aws>`__.
* Verify that your AWS role is set up correctly for `SageMaker <https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-roles.html>`__.
* Install the `AWS SageMaker k8s operator <https://github.com/aws/amazon-sagemaker-operator-for-k8s>`__ in your Kubernetes cluster.
* Confirm that you have the correct kubeconfig and have selected the appropriate Kubernetes context.
* Verify the presence of the correct Flytectl configuration at ``~/.flyte/config.yaml``.

Specify the plugin configuration
--------------------------------

.. tabs::

  .. tab:: Flyte binary

    Edit the relevant YAML file to specify the plugin.

    .. code-block:: yaml
      :emphasize-lines: 7,8

      tasks:
        task-plugins:
          enabled-plugins:
            - container
            - sidecar
            - k8s-array
            - sagemaker_training
            - sagemaker_hyperparameter_tuning
          default-for-task-types:
            - container: container
            - container_array: k8s-array

  .. tab:: Flyte core

    Create a file named ``values-override.yaml`` and add the following configuration to it. 

    .. code-block:: yaml

        configmap:
          enabled_plugins:
            # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
            tasks:
              # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
              task-plugins:
                # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config).
                # plugins
                enabled-plugins:
                  - container
                  - sidecar
                  - k8s-array
                  - sagemaker_training
                  - sagemaker_hyperparameter_tuning
                default-for-task-types:
                  container: container
                  sidecar: sidecar
                  container_array: k8s-array

Ensure that the propeller has the correct service account for Sagemaker.

Upgrade the Flyte Helm release
------------------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. code-block:: bash

      helm upgrade <RELEASE_NAME> flyteorg/flyte-binary -n <YOUR_NAMESPACE> --values <YOUR_YAML_FILE>

    Replace ``<RELEASE_NAME>`` with the name of your release (e.g., ``flyte-backend``),
    ``<YOUR_NAMESPACE>`` with the name of your namespace (e.g., ``flyte``),
    and ``<YOUR_YAML_FILE>`` with the name of your YAML file.

  .. group-tab:: Flyte core

    .. code-block:: bash

      helm upgrade <RELEASE_NAME> flyte/flyte-core -n <YOUR_NAMESPACE> --values values-override.yaml

    Replace ``<RELEASE_NAME>`` with the name of your release (e.g., ``flyte``)
    and ``<YOUR_NAMESPACE>`` with the name of your namespace (e.g., ``flyte``).
