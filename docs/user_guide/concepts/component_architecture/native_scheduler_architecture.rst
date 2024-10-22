.. _native-scheduler-architecture:

###################################
Flyte Native Scheduler Architecture
###################################

.. tags:: Advanced, Design

Introduction
============
Any workflow engine needs functionality to support scheduled executions. Flyte
fulfills this using an in-built native scheduler, which schedules fixed rate and
cron-based schedules. The workflow author specifies the schedule during the
:ref:`launchplan creation <cookbook:cron-schedules>`
and :ref:`activates or deactivates <cookbook:activating-schedules>`
the schedule using the
:ref:`admin APIs <flyteidl:ref_flyteidl.admin.LaunchPlanUpdateRequest>`
exposed for the launch plan.

Characteristics
===============

#. Cloud provider independent
#. Standard `cron <https://en.wikipedia.org/wiki/Cron#CRON_expression>`__ support
#. Independently scalable
#. Small memory footprint
#. Schedules run as lightweight goroutines
#. Fault tolerant and available
#. Support in sandbox environment


Components
==========

Schedule Management
-------------------

This component supports creation/activation and deactivation of schedules. Each schedule is tied to a launch plan and is versioned in a similar manner. The schedule is created or its state is changed to activated/deactivated whenever the `admin API <https://docs.flyte.org/en/latest/protos/docs/admin/admin.html#launchplanupdaterequest>`__ is invoked for it with `ACTIVE/INACTIVE state <https://docs.flyte.org/en/latest/protos/docs/admin/admin.html#ref-flyteidl-admin-launchplanstate>`__. This is done either through `flytectl <https://docs.flyte.org/en/latest/flytectl/gen/flytectl_update_launchplan.html#synopsis>`__ or through any other client that calls the GRPC API.
The API is similar to a launchplan, ensuring that only one schedule is active for a given launchplan.


Scheduler
---------

This component is a singleton and is responsible for reading the schedules from the DB and running them at the cadence defined by the schedule. The lowest granularity supported is `minutes` for scheduling through both cron and fixed rate schedulers. The scheduler can run in one replica, two at the most during redeployment. Multiple replicas will only duplicate the work, since each execution for a scheduleTime will have a unique identifier derived from the schedule name and the time of the schedule. The idempotency aspect of the admin for the same identifier prevents duplication on the admin side. The scheduler runs continuously in a loop reading the updated schedule entries in the data store and adding or removing the schedules. Removing a schedule will not alter the in-flight goroutines launched by the scheduler. Thus, the behavior of these executions is undefined.


Snapshoter
**********

This component is responsible for writing the snapshot state of all schedules at a regular cadence to a persistent store. It uses a DB to store the GOB format of the snapshot, which is versioned. The snapshot is a map[string]time.Time, which stores a map of schedule names to their last execution times. During bootup, the snapshot is bootstrapped from the data store and loaded into memory. The Scheduler uses this snapshot to schedule any missed schedules.

CatchupAll-System
*****************
This component runs at bootup and catches up all the schedules to current time, i.e., time.Now(). New runs for the schedules are sent to the admin in parallel.
Any failure in catching up is considered a hard failure and stops the scheduler. The rerun tries to catchup from the last snapshot of data.

GOCronWrapper
*************

This component is responsible for locking in the time for the scheduled job to be invoked and adding those to the cron scheduler. It is a wrapper around `this framework <https://github.com/robfig/cron>`__ for fixed rate and cron schedules that creates in-memory representation of the scheduled job functions. The scheduler schedules a function with scheduleTime parameters. When this scheduled function is invoked, the scheduleTime parameters provide the current schedule time used by the scheduler. This scheduler supports standard cron scheduling which has 5 `fields <https://en.wikipedia.org/wiki/Cron>`__. It requires 5 entries representing ``minute``, ``hour``, ``day of month``, ``month`` and ``day of week``, in that order.

Job Executor
************

The job executor component is responsible for sending the scheduled executions to FlyteAdmin. The job function accepts ``scheduleTime`` and the schedule which is used to create an execution request to the admin. Each job function is tied to the schedule which is executed in a separate goroutine in accordance with the schedule cadence.

Monitoring
----------

To monitor the system health, the following metrics are published by the native scheduler:

#. JobFuncPanicCounter  : count of crashes of the job functions executed by the scheduler.
#. JobScheduledFailedCounter  : count of scheduling failures by the scheduler.
#. CatchupErrCounter  : count of unsuccessful attempts to catchup on the schedules.
#. FailedExecutionCounter  : count of unsuccessful attempts to fire executions of a schedule.
#. SuccessfulExecutionCounter  : count of successful attempts to fire executions of a schedule.
