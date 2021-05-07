.. _howto-lanuchplans:

##########################
How do I add launch plans?
##########################

When to use launchplans?
========================

- I want multiple schedules for my workflow with zero or more predefined inputs.
- I want to run the same workflow but with a different set of notifications.
- I want to share my workflow with another user with the inputs already set so that the other user can simply kick off an execution.
- I want to share my workflow to another user but also make sure that some inputs can be overridden if needed.
- I want to share my workflow with another user but make sure that some inputs are never changed.

For preliminary examples on using launch plans in code, check out the canonical :std:ref:`User Guide <cookbook:sphx_glr_auto_core_flyte_basics_lp.py>` examples.

Partial Inputs for Launchplans
==============================
Launch plans bind a partial or complete list of inputs necessary to launch a workflow. Launch plan inputs must only assign inputs already defined in the reference workflow definition.  Refer to to :ref:`primary launch plan concept documentation <concepts-launchplans-inputs>` for a detailed introduction to launch plan input types.

For example, let's say you had the following workflow definition:

.. code:: python

    from datetime import datetime
    from flytekit import workflow

    @workflow
    def MyWorkflow(region: str, run_date: datetime, sample_size: int=100):
        ...

If you wanted to run the workflow over a set of date ranges for a specific region, you could define the following launch plan:

.. code:: python

    from flytekit import LaunchPlan

    sea_launch_plan = LaunchPlan.create(
        "sea_lp",
        MyWorkflow,
        default_inputs={'sample_size': 1000},
        fixed_inputs={'region': 'SEA'},                
    )

Some things to note here, we redefine the ``sample_size`` input in the launch plan - this is perfectly fine.
Workflow inputs with default values can be redefined in launch plans. If we decide at execution creation time to adjust
``sample_size`` once more that's also perfectly fine because default_inputs in a launch plan can also be overriden.
However the region input is *fixed* and cannot be overriden.

The launch plan doesn't assign run_date but there is nothing wrong with creating a launch plan that assigns
all workflow inputs (either as default or fixed inputs). The only requirement is that all required inputs (that is, those
without default values) must be resolved at execution time. The call to create an execution can still accept inputs
should your launch plan not define the complete set.

Backfills with Launchplans
==========================

Let's take a look at how to use the launch plan to create executions programmatically, for example to backfill:

.. code:: python

     from datetime import timedelta, date

     run_date = date(2020, 1, 1)
     end_date = date(2021, 1, 1)
     one_day = timedelta(days=1)
     while run_date < end_date:
         sea_launch_plan(run_date=run_date)
         run_date += one_day

And boom, you've got a year's worth of executions created!
