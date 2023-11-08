.. _deployment-agent-setup-databricks:

Databricks Agent
=================

This guide provides an overview of how to set up Databricks agent in your Flyte deployment.

Spin up a cluster
-----------------

.. tabs::

  .. group-tab:: Flyte binary
      
    You can spin up a demo cluster using the following command:
   
    .. code-block:: bash
  
      flytectl demo start

    Or install Flyte using the :ref:`flyte-binary helm chart <deployment-deployment-cloud-simple>`.

  .. group-tab:: Flyte core

    If you've installed Flyte using the 
    `flyte-core helm chart <https://github.com/flyteorg/flyte/tree/master/charts/flyte-core>`__, please ensure:

    * You have the correct kubeconfig and have selected the correct Kubernetes context.
    * You have configured the correct flytectl settings in ``~/.flyte/config.yaml``.

.. note::

  Add the Flyte chart repo to Helm if you're installing via the Helm charts.

  .. code-block:: bash

    helm repo add flyteorg https://flyteorg.github.io/flyte

Databricks workspace
--------------------

To set up your Databricks account, follow these steps:

1. Create a `Databricks account <https://www.databricks.com/>`__.
2. Ensure that you have a Databricks workspace up and running.
3. Generate a `personal access token 
   <https://docs.databricks.com/dev-tools/auth.html#databricks-personal-ACCESS_TOKEN-authentication>`__ to be used in the Flyte configuration.
   You can find the personal access token in the user settings within the workspace.

.. note::

  When testing the Databricks plugin on the demo cluster, create an S3 bucket because the local demo 
  cluster utilizes MinIO. Follow the `AWS instructions 
  <https://docs.aws.amazon.com/powershell/latest/userguide/pstools-appendix-sign-up.html>`__
  to generate access and secret keys, which can be used to access your preferred S3 bucket.

Create an `instance profile 
<https://docs.databricks.com/administration-guide/cloud-configurations/aws/instance-profiles.html>`__ 
for the Spark cluster. This profile enables the Spark job to access your data in the S3 bucket.
Please follow all four steps specified in the documentation.

Upload the following entrypoint.py file to either 
`DBFS <https://docs.databricks.com/archive/legacy/data-tab.html>`__ 
(the final path can be ``dbfs:///FileStore/tables/entrypoint.py``) or S3. 
This file will be executed by the Spark driver node, overriding the default command in the 
`dbx <https://docs.databricks.com/dev-tools/dbx.html>`__ job.

.. TODO: A quick-and-dirty workaround to resolve https://github.com/flyteorg/flyte/issues/3853 issue is to import pandas.

.. code-block:: python

  import os
  import sys
  from typing import List

  import click
  import pandas
  from flytekit.bin.entrypoint import fast_execute_task_cmd as _fast_execute_task_cmd
  from flytekit.bin.entrypoint import execute_task_cmd as _execute_task_cmd
  from flytekit.exceptions.user import FlyteUserException
  from flytekit.tools.fast_registration import download_distribution


  def fast_execute_task_cmd(additional_distribution: str, dest_dir: str, task_execute_cmd: List[str]):
      if additional_distribution is not None:
          if not dest_dir:
              dest_dir = os.getcwd()
          download_distribution(additional_distribution, dest_dir)

      # Insert the call to fast before the unbounded resolver args
      cmd = []
      for arg in task_execute_cmd:
          if arg == "--resolver":
              cmd.extend(["--dynamic-addl-distro", additional_distribution, "--dynamic-dest-dir", dest_dir])
          cmd.append(arg)

      click_ctx = click.Context(click.Command("dummy"))
      parser = _execute_task_cmd.make_parser(click_ctx)
      args, _, _ = parser.parse_args(cmd[1:])
      _execute_task_cmd.callback(test=False, **args)


  def main():

      args = sys.argv

      click_ctx = click.Context(click.Command("dummy"))
      if args[1] == "pyflyte-fast-execute":
          parser = _fast_execute_task_cmd.make_parser(click_ctx)
          args, _, _ = parser.parse_args(args[2:])
          fast_execute_task_cmd(**args)
      elif args[1] == "pyflyte-execute":
          parser = _execute_task_cmd.make_parser(click_ctx)
          args, _, _ = parser.parse_args(args[2:])
          _execute_task_cmd.callback(test=False, dynamic_addl_distro=None, dynamic_dest_dir=None, **args)
      else:
          raise FlyteUserException(f"Unrecognized command: {args[1:]}")


  if __name__ == '__main__':
      main()

Specify agent configuration
----------------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. tabs::
      
      .. group-tab:: Demo cluster

        Enable the Databricks agent on the demo cluster by updating the ConfigMap:

        .. code-block:: bash

          kubectl edit configmap flyte-sandbox-config -n flyte

        .. code-block:: yaml
          :emphasize-lines: 7,12,16

          tasks:
            task-plugins:
              default-for-task-types:
                container: container
                container_array: k8s-array
                sidecar: sidecar
                spark: agent-service
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - agent-service
          plugins:
            agent-service:
              supportedTaskTypes:
              - spark

      .. group-tab:: Helm chart

        Edit the relevant YAML file to specify the plugin.

        .. code-block:: yaml
          :emphasize-lines: 7,11,15

          tasks:
            task-plugins:
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - agent-service
              default-for-task-types:
                - container: container
                - container_array: k8s-array
                - spark: agent-service
            plugins:
              agent-service:
                supportedTaskTypes:
                - spark

  .. group-tab:: Flyte core

    Create a file named ``values-override.yaml`` and add the following config to it:

    .. code-block:: yaml
      :emphasize-lines: 9,14,15-21

        enabled_plugins:
          tasks:
            task-plugins:
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - agent-service
              default-for-task-types:
                container: container
                sidecar: sidecar
                container_array: k8s-array
                spark: agent-service
          plugins:
            agent-service:
              supportedTaskTypes:
              - spark

Add the Databricks access token
-------------------------------

You have to set the Databricks token to the Flyte configuration.

1. Install flyteagent pod using helm
  
.. code-block::
  
  helm repo add flyteorg https://flyteorg.github.io/flyte
  helm install flyteagent flyteorg/flyteagent --namespace flyte

2. Get the base64 value of your Databricks token.

.. code-block::

  echo -n "<DATABRICKS_TOKEN>" | base64

3. Edit the flyteagent secret
  
      .. code-block:: bash
    
        kubectl edit secret flyteagent -n flyte
    
      .. code-block:: yaml
        :emphasize-lines: 3

        apiVersion: v1
        data:
          flyte_databricks_access_token: <BASE64_ENCODED_DATABRICKS_TOKEN>
        kind: Secret
        metadata:
          annotations:
            meta.helm.sh/release-name: flyteagent
            meta.helm.sh/release-namespace: flyte
          creationTimestamp: "2023-10-04T04:09:03Z"
          labels:
            app.kubernetes.io/managed-by: Helm
          name: flyteagent
          namespace: flyte
          resourceVersion: "753"
          uid: 5ac1e1b6-2a4c-4e26-9001-d4ba72c39e54
        type: Opaque


Upgrade the deployment
----------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. tabs::

      .. group-tab:: Demo cluster

        .. code-block::

          kubectl rollout restart deployment flyte-sandbox -n flyte

      .. group-tab:: Helm chart

        .. code-block::

          helm upgrade <RELEASE_NAME> flyteorg/flyte-binary -n <YOUR_NAMESPACE> --values <YOUR_YAML_FILE>

        Replace ``<RELEASE_NAME>`` with the name of your release (e.g., ``flyte-backend``),
        ``<YOUR_NAMESPACE>`` with the name of your namespace (e.g., ``flyte``),
        and ``<YOUR_YAML_FILE>`` with the name of your YAML file.

  .. group-tab:: Flyte core

    .. code-block::

      helm upgrade <RELEASE_NAME> flyte/flyte-core -n <YOUR_NAMESPACE> --values values-override.yaml

    Replace ``<RELEASE_NAME>`` with the name of your release (e.g., ``flyte``)
    and ``<YOUR_NAMESPACE>`` with the name of your namespace (e.g., ``flyte``).

Wait for the upgrade to complete. You can check the status of the deployment pods by running the following command:

.. code-block::

  kubectl get pods -n flyte

.. note::

  Make sure you enable `custom containers 
  <https://docs.databricks.com/administration-guide/clusters/container-services.html>`__
  on your Databricks cluster before you trigger the workflow.
