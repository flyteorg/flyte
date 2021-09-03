package aws

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

func hashIdentifier(identifier core.Identifier) uint64 {
	h := fnv.New64()
	_, err := h.Write([]byte(fmt.Sprintf(scheduleNameInputsFormat,
		identifier.Project, identifier.Domain, identifier.Name)))
	if err != nil {
		// This shouldn't occur.
		logger.Errorf(context.Background(),
			"failed to hash launch plan identifier: %+v to get schedule name with err: %v", identifier, err)
		return 0
	}
	logger.Debugf(context.Background(), "Returning hash for [%+v]: %d", identifier, h.Sum64())
	return h.Sum64()
}
