package common

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestSortParameter_Ascending(t *testing.T) {
	sortParameter, err := NewSortParameter(admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "name",
	})
	assert.Nil(t, err)
	assert.Equal(t, "name asc", sortParameter.GetGormOrderExpr())
}

func TestSortParameter_Descending(t *testing.T) {
	sortParameter, err := NewSortParameter(admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "project",
	})
	assert.Nil(t, err)
	assert.Equal(t, "project desc", sortParameter.GetGormOrderExpr())
}
