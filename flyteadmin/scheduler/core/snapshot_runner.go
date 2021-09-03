package core

import (
	"context"

	sImpl "github.com/flyteorg/flyteadmin/scheduler/snapshoter"
)

const snapShotVersion = 1

// Snapshotrunner allows the ability to snapshot the scheduler state and save it to the db.
// Its invoked periodically from the scheduledExecutor
type Snapshotrunner struct {
	snapshoter sImpl.Persistence
	scheduler  Scheduler
}

func (u Snapshotrunner) Run(ctx context.Context) {
	snapshot := u.scheduler.CalculateSnapshot(ctx)
	snapshotWriter := &sImpl.VersionedSnapshot{Version: snapShotVersion}
	u.snapshoter.Save(ctx, snapshotWriter, snapshot)
}

func NewSnapshotRunner(snapshoter sImpl.Persistence, scheduler Scheduler) Snapshotrunner {
	return Snapshotrunner{snapshoter: snapshoter, scheduler: scheduler}
}
