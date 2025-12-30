package impl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

func TestNewEqualFilter(t *testing.T) {
	filter := NewEqualFilter("org", "test-org")

	expr, err := filter.GormQueryExpression("")
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

	expr, err := filter.GormQueryExpression("")
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

	expr, err := filter.GormQueryExpression("")
	require.NoError(t, err)
	assert.Equal(t, "status IN ?", expr.Query)
}

func TestCompositeFilter_And(t *testing.T) {
	f1 := NewEqualFilter("org", "test-org")
	f2 := NewEqualFilter("project", "test-project")

	combined := f1.And(f2)
	expr, err := combined.GormQueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "AND")
	assert.Len(t, expr.Args, 2)
}

func TestCompositeFilter_Or(t *testing.T) {
	f1 := NewEqualFilter("status", "active")
	f2 := NewEqualFilter("status", "pending")

	combined := f1.Or(f2)
	expr, err := combined.GormQueryExpression("")
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
	expr, err := filter.GormQueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "org = ?")
	assert.Contains(t, expr.Query, "project = ?")
	assert.Contains(t, expr.Query, "domain = ?")
}

func TestNewTaskNameFilter(t *testing.T) {
	taskName := &task.TaskName{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "test-task",
	}

	filter := NewTaskNameFilter(taskName)
	expr, err := filter.GormQueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "name = ?")
}

func TestConvertProtoFiltersToGormFilters(t *testing.T) {
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

	filter, err := ConvertProtoFiltersToGormFilters(protoFilters)
	require.NoError(t, err)
	assert.NotNil(t, filter)

	expr, err := filter.GormQueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "org = ?")
	assert.Contains(t, expr.Query, "name LIKE ?")
}

func TestConvertProtoFilters_EmptyList(t *testing.T) {
	filter, err := ConvertProtoFiltersToGormFilters([]*common.Filter{})
	require.NoError(t, err)
	assert.Nil(t, filter)
}
