package identifier

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/google/uuid"
)

// Utility functions used by the flyte native scheduler

const (
	scheduleNameInputsFormat = "%s:%s:%s:%s"
	executionIDInputsFormat  = scheduleNameInputsFormat + ":%d"
)

// GetScheduleName generate the schedule name to be used as unique identification string within the scheduler
func GetScheduleName(ctx context.Context, s models.SchedulableEntity) string {
	return strconv.FormatUint(hashIdentifier(ctx, core.Identifier{
		Project: s.Project,
		Domain:  s.Domain,
		Name:    s.Name,
		Version: s.Version,
	}), 10)
}

// GetExecutionIdentifier returns UUID using the hashed value of the schedule identifier and the scheduledTime
func GetExecutionIdentifier(ctx context.Context, identifier core.Identifier, scheduledTime time.Time) (uuid.UUID, error) {
	hashValue := hashScheduledTimeStamp(ctx, identifier, scheduledTime)
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, hashValue)
	return uuid.FromBytes(b)
}

// hashIdentifier returns the hash of the identifier
func hashIdentifier(ctx context.Context, identifier core.Identifier) uint64 {
	h := fnv.New64()
	_, err := h.Write([]byte(fmt.Sprintf(scheduleNameInputsFormat,
		identifier.Project, identifier.Domain, identifier.Name, identifier.Version)))
	if err != nil {
		// This shouldn't occur.
		logger.Errorf(ctx,
			"failed to hash launch plan identifier: %+v to get schedule name with err: %v", identifier, err)
		return 0
	}
	logger.Debugf(ctx, "Returning hash for [%+v]: %d", identifier, h.Sum64())
	return h.Sum64()
}

// hashScheduledTimeStamp return the hash of the identifier and the scheduledTime
func hashScheduledTimeStamp(ctx context.Context, identifier core.Identifier, scheduledTime time.Time) uint64 {
	h := fnv.New64()
	_, err := h.Write([]byte(fmt.Sprintf(executionIDInputsFormat,
		identifier.Project, identifier.Domain, identifier.Name, identifier.Version, scheduledTime.Unix())))
	if err != nil {
		// This shouldn't occur.
		logger.Errorf(ctx,
			"failed to hash launch plan identifier: %+v  with scheduled time %v to get execution identifier with err: %v", identifier, scheduledTime, err)
		return 0
	}
	logger.Debugf(ctx, "Returning hash for [%+v] %v: %d", identifier, scheduledTime, h.Sum64())
	return h.Sum64()
}
