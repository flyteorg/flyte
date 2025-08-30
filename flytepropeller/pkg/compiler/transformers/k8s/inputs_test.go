package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"
)

func TestValidateInputs_InvalidLiteralType(t *testing.T) {
	nodeID := "test-node"

	iface := &core.TypedInterface{
		Inputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"input1": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: 1000,
						},
					},
				},
			},
		},
	}

	inputs := core.LiteralMap{
		Literals: map[string]*core.Literal{
			"input1": nil, // Set this to nil to trigger the nil case
		},
	}

	errs := errors.NewCompileErrors()
	ok := validateInputs(nodeID, iface, inputs, errs)

	assert.False(t, ok)
	assert.True(t, errs.HasErrors())

	idlNotFound := false
	var errMsg string
	for _, err := range errs.Errors().List() {
		if err.Code() == "InvalidLiteralType" {
			idlNotFound = true
			errMsg = err.Error()
			break
		}
	}
	assert.True(t, idlNotFound, "Expected InvalidLiteralType error was not found in errors")

	expectedContainedErrorMsg := "Failed to validate literal type"
	assert.Contains(t, errMsg, expectedContainedErrorMsg)
}
