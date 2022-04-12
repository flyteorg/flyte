"""
.. _larger_apps_iterate:

Iterate and Re-deploy
----------------------

In this guide, you'll learn how to iterate on and re-deploy your tasks and workflows.

Modify Code and Test Locally
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Open ``example.py`` in your favorite editor.

.. dropdown:: flyte/workflows/example.py

   .. rli:: https://raw.githubusercontent.com/flyteorg/flytekit-python-template/simplify-template/myapp/workflows/example.py
      :language: python

Add ``message: str`` as an argument to both ``my_wf`` and ``say_hello`` functions. Then update the body of
``say_hello`` to consume that argument.

.. code-block:: python

   @task
   def say_hello(message: str) -> str:
      return f"hello world, {message}"

.. code-block:: python

   @workflow
   def my_wf(message: str) -> str:
      res = say_hello(message=message)
      return res

Update the simple test at the bottom of the file to pass in a message, e.g.

.. code-block:: python

   print(f"Running my_wf(message='what a nice day it is!') {my_wf(message='what a nice day it is!')}")

When you run this file locally, it should output ``hello world, what a nice day it is!``.

.. prompt:: bash (flyte)$

   python flyte/workflows/example.py


.. dropdown:: Expected output

      .. prompt:: text

         Running my_wf(message='what a nice day it is!') hello world, what a nice day it is!


Quickly Re-deploy Your Application
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

To re-deploy this workflow to the sandbox Flyte cluster, you can repeat the steps previously covered in
:ref:`getting-started-build-deploy`. Flyte provides a **faster** way to iterate on your workflows. Since you have not
updated any of the dependencies in your requirements file, it is possible to push just the code to Flyte backend without
re-building the entire Docker container. To do so, run the following commands.

.. prompt:: bash (flyte)$

   pyflyte --pkgs flyte.workflows package --image my_flyte_project:v1 --fast --force

.. note::

   ``--fast`` flag will take the code from your local machine and provide it for execution without having to build the container and push it. The ``--force`` flag allows overriding your previously created package.

.. caution::

   The ``fast`` registration method can only be used if you do not modify any requirements (that is, you re-use an existing environment). But, if you add a dependency to your requirements file or env you have to follow the :ref:`getting-started-build-deploy` method.

The code can now be deployed using Flytectl, similar to what we've done previously. ``flytectl`` automatically understands
that the package is for fast registration.

For this to work, a new ``storage`` block has to be added to the Flytectl configuration with appropriate permissions at
runtime. The storage block configures Flytectl to write to a specific ``S3 / GCS bucket``.

.. tip::

   If you're using the sandbox, this is automatically configured by Flytectl.

   The dropdown below provides more information on the required configuration depending on your cloud infrastructure.

   .. dropdown:: Flytectl configuration with ``storage`` block for Fast registration

      .. tabbed:: Local Flyte Sandbox

         Automatically configured for you by ``flytectl sandbox`` command.

         .. code-block:: yaml

            admin:
               # For GRPC endpoints you might want to use dns:///flyte.myexample.com
               endpoint: dns:///localhost:30081
               insecure: true
            storage:
               connection:
                  access-key: minio
                  auth-type: accesskey
                  disable-ssl: true
                  endpoint: http://localhost:30084
                  region: my-region-here
                  secret-key: miniostorage
               container: my-s3-bucket
               type: minio

      .. tabbed:: S3 Configuration

         .. code-block:: yaml

            admin:
               # For GRPC endpoints you might want to use dns:///flyte.myexample.com
               endpoint: dns:///<replace-me>
               authType: Pkce # authType: Pkce # if using authentication or just drop this.
               insecure: true # insecure: True # Set to true if the endpoint isn't accessible through TLS/SSL connection (not recommended except on local sandbox deployment)
            storage:
               type: stow
               stow:
                  kind: s3
                  config:
                     auth_type: iam
                     region: <REGION> # Example: us-east-2
               container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have read access to this bucket


      .. tabbed:: GCS Configuration

         .. code-block:: yaml

            admin:
               # For GRPC endpoints you might want to use dns:///flyte.myexample.com
               endpoint: dns:///<replace-me>
               authType: Pkce # authType: Pkce # if using authentication or just drop this.
               insecure: false # insecure: True # Set to true if the endpoint isn't accessible through TLS/SSL connection (not recommended except on local sandbox deployment)
            storage:
               type: stow
               stow:
                  kind: google
                  config:
                     json: ""
                     project_id: <replace-me> # replace <project-id> with the GCP project ID
                     scopes: https://www.googleapis.com/auth/devstorage.read_write
               container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have access to this bucket

      .. tabbed:: Others

         For other supported storage backends like Oracle, Azure, etc., refer to the configuration structure `here <https://pkg.go.dev/github.com/flyteorg/flytestdlib/storage#Config>`__.


Next, you can simply register your packaged workflows with:

.. prompt:: bash $

   flytectl register files --project flytesnacks --domain development --archive flyte-package.tgz  --version v1-fast1


Execute Your Re-deployed Workflow
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Finally, you can execute the updated workflow programmatically with ``flytectl``.

To pass arguments to the workflow, update the execution spec file that we previously generated in the
:ref:`Deploying to the Coud <larger_apps_deploy>` step.

Generate an execution spec file. This will prompt you to overwrite and answer 'y' on it.

.. prompt:: bash $

   flytectl get launchplan --project flytesnacks --domain development flyte.workflows.example.my_wf --latest --execFile exec_spec.yaml

Modify the execution spec file and update the input params and save the file. Notice that the version would be changed to your latest one.

.. code-block:: yaml

   ....
   inputs:
      message: "what's up doc? 🐰🥕"
   ....
   version: v1-fast1

Create an execution using the exec spec file.

.. prompt:: bash $

   flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml

You should see an output that looks like:

.. prompt:: text

   execution identifier project:"flytesnacks" domain:"development" name:"<execution_name>"

Monitor the execution by providing the execution name from the ``create execution`` command.

.. prompt:: bash $

   flytectl get execution --project flytesnacks --domain development <execution_name>

.. tip::

   Alternatively, visit `the Flyte sandbox console <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/flyte.workflows.example.my_wf>`__,
   click launch, and provide a ``message`` as input.

   .. TODO: update so that it reflects the new workflow with the message input https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/getting_started/getting_started_fastreg.gif


Recap
^^^^^^

In this guide, we:

1. Setup a Flyte project with ``pyflyte init my_flyte_project`` and
   ran your workflows locally.
2. Started a Flyte sandbox cluster and ran a Flyte workflow on a cluster.
3. Iterated on a Flyte workflow and updated the workflows on the cluster.


What's Next?
^^^^^^^^^^^^^

To experience the full power of Flyte on distributed compute, take a look at the
`Deployment Guides <https://docs.flyte.org/en/latest/deployment/index.html>`__.

"""
