package gormimpl

import (
	"testing"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestGormValueFilter(t *testing.T) {
	filter := NewGormValueFilter(common.Partition, common.Equal, "key", "region")
	expression, err := filter.GetDBQueryExpression("partitions")
	assert.NoError(t, err)
	assert.Equal(t, filter.GetDBEntity(), common.Partition)
	assert.Equal(t, expression.Query, "partitions.key = ?")
	assert.Equal(t, expression.Args, "region")
}

func TestGormValueFilterInvalidOperator(t *testing.T) {
	filter := NewGormValueFilter(common.Partition, 123, "key", "region")
	_, err := filter.GetDBQueryExpression("partitions")
	assert.Error(t, err)
}
