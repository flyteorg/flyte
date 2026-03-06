package impl

import (
	"fmt"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// basicFilter implements the Filter interface for simple field comparisons
type basicFilter struct {
	field      string
	expression interfaces.FilterExpression
	value      interface{}
}

func (f *basicFilter) GormQueryExpression(table string) (interfaces.GormQueryExpr, error) {
	var query string
	column := f.field
	if table != "" {
		column = table + "." + f.field
	}

	switch f.expression {
	case interfaces.FilterExpressionEqual:
		query = fmt.Sprintf("%s = ?", column)
	case interfaces.FilterExpressionNotEqual:
		query = fmt.Sprintf("%s != ?", column)
	case interfaces.FilterExpressionGreaterThan:
		query = fmt.Sprintf("%s > ?", column)
	case interfaces.FilterExpressionGreaterThanOrEqual:
		query = fmt.Sprintf("%s >= ?", column)
	case interfaces.FilterExpressionLessThan:
		query = fmt.Sprintf("%s < ?", column)
	case interfaces.FilterExpressionLessThanOrEqual:
		query = fmt.Sprintf("%s <= ?", column)
	case interfaces.FilterExpressionContains:
		query = fmt.Sprintf("%s LIKE ?", column)
		f.value = fmt.Sprintf("%%%v%%", f.value)
	case interfaces.FilterExpressionValueIn:
		query = fmt.Sprintf("%s IN ?", column)
	case interfaces.FilterExpressionEndsWith:
		query = fmt.Sprintf("%s LIKE ?", column)
		f.value = fmt.Sprintf("%%%v", f.value)
	case interfaces.FilterExpressionNotEndsWith:
		query = fmt.Sprintf("%s NOT LIKE ?", column)
		f.value = fmt.Sprintf("%%%v", f.value)
	case interfaces.FilterExpressionContainsCaseInsensitive:
		query = fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", column)
		f.value = fmt.Sprintf("%%%v%%", f.value)
	default:
		return interfaces.GormQueryExpr{}, fmt.Errorf("unsupported filter expression: %d", f.expression)
	}

	return interfaces.GormQueryExpr{
		Query: query,
		Args:  []interface{}{f.value},
	}, nil
}

func (f *basicFilter) And(filter interfaces.Filter) interfaces.Filter {
	return &compositeFilter{
		left:     f,
		right:    filter,
		operator: "AND",
	}
}

func (f *basicFilter) Or(filter interfaces.Filter) interfaces.Filter {
	return &compositeFilter{
		left:     f,
		right:    filter,
		operator: "OR",
	}
}

// compositeFilter implements the Filter interface for AND/OR operations
type compositeFilter struct {
	left     interfaces.Filter
	right    interfaces.Filter
	operator string // "AND" or "OR"
}

func (f *compositeFilter) GormQueryExpression(table string) (interfaces.GormQueryExpr, error) {
	leftExpr, err := f.left.GormQueryExpression(table)
	if err != nil {
		return interfaces.GormQueryExpr{}, err
	}

	rightExpr, err := f.right.GormQueryExpression(table)
	if err != nil {
		return interfaces.GormQueryExpr{}, err
	}

	query := fmt.Sprintf("(%s) %s (%s)", leftExpr.Query, f.operator, rightExpr.Query)
	args := append(leftExpr.Args, rightExpr.Args...)

	return interfaces.GormQueryExpr{
		Query: query,
		Args:  args,
	}, nil
}

func (f *compositeFilter) And(filter interfaces.Filter) interfaces.Filter {
	return &compositeFilter{
		left:     f,
		right:    filter,
		operator: "AND",
	}
}

func (f *compositeFilter) Or(filter interfaces.Filter) interfaces.Filter {
	return &compositeFilter{
		left:     f,
		right:    filter,
		operator: "OR",
	}
}

// Helper functions to create filters

// NewEqualFilter creates a filter for field = value
func NewEqualFilter(field string, value interface{}) interfaces.Filter {
	return &basicFilter{
		field:      field,
		expression: interfaces.FilterExpressionEqual,
		value:      value,
	}
}

// NewOrgFilter creates a filter for org = value
func NewOrgFilter(org string) interfaces.Filter {
	return NewEqualFilter("org", org)
}

// NewProjectIdFilter creates a filter for project identifier (org, project, domain)
func NewProjectIdFilter(projectId *common.ProjectIdentifier) interfaces.Filter {
	orgFilter := NewOrgFilter(projectId.GetOrganization())
	projectFilter := NewEqualFilter("project", projectId.GetName())
	domainFilter := NewEqualFilter("domain", projectId.GetDomain())

	return orgFilter.And(projectFilter).And(domainFilter)
}

// NewTaskNameFilter creates a filter for task name (org, project, domain, name)
func NewTaskNameFilter(taskName *task.TaskName) interfaces.Filter {
	orgFilter := NewOrgFilter(taskName.GetOrg())
	projectFilter := NewEqualFilter("project", taskName.GetProject())
	domainFilter := NewEqualFilter("domain", taskName.GetDomain())
	nameFilter := NewEqualFilter("name", taskName.GetName())

	return orgFilter.And(projectFilter).And(domainFilter).And(nameFilter)
}

// NewDeployedByFilter creates a filter for deployed_by = value
func NewDeployedByFilter(deployedBy string) interfaces.Filter {
	return NewEqualFilter("deployed_by", deployedBy)
}

// ConvertProtoFiltersToGormFilters converts proto filters to our Filter interfaces
func ConvertProtoFiltersToGormFilters(protoFilters []*common.Filter) (interfaces.Filter, error) {
	if len(protoFilters) == 0 {
		return nil, nil
	}

	filters := make([]interfaces.Filter, 0, len(protoFilters))

	for _, protoFilter := range protoFilters {
		// Convert filter function to expression
		var expression interfaces.FilterExpression
		switch protoFilter.Function {
		case common.Filter_EQUAL:
			expression = interfaces.FilterExpressionEqual
		case common.Filter_NOT_EQUAL:
			expression = interfaces.FilterExpressionNotEqual
		case common.Filter_GREATER_THAN:
			expression = interfaces.FilterExpressionGreaterThan
		case common.Filter_GREATER_THAN_OR_EQUAL:
			expression = interfaces.FilterExpressionGreaterThanOrEqual
		case common.Filter_LESS_THAN:
			expression = interfaces.FilterExpressionLessThan
		case common.Filter_LESS_THAN_OR_EQUAL:
			expression = interfaces.FilterExpressionLessThanOrEqual
		case common.Filter_CONTAINS:
			expression = interfaces.FilterExpressionContains
		case common.Filter_VALUE_IN:
			expression = interfaces.FilterExpressionValueIn
		case common.Filter_ENDS_WITH:
			expression = interfaces.FilterExpressionEndsWith
		case common.Filter_NOT_ENDS_WITH:
			expression = interfaces.FilterExpressionNotEndsWith
		case common.Filter_CONTAINS_CASE_INSENSITIVE:
			expression = interfaces.FilterExpressionContainsCaseInsensitive
		default:
			return nil, fmt.Errorf("unsupported filter function: %s", protoFilter.Function)
		}

		// Get filter value(s)
		var value interface{}
		if len(protoFilter.Values) == 1 {
			value = protoFilter.Values[0]
		} else if len(protoFilter.Values) > 1 {
			// For VALUE_IN, pass the array
			value = protoFilter.Values
		} else {
			return nil, fmt.Errorf("filter %s has no values", protoFilter.Field)
		}

		// Create basic filter
		filter := &basicFilter{
			field:      protoFilter.Field,
			expression: expression,
			value:      value,
		}

		filters = append(filters, filter)
	}

	// Combine all filters with AND
	if len(filters) == 0 {
		return nil, nil
	}

	combinedFilter := filters[0]
	for i := 1; i < len(filters); i++ {
		combinedFilter = combinedFilter.And(filters[i])
	}

	return combinedFilter, nil
}
