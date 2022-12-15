.. _deployment-plugin-setup-webapi-databricks:

Databricks Plugin Setup
-----------------------

This guide gives an overview of how to set up Databricks in your Flyte deployment.

1. Add Flyte chart repo to Helm

.. code-block::

 helm repo add flyteorg https://flyteorg.github.io/flyte


2. Setup the cluster

.. tabbed:: Sandbox

  * Start the sandbox cluster

    .. code-block:: bash

       flytectl sandbox start

  * Generate Flytectl sandbox config

    .. code-block:: bash

       flytectl config init

.. tabbed:: AWS/GCP

  * Make sure you have up and running flyte cluster in `AWS <https://docs.flyte.org/en/latest/deployment/aws/index.html#deployment-aws>`__ / `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__
  * Make sure you have correct kubeconfig and selected the correct kubernetes context
  * make sure you have the correct flytectl config at ~/.flyte/config.yaml

3. Upload an ``entrypoint.py`` to dbfs or s3. Spark driver node run this file to override the default command in the dbx job.

   .. code-block:: python

       # entrypoint.py
       import os
       import sys
       from typing import List

       import click
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
           _execute_task_cmd.callback(**args)


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
               _execute_task_cmd.callback(**args)
           else:
               raise FlyteUserException(f"Unrecognized command: {args[1:]}")


       if __name__ == '__main__':
           main()



4. Create a file named ``values-override.yaml`` and add the following config to it:

.. code-block:: yaml

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
          entrypointFile: dbfs:///FileStore/tables/entrypoint-4.py
          databricksInstance: dbc-a53b7a3c-614c

5. Create a Databricks account and follow the docs for creating an Access token.

6. Create a `Instance Profile <https://docs.databricks.com/administration-guide/cloud-configurations/aws/instance-profiles.html>`_ for the Spark cluster, it allows the spark job to access your data in the s3 bucket.

7. Add Databricks access token to FlytePropeller.

.. note::
        Refer to the `access token <https://docs.databricks.com/dev-tools/auth.html#databricks-personal-access-tokens>`__ to understand setting up the Databricks access token.

.. code-block:: bash

    kubectl edit secret -n flyte flyte-secret-auth

The configuration will look as follows:

.. code-block:: yaml

    apiVersion: v1
    data:
      FLYTE_DATABRICKS_API_TOKEN: <ACCESS_TOKEN>
      client_secret: Zm9vYmFy
    kind: Secret
    metadata:
      annotations:
        meta.helm.sh/release-name: flyte
        meta.helm.sh/release-namespace: flyte
    ...

Replace ``<ACCESS_TOKEN>`` with your access token.

8. Upgrade the Flyte Helm release.

.. code-block:: bash

    helm upgrade -n flyte -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-sandbox.yaml -f values-override.yaml flyteorg/flyte-core
