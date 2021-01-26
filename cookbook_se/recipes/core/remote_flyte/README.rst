Getting setup for remote execution
-------------------------------------
Flytekit provides a python SDK for authoring and executing workflows and tasks in python.
Flytekit comes with a simplistic local scheduler that executes code in a local environment.
But, to leverage the full power of Flyte, we recommend using a deployed backend of Flyte. Flyte can be run
on a kubernetes cluster - locally, in a cloud environment or on-prem.

Please refer to the `Installing Flyte <https://lyft.github.io/flyte/administrator/install/index.html>`_ for details on getting started with a Flyte installation.
This section walks through steps on deploying your local workflow to a distributed Flyte environment, with ``NO CODE CHANGES``.

Build your Dockerfile
^^^^^^^^^^^^^^^^^^^^^^
Now that you have workflows running locally, its time to take them for a spin onto a Hosted Flyte backend.

.. literalinclude:: ../../Dockerfile
    :language: dockerfile
    :emphasize-lines: 1
    :linenos:


Serialize your workflows and tasks
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Once you've built a Docker container image with your updated code changes, you can use the predefined make target to easily serialize your tasks:

.. code-block::

   make serialize

This runs the `pyflyte serialize` command to convert your workflow and task definitions to registerable protos.
The make target is a handy wrapper around the following:

.. code-block::

   pyflyte -c sandbox.config --pkgs recipes serialize --in-container-config-path /root/sandbox.config --local-source-root ${CURDIR} --image ${FULL_IMAGE_NAME}:${VERSION} workflows -f _pb_output/

- the `-c` is the path to the config definition on your machine. This config specifies SDK default attributes.
- the `--pkgs` arg points to the packages within the
- `--local-source-root` which contains the code copied over into your Docker container image that will be serialized (and later, executed)
- `--in-container-config-path` maps to the location within your Docker container image where the above config file will be copied over too
- `--image` is the non-optional fully qualified name of the container image housing your code

To avoid mucking with with specifying out of container configs and code paths you can also use the handy in-container serialize recipe:

.. code-block::

   make serialize_sandbox

Register your Workflows and Tasks
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Once you've serialized your workflows and tasks to proto, you'll need to register them with your deployed Flyte installation.
Again, you can make use of the included make target like so:

.. code-block::

   OUTPUT_DATA_PREFIX=s3://my-s3-bucket/raw_data FLYTE_HOST=flyte.example.com make register

making sure to appropriately substitute the correct output data location (to persist workflow execution outputs) along
with the URL to your hosted Flyte deployment.

Under the hood this recipe again supplies some defaults you may find yourself wishing to customize. Specifically, this recipe calls:

.. code-block::

   flyte-cli register-files -p flytetester -d development -v ${VERSION} --kubernetes-service-account demo \
       --output-location-prefix s3://my-s3-bucket/raw_data -h flyte.example.com _pb_output/*


Of interest are the following args:

- `-p` specifies the project to register your entities. This project itself must already be registered on your Flyte deployment.
- `-d` specifies the domain to register your entities. This domain must already be configured in your Flyte deployment
- `-v` is a unique string used to identify this version of entities registered under a project and domain.
- If required, you can specify a `kubernetes-service-account` or `assumable_iam_role` which your tasks will run with.


Fast(er) iteration
^^^^^^^^^^^^^^^^^^
Re-building a new Docker container image for every code change you make can become cumbersome and slow.
If you're making purely code changes that **do not** require updating your container definition, you can make use of
fast serialization and registration to speed up your iteration process and reduce the time it takes to upload new entity
versions to your hosted Flyte deployment.

First, run the fast serialization target:

.. code-block::

   make fast_serialize

And then the fast register target:

.. code-block::

   OUTPUT_DATA_PREFIX=s3://my-s3-bucket/raw_data FLYTE_HOST=flyte.example.com ADDL_DISTRIBUTION_DIR=s3://my-s3-bucket/archives make register

and just like that you can update your code without requiring a rebuild of your container!

.. _working_hosted_service:

Features @Hosted Flyte: Schedules, Notifications etc
-----------------------------------------------------

Using remote Flyte gives you the ability to:

- Use caching to avoid calling the same task with the same inputs (for the same version)
- Portability: You can reference pre-registered entities under any domain or project within your workflow code
- Sharable executions: you can easily share links to your executions with your teammates