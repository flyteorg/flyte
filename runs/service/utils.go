package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"slices"
	"time"

	"connectrpc.com/connect"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	flyteIdlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

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

func CoalesceNullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

// foldRunStartTimeIntoHash incorporates the run start time into an offloaded inputs hash when the
// trigger binds its scheduled time to an input variable (kickoffArg, set via flyte.TriggerTime). That
// time-bound input is not part of the offloaded blob, so without this every fire of the trigger would
// share a cache key and a cacheable task would return stale outputs from the first fire. It is a no-op
// when the trigger has no kickoff arg (the time is not a declared input) or no run start time is set.
func foldRunStartTimeIntoHash(inputsHash, kickoffArg string, ts *timestamppb.Timestamp) string {
	if kickoffArg == "" || ts == nil {
		return inputsHash
	}
	return inputsHash + "|runStartTime=" + ts.AsTime().UTC().Format(time.RFC3339Nano)
}

// generateCacheKeyForTask generates a cache key for a task using a precomputed inputs hash.
func generateCacheKeyForTask(taskTemplate *flyteIdlCore.TaskTemplate, inputsHash string) (string, error) {
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

// computeFilteredInputsHash filters out cache-ignored inputs and returns the hash.
func computeFilteredInputsHash(taskTemplate *flyteIdlCore.TaskTemplate, inputs *task.Inputs) (string, error) {
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

	return hashInputs(&task.Inputs{Literals: filteredInputs})
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

// validateProjectExists checks that the given project ID exists by calling the ProjectService.
func validateProjectExists(ctx context.Context, projectClient projectconnect.ProjectServiceClient, projectID string) error {
	if _, err := projectClient.GetProject(ctx, connect.NewRequest(&project.GetProjectRequest{
		Id: projectID,
	})); err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("project %q not found", projectID))
		}
		logger.Errorf(ctx, "Failed to validate project %q: %v", projectID, err)
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to validate project: %w", err))
	}
	return nil
}
