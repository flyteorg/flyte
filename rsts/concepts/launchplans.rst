.. _divedeep-launchplans:
  
Launch plans
============
Launch plans are used to execute workflows. A workflow can have many launch plans, but an individual launch plan is always tied to a single, specific workflow. Once created, launch plans are easy to share and execute.

Launch plans provide a way to templatize Flyte workflow invocations. Launch plans contain a set of bound workflow inputs that are passed as arguments to create an execution. Launch plans do not necessarily contain the entire set of required workflow inputs, but a launch plan is always necessary to trigger an execution. Additional input arguments can be provided at execution time to supplement launch plan static input values.

In addition to templatized inputs, launch plans allow you to run your workflow on one or multiple schedules. Each launch
plan can optionally define a single schedule (which can be easily disabled by disabling the launch plan) as well as
optional notifications. Refer to the :ref:`deployment-cluster-config-notifications` for a deep dive into available notifications.

See `here <https://docs.google.com/drawings/d/1xtG7lyk3es2S42pNnh5OGXW59jvnRIyPXCrdjPJm-3c/edit?usp=sharing>`__ for an overview.

What do launch plans provide?
-----------------------------

- One click invocation of workflows with predefined inputs and friendly launch plan names.
- Multiple schedules with different default values for inputs per workflow.
- Ability to easily enable and disable schedules.
- Can be created dynamically with flyteclient or statically using the Flyte SDK.
- Associate different notifications with your workflows.
- Restrict inputs to be passed to the workflows at launch time using the fixed_inputs_. parameter.
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

