.. _gettingstarted:

Getting started
---------------

.. rubric:: Estimated time to complete: 10 minutes.


Prerequisites
***************

Make sure you have `docker <https://docs.docker.com/get-docker/>`__ , `git <https://git-scm.com/>`__ and ``python > 3.6`` installed.

.. caution::

    We have not yet tested this flow on a Windows machine.

Build -> Deploy -> Iterate mechanics
*************************************

Setup
^^^^^^^^^^^^^

#. Install Flyte's python SDK - `flytekit <https://pypi.org/project/flytekit/>`_ (recommended in a virtual environment) and clone the `flytekit-python-template <https://github.com/flyteorg/flytekit-python-template>`_ repo.

   .. prompt::

     pip install flytekit
     git clone git@github.com:flyteorg/flytekit-python-template.git myflyteapp
     cd myflyteapp


#. The repo comes with a sample workflow, which can be found under ``myapp/workflows/example.py``. The structure below shows the most important files and how a typical flyteapp should be laid out.

   .. raw:: html

      <details>
      <summary><a>Important files a typical flyteapp should have</a></summary>

   .. code-block:: text

       .
       ├── Dockerfile
       ├── docker_build_and_tag.sh
       ├── myapp
       │   ├── __init__.py
       │   └── workflows
       │       ├── __init__.py
       │       └── example.py
       └── requirements.txt

   .. note::

       You can use pip-compile to build your requirements file. the Dockerfile that comes with this is not GPU ready, but is a simple Dockerfile that should work for most apps.

   .. raw:: html

      </details>


   The workflow can be run locally simply by running it as a python script - ``note the __main__ at the bottom of the file``

   .. prompt::

       python myapp/workflows/example.py


   .. raw:: html

       <details>
       <summary><a>Expected Output</a></summary>

   .. prompt::

        Running my_wf() hello world

   .. raw:: html

       </details>


#. Install :std:ref:`flytectl`. ``flytectl`` is a commandline interface for flyte.

   .. tabs::

      .. tab:: OSX

        .. prompt::

           brew install flyteorg/homebrew-tap/flytectl

        *Upgrade* existing installation using

        .. prompt::

           brew upgrade flytectl

      .. tab:: Other Operating systems

        .. prompt::

            curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash

   **Test** if flytectl is installed correctly (Expected flytectl version > 0.1.28)::

      flytectl version


#. Flyte can be deployed locally using a single docker container - we refer to this as ``flyte-sandbox`` environment. You can also run this getting started against a hosted / pre-provisioned environment.

   .. tabs::

      .. tab:: Start a new sandbox Cluster

        .. tip:: Want to dive under the hood into flyte-sandbox, refer to the guide `here<>`_.

        .. prompt::

           flytectl sandbox start --sourcesPath <full-path-to-myflyteapp>

      .. tab:: Connect to an existing Flyte cluster

        .. prompt::

            flytectl setup-config


.. _getting-started-standard:

Standard: Build & Deploy your application
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. Flyte uses docker containers to package your workflows and tasks and send it to the remote Flyte cluster. Thus if you notice there is a ``Dockerfile`` already in the cloned repo. You can build the docker container and push the built image to a registry. Follow the instructions below

   .. tabs::

       .. tab:: If using flyte-sandbox

           Since ``flyte-sandbox`` is running locally in a docker container, you do not really need to push the docker image. You can combine the build and push step, by simply building the image inside the flyte-sandbox container. This can be done using

           .. note::

           .. prompt::

               flytectl sandbox exec -- docker build . --tag "myapp:v1"

           .. tip::
            #. Why are we not pushing the docker image? Want to understand details - Refer to guide `here <>`_
            #. *Recommended* use the bundled ./docker_build_and_tag.sh. It will automatically build the local Dockerfile, name it and tag it with the current git-SHA. This helps in gitOps style workflow.

       .. tab:: If using remote flyte cluster

           If you are using a remote flyte cluster, then you need to build your container and push it to a registry that is accessible by the Flyte kubernetes cluster.

           .. prompt::

               docker build . --tag registry/repo:version
               docker push registry/repo:version

#. Now that the container is built, lets provide this information to the Flyte backend. To do that you have to package the workflow using the ``pyflyte`` cli, that is bundled with flytekit. Also note, the image is the same as the one built in the previous step::

    pyflyte --pkgs myapp.workflows package --image myapp:v1

#. Now lets upload this package to flyte backend. We refer to this as ``registration`` ::

    flytectl register files -p flytesnacks -d development -a flyte-package.tgz  -v v1


.. _getting-started-execute:

Execute in remote cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^

Use FlyteConsole to launch an execution and watch the progress.

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/tutorial/exercise.gif
    :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.

**Alternatively** Launch and monitor from CLI using flytectl

Launch an execution using flytectl::

        TODO

Retrieve execution status using flytectl::

        TODO


Modify code: Modify and test locally
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

#. Open ``example.py`` in your favorite editor.

   .. code-block::

       myapp/workflows/example.py

   .. raw:: html

      <details>
      <summary><a>myapp/workflows/example.py</a></summary>

   .. rli:: https://raw.githubusercontent.com/flyteorg/flytekit-python-template/simplify-template/myapp/workflows/example.py
      :language: python

   .. raw:: html

      </details>

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

#. Update the simple test at the bottom of the file to pass in a name. E.g.

   .. code-block:: python

     print(f"Running my_wf(name='adam') {my_wf(name='adam')}")

#. When you run this file locally, it should output ``hello world, adam``. Run this command in your terminal:

   .. prompt::

     python myapp/workflows/example.py


   .. raw:: html

       <details>
       <summary><a>Expected Output</a></summary>

   .. prompt::

        Running my_wf(name='adam') hello world, adam

   .. raw:: html

       </details>


Fast: Deploy your application quickly
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

#. To deploy this workflow to the Flyte cluster (sandbox), you can repeat the previously explained :ref:`getting-started-standard`. But, Flyte provides a faster way to iterate on your workflows. Since you have not really updated any of the dependencies in your requirements file, it is possible to push just the code to Flyte backend, without really re-building the entire docker container.

   .. prompt::

       pyflyte --pkgs myapp.workflows package --image myapp:v1 --fast --force

   .. note::

     Note the ``--fast`` flag. This will take the code from your local machine and provide it for ``execution`` without having to build the container and push it. Also note the ``--force`` flag, this is to simply override your previously created package.

   .. caution::

     The ``fast`` registration method can only be used if you do not modify any requirements. This is because your container / environment is essentially same. But, if you add a dependency you have to follow the :ref:`getting-started-standard` method.


#. You can now deploy the code using flytectl similar to done previously. ``flytectl`` automatically guesses that the package is for ``fast`` registration. For this to work, a new ``storage`` block has to be added to the flytectl configuration with appropriate permissions at runtime. The Storage block configures flytectl to write to a specific ``S3 / GCS bucket``. For sandbox this is automatically configured for you.

   .. prompt::

       flytectl register files -p flytesnacks -d development -a flyte-package.tgz  -v v1-fast1

   .. tabs:: Flytectl configuration with ``storage`` block for Fast registration

       .. tab:: Local Flyte Sandbox

           Automatically configured for you by ``flytectl sandbox`` command

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
                 container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have access to this bucket

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

       .. tab:: *

               For other supported storage backends like Oracle, Azure etc refer to Configuration structure `here <https://pkg.go.dev/github.com/flyteorg/flytestdlib/storage#Config>`_


#. Visit `the console for sandbox <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/myapp.workflows.example.my_wf>`__, click launch, and enter your name as the input.



.. admonition:: TADA! Recap

  You have successfully:

  1. Run a flyte sandbox cluster,
  2. Run a flyte workflow locally,
  3. Run a flyte workflow on a cluster,
  4. Iterated on a flyte workflow.

  .. rubric:: 🎉 Congratulations, you just ran your first Flyte workflow 🎉

Next Steps: User Guide
***********************

To experience the full capabilities of Flyte, take a look at the `User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>`__