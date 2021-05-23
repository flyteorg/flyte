.. _deployment_cluster:

Flyte cluster execution - setup & features
-------------------------------------------
Locally, Flytekit relies on the Python interpreter to execute both tasks and workflows.
To leverage the full power of Flyte, we recommend using a deployed backend of Flyte. Flyte can be run
on any Kubernetes cluster - a local cluster like `kind <https://kind.sigs.k8s.io/>`__, in a cloud environment or on-prem.

Please refer to the :doc:`Getting Started <flyte:getting_started>` for details on getting started with the Flyte installation.

1. First commit your changes. Some of the steps below default to referencing the git sha.
2. Run `make serialize`. This will build the image tagged with just `flytecookbook:<sha>`, no registry will be prefixed. See the image building section below for additional information.

Build your Dockerfile
^^^^^^^^^^^^^^^^^^^^^^
The first step of this process is building a container image that holds your code.

.. code-block:: docker
   :emphasize-lines: 1
   :linenos:

   FROM python:3.8-slim-buster
   LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks

   WORKDIR /root
   ENV VENV /opt/venv
   ENV LANG C.UTF-8
   ENV LC_ALL C.UTF-8
   ENV PYTHONPATH /root

   # This is necessary for opencv to work
   RUN apt-get update && apt-get install -y libsm6 libxext6 libxrender-dev ffmpeg build-essential

   # Install the AWS cli separately to prevent issues with boto being written over
   RUN pip3 install awscli

   ENV VENV /opt/venv
   # Virtual environment
   RUN python3 -m venv ${VENV}
   ENV PATH="${VENV}/bin:$PATH"

   # Install Python dependencies
   COPY core/requirements.txt /root
   RUN pip install -r /root/requirements.txt

   # Copy the makefile targets to expose on the container. This makes it easier to register
   COPY in_container.mk /root/Makefile
   COPY core/sandbox.config /root

   # Copy the actual code
   COPY core /root/core

   # This tag is supplied by the build script and will be used to determine the version
   # when registering tasks, workflows, and launch plans
   ARG tag
   ENV FLYTE_INTERNAL_IMAGE $tag

.. note::
   ``core`` is the directory being considered in the above Dockerfile.

Serialize your workflows and tasks
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Getting your tasks, workflows, and launch plans to run on a Flyte platform is effectively a two-step process.  Serialization is the first step of that process. It is the translation of all your Flyte entities defined in Python, into Flyte IDL entities, defined in protobuf.

Once you've built a Docker container image with your updated code changes, you can use the predefined make target to easily serialize your tasks:

.. code-block::

   make serialize

This runs the `pyflyte serialize` command to convert your workflow and task definitions to registerable protos.
The make target is a handy wrapper around the following:

.. code-block::

   pyflyte -c sandbox.config --pkgs core serialize --in-container-config-path /root/sandbox.config --local-source-root ${CURDIR} --image ${FULL_IMAGE_NAME}:${VERSION} workflows -f _pb_output/

- the :code:`-c` is the path to the config definition on your machine. This config specifies SDK default attributes.
- the :code:`--pkgs` arg points to the packages within the
- :code:`--local-source-root` which contains the code copied over into your Docker container image that will be serialized (and later, executed)
- :code:`--in-container-config-path` maps to the location within your Docker container image where the above config file will be copied over too
- :code:`--image` is the non-optional fully qualified name of the container image housing your code

In-container serialization
""""""""""""""""""""""""""
Notice that the commands above are run locally, _not_ inside the container. Strictly speaking, to be rigourous, serialization should be done within the container for the following reasons.

1. It ensures that the versions of all libraries used at execution time on the Flyte platform, are the same that are used during serialization.
2. Since serialization runs part of flytekit, it helps ensure that your container is set up correctly.

Take a look at this make target to see how it's done.
.. code-block::

   make serialize

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

- :code:`-p` specifies the project to register your entities. This project itself must already be registered on your Flyte deployment.
- :code:`-d` specifies the domain to register your entities. This domain must already be configured in your Flyte deployment
- :code:`-v` is a unique string used to identify this version of entities registered under a project and domain.
- If required, you can specify a :code:`kubernetes-service-account` or :code:`assumable_iam_role` which your tasks will run with.


Fast(er) iteration
^^^^^^^^^^^^^^^^^^
Re-building a new Docker container image for every code change you make can become cumbersome and slow.
If you're making purely code changes that **do not** require updating your container definition, you can make use of
fast serialization and registration to speed up your iteration process and reduce the time it takes to upload new entity
versions and development code to your hosted Flyte deployment. 

First, run the fast serialization target:

.. code-block::

   make fast_serialize

And then the fast register target:

.. code-block::

   OUTPUT_DATA_PREFIX=s3://my-s3-bucket/raw_data FLYTE_HOST=flyte.example.com ADDL_DISTRIBUTION_DIR=s3://my-s3-bucket/archives make register

and just like that you can update your code without requiring a rebuild of your container!

As fast registration serializes code from your local workstation and uploads it to the hosted flyte deployment, make sure to specify the following arguments correctly to ensure that the changes are picked up when the workflow is run.

- :code:`pyflyte serialize` has a :code:`--local-source-root` option which specifies which code is uploaded during the fast registration step. This ensures that the files you want to modify are serialized. This is optional and should be used when your code lies outside of your current working directory.
- :code:`flyte-cli fast-register-files` has a :code:`--dest-dir` option which specifies which folder (in the container) the fast serialization will dump the code in at execution time. This ensures that the running workflow loads the code changes that were uploaded via :code:`fast registration`.


Building Images
^^^^^^^^^^^^^^^
If you are just iterating locally, there is no need to push your Docker image. For Docker for Desktop at least, locally built images will be available for use in its K8s cluster.

If you would like to later push your image to a registry (Dockerhub, ECR, etc.), you can run,

```bash
REGISTRY=docker.io/corp make all_docker_push
```

.. _working_hosted_service:

Some concepts available remote only
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Using remote Flyte gives you the ability to:

- Use caching to avoid calling the same task with the same inputs (for the same version)
- Portability: You can reference pre-registered entities under any domain or project within your workflow code
- Sharable executions: you can easily share links to your executions with your teammates
