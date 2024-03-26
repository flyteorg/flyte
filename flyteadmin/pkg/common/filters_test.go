package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEntity(t *testing.T) {
	filter, err := NewSingleValueFilter(Workflow, Equal, "domain", "production")
	assert.NoError(t, err)
	assert.Equal(t, Workflow, filter.GetEntity())
}

func TestNewSingleValueFilter(t *testing.T) {
	_, err := NewSingleValueFilter(Workflow, Equal, "domain", "production")
	assert.NoError(t, err)

	_, err = NewSingleValueFilter(Workflow, ValueIn, "project", "SuperAwesomeProject")
	assert.EqualError(t, err, "invalid single value filter expression: value in")
}

func TestCustomizedSingleValueFilter(t *testing.T) {
	filter, err := NewSingleValueFilter(Execution, Equal, "domain", "production")
	assert.NoError(t, err)
	expression, err := filter.GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, expression.Query, "execution_domain = ?")
}

func TestNewSingleIntValueFilter(t *testing.T) {
	filter, err := NewSingleValueFilter(Workflow, Equal, "num", float64(1.2))
	assert.NoError(t, err)

	expression, err := filter.GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, expression.Query, "num = ?")
	assert.Equal(t, expression.Args, float64(1.2))
}

func TestNewSingleBoolValueFilter(t *testing.T) {
	filter, err := NewSingleValueFilter(Workflow, Equal, "raining", true)
	assert.NoError(t, err)

	expression, err := filter.GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, expression.Query, "raining = ?")
	assert.Equal(t, expression.Args, true)
}

func TestNewSingleValueCustomizedFieldFilter(t *testing.T) {
	filter, err := NewSingleValueFilter(Execution, Equal, "project", "a project")
	assert.NoError(t, err)

	expression, err := filter.GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, "execution_project = ?", expression.Query)

	expression, err = filter.GetGormJoinTableQueryExpr("node_executions")
	assert.NoError(t, err)
	assert.Equal(t, "node_executions.execution_project = ?", expression.Query)
}

func TestNewSingleValueCustomizedEntityFilter(t *testing.T) {
	filter, err := NewSingleValueFilter(NamedEntity, Equal, "state", 1)
	assert.NoError(t, err)

	assert.Equal(t, NamedEntityMetadata, filter.GetEntity())

	filter, err = NewSingleValueFilter(NamedEntity, Equal, "description", "test value")
	assert.NoError(t, err)

	assert.Equal(t, NamedEntityMetadata, filter.GetEntity())
}

func TestNewRepeatedValueFilter(t *testing.T) {
	vals := []string{"SuperAwesomeProject", "AnotherAwesomeProject"}
	filter, err := NewRepeatedValueFilter(Workflow, ValueIn, "project", vals)
	assert.NoError(t, err)

	expression, err := filter.GetGormJoinTableQueryExpr("projects")
	assert.NoError(t, err)
	assert.Equal(t, "projects.project in (?)", expression.Query)
	assert.Equal(t, vals, expression.Args)

	_, err = NewRepeatedValueFilter(Workflow, Equal, "domain", []string{"production", "qa"})
	assert.EqualError(t, err, "invalid repeated value filter expression: equal")
}

func TestGetGormJoinTableQueryExpr(t *testing.T) {
	filter, err := NewSingleValueFilter(Task, Equal, "domain", "production")
	assert.NoError(t, err)

	gormQueryExpr, err := filter.GetGormJoinTableQueryExpr("workflows")
	assert.NoError(t, err)
	assert.Equal(t, "workflows.domain = ?", gormQueryExpr.Query)
}

var expectedQueriesForFilters = map[FilterExpression]string{
	Contains:           "field LIKE ?",
	GreaterThan:        "field > ?",
	GreaterThanOrEqual: "field >= ?",
	LessThan:           "field < ?",
	LessThanOrEqual:    "field <= ?",
	Equal:              "field = ?",
	NotEqual:           "field <> ?",
}

var expectedArgsForFilters = map[FilterExpression]string{
	Contains:           "%value%",
	GreaterThan:        "value",
	GreaterThanOrEqual: "value",
	LessThan:           "value",
	LessThanOrEqual:    "value",
	Equal:              "value",
	NotEqual:           "value",
}

func TestQueryExpressions(t *testing.T) {
	for expression, expectedQuery := range expectedQueriesForFilters {
		filter, err := NewSingleValueFilter(Workflow, expression, "field", "value")
		assert.NoError(t, err)

		gormQueryExpr, err := filter.GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, expectedQuery, gormQueryExpr.Query)

		expectedArg, ok := expectedArgsForFilters[expression]
		assert.True(t, ok, "Missing expected argument for expression %s", expression)
		assert.Equal(t, expectedArg, gormQueryExpr.Args)
	}

	// Also test the one repeated value filter
	filter, err := NewRepeatedValueFilter(Workflow, ValueIn, "field", []string{"value"})
	assert.NoError(t, err)

	gormQueryExpr, err := filter.GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, "field in (?)", gormQueryExpr.Query)
	assert.EqualValues(t, []string{"value"}, gormQueryExpr.Args)
}

func TestMapFilter(t *testing.T) {
	mapFilterValue := map[string]interface{}{
		"foo": "bar",
		"baz": 1,
	}
	assert.EqualValues(t, mapFilterValue, NewMapFilter(mapFilterValue).GetFilter())
}

func TestWithDefaultValueFilter(t *testing.T) {
	filter, err := NewSingleValueFilter(NamedEntityMetadata, Equal, "state", 1)
	assert.NoError(t, err)

	filterWithDefaultValue, err := NewWithDefaultValueFilter(0, filter)
	assert.NoError(t, err)

	queryExpression, err := filterWithDefaultValue.GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, "COALESCE(state, 0) = ?", queryExpression.Query)
	assert.Equal(t, 1, queryExpression.Args)

	queryExpression, err = filterWithDefaultValue.GetGormJoinTableQueryExpr(
		"named_entity_metadata")
	assert.NoError(t, err)
	assert.Equal(t, "COALESCE(named_entity_metadata.state, 0) = ?", queryExpression.Query)
	assert.Equal(t, 1, queryExpression.Args)
}
