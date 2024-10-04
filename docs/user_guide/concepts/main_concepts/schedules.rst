.. _concepts-schedules:

Schedules
=========

.. tags:: Basic, Glossary

Workflows can be run automatically using :ref:`schedules <cookbook:scheduling_launch_plan>` associated with launch plans.

Only one launch plan version for a given {Project, Domain, Name} combination can be active, which means only one schedule can be active for a launch plan. This is because a single active schedule can exist across all versions of the launch plan. 

A :ref:`workflow <divedeep-workflows>` version can have multiple schedules associated with it, given that these schedules exist as versions of different launch plans. 

Creating a new schedule creates a new version of the launch plan.
If you wish to change a schedule, you will have to create a new version of that launch plan since a **schedule cannot be edited**.

FlyteAdmin keeps track of the newly-added schedules, and searches through all the versions of launch plans to set them to 'deactivated'.

The launch plan versions with schedules that were previously deactivated can be manually used, by clicking on the launch button and selecting the specific launch plan version. 

Let's now look at how schedules can be defined through cron_expression_ or rate_unit_.

.. _cron_expression:

Cron Expression
---------------
Cron expression strings use the syntax shown below. They are validated at launch plan registration time.

Format
------

A cron expression represents a set of times, with the help of 5 space-separated fields.

.. raw:: html

    <table id="cron_expression_table" class="docutils align-default">
        <colgroup>
            <col style="width: 20%">
            <col style="width: 17%">
            <col style="width: 24%">
            <col style="width: 39%">
        </colgroup>
        <thead>
            <tr class="row-odd">
                <th class="head">
                    <p>Field name</p>
                </th>
                <th class="head">
                    <p>Required</p>
                </th>
                <th class="head">
                    <p>Allowed values</p>
                </th>
                <th class="head">
                    <p>Allowed special characters</p>
                </th>
            </tr>
        </thead>
        <tbody>
            <tr class="row-even">
                <td>
                    <p>Minutes</p>
                </td>
                <td>
                    <p>Yes</p>
                </td>
                <td>
                    <p>0-59</p>
                </td>
                <td>
                    <p>* / , -</p>
                </td>
            </tr>
            <tr class="row-odd">
            <td><p>Hours</p></td>
                <td>
                    <p>Yes</p>
                </td>
                <td>
                    <p>0-23</p>
                </td>
                <td>
                    <p>* / , -</p>
                </td>
            </tr>
            <tr class="row-even">
                <td>
                    <p>Day of month</p>
                </td>
                <td>
                    <p>Yes</p>
                </td>
                <td>
                    <p>1-31</p><
                </td>
                <td>
                    <p>/ , - ?</p>
                </td>
            </tr>
            <tr class="row-odd">
                <td>
                    <p>Month</p>
                </td>
                <td>
                    <p>Yes</p>
                </td>
                <td>
                    <p>1-12 or JAN-DEC</p>
                </td>
                <td>
                    <p>* / , -</p>
                </td>
            </tr>
            <tr class="row-even">
                <td>
                    <p>Day of week</p>
                </td>
                <td>
                    <p>Yes</p>
                </td>
                <td>
                    <p>0-6 or SUN-SAT</p>
                </td>
                <td>
                    <p>* / , - ?</p>
                </td>
            </tr>
        </tbody>
    </table>

**Note**: The 'Month' and 'Day of week' fields are case insensitive.


Cron schedules
--------------
An incorrect cron schedule expression leads to a failure in triggering the schedule. :ref:`Here <cron_expression>` is a table that shows the format of a cron expression.

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


.. _fixed_rate:

Fixed rate schedules
----------------------
Instead of cron schedules, fixed rate schedules can be used.

You can specify the duration in the schedule using `timedelta`, that supports `minutes`, `hours`, `days` and `weeks`.

:ref:`Here <Fixed Rate Intervals>` is an example with duration in `minutes`.

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

.. _rate_unit:

Rate Unit
---------

Schedules can also be defined using fixed rates in units of **days**, **hours** and **minutes**.
