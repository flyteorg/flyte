package validators

import (
	"fmt"

	flyte "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
)

func validateOperand(node c.NodeBuilder, paramName string, operand *flyte.Operand,
	requireParamType bool, errs errors.CompileErrors) (literalType *flyte.LiteralType, ok bool) {
	if operand == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), paramName))
	} else if operand.GetPrimitive() != nil {
		// no validation
		literalType = literalTypeForPrimitive(operand.GetPrimitive())
	} else if len(operand.GetVar()) > 0 {
		if node.GetInterface() != nil {
			if param, paramOk := validateInputVar(node, operand.GetVar(), requireParamType, errs.NewScope()); paramOk {
				if param != nil {
					literalType = param.GetType()
				}
			}
		} else {
			errs.Collect(errors.NewValueRequiredErr(node.GetId(), operand.GetVar()))
		}
	} else {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), fmt.Sprintf("%v.%v", paramName, "Val")))
	}

	return literalType, !errs.HasErrors()
}

func ValidateBooleanExpression(w c.WorkflowBuilder, node c.NodeBuilder, expr *flyte.BooleanExpression, requireParamType bool, errs errors.CompileErrors) (ok bool) {
	if expr == nil {
		errs.Collect(errors.NewBranchNodeHasNoCondition(node.GetId()))
	} else {
		if expr.GetComparison() != nil {
			op1Type, op1Valid := validateOperand(node, "RightValue",
				expr.GetComparison().GetRightValue(), requireParamType, errs.NewScope())
			op2Type, op2Valid := validateOperand(node, "LeftValue",
				expr.GetComparison().GetLeftValue(), requireParamType, errs.NewScope())
			if op1Valid && op2Valid && op1Type != nil && op2Type != nil {
				if op1Type.String() != op2Type.String() {
					errs.Collect(errors.NewMismatchingTypesErr(node.GetId(), "RightValue",
						op1Type.String(), op2Type.String()))
				}
			}
		} else if expr.GetConjunction() != nil {
			ValidateBooleanExpression(w, node, expr.GetConjunction().LeftExpression, requireParamType, errs.NewScope())
			ValidateBooleanExpression(w, node, expr.GetConjunction().RightExpression, requireParamType, errs.NewScope())
		} else {
			errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Expr"))
		}
	}

	return !errs.HasErrors()
}
