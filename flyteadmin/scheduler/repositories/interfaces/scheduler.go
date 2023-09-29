package interfaces

type SchedulerRepoInterface interface {
	SchedulableEntityRepo() SchedulableEntityRepoInterface
	ScheduleEntitiesSnapshotRepo() ScheduleEntitiesSnapShotRepoInterface
}
