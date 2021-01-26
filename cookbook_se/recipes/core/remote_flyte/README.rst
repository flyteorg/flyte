Getting setup for remote execution
-------------------------------------
Locally, Flytekit relies on the Python interpreter to execute both tasks and workflows.
To leverage the full power of Flyte, we recommend using a deployed backend of Flyte. Flyte can be run
on any Kubernetes cluster - a local cluster like `kind <https://kind.sigs.k8s.io/>`__, in a cloud environment or on-prem.

Please refer to the `Installing Flyte <https://lyft.github.io/flyte/administrator/install/index.html>`__ for details on getting started with a Flyte installation.


1. First commit your changes. Some of the steps below default to referencing the git sha.
1. Run `make serialize_sandbox`. This will build the image tagged with just `flytecookbook:<sha>`, no registry will be prefixed. See the image building section below for additional information.

Build your Dockerfile
^^^^^^^^^^^^^^^^^^^^^^
The first step of this process is building a container image that holds your code.

.. literalinclude:: ../../Dockerfile
    :language: dockerfile
    :emphasize-lines: 1
    :linenos:


Serialize your workflows and tasks
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Getting your tasks, workflows, and launch plans to run on a Flyte platform is effectively a two-step process.  Serialization is the first step of that process. It is the translation of all your Flyte entities defined in Python, into Flyte IDL entities, defined in protobuf.

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

In-container serialization
""""""""""""""""""""""""""
Notice that the commands above are run locally, _not_ inside the container. Strictly speaking, to be rigourous, serialization should be done within the container for the following reasons.

1. It ensures that the versions of all libraries used at execution time on the Flyte platform, are the same that are used during serialization.
1. Since serialization runs part of flytekit, it helps ensure that your container is set up correctly.

Take a look at this make target to see how it's done.
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


Building Images
^^^^^^^^^^^^^^^
If you are just iterating locally, there is no need to push your Docker image. For Docker for Desktop at least, locally built images will be available for use in its K8s cluster.

If you would like to later push your image to a registry (Dockerhub, ECR, etc.), you can run,

```bash
REGISTRY=docker.io/corp make all_docker_push
```

.. _working_hosted_service:

Some concepts available remote only
-----------------------------------

Using remote Flyte gives you the ability to:

- Use caching to avoid calling the same task with the same inputs (for the same version)
- Portability: You can reference pre-registered entities under any domain or project within your workflow code
- Sharable executions: you can easily share links to your executions with your teammates
