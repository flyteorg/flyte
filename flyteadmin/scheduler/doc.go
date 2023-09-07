// Package scheduler
// Flyte scheduler implementation that allows to schedule fixed rate and cron schedules on sandbox deployment
// Scheduler has two components
// 1] Schedule management
//    This component is part of the pkg/async/schedule/flytescheduler package
//    Role of this component is to create / activate / deactivate schedules
//    The above actions are exposed through launchplan activation/deactivation api's and donot have separate controls.
//    Whenever a launchplan with a schedule is activated, a new schedule entry is created in the datastore
//    On deactivation the created scheduled and launchplan is deactivated through a flag
//    Atmost one launchplan is active at any moment across its various versions and same semantics apply for the
//    schedules as well.
// 2] Scheduler
//    This component is a singleton and has its source in the current folder and is responsible for reading the schedules
//    from the DB and running them at the cadence defined by the schedule
//    The lowest granularity supported is minutes for scheduling through cron and fixed rate scheduler
// 	  The scheduler should be running in one replica , two at the most during redeployment. Multiple replicas will just
// 	  duplicate the work since each execution for a scheduleTime will have unique identifier derived from schedule name
//	  and time of the schedule. The idempotency aspect of the admin for same identifier prevents duplication on the admin
//	  side.
//    The scheduler runs continuously in a loop reading the updated schedule entries in the data store and adding or removing
//    the schedules. Removing a schedule will not alter in-flight go-routines launched by the scheduler.
//    Thus the behavior of these executions is undefined (most probably will get executed).
//    Sub components:
//		a) Snapshoter
// 			This component is responsible for writing the snapshot state of all the schedules at a regular cadence to a
//			persistent store. The current implementation uses DB to store the GOB format of the snapshot which is versioned.
//			The snapshot is map[string]time.Time which stores a map of schedules names to there last execution times
// 			During bootup the snapshot is bootstraped from the data store and loaded in memory
//			The Scheduler use this snapshot to schedule any missed schedules.
//
//			We cannot use global snapshot time since each time snapshot doesn't contain information on how many schedules
//			were executed till that point in time. And hence the need to maintain map[string]time.Time of schedules to there
//			lastExectimes
//   		In the future we may support global snapshots, such that we can record the last successfully considered
//  		time for each schedule and select the lowest as the watermark. currently since the underlying scheduler
// 			does not expose the last considered time, we just calculate our own watermark per schedule.
// 		b) CatchupAll-System :
//			This component runs at bootup and catches up all the schedules to there current time.Now()
//			The scheduler is not run until all the schedules have been caught up.
//			The current design is also not to snapshot until all the schedules are caught up.
//			This might be drawback in case catch up runs for a long time and hasn't been snapshotted.(reassess)
//		c) GOCronWrapper :
//			This component is responsible for locking in the time for the scheduled job to be invoked and adding those
//			to the cron scheduler. Right now this uses https://github.com/robfig/cron/v3 framework for fixed rate and cron
// 			schedules
// 			The scheduler provides ability to schedule a function with scheduleTime parameter. This is useful to know
//			once the scheduled function is invoked that what scheduled time is this invocation for.
// 			This scheduler supports standard cron scheduling which has 5 fields
//			https://en.wikipedia.org/wiki/Cron
//			It requires 5 entries
//          representing: minute, hour, day of month, month and day of week, in that order.
//
//          It accepts
//   			- Standard crontab specs, e.g. "* * * * ?"
//   			- Descriptors, e.g. "@midnight", "@every 1h30m"
//		d) Job function :
//			The job function accepts the scheduleTime and the schedule which is used for creating an execution request
//			to the admin. Each job function is tied to schedule which gets executed in separate go routine by the gogf
// 			framework in according the schedule cadence.

// 		Failure scenarios:
//		a) Case when the schedule is activated but launchplan is not. Ideally admin should throw an error here but it
//		   allows to launch the scheduled execution.Bug marked here https://github.com/flyteorg/flyte/issues/1354
//		   Once this issue is fixed, then the scheduler behavior would be find the specific new error defined for this
//		   scenario.Eg : LaunchPlanNotActivated and skip the scheduled time execution after the failure.
//		   It will continue to hit the admin with new future scheduled times where the problem can get fixed for the launchplan.
//		   Hence its expected to not schedule the executions during such a discrepancy. The user need to reactivate the
//		   launchplan to fix the issue.
//		   eg: activate launch plan L1 with version V1 and create schedule. (One Api call)
//			   - Create schedule for L1,V1 succeeds
//			   - Activate launchplan fails for L1, V1
//			   - API return failure
//			Reactivate the launchplan by calling the API again to fix the discrepancy between the schedule and launchplan
//		    During the discrepancy the executions won't be scheduled on admin once the bug(1354) is fixed.
//
//      b) Case when scheduled time T1 execution fails. The goroutine executing for T1 will go through 30 repetitions before
//		   aborting the run. In such a scenario its possible that furture scheduled time T2 succeeds and gets executed successfully
//		   by the admin. i.e admin could execute the schedules in this order T2, T1. This is rare case though
//
// 		c) Case when the scheduler goes down then once it comes back up it will run catch up on all the schedules using
//		   the last snapshoted timestamp to time.Now()
//
//		d) Case when the snapshoter fails to record the last execution at T2 but has recorded at T1, where T1 < T2 ,
//		   then new schedules would be created from T1 -> time.Now() during catchup and the idempotency aspect of the admin
//		   will take care of not rescheduling the already scheduled execution from T1 -> Crash time
//
//		e) Case when the scheduler is down and the old schedule gets deactivated, then during catchup the scheduler won't
// 		   create executions for it. It doesn't matter how many activation/deactivations have happened during the downtime,
// 		   but the scheduler will go by the recent activation state
//
//		f) Similarly in case of scheduler being down and an old schedule gets activated,then during catchup the scheduler
//		   would run catch from updated_at timestamp till now. It doesn't matter how many activation/deactivations have
//		   happened during the downtime, but the scheduler will go by the recent activation state.
//
//		g) Case there are multiple pod running with the scheduler , then we rely on the idempotency aspect of the executions
//		   which have a identifier derived from the hash of schedule time + launch plan identifier which would remain the same
//		   any other instance of the scheduler picks up and admin will return the AlreadyExists error.
//

package scheduler
