Welcome to Flyte! 
=================

Flyte is a structured programming and distributed processing platform that enables highly concurrent, scalable and maintainable workflows for machine learning and data processing.

Getting started
---------------

.. rubric:: Estimated time: 3 minutes

Prerequisites
#############

Make sure you have `docker installed <https://docs.docker.com/get-docker/>`__ and `git <https://git-scm.com/>`__ installed, then install flytekit:

.. code:: console

   pip install flytekit

Clone the `flytekit-python-template <https://github.com/flyteorg/flytekit-python-template>`__ repo to create our own git repository called ``flyteexamples``:

.. code:: console

   git clone git@github.com:flyteorg/flytekit-python-template.git flyteexamples
   cd flyteexamples


Write Your First Flyte Workflow
###############################


Let's take a look at the example workflow in `myapp/workflows/example.py <https://github.com/flyteorg/flytekit-python-template/blob/main/myapp/workflows/example.py>`__:

.. rli:: https://raw.githubusercontent.com/flyteorg/flytekit-python-template/main/myapp/workflows/example.py
   :language: python

As you can see, a Flyte :std:doc:`task <generated/flytekit.task>` is the most basic unit of work in Flyte,
and you can compose multiple tasks into a :std:doc:`workflow <generated/flytekit.workflow>`. Try running and
modifying the ``example.py`` script locally.

Start a Local Flyte Backend
###########################

Once you're happy with the ``example.py`` script, run the following command in your terminal:

.. code:: console

   docker run --rm --privileged -p 30081:30081 -p 30082:30082 -p 30084:30084 ghcr.io/flyteorg/flyte-sandbox

When you see the message ``Flyte is ready!``, your local sandbox should be ready on http://localhost:30081/console.

Register Your Workflows
###########################

Now we're ready to ship your code to the Flyte backend by running the following command:

.. code:: console

   FLYTE_AWS_ENDPOINT=http://localhost:30084/ FLYTE_AWS_ACCESS_KEY_ID=minio FLYTE_AWS_SECRET_ACCESS_KEY=miniostorage make fast_register

Run Your Workflows
##################

To run a workflow, go to http://localhost:30081/console/projects/flyteexamples/workflows and then follow these steps:

1. Select the ``hello_world`` workflow
2. Click the **Launch Workflow** button in the upper right corner
3. Update the ``name`` input argument
4. Proceed to **Launch** to trigger an execution

.. rubric:: 🎉 Congratulations, you just ran your first Flyte workflow 🎉


Next Steps: Tutorials
#####################

To experience the full capabilities of Flyte, try out the `Flytekit Tutorials <https://flytecookbook.readthedocs.io/en/latest/>`__ 🛫


.. toctree::
   :maxdepth: 1
   :name: external-links
   :hidden:

   Tutorials <https://flytecookbook.readthedocs.io>
   Flytekit Python <https://flytekit.readthedocs.io>

.. toctree::
   :caption: How-Tos
   :maxdepth: 1
   :name: howtotoc
   :hidden:

   plugins/index
   howto/index

.. toctree::
   :caption: Deep Dive
   :maxdepth: 1
   :name: divedeeptoc
   :hidden:

   dive_deep/index
   reference/index

.. toctree::
   :caption: Contributor Guide
   :maxdepth: 1
   :name: roadmaptoc
   :hidden:

   community/index
   community/docs
   community/roadmap
   community/compare
