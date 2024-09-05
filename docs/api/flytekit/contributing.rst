.. _contributing:

###########################
Flytekit Contribution Guide
###########################

.. tags:: Contribute, Basic

First off, thank you for thinking about contributing! Below you'll find instructions that will hopefully guide you through how to fix, improve, and extend Flytekit.

Please also take some time to read through the :std:ref:`design guides <design>`, which describe the various parts of Flytekit and should make contributing easier.

*******************
📜 Background
*******************

Below is a listing of the most important packages that comprise the flytekit SDK:

- ``flytekit/core``
  This holds all the core functionality of the new API.
- ``flytekit/types``
  We bundle some special types like ``FlyteFile, FlyteSchema etc`` by default here.
- ``flytekit/extend``
  This is the future home of extension points, and currently serves as the raw documentation for extensions.
- ``flytekit/extras``
  This contains code that we want bundled with Flytekit but not everyone may find useful (for example AWS and GCP
  specific logic).
- ``flytekit/remote``
  This implements the interface to interact with the Flyte service. Think of the code here as the Python-object version of Console.
- ``flytekit/testing``
  is the future home for testing functionality like ``mock`` etc, and currently serves as documentation.
  All test extensions should be imported from here.
- ``flytekit/models``
  Protobuf generated Python code is not terribly user-friendly, so we improve upon those ``flyteidl`` classes here.
- ``plugins``
  is the source of all plugins
- ``flytekit/bin/entrypoint.py``
  The run time entrypoint for flytekit. When a task kicks off, this is where the click command goes.
- ``flytekit/clis``
  This is the home for the CLIs.
- ``flytekit/configuration``
  This holds all the configuration objects, but dependency on configuration should be carefully considered as it
  makes compiled Flyte tasks and workflows less portable (i.e. if you run ``pyflyte package`` can someone else use
  those serialized objects).

Please also see the :std:ref:`design overview section <design>` for more in-depth information.


******************
💻 Contribute Code
******************

Setup (Do Once)
===============

We recommend using a virtual environment to develop Flytekit. Inside the top level Flytekit repo folder, run: ::

    virtualenv ~/.virtualenvs/flytekit
    source ~/.virtualenvs/flytekit/bin/activate
    make setup

This will install Flytekit dependencies and Flytekit in editable mode. This links your virtual Python's ``site-packages`` with your local repo folder, allowing your local changes to take effect when the same Python interpreter runs ``import flytekit``.

Plugin Development
==================

As discussed in the design component, Flytekit plugins currently live in this Flytekit repo, but under a different top level folder ``plugins``.
In the future, this will be separated out into a different repo. These plugins follow a `microlib <https://medium.com/@jherreras/python-microlibs-5be9461ad979>`__ structure, which will persist even if we move repos. ::

    source ~/.virtualenvs/flytekit/bin/activate
    cd plugins
    pip install -e .

This should install all the plugins in editable mode as well.

Iteration
=========

Make
^^^^
Some helpful make commands ::

    $ make
      setup        Install requirements
      fmt          Format code with ruff
      lint         Run linters
      test         Run tests
      requirements Compile requirements

Testing
^^^^^^^
Three levels of testing are available.

Unit Testing
------------
Running unit tests: ::

    source ~/.virtualenvs/flytekit/bin/activate
    make test

Flytesnacks Testing
-------------------
Please see `Flytesnacks <https://github.com/flyteorg/flytesnacks>`__ for more information.
This example repo can be cloned and run on a local Flyte cluster, or just in your IDE or other Python environment.

Follow the setup instructions for Flytesnacks and then override it with the version of Flytekit you're interested in testing by running something like: ::

    pip install https://github.com/flyteorg/flytekit/archive/a32ab82bef4d9ff53c2b7b4e69ff11f1e93858ea.zip#egg=flytekit
    # Or for a plugin
    pip install https://github.com/flyteorg/flytekit/archive/e128f66dda48bbfc6076d240d39e4221d6af2d2b.zip#subdirectory=plugins/pod&egg=flytekitplugins-pod

Change the actual link to be from your fork if you are using a fork.

End-to-end Testing
------------------

.. TODO: Replace this with actual instructions

The Flyte developer experience team has put together an end-to-end testing framework that will spin up a K8s cluster, install Flyte onto it, and run through a series of workflows.
Please contact us if you reach this stage and would like more information on this.


Pre-commit hooks
================

We use `pre-commit <https://pre-commit.com/>`__ to automate linting and code formatting on every commit.
Configured hooks include `ruff <https://github.com/astral-sh/ruff>`__ and also linters to check for the validity of YAML files and ensuring that newlines are added to the end of files.

We run all those hooks in CI, but if you want to run them locally on every commit, run `pre-commit install` after installing the dev environment requirements. In case you want to disable `pre-commit` hooks locally, for example, while you're iterating on some feature, run `pre-commit uninstall`. More info in https://pre-commit.com/.


Formatting
==========

We use `ruff <https://github.com/astral-sh/ruff>`__  to autoformat code. In fact, they have been configured as git hooks in `pre-commit`. Run the following commands to execute the formatters. ::

    source ~/.virtualenvs/flytekit/bin/activate
    make fmt

Spell-checking
==============

We use `codespell <https://github.com/codespell-project/codespell>`__ to catch spelling mistakes in both code and documentation. Run the following commands to spell-check changes. ::

    source ~/.virtualenvs/flytekit/bin/activate
    make spellcheck

******************************
📃 Contribute to Documentation
******************************

1. Install requirements by running ``make doc-requirements.txt`` in the root of the repo
2. Make the required changes
3. Verify if the documentation looks as expected by running ``make html`` in the `docs <https://github.com/flyteorg/flytekit/tree/master/docs>`__ directory
4. Open HTML pages present in the ``docs/build`` directory in the browser
5. After creating the pull request, check if the docs are rendered correctly by clicking on the documentation check

   .. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/common/test_docs_link.png
       :alt: Doc link in PR

**********************************
📝 Releases and Project Management
**********************************

Currently, Flytekit and all its plugins share one common version.
To release, contact a member of the Flytekit repo maintainers or committers, and request a release.
We will create a GitHub release off of master, which will automatically publish a Pypi package.
