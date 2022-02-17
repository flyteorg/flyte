package snapshoter

import (
	"bytes"
	"context"
	"time"

	repositoryInterfaces "github.com/flyteorg/flyteadmin/scheduler/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type Metrics struct {
	Scope                      promutils.Scope
	SnapshotSaveErrCounter     prometheus.Counter
	SnapshotCreationErrCounter prometheus.Counter
}

type snapshoter struct {
	metrics Metrics
	db      repositoryInterfaces.SchedulerRepoInterface
}

func (w *snapshoter) Save(ctx context.Context, writer Writer, snapshot Snapshot) {
	var bytesArray []byte
	f := bytes.NewBuffer(bytesArray)
	// Only write if the snapshot has contents and not equal to the previous snapshot
	if !snapshot.IsEmpty() {
		err := writer.WriteSnapshot(f, snapshot)
		// Just log the error
		if err != nil {
			w.metrics.SnapshotCreationErrCounter.Inc()
			logger.Errorf(ctx, "unable to write the snapshot to buffer due to %v", err)
		}
		err = w.db.ScheduleEntitiesSnapshotRepo().Write(ctx, models.ScheduleEntitiesSnapshot{
			Snapshot: f.Bytes(),
		})
		if err != nil {
			w.metrics.SnapshotSaveErrCounter.Inc()
			logger.Errorf(ctx, "unable to save the snapshot to the database due to %v", err)
		}
	}
}

func (w *snapshoter) Read(ctx context.Context, reader Reader) (Snapshot, error) {
	scheduleEntitiesSnapShot, err := w.db.ScheduleEntitiesSnapshotRepo().Read(ctx)
	var snapshot Snapshot
	snapshot = &SnapshotV1{LastTimes: map[string]*time.Time{}}
	// Just log the error but dont interrupt the startup of the scheduler
	if err != nil {
		if err.(errors.FlyteAdminError).Code() == codes.NotFound {
			// This is not an error condition and hence can be ignored.
			return snapshot, nil
		}
		logger.Errorf(ctx, "unable to read the snapshot from the DB due to %v", err)
		return nil, err
	}
	f := bytes.NewReader(scheduleEntitiesSnapShot.Snapshot)
	snapshot, err = reader.ReadSnapshot(f)
	// Similarly just log the error but dont interrupt the startup of the scheduler
	if err != nil {
		logger.Errorf(ctx, "unable to construct the snapshot struct from the file due to %v", err)
		return nil, err
	}
	return snapshot, nil
}

func New(scope promutils.Scope, db repositoryInterfaces.SchedulerRepoInterface) Persistence {
	return &snapshoter{
		metrics: getSnapshoterMetrics(scope),
		db:      db,
	}
}

func getSnapshoterMetrics(scope promutils.Scope) Metrics {
	return Metrics{
		Scope: scope,
		SnapshotSaveErrCounter: scope.MustNewCounter("checkpoint_save_error_counter",
			"count of unsuccessful attempts to save the created snapshot to the DB"),
		SnapshotCreationErrCounter: scope.MustNewCounter("checkpoint_creation_error_counter",
			"count of unsuccessful attempts to create the snapshot from the inmemory map"),
	}
}
