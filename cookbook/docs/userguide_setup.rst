.. _env_setup:

##################
Environment Setup
##################

Prerequisites
^^^^^^^^^^^^^

* Make sure you have `docker <https://docs.docker.com/get-docker/>`_ and `git <https://git-scm.com/>`_ installed.
* Install :doc:`flytectl <flytectl:index>`, the commandline interface for flyte.

Repo Setup
^^^^^^^^^^

Clone the ``flytesnacks`` repo and install its dependencies, which includes :doc:`flytekit <flytekit:index>` :

.. tip::

    **[Recommended]** Create a new python virtual environment to make sure it doesn't interfere with your
    development environment, e.g. with ``python -m venv <path/to/env>``.

.. prompt:: bash

    git clone https://github.com/flyteorg/flytesnacks
    cd flytesnacks/cookbook
    pip install -r core/requirements.txt

To make sure everything is working in your virtual environment, run ``hello_world.py`` locally:

.. prompt:: bash

    python core/flyte_basics/hello_world.py

Expected output:

.. prompt::

    Running my_wf() hello world

Create a Local Demo Flyte Cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Use ``flytectl`` to start a demo Flyte cluster:

.. prompt:: bash

    flytectl demo start

Running Workflows
^^^^^^^^^^^^^^^^^

Now you can run all of the example workflows locally using the default Docker image bundled with ``flytekit``:

.. prompt:: bash

    pyflyte run core/flyte_basics/hello_world.py my_wf

.. note::

    The first couple arguments of ``pyflyte run`` is in the form of ``path/to/script.py <workflow_name>``, where
    ``<workflow_name>`` is the function decorated with ``@workflow`` that you want to run.

To run the workflow on the demo Flyte cluster, all you need to do is supply the ``--remote`` flag:

.. prompt:: bash

    pyflyte run --remote core/flyte_basics/hello_world.py my_wf

You should see an output that looks like:

.. prompt::

    Go to https://<flyte_admin_url>/console/projects/flytesnacks/domains/development/executions/<execution_name> to see execution in the console.

You can visit this url to inspect the execution as it runs:

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/common/first_run_console.gif
        :alt: A quick visual tour for launching your first Workflow.

Finally, let's run a workflow that takes some inputs, for example the ``basic_workflow.py`` example:

.. prompt:: bash

    pyflyte run --remote core/flyte_basics/basic_workflow.py my_wf --a 5 --b hello

.. note::

    We're passing in the workflow inputs as additional options to ``pyflyte run``, in the above example the
    inputs are ``--a 5`` and ``--b hello``. For snake-case argument names like ``arg_name``, you can provide the
    option as ``--arg-name``.

Visualizing Workflows
^^^^^^^^^^^^^^^^^^^^^

Workflows can be visualized as DAGs on the UI. However, you can visualize workflows on the browser and in the terminal by *just* using your terminal.

To view workflow on the browser:

.. prompt:: bash $

   flytectl get workflows --project flytesnacks --domain development flyte.workflows.example.my_wf --version <version> -o doturl

To view workflow as a ``strict digraph`` on the command line:

.. prompt:: bash $

   flytectl get workflows --project flytesnacks --domain development flyte.workflows.example.my_wf --version <version> -o dot

Replace ``<version>`` with version from console UI, it may look something like ``BLrGKJaYsW2ME1PaoirK1g==``

.. tip::

    Running most of the examples in the **User Guide** only requires the default Docker image that ships with Flyte.
    Many examples in the :ref:`tutorials` and :ref:`integrations` section depend on additional libraries, ``sklearn``,
    ``pytorch``, or ``tensorflow``, which will not work with the default docker image used by ``pyflyte run``.

    These examples will explicitly show you which images to use for running these examples by passing in the docker
    image you want to use with the ``--image`` option in ``pyflyte run``.

🎉 Congrats! Now you can run all the examples in the :ref:`userguide` 🎉

What's Next?
^^^^^^^^^^^^

Try out the examples in :doc:`Flyte Basics <auto/core/flyte_basics/index>` section.
