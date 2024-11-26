.. _divedeep-launchplans:

Launch plans
============

.. tags:: Basic, Glossary, Design

Launch plans help execute workflows. A workflow can be associated with multiple launch plans and launch plan versions, but an individual launch plan is always associated with a single, specific workflow. After creating a launch plan, it is easy to share and execute them.

Launch plans provide a way to templatize Flyte workflow invocations. Launch plans contain a set of bound workflow inputs that are passed as arguments to create an execution. Launch plans do not necessarily contain the entire set of required workflow inputs, but a launch plan is always necessary to trigger an execution. Additional input arguments can be provided at execution time to supplement launch plan static input values.

In addition to templatized inputs, launch plans allow you to run your workflow on one or multiple schedules. Each launch
plan can optionally define a single schedule (which can be easily disabled by disabling the launch plan) as well as
optional notifications. Refer to the :ref:`deployment-configuration-notifications` for a deep dive into available notifications.

The Association between Workflows and LaunchPlans
-------------------------------------------------

Every workflow comes with a `default` launch plan that has the same name as that of a workflow. The default launch plan is authored (in code) as part of creating a new workflow.
A launch plan version can only ever be mapped to one workflow version; meaning a launch plan version cannot be used twice. This is because part of what makes a new launch plan version is the mapping to the specific workflow version.

.. note::
   Users rarely interact with the default launch plan.

Suppose we have ``Workflow A`` in ``version 1``, ``LaunchPlans`` ``A`` and ``B`` in ``version 1``, and ``LaunchPlan`` ``B`` in ``version 2``, then:

1. ``Workflow A`` can be associated with ``LaunchPlan A`` (version 1);
2. ``Workflow A`` can be associated with ``LaunchPlan B`` (different launch plan name; version 1);
3. ``Workflow A`` can be associated with ``LaunchPlan B`` (version 2).


What do Launch Plans Provide?
------------------------------

- One click invocation of workflows with predefined inputs and friendly launch plan names.
- Multiple schedules with different default values for inputs per workflow.
- Ability to easily enable and disable schedules.
- Can be created dynamically with flyteclient or statically using the Flyte SDK.
- Associate different notifications with your workflows.
- Restrict inputs to be passed to the workflows at launch time using the :ref:`fixed_inputs <fixed_inputs>` parameter.
- Multiple versions of the launch plan (with same name) with only one active version. Schedule will reflect only on the active launch plan version.

.. _concepts-launchplans-inputs:

Launch plan inputs
------------------
Generally launch plan inputs correspond to their related workflow definition's inputs, in that the variable type and names are expected to match. Launch plans cannot introduce any inputs not defined in the core workflow definition. However, launch plan inputs differ slightly from workflow inputs in that the former are categorized into **default inputs** and **fixed inputs**.

Default Inputs
^^^^^^^^^^^^^^
Default inputs behave much like default workflow inputs. As their name implies, default inputs provide default workflow input values at execution time in the absence of any dynamically provided values.

.. _fixed_inputs:

Fixed Inputs
^^^^^^^^^^^^
Fixed inputs cannot be overridden. If a workflow is executed with a launch plan and dynamic inputs that attempt to redefine the launch plan's fixed inputs, the execution creation request *will fail*.
