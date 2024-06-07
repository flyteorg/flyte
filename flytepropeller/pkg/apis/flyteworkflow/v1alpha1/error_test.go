package v1alpha1

import (
	"encoding/json"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestExecutionErrorJSONMarshalling(t *testing.T) {
	execError := &core.ExecutionError{
		Code:     "TestCode",
		Message:  "Test error message",
		ErrorUri: "Test error uri",
	}

	execErr := &ExecutionError{ExecutionError: execError}
	data, jErr := json.Marshal(execErr)
	assert.Nil(t, jErr)

	newExecErr := &ExecutionError{}
	uErr := json.Unmarshal(data, newExecErr)
	assert.Nil(t, uErr)

	assert.Equal(t, execError.Code, newExecErr.ExecutionError.Code)
	assert.Equal(t, execError.Message, newExecErr.ExecutionError.Message)
	assert.Equal(t, execError.ErrorUri, newExecErr.ExecutionError.ErrorUri)
}

func TestExecutionErrorDeepCopy(t *testing.T) {
	execError := &core.ExecutionError{
		Code:     "TestCode",
		Message:  "Test error message",
		ErrorUri: "Test error uri",
	}

	execErr := &ExecutionError{ExecutionError: execError}
	newExecErr := execErr.DeepCopy()

	// 4. Compare the pointers and the actual values
	// Assert that the pointers are different
	assert.True(t, execErr.ExecutionError != newExecErr.ExecutionError)
	// Assert that the values stored in the proto messages are equal
	assert.True(t, proto.Equal(execErr.ExecutionError, newExecErr.ExecutionError))
}
