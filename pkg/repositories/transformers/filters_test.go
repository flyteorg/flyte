package transformers

import (
	"context"
	"testing"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/stretchr/testify/assert"
)

func assertJoinExpression(t *testing.T, listInput models.ListModelsInput, joiningEntity common.Entity, sourceTableName string, joiningTableName string, expectedJoinStatement string) {
	joinCondition, ok := listInput.JoinEntityToConditionMap[joiningEntity]
	assert.True(t, ok)
	expr, err := joinCondition.GetJoinOnDBQueryExpression(sourceTableName, joiningTableName)
	assert.NoError(t, err)
	assert.Equal(t, expr, expectedJoinStatement)
}

func assertFilterExpression(t *testing.T, filter models.ModelValueFilter, expectedEntity common.Entity, tableName string, expectedStatement string, expectedArgs interface{}) {
	assert.Equal(t, filter.GetDBEntity(), expectedEntity)
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

	// Should have 5 filters: 2 for each partition filter, 1 for tag filter
	assert.Len(t, listInput.Filters, 5)

	assertFilterExpression(t, listInput.Filters[0], common.Partition, "partitions", "partitions.key = ?", "key1")
	assertFilterExpression(t, listInput.Filters[1], common.Partition, "partitions", "partitions.value = ?", "val1")
	assertFilterExpression(t, listInput.Filters[2], common.Partition, "partitions", "partitions.key = ?", "key2")
	assertFilterExpression(t, listInput.Filters[3], common.Partition, "partitions", "partitions.value = ?", "val2")
	assertFilterExpression(t, listInput.Filters[4], common.Tag, "tags", "tags.tag_name = ?", "special")

	// even though there are 5 filters, we only need 2 joins on Partition and Tag
	assert.Len(t, listInput.JoinEntityToConditionMap, 2)

	assertJoinExpression(t, listInput, common.Partition, "artifacts", "partitions", "JOIN partitions ON artifacts.artifact_id = partitions.artifact_id")
	assertJoinExpression(t, listInput, common.Tag, "artifacts", "tags", "JOIN tags ON artifacts.artifact_id = tags.artifact_id")

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
