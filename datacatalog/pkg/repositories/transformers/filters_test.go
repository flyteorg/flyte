package transformers

import (
	"context"
	"testing"

	"github.com/flyteorg/datacatalog/pkg/common"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/stretchr/testify/assert"
)

func assertJoinExpression(t *testing.T, joinCondition models.ModelJoinCondition, sourceTableName string, joiningTableName string, joiningTableAlias string, expectedJoinStatement string) {
	expr, err := joinCondition.GetJoinOnDBQueryExpression(sourceTableName, joiningTableName, joiningTableAlias)
	assert.NoError(t, err)
	assert.Equal(t, expr, expectedJoinStatement)
}

func assertFilterExpression(t *testing.T, filter models.ModelValueFilter, tableName string, expectedStatement string, expectedArgs interface{}) {
	expr, err := filter.GetDBQueryExpression(tableName)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatement, expr.Query)
	assert.EqualValues(t, expectedArgs, expr.Args)
}

func TestListInputWithPartitionsAndTags(t *testing.T) {
	filter := &datacatalog.FilterExpression{
		Filters: []*datacatalog.SinglePropertyFilter{
			{
				PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
					PartitionFilter: &datacatalog.PartitionPropertyFilter{
						Property: &datacatalog.PartitionPropertyFilter_KeyVal{
							KeyVal: &datacatalog.KeyValuePair{Key: "key1", Value: "val1"},
						},
					},
				},
			},
			{
				PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
					PartitionFilter: &datacatalog.PartitionPropertyFilter{
						Property: &datacatalog.PartitionPropertyFilter_KeyVal{
							KeyVal: &datacatalog.KeyValuePair{Key: "key2", Value: "val2"},
						},
					},
				},
			},
			{
				PropertyFilter: &datacatalog.SinglePropertyFilter_TagFilter{
					TagFilter: &datacatalog.TagPropertyFilter{
						Property: &datacatalog.TagPropertyFilter_TagName{
							TagName: "special",
						},
					},
				},
			},
		},
	}
	listInput, err := FilterToListInput(context.Background(), common.Artifact, filter)
	assert.NoError(t, err)

	// Should have 3 filters: 2 for partitions, 1 for tag
	assert.Len(t, listInput.ModelFilters, 3)

	assertFilterExpression(t, listInput.ModelFilters[0].ValueFilters[0], "partitions",
		"partitions.key = ?", "key1")
	assertFilterExpression(t, listInput.ModelFilters[0].ValueFilters[1], "partitions",
		"partitions.value = ?", "val1")
	assertJoinExpression(t, listInput.ModelFilters[0].JoinCondition, "artifacts", "partitions",
		"p1", "JOIN partitions p1 ON artifacts.artifact_id = p1.artifact_id")

	assertFilterExpression(t, listInput.ModelFilters[1].ValueFilters[0], "partitions",
		"partitions.key = ?", "key2")
	assertFilterExpression(t, listInput.ModelFilters[1].ValueFilters[1], "partitions",
		"partitions.value = ?", "val2")
	assertJoinExpression(t, listInput.ModelFilters[1].JoinCondition, "artifacts", "partitions",
		"p2", "JOIN partitions p2 ON artifacts.artifact_id = p2.artifact_id")

	assertFilterExpression(t, listInput.ModelFilters[2].ValueFilters[0], "tags",
		"tags.tag_name = ?", "special")
	assertJoinExpression(t, listInput.ModelFilters[2].JoinCondition, "artifacts", "tags",
		"t1", "JOIN tags t1 ON artifacts.artifact_id = t1.artifact_id")

}

func TestEmptyFiledListInput(t *testing.T) {
	filter := &datacatalog.FilterExpression{
		Filters: []*datacatalog.SinglePropertyFilter{
			{
				PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
					PartitionFilter: &datacatalog.PartitionPropertyFilter{
						Property: &datacatalog.PartitionPropertyFilter_KeyVal{
							KeyVal: &datacatalog.KeyValuePair{Key: "", Value: ""},
						},
					},
				},
			},
		},
	}
	_, err := FilterToListInput(context.Background(), common.Artifact, filter)
	assert.Error(t, err)
}
