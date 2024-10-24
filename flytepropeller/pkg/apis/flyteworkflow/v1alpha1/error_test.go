package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestExecutionErrorJSONMarshalling(t *testing.T) {
	execError := ExecutionError{
		&core.ExecutionError{
			Code:     "TestCode",
			Message:  "Test error message",
			ErrorUri: "Test error uri",
		},
	}

	expected, mockErr := mockMarshalPbToBytes(execError.ExecutionError)
	assert.Nil(t, mockErr)

	// MarshalJSON
	execErrorBytes, mErr := execError.MarshalJSON()
	assert.Nil(t, mErr)
	assert.Equal(t, expected, execErrorBytes)

	// UnmarshalJSON
	execErrorObj := &ExecutionError{}
	uErr := execErrorObj.UnmarshalJSON(execErrorBytes)
	assert.Nil(t, uErr)

	assert.Equal(t, execError.Code, execErrorObj.Code)
	assert.Equal(t, execError.Message, execError.Message)
	assert.Equal(t, execError.ErrorUri, execErrorObj.ErrorUri)
}
