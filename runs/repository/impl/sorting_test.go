package impl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

func TestNewSortParameter(t *testing.T) {
	sp := NewSortParameter("created_at", interfaces.SortOrderDescending)
	assert.Equal(t, "created_at DESC", sp.GetGormOrderExpr())
}

func TestSortParameter_Ascending(t *testing.T) {
	sp := NewSortParameter("name", interfaces.SortOrderAscending)
	assert.Equal(t, "name ASC", sp.GetGormOrderExpr())
}

func TestGetSortByFieldsV2_SingleField(t *testing.T) {
	request := &common.ListRequest{
		SortByFields: []*common.Sort{
			{
				Key:       "created_at",
				Direction: common.Sort_DESCENDING,
			},
		},
	}

	allowedColumns := sets.New("created_at", "name", "version")
	sortParams, err := GetSortByFieldsV2(request, allowedColumns)
	require.NoError(t, err)
	require.Len(t, sortParams, 1)
	assert.Equal(t, "created_at DESC", sortParams[0].GetGormOrderExpr())
}

func TestGetSortByFieldsV2_MultipleFields(t *testing.T) {
	request := &common.ListRequest{
		SortByFields: []*common.Sort{
			{
				Key:       "name",
				Direction: common.Sort_ASCENDING,
			},
			{
				Key:       "version",
				Direction: common.Sort_DESCENDING,
			},
		},
	}

	allowedColumns := sets.New("name", "version")
	sortParams, err := GetSortByFieldsV2(request, allowedColumns)
	require.NoError(t, err)
	require.Len(t, sortParams, 2)
	assert.Equal(t, "name ASC", sortParams[0].GetGormOrderExpr())
	assert.Equal(t, "version DESC", sortParams[1].GetGormOrderExpr())
}

func TestGetSortByFieldsV2_InvalidField(t *testing.T) {
	request := &common.ListRequest{
		SortByFields: []*common.Sort{
			{
				Key:       "invalid_field",
				Direction: common.Sort_ASCENDING,
			},
		},
	}

	allowedColumns := sets.New("name", "version")
	_, err := GetSortByFieldsV2(request, allowedColumns)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort field")
}

func TestGetSortByFieldsV2_NilRequest(t *testing.T) {
	allowedColumns := sets.New("name")
	sortParams, err := GetSortByFieldsV2(nil, allowedColumns)
	require.NoError(t, err)
	assert.Nil(t, sortParams)
}

func TestGetSortByFieldsV2_EmptyFields(t *testing.T) {
	request := &common.ListRequest{
		SortByFields: []*common.Sort{},
	}

	allowedColumns := sets.New("name")
	sortParams, err := GetSortByFieldsV2(request, allowedColumns)
	require.NoError(t, err)
	assert.Nil(t, sortParams)
}
