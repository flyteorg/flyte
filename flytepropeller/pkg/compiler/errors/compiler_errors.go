package errors

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	// A cycle is detected in the Workflow, the error description should detail the nodes involved.
	CycleDetected ErrorCode = "CycleDetected"

	// BranchNode is missing a case with ThenNode populated.
	BranchNodeIDNotFound ErrorCode = "BranchNodeIdNotFound"

	// BranchNode is missing a condition.
	BranchNodeHasNoCondition ErrorCode = "BranchNodeHasNoCondition"

	// BranchNode is missing the else case & the else fail
	BranchNodeHasNoDefault ErrorCode = "BranchNodeHasNoDefault"

	// An expected field isn't populated.
	ValueRequired ErrorCode = "ValueRequired"

	// An expected field is malformed or contains invalid inputs
	InvalidValue ErrorCode = "InvalidValue"

	// A nodeBuilder referenced by an edge doesn't belong to the Workflow.
	NodeReferenceNotFound ErrorCode = "NodeReferenceNotFound"

	// A Task referenced by a node wasn't found.
	TaskReferenceNotFound ErrorCode = "TaskReferenceNotFound"

	// A Workflow referenced by a node wasn't found.
	WorkflowReferenceNotFound ErrorCode = "WorkflowReferenceNotFound"

	// A referenced variable (in a parameter or a condition) wasn't found.
	VariableNameNotFound ErrorCode = "VariableNameNotFound"

	// An alias existed twice.
	DuplicateAlias ErrorCode = "DuplicateAlias"

	// An Id existed twice.
	DuplicateNodeID ErrorCode = "DuplicateId"

	// Two types expected to be compatible but aren't.
	MismatchingTypes ErrorCode = "MismatchingTypes"

	// A binding is attempted via a list or map syntax, but the underlying type isn't a list or map.
	MismatchingBindings ErrorCode = "MismatchingBindings"

	// Two interfaced expected to be compatible but aren't.
	MismatchingInterfaces ErrorCode = "MismatchingInterfaces"

	// Expected types to be consistent.
	InconsistentTypes ErrorCode = "InconsistentTypes"

	// An input/output parameter was assigned a value through an edge more than once.
	ParameterBoundMoreThanOnce ErrorCode = "ParameterBoundMoreThanOnce"

	// One of the required input parameters or a Workflow output parameter wasn't bound.
	ParameterNotBound ErrorCode = "ParameterNotBound"

	// When we couldn't assign an entry point to the Workflow.
	NoEntryNodeFound ErrorCode = "NoEntryNodeFound"

	// When one more more unreachable node are detected.
	UnreachableNodes ErrorCode = "UnreachableNodes"

	// A Value doesn't fall within the expected range.
	UnrecognizedValue ErrorCode = "UnrecognizedValue"

	// An unknown error occurred while building the workflow.
	WorkflowBuildError ErrorCode = "WorkflowBuildError"

	// A value is expected to be unique but wasnt.
	ValueCollision ErrorCode = "ValueCollision"

	// A value isn't on the right syntax.
	SyntaxError ErrorCode = "SyntaxError"

	// A workflow is missing any nodes to execute
	NoNodesFound ErrorCode = "NoNodesFound"

	// Given value is not a legal Enum value (or not part of the defined set of enum values)
	IllegalEnumValue ErrorCode = "IllegalEnumValue"

	// Given value cannot be unambiguously assigned to a union variant in a binding
	AmbiguousBindingUnionValue ErrorCode = "AmbiguousBindingUnionValue"

	// Given value cannot be assigned to any union variant in a binding
	IncompatibleBindingUnionValue ErrorCode = "IncompatibleBindingUnionValue"

	// A gate node is missing a condition.
	NoConditionFound ErrorCode = "NoConditionFound"
)

func NewBranchNodeNotSpecified(branchNodeID string) *CompileError {
	return newError(
		BranchNodeIDNotFound,
		"BranchNode not assigned",
		branchNodeID,
	)
}

func NewBranchNodeHasNoCondition(branchNodeID string) *CompileError {
	return newError(
		BranchNodeHasNoCondition,
		"One of the branches on the node doesn't have a condition.",
		branchNodeID,
	)
}

func NewBranchNodeHasNoDefault(branchNodeID string) *CompileError {
	return newError(
		BranchNodeHasNoDefault,
		"Branch Node must have either the else case set or a default error.",
		branchNodeID,
	)
}

func NewValueRequiredErr(nodeID, paramName string) *CompileError {
	return newError(
		ValueRequired,
		fmt.Sprintf("Value required [%v].", paramName),
		nodeID,
	)
}

func NewInvalidValueErr(nodeID, paramName string) *CompileError {
	return newError(
		InvalidValue,
		fmt.Sprintf("Invalid value [%v].", paramName),
		nodeID,
	)
}

func NewParameterNotBoundErr(nodeID, paramName string) *CompileError {
	return newError(
		ParameterNotBound,
		fmt.Sprintf("Parameter not bound [%v].", paramName),
		nodeID,
	)
}

func NewNodeReferenceNotFoundErr(nodeID, referenceID string) *CompileError {
	return newError(
		NodeReferenceNotFound,
		fmt.Sprintf("Referenced node [%v] not found.", referenceID),
		nodeID,
	)
}

func NewWorkflowReferenceNotFoundErr(nodeID, referenceID string) *CompileError {
	return newError(
		WorkflowReferenceNotFound,
		fmt.Sprintf("Referenced Workflow [%v] not found.", referenceID),
		nodeID,
	)
}

func NewTaskReferenceNotFoundErr(nodeID, referenceID string) *CompileError {
	return newError(
		TaskReferenceNotFound,
		fmt.Sprintf("Referenced Task [%v] not found.", referenceID),
		nodeID,
	)
}

func NewVariableNameNotFoundErr(nodeID, referenceID, variableName string) *CompileError {
	return newError(
		VariableNameNotFound,
		fmt.Sprintf("Variable [%v] not found on node [%v].", variableName, referenceID),
		nodeID,
	)
}

func NewParameterBoundMoreThanOnceErr(nodeID, paramName string) *CompileError {
	return newError(
		ParameterBoundMoreThanOnce,
		fmt.Sprintf("Input [%v] is bound more than once.", paramName),
		nodeID,
	)
}

func NewDuplicateAliasErr(nodeID, alias string) *CompileError {
	return newError(
		DuplicateAlias,
		fmt.Sprintf("Duplicate alias [%v] found. An output alias can only be used once in the Workflow.", alias),
		nodeID,
	)
}

func NewDuplicateIDFoundErr(nodeID string) *CompileError {
	return newError(
		DuplicateNodeID,
		"Trying to insert two nodes with the same id.",
		nodeID,
	)
}

func NewMismatchingTypesErr(nodeID, fromVar, fromType, toType string) *CompileError {
	return newError(
		MismatchingTypes,
		fmt.Sprintf("Variable [%v] (type [%v]) doesn't match expected type [%v].", fromVar, fromType,
			toType),
		nodeID,
	)
}

func NewMismatchingBindingsErr(nodeID, sinkParam, expectedType, receivedType string) *CompileError {
	return newError(
		MismatchingBindings,
		fmt.Sprintf("Input [%v] on node [%v] expects bindings of type [%v].  Received [%v]", sinkParam, nodeID, expectedType, receivedType),
		nodeID,
	)
}

func NewIllegalEnumValueError(nodeID, sinkParam, receivedVal string, expectedVals []string) *CompileError {
	return newError(
		IllegalEnumValue,
		fmt.Sprintf("Input [%v] on node [%v] is an Enum and expects value to be one of [%v].  Received [%v]", sinkParam, nodeID, expectedVals, receivedVal),
		nodeID,
	)
}

func NewMismatchingInterfacesErr(nodeID1, nodeID2 string) *CompileError {
	return newError(
		MismatchingInterfaces,
		fmt.Sprintf("Interfaces of nodes [%v] and [%v] do not match.", nodeID1, nodeID2),
		nodeID1,
	)
}

func NewInconsistentTypesErr(nodeID, expectedType, actualType string) *CompileError {
	return newError(
		InconsistentTypes,
		fmt.Sprintf("Expected type: %v but found %v", expectedType, actualType),
		nodeID,
	)
}

func NewWorkflowHasNoEntryNodeErr(graphID string) *CompileError {
	return newError(
		NoEntryNodeFound,
		fmt.Sprintf("Can't find a node to start executing Workflow [%v].", graphID),
		graphID,
	)
}

func NewCycleDetectedInWorkflowErr(nodeID, cycle string) *CompileError {
	return newError(
		CycleDetected,
		fmt.Sprintf("A cycle has been detected while traversing the Workflow [%v].", cycle),
		nodeID,
	)
}

func NewUnreachableNodesErr(nodeID, nodes string) *CompileError {
	return newError(
		UnreachableNodes,
		fmt.Sprintf("The Workflow contain unreachable nodes [%v].", nodes),
		nodeID,
	)
}

func NewUnrecognizedValueErr(nodeID, value string) *CompileError {
	return newError(
		UnrecognizedValue,
		fmt.Sprintf("Unrecognized value [%v].", value),
		nodeID,
	)
}

func NewWorkflowBuildError(err error) *CompileError {
	return newError(WorkflowBuildError, err.Error(), "")
}

func NewValueCollisionError(nodeID string, valueName, value string) *CompileError {
	return newError(
		ValueCollision,
		fmt.Sprintf("%v is expected to be unique. %v already exists.", valueName, value),
		nodeID,
	)
}

func NewSyntaxError(nodeID string, element string, err error) *CompileError {
	return newError(SyntaxError,
		fmt.Sprintf("Failed to parse element [%v].", element),
		nodeID,
	)
}

func NewNoNodesFoundErr(graphID string) *CompileError {
	return newError(
		NoNodesFound,
		fmt.Sprintf("Can't find any nodes in workflow [%v].", graphID),
		graphID,
	)
}

func NewAmbiguousBindingUnionValue(nodeID, sinkParam, expectedType, binding, match1, match2 string) *CompileError {
	return newError(
		AmbiguousBindingUnionValue,
		fmt.Sprintf(
			"Input [%v] on node [%v] expects bindings of union type [%v].  Received [%v] which is ambiguous as both variants [%v] and [%v] match.",
			sinkParam, nodeID, expectedType, binding, match1, match2),
		nodeID,
	)
}

func NewIncompatibleBindingUnionValue(nodeID, sinkParam, expectedType, binding string) *CompileError {
	return newError(
		IncompatibleBindingUnionValue,
		fmt.Sprintf(
			"Input [%v] on node [%v] expects bindings of union type [%v].  Received [%v] which does not match any of the variants.",
			sinkParam, nodeID, expectedType, binding),
		nodeID,
	)
}

func NewNoConditionFound(nodeID string) *CompileError {
	return newError(
		NoConditionFound,
		fmt.Sprintf("Can't find any condition in gate node [%v].", nodeID),
		nodeID,
	)
}

func newError(code ErrorCode, description, nodeID string) (err *CompileError) {
	err = &CompileError{
		code:        code,
		description: description,
		nodeID:      nodeID,
	}

	if GetConfig().IncludeSource {
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "???"
			line = 1
		} else {
			slash := strings.LastIndex(file, "/")
			if slash >= 0 {
				file = file[slash+1:]
			}
		}

		err.source = fmt.Sprintf("%v:%v", file, line)
	}

	return
}
