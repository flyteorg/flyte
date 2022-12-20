package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func TestTransformFilter(t *testing.T) {
	tests := []TestCase{
		{
			Input:  "project.name=flytesnacks,execution.duration<200,execution.duration<=200,execution.duration>=200,name contains flyte,name!=flyte",
			Output: "eq(project.name,flytesnacks)+lt(execution.duration,200)+lte(execution.duration,200)+gte(execution.duration,200)+contains(name,flyte)+ne(name,flyte)",
		},
		{
			Input:  "execution.phase in (FAILED;SUCCEEDED),execution.name=y8n2wtuspj,execution.duration>200",
			Output: "value_in(execution.phase,FAILED;SUCCEEDED)+eq(execution.name,y8n2wtuspj)+gt(execution.duration,200)",
		},
		{
			Input:  `k=\\,,k2=v2`,
			Output: "eq(k,\\)+eq(k2,v2)",
		},
	}
	for _, test := range tests {
		filters := SplitTerms(test.Input)

		result, err := Transform(filters)
		assert.Nil(t, err)
		assert.Equal(t, test.Output, result)
	}
}

func TestTransformFilterError(t *testing.T) {
	tests := []TestCase{
		{
			Input:  `\=\,\`,
			Output: "",
		},
		{
			Input:  `foo=bar,baz=blah,complex=\=value\\\,\\`,
			Output: "",
		},
	}
	for _, test := range tests {
		filters := SplitTerms(test.Input)
		result, err := Transform(filters)
		assert.NotNil(t, err)
		assert.Equal(t, "", result)
	}
}

func TestParseFailed(t *testing.T) {
	tests := []TestCase{
		{
			Input:  ``,
			Output: "",
		},
	}
	for _, test := range tests {
		lhs, op, rhs, ok := parse(test.Input)
		result := transform(lhs, op, rhs)
		assert.Equal(t, "", result)
		assert.Equal(t, false, ok)
		assert.Equal(t, "", lhs)
		assert.Equal(t, "", rhs)
		assert.Equal(t, "", op)
	}
}
