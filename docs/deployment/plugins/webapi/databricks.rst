.. _deployment-plugin-setup-webapi-databricks:

Databricks Plugin
=================

This guide provides an overview of how to set up Databricks in your Flyte deployment.

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

Specify plugin configuration
----------------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. tabs::
      
      .. group-tab:: Demo cluster

        Enable the Databricks plugin on the demo cluster by adding the following config to ``~/.flyte/sandbox/config.yaml``:

        .. code-block:: yaml

          tasks:
            task-plugins:
              default-for-task-types:
                container: container
                container_array: k8s-array
                sidecar: sidecar
                spark: databricks
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - databricks
          plugins:
            databricks:
              entrypointFile: dbfs:///FileStore/tables/entrypoint.py
              databricksInstance: <DATABRICKS_ACCOUNT>.cloud.databricks.com
            k8s:
              default-env-vars:
                - FLYTE_AWS_ACCESS_KEY_ID: <AWS_ACCESS_KEY_ID>
                - FLYTE_AWS_SECRET_ACCESS_KEY: <AWS_SECRET_ACCESS_KEY>
                - AWS_DEFAULT_REGION: <AWS_REGION>
          remoteData:
            region: <AWS_REGION>
            scheme: aws
            signedUrls:
              durationMinutes: 3
          propeller:
            rawoutput-prefix: s3://<S3_BUCKET_NAME>/
          storage:
            container: "<S3_BUCKET_NAME>"
            type: s3
            stow:
              kind: s3
              config:
                region: <AWS_REGION>
                disable_ssl: true
                v2_signing: false
                auth_type: accesskey
                access_key_id: <AWS_ACCESS_KEY_ID>
                secret_key: <AWS_SECRET_ACCESS_KEY>
            signedURL:
              stowConfigOverride:
                endpoint: ""

        Substitute ``<DATABRICKS_ACCOUNT>`` with the name of your Databricks account, 
        ``<AWS_REGION>`` with the region where you created your AWS bucket,
        ``<AWS_ACCESS_KEY_ID>`` with your AWS access key ID,
        ``<AWS_SECRET_ACCESS_KEY>`` with your AWS secret access key,
        and ``<S3_BUCKET_NAME>`` with the name of your S3 bucket.

      .. group-tab:: Helm chart

        Edit the relevant YAML file to specify the plugin.

        .. code-block:: yaml
          :emphasize-lines: 7,11

          tasks:
            task-plugins:
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - databricks
              default-for-task-types:
                - container: container
                - container_array: k8s-array
                - spark: databricks
        
        .. code-block:: yaml
          :emphasize-lines: 3-5

          inline:
            plugins:
              databricks:
                entrypointFile: dbfs:///FileStore/tables/entrypoint.py
                databricksInstance: <DATABRICKS_ACCOUNT>.cloud.databricks.com
        
        Substitute ``<DATABRICKS_ACCOUNT>`` with the name of your Databricks account.

  .. group-tab:: Flyte core

    Create a file named ``values-override.yaml`` and add the following config to it:

    .. code-block:: yaml
      :emphasize-lines: 9,14,15-21

      configmap:
        enabled_plugins:
          tasks:
            task-plugins:
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - databricks
              default-for-task-types:
                container: container
                sidecar: sidecar
                container_array: k8s-array
                spark: databricks
      databricks:
        enabled: True
        plugin_config:
          plugins:
            databricks:
              entrypointFile: dbfs:///FileStore/tables/entrypoint.py
              databricksInstance: <DATABRICKS_ACCOUNT>.cloud.databricks.com
    
    Substitute ``<DATABRICKS_ACCOUNT>`` with the name of your Databricks account.

Add the Databricks access token
-------------------------------

Add the Databricks access token to FlytePropeller:

.. tabs::

  .. group-tab:: Flyte binary

    .. tabs::

      .. group-tab:: Demo cluster

        Add the access token as an environment variable to the ``flyte-sandbox`` deployment.

        .. code-block:: bash

          kubectl edit deploy flyte-sandbox -n flyte

        Update the ``env`` configuration:

        .. code-block:: yaml
          :emphasize-lines: 12-13

          env:
          - name: POD_NAME
            valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
          - name: FLYTE_SECRET_FLYTE_DATABRICKS_API_TOKEN
            value: <ACCESS_TOKEN>
          image: flyte-binary:sandbox
          ...

      .. group-tab:: Helm chart

        Create a secret as follows (or add to it if it already exists from other plugins):

        .. code-block:: bash

          cat <<EOF | kubectl apply -f -
          apiVersion: v1
          kind: Secret
          metadata:
            name: flyte-binary-external-services
            namespace: flyte
          type: Opaque
          stringData:
            FLYTE_DATABRICKS_API_TOKEN: <ACCESS_TOKEN>
          EOF
        
        Reference the newly created secret in 
        ``.Values.configuration.inlineSecretRef``
        in your YAML file as follows:

        .. code-block:: yaml
          :emphasize-lines: 2

          configuration:
            inlineSecretRef: flyte-binary-external-services

    Replace ``<ACCESS_TOKEN>`` with your access token.

  .. group-tab:: Flyte core

    Add the access token as a secret to ``flyte-secret-auth``.

    .. code-block:: bash

      kubectl edit secret -n flyte flyte-secret-auth

    .. code-block:: yaml
      :emphasize-lines: 3

      apiVersion: v1
      data:
        FLYTE_DATABRICKS_API_TOKEN: <ACCESS_TOKEN>
        client_secret: Zm9vYmFy
      kind: Secret
      ...

    Replace ``<ACCESS_TOKEN>`` with your access token.

Upgrade the deployment
----------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. tabs::

      .. group-tab:: Demo cluster

        .. code-block:: bash

          kubectl rollout restart deployment flyte-sandbox -n flyte

      .. group-tab:: Helm chart

        .. code-block:: bash

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