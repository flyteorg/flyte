package v1alpha1

import (
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

// Wrapper around core.Execution error. Execution Error has a protobuf enum and hence needs to be wrapped by custom marshaller
type ExecutionError struct {
	*core.ExecutionError
}

func (in *ExecutionError) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.ExecutionError)
}

func (in *ExecutionError) UnmarshalJSON(b []byte) error {
	in.ExecutionError = &core.ExecutionError{}
	return utils.UnmarshalBytesToPb(b, in.ExecutionError)
}

func (in *ExecutionError) DeepCopyInto(out *ExecutionError) {
	*out = *in
}

func (in *ExecutionError) Equals(other *ExecutionError) bool {
	if in == nil && other == nil {
		return true
	}

	if in == nil || other == nil {
		return false
	}

	if in.ExecutionError != nil && other.ExecutionError != nil {
		if !proto.Equal(in.ExecutionError, other.ExecutionError) {
			return false
		}
	} else if in.ExecutionError != other.ExecutionError {
		return false
	}

	return true
}
