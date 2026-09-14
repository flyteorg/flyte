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

// NewHasPausedActionFilter matches root actions whose run contains at least one
// action in the PAUSED phase (e.g. a human-in-the-loop gate node awaiting input).
func NewHasPausedActionFilter() interfaces.Filter {
	return NewRawFilter(
		`EXISTS (SELECT 1 FROM actions a2 `+
			`WHERE a2.project = actions.project `+
			`AND a2.domain = actions.domain `+
			`AND a2.run_name = actions.run_name `+
			`AND a2.phase = ?)`,
		int32(common.ActionPhase_ACTION_PHASE_PAUSED),
	)
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

// rawFilter implements the Filter interface for hand-written SQL predicates that
// the field/expression filters cannot express (e.g. correlated EXISTS subqueries).
// The query must qualify its own columns; the table argument is ignored.
type rawFilter struct {
	query string
	args  []interface{}
}

// NewRawFilter creates a filter from a raw SQL predicate and its bind arguments.
func NewRawFilter(query string, args ...interface{}) interfaces.Filter {
	return &rawFilter{query: query, args: args}
}

func (f *rawFilter) QueryExpression(table string) (interfaces.QueryExpr, error) {
	return interfaces.QueryExpr{Query: f.query, Args: f.args}, nil
}

func (f *rawFilter) And(filter interfaces.Filter) interfaces.Filter {
	return &compositeFilter{left: f, right: filter, operator: "AND"}
}

func (f *rawFilter) Or(filter interfaces.Filter) interfaces.Filter {
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

// SearchField is the synthetic filter field behind the console's run search box. A filter
// on it matches runs whose run name OR task name satisfies the predicate, which the flat,
// AND-ed common.Filter list cannot express on its own. It is not a column: the run service
// resolves it with SplitSearchFilters before the generic allow-list conversion
// (ConvertProtoFilters) sees the request. The Union cloud run service accepts the same
// field, so one console request works against both backends.
const SearchField = "search"

// searchFunctions maps the proto filter functions accepted on SearchField to their SQL
// expression. Contains (case-sensitive or not) is what the search box sends; EQUAL lets a
// caller holding a full run name match it exactly instead of with a wildcard scan.
var searchFunctions = map[common.Filter_Function]interfaces.FilterExpression{
	common.Filter_EQUAL:                     interfaces.FilterExpressionEqual,
	common.Filter_CONTAINS:                  interfaces.FilterExpressionContains,
	common.Filter_CONTAINS_CASE_INSENSITIVE: interfaces.FilterExpressionContainsCaseInsensitive,
}

// NewSearchFilter builds the `run_name <op> ? OR task_name <op> ?` predicate for a
// SearchField filter. The OR is parenthesized by compositeFilter, so it composes safely
// with the AND-ed scope filters callers add around it.
func NewSearchFilter(fn common.Filter_Function, values []string) (interfaces.Filter, error) {
	expression, ok := searchFunctions[fn]
	if !ok {
		return nil, fmt.Errorf("unsupported filter function %s for field %q; expected EQUAL, CONTAINS or CONTAINS_CASE_INSENSITIVE", fn, SearchField)
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("filter on field %q expects a single value, got %d", SearchField, len(values))
	}
	term := values[0]
	if term == "" {
		return nil, fmt.Errorf("filter on field %q requires a non-empty value", SearchField)
	}
	runName := &basicFilter{field: "run_name", expression: expression, value: term}
	taskName := &basicFilter{field: "task_name", expression: expression, value: term}
	return runName.Or(taskName), nil
}

// SplitSearchFilters partitions proto filters into a single combined SearchField predicate
// (nil when there is none; several are AND-ed) and the remaining plain column filters.
// Callers pass the remaining filters to ConvertProtoFilters, which validates fields against
// the column allow-list and would otherwise reject the synthetic field, and AND the returned
// predicate in separately.
func SplitSearchFilters(filters []*common.Filter) (interfaces.Filter, []*common.Filter, error) {
	var search interfaces.Filter
	remaining := make([]*common.Filter, 0, len(filters))
	for _, f := range filters {
		if f.GetField() != SearchField {
			remaining = append(remaining, f)
			continue
		}
		sf, err := NewSearchFilter(f.GetFunction(), f.GetValues())
		if err != nil {
			return nil, nil, err
		}
		if search == nil {
			search = sf
		} else {
			search = search.And(sf)
		}
	}
	return search, remaining, nil
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
