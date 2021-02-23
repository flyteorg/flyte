.. _howto-flytecli:

########################
How do I use Flyte CLI?
########################

.. note::

    We are working hard on replacing flyte-cli, with a more robust, better designed and cross platform CLI.
    Refer to :std:ref:`flytectl`.

***************************************************
A command-line interface for interacting with Flyte
***************************************************

The FlyteCLI is a command-line tool that allows users to perform administrative
tasks on their Flyte workflows and executions. It is an independent module but
installed as part of the `Flyte Kit <components-flytekit>`. It primarily
iteracts with the `FlyteAdmin <components-flyteadmin>` service over its gRPC
interface, allowing users to list registered workflows, or get a currently
running execution.

------

Installation
============

The easist way to install FlyteCLI is using virtual environments.
Follow the official doc_ to install the ``virtualenv`` package if
you don't already have it in your development environment.

Install from source
-------------------
Now that you have virtualenv, you can either install flyte-cli from source.
To do this first clone the git repository and
after setting up and activating your virtual environment, change directory to
the root directory of the flytecli package, and install the dependencies with
``pip install -e .``.


.. _doc: https://virtualenv.pypi.io/en/latest/installation/

Install from pypi [recommended]
-------------------------------
Another option is to just install flyte-cli from prebuilt binaries

Testing if you have a working installation
------------------------------------------

To test whether you have a successful installation of flytecli, run
``flyte-cli`` or ``flyte-cli --help``.

If you see the following output, you have installed the FlyteCLI successfully.

.. code-block:: console

    Usage: flyte-cli [OPTIONS] COMMAND [ARGS]...

        Command line tool for interacting with all entities on the Flyte Platform.

    Options:
        -n, --name TEXT     [Optional] The name to pass to the sub-command (if
    ...

    Commands:
        execute-launch-plan        Kick off a launch plan.
    ...


------

Terminology
===========

This section introduces and explains the most commonly used terms and concepts
the users will see in FlyteCLI.

Host
----
``Host`` refers to your running Flyte instance and is a common
argument for the commands in FlyteCLI. The FlyteCLI will only be interacting
with the Flyte instance at the URL you specify with the ``host`` argument.
parameter. This is a required argument for most of the FlyteCLI commands.

Project
-------
``Project`` is a multi-tenancy primitive in Flyte and allows logical grouping
of instances of Flyte entities by users. Within Lyft's context, this term
usually refers to the name of the Github repository in which your workflow
code resides.

For more information see :ref:`Projects <divedeep-projects>`

Domain
------
The term ``domain`` refers to development environment (or the service instance)
of your workflow/execution/launch plan/etc. You can specify it with the
``domain`` argument. Values can be either ``development``, ``staging``, or
``production``. See :ref:`Domains <divedeep-domains>`


Name
----
The ``name`` of a named entity is a randomly generated hash assigned
automatically by the system at the creation time of the named entity. For some
commands, this is an optional argument.


Named Entity
------------
``Name Entity`` is a primitive in Flyte that allows logical grouping of
processing entities across versions. The processing entities to which this term
can refer include unversioned ``launch plans``, ``workflows``,
``executions``, and ``tasks``. In other words, an unversioned ``workflow`` named
entity is essentially a group of multiple workflows that
have the same ``Project``, ``Domain``, and ``Name``, but different versions.


URN
---

.. note::

    URN is a FlyteCLI-only concept. You won't see this term in other flyte modules.

URN stands for "unique resource name", and is the identifier of
a version of a given named entity, such as a workflow, a launch plan,
an execution, or a task. Each URN uniquely identifies a named entity.
URNs are often used in FlyteCLI to interact with specific named entities.

The URN of a version of a name entity is composible from the entity's
attributes. For example, the URN of a workflow can be composed of a prefix
`wf` and the workflow's ``project``, ``domain``, ``name``, and ``version``,
in the form of ``wf:<project>:<domain>:<name>:<version>``.

Note that execution is the sole exception here as an execution does not
have versions. The URN of an execution, therefore, is in the form of
``ex:<project>:<domain>:<name>``.

------

Flyte CLI User Configuration
==============================
The ``flyte-cli`` command line utility also supports default user-level configuration settings if the Admin service it accesses supports authentication.  To get started either create or activate a Python 3 virtual environment ::

    $ python3 -m venv ~/envs/flyte
    $ source ~/envs/flyte/bin/activate

In general, we recommend installing and using Flyte CLI inside a virtualenv.  Install ``flytekit`` (which installs ``flyte-cli``) as follows ::

    $ pip install wheel flytekit

Use the setup-config command to create yourself a default config file.  This will pull the necessary settings from Flyte's oauth metadata endpoint. ::

    (flyte) username:~ $ flyte-cli setup-config -h flyte.company.net

------

Commands
========

For information on available commands in FlyteCLI, refer to FlyteCLI's help message.

Subcommand Help
---------------

FlyteCLI uses subcommands. Whenever you feel unsure about the usage or
the arguments of a command or a subcommand, get help by running
``flyte-cli --help`` or ``flyte-cli <subcommand> --help``
