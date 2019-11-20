.. _fetaures-lanuchplans:

Flyte Launchplans
#################

When to use launchplans?
========================

- I want multiple schedules for my workflow with zero or more predefined inputs.
- I want to run the same workflow but with a different set of notifications.
- I want to share my workflow with another user with the inputs already set so that the other user can simply kick off an execution.
- I want to share my workflow to another user but also make sure that some inputs can be overridden if needed.
- I want to share my workflow with another user but make sure that some inputs are never changed.


Partial Inputs for Launchplans
==============================
Launch plans bind a partial or complete list of inputs necessary to launch a workflow. Launch plan inputs must only assign inputs already defined in the reference workflow definition.  Refer to to :ref:`primary launch plan concept documentation <concepts-launchplans-inputs>` for a detailed introduction to launch plan input types.

For example, let's say you had the following workflow definition:

.. code:: python

    @workflow_class
    class MyWorkflow(object):
        sample_size = Input(Types.Integer, required=False, default=100)
        region = Input(Types.String)
        run_date = Input(Types.Datetime)
        ...

If you wanted to run the workflow over a set of date ranges for a specific region, you could define the following launch plan:

.. code:: python

    sea_launch_plan = MyWorkflow.create_launch_plan(
        default_inputs={'sample_size': 1000},
        fixed_inputs={'region': 'SEA'},                
    )

Some things to note here, we redefine the ``sample_size`` input in the launch plan - this is perfectly fine. Workflow inputs with default values can be redefined in launch plans. If we decide at execution creation time to adjust ``sample_size`` once more that's also perfectly fine because default_inputs can be overriden. However the region input is *fixed* and cannot be overriden.

The launch plan doesn't assign run_date but there is nothing wrong with creating a launch plan that assigns all workflow inputs (either as default or fixed inputs). Of course all required inputs must be resolved at execution time, but execution creation requests can always accept dynamic inputs should your launch plan not define the complete set.

Let's take a look at how to use the launch plan to create executions programmatically:

.. code:: python

     from datetime import timedelta, date

     run_date = date(2018, 1, 1)
     end_date = date(2019, 1, 1)
     one_day = timedelta(days=1)
     while run_date <= end_date:
         sea_launch_plan.execute('myflyteproject', 'myflytedomain', inputs={'run_date': run_date})
         run_date += one_day

And boom, you've got a year's worth of executions created!

Let's say for 2019 data you've realized it makes sense to change the sample size. All your other data is in place and you can't wait to be notified once this last workflow has completed. Easy! you can configure your launch plan like so:

.. code:: python

     from datetime import timedelta, date
     
     improved_sea_launch_plan = MyWorkflow.create_launch_plan(
        default_inputs={'sample_size': 2000},
        fixed_inputs={'region': 'SEA', 'run_date': date(2019, 1, 1)},
        notifications=[notifications.Email([WorkflowExecutionPhase.SUCCEEDED], ["me@myemail.com"])]
     )
     improved_sea_launch_plan.execute('myflyteproject', 'myflytedomain', name="one_last_execution")

Schedules for Launchplans
==========================

Schedules can be expressed as :ref:`rate units or cron expressions <concepts-schedules>`. 

For example, considering the following workflow definition with a triggered-at time input:

.. code:: python

    @workflow_class
    class MyWorkflow(object):
        triggered_time = Input(Types.Datetime)
        an_integer_input = Input(Types.Integer)
        an_alternative_input = Input(Types.Integer, required=False, default=10)
        ...

To run ``MyScheduledWorkflow`` every 5 minutes with a value set for ``an_integer_input`` and the scheduled execution time assigned to the ``triggered_time`` input you could define the following launch plan:

.. code:: python

    my_schedule = schedule.FixedRate(datetime.timedelta(minutes=5), kickoff_time_input_arg="triggered_time")
    my_fixed_rate_launch_plan = MyWorkflow.create_launch_plan(
        default_inputs={'an_integer_input': Input(Types.Integer, default=900)},
        schedule=my_schedule,
    )

Workflows that run on a schedule do not always need a corresponding triggered at input value.
For example, take the following workflow:

.. code:: python

    @workflow_class
    class MyWorkflow(object):
        an_integer_input = Input(Types.Integer)
        an_alternative_input = Input(Types.Integer, required=False, default=10)
        ...


The above can be run on a cron schedule every 5 minutes like so:

.. code:: python

    my_cron_launch_plan = MyWorkflow.create_launch_plan(
        default_inputs={'an_integer_input': Input(Types.Integer, default=5)},
        schedule=schedules.CronSchedule('0/5 * * * ? *'),,
    )

Once you've initialized your launch plan, don't forget to set it to active so that the schedule is run.

.. code:: python

    my_fixed_rate_launch_plan.update(LaunchPlanState.ACTIVE)
    # Our fixed rate schedule is now active.
    
    my_cron_launch_plan.update(LaunchPlanState.ACTIVE)
    # Our cron schedule is now active too. 
    
    # Too many active, let's disable one!
    my_fixed_rate_launch_plan.update(LaunchPlanState.INACTIVE)
    
