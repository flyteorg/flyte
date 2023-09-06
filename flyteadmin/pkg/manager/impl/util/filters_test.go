package util

import (
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
)

func TestParseRepeatedValues(t *testing.T) {
	assert.EqualValues(t, []string{"foo", "bar"}, parseRepeatedValues("foo;bar"))
}

func TestPrepareValues_WithTimestamp(t *testing.T) {
	timestampValue := "2018-07-27T00:30:31Z"
	values, err := prepareValues("CreatedAt", []string{timestampValue})
	if err != nil {
		t.Fatalf("failed to prepare value for CreatedAt: %+v with err %v", timestampValue, err)
	}
	expectedTime, err := time.Parse(time.RFC3339Nano, timestampValue)
	if err != nil {
		t.Fatalf("Native time library failed to parse test timestamp %s with err %v", timestampValue, err)
	}
	assert.EqualValues(t, expectedTime, values)

	badTimestampValue := "not a valid timestamp"
	_, err = prepareValues("CreatedAt", []string{badTimestampValue})
	assert.Error(t, err)
}

func TestPrepareValues_WithDuration(t *testing.T) {
	duration := "3600.5s"
	values, err := prepareValues("duration", []string{duration})
	assert.Nil(t, err)
	expectedDuration, err := time.ParseDuration(duration)
	if err != nil {
		t.Fatalf("Native time library failed to parse test timestamp %s with err %v", duration, err)
	}
	assert.EqualValues(t, expectedDuration, values)

	duration = "3600.5"
	values, err = prepareValues("duration", []string{duration})
	assert.Nil(t, err)
	assert.EqualValues(t, expectedDuration, values)

	badDurationValue := "not a valid duration"
	_, err = prepareValues("duration", []string{badDurationValue})
	assert.Error(t, err)
}

func TestPrepareValues_RepeatedValues(t *testing.T) {
	values, err := prepareValues("field", []string{"value"})
	assert.NoError(t, err)
	assert.Equal(t, "value", values)

	values, err = prepareValues("field", []string{"value a", "value b"})
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"value a", "value b"}, values)
}

func Test_ParseFilters_Success(t *testing.T) {
	filterExpression := "eq(project, flytesnacks)+ne(domain, development)+value_in(type, 4;5;6)"

	taskFilters, err := ParseFilters(filterExpression, common.Task)

	assert.NoError(t, err)
	require.Len(t, taskFilters, 3)

	actualFilterExpression, _ := taskFilters[0].GetGormQueryExpr()
	assert.Equal(t, "project = ?", actualFilterExpression.Query)
	assert.Equal(t, "flytesnacks", actualFilterExpression.Args)

	actualFilterExpression, _ = taskFilters[1].GetGormQueryExpr()
	assert.Equal(t, "domain <> ?", actualFilterExpression.Query)
	assert.Equal(t, "development", actualFilterExpression.Args)

	actualFilterExpression, _ = taskFilters[2].GetGormQueryExpr()
	assert.Equal(t, "type in (?)", actualFilterExpression.Query)
	assert.Equal(t, []interface{}{"4", "5", "6"}, actualFilterExpression.Args)
}

func Test_ParseFilters_InvalidFunction(t *testing.T) {
	filterExpression := "invalid_function(type,bar)"

	_, err := ParseFilters(filterExpression, common.Task)

	assert.EqualError(t, err, "unrecognized filter function: invalid_function")
}

func Test_ParseFilters_UnsupportedEntity(t *testing.T) {
	filterExpression := "eq(foo, 123)"

	_, err := ParseFilters(filterExpression, "wrong")

	assert.EqualError(t, err, "unsupported entity 'wrong'")
}

func Test_ParseFilters_InvalidJoinEntity(t *testing.T) {
	filterExpression := "eq(project.name, 123)"

	_, err := ParseFilters(filterExpression, common.Workflow)

	assert.EqualError(t, err, "'p' entity is not allowed in filters")
}

func Test_ParseFilters_InvalidFilter(t *testing.T) {
	filterExpression := "eq(foo, 123)"

	_, err := ParseFilters(filterExpression, common.Task)

	assert.EqualError(t, err, "'t.foo' is invalid filter")
}

func TestGetEqualityFilter(t *testing.T) {
	filter, err := GetSingleValueEqualityFilter(common.Task, "field", "value")
	assert.NoError(t, err)

	actualFilterExpression, _ := filter.GetGormQueryExpr()
	assert.Equal(t, "field = ?", actualFilterExpression.Query)
	assert.Equal(t, "value", actualFilterExpression.Args)
}

func Test_AddRequestFilters(t *testing.T) {
	filters, err := AddRequestFilters(
		"ne(cluster, TheWorst)+eq(workflow.name, workflow)", common.Execution, make([]common.InlineFilter, 0))

	assert.NoError(t, err)
	require.Len(t, filters, 2)

	expression, err := filters[0].GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, "cluster <> ?", expression.Query)
	assert.Equal(t, "TheWorst", expression.Args)

	expression, err = filters[1].GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, testutils.NameQueryPattern, expression.Query)
	assert.Equal(t, "workflow", expression.Args)
}

func TestGetDbFilters(t *testing.T) {
	actualFilters, err := GetDbFilters(FilterSpec{
		Project:        "project",
		Domain:         "domain",
		Name:           "name",
		RequestFilters: "ne(version, TheWorst)+eq(workflow.name, workflow)",
	}, common.LaunchPlan)
	assert.NoError(t, err)

	// Init expected values for filters.
	projectFilter, _ := GetSingleValueEqualityFilter(common.LaunchPlan, shared.Project, "project")
	domainFilter, _ := GetSingleValueEqualityFilter(common.LaunchPlan, shared.Domain, "domain")
	nameFilter, _ := GetSingleValueEqualityFilter(common.LaunchPlan, shared.Name, "name")
	versionFilter, _ := common.NewSingleValueFilter(common.LaunchPlan, common.NotEqual, shared.Version, "TheWorst")
	workflowNameFilter, _ := common.NewSingleValueFilter(common.Workflow, common.Equal, shared.Name, "workflow")
	expectedFilters := []common.InlineFilter{
		projectFilter,
		domainFilter,
		nameFilter,
		versionFilter,
		workflowNameFilter,
	}
	assert.EqualValues(t, expectedFilters, actualFilters)
}

func TestGetWorkflowExecutionIdentifierFilters(t *testing.T) {
	identifierFilters, err := GetWorkflowExecutionIdentifierFilters(
		context.Background(), core.WorkflowExecutionIdentifier{
			Project: "ex project",
			Domain:  "ex domain",
			Name:    "ex name",
		})
	assert.Nil(t, err)

	assert.Len(t, identifierFilters, 3)
	assert.Equal(t, common.Execution, identifierFilters[0].GetEntity())
	queryExpr, _ := identifierFilters[0].GetGormQueryExpr()
	assert.Equal(t, "ex project", queryExpr.Args)
	assert.Equal(t, "execution_project = ?", queryExpr.Query)

	assert.Equal(t, common.Execution, identifierFilters[1].GetEntity())
	queryExpr, _ = identifierFilters[1].GetGormQueryExpr()
	assert.Equal(t, "ex domain", queryExpr.Args)
	assert.Equal(t, "execution_domain = ?", queryExpr.Query)

	assert.Equal(t, common.Execution, identifierFilters[2].GetEntity())
	queryExpr, _ = identifierFilters[2].GetGormQueryExpr()
	assert.Equal(t, "ex name", queryExpr.Args)
	assert.Equal(t, "execution_name = ?", queryExpr.Query)
}

func TestGetNodeExecutionIdentifierFilters(t *testing.T) {
	identifierFilters, err := GetNodeExecutionIdentifierFilters(
		context.Background(), core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "ex project",
				Domain:  "ex domain",
				Name:    "ex name",
			},
			NodeId: "nodey",
		})
	assert.Nil(t, err)

	assert.Len(t, identifierFilters, 4)
	assert.Equal(t, common.Execution, identifierFilters[0].GetEntity())
	queryExpr, _ := identifierFilters[0].GetGormQueryExpr()
	assert.Equal(t, "ex project", queryExpr.Args)
	assert.Equal(t, "execution_project = ?", queryExpr.Query)

	assert.Equal(t, common.Execution, identifierFilters[1].GetEntity())
	queryExpr, _ = identifierFilters[1].GetGormQueryExpr()
	assert.Equal(t, "ex domain", queryExpr.Args)
	assert.Equal(t, "execution_domain = ?", queryExpr.Query)

	assert.Equal(t, common.Execution, identifierFilters[2].GetEntity())
	queryExpr, _ = identifierFilters[2].GetGormQueryExpr()
	assert.Equal(t, "ex name", queryExpr.Args)
	assert.Equal(t, "execution_name = ?", queryExpr.Query)

	assert.Equal(t, common.NodeExecution, identifierFilters[3].GetEntity())
	queryExpr, _ = identifierFilters[3].GetGormQueryExpr()
	assert.Equal(t, "nodey", queryExpr.Args)
	assert.Equal(t, "node_id = ?", queryExpr.Query)
}
