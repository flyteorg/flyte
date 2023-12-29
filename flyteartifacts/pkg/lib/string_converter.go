package lib

import (
	"fmt"
	"strings"
	"time"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const DateFormat string = "2006-01-02"

func RenderLiteral(lit *core.Literal) (string, error) {
	if lit == nil {
		return "", fmt.Errorf("can't RenderLiteral, input is nil")
	}

	switch lit.Value.(type) {
	case *core.Literal_Scalar:
		scalar := lit.GetScalar()
		if scalar.GetPrimitive() == nil {
			return "", fmt.Errorf("rendering only works for primitives, got [%v]", scalar)
		}
		// todo: figure out how to expose more formatting
		// todo: maybe add a metric to each one of these, or this whole block.
		switch scalar.GetPrimitive().GetValue().(type) {
		case *core.Primitive_StringValue:
			return scalar.GetPrimitive().GetStringValue(), nil
		case *core.Primitive_Integer:
			return fmt.Sprintf("%d", scalar.GetPrimitive().GetInteger()), nil
		case *core.Primitive_FloatValue:
			return fmt.Sprintf("%v", scalar.GetPrimitive().GetFloatValue()), nil
		case *core.Primitive_Boolean:
			if scalar.GetPrimitive().GetBoolean() {
				return "true", nil
			}
			return "false", nil
		case *core.Primitive_Datetime:
			// just date for now, not sure if we should support time...
			dt := scalar.GetPrimitive().GetDatetime().AsTime()
			txt := dt.Format(DateFormat)
			return txt, nil
		case *core.Primitive_Duration:
			dur := scalar.GetPrimitive().GetDuration().AsDuration()
			// Found somewhere as iso8601 representation of duration, but there's still lots of
			// possibilities for formatting.
			txt := "PT" + strings.ToUpper(dur.Truncate(time.Millisecond).String())
			return txt, nil
		default:
			return "", fmt.Errorf("unknown primitive type [%v]", scalar.GetPrimitive())
		}
	case *core.Literal_Collection:
		return "", fmt.Errorf("can't RenderLiteral for collections")
	case *core.Literal_Map:
		return "", fmt.Errorf("can't RenderLiteral for maps")
	}

	return "", fmt.Errorf("unknown literal type [%v]", lit)
}
