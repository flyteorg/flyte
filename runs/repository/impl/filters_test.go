package impl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestNewEqualFilter(t *testing.T) {
	filter := NewEqualFilter("org", "test-org")

	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Equal(t, "org = ?", expr.Query)
	assert.Equal(t, []interface{}{"test-org"}, expr.Args)
}

func TestBasicFilter_Contains(t *testing.T) {
	filter := &basicFilter{
		field:      "name",
		expression: interfaces.FilterExpressionContains,
		value:      "test",
	}

	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Equal(t, "name LIKE ?", expr.Query)
	assert.Equal(t, []interface{}{"%test%"}, expr.Args)
}

func TestBasicFilter_ValueIn(t *testing.T) {
	filter := &basicFilter{
		field:      "status",
		expression: interfaces.FilterExpressionValueIn,
		value:      []string{"active", "pending"},
	}

	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Equal(t, "status = ANY(?)", expr.Query)
}

func TestCompositeFilter_And(t *testing.T) {
	f1 := NewEqualFilter("org", "test-org")
	f2 := NewEqualFilter("project", "test-project")

	combined := f1.And(f2)
	expr, err := combined.QueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "AND")
	assert.Len(t, expr.Args, 2)
}

func TestCompositeFilter_Or(t *testing.T) {
	f1 := NewEqualFilter("status", "active")
	f2 := NewEqualFilter("status", "pending")

	combined := f1.Or(f2)
	expr, err := combined.QueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "OR")
}

func TestNewProjectIdFilter(t *testing.T) {
	projectId := &common.ProjectIdentifier{
		Organization: "test-org",
		Name:         "test-project",
		Domain:       "test-domain",
	}

	filter := NewProjectIdFilter(projectId)
	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "project = ?")
	assert.Contains(t, expr.Query, "domain = ?")
	assert.NotContains(t, expr.Query, "org = ?")
}

func TestNewTaskNameFilter(t *testing.T) {
	taskName := &task.TaskName{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "test-task",
	}

	filter := NewTaskNameFilter(taskName)
	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "name = ?")
}

func TestConvertProtoFilters(t *testing.T) {
	protoFilters := []*common.Filter{
		{
			Field:    "org",
			Function: common.Filter_EQUAL,
			Values:   []string{"test-org"},
		},
		{
			Field:    "name",
			Function: common.Filter_CONTAINS,
			Values:   []string{"test"},
		},
	}

	allowedColumns := sets.New("org", "name")
	filter, err := ConvertProtoFilters(protoFilters, allowedColumns)
	require.NoError(t, err)
	assert.NotNil(t, filter)

	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "org = ?")
	assert.Contains(t, expr.Query, "name LIKE ?")
}

func TestConvertProtoFilters_DisallowedColumn(t *testing.T) {
	protoFilters := []*common.Filter{
		{
			Field:    "malicious_field",
			Function: common.Filter_EQUAL,
			Values:   []string{"value"},
		},
	}

	allowedColumns := sets.New("org", "name")
	_, err := ConvertProtoFilters(protoFilters, allowedColumns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter field")
}

func TestConvertProtoFilters_ActionTypeColumn(t *testing.T) {
	protoFilters := []*common.Filter{
		{
			Field:    "action_type",
			Function: common.Filter_EQUAL,
			Values:   []string{"3"},
		},
	}

	filter, err := ConvertProtoFilters(protoFilters, models.ActionColumnsSet)
	require.NoError(t, err)

	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Equal(t, "action_type = ?", expr.Query)
	assert.Equal(t, []interface{}{"3"}, expr.Args)
}

func TestConvertProtoFilters_EmptyList(t *testing.T) {
	filter, err := ConvertProtoFilters([]*common.Filter{}, sets.New[string]())
	require.NoError(t, err)
	assert.Nil(t, filter)
}

func TestParseStringFilters_StateNumericString(t *testing.T) {
	filter, err := ParseStringFilters("eq(state,1)", models.ProjectColumns)
	require.NoError(t, err)

	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Equal(t, "state = ?", expr.Query)
	assert.Equal(t, []interface{}{"1"}, expr.Args)
}

func TestParseStringFilters_ValueInState(t *testing.T) {
	filter, err := ParseStringFilters("value_in(state,0;1;2)", models.ProjectColumns)
	require.NoError(t, err)

	expr, err := filter.QueryExpression("")
	require.NoError(t, err)
	assert.Equal(t, "state = ANY(?)", expr.Query)
	require.Len(t, expr.Args, 1)
	// pq.Array wraps the slice, so check the underlying values via formatting
	assert.Contains(t, fmt.Sprintf("%v", expr.Args[0]), "0")
	assert.Contains(t, fmt.Sprintf("%v", expr.Args[0]), "1")
	assert.Contains(t, fmt.Sprintf("%v", expr.Args[0]), "2")
}

// TestNewSearchFilter verifies the run search predicate matches the term against the run
// name OR the task name, using the SQL operator selected by the proto function.
func TestNewSearchFilter(t *testing.T) {
	cases := []struct {
		name     string
		fn       common.Filter_Function
		query    string
		wantArgs []interface{}
	}{
		{
			name:     "contains case insensitive",
			fn:       common.Filter_CONTAINS_CASE_INSENSITIVE,
			query:    "(LOWER(run_name) LIKE LOWER(?)) OR (LOWER(task_name) LIKE LOWER(?))",
			wantArgs: []interface{}{"%abc%", "%abc%"},
		},
		{
			name:     "contains case sensitive",
			fn:       common.Filter_CONTAINS,
			query:    "(run_name LIKE ?) OR (task_name LIKE ?)",
			wantArgs: []interface{}{"%abc%", "%abc%"},
		},
		{
			name:     "equal",
			fn:       common.Filter_EQUAL,
			query:    "(run_name = ?) OR (task_name = ?)",
			wantArgs: []interface{}{"abc", "abc"},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewSearchFilter(tt.fn, []string{"abc"})
			require.NoError(t, err)

			expr, err := filter.QueryExpression("")
			require.NoError(t, err)
			assert.Equal(t, tt.query, expr.Query)
			assert.Equal(t, tt.wantArgs, expr.Args)
		})
	}
}

// TestNewSearchFilter_WithTablePrefix verifies both columns are table-qualified when the
// predicate is rendered for a query that aliases the actions table.
func TestNewSearchFilter_WithTablePrefix(t *testing.T) {
	filter, err := NewSearchFilter(common.Filter_EQUAL, []string{"abc"})
	require.NoError(t, err)

	expr, err := filter.QueryExpression("actions")
	require.NoError(t, err)
	assert.Equal(t, "(actions.run_name = ?) OR (actions.task_name = ?)", expr.Query)
}

// TestNewSearchFilter_Errors verifies the search field rejects operators that make no sense
// for a free-text search and malformed value lists, so a bad request fails loudly instead
// of silently matching everything.
func TestNewSearchFilter_Errors(t *testing.T) {
	cases := []struct {
		name   string
		fn     common.Filter_Function
		values []string
	}{
		{name: "unsupported function", fn: common.Filter_GREATER_THAN, values: []string{"abc"}},
		{name: "value in is not a search", fn: common.Filter_VALUE_IN, values: []string{"a", "b"}},
		{name: "no values", fn: common.Filter_CONTAINS_CASE_INSENSITIVE, values: nil},
		{name: "multiple values", fn: common.Filter_CONTAINS_CASE_INSENSITIVE, values: []string{"a", "b"}},
		{name: "empty term", fn: common.Filter_CONTAINS_CASE_INSENSITIVE, values: []string{""}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewSearchFilter(tt.fn, tt.values)
			require.Error(t, err)
			assert.Nil(t, filter)
		})
	}
}

// TestSplitSearchFilters verifies the "search" field is pulled out of the filter list and
// resolved to the run-name-or-task-name predicate, while plain column filters pass through
// for ConvertProtoFilters.
func TestSplitSearchFilters(t *testing.T) {
	filters := []*common.Filter{
		{Function: common.Filter_EQUAL, Field: "phase", Values: []string{"2"}},
		{Function: common.Filter_CONTAINS_CASE_INSENSITIVE, Field: SearchField, Values: []string{"abc"}},
	}

	search, remaining, err := SplitSearchFilters(filters)
	require.NoError(t, err)

	require.Len(t, remaining, 1)
	assert.Equal(t, "phase", remaining[0].GetField())

	require.NotNil(t, search)
	expr, err := search.QueryExpression("")
	require.NoError(t, err)
	assert.Equal(t, "(LOWER(run_name) LIKE LOWER(?)) OR (LOWER(task_name) LIKE LOWER(?))", expr.Query)
	assert.Equal(t, []interface{}{"%abc%", "%abc%"}, expr.Args)
}

// TestSplitSearchFilters_NoSearch verifies a nil predicate is returned when there is no
// search filter, so the caller skips the extra AND and leaves the request untouched.
func TestSplitSearchFilters_NoSearch(t *testing.T) {
	filters := []*common.Filter{
		{Function: common.Filter_EQUAL, Field: "phase", Values: []string{"2"}},
	}

	search, remaining, err := SplitSearchFilters(filters)
	require.NoError(t, err)
	assert.Nil(t, search)
	assert.Len(t, remaining, 1)
}

// TestSplitSearchFilters_Error verifies a malformed search filter fails the whole split, so
// the handler returns InvalidArgument instead of dropping the predicate.
func TestSplitSearchFilters_Error(t *testing.T) {
	filters := []*common.Filter{
		{Function: common.Filter_GREATER_THAN, Field: SearchField, Values: []string{"abc"}},
	}

	search, remaining, err := SplitSearchFilters(filters)
	require.Error(t, err)
	assert.Nil(t, search)
	assert.Nil(t, remaining)
}
