.. _gettingstarted:

Getting started
---------------

.. rubric:: Estimated time to complete: 3 minutes.


Prerequisites
***************

Make sure you have `docker installed <https://docs.docker.com/get-docker/>`__ and `git <https://git-scm.com/>`__ installed.

Steps
*****

#. First install the python `flytekit<https://pypi.org/project/flytekit/>`_ python SDK for Flyte (maybe in a virtual environment) and clone the `flytekit-python-template <https://github.com/flyteorg/flytekit-python-template>`_ repo

    .. prompt::

      pip install flytekit
      git clone git@github.com:flyteorg/flytekit-python-template.git myflyteapp
      cd myflyteapp


#. The repo comes with a sample workflow, which can be found under ``myapp/workflows/example.py``. The structure below shows the most important files and how a typical flyteapp should be laid out.

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

    The workflow can be run locally simply by running it as a python script - ``note the __main__ at the bottom of the file``

    .. prompt::

        python myapp/workflows/example.py

#. Let us install :std:ref:`flytectl`. ``flytectl`` is a commandline interface for flyte.

    .. tabs::

        .. tab:: OSX

            .. prompt::

                brew install flyteorg/homebrew-tap/flytectl

            To upgrade you can

            .. prompt::

                brew upgrade flytectl

        .. tab:: Most other platforms

            .. prompt::

                curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash


#. Test if flytectl is installed correctly::

    flytectl version


#. [Optional] Flyte can be deployed locally using a single docker container - we refer to this as flyte-sandbox. You can skip this step if you already have a Flyte sandbox or a hosted Flyte deployed.

    .. tip:: Want to dive under the hood into flyte-sandbox, refer to the guide `here<>`_.

    .. prompt::

        flytectl sandbox start --source myflyteapp

#. Setup flytectl config using ... doc to configuring flytectl::

    flytectl setup-config

#. Flyte uses docker containers to package your workflows and tasks and send it to the remote Flyte cluster. Thus if you notice there is a ``Dockerfile`` already in the cloned repo. You can build the docker container and push the built image to a registry. Follow the instructions below

    .. tabs::

        .. tab:: If using flyte-sandbox

            Since ``flyte-sandbox`` is running locally in a docker container, you do not really need to push the docker image. You can combine the build and push step, by simply building the image inside the flyte-sandbox container. This can be done using

            .. tip:: Is this confusing? Refer to guide `here<>`

            .. prompt::

                flytectl sandbox exec -- docker build .

        .. tab:: If using remote flyte cluster

            If you are using a remote flyte cluster, then you need to build your container and push it to a registry that is accessible by the Flyte kubernetes cluster.

            .. prompt::

                docker build . --tag registry/repo:version
                docker push registry/repo:version

#. Now that the container is built, lets provide this information to the Flyte backend. To do that you have to package the workflow using the pyflyte cli, that is bundled with flytekit::

    pyflyte package ...

#. Now lets upload this package to flyte backend. We call this process ::

    flytectl register files my_wf.pb

#. You can create an execution using flytectl as follows::

    blah


#. You can use the FlyteConsole to launch an execution and watch the progress.

    .. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/tutorial/exercise.gif
        :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.

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

    .. tip::

      .. code-block:: python

        @task
        def say_hello(name: str) -> str:
            return f"hello world, {name}"

    .. tip::

      .. code-block:: python

        @workflow
        def my_wf(name: str) -> str:
            res = say_hello(name=name)
            return res

#. Update the simple test at the bottom of the file to pass in a name. E.g.

    .. tip::

      .. code-block:: python

        print(f"Running my_wf(name='adam') {my_wf(name='adam')}")

#. When you run this file locally, it should output ``hello world, adam``. Run this command in your terminal:

    .. prompt::

      python myapp/workflows/example.py

    *Congratulations!* You have just run your first workflow. Now, let's run it on the sandbox cluster deployed earlier.


#. To deploy this workflow to the Flyte cluster (sandbox), you can repeat the previous step of docker build -> package -> register. But, since you have not really updated any of the dependencies in your requirements file, it is possible to push just the code to flyte, without really re-building the entire docker container. The docker container that was built previously is enough.

    .. prompt::

      pyflyte package ... --fast

#. You can now deploy the code using flytectl, with an additional argument called --fast

    .. prompt::

        flytectl register --fast

#. Visit `the console <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/core.basic.hello_world.my_wf>`__, click launch, and enter your name as the input.




.. admonition:: Recap

  You have successfully:

  1. Run a flyte sandbox cluster,
  2. Run a flyte workflow locally,
  3. Run a flyte workflow on a cluster.

  .. rubric:: 🎉 Congratulations, you just ran your first Flyte workflow 🎉

Next Steps: User Guide
#######################

To experience the full capabilities of Flyte, take a look at the `User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>`__