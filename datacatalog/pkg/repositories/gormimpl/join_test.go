package gormimpl

import (
	"testing"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestGormJoinCondition(t *testing.T) {
	filter := NewGormJoinCondition(common.Artifact, common.Partition)
	joinQuery, err := filter.GetJoinOnDBQueryExpression("artifacts", "partitions")
	assert.NoError(t, err)
	assert.Equal(t, joinQuery, "JOIN partitions ON artifacts.artifact_id = partitions.artifact_id")
}

// Tag cannot be joined with partitions
func TestInvalidGormJoinCondition(t *testing.T) {
	filter := NewGormJoinCondition(common.Tag, common.Partition)

	_, err := filter.GetJoinOnDBQueryExpression("tags", "partitions")
	assert.Error(t, err)
}
