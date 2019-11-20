.. _concepts-notifications:

Notifications
=============

Notifications are customizable events triggered on workflow termination.

Notifications can be of three flavors:

- `Email <#howto-email-notifications>`__
- `PagerDuty <#howto-pagerduty-notifications>`__ 
- `Slack <#howto-slack>`__

You can combine notifications in a single workflow to trigger for different
combinations of recipients, domains and terminal phases.

Terminal workflow phases include:

- ``WorkflowExecution.Phase.FAILED``

- ``WorkflowExecution.Phase.TIMED_OUT``

- ``WorkflowExecution.Phase.ABORTED``

- ``WorkflowExecution.Phase.SUCCESS``


Future work
-----------
Currently the notification email subject and body are hard-coded. In the future passing these as customizable parameters in the notification proto message would be helpful.
