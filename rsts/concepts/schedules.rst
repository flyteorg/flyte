.. _concepts-schedules:

Schedules
==========
Workflows can be run automatically using `schedules <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/scheduled_workflows/lp_schedules.html#scheduling-workflows>`_ associated with launch plans. Schedules can define a cron_expression_. or rate_unit_.

At most, only one launch plan version for a given {Project, Domain, Name} combination can be active, which means at most, only one schedule can be active for a launch plan. However, many unique launch plans and corresponding schedules can be defined for the same workflow.

.. _cron_expression:

Cron Expression
---------------
Cron expression strings use :ref:`this <cron_expression_table>` syntax. They are validated at launch plan registration time.

.. _rate_unit:

Format
---------------

A cron expression represents a set of times, with the help of 5 space-separated fields.

.. _cron_expression_table:

+--------------+------------+-----------------+----------------------------+
| Field name   | Mandatory? | Allowed values  | Allowed special characters |
+==============+============+=================+============================+
| Minutes      | Yes        | 0-59            | * / , -                    |
+--------------+------------+-----------------+----------------------------+
| Hours        | Yes        | 0-23            | * / , -                    |
+--------------+------------+-----------------+----------------------------+
| Day of month | Yes        | 1-31            | * / , - ?                  |
+--------------+------------+-----------------+----------------------------+
| Month        | Yes        | 1-12 or JAN-DEC | * / , -                    |
+--------------+------------+-----------------+----------------------------+
| Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?                  |
+--------------+------------+-----------------+----------------------------+

**Note**: The 'Month' and 'Day of week' fields are case insensitive.

Short hand cron formats/predefined schedules
----------------------------------------------

+------------------------+--------------------------------------------+---------------+
| Entry                  | Description                                | Equivalent To |
+========================+============================================+===============+
| @yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 1 1 *     |
+------------------------+--------------------------------------------+---------------+
| @monthly               | Run once a month, midnight, first of month | 0 0 1 * *     |
+------------------------+--------------------------------------------+---------------+
| @weekly                | Run once a week, midnight between Sat/Sun  | 0 0 * * 0     |
+------------------------+--------------------------------------------+---------------+
| @daily (or @midnight)  | Run once a day, midnight                   | 0 0 * * *     |
+------------------------+--------------------------------------------+---------------+
| @hourly                | Run once an hour, beginning of hour        | 0 * * * *     |
+------------------------+--------------------------------------------+---------------+

The string field of the schedule should be populated. `Here`_ is an example.
.. _Here: https://docs.flyte.org/projects/cookbook/en/latest/auto/core/scheduled_workflows/lp_schedules.html#cron-schedules

Cron schedules
----------------
An incorrect cron schedule expression leads to a failure in triggering the schedule. :ref:`Here <cron_expression_table>` is a table that shows the format of a cron expression.

Below is another example:

.. code-block:: default
    
    cron_lp_every_min_of_hour = LaunchPlan.get_or_create(
    name="my_cron_scheduled_lp",
    workflow=date_formatter_wf,
    schedule=CronSchedule(
        # Note that kickoff_time_input_arg matches the workflow input we defined above: kickoff_time
        # But in case you are using the AWS scheme of schedules and not using the native scheduler then switch over the schedule parameter with cron_expression
        schedule="@hourly", # Following schedule runs every hour at beginning of the hour
        kickoff_time_input_arg="kickoff_time",
    ),
)

Fixed rate schedules
----------------------
Instead of cron schedules, fixed rate schedules can be used.

You can specify the duration in the schedule using `timedelta`, that supports `minutes`, `hours`, `days` and `weeks`.

`Here <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/scheduled_workflows/lp_schedules.html#fixed-rate-intervals>`_ is an example with duration in `minutes`.

Below is an example with duration in `days`.

.. code-block:: default

	fixed_rate_lp_days = LaunchPlan.get_or_create(
	    name="my_fixed_rate_lp_days",
	    workflow=positive_wf,
	    # Note that the above workflow doesn't accept any kickoff time arguments.
	    # We omit the ``kickoff_time_input_arg`` from the FixedRate schedule invocation
	    schedule=FixedRate(duration=timedelta(days=1)),
	    fixed_inputs={"name": "you"},
)


Rate Unit
---------

Schedules can also be defined using fixed rates in units of **days**, **hours** and **minutes**.