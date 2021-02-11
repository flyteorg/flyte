.. _divedeep_overview:

####################
Logical Overview
####################

Illustration of a workflow with tasks
----------------------------------------

.. image:: ./flyte_wf_tasks_high_level.png


:ref:`Tasks <divedeep-tasks>` are at the core of Flyte. A Task is any independent unit of
processing. Tasks can be pure functions or functions with side-effects. Tasks also have
configuration and requirements specification associated with each definition of the task.

:ref:`Workflows <divedeep-workflows>` are programs that are guaranteed to reach a terminal
state eventually. They are represented as Directed Acyclic Graphs (DAGs) expressed in protobuf.
The Flyte specification language expresses DAGs with branches, parallel steps and nested
Workflows. Workflow can optionally specify typed inputs and produce typed outputs, which
are captured by the framework. Workflows are composed of one or more
:ref:`Nodes <divedeep-nodes>`. A Node is an encapsulation of an instance of a Task.

:ref:`Executions <divedeep-executions>` are instances of workflows, nodes or tasks created
in the system as a result of a user-requested execution or a scheduled execution.

:ref:`Projects <divedeep-projects>` are a multi-tenancy primitive in Flyte that allow
logical grouping of Flyte workflows and tasks. Projects often correspond to source code
repositories. For example the project *Save Water* may include multiple `Workflows`
that analyze wastage of water etc.

:ref:`Domains <divedeep-domains>` enable workflows to be executed in different environments,
with separate resource isolation and feature configuration.

:ref:`Launchplans <divedeep-launchplans>` provide a mechanism to specialize input parameters
for workflows associated different schedules.

:ref:`Registration <divedeep-registrations>` is the process of uploading a workflow and its
task definitions to the :ref:`FlyteAdmin <divedeep-admin>` service. Registration creates
an inventory of available tasks, workflows and launchplans declared per project
and domain. A scheduled or on-demand execution can then be launched against one of
the registered entities.

Refer to the `divedeep-architecture-overview`_ to get an overview of the system architecture.