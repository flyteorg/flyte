package impl

import (
	"fmt"

	"github.com/lib/pq"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// NewIsRootActionFilter creates a filter for root actions (runs) only.
func NewIsRootActionFilter() interfaces.Filter {
	return &nullFilter{field: "parent_action_name", isNull: true}
}

// NewRunActionsFilter creates a filter for all actions belonging to a specific run.
func NewRunActionsFilter(runID *common.RunIdentifier) interfaces.Filter {
	return NewEqualFilter("project", runID.GetProject()).
		And(NewEqualFilter("domain", runID.GetDomain())).
		And(NewEqualFilter("run_name", runID.GetName()))
}

// basicFilter implements the Filter interface for simple field comparisons
type basicFilter struct {
	field      string
	expression interfaces.FilterExpression
	value      interface{}
}

func (f *basicFilter) QueryExpression(table string) (interfaces.QueryExpr, error) {
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
		query = fmt.Sprintf("%s = ANY(?)", column)
		f.value = pq.Array(f.value)
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
		return interfaces.QueryExpr{}, fmt.Errorf("unsupported filter expression: %d", f.expression)
	}

	return interfaces.QueryExpr{
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

// nullFilter implements the Filter interface for IS NULL / IS NOT NULL checks
type nullFilter struct {
	field  string
	isNull bool
}

func (f *nullFilter) QueryExpression(table string) (interfaces.QueryExpr, error) {
	column := f.field
	if table != "" {
		column = table + "." + f.field
	}
	op := "IS NULL"
	if !f.isNull {
		op = "IS NOT NULL"
	}
	return interfaces.QueryExpr{
		Query: fmt.Sprintf("%s %s", column, op),
	}, nil
}

func (f *nullFilter) And(filter interfaces.Filter) interfaces.Filter {
	return &compositeFilter{left: f, right: filter, operator: "AND"}
}

func (f *nullFilter) Or(filter interfaces.Filter) interfaces.Filter {
	return &compositeFilter{left: f, right: filter, operator: "OR"}
}

// compositeFilter implements the Filter interface for AND/OR operations
type compositeFilter struct {
	left     interfaces.Filter
	right    interfaces.Filter
	operator string // "AND" or "OR"
}

func (f *compositeFilter) QueryExpression(table string) (interfaces.QueryExpr, error) {
	leftExpr, err := f.left.QueryExpression(table)
	if err != nil {
		return interfaces.QueryExpr{}, err
	}

	rightExpr, err := f.right.QueryExpression(table)
	if err != nil {
		return interfaces.QueryExpr{}, err
	}

	query := fmt.Sprintf("(%s) %s (%s)", leftExpr.Query, f.operator, rightExpr.Query)
	args := append(leftExpr.Args, rightExpr.Args...)

	return interfaces.QueryExpr{
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

// NewNotEqualFilter creates a filter for field != value.
func NewNotEqualFilter(field string, value interface{}) interfaces.Filter {
	return &basicFilter{
		field:      field,
		expression: interfaces.FilterExpressionNotEqual,
		value:      value,
	}
}

// NewProjectIdFilter creates a filter for project identifier (project, domain)
func NewProjectIdFilter(projectId *common.ProjectIdentifier) interfaces.Filter {
	projectFilter := NewEqualFilter("project", projectId.GetName())
	domainFilter := NewEqualFilter("domain", projectId.GetDomain())

	return projectFilter.And(domainFilter)
}

// NewTaskNameFilter creates a filter for task name on the tasks table (project, domain, name).
func NewTaskNameFilter(taskName *task.TaskName) interfaces.Filter {
	return NewEqualFilter("project", taskName.GetProject()).
		And(NewEqualFilter("domain", taskName.GetDomain())).
		And(NewEqualFilter("name", taskName.GetName()))
}

// NewRunTaskNameFilter creates a filter matching runs by task name columns on the actions table.
func NewRunTaskNameFilter(taskName *task.TaskName) interfaces.Filter {
	return NewEqualFilter("task_project", taskName.GetProject()).
		And(NewEqualFilter("task_domain", taskName.GetDomain())).
		And(NewEqualFilter("task_name", taskName.GetName()))
}

// NewRunTaskIdFilter creates a filter matching runs by full task identifier on the actions table.
func NewRunTaskIdFilter(taskId *task.TaskIdentifier) interfaces.Filter {
	return NewEqualFilter("task_project", taskId.GetProject()).
		And(NewEqualFilter("task_domain", taskId.GetDomain())).
		And(NewEqualFilter("task_name", taskId.GetName())).
		And(NewEqualFilter("task_version", taskId.GetVersion()))
}

// NewTriggerNameFilter creates a filter matching runs by trigger_name on the actions table.
func NewTriggerNameFilter(triggerName *common.TriggerName) interfaces.Filter {
	return NewEqualFilter("project", triggerName.GetProject()).
		And(NewEqualFilter("domain", triggerName.GetDomain())).
		And(NewEqualFilter("trigger_task_name", triggerName.GetTaskName())).
		And(NewEqualFilter("trigger_name", triggerName.GetName()))
}

// NewDeployedByFilter creates a filter for deployed_by = value
func NewDeployedByFilter(deployedBy string) interfaces.Filter {
	return NewEqualFilter("deployed_by", deployedBy)
}

// ConvertProtoFilters converts proto filters to our Filter interfaces.
// allowedColumns is checked to prevent SQL injection via user-supplied field names.
func ConvertProtoFilters(protoFilters []*common.Filter, allowedColumns sets.Set[string]) (interfaces.Filter, error) {
	if len(protoFilters) == 0 {
		return nil, nil
	}

	filters := make([]interfaces.Filter, 0, len(protoFilters))

	for _, protoFilter := range protoFilters {
		if !allowedColumns.Has(protoFilter.Field) {
			return nil, fmt.Errorf("invalid filter field: %s", protoFilter.Field)
		}
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
