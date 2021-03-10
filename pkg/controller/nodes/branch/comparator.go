package branch

import (
	"reflect"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/errors"
)

type comparator func(lValue *core.Primitive, rValue *core.Primitive) bool
type comparators struct {
	gt comparator
	eq comparator
}

var primitiveBooleanType = reflect.TypeOf(&core.Primitive_Boolean{}).String()

var perTypeComparators = map[string]comparators{
	reflect.TypeOf(&core.Primitive_FloatValue{}).String(): {
		gt: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetFloatValue() > rValue.GetFloatValue()
		},
		eq: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetFloatValue() == rValue.GetFloatValue()
		},
	},
	reflect.TypeOf(&core.Primitive_Integer{}).String(): {
		gt: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetInteger() > rValue.GetInteger()
		},
		eq: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetInteger() == rValue.GetInteger()
		},
	},
	reflect.TypeOf(&core.Primitive_Boolean{}).String(): {
		eq: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetBoolean() == rValue.GetBoolean()
		},
	},
	reflect.TypeOf(&core.Primitive_StringValue{}).String(): {
		gt: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetStringValue() > rValue.GetStringValue()
		},
		eq: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetStringValue() == rValue.GetStringValue()
		},
	},
	reflect.TypeOf(&core.Primitive_StringValue{}).String(): {
		gt: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetStringValue() > rValue.GetStringValue()
		},
		eq: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetStringValue() == rValue.GetStringValue()
		},
	},
	reflect.TypeOf(&core.Primitive_Datetime{}).String(): {
		gt: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetDatetime().GetSeconds() > rValue.GetDatetime().GetSeconds()
		},
		eq: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetDatetime().GetSeconds() == rValue.GetDatetime().GetSeconds()
		},
	},
	reflect.TypeOf(&core.Primitive_Duration{}).String(): {
		gt: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetDuration().GetSeconds() > rValue.GetDuration().GetSeconds()
		},
		eq: func(lValue *core.Primitive, rValue *core.Primitive) bool {
			return lValue.GetDuration().GetSeconds() == rValue.GetDuration().GetSeconds()
		},
	},
}

func Evaluate(lValue *core.Primitive, rValue *core.Primitive, op core.ComparisonExpression_Operator) (bool, error) {
	lValueType := reflect.TypeOf(lValue.Value)
	rValueType := reflect.TypeOf(rValue.Value)
	if lValueType != rValueType {
		return false, errors.Errorf(ErrorCodeMalformedBranch, "Comparison between different primitives types. lVal[%v]:rVal[%v]", lValueType, rValueType)
	}
	comps, ok := perTypeComparators[lValueType.String()]
	if !ok {
		return false, errors.Errorf("Comparator not defined for type: [%v]", lValueType.String())
	}
	isBoolean := false
	if lValueType.String() == primitiveBooleanType {
		isBoolean = true
	}
	switch op {
	case core.ComparisonExpression_GT:
		if isBoolean {
			return false, errors.Errorf(ErrorCodeMalformedBranch, "[GT] not defined for boolean operands.")
		}
		return comps.gt(lValue, rValue), nil
	case core.ComparisonExpression_GTE:
		if isBoolean {
			return false, errors.Errorf(ErrorCodeMalformedBranch, "[GTE] not defined for boolean operands.")
		}
		return comps.eq(lValue, rValue) || comps.gt(lValue, rValue), nil
	case core.ComparisonExpression_LT:
		if isBoolean {
			return false, errors.Errorf(ErrorCodeMalformedBranch, "[LT] not defined for boolean operands.")
		}
		return !(comps.gt(lValue, rValue) || comps.eq(lValue, rValue)), nil
	case core.ComparisonExpression_LTE:
		if isBoolean {
			return false, errors.Errorf(ErrorCodeMalformedBranch, "[LTE] not defined for boolean operands.")
		}
		return !comps.gt(lValue, rValue), nil
	case core.ComparisonExpression_EQ:
		return comps.eq(lValue, rValue), nil
	case core.ComparisonExpression_NEQ:
		return !comps.eq(lValue, rValue), nil
	}
	return false, errors.Errorf(ErrorCodeMalformedBranch, "Unsupported operator type in Propeller. System error.")
}

func Evaluate1(lValue *core.Primitive, rValue *core.Literal, op core.ComparisonExpression_Operator) (bool, error) {
	if rValue.GetScalar() == nil || rValue.GetScalar().GetPrimitive() == nil {
		return false, errors.Errorf(ErrorCodeMalformedBranch, "Only primitives can be compared. RHS Variable is non primitive.")
	}
	return Evaluate(lValue, rValue.GetScalar().GetPrimitive(), op)
}

func Evaluate2(lValue *core.Literal, rValue *core.Primitive, op core.ComparisonExpression_Operator) (bool, error) {
	if lValue.GetScalar() == nil || lValue.GetScalar().GetPrimitive() == nil {
		return false, errors.Errorf(ErrorCodeMalformedBranch, "Only primitives can be compared. LHS Variable is non primitive.")
	}
	return Evaluate(lValue.GetScalar().GetPrimitive(), rValue, op)
}

func EvaluateLiterals(lValue *core.Literal, rValue *core.Literal, op core.ComparisonExpression_Operator) (bool, error) {
	if lValue.GetScalar() == nil || lValue.GetScalar().GetPrimitive() == nil {
		return false, errors.Errorf(ErrorCodeMalformedBranch, "Only primitives can be compared. LHS Variable is non primitive.")
	}
	if rValue.GetScalar() == nil || rValue.GetScalar().GetPrimitive() == nil {
		return false, errors.Errorf(ErrorCodeMalformedBranch, "Only primitives can be compared. RHS Variable is non primitive")
	}
	return Evaluate(lValue.GetScalar().GetPrimitive(), rValue.GetScalar().GetPrimitive(), op)
}
