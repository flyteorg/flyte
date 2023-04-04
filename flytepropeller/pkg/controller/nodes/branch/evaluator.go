package branch

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
)

const ErrorCodeUserProvidedError = "UserProvidedError"
const ErrorCodeMalformedBranch = "MalformedBranchUserError"
const ErrorCodeCompilerError = "CompilerError"
const ErrorCodeFailedFetchOutputs = "FailedFetchOutputs"

func EvaluateComparison(expr *core.ComparisonExpression, nodeInputs *core.LiteralMap) (bool, error) {
	var lValue *core.Literal
	var rValue *core.Literal
	var lPrim *core.Primitive
	var rPrim *core.Primitive

	if expr.GetLeftValue().GetPrimitive() == nil {
		if nodeInputs == nil {
			return false, errors.Errorf(ErrorCodeMalformedBranch, "Failed to find Value for Variable [%v]", expr.GetLeftValue().GetVar())
		}
		lValue = nodeInputs.Literals[expr.GetLeftValue().GetVar()]
		if lValue == nil {
			return false, errors.Errorf(ErrorCodeMalformedBranch, "Failed to find Value for Variable [%v]", expr.GetLeftValue().GetVar())
		}
	} else {
		lPrim = expr.GetLeftValue().GetPrimitive()
	}

	if expr.GetRightValue().GetPrimitive() == nil {
		if nodeInputs == nil {
			return false, errors.Errorf(ErrorCodeMalformedBranch, "Failed to find Value for Variable [%v]", expr.GetLeftValue().GetVar())
		}
		rValue = nodeInputs.Literals[expr.GetRightValue().GetVar()]
		if rValue == nil {
			return false, errors.Errorf(ErrorCodeMalformedBranch, "Failed to find Value for Variable [%v]", expr.GetRightValue().GetVar())
		}
	} else {
		rPrim = expr.GetRightValue().GetPrimitive()
	}

	if lValue != nil && rValue != nil {
		return EvaluateLiterals(lValue, rValue, expr.GetOperator())
	}
	if lValue != nil && rPrim != nil {
		return Evaluate2(lValue, rPrim, expr.GetOperator())
	}
	if lPrim != nil && rValue != nil {
		return Evaluate1(lPrim, rValue, expr.GetOperator())
	}
	return Evaluate(lPrim, rPrim, expr.GetOperator())
}

func EvaluateBooleanExpression(expr *core.BooleanExpression, nodeInputs *core.LiteralMap) (bool, error) {
	if expr.GetComparison() != nil {
		return EvaluateComparison(expr.GetComparison(), nodeInputs)
	}
	if expr.GetConjunction() == nil {
		return false, errors.Errorf(ErrorCodeMalformedBranch, "No Comparison or Conjunction found in Branch node expression.")
	}
	lvalue, err := EvaluateBooleanExpression(expr.GetConjunction().GetLeftExpression(), nodeInputs)
	if err != nil {
		return false, err
	}
	rvalue, err := EvaluateBooleanExpression(expr.GetConjunction().GetRightExpression(), nodeInputs)
	if err != nil {
		return false, err
	}
	if expr.GetConjunction().GetOperator() == core.ConjunctionExpression_OR {
		return lvalue || rvalue, nil
	}
	return lvalue && rvalue, nil
}

func EvaluateIfBlock(block v1alpha1.ExecutableIfBlock, nodeInputs *core.LiteralMap, skippedNodeIds []*v1alpha1.NodeID) (*v1alpha1.NodeID, []*v1alpha1.NodeID, error) {
	if ok, err := EvaluateBooleanExpression(block.GetCondition(), nodeInputs); err != nil {
		return nil, skippedNodeIds, err
	} else if ok {
		// Set status to running
		return block.GetThenNode(), skippedNodeIds, err
	}
	// This branch is not taken
	return nil, append(skippedNodeIds, block.GetThenNode()), nil
}

// Decides the branch to be taken, returns the nodeId of the selected node or an error
// The branchNode is marked as success. This is used by downstream node to determine if it can be executed
// All downstream nodes are marked as skipped
func DecideBranch(ctx context.Context, nl executors.NodeLookup, nodeID v1alpha1.NodeID, node v1alpha1.ExecutableBranchNode, nodeInputs *core.LiteralMap) (*v1alpha1.NodeID, error) {
	var selectedNodeID *v1alpha1.NodeID
	var skippedNodeIds []*v1alpha1.NodeID
	var err error

	selectedNodeID, skippedNodeIds, err = EvaluateIfBlock(node.GetIf(), nodeInputs, skippedNodeIds)
	if err != nil {
		return nil, err
	}

	for _, block := range node.GetElseIf() {
		if selectedNodeID != nil {
			skippedNodeIds = append(skippedNodeIds, block.GetThenNode())
		} else {
			selectedNodeID, skippedNodeIds, err = EvaluateIfBlock(block, nodeInputs, skippedNodeIds)
			if err != nil {
				return nil, err
			}
		}
	}
	if node.GetElse() != nil {
		if selectedNodeID == nil {
			selectedNodeID = node.GetElse()
		} else {
			skippedNodeIds = append(skippedNodeIds, node.GetElse())
		}
	}
	for _, nodeIDPtr := range skippedNodeIds {
		skippedNodeID := *nodeIDPtr
		n, ok := nl.GetNode(skippedNodeID)
		if !ok {
			return nil, errors.Errorf(ErrorCodeCompilerError, "Downstream node [%v] not found", skippedNodeID)
		}
		nStatus := nl.GetNodeExecutionStatus(ctx, n.GetID())
		logger.Infof(ctx, "Branch Setting Node[%v] status to Skipped!", skippedNodeID)
		nStatus.UpdatePhase(v1alpha1.NodePhaseSkipped, v1.Now(), "Branch evaluated to false", nil)
	}

	if selectedNodeID == nil {
		if node.GetElseFail() != nil {
			return nil, errors.Errorf(ErrorCodeUserProvidedError, node.GetElseFail().Message)
		}
		return nil, errors.Errorf(ErrorCodeMalformedBranch, "No branch satisfied")
	}
	logger.Infof(ctx, "Branch Node[%v] selected!", *selectedNodeID)
	return selectedNodeID, nil
}
