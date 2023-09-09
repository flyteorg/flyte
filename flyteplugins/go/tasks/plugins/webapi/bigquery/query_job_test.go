package bigquery

import (
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/bigquery/v2"
)

func TestGetQueryParameter(t *testing.T) {
	t.Run("get integer parameter", func(t *testing.T) {
		literal, _ := coreutils.MakePrimitiveLiteral(42)

		tpe, value, err := getQueryParameter(literal)

		assert.NoError(t, err)
		assert.Equal(t, bigquery.QueryParameterType{Type: "INT64"}, *tpe)
		assert.Equal(t, bigquery.QueryParameterValue{Value: "42"}, *value)
	})

	t.Run("get string parameter", func(t *testing.T) {
		literal, _ := coreutils.MakePrimitiveLiteral("abc")

		tpe, value, err := getQueryParameter(literal)

		assert.NoError(t, err)
		assert.Equal(t, bigquery.QueryParameterType{Type: "STRING"}, *tpe)
		assert.Equal(t, bigquery.QueryParameterValue{Value: "abc"}, *value)
	})

	t.Run("get float parameter", func(t *testing.T) {
		literal, _ := coreutils.MakePrimitiveLiteral(42.5)

		tpe, value, err := getQueryParameter(literal)

		assert.NoError(t, err)
		assert.Equal(t, bigquery.QueryParameterType{Type: "FLOAT64"}, *tpe)
		assert.Equal(t, bigquery.QueryParameterValue{Value: "42.5"}, *value)
	})

	t.Run("get true parameter", func(t *testing.T) {
		literal, _ := coreutils.MakePrimitiveLiteral(true)

		tpe, value, err := getQueryParameter(literal)

		assert.NoError(t, err)
		assert.Equal(t, bigquery.QueryParameterType{Type: "BOOL"}, *tpe)
		assert.Equal(t, bigquery.QueryParameterValue{Value: "TRUE"}, *value)
	})

	t.Run("get false parameter", func(t *testing.T) {
		literal, _ := coreutils.MakePrimitiveLiteral(false)

		tpe, value, err := getQueryParameter(literal)

		assert.NoError(t, err)
		assert.Equal(t, bigquery.QueryParameterType{Type: "BOOL"}, *tpe)
		assert.Equal(t, bigquery.QueryParameterValue{Value: "FALSE"}, *value)
	})
}

func TestGetJobConfigurationQuery(t *testing.T) {
	t.Run("get job configuration query", func(t *testing.T) {
		config := QueryJobConfig{}
		inputs, _ := coreutils.MakeLiteralMap(map[string]interface{}{
			"integer": 42,
		})

		jobConfigurationQuery, err := getJobConfigurationQuery(&config, inputs)
		useLegacySQL := false

		assert.NoError(t, err)
		assert.Equal(t, "NAMED", jobConfigurationQuery.ParameterMode)
		assert.Equal(t, &useLegacySQL, jobConfigurationQuery.UseLegacySql)
		assert.Equal(t, 1, len(jobConfigurationQuery.QueryParameters))
		assert.Equal(t, bigquery.QueryParameter{
			Name:           "integer",
			ParameterType:  &bigquery.QueryParameterType{Type: "INT64"},
			ParameterValue: &bigquery.QueryParameterValue{Value: "42"},
		}, *jobConfigurationQuery.QueryParameters[0])
	})
}
