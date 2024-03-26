package errors

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/pkg/errors"
)

func mustErrorCode(t *testing.T, compileError *CompileError, code ErrorCode) {
	assert.Equal(t, code, compileError.Code())
}

func TestErrorCodes(t *testing.T) {
	testCases := map[ErrorCode]*CompileError{
		CycleDetected:              NewCycleDetectedInWorkflowErr("", ""),
		BranchNodeIDNotFound:       NewBranchNodeNotSpecified(""),
		BranchNodeHasNoCondition:   NewBranchNodeHasNoCondition(""),
		ValueRequired:              NewValueRequiredErr("", ""),
		NodeReferenceNotFound:      NewNodeReferenceNotFoundErr("", ""),
		TaskReferenceNotFound:      NewTaskReferenceNotFoundErr("", ""),
		WorkflowReferenceNotFound:  NewWorkflowReferenceNotFoundErr("", ""),
		VariableNameNotFound:       NewVariableNameNotFoundErr("", "", ""),
		DuplicateAlias:             NewDuplicateAliasErr("", ""),
		DuplicateNodeID:            NewDuplicateIDFoundErr(""),
		MismatchingTypes:           NewMismatchingTypesErr("", "", "", ""),
		MismatchingInterfaces:      NewMismatchingInterfacesErr("", ""),
		InconsistentTypes:          NewInconsistentTypesErr("", "", ""),
		ParameterBoundMoreThanOnce: NewParameterBoundMoreThanOnceErr("", ""),
		ParameterNotBound:          NewParameterNotBoundErr("", ""),
		NoEntryNodeFound:           NewWorkflowHasNoEntryNodeErr(""),
		UnreachableNodes:           NewUnreachableNodesErr("", ""),
		UnrecognizedValue:          NewUnrecognizedValueErr("", ""),
		WorkflowBuildError:         NewWorkflowBuildError(errors.New("")),
		NoNodesFound:               NewNoNodesFoundErr(""),
	}

	for key, value := range testCases {
		t.Run(string(key), func(t *testing.T) {
			mustErrorCode(t, value, key)
		})
	}
}

func TestIncludeSource(t *testing.T) {
	e := NewCycleDetectedInWorkflowErr("", "")
	assert.Equal(t, e.source, "")

	SetConfig(Config{IncludeSource: true})
	e = NewCycleDetectedInWorkflowErr("", "")
	assert.Equal(t, e.source, "compiler_error_test.go:50")
	SetConfig(Config{})
}
