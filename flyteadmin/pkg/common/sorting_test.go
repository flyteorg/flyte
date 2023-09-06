package common

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
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

	assert.EqualError(t, err, `invalid sort order specified: key:"name" direction:2 `)
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
