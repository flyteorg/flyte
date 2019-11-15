package gormimpl

import (
	"database/sql/driver"
	"strings"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/utils"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/stretchr/testify/assert"
)

func TestApplyFilter(t *testing.T) {
	testDB := utils.GetDbForTest(t)
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	validInputApply := false

	GlobalMock.NewMock().WithQuery(
		`SELECT "artifacts".* FROM "artifacts"`).WithCallback(
		func(s string, values []driver.NamedValue) {
			// separate the regex matching because the joins reorder on different test runs
			validInputApply = strings.Contains(s, `JOIN tags ON artifacts.artifact_id = tags.artifact_id`) &&
				strings.Contains(s, `JOIN partitions ON artifacts.artifact_id = partitions.artifact_id`) &&
				strings.Contains(s, `WHERE "artifacts"."deleted_at" IS NULL AND `+
					`((partitions.key1 = val1) AND (partitions.key2 = val2) AND (tags.tag_name = special)) `+
					`ORDER BY artifacts.created_at desc LIMIT 10 OFFSET 10`)
		})

	listInput := models.ListModelsInput{
		JoinEntityToConditionMap: map[common.Entity]models.ModelJoinCondition{
			common.Partition: NewGormJoinCondition(common.Artifact, common.Partition),
			common.Tag:       NewGormJoinCondition(common.Artifact, common.Tag),
		},
		Filters: []models.ModelValueFilter{
			NewGormValueFilter(common.Partition, common.Equal, "key1", "val1"),
			NewGormValueFilter(common.Partition, common.Equal, "key2", "val2"),
			NewGormValueFilter(common.Tag, common.Equal, "tag_name", "special"),
		},
		Offset:        10,
		Limit:         10,
		SortParameter: NewGormSortParameter(datacatalog.PaginationOptions_CREATION_TIME, datacatalog.PaginationOptions_DESCENDING),
	}

	tx, err := applyListModelsInput(testDB, common.Artifact, listInput)
	assert.NoError(t, err)

	tx.Find(models.Artifact{})
	assert.True(t, validInputApply)
}

func TestApplyFilterEmpty(t *testing.T) {
	testDB := utils.GetDbForTest(t)
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	validInputApply := false

	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts"  WHERE "artifacts"."deleted_at" IS NULL LIMIT 10 OFFSET 10`).WithCallback(
		func(s string, values []driver.NamedValue) {
			// separate the regex matching because the joins reorder on different test runs
			validInputApply = true
		})

	listInput := models.ListModelsInput{
		JoinEntityToConditionMap: nil,
		Filters:                  nil,
		Offset:                   10,
		Limit:                    10,
	}

	tx, err := applyListModelsInput(testDB, common.Artifact, listInput)
	assert.NoError(t, err)

	tx.Find(models.Artifact{})
	assert.True(t, validInputApply)
}
