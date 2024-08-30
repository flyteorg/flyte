package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"
)

func TestValidateInputs_IDLTypeNotFound(t *testing.T) {
	nodeID := common.NodeID("test-node")

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
			"input1": &core.Literal{}, // This will cause LiteralTypeForLiteral to return nil
		},
	}

	errs := errors.NewCompileErrors()
	ok := validateInputs(nodeID, iface, inputs, errs)

	assert.False(t, ok)
	assert.True(t, errs.HasErrors())

	idlNotFound := false
	var errMsg string
	for _, err := range errs.Errors().List() {
		if err.Code() == "IDLTypeNotFound" {
			idlNotFound = true
			errMsg = err.Error()
			break
		}
	}
	assert.True(t, idlNotFound, "Expected IDLTypeNotFound error was not found in errors")

	expectedErrMsg := "Code: IDLTypeNotFound, Node Id: test-node, Description: Input type IDL is not found, please update all of your Flyte images to the latest version and try again."
	assert.Equal(t, expectedErrMsg, errMsg, "Error message does not match expected value")
}
