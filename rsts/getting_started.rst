.. _gettingstarted:

Getting Started
---------------

.. raw:: html
  
    <p style="color: #808080; font-weight: 500; font-size: 20px; padding-top: 10px;">A step-by-step guide to get acquainted with the Flyte environment</p>

First Things First
******************

Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ , `Git <https://git-scm.com/>`__, and `Python <https://www.python.org/downloads/>`__ >= 3.7 installed.

.. caution::

    We have not yet tested this flow on a Windows machine.

The Standard Process
********************

Set Up the Flyte Environment
^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. Install Flyte's Python SDK —— `Flytekit <https://pypi.org/project/flytekit/>`__ (recommended in a virtual environment) and clone the `flytekit-python-template <https://github.com/flyteorg/flytekit-python-template>`__ repo.

   .. prompt::

     pip install flytekit
     git clone https://github.com/flyteorg/flytekit-python-template.git myflyteapp
     cd myflyteapp


#. The repo comes with a sample workflow, which can be found under ``myapp/workflows/example.py``. The structure below shows the most important files and how a typical Flyte app should be laid out.

   .. dropdown:: A typical Flyte app should have these files

       .. code-block:: text

           .
           ├── Dockerfile
           ├── docker_build_and_tag.sh
           ├── myapp
           │         ├── __init__.py
           │         └── workflows
           │             ├── __init__.py
           │             └── example.py
           └── requirements.txt

       .. note::

           Two things to note here:

           * You can use `pip-compile` to build your requirements file. 
           * The Dockerfile that comes with this is not GPU ready, but is a simple Dockerfile that should work for most of your apps.

   The workflow can be run locally, simply by running it as a Python script —— note the ``__main__`` entry point at the `bottom of the file <https://github.com/flyteorg/flytekit-python-template/blob/main/myapp/workflows/example.py#L58>`__.

   .. prompt::

       python myapp/workflows/example.py

   .. dropdown:: Expected output

      .. prompt::

         Running my_wf() hello world

#. Next, install :std:ref:`flytectl`. ``Flytectl`` is a commandline interface for Flyte.

   .. tabs::

      .. tab:: OSX

        .. prompt::

           brew install flyteorg/homebrew-tap/flytectl

        *Upgrade* existing installation using the following command:

        .. prompt::

           brew upgrade flytectl

      .. tab:: Other Operating systems

        .. prompt::

            curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash

   **Test** if Flytectl is installed correctly (your Flytectl installation should have version > 0.1.28) using the following command: ::

      flytectl version


#. Flyte can be deployed locally using a single docker container —— we refer to this as the ``flyte-sandbox`` environment. You can also run this getting started against a hosted or pre-provisioned environment.

   .. tabs::

      .. tab:: Start a new sandbox cluster

        .. tip:: Want to dive under the hood into flyte-sandbox, refer to the guide `here<>`_.

        .. prompt::

           flytectl sandbox start --sourcesPath <full-path-to-myflyteapp>

      .. tab:: Connect to an existing Flyte cluster

        .. prompt::

            flytectl init


.. _getting-started-build-deploy:

Build & Deploy Your Application
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. Flyte uses Docker containers to package your workflows and tasks and sends them to the remote Flyte cluster. Thus, there is a ``Dockerfile`` already included in the cloned repo. You can build the docker container and push the built image to a registry. Read further to know how to do so.

   .. tabs::

       .. tab:: Flyte Sandbox

           Since ``flyte-sandbox`` runs locally in a docker container, you do not need to push the docker image. You can combine the build and push step by simply building the image inside the Flyte-sandbox container. This can be done using the following command:

           .. prompt::

               flytectl sandbox exec -- docker build . --tag "myapp:v1"

           .. tip::
            #. Why are we not pushing the docker image? Want to understand the details —— refer to guide `here <>`_
            #. *Recommended:* Use the bundled `./docker_build_and_tag.sh`. It will automatically build the local Dockerfile, name it and tag it with the current git-SHA. This helps in achieving GitOps style workflows.

       .. tab:: Remote Flyte Cluster

           If you are using a remote Flyte cluster, then you need to build your container and push it to a registry that is accessible by the Flyte Kubernetes cluster.

           .. prompt::

               docker build . --tag registry/repo:version
               docker push registry/repo:version

#. Now that the container is built let's provide this information to the Flyte backend, for which you have to package the workflow using the ``pyflyte`` cli bundled with Flytekit. Also, note that the image is the same as the one built in the previous step. ::

    pyflyte --pkgs myapp.workflows package --image myapp:v1

#. Let's now upload this package to the Flyte backend. We refer to this as ``registration``. ::

    flytectl register files -p flytesnacks -d development -a flyte-package.tgz  -v v1

#. Visualize the registered workflow. ::

    flytectl get workflows -p flytesnacks -d development myapp.workflows.example.my_wf -o doturl


.. _getting-started-execute:

Execute on Flyte
^^^^^^^^^^^^^^^^

Finally, use FlyteConsole to launch an execution and keep tabs on the window! 

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/tutorial/exercise.gif
    :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.

**Alternatively,** 

Launch and monitor from CLI using flytectl. This is how you will have to proceed.
More details can be found `here <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html>`__

#. Generate execution spec file. ::

    flytectl get launchplan -p flytesnacks -d development myapp.workflows.example.my_wf  --execFile exec_spec.yaml

#. Update the input spec file for arguments to the workflow. ::

    .. code-block:: text
            ....
            inputs:
              name: "adam"
            ....

#. Create execution using the exec spec file. ::

    flytectl create execution -p flytesnacks -d development --execFile exec_spec.yaml


#. Monitor the execution by providing the execution id from create command. ::

    flytectl get execution -p flytesnacks -d development <execid>



For the Explorers
*****************

If you're interested in poking around the various ways to make running code on Flyte comfortable, look further.

Modify and Test Locally
^^^^^^^^^^^^^^^^^^^^^^^
#. Open ``example.py`` in your favorite editor.

   .. code-block::

       myapp/workflows/example.py

   .. dropdown:: myapp/workflows/example.py

      .. rli:: https://raw.githubusercontent.com/flyteorg/flytekit-python-template/simplify-template/myapp/workflows/example.py
         :language: python

#. Add ``name: str`` as an argument to both ``my_wf`` and ``say_hello`` functions. Then update the body of ``say_hello`` to consume that argument.

   .. code-block:: python

     @task
     def say_hello(name: str) -> str:
         return f"hello world, {name}"

   .. code-block:: python

     @workflow
     def my_wf(name: str) -> str:
         res = say_hello(name=name)
         return res

#. Update the simple test at the bottom of the file to pass in a name, e.g.

   .. code-block:: python

     print(f"Running my_wf(name='adam') {my_wf(name='adam')}")

#. When you run this file locally, it should output ``hello world, adam``.

   .. prompt::

     python myapp/workflows/example.py


   .. dropdown:: Expected output

       .. prompt::

            Running my_wf(name='adam') hello world, adam

Build & Deploy Your Application "Fast"er!
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. To deploy this workflow to the Flyte cluster (sandbox), you can repeat the steps previously covered in :ref:`getting-started-build-deploy`. Flyte provides a faster way to iterate on your workflows. Since you have not updated any of the dependencies in your requirements file, it is possible to push just the code to Flyte backend without re-building the entire docker container. To do so, run the following commands.

   .. prompt::

       pyflyte --pkgs myapp.workflows package --image myapp:v1 --fast --force

   .. note::

      ``--fast`` flag —— this will take the code from your local machine and provide it for execution without having to build the container and push it. The ``--force`` flag allows overriding your previously created package.

   .. caution::

      The ``fast`` registration method can only be used if you do not modify any requirements (that is, you re-use an existing environment). But, if you add a dependency to your requirements file or env you have to follow the :ref:`getting-started-build-deploy` method.

#. The code can now be deployed using Flytectl, similar to what we've done previously. ``Flytectl`` automatically understands that the package is for ``fast`` registration.
   For this to work, a new ``storage`` block has to be added to the Flytectl configuration with appropriate permissions at runtime. The storage block configures Flytectl to write to a specific ``S3 / GCS bucket``. If you're using the sandbox, this is automatically configured by Flytectl, so you can skip this for now. But do take a note for the future.

   .. prompt::

       flytectl register files -p flytesnacks -d development -a flyte-package.tgz  -v v1-fast1

   .. tabs:: Flytectl configuration with ``storage`` block for Fast registration

       .. tab:: Local Flyte Sandbox

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

       .. tab:: S3 Configuration

           .. code-block:: yaml

               admin:
                 # For GRPC endpoints you might want to use dns:///flyte.myexample.com
                 endpoint: dns:///<replace-me>
                 authType: Pkce # if using authentication or just drop this. If insecure set insecure: True
               storage:
                 kind: s3
                 config:
                   auth_type: iam
                   region: <replace> # Example: us-east-2
                 container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have read access to this bucket

       .. tab:: GCS Configuration

           .. code-block:: yaml

               admin:
                 # For GRPC endpoints you might want to use dns:///flyte.myexample.com
                 endpoint: dns:///<replace-me>
                 authType: Pkce # if using authentication or just drop this. If insecure set insecure: True
               storage:
                 kind: google
                 config:
                   json: ""
                   project_id: <replace-me> # TODO: replace <project-id> with the GCP project ID
                   scopes: https://www.googleapis.com/auth/devstorage.read_write
                 container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have access to this bucket

       .. tab:: Others

               For other supported storage backends like Oracle, Azure, etc., refer to the configuration structure `here <https://pkg.go.dev/github.com/flyteorg/flytestdlib/storage#Config>`__.


#. Finally, visit `the sandbox console <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/myapp.workflows.example.my_wf>`__, click launch, and give your name as the input. Let the magic happen!


.. admonition:: TADA!

  You have successfully:

  1. Run a Flyte sandbox cluster,
  2. Run a Flyte workflow locally,
  3. Run a Flyte workflow on a cluster,
  4. Iterated on a Flyte workflow.

  .. rubric:: 🎉 Congratulations! you just ran your first Flyte workflow! 🎉

Next Steps: User Guide
**********************

To experience the full capabilities of Flyte, take a look at the `User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>`__.