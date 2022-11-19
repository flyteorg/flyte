Example Contribution Guide
###########################

.. tags:: Contribute, Basic

The examples documentation provides an easy way for the community to learn about the rich set of
features that Flyte offers, and we are constantly improving them with your help!

Whether you're a novice or experienced software engineer, data scientist, or machine learning
practitioner, all contributions are welcome!

How to Contribute
=================

The Flyte documentation examples guides are broken up into three types:

1. :ref:`User Guides <userguide>`: These are short, simple guides that demonstrate how to use a particular Flyte feature.
   These examples should be runnable locally.
2. :ref:`Tutorials <tutorials>`: These are longer, more advanced guides that use multiple Flyte features to solve
   real-world problems. Tutorials are generally more complex examples that may require extra setup or that can only run
   on larger clusters.
3. :ref:`Integrations <integrations>`: These examples showcase how to use the Flyte plugins that integrate with the
   broader data and ML ecosystem.

The first step to contributing an example is to open up a
`documentation issue <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=documentation%2Cuntriaged&template=docs_issue.yaml&title=%5BDocs%5D+)>`_
to articulate the kind of example you want to write. The Flyte maintainers will guide and help you figure out where
your example would fit best.

Creating an Example
===================

.. admonition:: Prerequisites

   Follow the :ref:`env_setup` guide to get your development environment ready.

The ``flytesnacks`` repo examples live in the ``cookbook`` directory, and are organized as
follows:

.. code-block::

   cookbook
   ├── core          # User Guide Basics features
   ├── deployment    # User Guide Production Config guides
   ├── larger_apps   # User Guide Building Large Apps
   ├── remote_access # User Guide Remote Access guides
   ├── testing       # User Guide Testing guides
   ├── case_studies  # Tutorials live here
   └── integrations  # Integrations live here

.. important::

   If you're creating a new example in ``integrations`` or ``case_studies`` that doesn'takes
   fit into any of the existing subdirectories, you'll need to setup a new example directory.

   Create a new directory with ``mkdir {integrations, case_studies}/path/to/new/example_dir``
   
   Each example directory should contain:

   * ``Dockerfile``
   * ``Makefile``
   * ``README.rst``
   * ``__init__.py``
   * ``requirements.in``
   * ``sandbox.config``

   You can copy one of the existing examples and modify it to your needs.

Create your example by going to the appropriate example directory and creating a ``.py`` file with
a descriptive name for your example. Be sure to follow the `percent <https://jupytext.readthedocs.io/en/latest/formats.html#the-percent-format>`_
format for delimiting code and markdown in your script.

.. note::
   
   ``flytesnacks`` uses `sphinx gallery <https://sphinx-gallery.github.io/stable/index.html>`_
   to convert the python files to ``.rst`` format so that the examples can be rendered in the
   documentation.

Write a README
===============

The ``README.md`` file needs to capture the *what*, *why*, and *how* of the example.

* What is the integration about? Its features, etc.
* Why do we need this integration? How is it going to benefit the Flyte users?
* Showcase the uniqueness of the integration
* How to install the plugin?
  
.. tip::
   Refer to any subdirectory in the ``cookbook`` directory for examples

Explain What the Code Does
===========================

Following the `literate programming <https://en.wikipedia.org/wiki/Literate_programming>`__ paradigm, make sure to
interleave explanations in the ``*.py`` files containing the code example.

.. admonition:: A Simple Example
   :class: tip

   Here's a code snippet that defines a function that takes two positional arguments and one keyword argument:

   .. code-block:: python

      def function(x, y, z=3):
          return x + y * z

   As you can see, ``function`` adds the two first arguments and multiplies the sum with the third keyword
   argument. Can you think of a better name for this ``function``?

Explanations don't have to be this detailed for such a simple example, but you can imagine how this makes for a better
reading experience for more complicated examples.

Test your code
===============

If the example code can be run locally, just use ``python <my file>.py`` to run it.

Testing on a Cluster
---------------------

Install :doc:`flytectl <flytectl:index>`, the commandline interface for flyte.

.. note::

   Learn more about installation and configuration of Flytectl `here <https://docs.flyte.org/projects/flytectl/en/latest/index.html>`__.

Start a Flyte demo cluster with:

.. code-block::

   flytectl demo start


Testing ``core`` directory examples on sandbox
-----------------------------------------------

Build Docker container:

.. prompt:: bash

   flytectl demo exec -- docker build . --tag "core:v1" -f core/Dockerfile

Package the examples by running

.. prompt:: bash

   pyflyte --pkgs core package --image core:v1 -f

Register the examples by running

.. prompt:: bash

   flytectl register files --archive -p flytesnacks -d development --archive flyte-package.tgz --version v1

Visit ``https://localhost:30081/console`` to view the Flyte console, which consists of the examples present in the
``flytesnacks/cookbook/core`` directory.

To fetch new dependencies and rebuild the image, run 

.. prompt:: bash

   flytectl demo exec -- docker build . --tag "core:v2" -f core/Dockerfile
   pyflyte --pkgs core package --image core:v2 -f
   flytectl register files --archive -p flytesnacks -d development --archive flyte-package.tgz --version v2

Refer to `this guide <https://docs.flyte.org/projects/cookbook/en/latest/auto/larger_apps/larger_apps_iterate.html#quickly-re-deploy-your-application>`__
if the code in itself is updated and requirements.txt is the same.


Pre-commit hooks
----------------

We use `pre-commit <https://pre-commit.com/>`__ to automate linting and code formatting on every commit.
Configured hooks include `black <https://github.com/psf/black>`__, `isort <https://github.com/PyCQA/isort>`__,
`flake8 <https://github.com/PyCQA/flake8>`__ and linters to ensure newlines are added to the end of files, and there is
proper spacing in files.

We run all those hooks in CI, but if you want to run them locally on every commit, run `pre-commit install` after
installing the dev environment requirements. In case you want to disable `pre-commit` hooks locally, run
`pre-commit uninstall`. More info `here <https://pre-commit.com/>`__.


Formatting
----------

We use `black <https://github.com/psf/black>`__ and `isort <https://github.com/PyCQA/isort>`__ to autoformat code. They
are configured as git hooks in `pre-commit`. Run ``make fmt`` to format your code.

Spell-checking
--------------

We use `codespell <https://github.com/codespell-project/codespell>`__ to catch common misspellings. Run
``make spellcheck`` to spell-check the changes.

Update Documentation Pages
==========================

The ``cookbook/docs/conf.py`` contains the sphinx configuration for building the ``flytesnacks`` documentation.

For example, if you added the ``snowflake`` directory to the ``integrations/external_services`` folder, you then need
to:
   
- Add the Python file names to the ``CUSTOM_FILE_SORT_ORDER`` list
- Add ``../integrations/external_services/snowflake`` to ``example_dirs``
- Add ``auto/integrations/external_services/snowflake`` to ``gallery_dirs``

If you've created a new section in the examples guides, you need to update the table of contents and navigation panels in
the appropriate ``rst`` file.

.. note::

   You will need to update the entries in the ``.. toc::`` directive *and* ``.. panels::`` directive.

   .. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/panel_and_toc.png
      :alt: panel and TOC

Update CI Workflows
===================

To make sure your example is tested in CI/CD, add the name and path to ``.github/workflows/ghcr_push.yml`` if you're
adding an integration or a tutorial.

QA your Changes
===============

Verify that the code and documentation look as expected:
   
- Learn about the documentation tools `here <https://docs.flyte.org/en/latest/community/contribute.html#documentation>`__
- Install the requirements by running ``pip install -r docs-requirements.txt`` in the ``cookbook`` folder
- Run ``make html`` in the ``docs`` folder

   .. tip::
      For implicit targets, run ``make -C docs html``.

- Open the HTML pages present in the ``docs/_build`` directory in the browser

Create a Pull request
======================

Create the pull request, then ensure that the docs are rendered correctly by clicking on the documentation check. 
   
   .. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/common/test_docs_link.png
       :alt: Docs link in a PR

You can refer to `this PR <https://github.com/flyteorg/flytesnacks/pull/332>`__ for the exact changes required.
