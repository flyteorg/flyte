.. _native-scheduler-architecture:

###################################
Flyte Native Scheduler Architecture
###################################

Introduction
============
Any workflow engine needs functionality to support scheduled executions. Flyte fulfills this need using an in-built native scheduler, which allows the scheduling of fixed rate as well as cron based schedules. The workflow author specifies the schedule during the `launchplan creation <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/scheduled_workflows/lp_schedules.html#cron-schedules>`__ and `activates or deactivates <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/scheduled_workflows/lp_schedules.html#activating-a-schedule>`__ the schedule using the `admin API's <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#launchplanupdaterequest>`__ exposed for the launchplan.

Characteristics
===============

#. Cloud provider independent
#. Standard `cron <https://en.wikipedia.org/wiki/Cron#CRON_expression>`__ support
#. Independently scalable
#. Small memory footprint
#. Schedules run as lightweight go routines
#. Fault tolerant and available
#. Support in sandbox environment


Components
==========

Schedule Management
-------------------

This component supports creation/activation and deactivation of schedules. Each schedule is tied to a launchplan and is versioned in a similar manner. The schedule is created or its state is changed to activated/deactivated whenever the `admin API <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#launchplanupdaterequest>`__ is invoked for it with `ACTIVE/INACTIVE state <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#ref-flyteidl-admin-launchplanstate>`__. This is done either through `flytectl <https://docs.flyte.org/projects/flytectl/en/latest/gen/flytectl_update_launchplan.html#synopsis>`__ or through any other client calling the GRPC API.
The API is similar to that of a launchplan, which makes sure one schedule at most is active for a given launchplan.


Scheduler
---------

This component is a singleton and is responsible for reading the schedules from the DB and running them at the cadence defined by the schedule. The lowest granularity supported is minutes for scheduling through both cron and fixed rate schedulers. The scheduler would be running in one replica, two at the most during redeployment. Multiple replicas will just duplicate the work, since each execution for a scheduleTime will have a unique identifier derived from the schedule name and the time of the schedule. The idempotency aspect of the admin for the same identifier prevents duplication on the admin side. The scheduler runs continuously in a loop reading the updated schedule entries in the data store and adding or removing the schedules. Removing a schedule will not alter in-flight go-routines launched by the scheduler. Thus the behavior of these executions is undefined.


Snapshoter
**********

This component is responsible for writing the snapshot state of all schedules at a regular cadence to a persistent store. It uses a DB to store the GOB format of the snapshot, which is versioned. The snapshot is a map[string]time.Time, which stores a map of schedule names to their last execution times. During bootup the snapshot is bootstraped from the data store and loaded in the memory. The Scheduler uses this snapshot to schedule any missed schedules.

CatchupAll-System
*****************
This component runs at bootup and catches up all the schedules to the current time.Now(). New runs for the schedules are also sent to the admin in parallel.
But any failure in catching up is considered to be a hard failure and stops the scheduler. The rerun tries to catchup from the last snapshotted data.

GOCronWrapper
*************

This component is responsible for locking in the time for the scheduled job to be invoked and adding those to the cron scheduler. It is a wrapper around the `following framework <https://github.com/robfig/cron/v3>`__ for fixed rate and cron schedules and creates in-memory representation of the scheduled job functions. The scheduler provides the ability to schedule a function with scheduleTime parameters. This is useful to know once the scheduled function is invoked as to what scheduled time this invocation is for. This scheduler supports standard cron scheduling which has 5 `fields <https://en.wikipedia.org/wiki/Cron>`__. It requires 5 entries representing: minute, hour, day of month, month and day of week, in that order.

Job Executor
************

This component is responsible for sending the scheduled executions to flyteadmin. The job function accepts the scheduleTime and the schedule which is used for creating an execution request to the admin. Each job function is tied to the schedule, which is executed in separate go routine according the schedule cadence.

Monitoring
----------

The following metrics are published by the native scheduler for easier monitoring of the health of the system:

#. JobFuncPanicCounter  : count of crashes of the job functions executed by the scheduler
#. JobScheduledFailedCounter  : count of scheduling failures by the scheduler
#. CatchupErrCounter  : count of unsuccessful attempts to catchup on the schedules
#. FailedExecutionCounter  : count of unsuccessful attempts to fire executions of a schedule
#. SuccessfulExecutionCounter  : count of successful attempts to fire executions of a schedule
