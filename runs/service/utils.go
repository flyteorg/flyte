package service

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"slices"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	flyteIdlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

// GenerateCacheKeyForTask generates a cache key for a task matching the cache key generation logic in the sdk
func generateCacheKeyForTask(taskTemplate *flyteIdlCore.TaskTemplate, inputs *task.Inputs) (string, error) {
	ignoredInputsVars := taskTemplate.GetMetadata().GetCacheIgnoreInputVars()
	if ignoredInputsVars == nil {
		ignoredInputsVars = []string{}
	}

	var filteredInputs []*task.NamedLiteral
	for _, namedLiteral := range inputs.GetLiterals() {
		if !slices.Contains(ignoredInputsVars, namedLiteral.GetName()) {
			filteredInputs = append(filteredInputs, namedLiteral)
		}
	}

	inputsHash, err := hashInputs(&task.Inputs{
		Literals: filteredInputs,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to hash inputs")
	}

	taskInterface := taskTemplate.GetInterface()
	interfaceHash, err := hashInterface(taskInterface)
	if err != nil {
		return "", errors.Wrap(err, "failed to hash interface")
	}

	taskName := taskTemplate.GetId().GetName()
	cacheVersion := taskTemplate.GetMetadata().GetDiscoveryVersion()

	data := fmt.Sprintf("%s%s%s%s", inputsHash, taskName, interfaceHash, cacheVersion)
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

// hashInterface computes a SHA-256 hash of the given TypedInterface matching the hashing logic in the sdk
func hashInterface(iface *flyteIdlCore.TypedInterface) (string, error) {
	if iface == nil {
		return "", nil
	}

	marshaller := proto.MarshalOptions{Deterministic: true}
	serializedInterface, err := marshaller.Marshal(iface)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal interface")
	}

	hash := sha256.Sum256(serializedInterface)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

// hashInputs computes a SHA-256 hash of the given Inputs matching the hashing logic in the sdk
func hashInputs(inputs *task.Inputs) (string, error) {
	if inputs == nil {
		return "", nil
	}

	marshaller := proto.MarshalOptions{Deterministic: true}
	marshaledInputs, err := marshaller.Marshal(inputs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal inputs")
	}

	hash := sha256.Sum256(marshaledInputs)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

// validateDateRange validates that start is before end and duration doesn't exceed maxDuration.
// If startDate is nil, skip validation (normalizeDateRange will handle filling defaults).
func validateDateRange(startDate, endDate *timestamppb.Timestamp, maxDuration time.Duration) error {
	if startDate == nil {
		// No validation needed if startDate is nil - normalizeDateRange will create a valid default
		return nil
	}

	now := time.Now()
	startTime := startDate.AsTime()
	if startTime.After(now) {
		return fmt.Errorf("start_date cannot be in the future. got %v", startDate.AsTime())
	}

	// check if startDate before current time and time difference is less than 30 days
	if endDate == nil {
		duration := now.Sub(startTime)
		if duration > maxDuration {
			return fmt.Errorf("date range from start_date to now cannot exceed %v, got %v", maxDuration, duration)
		}
		return nil
	}

	endTime := endDate.AsTime()
	// Validate that start is before end
	if !startTime.Before(endTime) {
		return fmt.Errorf("start_date must be before end_date. Got start_date: %v, end_date: %v",
			startTime.Format(time.DateTime), endTime.Format(time.DateTime))
	}

	// Validate duration doesn't exceed maxDuration
	duration := endTime.Sub(startTime)
	if duration > maxDuration {
		return fmt.Errorf("date range cannot exceed %v, got %v", maxDuration, duration)
	}

	return nil
}

// normalizeDateRange fills in missing dates with defaults based on the defaultDuration.
// The first return value (startDate) is guaranteed to be non-nil.
// If both dates are nil, returns last defaultDuration from now as start, and nil as end (to include running tasks).
// If only start is provided, leaves end as nil (to include running tasks).
// If only end is provided, sets start to defaultDuration before end.
// If both are provided, returns them as-is.
func normalizeDateRange(startDate, endDate *timestamppb.Timestamp, defaultDuration time.Duration) (*timestamppb.Timestamp, *timestamppb.Timestamp) {
	if startDate == nil && endDate == nil {
		// Both nil: use last defaultDuration from now as start, leave end nil to include running tasks
		now := time.Now()
		return timestamppb.New(now.Add(-defaultDuration)), nil
	}

	if startDate != nil && endDate == nil {
		// Only start provided: leave end as nil to include running tasks
		return startDate, nil
	}

	if startDate == nil && endDate != nil {
		// Only end provided: set start to defaultDuration before end
		startTime := endDate.AsTime().Add(-defaultDuration)
		return timestamppb.New(startTime), endDate
	}

	return startDate, endDate
}

// TruncateShortDescription truncates the short description to 255 characters if it exceeds the maximum length.
func truncateShortDescription(description string) string {
	if len(description) > 255 {
		return description[:255]
	}
	return description
}

// truncateLongDescription truncates the long description to 2048 characters if it exceeds the maximum length.
func truncateLongDescription(description string) string {
	if len(description) > 2048 {
		return description[:2048]
	}
	return description
}
