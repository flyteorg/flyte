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
