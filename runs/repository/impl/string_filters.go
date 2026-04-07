package impl

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

const (
	filterSeparator = "+"
	valueSeparator  = ";"
)

// ParseStringFilters parses a compact filter DSL like eq(state,1)+contains(name,foo).
func ParseStringFilters(filterString string, allowedColumns sets.Set[string]) (interfaces.Filter, error) {
	filterString = strings.TrimSpace(filterString)
	if filterString == "" {
		return nil, nil
	}

	var combined interfaces.Filter
	for _, expr := range strings.Split(filterString, filterSeparator) {
		filter, err := parseSingleStringFilter(expr, allowedColumns)
		if err != nil {
			return nil, err
		}
		if combined == nil {
			combined = filter
		} else {
			combined = combined.And(filter)
		}
	}

	return combined, nil
}

func parseSingleStringFilter(expr string, allowedColumns sets.Set[string]) (interfaces.Filter, error) {
	expr = strings.TrimSpace(expr)
	open := strings.IndexByte(expr, '(')
	close := strings.LastIndexByte(expr, ')')
	if open <= 0 || close != len(expr)-1 {
		return nil, fmt.Errorf("invalid filter format: %s", expr)
	}

	fn := strings.ToLower(strings.TrimSpace(expr[:open]))
	args := expr[open+1 : close]
	parts := strings.SplitN(args, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid filter arguments: %s", expr)
	}

	field := strings.TrimSpace(parts[0])
	if !allowedColumns.Has(field) {
		return nil, fmt.Errorf("invalid filter field: %s", field)
	}

	rawValue := strings.TrimSpace(parts[1])
	switch fn {
	case "eq":
		return NewEqualFilter(field, rawValue), nil
	case "ne":
		return NewNotEqualFilter(field, rawValue), nil
	case "contains":
		return &basicFilter{
			field:      field,
			expression: interfaces.FilterExpressionContains,
			value:      rawValue,
		}, nil
	case "value_in":
		rawValues := strings.Split(rawValue, valueSeparator)
		values := make([]string, 0, len(rawValues))
		for _, v := range rawValues {
			values = append(values, strings.TrimSpace(v))
		}
		return &basicFilter{
			field:      field,
			expression: interfaces.FilterExpressionValueIn,
			value:      values,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported filter function: %s", fn)
	}
}
