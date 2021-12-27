.. _native-scheduler-architecture:

###################################
Flyte Native Scheduler Architecture
###################################

Introduction
============
Any workflow engine needs functionality to support scheduled executions.Flyte fulfills this need using an inbuilt native scheduler which allows scheduling fixed rate aswell as cron based schedules. The workflow author specifies the schedule during the `launchplan creation <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/scheduled_workflows/lp_schedules.html#cron-schedules>`__ and `activates or deactivates <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/scheduled_workflows/lp_schedules.html#activating-a-schedule>`__ the schedule using the `admin API's <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#launchplanupdaterequest>`__ exposed for the launchplan

Characteristics
===============

#. Cloud provider independent
#. Standard `cron <https://en.wikipedia.org/wiki/Cron#CRON_expression>`__ support
#. Scalable independently
#. Small memory footprint
#. Schedules run as lightweight go routines
#. Fault tolerant and available
#. Support in sandbox environment


Components
==========

Schedule Management
-------------------

This component supports creation/activation and deactivation of schedules. Each schedule is tied to a launchplan and is versioned in a similar manner.Also Schedule is created or it state changed to activated/deactivated whenever the `admin API <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#launchplanupdaterequest>`__ is invoked for it with `ACTIVE/INACTIVE state <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#ref-flyteidl-admin-launchplanstate>`__ either through `flytectl <https://docs.flyte.org/projects/flytectl/en/latest/gen/flytectl_update_launchplan.html#synopsis>`__ or through any other client calling the GRPC api.
The API similar to that of launchplan, makes sure at the most one schedule is active for a launchplan.


Scheduler
---------

This component is a singleton and is responsible for reading the schedules from the DB and running them at the cadence defined by the schedule.The lowest granularity supported is minutes for scheduling through cron and fixed rate scheduler.The scheduler would be running in one replica, two at the most during redeployment. Multiple replicas will just duplicate the work since each execution for a scheduleTime will have unique identifier derived from schedule name and time of the schedule. The idempotency aspect of the admin for same identifier prevents duplication on the admin side.The scheduler runs continuously in a loop reading the updated schedule entries in the data store and adding or removing the schedules. Removing a schedule will not alter in-flight go-routines launched by the scheduler.Thus the behavior of these executions is undefined.


Snapshoter
**********

This component is responsible for writing the snapshot state of all the schedules at a regular cadence to a persistent store. It uses DB to store the GOB format of the snapshot which is versioned.The snapshot is map[string]time.Time which stores a map of schedules names to there last execution times.During bootup the snapshot is bootstraped from the data store and loaded in memory.The Scheduler uses this snapshot to schedule any missed schedules.

CatchupAll-System
*****************
This component runs at bootup and catches up all the schedules to the current time.Now().Parallely new runs for the schedules also get sent to the admin.
But any failure in catching up is considered to be hard failure and stops the scheduler. The rerun tries to catchup from the last snapshoted data.

GOCronWrapper
*************

This component is responsible for locking in the time for the scheduled job to be invoked and adding those to the cron scheduler. Its a wrapper around the `following framework <https://github.com/robfig/cron/v3>`__ for fixed rate and cron schedules and creates in memory representation for the scheduled job functions.The scheduler provides ability to schedule a function with scheduleTime parameter. This is useful to know once the scheduled function is invoked that what scheduled time is this invocation for. This scheduler supports standard cron scheduling which has 5 `fields <https://en.wikipedia.org/wiki/Cron>`__. It requires 5 entries representing: minute, hour, day of month, month and day of week, in that order.

Job Executor
************

This component is responsible for sending the scheduled executions to flyteadmin.The job function accepts the scheduleTime and the schedule which is used for creating an execution request to the admin. Each job function is tied to schedule which gets executed in separate go routine according the schedule cadence.

Monitoring
----------

Following metrics are published by the native scheduler for easier monitoring of the health of the system.

#. JobFuncPanicCounter  : count of crashes for the job functions executed by the scheduler
#. JobScheduledFailedCounter  : count of scheduling failures by the scheduler
#. CatchupErrCounter  : count of unsuccessful attempts to catchup on the schedules
#. FailedExecutionCounter  : count of unsuccessful attempts to fire execution for a schedules
#. SuccessfulExecutionCounter  : count of successful attempts to fire execution for a schedules
