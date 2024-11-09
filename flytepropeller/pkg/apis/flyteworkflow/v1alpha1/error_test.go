package v1alpha1

import (
	"encoding/json"
	"testing"

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

	assert.Equal(t, execError.GetCode(), newExecErr.ExecutionError.GetCode())
	assert.Equal(t, execError.GetMessage(), newExecErr.ExecutionError.GetMessage())
	assert.Equal(t, execError.GetErrorUri(), newExecErr.ExecutionError.GetErrorUri())
}
