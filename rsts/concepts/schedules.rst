.. _concepts-schedules:

Schedules
---------
Workflows can be run automatically using schedules associated with launch plans. Schedules can either define a cron_expression_. or rate_unit_.

At most one launch plan version for a given {Project, Domain, Name} combination can be active, which means at most one schedule can be active for a launch plan. However, many unique launch plans and corresponding schedules can be defined for the same workflow.

.. _cron_expression:

Cron Expression
---------------
Cron expression strings use the `AWS syntax <http://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions>`__. These are validated at launch plan registration time.

.. _rate_unit:

Rate Unit
---------

Schedules can also be defined using fixed rates in units of **days**, **hours** and **minutes**.