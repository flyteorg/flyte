package common

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestSortParameter_Nil(t *testing.T) {
	sortParameter, err := NewSortParameter(nil, nil)

	assert.NoError(t, err)
	assert.Nil(t, sortParameter)
}

func TestSortParameter_InvalidSortKey(t *testing.T) {
	_, err := NewSortParameter(&admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "wrong",
	}, sets.NewString("name"))

	assert.EqualError(t, err, "invalid sort key 'wrong'")
}

func TestSortParameter_InvalidSortDirection(t *testing.T) {
	_, err := NewSortParameter(&admin.Sort{
		Direction: 2,
		Key:       "name",
	}, sets.NewString("name"))

	regex := regexp.MustCompile(`\s+`)
	expected := `invalid sort order specified: key:"name" direction:2`
	assert.Equal(t, regex.ReplaceAllString(err.Error(), " "), regex.ReplaceAllString(expected, " "))
}

func TestSortParameter_Ascending(t *testing.T) {
	sortParameter, err := NewSortParameter(&admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "name",
	}, sets.NewString("name"))

	assert.NoError(t, err)
	assert.Equal(t, "name asc", sortParameter.GetGormOrderExpr())
}

func TestSortParameter_Descending(t *testing.T) {
	sortParameter, err := NewSortParameter(&admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "project",
	}, sets.NewString("project"))

	assert.NoError(t, err)
	assert.Equal(t, "project desc", sortParameter.GetGormOrderExpr())
}
