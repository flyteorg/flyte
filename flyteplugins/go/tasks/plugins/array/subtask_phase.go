package array

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
)

const (
	ErrSystem errors.ErrorCode = "SYSTEM_ERROR"
)

func CheckTaskOutput(ctx context.Context, dataStore *storage.DataStore, outputPrefix, baseOutputSandbox storage.DataReference, childIdx, originalIdx int) (core.Phase, error) {
	or, err := ConstructOutputReader(ctx, dataStore, outputPrefix, baseOutputSandbox, originalIdx)
	if err != nil {
		return core.PhaseUndefined, errors.Wrapf(ErrSystem, err, "Failed to build output reader for sub task [%v] with original index [%v].", childIdx, originalIdx)
	}

	outputExists, err := or.Exists(ctx)
	if err != nil {
		return core.PhaseUndefined, errors.Wrapf(ErrSystem, err, "Failed to check if output file exists for sub task [%v] with original index [%v].", childIdx, originalIdx)
	}

	if !outputExists {
		errExists, err := or.IsError(ctx)
		if err != nil {
			return core.PhaseUndefined, errors.Wrapf(ErrSystem, err, "Failed to check if error file exists for sub task [%v] with original index [%v].", childIdx, originalIdx)
		}

		if errExists {
			logger.Debugf(ctx, "Found error file for sub task [%v] with original index [%v]. Marking as failure.",
				childIdx, originalIdx)
			return core.PhaseRetryableFailure, nil
		}
	}

	return core.PhaseSuccess, nil
}
