package validators

import (
	"fmt"

	flyte "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
)

func validateOperand(node c.NodeBuilder, paramName string, operand *flyte.Operand,
	errs errors.CompileErrors) (literalType *flyte.LiteralType, ok bool) {
	if operand == nil {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), paramName))
	} else if operand.GetPrimitive() != nil {
		// no validation
		literalType = literalTypeForPrimitive(operand.GetPrimitive())
	} else if operand.GetVar() != "" {
		if node.GetInterface() != nil {
			if param, paramOk := validateInputVar(node, operand.GetVar(), errs.NewScope()); paramOk {
				literalType = param.GetType()
			}
		}
	} else {
		errs.Collect(errors.NewValueRequiredErr(node.GetId(), fmt.Sprintf("%v.%v", paramName, "Val")))
	}

	return literalType, !errs.HasErrors()
}

func ValidateBooleanExpression(node c.NodeBuilder, expr *flyte.BooleanExpression, errs errors.CompileErrors) (ok bool) {
	if expr == nil {
		errs.Collect(errors.NewBranchNodeHasNoCondition(node.GetId()))
	} else {
		if expr.GetComparison() != nil {
			op1Type, op1Valid := validateOperand(node, "RightValue",
				expr.GetComparison().GetRightValue(), errs.NewScope())
			op2Type, op2Valid := validateOperand(node, "LeftValue",
				expr.GetComparison().GetLeftValue(), errs.NewScope())
			if op1Valid && op2Valid {
				if op1Type.String() != op2Type.String() {
					errs.Collect(errors.NewMismatchingTypesErr(node.GetId(), "RightValue",
						op1Type.String(), op2Type.String()))
				}
			}
		} else if expr.GetConjunction() != nil {
			ValidateBooleanExpression(node, expr.GetConjunction().LeftExpression, errs.NewScope())
			ValidateBooleanExpression(node, expr.GetConjunction().RightExpression, errs.NewScope())
		} else {
			errs.Collect(errors.NewValueRequiredErr(node.GetId(), "Expr"))
		}
	}

	return !errs.HasErrors()
}
