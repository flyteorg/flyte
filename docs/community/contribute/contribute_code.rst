#################
Contributing code
#################

.. _component_reference:

🧱 Component reference
======================

To understand how the below components interact with each other, refer to :ref:`Understand the lifecycle of a workflow <workflow-lifecycle>`.

.. note::
    With the exception of ``flytekit``, the below components are maintained in the `flyte <https://github.com/flyteorg/flyte>`__ monorepo.

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/contribution_guide/dependency_graph.png
    :alt: Dependency graph between various flyteorg repos
    :align: center
    :figclass: align-center

    The dependency graph between various flyte repos

``flyte``
*********

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyte>`__
    * - **Purpose**: Deployment, Documentation, and Issues
    * - **Languages**: RST

``flyteidl``
************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyteidl>`__
    * - **Purpose**: Flyte workflow specification is in `protocol buffers <https://developers.google.com/protocol-buffers>`__ which forms the core of Flyte
    * - **Language**: Protobuf
    * - **Guidelines**: Refer to the `README <https://github.com/flyteorg/flyteidl#generate-code-from-protobuf>`__

``flytepropeller``
******************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytepropeller>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flytepropeller>`__
    * - **Purpose**: Kubernetes-native operator
    * - **Language**: Go
    * - **Guidelines:**

        * Check for Makefile in the root repo
        * Run the following commands:
           * ``make generate``
           * ``make test_unit``
           * ``make lint``
        * To compile, run ``make compile``

``flyteadmin``
**************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyteadmin>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flyteadmin>`__
    * - **Purpose**: Control Plane
    * - **Language**: Go
    * - **Guidelines**:

        * Check for Makefile in the root repo
        * If the service code has to be tested, run it locally:
            * ``make compile``
            * ``make server``
        * To seed data locally:
            * ``make compile``
            * ``make seed_projects``
            * ``make migrate``
        * To run integration tests locally:
            * ``make integration``
            * (or to run in containerized dockernetes): ``make k8s_integration``

``flytekit``
************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytekit>`__
    * - **Purpose**: Python SDK & Tools
    * - **Language**: Python
    * - **Guidelines**: Refer to the `Flytekit Contribution Guide <https://docs.flyte.org/en/latest/api/flytekit/contributing.html>`__

``flyteconsole``
****************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyteconsole>`__
    * - **Purpose**: Admin Console
    * - **Language**: Typescript
    * - **Guidelines**: Refer to the `README <https://github.com/flyteorg/flyteconsole/blob/master/README.md>`__

``datacatalog``
***************

.. list-table::

    * - `Repo <https://github.com/flyteorg/datacatalog>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/datacatalog>`__
    * - **Purpose**: Manage Input & Output Artifacts
    * - **Language**: Go

``flyteplugins``
****************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyteplugins>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flyteplugins>`__
    * - **Purpose**: Flyte Plugins
    * - **Language**: Go
    * - **Guidelines**:

        * Check for Makefile in the root repo
        * Run the following commands:
            * ``make generate``
            * ``make test_unit``
            * ``make lint``

``flytestdlib``
***************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytestdlib>`__
    * - **Purpose**: Standard Library for Shared Components
    * - **Language**: Go

``flytectl``
************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytectl>`__
    * - **Purpose**: A standalone Flyte CLI
    * - **Language**: Go
    * - **Guidelines**: Refer to the `FlyteCTL Contribution Guide <https://docs.flyte.org/en/latest/flytectl/contribute.html>`__


🔮 Development Environment Setup Guide
======================================

This guide provides a step-by-step approach to setting up a local development environment for
`flyteidl <https://github.com/flyteorg/flyteidl>`_, `flyteadmin <https://github.com/flyteorg/flyteadmin>`_,
`flyteplugins <https://github.com/flyteorg/flyteplugins>`_, `flytepropeller <https://github.com/flyteorg/flytepropeller>`_,
`flytekit <https://github.com/flyteorg/flytekit>`_ , `flyteconsole <https://github.com/flyteorg/flyteconsole>`_,
`datacatalog <https://github.com/flyteorg/datacatalog>`_, and `flytestdlib <https://github.com/flyteorg/flytestdlib>`_.

The video below is a tutorial on how to set up a local development environment for Flyte.

..  youtube:: V-KlVQmQAjE

Requirements
************

This guide has been tested and used on AWS EC2 with an Ubuntu 22.04
image. The following tools are required:

- `Docker <https://docs.docker.com/install/>`__
- `Kubectl <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`__
- `Go <https://golang.org/doc/install>`__

Content
*******

-  `How to setup dev environment for flyteidl, flyteadmin, flyteplugins,
   flytepropeller, datacatalog and flytestdlib? <#how-to-setup-dev-environment-for-flyteidl-flyteadmin-flyteplugins-flytepropeller-datacatalog-and-flytestdlib>`__

-  `How to setup dev environment for
   flytekit? <#how-to-setup-dev-environment-for-flytekit>`__

-  `How to setup dev environment for
   flyteconsole? <#how-to-setup-dev-environment-for-flyteconsole>`__

-  `How to access Flyte UI, minio, postgres, k3s, and endpoints?
   <#how-to-access-flyte-ui-minio-postgres-k3s-and-endpoints>`__

How to setup dev environment for flyteidl, flyteadmin, flyteplugins, flytepropeller, datacatalog and flytestdlib?
******************************************************************************************************************************

**1. Install flytectl**


`Flytectl <https://github.com/flyteorg/flytectl>`__ is a portable and lightweight command-line interface to work with Flyte.

.. code:: shell

   # Step 1: Install the latest version of flytectl
   curl -sL https://ctl.flyte.org/install | bash
   # flyteorg/flytectl info checking GitHub for latest tag
   # flyteorg/flytectl info found version: 0.6.39 for v0.6.39/Linux/x86_64
   # flyteorg/flytectl info installed ./bin/flytectl

   # Step 2: Export flytectl path based on the previous log "flyteorg/flytectl info installed ./bin/flytectl"
   export PATH=$PATH:/home/ubuntu/bin # replace with your path

**2. Build a k3s cluster that runs minio and postgres Pods.**


| `Minio <https://min.io/>`__ is an S3-compatible object store that will be used later to store task output, input, etc.
| `Postgres <https://www.postgresql.org/>`__ is an open-source object-relational database that will later be used by flyteadmin/dataCatalog to
  store all Flyte information.

.. code:: shell

   # Step 1: Start k3s cluster, create Pods for postgres and minio. Note: We cannot access Flyte UI yet! but we can access the minio console now.
   flytectl demo start --dev
   # 👨‍💻 Flyte is ready! Flyte UI is available at http://localhost:30080/console 🚀 🚀 🎉
   # ❇️ Run the following command to export demo environment variables for accessing flytectl
   #         export FLYTECTL_CONFIG=/home/ubuntu/.flyte/config-sandbox.yaml
   # 🐋 Flyte sandbox ships with a Docker registry. Tag and push custom workflow images to localhost:30000
   # 📂 The Minio API is hosted on localhost:30002. Use http://localhost:30080/minio/login for Minio console

   # Step 2: Export FLYTECTL_CONFIG as the previous log indicated.
   FLYTECTL_CONFIG=/home/ubuntu/.flyte/config-sandbox.yaml

   # Step 3: The kubeconfig will be automatically copied to the user's main kubeconfig (default is `/.kube/config`) with "flyte-sandbox" as the context name.
   # Check that we can access the K3s cluster. Verify that postgres and minio are running.
   kubectl get pod -n flyte
   # NAME                                                  READY   STATUS    RESTARTS   AGE
   # flyte-sandbox-docker-registry-85745c899d-dns8q        1/1     Running   0          5m
   # flyte-sandbox-kubernetes-dashboard-6757db879c-wl4wd   1/1     Running   0          5m
   # flyte-sandbox-proxy-d95874857-2wc5n                   1/1     Running   0          5m
   # flyte-sandbox-minio-645c8ddf7c-sp6cc                  1/1     Running   0          5m
   # flyte-sandbox-postgresql-0                            1/1     Running   0          5m


**3. Run all Flyte components (flyteadmin, flytepropeller, datacatalog, flyteconsole, etc) in a single binary.**

The `Flyte repository <https://github.com/flyteorg/flyte>`__ includes Go code
that integrates all Flyte components into a single binary.

.. code:: shell

   # Step 1: Clone flyte repo
   git clone https://github.com/flyteorg/flyte.git
   cd flyte

   # Step 2: Build a single binary that bundles all the Flyte components.
   # The version of each component/library used to build the single binary are defined in `go.mod`.
   sudo apt-get -y install jq # You may need to install jq
   make clean # (Optional) Run this only if you want to run the newest version of flyteconsole
   make go-tidy
   make compile

   # Step 3: Prepare a namespace template for the cluster resource controller.
   # The configuration file "flyte-single-binary-local.yaml" has an entry named cluster_resources.templatePath.
   # This entry needs to direct to a directory containing the templates for the cluster resource controller to use.
   # We will now create a simple template that allows the automatic creation of required namespaces for projects.
   # For example, with Flyte's default project "flytesnacks", the controller will auto-create the following namespaces:
   # flytesnacks-staging, flytesnacks-development, and flytesnacks-production.
   mkdir $HOME/.flyte/sandbox/cluster-resource-templates/
   echo "apiVersion: v1
   kind: Namespace
   metadata:
     name: '{{ namespace }}'" > $HOME/.flyte/sandbox/cluster-resource-templates/namespace.yaml

   # Step 4: Running the single binary.
   # The POD_NAMESPACE environment variable is necessary for the webhook to function correctly.
   # You may encounter an error due to `ERROR: duplicate key value violates unique constraint`. Running the command again will solve the problem.
   POD_NAMESPACE=flyte flyte start --config flyte-single-binary-local.yaml
   # All logs from flyteadmin, flyteplugins, flytepropeller, etc. will appear in the terminal.


**4. Build single binary with your own code.**


The following instructions provide guidance on how to build single binary with your customized code under the ``flyteadmin`` as an example.


- **Note** Although we'll use ``flyteadmin`` as an example, these steps can be applied to other Flyte components or libraries as well.
  ``{flyteadmin}`` below can be substituted with other Flyte components/libraries: ``flyteidl``, ``flyteplugins``, ``flytepropeller``, ``datacatalog``, or ``flytestdlib``.
- **Note** If you want to learn how flyte compiles those components and replace the repositories, you can study how ``go mod edit`` works.

.. code:: shell

   # Step 1: Install Go. Flyte uses Go 1.19, so make sure to switch to Go 1.19.
   export PATH=$PATH:$(go env GOPATH)/bin
   go install golang.org/dl/go1.19@latest
   go1.19 download
   export GOROOT=$(go1.19 env GOROOT)
   export PATH="$GOROOT/bin:$PATH"

   # You may need to install goimports to fix lint errors.
   # Refer to https://pkg.go.dev/golang.org/x/tools/cmd/goimports
   go install golang.org/x/tools/cmd/goimports@latest
   export PATH=$(go env GOPATH)/bin:$PATH

   # Step 2: Go to the {flyteadmin} repository, modify the source code accordingly.
   cd flyte/flyteadmin

   # Step 3: Now, you can build the single binary. Go back to Flyte directory.
   make go-tidy
   make compile
   POD_NAMESPACE=flyte flyte start --config flyte-single-binary-local.yaml

**5. Test by running a hello world workflow.**


.. code:: shell

   # Step 1: Install flytekit
   pip install flytekit && export PATH=$PATH:/home/ubuntu/.local/bin

   # Step 2: Run a hello world example
   pyflyte run --remote https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/basics/basics/hello_world.py  hello_world_wf
   # Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/fd63f88a55fed4bba846 to see execution in the console.
   # You can go to the [flytesnacks repository](https://github.com/flyteorg/flytesnacks) to see more useful examples.

**6. Tear down the k3s cluster after finishing developing.**


.. code:: shell

   flytectl demo teardown
   # context removed for "flyte-sandbox".
   # 🧹 🧹 Sandbox cluster is removed successfully.
   # ❇️ Run the following command to unset sandbox environment variables for accessing flytectl
   #        unset FLYTECTL_CONFIG

How to setup dev environment for flytekit?
*******************************************

**1. Set up local Flyte Cluster.**


If you are also modifying the code for flyteidl, flyteadmin, flyteplugins, flytepropeller datacatalog, or flytestdlib,
refer to the instructions in the  `previous section <#how-to-setup-dev-environment-for-flyteidl-flyteadmin-flyteplugins-flytepropeller-datacatalog-and-flytestdlib>`__ to set up a local Flyte cluster.

If not, we can start backends with a single command.

.. code:: shell

   # Step 1: Install the latest version of flytectl, a portable and lightweight command-line interface to work with Flyte.
   curl -sL https://ctl.flyte.org/install | bash
   # flyteorg/flytectl info checking GitHub for latest tag
   # flyteorg/flytectl info found version: 0.6.39 for v0.6.39/Linux/x86_64
   # flyteorg/flytectl info installed ./bin/flytectl

   # Step 2: Export flytectl path based on the previous log "flyteorg/flytectl info installed ./bin/flytectl"
   export PATH=$PATH:/home/ubuntu/bin # replace with your path

   # Step 3: Starts the Flyte demo cluster. This will setup a k3s cluster running minio, postgres Pods, and all Flyte components: flyteadmin, flyteplugins, flytepropeller, etc.
   # See https://docs.flyte.org/en/latest/flytectl/gen/flytectl_demo_start.html for more details.
   flytectl demo start
   # 👨‍💻 Flyte is ready! Flyte UI is available at http://localhost:30080/console 🚀 🚀 🎉
   # ❇️ Run the following command to export demo environment variables for accessing flytectl
   #         export FLYTECTL_CONFIG=/home/ubuntu/.flyte/config-sandbox.yaml
   # 🐋 Flyte sandbox ships with a Docker registry. Tag and push custom workflow images to localhost:30000
   # 📂 The Minio API is hosted on localhost:30002. Use http://localhost:30080/minio/login for Minio console

**2. Run workflow locally.**


.. code:: shell

   # Step 1: Build a virtual environment for developing Flytekit. This will allow your local changes to take effect when the same Python interpreter runs `import flytekit`.
   git clone https://github.com/flyteorg/flytekit.git # replace with your own repo
   cd flytekit
   virtualenv ~/.virtualenvs/flytekit
   source ~/.virtualenvs/flytekit/bin/activate
   make setup
   pip install -e .

   # If you are also developing the plugins, consider the following:

   # Installing Specific Plugins:
   # If you wish to only use few plugins, you can install them individually.
   # Take [Flytekit BigQuery Plugin](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-bigquery#flytekit-bigquery-plugin) for example:
   # You have to go to the bigquery plugin folder and install it.
   cd plugins/flytekit-bigquery/
   pip install -e .
   # Now you can use the bigquery plugin, and the performance is fast.

   # (Optional) Installing All Plugins:
   # If you wish to install all available plugins, you can execute the command below.
   # However, it's not typically recommended because the current version of plugins does not support
   # lazy loading. This can lead to a slowdown in the performance of your Python engine.
   cd plugins
   pip install -e .
   # Now you can use all plugins, but the performance is slow.

   # Step 2: Modify the source code for flytekit, then run unit tests and lint.
   make lint
   make test

   # Step 3: Run a hello world sample to test locally
   pyflyte run https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/basics/basics/hello_world.py hello_world_wf
   # Running hello_world_wf() hello world

**3-1. Run workflow in sandbox - with `ImageSpec`**

Before running your workflow in the sandbox, make sure you're able to successfully run it
locally. To run the workflow in the sandbox with the newest modification, you can use
`ImageSpec` to define your container image directly in Python code.

Create a workflow file that uses `ImageSpec` to define the image with your custom flytekit version:

.. code:: python
    from flytekit import ImageSpec, task, workflow

    # Define your custom flytekit version - replace with your own repo and commit hash
    # You need to push your modifications to your fork and get the commit hash first
    new_flytekit = "git+https://github.com/your-github-username/flytekit.git@your-commit-hash"

    # To install your modified plugins, use following:
    # new_deck_plugin = git+https://github.com/your-github-username/flytekit.git@your-commit-hash#subdirectory=plugins/flytekit-kf-pytorch

    # Create ImageSpec with your custom flytekit
    image_spec = ImageSpec(
        registry="localhost:30000",
        packages=[
            new_flytekit,
        ],
        apt_packages=["git"],
    )

    @task(container_image=image_spec)
    def hello_world_task(name: str) -> str:
        return f"Hello {name}!"

    @workflow
    def wf(name: str = "World") -> str:
        return hello_world_task(name=name)


Then submit the workflow to the Flyte cluster:

.. code:: shell
    # ImageSpec will automatically build and push the image to the local registry
    pyflyte run --remote custom_workflow.py wf
    # Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/<execution-id> to see execution in the console.


**3-2. Run workflow in sandbox - with `Dockerfile`**

Before running your workflow in the sandbox, make sure you're able to successfully run it locally.
To deploy the workflow in the sandbox, you'll need to build a Flytekit image.
Create a Dockerfile in your Flytekit directory with the minimum required configuration to run a task, as shown below.
If your task requires additional components, such as plugins, you may find it useful to refer to the construction of the `official flytekit image <https://github.com/flyteorg/flytekit/blob/master/Dockerfile>`__

.. code:: Dockerfile

   FROM python:3.9-slim-buster
   USER root
   WORKDIR /root
   ENV PYTHONPATH /root
   RUN apt-get update && apt-get install build-essential -y
   RUN apt-get install git -y
   # The following line is an example of how to install your modified plugins. In this case, it demonstrates how to install the 'deck' plugin.
   # RUN pip install -U git+https://github.com/Yicheng-Lu-llll/flytekit.git@"demo#egg=flytekitplugins-deck-standard&subdirectory=plugins/flytekit-deck-standard" # replace with your own repo and branch
   RUN pip install -U git+https://github.com/Yicheng-Lu-llll/flytekit.git@demo # replace with your own repo and branch
   ENV FLYTE_INTERNAL_IMAGE "localhost:30000/flytekit:demo" # replace with your own image name and tag

The instructions below explain how to build the image, push the image to
the Flyte cluster, and finally submit the workflow.

.. code:: shell

   # Step 1: Ensure you have pushed your changes to the remote repo
   # In the flytekit folder
   git add . && git commit -s -m "develop" && git push

   # Step 2: Build the image
   # In the flytekit folder
   export FLYTE_INTERNAL_IMAGE="localhost:30000/flytekit:demo" # replace with your own image name and tag
   docker build --no-cache -t  "${FLYTE_INTERNAL_IMAGE}" -f ./Dockerfile .

   # Step 3: Push the image to the Flyte cluster
   docker push ${FLYTE_INTERNAL_IMAGE}

   # Step 4: Submit a hello world workflow to the Flyte cluster
   cd flytesnacks
   pyflyte run --image ${FLYTE_INTERNAL_IMAGE} --remote https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/basics/basics/hello_world.py hello_world_wf
   # Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/f5c17e1b5640c4336bf8 to see execution in the console.


How to debug flytekit remote workflow?
**************************************

For running locally, we can easily debug with the help of debugger. However,
this approach does not work if we want to debug the remote workflow. When you
need to debug your flytekit changes while running a remote workflow and set
breakpoints to inspect variables, you can follow these steps:

**1. Add breakpoints to your flytekit code**

Insert `breakpoint()` function calls at the locations in your flytekit modifications where
you want to pause execution and inspect variables. This is particularly useful when
debugging changes you've made to flytekit's core functionality:


**2. Run your workflow remotely**

Execute your workflow using the remote flag:

.. code:: shell
    pyflyte run --remote custom_workflow.py wf

**3. Describe the pod to get container arguments**

First, find the pod name from your execution. You can get this from the Flyte UI execution page or by listing pods:

.. code:: shell
    # List pods to find your execution pod
    kubectl get pods -n flytesnacks-development

    # Look for pods with names matching your execution ID
    # Example output:
    # NAME                           READY   STATUS    RESTARTS   AGE
    # ab5mg9lzgth62h82qprp-n0-0     1/1     Running   0          2m

Then describe the pod to get the container arguments:

.. code:: shell
    # Replace with your actual pod name from the execution
    kubectl describe pod -n flytesnacks-development ab5mg9lzgth62h82qprp-n0-0

Look for the `Args:` section in the container specification and copy all the arguments. Example output:

.. code:: shell
    Args:
      pyflyte-fast-execute
      --additional-distribution
      /opt/venv
      --dest-dir
      /tmp/flyte
      --input
      s3://my-bucket/metadata/...

**Step 4: Set up environment variables**

Export the necessary Minio environment variables to access the object store if you are using demo cluster:

.. code:: shell
    export FLYTE_AWS_ENDPOINT="http://localhost:30002"
    export FLYTE_AWS_ACCESS_KEY_ID="minio" 
    export FLYTE_AWS_SECRET_ACCESS_KEY="miniostorage"

**Step 5: Run the container arguments locally**

Take the container arguments from step 3, combine them into a single line, and execute
them in your terminal. This will run the task locally with the same configuration as the
remote execution, allowing you to hit your breakpoints and inspect variables.

.. code:: shell
    # Example: Convert the multi-line args to a single command
    pyflyte-fast-execute --additional-distribution /opt/venv --dest-dir /tmp/flyte --input s3://my-bucket/metadata/... --output-prefix s3://my-bucket/data/...

This approach allows you to debug your flytekit changes in a remote workflow execution
locally while maintaining the same execution context and data access patterns. This is
especially valuable when you need to step through your modifications to flytekit's
internals during remote execution.



How to setup dev environment for flyteconsole?
**********************************************

**1. Set up local Flyte cluster.**

Depending on your needs, refer to one of the following guides to setup up the Flyte cluster:

- If you do not need to change the backend code, refer to the section on `How to Set Up a Dev Environment for Flytekit? <#how-to-setup-dev-environment-for-flytekit>`__
- If you need to change the backend code, refer to the section on `How to setup dev environment for flyteidl, flyteadmin, flyteplugins, flytepropeller, datacatalog and flytestdlib? <#how-to-setup-dev-environment-for-flyteidl-flyteadmin-flyteplugins-flytepropeller-datacatalog-and-flytestdlib>`__


**2. Start flyteconsole.**


.. code:: shell

   # Step 1: Clone the repo and navigate to the Flyteconsole folder
   git clone https://github.com/flyteorg/flyteconsole.git
   cd flyteconsole

   # Step 2: Install Node.js 18. Refer to https://github.com/nodesource/distributions/blob/master/README.md#using-ubuntu-2.
   curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - &&\
   sudo apt-get install -y nodejs

   # Step 3: Install yarn. Refer to https://classic.yarnpkg.com/lang/en/docs/install/#debian-stable.
   curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
   echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
   sudo apt update && sudo apt install yarn

   # Step 4: Add environment variables
   export BASE_URL=/console
   export ADMIN_API_URL=http://localhost:30080
   export DISABLE_AUTH=1
   export ADMIN_API_USE_SSL="http"

   # Step 5: Generate SSL certificate
   # Note, since we will use HTTP, SSL is not required. However, missing an SSL certificate will cause an error when starting Flyteconsole.
   make generate_ssl

   # Step 6: Install node packages
   yarn install
   yarn build:types # It is fine if seeing error `Property 'at' does not exist on type 'string[]'`
   yarn run build:prod

   # Step 7: Start flyteconsole
   yarn start

**3. Install the Chrome plugin:** `Moesif Origin & CORS Changer <https://chrome.google.com/webstore/detail/moesif-origin-cors-change/digfbfaphojjndkpccljibejjbppifbc>`__.


We need to disable `CORS <https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS>`__ to load resources.

::

   1. Activate plugin (toggle to "on")
   2. Open 'Advanced Settings':
   3. set Access-Control-Allow-Credentials: true

**4. Go to** http://localhost:3000/console/.


How to access Flyte UI, minio, postgres, k3s, and endpoints?
*************************************************************************


This section presumes a local Flyte cluster is already setup. If it isn't, refer to either:

- `How to setup dev environment for flytekit? <#how-to-setup-dev-environment-for-flytekit>`__
- `How to setup dev environment for flyteidl, flyteadmin, flyteplugins, flytepropeller, datacatalog and flytestdlib? <#how-to-setup-dev-environment-for-flyteidl-flyteadmin-flyteplugins-flytepropeller-datacatalog-and-flytestdlib>`__


**1. Access the Flyte UI.**


`Flyte UI <https://docs.flyte.org/en/latest/concepts/flyte_console.html>`__ is a web-based user interface for Flyte that lets you interact with Flyte objects and build directed acyclic graphs (DAGs) for your workflows.

You can access it via http://localhost:30080/console.

**2. Access the minio console.**


Core Flyte components, such as admin, propeller, and datacatalog, as well as user runtime containers rely on an object store (in this case, minio) to hold files.
During development, you might need to examine files such as `input.pb/output.pb <https://docs.flyte.org/en/latest/concepts/data_management.html#serialization-time>`__, or `deck.html <https://docs.flyte.org/en/latest/user_guide/development_lifecycle/decks.html#id1>`__ stored in minio.

Access the minio console at: http://localhost:30080/minio/login. The default credentials are:

- Username: ``minio``
- Password: ``miniostorage``


**3. Access the postgres.**


FlyteAdmin and datacatalog use postgres to store persistent records, and you can interact with postgres on port ``30001``. Here is an example of using `psql` to connect:

.. code:: shell

    # Step 1: Install the PostgreSQL client.
    sudo apt-get update
    sudo apt-get install postgresql-client

    # Step 2: Connect to the PostgreSQL server. The password is "postgres".
    psql -h localhost -p 30001 -U postgres -d flyte


**4. Access the k3s dashboard.**


Access the k3s dashboard at: http://localhost:30080/kubernetes-dashboard.

**5. Access the endpoints.**


Service endpoints are defined in the `flyteidl` repository under the `service` directory. You can browse them at `here <https://github.com/flyteorg/flyteidl/tree/master/protos/flyteidl/service>`__.

For example, the endpoint for the `ListTaskExecutions <https://github.com/flyteorg/flyteidl/blob/b219c2ab37886801039fda67d913760ac6fc4c8b/protos/flyteidl/service/admin.proto#L442>`__ API is:

.. code:: shell

   /api/v1/task_executions/{node_execution_id.execution_id.project}/{node_execution_id.execution_id.domain}/{node_execution_id.execution_id.name}/{node_execution_id.node_id}

You can access this endpoint at:

.. code:: shell

   # replace with your specific task execution parameters
   http://localhost:30080/api/v1/task_executions/flytesnacks/development/fe92c0a8cbf684ad19a8/n0?limit=10000

