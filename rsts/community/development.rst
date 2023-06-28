Development Environment Setup Guide
===================================================================

This guide provides a step-by-step approach to setting up a local development environment for `flyteidl <https://github.com/flyteorg/flyteidl>`_, `flyteadmin <https://github.com/flyteorg/flyteadmin>`_, `flyteplugins <https://github.com/flyteorg/flyteplugins>`_, `flytepropeller <https://github.com/flyteorg/flytepropeller>`_, `flytekit <https://github.com/flyteorg/flytekit>`_ and `flyteconsole <https://github.com/flyteorg/flyteconsole>`_.

Requirements
------------

This guide has been tested and used on AWS EC2 with an Ubuntu 22.04
image. The following tools are required:

- `Docker <https://docs.docker.com/install/>`__
- `Kubectl <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`__
- `Go <https://golang.org/doc/install>`__

Content
-------

-  `How to set dev environment for flyteidl, flyteadmin, flyteplugins,
   flytepropeller? <#how-to-set-dev-environment-for-flyteidl-flyteadmin-flyteplugins-flytepropeller>`__

-  `How to set dev environment for
   flytekit? <#how-to-set-dev-environment-for-flytekit>`__

   -  `Run workflow locally <#run-workflow-locally>`__
   -  `Run workflow in sandbox <#run-workflow-in-sandbox>`__

-  `How to set dev environment for
   flyteconsole? <#how-to-set-dev-environment-for-flyteconsole>`__

How to set dev environment for flyteidl, flyteadmin, flyteplugins, flytepropeller?
----------------------------------------------------------------------------------

1. install `Flytectl <https://github.com/flyteorg/flytectl>`__, a portable and lightweight command-line interface to work with Flyte.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: shell

   # Step1: Install the latest version of Flytectl
   curl -sL https://ctl.flyte.org/install | bash
   # flyteorg/flytectl info checking GitHub for latest tag
   # flyteorg/flytectl info found version: 0.6.39 for v0.6.39/Linux/x86_64
   # flyteorg/flytectl info installed ./bin/flytectl

   # Step2: Export Flytectl path based on the previous log "flyteorg/flytectl info installed ./bin/flytectl"
   export PATH=$PATH:/home/ubuntu/bin # replace with your path

2. Build a k3s cluster that runs Minio and Postgres pods.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

| `minio <https://min.io/>`__ is an S3-compatible object store that will
  be used later to store task output, input, etc.
| `postgres <https://www.postgresql.org/>`__ is an open-source
  object-relational database that will later be used by flyteadmin to
  store all Flyte information.

.. code:: shell

   # Step1: Start k3s cluster, create pods for Postgres and Minio. Note: We cannot access Flyte UI yet! but we can access the Minio console now.
   flytectl demo start --dev
   # 👨‍💻 Flyte is ready! Flyte UI is available at http://localhost:30080/console 🚀 🚀 🎉 
   # ❇️ Run the following command to export demo environment variables for accessing flytectl
   #         export FLYTECTL_CONFIG=/home/ubuntu/.flyte/config-sandbox.yaml 
   # 🐋 Flyte sandbox ships with a Docker registry. Tag and push custom workflow images to localhost:30000
   # 📂 The Minio API is hosted on localhost:30002. Use http://localhost:30080/minio/login for Minio console

   # Step2: Export FLYTECTL_CONFIG as the previous log indicated.
   FLYTECTL_CONFIG=/home/ubuntu/.flyte/config-sandbox.yaml

   # Step3: The kubeconfig will be automatically copied to the user's main kubeconfig (default is `~/.kube/config`) with "flyte-sandbox" as the context name.
   # Check that we can access the K3s cluster. Verify that Postgres and Minio are running.
   kubectl get pod -n flyte
   # NAME                                                  READY   STATUS    RESTARTS   AGE
   # flyte-sandbox-docker-registry-85745c899d-dns8q        1/1     Running   0          5m
   # flyte-sandbox-kubernetes-dashboard-6757db879c-wl4wd   1/1     Running   0          5m
   # flyte-sandbox-proxy-d95874857-2wc5n                   1/1     Running   0          5m
   # flyte-sandbox-minio-645c8ddf7c-sp6cc                  1/1     Running   0          5m
   # flyte-sandbox-postgresql-0                            1/1     Running   0          5m

3. [Optional] Access the Minio console via http://localhost:30080/minio/login.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The default Username is ``minio`` and the default Password is
``miniostorage``. You might need to look at input.pb, output.pb or
deck.html, etc in Minio when you are developing.

4. Run all backends(flyteidl, flyteadmin, flyteplugins, flytepropeller) and HTTP Server in a single binary.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: shell

   # Step1: Download flyte repo
   git clone https://github.com/flyteorg/flyte.git
   cd flyte

   # Step2: Build a single binary that bundles all the backends (flyteidl, flyteadmin, flyteplugins, flytepropeller) and HTTP Server.
   # The versions of flyteidl, flyteadmin, flyteplugins, and flytepropeller used to build the single binary are defined in `go.mod`.
   sudo apt-get -y install jq # You may need to install jq
   go mod tidy
   sudo make compile

   # Step3: Running the single binary. `flyte_local.yaml` is the config file. It is written to fit all your previous builds. So, you don't need to change `flyte_local.yaml`.
   # Note: Replace `flyte_local.yaml` with file in this PR:https://github.com/flyteorg/flyte/pull/3808. Once it is merged, there is no need to change.
   # Note: You may encounter an error due to database `flyteadmin` does not exists. Run the command again will solve the problem.
   flyte start --config flyte_local.yaml
   # All logs from flyteadmin, flyteplugins, flytepropeller, etc. will appear in the terminal.

5. [Optional] Access the Flyte UI at http://localhost:30080/console.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

6. Build single binary with your own code.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The following instructions assume that you’ll change flyteidl,
flyteadmin, flyteplugins, and flytepropeller simultaneously (features
that involve multiple components). If you don’t need to change some
components, simply ignore the instruction for that component.

.. code:: shell

   # Step1: Modify the source code for flyteidl, flyteadmin, flyteplugins, and flytepropeller.

   # Step2: Flyteidl, flyteadmin, flyteplugins, and flytepropeller use go1.19, so make sure to switch to go1.19.
   export PATH=$PATH:$(go env GOPATH)/bin
   go install golang.org/dl/go1.19@latest
   go1.19 download
   export GOROOT=$(go1.19 env GOROOT)
   export PATH="$GOROOT/bin:$PATH"


   # Step3.1: In the flyteidl folder, before building the single binary, you should run:
   make lint
   make generate

   # Step3.2: In the flyteadmin folder, before building the single binary, you should run:
   go mod edit -replace github.com/flyteorg/flytepropeller=/home/ubuntu/flytepropeller #replace with your own local path to flytepropeller
   go mod edit -replace github.com/flyteorg/flyteidl=/home/ubuntu/flyteidl #replace with your own local path to flyteidl
   go mod edit -replace github.com/flyteorg/flyteplugins=/home/ubuntu/flyteplugins # replace with your own local path to flyteplugins
   make lint
   make generate
   make test_unit

   # Step3.3: In the flyteplugins folder, before building the single binary, you should run:
   go mod edit -replace github.com/flyteorg/flyteidl=/home/ubuntu/flyteidl #replace with your own local path to flyteidl

   # Step3.4: In the flytepropeller folder, before building the single binary, you should run:
   go mod edit -replace github.com/flyteorg/flyteidl=/home/ubuntu/flyteidl #replace with your own local path to flyteidl
   go mod edit -replace github.com/flyteorg/flyteplugins=/home/ubuntu/flyteplugins # replace with your own local path to flyteplugins
   make lint
   make generate
   make test_unit

   # Step4: Now, you can build the single binary. In the Flyte folder, run `go mod edit -replace`. This will replace the code with your own.
   go mod edit -replace github.com/flyteorg/flyteadmin=/home/ubuntu/flyteadmin #replace with your own local path to flyteadmin
   go mod edit -replace github.com/flyteorg/flytepropeller=/home/ubuntu/flytepropeller #replace with your own local path to flytepropeller
   go mod edit -replace github.com/flyteorg/flyteidl=/home/ubuntu/flyteidl #replace with your own local path to flyteidl
   go mod edit -replace github.com/flyteorg/flyteplugins=/home/ubuntu/flyteplugins # replace with your own local path to flyteplugins

   # Step5: Rebuild and rerun the single binary based on your own code.
   go mod tidy
   sudo make compile
   flyte start --config flyte_local.yaml

7. Test it by running a Hello World workflow.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: shell

   # Step1: Install flytekit
   pip install flytekit && export PATH=$PATH:~/.local/bin

   # Step2: The flytesnacks repository provides a lot of useful examples.
   git clone https://github.com/flyteorg/flytesnacks && cd flytesnacks/cookbook

   # Step3: Before running the Hello World workflow, create the flytesnacks-development namespace. 
   # This is necessary because, by default (without creating a new project), task pods will run in the flytesnacks-development namespace.
   kubectl create namespace flytesnacks-development

   # Step4: Run a Hello World example
   pyflyte run --remote core/flyte_basics/hello_world.py my_wf
   # Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/fd63f88a55fed4bba846 to see execution in the console.

8. Tear down the k3s cluster After finishing developing.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: shell

   flytectl demo teardown
   # context removed for "flyte-sandbox".
   # 🧹 🧹 Sandbox cluster is removed successfully.
   # ❇️ Run the following command to unset sandbox environment variables for accessing flytectl
   #        unset FLYTECTL_CONFIG 

How to set dev environment for flytekit?
----------------------------------------

1. Set up local Flyte Cluster.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If you are modifying the code for flyteidl, flyteadmin, flyteplugins, or
flytepropeller, you can refer to `How to set up a development
environment for flyteidl, flyteadmin, flyteplugins, and
flytepropeller? <#how-to-set-dev-environment-for-flyteidl-flyteadmin-flyteplugins-flytepropeller>`__
to build the backends.

If not, we can start backends with a single command.

.. code:: shell

   # Step1: Install the latest version of Flytectl, a portable and lightweight command-line interface to work with Flyte.
   curl -sL https://ctl.flyte.org/install | bash
   # flyteorg/flytectl info checking GitHub for latest tag
   # flyteorg/flytectl info found version: 0.6.39 for v0.6.39/Linux/x86_64
   # flyteorg/flytectl info installed ./bin/flytectl

   # Step2: Export Flytectl path based on the previous log "flyteorg/flytectl info installed ./bin/flytectl"
   export PATH=$PATH:/home/ubuntu/bin # replace with your path

   # Step3: Create backends. This will set up a k3s cluster running Minio, Postgres pods, and all Flyte components: flyteadmin, flyteplugins, flytepropeller, etc.
   flytectl demo start
   # 👨‍💻 Flyte is ready! Flyte UI is available at http://localhost:30080/console 🚀 🚀 🎉 
   # ❇️ Run the following command to export demo environment variables for accessing flytectl
   #         export FLYTECTL_CONFIG=/home/ubuntu/.flyte/config-sandbox.yaml 
   # 🐋 Flyte sandbox ships with a Docker registry. Tag and push custom workflow images to localhost:30000
   # 📂 The Minio API is hosted on localhost:30002. Use http://localhost:30080/minio/login for Minio console

2. Run workflow locally.
~~~~~~~~~~~~~~~~~~~~~~~

.. code:: shell

   # Step1: Build a virtual environment for developing Flytekit. This will allow your local changes to take effect when the same Python interpreter runs `import flytekit`.
   git clone https://github.com/flyteorg/flytekit.git # replace with your own repo
   cd flytekit
   virtualenv ~/.virtualenvs/flytekit
   source ~/.virtualenvs/flytekit/bin/activate
   make setup
   pip install -e .
   pip install gsutil awscli
   # If you are also developing the plugins, execute the following:
   cd plugins
   pip install -e .

   # Step2: Modify the source code for flytekit, then run unit tests and lint.
   make lint
   make test

   # Step3: Run a hello world sample to test locally
   git clone https://github.com/flyteorg/flytesnacks
   cd flytesnacks/cookbook
   python3 core/flyte_basics/hello_world.py
   # Running my_wf() hello world

3. Run workflow in sandbox.
~~~~~~~~~~~~~~~~~~~~~~~~~~

| Before running a workflow in the sandbox, make sure you can run it
  locally.
| To run the workflow in the sandbox, we need to build the flytekit
  image. The following Dockerfile is the minimum setting required to run
  a task.
| You can refer to how the `officail flitekit
  image <https://github.com/flyteorg/flytekit/blob/master/Dockerfile>`__
  is built to add more components (like plugins) if needed.
| Please create the following Dockerfile in your flytekit folder.

.. code:: dockerfile

   FROM python:3.9-slim-buster
   USER root
   WORKDIR /root
   ENV PYTHONPATH /root
   RUN apt-get update && apt-get install build-essential -y
   RUN apt-get install git -y
   RUN pip install -U git+https://github.com/Yicheng-Lu-llll/flytekit.git@demo
   ENV FLYTE_INTERNAL_IMAGE "localhost:30000/flytekit:demo"

The instructions below explain how to build the image, push the image to
the Flyte Cluster, and finally submit the workflow to the Flyte Cluster.

.. code:: shell

   # Step1: Ensure you have pushed your changes to the remote repo
   # In the flytekit folder
   git add . && git commit -s -m "develop" && git push

   # Step2: Build the image
   # In the flytekit folder
   export FLYTE_INTERNAL_IMAGE="localhost:30000/flytekit:demo"
   docker build --no-cache -t  "${FLYTE_INTERNAL_IMAGE}" -f ./Dockerfile .

   # Step3: Push the image to the Flyte Cluster
   docker push ${FLYTE_INTERNAL_IMAGE}

   # Step4: Submit a hello world workflow to the Flyte Cluster
   git clone https://github.com/flyteorg/flytesnacks
   cd flytesnacks/cookbook
   # Note create the flytesnacks-development namespace if not exists: 
   # This is necessary because, by default (without creating a new project), task pods will run in the flytesnacks-development namespace.
   # kubectl create namespace flytesnacks-development
   pyflyte run --image ${FLYTE_INTERNAL_IMAGE} --remote core/flyte_basics/hello_world.py  my_wf
   # Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/f5c17e1b5640c4336bf8 to see execution in the console.

How to set dev environment for flyteconsole?
--------------------------------------------

1. Set up local Flyte Cluster.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Refer to `How to Set Up a Dev Environment for Flytekit? <#how-to-set-dev-environment-for-flytekit>`__ or `How to Set Up a Development Environment for Flyteidl, Flyteadmin, Flyteplugins, and Flytepropeller? <#how-to-set-dev-environment-for-flyteidl-flyteadmin-flyteplugins-flytepropeller>`__ to start the backend.

2. Start Flyteconsole.
~~~~~~~~~~~~~~~~~~~~~~

.. code:: shell

   # Step1: Clone the repo and navigate to the Flyteconsole folder
   git clone https://github.com/flyteorg/flyteconsole.git
   cd flyteconsole


   # Step2: Install Node.js 18. Refer to https://github.com/nodesource/distributions/blob/master/README.md#using-ubuntu-2.
   curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - &&\
   sudo apt-get install -y nodejs

   # Step3: Install yarn. Refer to https://classic.yarnpkg.com/lang/en/docs/install/#debian-stable.
   curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
   echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
   sudo apt update && sudo apt install yarn

   # Step4: Add environment variables
   export BASE_URL=/console
   export ADMIN_API_URL=http://localhost:30080
   export DISABLE_AUTH=1
   export ADMIN_API_USE_SSL="http"

   # Step5: Generate SSL certificate
   # Note, since we will use HTTP, SSL is not required. However, missing an SSL certificate will cause an error when starting Flyteconsole.
   make generate_ssl

   # Step6: Install node packages
   yarn install
   yarn build:types # It is fine if seeing error `Property 'at' does not exist on type 'string[]'`
   yarn run build:prod

   # Step7: Start Flyteconsole
   yarn start

3: Final Step: Install the Chrome plugin: `Moesif Origin & CORS Changer <https://chrome.google.com/webstore/detail/moesif-origin-cors-change/digfbfaphojjndkpccljibejjbppifbc>`__.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

We need to disable
`CORS <https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS>`__ to
load resources.

::

   1. Activate plugin (toggle to "on")
   2. Open 'Advanced Settings':
   3. set Access-Control-Allow-Credentials: true

4: Go to http://localhost:3000/console/.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
